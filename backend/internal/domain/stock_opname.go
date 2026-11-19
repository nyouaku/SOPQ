// ===== FILE: backend/internal/domain/stock_opname.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type StockOpname struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code            string     `json:"code" gorm:"uniqueIndex;not null"`
	WarehouseID     uuid.UUID  `json:"warehouse_id" gorm:"type:uuid;not null;index:idx_stock_opname_warehouse"`
	Warehouse       *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	Location        *Location  `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	Status          string     `json:"status" gorm:"not null;default:'in_progress';index:idx_stock_opname_status"`
	StartedBy       uuid.UUID  `json:"started_by" gorm:"type:uuid;not null"`
	StartedUser     *User      `json:"started_user,omitempty" gorm:"foreignKey:StartedBy"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty" gorm:"type:uuid"`
	ApprovedUser    *User      `json:"approved_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectedBy      *uuid.UUID `json:"rejected_by,omitempty" gorm:"type:uuid"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	RecountCount    int        `json:"recount_count" gorm:"not null;default:0"`
	MaxRecount      int        `json:"max_recount" gorm:"not null;default:2"`
	SnapshotTakenAt *time.Time `json:"snapshot_taken_at,omitempty"`
	CutOffAt        *time.Time `json:"cut_off_at,omitempty"`
	IsFrozen        bool       `json:"is_frozen" gorm:"not null;default:false"`
	IsBlindCount    bool       `json:"is_blind_count" gorm:"not null;default:false"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Session status constants
const (
	SessionStatusDraft            = "draft"
	SessionStatusScheduled        = "scheduled"
	SessionStatusInProgress       = "in_progress"
	SessionStatusPendingRecount   = "pending_recount"
	SessionStatusPendingApproval  = "pending_approval"
	SessionStatusCompleted        = "completed"
	SessionStatusApproved         = "approved"
	SessionStatusRejected         = "rejected"
	SessionStatusCancelled        = "cancelled"
	SessionStatusReopened         = "reopened"
)

// Transaction types untuk stock_ledger
const (
	LedgerTypeInitial         = "initial"
	LedgerTypePurchaseIn      = "purchase_in"
	LedgerTypeSaleOut         = "sale_out"
	LedgerTypeTransferIn      = "transfer_in"
	LedgerTypeTransferOut     = "transfer_out"
	LedgerTypeAdjustment      = "adjustment"
	LedgerTypeOpnameCorrect   = "opname_correction"
	LedgerTypeReturnIn        = "return_in"
	LedgerTypeReturnOut       = "return_out"
)

func (StockOpname) TableName() string {
	return "stock_opname"
}

type StockOpnameDetail struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	StockOpnameID uuid.UUID    `json:"stock_opname_id" gorm:"type:uuid;not null;index:idx_so_detail_opname"`
	StockOpname   *StockOpname `json:"stock_opname,omitempty" gorm:"foreignKey:StockOpnameID"`
	ProductID     uuid.UUID    `json:"product_id" gorm:"type:uuid;not null;index:idx_so_detail_product"`
	Product       *Product     `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	LocationID    *uuid.UUID  `json:"location_id,omitempty" gorm:"type:uuid"`
	Location      *Location   `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	StockSystem   int         `json:"stock_system" gorm:"not null"`
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

// LocationProgress — progress scan satu lokasi/area dalam sesi SO.
type LocationProgress struct {
	LocationID   uuid.UUID `json:"location_id"`
	LocationCode string    `json:"location_code"`
	LocationName string    `json:"location_name"`
	Scanned      int       `json:"scanned"`   // jumlah produk sudah di-scan di lokasi tsb
	Target       int       `json:"target"`    // jumlah produk seharusnya ada di lokasi tsb
	Percent      float64   `json:"percent"`   // scanned/target * 100
}

// SessionCompareRow — perbandingan 1 produk antara 2 sesi SO.
type SessionCompareRow struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductCode string    `json:"product_code"`
	ProductName string    `json:"product_name"`
	PrevPhysical int      `json:"prev_physical"`
	CurrPhysical int      `json:"curr_physical"`
	Change       int      `json:"change"` // curr - prev
	PrevSystem   int      `json:"prev_system"`
	CurrSystem   int      `json:"curr_system"`
}

// SessionCompare — hasil perbandingan 2 sesi SO.
type SessionCompare struct {
	PrevSession *StockOpname `json:"prev_session"`
	CurrSession *StockOpname `json:"curr_session"`
	Rows        []SessionCompareRow `json:"rows"`
	Summary     SessionCompareSummary `json:"summary"`
}

type SessionCompareSummary struct {
	OnlyInPrev int `json:"only_in_prev"` // produk ada di prev, tak ada di curr
	OnlyInCurr int `json:"only_in_curr"`
	InBoth     int `json:"in_both"`
	Changed    int `json:"changed"` // physical berubah
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

// StockLedger — satu baris pergerakan stok (kartu stok).
type StockLedger struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProductID       uuid.UUID  `json:"product_id" gorm:"type:uuid;not null"`
	Product         *Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	WarehouseID     uuid.UUID  `json:"warehouse_id" gorm:"type:uuid;not null"`
	Warehouse       *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	TransactionType string     `json:"transaction_type" gorm:"not null"`
	ReferenceType   string     `json:"reference_type"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceCode   string     `json:"reference_code"`
	QuantityIn      int        `json:"quantity_in"`
	QuantityOut     int        `json:"quantity_out"`
	BalanceAfter    int        `json:"balance_after"`
	UnitCost        float64    `json:"unit_cost"`
	TotalCost       float64    `json:"total_cost"`
	Note            string     `json:"note"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (StockLedger) TableName() string {
	return "stock_ledger"
}

// StockAdjustmentJournal — dokumen formal penyesuaian stok.
type StockAdjustmentJournal struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProductID     uuid.UUID `json:"product_id"`
	Product       *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	WarehouseID   uuid.UUID `json:"warehouse_id"`
	Warehouse     *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	SessionID     *uuid.UUID `json:"session_id,omitempty"`
	Reason        string    `json:"reason"`
	Note          string    `json:"note"`
	QuantityBefore int      `json:"quantity_before"`
	QuantityAfter  int      `json:"quantity_after"`
	Delta         int       `json:"delta"`
	AdjustedBy    uuid.UUID `json:"adjusted_by"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	IPAddress     string    `json:"ip_address"`
	CreatedAt     time.Time `json:"created_at"`
}

func (StockAdjustmentJournal) TableName() string {
	return "stock_adjustment_journal"
}

// ===== END FILE =====
