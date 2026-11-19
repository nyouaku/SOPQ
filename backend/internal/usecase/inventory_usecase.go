// ===== FILE: backend/internal/usecase/inventory_usecase.go =====
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
)

// InventoryUsecase mencatat pergerakan stok harian (IN/OUT/transfer antar gudang).
// Ini terpisah dari Stock Opname: SO = audit fisik berkala; inventory = pergerakan harian.
// Setiap posting mengubah quantity_system DAN menulis stock_ledger dalam SATU transaksi,
// sehingga saldo ledger tidak pernah menyimpang dari stok sistem.
type InventoryUsecase interface {
	PostIn(ctx context.Context, input PostMovementInput) (*domain.StockMovement, error)
	PostOut(ctx context.Context, input PostMovementInput) (*domain.StockMovement, error)
	PostTransfer(ctx context.Context, input PostTransferInput) (*domain.StockMovement, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error)
	List(ctx context.Context, filter repository.InventoryFilter) ([]domain.StockMovement, int64, error)
	// PostGrid menyimpan banyak isian grid (produk × tanggal) sekaligus. Item
	// dikelompokkan per tanggal -> satu dokumen per tanggal, tiap dokumen atomic.
	PostGrid(ctx context.Context, input PostGridInput) (*GridResult, error)
	// CopyGrid menyalin seluruh isian grid dari bulan sumber ke bulan tujuan
	// (produk & hari sama). Mengembalikan ringkasan hasil.
	CopyGrid(ctx context.Context, input CopyGridInput) (*GridResult, error)
	// GetGrid mengembalikan agregat isian per produk × tanggal utk mengisi grid.
	GetGrid(ctx context.Context, q GridQuery) ([]domain.GridCell, error)
	// GetSummary mengembalikan KPI dashboard (server-side, bukan agregat klien).
	GetSummary(ctx context.Context, q SummaryQuery) (*domain.InventorySummary, error)
	// SetLowStockDeps menyuntikkan depensi utk peringatan stok menipis otomatis.
	SetLowStockDeps(notifRepo repository.NotificationRepository, productRepo repository.ProductRepository, stockRepo repository.StockRepository)
	// ExportGridToExcel menghasilkan file Excel grid (produk × tanggal) utk unduh.
	ExportGridToExcel(ctx context.Context, q GridQuery) (*excelize.File, string, error)
	// CancelMovement membatalkan dokumen pergerakan: stok dikembalikan ke posisi
	// sebelum dokumen, dan dokumen ditandai 'cancelled'. Dokumen tidak dihapus.
	CancelMovement(ctx context.Context, movementID, userID uuid.UUID, reason, ip string) error
}

type inventoryUsecase struct {
	repo           repository.InventoryRepository
	activityLogRepo repository.ActivityLogRepository
	// notifRepo + productRepo + stockRepo dipakai utk peringatan stok menipis.
	// Nil (mis. di sebagian test) = lewati notifikasi tanpa error.
	notifRepo   repository.NotificationRepository
	productRepo repository.ProductRepository
	stockRepo   repository.StockRepository
}

func NewInventoryUsecase(
	repo repository.InventoryRepository,
	activityLogRepo repository.ActivityLogRepository,
) InventoryUsecase {
	return &inventoryUsecase{repo: repo, activityLogRepo: activityLogRepo}
}

// SetLowStockDeps menyuntikkan notifikasi + product + stock repo (dipanggil dari wiring).
// Bila tak diset, peringatan stok menipis dilewati.
func (u *inventoryUsecase) SetLowStockDeps(
	notifRepo repository.NotificationRepository,
	productRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
) {
	u.notifRepo = notifRepo
	u.productRepo = productRepo
	u.stockRepo = stockRepo
}

// MovementItemInput satu baris produk dalam dokumen pergerakan.
type MovementItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitCost  float64
	Note      string
	// MovementDate: tanggal efektif transaksi (dipakai grid entry). Nil = form biasa.
	MovementDate *time.Time
}

// PostMovementInput untuk IN dan OUT (satu gudang).
type PostMovementInput struct {
	WarehouseID   uuid.UUID
	SubType       string
	ReferenceCode string
	Note          string
	UserID        uuid.UUID
	IPAddress     string
	Items         []MovementItemInput
}

