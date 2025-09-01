package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/auth"
	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/pkg/services/device"
	"github.com/fawad-mazhar/onvif-go/pkg/services/deviceio"
	"github.com/fawad-mazhar/onvif-go/pkg/services/events"
	"github.com/fawad-mazhar/onvif-go/pkg/services/media"
	"github.com/fawad-mazhar/onvif-go/pkg/services/ptz"
)

func StartONVIFServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)
	// logger.SetLevel(logger.INFO) - level already set during initialization
	
	// Configuration is passed as parameter
	// Set up service contexts with converted types
	
	// Parse the request method
	requestMethod := os.Getenv("REQUEST_METHOD")
	if requestMethod != "POST" {
		logger.Warn("Invalid request method: %s", requestMethod)
		handleError("Invalid request method")
		return nil
	}
	
	// Read the SOAP request from stdin
	soapRequest := readSOAPRequest()
	if soapRequest == "" {
		logger.Warn("Empty SOAP request")
		handleError("Empty SOAP request")
		return nil
	}
	
	// Parse the SOAP action from the request
	soapAction := parseSOAPAction(soapRequest)
	if soapAction == "" {
		logger.Warn("Failed to parse SOAP action")
		handleError("Failed to parse SOAP action")
		return nil
	}
	
	// Validate authentication if required
	if cfg.User != "" && cfg.Password != "" {
		usernameToken, err := auth.ParseSOAPHeader(soapRequest)
		if err != nil {
			logger.Warn("Failed to parse SOAP header: %v", err)
			handleAuthError()
			return nil
		}
		
		authContext := &auth.ServiceContext{
			Username: cfg.User,
			Password: cfg.Password,
		}
		
		if !authContext.ValidateUsernameToken(usernameToken) {
			logger.Warn("Authentication failed")
			handleAuthError()
		}
		
		// Validate timestamp to prevent replay attacks
		if !auth.ValidateNonceTimestamp(usernameToken.Created, 300) { // 5 minutes max age
			logger.Warn("Nonce timestamp validation failed")
			return fmt.Errorf("nonce timestamp validation failed")
		}
	}
	
	// Route the request to the appropriate service handler
	switch {
	case strings.Contains(soapAction, "GetServices"):
		deviceService := &device.ServiceContext{
			Port: cfg.Port,
		}
		deviceService.GetServices()
	case strings.Contains(soapAction, "GetDeviceInformation"):
		deviceService := &device.ServiceContext{
			Port:         cfg.Port,
			Manufacturer: cfg.Manufacturer,
			Model:        cfg.Model,
			FirmwareVer:  cfg.FirmwareVer,
			SerialNum:    cfg.SerialNum,
			HardwareId:   cfg.HardwareId,
			Scopes:       cfg.Scopes,
			PTZEnable:    cfg.PTZNode.Enable == 1,
			Media2Enable: cfg.AdvEnableMedia2 == 1,
		}
		deviceService.GetDeviceInformation()
	case strings.Contains(soapAction, "GetCapabilities"):
		deviceService := &device.ServiceContext{
			Port: cfg.Port,
		}
		deviceService.GetCapabilities()
	case strings.Contains(soapAction, "GetScopes"):
		deviceService := &device.ServiceContext{
			Port: cfg.Port,
			Scopes: cfg.Scopes,
		}
		deviceService.GetScopes()
	case strings.Contains(soapAction, "SystemReboot"):
		deviceService := &device.ServiceContext{}
		deviceService.SystemReboot()
	case strings.Contains(soapAction, "GetSystemDateAndTime"):
		deviceService := &device.ServiceContext{}
		deviceService.GetSystemDateAndTime()
	case strings.Contains(soapAction, "GetUsers"):
		deviceService := &device.ServiceContext{}
		deviceService.GetUsers()
	case strings.Contains(soapAction, "GetWsdlUrl"):
		deviceService := &device.ServiceContext{}
		deviceService.GetWsdlUrl()
	case strings.Contains(soapAction, "GetNetworkInterfaces"):
		deviceService := &device.ServiceContext{}
		deviceService.GetNetworkInterfaces()
	case strings.Contains(soapAction, "GetDiscoveryMode"):
		deviceService := &device.ServiceContext{}
		deviceService.GetDiscoveryMode()
	case strings.Contains(soapAction, "GetServiceCapabilities") && strings.Contains(soapAction, "media"):
		mediaService := &media.ServiceContext{
			Port: cfg.Port,
		}
		mediaService.GetServiceCapabilities()
	case strings.Contains(soapAction, "GetProfiles"):
		mediaService := &media.ServiceContext{
			Port: cfg.Port,
			Profiles: convertMediaProfiles(cfg.Profiles),
		}
		mediaService.GetProfiles()
	case strings.Contains(soapAction, "GetServiceCapabilities") && strings.Contains(soapAction, "ptz"):
		ptzService := &ptz.ServiceContext{
			Port: cfg.Port,
		}
		ptzService.GetServiceCapabilities()
	case strings.Contains(soapAction, "GetNodes"):
		ptzService := &ptz.ServiceContext{
			Port: cfg.Port,
			PTZNodes: convertPTZNodes(cfg.PTZNode),
		}
		ptzService.GetNodes()
	case strings.Contains(soapAction, "GetServiceCapabilities") && strings.Contains(soapAction, "events"):
		eventsService := &events.ServiceContext{
			Port: cfg.Port,
		}
		eventsService.GetServiceCapabilities()
	case strings.Contains(soapAction, "GetEventProperties"):
		eventsService := &events.ServiceContext{
			Port: cfg.Port,
			Events: convertEvents(cfg.Events),
		}
		eventsService.GetEventProperties()
	case strings.Contains(soapAction, "GetServiceCapabilities") && strings.Contains(soapAction, "deviceio"):
		deviceioService := &deviceio.ServiceContext{
			Port: cfg.Port,
		}
		deviceioService.GetServiceCapabilities()
	case strings.Contains(soapAction, "GetRelayOutputs"):
		deviceioService := &deviceio.ServiceContext{
			Port: cfg.Port,
			RelayOutputs: convertRelayOutputs(cfg.RelayOutputs),
		}
		deviceioService.GetRelayOutputs()
	default:
		logger.Warn("Unsupported SOAP action: %s", soapAction)
		handleError("Unsupported SOAP action")
	}
	return nil
}

