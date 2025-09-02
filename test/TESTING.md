# ONVIF Go Server Testing Guide

This guide provides comprehensive testing scenarios for the ONVIF Go server implementation.

## 🔧 Setup & Configuration

### Initial Setup

First, create a proper configuration file:
```bash
# Create a test configuration
cp internal/config/onvif_simple_server.conf.example internal/config/onvif_simple_server.conf
```

Edit the config with your network settings:
```ini
port=8080
user=admin
password=admin123
manufacturer=YourCompany
model=TestCamera
interface=eth0
uuid=12345678-1234-1234-1234-123456789012
```

### Build the Server

```bash
# Build using Makefile
make build

# Or build manually
go build -o bin/onvif_server cmd/onvif-server/main.go
```

## 🚀 Deployment Options

### 1. Standalone HTTP Server

```bash
# Run the server directly
./bin/onvif_server

# Or with custom config and debug logging
./bin/onvif_server -c /path/to/your.conf -d 3
```

### 2. CGI with Apache/Nginx

```bash
# Install in CGI directory
sudo cp bin/onvif_server /usr/lib/cgi-bin/
sudo chmod +x /usr/lib/cgi-bin/onvif_server

# Configure web server to handle ONVIF requests
```

### 3. Docker Container

Create a `Dockerfile`:
```dockerfile
FROM golang:alpine AS builder
COPY . /app
WORKDIR /app
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/onvif_server .
COPY --from=builder /app/service_files ./service_files
COPY --from=builder /app/internal/config ./internal/config
EXPOSE 8080 3702
CMD ["./onvif_server"]
```

Build and run:
```bash
docker build -t onvif-go-server .
docker run -p 8080:8080 -p 3702:3702/udp onvif-go-server
```

## 🧪 Testing Tools

### 1. ONVIF Device Manager (Windows)

**Best for comprehensive testing**

1. Download from ONVIF official website
2. Launch ONVIF Device Manager
3. Add device with your server's IP:8080
4. Test discovery, authentication, and service calls
5. Verify all ONVIF services are detected correctly

### 2. Command Line Testing

#### Test WS-Discovery
```bash
# Create probe request XML
cat > probe_request.xml << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" 
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery">
  <soap:Header>
    <wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
    <wsa:MessageID>uuid:12345678-1234-1234-1234-123456789012</wsa:MessageID>
    <wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
  </soap:Header>
  <soap:Body>
    <wsd:Probe/>
  </soap:Body>
</soap:Envelope>
EOF

# Send probe request
curl -X POST http://your-ip:3702 \
  -H "Content-Type: application/soap+xml" \
  -d @probe_request.xml
```

#### Test Device Service
```bash
# Create device info request
cat > get_device_info.xml << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
  <soap:Header>
    <wsa:Action>http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation</wsa:Action>
  </soap:Header>
  <soap:Body>
    <tds:GetDeviceInformation/>
  </soap:Body>
</soap:Envelope>
EOF

# Test device service
curl -X POST http://your-ip:8080/onvif/device_service \
  -H "Content-Type: application/soap+xml" \
  -d @get_device_info.xml
```

### 3. Python ONVIF Library Testing

Install and test with Python:
```bash
pip install onvif-zeep
```

```python
# test_onvif_client.py
from onvif import ONVIFCamera

try:
    # Connect to your ONVIF server
    cam = ONVIFCamera('your-ip', 8080, 'admin', 'admin123')
    
    # Test device information
    device_info = cam.devicemgmt.GetDeviceInformation()
    print("Device Info:", device_info)
    
    # Test capabilities
    capabilities = cam.devicemgmt.GetCapabilities()
    print("Capabilities:", capabilities)
    
    # Test media profiles
    media_service = cam.create_media_service()
    profiles = media_service.GetProfiles()
    print("Profiles:", profiles)
    
    print("✅ All tests passed!")
    
except Exception as e:
    print(f"❌ Test failed: {e}")
```

Run the test:
```bash
python test_onvif_client.py
```

### 4. Mobile Apps Testing

Test with these mobile applications:

- **Onvier** (Android/iOS) - Free ONVIF client
- **ONVIF Viewer** - Basic ONVIF testing
- **IP Cam Viewer** - Multi-protocol support

## 📱 Integration Testing

### Video Management Systems

#### Frigate
```yaml
# Add to Frigate config
cameras:
  onvif_test:
    ffmpeg:
      inputs:
        - path: rtsp://admin:admin123@your-ip:8080/stream1
          roles:
            - detect
            - record
    onvif:
      host: your-ip
      port: 8080
      user: admin
      password: admin123
```

#### Synology Surveillance Station
1. Add camera via IP Camera wizard
2. Select ONVIF protocol
3. Enter your server IP and credentials
4. Test connection and verify streams

#### UniFi Protect
1. Add generic ONVIF camera
2. Configure with server IP:8080
3. Test device detection and stream access

### Network Testing

```bash
# Check service discovery
nmap -sU -p 3702 your-network/24

# Verify ONVIF ports are open
nmap -p 8080,3702 your-ip

# Test multicast discovery (requires root)
sudo tcpdump -i any -n udp port 3702
```

## 🔍 Debugging & Monitoring

### Enable Debug Logging
```bash
# Maximum debug level
./bin/onvif_server -d 5

# Monitor logs in real-time
tail -f /var/log/onvif_server.log
```

### Packet Capture Analysis
```bash
# Capture ONVIF traffic
sudo tcpdump -i eth0 -w onvif_capture.pcap port 8080 or port 3702

# Analyze with Wireshark
wireshark onvif_capture.pcap

# Or use tshark for command line analysis
tshark -r onvif_capture.pcap -Y "http or soap"
```

### Test Individual Services

