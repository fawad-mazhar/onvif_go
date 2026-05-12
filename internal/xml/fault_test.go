package xml

import (
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resolves service_files/ relative to the test file so the
// tests work regardless of `go test`'s working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	// this file is /<repo>/internal/xml/fault_test.go -> up 3
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func withTemplates(t *testing.T) func() {
	t.Helper()
	prev := GenericTemplateDir
	GenericTemplateDir = filepath.Join(repoRoot(t), "service_files", "generic")
	return func() { GenericTemplateDir = prev }
}

func TestRenderFault(t *testing.T) {
	defer withTemplates(t)()

	body, status, err := RenderFault(Fault{
		Service:        "media_service",
		DeviceAddress:  "http://example.test:8080/onvif",
		ServiceAddress: "http://example.test:8080/onvif/media_service",
		RecSend:        "Receiver",
		Subcode:        "ter:Action",
		SubcodeEx:      "ter:MaxNVTProfiles",
		Reason:         "Max profile number reached",
		Detail:         "The maximum number of supported profiles supported by the device has been reached",
	})
	if err != nil {
		t.Fatalf("RenderFault: %v", err)
	}
	if status != 500 {
		t.Errorf("status = %d; want 500", status)
	}

	s := string(body)
	for _, want := range []string{
		`"urn:uuid:`, // C-quirk: literal quotes around UUID
		`<wsa5:Address>http://example.test:8080/onvif</wsa5:Address>`,
		`<wsa5:To SOAP-ENV:mustUnderstand="true">http://example.test:8080/onvif/media_service</wsa5:To>`,
		`<SOAP-ENV:Value>SOAP-ENV:Receiver</SOAP-ENV:Value>`,
		`<SOAP-ENV:Value>ter:Action</SOAP-ENV:Value>`,
		`<SOAP-ENV:Value>ter:MaxNVTProfiles</SOAP-ENV:Value>`,
		`Max profile number reached`,
		`maximum number of supported profiles`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("body missing %q", want)
		}
	}

	// C-quirk: Fault.xml has no <?xml ...?> prolog.
	if strings.HasPrefix(s, "<?xml") {
		t.Errorf("body unexpectedly starts with XML prolog:\n%s", s[:60])
	}
}

func TestRenderAuthenticationError(t *testing.T) {
	defer withTemplates(t)()

	body, status, err := RenderAuthenticationError()
	if err != nil {
		t.Fatalf("RenderAuthenticationError: %v", err)
	}
	if status != 400 {
		t.Errorf("status = %d; want 400", status)
	}
	s := string(body)
	for _, want := range []string{
		`<?xml version="1.0" encoding="utf-8"?>`, // C quirk: prolog present, lowercase
		`<env:Value>ter:NotAuthorized</env:Value>`,
		`Sender not Authorized`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	defer withTemplates(t)()

	body, status, err := RenderEmpty("tds", "SystemReboot")
	if err != nil {
		t.Fatalf("RenderEmpty: %v", err)
	}
	if status != 200 {
		t.Errorf("status = %d; want 200", status)
	}
	s := string(body)
	if !strings.Contains(s, `<tds:SystemRebootResponse />`) {
		t.Errorf("body missing expected empty response tag:\n%s", s)
	}
}

func TestRenderPullMessagesFault(t *testing.T) {
	defer withTemplates(t)()

	body, _, err := RenderPullMessagesFault("PT60S", "1024")
	if err != nil {
		t.Fatalf("RenderPullMessagesFault: %v", err)
	}
	s := string(body)
	for _, want := range []string{
		`<MaxTimeout>PT60S</MaxTimeout>`,
		`<MaxMessageLimit>1024</MaxMessageLimit>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestGenUUIDShape(t *testing.T) {
	id, err := genUUID()
	if err != nil {
		t.Fatalf("genUUID: %v", err)
	}
	// 8-4-4-4-12 hex groups.
	parts := strings.Split(id, "-")
	wantLens := []int{8, 4, 4, 4, 12}
	if len(parts) != 5 {
		t.Fatalf("uuid groups = %d; want 5 (%q)", len(parts), id)
	}
	for i, p := range parts {
		if len(p) != wantLens[i] {
			t.Errorf("group[%d] len = %d; want %d (%q)", i, len(p), wantLens[i], id)
		}
	}
	// Two separate IDs must differ (probability of collision: negligible).
	id2, _ := genUUID()
	if id == id2 {
		t.Errorf("genUUID returned identical ids twice: %q", id)
	}
}

func TestWriteFault(t *testing.T) {
	defer withTemplates(t)()

	rec := httptest.NewRecorder()
	err := WriteFault(rec, Fault{
		Service:        "device_service",
		DeviceAddress:  "http://h/onvif",
		ServiceAddress: "http://h/onvif/device_service",
		RecSend:        "Sender",
		Subcode:        "ter:Action",
		SubcodeEx:      "ter:ActionNotSupported",
		Reason:         "unsupported",
		Detail:         "detail",
	})
	if err != nil {
		t.Fatalf("WriteFault: %v", err)
	}
	if got := rec.Code; got != 500 {
		t.Errorf("status = %d; want 500", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/soap+xml; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}
