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
	Port       int
	Profiles   []Profile
	PTZEnabled bool
	PTZMinX    float64
	PTZMaxX    float64
	PTZMinY    float64
	PTZMaxY    float64
	PTZMinZ    float64
	PTZMaxZ    float64
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

// buildProfileXML assembles one complete profile entry matching the C reference output.
// outerTag is "trt:Profile" for GetProfile responses, "trt:Profiles" for GetProfiles.
// VSC always uses Profiles[0] dimensions (C reference behavior — one physical source).
// H264Profile is "High" for index 0, "Main" for index 1+ (hardcoded in C reference).
// ASC, AEC, and PTZConfiguration are included only when audio / PTZ are configured.
func (s *ServiceContext) buildProfileXML(profile Profile, token string, index int, outerTag string) string {
	total := len(s.Profiles)
	vscW, vscH := s.Profiles[0].Width, s.Profiles[0].Height

	vsc := fmt.Sprintf(
		`<tt:VideoSourceConfiguration token="VideoSourceConfigToken">`+
			`<tt:Name>VideoSourceConfig</tt:Name>`+
			`<tt:UseCount>%d</tt:UseCount>`+
			`<tt:SourceToken>VideoSourceToken</tt:SourceToken>`+
			`<tt:Bounds x="0" y="0" width="%d" height="%d"/>`+
			`</tt:VideoSourceConfiguration>`,
		total, vscW, vscH,
	)

	var asc string
	if profile.AudioEncoder != "AudioNone" {
		asc = fmt.Sprintf(
			`<tt:AudioSourceConfiguration token="AudioSourceConfigToken">`+
				`<tt:Name>AudioSourceConfig</tt:Name>`+
				`<tt:UseCount>%d</tt:UseCount>`+
				`<tt:SourceToken>AudioSourceToken</tt:SourceToken>`+
				`</tt:AudioSourceConfiguration>`,
			total,
		)
	}

	h264Profile := "High"
	if index > 0 {
		h264Profile = "Main"
	}
	vec := fmt.Sprintf(
		`<tt:VideoEncoderConfiguration token="%s_VideoEncoderToken">`+
			`<tt:Name>%s_VideoEncoder</tt:Name>`+
			`<tt:UseCount>1</tt:UseCount>`+
			`<tt:Encoding>H264</tt:Encoding>`+
			`<tt:Resolution><tt:Width>%d</tt:Width><tt:Height>%d</tt:Height></tt:Resolution>`+
			`<tt:Quality>100</tt:Quality>`+
			`<tt:RateControl><tt:FrameRateLimit>30</tt:FrameRateLimit>`+
			`<tt:EncodingInterval>1</tt:EncodingInterval>`+
			`<tt:BitrateLimit>5000</tt:BitrateLimit></tt:RateControl>`+
			`<tt:H264><tt:GovLength>40</tt:GovLength>`+
			`<tt:H264Profile>%s</tt:H264Profile></tt:H264>`+
			`<tt:Multicast><tt:Address><tt:Type>IPv4</tt:Type></tt:Address>`+
			`<tt:Port>0</tt:Port><tt:TTL>0</tt:TTL>`+
			`<tt:AutoStart>false</tt:AutoStart></tt:Multicast>`+
			`<tt:SessionTimeout>PT0S</tt:SessionTimeout>`+
			`</tt:VideoEncoderConfiguration>`,
		token, token, profile.Width, profile.Height, h264Profile,
	)

	var aec string
	if profile.AudioEncoder != "AudioNone" {
		aec = fmt.Sprintf(
			`<tt:AudioEncoderConfiguration token="%s_AudioEncoderToken">`+
				`<tt:Name>%s_AudioEncoder</tt:Name>`+
				`<tt:UseCount>1</tt:UseCount>`+
				`<tt:Encoding>%s</tt:Encoding>`+
				`<tt:Bitrate>50</tt:Bitrate>`+
				`<tt:SampleRate>16</tt:SampleRate>`+
				`<tt:Multicast><tt:Address><tt:Type>IPv4</tt:Type></tt:Address>`+
				`<tt:Port>0</tt:Port><tt:TTL>0</tt:TTL>`+
				`<tt:AutoStart>false</tt:AutoStart></tt:Multicast>`+
				`<tt:SessionTimeout>PT0S</tt:SessionTimeout>`+
				`</tt:AudioEncoderConfiguration>`,
			token, token, profile.AudioEncoder,
		)
	}

	var ptz string
	if s.PTZEnabled {
		ptz = fmt.Sprintf(
			`<tt:PTZConfiguration token="PTZCfgToken" MoveRamp="0" PresetRamp="0" PresetTourRamp="0">`+
				`<tt:Name>PTZCfg</tt:Name>`+
				`<tt:UseCount>0</tt:UseCount>`+
				`<tt:NodeToken>PTZNodeToken</tt:NodeToken>`+
				`<tt:DefaultAbsolutePantTiltPositionSpace>http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace</tt:DefaultAbsolutePantTiltPositionSpace>`+
				`<tt:DefaultAbsoluteZoomPositionSpace>http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace</tt:DefaultAbsoluteZoomPositionSpace>`+
				`<tt:DefaultRelativePanTiltTranslationSpace>http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace</tt:DefaultRelativePanTiltTranslationSpace>`+
				`<tt:DefaultRelativeZoomTranslationSpace>http://www.onvif.org/ver10/tptz/ZoomSpaces/TranslationGenericSpace</tt:DefaultRelativeZoomTranslationSpace>`+
				`<tt:DefaultContinuousPanTiltVelocitySpace>http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace</tt:DefaultContinuousPanTiltVelocitySpace>`+
				`<tt:DefaultContinuousZoomVelocitySpace>http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace</tt:DefaultContinuousZoomVelocitySpace>`+
				`<tt:DefaultPTZSpeed>`+
				`<tt:PanTilt x="1.0" y="1.0" space="http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"/>`+
				`<tt:Zoom x="1.0" space="http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"/>`+
				`</tt:DefaultPTZSpeed>`+
				`<tt:DefaultPTZTimeout>PT00H00M05S</tt:DefaultPTZTimeout>`+
				`<tt:PanTiltLimits><tt:Range>`+
				`<tt:URI>http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace</tt:URI>`+
				`<tt:XRange><tt:Min>%.1f</tt:Min><tt:Max>%.1f</tt:Max></tt:XRange>`+
				`<tt:YRange><tt:Min>%.1f</tt:Min><tt:Max>%.1f</tt:Max></tt:YRange>`+
				`</tt:Range></tt:PanTiltLimits>`+
				`<tt:ZoomLimits><tt:Range>`+
				`<tt:URI>http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace</tt:URI>`+
				`<tt:XRange><tt:Min>%.1f</tt:Min><tt:Max>%.1f</tt:Max></tt:XRange>`+
				`</tt:Range></tt:ZoomLimits>`+
				`<tt:Extension><tt:PTControlDirection>`+
				`<tt:EFlip><tt:Mode>OFF</tt:Mode></tt:EFlip>`+
				`<tt:Reverse><tt:Mode>OFF</tt:Mode></tt:Reverse>`+
				`</tt:PTControlDirection></tt:Extension>`+
				`</tt:PTZConfiguration>`,
			s.PTZMinX, s.PTZMaxX, s.PTZMinY, s.PTZMaxY,
			s.PTZMinZ, s.PTZMaxZ,
		)
	}

	return fmt.Sprintf(`<%s token="%s" fixed="true"><tt:Name>%s</tt:Name>%s%s%s%s%s</%s>`,
		outerTag, token, profile.Name, vsc, asc, vec, aec, ptz, outerTag,
	)
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
		profileElements[i] = s.buildProfileXML(profile, profile.Name, i, "trt:Profiles")
	}

	profilesXML := strings.Join(profileElements, "")

	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILES%": profilesXML,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", "GetProfiles", replacements)
}