**Using the Test Script (Recommended)**
```bash
# Run the comprehensive test script
cd test/
./test_requests.sh
```

**Manual SOAP Requests**
```bash
# Test Device Service - GetDeviceInformation
curl -X POST localhost:8080/onvif/device_service \
  -H "Content-Type: application/soap+xml; charset=utf-8" \
  -H "SOAPAction: http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation" \
  -d @test/soap_requests/device_get_info.xml

# Test Device Service - GetCapabilities
curl -X POST localhost:8080/onvif/device_service \
  -H "Content-Type: application/soap+xml; charset=utf-8" \
  -H "SOAPAction: http://www.onvif.org/ver10/device/wsdl/GetCapabilities" \
  -d @test/soap_requests/device_get_capabilities.xml

# Test Media Service - GetProfiles  
curl -X POST localhost:8080/onvif/media_service \
  -H "Content-Type: application/soap+xml; charset=utf-8" \
  -H "SOAPAction: http://www.onvif.org/ver10/media/wsdl/GetProfiles" \
  -d @test/soap_requests/media_get_profiles.xml
```

**Testing CGI Mode**
```bash
# Test as CGI subprocess with proper environment variables
echo '<?xml version="1.0"?><soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"><soap:Body><GetDeviceInformation/></soap:Body></soap:Envelope>' | \
CONTENT_LENGTH=150 REQUEST_METHOD=POST ./bin/onvif_server device_service
```

## 📋 Real-World Test Checklist

### Basic Functionality
- [ ] Server starts without errors
- [ ] Configuration file loads correctly
- [ ] All required ports are bound (8080, 3702)
- [ ] Service endpoints respond to HTTP requests

### ONVIF Discovery
- [ ] Device responds to WS-Discovery probe requests
- [ ] Device appears in ONVIF Device Manager
- [ ] Proper device information is returned
- [ ] UUID and scopes are correctly configured

### Authentication
- [ ] Username/password authentication works
- [ ] WS-Security digest authentication functions
- [ ] Invalid credentials are properly rejected
- [ ] Nonce timestamp validation prevents replay attacks

### Device Service
- [ ] GetDeviceInformation returns correct data
- [ ] GetCapabilities shows all enabled services
- [ ] GetServices lists available ONVIF services
- [ ] GetSystemDateAndTime returns current time
- [ ] GetScopes returns configured scopes

### Media Service
- [ ] GetProfiles lists configured media profiles
- [ ] Profile configurations match config file
- [ ] Stream URLs are properly generated
- [ ] Snapshot URLs work correctly

### PTZ Service (if enabled)
- [ ] GetNodes returns PTZ node information
- [ ] PTZ commands execute without errors
- [ ] Movement boundaries are respected
- [ ] Preset operations function correctly

### Events Service (if enabled)
- [ ] GetEventProperties returns event information
- [ ] Event subscriptions work properly
- [ ] Notifications are delivered correctly
- [ ] Pull-point mechanism functions

### Network & Performance
- [ ] Multiple simultaneous client connections
- [ ] Server handles network interface changes
- [ ] Memory usage remains stable under load
- [ ] No resource leaks during extended operation

## ⚠️ Common Issues & Troubleshooting

### Network Issues
```bash
# Check if ports are open
netstat -tulpn | grep -E ':(8080|3702)'

# Verify firewall settings
sudo ufw status
sudo firewall-cmd --list-ports

# Test connectivity
telnet your-ip 8080
```

### Configuration Issues
- **XML Templates**: Ensure all required template files exist in `service_files/`
- **Interface Binding**: Verify network interface name in config
- **IP Addresses**: Replace any hardcoded IPs with dynamic detection
- **File Permissions**: Check that template files are readable

### Authentication Problems
- **Password Encoding**: Verify digest calculation matches ONVIF spec
- **Timestamp Validation**: Check system clock synchronization
- **Nonce Handling**: Ensure proper base64 encoding/decoding

### Service Discovery Issues
- **Multicast**: Verify multicast routing is enabled
- **UUID Format**: Ensure UUID follows proper format
- **Network Scope**: Check if discovery works across subnets

## 🎯 Performance Testing

### Load Testing
```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Test device service under load
ab -n 100 -c 10 -H "Content-Type: application/soap+xml" \
   -p get_device_info.xml \
   http://your-ip:8080/onvif/device_service
```

### Memory Profiling
```bash
# Build with profiling enabled
go build -o bin/onvif_server_debug cmd/onvif-server/main.go

# Run with memory profiling
./bin/onvif_server_debug &
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 📊 Monitoring & Metrics

### Health Check Endpoint
Consider adding a health check endpoint for monitoring:
```bash
curl http://your-ip:8080/health
```

### Log Analysis
```bash
# Monitor error rates
grep -c "ERROR" /var/log/onvif_server.log

# Check authentication failures
grep "Authentication failed" /var/log/onvif_server.log

# Monitor service usage
grep "SOAP action" /var/log/onvif_server.log | sort | uniq -c
```

## 🚀 Production Deployment

### Systemd Service
Create `/etc/systemd/system/onvif-server.service`:
```ini
[Unit]
Description=ONVIF Go Server
After=network.target

[Service]
Type=simple
User=onvif
ExecStart=/usr/local/bin/onvif_server -c /etc/onvif/onvif_simple_server.conf
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable onvif-server
sudo systemctl start onvif-server
sudo systemctl status onvif-server
```

### Security Considerations
- Run as non-root user
- Use strong authentication credentials
- Implement rate limiting for SOAP requests
- Regular security updates and monitoring
- Network segmentation for device isolation

---

This testing guide should help you thoroughly validate your ONVIF Go server implementation in various real-world scenarios. Start with basic functionality tests and gradually move to more complex integration scenarios.