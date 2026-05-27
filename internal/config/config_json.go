package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// jsonConfig is the JSON wire representation of ServiceContext.
// It uses string values for enum fields (stream type, audio type, idle state,
// events enable) so the file is human-readable without needing to know integer
// ordinals. It is intentionally package-private; callers use LoadConfigJSON.
type jsonConfig struct {
	Port              int               `json:"port"`
	NotificationPort  int               `json:"notification_port,omitempty"`
	WSDPort           int               `json:"wsd_port,omitempty"`
	User              string            `json:"user,omitempty"`
	Password          string            `json:"password,omitempty"`
	Manufacturer      string            `json:"manufacturer,omitempty"`
	Model             string            `json:"model,omitempty"`
	FirmwareVer       string            `json:"firmware_ver,omitempty"`
	SerialNum         string            `json:"serial_num,omitempty"`
	HardwareID        string            `json:"hardware_id,omitempty"`
	UUID              string            `json:"uuid,omitempty"`
	Interface         string            `json:"ifs,omitempty"`
	Scopes            []string          `json:"scopes,omitempty"`
	AdvFaultIfUnknown int               `json:"adv_fault_if_unknown,omitempty"`
	AdvFaultIfSet     int               `json:"adv_fault_if_set,omitempty"`
	AdvEnableMedia2   int               `json:"adv_enable_media2,omitempty"`
	AdvSynologyNVR    int               `json:"adv_synology_nvr,omitempty"`
	Profiles          []jsonProfile     `json:"profiles,omitempty"`
	PTZ               *jsonPTZ          `json:"ptz,omitempty"`
	RelayOutputs      []jsonRelayOutput `json:"relay_outputs,omitempty"`
	EventsEnable      int               `json:"events_enable,omitempty"`
	Events            []jsonEvent       `json:"events,omitempty"`
}

