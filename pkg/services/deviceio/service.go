// Package deviceio provides ONVIF Device I/O service implementation.
package deviceio

import (
	"fmt"
	"net/http"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

// ServiceContext holds the configuration and state for the deviceio service
type ServiceContext struct {
	Port         int
	RelayOutputs []RelayOutput
}

// RelayOutput represents a relay output configuration
type RelayOutput struct {
	Name      string
	Token     string
	IdleState string // "open" or "closed"
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
	return utils.ProcessServiceTemplate(w, "deviceio", "GetServiceCapabilities", replacements)
}

// GetRelayOutputsHTTP handles the GetRelayOutputs ONVIF deviceio service method via HTTP
func (s *ServiceContext) GetRelayOutputsHTTP(w http.ResponseWriter) error {
	return utils.ProcessCollectionServiceTemplate(
		w,
		s.RelayOutputs,
		s.createRelayOutputElement,
		"%RELAY_OUTPUTS%",
		"deviceio",
		"GetRelayOutputs",
	)
}
