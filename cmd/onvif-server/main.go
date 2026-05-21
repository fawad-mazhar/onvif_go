// Package main provides the ONVIF server application entry point.
package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/server"
	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

func main() {
	configPath := flag.String("config", "internal/config/onvif_simple_server.conf", "path to config file (.conf or .json)")
	logLevelStr := flag.String("log-level", "info", "log level: trace, debug, info, warn, error, fatal")
	templateDir := flag.String("template-dir", xml.ServiceTemplateDir, "path to service template directory")
	flag.Parse()

	// Initialize logger at requested level
	level, err := logger.ParseLevel(*logLevelStr)
	if err != nil {
		logger.InitLogger(logger.INFO)
		logger.Warnf("Unknown log level %q, defaulting to info", *logLevelStr)
	} else {
		logger.InitLogger(level)
	}

	// Override template directory if provided
	xml.ServiceTemplateDir = *templateDir

	// Load configuration (extension-agnostic: .json or flat key-value)
	cfg, err := config.Load(*configPath)
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

	// Single signal handler for the whole process. Cancelling ctx propagates
	// to StartWSDServer which sends Bye and returns.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start WSD server
	go func() {
		if err := server.StartWSDServer(ctx, cfg); err != nil {
			logger.Fatalf("WSD server error: %v", err)
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	logger.Infof("Shutting down servers...")
}
