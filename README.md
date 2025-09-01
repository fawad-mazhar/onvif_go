# ONVIF Simple Server - Go Implementation

This is a Go implementation of the ONVIF simple server, originally written in C. It is a lightweight ONVIF server intended for use in resource-constrained devices, supporting Profile S.

## Features

- Device service implementation
- Media service implementation (Profile S)
- PTZ service implementation
- Events service implementation
- DeviceIO service implementation
- WS-UsernameToken authentication support
- Configuration file parsing (.conf format)
- XML template-based response system
- CGI-compatible server interface

## Architecture

The server is organized into modules that mirror the ONVIF service structure:

```
onvif-go/
├── cmd/
│   └── onvif-server/      # Main application
├── internal/
│   ├── config/            # Configuration parsing
│   ├── logger/            # Logging system
│   ├── auth/             # Authentication system
│   ├── utils/            # Utility functions
│   ├── xml/              # XML processing
│   ├── performance/      # Performance optimization
│   └── server/           # Server implementations
└── pkg/
    └── services/
        ├── device/       # Device service methods
        ├── media/        # Media service methods
        ├── ptz/          # PTZ service methods
        ├── events/       # Events service methods
        └── deviceio/     # DeviceIO service methods
```

## Building

Using the provided Makefile:

```bash
make
```

Or manually:

```bash
go build -o onvif_simple_server ./cmd/onvif-server
go build -o wsd_simple_server ./cmd/onvif-server/server/wsd_server.go
go build -o onvif_notify_server ./cmd/onvif-server/server/notification_server.go
```

## Usage

### ONVIF Simple Server

The main ONVIF server runs as a CGI application and needs an HTTP server that supports the CGI standard.

```bash
onvif_simple_server [-c CONF_FILE] [-d LEVEL] [-f]
```

Options:
- `-c CONF_FILE, --conf_file CONF_FILE`: Path of the configuration file
- `-d LEVEL, --debug LEVEL`: Enable debug with LEVEL = 0..5 (default 0 = fatal errors)
- `-f, --conf_help`: Print the help for the configuration file
- `-h, --help`: Print help information

### WSD Simple Server

Web Service Discovery daemon.

```bash
wsd_simple_server -i INTERFACE -x XADDR [-m MODEL] [-n MANUFACTURER] -p PID_FILE [-f] [-d LEVEL]
```

### ONVIF Notify Server

Event notification server that monitors files and sends notify messages to subscribers.

```bash
onvif_notify_server [-p PID_FILE] [-q NUM] [-f] [-d LEVEL]
```

## Configuration

The server supports the same configuration file format as the original C implementation. See `onvif_simple_server.conf.example` for a detailed example.

## Compatibility

Tested with:
- Onvif Device Manager (Windows)
- Synology Surveillance Station (DSM 6.x and 7.x)
- Onvier (Android)
- Frigate
- Unifi Protect

## License

[GPLv3](https://choosealicense.com/licenses/gpl-3.0/)
