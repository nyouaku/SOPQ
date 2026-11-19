// ===== FILE: backend/internal/domain/product.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string     `json:"name" gorm:"not null"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	Parent    *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Category) TableName() string {
	return "categories"
}

type Supplier struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code          string    `json:"code" gorm:"uniqueIndex;not null"`
	Name          string    `json:"name" gorm:"not null"`
	ContactPerson string    `json:"contact_person"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Address       string    `json:"address"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Supplier) TableName() string {
	return "suppliers"
}

type Product struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProductCode  string     `json:"product_code" gorm:"uniqueIndex:idx_products_code;not null"`
	Name         string     `json:"name" gorm:"not null"`
	Brand        string     `json:"brand"`
	Color        string     `json:"color"`
	Size         string     `json:"size"`
	Unit         string     `json:"unit" gorm:"not null;default:'PCS'"`
	LocationID   *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	Location     *Location  `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	MinimumStock int        `json:"minimum_stock" gorm:"default:0"`
	SupplierID   *uuid.UUID `json:"supplier_id,omitempty" gorm:"type:uuid"`
	Supplier     *Supplier  `json:"supplier,omitempty" gorm:"foreignKey:SupplierID"`
	PhotoURL     string     `json:"photo_url"`
	Status       string     `json:"status" gorm:"default:'active'"` // active, inactive, discontinued
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Product) TableName() string {
	return "products"
}
// ===== END FILE =====
