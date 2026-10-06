package mastery

import (
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func TestOneContributionPerConceptPerInstance(t *testing.T) {
	ledger := NewEvidenceLedger(3.0)
	now := time.Now()

	// Two exposures for the SAME instance and concept (e.g. repeated submission / retry)
	exposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_1",
			SettingGroup: "group_a",
			StageID:      "stage_calc",
			ConceptIDs:   []string{"calc_concept"},
			AttemptNumber: 1,
			IsCorrect:    false,
			Timestamp:    now.Add(-20 * time.Minute),
		},
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_1",
			SettingGroup: "group_a",
			StageID:      "stage_calc",
			ConceptIDs:   []string{"calc_concept"},
			AttemptNumber: 2,
			IsCorrect:    true,
			Assistance:   []domain.AssistanceType{domain.AssistanceRetry},
			Timestamp:    now.Add(-19 * time.Minute),
		},
	}

	projections := ledger.ProcessExposures(exposures, now)
	cm, ok := projections["calc_concept"]
	if !ok {
		t.Fatalf("expected projection for calc_concept")
	}

	// Invariant: Total evidence count must be exactly 1 per instance
	if cm.TotalEvidenceCount != 1 {
		t.Errorf("expected TotalEvidenceCount=1, got %d", cm.TotalEvidenceCount)
	}

	// Invariant: Hinted retry never erases the first error
	if cm.IndependentErrors != 1 {
		t.Errorf("expected IndependentErrors=1, got %d", cm.IndependentErrors)
	}
	if cm.IndependentSuccesses != 0 {
		t.Errorf("expected IndependentSuccesses=0, got %d", cm.IndependentSuccesses)
	}
}

func TestAssistedAttemptsDoNotAwardIndependentCredit(t *testing.T) {
	ledger := NewEvidenceLedger(3.0)
	now := time.Now()

	exposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_1",
			SettingGroup: "group_a",
			StageID:      "stage_1",
			ConceptIDs:   []string{"concept_hinted"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Assistance:   []domain.AssistanceType{domain.AssistanceHint}, // Pre-answer hint requested
			Timestamp:    now.Add(-10 * time.Minute),
		},
		{
			SessionID:    "s2",
			InstanceID:   "inst_2",
			TemplateID:   "tmpl_2",
			SettingGroup: "group_b",
			StageID:      "stage_1",
			ConceptIDs:   []string{"concept_contrast"},
			AttemptNumber: 1,
			IsCorrect:    true,
			IsContrast:   true, // Guided contrast problem
			Timestamp:    now.Add(-5 * time.Minute),
		},
	}

	projections := ledger.ProcessExposures(exposures, now)

	ch := projections["concept_hinted"]
	if ch.IndependentSuccesses != 0 || ch.AssistedCount != 1 {
		t.Errorf("hinted attempt should be counted as assisted, got indSuccess=%d, assisted=%d",
			ch.IndependentSuccesses, ch.AssistedCount)
	}

	cc := projections["concept_contrast"]
	if cc.IndependentSuccesses != 0 || cc.AssistedCount != 1 {
		t.Errorf("contrast attempt should be counted as assisted, got indSuccess=%d, assisted=%d",
			cc.IndependentSuccesses, cc.AssistedCount)
	}
}

func TestDecayCalculationAndClockRollbackProtection(t *testing.T) {
	baseScore := 0.80 // 80%
	halfLife := 3.0   // 3 days
	t0 := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	// Exactly 3 days elapsed: retention should be 0.50, decayed score 0.40
	t3Days := t0.Add(72 * time.Hour)
	score3d, ret3d := CalculateDecay(baseScore, &t0, t3Days, halfLife)
	if ret3d < 0.499 || ret3d > 0.501 {
		t.Errorf("expected retention ~0.50 at 3 days, got %f", ret3d)
	}
	if score3d < 0.399 || score3d > 0.401 {
		t.Errorf("expected decayed score ~0.40 at 3 days, got %f", score3d)
	}

	// Exactly 6 days elapsed (2 half-lives): retention should be 0.25, score 0.20
	t6Days := t0.Add(144 * time.Hour)
	score6d, ret6d := CalculateDecay(baseScore, &t0, t6Days, halfLife)
	if ret6d < 0.249 || ret6d > 0.251 {
		t.Errorf("expected retention ~0.25 at 6 days, got %f", ret6d)
	}
	if score6d < 0.199 || score6d > 0.201 {
		t.Errorf("expected decayed score ~0.20 at 6 days, got %f", score6d)
	}

	// Clock rollback: current time is BEFORE t0 (e.g. system clock set back)
	tPast := t0.Add(-24 * time.Hour)
	scoreRollback, retRollback := CalculateDecay(baseScore, &t0, tPast, halfLife)
	if retRollback != 1.0 {
		t.Errorf("clock rollback must not inflate or corrupt retention, got %f", retRollback)
	}
	if scoreRollback != baseScore {
		t.Errorf("clock rollback must return baseScore %f, got %f", baseScore, scoreRollback)
	}
}

