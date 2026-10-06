package mastery

import (
	"sort"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// EvidenceLedger derives deterministic mastery projections from an immutable chronological stream of exposures.
type EvidenceLedger struct {
	policyVersion int
	halfLifeDays  float64
}

// NewEvidenceLedger creates a ledger with configured policy version and half-life.
func NewEvidenceLedger(halfLifeDays float64) *EvidenceLedger {
	if halfLifeDays <= 0 {
		halfLifeDays = DefaultHalfLifeDays
	}
	return &EvidenceLedger{
		policyVersion: EvidencePolicyVersion,
		halfLifeDays:  halfLifeDays,
	}
}

type conceptHistory struct {
	independentSuccesses int
	independentErrors    int
	assistedCount        int
	totalEvidenceCount   int
	lastTestedAt         *time.Time
	lastExposureAt       *time.Time
	settingGroupsSeen    map[string]bool
	lastSuccessGroup     string
	delayedTransfer      bool
	recentError          bool
}

// ProcessExposures derives authoritative ConceptMastery records from raw chronological exposures.
// Invariants enforced:
// 1. Designate at most one evidence stage per concept per instance.
// 2. An independent first success adds one success, an independent first error adds one error.
// 3. Hinted retry never erases the first error.
// 4. Pre-answer hints, reference use, earlier answer reveal or guided contrast make dependent responses assisted.
// 5. Review, duplicate commands, and repeated visits cannot add counts.
// 6. Historical attempts without reviewed setting metadata retain original results but receive no invented transfer credit.
func (l *EvidenceLedger) ProcessExposures(exposures []RawExposure, now time.Time) map[string]*ConceptMastery {
	// Sort exposures chronologically
	sorted := make([]RawExposure, len(exposures))
	copy(sorted, exposures)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Track instance-to-concept contributions: instanceID + ":" + conceptID -> true
	instanceConceptEvaluated := make(map[string]bool)

	histories := make(map[string]*conceptHistory)

	for _, exp := range sorted {
		if len(exp.ConceptIDs) == 0 {
			continue
		}

		// Check if exposure has assistance
		isAssisted := exp.IsContrast
		if !isAssisted {
			for _, a := range exp.Assistance {
				if a == domain.AssistanceHint ||
					a == domain.AssistanceReference ||
					a == domain.AssistanceGuidedContrast ||
					a == domain.AssistanceSolutionReveal ||
					a == domain.AssistanceTutor {
					isAssisted = true
					break
				}
			}
		}

		for _, cid := range exp.ConceptIDs {
			if cid == "" {
				continue
			}

			ch, exists := histories[cid]
			if !exists {
				ch = &conceptHistory{
					settingGroupsSeen: make(map[string]bool),
				}
				histories[cid] = ch
			}

			// Invariant 1: Designate at most one evidence stage per concept per instance
			dedupKey := exp.InstanceID + ":" + cid
			if instanceConceptEvaluated[dedupKey] {
				// Record exposure timestamp for retrieval spacing tracking, but do not award additional counts
				ts := exp.Timestamp
				ch.lastExposureAt = &ts
				continue
			}

			// First evaluation for this concept in this instance
			instanceConceptEvaluated[dedupKey] = true
			ts := exp.Timestamp
			ch.lastTestedAt = &ts
			ch.totalEvidenceCount++

			if isAssisted {
				// Invariant 4: Assisted response
				ch.assistedCount++
			} else {
				// Independent attempt
				if exp.AttemptNumber == 1 && exp.IsCorrect {
					// Invariant 2: Independent first success
					ch.independentSuccesses++
					ch.recentError = false

					if exp.SettingGroup != "" {
						ch.settingGroupsSeen[exp.SettingGroup] = true

						// Check delayed transfer: success in a different setting group with >= 10m elapsed
						if ch.lastSuccessGroup != "" && ch.lastSuccessGroup != exp.SettingGroup {
							if ch.lastExposureAt != nil && exp.Timestamp.Sub(*ch.lastExposureAt) >= DelayedInterval {
								ch.delayedTransfer = true
							}
						}
						ch.lastSuccessGroup = exp.SettingGroup
					}
				} else {
					// Invariant 2 & 3: Independent first error (hinted retry never erases first error)
					ch.independentErrors++
					ch.recentError = true
				}
			}

			ch.lastExposureAt = &ts
		}
	}

	result := make(map[string]*ConceptMastery, len(histories))
	for cid, ch := range histories {
		groups := make([]string, 0, len(ch.settingGroupsSeen))
		for g := range ch.settingGroupsSeen {
			groups = append(groups, g)
		}
		sort.Strings(groups)

		baseScore := CalculateBaseScore(ch.independentSuccesses, ch.independentErrors)
		decayedScore, retention := CalculateDecay(baseScore, ch.lastTestedAt, now, l.halfLifeDays)

		cm := &ConceptMastery{
			ConceptID:               cid,
			PolicyVersion:           l.policyVersion,
			IndependentSuccesses:    ch.independentSuccesses,
			IndependentErrors:       ch.independentErrors,
			AssistedCount:           ch.assistedCount,
			TotalEvidenceCount:      ch.totalEvidenceCount,
			LastTestedAt:            ch.lastTestedAt,
			BaseScore:               baseScore,
			DecayedScore:            decayedScore,
			RetentionFactor:         retention,
			HalfLifeDays:            l.halfLifeDays,
			SettingGroupsSeen:       groups,
			DelayedTransferAchieved: ch.delayedTransfer,
			RecentError:             ch.recentError,
		}

		cm.ScaffoldLevel = DetermineScaffold(cm)
		cm.Status = DetermineStatus(cm)
		cm.PriorityScore = CalculatePriorityScore(cm)

		result[cid] = cm
	}

	return result
}
