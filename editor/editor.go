// Package editor contains UI-independent operations on tier lists.
package editor

import (
	"errors"
	"fmt"
	"strings"

	"tierlist-app/models"
)

var (
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidTier  = errors.New("invalid tier index")
	ErrDuplicateID  = errors.New("duplicate item ID")
	ErrInvalidName  = errors.New("name must not be empty")
)

// Clone creates a deep copy, so unsuccessful persistence cannot mutate the live document.
func Clone(src models.TierList) models.TierList {
	dst := src
	if src.ShowLabels != nil {
		show := *src.ShowLabels
		dst.ShowLabels = &show
	}
	dst.Tiers = make([]models.Tier, len(src.Tiers))
	for i, tier := range src.Tiers {
		dst.Tiers[i] = tier
		dst.Tiers[i].Items = append([]models.TierItem(nil), tier.Items...)
	}
	dst.Pool = append([]models.TierItem(nil), src.Pool...)
	return dst
}

// Move transfers an item to a tier, or to the pool when destination is -1.
// It preserves order of the remaining items and never changes the document on failure.
func Move(list *models.TierList, id string, destination int) error {
	return MoveAt(list, id, destination, -1)
}

// Remove deletes an item from either the pool or a tier and returns its image path.
func Remove(list *models.TierList, id string) (string, error) {
	if list == nil {
		return "", errors.New("nil tier list")
	}
	next := Clone(*list)
	for i := range next.Tiers {
		for j, item := range next.Tiers[i].Items {
			if item.ID == id {
				next.Tiers[i].Items = append(next.Tiers[i].Items[:j], next.Tiers[i].Items[j+1:]...)
				*list = next
				return item.ImagePath, nil
			}
		}
	}
	for j, item := range next.Pool {
		if item.ID == id {
			next.Pool = append(next.Pool[:j], next.Pool[j+1:]...)
			*list = next
			return item.ImagePath, nil
		}
	}
	return "", ErrItemNotFound
}

func RenameItem(list *models.TierList, id, name string) error {
	if list == nil {
		return errors.New("nil tier list")
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	for i := range list.Tiers {
		for j := range list.Tiers[i].Items {
			if list.Tiers[i].Items[j].ID == id {
				list.Tiers[i].Items[j].Label = name
				return nil
			}
		}
	}
	for j := range list.Pool {
		if list.Pool[j].ID == id {
			list.Pool[j].Label = name
			return nil
		}
	}
	return ErrItemNotFound
}

func Validate(list models.TierList) error {
	seen := make(map[string]struct{})
	check := func(item models.TierItem) error {
		if item.ID == "" {
			return errors.New("item has empty ID")
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateID, item.ID)
		}
		seen[item.ID] = struct{}{}
		return nil
	}
	for _, tier := range list.Tiers {
		for _, item := range tier.Items {
			if err := check(item); err != nil {
				return err
			}
		}
	}
	for _, item := range list.Pool {
		if err := check(item); err != nil {
			return err
		}
	}
	return nil
}

// MoveAt moves an item to the requested insertion position. A negative tier
// means the pool. The index is interpreted after removal from the source.
func MoveAt(list *models.TierList, id string, destination, index int) error {
	if list == nil {
		return errors.New("nil tier list")
	}
	if destination < -1 || destination >= len(list.Tiers) {
		return ErrInvalidTier
	}
	next := Clone(*list)
	var item models.TierItem
	found := false
	remove := func(items []models.TierItem) []models.TierItem {
		for i, current := range items {
			if current.ID == id {
				item, found = current, true
				return append(items[:i:i], items[i+1:]...)
			}
		}
		return items
	}
	next.Pool = remove(next.Pool)
	if !found {
		for i := range next.Tiers {
			next.Tiers[i].Items = remove(next.Tiers[i].Items)
			if found {
				break
			}
		}
	}
	if !found {
		return ErrItemNotFound
	}
	insert := func(items []models.TierItem) []models.TierItem {
		if index < 0 || index > len(items) {
			index = len(items)
		}
		items = append(items, models.TierItem{})
		copy(items[index+1:], items[index:])
		items[index] = item
		return items
	}
	if destination == -1 {
		next.Pool = insert(next.Pool)
	} else {
		next.Tiers[destination].Items = insert(next.Tiers[destination].Items)
	}
	*list = next
	return nil
}
