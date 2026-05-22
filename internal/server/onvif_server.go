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
	xmlfault "github.com/fawad-mazhar/onvif-go/internal/xml"
	"github.com/fawad-mazhar/onvif-go/pkg/services/device"
	"github.com/fawad-mazhar/onvif-go/pkg/services/deviceio"
	"github.com/fawad-mazhar/onvif-go/pkg/services/events"
	"github.com/fawad-mazhar/onvif-go/pkg/services/media"
	"github.com/fawad-mazhar/onvif-go/pkg/services/ptz"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	// RequestTimeoutSeconds defines the HTTP request timeout in seconds
	RequestTimeoutSeconds = 60
	// CORSMaxAgeSeconds defines the CORS preflight cache duration in seconds
	CORSMaxAgeSeconds = 300
	// NonceMaxAgeSeconds defines the maximum age for nonce validation in seconds (5 minutes)
	NonceMaxAgeSeconds = 300
)

// BuildRouter wires all ONVIF routes and middleware onto a Chi router,
// returning it without starting the listener. Exposed for test harnesses
// (see test/golden_diff_test.go).
func BuildRouter(cfg *config.ServiceContext) chi.Router {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(RequestTimeoutSeconds * time.Second))

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify allowed origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "SOAPAction"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           CORSMaxAgeSeconds, // Maximum value not ignored by any major browsers
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

	return r
}

