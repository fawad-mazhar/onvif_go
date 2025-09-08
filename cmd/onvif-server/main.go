// Package main provides the ONVIF server application entry point.
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/server"
)

func main() {
	// Initialize logger
	logger.InitLogger(logger.INFO)

	// Load configuration
	cfg, err := config.LoadConfig("internal/config/onvif_simple_server.conf")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Start integrated ONVIF HTTP server
	go func() {
		err := server.StartHTTPServer(cfg)
		if err != nil {
			logger.Fatalf("ONVIF HTTP server error: %v", err)
		}
	}()

	// Start Notification server
	go func() {
		err := server.StartNotificationServer(cfg)
		if err != nil {
			logger.Fatalf("Notification server error: %v", err)
		}
	}()

	// Start WSD server
	go func() {
		err := server.StartWSDServer(cfg)
		if err != nil {
			logger.Fatalf("WSD server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Infof("Shutting down servers...")
	time.Sleep(1 * time.Second)
}
