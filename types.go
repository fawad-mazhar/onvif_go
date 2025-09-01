package main

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

// UsernameToken represents the security token from the SOAP header
type UsernameToken struct {
	Enable  int
	Username string
	Password string
	Nonce   string
	Created string
	Type    int
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
	Enable       int
	MinStepX     float64
	MaxStepX     float64
	MinStepY     float64
	MaxStepY     float64
	MinStepZ     float64
	MaxStepZ     float64
	GetPosition  string
	IsMoving     string
	MoveLeft     string
	MoveRight    string
	MoveUp       string
	MoveDown     string
	MoveIn       string
	MoveOut      string
	MoveStop     string
	MovePreset   string
	GotoHomePosition string
	SetPreset    string
	SetHomePosition string
	RemovePreset string
	JumpToAbs    string
	JumpToRel    string
	GetPresets   string
}

// Event represents an event configuration
type Event struct {
	Topic       string
	SourceName  string
	SourceType  string
	SourceValue string
	InputFile   string
}

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
}
