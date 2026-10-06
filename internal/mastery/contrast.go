package mastery

// ContrastPair defines a reviewed pairing between a source misconception and a target contrast partner template.
type ContrastPair struct {
	OriginTemplateID string `json:"origin_template_id"`
	MisconceptionID  string `json:"misconception_id"`
	TargetTemplateID string `json:"target_template_id"`
	Description      string `json:"description"`
}

var reviewedContrastRegistry = []ContrastPair{
	{
		OriginTemplateID: "binomial_fair_coin_exactly_two",
		MisconceptionID:  "exactly_as_at_most",
		TargetTemplateID: "binomial_defective_at_most_one",
		Description:      "Contrast exactly two ($X=2$) with at most one ($X\\le 1$) cumulative inspection",
	},
	{
		OriginTemplateID: "binomial_fair_coin_exactly_two",
		MisconceptionID:  "exactly_as_at_least",
		TargetTemplateID: "binomial_defective_at_most_one",
		Description:      "Contrast exactly two ($X=2$) with boundary inspection",
	},
	{
		OriginTemplateID: "binomial_defective_at_most_one",
		MisconceptionID:  "missing_combination",
		TargetTemplateID: "binomial_fair_coin_exactly_two",
		Description:      "Contrast cumulative combinations with exact coin permutations",
	},
	{
		OriginTemplateID: "set_probability_conditional",
		MisconceptionID:  "independence_concept",
		TargetTemplateID: "set_probability_union_rule",
		Description:      "Contrast conditional dependence ($P(A|B)$) with addition rule overlap ($P(A\\cap B)$)",
	},
	{
		OriginTemplateID: "set_probability_union_rule",
		MisconceptionID:  "disjoint_concept",
		TargetTemplateID: "set_probability_conditional",
		Description:      "Contrast disjoint events with independent conditional events",
	},
	{
		OriginTemplateID: "poisson_rare_event_one",
		MisconceptionID:  "poisson_rate_scaling",
		TargetTemplateID: "poisson_server_failures_zero",
		Description:      "Contrast unit rate Poisson with zero-rate boundary Poisson",
	},
	{
		OriginTemplateID: "poisson_server_failures_zero",
		MisconceptionID:  "poisson_zero_rate",
		TargetTemplateID: "poisson_call_center_arrivals",
		Description:      "Contrast zero event boundary with discrete arrival counts",
	},
}

// FindContrastPartner looks up an approved contrast partner template ID for an eligible misconception.
func FindContrastPartner(originTmplID, misconceptionID string) (targetTmplID string, found bool) {
	if originTmplID == "" || misconceptionID == "" {
		return "", false
	}

	for _, pair := range reviewedContrastRegistry {
		if pair.OriginTemplateID == originTmplID && pair.MisconceptionID == misconceptionID {
			return pair.TargetTemplateID, true
		}
	}
	return "", false
}

// AllReviewedContrasts returns all configured reviewed contrast pairs.
func AllReviewedContrasts() []ContrastPair {
	res := make([]ContrastPair, len(reviewedContrastRegistry))
	copy(res, reviewedContrastRegistry)
	return res
}