// jsonProfile is the JSON representation of a stream profile.
// Type and AudioEncoder/AudioDecoder use the same string values as the conf
// file: "H264", "H265", "MPEG4", "JPEG", "AAC", "G711", "G726", "NONE".
type jsonProfile struct {
	Name         string `json:"name"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	URL          string `json:"url,omitempty"`
	SnapURL      string `json:"snapurl,omitempty"`
	Type         string `json:"type,omitempty"`
	AudioEncoder string `json:"audio_encoder,omitempty"`
	AudioDecoder string `json:"audio_decoder,omitempty"`
}

// jsonPTZ is the JSON representation of PTZ configuration.
// Present only when PTZ is configured; enable=1 activates the node.
type jsonPTZ struct {
	Enable           int     `json:"enable"`
	MinStepX         float64 `json:"min_step_x,omitempty"`
	MaxStepX         float64 `json:"max_step_x,omitempty"`
	MinStepY         float64 `json:"min_step_y,omitempty"`
	MaxStepY         float64 `json:"max_step_y,omitempty"`
	MinStepZ         float64 `json:"min_step_z,omitempty"`
	MaxStepZ         float64 `json:"max_step_z,omitempty"`
	GetPosition      string  `json:"get_position,omitempty"`
	IsMoving         string  `json:"is_moving,omitempty"`
	MoveLeft         string  `json:"move_left,omitempty"`
	MoveRight        string  `json:"move_right,omitempty"`
	MoveUp           string  `json:"move_up,omitempty"`
	MoveDown         string  `json:"move_down,omitempty"`
	MoveIn           string  `json:"move_in,omitempty"`
	MoveOut          string  `json:"move_out,omitempty"`
	MoveStop         string  `json:"move_stop,omitempty"`
	MovePreset       string  `json:"move_preset,omitempty"`
	GotoHomePosition string  `json:"goto_home_position,omitempty"`
	SetPreset        string  `json:"set_preset,omitempty"`
	SetHomePosition  string  `json:"set_home_position,omitempty"`
	RemovePreset     string  `json:"remove_preset,omitempty"`
	JumpToAbs        string  `json:"jump_to_abs,omitempty"`
	JumpToRel        string  `json:"jump_to_rel,omitempty"`
	GetPresets       string  `json:"get_presets,omitempty"`
}

// jsonRelayOutput is the JSON representation of a relay output.
// IdleState uses the strings "open" or "close" matching the conf-file values.
type jsonRelayOutput struct {
	IdleState string `json:"idle_state"`
	CloseCmd  string `json:"close,omitempty"`
	OpenCmd   string `json:"open,omitempty"`
}

// jsonEvent is the JSON representation of a single event definition.
type jsonEvent struct {
	Topic       string `json:"topic"`
	SourceName  string `json:"source_name,omitempty"`
	SourceType  string `json:"source_type,omitempty"`
	SourceValue string `json:"source_value,omitempty"`
	InputFile   string `json:"input_file,omitempty"`
}

// LoadConfigJSON loads a ServiceContext from a JSON file.
//
// The JSON format uses human-readable string values for enum fields (stream
// type, audio type, idle state) and a nested "ptz" object rather than the
// sequential flat-key approach of the .conf format. Both formats populate the
// same ServiceContext struct and are fully interchangeable at runtime.
//
// See test/fixtures/config/server.json for a complete example that mirrors
// test/fixtures/config/server.conf.
func LoadConfigJSON(filename string) (*ServiceContext, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON config: %v", err)
	}

	var jcfg jsonConfig
	if err := json.Unmarshal(data, &jcfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %v", err)
	}

	return convertJSONConfig(&jcfg)
}

// convertJSONConfig maps a jsonConfig DTO to ServiceContext, applying the same
// enum conversions and defaults as the flat-config parser:
//   - Profile with no type defaults to H264; no audio_encoder defaults to AAC.
//     (Mirrors parseFlatConfig's append-time defaults.)
//   - PTZ with enable=1 and zero MaxStepX/MaxStepY defaults to 360/180.
//     (Mirrors the `ptz=1` branch in parseFlatConfig.)
//   - Port validated against 0-65535 like parsePortValue.
func convertJSONConfig(jcfg *jsonConfig) (*ServiceContext, error) {
	if jcfg.Port < 0 || jcfg.Port > 65535 {
		return nil, fmt.Errorf("port value out of range: %d", jcfg.Port)
	}
	ctx := &ServiceContext{
		Port:              jcfg.Port,
		NotificationPort:  jcfg.NotificationPort,
		WSDPort:           jcfg.WSDPort,
		User:              jcfg.User,
		Password:          jcfg.Password,
		Manufacturer:      jcfg.Manufacturer,
		Model:             jcfg.Model,
		FirmwareVer:       jcfg.FirmwareVer,
		SerialNum:         jcfg.SerialNum,
		HardwareID:        jcfg.HardwareID,
		UUID:              jcfg.UUID,
		Interface:         jcfg.Interface,
		AdvFaultIfUnknown: jcfg.AdvFaultIfUnknown,
		AdvFaultIfSet:     jcfg.AdvFaultIfSet,
		AdvEnableMedia2:   jcfg.AdvEnableMedia2,
		AdvSynologyNVR:    jcfg.AdvSynologyNVR,
		Scopes:            jcfg.Scopes,
		Profiles:          make([]StreamProfile, 0, len(jcfg.Profiles)),
		RelayOutputs:      make([]RelayOutput, 0, len(jcfg.RelayOutputs)),
		Events:            make([]Event, 0, len(jcfg.Events)),
	}
	if ctx.Scopes == nil {
		ctx.Scopes = make([]string, 0)
	}

	// Profiles — apply same append-time defaults as parseFlatConfig
	for _, jp := range jcfg.Profiles {
		sp := StreamProfile{
			Name:         jp.Name,
			Width:        jp.Width,
			Height:       jp.Height,
			URL:          jp.URL,
			SnapURL:      jp.SnapURL,
			Type:         parseStreamType(jp.Type),
			AudioEncoder: parseAudioType(jp.AudioEncoder),
			AudioDecoder: parseAudioType(jp.AudioDecoder),
		}
		if jp.Type == "" {
			sp.Type = H264
		}
		if jp.AudioEncoder == "" {
			sp.AudioEncoder = AAC
		}
		ctx.Profiles = append(ctx.Profiles, sp)
	}

	// PTZ node — apply parseFlatConfig defaults (360/180) when enable=1 and steps unset
	if jcfg.PTZ != nil {
		p := jcfg.PTZ
		maxX, maxY := p.MaxStepX, p.MaxStepY
		if p.Enable == 1 {
			if maxX == 0 {
				maxX = 360.0
			}
			if maxY == 0 {
				maxY = 180.0
			}
		}
		ctx.PTZNode = PTZNode{
			Enable:           p.Enable,
			MinStepX:         p.MinStepX,
			MaxStepX:         maxX,
			MinStepY:         p.MinStepY,
			MaxStepY:         maxY,
			MinStepZ:         p.MinStepZ,
			MaxStepZ:         p.MaxStepZ,
			GetPosition:      p.GetPosition,
			IsMoving:         p.IsMoving,
			MoveLeft:         p.MoveLeft,
			MoveRight:        p.MoveRight,
			MoveUp:           p.MoveUp,
			MoveDown:         p.MoveDown,
			MoveIn:           p.MoveIn,
			MoveOut:          p.MoveOut,
			MoveStop:         p.MoveStop,
			MovePreset:       p.MovePreset,
			GotoHomePosition: p.GotoHomePosition,
			SetPreset:        p.SetPreset,
			SetHomePosition:  p.SetHomePosition,
			RemovePreset:     p.RemovePreset,
			JumpToAbs:        p.JumpToAbs,
			JumpToRel:        p.JumpToRel,
			GetPresets:       p.GetPresets,
		}
	}

	// Relay outputs
	for _, jr := range jcfg.RelayOutputs {
		ro := RelayOutput{
			CloseCmd: jr.CloseCmd,
			OpenCmd:  jr.OpenCmd,
		}
		if strings.EqualFold(jr.IdleState, "open") {
			ro.IdleState = IdleStateOpen
		}
		ctx.RelayOutputs = append(ctx.RelayOutputs, ro)
	}
	ctx.RelayOutputsNum = len(ctx.RelayOutputs)

	// Events enable: JSON stores the integer (0-3) matching C values
	ctx.EventsEnable = EventsEnable(jcfg.EventsEnable)

	// Events
	for _, je := range jcfg.Events {
		ctx.Events = append(ctx.Events, Event(je))
	}
	ctx.EventsNum = len(ctx.Events)

	return ctx, nil
}
