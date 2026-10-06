package mastery

import (
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// EvidencePolicyVersion is the canonical version identifier for evidence policy v0.
const EvidencePolicyVersion = 1

// ScaffoldLevel defines the level of pedagogical scaffolding for a problem.
type ScaffoldLevel string

const (
	ScaffoldFull         ScaffoldLevel = "full"
	ScaffoldIntermediate ScaffoldLevel = "intermediate"
	ScaffoldFaded        ScaffoldLevel = "faded"
)

// MasteryStatus describes the learner's overall mastery category for a concept.
type MasteryStatus string

const (
	StatusNew          MasteryStatus = "new"
	StatusLearning     MasteryStatus = "learning"
	StatusTransferring MasteryStatus = "transferring"
	StatusMastered     MasteryStatus = "mastered"
)

// EvidenceOutcome defines the evaluation category of a single concept evidence event.
type EvidenceOutcome string

const (
	OutcomeIndependentSuccess EvidenceOutcome = "independent_success"
	OutcomeIndependentError   EvidenceOutcome = "independent_error"
	OutcomeAssisted           EvidenceOutcome = "assisted"
	OutcomeNone               EvidenceOutcome = "none"
)

// ConceptMastery holds the authoritative projected mastery state for a single concept.
type ConceptMastery struct {
	ConceptID               string        `json:"concept_id"`
	PolicyVersion           int           `json:"policy_version"`
	IndependentSuccesses    int           `json:"independent_successes"`
	IndependentErrors       int           `json:"independent_errors"`
	AssistedCount           int           `json:"assisted_count"`
	TotalEvidenceCount      int           `json:"total_evidence_count"`
	LastTestedAt            *time.Time    `json:"last_tested_at,omitempty"`
	BaseScore               float64       `json:"base_score"`
	DecayedScore            float64       `json:"decayed_score"`
	RetentionFactor         float64       `json:"retention_factor"`
	HalfLifeDays            float64       `json:"half_life_days"`
	Status                  MasteryStatus `json:"status"`
	ScaffoldLevel           ScaffoldLevel `json:"scaffold_level"`
	SettingGroupsSeen       []string      `json:"setting_groups_seen"`
	DelayedTransferAchieved bool          `json:"delayed_transfer_achieved"`
	RecentError             bool          `json:"recent_error"`
	PriorityScore           float64       `json:"priority_score"`
}

// RawExposure captures an immutable event for concept evidence derivation.
type RawExposure struct {
	SessionID     string                  `json:"session_id"`
	InstanceID    string                  `json:"instance_id"`
	TemplateID    string                  `json:"template_id"`
	SettingGroup  string                  `json:"setting_group"`
	StageID       string                  `json:"stage_id"`
	ConceptIDs    []string                `json:"concept_ids"`
	AttemptNumber int                     `json:"attempt_number"`
	IsCorrect     bool                    `json:"is_correct"`
	Assistance    []domain.AssistanceType `json:"assistance"`
	Timestamp     time.Time               `json:"timestamp"`
	IsContrast    bool                    `json:"is_contrast"`
}

// MasterySummary bundles the complete curriculum mastery projection for the client.
type MasterySummary struct {
	PolicyVersion     int              `json:"policy_version"`
	OverallScore      float64          `json:"overall_score"`
	TotalMastered     int              `json:"total_mastered"`
	TotalTransferring int              `json:"total_transferring"`
	TotalLearning     int              `json:"total_learning"`
	TotalNew          int              `json:"total_new"`
	Concepts          []ConceptMastery `json:"concepts"`
	GeneratedAt       time.Time        `json:"generated_at"`
}
