package media

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	
	"github.com/fawad-mazhar/onvif-go/internal/xml"
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

// GetServiceCapabilities handles the GetServiceCapabilities ONVIF media service method
func (s *ServiceContext) GetServiceCapabilities() error {
	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILE_COUNT%": fmt.Sprintf("%d", len(s.Profiles)),
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "media", "GetServiceCapabilities.xml")
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

// GetProfiles handles the GetProfiles ONVIF media service method
func (s *ServiceContext) GetProfiles() error {
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
	templatePath := filepath.Join("service_files", "media", "GetProfiles.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetProfiles template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetStreamUri handles the GetStreamUri ONVIF media service method
func (s *ServiceContext) GetStreamUri(profileToken string) error {
	// Find the profile by token
	var profile *Profile
	for i, p := range s.Profiles {
		if fmt.Sprintf("Profile%d", i) == profileToken {
			profile = &p
			break
		}
	}
	
	if profile == nil {
		return fmt.Errorf("profile not found: %s", profileToken)
	}
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%STREAM_URL%": profile.URL,
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "media", "GetStreamUri.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetStreamUri template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
}

// GetSnapshotUri handles the GetSnapshotUri ONVIF media service method
func (s *ServiceContext) GetSnapshotUri(profileToken string) error {
	// Find the profile by token
	var profile *Profile
	for i, p := range s.Profiles {
		if fmt.Sprintf("Profile%d", i) == profileToken {
			profile = &p
			break
		}
	}
	
	if profile == nil {
		return fmt.Errorf("profile not found: %s", profileToken)
	}
	
	// Create replacements map for template processing
	replacements := map[string]string{
		"%SNAPSHOT_URL%": profile.SnapURL,
	}
	
	// Process template and write response
	templatePath := filepath.Join("service_files", "media", "GetSnapshotUri.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetSnapshotUri template: %v", err)
	}
	
	xml.WriteResponse(response, "")
	return nil
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
	
	// Create video encoder configuration based on profile type
	videoEncoderConfig := ""
	switch profile.Type {
	case "JPEG":
		videoEncoderConfig = fmt.Sprintf(`
                    <trt:VideoEncoderConfiguration token="%s_veconf">
                        <tt:Name>VideoEncoderConfiguration</tt:Name>
                        <tt:UseCount>1</tt:UseCount>
                        <tt:Encoding>JPEG</tt:Encoding>
                        <tt:Resolution>
                            <tt:Width>%d</tt:Width>
                            <tt:Height>%d</tt:Height>
                        </tt:Resolution>
                        <tt:Quality>5.0</tt:Quality>
                        <tt:RateControl>
                            <tt:FrameRateLimit>25</tt:FrameRateLimit>
                            <tt:EncodingInterval>1</tt:EncodingInterval>
                            <tt:BitrateLimit>10240</tt:BitrateLimit>
                        </tt:RateControl>
                        <tt:Multicast>
                            <tt:Address>
                                <tt:Type>IPv4</tt:Type>
                            </tt:Address>
                            <tt:Port>0</tt:Port>
                            <tt:TTL>1</tt:TTL>
                            <tt:AutoStart>false</tt:AutoStart>
                        </tt:Multicast>
                        <tt:SessionTimeout>PT60S</tt:SessionTimeout>
                    </trt:VideoEncoderConfiguration>`, token, profile.Width, profile.Height)
	case "MPEG4":
		videoEncoderConfig = fmt.Sprintf(`
                    <trt:VideoEncoderConfiguration token="%s_veconf">
                        <tt:Name>VideoEncoderConfiguration</tt:Name>
                        <tt:UseCount>1</tt:UseCount>
                        <tt:Encoding>MPEG4</tt:Encoding>
                        <tt:Resolution>
                            <tt:Width>%d</tt:Width>
                            <tt:Height>%d</tt:Height>
                        </tt:Resolution>
                        <tt:Quality>5.0</tt:Quality>
                        <tt:RateControl>
                            <tt:FrameRateLimit>25</tt:FrameRateLimit>
                            <tt:EncodingInterval>1</tt:EncodingInterval>
                            <tt:BitrateLimit>10240</tt:BitrateLimit>
                        </tt:RateControl>
                        <tt:MPEG4>
                            <tt:GovLength>60</tt:GovLength>
                            <tt:Mpeg4Profile>Main</tt:Mpeg4Profile>
                        </tt:MPEG4>
                        <tt:Multicast>
                            <tt:Address>
                                <tt:Type>IPv4</tt:Type>
                            </tt:Address>
                            <tt:Port>0</tt:Port>
                            <tt:TTL>1</tt:TTL>
                            <tt:AutoStart>false</tt:AutoStart>
                        </tt:Multicast>
                        <tt:SessionTimeout>PT60S</tt:SessionTimeout>
                    </trt:VideoEncoderConfiguration>`, token, profile.Width, profile.Height)
	case "H264":
		videoEncoderConfig = fmt.Sprintf(`
                    <trt:VideoEncoderConfiguration token="%s_veconf">
                        <tt:Name>VideoEncoderConfiguration</tt:Name>
                        <tt:UseCount>1</tt:UseCount>
                        <tt:Encoding>H264</tt:Encoding>
                        <tt:Resolution>
                            <tt:Width>%d</tt:Width>
                            <tt:Height>%d</tt:Height>
                        </tt:Resolution>
                        <tt:Quality>5.0</tt:Quality>
                        <tt:RateControl>
                            <tt:FrameRateLimit>25</tt:FrameRateLimit>
                            <tt:EncodingInterval>1</tt:EncodingInterval>
                            <tt:BitrateLimit>10240</tt:BitrateLimit>
                        </tt:RateControl>
                        <tt:H264>
                            <tt:GovLength>60</tt:GovLength>
                            <tt:H264Profile>Main</tt:H264Profile>
                        </tt:H264>
                        <tt:Multicast>
                            <tt:Address>
                                <tt:Type>IPv4</tt:Type>
                            </tt:Address>
                            <tt:Port>0</tt:Port>
                            <tt:TTL>1</tt:TTL>
                            <tt:AutoStart>false</tt:AutoStart>
                        </tt:Multicast>
                        <tt:SessionTimeout>PT60S</tt:SessionTimeout>
                    </trt:VideoEncoderConfiguration>`, token, profile.Width, profile.Height)
	}
	
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
	templatePath := filepath.Join("service_files", "media", "GetServiceCapabilities.xml")
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
	templatePath := filepath.Join("service_files", "media", "GetProfiles.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}
	
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetProfiles template: %v", err)
	}
	
	w.Write([]byte(response))
	return nil
}
