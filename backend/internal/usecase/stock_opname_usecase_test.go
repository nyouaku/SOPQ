// ===== FILE: backend/internal/usecase/stock_opname_usecase_test.go =====
package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

// ---------------------------------------------------------------- mocks

type mockStockOpnameRepository struct {
	sessions  map[uuid.UUID]*domain.StockOpname
	codes     map[string]int // day -> last value
	createErr error
	// createErrHook dipakai untuk mensimulasikan bentrok kode sementara.
	createErrHook func(*domain.StockOpname) error
}

func newMockStockOpnameRepo() *mockStockOpnameRepository {
	return &mockStockOpnameRepository{
		sessions: make(map[uuid.UUID]*domain.StockOpname),
		codes:    make(map[string]int),
	}
}

func (m *mockStockOpnameRepository) Create(ctx context.Context, opname *domain.StockOpname) error {
	if m.createErrHook != nil {
		if err := m.createErrHook(opname); err != nil {
			return err
		}
	}
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[opname.ID] = opname
	return nil
}

func (m *mockStockOpnameRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpname, error) {
	return m.sessions[id], nil
}

func (m *mockStockOpnameRepository) Update(ctx context.Context, opname *domain.StockOpname) error {
	m.sessions[opname.ID] = opname
	return nil
}

func (m *mockStockOpnameRepository) List(ctx context.Context, filter repository.StockOpnameFilter) ([]domain.StockOpname, int64, error) {
	return nil, 0, nil
}

// NextCode meniru ON CONFLICT ... RETURNING: counter per hari.
func (m *mockStockOpnameRepository) NextCode(ctx context.Context, day time.Time) (string, error) {
	key := day.Format("2006-01-02")
	m.codes[key]++
	return fmt.Sprintf("SO-%s-%04d", day.Format("20060102"), m.codes[key]), nil
}

// ApproveWithStock meniru transisi completed -> approved + tulis stock.
func (m *mockStockOpnameRepository) ApproveWithStock(ctx context.Context, sessionID, approverID uuid.UUID) error {
	session, ok := m.sessions[sessionID]
	if !ok {
		return domain.ErrSessionNotFound
	}
	if session.Status == "approved" {
		return domain.ErrSessionAlreadyApproved
	}
	if session.Status != "completed" {
		return domain.ErrSessionNotApprovable
	}

	now := time.Now()
	session.Status = "approved"
	session.ApprovedBy = &approverID
	session.ApprovedAt = &now
	return nil
}

type mockStockOpnameDetailRepository struct {
	details map[uuid.UUID]*domain.StockOpnameDetail
}

func (m *mockStockOpnameDetailRepository) Create(ctx context.Context, detail *domain.StockOpnameDetail) error {
	if detail.ID == uuid.Nil {
		detail.ID = uuid.New()
	}
	m.details[detail.ID] = detail
	return nil
}

func (m *mockStockOpnameDetailRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockOpnameDetail, error) {
	return m.details[id], nil
}

func (m *mockStockOpnameDetailRepository) FindByIdempotencyKey(ctx context.Context, sessionID uuid.UUID, key string) (*domain.StockOpnameDetail, error) {
	for _, detail := range m.details {
		if detail.StockOpnameID == sessionID && detail.IdempotencyKey != nil && *detail.IdempotencyKey == key {
			return detail, nil
		}
	}
	return nil, nil
}

