// ===== FILE: backend/internal/domain/stock.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Stock struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProductID      uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;uniqueIndex:idx_stock_product_warehouse"`
	Product        *Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	WarehouseID    uuid.UUID  `json:"warehouse_id" gorm:"type:uuid;not null;uniqueIndex:idx_stock_product_warehouse"`
	Warehouse      *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	QuantitySystem int        `json:"quantity_system" gorm:"not null;default:0"`
	Version        int        `json:"version" gorm:"not null;default:0"` // optimistic locking
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Stock) TableName() string {
	return "stock"
}

// WarehouseReconciliation — hasil rekonsiliasi antar gudang untuk satu produk.
type WarehouseReconciliation struct {
	ProductID       uuid.UUID `json:"product_id"`
	ProductCode     string    `json:"product_code"`
	ProductName     string    `json:"product_name"`
	WarehouseID     uuid.UUID `json:"warehouse_id"`
	WarehouseName   string    `json:"warehouse_name"`
	QuantitySystem  int       `json:"quantity_system"`
	TotalAll        int       `json:"total_all"`
	PresentInCount  int       `json:"present_in_count"`
	TotalWarehouses int       `json:"total_warehouses"`
	MissingFrom     []string  `json:"missing_from"` // nama gudang di mana produk tidak ada
}

// ===== END FILE =====
