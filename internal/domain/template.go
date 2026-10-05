package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TemplateKind represents the structural variety of a template.
type TemplateKind string

const (
	TemplateKindFixedReference TemplateKind = "fixed_reference"
)

// TemplateStatus represents the approval and lifecycle status of a template.
type TemplateStatus string

const (
	StatusDraft   TemplateStatus = "draft"
	StatusActive  TemplateStatus = "active"
	StatusRetired TemplateStatus = "retired"
)

// StageKind represents the input type required by a stage.
type StageKind string

const (
	StageKindChoice  StageKind = "choice"
	StageKindNumeric StageKind = "numeric"
)

// AnswerKind represents the kind of expected answer.
type AnswerKind string

const (
	AnswerKindChoice  AnswerKind = "choice"
	AnswerKindNumeric AnswerKind = "numeric"
)

// NumericForm represents an allowed numeric input representation.
type NumericForm string

const (
	NumericFormDecimal  NumericForm = "decimal"
	NumericFormPercent  NumericForm = "percent"
	NumericFormFraction NumericForm = "fraction"
)

// ApprovalRecord captures human review provenance for active content.
type ApprovalRecord struct {
	Reviewer        string    `json:"reviewer"`
	ReviewedAt      time.Time `json:"reviewed_at"`
	Notes           string    `json:"notes"`
	DelegationBasis *string   `json:"delegation_basis"`
}

// Option represents a choice in a multiple-choice stage.
type Option struct {
	ID              string  `json:"id"`
	TextMarkdown    string  `json:"text_markdown"`
	MisconceptionID *string `json:"misconception_id"`
	HintMarkdown    *string `json:"hint_markdown"`
}

// ExpectedAnswer represents the canonical answer specification for a stage.
type ExpectedAnswer struct {
	Kind     AnswerKind `json:"kind"`
	OptionID string     `json:"option_id,omitempty"`
	Value    *float64   `json:"value,omitempty"`
	Units    string     `json:"units,omitempty"`
}

// UnmarshalJSON strictly validates expected answer variants and rejects extra properties.
func (ea *ExpectedAnswer) UnmarshalJSON(data []byte) error {
	var probe struct {
		Kind AnswerKind `json:"kind"`
	}
	probeDec := json.NewDecoder(bytes.NewReader(data))
	if err := probeDec.Decode(&probe); err != nil {
		return fmt.Errorf("reading expected_answer kind: %w", err)
	}

	switch probe.Kind {
	case AnswerKindChoice:
		type rawChoice struct {
			Kind     AnswerKind `json:"kind"`
			OptionID string     `json:"option_id"`
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var rc rawChoice
		if err := dec.Decode(&rc); err != nil {
			return fmt.Errorf("strict decode choice expected_answer: %w", err)
		}
		if rc.OptionID == "" {
			return errors.New("choice expected_answer must specify option_id")
		}
		ea.Kind = AnswerKindChoice
		ea.OptionID = rc.OptionID
		ea.Value = nil
		ea.Units = ""
		return nil

	case AnswerKindNumeric:
		type rawNumeric struct {
			Kind  AnswerKind `json:"kind"`
			Value *float64   `json:"value"`
			Units string     `json:"units"`
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var rn rawNumeric
		if err := dec.Decode(&rn); err != nil {
			return fmt.Errorf("strict decode numeric expected_answer: %w", err)
		}
		if rn.Value == nil {
			return errors.New("numeric expected_answer must specify value")
		}
		if rn.Units != "probability" {
			return fmt.Errorf("numeric expected_answer units must be \"probability\", got %q", rn.Units)
		}
		ea.Kind = AnswerKindNumeric
		ea.OptionID = ""
		ea.Value = rn.Value
		ea.Units = rn.Units
		return nil

	default:
		return fmt.Errorf("unknown expected_answer kind %q", probe.Kind)
	}
}

// NumericPolicy configures tolerance and format constraints for numeric answers.
type NumericPolicy struct {
	Version           int           `json:"version"`
	AbsoluteTolerance float64       `json:"absolute_tolerance"`
	RelativeTolerance float64       `json:"relative_tolerance"`
	AllowedForms      []NumericForm `json:"allowed_forms"`
	DisplayDecimals   *int          `json:"display_decimals"`
}

// StageTemplate describes one question stage in the question bank.
type StageTemplate struct {
	ID                  string         `json:"id"`
	Kind                StageKind      `json:"kind"`
	PromptMarkdown      string         `json:"prompt_markdown"`
	Options             []Option       `json:"options"`
	ExpectedAnswer      ExpectedAnswer `json:"expected_answer"`
	EvidenceConceptIDs  []string       `json:"evidence_concept_ids"`
	ExplanationMarkdown string         `json:"explanation_markdown"`
	NumericPolicy       *NumericPolicy `json:"numeric_policy"`
}

// QuestionTemplate describes a complete multi-stage question template.
type QuestionTemplate struct {
	SchemaVersion    int                    `json:"schema_version"`
	TemplateKind     TemplateKind           `json:"template_kind"`
	ID               string                 `json:"id"`
	Version          int                    `json:"version"`
	Status           TemplateStatus         `json:"status"`
	FamilyID         string                 `json:"family_id"`
	RuleVersion      int                    `json:"rule_version"`
	ModuleID         string                 `json:"module_id"`
	Title            string                 `json:"title"`
	ScenarioMarkdown string                 `json:"scenario_markdown"`
	Parameters       map[string]interface{} `json:"parameters"`
	Assumptions      []string               `json:"assumptions"`
	ConceptIDs       []string               `json:"concept_ids"`
	SettingGroup     string                 `json:"setting_group"`
	SourceRefs       []string               `json:"source_refs"`
	Approval         *ApprovalRecord        `json:"approval"`
	Stages           []StageTemplate        `json:"stages"`
}
