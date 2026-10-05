package storage

import (
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// SessionDraft represents an unsaved learner draft on a specific stage.
type SessionDraft struct {
	SessionID      string                 `json:"session_id"`
	StageID        string                 `json:"stage_id"`
	DraftAnswer    domain.SubmittedAnswer `json:"draft_answer"`
	ActivePosition int                    `json:"active_position"`
	Revision       int64                  `json:"revision"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// SettingsRecord represents a key-value configuration preference.
type SettingsRecord struct {
	Key       string    `json:"key"`
	ValueJSON string    `json:"value_json"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AssistanceEventRecord captures an append-only assistance log entry.
type AssistanceEventRecord struct {
	ID         string                `json:"id"`
	SessionID  string                `json:"session_id"`
	InstanceID string                `json:"instance_id"`
	StageID    string                `json:"stage_id"`
	EventType  domain.AssistanceType `json:"event_type"`
	Scope      string                `json:"scope"`
	CreatedAt  time.Time             `json:"created_at"`
}
