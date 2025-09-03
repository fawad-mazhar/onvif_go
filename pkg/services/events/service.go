package events

import (
	"fmt"
	"net/http"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

// ServiceContext holds the configuration and state for the events service
type ServiceContext struct {
	Port   int
	Events []Event
}

// Event represents an event configuration
type Event struct {
	Topic    string
	Producer string
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

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF events service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "events", "GetServiceCapabilities", replacements)
}

// GetEventPropertiesHTTP handles the GetEventProperties ONVIF events service method via HTTP
func (s *ServiceContext) GetEventPropertiesHTTP(w http.ResponseWriter) error {
	return utils.ProcessCollectionServiceTemplate(
		w,
		s.Events,
		s.createEventElement,
		"%EVENTS%",
		"events",
		"GetEventProperties",
	)
}
