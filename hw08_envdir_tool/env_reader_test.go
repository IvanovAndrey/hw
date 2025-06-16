package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", name, err)
	}
}

func setupDir(t *testing.T, fileName, content string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, fileName, content)
	return dir
}

func TestReadDir(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    string
		expectedEnv Environment
		expectError bool
	}{
		{
			name: "Not a directory",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tmpFile, err := os.CreateTemp("", "not-a-dir")
				if err != nil {
					t.Fatal(err)
				}
				tmpFile.Close()
				return tmpFile.Name()
			}(t),
			expectedEnv: nil,
			expectError: true,
		},
		{
			name:        "Empty directory",
			setupDir:    t.TempDir(),
			expectedEnv: nil,
			expectError: false,
		},
		{
			name:     "Valid single file",
			setupDir: setupDir(t, "FOO", "bar\nbaz\nqux"),
			expectedEnv: Environment{
				"FOO": {Value: "bar", NeedRemove: false},
			},
			expectError: false,
		},
		{
			name:     "Empty file",
			setupDir: setupDir(t, "EMPTY", ""),
			expectedEnv: Environment{
				"EMPTY": {Value: "", NeedRemove: true},
			},
			expectError: false,
		},
		{
			name:        "Invalid filename with '='",
			setupDir:    setupDir(t, "BAD=NAME", "data"),
			expectedEnv: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir
			env, err := ReadDir(dir)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(env) != len(tt.expectedEnv) {
				t.Errorf("expected %d env variables, got %d", len(tt.expectedEnv), len(env))
			}

			for k, v := range tt.expectedEnv {
				got, ok := env[k]
				if !ok {
					t.Errorf("expected key %s in env", k)
					continue
				}
				if got.Value != v.Value {
					t.Errorf("expected value for %s: %q, got %q", k, v.Value, got.Value)
				}
				if got.NeedRemove != v.NeedRemove {
					t.Errorf("expected NeedRemove for %s: %v, got %v", k, v.NeedRemove, got.NeedRemove)
				}
			}
		})
	}
}
