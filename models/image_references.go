package models

// ImageInUse reports whether any item in any category references path.
// The comparison is exact because image paths are persisted identifiers.
func ImageInUse(categories []Category, path string) bool {
	if path == "" {
		return false
	}
	for _, category := range categories {
		for _, list := range category.TierLists {
			if tierListUsesImage(list, path) {
				return true
			}
		}
	}
	return false
}

func tierListUsesImage(list TierList, path string) bool {
	if itemsUseImage(list.Pool, path) {
		return true
	}
	for _, tier := range list.Tiers {
		if itemsUseImage(tier.Items, path) {
			return true
		}
	}
	return false
}

func itemsUseImage(items []TierItem, path string) bool {
	for _, item := range items {
		if item.ImagePath == path {
			return true
		}
	}
	return false
}
