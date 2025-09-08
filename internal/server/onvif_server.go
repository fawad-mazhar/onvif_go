package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/auth"
	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/pkg/services/device"
	"github.com/fawad-mazhar/onvif-go/pkg/services/deviceio"
	"github.com/fawad-mazhar/onvif-go/pkg/services/events"
	"github.com/fawad-mazhar/onvif-go/pkg/services/media"
	"github.com/fawad-mazhar/onvif-go/pkg/services/ptz"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// StartHTTPServer starts the integrated HTTP ONVIF server with Chi router and CORS
func StartHTTPServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)

	// Create Chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify allowed origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "SOAPAction"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	}))

	// Create ONVIF service handlers
	r.Route("/onvif", func(r chi.Router) {
		// Add custom middleware for ONVIF requests
		r.Use(onvifMiddleware(cfg))

		// ONVIF service endpoints
		r.Post("/device_service", createONVIFHandler(cfg, "device_service"))
		r.Post("/media_service", createONVIFHandler(cfg, "media_service"))
		r.Post("/ptz_service", createONVIFHandler(cfg, "ptz_service"))
		r.Post("/events_service", createONVIFHandler(cfg, "events_service"))
		r.Post("/deviceio_service", createONVIFHandler(cfg, "deviceio_service"))
	})

	// Add health check endpoint
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","service":"onvif-server"}`, time.Now().Format(time.RFC3339)); err != nil {
			// Ignore write error for health check endpoint
			_ = err
		}
	})

	// Add info endpoint
	r.Get("/info", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, `{"manufacturer":"%s","model":"%s","firmware":"%s","serial":"%s"}`,
			cfg.Manufacturer, cfg.Model, cfg.FirmwareVer, cfg.SerialNum); err != nil {
			// Ignore write error for info endpoint
			_ = err
		}
	})

	// Start the HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Infof("Starting integrated HTTP ONVIF server on port %d", cfg.Port)
	logger.Infof("ONVIF server listening on address: %s", addr)
	logger.Infof("Health check available at: http://localhost:%d/health", cfg.Port)
	logger.Infof("Device info available at: http://localhost:%d/info", cfg.Port)
	return http.ListenAndServe(addr, r)
}

// handleServiceError handles errors from service calls
func handleServiceError(w http.ResponseWriter, err error, actionName string) {
	if err != nil {
		logger.Errorf("Error handling %s: %v", actionName, err)
		sendSOAPError(w, "Internal server error")
	}
}

