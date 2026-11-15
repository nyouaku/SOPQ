// ===== FILE: backend/internal/usecase/auth_usecase.go =====
package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"gorm.io/datatypes"
)

type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*domain.User, error)
	Login(ctx context.Context, input LoginInput) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, deviceInfo DeviceInfo) (*LoginResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	ListRoles(ctx context.Context) ([]domain.Role, error)
	// ToggleUserStatus aktif/nonaktifkan user. Return status baru.
	ToggleUserStatus(ctx context.Context, userID uuid.UUID, isActive bool, adminID uuid.UUID) error
	// ResetUserPassword mengganti password user. Return password baru yang di-generate.
	ResetUserPassword(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) (string, error)
}

type authUsecase struct {
	userRepo        repository.UserRepository
	sessionRepo     repository.SessionRepository
	activityLogRepo repository.ActivityLogRepository
	cfg             *config.Config
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	activityLogRepo repository.ActivityLogRepository,
	cfg *config.Config,
) AuthUsecase {
	return &authUsecase{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		activityLogRepo: activityLogRepo,
		cfg:             cfg,
	}
}

type RegisterInput struct {
	Username string
	Password string
	FullName string
	RoleID   int
}

type LoginInput struct {
	Username   string
	Password   string
	DeviceInfo DeviceInfo
}

type DeviceInfo struct {
	DeviceID  *uuid.UUID
	IPAddress string
	UserAgent string
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
	ExpiresIn    int64        `json:"expires_in"` // seconds
}

// minimalPasswordLength: 12 karakter. Batas 6 lama terlalu lemah untuk akun
// yang bisa menyetujui (approve) stok.
const minimalPasswordLength = 12

// Register membuat user baru (hanya admin yang bisa memanggil, cek di handler)
func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if len(input.Password) < minimalPasswordLength {
		return nil, domain.ErrWeakPassword
	}

	// RoleID berasal dari body request. Tanpa cek ini, klien bisa mengirim
	// role_id sembarang (mis. 99) dan user lahir tanpa role yang sah.
	roles, err := u.userRepo.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load roles: %w", err)
	}
	roleExists := false
	for _, r := range roles {
		if r.ID == input.RoleID {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return nil, domain.ErrRoleNotFound
	}

	// Cek apakah username sudah ada
	existingUser, err := u.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUsernameTaken
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Buat user baru
	user := &domain.User{
		Username:     input.Username,
		PasswordHash: hashedPassword,
		FullName:     input.FullName,
		RoleID:       input.RoleID,
		IsActive:     true,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		// Balapan dua admin yang mendaftarkan username sama: unique index
		// menangkapnya di sini, bukan di pengecekan sebelumnya.
		if errors.Is(err, domain.ErrUsernameTaken) {
			return nil, domain.ErrUsernameTaken
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Load role
	user, err = u.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	return user, nil
}

// Login melakukan autentikasi dan mengembalikan token
func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*LoginResponse, error) {
	// Cari user berdasarkan username
	user, err := u.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("username atau password salah")
	}

	// Cek apakah user aktif
	if !user.IsActive {
		return nil, fmt.Errorf("user tidak aktif")
	}

	// Verifikasi password
	if err := utils.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		return nil, fmt.Errorf("username atau password salah")
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.RoleID,
		user.Role.Name,
		u.cfg.JWT.AccessSecret,
		u.cfg.JWT.AccessTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		u.cfg.JWT.RefreshSecret,
		u.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token untuk disimpan di database
	// Gunakan SHA256 karena refresh token terlalu panjang untuk bcrypt (>72 bytes)
	refreshTokenHash := utils.HashToken(refreshToken)

	// Simpan session
	session := &domain.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		DeviceID:         input.DeviceInfo.DeviceID,
		IPAddress:        input.DeviceInfo.IPAddress,
		UserAgent:        input.DeviceInfo.UserAgent,
		ExpiresAt:        time.Now().Add(u.cfg.JWT.RefreshTTL),
	}
	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int64(u.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

// RefreshToken menukar refresh token dengan access token baru
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string, deviceInfo DeviceInfo) (*LoginResponse, error) {
	// 1. Token harus valid secara kriptografis.
	userID, err := utils.ParseRefreshToken(refreshToken, u.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, domain.ErrRefreshTokenInvalid
	}

	// 2. Token harus cocok dengan sesi di DB. Sesi yang sudah dicabut/kedaluwarsa
	//    ditolak; yang sudah dipakai (revoked karena rotasi) dianggap kompromi.
	session, err := u.sessionRepo.FindByRefreshTokenHash(ctx, utils.HashToken(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("failed to look up session: %w", err)
	}
	if session == nil || session.UserID != userID {
		return nil, domain.ErrRefreshTokenInvalid
	}
	if session.RevokedAt != nil {
		if revokeErr := u.sessionRepo.RevokeByUserID(ctx, session.UserID); revokeErr != nil {
			return nil, fmt.Errorf("failed to revoke sessions after token reuse: %w", revokeErr)
		}
		u.audit(ctx, &session.UserID, "refresh_reuse_detected", deviceInfo)
		return nil, domain.ErrRefreshTokenReused
	}
	if !session.ExpiresAt.After(time.Now()) {
		return nil, domain.ErrRefreshTokenInvalid
	}

	// 3. User harus masih ada dan aktif.
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.RoleID,
		user.Role.Name,
		u.cfg.JWT.AccessSecret,
		u.cfg.JWT.AccessTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		u.cfg.JWT.RefreshSecret,
		u.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 4. Rotasi sekali pakai: sesi lama dicabut, pengganti dibuat, satu transaksi.
	newSession := &domain.Session{
		UserID:           user.ID,
		RefreshTokenHash: utils.HashToken(newRefreshToken),
		DeviceID:         deviceInfo.DeviceID,
		IPAddress:        deviceInfo.IPAddress,
		UserAgent:        deviceInfo.UserAgent,
		ExpiresAt:        time.Now().Add(u.cfg.JWT.RefreshTTL),
	}
	if err := u.sessionRepo.RotateRefreshToken(ctx, session.ID, newSession); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresIn:    int64(u.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

// Logout mencabut SEMUA sesi milik user. Handler mengirim user_id, bukan
// session_id — dulu nilai itu masuk ke RevokeByID dan tidak mencabut apa pun.
func (u *authUsecase) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := u.sessionRepo.RevokeByUserID(ctx, userID); err != nil {
		return err
	}
	u.audit(ctx, &userID, "logout", DeviceInfo{})
	return nil
}

// audit mencatat peristiwa autentikasi ke activity_logs. Kegagalan audit
// tidak menggagalkan request: ia hanya dicatat sebagai peringatan.
func (u *authUsecase) audit(ctx context.Context, userID *uuid.UUID, action string, device DeviceInfo) {
	if u.activityLogRepo == nil {
		return
	}
	if err := u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     userID,
		Action:     action,
		EntityType: "auth",
		IPAddress:  device.IPAddress,
	}); err != nil {
		logger.Warn("Gagal mencatat audit %s: %v", action, err)
	}
}

