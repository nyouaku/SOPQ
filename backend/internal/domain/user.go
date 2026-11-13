// ===== FILE: backend/internal/domain/user.go =====
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Username     string     `json:"username" gorm:"uniqueIndex:idx_users_username;not null"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;not null"`
	FullName     string     `json:"full_name" gorm:"not null"`
	RoleID       int        `json:"role_id" gorm:"not null"`
	Role         *Role      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

type Session struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_sessions_user"`
	RefreshTokenHash  string     `json:"-" gorm:"not null"`
	DeviceID          *uuid.UUID `json:"device_id,omitempty" gorm:"type:uuid"`
	IPAddress         string     `json:"ip_address"`
	UserAgent         string     `json:"user_agent"`
	ExpiresAt         time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (Session) TableName() string {
	return "sessions"
}
// ===== END FILE =====
