package models

import (
	"fmt"
	"sync"
	"time"
)

var (
	idCounter uint64
	idMu      sync.Mutex
)

func NewItemID() string {
	idMu.Lock()
	idCounter++
	n := idCounter
	idMu.Unlock()
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), n)
}

func (tl *TierList) RepairIDs() {
	seen := make(map[string]bool)
	repair := func(items []TierItem) {
		for i := range items {
			if items[i].ID == "" || seen[items[i].ID] {
				items[i].ID = NewItemID()
			}
			seen[items[i].ID] = true
		}
	}
	for t := range tl.Tiers {
		repair(tl.Tiers[t].Items)
	}
	repair(tl.Pool)
}

type TierItem struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	ImagePath string `json:"image_path"`
}

type Tier struct {
	Name  string     `json:"name"`
	Items []TierItem `json:"items"`
}

type TierList struct {
	ShowLabels *bool      `json:"show_labels,omitempty"`
	ImageFit   string     `json:"image_fit,omitempty"`
	Name       string     `json:"name"`
	Tiers      []Tier     `json:"tiers"`
	Pool       []TierItem `json:"pool,omitempty"`
}

type Category struct {
	Name      string     `json:"name"`
	TierLists []TierList `json:"tier_lists"`
}

func NewTierList(name string) TierList {
	return TierList{
		Name: name,
		Tiers: []Tier{
			{Name: "S Tier", Items: []TierItem{}},
			{Name: "A Tier", Items: []TierItem{}},
			{Name: "B Tier", Items: []TierItem{}},
			{Name: "C Tier", Items: []TierItem{}},
			{Name: "D Tier", Items: []TierItem{}},
		},
		Pool: []TierItem{},
	}
}

// LabelsVisible defaults to true for tier lists created before display settings existed.
func (tl TierList) LabelsVisible() bool { return tl.ShowLabels == nil || *tl.ShowLabels }

// ImageFillMode defaults to cover, which avoids letterboxing.
func (tl TierList) ImageFillMode() string {
	if tl.ImageFit == "contain" {
		return "contain"
	}
	return "cover"
}
