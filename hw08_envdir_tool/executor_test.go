package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

const script = `#!/bin/sh
echo "VAR1=$VAR1"
echo "VAR2=$VAR2"
echo "ARGS=$@"
[ "$REMOVE_ME" = "" ] && exit 0 || exit 42
`

func writeTempScript(t *testing.T) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-script-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(tmpFile.Name(), 0o755); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

func TestRunCmd(t *testing.T) {
	scriptPath := writeTempScript(t)

	os.Setenv("REMOVE_ME", "bye")

	env := Environment{
		"VAR1":      {Value: "hello", NeedRemove: false},
		"VAR2":      {Value: "world", NeedRemove: false},
		"REMOVE_ME": {Value: "", NeedRemove: true},
	}

	cmd := []string{scriptPath, "arg1", "arg2"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout := os.Stdout
	os.Stdout = w

	code := RunCmd(cmd, env)

	_ = w.Close()
	os.Stdout = stdout

	out, _ := io.ReadAll(r)
	output := string(out)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "VAR1=hello") {
		t.Errorf("expected VAR1=hello, got: %s", output)
	}
	if !strings.Contains(output, "VAR2=world") {
		t.Errorf("expected VAR2=world, got: %s", output)
	}
	if strings.Contains(output, "REMOVE_ME") {
		t.Errorf("expected REMOVE_ME to be unset, got: %s", output)
	}
	if !strings.Contains(output, "ARGS=arg1 arg2") {
		t.Errorf("expected args to be passed, got: %s", output)
	}
}

func TestRunCmd_CommandNotFound(t *testing.T) {
	env := Environment{}
	cmd := []string{"/nonexistent/command"}
	code := RunCmd(cmd, env)
	if code != 127 {
		t.Errorf("expected exit code 127 for command not found, got %d", code)
	}
}

func TestRunCmd_NoCommand(t *testing.T) {
	env := Environment{}
	var cmd []string
	code := RunCmd(cmd, env)
	if code != 1 {
		t.Errorf("expected exit code 1 for empty command, got %d", code)
	}
}
