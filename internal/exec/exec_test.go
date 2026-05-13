package exec

import (
	"strings"
	"testing"
)

func TestRun_OK(t *testing.T) {
	if err := Run("/bin/echo hello"); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunWithFloat_OK(t *testing.T) {
	if err := RunWithFloat("/bin/echo %f", 1.5); err != nil {
		t.Fatalf("RunWithFloat: %v", err)
	}
}

func TestOutput_OK(t *testing.T) {
	out, err := Output("/bin/echo hello world")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if out != "hello world" {
		t.Errorf("got %q; want %q", out, "hello world")
	}
}

func TestOutput_TrimSpace(t *testing.T) {
	out, err := Output("/bin/echo")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("output not trimmed, got %q", out)
	}
}

func TestRun_EmptyCommand(t *testing.T) {
	if err := Run(""); err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestRun_RelativePath(t *testing.T) {
	if err := Run("echo hello"); err == nil {
		t.Fatal("expected error for relative argv[0]")
	}
}

func TestOutput_RelativePath(t *testing.T) {
	if _, err := Output("echo hello"); err == nil {
		t.Fatal("expected error for relative argv[0]")
	}
}

func TestOutput_CommandFails(t *testing.T) {
	if _, err := Output("/bin/false"); err == nil {
		t.Fatal("expected error from /bin/false")
	}
}

func TestRun_CommandFails(t *testing.T) {
	if err := Run("/bin/false"); err == nil {
		t.Fatal("expected error from /bin/false")
	}
}

func TestRunWithFloat_RelativePath(t *testing.T) {
	if err := RunWithFloat("echo %f", 1.0); err == nil {
		t.Fatal("expected error for relative argv[0] after Sprintf")
	}
}
