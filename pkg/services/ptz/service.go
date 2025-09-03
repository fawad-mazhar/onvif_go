package ptz

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/fawad-mazhar/onvif-go/internal/xml"
)

// ServiceContext holds the configuration and state for the PTZ service
type ServiceContext struct {
	Port     int
	PTZNodes []PTZNode
}

// PTZNode represents a PTZ node configuration
type PTZNode struct {
	Name    string
	Token   string
	PTZType string // "relative", "absolute", or "continuous"
	MinPan  float64
	MaxPan  float64
	MinTilt float64
	MaxTilt float64
	MinZoom float64
	MaxZoom float64
}


// createNodeElement creates an XML element for a PTZ node
func (s *ServiceContext) createNodeElement(node PTZNode) string {
	// Create node element
	nodeElement := fmt.Sprintf(`
                <tptz:PTZNode token="%s">
                    <tt:Name>%s</tt:Name>
                    <tt:SupportedPTZSpaces>
                        <tt:AbsolutePanTiltPositionSpace>
                            <tt:URI>http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace</tt:URI>
                            <tt:XRange>
                                <tt:Min>%f</tt:Min>
                                <tt:Max>%f</tt:Max>
                            </tt:XRange>
                            <tt:YRange>
                                <tt:Min>%f</tt:Min>
                                <tt:Max>%f</tt:Max>
                            </tt:YRange>
                        </tt:AbsolutePanTiltPositionSpace>
                        <tt:AbsoluteZoomPositionSpace>
                            <tt:URI>http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace</tt:URI>
                            <tt:XRange>
                                <tt:Min>%f</tt:Min>
                                <tt:Max>%f</tt:Max>
                            </tt:XRange>
                        </tt:AbsoluteZoomPositionSpace>
                    </tt:SupportedPTZSpaces>
                    <tt:MaximumNumberOfPresets>0</tt:MaximumNumberOfPresets>
                    <tt:HomeSupported>false</tt:HomeSupported>
                </tptz:PTZNode>`,
		node.Token, node.Name,
		node.MinPan, node.MaxPan,
		node.MinTilt, node.MaxTilt,
		node.MinZoom, node.MaxZoom)

	return nodeElement
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF PTZ service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	// Create replacements map for template processing
	replacements := map[string]string{}

	// Process template and write response
	templatePath := filepath.Join("service_files", "ptz", "GetServiceCapabilities.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}

	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetServiceCapabilities template: %v", err)
	}

	w.Write([]byte(response))
	return nil
}

// GetNodesHTTP handles the GetNodes ONVIF PTZ service method via HTTP
func (s *ServiceContext) GetNodesHTTP(w http.ResponseWriter) error {
	// Create node elements
	nodeElements := make([]string, len(s.PTZNodes))
	for i, node := range s.PTZNodes {
		nodeElements[i] = s.createNodeElement(node)
	}

	nodesXML := ""
	if len(nodeElements) > 0 {
		nodesXML = nodeElements[0] // For simplicity, we're only using the first node
	}

	// Create replacements map for template processing
	replacements := map[string]string{
		"%NODES%": nodesXML,
	}

	// Process template and write response
	templatePath := filepath.Join("service_files", "ptz", "GetNodes.xml")
	if !xml.FileExists(templatePath) {
		// Fallback to generic template if service-specific one doesn't exist
		templatePath = filepath.Join("generic_files", "Empty.xml")
	}

	response, err := xml.ProcessTemplate(templatePath, replacements)
	if err != nil {
		return fmt.Errorf("failed to process GetNodes template: %v", err)
	}

	w.Write([]byte(response))
	return nil
}
