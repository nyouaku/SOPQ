package repository

import (
	"slices"
	"testing"
)

// Regresi: Create pernah tidak menyertakan idempotency_key di Select, sehingga
// kolomnya tidak pernah tertulis dan retry klien selalu membuat baris baru.
func TestStockOpnameDetailInsertFields(t *testing.T) {
	if !slices.Contains(stockOpnameDetailInsertFields, "idempotency_key") {
		t.Fatalf("idempotency_key harus ikut di-INSERT; fields = %v", stockOpnameDetailInsertFields)
	}
	// difference adalah generated column — DB yang menghitung, bukan aplikasi.
	if slices.Contains(stockOpnameDetailInsertFields, "difference") {
		t.Fatalf("difference tidak boleh di-INSERT; fields = %v", stockOpnameDetailInsertFields)
	}
}
