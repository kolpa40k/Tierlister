package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// OpenFileAtPath opens a single file inside its parent directory, without
// allowing the final path component to escape that directory.
func OpenFileAtPath(path string) (*os.File, error) {
	if path == "" {
		return nil, errors.New("empty file path")
	}
	parent, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("open parent directory: %w", err)
	}
	file, openErr := parent.Open(filepath.Base(path))
	closeErr := parent.Close()
	if openErr != nil {
		return nil, fmt.Errorf("open file: %w", openErr)
	}
	if closeErr != nil {
		_ = file.Close()
		return nil, fmt.Errorf("close parent directory: %w", closeErr)
	}
	return file, nil
}

func readFileAtPath(path string) ([]byte, error) {
	file, err := OpenFileAtPath(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, 32<<20+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(data) > 32<<20 {
		return nil, errors.New("file exceeds 32 MiB")
	}
	return data, nil
}

// ReadImageFile reads a locally selected or previously imported image.
func ReadImageFile(path string) ([]byte, error) { return readFileAtPath(path) }
