# ONVIF Go Server Makefile

# Variables
MODULE := github.com/fawad-mazhar/onvif-go
BINARY_NAME := onvif_server
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null || echo "0.0.0")
BUILD_DIR := bin
MAIN_DIR := cmd/onvif-server

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
all: help

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build main ONVIF server
onvif-server: $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_DIR)/main.go

# Build all binaries
build: onvif-server

# Run tests
test:
	go test -v ./test/...

# Run tests with coverage
test-coverage:
	go test -cover ./test/...

# Help target
help:
	@echo "ONVIF Go Server Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build           - Build ONVIF server"
	@echo "  onvif-server    - Build main ONVIF server"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"

# Phony targets
.PHONY: all build onvif-server install test test-coverage clean help
