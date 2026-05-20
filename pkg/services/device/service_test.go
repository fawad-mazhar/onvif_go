package device

import (
	"testing"
)

// --- servicesTemplate variant selection ---

func TestServicesTemplate(t *testing.T) {
	cases := []struct {
		ptz, media2, withCap bool
		want                 string
	}{
		{false, false, false, "GetServices_no_ptz_no_media2"},
		{false, true, false, "GetServices_no_ptz_media2"},
		{true, false, false, "GetServices_ptz_no_media2"},
		{true, true, false, "GetServices_ptz_media2"},
		{false, false, true, "GetServices_with_capabilities_no_ptz_no_media2"},
		{false, true, true, "GetServices_with_capabilities_no_ptz_media2"},
		{true, false, true, "GetServices_with_capabilities_ptz_no_media2"},
		{true, true, true, "GetServices_with_capabilities_ptz_media2"},
	}
	for _, c := range cases {
		got := servicesTemplate(c.ptz, c.media2, c.withCap)
		if got != c.want {
			t.Errorf("servicesTemplate(%v,%v,%v) = %q; want %q", c.ptz, c.media2, c.withCap, got, c.want)
		}
	}
}

// --- eventsFlags ---

func TestEventsFlags(t *testing.T) {
	cases := []struct {
		ev          EventsEnable
		pull, bsub  string
	}{
		{EventsNone, "false", "false"},
		{EventsPullPoint, "true", "false"},
		{EventsBaseSubscription, "false", "true"},
		{EventsBoth, "true", "true"},
	}
	for _, c := range cases {
		s := &ServiceContext{EventsEnable: c.ev}
		pull, bsub := s.eventsFlags()
		if pull != c.pull || bsub != c.bsub {
			t.Errorf("eventsFlags(%d) = (%q,%q); want (%q,%q)", c.ev, pull, bsub, c.pull, c.bsub)
		}
	}
}

// --- getInterfaceIP fallback ---

func TestGetInterfaceIP_UnknownFallback(t *testing.T) {
	ip, err := getInterfaceIP("nonexistent_interface_xyz")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1 fallback, got %q", ip)
	}
}

func TestGetInterfaceIP_Empty(t *testing.T) {
	ip, err := getInterfaceIP("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1 for empty iface, got %q", ip)
	}
}

// --- serviceAddrs port formatting ---

func TestServiceAddrs_Port80Omitted(t *testing.T) {
	s := &ServiceContext{Port: 80, Interface: ""}
	_, dev, _, _, _, _, err := s.serviceAddrs()
	if err != nil {
		t.Fatalf("serviceAddrs: %v", err)
	}
	if got, notWant := dev, "127.0.0.1:80"; got == notWant {
		t.Errorf("port 80 should be omitted, got %q", got)
	}
}

func TestServiceAddrs_NonDefaultPort(t *testing.T) {
	s := &ServiceContext{Port: 8080, Interface: ""}
	_, dev, _, _, _, _, err := s.serviceAddrs()
	if err != nil {
		t.Fatalf("serviceAddrs: %v", err)
	}
	want := "http://127.0.0.1:8080/onvif/device_service"
	if dev != want {
		t.Errorf("serviceAddrs device = %q; want %q", dev, want)
	}
}

// --- GetCapabilitiesHTTP category dispatch (icategory derivation) ---

func TestCategoryDispatch(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"Device", 1}, {"device", 1}, {"DEVICE", 1},
		{"Media", 2}, {"media", 2},
		{"PTZ", 4}, {"ptz", 4},
		{"Events", 8}, {"events", 8},
		{"All", 15}, {"all", 15}, {"ALL", 15},
		{"", 15},
	}
	for _, c := range cases {
		got := categoryCode(c.input)
		if got != c.want {
			t.Errorf("categoryCode(%q) = %d; want %d", c.input, got, c.want)
		}
	}
}
