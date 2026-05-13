package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeConf(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write conf: %v", err)
	}
	return p
}

func TestLoadConfig_Defaults(t *testing.T) {
	p := writeConf(t, t.TempDir(), "")
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig empty: %v", err)
	}
	// C initialises adv_fault_if_unknown to 0 (conf.c:59).
	if cfg.AdvFaultIfUnknown != 0 {
		t.Errorf("AdvFaultIfUnknown default = %d; want 0", cfg.AdvFaultIfUnknown)
	}
	if cfg.AdvFaultIfSet != 0 {
		t.Errorf("AdvFaultIfSet default = %d; want 0", cfg.AdvFaultIfSet)
	}
	if cfg.AdvEnableMedia2 != 0 {
		t.Errorf("AdvEnableMedia2 default = %d; want 0", cfg.AdvEnableMedia2)
	}
	if cfg.AdvSynologyNVR != 0 {
		t.Errorf("AdvSynologyNVR default = %d; want 0", cfg.AdvSynologyNVR)
	}
}

func TestLoadConfig_AdvFlags(t *testing.T) {
	p := writeConf(t, t.TempDir(), `
adv_fault_if_unknown=1
adv_fault_if_set=1
adv_enable_media2=1
adv_synology_nvr=1
`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.AdvFaultIfUnknown != 1 {
		t.Errorf("AdvFaultIfUnknown = %d; want 1", cfg.AdvFaultIfUnknown)
	}
	if cfg.AdvFaultIfSet != 1 {
		t.Errorf("AdvFaultIfSet = %d; want 1", cfg.AdvFaultIfSet)
	}
	if cfg.AdvEnableMedia2 != 1 {
		t.Errorf("AdvEnableMedia2 = %d; want 1", cfg.AdvEnableMedia2)
	}
	if cfg.AdvSynologyNVR != 1 {
		t.Errorf("AdvSynologyNVR = %d; want 1", cfg.AdvSynologyNVR)
	}
}

func TestLoadConfig_BasicFields(t *testing.T) {
	p := writeConf(t, t.TempDir(), `
manufacturer=Acme
model=Cam01
firmware_ver=1.0.0
serial_num=SN123
hardware_id=HW456
uuid=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
port=8080
`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	checks := map[string]string{
		"Manufacturer": cfg.Manufacturer,
		"Model":        cfg.Model,
		"FirmwareVer":  cfg.FirmwareVer,
		"SerialNum":    cfg.SerialNum,
		"HardwareID":   cfg.HardwareID,
		"UUID":         cfg.UUID,
	}
	want := map[string]string{
		"Manufacturer": "Acme",
		"Model":        "Cam01",
		"FirmwareVer":  "1.0.0",
		"SerialNum":    "SN123",
		"HardwareID":   "HW456",
		"UUID":         "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
	}
	for field, got := range checks {
		if got != want[field] {
			t.Errorf("%s = %q; want %q", field, got, want[field])
		}
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d; want 8080", cfg.Port)
	}
}

func TestLoadConfig_ProfileAndScope(t *testing.T) {
	p := writeConf(t, t.TempDir(), `
profile.0.name=main
profile.0.width=1920
profile.0.height=1080
profile.0.url=rtsp://cam/stream0
profile.0.type=h264
scope.0=onvif://www.onvif.org/Profile/Streaming
scope.1=onvif://www.onvif.org/name/Cam01
`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("Profiles len = %d; want 1", len(cfg.Profiles))
	}
	p0 := cfg.Profiles[0]
	if p0.Name != "main" {
		t.Errorf("profile.Name = %q; want main", p0.Name)
	}
	if p0.Width != 1920 || p0.Height != 1080 {
		t.Errorf("profile dims = %dx%d; want 1920x1080", p0.Width, p0.Height)
	}
	if p0.URL != "rtsp://cam/stream0" {
		t.Errorf("profile.URL = %q", p0.URL)
	}
	if p0.Type != H264 {
		t.Errorf("profile.Type = %v; want H264", p0.Type)
	}
	if len(cfg.Scopes) != 2 {
		t.Fatalf("Scopes len = %d; want 2", len(cfg.Scopes))
	}
}

func TestLoadConfig_CommentAndBlankLines(t *testing.T) {
	p := writeConf(t, t.TempDir(), `
# This is a comment
manufacturer=Test

# Another comment
model=Cam
`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Manufacturer != "Test" || cfg.Model != "Cam" {
		t.Errorf("comments/blanks not skipped: %+v", cfg)
	}
}

// TestLoadConfig_CanonicalFixture is the canary test for the parity project.
// It loads the shared fixture config (C-format flat keys) and asserts that
// every section is parsed — not silently dropped. A failure here means Go
// will render fewer items than the C reference server for the same config.
func TestLoadConfig_CanonicalFixture(t *testing.T) {
	// The fixture lives at test/fixtures/config/server.conf relative to the
	// repo root; from internal/config the root is ../..
	repoRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	confPath := filepath.Join(repoRoot, "test", "fixtures", "config", "server.conf")
	if _, err := os.Stat(confPath); err != nil {
		t.Fatalf("canonical fixture config not found at %s: %v", confPath, err)
	}
	cfg, err := LoadConfig(confPath)
	if err != nil {
		t.Fatalf("LoadConfig canonical fixture: %v", err)
	}

	// Basic scalar fields
	if cfg.Manufacturer != "Manufacturer" {
		t.Errorf("Manufacturer = %q; want Manufacturer", cfg.Manufacturer)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d; want 8080", cfg.Port)
	}

	// adv defaults (fixture sets all to 0)
	if cfg.AdvFaultIfUnknown != 0 {
		t.Errorf("AdvFaultIfUnknown = %d; want 0", cfg.AdvFaultIfUnknown)
	}

	// 4 scope= lines
	if len(cfg.Scopes) != 4 {
		t.Errorf("Scopes len = %d; want 4", len(cfg.Scopes))
	}

	// 2 profiles (Profile_0 and Profile_1)
	if len(cfg.Profiles) != 2 {
		t.Fatalf("Profiles len = %d; want 2 (config parser is silently dropping profiles)", len(cfg.Profiles))
	}
	p0 := cfg.Profiles[0]
	if p0.Name != "Profile_0" {
		t.Errorf("Profiles[0].Name = %q; want Profile_0", p0.Name)
	}
	if p0.Width != 1920 || p0.Height != 1080 {
		t.Errorf("Profiles[0] dims = %dx%d; want 1920x1080", p0.Width, p0.Height)
	}
	if p0.Type != H264 {
		t.Errorf("Profiles[0].Type = %v; want H264", p0.Type)
	}
	p1 := cfg.Profiles[1]
	if p1.Name != "Profile_1" {
		t.Errorf("Profiles[1].Name = %q; want Profile_1", p1.Name)
	}
	if p1.Width != 640 || p1.Height != 360 {
		t.Errorf("Profiles[1] dims = %dx%d; want 640x360", p1.Width, p1.Height)
	}

	// PTZ enabled
	if cfg.PTZNode.Enable != 1 {
		t.Errorf("PTZNode.Enable = %d; want 1", cfg.PTZNode.Enable)
	}
	if cfg.PTZNode.MaxStepX != 360.0 {
		t.Errorf("PTZNode.MaxStepX = %v; want 360", cfg.PTZNode.MaxStepX)
	}
	if cfg.PTZNode.GetPosition == "" {
		t.Errorf("PTZNode.GetPosition is empty; want command path")
	}

	// 2 relay outputs
	if len(cfg.RelayOutputs) != 2 {
		t.Fatalf("RelayOutputs len = %d; want 2", len(cfg.RelayOutputs))
	}
	if cfg.RelayOutputs[0].IdleState != IdleStateOpen {
		t.Errorf("RelayOutputs[0].IdleState = %v; want IdleStateOpen", cfg.RelayOutputs[0].IdleState)
	}
	if cfg.RelayOutputs[0].CloseCmd == "" {
		t.Errorf("RelayOutputs[0].CloseCmd is empty")
	}

	// events=3 → EventsBoth; 3 explicit event entries
	if cfg.EventsEnable != EventsBoth {
		t.Errorf("EventsEnable = %v; want EventsBoth", cfg.EventsEnable)
	}
	if len(cfg.Events) != 3 {
		t.Fatalf("Events len = %d; want 3", len(cfg.Events))
	}
	if cfg.Events[0].Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("Events[0].Topic = %q", cfg.Events[0].Topic)
	}
	if cfg.Events[0].InputFile == "" {
		t.Errorf("Events[0].InputFile is empty")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/to.conf")
	if err == nil {
		t.Errorf("expected error for missing config file")
	}
}

// ---- Phase 0c: JSON config loader tests -----------------------------------

func writeJSON(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "test.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write json: %v", err)
	}
	return p
}

// TestLoadConfigJSON_CanonicalFixture is the JSON counterpart of
// TestLoadConfig_CanonicalFixture. Both must assert the same values.
func TestLoadConfigJSON_CanonicalFixture(t *testing.T) {
	repoRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	jsonPath := filepath.Join(repoRoot, "test", "fixtures", "config", "server.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("canonical JSON fixture not found at %s: %v", jsonPath, err)
	}
	cfg, err := LoadConfigJSON(jsonPath)
	if err != nil {
		t.Fatalf("LoadConfigJSON: %v", err)
	}

	if cfg.Manufacturer != "Manufacturer" {
		t.Errorf("Manufacturer = %q; want Manufacturer", cfg.Manufacturer)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d; want 8080", cfg.Port)
	}
	if len(cfg.Scopes) != 4 {
		t.Errorf("Scopes len = %d; want 4", len(cfg.Scopes))
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("Profiles len = %d; want 2", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Name != "Profile_0" {
		t.Errorf("Profiles[0].Name = %q; want Profile_0", cfg.Profiles[0].Name)
	}
	if cfg.Profiles[0].Width != 1920 || cfg.Profiles[0].Height != 1080 {
		t.Errorf("Profiles[0] dims = %dx%d; want 1920x1080", cfg.Profiles[0].Width, cfg.Profiles[0].Height)
	}
	if cfg.Profiles[0].Type != H264 {
		t.Errorf("Profiles[0].Type = %v; want H264", cfg.Profiles[0].Type)
	}
	if cfg.Profiles[0].AudioDecoder != G711 {
		t.Errorf("Profiles[0].AudioDecoder = %v; want G711", cfg.Profiles[0].AudioDecoder)
	}
	if cfg.Profiles[1].Name != "Profile_1" {
		t.Errorf("Profiles[1].Name = %q; want Profile_1", cfg.Profiles[1].Name)
	}
	if cfg.Profiles[1].AudioDecoder != AudioNone {
		t.Errorf("Profiles[1].AudioDecoder = %v; want AudioNone", cfg.Profiles[1].AudioDecoder)
	}
	if cfg.PTZNode.Enable != 1 {
		t.Errorf("PTZNode.Enable = %d; want 1", cfg.PTZNode.Enable)
	}
	if cfg.PTZNode.MaxStepX != 360.0 {
		t.Errorf("PTZNode.MaxStepX = %v; want 360", cfg.PTZNode.MaxStepX)
	}
	if cfg.PTZNode.GetPosition == "" {
		t.Errorf("PTZNode.GetPosition is empty")
	}
	if len(cfg.RelayOutputs) != 2 {
		t.Fatalf("RelayOutputs len = %d; want 2", len(cfg.RelayOutputs))
	}
	if cfg.RelayOutputs[0].IdleState != IdleStateOpen {
		t.Errorf("RelayOutputs[0].IdleState = %v; want IdleStateOpen", cfg.RelayOutputs[0].IdleState)
	}
	if cfg.RelayOutputs[0].CloseCmd == "" {
		t.Errorf("RelayOutputs[0].CloseCmd is empty")
	}
	if cfg.EventsEnable != EventsBoth {
		t.Errorf("EventsEnable = %v; want EventsBoth", cfg.EventsEnable)
	}
	if len(cfg.Events) != 3 {
		t.Fatalf("Events len = %d; want 3", len(cfg.Events))
	}
	if cfg.Events[0].Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("Events[0].Topic = %q", cfg.Events[0].Topic)
	}
}

// TestLoadConfigJSON_RoundTrip is the oracle for Phase 0c: load the same
// logical config from both formats and assert the resulting ServiceContext
// structs are deeply equal (every field, not just a hand-picked subset).
func TestLoadConfigJSON_RoundTrip(t *testing.T) {
	repoRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	confPath := filepath.Join(repoRoot, "test", "fixtures", "config", "server.conf")
	jsonPath := filepath.Join(repoRoot, "test", "fixtures", "config", "server.json")
	for _, p := range []string{confPath, jsonPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("fixture not found at %s: %v", p, err)
		}
	}

	conf, err := LoadConfig(confPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	jcfg, err := LoadConfigJSON(jsonPath)
	if err != nil {
		t.Fatalf("LoadConfigJSON: %v", err)
	}

	if !reflect.DeepEqual(conf, jcfg) {
		t.Errorf("conf and json configs produce different ServiceContext structs:\nconf: %+v\njson: %+v", conf, jcfg)
	}
}

// TestLoadConfigJSON_SparseDefaults verifies that the JSON loader applies the
// same defaults as parseFlatConfig for fields omitted from the JSON:
//   - profile with no "type" → H264
//   - profile with no "audio_encoder" → AAC
//   - ptz with enable=1 and no max_step_x/y → 360/180
func TestLoadConfigJSON_SparseDefaults(t *testing.T) {
	dir := t.TempDir()

	// --- sparse profile ---
	jsonP := writeJSON(t, dir, `{"profiles":[{"name":"x"}]}`)
	jcfg, err := LoadConfigJSON(jsonP)
	if err != nil {
		t.Fatalf("LoadConfigJSON sparse profile: %v", err)
	}
	if len(jcfg.Profiles) != 1 {
		t.Fatalf("Profiles len = %d; want 1", len(jcfg.Profiles))
	}
	if jcfg.Profiles[0].Type != H264 {
		t.Errorf("sparse profile Type = %v; want H264 (default)", jcfg.Profiles[0].Type)
	}
	if jcfg.Profiles[0].AudioEncoder != AAC {
		t.Errorf("sparse profile AudioEncoder = %v; want AAC (default)", jcfg.Profiles[0].AudioEncoder)
	}
	if jcfg.Profiles[0].AudioDecoder != AudioNone {
		t.Errorf("sparse profile AudioDecoder = %v; want AudioNone (default)", jcfg.Profiles[0].AudioDecoder)
	}

	// compare against equivalent sparse .conf
	confP := writeConf(t, dir, "name=x\n")
	ccfg, err := LoadConfig(confP)
	if err != nil {
		t.Fatalf("LoadConfig sparse profile: %v", err)
	}
	if !reflect.DeepEqual(ccfg.Profiles, jcfg.Profiles) {
		t.Errorf("sparse profile mismatch: conf=%+v json=%+v", ccfg.Profiles, jcfg.Profiles)
	}

	// --- sparse PTZ ---
	jsonPTZ := writeJSON(t, dir, `{"ptz":{"enable":1}}`)
	ptz, err := LoadConfigJSON(jsonPTZ)
	if err != nil {
		t.Fatalf("LoadConfigJSON sparse ptz: %v", err)
	}
	if ptz.PTZNode.MaxStepX != 360.0 {
		t.Errorf("sparse ptz MaxStepX = %v; want 360 (default)", ptz.PTZNode.MaxStepX)
	}
	if ptz.PTZNode.MaxStepY != 180.0 {
		t.Errorf("sparse ptz MaxStepY = %v; want 180 (default)", ptz.PTZNode.MaxStepY)
	}

	// compare against equivalent sparse .conf
	confPTZ := writeConf(t, dir, "ptz=1\n")
	ccfgPTZ, err := LoadConfig(confPTZ)
	if err != nil {
		t.Fatalf("LoadConfig sparse ptz: %v", err)
	}
	if !reflect.DeepEqual(ccfgPTZ.PTZNode, ptz.PTZNode) {
		t.Errorf("sparse ptz mismatch: conf=%+v json=%+v", ccfgPTZ.PTZNode, ptz.PTZNode)
	}
}

// TestLoad_ByExtension verifies that Load dispatches to the correct loader
// based on the file extension (.json or anything else).
func TestLoad_ByExtension(t *testing.T) {
	dir := t.TempDir()

	// .json → LoadConfigJSON
	jsonFile := writeJSON(t, dir, `{"port":9090}`)
	cfg, err := Load(jsonFile)
	if err != nil {
		t.Fatalf("Load(.json): %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Load(.json) port = %d; want 9090", cfg.Port)
	}

	// .conf → LoadConfig
	confFile := writeConf(t, dir, "port=7070\n")
	cfg, err = Load(confFile)
	if err != nil {
		t.Fatalf("Load(.conf): %v", err)
	}
	if cfg.Port != 7070 {
		t.Errorf("Load(.conf) port = %d; want 7070", cfg.Port)
	}

	// .JSON (upper-case) → LoadConfigJSON (case-insensitive extension check)
	jsonUpper := filepath.Join(dir, "upper.JSON")
	if err := os.WriteFile(jsonUpper, []byte(`{"port":5050}`), 0o600); err != nil {
		t.Fatalf("write upper.JSON: %v", err)
	}
	cfg, err = Load(jsonUpper)
	if err != nil {
		t.Fatalf("Load(.JSON): %v", err)
	}
	if cfg.Port != 5050 {
		t.Errorf("Load(.JSON) port = %d; want 5050", cfg.Port)
	}
}

func TestLoadConfigJSON_Defaults(t *testing.T) {
	p := writeJSON(t, t.TempDir(), `{"port": 8888}`)
	cfg, err := LoadConfigJSON(p)
	if err != nil {
		t.Fatalf("LoadConfigJSON: %v", err)
	}
	if cfg.Port != 8888 {
		t.Errorf("Port = %d; want 8888", cfg.Port)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("Profiles should be empty for minimal config")
	}
	if cfg.PTZNode.Enable != 0 {
		t.Errorf("PTZNode.Enable should default to 0")
	}
	if cfg.EventsEnable != EventsNone {
		t.Errorf("EventsEnable should default to EventsNone")
	}
}

func TestLoadConfigJSON_InvalidJSON(t *testing.T) {
	p := writeJSON(t, t.TempDir(), `{not valid json}`)
	_, err := LoadConfigJSON(p)
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

func TestLoadConfigJSON_MissingFile(t *testing.T) {
	_, err := LoadConfigJSON("/nonexistent/path/to.json")
	if err == nil {
		t.Errorf("expected error for missing JSON config file")
	}
}

func TestLoadConfig_PortOutOfRange(t *testing.T) {
	p := writeConf(t, t.TempDir(), "port=99999\n")
	_, err := LoadConfig(p)
	if err == nil {
		t.Errorf("expected error for out-of-range port")
	}
}
