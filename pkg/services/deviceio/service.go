// Package deviceio provides ONVIF Device I/O service implementation.
package deviceio

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

// ServiceContext holds the configuration and state for the deviceio service
type ServiceContext struct {
	Port         int
	RelayOutputs []RelayOutput
	AudioSources int
	AudioOutputs int
}

// RelayOutput represents a relay output configuration
type RelayOutput struct {
	Token     string
	IdleState string // "open" or "closed"
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF deviceio service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "deviceio", "GetServiceCapabilities", map[string]string{
		"%RELAY_OUTPUTS%": fmt.Sprintf("%d", len(s.RelayOutputs)),
		"%AUDIO_OUTPUTS%": fmt.Sprintf("%d", s.AudioOutputs),
		"%AUDIO_SOURCES%": fmt.Sprintf("%d", s.AudioSources),
	})
}

// GetRelayOutputsHTTP handles the GetRelayOutputs ONVIF deviceio service method via HTTP.
// Each relay element uses tds:RelayOutputs + tt:Properties matching the C reference template.
func (s *ServiceContext) GetRelayOutputsHTTP(w http.ResponseWriter) error {
	var sb strings.Builder
	for _, relay := range s.RelayOutputs {
		sb.WriteString(fmt.Sprintf(
			`<tds:RelayOutputs token="%s"><tt:Properties>`+
				`<tt:Mode>Bistable</tt:Mode>`+
				`<tt:DelayTime>PT1S</tt:DelayTime>`+
				`<tt:IdleState>%s</tt:IdleState>`+
				`</tt:Properties></tds:RelayOutputs>`,
			relay.Token, relay.IdleState,
		))
	}
	return utils.ProcessServiceTemplate(w, "deviceio", "GetRelayOutputs", map[string]string{
		"%RELAYS%": sb.String(),
	})
}
