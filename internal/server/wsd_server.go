package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

const (
	// MulticastAddress is the ONVIF WS-Discovery multicast address
	MulticastAddress = "239.255.255.250"
	// MulticastPort is the ONVIF WS-Discovery multicast port
	MulticastPort = 3702

	// DeviceType is the ONVIF device type
	DeviceType = "tdn:NetworkVideoTransmitter"

	// ActionHello is the WS-Discovery Hello action
	ActionHello        = "http://schemas.xmlsoap.org/ws/2005/04/discovery/Hello"
	// ActionBye is the WS-Discovery Bye action
	ActionBye          = "http://schemas.xmlsoap.org/ws/2005/04/discovery/Bye"
	// ActionProbe is the WS-Discovery Probe action
	ActionProbe        = "http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe"
	// ActionProbeMatch is the WS-Discovery ProbeMatches action
	ActionProbeMatch   = "http://schemas.xmlsoap.org/ws/2005/04/discovery/ProbeMatches"
	// ActionResolve is the WS-Discovery Resolve action
	ActionResolve      = "http://schemas.xmlsoap.org/ws/2005/04/discovery/Resolve"
	// ActionResolveMatch is the WS-Discovery ResolveMatches action
	ActionResolveMatch = "http://schemas.xmlsoap.org/ws/2005/04/discovery/ResolveMatches"

	// Network buffer and UUID generation constants
	// MessageBufferSize defines the buffer size for UDP messages
	MessageBufferSize = 4096
	// UUIDByteLength defines the length of UUID byte array
	UUIDByteLength = 16
	// UUID bit masks and shifts for timestamp-based generation
	UUIDMask16Bits    = 0xFFFF
	UUIDMask48Bits    = 0xFFFFFFFFFFFF
	UUIDShift16Bits   = 16
	UUIDShift32Bits   = 32
	UUIDShift48Bits   = 48
	// UUID version and variant bits
	UUIDVersionMask   = 0x0F
	UUIDVersion4      = 0x40
	UUIDVariantMask   = 0x3F
	UUIDVariant10     = 0x80
)

type WSDServer struct {
	config        *config.ServiceContext
	conn          *net.UDPConn
	multicastAddr *net.UDPAddr
	localIP       string
	deviceUUID    string
	msgNumber     int64
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// StartWSDServer starts the WS-Discovery UDP multicast server
func StartWSDServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)

	// Create WS-Discovery server instance
	server, err := NewWSDServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create WS-Discovery server: %v", err)
	}

	// Start the server
	return server.Start()
}

// NewWSDServer creates a new WS-Discovery server instance
func NewWSDServer(cfg *config.ServiceContext) (*WSDServer, error) {
	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP: %v", err)
	}

	// Generate device UUID if not provided
	deviceUUID := cfg.UUID
	if deviceUUID == "" {
		deviceUUID = generateUUID()
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	server := &WSDServer{
		config:     cfg,
		localIP:    localIP,
		deviceUUID: deviceUUID,
		msgNumber:  0,
		ctx:        ctx,
		cancel:     cancel,
	}

	logger.Infof("WS-Discovery server initialized - IP: %s, UUID: %s", localIP, deviceUUID)
	return server, nil
}

// Start starts the WS-Discovery server
func (s *WSDServer) Start() error {
	// Set up UDP multicast socket
	if err := s.setupMulticastSocket(); err != nil {
		return fmt.Errorf("failed to setup multicast socket: %v", err)
	}
	defer func() {
		if err := s.conn.Close(); err != nil {
			logger.Errorf("Error closing connection: %v", err)
		}
	}()

	// Set up signal handler for graceful shutdown
	s.setupSignalHandler()

	// Send Hello announcement
	if err := s.sendHello(); err != nil {
		logger.Warnf("Failed to send Hello message: %v", err)
	}

	// Start listening for discovery messages
	s.wg.Add(1)
	go s.listenForMessages()

	logger.Infof("WS-Discovery server started on %s:%d", MulticastAddress, MulticastPort)

	// Wait for shutdown
	s.wg.Wait()

	return nil
}