// PostTransferInput untuk transfer antar gudang.
type PostTransferInput struct {
	FromWarehouseID uuid.UUID
	ToWarehouseID   uuid.UUID
	ReferenceCode   string
	Note            string
	UserID          uuid.UUID
	IPAddress       string
	Items           []MovementItemInput
}

// validasiSubType memastikan sub_type diizinkan untuk tipe pergerakan tsb.
func validasiSubType(movementType, subType string) error {
	allowed := map[string][]string{
		domain.MovementTypeIn:       {domain.MovementSubPurchase, domain.MovementSubReturn, domain.MovementSubOther},
		domain.MovementTypeOut:      {domain.MovementSubSale, domain.MovementSubReturn, domain.MovementSubDamaged, domain.MovementSubLost, domain.MovementSubExpired, domain.MovementSubOther},
		domain.MovementTypeTransfer: {domain.MovementSubTransfer},
	}
	for _, s := range allowed[movementType] {
		if s == subType {
			return nil
		}
	}
	return fmt.Errorf("sub_type '%s' tidak valid untuk tipe '%s'", subType, movementType)
}

// build membangun entity movement + efek stok + baris ledger untuk satu gudang (IN/OUT).
func (u *inventoryUsecase) build(
	ctx context.Context,
	movementType string,
	input PostMovementInput,
) (*domain.StockMovement, []repository.StockEffect, []*domain.StockLedger, error) {
	if len(input.Items) == 0 {
		return nil, nil, nil, fmt.Errorf("minimal satu item produk")
	}
	if input.WarehouseID == uuid.Nil {
		return nil, nil, nil, fmt.Errorf("gudang wajib dipilih")
	}
	if err := validasiSubType(movementType, input.SubType); err != nil {
		return nil, nil, nil, err
	}

	code, err := u.repo.NextCode(ctx, time.Now())
	if err != nil {
		return nil, nil, nil, err
	}

	movement := &domain.StockMovement{
		Code:          code,
		MovementType:  movementType,
		SubType:       input.SubType,
		WarehouseID:   input.WarehouseID,
		ReferenceCode: input.ReferenceCode,
		Note:          input.Note,
		Status:        domain.MovementStatusPosted,
		CreatedBy:     input.UserID,
	}

	ledgerType := domain.LedgerTypeFor(movementType, input.SubType)
	var effects []repository.StockEffect
	var entries []*domain.StockLedger
	userID := input.UserID

	for _, it := range input.Items {
		if it.Quantity <= 0 {
			return nil, nil, nil, fmt.Errorf("jumlah harus lebih dari 0")
		}
		movement.Items = append(movement.Items, domain.StockMovementItem{
			ProductID:    it.ProductID,
			Quantity:     it.Quantity,
			UnitCost:     it.UnitCost,
			MovementDate: it.MovementDate,
			Note:         it.Note,
		})

		delta := it.Quantity // IN: naik
		qtyIn := it.Quantity
		qtyOut := 0
		if movementType == domain.MovementTypeOut {
			delta = -it.Quantity // OUT: turun
			qtyIn = 0
			qtyOut = it.Quantity
		}
		effects = append(effects, repository.StockEffect{
			ProductID:   it.ProductID,
			WarehouseID: input.WarehouseID,
			Delta:       delta,
		})
		entries = append(entries, &domain.StockLedger{
			ProductID:       it.ProductID,
			WarehouseID:     input.WarehouseID,
			TransactionType: ledgerType,
			ReferenceType:   "inventory",
			QuantityIn:      qtyIn,
			QuantityOut:     qtyOut,
			UnitCost:        it.UnitCost,
			TotalCost:       float64(it.Quantity) * it.UnitCost,
			Note:            it.Note,
			CreatedBy:       &userID,
		})
	}

	return movement, effects, entries, nil
}

func (u *inventoryUsecase) PostIn(ctx context.Context, input PostMovementInput) (*domain.StockMovement, error) {
	movement, effects, entries, err := u.build(ctx, domain.MovementTypeIn, input)
	if err != nil {
		return nil, err
	}
	if err := u.repo.PostMovementTx(ctx, movement, effects, entries); err != nil {
		return nil, err
	}
	u.log(ctx, input.UserID, input.IPAddress, movement)
	return movement, nil
}

