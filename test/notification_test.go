package tests

import (
	"testing"
)

// TestNotificationServerSubscribe tests the notification server's subscription functionality
func TestNotificationServerSubscribe(t *testing.T) {
	runSOAPTest(t, TestCase{
		Name: "Notification server subscribe",
		URL:  "http://localhost:8082/notification",
		Request: `<?xml version="1.0" encoding="UTF-8"?>
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
	</SOAP-ENV:Envelope>`,
	})
}

// TestNotificationServerPullMessages tests the notification server's PullMessages functionality
func TestNotificationServerPullMessages(t *testing.T) {
	runSOAPTest(t, TestCase{
		Name: "Notification server PullMessages",
		URL:  "http://localhost:8082/notification",
		Request: `<?xml version="1.0" encoding="UTF-8"?>
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
	</SOAP-ENV:Envelope>`,
	})
}

// TestNotificationServerGetEventProperties tests the notification server's GetEventProperties functionality
func TestNotificationServerGetEventProperties(t *testing.T) {
	runSOAPTest(t, TestCase{
		Name: "Notification server GetEventProperties",
		URL:  "http://localhost:8082/notification",
		Request: `<?xml version="1.0" encoding="UTF-8"?>
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
	</SOAP-ENV:Envelope>`,
	})
}
