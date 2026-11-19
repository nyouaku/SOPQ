// ===== FILE: backend/internal/usecase/inventory_usecase_test.go =====
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

// mockInventoryRepo mencatat panggilan PostMovementTx untuk diverifikasi test.
type mockInventoryRepo struct {
	movements   []*domain.StockMovement
	effects     [][]repository.StockEffect
	entries     [][]*domain.StockLedger
	nextCodeSeq int
}

func (m *mockInventoryRepo) Create(ctx context.Context, mv *domain.StockMovement) error {
	m.movements = append(m.movements, mv)
	return nil
}

func (m *mockInventoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error) {
	return nil, nil
}

func (m *mockInventoryRepo) List(ctx context.Context, f repository.InventoryFilter) ([]domain.StockMovement, int64, error) {
	return nil, 0, nil
}

func (m *mockInventoryRepo) GridCells(ctx context.Context, warehouseID uuid.UUID, month time.Time, movementType string) ([]domain.GridCell, error) {
	return nil, nil
}

func (m *mockInventoryRepo) Summary(ctx context.Context, warehouseID uuid.UUID, month time.Time) (*domain.InventorySummary, error) {
	return &domain.InventorySummary{}, nil
}

func (m *mockInventoryRepo) NextCode(ctx context.Context, day time.Time) (string, error) {
	m.nextCodeSeq++
	return "MV-TEST-0001", nil
}

func (m *mockInventoryRepo) PostMovementTx(ctx context.Context, mv *domain.StockMovement, effects []repository.StockEffect, entries []*domain.StockLedger) error {
	m.movements = append(m.movements, mv)
	m.effects = append(m.effects, effects)
	m.entries = append(m.entries, entries)
	return nil
}

func (m *mockInventoryRepo) CancelMovement(ctx context.Context, id, userID uuid.UUID, reason string) error {
	return nil
}

func TestPostIn_BuildsEffectsAndLedger(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	whID := uuid.New()
	pid := uuid.New()

	mv, err := uc.PostIn(context.Background(), PostMovementInput{
		WarehouseID: whID,
		SubType:     domain.MovementSubPurchase,
		UserID:      uuid.New(),
		Items: []MovementItemInput{
			{ProductID: pid, Quantity: 10, UnitCost: 5000},
		},
	})
	if err != nil {
		t.Fatalf("PostIn() error = %v", err)
	}
	if mv.MovementType != domain.MovementTypeIn {
		t.Errorf("movement_type = %q, want in", mv.MovementType)
	}
	if mv.Code != "MV-TEST-0001" {
		t.Errorf("code = %q, want MV-TEST-0001", mv.Code)
	}

	// Efek stok: satu item, delta +10.
	if len(repo.effects) != 1 || len(repo.effects[0]) != 1 {
		t.Fatalf("effects recorded = %d groups", len(repo.effects))
	}
	if repo.effects[0][0].Delta != 10 {
		t.Errorf("delta = %+d, want +10", repo.effects[0][0].Delta)
	}

	// Ledger: satu baris purchase_in, qtyIn=10, qtyOut=0.
	if len(repo.entries) != 1 || len(repo.entries[0]) != 1 {
		t.Fatalf("entries recorded = %d groups", len(repo.entries))
	}
	e := repo.entries[0][0]
	if e.TransactionType != domain.LedgerTypePurchaseIn {
		t.Errorf("transaction_type = %q, want purchase_in", e.TransactionType)
	}
	if e.QuantityIn != 10 || e.QuantityOut != 0 {
		t.Errorf("qtyIn/qtyOut = %d/%d, want 10/0", e.QuantityIn, e.QuantityOut)
	}
	if e.TotalCost != 50000 {
		t.Errorf("total_cost = %f, want 50000", e.TotalCost)
	}
}

