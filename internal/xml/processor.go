package xml

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ProcessTemplate reads an XML template file and replaces placeholders with provided values
func ProcessTemplate(filename string, replacements map[string]string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %v", err)
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Replace all placeholders in the line
		for placeholder, replacement := range replacements {
			line = strings.ReplaceAll(line, placeholder, replacement)
		}
		
		result.WriteString(line)
		result.WriteString("\n")
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading template file: %v", err)
	}
	
	return result.String(), nil
}

// WriteResponse writes the processed XML response to stdout with proper HTTP headers
func WriteResponse(xmlContent string, contentType string) {
	if contentType == "" {
		contentType = "application/soap+xml"
	}
	
	fmt.Printf("Content-Type: %s\n", contentType)
	fmt.Printf("Content-Length: %d\n", len(xmlContent))
	fmt.Printf("\n")
	fmt.Print(xmlContent)
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
