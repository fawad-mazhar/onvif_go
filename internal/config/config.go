// Package config provides configuration parsing and management for ONVIF services.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	unknownType = "UNKNOWN"

	// Configuration parsing constants
	configKeyValueParts = 2 // Expected parts when splitting key=value
	minProfileParts     = 3 // Minimum parts for profile configuration
	minRelayParts       = 2 // Minimum parts for relay configuration
	minEventParts       = 3 // Minimum parts for event configuration
)

// ServiceContext represents the main service context
type ServiceContext struct {
	Port              int
	User              string
	Password          string
	Manufacturer      string
	Model             string
	FirmwareVer       string
	SerialNum         string
	HardwareID        string
	UUID              string
	Interface         string
	AdvEnableMedia2   int
	AdvFaultIfUnknown int
	AdvFaultIfSet     int
	AdvSynologyNVR    int
	Profiles          []StreamProfile
	Scopes            []string
	RelayOutputs      []RelayOutput
	RelayOutputsNum   int
	PTZNode           PTZNode
	Events            []Event
	EventsEnable      EventsEnable
	EventsNum         int
	NotificationPort  int
	WSDPort           int
}

// StreamProfile represents a media profile configuration
type StreamProfile struct {
	Name         string
	Width        int
	Height       int
	URL          string
	SnapURL      string
	Type         StreamType
	AudioEncoder AudioType
	AudioDecoder AudioType
}

// RelayOutput represents a relay output configuration
type RelayOutput struct {
	IdleState IdleState
	CloseCmd  string
	OpenCmd   string
}

// PTZNode represents PTZ configuration
type PTZNode struct {
	Enable           int
	MinStepX         float64
	MaxStepX         float64
	MinStepY         float64
	MaxStepY         float64
	MinStepZ         float64
	MaxStepZ         float64
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

// Event represents an event configuration
type Event struct {
	Topic       string
	SourceName  string
	SourceType  string
	SourceValue string
	InputFile   string
}

// StreamType represents the video stream type
type StreamType int

const (
	// VideoNone represents no video stream
	VideoNone StreamType = iota
	// JPEG represents JPEG video stream
	JPEG
	// MPEG4 represents MPEG4 video stream
	MPEG4
	// H264 represents H264 video stream
	H264
	// H265 represents H265 video stream
	H265
)

// String returns the string representation of StreamType
func (s StreamType) String() string {
	switch s {
	case VideoNone:
		return "VideoNone"
	case JPEG:
		return "JPEG"
	case MPEG4:
		return "MPEG4"
	case H264:
		return "H264"
	case H265:
		return "H265"
	default:
		return unknownType
	}
}

// AudioType represents the audio stream type
type AudioType int

const (
	// AudioNone represents no audio stream
	AudioNone AudioType = iota
	// G711 represents G711 audio codec
	G711
	// G726 represents G726 audio codec
	G726
	// AAC represents AAC audio codec
	AAC
)

// String returns the string representation of AudioType
func (a AudioType) String() string {
	switch a {
	case AudioNone:
		return "AudioNone"
	case G711:
		return "G711"
	case G726:
		return "G726"
	case AAC:
		return "AAC"
	default:
		return unknownType
	}
}

// IdleState represents the relay output idle state
type IdleState int

const (
	// IdleStateClose represents a closed idle state
	IdleStateClose IdleState = iota
	// IdleStateOpen represents an open idle state
	IdleStateOpen
)

// EventsEnable represents the events service enable state
type EventsEnable int

const (
	// EventsNone represents no events enabled
	EventsNone EventsEnable = iota
	// EventsPullPoint represents pull point events enabled
	EventsPullPoint
	// EventsBaseSubscription represents base subscription events enabled
	EventsBaseSubscription
	// EventsBoth represents both pull point and base subscription events enabled
	EventsBoth
)

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*ServiceContext, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Ignore close error - config file was successfully read
			_ = closeErr
		}
	}()

	context := &ServiceContext{
		Profiles:     make([]StreamProfile, 0),
		Scopes:       make([]string, 0),
		RelayOutputs: make([]RelayOutput, 0),
		Events:       make([]Event, 0),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", configKeyValueParts)
		if len(parts) != configKeyValueParts {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) > 1 {
			value = value[1 : len(value)-1]
		}

		err := parseConfigValue(context, key, value)
		if err != nil {
			return nil, fmt.Errorf("error parsing config key %s: %v", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	return context, nil
}

// parsePortValue parses and validates a port value
func parsePortValue(value, fieldName string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value: %v", fieldName, err)
	}
	if port < 0 || port > 65535 {
		return 0, fmt.Errorf("%s value out of range: %d", fieldName, port)
	}
	return port, nil
}

// parseConfigValue parses a single key=value configuration pair
func parseConfigValue(context *ServiceContext, key, value string) error {
	// Try parsing basic string configurations
	if err := parseBasicConfig(context, key, value); err == nil {
		return nil
	}

	// Try parsing integer configurations
	if err := parseIntConfig(context, key, value); err != nil {
		return err
	}

	// Try parsing port configurations
	if err := parsePortConfig(context, key, value); err != nil {
		return err
	}

	// Handle prefixed configurations
	return parsePrefixedConfig(context, key, value)
}

// parseBasicConfig handles basic string configuration values
func parseBasicConfig(context *ServiceContext, key, value string) error {
	basicConfigs := map[string]*string{
		"user":         &context.User,
		"password":     &context.Password,
		"manufacturer": &context.Manufacturer,
		"model":        &context.Model,
		"firmware_ver": &context.FirmwareVer,
		"serial_num":   &context.SerialNum,
		"hardware_id":  &context.HardwareID,
		"uuid":         &context.UUID,
		"interface":    &context.Interface,
	}

	if target, exists := basicConfigs[key]; exists {
		*target = value
		return nil
	}

	return fmt.Errorf("not a basic config: %s", key)
}

// parseIntConfig handles integer configuration values
func parseIntConfig(context *ServiceContext, key, value string) error {
	intConfigs := map[string]*int{
		"adv_enable_media2":    &context.AdvEnableMedia2,
		"adv_fault_if_unknown": &context.AdvFaultIfUnknown,
		"adv_fault_if_set":     &context.AdvFaultIfSet,
		"adv_synology_nvr":     &context.AdvSynologyNVR,
	}

	if target, exists := intConfigs[key]; exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid %s value: %v", key, err)
		}
		*target = intValue
		return nil
	}

	return nil
}

