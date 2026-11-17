// ===== FILE: backend/internal/usecase/stock_opname_usecase.go =====
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"gorm.io/datatypes"
)

type StockOpnameUsecase interface {
	StartSession(ctx context.Context, input StartSessionInput) (*domain.StockOpname, error)
	SubmitScan(ctx context.Context, input SubmitScanInput) (*domain.StockOpnameDetail, error)
	CompleteSession(ctx context.Context, sessionID, userID uuid.UUID) error
	ApproveSession(ctx context.Context, sessionID, approverID uuid.UUID) error
	RejectSession(ctx context.Context, sessionID, rejectedBy uuid.UUID, reason string) error
	ReopenSession(ctx context.Context, sessionID, reopenedBy uuid.UUID) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.StockOpname, error)
	GetSessionDetails(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]domain.StockOpnameDetail, int64, error)
	GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*domain.SessionSummary, error)
	// GetProgressByBox mengembalikan progress scan per rak/box dalam satu sesi.
	GetProgressByBox(ctx context.Context, sessionID uuid.UUID) ([]domain.BoxProgress, error)
	// CompareSessions membandingkan hasil fisik 2 sesi SO per produk.
	CompareSessions(ctx context.Context, prevID, currID uuid.UUID) (*domain.SessionCompare, error)
	ListSessions(ctx context.Context, filter repository.StockOpnameFilter) ([]domain.StockOpname, int64, error)
	// GetProductTrend mengembalikan riwayat stok fisik satu produk untuk grafik tren.
	// from/to opsional (nil = tanpa batas awal/akhir). limit default 100, max 500.
	GetProductTrend(ctx context.Context, productID uuid.UUID, from, to *time.Time, limit int) ([]domain.ProductStockTrend, error)
	// ReconcileWarehouses merekap stok produk di seluruh gudang yang dipilih.
	ReconcileWarehouses(ctx context.Context, warehouseIDs []uuid.UUID) ([]domain.WarehouseReconciliation, error)
	// AdjustStock menyesuaikan stok sistem secara manual dengan alasan wajib.
	AdjustStock(ctx context.Context, input AdjustStockInput) (*domain.Stock, error)
	// SetLedgerRepo menyuntikkan ledger repo agar perubahan stok tercatat di stock_ledger.
	SetLedgerRepo(r repository.StockLedgerRepository)
}

// AdjustStockInput untuk penyesuaian stok manual.
type AdjustStockInput struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Delta       int    // bisa negatif atau positif
	Reason      string // damaged, lost, expired, found, typo_fix, other
	Note        string // wajib jika reason = other
	UserID      uuid.UUID
	IPAddress   string
}

// ScanBroadcaster menerima update real-time. Dideklarasikan di sini (bukan
// diimpor dari package websocket) supaya usecase tidak bergantung ke delivery.
type ScanBroadcaster interface {
	BroadcastNewScan(sessionID uuid.UUID, detail *domain.StockOpnameDetail)
	BroadcastSessionUpdate(sessionID uuid.UUID, session *domain.StockOpname)
	BroadcastDiscrepancyAlert(sessionID uuid.UUID, productName string, difference int, percent float64)
}

type stockOpnameUsecase struct {
	stockOpnameRepo       repository.StockOpnameRepository
	stockOpnameDetailRepo repository.StockOpnameDetailRepository
	stockRepo             repository.StockRepository
	productRepo           repository.ProductRepository
	activityLogRepo       repository.ActivityLogRepository
	// ledgerRepo opsional: bila tersedia, setiap perubahan quantity_system (adjustment
	// maupun approve SO) juga dicatat ke stock_ledger agar saldo ledger tak menyimpang.
	// Nil (mis. di sebagian test) berarti ledger dilewati tanpa menggagalkan operasi.
	ledgerRepo repository.StockLedgerRepository
	broadcaster ScanBroadcaster
}

func NewStockOpnameUsecase(
	stockOpnameRepo repository.StockOpnameRepository,
	stockOpnameDetailRepo repository.StockOpnameDetailRepository,
	stockRepo repository.StockRepository,
	productRepo repository.ProductRepository,
	activityLogRepo repository.ActivityLogRepository,
	broadcaster ScanBroadcaster,
) StockOpnameUsecase {
	return &stockOpnameUsecase{
		stockOpnameRepo:       stockOpnameRepo,
		stockOpnameDetailRepo: stockOpnameDetailRepo,
		stockRepo:             stockRepo,
		productRepo:           productRepo,
		activityLogRepo:       activityLogRepo,
		broadcaster:           broadcaster,
	}
}