func TestTransferRequiresDistinctGroupsAndDelayedRetrieval(t *testing.T) {
	ledger := NewEvidenceLedger(3.0)
	t0 := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)

	// Scenario A: 3 successes in the SAME setting group -> CANNOT graduate (repeated variants cannot graduate)
	sameGroupExposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_coin_1",
			SettingGroup: "coin_toss",
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_a"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0,
		},
		{
			SessionID:    "s2",
			InstanceID:   "inst_2",
			TemplateID:   "tmpl_coin_2",
			SettingGroup: "coin_toss", // Same group!
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_a"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(15 * time.Minute),
		},
		{
			SessionID:    "s3",
			InstanceID:   "inst_3",
			TemplateID:   "tmpl_coin_3",
			SettingGroup: "coin_toss", // Same group!
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_a"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(30 * time.Minute),
		},
	}

	resA := ledger.ProcessExposures(sameGroupExposures, t0.Add(35*time.Minute))
	cmA := resA["concept_a"]
	if len(cmA.SettingGroupsSeen) != 1 {
		t.Errorf("expected 1 setting group, got %d", len(cmA.SettingGroupsSeen))
	}
	if cmA.DelayedTransferAchieved {
		t.Errorf("repeated variants in single group must not achieve transfer")
	}
	if cmA.ScaffoldLevel != ScaffoldFull {
		t.Errorf("without transfer across distinct groups, scaffold must remain Full, got %s", cmA.ScaffoldLevel)
	}

	// Scenario B: Successes in 2 distinct setting groups, but rapid repetition (< 10m) -> CANNOT graduate
	rapidExposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_coin",
			SettingGroup: "coin_toss",
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_b"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0,
		},
		{
			SessionID:    "s2",
			InstanceID:   "inst_2",
			TemplateID:   "tmpl_defect",
			SettingGroup: "inspection_batch", // Distinct group!
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_b"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(3 * time.Minute), // Only 3 minutes! < 10m
		},
	}

	resB := ledger.ProcessExposures(rapidExposures, t0.Add(5*time.Minute))
	cmB := resB["concept_b"]
	if cmB.DelayedTransferAchieved {
		t.Errorf("rapid repetition (<10m) must not count as delayed transfer")
	}
	if cmB.ScaffoldLevel != ScaffoldFull {
		t.Errorf("without delayed interval, scaffold must remain Full, got %s", cmB.ScaffoldLevel)
	}

	// Scenario C: Successes in 2 distinct setting groups with >= 10m delayed retrieval -> GRADUATES to Intermediate!
	transferExposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_coin",
			SettingGroup: "coin_toss",
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_c"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0,
		},
		{
			SessionID:    "s2",
			InstanceID:   "inst_2",
			TemplateID:   "tmpl_defect",
			SettingGroup: "inspection_batch", // Distinct group!
			StageID:      "st_1",
			ConceptIDs:   []string{"concept_c"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(15 * time.Minute), // 15 min > 10 min
		},
	}

	resC := ledger.ProcessExposures(transferExposures, t0.Add(16*time.Minute))
	cmC := resC["concept_c"]
	if !cmC.DelayedTransferAchieved {
		t.Errorf("delayed group-changing retrieval should be achieved")
	}
	if cmC.ScaffoldLevel != ScaffoldIntermediate {
		t.Errorf("with 2 successes and delayed transfer, expected ScaffoldIntermediate, got %s", cmC.ScaffoldLevel)
	}
}

