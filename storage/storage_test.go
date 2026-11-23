package storage

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"tierlist-app/models"
)

func TestCategoriesRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	want := []models.Category{{Name: "Movies", TierLists: []models.TierList{models.NewTierList("Favorites")}}}
	if err := SaveCategories(want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCategories()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != want[0].Name || len(got[0].TierLists) != 1 {
		t.Fatalf("unexpected categories: %+v", got)
	}
	path, err := dataPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0077 != 0 {
		t.Fatalf("private data has permissions %o", info.Mode().Perm())
	}
}

func TestCorruptedCategoriesAreNotOverwrittenOnLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := dataPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCategories(); err == nil {
		t.Fatal("expected decode error")
	}
	data, err := readFileAtPath(path)
	if err != nil || string(data) != "{broken" {
		t.Fatalf("original file changed: %q (%v)", data, err)
	}
}

func TestImportImage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sourceDir := t.TempDir()
	source := filepath.Join(sourceDir, "source.png")
	f, err := os.CreateTemp(filepath.Dir(source), "source-*.png")
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	source = f.Name()
	path, err := ImportImage(source, "test-id")
	if err != nil {
		t.Fatal(err)
	}
	imported, err := OpenFileAtPath(path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(imported)
	if err != nil {
		t.Fatal(err)
	}
	if err := imported.Close(); err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 200 || decoded.Bounds().Dy() != 200 {
		t.Fatalf("wrong size: %v", decoded.Bounds())
	}
	if _, err := ImportImage(source, "../escape"); err == nil {
		t.Fatal("expected invalid ID error")
	}
}

func TestIsInside(t *testing.T) {
	dir := t.TempDir()
	if !isInside(filepath.Join(dir, "image.png"), dir) {
		t.Fatal("expected inside")
	}
	if isInside(filepath.Join(dir, "..", "outside.png"), dir) {
		t.Fatal("path traversal accepted")
	}
}
