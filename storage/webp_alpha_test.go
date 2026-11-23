package storage

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestWebPAlphaPreservedOnImport(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	fixture, err := root.ReadFile("alpha.webp")
	if err != nil {
		t.Fatal(err)
	}
	// The source WebP contains fully transparent and semi-transparent pixels.
	original, _, err := image.Decode(bytes.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, a := original.At(0, 0).RGBA(); a != 0 {
		t.Fatal("fixture is not transparent")
	}
	if _, _, _, a := original.At(4, 4).RGBA(); a == 0 || a == 0xffff {
		t.Fatal("fixture has no partial alpha")
	}

	filePath := filepath.Join(t.TempDir(), "alpha.webp")
	if err := os.WriteFile(filePath, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	for name, importFn := range map[string]func() (string, error){
		"file":      func() (string, error) { return ImportImage(filePath, "file-alpha") },
		"clipboard": func() (string, error) { return ImportImageBytes(fixture, "clipboard-alpha") },
	} {
		t.Run(name, func(t *testing.T) {
			storedPath, err := importFn()
			if err != nil {
				t.Fatal(err)
			}
			f, err := OpenFileAtPath(storedPath)
			if err != nil {
				t.Fatal(err)
			}
			stored, _, decodeErr := image.Decode(f)
			closeErr := f.Close()
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			if _, _, _, a := stored.At(0, 0).RGBA(); a != 0 {
				t.Fatalf("transparent pixel flattened: alpha=%d", a)
			}
			// The imported thumbnail is 200x200 and the 8x8 source fills it.
			if _, _, _, a := stored.At(112, 112).RGBA(); a == 0 || a == 0xffff {
				t.Fatalf("partial alpha flattened: alpha=%d", a)
			}
		})
	}
}
