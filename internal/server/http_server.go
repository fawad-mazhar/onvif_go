package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
)

// StartHTTPServer starts an HTTP server that wraps the ONVIF server functionality
func StartHTTPServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)

	// Create HTTP server for ONVIF
	http.HandleFunc("/onvif/device_service", func(w http.ResponseWriter, r *http.Request) {
		// Handle ONVIF requests
		handleHTTPONVIFRequest(w, r, cfg)
	})

	// Start the HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Starting HTTP ONVIF server on port %d", cfg.Port)
	srv := &http.Server{Addr: addr}
	logger.Info("HTTP ONVIF server listening on address: %s", addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Fatal("Failed to start HTTP ONVIF server: %v", err)
	}
	return nil
}

// handleHTTPONVIFRequest handles HTTP ONVIF requests by wrapping the CGI-based ONVIF server
func handleHTTPONVIFRequest(w http.ResponseWriter, r *http.Request, cfg *config.ServiceContext) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Failed to read ONVIF request: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	// Create a command to run the ONVIF server
	cmd := exec.Command(os.Args[0])
	
	// Set environment variables that the ONVIF server expects
	cmd.Env = append(os.Environ(),
		"REQUEST_METHOD="+r.Method,
		"CONTENT_LENGTH="+fmt.Sprintf("%d", len(body)))
	
	// Set the request body as stdin
	cmd.Stdin = bytes.NewReader(body)
	
	// Capture stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	// Capture stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	// Run the command
	if err := cmd.Run(); err != nil {
		logger.Warn("ONVIF server execution failed: %v, stderr: %s", err, stderr.String())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Parse the output to separate headers from body
	output := stdout.Bytes()
	
	// Set default content type
	w.Header().Set("Content-Type", "application/soap+xml")
	
	// Check if output contains HTTP headers
	if bytes.Contains(output, []byte("Content-Type:")) {
		// Split headers and body
		if bytes.Contains(output, []byte("\r\n\r\n")) {
			parts := bytes.SplitN(output, []byte("\r\n\r\n"), 2)
			if len(parts) == 2 {
				// Parse headers
				headerLines := bytes.Split(parts[0], []byte("\r\n"))
				for _, line := range headerLines {
					if bytes.Contains(line, []byte(":")) {
						// Split on first colon
						headerParts := bytes.SplitN(line, []byte(":"), 2)
						key := string(bytes.TrimSpace(headerParts[0]))
						value := string(bytes.TrimSpace(headerParts[1]))
						w.Header().Set(key, value)
					}
				}
				// Write the body
				w.Write(parts[1])
				return
			}
		}
	}
	
	// Write the response directly if no headers found
	w.Write(output)
}
