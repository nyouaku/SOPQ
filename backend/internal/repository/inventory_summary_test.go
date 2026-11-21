package repository

import (
	"os"
	"strings"
	"testing"
)

// Regresi: Summary() sebelumnya memakai `smi.total_cost` padahal kolom
// stock_movement_item hanya punya `unit_cost` → nilai masuk/keluar salah.
// Test ini menjaga query tetap memakai quantity * unit_cost.
func TestSummaryQueryUsesUnitCostNotTotalCost(t *testing.T) {
	src, err := os.ReadFile("inventory_repository.go")
	if err != nil {
		t.Fatalf("baca inventory_repository.go: %v", err)
	}
	s := string(src)

	if strings.Contains(s, "smi.total_cost") {
		t.Fatal("Summary query masih memakai smi.total_cost (kolom tidak ada)")
	}
	if !strings.Contains(s, "smi.quantity * smi.unit_cost") {
		t.Fatal("Summary query tidak memakai smi.quantity * smi.unit_cost")
	}
}
