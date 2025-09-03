package utils

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

// ProcessServiceTemplate processes an ONVIF service template with fallback to generic template
func ProcessServiceTemplate(w http.ResponseWriter, serviceName, methodName string, replacements map[string]string) error {
	// Build template path
	templatePath := filepath.Join("service_files", serviceName, methodName+".xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}

	// Process template
	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process %s template: %v", methodName, err)
	}

	// Write response
	w.Write([]byte(response))
	return nil
}

// ReplaceTemplatePlaceholders replaces multiple placeholders in a template string
func ReplaceTemplatePlaceholders(template string, replacements map[string]string) string {
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// GetCurrentTimeRFC3339 returns current time in RFC3339 format
func GetCurrentTimeRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

// GetTimeAfterHoursRFC3339 returns time after specified hours in RFC3339 format
func GetTimeAfterHoursRFC3339(hours int) string {
	return time.Now().Add(time.Duration(hours) * time.Hour).Format(time.RFC3339)
}

// ReadSOAPRequest reads and validates SOAP request body
func ReadSOAPRequest(r *http.Request, requestType string) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read %s request: %v", requestType, err)
	}
	return string(body), nil
}

// SetSOAPHeaders sets common SOAP response headers
func SetSOAPHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/soap+xml")
}

// CreateVideoEncoderConfig creates video encoder configuration XML for different types
func CreateVideoEncoderConfig(encodingType, token string, width, height int) string {
	// Common parts
	commonConfig := fmt.Sprintf(`
                    <trt:VideoEncoderConfiguration token="%s_veconf">
                        <tt:Name>VideoEncoderConfiguration</tt:Name>
                        <tt:UseCount>1</tt:UseCount>
                        <tt:Encoding>%s</tt:Encoding>
                        <tt:Resolution>
                            <tt:Width>%d</tt:Width>
                            <tt:Height>%d</tt:Height>
                        </tt:Resolution>
                        <tt:Quality>5.0</tt:Quality>
                        <tt:RateControl>
                            <tt:FrameRateLimit>25</tt:FrameRateLimit>
                            <tt:EncodingInterval>1</tt:EncodingInterval>
                            <tt:BitrateLimit>10240</tt:BitrateLimit>
                        </tt:RateControl>`, token, encodingType, width, height)

	// Type-specific configuration
	var typeSpecificConfig string
	switch encodingType {
	case "JPEG":
		typeSpecificConfig = ""
	case "MPEG4":
		typeSpecificConfig = `
                        <tt:MPEG4>
                            <tt:GovLength>60</tt:GovLength>
                            <tt:Mpeg4Profile>Main</tt:Mpeg4Profile>
                        </tt:MPEG4>`
	case "H264":
		typeSpecificConfig = `
                        <tt:H264>
                            <tt:GovLength>60</tt:GovLength>
                            <tt:H264Profile>Main</tt:H264Profile>
                        </tt:H264>`
	}

	// Common ending
	commonEnding := `
                        <tt:Multicast>
                            <tt:Address>
                                <tt:Type>IPv4</tt:Type>
                            </tt:Address>
                            <tt:Port>0</tt:Port>
                            <tt:TTL>1</tt:TTL>
                            <tt:AutoStart>false</tt:AutoStart>
                        </tt:Multicast>
                        <tt:SessionTimeout>PT60S</tt:SessionTimeout>
                    </trt:VideoEncoderConfiguration>`

	return commonConfig + typeSpecificConfig + commonEnding
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

// ProcessCollectionServiceTemplate processes a service template with a collection of items
// T is the type of items in the collection
// items: the collection of items to process
// createElement: function that converts an item to XML element string
// placeholder: the template placeholder to replace (e.g., "%ITEMS%")
// serviceName: the service name for template lookup
// methodName: the method name for template lookup
func ProcessCollectionServiceTemplate[T any](
	w http.ResponseWriter,
	items []T,
	createElement func(T) string,
	placeholder string,
	serviceName string,
	methodName string,
) error {
	// Create elements from the collection
	elements := make([]string, len(items))
	for i, item := range items {
		elements[i] = createElement(item)
	}

	// Use first element or empty string if no items
	resultXML := ""
	if len(elements) > 0 {
		resultXML = elements[0] // For simplicity, using the first element
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		placeholder: resultXML,
	}

	// Process template and write response
	return ProcessServiceTemplate(w, serviceName, methodName, replacements)
}