// SetLedgerRepo menyuntikkan ledger repo (opsional). Dipanggil dari wiring routes.go.
// Bila tidak diset, perubahan stok tak mencatat ledger (kontrak lama tetap berlaku).
func (u *stockOpnameUsecase) SetLedgerRepo(r repository.StockLedgerRepository) {
	u.ledgerRepo = r
}

type StartSessionInput struct {
	WarehouseID  uuid.UUID
	LocationID   *uuid.UUID
	StartedBy    uuid.UUID
	IPAddress    string
	IsBlindCount bool
}

type SubmitScanInput struct {
	SessionID     uuid.UUID
	ProductCode   string
	StockPhysical int
	InputMethod   string // manual, voice
	Note          string
	ScannedBy     uuid.UUID
	DeviceID      *uuid.UUID
	IPAddress     string
	// IdempotencyKey opsional (UUID dari klien). Bila diisi dan baris dengan
	// kunci yang sama sudah ada di sesi ini, baris itu yang dikembalikan.
	IdempotencyKey string
}

func (u *stockOpnameUsecase) StartSession(ctx context.Context, input StartSessionInput) (*domain.StockOpname, error) {
	now := time.Now()

	stockOpname := &domain.StockOpname{
		WarehouseID:     input.WarehouseID,
		LocationID:      input.LocationID,
		Status:          "in_progress",
		StartedBy:       input.StartedBy,
		StartedAt:       now,
		SnapshotTakenAt: &now, // snapshot stok sistem saat sesi dimulai
		IsBlindCount:    input.IsBlindCount,
	}

	// Kode sesi: SO-YYYYMMDD-NNNN, counter harian atomik di DB. Counter bisa
	// tertinggal dari kode yang sudah terpakai (mis. seq dikosongkan/restore DB),
	// jadi kalau bentrok ambil nomor berikutnya — jangan gagal ke user.
	const maxAttempts = 5
	var code string
	for attempt := 1; ; attempt++ {
		c, err := u.stockOpnameRepo.NextCode(ctx, now)
		if err != nil {
			return nil, err
		}
		code = c
		stockOpname.Code = code

		err = u.stockOpnameRepo.Create(ctx, stockOpname)
		if err == nil {
			break
		}
		if !errors.Is(err, domain.ErrSessionCodeTaken) || attempt >= maxAttempts {
			return nil, fmt.Errorf("failed to create stock opname session: %w", err)
		}
	}

	// Activity log
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &input.StartedBy,
		Action:     "start_stock_opname",
		EntityType: "stock_opname",
		EntityID:   &stockOpname.ID,
		NewValue:   datatypes.JSON(fmt.Sprintf(`{"code":"%s","warehouse_id":"%s"}`, code, input.WarehouseID)),
		IPAddress:  input.IPAddress,
	})

	return stockOpname, nil
}