// processSOAPRequest routes SOAP requests to the appropriate service handler
func processSOAPRequest(w http.ResponseWriter, r *http.Request, soapAction, serviceName string, cfg *config.ServiceContext) {
	switch serviceName {
	case "device_service":
		switch {
		case strings.Contains(soapAction, "GetServices"):
			deviceService := &device.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, deviceService.GetServicesHTTP(w), "GetServices")
		case strings.Contains(soapAction, "GetDeviceInformation"):
			deviceService := &device.ServiceContext{
				Port:         cfg.Port,
				Manufacturer: cfg.Manufacturer,
				Model:        cfg.Model,
				FirmwareVer:  cfg.FirmwareVer,
				SerialNum:    cfg.SerialNum,
				HardwareID:   cfg.HardwareID,
				Scopes:       cfg.Scopes,
				PTZEnable:    cfg.PTZNode.Enable == 1,
				Media2Enable: cfg.AdvEnableMedia2 == 1,
			}
			handleServiceError(w, deviceService.GetDeviceInformationHTTP(w), "GetDeviceInformation")
		case strings.Contains(soapAction, "GetCapabilities"):
			deviceService := &device.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, deviceService.GetCapabilitiesHTTP(w), "GetCapabilities")
		case strings.Contains(soapAction, "GetScopes"):
			deviceService := &device.ServiceContext{
				Port:   cfg.Port,
				Scopes: cfg.Scopes,
			}
			handleServiceError(w, deviceService.GetScopesHTTP(w), "GetScopes")
		case strings.Contains(soapAction, "SystemReboot"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.SystemRebootHTTP(w), "SystemReboot")
		case strings.Contains(soapAction, "GetSystemDateAndTime"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.GetSystemDateAndTimeHTTP(w), "GetSystemDateAndTime")
		case strings.Contains(soapAction, "GetUsers"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.GetUsersHTTP(w), "GetUsers")
		case strings.Contains(soapAction, "GetWsdlUrl"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.GetWsdlURLHTTP(w), "GetWsdlURL")
		case strings.Contains(soapAction, "GetNetworkInterfaces"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.GetNetworkInterfacesHTTP(w), "GetNetworkInterfaces")
		case strings.Contains(soapAction, "GetDiscoveryMode"):
			deviceService := &device.ServiceContext{}
			handleServiceError(w, deviceService.GetDiscoveryModeHTTP(w), "GetDiscoveryMode")
		default:
			handleUnsupportedSOAPAction(w, soapAction)
		}
	case "media_service":
		switch {
		case strings.Contains(soapAction, "GetServiceCapabilities"):
			mediaService := &media.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, mediaService.GetServiceCapabilitiesHTTP(w), "Media.GetServiceCapabilities")
		case strings.Contains(soapAction, "GetProfiles"):
			mediaService := &media.ServiceContext{
				Port:     cfg.Port,
				Profiles: convertMediaProfiles(cfg.Profiles),
			}
			handleServiceError(w, mediaService.GetProfilesHTTP(w), "Media.GetProfiles")
		default:
			handleUnsupportedSOAPAction(w, soapAction)
		}
	case "ptz_service":
		switch {
		case strings.Contains(soapAction, "GetServiceCapabilities"):
			ptzService := &ptz.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, ptzService.GetServiceCapabilitiesHTTP(w), "PTZ.GetServiceCapabilities")
		case strings.Contains(soapAction, "GetNodes"):
			ptzService := &ptz.ServiceContext{
				Port:     cfg.Port,
				PTZNodes: convertPTZNodes(cfg.PTZNode),
			}
			handleServiceError(w, ptzService.GetNodesHTTP(w), "PTZ.GetNodes")
		default:
			handleUnsupportedSOAPAction(w, soapAction)
		}
	case "events_service":
		eventsService := events.NewServiceContext()
		eventsService.Port = cfg.Port
		eventsService.Events = convertEvents(cfg.Events)
		eventsService.StartEventGenerator() // Start generating test events

		switch {
		case strings.Contains(soapAction, "GetServiceCapabilities"):
			handleServiceError(w, eventsService.GetServiceCapabilitiesHTTP(w), "Events.GetServiceCapabilities")
		case strings.Contains(soapAction, "CreatePullPointSubscription"):
			handleServiceError(w, eventsService.CreatePullPointSubscriptionHTTP(w, r), "Events.CreatePullPointSubscription")
		case strings.Contains(soapAction, "PullMessages"):
			handleServiceError(w, eventsService.PullMessagesHTTP(w, r), "Events.PullMessages")
		case strings.Contains(soapAction, "Subscribe"):
			handleServiceError(w, eventsService.SubscribeHTTP(w, r), "Events.Subscribe")
		case strings.Contains(soapAction, "Renew"):
			handleServiceError(w, eventsService.RenewHTTP(w, r), "Events.Renew")
		case strings.Contains(soapAction, "Unsubscribe"):
			handleServiceError(w, eventsService.UnsubscribeHTTP(w, r), "Events.Unsubscribe")
		case strings.Contains(soapAction, "GetEventProperties"):
			handleServiceError(w, eventsService.GetEventPropertiesHTTP(w), "Events.GetEventProperties")
		case strings.Contains(soapAction, "SetSynchronizationPoint"):
			handleServiceError(w, eventsService.SetSynchronizationPointHTTP(w, r), "Events.SetSynchronizationPoint")
		default:
			handleUnsupportedSOAPAction(w, soapAction)
		}
	case "deviceio_service":
		switch {
		case strings.Contains(soapAction, "GetServiceCapabilities"):
			deviceioService := &deviceio.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, deviceioService.GetServiceCapabilitiesHTTP(w), "DeviceIO.GetServiceCapabilities")
		case strings.Contains(soapAction, "GetRelayOutputs"):
			deviceioService := &deviceio.ServiceContext{
				Port:         cfg.Port,
				RelayOutputs: convertRelayOutputs(cfg.RelayOutputs),
			}
			handleServiceError(w, deviceioService.GetRelayOutputsHTTP(w), "DeviceIO.GetRelayOutputs")
		default:
			handleUnsupportedSOAPAction(w, soapAction)
		}
	default:
		handleUnsupportedService(w, serviceName)
	}
}

// handleUnsupportedSOAPAction handles unsupported SOAP actions with consistent logging
func handleUnsupportedSOAPAction(w http.ResponseWriter, soapAction string) {
	logger.Warnf("Unsupported SOAP action: %s", soapAction)
	sendSOAPError(w, "Unsupported SOAP action")
}

// handleUnsupportedService handles unsupported services with consistent logging
func handleUnsupportedService(w http.ResponseWriter, serviceName string) {
	logger.Warnf("Unsupported service: %s", serviceName)
	sendSOAPError(w, "Unsupported service")
}

// sendSOAPError sends a SOAP fault response for errors
func sendSOAPError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tns="http://www.onvif.org/ver10/device/wsdl">
  <soap:Body>
    <soap:Fault>
      <soap:Code>
        <soap:Value>soap:Receiver</soap:Value>
      </soap:Code>
      <soap:Reason>
        <soap:Text>%s</soap:Text>
      </soap:Reason>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`, message); err != nil {
		// Ignore write error in error handler
		_ = err
	}
}

// sendSOAPAuthError sends a SOAP fault response for authentication errors
func sendSOAPAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tns="http://www.onvif.org/ver10/device/wsdl">
  <soap:Body>
    <soap:Fault>
      <soap:Code>
        <soap:Value>soap:Sender</soap:Value>
      </soap:Code>
      <soap:Reason>
        <soap:Text>Authentication failed</soap:Text>
      </soap:Reason>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`); err != nil {
		// Ignore write error in auth error handler
		_ = err
	}
}

