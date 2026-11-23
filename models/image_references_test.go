package models

import "testing"

func TestImageInUse(t *testing.T) {
	categories := []Category{
		{TierLists: []TierList{{Pool: []TierItem{{ImagePath: "pool.png"}}}}},
		{TierLists: []TierList{{Tiers: []Tier{{Items: []TierItem{{ImagePath: "tier.png"}}}}}}},
	}
	cases := []struct {
		name, path string
		want       bool
	}{
		{"pool", "pool.png", true},
		{"tier", "tier.png", true},
		{"missing", "missing.png", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ImageInUse(categories, tc.path); got != tc.want {
				t.Fatalf("ImageInUse(%q) = %t, want %t", tc.path, got, tc.want)
			}
		})
	}
}