func (u *inventoryUsecase) PostOut(ctx context.Context, input PostMovementInput) (*domain.StockMovement, error) {
	movement, effects, entries, err := u.build(ctx, domain.MovementTypeOut, input)
	if err != nil {
		return nil, err
	}
	if err := u.repo.PostMovementTx(ctx, movement, effects, entries); err != nil {
		return nil, err
	}
	u.log(ctx, input.UserID, input.IPAddress, movement)
	// Peringatan stok menipis otomatis (event-driven, tanpa cron). Gagal notifikasi
	// tak menggagalkan posting — stok sudah tersimpan.
	u.notifyLowStock(ctx, input.WarehouseID, effects)
	return movement, nil
}

func (u *inventoryUsecase) PostTransfer(ctx context.Context, input PostTransferInput) (*domain.StockMovement, error) {
	if input.FromWarehouseID == uuid.Nil || input.ToWarehouseID == uuid.Nil {
		return nil, fmt.Errorf("gudang asal dan tujuan wajib dipilih")
	}
	if input.FromWarehouseID == input.ToWarehouseID {
		return nil, fmt.Errorf("gudang asal dan tujuan tidak boleh sama")
	}
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("minimal satu item produk")
	}

	code, err := u.repo.NextCode(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	movement := &domain.StockMovement{
		Code:          code,
		MovementType:  domain.MovementTypeTransfer,
		SubType:       domain.MovementSubTransfer,
		WarehouseID:   input.FromWarehouseID,
		ToWarehouseID: &input.ToWarehouseID,
		ReferenceCode: input.ReferenceCode,
		Note:          input.Note,
		Status:        domain.MovementStatusPosted,
		CreatedBy:     input.UserID,
	}

	var effects []repository.StockEffect
	var entries []*domain.StockLedger
	userID := input.UserID

	for _, it := range input.Items {
		if it.Quantity <= 0 {
			return nil, fmt.Errorf("jumlah harus lebih dari 0")
		}
		movement.Items = append(movement.Items, domain.StockMovementItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitCost:  it.UnitCost,
			Note:      it.Note,
		})

		// Gudang asal: turun (transfer_out)
		effects = append(effects, repository.StockEffect{
			ProductID:   it.ProductID,
			WarehouseID: input.FromWarehouseID,
			Delta:       -it.Quantity,
		})
		entries = append(entries, &domain.StockLedger{
			ProductID:       it.ProductID,
			WarehouseID:     input.FromWarehouseID,
			TransactionType: domain.LedgerTypeTransferOut,
			ReferenceType:   "inventory",
			QuantityOut:     it.Quantity,
			UnitCost:        it.UnitCost,
			TotalCost:       float64(it.Quantity) * it.UnitCost,
			Note:            it.Note,
			CreatedBy:       &userID,
		})

		// Gudang tujuan: naik (transfer_in)
		effects = append(effects, repository.StockEffect{
			ProductID:   it.ProductID,
			WarehouseID: input.ToWarehouseID,
			Delta:       it.Quantity,
		})
		entries = append(entries, &domain.StockLedger{
			ProductID:       it.ProductID,
			WarehouseID:     input.ToWarehouseID,
			TransactionType: domain.LedgerTypeTransferIn,
			ReferenceType:   "inventory",
			QuantityIn:      it.Quantity,
			UnitCost:        it.UnitCost,
			TotalCost:       float64(it.Quantity) * it.UnitCost,
			Note:            it.Note,
			CreatedBy:       &userID,
		})
	}

	if err := u.repo.PostMovementTx(ctx, movement, effects, entries); err != nil {
		return nil, err
	}
	u.log(ctx, input.UserID, input.IPAddress, movement)
	return movement, nil
}

func (u *inventoryUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *inventoryUsecase) List(ctx context.Context, filter repository.InventoryFilter) ([]domain.StockMovement, int64, error) {
	return u.repo.List(ctx, filter)
}

// GridEntryInput satu sel isian grid (produk × tanggal).
type GridEntryInput struct {
	ProductID    uuid.UUID
	Date         time.Time
	Quantity     int
	UnitCost     float64
}

// PostGridInput — body untuk posting seluruh isian grid sekaligus.
type PostGridInput struct {
	WarehouseID uuid.UUID
	MovementType string // in | out
	SubType      string
	ReferenceCode string
	UserID       uuid.UUID
	IPAddress    string
	Entries      []GridEntryInput
}

