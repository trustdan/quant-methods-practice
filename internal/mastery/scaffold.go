package mastery

import (
	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// DegradeStages applies the scaffold degradation policy to a question template's stages.
// - Full: preserves all original template stages (up to 7).
// - Intermediate: reduces 7-stage drills to 4 core decisions (conditions, expression, calculation, interpretation).
// - Faded: reduces drills to 2 core decisions (calculation and interpretation).
// Invariant: concept attribution is preserved on all retained stages; fewer stages does not mean weaker numerical validation.
func DegradeStages(tmpl *domain.QuestionTemplate, scaffold ScaffoldLevel) []domain.StageTemplate {
	if tmpl == nil || len(tmpl.Stages) == 0 {
		return nil
	}

	switch scaffold {
	case ScaffoldFaded:
		if len(tmpl.Stages) <= 2 {
			return copyStages(tmpl.Stages)
		}

		// Find calculation and interpretation stages if present
		var calcStage, interpStage *domain.StageTemplate
		for i := range tmpl.Stages {
			st := &tmpl.Stages[i]
			if st.Kind == domain.StageKindNumeric || st.ID == "calculate_probability" || st.ID == "calculate_complement" {
				calcStage = st
			}
			if st.ID == "interpret_probability" || st.ID == "disjoint_vs_independent" || st.ID == "independence_check" || st.ID == "covariance_property" || st.ID == "rate_boundary_interpretation" {
				interpStage = st
			}
		}

		if calcStage != nil && interpStage != nil && calcStage.ID != interpStage.ID {
			return []domain.StageTemplate{*calcStage, *interpStage}
		}

		// Fallback: take the final 2 stages
		n := len(tmpl.Stages)
		return []domain.StageTemplate{tmpl.Stages[n-2], tmpl.Stages[n-1]}

	case ScaffoldIntermediate:
		if len(tmpl.Stages) <= 4 {
			return copyStages(tmpl.Stages)
		}

		// For 7-stage template: conditions, expression, calculation, interpretation
		if tmpl.ID == "binomial_fair_coin_exactly_two" && len(tmpl.Stages) == 7 {
			selectedIDs := []string{"check_conditions", "build_expression", "calculate_probability", "interpret_probability"}
			result := make([]domain.StageTemplate, 0, 4)
			for _, id := range selectedIDs {
				for _, st := range tmpl.Stages {
					if st.ID == id {
						result = append(result, st)
						break
					}
				}
			}
			if len(result) == 4 {
				return result
			}
		}

		// Generic reduction for > 4 stages: first, middle, second-to-last, last
		n := len(tmpl.Stages)
		return []domain.StageTemplate{
			tmpl.Stages[1], // e.g. conditions or model
			tmpl.Stages[n/2],
			tmpl.Stages[n-2],
			tmpl.Stages[n-1],
		}

	case ScaffoldFull:
		fallthrough
	default:
		return copyStages(tmpl.Stages)
	}
}

// DetermineTemplateScaffold derives the scaffold level for a template given the learner's mastery profile.
// Invariant: if any tested concept requires full guidance (e.g. due to recent error or insufficient evidence),
// full guidance is chosen.
func DetermineTemplateScaffold(tmpl *domain.QuestionTemplate, masteryMap map[string]*ConceptMastery) ScaffoldLevel {
	if tmpl == nil || len(tmpl.ConceptIDs) == 0 {
		return ScaffoldFull
	}

	// Check all concepts required by this template
	allFaded := true
	allIntermediate := true

	for _, cid := range tmpl.ConceptIDs {
		cm, ok := masteryMap[cid]
		if !ok || cm == nil {
			return ScaffoldFull
		}

		if cm.RecentError {
			return ScaffoldFull
		}

		scaffold := DetermineScaffold(cm)
		if scaffold != ScaffoldFaded {
			allFaded = false
		}
		if scaffold != ScaffoldFaded && scaffold != ScaffoldIntermediate {
			allIntermediate = false
		}
	}

	if allFaded {
		return ScaffoldFaded
	}
	if allIntermediate {
		return ScaffoldIntermediate
	}
	return ScaffoldFull
}

func copyStages(stages []domain.StageTemplate) []domain.StageTemplate {
	copied := make([]domain.StageTemplate, len(stages))
	copy(copied, stages)
	return copied
}
