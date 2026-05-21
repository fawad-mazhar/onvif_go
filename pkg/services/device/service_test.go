package device

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fawad-mazhar/onvif-go/internal/xml"
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
		ev         EventsEnable
		pull, bsub string
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
	want := "http://127.0.0.1/onvif/device_service"
	if dev != want {
		t.Errorf("serviceAddrs device = %q; want %q", dev, want)
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

// setTemplateDirs configures xml template dirs relative to this package's
// location (pkg/services/device/ → ../../../ is the repo root).
// NOT safe under t.Parallel() — mutates package-level xml vars.
func setTemplateDirs(t *testing.T) func() {
	t.Helper()
	const root = "../../../service_files"
	prevSvc := xml.ServiceTemplateDir
	prevGen := xml.GenericTemplateDir
	xml.ServiceTemplateDir = root
	xml.GenericTemplateDir = root + "/generic"
	return func() {
		xml.ServiceTemplateDir = prevSvc
		xml.GenericTemplateDir = prevGen
	}
}

// --- GetCapabilitiesHTTP fault-response shape ---

func TestGetCapabilitiesHTTP_UnknownCategory_Fault(t *testing.T) {
	defer setTemplateDirs(t)()
	s := &ServiceContext{Port: 8080, Interface: "", PTZEnable: false}
	soap := `<s:Envelope><s:Body><tds:GetCapabilities><Category>Bogus</Category></tds:GetCapabilities></s:Body></s:Envelope>`
	r := httptest.NewRequest(http.MethodPost, "/onvif/device_service", nil)
	r.Host = "192.0.2.1:8080"
	rec := httptest.NewRecorder()
	if err := s.GetCapabilitiesHTTP(rec, r, soap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("unknown category: status = %d; want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ter:NoSuchService") {
		t.Errorf("unknown category: body missing ter:NoSuchService:\n%s", rec.Body.String())
	}
}

func TestGetCapabilitiesHTTP_PTZCategoryWhenDisabled_Fault(t *testing.T) {
	defer setTemplateDirs(t)()
	s := &ServiceContext{Port: 8080, Interface: "", PTZEnable: false}
	soap := `<s:Envelope><s:Body><tds:GetCapabilities><Category>PTZ</Category></tds:GetCapabilities></s:Body></s:Envelope>`
	r := httptest.NewRequest(http.MethodPost, "/onvif/device_service", nil)
	r.Host = "192.0.2.1:8080"
	rec := httptest.NewRecorder()
	if err := s.GetCapabilitiesHTTP(rec, r, soap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("PTZ disabled: status = %d; want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ter:NoSuchService") {
		t.Errorf("PTZ disabled: body missing ter:NoSuchService:\n%s", rec.Body.String())
	}
}

func TestGetCapabilitiesHTTP_PTZCategoryWhenEnabled_OK(t *testing.T) {
	defer setTemplateDirs(t)()
	s := &ServiceContext{Port: 8080, Interface: "", PTZEnable: true}
	soap := `<s:Envelope><s:Body><tds:GetCapabilities><Category>PTZ</Category></tds:GetCapabilities></s:Body></s:Envelope>`
	r := httptest.NewRequest(http.MethodPost, "/onvif/device_service", nil)
	r.Host = "192.0.2.1:8080"
	rec := httptest.NewRecorder()
	if err := s.GetCapabilitiesHTTP(rec, r, soap); err != nil {
		t.Fatalf("PTZ enabled: unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("PTZ enabled: status = %d; want 200", rec.Code)
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
		{"bogus", -1}, {"BadCategory", -1}, {"MEDIA2", -1},
	}
	for _, c := range cases {
		got := categoryCode(c.input)
		if got != c.want {
			t.Errorf("categoryCode(%q) = %d; want %d", c.input, got, c.want)
		}
	}
}
