package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// parseSOAPAction extracts the SOAP action from the request (current implementation)
func parseSOAPActionCurrent(soapRequest string) string {
	// Look for the SOAP action in the request
	actionStart := strings.Index(soapRequest, "<soap:Body>")
	if actionStart == -1 {
		actionStart = strings.Index(soapRequest, "<SOAP-ENV:Body>")
		if actionStart == -1 {
			return ""
		}
	}
	
	// Extract the first tag after <soap:Body> or <SOAP-ENV:Body>
	actionStart += 10
	if strings.Contains(soapRequest[actionStart:], "<soap:Body>") {
		actionStart += 12
	}
	
	actionEnd := strings.Index(soapRequest[actionStart:], ">")
	if actionEnd == -1 {
		return ""
	}
	
	// Extract the tag name
	tag := soapRequest[actionStart:actionStart+actionEnd+1]
	
	// Remove attributes if any
	if strings.Contains(tag, " ") {
		tag = tag[:strings.Index(tag, " ")]
	}
	
	// Remove opening bracket
	tag = strings.TrimPrefix(tag, "<")
	
	// Remove closing bracket or slash
	if strings.HasSuffix(tag, ">") {
		tag = tag[:len(tag)-1]
	} else if strings.HasSuffix(tag, "/") {
		tag = tag[:len(tag)-1]
	}
	
	return tag
}

// parseSOAPAction extracts the SOAP action from the request (fixed implementation)
func parseSOAPActionFixed(soapRequest string) string {
	// Look for the SOAP action in the request
	actionStart := strings.Index(soapRequest, "<soap:Body>")
	if actionStart == -1 {
		actionStart = strings.Index(soapRequest, "<SOAP-ENV:Body>")
		if actionStart == -1 {
			return ""
		}
		actionStart += len("<SOAP-ENV:Body>")
	} else {
		actionStart += len("<soap:Body>")
	}
	
	// Skip whitespace and newlines
	for actionStart < len(soapRequest) && (soapRequest[actionStart] == ' ' || 
		soapRequest[actionStart] == '\n' || soapRequest[actionStart] == '\r' || 
		soapRequest[actionStart] == '\t') {
		actionStart++
	}
	
	// Find the next opening tag
	if actionStart >= len(soapRequest) || soapRequest[actionStart] != '<' {
		return ""
	}
	
	actionEnd := strings.Index(soapRequest[actionStart:], ">")
	if actionEnd == -1 {
		return ""
	}
	
	// Extract the tag name
	tag := soapRequest[actionStart+1:actionStart+actionEnd]
	
	// Remove attributes if any
	if strings.Contains(tag, " ") {
		tag = tag[:strings.Index(tag, " ")]
	}
	
	// Remove namespace prefix for comparison
	if strings.Contains(tag, ":") {
		parts := strings.Split(tag, ":")
		if len(parts) > 1 {
			tag = parts[len(parts)-1]
		}
	}
	
	// Handle self-closing tags
	if strings.HasSuffix(tag, "/") {
		tag = tag[:len(tag)-1]
	}
	
	return tag
}

func main() {
	// Read the SOAP request file
	content, err := ioutil.ReadFile("soap_requests/device_get_info.xml")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	soapRequest := string(content)
	
	fmt.Println("SOAP Request Content:")
	fmt.Println("====================")
	fmt.Println(soapRequest)
	fmt.Println()
	
	fmt.Println("Current Implementation Result:")
	fmt.Println("=============================")
	currentResult := parseSOAPActionCurrent(soapRequest)
	fmt.Printf("Parsed Action: '%s'\n", currentResult)
	fmt.Println()
	
	fmt.Println("Fixed Implementation Result:")
	fmt.Println("===========================")
	fixedResult := parseSOAPActionFixed(soapRequest)
	fmt.Printf("Parsed Action: '%s'\n", fixedResult)
	fmt.Println()
	
	// Debug the parsing steps
	fmt.Println("Debug Information:")
	fmt.Println("==================")
	
	actionStart := strings.Index(soapRequest, "<soap:Body>")
	fmt.Printf("soap:Body found at position: %d\n", actionStart)
	
	if actionStart != -1 {
		actionStart += len("<soap:Body>")
		fmt.Printf("Content after soap:Body: '%s...'\n", soapRequest[actionStart:actionStart+50])
		
		// Skip whitespace
		for actionStart < len(soapRequest) && (soapRequest[actionStart] == ' ' || 
			soapRequest[actionStart] == '\n' || soapRequest[actionStart] == '\r' || 
			soapRequest[actionStart] == '\t') {
			actionStart++
		}
		fmt.Printf("After skipping whitespace: '%s...'\n", soapRequest[actionStart:actionStart+30])
	}
}