package deviceio

import (
	"fmt"
	"net/http"
	"path/filepath"
	
	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

// ServiceContext holds the configuration and state for the deviceio service
type ServiceContext struct {
	Port        int
	RelayOutputs []RelayOutput
}

// RelayOutput represents a relay output configuration
type RelayOutput struct {
	Name        string
	Token       string
	IdleState   string  // "open" or "closed"
}

// GetServiceCapabilities handles the GetServiceCapabilities ONVIF deviceio service method
func (s *ServiceContext) GetServiceCapabilities() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetServiceCapabilities.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetServiceCapabilities template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetRelayOutputs handles the GetRelayOutputs ONVIF deviceio service method
func (s *ServiceContext) GetRelayOutputs() error {
	// Create relay output elements
	relayElements := make([]string, len(s.RelayOutputs))
	for i, relay := range s.RelayOutputs {
		relayElements[i] = s.createRelayOutputElement(relay)
	}
	
	relaysXML := ""
	if len(relayElements) > 0 {
		relaysXML = relayElements[0] // For simplicity, we're only using the first relay
	}
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%RELAY_OUTPUTS%": relaysXML,
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetRelayOutputs.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetRelayOutputs template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetAudioSources handles the GetAudioSources ONVIF deviceio service method
func (s *ServiceContext) GetAudioSources() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetAudioSources.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetAudioSources template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetAudioOutputs handles the GetAudioOutputs ONVIF deviceio service method
func (s *ServiceContext) GetAudioOutputs() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetAudioOutputs.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetAudioOutputs template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// createRelayOutputElement creates an XML element for a relay output
func (s *ServiceContext) createRelayOutputElement(relay RelayOutput) string {
	// Create relay output element
	relayElement := fmt.Sprintf(`
                <tdio:RelayOutput token="%s">
                    <tt:Name>%s</tt:Name>
                    <tt:RelayMode>Monostable</tt:RelayMode>
                    <tt:IdleState>%s</tt:IdleState>
                </tdio:RelayOutput>`, relay.Token, relay.Name, relay.IdleState)
	
	return relayElement
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF deviceio service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetServiceCapabilities.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetServiceCapabilities template: %v", err)
	}
	
	w.Write([]byte(response))
	return nil
}

// GetRelayOutputsHTTP handles the GetRelayOutputs ONVIF deviceio service method via HTTP
func (s *ServiceContext) GetRelayOutputsHTTP(w http.ResponseWriter) error {
	// Create relay output elements
	relayElements := make([]string, len(s.RelayOutputs))
	for i, relay := range s.RelayOutputs {
		relayElements[i] = s.createRelayOutputElement(relay)
	}
	
	relaysXML := ""
	if len(relayElements) > 0 {
		relaysXML = relayElements[0] // For simplicity, we're only using the first relay
	}
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%RELAY_OUTPUTS%": relaysXML,
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "deviceio", "GetRelayOutputs.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetRelayOutputs template: %v", err)
	}
	
	w.Write([]byte(response))
	return nil
}
