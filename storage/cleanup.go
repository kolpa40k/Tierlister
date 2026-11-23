package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tierlist-app/models"
)

// CleanupReport describes owned image files found and removed. Unrelated files
// and symbolic links are never touched.
type CleanupReport struct {
	Files          int
	Used           int
	Orphaned       int
	Deleted        int
	BytesReclaimed int64
}

// CleanupImages removes unreferenced application-owned PNGs. It never deletes
// a file referenced by any category, tier, or pool. Set dryRun to preview.
func CleanupImages(categories []models.Category, dryRun bool) (CleanupReport, error) {
	dir, err := imagesDir()
	if err != nil {
		return CleanupReport{}, err
	}
	return cleanupImagesAt(dir, categories, dryRun)
}

func cleanupImagesAt(dir string, categories []models.Category, dryRun bool) (CleanupReport, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return CleanupReport{}, nil
	}
	if err != nil {
		return CleanupReport{}, fmt.Errorf("list image directory: %w", err)
	}
	references := imageReferences(dir, categories)
	var report CleanupReport
	var failures []error
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		if err := processImageEntry(dir, entry, references, dryRun, &report); err != nil {
			failures = append(failures, err)
		}
	}
	return report, errors.Join(failures...)
}

func imageReferences(dir string, categories []models.Category) map[string]struct{} {
	references := make(map[string]struct{})
	add := func(items []models.TierItem) {
		for _, item := range items {
			if item.ImagePath != "" && isInside(item.ImagePath, dir) {
				references[filepath.Clean(item.ImagePath)] = struct{}{}
			}
		}
	}
	for _, category := range categories {
		for _, list := range category.TierLists {
			add(list.Pool)
			for _, tier := range list.Tiers {
				add(tier.Items)
			}
		}
	}
	return references
}

func processImageEntry(dir string, entry os.DirEntry, references map[string]struct{}, dryRun bool, report *CleanupReport) error {
	info, err := entry.Info()
	if err != nil {
		return fmt.Errorf("stat %s: %w", entry.Name(), err)
	}
	path := filepath.Join(dir, entry.Name())
	report.Files++
	if _, used := references[filepath.Clean(path)]; used {
		report.Used++
		return nil
	}
	report.Orphaned++
	if dryRun {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %s: %w", entry.Name(), err)
	}
	report.Deleted++
	report.BytesReclaimed += info.Size()
	return nil
}
