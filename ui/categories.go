package ui

import (
	"fmt"
	"tierlist-app/models"
	"tierlist-app/storage"
)

// persistCategories makes changes to a detached snapshot and publishes them
// only after successful persistence.
func persistCategories(current []models.Category, mutate func([]models.Category) ([]models.Category, error)) ([]models.Category, error) {
	next := cloneCategories(current)
	next, err := mutate(next)
	if err != nil {
		return nil, err
	}
	if err := storage.SaveCategories(next); err != nil {
		return nil, fmt.Errorf("save categories: %w", err)
	}
	return next, nil
}
