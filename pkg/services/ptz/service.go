// Package ptz provides ONVIF PTZ (Pan-Tilt-Zoom) service implementation.
package ptz

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/exec"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/utils"
	xmlfault "github.com/fawad-mazhar/onvif-go/internal/xml"
)

// ServiceContext holds the configuration and state for the PTZ service.
type ServiceContext struct {
	Port int
	Node Node
}

// Node holds the full PTZ configuration for a single PTZ node.
// Fields map 1-to-1 with config.PTZNode; populated by the dispatcher.
type Node struct {
	MinX, MaxX float64
	MinY, MaxY float64
	MinZ, MaxZ float64

	GetPosition      string
	IsMoving         string
	MoveLeft         string
	MoveRight        string
	MoveUp           string
	MoveDown         string
	MoveIn           string
	MoveOut          string
	MoveStop         string
	MovePreset       string
	GotoHomePosition string
	SetPreset        string
	SetHomePosition  string
	RemovePreset     string
	JumpToAbs        string
	JumpToRel        string
	GetPresets       string
}

// validateSOAPArg rejects shell meta-characters that could be injected through
// SOAP-derived string arguments before they reach exec.RunFmt.
//
// G-010 block-merge condition: this gate must exist at the handler layer.
// Characters blocked: ; $( ` && || >> | and --prefix tokens.
func validateSOAPArg(arg string) error {
	banned := []string{";", "$(", "`", "&&", "||", ">>", "|"}
	for _, b := range banned {
		if strings.Contains(arg, b) {
			return fmt.Errorf("rejected SOAP arg: contains %q", b)
		}
	}
	if strings.HasPrefix(arg, "--") {
		return fmt.Errorf("rejected SOAP arg: --flag-style tokens not allowed")
	}
	return nil
}

// ptzFault sends a SOAP fault for the ptz_service using the shared xmlfault renderer.
func ptzFault(w http.ResponseWriter, r *http.Request, recSend, subcode, subcodeEx, reason, detail string) error {
	devAddr, svcAddr := xmlfault.FaultAddrs(r)
	return xmlfault.WriteFault(w, xmlfault.Fault{
		Service:        "ptz_service",
		DeviceAddress:  devAddr,
		ServiceAddress: svcAddr,
		RecSend:        recSend,
		Subcode:        subcode,
		SubcodeEx:      subcodeEx,
		Reason:         reason,
		Detail:         detail,
	})
}

// actionFailed emits the canonical "ActionFailed" fault.
func actionFailed(w http.ResponseWriter, r *http.Request) error {
	return ptzFault(w, r, "Receiver", "ter:Action", "ter:ActionFailed", "Action failed", "PTZ action failed")
}

// step1 formats a float64 to C's %.1f notation used in PTZ space coordinate placeholders.
func step1(f float64) string { return fmt.Sprintf("%.1f", f) }

// ptzStepReplacements returns the standard %MIN_X%..%MAX_Z% placeholder map
// used by GetNodes, GetNode, GetConfigurations, GetConfiguration,
// GetConfigurationOptions templates.
func (s *ServiceContext) ptzStepReplacements() map[string]string {
	return map[string]string{
		"%MIN_X%": step1(s.Node.MinX),
		"%MAX_X%": step1(s.Node.MaxX),
		"%MIN_Y%": step1(s.Node.MinY),
		"%MAX_Y%": step1(s.Node.MaxY),
		"%MIN_Z%": step1(s.Node.MinZ),
		"%MAX_Z%": step1(s.Node.MaxZ),
	}
}

// requirePTZProfile validates ProfileToken presence in the SOAP request body.
// Returns a non-nil error (fault already written) when the token is absent.
func (s *ServiceContext) requirePTZProfile(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	pt, _ := xmlfault.ExtractElement([]byte(soapRequest), "ProfileToken")
	if pt == "" {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoProfile",
			"No profile", "The requested profile token ProfileToken does not exist")
	}
	return nil
}

// GetServiceCapabilitiesHTTP handles GetServiceCapabilities.
// MoveStatus / StatusPosition reflect whether is_moving / get_position scripts are configured.
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	moveStatus := "false"
	if s.Node.IsMoving != "" {
		moveStatus = "true"
	}
	statusPosition := "false"
	if s.Node.GetPosition != "" {
		statusPosition = "true"
	}
	return utils.ProcessServiceTemplate(w, "ptz", "GetServiceCapabilities", map[string]string{
		"%MOVE_STATUS%":     moveStatus,
		"%STATUS_POSITION%": statusPosition,
	})
}

// GetNodesHTTP handles GetNodes — returns the single PTZ node description.
func (s *ServiceContext) GetNodesHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "ptz", "GetNodes", s.ptzStepReplacements())
}