// readSOAPRequest reads the SOAP request from stdin
func readSOAPRequest() string {
	// Get content length from environment
	contentLength := os.Getenv("CONTENT_LENGTH")
	if contentLength == "" {
		return ""
	}
	
	// Read the request body
	data := make([]byte, 1024)
	n, err := os.Stdin.Read(data)
	if err != nil {
		logger.Warn("Failed to read SOAP request: %v", err)
		return ""
	}
	
	return string(data[:n])
}

// parseSOAPAction extracts the SOAP action from the request
func parseSOAPAction(soapRequest string) string {
	// Look for the SOAP action in the request
	actionStart := strings.Index(soapRequest, "<soap:Body>")
	if actionStart == -1 {
		actionStart = strings.Index(soapRequest, "<SOAP-ENV:Body>")
		if actionStart == -1 {
			return ""
		}
	}
	
	// Extract the first tag after <soap:Body> or <SOAP-ENV:Body>
	actionStart += 10
	if strings.Contains(soapRequest[actionStart:], "<soap:Body>") {
		actionStart += 12
	}
	
	actionEnd := strings.Index(soapRequest[actionStart:], ">")
	if actionEnd == -1 {
		return ""
	}
	
	// Extract the tag name
	tag := soapRequest[actionStart:actionStart+actionEnd+1]
	
	// Remove attributes if any
	if strings.Contains(tag, " ") {
		tag = tag[:strings.Index(tag, " ")]
	}
	
	// Remove opening bracket
	tag = strings.TrimPrefix(tag, "<")
	
	// Remove closing bracket or slash
	if strings.HasSuffix(tag, ">") {
		tag = tag[:len(tag)-1]
	} else if strings.HasSuffix(tag, "/") {
		tag = tag[:len(tag)-1]
	}
	
	return tag
}

// handleError sends a generic error response
func handleError(message string) {
	// For now, we'll just send a simple error response
	fmt.Printf("Content-Type: application/soap+xml\r\n\r\n")
	fmt.Printf("<error>%s</error>", message)
}

// handleAuthError sends an authentication error response
func handleAuthError() {
	// For now, we'll just send a simple authentication error response
	fmt.Printf("Content-Type: application/soap+xml\r\n\r\n")
	fmt.Printf("<error>Authentication failed</error>")
}

// convertMediaProfiles converts config.StreamProfile to media.Profile
func convertMediaProfiles(configProfiles []config.StreamProfile) []media.Profile {
	mediaProfiles := make([]media.Profile, len(configProfiles))
	for i, cp := range configProfiles {
		mediaProfiles[i] = media.Profile{
			Name:      cp.Name,
			Width:     cp.Width,
			Height:    cp.Height,
			URL:       cp.URL,
			SnapURL:   cp.SnapURL,
			Type:      cp.Type.String(),
			AudioEncoder: cp.AudioEncoder.String(),
			AudioDecoder: cp.AudioDecoder.String(),
		}
	}
	return mediaProfiles
}

// convertDeviceProfiles converts config.StreamProfile to device.Profile
func convertDeviceProfiles(configProfiles []config.StreamProfile) []device.Profile {
	deviceProfiles := make([]device.Profile, len(configProfiles))
	for i, cp := range configProfiles {
		deviceProfiles[i] = device.Profile{
			Name: cp.Name,
			Type: cp.Type.String(),
		}
	}
	return deviceProfiles
}

// convertPTZNodes converts config.PTZNode to ptz.PTZNode
func convertPTZNodes(configPTZNode config.PTZNode) []ptz.PTZNode {
	// For now, we're creating a single PTZ node from the config
	ptzNode := ptz.PTZNode{
		Name:      "PTZ Node",
		Token:     "PTZToken",
		PTZType:   "PanTiltZoom",
		MinPan:    configPTZNode.MinStepX,
		MaxPan:    configPTZNode.MaxStepX,
		MinTilt:   configPTZNode.MinStepY,
		MaxTilt:   configPTZNode.MaxStepY,
		MinZoom:   configPTZNode.MinStepZ,
		MaxZoom:   configPTZNode.MaxStepZ,
	}
	return []ptz.PTZNode{ptzNode}
}

// convertEvents converts config.Event to events.Event
func convertEvents(configEvents []config.Event) []events.Event {
	eventsList := make([]events.Event, len(configEvents))
	for i, ce := range configEvents {
		eventsList[i] = events.Event{
			Topic:    ce.Topic,
			Producer: ce.SourceName,
		}
	}
	return eventsList
}

// convertRelayOutputs converts config.RelayOutput to deviceio.RelayOutput
func convertRelayOutputs(configRelays []config.RelayOutput) []deviceio.RelayOutput {
	relayOutputs := make([]deviceio.RelayOutput, len(configRelays))
	for i, cr := range configRelays {
		idleState := "closed"
		if cr.IdleState == config.IDLE_STATE_OPEN {
			idleState = "open"
		}
		relayOutputs[i] = deviceio.RelayOutput{
			Name:      "RelayOutput",
			Token:     "RelayToken",
			IdleState: idleState,
		}
	}
	return relayOutputs
}
