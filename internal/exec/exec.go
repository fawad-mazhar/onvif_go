// Package exec provides a secure shell-command runner for PTZ and relay
// operations. Commands are tokenized by whitespace and executed directly via
// os/exec — no shell is invoked. argv[0] must be an absolute path.
//
// This mirrors the semantics of the C reference server's use of system() and
// popen() for PTZ commands while eliminating shell-injection risk:
//
//	C: sprintf(sys_command, cmd_template, step); system(sys_command);
//	Go: RunWithFloat(cmd_template, step)
//
//	C: fp = popen(cmd, "r"); fgets(out, ...);
//	Go: out, err := Output(cmd)
//
// All three public functions are bounded by DefaultTimeout so a hanging script
// (firmware bug, network-mount glitch) cannot block the calling goroutine until
// chi's 60-second request timeout fires.
package exec

import (
	"bytes"
	"context"
	"fmt"
	goexec "os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the per-command execution deadline applied by Run, RunFmt,
// RunWithFloat, and Output. Override in tests with a shorter value.
var DefaultTimeout = 5 * time.Second

// tokenize splits cmdStr into argv by whitespace and validates that argv[0]
// is an absolute path. Returns an error for empty strings and relative paths.
func tokenize(cmdStr string) ([]string, error) {
	args := strings.Fields(cmdStr)
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command string")
	}
	if !strings.HasPrefix(args[0], "/") {
		return nil, fmt.Errorf("argv[0] must be an absolute path, got: %q", args[0])
	}
	return args, nil
}

// Run executes cmdStr as a fire-and-forget process (stdout/stderr not captured).
// cmdStr is split by whitespace; argv[0] must be an absolute path.
// No shell is invoked, eliminating injection risk vs C's system().
// Execution is bounded by DefaultTimeout.
func Run(cmdStr string) error {
	args, err := tokenize(cmdStr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return goexec.CommandContext(ctx, args[0], args[1:]...).Run() //nolint:gosec // PTZ command is user-configured; gated by validateSOAPArg
}

// RunFmt formats template via fmt.Sprintf(template, args...) and executes the
// result the same way as Run. This is the general multi-argument variant for
// PTZ commands whose C templates contain mixed printf verbs:
//
//	set_preset  = /usr/local/bin/ptz_presets.sh -a add_preset -n %d -m %s
//	jump_to_abs = /usr/local/bin/ptz_move -j %f,%f,%f
//
// Warning: if template contains no format verbs and args is non-empty,
// fmt.Sprintf appends %!(EXTRA ...) to the string, making argv[0] invalid.
// Ensure the template format verbs match the supplied arguments.
func RunFmt(template string, args ...any) error {
	return Run(fmt.Sprintf(template, args...))
}

// RunWithFloat substitutes a single %f placeholder in cmdTemplate and executes
// the result. Delegates to RunFmt; kept for backward compatibility.
// Mirrors the C pattern used for PTZ move commands:
//
//	spprintf(sys_command, cmd_template, dx); system(sys_command);
func RunWithFloat(cmdTemplate string, arg float64) error {
	return RunFmt(cmdTemplate, arg)
}

// Output executes cmdStr and returns its trimmed standard output.
// Mirrors C's popen(cmd, "r") + fread pattern used for get_position / is_moving
// / get_presets — commands that return their result to stdout.
// Execution is bounded by DefaultTimeout. stderr is not captured (see G-010
// for a future OutputCombined variant). stdout is read unbounded — scripts
// are expected to return short single-line values (see G-010).
func Output(cmdStr string) (string, error) {
	args, err := tokenize(cmdStr)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	cmd := goexec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // PTZ command is user-configured; gated by validateSOAPArg
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
