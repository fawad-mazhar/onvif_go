# ONVIF Simple Server - Go Implementation Development Plan

## Overview
This document outlines a structured and incremental approach to rewrite the ONVIF simple server project from C to Go. The original C implementation is a lightweight ONVIF server intended for resource-constrained devices, supporting Profile S and Profile T.

## Architecture Overview
The Go implementation will follow a similar architecture but with Go idioms:
- Configuration parsing (supporting both .conf and .json formats)
- XML template processing system
- Service modules (Device, Media, PTZ, Events, DeviceIO)
- Authentication and security handling
- CGI-based server implementation

## Phase 1: Project Setup and Core Structures

### 1.1 Project Structure
- Create directory structure mirroring the C implementation
- Set up go.mod file with appropriate module name
- Create README.md with project description and usage instructions

### 1.2 Data Structures
- Define Go structs equivalent to C structs in `onvif_simple_server.h`
- Create `types.go` file with all data structures:
  - `UsernameToken` struct
  - `StreamProfile` struct
  - `RelayOutput` struct
  - `PTZNode` struct
  - `Event` struct
  - `ServiceContext` struct (main context struct)

### 1.3 Configuration Parsing
- Create `config.go` to handle .conf file parsing
- Implement functions to parse all configuration options
- Add support for environment variables override

## Phase 2: XML Processing System

### 2.1 XML Template Engine
- Create `xml_processor.go` to handle XML template processing
- Implement placeholder replacement functionality (equivalent to `cat()` function)
- Support for loading XML templates from files

### 2.2 XML Parser
- Create `xml_parser.go` for parsing incoming SOAP requests
- Implement method extraction functionality
- Add security header parsing

## Phase 3: Core Utilities and Helpers

### 3.1 Logging System
- Create `logger.go` with logging functionality equivalent to log.c
- Implement different log levels (trace, debug, info, warn, error, fatal)
- Add log rotation feature

### 3.2 System Utilities
- Create `utils.go` with utility functions:
  - IP address retrieval
  - String trimming
  - Base64 encoding/decoding
  - SHA1 hashing
  - System command execution

## Phase 4: Service Implementations

### 4.1 Device Service
- Create `device_service.go` implementing all Device methods:
  - GetServices
  - GetServiceCapabilities
  - GetDeviceInformation
  - GetSystemDateAndTime
  - SystemReboot
  - GetScopes
  - GetUsers
  - GetWsdlUrl
  - GetCapabilities
  - GetNetworkInterfaces
  - GetDiscoveryMode

### 4.2 Media Service
- Create `media_service.go` implementing all Media methods:
  - GetServiceCapabilities
  - GetVideoSources
  - GetVideoSourceConfigurations
  - GetProfiles
  - GetStreamUri
  - GetSnapshotUri
  - etc.

### 4.3 Media2 Service
- Create `media2_service.go` implementing all Media2 methods
- Conditional compilation based on configuration

### 4.4 PTZ Service
- Create `ptz_service.go` implementing all PTZ methods:
  - GetServiceCapabilities
  - GetConfigurations
  - ContinuousMove
  - AbsoluteMove
  - RelativeMove
  - Stop
  - GetStatus
  - SetPreset
  - RemovePreset
  - GotoPreset
  - GotoHomePosition

### 4.5 Events Service
- Create `events_service.go` implementing all Events methods:
  - GetServiceCapabilities
  - CreatePullPointSubscription
  - PullMessages
  - Subscribe
  - Renew
  - Unsubscribe
  - GetEventProperties
  - SetSynchronizationPoint

### 4.6 DeviceIO Service
- Create `deviceio_service.go` implementing all DeviceIO methods:
  - GetVideoSources
  - GetServiceCapabilities
  - GetAudioOutputs
  - GetAudioSources
  - GetRelayOutputs
  - SetRelayOutputSettings
  - SetRelayOutputState

## Phase 5: Security Implementation

### 5.1 Authentication
- Create `auth.go` to handle WS-UsernameToken authentication
- Implement nonce decoding
- Add SHA1 hashing for password digest calculation
- Create authentication error handling

## Phase 6: Main Server Implementation

### 6.1 CGI Server
- Create `onvif_simple_server.go` as the main entry point
- Implement CGI request handling
- Add command-line argument parsing
- Integrate all service modules

### 6.2 WSD Server
- Create `wsd_simple_server.go` for Web Service Discovery
- Implement discovery protocol handling

### 6.3 Notification Server
- Create `onvif_notify_server.go` for event notifications
- Implement file monitoring for events
- Add subscriber management

## Phase 7: Testing and Validation

### 7.1 Unit Tests
- Create unit tests for each module
- Test configuration parsing
- Test XML processing functions
- Test authentication mechanisms

### 7.2 Integration Tests
- Create integration tests for service methods
- Test with sample XML requests
- Validate responses match expected format

## Phase 8: Documentation and Examples

### 8.1 Configuration Examples
- Convert .conf.example to Go equivalent
- Convert .json.example to Go equivalent
- Document all configuration options

### 8.2 Build Instructions
- Create Makefile equivalent in Go
- Document build process
- Add cross-compilation support

### 8.3 Usage Documentation
- Document command-line options
- Provide usage examples
- Explain deployment with different HTTP servers

## Phase 9: Optimization and Refinement

### 9.1 Performance Optimization
- Profile and optimize critical paths
- Implement efficient memory management
- Optimize XML processing

### 9.2 Code Quality
- Add proper error handling
- Implement logging throughout
- Add comments and documentation
- Follow Go best practices and idioms

## Phase 10: Compatibility Testing

### 10.1 Client Compatibility
- Test with ONVIF Device Manager
- Test with Synology Surveillance Station
- Test with other common ONVIF clients

### 10.2 Platform Compatibility
- Test on different Linux distributions
- Verify resource usage on constrained devices
- Test with different HTTP servers (lighttpd, busybox httpd)

## Implementation Timeline

### Week 1: Phases 1-2
- Project setup and core structures
- XML processing system

### Week 2: Phases 3-4
- Utilities and helpers
- Device and Media service implementations

### Week 3: Phases 4-5
- PTZ, Events, and DeviceIO services
- Security implementation

### Week 4: Phase 6
- Main server implementation
- WSD and notification servers

### Week 5: Phases 7-8
- Testing framework
- Documentation and examples

### Week 6: Phases 9-10
- Optimization
- Compatibility testing

## Key Considerations

1. **Resource Constraints**: The Go implementation should maintain the lightweight nature of the original C version
2. **CGI Compatibility**: Must work with existing HTTP servers that support CGI
3. **Configuration Compatibility**: Should support the same configuration file format as the C version
4. **XML Template System**: Maintain the same template-based approach for SOAP responses
5. **Authentication**: Support the same WS-UsernameToken authentication mechanism
6. **Cross-Platform**: Should work on Linux-based systems targeting embedded devices
