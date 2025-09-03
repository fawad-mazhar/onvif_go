package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

// TestNotificationServerSubscribe tests the notification server's subscription functionality
func TestNotificationServerSubscribe(t *testing.T) {
	// Give the server some time to start
	time.Sleep(2 * time.Second)

	// Create a notification subscription request
	subscribeRequest := `<?xml version="1.0" encoding="UTF-8"?>
	<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
					   xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
					   xmlns:wse="http://schemas.xmlsoap.org/ws/2004/08/eventing">
		<SOAP-ENV:Header>
			<wsa:Action>http://schemas.xmlsoap.org/ws/2004/08/eventing/Subscribe</wsa:Action>
			<wsa:MessageID>urn:uuid:test-message-id</wsa:MessageID>
			<wsa:To>http://localhost:8082/notification</wsa:To>
		</SOAP-ENV:Header>
		<SOAP-ENV:Body>
			<wse:Subscribe>
				<wse:ConsumerReference>
					<wsa:Address>http://localhost:8082/notification_consumer</wsa:Address>
				</wse:ConsumerReference>
				<wse:InitialTerminationTime>PT60S</wse:InitialTerminationTime>
			</wse:Subscribe>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`

	// Send the request to the notification server
	resp, err := http.Post("http://localhost:8082/notification", "application/soap+xml", bytes.NewBufferString(subscribeRequest))
	if err != nil {
		t.Logf("Notification server subscribe test failed (server may not be running): %v", err)
		t.Skip("Skipping notification subscribe test - server not running")
		return
	}
	defer resp.Body.Close()

	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check that the response contains expected notification elements
	// This is a basic check - in a real test we would parse the XML response
	t.Logf("Notification server subscribe test completed with status code: %d", resp.StatusCode)
}

// TestNotificationServerPullMessages tests the notification server's PullMessages functionality
func TestNotificationServerPullMessages(t *testing.T) {
	// Give the server some time to start
	time.Sleep(2 * time.Second)

	// Create a PullMessages request
	pullMessagesRequest := `<?xml version="1.0" encoding="UTF-8"?>
	<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
					   xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
					   xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
		<SOAP-ENV:Header>
			<wsa:Action>http://www.onvif.org/ver10/events/wsdl/PullPointSubscription/PullMessagesRequest</wsa:Action>
			<wsa:MessageID>urn:uuid:test-message-id</wsa:MessageID>
		</SOAP-ENV:Header>
		<SOAP-ENV:Body>
			<tev:PullMessages>
				<tev:Timeout>PT5S</tev:Timeout>
				<tev:MessageLimit>10</tev:MessageLimit>
			</tev:PullMessages>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`

	// Send the request to the notification server
	resp, err := http.Post("http://localhost:8082/notification", "application/soap+xml", bytes.NewBufferString(pullMessagesRequest))
	if err != nil {
		t.Logf("Notification server PullMessages test failed (server may not be running): %v", err)
		t.Skip("Skipping notification PullMessages test - server not running")
		return
	}
	defer resp.Body.Close()

	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check that the response contains expected notification elements
	// This is a basic check - in a real test we would parse the XML response
	t.Logf("Notification server PullMessages test completed with status code: %d", resp.StatusCode)
}

// TestNotificationServerGetEventProperties tests the notification server's GetEventProperties functionality
func TestNotificationServerGetEventProperties(t *testing.T) {
	// Give the server some time to start
	time.Sleep(2 * time.Second)

	// Create a GetEventProperties request
	getEventPropertiesRequest := `<?xml version="1.0" encoding="UTF-8"?>
	<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
					   xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
					   xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
		<SOAP-ENV:Header>
			<wsa:Action>http://www.onvif.org/ver10/events/wsdl/GetEventProperties</wsa:Action>
			<wsa:MessageID>urn:uuid:test-message-id</wsa:MessageID>
		</SOAP-ENV:Header>
		<SOAP-ENV:Body>
			<tev:GetEventProperties/>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`

	// Send the request to the notification server
	resp, err := http.Post("http://localhost:8082/notification", "application/soap+xml", bytes.NewBufferString(getEventPropertiesRequest))
	if err != nil {
		t.Logf("Notification server GetEventProperties test failed (server may not be running): %v", err)
		t.Skip("Skipping notification GetEventProperties test - server not running")
		return
	}
	defer resp.Body.Close()

	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check that the response contains expected notification elements
	// This is a basic check - in a real test we would parse the XML response
	t.Logf("Notification server GetEventProperties test completed with status code: %d", resp.StatusCode)
}