// onvifMiddleware provides ONVIF-specific middleware
func onvifMiddleware(_ *config.ServiceContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set SOAP-specific headers
			w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")

			// Log ONVIF requests
			logger.Debugf("ONVIF request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

			next.ServeHTTP(w, r)
		})
	}
}

// createONVIFHandler creates a handler function for ONVIF services
func createONVIFHandler(cfg *config.ServiceContext, serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the SOAP request from the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warnf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		soapRequest := string(body)
		if soapRequest == "" {
			logger.Warnf("Empty SOAP request")
			sendSOAPError(w, "Empty SOAP request")
			return
		}

		// Parse the SOAP action from the request
		soapAction := parseSOAPAction(soapRequest)
		if soapAction == "" {
			logger.Warnf("Failed to parse SOAP action")
			sendSOAPError(w, "Failed to parse SOAP action")
			return
		}

		// Validate authentication if required
		if cfg.User != "" && cfg.Password != "" {
			usernameToken, err := auth.ParseSOAPHeader(soapRequest)
			if err != nil {
				logger.Warnf("Failed to parse SOAP header: %v", err)
				sendSOAPAuthError(w)
				return
			}

			authContext := &auth.ServiceContext{
				Username: cfg.User,
				Password: cfg.Password,
			}

			if !authContext.ValidateUsernameToken(usernameToken) {
				logger.Warnf("Authentication failed")
				sendSOAPAuthError(w)
				return
			}

			// Validate timestamp to prevent replay attacks
			if !auth.ValidateNonceTimestamp(usernameToken.Created, 300) { // 5 minutes max age
				logger.Warnf("Nonce timestamp validation failed")
				sendSOAPError(w, "Nonce timestamp validation failed")
				return
			}
		}

		// Route the request to the appropriate service handler
		processSOAPRequest(w, r, soapAction, serviceName, cfg)
	}
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
		actionStart += len("<SOAP-ENV:Body>")
	} else {
		actionStart += len("<soap:Body>")
	}

	// Skip whitespace and newlines
	for actionStart < len(soapRequest) && (soapRequest[actionStart] == ' ' ||
		soapRequest[actionStart] == '\n' || soapRequest[actionStart] == '\r' ||
		soapRequest[actionStart] == '\t') {
		actionStart++
	}

	// Find the next opening tag
	if actionStart >= len(soapRequest) || soapRequest[actionStart] != '<' {
		return ""
	}

	actionEnd := strings.Index(soapRequest[actionStart:], ">")
	if actionEnd == -1 {
		return ""
	}

	// Extract the tag name
	tag := soapRequest[actionStart+1 : actionStart+actionEnd]

	// Remove attributes if any
	if strings.Contains(tag, " ") {
		tag = tag[:strings.Index(tag, " ")]
	}

	// Remove namespace prefix for comparison
	if strings.Contains(tag, ":") {
		parts := strings.Split(tag, ":")
		if len(parts) > 1 {
			tag = parts[len(parts)-1]
		}
	}

	// Handle self-closing tags
	tag = strings.TrimSuffix(tag, "/")

	return tag
}

// convertMediaProfiles converts config.StreamProfile to media.Profile
func convertMediaProfiles(configProfiles []config.StreamProfile) []media.Profile {
	mediaProfiles := make([]media.Profile, len(configProfiles))
	for i, cp := range configProfiles {
		mediaProfiles[i] = media.Profile{
			Name:         cp.Name,
			Width:        cp.Width,
			Height:       cp.Height,
			URL:          cp.URL,
			SnapURL:      cp.SnapURL,
			Type:         cp.Type.String(),
			AudioEncoder: cp.AudioEncoder.String(),
			AudioDecoder: cp.AudioDecoder.String(),
		}
	}
	return mediaProfiles
}

// convertPTZNodes converts config.PTZNode to ptz.PTZNode
func convertPTZNodes(configPTZNode config.PTZNode) []ptz.PTZNode {
	// For now, we're creating a single PTZ node from the config
	ptzNode := ptz.PTZNode{
		Name:    "PTZ Node",
		Token:   "PTZToken",
		PTZType: "PanTiltZoom",
		MinPan:  configPTZNode.MinStepX,
		MaxPan:  configPTZNode.MaxStepX,
		MinTilt: configPTZNode.MinStepY,
		MaxTilt: configPTZNode.MaxStepY,
		MinZoom: configPTZNode.MinStepZ,
		MaxZoom: configPTZNode.MaxStepZ,
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
		if cr.IdleState == config.IdleStateOpen {
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
