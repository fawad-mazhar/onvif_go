package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

func StartWSDServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)
	// logger.SetLevel(logger.INFO) - level already set during initialization
	
	// Configuration is passed as parameter
	// Get the IP address of the first network interface
	ip := getLocalIP()
	if ip == "" {
		err := fmt.Errorf("failed to get local IP address")
		logger.Error("Failed to get local IP address")
		return err
	}
	logger.Info("WSD server detected local IP: %s", ip)
	
		
	// Create HTTP server for WSD
	http.HandleFunc("/wsd", func(w http.ResponseWriter, r *http.Request) {
		// Handle WSD requests
		handleWSDRequest(w, r, cfg, ip)
	})
	
	// Start the HTTP server
	addr := fmt.Sprintf(":%d", cfg.WSDPort)
	logger.Info("Starting WSD server on port %d", cfg.WSDPort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Fatal("Failed to start WSD server: %v", err)
	}
	return nil
}

// getLocalIP returns the IP address of the first network interface
func getLocalIP() string {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Warn("Failed to get network interfaces: %v", err)
		return ""
	}
	
	// Iterate through interfaces to find the first one with an IP address
	for _, iface := range interfaces {
		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		
		// Get addresses for the interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		// Find the first IPv4 address
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	
	return ""
}

// handleWSDRequest handles WSD requests and sends appropriate responses
func handleWSDRequest(w http.ResponseWriter, r *http.Request, cfg *config.ServiceContext, ip string) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Failed to read WSD request: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	request := string(body)
	logger.Debug("WSD request: %s", request)
	
	// Set response headers
	w.Header().Set("Content-Type", "application/soap+xml")
	
	// Handle different WSD actions
	switch {
	case strings.Contains(request, "Probe"):
		// Handle probe request
		response := generateProbeMatchesResponse(cfg, ip)
		fmt.Fprint(w, response)
	case strings.Contains(request, "Resolve"):
		// Handle resolve request
		response := generateResolveResponse(cfg, ip)
		fmt.Fprint(w, response)
	default:
		// Send a generic response
		response := utils.GenerateGenericResponse()
		fmt.Fprint(w, response)
	}
}

// generateProbeMatchesResponse generates a WSD probe matches response
func generateProbeMatchesResponse(cfg *config.ServiceContext, ip string) string {
	// Read the ProbeMatches template
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope
	xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
	xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
	xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
	xmlns:tt="http://www.onvif.org/ver10/schema">
	<soap:Header>
		<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/ProbeMatches</wsa:Action>
		<wsa:MessageID>%MESSAGE_ID%</wsa:MessageID>
		<wsa:RelatesTo>%RELATES_TO%</wsa:RelatesTo>
		<wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
	</soap:Header>
	<soap:Body>
		<wsd:ProbeMatches>
			<wsd:ProbeMatch>
				<wsa:EndpointReference>
					<wsa:Address>urn:uuid:%UUID%</wsa:Address>
				</wsa:EndpointReference>
				<wsd:Types>tds:Device</wsd:Types>
				<wsd:MetadataVersion>1</wsd:MetadataVersion>
			</wsd:ProbeMatch>
		</wsd:ProbeMatches>
	</soap:Body>
</soap:Envelope>`
	
	// Replace placeholders with actual values
	response := strings.Replace(template, "%MESSAGE_ID%", "urn:uuid:"+generateUUID(), -1)
	response = strings.Replace(response, "%RELATES_TO%", "urn:uuid:example", -1)
	response = strings.Replace(response, "%UUID%", cfg.UUID, -1)
	
	return response
}

// generateResolveResponse generates a WSD resolve response
func generateResolveResponse(cfg *config.ServiceContext, ip string) string {
	// Read the Resolve template
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope
	xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
	xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
	xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery">
	<soap:Header>
		<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/ResolveMatches</wsa:Action>
		<wsa:MessageID>%MESSAGE_ID%</wsa:MessageID>
	</soap:Header>
	<soap:Body>
		<wsd:ResolveMatches>
			<wsd:ResolveMatch>
				<wsa:EndpointReference>
					<wsa:Address>urn:uuid:%UUID%</wsa:Address>
				</wsa:EndpointReference>
				<wsd:Types>tds:Device</wsd:Types>
				<wsd:MetadataVersion>1</wsd:MetadataVersion>
			</wsd:ResolveMatch>
		</wsd:ResolveMatches>
	</soap:Body>
</soap:Envelope>`
	
	// Replace placeholders with actual values
	response := strings.Replace(template, "%MESSAGE_ID%", "urn:uuid:"+generateUUID(), -1)
	response = strings.Replace(response, "%UUID%", cfg.UUID, -1)
	
	return response
}


// generateUUID generates a simple UUID for WSD responses
func generateUUID() string {
	// For simplicity, we'll generate a fixed UUID
	// In a real implementation, this should be a proper UUID generator
	return "12345678-1234-1234-1234-123456789012"
}
