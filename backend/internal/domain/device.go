// ===== FILE: backend/internal/domain/device.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID     *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DeviceName string     `json:"device_name"`
	Platform   string     `json:"platform"` // android, ios
	AppVersion string     `json:"app_version"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (Device) TableName() string {
	return "devices"
}
// ===== END FILE =====
