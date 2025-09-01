package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Daemonize detaches the process from the controlling terminal
func Daemonize() error {
	// Fork the process
	pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
		Dir:   "",
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
		Sys:   &syscall.SysProcAttr{Setsid: true},
	})
	
	if err != nil {
		return fmt.Errorf("failed to fork process: %v", err)
	}
	
	// If we're in the parent process, exit
	if pid > 0 {
		os.Exit(0)
	}
	
	return nil
}

// GetFirstIP returns the first IP address of the specified network interface
func GetFirstIP(interfaceName string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}
	
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", fmt.Errorf("failed to get interface addresses: %v", err)
			}
			
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}
	
	return "", fmt.Errorf("interface %s not found or has no IP addresses", interfaceName)
}

// GetIPFromRoute returns the IP address from routing table for the specified destination
func GetIPFromRoute(destination string) (string, error) {
	cmd := exec.Command("ip", "route", "get", destination)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute ip route command: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "src") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "src" && i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("source IP not found in routing table")
}

// Trim removes leading and trailing whitespace from a string
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// SHA1Hash computes the SHA1 hash of the input string
func SHA1Hash(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Base64Decode decodes a base64 encoded string
func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// Base64Encode encodes a byte array to base64 string
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ExecuteCommand executes a system command
func ExecuteCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return string(bytes)
}

// GetHostname returns the system hostname
func GetHostname() (string, error) {
	return os.Hostname()
}

// GenerateGenericResponse generates a generic SOAP response
func GenerateGenericResponse() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
    <SOAP-ENV:Body>
        <GenericResponse/>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
}
