package device

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/fawad-mazhar/onvif_go/xml"
)

// ServiceContext holds the configuration and state for the device service
type ServiceContext struct {
	Port        int
	Manufacturer string
	Model       string
	FirmwareVer string
	SerialNum   string
	HardwareId  string
	Profiles    []Profile
	Scopes      []string
	PTZEnable   bool
	Media2Enable bool
}

// Profile represents a media profile configuration
type Profile struct {
	Name string
	Type string
}

// GetServices handles the GetServices ONVIF device service method
func (s *ServiceContext) GetServices() error {
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
	templatePath := filepath.Join("device_service_files", "GetServices.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetServices template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetDeviceInformation handles the GetDeviceInformation ONVIF device service method
func (s *ServiceContext) GetDeviceInformation() error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%MANUFACTURER%":      s.Manufacturer,
		"%MODEL%":            s.Model,
		"%FIRMWARE_VERSION%": s.FirmwareVer,
		"%SERIAL_NUMBER%":    s.SerialNum,
		"%HARDWARE_ID%":      s.HardwareId,
	}
	
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetDeviceInformation.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetDeviceInformation template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetCapabilities handles the GetCapabilities ONVIF device service method
func (s *ServiceContext) GetCapabilities() error {
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
		"%MEDIA_SERVICE_URL%": mediaURL,
		"%PTZ_SERVICE_URL%":   ptzURL,
		"%EVENTS_SERVICE_URL%": eventsURL,
	}
	
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetCapabilities.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetCapabilities template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetScopes handles the GetScopes ONVIF device service method
func (s *ServiceContext) GetScopes() error {
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
	templatePath := filepath.Join("device_service_files", "GetScopes.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetScopes template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// SystemReboot handles the SystemReboot ONVIF device service method
func (s *ServiceContext) SystemReboot() error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%MESSAGE%": "Rebooting",
	}
	
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "SystemReboot.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process SystemReboot template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	
	// In a real implementation, we would execute the reboot command
	// For now, we'll just log that reboot was requested
	fmt.Fprintf(os.Stderr, "System reboot requested\n")
	return nil
}

// GetSystemDateAndTime handles the GetSystemDateAndTime ONVIF device service method
func (s *ServiceContext) GetSystemDateAndTime() error {
	// In a real implementation, we would get the actual system time
	// For now, we'll use a placeholder
	currentTime := "2024-01-01T12:00:00Z"
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%CURRENT_TIME%": currentTime,
	}
	
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetSystemDateAndTime.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetSystemDateAndTime template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetUsers handles the GetUsers ONVIF device service method
func (s *ServiceContext) GetUsers() error {
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetUsers.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, nil)
	if err != nil {
		return fmt.Errorf("failed to process GetUsers template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetWsdlUrl handles the GetWsdlUrl ONVIF device service method
func (s *ServiceContext) GetWsdlUrl() error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%WSDL_URL%": "http://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl",
	}
	
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetWsdlUrl.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetWsdlUrl template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetNetworkInterfaces handles the GetNetworkInterfaces ONVIF device service method
func (s *ServiceContext) GetNetworkInterfaces() error {
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetNetworkInterfaces.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, nil)
	if err != nil {
		return fmt.Errorf("failed to process GetNetworkInterfaces template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetDiscoveryMode handles the GetDiscoveryMode ONVIF device service method
func (s *ServiceContext) GetDiscoveryMode() error {
	// Process template and write response
	templatePath := filepath.Join("device_service_files", "GetDiscoveryMode.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, nil)
	if err != nil {
		return fmt.Errorf("failed to process GetDiscoveryMode template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}
