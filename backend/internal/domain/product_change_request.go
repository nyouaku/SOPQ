// ===== FILE: backend/internal/domain/product_change_request.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Status pengajuan perubahan master produk.
const (
	ChangeStatusPending  = "pending"
	ChangeStatusApproved = "approved"
	ChangeStatusRejected = "rejected"
)

// Tipe perubahan.
const (
	ChangeTypeCreate = "create"
	ChangeTypeUpdate = "update"
	ChangeTypeDelete = "delete"
)

// ProductChangeRequest — pengajuan perubahan master produk yang perlu
// disetujui supervisor/admin sebelum ditulis ke tabel products.
type ProductChangeRequest struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ChangeType  string         `json:"change_type" gorm:"not null"`
	ProductID   *uuid.UUID     `json:"product_id,omitempty" gorm:"type:uuid"`
	Product     *Product       `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Payload     datatypes.JSON `json:"payload" gorm:"type:jsonb;not null"`
	Reason      string         `json:"reason"`
	Status      string         `json:"status" gorm:"not null;default:'pending'"`
	RequestedBy uuid.UUID      `json:"requested_by" gorm:"type:uuid;not null"`
	Requester   *User          `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
	ReviewedBy  *uuid.UUID     `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	Reviewer    *User          `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
	ReviewedAt  *time.Time     `json:"reviewed_at,omitempty"`
	ReviewNote  string         `json:"review_note,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (ProductChangeRequest) TableName() string {
	return "product_change_requests"
}

// ===== END FILE =====