func (m *mockStockOpnameDetailRepository) FindByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID, limit, offset int) ([]domain.StockOpnameDetail, int64, error) {
	var result []domain.StockOpnameDetail
	for _, detail := range m.details {
		if detail.StockOpnameID == stockOpnameID {
			result = append(result, *detail)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockStockOpnameDetailRepository) CountByStockOpnameID(ctx context.Context, stockOpnameID uuid.UUID) (int64, error) {
	var n int64
	for _, detail := range m.details {
		if detail.StockOpnameID == stockOpnameID {
			n++
		}
	}
	return n, nil
}

func (m *mockStockOpnameDetailRepository) FindProductTrend(ctx context.Context, productID uuid.UUID, from, to *time.Time, limit int) ([]domain.ProductStockTrend, error) {
	return nil, nil
}

type mockStockRepository struct {
	stocks map[uuid.UUID]*domain.Stock
}

// errOptimisticLockForTest meniru error yang dikembalikan repository asli
// (repository mengembalikan fmt.Errorf, bukan sentinel domain).
var errOptimisticLockForTest = errors.New("optimistic lock failed: stock was modified by another transaction")

func (m *mockStockRepository) FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*domain.Stock, error) {
	for _, stock := range m.stocks {
		if stock.WarehouseID == warehouseID && stock.ProductID == productID {
			return stock, nil
		}
	}
	return nil, nil
}

func (m *mockStockRepository) UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock) error {
	existing, exists := m.stocks[stock.ID]
	if exists && existing.Version != stock.Version {
		return errOptimisticLockForTest
	}
	stock.Version++
	m.stocks[stock.ID] = stock
	return nil
}

func (m *mockStockRepository) CreateOrUpdate(ctx context.Context, stock *domain.Stock) error {
	if stock.ID == uuid.Nil {
		stock.ID = uuid.New()
	}
	m.stocks[stock.ID] = stock
	return nil
}

func (m *mockStockRepository) ReconcileWarehouses(ctx context.Context, warehouseIDs []uuid.UUID) ([]domain.WarehouseReconciliation, error) {
	return nil, nil
}

func (m *mockStockRepository) AdjustStock(ctx context.Context, productID, warehouseID uuid.UUID, delta int, version int) (oldVal, newVal int, err error) {
	return 0, 0, nil
}

type mockProductRepository struct {
	products map[string]*domain.Product // key: barcode
}

func (m *mockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	return nil
}

func (m *mockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProductRepository) FindByCode(ctx context.Context, code string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.ProductCode == code {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProductRepository) Update(ctx context.Context, product *domain.Product) error {
	return nil
}
func (m *mockProductRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func (m *mockProductRepository) List(ctx context.Context, filter repository.ProductFilter) ([]domain.Product, int64, error) {
	return nil, 0, nil
}

func (m *mockProductRepository) BulkUpdateBoxID(ctx context.Context, productIDs []uuid.UUID, boxID *uuid.UUID) (int64, error) {
	return int64(len(productIDs)), nil
}

type mockActivityLogRepository struct{}

func (m *mockActivityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	return nil
}

func (m *mockActivityLogRepository) List(ctx context.Context, filter repository.ActivityLogFilter) ([]domain.ActivityLog, int64, error) {
	return nil, 0, nil
}

// ---------------------------------------------------------------- helpers

func newTestStockOpnameUsecase(t *testing.T) (StockOpnameUsecase, *mockStockOpnameRepository, *mockStockOpnameDetailRepository, *mockStockRepository, *mockProductRepository) {
	t.Helper()
	opnameRepo := newMockStockOpnameRepo()
	detailRepo := &mockStockOpnameDetailRepository{details: make(map[uuid.UUID]*domain.StockOpnameDetail)}
	stockRepo := &mockStockRepository{stocks: make(map[uuid.UUID]*domain.Stock)}
	productRepo := &mockProductRepository{products: make(map[string]*domain.Product)}
	uc := NewStockOpnameUsecase(opnameRepo, detailRepo, stockRepo, productRepo, &mockActivityLogRepository{}, nil)
	return uc, opnameRepo, detailRepo, stockRepo, productRepo
}

// ---------------------------------------------------------------- tests

// Regresi BE-01: sesi kedua di hari yang sama harus mendapat kode berbeda.
func TestStartSession_CodeIncrementsSameDay(t *testing.T) {
	uc, _, _, _, _ := newTestStockOpnameUsecase(t)
	warehouseID := uuid.New()
	userID := uuid.New()

	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		session, err := uc.StartSession(context.Background(), StartSessionInput{
			WarehouseID: warehouseID,
			StartedBy:   userID,
		})
		if err != nil {
			t.Fatalf("StartSession #%d error = %v", i+1, err)
		}
		if seen[session.Code] {
			t.Fatalf("duplicate session code %q", session.Code)
		}
		seen[session.Code] = true
	}

	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct codes, got %d", len(seen))
	}
}

