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
package exec

import (
	"bytes"
	"fmt"
	goexec "os/exec"
	"strings"
)

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
func Run(cmdStr string) error {
	args, err := tokenize(cmdStr)
	if err != nil {
		return err
	}
	return goexec.Command(args[0], args[1:]...).Run()
}

// RunWithFloat substitutes a single printf-style placeholder in cmdTemplate via
// fmt.Sprintf(cmdTemplate, arg), then executes the result the same way as Run.
// Mirrors the C pattern used for PTZ move commands:
//
//	sprintf(sys_command, cmd_template, dx); system(sys_command);
func RunWithFloat(cmdTemplate string, arg float64) error {
	return Run(fmt.Sprintf(cmdTemplate, arg))
}

// Output executes cmdStr and returns its trimmed standard output.
// Mirrors C's popen(cmd, "r") + fread pattern used for get_position / is_moving
// / get_presets — commands that return their result to stdout.
func Output(cmdStr string) (string, error) {
	args, err := tokenize(cmdStr)
	if err != nil {
		return "", err
	}
	cmd := goexec.Command(args[0], args[1:]...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
