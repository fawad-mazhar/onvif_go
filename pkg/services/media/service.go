// Package media provides ONVIF Media service implementation.
package media

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

// ServiceContext holds the configuration and state for the media service
type ServiceContext struct {
	Port     int
	Profiles []Profile
}

// Profile represents a media profile configuration
type Profile struct {
	Name         string
	Width        int
	Height       int
	URL          string
	SnapURL      string
	Type         string
	AudioEncoder string
	AudioDecoder string
}

// createProfileElement creates an XML element for a profile
func (s *ServiceContext) createProfileElement(profile Profile, token string) string {
	// Create video source configuration
	videoSourceConfig := fmt.Sprintf(`
                    <trt:VideoSourceConfiguration token="%s_vsconf">
                        <tt:Name>VideoSourceConfiguration</tt:Name>
                        <tt:UseCount>1</tt:UseCount>
                        <tt:SourceToken>VideoSource</tt:SourceToken>
                        <tt:Bounds x="0" y="0" width="%d" height="%d"/>
                    </trt:VideoSourceConfiguration>`, token, profile.Width, profile.Height)

	// Create video encoder configuration using shared utility
	videoEncoderConfig := utils.CreateVideoEncoderConfig(profile.Type, token, profile.Width, profile.Height)

	// Create profile element
	profileElement := fmt.Sprintf(`
                <trt:Profiles token="%s" fixed="true">
                    <tt:Name>%s</tt:Name>
                    %s
                    %s
                </trt:Profiles>`, token, profile.Name, videoSourceConfig, videoEncoderConfig)

	return profileElement
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF media service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILE_COUNT%": fmt.Sprintf("%d", len(s.Profiles)),
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", "GetServiceCapabilities", replacements)
}

// GetProfilesHTTP handles the GetProfiles ONVIF media service method via HTTP
func (s *ServiceContext) GetProfilesHTTP(w http.ResponseWriter) error {
	// Create profile elements
	profileElements := make([]string, len(s.Profiles))
	for i, profile := range s.Profiles {
		profileElements[i] = s.createProfileElement(profile, fmt.Sprintf("Profile%d", i))
	}

	profilesXML := strings.Join(profileElements, "\n")

	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILES%": profilesXML,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", "GetProfiles", replacements)
}
