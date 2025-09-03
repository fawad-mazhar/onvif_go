package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	HardwareId        string
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
	VIDEO_NONE StreamType = iota
	JPEG
	MPEG4
	H264
	H265
)

// String returns the string representation of StreamType
func (s StreamType) String() string {
	switch s {
	case VIDEO_NONE:
		return "VIDEO_NONE"
	case JPEG:
		return "JPEG"
	case MPEG4:
		return "MPEG4"
	case H264:
		return "H264"
	case H265:
		return "H265"
	default:
		return "UNKNOWN"
	}
}

// AudioType represents the audio stream type
type AudioType int

const (
	AUDIO_NONE AudioType = iota
	G711
	G726
	AAC
)

// String returns the string representation of AudioType
func (a AudioType) String() string {
	switch a {
	case AUDIO_NONE:
		return "AUDIO_NONE"
	case G711:
		return "G711"
	case G726:
		return "G726"
	case AAC:
		return "AAC"
	default:
		return "UNKNOWN"
	}
}

// IdleState represents the relay output idle state
type IdleState int

const (
	IDLE_STATE_CLOSE IdleState = iota
	IDLE_STATE_OPEN
)

// EventsEnable represents the events service enable state
type EventsEnable int

const (
	EVENTS_NONE EventsEnable = iota
	EVENTS_PULLPOINT
	EVENTS_BASESUBSCRIPTION
	EVENTS_BOTH
)

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*ServiceContext, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

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
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
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

// parseConfigValue parses a single key=value configuration pair
func parseConfigValue(context *ServiceContext, key, value string) error {
	switch key {
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port value: %v", err)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("port value out of range: %d", port)
		}
		context.Port = port

	case "user":
		context.User = value

	case "password":
		context.Password = value

	case "manufacturer":
		context.Manufacturer = value

	case "model":
		context.Model = value

	case "firmware_ver":
		context.FirmwareVer = value

	case "serial_num":
		context.SerialNum = value

	case "hardware_id":
		context.HardwareId = value

	case "uuid":
		context.UUID = value

	case "interface":
		context.Interface = value

	case "adv_enable_media2":
		enable, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid adv_enable_media2 value: %v", err)
		}
		context.AdvEnableMedia2 = enable

	case "adv_fault_if_unknown":
		fault, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid adv_fault_if_unknown value: %v", err)
		}
		context.AdvFaultIfUnknown = fault

	case "adv_fault_if_set":
		fault, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid adv_fault_if_set value: %v", err)
		}
		context.AdvFaultIfSet = fault

	case "adv_synology_nvr":
		synology, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid adv_synology_nvr value: %v", err)
		}
		context.AdvSynologyNVR = synology

	case "notification_port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid notification_port value: %v", err)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("notification_port value out of range: %d", port)
		}
		context.NotificationPort = port

	case "wsd_port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid wsd_port value: %v", err)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("wsd_port value out of range: %d", port)
		}
		context.WSDPort = port

	default:
		// Handle profile configurations
		if strings.HasPrefix(key, "profile.") {
			return parseProfileConfig(context, key, value)
		}

		// Handle scope configurations
		if strings.HasPrefix(key, "scope.") {
			return parseScopeConfig(context, key, value)
		}

		// Handle relay output configurations
		if strings.HasPrefix(key, "relayoutput.") {
			return parseRelayOutputConfig(context, key, value)
		}

		// Handle PTZ configurations
		if strings.HasPrefix(key, "ptz.") {
			return parsePTZConfig(context, key, value)
		}

		// Handle event configurations
		if strings.HasPrefix(key, "event.") {
			return parseEventConfig(context, key, value)
		}

		// Handle events enable configuration
		if key == "events_enable" {
			return parseEventsEnableConfig(context, value)
		}
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
			relay.IdleState = IDLE_STATE_OPEN
		} else {
			relay.IdleState = IDLE_STATE_CLOSE
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

	switch property {
	case "enable":
		enable, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid ptz enable value: %v", err)
		}
		context.PTZNode.Enable = enable
	case "min_step_x":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid min_step_x value: %v", err)
		}
		context.PTZNode.MinStepX = step
	case "max_step_x":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid max_step_x value: %v", err)
		}
		context.PTZNode.MaxStepX = step
	case "min_step_y":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid min_step_y value: %v", err)
		}
		context.PTZNode.MinStepY = step
	case "max_step_y":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid max_step_y value: %v", err)
		}
		context.PTZNode.MaxStepY = step
	case "min_step_z":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid min_step_z value: %v", err)
		}
		context.PTZNode.MinStepZ = step
	case "max_step_z":
		step, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid max_step_z value: %v", err)
		}
		context.PTZNode.MaxStepZ = step
	case "get_position":
		context.PTZNode.GetPosition = value
	case "is_moving":
		context.PTZNode.IsMoving = value
	case "move_left":
		context.PTZNode.MoveLeft = value
	case "move_right":
		context.PTZNode.MoveRight = value
	case "move_up":
		context.PTZNode.MoveUp = value
	case "move_down":
		context.PTZNode.MoveDown = value
	case "move_in":
		context.PTZNode.MoveIn = value
	case "move_out":
		context.PTZNode.MoveOut = value
	case "move_stop":
		context.PTZNode.MoveStop = value
	case "move_preset":
		context.PTZNode.MovePreset = value
	case "goto_home_position":
		context.PTZNode.GotoHomePosition = value
	case "set_preset":
		context.PTZNode.SetPreset = value
	case "set_home_position":
		context.PTZNode.SetHomePosition = value
	case "remove_preset":
		context.PTZNode.RemovePreset = value
	case "jump_to_abs":
		context.PTZNode.JumpToAbs = value
	case "jump_to_rel":
		context.PTZNode.JumpToRel = value
	case "get_presets":
		context.PTZNode.GetPresets = value
	}

	return nil
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
		context.EventsEnable = EVENTS_PULLPOINT
	case "basesubscription":
		context.EventsEnable = EVENTS_BASESUBSCRIPTION
	case "both":
		context.EventsEnable = EVENTS_BOTH
	default:
		context.EventsEnable = EVENTS_NONE
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
		return VIDEO_NONE
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
		return AUDIO_NONE
	}
}
