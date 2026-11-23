package ui

import (
	"errors"
	"tierlist-app/editor"
	"tierlist-app/models"
	"tierlist-app/storage"

	"fyne.io/fyne/v2"
)

// change applies a domain operation to a detached copy and publishes it only
// after the entire category document has been durably persisted.
func (ts *TierListScreen) change(operation func(*models.TierList) error) bool {
	next := editor.Clone(ts.tierList)
	if err := operation(&next); err != nil {
		ts.showError(err)
		return false
	}
	if err := editor.Validate(next); err != nil {
		ts.showError(err)
		return false
	}
	categories := cloneCategories(*ts.categories)
	if ts.catIdx < 0 || ts.catIdx >= len(categories) || ts.tlIdx < 0 || ts.tlIdx >= len(categories[ts.catIdx].TierLists) {
		ts.showError(errors.New("selected tier list no longer exists"))
		return false
	}
	categories[ts.catIdx].TierLists[ts.tlIdx] = next
	if err := storage.SaveCategories(categories); err != nil {
		ts.showError(err)
		return false
	}
	*ts.categories = categories
	ts.tierList = next
	ts.refreshRows()
	return true
}

func cloneCategories(src []models.Category) []models.Category {
	dst := make([]models.Category, len(src))
	for i := range src {
		dst[i].Name = src[i].Name
		dst[i].TierLists = make([]models.TierList, len(src[i].TierLists))
		for j := range src[i].TierLists {
			dst[i].TierLists[j] = editor.Clone(src[i].TierLists[j])
		}
	}
	return dst
}

func (ts *TierListScreen) refreshRows() {
	// Update the existing scene graph rather than replacing the entire window.
	// Rebuilding the window used to close the image panel after every paste,
	// recreated the keyboard shortcuts and caused visible flickering.
	if ts.tiersBox == nil || ts.poolBox == nil {
		return
	}
	ts.tierRows = make([]*TierRow, 0, len(ts.tierList.Tiers))
	rows := make([]fyne.CanvasObject, 0, len(ts.tierList.Tiers))
	for i := range ts.tierList.Tiers {
		row := NewTierRow(&ts.tierList.Tiers[i], ts)
		ts.tierRows = append(ts.tierRows, row)
		rows = append(rows, row)
	}
	ts.tiersBox.Objects = rows
	ts.tiersBox.Refresh()
	ts.rebuildPool()
}