// GridQuery — parameter baca isian grid satu bulan.
type GridQuery struct {
	WarehouseID  uuid.UUID
	Month        time.Time
	MovementType string
}

// SummaryQuery — parameter KPI dashboard.
type SummaryQuery struct {
	WarehouseID uuid.UUID
	Month       time.Time
}

// CopyGridInput — parameter salin isian grid bulan sumber ke bulan tujuan.
type CopyGridInput struct {
	WarehouseID  uuid.UUID
	SrcMonth     time.Time
	DstMonth     time.Time
	MovementType string // in | out
	SubType      string
	UserID       uuid.UUID
	IPAddress    string
}

// GridResult — ringkasan hasil posting grid.
type GridResult struct {
	MovementsCreated int   `json:"movements_created"`
	ItemsPosted      int   `json:"items_posted"`
}

// PostGrid menyimpan seluruh isian grid (produk × tanggal) sekaligus.
// Item dikelompokkan per tanggal -> satu dokumen per tanggal (tiap dokumen atomic),
// sehingga stok + ledger per tanggal tetap konsisten dan bisa diaudit per hari.
func (u *inventoryUsecase) PostGrid(ctx context.Context, input PostGridInput) (*GridResult, error) {
	if len(input.Entries) == 0 {
		return nil, fmt.Errorf("tidak ada isian grid untuk disimpan")
	}
	if input.WarehouseID == uuid.Nil {
		return nil, fmt.Errorf("gudang wajib dipilih")
	}
	if input.MovementType != domain.MovementTypeIn && input.MovementType != domain.MovementTypeOut {
		return nil, fmt.Errorf("movement_type harus 'in' atau 'out'")
	}
	if err := validasiSubType(input.MovementType, input.SubType); err != nil {
		return nil, err
	}

	// Kelompokkan isian per tanggal.
	byDate := make(map[string][]MovementItemInput)
	var dates []time.Time
	for _, e := range input.Entries {
		if e.Quantity <= 0 {
			return nil, fmt.Errorf("jumlah harus lebih dari 0")
		}
		if e.Date.IsZero() {
			return nil, fmt.Errorf("setiap isian wajib punya tanggal")
		}
		day := e.Date.Format("2006-01-02")
		if _, ok := byDate[day]; !ok {
			dates = append(dates, e.Date)
		}
		byDate[day] = append(byDate[day], MovementItemInput{
			ProductID:    e.ProductID,
			Quantity:     e.Quantity,
			UnitCost:     e.UnitCost,
			MovementDate: &e.Date,
		})
	}

	created := 0
	itemsPosted := 0
	for _, d := range dates {
		day := d
		items := byDate[day.Format("2006-01-02")]

		var err error
		if input.MovementType == domain.MovementTypeOut {
			_, err = u.PostOut(ctx, PostMovementInput{
				WarehouseID:   input.WarehouseID,
				SubType:       input.SubType,
				ReferenceCode: input.ReferenceCode,
				UserID:        input.UserID,
				IPAddress:     input.IPAddress,
				Items:         items,
			})
		} else {
			_, err = u.PostIn(ctx, PostMovementInput{
				WarehouseID:   input.WarehouseID,
				SubType:       input.SubType,
				ReferenceCode: input.ReferenceCode,
				UserID:        input.UserID,
				IPAddress:     input.IPAddress,
				Items:         items,
			})
		}
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan tanggal %s: %w", day.Format("02 Jan"), err)
		}
		created++
		itemsPosted += len(items)
	}

	return &GridResult{MovementsCreated: created, ItemsPosted: itemsPosted}, nil
}

// GetGrid mengembalikan agregat isian per produk × tanggal (pre-fill grid).
func (u *inventoryUsecase) GetGrid(ctx context.Context, q GridQuery) ([]domain.GridCell, error) {
	if q.WarehouseID == uuid.Nil {
		return nil, fmt.Errorf("gudang wajib dipilih")
	}
	if q.MovementType != domain.MovementTypeIn && q.MovementType != domain.MovementTypeOut {
		return nil, fmt.Errorf("movement_type harus 'in' atau 'out'")
	}
	return u.repo.GridCells(ctx, q.WarehouseID, q.Month, q.MovementType)
}

