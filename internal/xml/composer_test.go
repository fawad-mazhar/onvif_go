package xml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompose(t *testing.T) {
	got := Compose("H", []string{"a", "b", "c"}, "F")
	if got != "HabcF" {
		t.Errorf("Compose = %q; want HabcF", got)
	}
	if got := Compose("H", nil, "F"); got != "HF" {
		t.Errorf("Compose with no items = %q; want HF", got)
	}
}

func writeTempTemplate(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestComposeFromFiles_Items(t *testing.T) {
	dir := t.TempDir()
	hdr := writeTempTemplate(t, dir, "h.xml", "<List service=\"%SVC%\">\n")
	mid := writeTempTemplate(t, dir, "m.xml", "  <Item token=\"%TOKEN%\">%NAME%</Item>\n")
	ftr := writeTempTemplate(t, dir, "f.xml", "</List>\n")

	out, err := ComposeFromFiles(hdr, mid, ftr, "",
		map[string]string{"%SVC%": "media"},
		[]map[string]string{
			{"%TOKEN%": "t0", "%NAME%": "first"},
			{"%TOKEN%": "t1", "%NAME%": "second"},
		},
	)
	if err != nil {
		t.Fatalf("ComposeFromFiles: %v", err)
	}
	for _, want := range []string{
		`<List service="media">`,
		`<Item token="t0">first</Item>`,
		`<Item token="t1">second</Item>`,
		`</List>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestComposeFromFiles_NoneTemplate(t *testing.T) {
	dir := t.TempDir()
	hdr := writeTempTemplate(t, dir, "h.xml", "<List>")
	mid := writeTempTemplate(t, dir, "m.xml", "<Item/>")
	ftr := writeTempTemplate(t, dir, "f.xml", "</List>")
	non := writeTempTemplate(t, dir, "n.xml", "<EmptyResponse for=\"%SVC%\"/>")

	out, err := ComposeFromFiles(hdr, mid, ftr, non,
		map[string]string{"%SVC%": "device"},
		nil,
	)
	if err != nil {
		t.Fatalf("ComposeFromFiles: %v", err)
	}
	if !strings.Contains(out, `<EmptyResponse for="device"/>`) {
		t.Errorf("empty-list path did not use none template: %q", out)
	}
}

func TestComposeFromFiles_EmptyWithoutNone(t *testing.T) {
	dir := t.TempDir()
	hdr := writeTempTemplate(t, dir, "h.xml", "H")
	mid := writeTempTemplate(t, dir, "m.xml", "M")
	ftr := writeTempTemplate(t, dir, "f.xml", "F")

	out, err := ComposeFromFiles(hdr, mid, ftr, "", nil, nil)
	if err != nil {
		t.Fatalf("ComposeFromFiles: %v", err)
	}
	// ProcessTemplate terminates each source line with \n, so single-line
	// "H" and "F" templates render as "H\n" and "F\n" respectively.
	if out != "H\nF\n" {
		t.Errorf("empty-list without none = %q; want %q", out, "H\nF\n")
	}
}

func TestComposeFromFiles_PerItemOverridesGlobal(t *testing.T) {
	dir := t.TempDir()
	hdr := writeTempTemplate(t, dir, "h.xml", "")
	mid := writeTempTemplate(t, dir, "m.xml", "[%X%]")
	ftr := writeTempTemplate(t, dir, "f.xml", "")

	out, err := ComposeFromFiles(hdr, mid, ftr, "",
		map[string]string{"%X%": "global"},
		[]map[string]string{
			{},               // uses global
			{"%X%": "local"}, // overrides
		},
	)
	if err != nil {
		t.Fatalf("ComposeFromFiles: %v", err)
	}
	want := "[global][local]\n\n" // ProcessTemplate appends \n per line; header+footer empty lines trail
	if !strings.Contains(out, "[global]") || !strings.Contains(out, "[local]") {
		t.Errorf("got %q; want substrings %q", out, want)
	}
}

func TestServiceTemplatePath(t *testing.T) {
	prev := ServiceTemplateDir
	ServiceTemplateDir = "sf"
	defer func() { ServiceTemplateDir = prev }()

	if got := ServiceTemplatePath("media", "GetProfiles.xml"); got != filepath.Join("sf", "media", "GetProfiles.xml") {
		t.Errorf("ServiceTemplatePath = %q", got)
	}
}
