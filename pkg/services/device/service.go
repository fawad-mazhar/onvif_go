// Package device provides ONVIF Device service implementation.
package device

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
	xmlfault "github.com/fawad-mazhar/onvif-go/internal/xml"
)

// EventsEnable mirrors config.EventsEnable without a circular import.
type EventsEnable int

const (
	EventsNone             EventsEnable = 0
	EventsPullPoint        EventsEnable = 1
	EventsBaseSubscription EventsEnable = 2
	EventsBoth             EventsEnable = 3
)

// ServiceContext holds the configuration and state for the device service.
type ServiceContext struct {
	Port            int
	Interface       string
	Manufacturer    string
	Model           string
	FirmwareVer     string
	SerialNum       string
	HardwareID      string
	Scopes          []string
	PTZEnable       bool
	Media2Enable    bool
	EventsEnable    EventsEnable
	AudioSources    int
	AudioOutputs    int
	RelayOutputsNum int
}

// serviceAddrs computes the canonical service URLs for this device.
// Port 80 is omitted from the URL per C reference convention.
func (s *ServiceContext) serviceAddrs() (ip, device, media, ptz, events, deviceio string, err error) {
	ip, err = getInterfaceIP(s.Interface)
	if err != nil {
		return
	}
	port := ""
	if s.Port != 80 {
		port = fmt.Sprintf(":%d", s.Port)
	}
	base := "http://" + ip + port + "/onvif"
	device = base + "/device_service"
	media = base + "/media_service"
	ptz = base + "/ptz_service"
	events = base + "/events_service"
	deviceio = base + "/deviceio_service"
	return
}

// getInterfaceIP returns the first unicast IPv4 address on iface.
// Falls back to 127.0.0.1 when the interface is not found (e.g. "lo" on macOS).
func getInterfaceIP(iface string) (string, error) {
	if iface == "" {
		return "127.0.0.1", nil
	}
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return "127.0.0.1", nil
	}
	addrs, err := ifi.Addrs()
	if err != nil {
		return "", fmt.Errorf("addrs %q: %w", iface, err)
	}
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip4 := ip.To4(); ip4 != nil {
			return ip4.String(), nil
		}
	}
	return "127.0.0.1", nil
}

// eventsFlags returns the epullpoint / ebasesubscription strings for templates.
func (s *ServiceContext) eventsFlags() (pullpoint, basesubscription string) {
	if s.EventsEnable == EventsPullPoint || s.EventsEnable == EventsBoth {
		pullpoint = "true"
	} else {
		pullpoint = "false"
	}
	if s.EventsEnable == EventsBaseSubscription || s.EventsEnable == EventsBoth {
		basesubscription = "true"
	} else {
		basesubscription = "false"
	}
	return
}

// GetServicesHTTP returns a GetServicesResponse selecting the correct
// template variant based on PTZ / Media2 / IncludeCapability flags.
func (s *ServiceContext) GetServicesHTTP(w http.ResponseWriter, soapRequest string) error {
	ip, devAddr, mediaAddr, ptzAddr, eventsAddr, deviceioAddr, err := s.serviceAddrs()
	if err != nil {
		return err
	}

	port := ""
	if s.Port != 80 {
		port = fmt.Sprintf(":%d", s.Port)
	}
	media2Addr := "http://" + ip + port + "/onvif/media2_service"

	includeCapability, _ := xmlfault.ExtractBodyElement([]byte(soapRequest), "IncludeCapability")
	withCap := strings.EqualFold(includeCapability, "true")

	pull, base := s.eventsFlags()

	repl := map[string]string{
		"%DEVICE_SERVICE_ADDRESS%":   devAddr,
		"%MEDIA_SERVICE_ADDRESS%":    mediaAddr,
		"%PTZ_SERVICE_ADDRESS%":      ptzAddr,
		"%MEDIA2_SERVICE_ADDRESS%":   media2Addr,
		"%EVENTS_SERVICE_ADDRESS%":   eventsAddr,
		"%DEVICEIO_SERVICE_ADDRESS%": deviceioAddr,
		"%EVENTS_PULLPOINT%":         pull,
		"%EVENTS_BASESUBSCRIPTION%":  base,
		"%AUDIO_SOURCES%":            strconv.Itoa(s.AudioSources),
		"%AUDIO_OUTPUTS%":            strconv.Itoa(s.AudioOutputs),
		"%RELAY_OUTPUTS%":            strconv.Itoa(s.RelayOutputsNum),
	}

	tmpl := servicesTemplate(s.PTZEnable, s.Media2Enable, withCap)
	return utils.ProcessServiceTemplate(w, "device", tmpl, repl)
}

