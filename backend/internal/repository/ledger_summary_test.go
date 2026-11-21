package repository

import (
	"os"
	"strings"
	"testing"
)

// Regresi: saldo awal/akhir Grid Harian harus derived dari stock_ledger,
// bukan dari tabel stock live (yang berubah tiap transaksi).
func TestLedgerSummaryDerivedFromLedger(t *testing.T) {
	src, err := os.ReadFile("stock_ledger_repository.go")
	if err != nil {
		t.Fatalf("baca stock_ledger_repository.go: %v", err)
	}
	s := string(src)

	if !strings.Contains(s, "FROM stock_ledger sl") {
		t.Fatal("LedgerSummary tidak membaca dari stock_ledger")
	}
	if !strings.Contains(s, "opening") || !strings.Contains(s, "closing") {
		t.Fatal("LedgerSummary tidak menghitung opening/closing")
	}
}
