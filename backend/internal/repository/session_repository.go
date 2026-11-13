// ===== FILE: backend/internal/repository/session_repository.go =====
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	FindActiveByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	// FindByRefreshTokenHash mengembalikan sesi terbaru dengan hash tsb tanpa
	// memfilter status — dipakai untuk membedakan "token tidak dikenal" dari
	// "token sudah dipakai" (indikasi pencurian).
	FindByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	// RotateRefreshToken mencabut sesi lama dan membuat sesi baru dalam satu
	// transaksi. Gagal bila sesi lama sudah dicabut (token dipakai ulang).
	RotateRefreshToken(ctx context.Context, oldSessionID uuid.UUID, newSession *domain.Session) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND expires_at > ? AND revoked_at IS NULL", tokenHash, now).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find active session: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) FindByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", tokenHash).
		Order("created_at DESC").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session by token hash: %w", err)
	}
	return &session, nil
}

// RevokeByID mencabut satu sesi. Sesi yang sudah dicabut atau tidak ada
// dilaporkan lewat RowsAffected == 0; pemanggil memutuskan apakah itu error.
func (r *sessionRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to revoke session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("failed to revoke sessions by user_id: %w", err)
	}
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&domain.Session{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// RotateRefreshToken mencabut sesi lama dan membuat penggantinya secara atomik.
// UPDATE bersyarat `revoked_at IS NULL` membuat token yang dipakai dua kali
// gagal pada panggilan kedua — rotasi bersifat sekali pakai.
func (r *sessionRepository) RotateRefreshToken(ctx context.Context, oldSessionID uuid.UUID, newSession *domain.Session) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&domain.Session{}).
			Where("id = ? AND revoked_at IS NULL", oldSessionID).
			Update("revoked_at", now)
		if result.Error != nil {
			return fmt.Errorf("failed to revoke old session: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return domain.ErrRefreshTokenReused
		}

		if err := tx.Create(newSession).Error; err != nil {
			return fmt.Errorf("failed to create replacement session: %w", err)
		}
		return nil
	})
}

// ===== END FILE =====
