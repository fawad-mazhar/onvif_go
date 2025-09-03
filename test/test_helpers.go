package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

// TestCase represents a SOAP test case with request and metadata
type TestCase struct {
	Name    string
	URL     string
	Request string
}

// testServerStartupDelay waits for servers to start up before running tests
func testServerStartupDelay() {
	time.Sleep(2 * time.Second)
}

// sendSOAPRequest sends a SOAP request to the specified URL and returns the response
func sendSOAPRequest(t *testing.T, url, request, testName string) *http.Response {
	resp, err := http.Post(url, "application/soap+xml", bytes.NewBufferString(request))
	if err != nil {
		t.Logf("%s test failed (server may not be running): %v", testName, err)
		t.Skipf("Skipping %s test - server not running", testName)
		return nil
	}
	return resp
}

// validateSOAPResponse checks the response status code and logs the result
func validateSOAPResponse(t *testing.T, resp *http.Response, testName string) {
	defer resp.Body.Close()

	// Check that we got a response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Log test completion
	t.Logf("%s test completed with status code: %d", testName, resp.StatusCode)
}

// runSOAPTest executes a complete SOAP test case
func runSOAPTest(t *testing.T, testCase TestCase) {
	testServerStartupDelay()

	// Send the request and validate response
	resp := sendSOAPRequest(t, testCase.URL, testCase.Request, testCase.Name)
	if resp != nil {
		validateSOAPResponse(t, resp, testCase.Name)
	}
}
