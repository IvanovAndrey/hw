package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %q is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", dir, err)
	}
	if len(entries) == 0 {
		return nil, nil
	}

	env := make(Environment)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("invalid filename %q: contains '='", name)
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", path, err)
		}

		if len(data) == 0 {
			env[name] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		firstLine := ""
		if scanner.Scan() {
			firstLine = scanner.Text()
		}
		firstLine = strings.ReplaceAll(firstLine, "\x00", "\n")
		firstLine = strings.TrimRight(firstLine, " \t")

		env[name] = EnvValue{
			Value:      firstLine,
			NeedRemove: false,
		}
	}

	return env, nil
}
