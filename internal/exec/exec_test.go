package exec

import (
	"strings"
	"testing"
	"time"
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

func TestRunFmt_MultiArg(t *testing.T) {
	// Exercises the multi-arg shape used by set_preset (%d + %s) and jump_to_abs (%f,%f,%f).
	if err := RunFmt("/bin/echo %d %s", 42, "hello"); err != nil {
		t.Fatalf("RunFmt(int+string): %v", err)
	}
	if err := RunFmt("/bin/echo %f %f %f", 1.0, 2.0, 3.0); err != nil {
		t.Fatalf("RunFmt(three floats): %v", err)
	}
}

func TestRun_Timeout(t *testing.T) {
	orig := DefaultTimeout
	DefaultTimeout = 50 * time.Millisecond
	defer func() { DefaultTimeout = orig }()

	if err := Run("/bin/sleep 10"); err == nil {
		t.Fatal("expected timeout error from /bin/sleep 10")
	}
}

func TestOutput_Timeout(t *testing.T) {
	orig := DefaultTimeout
	DefaultTimeout = 50 * time.Millisecond
	defer func() { DefaultTimeout = orig }()

	if _, err := Output("/bin/sleep 10"); err == nil {
		t.Fatal("expected timeout error from /bin/sleep 10")
	}
}

func TestRun_MetacharactersAreInert(t *testing.T) {
	// Without a shell, these metacharacters become literal argv tokens.
	// /bin/echo exits 0 regardless; the test confirms no shell escape occurs.
	corpus := []string{
		"/bin/echo ;",
		"/bin/echo $(id)",
		"/bin/echo `id`",
		"/bin/echo &&id",
		"/bin/echo |cat",
		"/bin/echo >>/etc/passwd",
	}
	for _, cmd := range corpus {
		if err := Run(cmd); err != nil {
			t.Errorf("Run(%q): unexpected error %v", cmd, err)
		}
	}
}

func TestOutput_MetacharactersAreLiteral(t *testing.T) {
	// $(id) must appear verbatim in output — not shell-expanded.
	out, err := Output("/bin/echo $(id)")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if out != "$(id)" {
		t.Errorf("got %q; expected literal $(id), not shell-expanded", out)
	}
}
