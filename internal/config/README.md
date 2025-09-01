# Configuration Package

This package handles the loading and parsing of the ONVIF server configuration file.

## Configuration File

The main configuration file is `onvif_simple_server.conf`, which should be located in the config directory. This file contains all the settings needed for the ONVIF server to operate.

### File Structure

The configuration file uses a simple key=value format with the following sections:

1. **Server Settings**:
   - `port`: Main server port (default 8080)
   - `wsd_port`: WSD discovery server port (default 3702)
   - `notification_port`: Event notification server port (default 8081)

2. **Device Information**:
   - `manufacturer`: Device manufacturer name
   - `model`: Device model name
   - `firmware_version`: Firmware version string
   - `serial_number`: Device serial number
   - `hardware_id`: Device hardware ID

3. **Authentication**:
   - `username`: Authentication username (empty to disable auth)
   - `password`: Authentication password (empty to disable auth)
   - `uuid`: Device UUID (should be unique for each device)

4. **Media Profiles**:
   - Configured under `[profiles]` section with profile-specific settings

5. **PTZ Nodes**:
   - Configured under `[ptz_nodes]` section with PTZ-specific settings

## Usage

The configuration is loaded in the main application using:

```go
cfg, err := config.LoadConfig("onvif_simple_server.conf")
```

The `LoadConfig` function parses the configuration file and returns a `ServiceContext` struct containing all the configuration values that are used throughout the ONVIF services.
