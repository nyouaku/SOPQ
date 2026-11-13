// ===== FILE: backend/internal/domain/activity_log.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ActivityLog struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID     *uuid.UUID      `json:"user_id,omitempty" gorm:"type:uuid;index:idx_activity_logs_user"`
	User       *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action     string          `json:"action" gorm:"not null"` // scan_product, login, update_product, dll
	EntityType string          `json:"entity_type"`
	EntityID   *uuid.UUID      `json:"entity_id,omitempty" gorm:"type:uuid;index:idx_activity_logs_entity"`
	OldValue   datatypes.JSON  `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue   datatypes.JSON  `json:"new_value,omitempty" gorm:"type:jsonb"`
	IPAddress  string          `json:"ip_address"`
	CreatedAt  time.Time       `json:"created_at" gorm:"index:idx_activity_logs_created"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}
// ===== END FILE =====
