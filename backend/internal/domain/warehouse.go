// ===== FILE: backend/internal/domain/warehouse.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Warehouse struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code       string         `json:"code" gorm:"uniqueIndex;not null"`
	Name       string         `json:"name" gorm:"not null"`
	Address    string         `json:"address"`
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	LayoutJSON datatypes.JSON `json:"layout_json,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (Warehouse) TableName() string {
	return "warehouses"
}

type Location struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	WarehouseID uuid.UUID `json:"warehouse_id" gorm:"type:uuid;not null"`
	Warehouse   *Warehouse `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	Code        string    `json:"code" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Location) TableName() string {
	return "locations"
}

type Box struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	LocationID uuid.UUID `json:"location_id" gorm:"type:uuid;not null"`
	Location   *Location `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	Code       string    `json:"code" gorm:"not null"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Box) TableName() string {
	return "boxes"
}
// ===== END FILE =====
