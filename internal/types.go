package main

import (
	"github.com/fawad-mazhar/onvif-go/internal/config"
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
	Type         config.StreamType
	AudioEncoder config.AudioType
	AudioDecoder config.AudioType
}

// RelayOutput represents a relay output configuration
type RelayOutput struct {
	IdleState config.IdleState
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
