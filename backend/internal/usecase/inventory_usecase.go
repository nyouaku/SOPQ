// ===== FILE: backend/internal/usecase/inventory_usecase.go =====
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
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
}

type inventoryUsecase struct {
	repo           repository.InventoryRepository
	activityLogRepo repository.ActivityLogRepository
}

func NewInventoryUsecase(
	repo repository.InventoryRepository,
	activityLogRepo repository.ActivityLogRepository,
) InventoryUsecase {
	return &inventoryUsecase{repo: repo, activityLogRepo: activityLogRepo}
}

// MovementItemInput satu baris produk dalam dokumen pergerakan.
type MovementItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitCost  float64
	Note      string
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
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitCost:  it.UnitCost,
			Note:      it.Note,
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

func (u *inventoryUsecase) log(ctx context.Context, userID uuid.UUID, ip string, movement *domain.StockMovement) {
	if u.activityLogRepo == nil {
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
