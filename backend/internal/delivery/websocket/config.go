package websocket

import (
	"context"

	"github.com/google/uuid"
)

// WSConfig membawa dependensi yang dibutuhkan handler WebSocket.
type WSConfig struct {
	// JWTSecret dipakai memvalidasi access token dari handshake.
	JWTSecret string
	// AllowedOrigins berasal dari CORS_ALLOWED_ORIGINS, bukan daftar hardcode.
	AllowedOrigins []string
	// CanAccessSession memeriksa apakah user boleh melihat sesi ini.
	// Nil = semua user terautentikasi boleh (dipakai saat wiring awal).
	CanAccessSession func(ctx context.Context, userID uuid.UUID, roleName string, sessionID uuid.UUID) (bool, error)
}

// BroadcastNewScan menyusun payload update dan mengirimnya ke room sesi.
func BroadcastNewScan(sessionID uuid.UUID, detailPayload []byte) {
	payload := make([]byte, 0, len(detailPayload)+64)
	payload = append(payload, `{"type":"new_scan","detail":`...)
	payload = append(payload, detailPayload...)
	payload = append(payload, '}')
	BroadcastProgress(sessionID, payload)
}

// BroadcastSessionUpdate menyusun payload perubahan status sesi.
func BroadcastSessionUpdate(sessionID uuid.UUID, sessionPayload []byte) {
	payload := make([]byte, 0, len(sessionPayload)+64)
	payload = append(payload, `{"type":"session_update","session":`...)
	payload = append(payload, sessionPayload...)
	payload = append(payload, '}')
	BroadcastProgress(sessionID, payload)
}
