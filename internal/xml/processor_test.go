package xml

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func writeXMLFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func writeXMLGz(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	gw := gzip.NewWriter(f)
	if _, err := gw.Write([]byte(content)); err != nil {
		_ = f.Close()
		t.Fatalf("gzip write %s: %v", name, err)
	}
	if err := gw.Close(); err != nil {
		_ = f.Close()
		t.Fatalf("gzip close %s: %v", name, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", name, err)
	}
	return p
}

func TestProcessTemplate_PlainText(t *testing.T) {
	dir := t.TempDir()
	p := writeXMLFile(t, dir, "tpl.xml", "<resp>%VALUE%</resp>")
	got, err := ProcessTemplate(p, map[string]string{"%VALUE%": "hello"})
	if err != nil {
		t.Fatalf("ProcessTemplate: %v", err)
	}
	if want := "<resp>hello</resp>\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestProcessTemplate_GzipDirect(t *testing.T) {
	dir := t.TempDir()
	p := writeXMLGz(t, dir, "tpl.xml.gz", "<resp>%VALUE%</resp>")
	got, err := ProcessTemplate(p, map[string]string{"%VALUE%": "world"})
	if err != nil {
		t.Fatalf("ProcessTemplate(.gz direct): %v", err)
	}
	if want := "<resp>world</resp>\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestProcessTemplate_GzipFallback(t *testing.T) {
	dir := t.TempDir()
	writeXMLGz(t, dir, "tpl.xml.gz", "<resp>%VALUE%</resp>")
	// Ask for .xml but only .xml.gz exists — should fall back transparently.
	got, err := ProcessTemplate(filepath.Join(dir, "tpl.xml"), map[string]string{"%VALUE%": "fallback"})
	if err != nil {
		t.Fatalf("ProcessTemplate(.gz fallback): %v", err)
	}
	if want := "<resp>fallback</resp>\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestProcessTemplate_PlainPreferredOverGzip(t *testing.T) {
	dir := t.TempDir()
	writeXMLFile(t, dir, "tpl.xml", "<resp>plain</resp>")
	writeXMLGz(t, dir, "tpl.xml.gz", "<resp>gzip</resp>")
	got, err := ProcessTemplate(filepath.Join(dir, "tpl.xml"), nil)
	if err != nil {
		t.Fatalf("ProcessTemplate: %v", err)
	}
	if want := "<resp>plain</resp>\n"; got != want {
		t.Errorf("plain file should be preferred over .gz; got %q", got)
	}
}

func TestProcessTemplate_MissingFile(t *testing.T) {
	_, err := ProcessTemplate("/nonexistent/path/tpl.xml", nil)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFileExists_PlainFile(t *testing.T) {
	dir := t.TempDir()
	p := writeXMLFile(t, dir, "tpl.xml", "content")
	if !FileExists(p) {
		t.Error("FileExists should return true for existing plain file")
	}
}

func TestFileExists_GzipFallback(t *testing.T) {
	dir := t.TempDir()
	writeXMLGz(t, dir, "tpl.xml.gz", "content")
	if !FileExists(filepath.Join(dir, "tpl.xml")) {
		t.Error("FileExists should return true when only .gz exists")
	}
}

func TestFileExists_Missing(t *testing.T) {
	if FileExists("/nonexistent/path/tpl.xml") {
		t.Error("FileExists should return false for missing file and missing .gz")
	}
}

func TestProcessTemplate_CorruptGzip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.xml.gz")
	if err := os.WriteFile(p, []byte("not a valid gzip stream"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := ProcessTemplate(p, nil); err == nil {
		t.Error("expected error for corrupt gzip header")
	}
}

func TestProcessTemplate_EmptyGzip(t *testing.T) {
	dir := t.TempDir()
	writeXMLGz(t, dir, "empty.xml.gz", "")
	got, err := ProcessTemplate(filepath.Join(dir, "empty.xml"), nil)
	if err != nil {
		t.Fatalf("empty gzip: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestProcessTemplate_MultiLinePlaceholders(t *testing.T) {
	dir := t.TempDir()
	writeXMLGz(t, dir, "multi.xml.gz", "line1 %A%\nline2 %B%\n")
	got, err := ProcessTemplate(filepath.Join(dir, "multi.xml"), map[string]string{
		"%A%": "alpha",
		"%B%": "beta",
	})
	if err != nil {
		t.Fatalf("ProcessTemplate multi: %v", err)
	}
	if want := "line1 alpha\nline2 beta\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