// parsePortConfig handles port configuration values
func parsePortConfig(context *ServiceContext, key, value string) error {
	portConfigs := map[string]*int{
		"port":              &context.Port,
		"notification_port": &context.NotificationPort,
		"wsd_port":          &context.WSDPort,
	}

	if target, exists := portConfigs[key]; exists {
		port, err := parsePortValue(value, key)
		if err != nil {
			return err
		}
		*target = port
		return nil
	}

	return nil
}

// parsePrefixedConfig handles configuration keys with prefixes
func parsePrefixedConfig(context *ServiceContext, key, value string) error {
	prefixHandlers := []struct {
		prefix  string
		handler func(*ServiceContext, string, string) error
	}{
		{"profile.", parseProfileConfig},
		{"scope.", parseScopeConfig},
		{"relayoutput.", parseRelayOutputConfig},
		{"ptz.", parsePTZConfig},
		{"event.", parseEventConfig},
	}

	for _, handler := range prefixHandlers {
		if strings.HasPrefix(key, handler.prefix) {
			return handler.handler(context, key, value)
		}
	}

	// Handle events enable configuration
	if key == "events_enable" {
		return parseEventsEnableConfig(context, value)
	}

	return nil
}

// parseProfileConfig parses profile-related configuration values
func parseProfileConfig(context *ServiceContext, key, value string) error {
	// Extract profile index and property
	parts := strings.Split(key, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid profile key format: %s", key)
	}

	profileIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid profile index: %v", err)
	}

	// Ensure we have enough profiles
	for len(context.Profiles) <= profileIndex {
		context.Profiles = append(context.Profiles, StreamProfile{})
	}

	property := parts[2]
	profile := &context.Profiles[profileIndex]

	switch property {
	case "name":
		profile.Name = value
	case "width":
		width, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid width value: %v", err)
		}
		profile.Width = width
	case "height":
		height, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid height value: %v", err)
		}
		profile.Height = height
	case "url":
		profile.URL = value
	case "snapurl":
		profile.SnapURL = value
	case "type":
		profile.Type = parseStreamType(value)
	case "audio_encoder":
		profile.AudioEncoder = parseAudioType(value)
	case "audio_decoder":
		profile.AudioDecoder = parseAudioType(value)
	}

	return nil
}

// parseScopeConfig parses scope-related configuration values
func parseScopeConfig(context *ServiceContext, key, value string) error {
	// Extract scope index
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid scope key format: %s", key)
	}

	scopeIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid scope index: %v", err)
	}

	// Ensure we have enough scopes
	for len(context.Scopes) <= scopeIndex {
		context.Scopes = append(context.Scopes, "")
	}

	context.Scopes[scopeIndex] = value
	return nil
}