// GetProfileHTTP handles the GetProfile ONVIF media service method via HTTP
func (s *ServiceContext) GetProfileHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	// Extract ProfileToken from SOAP request
	profileToken, _ := xmlfault.ExtractElement([]byte(soapRequest), "ProfileToken")
	if profileToken == "" {
		return sendMediaFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile token does not exist")
	}

	// Find the profile by token
	profile, token, index := s.findProfileByToken(profileToken)
	if profile == nil {
		return sendMediaFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile token does not exist")
	}

	// Create the full profile XML with all configurations
	profileXML := s.buildProfileXML(*profile, token, index, "trt:Profile")

	// Create replacements map for template processing
	replacements := map[string]string{
		"%PROFILE%": profileXML,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", "GetProfile", replacements)
}

// findProfileByToken finds a profile by its name/token and returns profile, token, and index.
// The profile Name is used as token (C reference uses config name as the profile token).
func (s *ServiceContext) findProfileByToken(profileToken string) (*Profile, string, int) {
	for i := range s.Profiles {
		if s.Profiles[i].Name == profileToken {
			return &s.Profiles[i], s.Profiles[i].Name, i
		}
	}
	return nil, "", 0
}

// getMediaUriHTTP is a generic helper for GetStreamUri and GetSnapshotUri
func (s *ServiceContext) getMediaUriHTTP(w http.ResponseWriter, r *http.Request, soapRequest, methodName, urlField, placeholder, errorMsg string) error {
	// Extract ProfileToken from SOAP request
	profileToken, _ := xmlfault.ExtractElement([]byte(soapRequest), "ProfileToken")
	if profileToken == "" {
		return sendMediaFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile does not exist")
	}

	// Find the profile by token
	profile, _, _ := s.findProfileByToken(profileToken)
	if profile == nil {
		return sendMediaFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoProfile", "No profile", "The requested profile does not exist")
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
		return sendMediaFault(w, r, "Receiver", "ter:Action", "ter:IncompleteConfiguration", "Incomplete configuration", errorMsg)
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		placeholder: url,
	}

	// Process template and write response
	return utils.ProcessServiceTemplate(w, "media", methodName, replacements)
}