// servicesTemplate returns the template basename to use for GetServices.
func servicesTemplate(ptz, media2, withCap bool) string {
	prefix := "GetServices"
	if withCap {
		prefix = "GetServices_with_capabilities"
	}
	switch {
	case !ptz && !media2:
		return prefix + "_no_ptz_no_media2"
	case !ptz && media2:
		return prefix + "_no_ptz_media2"
	case ptz && !media2:
		return prefix + "_ptz_no_media2"
	default:
		return prefix + "_ptz_media2"
	}
}

// GetServiceCapabilitiesHTTP returns a static GetServiceCapabilitiesResponse.
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "device", "GetServiceCapabilities", nil)
}

// GetDeviceInformationHTTP returns a GetDeviceInformationResponse.
func (s *ServiceContext) GetDeviceInformationHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "device", "GetDeviceInformation", map[string]string{
		"%MANUFACTURER%":     s.Manufacturer,
		"%MODEL%":            s.Model,
		"%FIRMWARE_VERSION%": s.FirmwareVer,
		"%SERIAL_NUMBER%":    s.SerialNum,
		"%HARDWARE_ID%":      s.HardwareID,
	})
}

// categoryCode maps a GetCapabilities Category string to the C icategory int.
// Returns -1 for unknown/invalid category values (caller should fault).
// Empty string → 15 (All), matching C's "else { icategory = 15 }".
func categoryCode(category string) int {
	switch strings.ToLower(category) {
	case "":
		return 15
	case "device":
		return 1
	case "media":
		return 2
	case "ptz":
		return 4
	case "events":
		return 8
	case "all":
		return 15
	default:
		return -1
	}
}

// GetCapabilitiesHTTP dispatches on the Category element, matching
// device_service.c:444-596. Unknown categories return a SOAP fault.
func (s *ServiceContext) GetCapabilitiesHTTP(w http.ResponseWriter, soapRequest string) error {
	_, devAddr, mediaAddr, ptzAddr, eventsAddr, deviceioAddr, err := s.serviceAddrs()
	if err != nil {
		return err
	}

	pull, base := s.eventsFlags()

	category, _ := xmlfault.ExtractBodyElement([]byte(soapRequest), "Category")
	icategory := categoryCode(category)
	if icategory == -1 {
		return xmlfault.WriteFault(w, xmlfault.Fault{
			Service:   "device_service",
			RecSend:   "Receiver",
			Subcode:   "ter:ActionNotSupported",
			SubcodeEx: "ter:NoSuchService",
			Reason:    "No such service",
			Detail:    "The requested WSDL service category is not supported by the device",
		})
	}

	switch icategory {
	case 1:
		return utils.ProcessServiceTemplate(w, "device", "GetDeviceCapabilities", map[string]string{
			"%DEVICE_SERVICE_ADDRESS%": devAddr,
		})
	case 2:
		return utils.ProcessServiceTemplate(w, "device", "GetMediaCapabilities", map[string]string{
			"%MEDIA_SERVICE_ADDRESS%": mediaAddr,
		})
	case 4:
		if !s.PTZEnable {
			return xmlfault.WriteFault(w, xmlfault.Fault{
				Service:   "device_service",
				RecSend:   "Receiver",
				Subcode:   "ter:ActionNotSupported",
				SubcodeEx: "ter:NoSuchService",
				Reason:    "No such service",
				Detail:    "The requested WSDL service category is not supported by the device",
			})
		}
		return utils.ProcessServiceTemplate(w, "device", "GetPTZCapabilities", map[string]string{
			"%PTZ_SERVICE_ADDRESS%": ptzAddr,
		})
	case 8:
		return utils.ProcessServiceTemplate(w, "device", "GetEventsCapabilities", map[string]string{
			"%EVENTS_SERVICE_ADDRESS%":  eventsAddr,
			"%EVENTS_BASESUBSCRIPTION%": base,
			"%EVENTS_PULLPOINT%":        pull,
		})
	default:
		repl := map[string]string{
			"%DEVICE_SERVICE_ADDRESS%":   devAddr,
			"%MEDIA_SERVICE_ADDRESS%":    mediaAddr,
			"%PTZ_SERVICE_ADDRESS%":      ptzAddr,
			"%EVENTS_SERVICE_ADDRESS%":   eventsAddr,
			"%DEVICEIO_SERVICE_ADDRESS%": deviceioAddr,
			"%EVENTS_PULLPOINT%":         pull,
			"%EVENTS_BASESUBSCRIPTION%":  base,
			"%AUDIO_SOURCES%":            strconv.Itoa(s.AudioSources),
			"%AUDIO_OUTPUTS%":            strconv.Itoa(s.AudioOutputs),
			"%RELAY_OUTPUTS%":            strconv.Itoa(s.RelayOutputsNum),
		}
		tmpl := "GetCapabilities_no_ptz"
		if s.PTZEnable {
			tmpl = "GetCapabilities_ptz"
		}
		return utils.ProcessServiceTemplate(w, "device", tmpl, repl)
	}
}

