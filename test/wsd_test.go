package tests

import (
	"testing"
)

// TestWSDServerProbe sends a WSD probe request to verify the server responds correctly
func TestWSDServerProbe(t *testing.T) {
	runSOAPTest(t, TestCase{
		Name: "WSD server probe",
		URL:  "http://localhost:3703/wsd",
		Request: `<?xml version="1.0" encoding="utf-8"?>
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
	</soap:Envelope>`,
	})
}

// TestWSDServerResolve sends a WSD resolve request to verify the server responds correctly
func TestWSDServerResolve(t *testing.T) {
	runSOAPTest(t, TestCase{
		Name: "WSD server resolve",
		URL:  "http://localhost:3703/wsd",
		Request: `<?xml version="1.0" encoding="utf-8"?>
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
	</soap:Envelope>`,
	})
}
