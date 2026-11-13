package websocket

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

// wsSubprotocol adalah nama sub-protokol yang dipakai klien untuk mengirim
// access token di header handshake. Token TIDAK boleh lewat query string:
// URL tercatat di access log nginx dan riwayat browser.
const wsSubprotocol = "access_token"

// StockOpnameWSHub menyimpan koneksi per sesi opname. Klien yang berlangganan
// sesi A tidak boleh menerima update sesi B.
type StockOpnameWSHub struct {
	rooms      map[uuid.UUID]map[*websocket.Conn]bool
	broadcast  chan broadcastMessage
	register   chan registration
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

type registration struct {
	sessionID uuid.UUID
	conn      *websocket.Conn
}

type broadcastMessage struct {
	sessionID uuid.UUID
	payload   []byte
}

var hub = &StockOpnameWSHub{
	rooms:      make(map[uuid.UUID]map[*websocket.Conn]bool),
	broadcast:  make(chan broadcastMessage),
	register:   make(chan registration),
	unregister: make(chan *websocket.Conn),
}

func init() {
	go hub.run()
}

func (h *StockOpnameWSHub) run() {
	for {
		select {
		case reg := <-h.register:
			h.mu.Lock()
			if h.rooms[reg.sessionID] == nil {
				h.rooms[reg.sessionID] = make(map[*websocket.Conn]bool)
			}
			h.rooms[reg.sessionID][reg.conn] = true
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			for sessionID, conns := range h.rooms {
				if _, ok := conns[conn]; ok {
					delete(conns, conn)
					if len(conns) == 0 {
						delete(h.rooms, sessionID)
					}
					break
				}
			}
			conn.Close()
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			var failed []*websocket.Conn
			for conn := range h.rooms[msg.sessionID] {
				if err := conn.WriteMessage(websocket.TextMessage, msg.payload); err != nil {
					failed = append(failed, conn)
				}
			}
			h.mu.RUnlock()

			if len(failed) > 0 {
				h.mu.Lock()
				for _, conn := range failed {
					if conns := h.rooms[msg.sessionID]; conns != nil {
						delete(conns, conn)
						if len(conns) == 0 {
							delete(h.rooms, msg.sessionID)
						}
					}
					conn.Close()
				}
				h.mu.Unlock()
			}
		}
	}
}

// BroadcastProgress mengirim update ke klien yang berlangganan sesi tsb saja.
// ShutdownWS menutup semua koneksi WS aktif. Hub tetap jalan supaya handler
// yang sedang berjalan bisa tetap mengirim unregister tanpa deadlock.
func ShutdownWS() {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for _, conns := range hub.rooms {
		for conn := range conns {
			conn.Close()
		}
	}
}

func BroadcastProgress(sessionID uuid.UUID, payload []byte) {
	select {
	case hub.broadcast <- broadcastMessage{sessionID: sessionID, payload: payload}:
	default:
		// Antrean penuh: lewati update ini daripada memblokir pemanggil.
		logger.Warn("WebSocket broadcast queue penuh, update dilewati")
	}
}

// newUpgrader membangun upgrader dengan origin whitelist dari konfigurasi.
// Sebelumnya daftar origin di-hardcode ke alamat dev.
func newUpgrader(allowedOrigins []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{wsSubprotocol},
		CheckOrigin: func(r *http.Request) bool {
			return originAllowed(r.Header.Get("Origin"), allowedOrigins)
		},
	}
}

// originAllowed: origin kosong (klien non-browser) ditolak karena token tetap
// wajib; yang dicek hanya apakah origin cocok dengan daftar.
func originAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	origin = strings.TrimSuffix(origin, "/")
	for _, a := range allowed {
		a = strings.TrimSuffix(strings.TrimSpace(a), "/")
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

// tokenFromHandshake mengambil token dari sub-protokol `access_token, <jwt>`.
func tokenFromHandshake(r *http.Request) string {
	header := r.Header.Get("Sec-WebSocket-Protocol")
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	for i, p := range parts {
		if strings.TrimSpace(p) == wsSubprotocol && i+1 < len(parts) {
			return strings.TrimSpace(parts[i+1])
		}
	}
	return ""
}

func HandleStockOpnameWS(cfg *WSConfig) gin.HandlerFunc {
	upgrader := newUpgrader(cfg.AllowedOrigins)

	return func(c *gin.Context) {
		sessionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
			return
		}

		token := tokenFromHandshake(c.Request)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
			return
		}

		claims, err := utils.ParseAccessToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		if cfg.CanAccessSession != nil {
			ok, accessErr := cfg.CanAccessSession(c.Request.Context(), claims.UserID, claims.RoleName, sessionID)
			if accessErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access"})
				return
			}
			if !ok {
				c.JSON(http.StatusForbidden, gin.H{"error": "no access to this session"})
				return
			}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("WebSocket upgrade error: %v", err)
			return
		}

		select {
		case hub.register <- registration{sessionID: sessionID, conn: conn}:
		case <-time.After(5 * time.Second):
			conn.Close()
			return
		}

		// Satu jalur berhenti saja: dulu read-loop dan ping-loop sama-sama
		// mengirim ke unregister, sehingga koneksi kedua kali menutup.
		done := make(chan struct{})
		var closeOnce sync.Once
		stop := func() {
			closeOnce.Do(func() {
				close(done)
				hub.unregister <- conn
			})
		}
		defer stop()

		go func() {
			defer stop()
			conn.SetReadLimit(512)
			_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
			conn.SetPongHandler(func(string) error {
				return conn.SetReadDeadline(time.Now().Add(90 * time.Second))
			})
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}
}
