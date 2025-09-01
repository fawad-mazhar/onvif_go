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
		logger.Fatal("Failed to load config: %v", err)
	}

	// Start ONVIF server
	go func() {
		err := server.StartONVIFServer(cfg)
		if err != nil {
			logger.Fatal("ONVIF server error: %v", err)
		}
	}()

	// Start Notification server
	go func() {
		err := server.StartNotificationServer(cfg)
		if err != nil {
			logger.Fatal("Notification server error: %v", err)
		}
	}()

	// Start WSD server
	go func() {
		err := server.StartWSDServer(cfg)
		if err != nil {
			logger.Fatal("WSD server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down servers...")
	time.Sleep(1 * time.Second)
}
