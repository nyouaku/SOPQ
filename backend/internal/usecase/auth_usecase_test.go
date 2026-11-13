// ===== FILE: backend/internal/usecase/auth_usecase_test.go =====
package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
)

// ---------------------------------------------------------------- mocks

type mockUserRepository struct {
	users map[string]*domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.Username]; exists {
		return errors.New("username already exists")
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return m.users[username], nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for username, user := range m.users {
		if user.ID == id {
			delete(m.users, username)
			return nil
		}
	}
	return errors.New("user not found")
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *mockUserRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return []domain.Role{{ID: 1, Name: "staff"}}, nil
}

type mockSessionRepository struct {
	sessions map[uuid.UUID]*domain.Session
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return m.sessions[id], nil
}

func (m *mockSessionRepository) FindByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	for _, session := range m.sessions {
		if session.RefreshTokenHash == tokenHash {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	for _, session := range m.sessions {
		if session.RefreshTokenHash == tokenHash &&
			session.RevokedAt == nil &&
			session.ExpiresAt.After(time.Now()) {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockSessionRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	if session, ok := m.sessions[id]; ok {
		now := time.Now()
		session.RevokedAt = &now
	}
	return nil
}

func (m *mockSessionRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	for _, session := range m.sessions {
		if session.UserID == userID {
			session.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

func (m *mockSessionRepository) RotateRefreshToken(ctx context.Context, oldSessionID uuid.UUID, newSession *domain.Session) error {
	old, ok := m.sessions[oldSessionID]
	if !ok || old.RevokedAt != nil {
		return domain.ErrRefreshTokenReused
	}
	now := time.Now()
	old.RevokedAt = &now
	if newSession.ID == uuid.Nil {
		newSession.ID = uuid.New()
	}
	m.sessions[newSession.ID] = newSession
	return nil
}

// ---------------------------------------------------------------- helpers

func newTestAuthUsecase() (AuthUsecase, *mockUserRepository, *mockSessionRepository) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	sessionRepo := &mockSessionRepository{sessions: make(map[uuid.UUID]*domain.Session)}
	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret",
			RefreshSecret: "test-refresh-secret",
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    7 * 24 * time.Hour,
		},
	}
	return NewAuthUsecase(userRepo, sessionRepo, &mockActivityLogRepository{}, cfg), userRepo, sessionRepo
}

func mustRegister(t *testing.T, uc AuthUsecase, username, password string) *domain.User {
	t.Helper()
	user, err := uc.Register(context.Background(), RegisterInput{
		Username: username,
		Password: password,
		FullName: "Test User",
		RoleID:   1,
	})
	if err != nil {
		t.Fatalf("Register(%q) error = %v", username, err)
	}
	return user
}

// ---------------------------------------------------------------- tests

func TestAuthUsecase_Register(t *testing.T) {
	uc, _, _ := newTestAuthUsecase()

	user := mustRegister(t, uc, "testuser", "password1234")
	if user.Username != "testuser" {
		t.Errorf("username = %q, want testuser", user.Username)
	}
	if user.PasswordHash == "password1234" {
		t.Error("password stored in plaintext")
	}

	// Username duplikat harus ditolak.
	_, err := uc.Register(context.Background(), RegisterInput{
		Username: "testuser",
		Password: "password1234",
		FullName: "Test User 2",
		RoleID:   1,
	})
	if err == nil {
		t.Error("Register() with duplicate username should fail")
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	uc, _, _ := newTestAuthUsecase()
	user := mustRegister(t, uc, "testuser", "password1234")
	// Role di-preload oleh repository asli; mock harus menyediakannya
	// karena Login membaca user.Role.Name.
	user.Role = &domain.Role{ID: 1, Name: "staff"}

	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{"Valid login", "testuser", "password1234", false},
		{"Invalid password", "testuser", "wrongpassword", true},
		{"Non-existent user", "nonexistent", "password1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Login(context.Background(), LoginInput{
				Username: tt.username,
				Password: tt.password,
			})
			if (err != nil) != tt.expectError {
				t.Fatalf("Login() error = %v, expectError %v", err, tt.expectError)
			}
			if tt.expectError {
				return
			}
			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Error("Login() returned empty tokens")
			}
			if resp.ExpiresIn != int64((15 * time.Minute).Seconds()) {
				t.Errorf("ExpiresIn = %d, want 900", resp.ExpiresIn)
			}
		})
	}
}

func TestAuthUsecase_RefreshToken(t *testing.T) {
	uc, _, sessionRepo := newTestAuthUsecase()
	user := mustRegister(t, uc, "testuser", "password1234")
	user.Role = &domain.Role{ID: 1, Name: "staff"}

	if _, err := uc.Login(context.Background(), LoginInput{Username: "testuser", Password: "password1234"}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// Refresh token asli.
	refreshResp, err := uc.Login(context.Background(), LoginInput{Username: "testuser", Password: "password1234"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	newResp, err := uc.RefreshToken(context.Background(), refreshResp.RefreshToken, DeviceInfo{
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if newResp.AccessToken == "" || newResp.RefreshToken == "" {
		t.Error("RefreshToken() returned empty tokens")
	}

	// Token tidak valid harus ditolak.
	if _, err := uc.RefreshToken(context.Background(), "invalid-token", DeviceInfo{}); err == nil {
		t.Error("RefreshToken() with invalid token should fail")
	}

	// Regresi BE-03: token yang sudah dipakai dirotasi tidak boleh dipakai lagi.
	if _, err := uc.RefreshToken(context.Background(), refreshResp.RefreshToken, DeviceInfo{}); !errors.Is(err, domain.ErrRefreshTokenReused) {
		t.Errorf("reusing a rotated refresh token: got %v, want ErrRefreshTokenReused", err)
	}

	// Setelah pemakaian ulang terdeteksi, SEMUA sesi user dicabut.
	for id, session := range sessionRepo.sessions {
		if session.UserID == user.ID && session.RevokedAt == nil {
			t.Errorf("session %s still active after refresh-token reuse", id)
		}
	}
}

// Regresi SEC-13: logout mencabut sesi user, bukan mencari baris ber-ID user.
func TestAuthUsecase_Logout_RevokesAllSessions(t *testing.T) {
	uc, _, sessionRepo := newTestAuthUsecase()
	user := mustRegister(t, uc, "testuser", "password1234")
	user.Role = &domain.Role{ID: 1, Name: "staff"}

	for range 2 {
		if _, err := uc.Login(context.Background(), LoginInput{Username: "testuser", Password: "password1234"}); err != nil {
			t.Fatalf("Login() error = %v", err)
		}
	}
	if len(sessionRepo.sessions) == 0 {
		t.Fatal("expected sessions after login")
	}

	if err := uc.Logout(context.Background(), user.ID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	for id, session := range sessionRepo.sessions {
		if session.RevokedAt == nil {
			t.Errorf("session %s not revoked by logout", id)
		}
	}
}

// ===== END FILE =====