func (u *stockOpnameUsecase) SubmitScan(ctx context.Context, input SubmitScanInput) (*domain.StockOpnameDetail, error) {
	// Retry jaringan: kalau kunci ini sudah pernah dipakai di sesi ini,
	// kembalikan baris lama — jangan menambah scan duplikat.
	if input.IdempotencyKey != "" {
		existing, err := u.stockOpnameDetailRepo.FindByIdempotencyKey(ctx, input.SessionID, input.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	// Cek session exists dan status in_progress
	session, err := u.stockOpnameRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}
	if session.Status != "in_progress" {
		return nil, fmt.Errorf("session is not active")
	}
	// Cut-off / frozen: tolak scan baru jika sesi dibekukan.
	if session.IsFrozen {
		return nil, fmt.Errorf("session is frozen; no new scans accepted")
	}

	// QR berisi product_code.
	product, err := u.productRepo.FindByCode(ctx, input.ProductCode)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Ambil stock sistem dari tabel stock
	stock, err := u.stockRepo.FindByProductAndWarehouse(ctx, product.ID, session.WarehouseID)
	if err != nil {
		return nil, err
	}

	stockSystem := 0
	if stock != nil {
		stockSystem = stock.QuantitySystem
	}

	// Hitung selisih (dilakukan di backend, JANGAN percaya nilai dari client)
	// difference := input.StockPhysical - stockSystem
	// Note: difference dihitung otomatis di DB sebagai generated column

	// Simpan detail scan
	detail := &domain.StockOpnameDetail{
		StockOpnameID: input.SessionID,
		ProductID:     product.ID,
		BoxID:         product.BoxID,
		StockSystem:   stockSystem,
		StockPhysical: input.StockPhysical,
		InputMethod:   input.InputMethod,
		Note:          input.Note,
		ScannedBy:     input.ScannedBy,
		DeviceID:      input.DeviceID,
		ScannedAt:     time.Now(),
		SyncStatus:    "synced",
	}

	if input.IdempotencyKey != "" {
		key := input.IdempotencyKey
		detail.IdempotencyKey = &key
	}

	if err := u.stockOpnameDetailRepo.Create(ctx, detail); err != nil {
		// Balapan: request kembar menang lebih dulu. Ambil baris pemenangnya.
		if input.IdempotencyKey != "" {
			if existing, findErr := u.stockOpnameDetailRepo.FindByIdempotencyKey(ctx, input.SessionID, input.IdempotencyKey); findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("failed to save scan: %w", err)
	}

	// Activity log
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &input.ScannedBy,
		Action:     "scan_product",
		EntityType: "stock_opname_detail",
		EntityID:   &detail.ID,
		NewValue:   datatypes.JSON(fmt.Sprintf(`{"product_code":"%s","stock_physical":%d,"stock_system":%d}`, input.ProductCode, input.StockPhysical, stockSystem)),
		IPAddress:  input.IPAddress,
	})

	// Reload baris yang BARU dibuat (bukan baris terbaru sesi — bisa milik scanner lain).
	created, err := u.stockOpnameDetailRepo.FindByID(ctx, detail.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload scan detail: %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("scan detail disappeared after create")
	}

	if u.broadcaster != nil {
		u.broadcaster.BroadcastNewScan(input.SessionID, created)

		// Real-time discrepancy alert: selisih > 10% dari stok sistem.
		// Threshold konsisten dengan DIFF_WARNING_THRESHOLD_PERCENT di .env.
		productName := product.Name
		if created.Product != nil {
			productName = created.Product.Name
		}

		if stockSystem > 0 {
			diff := input.StockPhysical - stockSystem
			absDiff := diff
			if absDiff < 0 {
				absDiff = -absDiff
			}
			percent := float64(absDiff) / float64(stockSystem) * 100.0
			if percent > 10.0 {
				u.broadcaster.BroadcastDiscrepancyAlert(input.SessionID, productName, diff, percent)
			}
		} else if input.StockPhysical > 10 {
			// Stok sistem 0 tapi fisik banyak — produk "misterius".
			u.broadcaster.BroadcastDiscrepancyAlert(input.SessionID, productName, input.StockPhysical, 100.0)
		}
	}

	return created, nil
}

func (u *stockOpnameUsecase) CompleteSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	session, err := u.stockOpnameRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != "in_progress" {
		return fmt.Errorf("session is not active")
	}

	now := time.Now()
	session.Status = "completed"
	session.CompletedAt = &now
	session.CutOffAt = &now   // cut-off: stok dibekukan pada titik ini
	session.IsFrozen = true   // tolak scan baru setelah cut-off

	if err := u.stockOpnameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to complete session: %w", err)
	}

	if u.broadcaster != nil {
		u.broadcaster.BroadcastSessionUpdate(sessionID, session)
	}

	// Activity log
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &userID,
		Action:     "complete_stock_opname",
		EntityType: "stock_opname",
		EntityID:   &sessionID,
	})

	return nil
}

func (u *stockOpnameUsecase) ApproveSession(ctx context.Context, sessionID, approverID uuid.UUID) error {
	// Approve + tulis stock dalam satu transaksi.
	// Stok sistem ikut diperbarui supaya opname berikutnya membandingkan ke angka baru.
	if err := u.stockOpnameRepo.ApproveWithStock(ctx, sessionID, approverID); err != nil {
		return err
	}

	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &approverID,
		Action:     "approve_stock_opname",
		EntityType: "stock_opname",
		EntityID:   &sessionID,
	})

	return nil
}

