package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	xmlpkg "github.com/fawad-mazhar/onvif-go/internal/xml"
)

// envelope wraps body in a minimal SOAP 1.2 Envelope using the <s:> prefix
// that the captured fixtures use — exercises ExtractBodyAction on the real path.
func envelope(action string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">` +
		`<s:Body><tds:` + action + ` xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/></s:Body>` +
		`</s:Envelope>`)
}

func testCfg(t *testing.T, faultIfUnknown int) *config.ServiceContext {
	t.Helper()
	// Load the fixture config; it lives at test/fixtures/config/server.conf
	// relative to repo root. From internal/server the repo root is ../..
	repoRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	xmlpkg.ServiceTemplateDir = filepath.Join(repoRoot, "service_files")
	xmlpkg.GenericTemplateDir = filepath.Join(repoRoot, "service_files", "generic")
	cfg, err := config.LoadConfig(filepath.Join(repoRoot, "test", "fixtures", "config", "server.conf"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cfg.AdvFaultIfUnknown = faultIfUnknown
	return cfg
}

func post(t *testing.T, ts *httptest.Server, path string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest("POST", ts.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	b, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	return string(b)
}

// TestAdvFaultIfUnknown_DefaultEmptyResponse verifies that when
// adv_fault_if_unknown == 0 (C default) an unsupported op returns HTTP 200
// with an empty <tds:FooResponse/> body, matching C's send_empty_response().
func TestAdvFaultIfUnknown_DefaultEmptyResponse(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "service_files", "generic", "Empty.xml")); err != nil {
		t.Skip("service_files/generic/Empty.xml not found — run from repo root or adjust path")
	}
	cfg := testCfg(t, 0)
	ts := httptest.NewServer(BuildRouter(cfg))
	defer ts.Close()

	resp := post(t, ts, "/onvif/device_service", envelope("FooUnknownOp"))
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d; want 200", resp.StatusCode)
	}
	if !strings.Contains(body, "FooUnknownOp") {
		t.Errorf("empty response missing op name:\n%s", body)
	}
	if strings.Contains(body, "<soap:Fault") || strings.Contains(body, "SOAP-ENV:Fault") {
		t.Errorf("unexpected fault in default (adv_fault_if_unknown=0) path:\n%s", body)
	}
}

// TestAdvFaultIfUnknown_FaultWhenSet verifies that when
// adv_fault_if_unknown == 1 an unsupported op returns a SOAP fault, matching
// C's send_action_failed_fault().
func TestAdvFaultIfUnknown_FaultWhenSet(t *testing.T) {
	cfg := testCfg(t, 1)
	ts := httptest.NewServer(BuildRouter(cfg))
	defer ts.Close()

	resp := post(t, ts, "/onvif/device_service", envelope("FooUnknownOp"))
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d; want 500", resp.StatusCode)
	}
	if !strings.Contains(body, "Fault") {
		t.Errorf("expected SOAP fault body:\n%s", body)
	}
}

// TestAdvFaultIfUnknown_AllServices verifies all five known services use
// sendUnsupportedResponse (and not a hard-coded fault).
func TestAdvFaultIfUnknown_AllServices(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "service_files", "generic", "Empty.xml")); err != nil {
		t.Skip("service_files/generic/Empty.xml not found")
	}
	cfg := testCfg(t, 0)
	ts := httptest.NewServer(BuildRouter(cfg))
	defer ts.Close()

	services := map[string][]byte{
		"/onvif/device_service":  envelope("NoSuchDeviceOp"),
		"/onvif/media_service":   envelope("NoSuchMediaOp"),
		"/onvif/ptz_service":     envelope("NoSuchPTZOp"),
		"/onvif/events_service":  envelope("NoSuchEventsOp"),
		"/onvif/deviceio_service": envelope("NoSuchDeviceIOOp"),
	}
	for path, body := range services {
		resp := post(t, ts, path, body)
		_ = readBody(t, resp)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s unsupported op: status = %d; want 200 (adv_fault_if_unknown=0)", path, resp.StatusCode)
		}
	}
}

// TestAdvFaultIfSet_MediaSetOps verifies that when adv_fault_if_set == 1
// the five Set* ops return a fault regardless of adv_fault_if_unknown.
func TestAdvFaultIfSet_MediaSetOps(t *testing.T) {
	cfg := testCfg(t, 0)
	cfg.AdvFaultIfSet = 1
	ts := httptest.NewServer(BuildRouter(cfg))
	defer ts.Close()

	setOps := []string{
		"SetVideoSourceConfiguration",
		"SetAudioSourceConfiguration",
		"SetVideoEncoderConfiguration",
		"SetAudioEncoderConfiguration",
		"SetAudioOutputConfiguration",
	}
	for _, op := range setOps {
		req := []byte(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">` +
			`<s:Body><trt:` + op + ` xmlns:trt="http://www.onvif.org/ver10/media/wsdl"/></s:Body>` +
			`</s:Envelope>`)
		resp := post(t, ts, "/onvif/media_service", req)
		body := readBody(t, resp)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("%s: status = %d; want 500 (adv_fault_if_set=1)", op, resp.StatusCode)
		}
		if !strings.Contains(body, "Fault") {
			t.Errorf("%s: expected SOAP fault:\n%s", op, body)
		}
	}
}
