# ONVIF Simple Server — Go Implementation

A lightweight ONVIF server (Profile S) written in Go. It runs as a single self-contained binary — the ONVIF HTTP server, WS-Discovery daemon, and event notification server all start in one process with no external dependencies.

## Table of Contents
- [Features](#features)
- [Architecture](#architecture)
- [Building](#building)
- [Running](#running)
- [Configuration](#configuration)
- [Testing](#testing)
- [Implemented ONVIF Operations](#implemented-onvif-operations)
- [Compatibility](#compatibility)
- [License](#license)

## Features

- **Single binary** — ONVIF HTTP server + WSD discovery + event notification server all in one process
- **Standalone HTTP server** — no CGI or external `httpd` required
- **Device, Media (Profile S), PTZ, Events, DeviceIO** service implementations
- **WS-UsernameToken** authentication (SHA-1 nonce, configurable user/password)
- **Dual config format** — flat `.conf` key-value (identical to C reference) or `.json`
- **XML template-based** response system byte-compatible with the C reference output
- **Golden-diff test suite** — all responses verified against the C reference XML

## Architecture

```
onvif-go/
├── cmd/onvif-server/       # main() — starts all three servers in one binary
├── internal/
│   ├── auth/               # WS-UsernameToken SHA-1 authentication
│   ├── config/             # .conf / .json config loading
│   ├── exec/               # PTZ shell-command execution
│   ├── logger/             # levelled logging
│   ├── server/             # HTTP, WSD, notification server wiring
│   └── xml/                # XML template engine + SOAP fault renderer
├── pkg/services/
│   ├── device/             # Device service methods
│   ├── deviceio/           # DeviceIO service methods
│   ├── events/             # Events service methods
│   ├── media/              # Media service methods
│   └── ptz/                # PTZ service methods
├── service_files/          # XML response templates
│   ├── generic/            #   shared (Fault.xml, Empty.xml, …)
│   └── <service>/          #   per-service subdirectories
└── test/                   # Golden-diff integration tests
```

## Building

```bash
make build          # produces bin/onvif_server
```

Or directly:

```bash
go build -o bin/onvif_server ./cmd/onvif-server
```

## Running

```bash
./bin/onvif_server [flags]
```

| Flag | Default | Description |
|---|---|---|
| `-config` | `internal/config/onvif_simple_server.conf` | Path to config file (`.conf` or `.json`) |
| `-log-level` | `info` | Log verbosity: `trace`, `debug`, `info`, `warn`, `error`, `fatal` |
| `-template-dir` | `service_files` | Path to XML template directory |

### Quick start example

```bash
# Build
make build

# Run with the bundled test config
./bin/onvif_server -config test/fixtures/config/server.conf -log-level debug
```

The server will listen on the port defined in the config file (default `8080`).  
Test with any ONVIF client at `http://<host>:<port>/onvif/device_service`.

### Graceful shutdown

The process responds to `SIGINT` / `SIGTERM`: the WSD server sends a `Bye` multicast before exiting.

## Configuration

The config file uses the same flat `key=value` format as the C reference.

```ini
model=MyCamera
manufacturer=Acme
firmware_ver=1.0.0
hardware_id=CAM001
serial_num=SN0001234
ifs=eth0
port=8080
uuid=550e8400-e29b-41d4-a716-446655440000
scope=onvif://www.onvif.org/Profile/Streaming
scope=onvif://www.onvif.org/Profile/S

# Authentication (leave blank to disable)
user=admin
password=secret

# Advanced options
adv_fault_if_unknown=0
adv_fault_if_set=0

# Profile 0 — main stream
name=Profile_0
width=1920
height=1080
url=rtsp://%s/stream0
snapurl=http://%s/snapshot.jpg
audio_encoder=AAC

# Profile 1 — sub stream
name=Profile_1
width=640
height=360
url=rtsp://%s/stream1
snapurl=http://%s/snapshot_low.jpg
audio_encoder=AudioNone

# PTZ
ptz=1
min_step_x=-1.0
max_step_x=1.0
min_step_y=-1.0
max_step_y=1.0
min_step_z=0.0
max_step_z=0.0
move_left=/usr/bin/ptz_ctrl left %f
move_right=/usr/bin/ptz_ctrl right %f
move_up=/usr/bin/ptz_ctrl up %f
move_down=/usr/bin/ptz_ctrl down %f
move_stop=/usr/bin/ptz_ctrl stop %s

# Relay outputs
idle_state=open
close=/usr/bin/relay 0 close
open=/usr/bin/relay 0 open

# Events (1=PullPoint, 2=BaseSubscription, 3=both)
events=1
topic=tns1:VideoSource/MotionAlarm
source_name=Source
source_type=tt:ReferenceToken
source_value=VideoSourceToken
input_file=/tmp/onvif_notify_server/motion_alarm
```

**JSON format** is also accepted — pass a `.json` file to `-config`.  
See `internal/config/README.md` for the full field reference.

| Key | Description |
|---|---|
| `port` | HTTP listen port |
| `ifs` | Network interface name for service URL generation and WSD (`lo`, `eth0`, …) |
| `uuid` | Device UUID embedded in WSD Hello/Bye and service URNs |
| `user` / `password` | WS-UsernameToken credentials (empty = auth disabled) |
| `adv_fault_if_unknown` | Send SOAP fault for unsupported actions instead of empty 200 |
| `adv_fault_if_set` | Send SOAP fault for unsupported Set* actions |
| `url` | RTSP stream URL (`%s` is replaced with the device IP at runtime) |
| `snapurl` | Snapshot HTTP URL |
| `ptz` | `1` to enable PTZ service |
| `min_step_*` / `max_step_*` | PTZ pan/tilt/zoom range |
| `move_*` | Shell commands for PTZ movement (`%f` = speed, `%s` = direction) |
| `idle_state` | Relay idle state: `open` or `close` |
| `events` | Event delivery mode: `1`=PullPoint, `2`=BaseSubscription, `3`=both |
| `input_file` | File whose creation/deletion triggers event notifications |

## Testing

```bash
make test                    # all tests
make test-coverage           # with coverage report
go test -v ./test/...        # verbose golden-diff + integration tests
go test ./...                # all packages
```

The `test/` suite runs **golden-diff tests** that compare every ONVIF response byte-for-byte against captured C reference XML. All 21 operation fixtures must pass before merging.

## Implemented ONVIF Operations

### Device
```
GetCapabilities          GetDeviceInformation     GetDiscoveryMode
GetNetworkInterfaces     GetScopes                GetServices
GetServiceCapabilities   GetSystemDateAndTime     GetUsers
GetWsdlUrl               SystemReboot
```

### Media (Profile S)
```
GetServiceCapabilities   GetProfiles              GetProfile
GetStreamUri             GetSnapshotUri           CreateProfile (→ fault)
```

### PTZ
```
GetServiceCapabilities   GetNodes                 GetNode
GetConfigurations        GetConfiguration         GetConfigurationOptions
GetPresets               GotoPreset               GotoHomePosition
ContinuousMove           RelativeMove             AbsoluteMove
Stop                     GetStatus                SetPreset
RemovePreset             SetHomePosition
```

### Events
```
GetServiceCapabilities         CreatePullPointSubscription
PullMessages                   Subscribe
Renew                          Unsubscribe
GetEventProperties             SetSynchronizationPoint
```

### DeviceIO
```
GetServiceCapabilities   GetRelayOutputs
```

All other actions return an appropriate SOAP fault (or an empty 200 response, depending on `adv_fault_if_unknown`).

## Compatibility

Tested with the same ONVIF clients as the C reference:
- Onvif Device Manager (Windows)
- Synology Surveillance Station (DSM 6.x and 7.x)
- Onvier (Android)
- Frigate
- Unifi Protect

## License

[GPLv3](https://choosealicense.com/licenses/gpl-3.0/)