// RejectSession menolak sesi yang sudah complete. Hanya supervisor/admin.
func (u *stockOpnameUsecase) RejectSession(ctx context.Context, sessionID, rejectedBy uuid.UUID, reason string) error {
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}
	session, err := u.stockOpnameRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != "completed" {
		return fmt.Errorf("only completed sessions can be rejected")
	}

	now := time.Now()
	session.Status = "rejected"
	session.RejectedBy = &rejectedBy
	session.RejectedAt = &now
	session.RejectionReason = reason

	if err := u.stockOpnameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to reject session: %w", err)
	}

	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &rejectedBy,
		Action:     "reject_stock_opname",
		EntityType: "stock_opname",
		EntityID:   &sessionID,
		NewValue:   datatypes.JSON(fmt.Sprintf(`{"reason":"%s"}`, reason)),
	})

	return nil
}

// ReopenSession membuka kembali sesi yang ditolak. Hanya admin.
func (u *stockOpnameUsecase) ReopenSession(ctx context.Context, sessionID, reopenedBy uuid.UUID) error {
	session, err := u.stockOpnameRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != "rejected" {
		return fmt.Errorf("only rejected sessions can be reopened")
	}

	session.Status = "in_progress"
	session.IsFrozen = false
	session.CutOffAt = nil
	session.RejectedBy = nil
	session.RejectedAt = nil
	session.RejectionReason = ""

	if err := u.stockOpnameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to reopen session: %w", err)
	}

	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &reopenedBy,
		Action:     "reopen_stock_opname",
		EntityType: "stock_opname",
		EntityID:   &sessionID,
	})

	return nil
}

func (u *stockOpnameUsecase) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.StockOpname, error) {
	session, err := u.stockOpnameRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (u *stockOpnameUsecase) GetSessionDetails(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]domain.StockOpnameDetail, int64, error) {
	return u.stockOpnameDetailRepo.FindByStockOpnameID(ctx, sessionID, limit, offset)
}