// setupMulticastSocket sets up UDP multicast socket for WS-Discovery
func (s *WSDServer) setupMulticastSocket() error {
	// Parse multicast address
	multicastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MulticastAddress, MulticastPort))
	if err != nil {
		return fmt.Errorf("failed to resolve multicast address: %v", err)
	}
	s.multicastAddr = multicastAddr

	// Try different approaches for multicast setup
	conn, err := s.setupMulticastWithFallback()
	if err != nil {
		return fmt.Errorf("failed to setup multicast socket: %v", err)
	}
	s.conn = conn

	logger.Infof("WS-Discovery multicast socket setup successful on %s:%d", MulticastAddress, MulticastPort)
	return nil
}

// setupMulticastWithFallback tries multiple approaches to setup multicast
func (s *WSDServer) setupMulticastWithFallback() (*net.UDPConn, error) {
	var lastErr error

	// Approach 1: Try with specific interface and multicast join
	conn, err := s.tryMulticastWithInterface()
	if err == nil {
		return conn, nil
	}
	lastErr = err
	logger.Warnf("Multicast with interface failed: %v, trying fallback approach", err)

	// Approach 2: Try simple UDP listen without explicit multicast join
	conn, err = s.trySimpleUDPListen()
	if err == nil {
		logger.Warnf("Using fallback UDP listen mode - multicast may be limited")
		return conn, nil
	}
	logger.Warnf("Simple UDP listen failed: %v", err)

	// Approach 3: Try with any available port (for testing)
	conn, err = s.tryUDPListenAnyPort()
	if err == nil {
		logger.Warnf("Using UDP listen on alternative port - WS-Discovery may not work with standard clients")
		return conn, nil
	}

	return nil, fmt.Errorf("all multicast setup approaches failed, last error: %v", lastErr)
}

// tryMulticastWithInterface attempts full multicast setup with interface
func (s *WSDServer) tryMulticastWithInterface() (*net.UDPConn, error) {
	// Create UDP connection
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: MulticastPort,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %v", err)
	}

	// Get the network interface
	iface, err := s.getNetworkInterface()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Errorf("Error closing connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to get network interface: %v", err)
	}

	// Create packet connection for multicast operations
	packetConn := ipv4.NewPacketConn(conn)

	// Join multicast group
	if err := packetConn.JoinGroup(iface, s.multicastAddr); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Errorf("Error closing connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to join multicast group: %v", err)
	}

	// Set multicast interface
	if err := packetConn.SetMulticastInterface(iface); err != nil {
		logger.Warnf("Failed to set multicast interface: %v", err)
	}

	// Set multicast loop (don't receive own messages)
	if err := packetConn.SetMulticastLoopback(false); err != nil {
		logger.Warnf("Failed to set multicast loopback: %v", err)
	}

	logger.Infof("Joined multicast group %s on interface %s", MulticastAddress, iface.Name)
	return conn, nil
}

// trySimpleUDPListen attempts simple UDP listen without multicast join
func (s *WSDServer) trySimpleUDPListen() (*net.UDPConn, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: MulticastPort,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create simple UDP socket: %v", err)
	}

	logger.Infof("Created UDP socket on port %d (without multicast join)", MulticastPort)
	return conn, nil
}

// tryUDPListenAnyPort attempts UDP listen on any available port
func (s *WSDServer) tryUDPListenAnyPort() (*net.UDPConn, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0, // Let system choose port
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket on any port: %v", err)
	}

	addr := conn.LocalAddr().(*net.UDPAddr)
	logger.Infof("Created UDP socket on port %d (fallback mode)", addr.Port)
	return conn, nil
}

// getNetworkInterface returns the network interface to use for multicast
func (s *WSDServer) getNetworkInterface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Find interface with our local IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.String() == s.localIP {
					return &iface, nil
				}
			}
		}
	}

	// Fallback to first non-loopback interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("no suitable network interface found")
}

