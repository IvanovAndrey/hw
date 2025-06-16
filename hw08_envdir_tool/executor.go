package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func isCmdValid(cmd []string) []string {
	for _, arg := range cmd {
		for _, ch := range arg {
			if !unicode.IsPrint(ch) || ch == ';' || ch == '&' || ch == '|' {
				return nil
			}
		}
	}
	return cmd
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	validCmd := isCmdValid(cmd)
	if validCmd == nil {
		return 1
	}

	finalEnv := make([]string, 0, len(os.Environ())+len(env))
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		key := parts[0]

		if val, ok := env[key]; ok {
			if val.NeedRemove {
				continue
			}
			finalEnv = append(finalEnv, key+"="+val.Value)
			delete(env, key)
		} else {
			finalEnv = append(finalEnv, e)
		}
	}

	for k, v := range env {
		if !v.NeedRemove {
			finalEnv = append(finalEnv, k+"="+v.Value)
		}
	}

	c := exec.Command(validCmd[0], validCmd[1:]...) // #nosec G204
	c.Env = finalEnv
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 127
	}

	return 0
}