// Regresi: counter kode bisa tertinggal dari kode yang sudah terpakai (seq
// dikosongkan / DB di-restore). StartSession harus ambil nomor berikutnya,
// bukan menyerahkan error 23505 ke user.
func TestStartSession_RetriesWhenCodeAlreadyTaken(t *testing.T) {
	uc, opnameRepo, _, _, _ := newTestStockOpnameUsecase(t)
	warehouseID := uuid.New()
	userID := uuid.New()

	// Dua kode pertama dianggap sudah dipakai baris lain.
	taken := 2
	opnameRepo.createErrHook = func(s *domain.StockOpname) error {
		if taken > 0 {
			taken--
			return domain.ErrSessionCodeTaken
		}
		return nil
	}

	session, err := uc.StartSession(context.Background(), StartSessionInput{
		WarehouseID: warehouseID,
		StartedBy:   userID,
	})
	if err != nil {
		t.Fatalf("StartSession error = %v", err)
	}
	if session.Code != "SO-"+time.Now().Format("20060102")+"-0003" {
		t.Fatalf("expected code ...-0003, got %q", session.Code)
	}
}

// Regresi BE-02: SubmitScan harus mengembalikan baris yang baru dibuat, bukan baris terbaru sesi.
func TestSubmitScan_ReturnsInsertedRow(t *testing.T) {
	uc, opnameRepo, detailRepo, stockRepo, productRepo := newTestStockOpnameUsecase(t)

	warehouseID := uuid.New()
	userID := uuid.New()
	sessionID := uuid.New()
	opnameRepo.sessions[sessionID] = &domain.StockOpname{
		ID:          sessionID,
		WarehouseID: warehouseID,
		Status:      "in_progress",
	}

	// Produk A sudah pernah discan oleh user lain (baris "terbaru" di sesi).
	otherProduct := &domain.Product{ID: uuid.New(), ProductCode: "OTHER-1", Name: "Other"}
	productRepo.products["OTHER-1"] = otherProduct
	detailRepo.details[uuid.New()] = &domain.StockOpnameDetail{
		ID:            uuid.New(),
		StockOpnameID: sessionID,
		ProductID:     otherProduct.ID,
		ScannedBy:     uuid.New(),
		ScannedAt:     time.Now().Add(-time.Minute),
	}

	// Produk yang akan discan sekarang.
	product := &domain.Product{ID: uuid.New(), ProductCode: "PRD-1", Name: "Produk"}
	productRepo.products["BC-1"] = product
	stockRepo.stocks[uuid.New()] = &domain.Stock{
		ProductID:      product.ID,
		WarehouseID:    warehouseID,
		QuantitySystem: 100,
		Version:        1,
	}

	detail, err := uc.SubmitScan(context.Background(), SubmitScanInput{
		SessionID:     sessionID,
		ProductCode:   "PRD-1",
		StockPhysical: 120,
		InputMethod:   "manual",
		ScannedBy:     userID,
	})
	if err != nil {
		t.Fatalf("SubmitScan() error = %v", err)
	}

	if detail.ProductID != product.ID {
		t.Fatalf("SubmitScan returned product %s, want %s (wrong-row bug)", detail.ProductID, product.ID)
	}
	if detail.StockSystem != 100 {
		t.Errorf("stock_system = %d, want 100", detail.StockSystem)
	}
	if detail.StockPhysical != 120 {
		t.Errorf("stock_physical = %d, want 120", detail.StockPhysical)
	}
}

