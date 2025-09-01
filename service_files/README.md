# Service Files Directory

This directory contains XML template files used by the ONVIF Go implementation for various service responses. Each subdirectory corresponds to a specific ONVIF service:

## Directory Structure

- `device/` - XML templates for ONVIF device service methods
- `deviceio/` - XML templates for ONVIF device I/O service methods
- `events/` - XML templates for ONVIF events service methods
- `media/` - XML templates for ONVIF media service methods
- `ptz/` - XML templates for ONVIF PTZ (Pan-Tilt-Zoom) service methods

## How XML Templates Function

The XML files in these directories serve as response templates for ONVIF service method calls. When a service method is invoked, the corresponding XML template is processed by replacing placeholder variables with actual values from the service context.

### Placeholder Format

Placeholders in the XML templates follow the format `%PLACEHOLDER_NAME%`. For example:
- `%PROFILE_COUNT%` - Replaced with the actual number of profiles
- `%STREAM_URL%` - Replaced with the actual stream URL
- `%SNAPSHOT_URL%` - Replaced with the actual snapshot URL

### Template Processing

The template processing is handled by the XML utility functions in `internal/xml` package:
1. Load the appropriate template file based on the service method being called
2. Replace placeholders with actual values from the service context
3. Generate and return the final XML response

This approach allows for clean separation between the service logic and the response formatting, making it easier to maintain and modify the ONVIF service responses.
