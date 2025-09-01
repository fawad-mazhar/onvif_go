package tests

import (
	"testing"
	"os"
	"path/filepath"

	"github.com/fawad-mazhar/onvif_go/config"
	"github.com/fawad-mazhar/onvif_go/services/device"
	"github.com/fawad-mazhar/onvif_go/services/events"
	"github.com/fawad-mazhar/onvif_go/services/media"
	"github.com/fawad-mazhar/onvif_go/services/ptz"
	"github.com/fawad-mazhar/onvif_go/services/deviceio"
)

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := os.TempDir()
	configPath := filepath.Join(tempDir, "test_config.conf")
	
	configContent := `# ONVIF Server Configuration
port = 8080
wsd_port = 3702
notification_port = 8081

manufacturer = "TestManufacturer"
model = "TestModel"
firmware_ver = "1.0.0"
serial_num = "1234567890"
hardware_id = "HW123"

user = "admin"
password = "admin123"

uuid = "12345678-1234-1234-1234-123456789012"

profile.0.name = "Profile1"
profile.0.width = 1920
profile.0.height = 1080
profile.0.url = "rtsp://localhost:554/stream1"
profile.0.snapurl = "http://localhost:8080/snapshot1"
profile.0.type = "h264"
profile.0.audio_encoder = "aac"
profile.0.audio_decoder = "aac"

ptz.enable = 1
ptz.min_step_x = -180.0
ptz.max_step_x = 180.0
ptz.min_step_y = -90.0
ptz.max_step_y = 90.0
ptz.min_step_z = 0.0
ptz.max_step_z = 100.0

event.0.topic = "tns1:VideoSource/MotionAlarm"

relayoutput.0.idle_state = "open"

scope.0 = "onvif://www.onvif.org/name/TestDevice"
scope.1 = "onvif://www.onvif.org/location/TestLocation"
`
	
	// Write the config content to the temporary file
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}
	
	// Clean up the temporary file after the test
	defer os.Remove(configPath)
	
	// Test loading the configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Verify configuration values
	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}
	
	if cfg.WSDPort != 3702 {
		t.Errorf("Expected WSD port 3702, got %d", cfg.WSDPort)
	}
	
	if cfg.NotificationPort != 8081 {
		t.Errorf("Expected notification port 8081, got %d", cfg.NotificationPort)
	}
	
	if cfg.Manufacturer != "TestManufacturer" {
		t.Errorf("Expected manufacturer TestManufacturer, got %s", cfg.Manufacturer)
	}
	
	if cfg.Model != "TestModel" {
		t.Errorf("Expected model TestModel, got %s", cfg.Model)
	}
	
	if cfg.FirmwareVer != "1.0.0" {
		t.Errorf("Expected firmware version 1.0.0, got %s", cfg.FirmwareVer)
	}
	
	if cfg.SerialNum != "1234567890" {
		t.Errorf("Expected serial number 1234567890, got %s", cfg.SerialNum)
	}
	
	if cfg.HardwareId != "HW123" {
		t.Errorf("Expected hardware ID HW123, got %s", cfg.HardwareId)
	}
	
	if cfg.User != "admin" {
		t.Errorf("Expected username admin, got %s", cfg.User)
	}
	
	if cfg.Password != "admin123" {
		t.Errorf("Expected password admin123, got %s", cfg.Password)
	}
	
	if cfg.UUID != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("Expected UUID 12345678-1234-1234-1234-123456789012, got %s", cfg.UUID)
	}
	
	// Verify profiles
	if len(cfg.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(cfg.Profiles))
	}
	
	// Verify PTZ nodes
	if cfg.PTZNode.Enable != 1 {
		t.Errorf("Expected PTZ node enabled, got %d", cfg.PTZNode.Enable)
	}
	
	// Verify events
	if len(cfg.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(cfg.Events))
	}
	
	// Verify relay outputs
	if len(cfg.RelayOutputs) != 1 {
		t.Errorf("Expected 1 relay output, got %d", len(cfg.RelayOutputs))
	}
	
	// Verify scopes
	if len(cfg.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(cfg.Scopes))
	}
}

func TestDeviceService(t *testing.T) {
	// Create a device service context
	deviceService := &device.ServiceContext{
		Manufacturer: "TestManufacturer",
		Model: "TestModel",
		FirmwareVer: "1.0.0",
		SerialNum: "1234567890",
		HardwareId: "HW123",
		Port: 8080,
		Scopes: []string{
			"onvif://www.onvif.org/name/TestDevice",
			"onvif://www.onvif.org/location/TestLocation",
		},
	}
	
	// Test GetDeviceInformation method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	deviceService.GetDeviceInformation()
}

func TestMediaService(t *testing.T) {
	// Create a media service context with a profile
	profiles := []media.Profile{
		{
			Name: "Profile1",
			Width: 1920,
			Height: 1080,
			URL: "rtsp://localhost:554/stream1",
			SnapURL: "http://localhost:8080/snapshot1",
			Type: "H264",
			AudioEncoder: "AAC",
			AudioDecoder: "AAC",
		},
	}
	
	mediaService := &media.ServiceContext{
		Port: 8080,
		Profiles: profiles,
	}
	
	// Test GetProfiles method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	mediaService.GetProfiles()
}

func TestPTZService(t *testing.T) {
	// Create a PTZ service context with a node
	ptzNodes := []ptz.PTZNode{
		{
			Name: "PTZ Node 1",
			Token: "ptz_node1",
			PTZType: "continuous",
			MinPan: -180,
			MaxPan: 180,
			MinTilt: -90,
			MaxTilt: 90,
			MinZoom: 0,
			MaxZoom: 100,
		},
	}
	
	ptzService := &ptz.ServiceContext{
		Port: 8080,
		PTZNodes: ptzNodes,
	}
	
	// Test GetNodes method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	ptzService.GetNodes()
}

func TestEventsService(t *testing.T) {
	// Create an events service context with an event
	eventList := []events.Event{
		{
			Topic: "tns1:VideoSource/MotionAlarm",
			Producer: "MotionDetector",
		},
	}
	
	eventsService := &events.ServiceContext{
		Port: 8080,
		Events: eventList,
	}
	
	// Test GetEventProperties method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	eventsService.GetEventProperties()
}

func TestDeviceIOService(t *testing.T) {
	// Create a deviceio service context with a relay output
	relayOutputs := []deviceio.RelayOutput{
		{
			Name: "Relay Output 1",
			Token: "relay1",
			IdleState: "open",
		},
	}
	
	deviceioService := &deviceio.ServiceContext{
		Port: 8080,
		RelayOutputs: relayOutputs,
	}
	
	// Test GetRelayOutputs method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	deviceioService.GetRelayOutputs()
}
