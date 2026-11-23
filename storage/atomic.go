package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// writeJSONAtomic replaces a JSON document only after the new content is durable.
func writeJSONAtomic(path string, value any) (err error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	file, err := os.CreateTemp(dir, ".tierlist-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.Remove(file.Name())
		}
	}()
	if err = file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	// Best effort: syncing directories is unsupported on some platforms.
	if parent, openErr := os.OpenRoot(filepath.Dir(dir)); openErr == nil {
		if d, dirErr := parent.Open(filepath.Base(dir)); dirErr == nil {
			_ = errors.Join(d.Sync(), d.Close())
		}
		_ = parent.Close()
	}
	return nil
}

func readJSON(path string, out any) error {
	data, err := readFileAtPath(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}