// setupSignalHandler sets up graceful shutdown on SIGTERM/SIGINT
func (s *WSDServer) setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		logger.Infof("Shutdown signal received, sending Bye message...")

		// Send Bye message
		if err := s.sendBye(); err != nil {
			logger.Warnf("Failed to send Bye message: %v", err)
		}

		// Cancel context to stop other goroutines
		s.cancel()
	}()
}

// listenForMessages listens for incoming WS-Discovery messages
func (s *WSDServer) listenForMessages() {
	defer s.wg.Done()

	buffer := make([]byte, MessageBufferSize)

	for {
		select {
		case <-s.ctx.Done():
			logger.Infof("WS-Discovery server shutting down...")
			return

		default:
			// Set read timeout
			if err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
				logger.Errorf("Error setting read deadline: %v", err)
				continue
			}

			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is expected, continue listening
				}
				logger.Warnf("Error reading UDP message: %v", err)
				continue
			}

			message := string(buffer[:n])
			logger.Debugf("Received message from %s: %s", addr.String(), message)

			// Process the message
			s.processMessage(message, addr)
		}
	}
}

// processMessage processes incoming WS-Discovery messages
func (s *WSDServer) processMessage(message string, from *net.UDPAddr) {
	// Parse SOAP action from message
	action := s.parseSOAPAction(message)
	if action == "" {
		logger.Debugf("Could not parse SOAP action from message")
		return
	}

	logger.Debugf("Processing WS-Discovery action: %s", action)

	switch action {
	case ActionProbe:
		s.handleProbe(message, from)
	case ActionResolve:
		s.handleResolve(message, from)
	default:
		logger.Debugf("Ignoring unknown WS-Discovery action: %s", action)
	}
}

// parseSOAPAction extracts the SOAP action from the message
func (s *WSDServer) parseSOAPAction(message string) string {
	// Look for wsa:Action in the SOAP header
	actionStart := strings.Index(message, "<wsa:Action>")
	if actionStart == -1 {
		actionStart = strings.Index(message, "<a:Action>")
	}
	if actionStart == -1 {
		return ""
	}

	actionStart += len("<wsa:Action>")
	actionEnd := strings.Index(message[actionStart:], "</")
	if actionEnd == -1 {
		return ""
	}

	return strings.TrimSpace(message[actionStart : actionStart+actionEnd])
}

// parseMessageID extracts the MessageID from incoming message for RelatesTo
func (s *WSDServer) parseMessageID(message string) string {
	// Look for wsa:MessageID
	idStart := strings.Index(message, "<wsa:MessageID>")
	if idStart == -1 {
		idStart = strings.Index(message, "<a:MessageID>")
	}
	if idStart == -1 {
		return ""
	}

	idStart += len("<wsa:MessageID>")
	idEnd := strings.Index(message[idStart:], "</")
	if idEnd == -1 {
		return ""
	}

	return strings.TrimSpace(message[idStart : idStart+idEnd])
}

// getLocalIP returns the local IP address
func getLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no local IP address found")
}

// generateUUID generates a simple UUID
func generateUUID() string {
	b := make([]byte, UUIDByteLength)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based UUID
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			time.Now().Unix(),
			time.Now().UnixNano()&UUIDMask16Bits,
			time.Now().UnixNano()>>UUIDShift16Bits&UUIDMask16Bits,
			time.Now().UnixNano()>>UUIDShift32Bits&UUIDMask16Bits,
			time.Now().UnixNano()>>UUIDShift48Bits&UUIDMask48Bits)
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & UUIDVersionMask) | UUIDVersion4 // Version 4
	b[8] = (b[8] & UUIDVariantMask) | UUIDVariant10 // Variant 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// getNextMessageNumber returns the next message sequence number
func (s *WSDServer) getNextMessageNumber() int64 {
	return atomic.AddInt64(&s.msgNumber, 1)
}

// handleProbe processes Probe messages and sends ProbeMatch responses
func (s *WSDServer) handleProbe(message string, from *net.UDPAddr) {
	s.handleRequestMessage(message, from, "Probe", s.generateProbeMatchResponse)
}

