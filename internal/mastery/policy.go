package mastery

import (
	"math"
	"time"
)

const (
	// DefaultHalfLifeDays defines the default half-life for time decay (3 days).
	DefaultHalfLifeDays = 3.0

	// IntermediateMinSuccesses is the minimum number of independent successes for intermediate guidance.
	IntermediateMinSuccesses = 2

	// IntermediateMinScore is the minimum decayed score for intermediate guidance.
	IntermediateMinScore = 0.60

	// FadedMinSuccesses is the minimum number of independent successes for faded guidance.
	FadedMinSuccesses = 3

	// FadedMinScore is the minimum decayed score for faded guidance.
	FadedMinScore = 0.80

	// MinSettingGroupsForTransfer is the minimum number of distinct setting groups required for transfer.
	MinSettingGroupsForTransfer = 2

	// DelayedInterval is the minimum duration required between exposures to award delayed retrieval credit (10 minutes).
	DelayedInterval = 10 * time.Minute
)

// CalculateBaseScore computes the Bayesian mean base score with alpha=1, beta=1 priors:
// (1 + successes) / (1 + successes + 1 + errors) = (1 + s) / (2 + s + e).
func CalculateBaseScore(successes, errors int) float64 {
	if successes < 0 {
		successes = 0
	}
	if errors < 0 {
		errors = 0
	}
	alpha := 1.0 + float64(successes)
	beta := 1.0 + float64(errors)
	return alpha / (alpha + beta)
}

// CalculateDecay computes the decayed score and retention factor at read time given the reference time 'now'.
// Invariant: clock rollback is protected and cannot inflate mastery (elapsed days clamped to >= 0).
func CalculateDecay(baseScore float64, lastTested *time.Time, now time.Time, halfLifeDays float64) (decayedScore, retention float64) {
	if halfLifeDays <= 0 {
		halfLifeDays = DefaultHalfLifeDays
	}

	if lastTested == nil {
		return baseScore, 1.0
	}

	// Clock rollback protection: if current time is before last tested, clamp elapsed to 0
	if now.Before(*lastTested) {
		return baseScore, 1.0
	}

	elapsedHours := now.Sub(*lastTested).Hours()
	elapsedDays := elapsedHours / 24.0

	// Retention factor: 2^(-elapsed_days / half_life_days)
	retention = math.Pow(2.0, -elapsedDays/halfLifeDays)
	if retention < 0 {
		retention = 0
	} else if retention > 1.0 {
		retention = 1.0
	}

	decayedScore = baseScore * retention
	return decayedScore, retention
}

// DetermineScaffold derives the appropriate scaffold level based on learner mastery and transfer.
// Invariant: an error restores full guidance immediately.
func DetermineScaffold(cm *ConceptMastery) ScaffoldLevel {
	if cm == nil {
		return ScaffoldFull
	}

	// An error restores full guidance
	if cm.RecentError {
		return ScaffoldFull
	}

	// Faded guidance requires: >= 3 independent successes, score >= 0.80, successes in >= 2 groups, delayed transfer
	if cm.IndependentSuccesses >= FadedMinSuccesses &&
		cm.DecayedScore >= (FadedMinScore-0.01) &&
		len(cm.SettingGroupsSeen) >= MinSettingGroupsForTransfer &&
		cm.DelayedTransferAchieved {
		return ScaffoldFaded
	}

	// Intermediate guidance requires: >= 2 independent successes, score >= 0.60, successes in >= 2 groups, delayed transfer
	if cm.IndependentSuccesses >= IntermediateMinSuccesses &&
		cm.DecayedScore >= (IntermediateMinScore-0.01) &&
		len(cm.SettingGroupsSeen) >= MinSettingGroupsForTransfer &&
		cm.DelayedTransferAchieved {
		return ScaffoldIntermediate
	}

	return ScaffoldFull
}

// DetermineStatus determines the learner-facing category status for a concept.
func DetermineStatus(cm *ConceptMastery) MasteryStatus {
	if cm == nil || cm.TotalEvidenceCount == 0 {
		return StatusNew
	}

	if cm.ScaffoldLevel == ScaffoldFaded {
		return StatusMastered
	}

	if cm.ScaffoldLevel == ScaffoldIntermediate {
		return StatusTransferring
	}

	return StatusLearning
}

// CalculatePriorityScore computes a weighted priority score for selection.
// Higher score indicates higher urgency/need for practice.
func CalculatePriorityScore(cm *ConceptMastery) float64 {
	if cm == nil || cm.TotalEvidenceCount == 0 {
		return 3.0 // High priority for unseen concepts
	}

	// Weak concept penalty (urgency boost)
	urgency := 0.0
	if cm.DecayedScore < 0.60 {
		urgency += 2.0 * (0.60 - cm.DecayedScore)
	}

	// Decay penalty (need for spaced review)
	decayNeed := 1.5 * (1.0 - cm.RetentionFactor)

	// If recent error, boost urgency to reinforce
	if cm.RecentError {
		urgency += 1.5
	}

	// If transfer not yet achieved, boost priority for transfer
	if !cm.DelayedTransferAchieved && cm.IndependentSuccesses >= 2 {
		urgency += 1.0
	}

	return 0.5 + urgency + decayNeed
}