// GetProfile mengambil profile user
func (u *authUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}
	return user, nil
}

// ListUsers mengambil semua user (bukan staff saja: filter pihak depan).
func (u *authUsecase) ListUsers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.List(ctx, 1000, 0)
}

func (u *authUsecase) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return u.userRepo.ListRoles(ctx)
}

func (u *authUsecase) ToggleUserStatus(ctx context.Context, userID uuid.UUID, isActive bool, adminID uuid.UUID) error {
	if userID == adminID {
		return fmt.Errorf("tidak bisa menonaktifkan akun sendiri")
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	user.IsActive = isActive
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &adminID,
		Action:     "update_user_status",
		EntityType: "user",
		EntityID:   &userID,
		NewValue:   datatypes.JSON(fmt.Sprintf(`{"is_active":%v}`, isActive)),
	})
	return nil
}

func (u *authUsecase) ResetUserPassword(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) (string, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("user tidak ditemukan")
	}
	// Generate password acak 12 karakter alfanumerik
	newPassword := generateRandomPassword(12)
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = hashed
	if err := u.userRepo.Update(ctx, user); err != nil {
		return "", err
	}
	// Revoke semua refresh token user agar sesi lama invalid
	_ = u.sessionRepo.RevokeByUserID(ctx, userID)
	_ = u.activityLogRepo.Create(ctx, &domain.ActivityLog{
		UserID:     &adminID,
		Action:     "reset_user_password",
		EntityType: "user",
		EntityID:   &userID,
	})
	return newPassword, nil
}

// generateRandomPassword menghasilkan password acak sepanjang n karakter.
func generateRandomPassword(n int) string {
	const charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	for i, b := range buf {
		buf[i] = charset[int(b)%len(charset)]
	}
	return string(buf)
}

// ===== END FILE =====
