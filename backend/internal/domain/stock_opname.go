// ===== FILE: backend/internal/domain/stock_opname.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type StockOpname struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code         string     `json:"code" gorm:"uniqueIndex;not null"` // SO-20260801-0001
	WarehouseID  uuid.UUID  `json:"warehouse_id" gorm:"type:uuid;not null;index:idx_stock_opname_warehouse"`
	Warehouse    *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	LocationID   *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	Location     *Location  `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	Status       string     `json:"status" gorm:"not null;default:'in_progress';index:idx_stock_opname_status"` // in_progress, completed, approved, cancelled
	StartedBy    uuid.UUID  `json:"started_by" gorm:"type:uuid;not null"`
	StartedUser  *User      `json:"started_user,omitempty" gorm:"foreignKey:StartedBy"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ApprovedBy   *uuid.UUID `json:"approved_by,omitempty" gorm:"type:uuid"`
	ApprovedUser *User      `json:"approved_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (StockOpname) TableName() string {
	return "stock_opname"
}

type StockOpnameDetail struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	StockOpnameID uuid.UUID    `json:"stock_opname_id" gorm:"type:uuid;not null;index:idx_so_detail_opname"`
	StockOpname   *StockOpname `json:"stock_opname,omitempty" gorm:"foreignKey:StockOpnameID"`
	ProductID     uuid.UUID    `json:"product_id" gorm:"type:uuid;not null;index:idx_so_detail_product"`
	Product       *Product     `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	BoxID         *uuid.UUID   `json:"box_id,omitempty" gorm:"type:uuid"`
	Box           *Box         `json:"box,omitempty" gorm:"foreignKey:BoxID"`
	StockSystem   int          `json:"stock_system" gorm:"not null"`
	StockPhysical int          `json:"stock_physical" gorm:"not null"`
	Difference    int          `json:"difference" gorm:"->"`                          // generated column di DB — hanya dibaca, jangan pernah ditulis
	InputMethod   string       `json:"input_method" gorm:"not null;default:'manual'"` // manual, voice
	Note          string       `json:"note"`
	PhotoURL      string       `json:"photo_url"`
	ScannedBy     uuid.UUID    `json:"scanned_by" gorm:"type:uuid;not null;index:idx_so_detail_scanned_by"`
	ScannedUser   *User        `json:"scanned_user,omitempty" gorm:"foreignKey:ScannedBy"`
	DeviceID      *uuid.UUID   `json:"device_id,omitempty" gorm:"type:uuid"`
	ScannedAt     time.Time    `json:"scanned_at"`
	SyncStatus    string       `json:"sync_status" gorm:"default:'synced'"` // synced, pending, failed
	// IdempotencyKey dikirim klien (UUID) supaya retry jaringan mengembalikan
	// baris yang sama, bukan menambah baris duplikat. NULL untuk scan web.
	IdempotencyKey *string `json:"idempotency_key,omitempty" gorm:"type:varchar(64)"`
}

func (StockOpnameDetail) TableName() string {
	return "stock_opname_detail"
}

// SessionSummary untuk completion status
type SessionSummary struct {
	SessionID      uuid.UUID  `json:"session_id"`
	Code           string     `json:"code"`
	TotalScanned   int        `json:"total_scanned"`
	TotalProducts  int        `json:"total_products"`
	MatchedCount   int        `json:"matched_count"`
	UnmatchedCount int        `json:"unmatched_count"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	DurationSeconds int64     `json:"duration_seconds"`
}

// ProductStockTrend — satu titik riwayat stok fisik produk dari hasil opname.
type ProductStockTrend struct {
	ScannedAt     time.Time `json:"scanned_at"`
	StockSystem   int       `json:"stock_system"`
	StockPhysical int       `json:"stock_physical"`
	Difference    int       `json:"difference"`
	SessionCode   string    `json:"session_code"`
}

// ===== END FILE =====
