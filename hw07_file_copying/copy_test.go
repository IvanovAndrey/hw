package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func createTempFileWithContent(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

func readFileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(data)
}

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "Hello, this is test content!")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := Copy(srcPath, dstPath, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := readFileContent(t, dstPath)
	want := "Hello, this is test content!"
	if got != want {
		t.Errorf("unexpected copy result: got %q, want %q", got, want)
	}
}

func TestCopyWithOffsetAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "abcdef123456")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := Copy(srcPath, dstPath, 3, 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := readFileContent(t, dstPath)
	want := "def1"
	if got != want {
		t.Errorf("unexpected copy result: got %q, want %q", got, want)
	}
}

func TestCopyOffsetExceedsFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "abc")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := Copy(srcPath, dstPath, 10, 0)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Fatalf("expected ErrOffsetExceedsFileSize, got %v", err)
	}
}

func TestCopyUnsupportedFile(t *testing.T) {
	tmpDir := t.TempDir()
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := Copy(tmpDir, dstPath, 0, 0)
	if !errors.Is(err, ErrUnsupportedFile) {
		t.Fatalf("expected ErrUnsupportedFile, got %v", err)
	}
}

func TestCopyLimitExceedsEOF(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "abc")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := Copy(srcPath, dstPath, 1, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := readFileContent(t, dstPath)
	want := "bc"
	if got != want {
		t.Errorf("unexpected copy result: got %q, want %q", got, want)
	}
}

func TestCopyCreatesDestinationFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "data")
	dstPath := filepath.Join(tmpDir, "created.txt")

	err := Copy(srcPath, dstPath, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("expected destination file to be created")
	}
}

func TestCopyTruncatesDestinationFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := createTempFileWithContent(t, tmpDir, "source.txt", "123")
	dstPath := createTempFileWithContent(t, tmpDir, "dest.txt", "123456789")

	err := Copy(srcPath, dstPath, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := readFileContent(t, dstPath)
	want := "123"
	if got != want {
		t.Errorf("expected destination file to be truncated: got %q, want %q", got, want)
	}
}
