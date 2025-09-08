// Package device provides ONVIF Device service implementation.
package device

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

// ServiceContext holds the configuration and state for the device service
type ServiceContext struct {
	Port         int
	Manufacturer string
	Model        string
	FirmwareVer  string
	SerialNum    string
	HardwareID   string
	Profiles     []Profile
	Scopes       []string
	PTZEnable    bool
	Media2Enable bool
}

// Profile represents a media profile configuration
type Profile struct {
	Name string
	Type string
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServicesHTTP handles the GetServices ONVIF device service method via HTTP
func (s *ServiceContext) GetServicesHTTP(w http.ResponseWriter) error {
	// Determine which services are enabled based on configuration
	services := []string{"Device"}
	if len(s.Profiles) > 0 {
		services = append(services, "Media")
	}
	if s.PTZEnable {
		services = append(services, "PTZ")
	}
	if s.Media2Enable {
		services = append(services, "Media2")
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		"%SERVICES%": strings.Join(services, ","),
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetServices", replacements)
}

// GetDeviceInformationHTTP handles the GetDeviceInformation ONVIF device service method via HTTP
func (s *ServiceContext) GetDeviceInformationHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%MANUFACTURER%":     s.Manufacturer,
		"%MODEL%":            s.Model,
		"%FIRMWARE_VERSION%": s.FirmwareVer,
		"%SERIAL_NUMBER%":    s.SerialNum,
		"%HARDWARE_ID%":      s.HardwareID,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetDeviceInformation", replacements)
}

// GetCapabilitiesHTTP handles the GetCapabilities ONVIF device service method via HTTP
func (s *ServiceContext) GetCapabilitiesHTTP(w http.ResponseWriter) error {
	// Get the host IP address
	hostIP := "127.0.0.1" // Default to localhost
	if s.Port > 0 {
		// In a real implementation, we would get the actual IP
		// This is a simplified version for now
		hostIP = "192.168.1.100"
	}

	// Create service URLs
	deviceURL := fmt.Sprintf("http://%s:%d/onvif/device_service", hostIP, s.Port)
	mediaURL := fmt.Sprintf("http://%s:%d/onvif/media_service", hostIP, s.Port)
	ptzURL := ""
	eventsURL := ""

	if s.PTZEnable {
		ptzURL = fmt.Sprintf("http://%s:%d/onvif/ptz_service", hostIP, s.Port)
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		"%DEVICE_SERVICE_URL%": deviceURL,
		"%MEDIA_SERVICE_URL%":  mediaURL,
		"%PTZ_SERVICE_URL%":    ptzURL,
		"%EVENTS_SERVICE_URL%": eventsURL,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetCapabilities", replacements)
}

// GetScopesHTTP handles the GetScopes ONVIF device service method via HTTP
func (s *ServiceContext) GetScopesHTTP(w http.ResponseWriter) error {
	// Create scope elements
	scopeElements := make([]string, len(s.Scopes))
	for i, scope := range s.Scopes {
		scopeElements[i] = fmt.Sprintf("<tt:ScopeItem>%s</tt:ScopeItem>", scope)
	}

	scopesXML := strings.Join(scopeElements, "\n")

	// Create replacements map for template processing
	replacements := map[string]string{
		"%SCOPES%": scopesXML,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetScopes", replacements)
}

// SystemRebootHTTP handles the SystemReboot ONVIF device service method via HTTP
func (s *ServiceContext) SystemRebootHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%MESSAGE%": "Rebooting",
	}

	// Process template and write response
	err := utils.ProcessServiceTemplate(w, "device", "SystemReboot", replacements)
	if err != nil {
		return err
	}

	// In a real implementation, we would execute the reboot command
	// For now, we'll just log that reboot was requested
	fmt.Fprintf(os.Stderr, "System reboot requested\n")
	return nil
}

// GetSystemDateAndTimeHTTP handles the GetSystemDateAndTime ONVIF device service method via HTTP
func (s *ServiceContext) GetSystemDateAndTimeHTTP(w http.ResponseWriter) error {
	// In a real implementation, we would get the actual system time
	// For now, we'll use a placeholder
	currentTime := "2024-01-01T12:00:00Z"

	// Create replacements map for template processing
	replacements := map[string]string{
		"%CURRENT_TIME%": currentTime,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetSystemDateAndTime", replacements)
}

// GetUsersHTTP handles the GetUsers ONVIF device service method via HTTP
func (s *ServiceContext) GetUsersHTTP(w http.ResponseWriter) error {
	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetUsers", nil)
}

// GetWsdlURLHTTP handles the GetWsdlUrl ONVIF device service method via HTTP
func (s *ServiceContext) GetWsdlURLHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%WSDL_URL%": "http://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl",
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetWsdlUrl", replacements)
}

// GetNetworkInterfacesHTTP handles the GetNetworkInterfaces ONVIF device service method via HTTP
func (s *ServiceContext) GetNetworkInterfacesHTTP(w http.ResponseWriter) error {
	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetNetworkInterfaces", nil)
}

// GetDiscoveryModeHTTP handles the GetDiscoveryMode ONVIF device service method via HTTP
func (s *ServiceContext) GetDiscoveryModeHTTP(w http.ResponseWriter) error {
	// Process template and write response
	return utils.ProcessServiceTemplate(w, "device", "GetDiscoveryMode", nil)
}
