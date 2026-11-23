package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tierlist-app/models"
)

func imagesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tierlist-app", "images"), nil
}

func isInside(path, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absDir, absPath)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func migrateItem(item *models.TierItem, dir string) (bool, error) {
	if item.ImagePath == "" || item.ID == "" || isInside(item.ImagePath, dir) {
		return false, nil
	}
	if _, err := os.Stat(item.ImagePath); err != nil {
		return false, fmt.Errorf("source image %q: %w", item.ImagePath, err)
	}
	newPath, err := ImportImage(item.ImagePath, item.ID)
	if err != nil {
		return false, fmt.Errorf("migrate image %q: %w", item.ImagePath, err)
	}
	item.ImagePath = newPath
	return true, nil
}

func MigrateAllImages(categories []models.Category) (bool, error) {
	dir, err := imagesDir()
	if err != nil {
		return false, err
	}
	changed := false
	var failures []error
	visit := func(item *models.TierItem) {
		updated, err := migrateItem(item, dir)
		changed = changed || updated
		if err != nil {
			failures = append(failures, err)
		}
	}
	for ci := range categories {
		for ti := range categories[ci].TierLists {
			list := &categories[ci].TierLists[ti]
			for tier := range list.Tiers {
				for i := range list.Tiers[tier].Items {
					visit(&list.Tiers[tier].Items[i])
				}
			}
			for i := range list.Pool {
				visit(&list.Pool[i])
			}
		}
	}
	return changed, errors.Join(failures...)
}
