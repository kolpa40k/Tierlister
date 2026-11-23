package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestImportImageBytes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	src := image.NewRGBA(image.Rect(0, 0, 32, 16))
	src.Set(1, 1, color.RGBA{R: 255, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, src); err != nil {
		t.Fatal(err)
	}
	path, err := ImportImageBytes(data.Bytes(), "clipboard-test")
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	f, err := root.Open(filepath.Base(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 200 || img.Bounds().Dy() != 200 {
		t.Fatalf("unexpected size: %v", img.Bounds())
	}
}

func TestImportImageBytesRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		data     []byte
	}{
		{name: "empty", id: "good"},
		{name: "invalid image", id: "good", data: []byte("not an image")},
		{name: "unsafe id", id: "../bad", data: []byte("not an image")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ImportImageBytes(tc.data, tc.id); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
