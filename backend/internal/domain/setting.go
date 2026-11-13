// ===== FILE: backend/internal/domain/setting.go =====
package domain

import (
	"time"
)

type Setting struct {
	Key         string    `json:"key" gorm:"primaryKey"`
	Value       string    `json:"value" gorm:"not null"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}
// ===== END FILE =====