func TestPostOut_BuildsNegativeDelta(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	whID := uuid.New()
	pid := uuid.New()

	_, err := uc.PostOut(context.Background(), PostMovementInput{
		WarehouseID: whID,
		SubType:     domain.MovementSubSale,
		UserID:      uuid.New(),
		Items:       []MovementItemInput{{ProductID: pid, Quantity: 4}},
	})
	if err != nil {
		t.Fatalf("PostOut() error = %v", err)
	}
	if repo.effects[0][0].Delta != -4 {
		t.Errorf("delta = %+d, want -4", repo.effects[0][0].Delta)
	}
	e := repo.entries[0][0]
	if e.TransactionType != domain.LedgerTypeSaleOut {
		t.Errorf("transaction_type = %q, want sale_out", e.TransactionType)
	}
	if e.QuantityOut != 4 || e.QuantityIn != 0 {
		t.Errorf("qtyIn/qtyOut = %d/%d, want 0/4", e.QuantityIn, e.QuantityOut)
	}
}

func TestPostTransfer_BuildsTwoEffectsAndLedgers(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	from, to, pid := uuid.New(), uuid.New(), uuid.New()

	_, err := uc.PostTransfer(context.Background(), PostTransferInput{
		FromWarehouseID: from,
		ToWarehouseID:   to,
		UserID:          uuid.New(),
		Items:           []MovementItemInput{{ProductID: pid, Quantity: 7}},
	})
	if err != nil {
		t.Fatalf("PostTransfer() error = %v", err)
	}
	if len(repo.effects[0]) != 2 {
		t.Fatalf("effects = %d, want 2 (out + in)", len(repo.effects[0]))
	}
	if repo.effects[0][0].WarehouseID != from || repo.effects[0][0].Delta != -7 {
		t.Errorf("from effect = %+d @ %v, want -7 @ from", repo.effects[0][0].Delta, repo.effects[0][0].WarehouseID)
	}
	if repo.effects[0][1].WarehouseID != to || repo.effects[0][1].Delta != 7 {
		t.Errorf("to effect = %+d @ %v, want +7 @ to", repo.effects[0][1].Delta, repo.effects[0][1].WarehouseID)
	}
	if len(repo.entries[0]) != 2 {
		t.Fatalf("ledger entries = %d, want 2", len(repo.entries[0]))
	}
	if repo.entries[0][0].TransactionType != domain.LedgerTypeTransferOut {
		t.Errorf("entry[0] = %q, want transfer_out", repo.entries[0][0].TransactionType)
	}
	if repo.entries[0][1].TransactionType != domain.LedgerTypeTransferIn {
		t.Errorf("entry[1] = %q, want transfer_in", repo.entries[0][1].TransactionType)
	}
}

func TestPostTransfer_RejectsSameWarehouse(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	same := uuid.New()
	_, err := uc.PostTransfer(context.Background(), PostTransferInput{
		FromWarehouseID: same,
		ToWarehouseID:   same,
		UserID:          uuid.New(),
		Items:           []MovementItemInput{{ProductID: uuid.New(), Quantity: 1}},
	})
	if err == nil {
		t.Error("PostTransfer() with same from/to should fail")
	}
}

func TestPostIn_RejectsZeroQuantity(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	_, err := uc.PostIn(context.Background(), PostMovementInput{
		WarehouseID: uuid.New(),
		SubType:     domain.MovementSubPurchase,
		UserID:      uuid.New(),
		Items:       []MovementItemInput{{ProductID: uuid.New(), Quantity: 0}},
	})
	if err == nil {
		t.Error("PostIn() with qty 0 should fail")
	}
}

func TestLedgerTypeFor(t *testing.T) {
	cases := []struct{ mt, st, want string }{
		{domain.MovementTypeIn, domain.MovementSubPurchase, domain.LedgerTypePurchaseIn},
		{domain.MovementTypeIn, domain.MovementSubReturn, domain.LedgerTypeReturnIn},
		{domain.MovementTypeOut, domain.MovementSubSale, domain.LedgerTypeSaleOut},
		{domain.MovementTypeOut, domain.MovementSubReturn, domain.LedgerTypeReturnOut},
		{domain.MovementTypeTransfer, domain.MovementSubTransfer, domain.LedgerTypeTransferOut},
	}
	for _, c := range cases {
		if got := domain.LedgerTypeFor(c.mt, c.st); got != c.want {
			t.Errorf("LedgerTypeFor(%q,%q) = %q, want %q", c.mt, c.st, got, c.want)
		}
	}
}

