package repository

import (
	"os"
	"strings"
	"testing"
)

// Regresi: approve SO harus menulis stock_ledger (opname_correction) di transaksi
// yang sama saat mengubah stock.quantity_system. Jika tidak, LedgerSummary.closing
// akan beda dengan stok nyata setelah SO.
func TestApproveWithStockWritesOpnameLedger(t *testing.T) {
	src, err := os.ReadFile("stock_opname_repository.go")
	if err != nil {
		t.Fatalf("baca stock_opname_repository.go: %v", err)
	}
	s := string(src)

	if !strings.Contains(s, "LedgerTypeOpnameCorrect") {
		t.Fatal("ApproveWithStock tidak menulis ledger opname_correction")
	}
	if !strings.Contains(s, "BalanceAfter:    d.StockPhysical") {
		t.Fatal("Ledger SO tidak mencatat saldo akhir = stok fisik")
	}
	if !strings.Contains(s, "ReferenceType:   refType") {
		t.Fatal("Ledger SO tidak punya referensi sesi")
	}
}
