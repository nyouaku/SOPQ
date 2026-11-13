package websocket

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
)

// Publisher mengimplementasikan usecase.ScanBroadcaster tanpa membuat usecase
// mengimpor package ini.
type Publisher struct{}

func NewPublisher() *Publisher { return &Publisher{} }

func (p *Publisher) BroadcastNewScan(sessionID uuid.UUID, detail *domain.StockOpnameDetail) {
	body, err := json.Marshal(detail)
	if err != nil {
		return
	}
	BroadcastNewScan(sessionID, body)
}

func (p *Publisher) BroadcastSessionUpdate(sessionID uuid.UUID, session *domain.StockOpname) {
	body, err := json.Marshal(session)
	if err != nil {
		return
	}
	BroadcastSessionUpdate(sessionID, body)
}