// StartHTTPServer builds the router and starts listening on cfg.Port.
// Thin wrapper around BuildRouter for production main().
func StartHTTPServer(cfg *config.ServiceContext) error {
	r := BuildRouter(cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Infof("Starting integrated HTTP ONVIF server on port %d", cfg.Port)
	logger.Infof("ONVIF server listening on address: %s", addr)
	logger.Infof("Health check available at: http://localhost:%d/health", cfg.Port)
	logger.Infof("Device info available at: http://localhost:%d/info", cfg.Port)
	return http.ListenAndServe(addr, r)
}

// handleServiceError handles errors from service calls
func handleServiceError(w http.ResponseWriter, r *http.Request, err error, actionName string) {
	if err != nil {
		logger.Errorf("Error handling %s: %v", actionName, err)
		sendSOAPError(w, r, "Internal server error")
	}
}

// buildDeviceService maps a config.ServiceContext to a device.ServiceContext.
// Audio source/output counts are derived from the profile list to match the
// C reference's runtime computation.
func buildDeviceService(cfg *config.ServiceContext) *device.ServiceContext {
	audioSources, audioOutputs := 0, 0
	for _, p := range cfg.Profiles {
		if p.AudioEncoder != config.AudioNone {
			audioSources = 1
		}
		if p.AudioDecoder != config.AudioNone {
			audioOutputs = 1
		}
	}
	return &device.ServiceContext{
		Port:            cfg.Port,
		Interface:       cfg.Interface,
		Manufacturer:    cfg.Manufacturer,
		Model:           cfg.Model,
		FirmwareVer:     cfg.FirmwareVer,
		SerialNum:       cfg.SerialNum,
		HardwareID:      cfg.HardwareID,
		Scopes:          cfg.Scopes,
		PTZEnable:       cfg.PTZNode.Enable == 1,
		Media2Enable:    cfg.AdvEnableMedia2 == 1,
		EventsEnable:    device.EventsEnable(cfg.EventsEnable),
		AudioSources:    audioSources,
		AudioOutputs:    audioOutputs,
		RelayOutputsNum: cfg.RelayOutputsNum,
	}
}

// processSOAPRequest routes SOAP requests to the appropriate service handler
func processSOAPRequest(w http.ResponseWriter, r *http.Request, soapRequest, soapAction, serviceName string, cfg *config.ServiceContext) {
	switch serviceName {
	case "device_service":
		ds := buildDeviceService(cfg)
		switch soapAction {
		case "GetServices":
			handleServiceError(w, r, ds.GetServicesHTTP(w, soapRequest), "GetServices")
		case "GetServiceCapabilities":
			handleServiceError(w, r, ds.GetServiceCapabilitiesHTTP(w), "GetServiceCapabilities")
		case "GetDeviceInformation":
			handleServiceError(w, r, ds.GetDeviceInformationHTTP(w), "GetDeviceInformation")
		case "GetCapabilities":
			handleServiceError(w, r, ds.GetCapabilitiesHTTP(w, r, soapRequest), "GetCapabilities")
		case "GetScopes":
			handleServiceError(w, r, ds.GetScopesHTTP(w), "GetScopes")
		case "SystemReboot":
			handleServiceError(w, r, ds.SystemRebootHTTP(w), "SystemReboot")
		case "GetSystemDateAndTime":
			handleServiceError(w, r, ds.GetSystemDateAndTimeHTTP(w), "GetSystemDateAndTime")
		case "GetUsers":
			handleServiceError(w, r, ds.GetUsersHTTP(w), "GetUsers")
		case "GetWsdlUrl":
			handleServiceError(w, r, ds.GetWsdlURLHTTP(w), "GetWsdlUrl")
		case "GetNetworkInterfaces":
			handleServiceError(w, r, ds.GetNetworkInterfacesHTTP(w), "GetNetworkInterfaces")
		case "GetDiscoveryMode":
			handleServiceError(w, r, ds.GetDiscoveryModeHTTP(w), "GetDiscoveryMode")
		default:
			sendUnsupportedResponse(w, r, cfg, "tds", soapAction)
		}
	case "media_service":
		switch {
		// adv_fault_if_set: these Set* ops fault unconditionally when the flag is set;
		// otherwise they fall through to the default unsupported path.
		case cfg.AdvFaultIfSet == 1 && (soapAction == "SetVideoSourceConfiguration" ||
			soapAction == "SetAudioSourceConfiguration" ||
			soapAction == "SetVideoEncoderConfiguration" ||
			soapAction == "SetAudioEncoderConfiguration" ||
			soapAction == "SetAudioOutputConfiguration"):
			sendSOAPError(w, r, "Action failed")
		case soapAction == "GetServiceCapabilities":
			mediaService := &media.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, r, mediaService.GetServiceCapabilitiesHTTP(w), "Media.GetServiceCapabilities")
		case soapAction == "GetProfiles":
			mediaService := &media.ServiceContext{
				Port:     cfg.Port,
				Profiles: convertMediaProfiles(cfg.Profiles),
			}
			handleServiceError(w, r, mediaService.GetProfilesHTTP(w), "Media.GetProfiles")
		case soapAction == "GetProfile":
			mediaService := &media.ServiceContext{
				Port:     cfg.Port,
				Profiles: convertMediaProfiles(cfg.Profiles),
			}
			handleServiceError(w, r, mediaService.GetProfileHTTP(w, r, soapRequest), "Media.GetProfile")
		case soapAction == "GetStreamUri":
			mediaService := &media.ServiceContext{
				Port:     cfg.Port,
				Profiles: convertMediaProfiles(cfg.Profiles),
			}
			handleServiceError(w, r, mediaService.GetStreamUriHTTP(w, r, soapRequest), "Media.GetStreamUri")
		case soapAction == "GetSnapshotUri":
			mediaService := &media.ServiceContext{
				Port:     cfg.Port,
				Profiles: convertMediaProfiles(cfg.Profiles),
			}
			handleServiceError(w, r, mediaService.GetSnapshotUriHTTP(w, r, soapRequest), "Media.GetSnapshotUri")
		case soapAction == "CreateProfile":
			mediaService := &media.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, r, mediaService.CreateProfileHTTP(w, r), "Media.CreateProfile")
		default:
			sendUnsupportedResponse(w, r, cfg, "trt", soapAction)
		}
	case "ptz_service":
		ptzService := &ptz.ServiceContext{
			Port: cfg.Port,
			Node: buildPTZNode(cfg.PTZNode),
		}
		switch {
		case soapAction == "GetServiceCapabilities":
			handleServiceError(w, r, ptzService.GetServiceCapabilitiesHTTP(w), "PTZ.GetServiceCapabilities")
		case soapAction == "GetNodes":
			handleServiceError(w, r, ptzService.GetNodesHTTP(w), "PTZ.GetNodes")
		case soapAction == "GetNode":
			handleServiceError(w, r, ptzService.GetNodeHTTP(w, r, soapRequest), "PTZ.GetNode")
		case soapAction == "GetConfigurations":
			handleServiceError(w, r, ptzService.GetConfigurationsHTTP(w), "PTZ.GetConfigurations")
		case soapAction == "GetConfiguration":
			handleServiceError(w, r, ptzService.GetConfigurationHTTP(w), "PTZ.GetConfiguration")
		case soapAction == "GetConfigurationOptions":
			handleServiceError(w, r, ptzService.GetConfigurationOptionsHTTP(w), "PTZ.GetConfigurationOptions")
		case soapAction == "GetPresets":
			handleServiceError(w, r, ptzService.GetPresetsHTTP(w, r, soapRequest), "PTZ.GetPresets")
		case soapAction == "GotoPreset":
			handleServiceError(w, r, ptzService.GotoPresetHTTP(w, r, soapRequest), "PTZ.GotoPreset")
		case soapAction == "GotoHomePosition":
			handleServiceError(w, r, ptzService.GotoHomePositionHTTP(w, r, soapRequest), "PTZ.GotoHomePosition")
		case soapAction == "ContinuousMove":
			handleServiceError(w, r, ptzService.ContinuousMoveHTTP(w, r, soapRequest), "PTZ.ContinuousMove")
		case soapAction == "RelativeMove":
			handleServiceError(w, r, ptzService.RelativeMoveHTTP(w, r, soapRequest), "PTZ.RelativeMove")
		case soapAction == "AbsoluteMove":
			handleServiceError(w, r, ptzService.AbsoluteMoveHTTP(w, r, soapRequest), "PTZ.AbsoluteMove")
		case soapAction == "Stop":
			handleServiceError(w, r, ptzService.StopHTTP(w, r, soapRequest), "PTZ.Stop")
		case soapAction == "GetStatus":
			handleServiceError(w, r, ptzService.GetStatusHTTP(w, r, soapRequest), "PTZ.GetStatus")
		case soapAction == "SetPreset":
			handleServiceError(w, r, ptzService.SetPresetHTTP(w, r, soapRequest), "PTZ.SetPreset")
		case soapAction == "RemovePreset":
			handleServiceError(w, r, ptzService.RemovePresetHTTP(w, r, soapRequest), "PTZ.RemovePreset")
		case soapAction == "SetHomePosition":
			handleServiceError(w, r, ptzService.SetHomePositionHTTP(w, r, soapRequest), "PTZ.SetHomePosition")
		default:
			sendUnsupportedResponse(w, r, cfg, "tptz", soapAction)
		}
	case "events_service":
		eventsService := events.NewServiceContext()
		eventsService.Port = cfg.Port
		eventsService.Events = convertEvents(cfg.Events)
		eventsService.StartEventGenerator() // Start generating test events

		switch {
		case soapAction == "GetServiceCapabilities":
			handleServiceError(w, r, eventsService.GetServiceCapabilitiesHTTP(w), "Events.GetServiceCapabilities")
		case soapAction == "CreatePullPointSubscription":
			handleServiceError(w, r, eventsService.CreatePullPointSubscriptionHTTP(w, r), "Events.CreatePullPointSubscription")
		case soapAction == "PullMessages":
			handleServiceError(w, r, eventsService.PullMessagesHTTP(w, r), "Events.PullMessages")
		case soapAction == "Subscribe":
			handleServiceError(w, r, eventsService.SubscribeHTTP(w, r), "Events.Subscribe")
		case soapAction == "Renew":
			handleServiceError(w, r, eventsService.RenewHTTP(w, r), "Events.Renew")
		case soapAction == "Unsubscribe":
			handleServiceError(w, r, eventsService.UnsubscribeHTTP(w, r), "Events.Unsubscribe")
		case soapAction == "GetEventProperties":
			handleServiceError(w, r, eventsService.GetEventPropertiesHTTP(w), "Events.GetEventProperties")
		case soapAction == "SetSynchronizationPoint":
			handleServiceError(w, r, eventsService.SetSynchronizationPointHTTP(w, r), "Events.SetSynchronizationPoint")
		default:
			sendUnsupportedResponse(w, r, cfg, "tev", soapAction)
		}
	case "deviceio_service":
		switch {
		case soapAction == "GetServiceCapabilities":
			deviceioService := &deviceio.ServiceContext{
				Port: cfg.Port,
			}
			handleServiceError(w, r, deviceioService.GetServiceCapabilitiesHTTP(w), "DeviceIO.GetServiceCapabilities")
		case soapAction == "GetRelayOutputs":
			deviceioService := &deviceio.ServiceContext{
				Port:         cfg.Port,
				RelayOutputs: convertRelayOutputs(cfg.RelayOutputs),
			}
			handleServiceError(w, r, deviceioService.GetRelayOutputsHTTP(w), "DeviceIO.GetRelayOutputs")
		default:
			sendUnsupportedResponse(w, r, cfg, "tmd", soapAction)
		}
	default:
		handleUnsupportedService(w, r, serviceName)
	}
}