// handleResolve processes Resolve messages and sends ResolveMatch responses
func (s *WSDServer) handleResolve(message string, from *net.UDPAddr) {
	s.handleRequestMessage(message, from, "Resolve", s.generateResolveMatchResponse)
}

// handleRequestMessage is a common handler for Probe and Resolve requests
func (s *WSDServer) handleRequestMessage(message string, from *net.UDPAddr, requestType string, generateResponse func(string) (string, error)) {
	logger.Debugf("Handling %s request from %s", requestType, from.String())

	// Parse MessageID from incoming message for RelatesTo
	messageID := s.parseMessageID(message)
	if messageID == "" {
		logger.Warnf("Could not parse MessageID from %s message", requestType)
		return
	}

	// Generate response
	response, err := generateResponse(messageID)
	if err != nil {
		logger.Errorf("Failed to generate %sMatch response: %v", requestType, err)
		return
	}

	// Send unicast response to the requester
	if err := s.sendUnicastMessage(response, from); err != nil {
		logger.Errorf("Failed to send %sMatch response: %v", requestType, err)
		return
	}

	logger.Debugf("Sent %sMatch response to %s", requestType, from.String())
}

// generateProbeMatchResponse generates a SOAP ProbeMatch response
func (s *WSDServer) generateProbeMatchResponse(relatesTo string) (string, error) {
	return s.generateMatchResponse(relatesTo, "service_files/wsd/ProbeMatches.xml", "ProbeMatch", s.generateHardcodedProbeMatch)
}

// generateResolveMatchResponse generates a SOAP ResolveMatch response
func (s *WSDServer) generateResolveMatchResponse(relatesTo string) (string, error) {
	return s.generateMatchResponse(relatesTo, "service_files/wsd/ResolveMatches.xml", "ResolveMatch", s.generateHardcodedResolveMatch)
}

// generateMatchResponse is a common function for generating match responses
func (s *WSDServer) generateMatchResponse(relatesTo, templatePath, responseType string, hardcodedGenerator func(map[string]string) string) (string, error) {
	// Build service URL (XAddr)
	serviceURL := fmt.Sprintf("http://%s:%d/onvif/device_service", s.localIP, s.config.Port)

	// Build scopes
	scopes := s.buildDeviceScopes()

	// Create replacements map
	replacements := map[string]string{
		"%MSG_UUID%":    generateUUID(),
		"%RELATES_TO%":  relatesTo,
		"%UUID%":        s.deviceUUID,
		"%DEVICE_TYPE%": DeviceType,
		"%SCOPES%":      scopes,
		"%XADDRS%":      serviceURL,
		"%MSG_NUMBER%":  strconv.FormatInt(s.getNextMessageNumber(), 10),
	}

	// Use template from service_files
	if !xml.FileExists(templatePath) {
		// Fallback to hardcoded template
		return hardcodedGenerator(replacements), nil
	}

	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return "", fmt.Errorf("failed to process %s template: %v", responseType, err)
	}

	return response, nil
}

// buildDeviceScopes builds ONVIF device scopes string
func (s *WSDServer) buildDeviceScopes() string {
	scopes := []string{
		"onvif://www.onvif.org/type/video_encoder",
		"onvif://www.onvif.org/type/audio_encoder",
	}

	// Add hardware scope if available
	if s.config.Manufacturer != "" {
		scopes = append(scopes, fmt.Sprintf("onvif://www.onvif.org/hardware/%s", s.config.Manufacturer))
	}

	// Add model scope if available
	if s.config.Model != "" {
		scopes = append(scopes, fmt.Sprintf("onvif://www.onvif.org/name/%s", s.config.Model))
	}

	return strings.Join(scopes, " ")
}

