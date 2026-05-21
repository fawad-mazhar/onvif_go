// Package media provides ONVIF Media service implementation.
package media

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/utils"
	xmlfault "github.com/fawad-mazhar/onvif-go/internal/xml"
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

// GetProfileHTTP handles the GetProfile ONVIF media service method via HTTP
func (s *ServiceContext) GetProfileHTTP(w http.ResponseWriter, soapRequest string) error {
	// Extract ProfileToken from SOAP request
	profileToken, _ := xmlfault.ExtractElement([]byte(soapRequest), "ProfileToken")
	if profileToken == "" {
		return sendMediaFault(w, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile token does not exist")
	}

	// Find the profile by token
	profile, token := s.findProfileByToken(profileToken)
	if profile == nil {
		return sendMediaFault(w, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile token does not exist")
	}

	// Create the full profile XML with all configurations
	profileXML := s.createProfileElement(*profile, token)

	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILE%": profileXML,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", "GetProfile", replacements)
}

// findProfileByToken finds a profile by its token or name and returns both profile and token
func (s *ServiceContext) findProfileByToken(profileToken string) (*Profile, string) {
	for i, p := range s.Profiles {
		token := fmt.Sprintf("Profile%d", i)
		if token == profileToken || p.Name == profileToken {
			return &p, token
		}
	}
	return nil, ""
}

// getMediaUriHTTP is a generic helper for GetStreamUri and GetSnapshotUri
func (s *ServiceContext) getMediaUriHTTP(w http.ResponseWriter, soapRequest, methodName, urlField, placeholder, errorMsg string) error {
	// Extract ProfileToken from SOAP request
	profileToken, _ := xmlfault.ExtractElement([]byte(soapRequest), "ProfileToken")
	if profileToken == "" {
		return sendMediaFault(w, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile does not exist")
	}

	// Find the profile by token
	profile, _ := s.findProfileByToken(profileToken)
	if profile == nil {
		return sendMediaFault(w, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile does not exist")
	}

	// Get the URL field value using reflection-free approach
	var url string
	switch urlField {
	case "URL":
		url = profile.URL
	case "SnapURL":
		url = profile.SnapURL
	}

	// Check if URL is configured
	if url == "" {
		return sendMediaFault(w, "Receiver", "ter:Action", "ter:IncompleteConfiguration", "Incomplete configuration", errorMsg)
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		placeholder: url,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", methodName, replacements)
}

// GetStreamUriHTTP handles the GetStreamUri ONVIF media service method via HTTP
func (s *ServiceContext) GetStreamUriHTTP(w http.ResponseWriter, soapRequest string) error {
	return s.getMediaUriHTTP(w, soapRequest, "GetStreamUri", "URL", "%STREAM_URL%",
		"The specified media profile does not contain either unused sources or encoder configurations without a corresponding source")
}

// GetSnapshotUriHTTP handles the GetSnapshotUri ONVIF media service method via HTTP
func (s *ServiceContext) GetSnapshotUriHTTP(w http.ResponseWriter, soapRequest string) error {
	return s.getMediaUriHTTP(w, soapRequest, "GetSnapshotUri", "SnapURL", "%SNAPSHOT_URL%",
		"The specified media profile does not contain either a reference to a video encoder configuration or a reference to a video source configuration")
}

// CreateProfileHTTP handles the CreateProfile ONVIF media service method via HTTP
// Returns a fault as profile creation is not supported (fixed profiles only)
func (s *ServiceContext) CreateProfileHTTP(w http.ResponseWriter) error {
	return sendMediaFault(w, "Receiver", "ter:Action", "ter:MaxNVTProfiles", "Max profile number reached", "The maximum number of supported profiles supported by the device has been reached")
}

// DeleteProfileHTTP handles the DeleteProfile ONVIF media service method via HTTP
// Returns a fault as profile deletion is not supported (fixed profiles only)
func (s *ServiceContext) DeleteProfileHTTP(w http.ResponseWriter) error {
	return sendMediaFault(w, "Receiver", "ter:Action", "ter:FixedProfile", "Fixed profile", "Deletion of fixed profiles is not allowed")
}

// sendMediaFault sends a SOAP fault response for media service errors
// using the shared structured xmlfault renderer (Fault.xml template).
func sendMediaFault(w http.ResponseWriter, recSend, subcode, subcodeEx, reason, detail string) error {
	return xmlfault.WriteFault(w, xmlfault.Fault{
		Service:   "media_service",
		RecSend:   recSend,
		Subcode:   subcode,
		SubcodeEx: subcodeEx,
		Reason:    reason,
		Detail:    detail,
	})
}