// GetNodeHTTP handles GetNode — validates NodeToken then returns the node.
func (s *ServiceContext) GetNodeHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	token, _ := xmlfault.ExtractElement([]byte(soapRequest), "NodeToken")
	if token != "PTZNodeToken" {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoEntity",
			"No entity", "No such node on the device")
	}
	return utils.ProcessServiceTemplate(w, "ptz", "GetNode", s.ptzStepReplacements())
}

// GetConfigurationsHTTP handles GetConfigurations.
func (s *ServiceContext) GetConfigurationsHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "ptz", "GetConfigurations", s.ptzStepReplacements())
}

// GetConfigurationHTTP handles GetConfiguration.
func (s *ServiceContext) GetConfigurationHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "ptz", "GetConfiguration", s.ptzStepReplacements())
}

// GetConfigurationOptionsHTTP handles GetConfigurationOptions.
func (s *ServiceContext) GetConfigurationOptionsHTTP(w http.ResponseWriter) error {
	return utils.ProcessServiceTemplate(w, "ptz", "GetConfigurationOptions", s.ptzStepReplacements())
}

// preset holds a parsed preset entry from the get_presets script output.
type preset struct {
	number int
	name   string
	x, y   float64
	z      float64
}

// parsePresets parses the stdout of a get_presets script.
// Each line is expected to be: number=name,pan,tilt[,zoom]
// (commas replaced by spaces before sscanf, matching C reference semantics).
func parsePresets(output string) []preset {
	var out []preset
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.ReplaceAll(line, ",", " ")
		var num int
		var name string
		var x, y, z float64
		z = 1.0
		n, err := fmt.Sscanf(line, "%d=%s %f %f %f", &num, &name, &x, &y, &z)
		if err != nil && n < 4 {
			continue
		}
		if name == "" {
			continue
		}
		out = append(out, preset{number: num, name: name, x: x, y: y, z: z})
	}
	return out
}

// GetPresetsHTTP handles GetPresets — runs get_presets script and assembles response.
func (s *ServiceContext) GetPresetsHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.GetPresets == "" {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoPTZProfile",
			"No PTZ profile", "The requested profile token does not reference a PTZ configuration")
	}

	raw, err := exec.Output(s.Node.GetPresets)
	if err != nil {
		logger.Warnf("PTZ GetPresets exec: %v", err)
	}

	presets := parsePresets(raw)
	var sb strings.Builder
	for _, p := range presets {
		sb.WriteString(fmt.Sprintf(
			`<tptz:Preset token="PresetToken_%d"><tt:Name>%s</tt:Name>`+
				`<tt:PTZPosition><tt:PanTilt x="%s" y="%s"/>`+
				`<tt:Zoom x="%s"/></tt:PTZPosition></tptz:Preset>`,
			p.number, p.name,
			fmt.Sprintf("%f", p.x), fmt.Sprintf("%f", p.y),
			fmt.Sprintf("%f", p.z),
		))
	}

	return utils.ProcessServiceTemplate(w, "ptz", "GetPresets", map[string]string{
		"%PRESETS%": sb.String(),
	})
}

// GotoPresetHTTP handles GotoPreset — validates token format then calls move_preset.
func (s *ServiceContext) GotoPresetHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	token, _ := xmlfault.ExtractElement([]byte(soapRequest), "PresetToken")
	var presetNumber int
	if n, _ := fmt.Sscanf(token, "PresetToken_%d", &presetNumber); n != 1 {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoToken",
			"No token", "The requested preset token does not exist")
	}
	if s.Node.MovePreset == "" {
		return actionFailed(w, r)
	}
	if err := exec.RunFmt(s.Node.MovePreset, presetNumber); err != nil {
		logger.Warnf("PTZ GotoPreset exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "GotoPreset", nil)
}

// GotoHomePositionHTTP handles GotoHomePosition.
func (s *ServiceContext) GotoHomePositionHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.GotoHomePosition == "" {
		return actionFailed(w, r)
	}
	if err := exec.Run(s.Node.GotoHomePosition); err != nil {
		logger.Warnf("PTZ GotoHomePosition exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "GotoHomePosition", nil)
}

// parseAttrFloat parses an XML attribute value (from ExtractAttr) to float64.
// Returns 0.0 and false if the value is absent or unparseable.
func parseAttrFloat(val string) (float64, bool) {
	if val == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// ContinuousMoveHTTP handles ContinuousMove.
// Reads PanTilt@x/y and Zoom@x from the Velocity element and dispatches the
// appropriate directional move command, matching C ptz_continuous_move() logic.
func (s *ServiceContext) ContinuousMoveHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	data := []byte(soapRequest)

	pxStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "x")
	pyStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "y")
	pzStr, _ := xmlfault.ExtractAttr(data, "Zoom", "x")

	if dx, ok := parseAttrFloat(pxStr); ok {
		if dx > 0 {
			if s.Node.MoveRight == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveRight, dx); err != nil {
				logger.Warnf("PTZ MoveRight: %v", err)
			}
		} else if dx < 0 {
			if s.Node.MoveLeft == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveLeft, -dx); err != nil {
				logger.Warnf("PTZ MoveLeft: %v", err)
			}
		}
	}
	if dy, ok := parseAttrFloat(pyStr); ok {
		if dy > 0 {
			if s.Node.MoveUp == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveUp, dy); err != nil {
				logger.Warnf("PTZ MoveUp: %v", err)
			}
		} else if dy < 0 {
			if s.Node.MoveDown == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveDown, -dy); err != nil {
				logger.Warnf("PTZ MoveDown: %v", err)
			}
		}
	}
	if dz, ok := parseAttrFloat(pzStr); ok {
		if dz > 0 {
			if s.Node.MoveIn == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveIn, dz); err != nil {
				logger.Warnf("PTZ MoveIn: %v", err)
			}
		} else if dz < 0 {
			if s.Node.MoveOut == "" {
				return actionFailed(w, r)
			}
			if err := exec.RunFmt(s.Node.MoveOut, -dz); err != nil {
				logger.Warnf("PTZ MoveOut: %v", err)
			}
		}
	}
	return utils.ProcessServiceTemplate(w, "ptz", "ContinuousMove", nil)
}