// CopyGrid menyalin seluruh isian grid bulan sumber ke bulan tujuan (produk & hari sama).
// Reuse GridCells (baca sumber) + PostGrid (tulis tujuan, per tanggal atomic).
func (u *inventoryUsecase) CopyGrid(ctx context.Context, input CopyGridInput) (*GridResult, error) {
	if input.WarehouseID == uuid.Nil {
		return nil, fmt.Errorf("gudang wajib dipilih")
	}
	if input.SrcMonth.Equal(input.DstMonth) {
		return nil, fmt.Errorf("bulan sumber dan tujuan tidak boleh sama")
	}
	if input.MovementType != domain.MovementTypeIn && input.MovementType != domain.MovementTypeOut {
		return nil, fmt.Errorf("movement_type harus 'in' atau 'out'")
	}

	cells, err := u.GetGrid(ctx, GridQuery{
		WarehouseID:  input.WarehouseID,
		Month:        input.SrcMonth,
		MovementType: input.MovementType,
	})
	if err != nil {
		return nil, err
	}
	if len(cells) == 0 {
		return nil, fmt.Errorf("bulan sumber tidak punya isian untuk disalin")
	}

	// Petakan tiap cell -> entry bulan tujuan (hari sama, tahun+bulan ikut tujuan).
	var entries []GridEntryInput
	for _, c := range cells {
		if c.InQty <= 0 && c.OutQty <= 0 {
			continue
		}
		qty := c.InQty
		if input.MovementType == domain.MovementTypeOut {
			qty = c.OutQty
		}
		entries = append(entries, GridEntryInput{
			ProductID: c.ProductID,
			Date: time.Date(
				input.DstMonth.Year(), input.DstMonth.Month(), c.Day,
				0, 0, 0, 0, time.UTC,
			),
			Quantity: qty,
			UnitCost: c.UnitCost,
		})
	}

	return u.PostGrid(ctx, PostGridInput{
		WarehouseID:   input.WarehouseID,
		MovementType:  input.MovementType,
		SubType:       input.SubType,
		ReferenceCode: "",
		UserID:        input.UserID,
		IPAddress:     input.IPAddress,
		Entries:       entries,
	})
}

// GetSummary mengembalikan KPI dashboard inventory.
func (u *inventoryUsecase) GetSummary(ctx context.Context, q SummaryQuery) (*domain.InventorySummary, error) {
	if q.WarehouseID == uuid.Nil {
		return nil, fmt.Errorf("gudang wajib dipilih")
	}
	return u.repo.Summary(ctx, q.WarehouseID, q.Month)
}

// ExportGridToExcel menghasilkan file Excel grid harian (produk × tanggal) untuk
// unduh — mempertahankan kebiasaan karyawan yang bekerja di Excel.
func (u *inventoryUsecase) ExportGridToExcel(ctx context.Context, q GridQuery) (*excelize.File, string, error) {
	if q.WarehouseID == uuid.Nil {
		return nil, "", fmt.Errorf("gudang wajib dipilih")
	}
	if q.MovementType != domain.MovementTypeIn && q.MovementType != domain.MovementTypeOut {
		return nil, "", fmt.Errorf("movement_type harus 'in' atau 'out'")
	}

	cells, err := u.repo.GridCells(ctx, q.WarehouseID, q.Month, q.MovementType)
	if err != nil {
		return nil, "", err
	}

	// Hitung jumlah hari di bulan tsb.
	daysInMonth := time.Date(q.Month.Year(), q.Month.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	f := excelize.NewFile()
	sheet := "Grid"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Header: Produk | Kode | 1 | 2 | ... | N
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"2E7D32"}, Pattern: 1},
	})
	f.SetCellValue(sheet, "A1", "Produk")
	f.SetCellValue(sheet, "B1", "Kode")
	f.SetCellStyle(sheet, "A1", "A1", headerStyle)
	f.SetCellStyle(sheet, "B1", "B1", headerStyle)
	for d := 1; d <= daysInMonth; d++ {
		cell, _ := excelize.CoordinatesToCellName(d+2, 1)
		f.SetCellValue(sheet, cell, d)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 14)

	// Isi baris: satu produk per baris, nilai per tanggal.
	byProduct := make(map[uuid.UUID][]domain.GridCell)
	var order []uuid.UUID
	for _, c := range cells {
		if _, ok := byProduct[c.ProductID]; !ok {
			order = append(order, c.ProductID)
		}
		byProduct[c.ProductID] = append(byProduct[c.ProductID], c)
	}

	for r, pid := range order {
		rowNum := r + 2
		f.SetCellValue(sheet, "A"+fmt.Sprint(rowNum), byProduct[pid][0].Name)
		f.SetCellValue(sheet, "B"+fmt.Sprint(rowNum), byProduct[pid][0].ProductCode)
		for _, c := range byProduct[pid] {
			qty := c.InQty
			if q.MovementType == domain.MovementTypeOut {
				qty = c.OutQty
			}
			if qty > 0 {
				cell, _ := excelize.CoordinatesToCellName(c.Day+2, rowNum)
				f.SetCellValue(sheet, cell, qty)
			}
		}
	}

	filename := fmt.Sprintf("inventory-grid-%s-%s.xlsx",
		q.MovementType, q.Month.Format("2006-01"))
	return f, filename, nil
}

