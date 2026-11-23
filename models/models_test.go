package models

import "testing"

func TestRepairIDs(t *testing.T) {
	const duplicateID = "duplicate"
	tl := NewTierList("Example")
	tl.Tiers[0].Items = []TierItem{{ID: duplicateID}, {ID: duplicateID}, {}}
	tl.Pool = []TierItem{{ID: duplicateID}, {}}
	tl.RepairIDs()
	seen := make(map[string]bool)
	for _, tier := range tl.Tiers {
		for _, item := range tier.Items {
			if item.ID == "" || seen[item.ID] {
				t.Fatalf("invalid ID %q", item.ID)
			}
			seen[item.ID] = true
		}
	}
	for _, item := range tl.Pool {
		if item.ID == "" || seen[item.ID] {
			t.Fatalf("invalid pool ID %q", item.ID)
		}
		seen[item.ID] = true
	}
	if tl.Tiers[0].Items[0].ID != duplicateID {
		t.Fatal("first occurrence changed")
	}
}
