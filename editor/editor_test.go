package editor

import (
	"errors"
	"reflect"
	"testing"
	"tierlist-app/models"
)

func fixture() models.TierList {
	return models.TierList{Tiers: []models.Tier{{Name: "S", Items: []models.TierItem{{ID: "a", ImagePath: "a.png"}}}, {Name: "A"}}, Pool: []models.TierItem{{ID: "b", Label: "B"}}}
}
func TestCloneIsIndependent(t *testing.T) {
	original := fixture()
	cloned := Clone(original)
	cloned.Tiers[0].Items[0].Label = "changed"
	cloned.Pool[0].ID = "changed"
	if original.Tiers[0].Items[0].Label != "" || original.Pool[0].ID != "b" {
		t.Fatal("clone aliases source")
	}
}
func TestMove(t *testing.T) {
	for _, tc := range []struct {
		name, id           string
		dest               int
		wantErr            error
		wantTier, wantPool int
	}{
		{"pool to tier", "b", 1, nil, 1, 0},
		{"tier to pool", "a", -1, nil, 0, 2},
		{"missing", "x", 0, ErrItemNotFound, 0, 1},
		{"bad destination", "a", 99, ErrInvalidTier, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list := fixture()
			before := Clone(list)
			err := Move(&list, tc.id, tc.dest)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error %v want %v", err, tc.wantErr)
			}
			if err != nil && !reflect.DeepEqual(list, before) {
				t.Fatal("failed move mutated list")
			}
			if err == nil {
				if len(list.Pool) != tc.wantPool {
					t.Fatal("unexpected pool size")
				}
				if tc.dest >= 0 && len(list.Tiers[tc.dest].Items) != tc.wantTier {
					t.Fatal("unexpected tier size")
				}
			}
		})
	}
}

func TestRemoveAndValidate(t *testing.T) {
	list := fixture()
	path, err := Remove(&list, "a")
	if err != nil || path != "a.png" {
		t.Fatalf("remove: %q %v", path, err)
	}
	if err := Validate(list); err != nil {
		t.Fatal(err)
	}
	list.Pool = append(list.Pool, list.Pool[0])
	if !errors.Is(Validate(list), ErrDuplicateID) {
		t.Fatal("duplicate ID accepted")
	}
}
func TestRename(t *testing.T) {
	list := fixture()
	if err := RenameItem(&list, "b", "Renamed"); err != nil {
		t.Fatal(err)
	}
	if list.Pool[0].Label != "Renamed" {
		t.Fatal("not renamed")
	}
	if !errors.Is(RenameItem(&list, "b", "  "), ErrInvalidName) {
		t.Fatal("empty name accepted")
	}
}

// TestRoundTripPoolTierPool guarantees that images can be returned to the
// unassigned collection without losing their identity or user-defined label.
func TestRoundTripPoolTierPool(t *testing.T) {
	list := fixture()
	if err := RenameItem(&list, "b", "Saab 9-5"); err != nil {
		t.Fatal(err)
	}
	if err := Move(&list, "b", 0); err != nil {
		t.Fatal(err)
	}
	if len(list.Pool) != 0 || len(list.Tiers[0].Items) != 2 {
		t.Fatal("image not moved into tier")
	}
	if err := Move(&list, "b", -1); err != nil {
		t.Fatal(err)
	}
	if len(list.Pool) != 1 || list.Pool[0].ID != "b" || list.Pool[0].Label != "Saab 9-5" {
		t.Fatalf("image was not restored: %+v", list.Pool)
	}
	if err := Validate(list); err != nil {
		t.Fatal(err)
	}
}
