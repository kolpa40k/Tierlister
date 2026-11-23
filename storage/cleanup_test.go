package storage

import (
	"os"
	"path/filepath"
	"testing"

	"tierlist-app/models"
)

func TestCleanupImages(t *testing.T) {
	dir := t.TempDir()
	create := func(name string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	usedPool := create("pool.png")
	usedTier := create("tier.png")
	orphan := create("orphan.png")
	other := create("notes.txt")
	categories := []models.Category{{TierLists: []models.TierList{{
		Pool:  []models.TierItem{{ImagePath: usedPool}},
		Tiers: []models.Tier{{Items: []models.TierItem{{ImagePath: usedTier}}}},
	}}}}
	preview, err := cleanupImagesAt(dir, categories, true)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Used != 2 || preview.Orphaned != 1 || preview.Deleted != 0 {
		t.Fatalf("preview: %+v", preview)
	}
	if _, err := os.Stat(orphan); err != nil {
		t.Fatal("dry run deleted image", err)
	}
	report, err := cleanupImagesAt(dir, categories, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Deleted != 1 || report.BytesReclaimed != 4 {
		t.Fatalf("cleanup: %+v", report)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("orphan not deleted: %v", err)
	}
	for _, path := range []string{usedPool, usedTier, other} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("retained %s: %v", path, err)
		}
	}
}

func TestCleanupImagesSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	external := filepath.Join(t.TempDir(), "external.png")
	if err := os.WriteFile(external, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(external, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	report, err := cleanupImagesAt(dir, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Deleted != 0 {
		t.Fatalf("deleted symlink: %+v", report)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(external); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupImagesMissingDirectory(t *testing.T) {
	report, err := cleanupImagesAt(filepath.Join(t.TempDir(), "missing"), nil, false)
	if err != nil || report != (CleanupReport{}) {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}
