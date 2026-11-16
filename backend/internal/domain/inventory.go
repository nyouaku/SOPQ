// ===== FILE: backend/internal/domain/inventory.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tipe pergerakan inventory (header stock_movement).
const (
	MovementTypeIn       = "in"
	MovementTypeOut      = "out"
	MovementTypeTransfer = "transfer"
)

// Sub-tipe: alasan/keperluan pergerakan. Dipetakan ke LedgerType* saat posting.
const (
	MovementSubPurchase = "purchase" // in
	MovementSubReturn   = "return"   // in (retur pelanggan) / out (retur ke supplier)
	MovementSubSale     = "sale"     // out
	MovementSubDamaged  = "damaged"  // out
	MovementSubLost     = "lost"     // out
	MovementSubExpired  = "expired"  // out
	MovementSubTransfer = "transfer" // transfer
	MovementSubOther    = "other"    // in/out
)

// Status dokumen pergerakan. Fase 1 hanya posted (langsung efektif).
const (
	MovementStatusPosted    = "posted"
	MovementStatusCancelled = "cancelled"
)

// StockMovement — header dokumen pergerakan stok (IN/OUT/transfer).
type StockMovement struct {
	ID             uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code           string              `json:"code" gorm:"uniqueIndex;not null"`
	MovementType   string              `json:"movement_type" gorm:"not null"`
	SubType        string              `json:"sub_type" gorm:"not null"`
	WarehouseID    uuid.UUID           `json:"warehouse_id" gorm:"type:uuid;not null"`
	Warehouse      *Warehouse          `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	ToWarehouseID  *uuid.UUID          `json:"to_warehouse_id,omitempty" gorm:"type:uuid"`
	ToWarehouse    *Warehouse          `json:"to_warehouse,omitempty" gorm:"foreignKey:ToWarehouseID"`
	ReferenceCode  string              `json:"reference_code"`
	Note           string              `json:"note"`
	Status         string              `json:"status" gorm:"not null;default:'posted'"`
	CreatedBy      uuid.UUID           `json:"created_by" gorm:"type:uuid;not null"`
	CreatedByUser  *User               `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	Items          []StockMovementItem `json:"items,omitempty" gorm:"foreignKey:MovementID"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      *time.Time          `json:"deleted_at,omitempty" gorm:"index"`
}

func (StockMovement) TableName() string {
	return "stock_movement"
}

// StockMovementItem — satu baris produk dalam dokumen pergerakan.
type StockMovementItem struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	MovementID uuid.UUID `json:"movement_id" gorm:"type:uuid;not null"`
	ProductID  uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Product    *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quantity   int       `json:"quantity" gorm:"not null"`
	UnitCost   float64   `json:"unit_cost" gorm:"not null;default:0"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

func (StockMovementItem) TableName() string {
	return "stock_movement_item"
}

// LedgerTypeFor memetakan (movement_type, sub_type) -> transaction_type stock_ledger.
// Reuse konstanta LedgerType* yang sudah ada; jangan buat nilai ledger baru.
func LedgerTypeFor(movementType, subType string) string {
	switch movementType {
	case MovementTypeIn:
		if subType == MovementSubReturn {
			return LedgerTypeReturnIn
		}
		return LedgerTypePurchaseIn
	case MovementTypeOut:
		if subType == MovementSubReturn {
			return LedgerTypeReturnOut
		}
		return LedgerTypeSaleOut
	case MovementTypeTransfer:
		return LedgerTypeTransferOut // gudang asal; gudang tujuan pakai TransferIn
	}
	return LedgerTypeAdjustment
}

// ===== END FILE =====
