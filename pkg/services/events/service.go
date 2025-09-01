package events

import (
	"fmt"
	"path/filepath"
	
	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

// ServiceContext holds the configuration and state for the events service
type ServiceContext struct {
	Port   int
	Events []Event
}

// Event represents an event configuration
type Event struct {
	Topic     string
	Producer  string
}

// GetServiceCapabilities handles the GetServiceCapabilities ONVIF events service method
func (s *ServiceContext) GetServiceCapabilities() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "events", "GetServiceCapabilities.xml")
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

// GetEventProperties handles the GetEventProperties ONVIF events service method
func (s *ServiceContext) GetEventProperties() error {
	// Create event elements
	eventElements := make([]string, len(s.Events))
	for i, event := range s.Events {
		eventElements[i] = s.createEventElement(event)
	}
	
	eventsXML := ""
	if len(eventElements) > 0 {
		eventsXML = eventElements[0] // For simplicity, we're only using the first event
	}
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%EVENTS%": eventsXML,
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "events", "GetEventProperties.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetEventProperties template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// CreatePullPointSubscription handles the CreatePullPointSubscription ONVIF events service method
func (s *ServiceContext) CreatePullPointSubscription() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "events", "CreatePullPointSubscription.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process CreatePullPointSubscription template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// PullMessages handles the PullMessages ONVIF events service method
func (s *ServiceContext) PullMessages() error {
	// Create replacements map for template processing
	replacements := map[string]string{}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "events", "PullMessages.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process PullMessages template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// createEventElement creates an XML element for an event
func (s *ServiceContext) createEventElement(event Event) string {
	// Create event element
	eventElement := fmt.Sprintf(`
                <tev:EventTopic>
                    <tt:Topic>%s</tt:Topic>
                    <tt:Producer>%s</tt:Producer>
                </tev:EventTopic>`, event.Topic, event.Producer)
	
	return eventElement
}
