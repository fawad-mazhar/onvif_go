# ONVIF Server Test Scripts

This directory contains test scripts to verify the functionality of the ONVIF server components:

## Test Files

1. **auth_test.go** - WS-UsernameToken authentication (digest, wrong user, wrong password, plain-text rejection)
2. **core_test.go** - Configuration loading with field-by-field assertions
3. **golden_diff_test.go** - End-to-end golden-diff suite; every ONVIF response compared byte-for-byte against C reference fixtures

## Running Tests

To run all tests:

```bash
go test -v ./test/...
```

To run a specific file:

```bash
go test -v -run TestAuth ./test/...
go test -v -run GoldenDiff ./test/...
```

## Test Requirements

No external server needed — all tests use `httptest.NewServer` internally.

## Test Coverage

- Authentication (all four auth paths)
- Configuration loading (flat `.conf` format, all field types)
- Golden-diff for all 21 happy-path ONVIF responses (Device, Media, PTZ, Events, DeviceIO)

For the full list of covered operations see the `test/fixtures/` subdirectories.
