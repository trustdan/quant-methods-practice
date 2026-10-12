package exam

import (
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

type ExamStatus string

const (
	StatusActive    ExamStatus = "active"
	StatusCompleted ExamStatus = "completed"
	StatusExpired   ExamStatus = "expired"
	StatusAbandoned ExamStatus = "abandoned"
)

// ExamConfig specifies the parameters to initialize an assessment.
type ExamConfig struct {
	Title           string   `json:"title,omitempty"`
	QuestionCount   int      `json:"question_count,omitempty"`
	ModuleIDs       []string `json:"module_ids,omitempty"`
	DurationMinutes int      `json:"duration_minutes,omitempty"` // 0 for untimed
	Seed            int64    `json:"seed,omitempty"`
	AllowedAid      string   `json:"allowed_aid,omitempty"` // e.g. "none", "formula_sheet"
}

// ExamRecord is the authoritative durable state of an exam.
type ExamRecord struct {
	ID                   string                            `json:"id"`
	Title                string                            `json:"title"`
	Revision             int                               `json:"revision"`
	Status               ExamStatus                        `json:"status"`
	DurationMinutes      int                               `json:"duration_minutes"`
	AllowedAid           string                            `json:"allowed_aid"`
	Seed                 int64                             `json:"seed"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
	StartedAt            time.Time                         `json:"started_at"`
	DeadlineAt           *time.Time                        `json:"deadline_at,omitempty"`
	CompletedAt          *time.Time                        `json:"completed_at,omitempty"`
	CurrentQuestionIndex int                               `json:"current_question_index"`
	Questions            []domain.QuestionInstance         `json:"questions"`
	Responses            map[string]domain.SubmittedAnswer `json:"responses"` // key: questionID:stageID
	BankTotalQuestions   int                               `json:"bank_total_questions,omitempty"`
	Report               *ExamReport                       `json:"report,omitempty"`
}

// ExamCommand mutates the exam state under revision control.
type ExamCommand struct {
	CommandID        string                  `json:"command_id"`
	ExpectedRevision int                     `json:"expected_revision"`
	Type             string                  `json:"type"` // "save_answer", "navigate", "submit_exam", "abandon_exam"
	QuestionIndex    *int                    `json:"question_index,omitempty"`
	StageID          string                  `json:"stage_id,omitempty"`
	Answer           *domain.SubmittedAnswer `json:"answer,omitempty"`
}

// ExamStageReport details grading and canonical explanation for a single stage.
type ExamStageReport struct {
	ID                  string                  `json:"id"`
	Kind                domain.StageKind        `json:"kind"`
	PromptMarkdown      string                  `json:"prompt_markdown"`
	Options             []domain.Option         `json:"options,omitempty"`
	SubmittedAnswer     *domain.SubmittedAnswer `json:"submitted_answer,omitempty"`
	ExpectedAnswer      domain.ExpectedAnswer   `json:"expected_answer"`
	IsCorrect           bool                    `json:"is_correct"`
	FeedbackMarkdown    string                  `json:"feedback_markdown"`
	ExplanationMarkdown string                  `json:"explanation_markdown"`
	NumericPolicy       *domain.NumericPolicy   `json:"numeric_policy,omitempty"`
}

// ExamQuestionReport summarizes evaluation of an entire question snapshot.
type ExamQuestionReport struct {
	Index            int               `json:"index"`
	QuestionID       string            `json:"question_id"`
	TemplateID       string            `json:"template_id"`
	Title            string            `json:"title"`
	ScenarioMarkdown string            `json:"scenario_markdown"`
	Assumptions      []string          `json:"assumptions,omitempty"`
	StageReports     []ExamStageReport `json:"stage_reports"`
	CorrectStages    int               `json:"correct_stages"`
	TotalStages      int               `json:"total_stages"`
	IsFullyCorrect   bool              `json:"is_fully_correct"`
}

// ExamReport contains full post-completion grading and evaluation.
type ExamReport struct {
	TotalQuestions   int                  `json:"total_questions"`
	TotalStages      int                  `json:"total_stages"`
	CorrectStages    int                  `json:"correct_stages"`
	ScorePercent     float64              `json:"score_percent"`
	TimeSpentSeconds int                  `json:"time_spent_seconds"`
	Expired          bool                 `json:"expired"`
	QuestionReports  []ExamQuestionReport `json:"question_reports"`
}

// PublicOptionView represents a sanitized choice option without hints or misconception tags.
type PublicOptionView struct {
	ID           string `json:"id"`
	TextMarkdown string `json:"text_markdown"`
}

// PublicStageView represents what the client is permitted to view for a stage.
type PublicStageView struct {
	ID              string                  `json:"id"`
	Number          int                     `json:"number"`
	Kind            domain.StageKind        `json:"kind"`
	PromptMarkdown  string                  `json:"prompt_markdown"`
	Options         []PublicOptionView      `json:"options,omitempty"`
	NumericPolicy   *domain.NumericPolicy   `json:"numeric_policy,omitempty"`
	SubmittedAnswer *domain.SubmittedAnswer `json:"submitted_answer,omitempty"`
	IsAnswered      bool                    `json:"is_answered"`
}

// PublicQuestionView represents what the client is permitted to view for a question.
type PublicQuestionView struct {
	Index            int               `json:"index"`
	QuestionID       string            `json:"question_id"`
	Title            string            `json:"title"`
	ScenarioMarkdown string            `json:"scenario_markdown"`
	Assumptions      []string          `json:"assumptions,omitempty"`
	Stages           []PublicStageView `json:"stages"`
	IsAnswered       bool              `json:"is_answered"`
}

// PublicExamView represents the complete sanitized client projection.
// While active, keys, feedback, hints, and explanations are withheld.
type PublicExamView struct {
	ID                   string               `json:"id"`
	Title                string               `json:"title"`
	Revision             int                  `json:"revision"`
	Status               ExamStatus           `json:"status"`
	DurationMinutes      int                  `json:"duration_minutes"`
	AllowedAid           string               `json:"allowed_aid"`
	StartedAt            time.Time            `json:"started_at"`
	DeadlineAt           *time.Time           `json:"deadline_at,omitempty"`
	CompletedAt          *time.Time           `json:"completed_at,omitempty"`
	TimeRemainingSeconds *int                 `json:"time_remaining_seconds,omitempty"`
	CurrentQuestionIndex int                  `json:"current_question_index"`
	TotalQuestions       int                  `json:"total_questions"`
	BankTotalQuestions   int                  `json:"bank_total_questions,omitempty"`
	AnsweredQuestions    int                  `json:"answered_questions"`
	Questions            []PublicQuestionView `json:"questions"`
	Report               *ExamReport          `json:"report,omitempty"` // populated only when not active
}

// ExamSummary is a high-level summary of an assessment for listings and history.
type ExamSummary struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Status          ExamStatus `json:"status"`
	DurationMinutes int        `json:"duration_minutes"`
	QuestionCount   int        `json:"question_count"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ScorePercent    *float64   `json:"score_percent,omitempty"`
	CorrectStages   *int       `json:"correct_stages,omitempty"`
	TotalStages     *int       `json:"total_stages,omitempty"`
}
