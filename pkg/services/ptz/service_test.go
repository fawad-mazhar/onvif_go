package ptz

import (
	"strings"
	"testing"
)

// TestValidateSOAPArg covers the G-010 corpus gate: every dangerous shell
// meta-character sequence must be rejected; clean inputs must pass.
func TestValidateSOAPArg(t *testing.T) {
	t.Run("clean_names_pass", func(t *testing.T) {
		clean := []string{"MyPreset", "Preset1", "home", "front-door", "cam_roof"}
		for _, s := range clean {
			if err := validateSOAPArg(s); err != nil {
				t.Errorf("clean arg %q rejected: %v", s, err)
			}
		}
	})

	t.Run("semicolon_rejected", func(t *testing.T) {
		if err := validateSOAPArg("ok;rm -rf /"); err == nil {
			t.Error("expected rejection of semicolon, got nil")
		}
	})

	t.Run("dollar_paren_rejected", func(t *testing.T) {
		if err := validateSOAPArg("$(whoami)"); err == nil {
			t.Error("expected rejection of $(...), got nil")
		}
	})

	t.Run("backtick_rejected", func(t *testing.T) {
		if err := validateSOAPArg("`id`"); err == nil {
			t.Error("expected rejection of backtick, got nil")
		}
	})

	t.Run("double_ampersand_rejected", func(t *testing.T) {
		if err := validateSOAPArg("name&&rm -rf /"); err == nil {
			t.Error("expected rejection of &&, got nil")
		}
	})

	t.Run("double_pipe_rejected", func(t *testing.T) {
		if err := validateSOAPArg("name||cat /etc/passwd"); err == nil {
			t.Error("expected rejection of ||, got nil")
		}
	})

	t.Run("append_redirect_rejected", func(t *testing.T) {
		if err := validateSOAPArg("name>>file"); err == nil {
			t.Error("expected rejection of >>, got nil")
		}
	})

	t.Run("pipe_rejected", func(t *testing.T) {
		if err := validateSOAPArg("name|cat"); err == nil {
			t.Error("expected rejection of |, got nil")
		}
	})

	t.Run("double_dash_prefix_rejected", func(t *testing.T) {
		if err := validateSOAPArg("--flag"); err == nil {
			t.Error("expected rejection of --flag token, got nil")
		}
	})

	t.Run("double_dash_embedded_allowed", func(t *testing.T) {
		// "--" in the middle of a name (not a prefix) is fine
		if err := validateSOAPArg("my--preset"); err != nil {
			t.Errorf("mid-string -- should be allowed, got: %v", err)
		}
	})
}

// TestParsePresets verifies that get_presets stdout is parsed correctly,
// including the comma→space substitution that matches C reference semantics.
func TestParsePresets(t *testing.T) {
	t.Run("single_preset_with_zoom", func(t *testing.T) {
		out := "1=Home,0.500000,0.250000,1.000000\n"
		ps := parsePresets(out)
		if len(ps) != 1 {
			t.Fatalf("expected 1 preset, got %d", len(ps))
		}
		p := ps[0]
		if p.number != 1 {
			t.Errorf("number: want 1, got %d", p.number)
		}
		if p.name != "Home" {
			t.Errorf("name: want Home, got %q", p.name)
		}
		if p.x != 0.5 {
			t.Errorf("x: want 0.5, got %f", p.x)
		}
		if p.y != 0.25 {
			t.Errorf("y: want 0.25, got %f", p.y)
		}
		if p.z != 1.0 {
			t.Errorf("z: want 1.0, got %f", p.z)
		}
	})

	t.Run("multiple_presets_no_zoom", func(t *testing.T) {
		out := "1=Garden,10.0,5.0\n2=Door,100.0,50.0\n"
		ps := parsePresets(out)
		if len(ps) != 2 {
			t.Fatalf("expected 2 presets, got %d", len(ps))
		}
		if ps[0].name != "Garden" || ps[1].name != "Door" {
			t.Errorf("names: want [Garden Door], got [%s %s]", ps[0].name, ps[1].name)
		}
		// zoom defaults to 1.0 when absent
		if ps[0].z != 1.0 {
			t.Errorf("z default: want 1.0, got %f", ps[0].z)
		}
	})

	t.Run("empty_output_returns_empty_slice", func(t *testing.T) {
		ps := parsePresets("")
		if len(ps) != 0 {
			t.Errorf("expected empty slice, got %d presets", len(ps))
		}
	})

	t.Run("malformed_line_skipped", func(t *testing.T) {
		out := "not-a-preset\n1=Valid,10.0,5.0\n"
		ps := parsePresets(out)
		if len(ps) != 1 {
			t.Fatalf("expected 1 valid preset, got %d", len(ps))
		}
		if ps[0].name != "Valid" {
			t.Errorf("name: want Valid, got %q", ps[0].name)
		}
	})
}

// TestParsePosition verifies that get_position stdout is parsed correctly.
func TestParsePosition(t *testing.T) {
	t.Run("x_y_z", func(t *testing.T) {
		x, y, z, ok := parsePosition("100.5,50.2,1.0")
		if !ok {
			t.Fatal("expected ok=true")
		}
		if x != 100.5 || y != 50.2 || z != 1.0 {
			t.Errorf("want 100.5,50.2,1.0 got %f,%f,%f", x, y, z)
		}
	})

	t.Run("x_y_only_z_defaults_to_1", func(t *testing.T) {
		x, y, z, ok := parsePosition("100.0,50.0")
		if !ok {
			t.Fatal("expected ok=true")
		}
		if z != 1.0 {
			t.Errorf("z default: want 1.0, got %f", z)
		}
		_ = x
		_ = y
	})

	t.Run("single_value_not_ok", func(t *testing.T) {
		_, _, _, ok := parsePosition("100.0")
		if ok {
			t.Error("expected ok=false for single value")
		}
	})

	t.Run("empty_not_ok", func(t *testing.T) {
		_, _, _, ok := parsePosition("")
		if ok {
			t.Error("expected ok=false for empty string")
		}
	})

	t.Run("whitespace_trimmed", func(t *testing.T) {
		_, _, _, ok := parsePosition("  10.0,20.0  ")
		if !ok {
			t.Error("expected ok=true after trim")
		}
	})
}

// TestStep1 ensures the %.1f formatter matches C reference output.
func TestStep1(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0.0"},
		{360, "360.0"},
		{180, "180.0"},
		{-180, "-180.0"},
		{1.5, "1.5"},
	}
	for _, tc := range cases {
		got := step1(tc.in)
		if got != tc.want {
			t.Errorf("step1(%v): want %q got %q", tc.in, tc.want, got)
		}
	}
}

// TestPtzStepReplacements verifies the -%MAX_X% pattern used in GetNodes template.
func TestPtzStepReplacements(t *testing.T) {
	svc := &ServiceContext{
		Node: Node{MinX: 0, MaxX: 360, MinY: 0, MaxY: 180, MinZ: 0, MaxZ: 0},
	}
	r := svc.ptzStepReplacements()
	if r["%MAX_X%"] != "360.0" {
		t.Errorf("MAX_X: want 360.0 got %s", r["%MAX_X%"])
	}
	// Simulate template substitution: "-%MAX_X%" becomes "-360.0"
	tmpl := "-%MAX_X%"
	for k, v := range r {
		tmpl = strings.ReplaceAll(tmpl, k, v)
	}
	if tmpl != "-360.0" {
		t.Errorf("negative range substitution: want -360.0 got %s", tmpl)
	}
}