// sendUnsupportedResponse implements C's <svc>_unsupported() semantics:
//   - adv_fault_if_unknown == 0 (default): send an empty 200 response,
//     identical to C's send_empty_response(ns, method).
//   - adv_fault_if_unknown == 1: send an action-failed SOAP fault,
//     identical to C's send_action_failed_fault(service, -1).
func sendUnsupportedResponse(w http.ResponseWriter, r *http.Request, cfg *config.ServiceContext, ns, action string) {
	logger.Warnf("Unsupported SOAP action: %s", action)
	if cfg.AdvFaultIfUnknown == 1 {
		sendSOAPError(w, r, "Action failed")
		return
	}
	body, status, err := xmlfault.RenderEmpty(ns, action)
	if err != nil {
		logger.Warnf("sendUnsupportedResponse RenderEmpty: %v", err)
		sendSOAPError(w, r, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// handleUnsupportedService handles unsupported services with consistent logging
func handleUnsupportedService(w http.ResponseWriter, r *http.Request, serviceName string) {
	logger.Warnf("Unsupported service: %s", serviceName)
	sendSOAPError(w, r, "Unsupported service")
}

// sendSOAPError emits an ONVIF-spec-compliant SOAP fault via the
// structured xmlfault library. Byte-compatible with the C reference
// server's send_fault() generic path.
//
// G-005 (Phase 3): DeviceAddress and ServiceAddress are now derived
// from r.Host + r.URL.Path so Fault.xml's %ADDRESS%/%SERVICE%
// placeholders are filled correctly instead of being left empty.
func sendSOAPError(w http.ResponseWriter, r *http.Request, message string) {
	devAddr, svcAddr := xmlfault.FaultAddrs(r)
	if err := xmlfault.WriteFault(w, xmlfault.Fault{
		DeviceAddress:  devAddr,
		ServiceAddress: svcAddr,
		RecSend:        "Receiver",
		Subcode:        "ter:Action",
		SubcodeEx:      "ter:ActionFailed",
		Reason:         "Action failed",
		Detail:         message,
	}); err != nil {
		logger.Warnf("sendSOAPError: %v", err)
	}
}

// sendSOAPAuthError emits the standard WS-Security auth-failed fault
// using the AuthenticationError.xml template (byte-compatible with
// the C reference's send_authentication_error()).
func sendSOAPAuthError(w http.ResponseWriter) {
	if err := xmlfault.WriteAuthenticationError(w); err != nil {
		logger.Warnf("sendSOAPAuthError: %v", err)
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
			sendSOAPError(w, r, "Empty SOAP request")
			return
		}

		// Parse the SOAP action from the request
		soapAction := parseSOAPAction(soapRequest)
		if soapAction == "" {
			logger.Warnf("Failed to parse SOAP action")
			sendSOAPError(w, r, "Failed to parse SOAP action")
			return
		}

		// G-001: device_service ops exempt from auth (matches C reference strcasecmp).
		authExempt := serviceName == "device_service" &&
			(strings.EqualFold(soapAction, "GetSystemDateAndTime") ||
				strings.EqualFold(soapAction, "GetUsers") ||
				strings.EqualFold(soapAction, "GetCapabilities") ||
				strings.EqualFold(soapAction, "GetServices") ||
				strings.EqualFold(soapAction, "GetServiceCapabilities"))

		// Validate authentication if required
		if cfg.User != "" && cfg.Password != "" && !authExempt {
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
			if !auth.ValidateNonceTimestamp(usernameToken.Created, NonceMaxAgeSeconds) { // 5 minutes max age
				logger.Warnf("Nonce timestamp validation failed")
				sendSOAPError(w, r, "Nonce timestamp validation failed")
				return
			}
		}

		// Route the request to the appropriate service handler
		processSOAPRequest(w, r, soapRequest, soapAction, serviceName, cfg)
	}
}

// parseSOAPAction extracts the local-name of the first child element
// inside soap:Body from a SOAP request. Delegates to the namespace-aware
// xmlfault.ExtractBodyAction so any envelope prefix (s:, soap:,
// SOAP-ENV:, env:) is accepted. Returns "" if extraction fails.
func parseSOAPAction(soapRequest string) string {
	action, err := xmlfault.ExtractBodyAction([]byte(soapRequest))
	if err != nil {
		logger.Debugf("ExtractBodyAction: %v", err)
		return ""
	}
	return action
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

// buildPTZNode converts config.PTZNode to ptz.Node, mapping all command strings
// and coordinate limits so PTZ handlers have everything they need.
func buildPTZNode(c config.PTZNode) ptz.Node {
	return ptz.Node{
		MinX: c.MinStepX,
		MaxX: c.MaxStepX,
		MinY: c.MinStepY,
		MaxY: c.MaxStepY,
		MinZ: c.MinStepZ,
		MaxZ: c.MaxStepZ,

		GetPosition:      c.GetPosition,
		IsMoving:         c.IsMoving,
		MoveLeft:         c.MoveLeft,
		MoveRight:        c.MoveRight,
		MoveUp:           c.MoveUp,
		MoveDown:         c.MoveDown,
		MoveIn:           c.MoveIn,
		MoveOut:          c.MoveOut,
		MoveStop:         c.MoveStop,
		MovePreset:       c.MovePreset,
		GotoHomePosition: c.GotoHomePosition,
		SetPreset:        c.SetPreset,
		SetHomePosition:  c.SetHomePosition,
		RemovePreset:     c.RemovePreset,
		JumpToAbs:        c.JumpToAbs,
		JumpToRel:        c.JumpToRel,
		GetPresets:       c.GetPresets,
	}
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
