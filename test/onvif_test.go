package tests

import (
	"testing"

	"github.com/fawad-mazhar/onvif-go/pkg/services/device"
	"github.com/fawad-mazhar/onvif-go/pkg/services/media"
)

// TestONVIFGetDeviceInformation tests the GetDeviceInformation method directly
func TestONVIFGetDeviceInformation(t *testing.T) {
	// Create a device service context
	deviceService := &device.ServiceContext{
		Manufacturer: "TestManufacturer",
		Model: "TestModel",
		FirmwareVer: "1.0.0",
		SerialNum: "1234567890",
		HardwareId: "HW123",
		Port: 8080,
	}
	
	// Test GetDeviceInformation method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	deviceService.GetDeviceInformation()
}

// TestONVIFGetProfiles tests the GetProfiles method directly
func TestONVIFGetProfiles(t *testing.T) {
	// Create a media service context
	mediaService := &media.ServiceContext{
		Port: 8080,
		Profiles: []media.Profile{
			{
				Name: "Profile1",
				Width: 1920,
				Height: 1080,
				URL: "rtsp://localhost:554/stream1",
				SnapURL: "http://localhost:8080/snapshot1.jpg",
				Type: "H264",
			},
		},
	}
	
	// Test GetProfiles method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	mediaService.GetProfiles()
}

// TestONVIFGetServices tests the GetServices method directly
func TestONVIFGetServices(t *testing.T) {
	// Create a device service context
	deviceService := &device.ServiceContext{
		Port: 8080,
	}
	
	// Test GetServices method
	// Note: In a real test, we would capture the output and verify it
	// For now, we're just ensuring the method doesn't panic
	deviceService.GetServices()
}