func TestRecentErrorRestoresFullScaffold(t *testing.T) {
	ledger := NewEvidenceLedger(3.0)
	t0 := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)

	// Learner achieved faded status (3 successes across 2 groups with delayed retrieval)
	exposures := []RawExposure{
		{
			SessionID:    "s1",
			InstanceID:   "inst_1",
			TemplateID:   "tmpl_1",
			SettingGroup: "group_1",
			StageID:      "st_1",
			ConceptIDs:   []string{"c1"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0,
		},
		{
			SessionID:    "s2",
			InstanceID:   "inst_2",
			TemplateID:   "tmpl_2",
			SettingGroup: "group_2",
			StageID:      "st_1",
			ConceptIDs:   []string{"c1"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(15 * time.Minute),
		},
		{
			SessionID:    "s3",
			InstanceID:   "inst_3",
			TemplateID:   "tmpl_3",
			SettingGroup: "group_2",
			StageID:      "st_1",
			ConceptIDs:   []string{"c1"},
			AttemptNumber: 1,
			IsCorrect:    true,
			Timestamp:    t0.Add(30 * time.Minute),
		},
	}

	resBefore := ledger.ProcessExposures(exposures, t0.Add(35*time.Minute))
	cmBefore := resBefore["c1"]
	if cmBefore.ScaffoldLevel != ScaffoldFaded {
		t.Fatalf("expected ScaffoldFaded before error, got %s", cmBefore.ScaffoldLevel)
	}

	// Now learner makes an error on instance 4
	exposures = append(exposures, RawExposure{
		SessionID:    "s4",
		InstanceID:   "inst_4",
		TemplateID:   "tmpl_4",
		SettingGroup: "group_1",
		StageID:      "st_1",
		ConceptIDs:   []string{"c1"},
		AttemptNumber: 1,
		IsCorrect:    false, // ERROR!
		Timestamp:    t0.Add(45 * time.Minute),
	})

	resAfter := ledger.ProcessExposures(exposures, t0.Add(50*time.Minute))
	cmAfter := resAfter["c1"]

	// Invariant: An error restores full guidance immediately
	if cmAfter.ScaffoldLevel != ScaffoldFull {
		t.Errorf("error must restore full guidance immediately, got %s", cmAfter.ScaffoldLevel)
	}
	if !cmAfter.RecentError {
		t.Errorf("expected RecentError=true")
	}
}

func TestScaffoldDegradation(t *testing.T) {
	tmpl := &domain.QuestionTemplate{
		ID: "binomial_fair_coin_exactly_two",
		Stages: []domain.StageTemplate{
			{ID: "define_target", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"target_c"}},
			{ID: "select_distribution", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"dist_c"}},
			{ID: "check_conditions", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"cond_c"}},
			{ID: "translate_event", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"event_c"}},
			{ID: "build_expression", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"expr_c"}},
			{ID: "calculate_probability", Kind: domain.StageKindNumeric, EvidenceConceptIDs: []string{"calc_c"}},
			{ID: "interpret_probability", Kind: domain.StageKindChoice, EvidenceConceptIDs: []string{"interp_c"}},
		},
	}

	// Full: all 7 stages
	fullStages := DegradeStages(tmpl, ScaffoldFull)
	if len(fullStages) != 7 {
		t.Errorf("expected 7 stages for Full, got %d", len(fullStages))
	}

	// Intermediate: 4 stages
	interStages := DegradeStages(tmpl, ScaffoldIntermediate)
	if len(interStages) != 4 {
		t.Errorf("expected 4 stages for Intermediate, got %d", len(interStages))
	}
	expectedInterIDs := []string{"check_conditions", "build_expression", "calculate_probability", "interpret_probability"}
	for i, id := range expectedInterIDs {
		if interStages[i].ID != id {
			t.Errorf("intermediate stage %d expected %q, got %q", i, id, interStages[i].ID)
		}
	}

	// Faded: 2 stages (calculation + interpretation)
	fadedStages := DegradeStages(tmpl, ScaffoldFaded)
	if len(fadedStages) != 2 {
		t.Errorf("expected 2 stages for Faded, got %d", len(fadedStages))
	}
	if fadedStages[0].ID != "calculate_probability" || fadedStages[1].ID != "interpret_probability" {
		t.Errorf("faded stages expected calc and interp, got %s and %s", fadedStages[0].ID, fadedStages[1].ID)
	}
	// Evidence concept preservation
	if len(fadedStages[0].EvidenceConceptIDs) != 1 || fadedStages[0].EvidenceConceptIDs[0] != "calc_c" {
		t.Errorf("evidence concepts must be preserved on degraded stage")
	}
}

func TestWeightedSelectionDeterministicSeed(t *testing.T) {
	tmpls := []*domain.QuestionTemplate{
		{ID: "t1", SettingGroup: "g1", ConceptIDs: []string{"c1"}},
		{ID: "t2", SettingGroup: "g2", ConceptIDs: []string{"c2"}},
		{ID: "t3", SettingGroup: "g3", ConceptIDs: []string{"c3"}},
		{ID: "t4", SettingGroup: "g4", ConceptIDs: []string{"c4"}},
		{ID: "t5", SettingGroup: "g5", ConceptIDs: []string{"c5"}},
	}

	optsA := SelectionOptions{
		CandidateTemplates: tmpls,
		QuestionCount:      5,
		Seed:               4242,
		Intensity:          "standard",
	}

	optsB := SelectionOptions{
		CandidateTemplates: tmpls,
		QuestionCount:      5,
		Seed:               4242,
		Intensity:          "standard",
	}

	selA := SelectQuestions(optsA)
	selB := SelectQuestions(optsB)

	if len(selA) != 5 || len(selB) != 5 {
		t.Fatalf("expected 5 selected questions")
	}

	for i := range selA {
		if selA[i].ID != selB[i].ID {
			t.Errorf("seeded selection must be deterministic: at %d got %s vs %s", i, selA[i].ID, selB[i].ID)
		}
	}
}

func TestContrastRegistry(t *testing.T) {
	partner, found := FindContrastPartner("binomial_fair_coin_exactly_two", "exactly_as_at_most")
	if !found {
		t.Fatalf("expected contrast partner for exactly_as_at_most")
	}
	if partner != "binomial_defective_at_most_one" {
		t.Errorf("expected binomial_defective_at_most_one, got %s", partner)
	}

	_, notFound := FindContrastPartner("unknown_origin", "fake_misconception")
	if notFound {
		t.Errorf("expected false for unknown contrast pairing")
	}
}
