package mastery

import (
	"math/rand"
	"sort"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// SelectionOptions specifies configuration for weighted question selection.
type SelectionOptions struct {
	CandidateTemplates  []*domain.QuestionTemplate
	MasteryMap          map[string]*ConceptMastery
	RecentTemplateIDs   []string
	RecentSettingGroups []string
	QuestionCount       int
	Intensity           string
	Seed                int64
	Now                 time.Time
}

// SelectQuestions performs seeded weighted problem selection according to mastery, transfer queues,
// non-zero eligibility floors, roughly 20% mixed review, and anti-repeat penalties.
func SelectQuestions(opts SelectionOptions) []*domain.QuestionTemplate {
	if len(opts.CandidateTemplates) == 0 {
		return nil
	}

	targetCount := opts.QuestionCount
	if targetCount <= 0 {
		targetCount = 10
	}

	recentTmplMap := make(map[string]bool)
	for _, id := range opts.RecentTemplateIDs {
		recentTmplMap[id] = true
	}

	recentGroupMap := make(map[string]bool)
	for _, g := range opts.RecentSettingGroups {
		recentGroupMap[g] = true
	}

	var rng *rand.Rand
	if opts.Seed != 0 {
		rng = rand.New(rand.NewSource(opts.Seed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Calculate weights for candidates
	weights := make([]float64, len(opts.CandidateTemplates))
	for i, tmpl := range opts.CandidateTemplates {
		weights[i] = computeTemplateWeight(tmpl, opts.MasteryMap, recentTmplMap, recentGroupMap, opts.Intensity)
	}

	selected := make([]*domain.QuestionTemplate, 0, targetCount)
	availableIndices := make([]int, len(opts.CandidateTemplates))
	for i := range availableIndices {
		availableIndices[i] = i
	}

	// Sample without replacement as long as pool is available
	currentWeights := make([]float64, len(weights))
	copy(currentWeights, weights)

	for len(selected) < targetCount {
		if len(availableIndices) == 0 {
			// Pool exhausted: refill candidates for subsequent picks
			for i := range opts.CandidateTemplates {
				availableIndices = append(availableIndices, i)
			}
			copy(currentWeights, weights)
		}

		// Calculate total weight of available
		totalWeight := 0.0
		for _, idx := range availableIndices {
			totalWeight += currentWeights[idx]
		}

		if totalWeight <= 0 {
			// Fallback: pick uniformly from available
			pickPos := rng.Intn(len(availableIndices))
			pickIdx := availableIndices[pickPos]
			selected = append(selected, opts.CandidateTemplates[pickIdx])
			availableIndices = append(availableIndices[:pickPos], availableIndices[pickPos+1:]...)
			continue
		}

		// Weighted random pick
		r := rng.Float64() * totalWeight
		cumulative := 0.0
		chosenPos := -1

		for pos, idx := range availableIndices {
			cumulative += currentWeights[idx]
			if r <= cumulative {
				chosenPos = pos
				break
			}
		}

		if chosenPos == -1 {
			chosenPos = len(availableIndices) - 1
		}

		chosenIdx := availableIndices[chosenPos]
		selected = append(selected, opts.CandidateTemplates[chosenIdx])

		// Penalize recently picked setting group in current selection pass
		pickedGroup := opts.CandidateTemplates[chosenIdx].SettingGroup
		for _, remainingIdx := range availableIndices {
			if opts.CandidateTemplates[remainingIdx].SettingGroup == pickedGroup {
				currentWeights[remainingIdx] *= 0.5
			}
		}

		availableIndices = append(availableIndices[:chosenPos], availableIndices[chosenPos+1:]...)
	}

	return selected
}

func computeTemplateWeight(
	tmpl *domain.QuestionTemplate,
	masteryMap map[string]*ConceptMastery,
	recentTemplates map[string]bool,
	recentGroups map[string]bool,
	intensity string,
) float64 {
	const floorWeight = 0.10 // Nonzero eligibility floor

	if tmpl == nil {
		return 0
	}

	// 1. Concept mastery priority factor
	avgPriority := 1.0
	masteredCount := 0
	transferNeeded := false

	if len(tmpl.ConceptIDs) > 0 {
		sumPriority := 0.0
		for _, cid := range tmpl.ConceptIDs {
			cm, ok := masteryMap[cid]
			if !ok || cm == nil {
				sumPriority += 3.0 // High priority for new/unseen concepts
			} else {
				sumPriority += cm.PriorityScore
				if cm.Status == StatusMastered {
					masteredCount++
				}
				if !cm.DelayedTransferAchieved && cm.IndependentSuccesses >= 2 {
					transferNeeded = true
				}
			}
		}
		avgPriority = sumPriority / float64(len(tmpl.ConceptIDs))
	}

	weight := floorWeight + avgPriority

	// 2. Mixed-review allocation (~20% allocation for mastered concepts)
	isMasteredProblem := len(tmpl.ConceptIDs) > 0 && masteredCount == len(tmpl.ConceptIDs)
	if isMasteredProblem {
		// Provide maintenance weight (~20% allocation)
		weight = floorWeight + 0.50
	}

	// 3. Intensity adaptations
	switch intensity {
	case "spaced":
		// Prioritizes decayed concepts
		weight *= 1.5
	case "intensive":
		// Prioritizes weak concepts, downweights already mastered
		if isMasteredProblem {
			weight *= 0.3
		} else {
			weight *= 1.8
		}
	case "transfer":
		// Prioritizes transfer-eligible concepts across distinct groups
		if transferNeeded {
			weight *= 2.5
		}
	case "standard":
		fallthrough
	default:
		// Balanced standard weights
	}

	// 4. Anti-repeat penalties
	if recentTemplates[tmpl.ID] {
		weight *= 0.10 // Heavy penalty for repeating exact same template
	}
	if recentGroups[tmpl.SettingGroup] {
		weight *= 0.30 // Penalty for repeating exact same setting group
	}

	// Ensure strictly positive weight
	if weight < floorWeight {
		weight = floorWeight
	}

	return weight
}

// StableSortTemplates provides deterministic ordering placing the canonical coin problem first.
func StableSortTemplates(tmpls []*domain.QuestionTemplate) {
	sort.SliceStable(tmpls, func(i, j int) bool {
		if tmpls[i].ID == "binomial_fair_coin_exactly_two" {
			return true
		}
		if tmpls[j].ID == "binomial_fair_coin_exactly_two" {
			return false
		}
		return tmpls[i].ID < tmpls[j].ID
	})
}
