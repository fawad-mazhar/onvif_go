package server

import (
	"testing"
)

// --- parseSOAPAction ---

const wsdProbeWSA = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery">
  <soap:Header>
    <wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
    <wsa:MessageID>uuid:abc-123</wsa:MessageID>
    <wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
  </soap:Header>
  <soap:Body><wsd:Probe/></soap:Body>
</soap:Envelope>`

// Same message but using the short "a:" prefix — this is the case that
// the old off-by-N parser silently mangled (extracted "tp://..." instead
// of "http://...").
const wsdProbeA = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery">
  <soap:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
    <a:MessageID>uuid:def-456</a:MessageID>
    <a:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
  </soap:Header>
  <soap:Body><wsd:Probe/></soap:Body>
</soap:Envelope>`

func newTestWSDServer() *WSDServer {
	return &WSDServer{}
}

func TestParseSOAPAction_WSAPrefix(t *testing.T) {
	s := newTestWSDServer()
	got := s.parseSOAPAction(wsdProbeWSA)
	want := ActionProbe
	if got != want {
		t.Errorf("parseSOAPAction(wsa:) = %q; want %q", got, want)
	}
}

func TestParseSOAPAction_APrefix(t *testing.T) {
	s := newTestWSDServer()
	got := s.parseSOAPAction(wsdProbeA)
	want := ActionProbe
	if got != want {
		// Specifically call out the off-by-N symptom the old parser produced.
		t.Errorf("parseSOAPAction(a:) = %q; want %q\n(old parser would return %q)",
			got, want, "tp://schemas.xmlsoap.org/ws/2005/04/discovery/Probe")
	}
}

func TestParseSOAPAction_Empty(t *testing.T) {
	s := newTestWSDServer()
	if got := s.parseSOAPAction("<soap:Envelope/>"); got != "" {
		t.Errorf("parseSOAPAction(no action) = %q; want empty", got)
	}
}

// --- parseMessageID ---

func TestParseMessageID_WSAPrefix(t *testing.T) {
	s := newTestWSDServer()
	got := s.parseMessageID(wsdProbeWSA)
	if got != "uuid:abc-123" {
		t.Errorf("parseMessageID(wsa:) = %q; want %q", got, "uuid:abc-123")
	}
}

func TestParseMessageID_APrefix(t *testing.T) {
	s := newTestWSDServer()
	got := s.parseMessageID(wsdProbeA)
	// Old parser: len("<wsa:MessageID>")=15 vs len("<a:MessageID>")=13 → skipped "uu"
	if got != "uuid:def-456" {
		t.Errorf("parseMessageID(a:) = %q; want %q\n(old parser would return %q)",
			got, "uuid:def-456", "id:def-456")
	}
}

func TestParseMessageID_Empty(t *testing.T) {
	s := newTestWSDServer()
	if got := s.parseMessageID("<soap:Envelope/>"); got != "" {
		t.Errorf("parseMessageID(no id) = %q; want empty", got)
	}
}