// generateHardcodedProbeMatch generates ProbeMatch response without template file
func (s *WSDServer) generateHardcodedProbeMatch(replacements map[string]string) string {
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
               xmlns:tdn="http://www.onvif.org/ver10/network/wsdl">
    <soap:Header>
        <wsa:Action>%ACTION_PROBE_MATCH%</wsa:Action>
        <wsa:MessageID>uuid:%MSG_UUID%</wsa:MessageID>
        <wsa:RelatesTo>%RELATES_TO%</wsa:RelatesTo>
        <wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
        <wsd:AppSequence InstanceId="0" MessageNumber="%MSG_NUMBER%"/>
    </soap:Header>
    <soap:Body>
        <wsd:ProbeMatches>
            <wsd:ProbeMatch>
                <wsa:EndpointReference>
                    <wsa:Address>urn:uuid:%UUID%</wsa:Address>
                </wsa:EndpointReference>
                <wsd:Types>%DEVICE_TYPE%</wsd:Types>
                <wsd:Scopes>%SCOPES%</wsd:Scopes>
                <wsd:XAddrs>%XADDRS%</wsd:XAddrs>
                <wsd:MetadataVersion>1</wsd:MetadataVersion>
            </wsd:ProbeMatch>
        </wsd:ProbeMatches>
    </soap:Body>
</soap:Envelope>`

	// Add action to replacements
	replacements["%ACTION_PROBE_MATCH%"] = ActionProbeMatch

	// Replace placeholders
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// generateHardcodedResolveMatch generates ResolveMatch response without template file
func (s *WSDServer) generateHardcodedResolveMatch(replacements map[string]string) string {
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
               xmlns:tdn="http://www.onvif.org/ver10/network/wsdl">
    <soap:Header>
        <wsa:Action>%ACTION_RESOLVE_MATCH%</wsa:Action>
        <wsa:MessageID>uuid:%MSG_UUID%</wsa:MessageID>
        <wsa:RelatesTo>%RELATES_TO%</wsa:RelatesTo>
        <wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
        <wsd:AppSequence InstanceId="0" MessageNumber="%MSG_NUMBER%"/>
    </soap:Header>
    <soap:Body>
        <wsd:ResolveMatches>
            <wsd:ResolveMatch>
                <wsa:EndpointReference>
                    <wsa:Address>urn:uuid:%UUID%</wsa:Address>
                </wsa:EndpointReference>
                <wsd:Types>%DEVICE_TYPE%</wsd:Types>
                <wsd:Scopes>%SCOPES%</wsd:Scopes>
                <wsd:XAddrs>%XADDRS%</wsd:XAddrs>
                <wsd:MetadataVersion>1</wsd:MetadataVersion>
            </wsd:ResolveMatch>
        </wsd:ResolveMatches>
    </soap:Body>
</soap:Envelope>`

	// Add action to replacements
	replacements["%ACTION_RESOLVE_MATCH%"] = ActionResolveMatch

	// Replace placeholders
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// sendUnicastMessage sends a message to a specific UDP address
func (s *WSDServer) sendUnicastMessage(message string, to *net.UDPAddr) error {
	_, err := s.conn.WriteToUDP([]byte(message), to)
	return err
}

// sendMulticastMessage sends a message to the multicast group
func (s *WSDServer) sendMulticastMessage(message string) error {
	_, err := s.conn.WriteToUDP([]byte(message), s.multicastAddr)
	return err
}

// sendHello sends a Hello announcement to the multicast group
func (s *WSDServer) sendHello() error {
	logger.Infof("Sending Hello announcement...")

	// Generate Hello message
	message, err := s.generateHelloMessage()
	if err != nil {
		return fmt.Errorf("failed to generate Hello message: %v", err)
	}

	// Send to multicast group
	if err := s.sendMulticastMessage(message); err != nil {
		return fmt.Errorf("failed to send Hello message: %v", err)
	}

	logger.Infof("Hello announcement sent successfully")
	return nil
}

// sendBye sends a Bye announcement to the multicast group
func (s *WSDServer) sendBye() error {
	logger.Infof("Sending Bye announcement...")

	// Generate Bye message
	message, err := s.generateByeMessage()
	if err != nil {
		return fmt.Errorf("failed to generate Bye message: %v", err)
	}

	// Send to multicast group
	if err := s.sendMulticastMessage(message); err != nil {
		return fmt.Errorf("failed to send Bye message: %v", err)
	}

	logger.Infof("Bye announcement sent successfully")
	return nil
}

// generateHelloMessage generates a SOAP Hello announcement
func (s *WSDServer) generateHelloMessage() (string, error) {
	return s.generateAnnouncementMessage("service_files/wsd/Hello.xml", "Hello", s.generateHardcodedHello)
}

// generateByeMessage generates a SOAP Bye announcement
func (s *WSDServer) generateByeMessage() (string, error) {
	return s.generateAnnouncementMessage("service_files/wsd/Bye.xml", "Bye", s.generateHardcodedBye)
}

// generateAnnouncementMessage is a common function for generating Hello/Bye messages
func (s *WSDServer) generateAnnouncementMessage(templatePath, messageType string, hardcodedGenerator func(map[string]string) string) (string, error) {
	// Build service URL (XAddr)
	serviceURL := fmt.Sprintf("http://%s:%d/onvif/device_service", s.localIP, s.config.Port)

	// Build scopes
	scopes := s.buildDeviceScopes()

	// Create replacements map
	replacements := map[string]string{
		"%MSG_UUID%":    generateUUID(),
		"%UUID%":        s.deviceUUID,
		"%DEVICE_TYPE%": DeviceType,
		"%SCOPES%":      scopes,
		"%XADDRS%":      serviceURL,
		"%MSG_NUMBER%":  strconv.FormatInt(s.getNextMessageNumber(), 10),
	}

	// Use template from service_files
	if !xml.FileExists(templatePath) {
		// Fallback to hardcoded template
		return hardcodedGenerator(replacements), nil
	}

	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return "", fmt.Errorf("failed to process %s template: %v", messageType, err)
	}

	return response, nil
}

