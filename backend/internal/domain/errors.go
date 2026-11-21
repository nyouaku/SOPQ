package domain

import "errors"

// Kesalahan domain yang dipetakan ke status HTTP di lapisan delivery.
var (
	ErrSessionNotFound        = errors.New("stock opname session not found")
	ErrSessionCodeTaken       = errors.New("session code already used")
	ErrSessionNotApprovable   = errors.New("session must be completed before approval")
	ErrSessionAlreadyApproved = errors.New("session is already approved")

	ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")
	ErrRefreshTokenReused  = errors.New("refresh token has already been used")
	ErrUserInactive        = errors.New("user is not active")

	ErrScanDuplicate  = errors.New("scan with this idempotency key already exists")
	ErrUsernameTaken  = errors.New("username sudah digunakan")
	ErrRoleNotFound   = errors.New("role tidak ditemukan")
	ErrWeakPassword   = errors.New("password minimal 12 karakter")
	ErrImportTooLarge = errors.New("jumlah baris melebihi batas")
	ErrExportTooLarge = errors.New("jumlah baris ekspor melebihi batas")

	ErrMovementNotFound         = errors.New("dokumen pergerakan tidak ditemukan")
	ErrMovementAlreadyCancelled = errors.New("dokumen pergerakan sudah dibatalkan")
	ErrMovementNotCancellable   = errors.New("hanya dokumen berstatus posted yang bisa dibatalkan")

	ErrChangeRequestNotFound   = errors.New("pengajuan perubahan tidak ditemukan")
	ErrChangeRequestNotPending = errors.New("pengajuan sudah diproses")
)
