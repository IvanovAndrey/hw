package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/schollz/progressbar/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	file, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer closeFile(file)

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return ErrUnsupportedFile
	}

	if info.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	if limit == 0 || limit > info.Size()-offset {
		limit = info.Size() - offset
	}

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	outFile, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer closeFile(outFile)

	bar := progressbar.DefaultBytes(
		limit,
		fmt.Sprintf("Copying %s → %s", fromPath, toPath),
	)

	writer := io.MultiWriter(outFile, bar)

	_, err = io.CopyN(writer, file, limit)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy failed: %w", err)
	}

	return nil
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatalf("failed to close file: %v", err)
	}
}
