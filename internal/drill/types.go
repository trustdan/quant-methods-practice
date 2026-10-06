package drill

import (
	"context"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

// SessionStore defines the storage capabilities required by the drill session and engine.
type SessionStore interface {
	SaveSession(ctx context.Context, session *DrillSession) error
	GetSession(ctx context.Context, id string) (*DrillSession, error)
	GetLatestActiveSession(ctx context.Context) (*DrillSession, error)
	SaveCommandResult(ctx context.Context, sessionID, cmdID string, cmdType CommandType, res *CommandResult) error
	GetCommandResult(ctx context.Context, cmdID string) (*CommandResult, error)
	SaveDraftAnswer(ctx context.Context, sessionID, stageID string, ans domain.SubmittedAnswer) error
	ClearSessionDraft(ctx context.Context, sessionID, stageID string) error
	SaveSettings(ctx context.Context, key, valueJSON string) error
	GetSettings(ctx context.Context, key string) (string, error)
}

// StageProgressStatus represents learner status on an individual stage.
type StageProgressStatus string

const (
	StageStatusUnvisited StageProgressStatus = "unvisited"
	StageStatusActive    StageProgressStatus = "active"
	StageStatusRetry     StageProgressStatus = "retry"
	StageStatusCompleted StageProgressStatus = "completed"
)

// StageState holds authoritative runtime state for one stage of a drill.
type StageState struct {
	Instance           domain.StageInstance    `json:"instance"`
	Status             StageProgressStatus     `json:"status"`
	Attempts           []domain.StageAttempt   `json:"attempts"`
	Assistance         []domain.AssistanceType `json:"assistance"`
	FirstTryCorrect    bool                    `json:"first_try_correct"`
	SolvedOnRetry      bool                    `json:"solved_on_retry"`
	Revealed           bool                    `json:"revealed"`
	ActiveHint         string                  `json:"active_hint,omitempty"`
	LastFeedback       string                  `json:"last_feedback,omitempty"`
	MisconceptionID    *string                 `json:"misconception_id,omitempty"`
	InvalidInputNotice string                  `json:"invalid_input_notice,omitempty"`
	DraftAnswer        *domain.SubmittedAnswer `json:"draft_answer,omitempty"`
}

// DrillSession manages the execution, history, and state machine of a practice drill.
type DrillSession struct {
	ID                string                  `json:"id"`
	TemplateID        string                  `json:"template_id"`
	TemplateVersion   int                     `json:"template_version"`
	Seed              int64                   `json:"seed"`
	Revision          int64                   `json:"revision"`
	QuestionInstance  domain.QuestionInstance `json:"question_instance"`
	Stages            []StageState            `json:"stages"`
	CurrentStageIndex int                     `json:"current_stage_index"`
	Completed         bool                    `json:"completed"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`

	// commandCache caches results by CommandID for idempotency
	commandCache map[string]*CommandResult
	store        SessionStore

	// Multi-question practice session state
	CurrentQuestionIndex int                    `json:"current_question_index"`
	Questions            []QuestionState        `json:"questions,omitempty"`
	Settings             domain.SessionSettings `json:"settings"`
	ContrastCount        int                    `json:"contrast_count,omitempty"`

	contrastFinder func(originID, misID string) (*domain.QuestionTemplate, bool)
}

// QuestionState tracks progress and stages for one question within a multi-question drill session.
type QuestionState struct {
	Index             int                     `json:"index"`
	TemplateID        string                  `json:"template_id"`
	TemplateVersion   int                     `json:"template_version"`
	QuestionInstance  domain.QuestionInstance `json:"question_instance"`
	Stages            []StageState            `json:"stages"`
	CurrentStageIndex int                     `json:"current_stage_index"`
	Completed         bool                    `json:"completed"`
	Recap             *DrillRecap             `json:"recap,omitempty"`
	ScaffoldLevel     string                  `json:"scaffold_level,omitempty"`
	IsContrast        bool                    `json:"is_contrast,omitempty"`
	ContrastPartnerID string                  `json:"contrast_partner_id,omitempty"`
}

// SetStore binds a persistence store to the session.
func (s *DrillSession) SetStore(st SessionStore) {
	s.store = st
}

// SetCachedResult records a command result in the in-memory cache.
func (s *DrillSession) SetCachedResult(cmdID string, res *CommandResult) {
	if s.commandCache == nil {
		s.commandCache = make(map[string]*CommandResult)
	}
	s.commandCache[cmdID] = res
}

// GetCachedResult retrieves a cached command result if present.
func (s *DrillSession) GetCachedResult(cmdID string) (*CommandResult, bool) {
	if s.commandCache == nil {
		return nil, false
	}
	res, ok := s.commandCache[cmdID]
	return res, ok
}

// SetContrastFinder registers a callback to locate contrast partners for misconceptions.
func (s *DrillSession) SetContrastFinder(fn func(originID, misID string) (*domain.QuestionTemplate, bool)) {
	s.contrastFinder = fn
}

// CurrentQuestionIsContrast returns true if the active question is a guided contrast problem.
func (s *DrillSession) CurrentQuestionIsContrast() bool {
	if len(s.Questions) > 0 && s.CurrentQuestionIndex >= 0 && s.CurrentQuestionIndex < len(s.Questions) {
		return s.Questions[s.CurrentQuestionIndex].IsContrast
	}
	return false
}

// CommandType represents actions dispatched by the client.
type CommandType string

const (
	CmdSubmitAnswer     CommandType = "submit_answer"
	CmdRequestHint      CommandType = "request_hint"
	CmdNavigateStage    CommandType = "navigate_stage"
	CmdNavigateQuestion CommandType = "navigate_question"
	CmdResetDrill       CommandType = "reset_drill"
	CmdSaveDraft        CommandType = "save_draft"
	CmdClearDraft       CommandType = "clear_draft"
)

// SessionCommand represents a mutation request sent to the session.
type SessionCommand struct {
	CommandID           string                  `json:"command_id"`
	ExpectedRevision    int64                   `json:"expected_revision"`
	Type                CommandType             `json:"type"`
	StageID             string                  `json:"stage_id,omitempty"`
	Answer              *domain.SubmittedAnswer `json:"answer,omitempty"`
	TargetStageIndex    *int                    `json:"target_stage_index,omitempty"`
	TargetQuestionIndex *int                    `json:"target_question_index,omitempty"`
}

// CommandResult represents the outcome of processing a SessionCommand.
type CommandResult struct {
	Success      bool              `json:"success"`
	CommandID    string            `json:"command_id"`
	SessionState PublicSessionView `json:"session_state"`
	ErrorMessage string            `json:"error_message,omitempty"`
	InvalidInput bool              `json:"invalid_input,omitempty"`
}

// PublicOptionView safely exposes choice options without leaking misconception keys or hints.
type PublicOptionView struct {
	ID           string `json:"id"`
	TextMarkdown string `json:"text_markdown"`
}

// PublicNumericPolicyView provides learner-facing input guidelines.
type PublicNumericPolicyView struct {
	Version         int      `json:"version"`
	AllowedForms    []string `json:"allowed_forms"`
	DisplayDecimals *int     `json:"display_decimals,omitempty"`
}

// PublicStageView projects a stage for safe transmission to the client.
// Unresolved stages withhold expected answers, unselected misconception IDs, and final explanations.
type PublicStageView struct {
	ID                  string                   `json:"id"`
	Number              int                      `json:"number"`
	Label               string                   `json:"label"`
	Kind                domain.StageKind         `json:"kind"`
	PromptMarkdown      string                   `json:"prompt_markdown"`
	Status              StageProgressStatus      `json:"status"`
	Options             []PublicOptionView       `json:"options"`
	NumericPolicy       *PublicNumericPolicyView `json:"numeric_policy,omitempty"`
	AttemptCount        int                      `json:"attempt_count"`
	MaxAttempts         int                      `json:"max_attempts"`
	IsCorrect           *bool                    `json:"is_correct,omitempty"`
	SolvedOnRetry       bool                     `json:"solved_on_retry"`
	Revealed            bool                     `json:"revealed"`
	ActiveHint          string                   `json:"active_hint,omitempty"`
	LastFeedback        string                   `json:"last_feedback,omitempty"`
	ExplanationMarkdown string                   `json:"explanation_markdown,omitempty"`
	ExpectedAnswer      *domain.ExpectedAnswer   `json:"expected_answer,omitempty"`
	Attempts            []domain.StageAttempt    `json:"attempts"`
	InvalidInputNotice  string                   `json:"invalid_input_notice,omitempty"`
	DraftAnswer         *domain.SubmittedAnswer  `json:"draft_answer,omitempty"`
}

// PublicQuestionInfo provides summary status of a question within a multi-question session.
type PublicQuestionInfo struct {
	Index         int    `json:"index"`
	Title         string `json:"title"`
	Status        string `json:"status"` // "pending", "in_progress", "completed", "skipped"
	ScaffoldLevel string `json:"scaffold_level,omitempty"`
	IsContrast    bool   `json:"is_contrast,omitempty"`
}

// PublicSessionView is the sanitized projection sent to the web frontend.
type PublicSessionView struct {
	ID                   string               `json:"id"`
	TemplateID           string               `json:"template_id"`
	Revision             int64                `json:"revision"`
	Title                string               `json:"title"`
	ScenarioMarkdown     string               `json:"scenario_markdown"`
	Parameters           map[string]any       `json:"parameters"`
	Assumptions          []string             `json:"assumptions"`
	CurrentStageIndex    int                  `json:"current_stage_index"`
	Completed            bool                 `json:"completed"`
	Stages               []PublicStageView    `json:"stages"`
	Recap                *DrillRecap          `json:"recap,omitempty"`
	UpdatedAt            time.Time            `json:"updated_at"`
	CurrentQuestionIndex int                  `json:"current_question_index"`
	TotalQuestions       int                  `json:"total_questions"`
	Questions            []PublicQuestionInfo `json:"questions"`
	AllCompleted         bool                 `json:"all_completed"`
	ScaffoldLevel        string               `json:"scaffold_level,omitempty"`
	IsContrast           bool                 `json:"is_contrast,omitempty"`
}

// StageSummary records stage performance for the post-drill recap.
type StageSummary struct {
	StageNumber            int     `json:"stage_number"`
	StageID                string  `json:"stage_id"`
	Label                  string  `json:"label"`
	PromptMarkdown         string  `json:"prompt_markdown"`
	LearnerAnswer          string  `json:"learner_answer"`
	CanonicalAnswer        string  `json:"canonical_answer"`
	Outcome                string  `json:"outcome"` // "first_try", "retry", "revealed"
	MisconceptionTriggered *string `json:"misconception_triggered,omitempty"`
	ExplanationMarkdown    string  `json:"explanation_markdown"`
}

// DrillRecap provides the full worked solution and performance review.
type DrillRecap struct {
	Title               string                         `json:"title"`
	ScenarioMarkdown    string                         `json:"scenario_markdown"`
	TotalStages         int                            `json:"total_stages"`
	FirstTryCount       int                            `json:"first_try_count"`
	RetryCount          int                            `json:"retry_count"`
	RevealedCount       int                            `json:"revealed_count"`
	CanonicalDerivation *mathengine.BinomialDerivation `json:"canonical_derivation,omitempty"`
	StageSummaries      []StageSummary                 `json:"stage_summaries"`
}
