package models

import (
	"encoding/json"
	"testing"
)

func TestDisplaySettingsCompatibility(t *testing.T) {
	var old TierList
	if err := json.Unmarshal([]byte(`{"name":"legacy","tiers":[]}`), &old); err != nil {
		t.Fatal(err)
	}
	if !old.LabelsVisible() || old.ImageFillMode() != "cover" {
		t.Fatal("legacy display defaults changed")
	}
	off := false
	old.ShowLabels = &off
	old.ImageFit = "contain"
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	var restored TierList
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.LabelsVisible() || restored.ImageFillMode() != "contain" {
		t.Fatal("display settings were not persisted")
	}
}
