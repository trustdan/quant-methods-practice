package tutor

import (
	"context"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// TutorAction specifies the pedagogical intent of a tutor interaction.
type TutorAction string

const (
	ActionHint     TutorAction = "hint"
	ActionExplain  TutorAction = "explain"
	ActionFollowUp TutorAction = "follow_up"
)

// FollowUpKind categorizes specific drill-down follow-up inquiries.
type FollowUpKind string

const (
	FollowUpExplainDifferently  TutorFollowUpKind = "explain_differently"
	FollowUpWorkedExample       TutorFollowUpKind = "worked_example"
	FollowUpWhyConditionMatters TutorFollowUpKind = "why_condition_matters"
	FollowUpCompareConcepts     TutorFollowUpKind = "compare_concepts"
	FollowUpCustom              TutorFollowUpKind = "custom"
)

type TutorFollowUpKind = FollowUpKind

// TutorRequest contains contextual problem and stage snapshots for the tutor.
type TutorRequest struct {
	RequestID       string                   `json:"request_id"`
	SessionID       string                   `json:"session_id"`
	InstanceID      string                   `json:"instance_id"`
	StageID         string                   `json:"stage_id"`
	Action          TutorAction              `json:"action"`
	FollowUpKind    FollowUpKind             `json:"follow_up_kind,omitempty"`
	CustomPrompt    string                   `json:"custom_prompt,omitempty"`
	Provider        string                   `json:"provider,omitempty"`
	Instance        *domain.QuestionInstance `json:"instance,omitempty"`
	Stage           *domain.StageInstance    `json:"stage,omitempty"`
	SubmittedAnswer *domain.SubmittedAnswer  `json:"submitted_answer,omitempty"`
	MisconceptionID string                   `json:"misconception_id,omitempty"`
}

// TutorEventType represents the lifecycle state of a streaming tutor response.
type TutorEventType string

const (
	EventStarted   TutorEventType = "started"
	EventTextDelta TutorEventType = "text_delta"
	EventComplete  TutorEventType = "complete"
	EventFallback  TutorEventType = "fallback"
	EventCancelled TutorEventType = "cancelled"
	EventError     TutorEventType = "error"
)

// TutorEvent represents a single streamed event dispatched to the learner.
type TutorEvent struct {
	Type          TutorEventType `json:"type"`
	RequestID     string         `json:"request_id"`
	SessionID     string         `json:"session_id,omitempty"`
	InstanceID    string         `json:"instance_id,omitempty"`
	StageID       string         `json:"stage_id,omitempty"`
	Delta         string         `json:"delta,omitempty"`
	Text          string         `json:"text,omitempty"`
	FallbackLabel string         `json:"fallback_label,omitempty"`
	Error         string         `json:"error,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
}

// ProviderCapabilities describes what features a tutor provider supports.
type ProviderCapabilities struct {
	SupportsStreaming    bool `json:"supports_streaming"`
	SupportsFollowUps    bool `json:"supports_follow_ups"`
	SupportsCancellation bool `json:"supports_cancellation"`
	IsOffline            bool `json:"is_offline"`
}

// TutorService defines the read-only AI tutor contract.
// Implementations MUST NOT have any execution, bank approval, grade, or mastery APIs.
type TutorService interface {
	Stream(ctx context.Context, req TutorRequest) (<-chan TutorEvent, error)
	ProviderID() string
	Capabilities() ProviderCapabilities
}
