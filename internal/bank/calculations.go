package bank

import (
	"fmt"
	"math"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func validateCalculations(t *domain.QuestionTemplate) error {
	var quantities map[string]float64
	if t.RuleVersion == 2 {
		var err error
		quantities, err = mathengine.CurriculumCalculations(t.FamilyID, t.Parameters)
		if err != nil {
			return err
		}
	}
	count := 0
	for _, s := range t.Stages {
		if t.RuleVersion == 2 {
			for _, o := range s.Options {
				if o.ID == s.ExpectedAnswer.OptionID {
					if o.MisconceptionID != nil || o.HintMarkdown != nil {
						return fmt.Errorf("correct option must not carry a misconception or hint")
					}
				} else if o.MisconceptionID == nil || o.HintMarkdown == nil {
					return fmt.Errorf("distractor %s requires a misconception and causal hint", o.ID)
				}
			}
		}
		if s.Calculation == "" {
			for _, o := range s.Options {
				if o.NumericValue != nil {
					return fmt.Errorf("numeric_value requires a calculation binding")
				}
			}
			continue
		}
		if t.RuleVersion != 2 || s.Kind != domain.StageKindChoice {
			return fmt.Errorf("calculation bindings require rule-v2 choice stages")
		}
		value, ok := quantities[s.Calculation]
		if !ok {
			return fmt.Errorf("unknown calculation %q", s.Calculation)
		}
		count++
		matches := 0
		for _, o := range s.Options {
			if o.NumericValue == nil || math.IsNaN(*o.NumericValue) || math.IsInf(*o.NumericValue, 0) {
				return fmt.Errorf("calculated option %s requires a finite numeric_value", o.ID)
			}
			// Four displayed decimals are part of rule v2. Labels cannot
			// contradict their values, and colliding distractors fail closed.
			if o.TextMarkdown != fmt.Sprintf("$%.4f$", *o.NumericValue) {
				return fmt.Errorf("calculated option %s label disagrees with numeric_value", o.ID)
			}
			match := fmt.Sprintf("%.4f", *o.NumericValue) == fmt.Sprintf("%.4f", value)
			if match {
				matches++
			}
			if (o.ID == s.ExpectedAnswer.OptionID) != match {
				return fmt.Errorf("stage %s key disagrees with engine calculation %s", s.ID, s.Calculation)
			}
		}
		if matches != 1 {
			return fmt.Errorf("stage %s has %d canonical matches", s.ID, matches)
		}
	}
	if t.RuleVersion == 2 && count == 0 {
		return fmt.Errorf("rule-v2 template requires a calculation binding")
	}
	return nil
}