// RelativeMoveHTTP handles RelativeMove — reads Translation PanTilt@x/y and calls jump_to_rel.
func (s *ServiceContext) RelativeMoveHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.JumpToRel == "" {
		return actionFailed(w, r)
	}
	data := []byte(soapRequest)
	pxStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "x")
	pyStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "y")
	pzStr, _ := xmlfault.ExtractAttr(data, "Zoom", "x")

	dx, okX := parseAttrFloat(pxStr)
	dy, okY := parseAttrFloat(pyStr)
	if !okX || !okY {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:InvalidTranslation",
			"Invalid translation", "The requested translation is out of bounds")
	}
	dz, _ := parseAttrFloat(pzStr)

	if err := exec.RunFmt(s.Node.JumpToRel, dx, dy, dz); err != nil {
		logger.Warnf("PTZ RelativeMove exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "RelativeMove", nil)
}

// AbsoluteMoveHTTP handles AbsoluteMove — reads Position PanTilt@x/y/Zoom@x and calls jump_to_abs.
func (s *ServiceContext) AbsoluteMoveHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.JumpToAbs == "" {
		return actionFailed(w, r)
	}
	data := []byte(soapRequest)
	pxStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "x")
	pyStr, _ := xmlfault.ExtractAttr(data, "PanTilt", "y")
	pzStr, _ := xmlfault.ExtractAttr(data, "Zoom", "x")

	dx, okX := parseAttrFloat(pxStr)
	dy, okY := parseAttrFloat(pyStr)
	if !okX || !okY {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:InvalidPosition",
			"Invalid position", "The requested position is out of bounds")
	}
	dz, _ := parseAttrFloat(pzStr)

	if err := exec.RunFmt(s.Node.JumpToAbs, dx, dy, dz); err != nil {
		logger.Warnf("PTZ AbsoluteMove exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "AbsoluteMove", nil)
}

// StopHTTP handles Stop — reads optional PanTilt/Zoom booleans and calls move_stop.
func (s *ServiceContext) StopHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.MoveStop == "" {
		return actionFailed(w, r)
	}
	data := []byte(soapRequest)
	pantilt := true
	zoom := true
	pt, _ := xmlfault.ExtractElement(data, "PanTilt")
	if strings.EqualFold(pt, "false") {
		pantilt = false
	}
	z, _ := xmlfault.ExtractElement(data, "Zoom")
	if strings.EqualFold(z, "false") {
		zoom = false
	}

	var arg string
	switch {
	case pantilt && zoom:
		arg = "all"
	case pantilt:
		arg = "pantilt"
	case zoom:
		arg = "zoom"
	}
	if arg != "" {
		if err := exec.RunFmt(s.Node.MoveStop, arg); err != nil {
			logger.Warnf("PTZ Stop exec: %v", err)
		}
	}
	return utils.ProcessServiceTemplate(w, "ptz", "Stop", nil)
}

// parsePosition parses "x,y[,z]" stdout from get_position script.
// Returns (x, y, z, ok). z defaults to 1.0 if absent.
func parsePosition(output string) (x, y, z float64, ok bool) {
	output = strings.TrimSpace(output)
	z = 1.0
	var n int
	n, _ = fmt.Sscanf(output, "%f,%f,%f", &x, &y, &z)
	return x, y, z, n >= 2
}