// parseRelayOutputConfig parses relay output-related configuration values
func parseRelayOutputConfig(context *ServiceContext, key, value string) error {
	// Extract relay output index and property
	parts := strings.Split(key, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid relay output key format: %s", key)
	}

	relayIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid relay output index: %v", err)
	}

	// Ensure we have enough relay outputs
	for len(context.RelayOutputs) <= relayIndex {
		context.RelayOutputs = append(context.RelayOutputs, RelayOutput{})
	}

	property := parts[2]
	relay := &context.RelayOutputs[relayIndex]

	switch property {
	case "idle_state":
		if value == "open" {
			relay.IdleState = IdleStateOpen
		} else {
			relay.IdleState = IdleStateClose
		}
	case "close_cmd":
		relay.CloseCmd = value
	case "open_cmd":
		relay.OpenCmd = value
	}

	return nil
}

// parsePTZConfig parses PTZ-related configuration values
func parsePTZConfig(context *ServiceContext, key, value string) error {
	property := strings.TrimPrefix(key, "ptz.")

	// Handle enable property
	if property == "enable" {
		return parsePTZEnable(context, value)
	}

	// Handle step properties
	if err := parsePTZStepProperty(context, property, value); err != nil {
		return err
	}

	// Handle command properties
	parsePTZCommand(context, property, value)

	return nil
}

// parsePTZEnable parses the PTZ enable configuration
func parsePTZEnable(context *ServiceContext, value string) error {
	enable, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid ptz enable value: %v", err)
	}
	context.PTZNode.Enable = enable
	return nil
}

// parsePTZStepProperty parses PTZ step-related properties
func parsePTZStepProperty(context *ServiceContext, property, value string) error {
	stepProperties := map[string]*float64{
		"min_step_x": &context.PTZNode.MinStepX,
		"max_step_x": &context.PTZNode.MaxStepX,
		"min_step_y": &context.PTZNode.MinStepY,
		"max_step_y": &context.PTZNode.MaxStepY,
		"min_step_z": &context.PTZNode.MinStepZ,
		"max_step_z": &context.PTZNode.MaxStepZ,
	}

	if target, exists := stepProperties[property]; exists {
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid %s value: %v", property, err)
		}
		*target = step
	}

	return nil
}

// parsePTZCommand parses PTZ command properties
func parsePTZCommand(context *ServiceContext, property, value string) {
	commandProperties := map[string]*string{
		"get_position":       &context.PTZNode.GetPosition,
		"is_moving":          &context.PTZNode.IsMoving,
		"move_left":          &context.PTZNode.MoveLeft,
		"move_right":         &context.PTZNode.MoveRight,
		"move_up":            &context.PTZNode.MoveUp,
		"move_down":          &context.PTZNode.MoveDown,
		"move_in":            &context.PTZNode.MoveIn,
		"move_out":           &context.PTZNode.MoveOut,
		"move_stop":          &context.PTZNode.MoveStop,
		"move_preset":        &context.PTZNode.MovePreset,
		"goto_home_position": &context.PTZNode.GotoHomePosition,
		"set_preset":         &context.PTZNode.SetPreset,
		"set_home_position":  &context.PTZNode.SetHomePosition,
		"remove_preset":      &context.PTZNode.RemovePreset,
		"jump_to_abs":        &context.PTZNode.JumpToAbs,
		"jump_to_rel":        &context.PTZNode.JumpToRel,
		"get_presets":        &context.PTZNode.GetPresets,
	}

	if target, exists := commandProperties[property]; exists {
		*target = value
	}
}

// parseEventConfig parses event-related configuration values
func parseEventConfig(context *ServiceContext, key, value string) error {
	// Extract event index and property
	parts := strings.Split(key, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid event key format: %s", key)
	}

	eventIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid event index: %v", err)
	}

	// Ensure we have enough events
	for len(context.Events) <= eventIndex {
		context.Events = append(context.Events, Event{})
	}

	property := parts[2]
	event := &context.Events[eventIndex]

	switch property {
	case "topic":
		event.Topic = value
	case "source_name":
		event.SourceName = value
	case "source_type":
		event.SourceType = value
	case "source_value":
		event.SourceValue = value
	case "input_file":
		event.InputFile = value
	}

	return nil
}

// parseEventsEnableConfig parses events enable configuration
func parseEventsEnableConfig(context *ServiceContext, value string) error {
	switch value {
	case "pullpoint":
		context.EventsEnable = EventsPullPoint
	case "basesubscription":
		context.EventsEnable = EventsBaseSubscription
	case "both":
		context.EventsEnable = EventsBoth
	default:
		context.EventsEnable = EventsNone
	}

	return nil
}

// parseStreamType converts string to StreamType enum
func parseStreamType(value string) StreamType {
	switch strings.ToLower(value) {
	case "jpeg":
		return JPEG
	case "mpeg4":
		return MPEG4
	case "h264":
		return H264
	case "h265":
		return H265
	default:
		return VideoNone
	}
}

// parseAudioType converts string to AudioType enum
func parseAudioType(value string) AudioType {
	switch strings.ToLower(value) {
	case "g711":
		return G711
	case "g726":
		return G726
	case "aac":
		return AAC
	default:
		return AudioNone
	}
}