// Regresi grid: isian grid dikelompokkan per tanggal -> satu dokumen per tanggal.
func TestPostGrid_GroupsByDate(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	wh := uuid.New()
	d1 := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)

	res, err := uc.PostGrid(context.Background(), PostGridInput{
		WarehouseID:  wh,
		MovementType: domain.MovementTypeOut,
		SubType:      domain.MovementSubSale,
		UserID:       uuid.New(),
		Entries: []GridEntryInput{
			{ProductID: uuid.New(), Date: d1, Quantity: 2},
			{ProductID: uuid.New(), Date: d1, Quantity: 3}, // sama tanggal d1
			{ProductID: uuid.New(), Date: d2, Quantity: 5},
		},
	})
	if err != nil {
		t.Fatalf("PostGrid() error = %v", err)
	}
	// 2 tanggal -> 2 dokumen.
	if res.MovementsCreated != 2 {
		t.Errorf("MovementsCreated = %d, want 2 (satu per tanggal)", res.MovementsCreated)
	}
	if res.ItemsPosted != 3 {
		t.Errorf("ItemsPosted = %d, want 3", res.ItemsPosted)
	}
	// Tiap dokumen: efek stok = jumlah item di tanggal itu.
	if len(repo.effects) != 2 {
		t.Fatalf("posting groups = %d, want 2", len(repo.effects))
	}
	if len(repo.effects[0]) != 2 {
		t.Errorf("tanggal pertama efek = %d, want 2 (2 item)", len(repo.effects[0]))
	}
	if len(repo.effects[1]) != 1 {
		t.Errorf("tanggal kedua efek = %d, want 1", len(repo.effects[1]))
	}
}

// Regresi grid: qty <= 0 ditolak.
func TestPostGrid_RejectsZeroQty(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	_, err := uc.PostGrid(context.Background(), PostGridInput{
		WarehouseID:  uuid.New(),
		MovementType: domain.MovementTypeOut,
		SubType:      domain.MovementSubSale,
		UserID:       uuid.New(),
		Entries:      []GridEntryInput{{ProductID: uuid.New(), Date: time.Now(), Quantity: 0}},
	})
	if err == nil {
		t.Error("PostGrid() dengan qty 0 harus gagal")
	}
}

// Regresi grid: tanpa tanggal ditolak.
func TestPostGrid_RejectsMissingDate(t *testing.T) {
	repo := &mockInventoryRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})
	_, err := uc.PostGrid(context.Background(), PostGridInput{
		WarehouseID:  uuid.New(),
		MovementType: domain.MovementTypeIn,
		SubType:      domain.MovementSubPurchase,
		UserID:       uuid.New(),
		Entries:      []GridEntryInput{{ProductID: uuid.New(), Quantity: 3}}, // tanpa Date
	})
	if err == nil {
		t.Error("PostGrid() tanpa tanggal harus gagal")
	}
}

// Mock pembatalan: cukup mencatat panggilan untuk memverifikasi validasi usecase.
type mockCancelRepo struct {
	repository.InventoryRepository // method lain tak dipakai di test ini
	calledWithID     uuid.UUID
	calledWithReason string
	returnErr        error
}

func (m *mockCancelRepo) CancelMovement(ctx context.Context, id, userID uuid.UUID, reason string) error {
	m.calledWithID = id
	m.calledWithReason = reason
	return m.returnErr
}

func TestCancelMovement_RequiresReason(t *testing.T) {
	repo := &mockCancelRepo{}
	uc := NewInventoryUsecase(repo, &mockActivityLogRepository{})

	err := uc.CancelMovement(context.Background(), uuid.New(), uuid.New(), "   ", "127.0.0.1")
	if err == nil {
		t.Fatal("alasan kosong harus ditolak")
	}
	if repo.calledWithID != uuid.Nil {
		t.Fatal("repo tak boleh dipanggil saat alasan kosong")
	}
}

func TestCancelMovement_TrimsReasonAndLogs(t *testing.T) {
	repo := &mockCancelRepo{}
	logRepo := &mockActivityLogRepository{}
	uc := NewInventoryUsecase(repo, logRepo)

	id := uuid.New()
	if err := uc.CancelMovement(context.Background(), id, uuid.New(), "  salah input  ", "127.0.0.1"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if repo.calledWithID != id {
		t.Fatal("repo tidak menerima ID yang benar")
	}
	if repo.calledWithReason != "salah input" {
		t.Fatalf("alasan tidak di-trim: %q", repo.calledWithReason)
	}
}
