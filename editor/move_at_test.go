package editor

import (
	"reflect"
	"testing"
	"tierlist-app/models"
)

func TestMoveAtReorderAndReturn(t *testing.T) {
	list := models.TierList{Pool: []models.TierItem{{ID: "p"}}, Tiers: []models.Tier{{Items: []models.TierItem{{ID: "a"}, {ID: "b"}, {ID: "c"}}}}}
	if err := MoveAt(&list, "p", 0, 1); err != nil {
		t.Fatal(err)
	}
	if got := ids(list.Tiers[0].Items); !reflect.DeepEqual(got, []string{"a", "p", "b", "c"}) {
		t.Fatal(got)
	}
	if err := MoveAt(&list, "a", 0, 3); err != nil {
		t.Fatal(err)
	}
	if got := ids(list.Tiers[0].Items); !reflect.DeepEqual(got, []string{"p", "b", "c", "a"}) {
		t.Fatal(got)
	}
	if err := MoveAt(&list, "a", -1, 0); err != nil {
		t.Fatal(err)
	}
	if got := ids(list.Pool); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatal(got)
	}
}
func ids(items []models.TierItem) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func TestMoveAtInvalidDoesNotMutate(t *testing.T) {
	list := models.TierList{Pool: []models.TierItem{{ID: "a"}}}
	before := list
	if err := MoveAt(&list, "a", 5, 0); err == nil {
		t.Fatal("expected error")
	}
	if !reflect.DeepEqual(list, before) {
		t.Fatal("mutated")
	}
}