// generateHardcodedHello generates Hello announcement without template file
func (s *WSDServer) generateHardcodedHello(replacements map[string]string) string {
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
               xmlns:tdn="http://www.onvif.org/ver10/network/wsdl">
    <soap:Header>
        <wsa:Action>%ACTION_HELLO%</wsa:Action>
        <wsa:MessageID>uuid:%MSG_UUID%</wsa:MessageID>
        <wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
        <wsd:AppSequence InstanceId="0" MessageNumber="%MSG_NUMBER%"/>
    </soap:Header>
    <soap:Body>
        <wsd:Hello>
            <wsa:EndpointReference>
                <wsa:Address>urn:uuid:%UUID%</wsa:Address>
            </wsa:EndpointReference>
            <wsd:Types>%DEVICE_TYPE%</wsd:Types>
            <wsd:Scopes>%SCOPES%</wsd:Scopes>
            <wsd:XAddrs>%XADDRS%</wsd:XAddrs>
            <wsd:MetadataVersion>1</wsd:MetadataVersion>
        </wsd:Hello>
    </soap:Body>
</soap:Envelope>`

	// Add action to replacements
	replacements["%ACTION_HELLO%"] = ActionHello

	// Replace placeholders
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// generateHardcodedBye generates Bye announcement without template file
func (s *WSDServer) generateHardcodedBye(replacements map[string]string) string {
	template := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
               xmlns:tdn="http://www.onvif.org/ver10/network/wsdl">
    <soap:Header>
        <wsa:Action>%ACTION_BYE%</wsa:Action>
        <wsa:MessageID>uuid:%MSG_UUID%</wsa:MessageID>
        <wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
        <wsd:AppSequence InstanceId="0" MessageNumber="%MSG_NUMBER%"/>
    </soap:Header>
    <soap:Body>
        <wsd:Bye>
            <wsa:EndpointReference>
                <wsa:Address>urn:uuid:%UUID%</wsa:Address>
            </wsa:EndpointReference>
            <wsd:Types>%DEVICE_TYPE%</wsd:Types>
            <wsd:Scopes>%SCOPES%</wsd:Scopes>
            <wsd:XAddrs>%XADDRS%</wsd:XAddrs>
            <wsd:MetadataVersion>1</wsd:MetadataVersion>
        </wsd:Bye>
    </soap:Body>
</soap:Envelope>`

	// Add action to replacements
	replacements["%ACTION_BYE%"] = ActionBye

	// Replace placeholders
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
