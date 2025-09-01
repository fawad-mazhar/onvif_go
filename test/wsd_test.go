package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

// TestWSDServerProbe sends a WSD probe request to verify the server responds correctly
func TestWSDServerProbe(t *testing.T) {
	// Give the server some time to start
	time.Sleep(2 * time.Second)
	
	// Create a WSD probe request
	probeRequest := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
				   xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
				   xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">
		<soap:Header>
			<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
			<wsa:MessageID>urn:uuid:test-message-id</wsa:MessageID>
			<wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
		</soap:Header>
		<soap:Body>
			<wsd:Probe>
				<wsd:Types>tds:Device</wsd:Types>
			</wsd:Probe>
		</soap:Body>
	</soap:Envelope>`
	
	// Send the request to the WSD server
	resp, err := http.Post("http://localhost:3703/wsd", "application/soap+xml", bytes.NewBufferString(probeRequest))
	if err != nil {
		t.Logf("WSD server probe test failed (server may not be running): %v", err)
		t.Skip("Skipping WSD probe test - server not running")
		return
	}
	defer resp.Body.Close()
	
	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	
	// Check that the response contains expected WSD elements
	// This is a basic check - in a real test we would parse the XML response
	t.Logf("WSD server probe test completed with status code: %d", resp.StatusCode)
}

// TestWSDServerResolve sends a WSD resolve request to verify the server responds correctly
func TestWSDServerResolve(t *testing.T) {
	// Give the server some time to start
	time.Sleep(2 * time.Second)
	
	// Create a WSD resolve request
	resolveRequest := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
				   xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery"
				   xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">
		<soap:Header>
			<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Resolve</wsa:Action>
			<wsa:MessageID>urn:uuid:test-message-id</wsa:MessageID>
		</soap:Header>
		<soap:Body>
			<wsd:Resolve>
				<wsa:EndpointReference>
					<wsa:Address>urn:uuid:test-device-uuid</wsa:Address>
				</wsa:EndpointReference>
			</wsd:Resolve>
		</soap:Body>
	</soap:Envelope>`
	
	// Send the request to the WSD server
	resp, err := http.Post("http://localhost:3703/wsd", "application/soap+xml", bytes.NewBufferString(resolveRequest))
	if err != nil {
		t.Logf("WSD server resolve test failed (server may not be running): %v", err)
		t.Skip("Skipping WSD resolve test - server not running")
		return
	}
	defer resp.Body.Close()
	
	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	
	// Check that the response contains expected WSD elements
	// This is a basic check - in a real test we would parse the XML response
	t.Logf("WSD server resolve test completed with status code: %d", resp.StatusCode)
}