// GetScopesHTTP returns a GetScopesResponse with C-format scope items.
func (s *ServiceContext) GetScopesHTTP(w http.ResponseWriter) error {
	var sb strings.Builder
	for _, scope := range s.Scopes {
		sb.WriteString("\t    <tds:Scopes>\n")
		sb.WriteString("\t\t<tt:ScopeDef>Fixed</tt:ScopeDef>\n")
		sb.WriteString("\t\t<tt:ScopeItem>")
		sb.WriteString(scope)
		sb.WriteString("</tt:ScopeItem>\n")
		sb.WriteString("\t    </tds:Scopes>\n")
	}
	return utils.ProcessServiceTemplate(w, "device", "GetScopes", map[string]string{
		"%SCOPES%": sb.String(),
	})
}

// SystemRebootHTTP returns a SystemRebootResponse (no substitution; template is static).
func (s *ServiceContext) SystemRebootHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "device", "SystemReboot", nil)
}

// GetSystemDateAndTimeHTTP returns the current UTC date/time using real system clock.
// Uses no zero-padding, matching C's sprintf("%d", ...) output.
func (s *ServiceContext) GetSystemDateAndTimeHTTP(w http.ResponseWriter) error {
	now := time.Now().UTC()
	dst := "false"
	if now.IsDST() {
		dst = "true"
	}
	return utils.ProcessServiceTemplate(w, "device", "GetSystemDateAndTime", map[string]string{
		"%DST%":    dst,
		"%HOUR%":   strconv.Itoa(now.Hour()),
		"%MINUTE%": strconv.Itoa(now.Minute()),
		"%SECOND%": strconv.Itoa(now.Second()),
		"%YEAR%":   strconv.Itoa(now.Year()),
		"%MONTH%":  strconv.Itoa(int(now.Month())),
		"%DAY%":    strconv.Itoa(now.Day()),
	})
}

// GetUsersHTTP returns an empty GetUsersResponse — matches C's send_empty_response.
func (s *ServiceContext) GetUsersHTTP(w http.ResponseWriter) error {
	return xmlfault.WriteEmpty(w, "tds", "GetUsers")
}

// GetWsdlURLHTTP returns the static GetWsdlUrlResponse (no substitution).
func (s *ServiceContext) GetWsdlURLHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "device", "GetWsdlUrl", nil)
}

// GetNetworkInterfacesHTTP returns the network interface details for s.Interface.
func (s *ServiceContext) GetNetworkInterfacesHTTP(w http.ResponseWriter) error {
	iface := s.Interface
	if iface == "" {
		iface = "lo"
	}

	ipStr, prefixLen, mac, mtu, err := getInterfaceDetails(iface)
	if err != nil {
		ipStr, prefixLen, mac, mtu = "127.0.0.1", "8", "", "65536"
	}

	return utils.ProcessServiceTemplate(w, "device", "GetNetworkInterfaces", map[string]string{
		"%INTERFACE%":   iface,
		"%MAC_ADDRESS%": mac,
		"%MTU%":         mtu,
		"%IP_ADDRESS%":  ipStr,
		"%NETMASK%":     prefixLen,
	})
}

// getInterfaceDetails returns ip, cidr-prefix-len, MAC, MTU for iface.
func getInterfaceDetails(iface string) (ip, prefixLen, mac, mtu string, err error) {
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return "", "", "", "", fmt.Errorf("interface %q: %w", iface, err)
	}
	mtu = strconv.Itoa(ifi.MTU)
	if hw := ifi.HardwareAddr; len(hw) > 0 {
		mac = hw.String()
	}
	addrs, err := ifi.Addrs()
	if err != nil {
		return "", "", mac, mtu, fmt.Errorf("addrs %q: %w", iface, err)
	}
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				ones, _ := ipNet.Mask.Size()
				return ip4.String(), strconv.Itoa(ones), mac, mtu, nil
			}
		}
	}
	return "127.0.0.1", "8", mac, mtu, nil
}

// GetDiscoveryModeHTTP returns the static GetDiscoveryModeResponse.
func (s *ServiceContext) GetDiscoveryModeHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "device", "GetDiscoveryMode", nil)
}
