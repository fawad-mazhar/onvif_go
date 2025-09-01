# ONVIF Go Server Makefile

# Variables
MODULE := github.com/fawad-mazhar/onvif-go
BINARY_NAME := onvif_simple_server
WSD_BINARY_NAME := wsd_simple_server
NOTIFY_BINARY_NAME := onvif_notify_server
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null || echo "0.0.0")
BUILD_DIR := bin
MAIN_DIR := cmd/onvif-server

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build all binaries
build: onvif-server wsd-server notify-server

# Build main ONVIF server
onvif-server:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_DIR)/main.go

# Build WSD server
wsd-server:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(WSD_BINARY_NAME) ./internal/server/wsd_server.go

# Build notification server
notify-server:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(NOTIFY_BINARY_NAME) ./internal/server/notification_server.go

# Install binaries to system
install: build
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	install -m 755 $(BUILD_DIR)/$(WSD_BINARY_NAME) /usr/local/bin/
	install -m 755 $(BUILD_DIR)/$(NOTIFY_BINARY_NAME) /usr/local/bin/

# Run tests
test:
	go test ./test/...

# Run tests with coverage
test-coverage:
	go test -cover ./test/...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build with build directory dependency
build: $(BUILD_DIR)

# Help target
help:
	@echo "ONVIF Go Server Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all             - Build all binaries (default target)"
	@echo "  build           - Build all binaries"
	@echo "  onvif-server    - Build main ONVIF server"
	@echo "  wsd-server      - Build WSD discovery server"
	@echo "  notify-server   - Build event notification server"
	@echo "  install         - Install binaries to /usr/local/bin"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Remove build artifacts"
	@echo "  help            - Display this help message"

# Phony targets
.PHONY: all build onvif-server wsd-server notify-server install test test-coverage clean help
