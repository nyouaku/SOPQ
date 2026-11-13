// ===== FILE: backend/internal/domain/notification.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid;index:idx_notifications_user"`
	User      *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Type      string         `json:"type" gorm:"not null"` // diff_over_threshold, sync_failed, dll
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Metadata  datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
// ===== END FILE =====
