package bank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// StrictDecodeTemplate deserializes JSON bytes into a QuestionTemplate,
// rejecting any unknown fields at all structural levels.
func StrictDecodeTemplate(data []byte) (*domain.QuestionTemplate, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var tmpl domain.QuestionTemplate
	if err := dec.Decode(&tmpl); err != nil {
		return nil, fmt.Errorf("strict decode template: %w", err)
	}

	// Ensure no trailing tokens
	if dec.More() {
		return nil, errors.New("strict decode template: unexpected content after JSON document")
	}

	return &tmpl, nil
}

// ValidateTemplate performs comprehensive semantic validation against domain and family contracts.
func ValidateTemplate(tmpl *domain.QuestionTemplate, reg *Registry) error {
	if tmpl == nil {
		return errors.New("template cannot be nil")
	}

	if reg == nil {
		reg = DefaultRegistry()
	}

	if tmpl.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version %d (expected 1)", tmpl.SchemaVersion)
	}

	if tmpl.TemplateKind != domain.TemplateKindFixedReference {
		return fmt.Errorf("unsupported template_kind %q (expected %q)", tmpl.TemplateKind, domain.TemplateKindFixedReference)
	}

	if err := domain.ValidateID("id", tmpl.ID); err != nil {
		return err
	}

	if tmpl.Version < 1 {
		return fmt.Errorf("version must be >= 1, got %d", tmpl.Version)
	}

	switch tmpl.Status {
	case domain.StatusDraft, domain.StatusActive, domain.StatusRetired:
		// Valid statuses
	default:
		return fmt.Errorf("invalid status %q; must be draft, active, or retired", tmpl.Status)
	}

	// Active templates require human approval
	if tmpl.Status == domain.StatusActive {
		if tmpl.Approval == nil {
			return errors.New("active template must have an approval record")
		}
		if tmpl.Approval.Reviewer == "" {
			return errors.New("approval record must specify reviewer")
		}
		if tmpl.Approval.ReviewedAt.IsZero() {
			return errors.New("approval record must specify non-zero reviewed_at")
		}
		if tmpl.Approval.Notes == "" {
			return errors.New("approval record must specify notes")
		}
	}

	// Check family registry
	fam, ok := reg.Get(tmpl.FamilyID)
	if !ok {
		return fmt.Errorf("unsupported family_id %q", tmpl.FamilyID)
	}

	supportedRule := false
	for _, v := range fam.SupportedRuleVersions() {
		if v == tmpl.RuleVersion {
			supportedRule = true
			break
		}
	}
	if !supportedRule {
		return fmt.Errorf("family %q does not support rule_version %d", tmpl.FamilyID, tmpl.RuleVersion)
	}

	if err := fam.ValidateParameters(tmpl.RuleVersion, tmpl.Parameters); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	if err := domain.ValidateID("module_id", tmpl.ModuleID); err != nil {
		return err
	}

	if tmpl.Title == "" {
		return errors.New("title cannot be empty")
	}

	if tmpl.ScenarioMarkdown == "" {
		return errors.New("scenario_markdown cannot be empty")
	}

	if len(tmpl.Assumptions) == 0 {
		return errors.New("assumptions must contain at least one assumption")
	}
	assumptionSet := make(map[string]bool)
	for _, a := range tmpl.Assumptions {
		if a == "" {
			return errors.New("assumption cannot be empty")
		}
		if assumptionSet[a] {
			return fmt.Errorf("duplicate assumption: %q", a)
		}
		assumptionSet[a] = true
	}

	if len(tmpl.SourceRefs) == 0 {
		return errors.New("source_refs must contain at least one reference")
	}
	for _, ref := range tmpl.SourceRefs {
		if ref == "" {
			return errors.New("source_ref cannot be empty")
		}
	}

	if err := domain.ValidateID("setting_group", tmpl.SettingGroup); err != nil {
		return err
	}

	if len(tmpl.ConceptIDs) == 0 {
		return errors.New("concept_ids must contain at least one concept")
	}
	conceptSet := make(map[string]bool)
	for _, c := range tmpl.ConceptIDs {
		if err := domain.ValidateID("concept_id", c); err != nil {
			return err
		}
		if conceptSet[c] {
			return fmt.Errorf("duplicate concept_id: %q", c)
		}
		conceptSet[c] = true
	}

	// Validate stages
	if len(tmpl.Stages) < 1 || len(tmpl.Stages) > 7 {
		return fmt.Errorf("stages count must be between 1 and 7, got %d", len(tmpl.Stages))
	}

	stageIDSet := make(map[string]bool)
	assignedEvidenceConcepts := make(map[string]string) // concept_id -> stage_id

	for i, stage := range tmpl.Stages {
		if err := domain.ValidateID(fmt.Sprintf("stage[%d].id", i), stage.ID); err != nil {
			return err
		}
		if stageIDSet[stage.ID] {
			return fmt.Errorf("duplicate stage id %q", stage.ID)
		}
		stageIDSet[stage.ID] = true

		if stage.PromptMarkdown == "" {
			return fmt.Errorf("stage %q prompt_markdown cannot be empty", stage.ID)
		}
		if stage.ExplanationMarkdown == "" {
			return fmt.Errorf("stage %q explanation_markdown cannot be empty", stage.ID)
		}

		// Evidence concepts
		for _, ec := range stage.EvidenceConceptIDs {
			if !conceptSet[ec] {
				return fmt.Errorf("stage %q evidence_concept_id %q is not in template concept_ids", stage.ID, ec)
			}
			if priorStage, seen := assignedEvidenceConcepts[ec]; seen {
				return fmt.Errorf("evidence_concept_id %q assigned to multiple stages (%q and %q)", ec, priorStage, stage.ID)
			}
			assignedEvidenceConcepts[ec] = stage.ID
		}

		switch stage.Kind {
		case domain.StageKindChoice:
			if len(stage.Options) < 2 || len(stage.Options) > 4 {
				return fmt.Errorf("stage %q is choice kind: must have between 2 and 4 options, got %d", stage.ID, len(stage.Options))
			}
			if stage.NumericPolicy != nil {
				return fmt.Errorf("stage %q is choice kind: numeric_policy must be null", stage.ID)
			}
			if stage.ExpectedAnswer.Kind != domain.AnswerKindChoice {
				return fmt.Errorf("stage %q is choice kind: expected_answer.kind must be choice", stage.ID)
			}

			optIDSet := make(map[string]bool)
			foundExpected := false
			for j, opt := range stage.Options {
				if err := domain.ValidateID(fmt.Sprintf("stage[%s].options[%d].id", stage.ID, j), opt.ID); err != nil {
					return err
				}
				if optIDSet[opt.ID] {
					return fmt.Errorf("stage %q duplicate option id %q", stage.ID, opt.ID)
				}
				optIDSet[opt.ID] = true

				if opt.TextMarkdown == "" {
					return fmt.Errorf("stage %q option %q text_markdown cannot be empty", stage.ID, opt.ID)
				}
				if opt.MisconceptionID != nil {
					if err := domain.ValidateID(fmt.Sprintf("stage[%s].options[%s].misconception_id", stage.ID, opt.ID), *opt.MisconceptionID); err != nil {
						return err
					}
				}
				if opt.HintMarkdown != nil && *opt.HintMarkdown == "" {
					return fmt.Errorf("stage %q option %q hint_markdown cannot be empty string (use null instead)", stage.ID, opt.ID)
				}
				if opt.ID == stage.ExpectedAnswer.OptionID {
					foundExpected = true
				}
			}
			if !foundExpected {
				return fmt.Errorf("stage %q expected_answer option_id %q not found in options", stage.ID, stage.ExpectedAnswer.OptionID)
			}

		case domain.StageKindNumeric:
			if len(stage.Options) != 0 {
				return fmt.Errorf("stage %q is numeric kind: options must be empty, got %d options", stage.ID, len(stage.Options))
			}
			if stage.NumericPolicy == nil {
				return fmt.Errorf("stage %q is numeric kind: numeric_policy must not be null", stage.ID)
			}
			if stage.ExpectedAnswer.Kind != domain.AnswerKindNumeric {
				return fmt.Errorf("stage %q is numeric kind: expected_answer.kind must be numeric", stage.ID)
			}
			if stage.ExpectedAnswer.Value == nil {
				return fmt.Errorf("stage %q numeric expected_answer must specify value", stage.ID)
			}
			if stage.ExpectedAnswer.Units != "probability" {
				return fmt.Errorf("stage %q numeric expected_answer units must be \"probability\", got %q", stage.ID, stage.ExpectedAnswer.Units)
			}

			pol := stage.NumericPolicy
			if pol.Version < 1 {
				return fmt.Errorf("stage %q numeric_policy version must be >= 1, got %d", stage.ID, pol.Version)
			}
			if pol.AbsoluteTolerance < 0 {
				return fmt.Errorf("stage %q numeric_policy absolute_tolerance cannot be negative", stage.ID)
			}
			if pol.RelativeTolerance < 0 {
				return fmt.Errorf("stage %q numeric_policy relative_tolerance cannot be negative", stage.ID)
			}
			if len(pol.AllowedForms) == 0 {
				return fmt.Errorf("stage %q numeric_policy must allow at least one form", stage.ID)
			}
			formSet := make(map[domain.NumericForm]bool)
			for _, form := range pol.AllowedForms {
				switch form {
				case domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction:
					// Valid form
				default:
					return fmt.Errorf("stage %q numeric_policy invalid allowed_form %q", stage.ID, form)
				}
				if formSet[form] {
					return fmt.Errorf("stage %q numeric_policy duplicate allowed_form %q", stage.ID, form)
				}
				formSet[form] = true
			}
			if pol.DisplayDecimals != nil {
				if *pol.DisplayDecimals < 0 || *pol.DisplayDecimals > 12 {
					return fmt.Errorf("stage %q numeric_policy display_decimals must be between 0 and 12, got %d", stage.ID, *pol.DisplayDecimals)
				}
			}

		default:
			return fmt.Errorf("stage %q has unknown kind %q", stage.ID, stage.Kind)
		}
	}

	// Verify all concept_ids were designated to a stage
	if len(assignedEvidenceConcepts) != len(tmpl.ConceptIDs) {
		var missing []string
		for _, c := range tmpl.ConceptIDs {
			if _, ok := assignedEvidenceConcepts[c]; !ok {
				missing = append(missing, c)
			}
		}
		return fmt.Errorf("the following concept_ids were not designated to any stage: %v", missing)
	}

	return nil
}