// GetStatusHTTP handles GetStatus — calls get_position and is_moving scripts.
func (s *ServiceContext) GetStatusHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.GetPosition == "" {
		return ptzFault(w, r, "Receiver", "ter:Action", "ter:NoStatus",
			"No status", "No PTZ status is available in the requested Media Profile")
	}

	posOut, err := exec.Output(s.Node.GetPosition)
	if err != nil {
		logger.Warnf("PTZ GetStatus get_position: %v", err)
		return ptzFault(w, r, "Receiver", "ter:Action", "ter:NoStatus",
			"No status", "No PTZ status is available in the requested Media Profile")
	}
	x, y, z, ok := parsePosition(posOut)
	if !ok {
		return ptzFault(w, r, "Receiver", "ter:Action", "ter:NoStatus",
			"No status", "No PTZ status is available in the requested Media Profile")
	}

	moveStatus := "IDLE"
	if s.Node.IsMoving != "" {
		movOut, err := exec.Output(s.Node.IsMoving)
		if err == nil {
			var i int
			if n, _ := fmt.Sscanf(movOut, "%d", &i); n == 1 && i == 1 {
				moveStatus = "MOVING"
			}
		}
	}

	now := time.Now().UTC()
	utcTime := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	return utils.ProcessServiceTemplate(w, "ptz", "GetStatus", map[string]string{
		"%X%":                fmt.Sprintf("%f", x),
		"%Y%":                fmt.Sprintf("%f", y),
		"%Z%":                fmt.Sprintf("%f", z),
		"%MOVE_STATUS_PT%":   moveStatus,
		"%MOVE_STATUS_ZOOM%": "IDLE",
		"%TIME%":             utcTime,
	})
}

// SetPresetHTTP handles SetPreset — validates preset name then calls set_preset.
func (s *ServiceContext) SetPresetHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.SetPreset == "" {
		return actionFailed(w, r)
	}
	data := []byte(soapRequest)
	presetName, _ := xmlfault.ExtractElement(data, "PresetName")
	presetTokenStr, _ := xmlfault.ExtractElement(data, "PresetToken")

	if presetName == "" {
		presetName = fmt.Sprintf("Preset_%s", generateSimpleID())
	}

	if strings.Contains(presetName, " ") || len(presetName) == 0 || len(presetName) > 64 {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:InvalidPresetName",
			"Invalid preset name", "The preset name is either too long or contains invalid characters")
	}
	if err := validateSOAPArg(presetName); err != nil {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:InvalidPresetName",
			"Invalid preset name", err.Error())
	}

	presetNumber := -1
	if presetTokenStr != "" {
		if n, _ := fmt.Sscanf(presetTokenStr, "PresetToken_%d", &presetNumber); n != 1 {
			return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoToken",
				"No token", "The requested preset token does not exist")
		}
	}

	if err := exec.RunFmt(s.Node.SetPreset, presetNumber, presetName); err != nil {
		logger.Warnf("PTZ SetPreset exec: %v", err)
	}

	presetTokenOut := ""
	if presetNumber >= 0 {
		presetTokenOut = fmt.Sprintf("PresetToken_%d", presetNumber)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "SetPreset", map[string]string{
		"%PRESET_TOKEN%": presetTokenOut,
	})
}

// RemovePresetHTTP handles RemovePreset — validates token format then calls remove_preset.
func (s *ServiceContext) RemovePresetHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	token, _ := xmlfault.ExtractElement([]byte(soapRequest), "PresetToken")
	var presetNumber int
	if n, _ := fmt.Sscanf(token, "PresetToken_%d", &presetNumber); n != 1 {
		return ptzFault(w, r, "Sender", "ter:InvalidArgVal", "ter:NoToken",
			"No token", "The requested preset token does not exist")
	}
	if s.Node.RemovePreset == "" {
		return actionFailed(w, r)
	}
	if err := exec.RunFmt(s.Node.RemovePreset, presetNumber); err != nil {
		logger.Warnf("PTZ RemovePreset exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "RemovePreset", nil)
}

// SetHomePositionHTTP handles SetHomePosition — calls set_home_position script.
func (s *ServiceContext) SetHomePositionHTTP(w http.ResponseWriter, r *http.Request, soapRequest string) error {
	if err := s.requirePTZProfile(w, r, soapRequest); err != nil {
		return nil
	}
	if s.Node.SetHomePosition == "" {
		return actionFailed(w, r)
	}
	if err := exec.Run(s.Node.SetHomePosition); err != nil {
		logger.Warnf("PTZ SetHomePosition exec: %v", err)
	}
	return utils.ProcessServiceTemplate(w, "ptz", "SetHomePosition", nil)
}

// generateSimpleID generates a short random-ish identifier for unnamed presets.
// Uses nanoseconds as a lightweight alternative to UUID generation here.
func generateSimpleID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}
