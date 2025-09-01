# ONVIF Server Test Scripts

This directory contains test scripts to verify the functionality of the ONVIF server components:

## Test Files

1. **auth_test.go** - Tests authentication functionality
2. **core_test.go** - Tests core configuration and service functionality
3. **onvif_test.go** - Tests ONVIF server functionality
4. **notification_test.go** - Tests notification server functionality
5. **wsd_test.go** - Tests WSD (Web Services Dynamic Discovery) server functionality

## Running Tests

To run all tests:

```bash
go test -v ./test/...
```

To run tests for a specific component:

```bash
# Run only authentication tests
go test -v ./test/auth_test.go

# Run only ONVIF tests
go test -v ./test/onvif_test.go

# Run only notification tests
go test -v ./test/notification_test.go

# Run only WSD tests
go test -v ./test/wsd_test.go
```

## Test Requirements

- The ONVIF server must be running on port 8080
- The WSD server must be running on port 3702
- The notification server must be running on port 8081

Start the server before running integration tests:

```bash
./bin/onvif_server
```

## Test Results

The integration tests (ONVIF, notification, WSD) will be skipped if the respective servers are not running.
When servers are running, these tests will verify that the servers respond correctly to standard ONVIF/WSD requests.

## Test Coverage

The tests cover basic functionality verification:
- Authentication validation
- Configuration loading
- Device information retrieval
- Media profile listing
- PTZ node information
- Event properties
- WSD probe and resolve operations
- Notification subscription and message pulling