// Regresi BE-04: approve harus menutup siklus (transisi status) dan menolak sesi yang belum completed.
func TestApproveSession_TransitionsAndGuards(t *testing.T) {
	uc, opnameRepo, _, _, _ := newTestStockOpnameUsecase(t)
	approverID := uuid.New()
	sessionID := uuid.New()
	opnameRepo.sessions[sessionID] = &domain.StockOpname{
		ID:          sessionID,
		WarehouseID: uuid.New(),
		Status:      "in_progress",
	}

	if err := uc.ApproveSession(context.Background(), sessionID, approverID); err != domain.ErrSessionNotApprovable {
		t.Fatalf("approve in_progress session: got %v, want ErrSessionNotApprovable", err)
	}

	opnameRepo.sessions[sessionID].Status = "completed"
	if err := uc.ApproveSession(context.Background(), sessionID, approverID); err != nil {
		t.Fatalf("approve completed session: %v", err)
	}

	session := opnameRepo.sessions[sessionID]
	if session.Status != "approved" {
		t.Errorf("status = %q, want approved", session.Status)
	}
	if session.ApprovedBy == nil || *session.ApprovedBy != approverID {
		t.Errorf("approved_by not set")
	}

	if err := uc.ApproveSession(context.Background(), sessionID, approverID); err != domain.ErrSessionAlreadyApproved {
		t.Fatalf("double approve: got %v, want ErrSessionAlreadyApproved", err)
	}
}

// Regresi FN-04: retry dengan idempotency key sama tidak boleh membuat baris baru.
func TestSubmitScan_IdempotentRetry(t *testing.T) {
	uc, opnameRepo, _, stockRepo, productRepo := newTestStockOpnameUsecase(t)

	warehouseID := uuid.New()
	sessionID := uuid.New()
	opnameRepo.sessions[sessionID] = &domain.StockOpname{
		ID:          sessionID,
		WarehouseID: warehouseID,
		Status:      "in_progress",
	}

	product := &domain.Product{ID: uuid.New(), ProductCode: "BC-IDEM", Name: "Produk Idem"}
	productRepo.products["BC-IDEM"] = product
	stockRepo.stocks[uuid.New()] = &domain.Stock{
		ProductID:      product.ID,
		WarehouseID:    warehouseID,
		QuantitySystem: 10,
		Version:        1,
	}

	input := SubmitScanInput{
		SessionID:      sessionID,
		ProductCode:    "BC-IDEM",
		StockPhysical:  12,
		InputMethod:    "manual",
		ScannedBy:      uuid.New(),
		IdempotencyKey: uuid.New().String(),
	}

	first, err := uc.SubmitScan(context.Background(), input)
	if err != nil {
		t.Fatalf("first SubmitScan() error = %v", err)
	}

	second, err := uc.SubmitScan(context.Background(), input)
	if err != nil {
		t.Fatalf("retry SubmitScan() error = %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("retry created a new row: %s vs %s", first.ID, second.ID)
	}
}

func TestStockOpnameUsecase_OptimisticLocking(t *testing.T) {
	stockRepo := &mockStockRepository{stocks: make(map[uuid.UUID]*domain.Stock)}

	stock := &domain.Stock{
		ID:             uuid.New(),
		ProductID:      uuid.New(),
		WarehouseID:    uuid.New(),
		QuantitySystem: 100,
		Version:        1,
	}
	stockRepo.stocks[stock.ID] = stock

	stock1 := *stock
	if err := stockRepo.UpdateWithOptimisticLock(context.Background(), &stock1); err != nil {
		t.Errorf("first update should succeed, got error: %v", err)
	}
	if stock1.Version != 2 {
		t.Errorf("expected version 2 after update, got %d", stock1.Version)
	}

	stock2 := *stock // still version 1
	if err := stockRepo.UpdateWithOptimisticLock(context.Background(), &stock2); err != errOptimisticLockForTest {
		t.Errorf("expected optimistic lock error, got: %v", err)
	}
}

// ===== END FILE =====
