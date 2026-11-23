package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFileAtPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "example.txt")
	if err := os.WriteFile(file, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := readFileAtPath(file)
	if err != nil || string(got) != "hello" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := OpenFileAtPath(""); err == nil {
		t.Fatal("empty path accepted")
	}
	if _, err := OpenFileAtPath(filepath.Join(dir, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected not exist, got %v", err)
	}
}

func TestReadFileRejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "large-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(32<<20 + 1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readFileAtPath(f.Name()); err == nil {
		t.Fatal("oversized file accepted")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := SaveSettings(&Settings{ThemeName: "Тёмная"}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings()
	if err != nil || got.ThemeName != "Тёмная" {
		t.Fatalf("got %+v, %v", got, err)
	}
}