func (u *stockOpnameUsecase) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*domain.SessionSummary, error) {
	// Get session
	session, err := u.stockOpnameRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Get all details (no limit, just for counting)
	details, _, err := u.stockOpnameDetailRepo.FindByStockOpnameID(ctx, sessionID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Count matched vs unmatched
	matched := 0
	unmatched := 0
	for _, detail := range details {
		if detail.Difference == 0 {
			matched++
		} else {
			unmatched++
		}
	}

	// Target sesi = jumlah produk yang PUNYA stok di gudang sesi (bukan semua produk
	// global). Ini yang menentukan progress & kapan sesi dinyatakan "selesai".
	totalTarget, err := u.stockRepo.CountProductsInWarehouse(ctx, session.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to count target products: %w", err)
	}
	// Jangan pernah lebih kecil dari yang sudah ter-scan (bisa terjadi bila stok
	// baris dihapus setelah scan), agar progress tidak mundur.
	if totalTarget < len(details) {
		totalTarget = len(details)
	}

	// Calculate duration
	var completedAt time.Time
	if session.CompletedAt != nil {
		completedAt = *session.CompletedAt
	} else {
		completedAt = time.Now()
	}
	duration := completedAt.Sub(session.StartedAt).Seconds()

	return &domain.SessionSummary{
		SessionID:       session.ID,
		Code:            session.Code,
		TotalScanned:    len(details),
		TotalProducts:   totalTarget,
		MatchedCount:    matched,
		UnmatchedCount:  unmatched,
		Status:          session.Status,
		StartedAt:       session.StartedAt,
		CompletedAt:     session.CompletedAt,
		DurationSeconds: int64(duration),
	}, nil
}

// GetProgressByBox mengembalikan progress scan per rak/box dalam satu sesi.
func (u *stockOpnameUsecase) GetProgressByBox(ctx context.Context, sessionID uuid.UUID) ([]domain.BoxProgress, error) {
	return u.stockOpnameDetailRepo.ProgressByBox(ctx, sessionID)
}

// CompareSessions membandingkan hasil fisik 2 sesi SO per produk + ringkasan.
func (u *stockOpnameUsecase) CompareSessions(ctx context.Context, prevID, currID uuid.UUID) (*domain.SessionCompare, error) {
	prev, err := u.stockOpnameRepo.FindByID(ctx, prevID)
	if err != nil || prev == nil {
		return nil, fmt.Errorf("sesi sebelumnya tidak ditemukan")
	}
	curr, err := u.stockOpnameRepo.FindByID(ctx, currID)
	if err != nil || curr == nil {
		return nil, fmt.Errorf("sesi sekarang tidak ditemukan")
	}

	rows, err := u.stockOpnameDetailRepo.CompareSessions(ctx, prevID, currID)
	if err != nil {
		return nil, err
	}

	summary := domain.SessionCompareSummary{}
	for _, r := range rows {
		switch {
		case r.PrevPhysical > 0 && r.CurrPhysical > 0:
			summary.InBoth++
		case r.PrevPhysical > 0:
			summary.OnlyInPrev++
		case r.CurrPhysical > 0:
			summary.OnlyInCurr++
		}
		if r.PrevPhysical != r.CurrPhysical {
			summary.Changed++
		}
	}

	return &domain.SessionCompare{
		PrevSession: prev,
		CurrSession: curr,
		Rows:        rows,
		Summary:     summary,
	}, nil
}

func (u *stockOpnameUsecase) ListSessions(ctx context.Context, filter repository.StockOpnameFilter) ([]domain.StockOpname, int64, error) {
	return u.stockOpnameRepo.List(ctx, filter)
}

func (u *stockOpnameUsecase) GetProductTrend(ctx context.Context, productID uuid.UUID, from, to *time.Time, limit int) ([]domain.ProductStockTrend, error) {
	return u.stockOpnameDetailRepo.FindProductTrend(ctx, productID, from, to, limit)
}

func (u *stockOpnameUsecase) ReconcileWarehouses(ctx context.Context, warehouseIDs []uuid.UUID) ([]domain.WarehouseReconciliation, error) {
	return u.stockRepo.ReconcileWarehouses(ctx, warehouseIDs)
}

func (u *stockOpnameUsecase) AdjustStock(ctx context.Context, input AdjustStockInput) (*domain.Stock, error) {
	// Validasi alasan
	validReasons := map[string]bool{
		domain.AdjustReasonDamaged: true,
		domain.AdjustReasonLost:    true,
		domain.AdjustReasonExpired: true,
		domain.AdjustReasonFound:   true,
		domain.AdjustReasonTypoFix: true,
		domain.AdjustReasonOther:   true,
	}
	if !validReasons[input.Reason] {
		return nil, fmt.Errorf("alasan tidak valid: %s", input.Reason)
	}
	if input.Reason == domain.AdjustReasonOther && input.Note == "" {
		return nil, fmt.Errorf("catatan wajib diisi jika alasan 'other'")
	}
	if input.Delta == 0 {
		return nil, fmt.Errorf("delta tidak boleh nol")
	}

	stock, err := u.stockRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	if stock == nil {
		return nil, fmt.Errorf("stok tidak ditemukan untuk produk di gudang ini")
	}

	oldVal, newVal, err := u.stockRepo.AdjustStock(ctx, input.ProductID, input.WarehouseID, input.Delta, stock.Version)
	if err != nil {
		return nil, err
	}

	// Catat ke stock_ledger agar saldo ledger tak menyimpang dari quantity_system.
	// Dilakukan SETELAH stok berhasil diubah; kegagalan ledger tidak menggagalkan
	// adjustment (stok sudah tersimpan), hanya dicatat sebagai peringatan.
	if u.ledgerRepo != nil {
		qtyIn, qtyOut := input.Delta, 0
		if input.Delta < 0 {
			qtyIn, qtyOut = 0, -input.Delta
		}
		if lerr := u.ledgerRepo.Create(ctx, &domain.StockLedger{
			ProductID:       input.ProductID,
			WarehouseID:     input.WarehouseID,
			TransactionType: domain.LedgerTypeAdjustment,
			ReferenceType:   "manual",
			QuantityIn:      qtyIn,
			QuantityOut:     qtyOut,
			BalanceAfter:    newVal,
			Note:            fmt.Sprintf("%s: %s", input.Reason, input.Note),
			CreatedBy:       &input.UserID,
		}); lerr != nil {
			fmt.Printf("warning: gagal mencatat ledger adjustment: %v\n", lerr)
		}
	}

	// Activity log untuk audit trail
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &input.UserID,
		Action:     "stock_adjustment",
		EntityType: "stock",
		EntityID:   &stock.ID,
		OldValue:   datatypes.JSON(fmt.Sprintf(`{"quantity_system":%d}`, oldVal)),
		NewValue:   datatypes.JSON(fmt.Sprintf(`{"quantity_system":%d,"delta":%d,"reason":"%s","note":"%s"}`, newVal, input.Delta, input.Reason, input.Note)),
		IPAddress:  input.IPAddress,
	})

	// Refresh stok dari DB
	updated, err := u.stockRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// ===== END FILE =====