// GetStreamUriHTTP handles the GetStreamUri ONVIF media service method via HTTP
func (s *ServiceContext) GetStreamUriHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	return s.getMediaUriHTTP(w, r, soapRequest, "GetStreamUri", "URL", "%STREAM_URL%",
		"The specified media profile does not contain either unused sources or encoder configurations without a corresponding source")
}

// GetSnapshotUriHTTP handles the GetSnapshotUri ONVIF media service method via HTTP
func (s *ServiceContext) GetSnapshotUriHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	return s.getMediaUriHTTP(w, r, soapRequest, "GetSnapshotUri", "SnapURL", "%SNAPSHOT_URL%",
		"The specified media profile does not contain either a reference to a video encoder configuration or a reference to a video source configuration")
}

// CreateProfileHTTP handles the CreateProfile ONVIF media service method via HTTP
// Returns a fault as profile creation is not supported (fixed profiles only)
func (s *ServiceContext) CreateProfileHTTP(w http.ResponseWriter, r *http.Request) error {
	return sendMediaFault(w, r, "Receiver", "ter:Action", "ter:MaxNVTProfiles", "Max profile number reached", "The maximum number of supported profiles supported by the device has been reached")
}

// DeleteProfileHTTP handles the DeleteProfile ONVIF media service method via HTTP
// Returns a fault as profile deletion is not supported (fixed profiles only)
func (s *ServiceContext) DeleteProfileHTTP(w http.ResponseWriter, r *http.Request) error {
	return sendMediaFault(w, r, "Receiver", "ter:Action", "ter:FixedProfile", "Fixed profile", "Deletion of fixed profiles is not allowed")
}

// sendMediaFault sends a SOAP fault response for media service errors
// using the shared structured xmlfault renderer (Fault.xml template).
func sendMediaFault(w http.ResponseWriter, r *http.Request, recSend, subcode, subcodeEx, reason, detail string) error {
	devAddr, svcAddr := xmlfault.FaultAddrs(r)
	return xmlfault.WriteFault(w, xmlfault.Fault{
		Service:        "media_service",
		DeviceAddress:  devAddr,
		ServiceAddress: svcAddr,
		RecSend:        recSend,
		Subcode:        subcode,
		SubcodeEx:      subcodeEx,
		Reason:         reason,
		Detail:         detail,
	})
}
