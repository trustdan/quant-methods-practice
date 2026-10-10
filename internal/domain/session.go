package domain

import (
	"time"
)

// AssistanceType enumerates kinds of pedagogical assistance recorded for an attempt.
type AssistanceType string

const (
	AssistanceNone           AssistanceType = "none"
	AssistanceHint           AssistanceType = "hint"
	AssistanceRetry          AssistanceType = "retry"
	AssistanceReference      AssistanceType = "reference"
	AssistanceGuidedContrast AssistanceType = "guided_contrast"
	AssistanceSolutionReveal AssistanceType = "solution_reveal"
	AssistanceTutor          AssistanceType = "tutor"
	AssistanceFullSolution   AssistanceType = "full_solution_form"
)

// SubmittedAnswer captures learner input across choice and numeric forms.
type SubmittedAnswer struct {
	Kind            StageKind    `json:"kind"`
	OptionID        string       `json:"option_id,omitempty"`
	NumericRaw      string       `json:"numeric_raw,omitempty"`
	NormalizedValue *float64     `json:"normalized_value,omitempty"`
	Form            *NumericForm `json:"form,omitempty"`
}

// StageAttempt records one evaluated submission to a stage within a session.
type StageAttempt struct {
	ID               string           `json:"id"`
	SessionID        string           `json:"session_id"`
	InstanceID       string           `json:"instance_id"`
	StageID          string           `json:"stage_id"`
	AttemptNumber    int              `json:"attempt_number"`
	SubmittedAnswer  SubmittedAnswer  `json:"submitted_answer"`
	Assistance       []AssistanceType `json:"assistance"`
	IsCorrect        bool             `json:"is_correct"`
	FeedbackMarkdown string           `json:"feedback_markdown"`
	CreatedAt        time.Time        `json:"created_at"`
}

// SessionSettings captures configuration for a generated practice session.
type SessionSettings struct {
	QuestionCount int      `json:"question_count"`
	ModuleIDs     []string `json:"module_ids,omitempty"`
	ConceptIDs    []string `json:"concept_ids,omitempty"`
	Intensity     string   `json:"intensity"`
}

// PracticeSession coordinates active instances and learner progress across stages.
type PracticeSession struct {
	ID                   string             `json:"id"`
	Seed                 int64              `json:"seed"`
	Revision             int64              `json:"revision"`
	Settings             SessionSettings    `json:"settings"`
	QuestionInstances    []QuestionInstance `json:"question_instances"`
	CurrentQuestionIndex int                `json:"current_question_index"`
	CurrentStageIndex    int                `json:"current_stage_index"`
	Completed            bool               `json:"completed"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}
