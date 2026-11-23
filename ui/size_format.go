package ui

import (
	"fmt"
	"math"
	"strings"
)

// formatDataSize reports binary sizes with Russian labels.
func formatDataSize(bytes int64) string {
	if bytes < 0 {
		return "0 Б"
	}
	if bytes < 1024 {
		return fmt.Sprintf("%d Б", bytes)
	}
	units := []string{"КиБ", "МиБ", "ГиБ", "ТиБ"}
	value := float64(bytes)
	unit := 0
	for value >= 1024 && unit < len(units) {
		value /= 1024
		unit++
	}
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return "0 Б"
	}
	return strings.ReplaceAll(fmt.Sprintf("%.2f %s", value, units[unit-1]), ".", ",")
}