// CancelMovement membatalkan dokumen pergerakan. Stok dikembalikan ke posisi
// sebelum dokumen dan dokumen ditandai 'cancelled' (tidak dihapus, agar audit utuh).
// Alasan wajib diisi — pembatalan menyentuh stok dan harus bisa dipertanggungjawabkan.
func (u *inventoryUsecase) CancelMovement(ctx context.Context, movementID, userID uuid.UUID, reason, ip string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("alasan pembatalan wajib diisi")
	}
	if err := u.repo.CancelMovement(ctx, movementID, userID, strings.TrimSpace(reason)); err != nil {
		return err
	}
	if u.activityLogRepo != nil {
		_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
			UserID:     &userID,
			Action:     "inventory_cancel",
			EntityType: "stock_movement",
			EntityID:   &movementID,
			NewValue:   datatypes.JSON(fmt.Sprintf(`{"reason":%q}`, strings.TrimSpace(reason))),
			IPAddress:  ip,
		})
	}
	return nil
}

func (u *inventoryUsecase) log(ctx context.Context, userID uuid.UUID, ip string, movement *domain.StockMovement) {	if u.activityLogRepo == nil {
		return
	}
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &userID,
		Action:     "inventory_" + movement.MovementType,
		EntityType: "stock_movement",
		EntityID:   &movement.ID,
		IPAddress:  ip,
	})
}

// notifyLowStock memeriksa tiap produk yang stoknya berkurang; bila stok kini di
// bawah minimum_stock, buat notifikasi (dengan guard dedup agar tak spam). Gagal
// notifikasi hanya dicatat, tak menggagalkan posting.
func (u *inventoryUsecase) notifyLowStock(ctx context.Context, warehouseID uuid.UUID, effects []repository.StockEffect) {
	if u.notifRepo == nil || u.productRepo == nil {
		return
	}
	for _, eff := range effects {
		if eff.Delta >= 0 {
			continue // hanya stok yang berkurang
		}
		stock, err := u.stockRepo.FindByProductAndWarehouse(ctx, eff.ProductID, warehouseID)
		if err != nil || stock == nil {
			continue
		}
		product, err := u.productRepo.FindByID(ctx, eff.ProductID)
		if err != nil || product == nil {
			continue
		}
		if stock.QuantitySystem >= product.MinimumStock {
			continue // masih di atas minimum
		}
		// Guard dedup: lewati bila sudah ada notifikasi low-stock belum dibaca.
		exists, err := u.notifRepo.HasUnreadLowStock(ctx, eff.ProductID, warehouseID)
		if err != nil || exists {
			continue
		}
		meta, _ := json.Marshal(map[string]any{
			"product_id":   eff.ProductID,
			"warehouse_id": warehouseID,
			"quantity_system": stock.QuantitySystem,
			"minimum_stock": product.MinimumStock,
		})
		_ = u.notifRepo.Create(ctx, &domain.Notification{
			Type:    domain.NotificationTypeLowStock,
			Title:   "Stok menipis: " + product.Name,
			Message: fmt.Sprintf("Stok %s tersisa %d (minimum %d)", product.Name, stock.QuantitySystem, product.MinimumStock),
			Metadata: meta,
		})
	}
}
