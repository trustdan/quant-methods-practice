package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mastery"
)

// GetHistoricalExposures queries all attempts, assistance events, and immutable instance snapshots
// from SQLite, returning a chronological stream of RawExposure records.
func (s *Store) GetHistoricalExposures(ctx context.Context) ([]mastery.RawExposure, error) {
	const query = `
SELECT 
    a.session_id,
    a.instance_id,
    qi.template_id,
    qi.stages_json,
    qi.title,
    a.stage_id,
    a.attempt_number,
    a.is_correct,
    a.assistance_json,
    a.created_at
FROM attempts a
JOIN question_instances qi ON a.instance_id = qi.id
ORDER BY a.created_at ASC;`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical attempts: %w", err)
	}
	defer rows.Close()

	// Cache parsed stage concepts by instance_id:stage_id -> concept_ids
	stageConceptsCache := make(map[string][]string)
	// Cache setting groups by instance_id
	instanceSettingGroupCache := make(map[string]string)

	var exposures []mastery.RawExposure

	for rows.Next() {
		var (
			sessionID      string
			instanceID     string
			templateID     string
			stagesJSON     string
			title          string
			stageID        string
			attemptNum     int
			isCorrectInt   int
			assistanceJSON string
			createdAtStr   string
		)

		if err := rows.Scan(
			&sessionID,
			&instanceID,
			&templateID,
			&stagesJSON,
			&title,
			&stageID,
			&attemptNum,
			&isCorrectInt,
			&assistanceJSON,
			&createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attempt row: %w", err)
		}

		// Parse stages if not yet cached for this instance
		cacheKey := instanceID + ":" + stageID
		concepts, ok := stageConceptsCache[cacheKey]
		if !ok {
			var stages []domain.StageInstance
			if err := json.Unmarshal([]byte(stagesJSON), &stages); err == nil {
				for _, st := range stages {
					stageConceptsCache[instanceID+":"+st.ID] = st.EvidenceConceptIDs
				}
			}
			concepts = stageConceptsCache[cacheKey]
		}

		// Retrieve setting group
		settingGroup := instanceSettingGroupCache[instanceID]
		if settingGroup == "" {
			// Extract setting group if stored in parameters or stages
			type probeInstance struct {
				SettingGroup string `json:"setting_group"`
			}
			var pi probeInstance
			_ = json.Unmarshal([]byte(stagesJSON), &pi)
			settingGroup = pi.SettingGroup
			instanceSettingGroupCache[instanceID] = settingGroup
		}

		var assistance []domain.AssistanceType
		if assistanceJSON != "" && assistanceJSON != "null" {
			_ = json.Unmarshal([]byte(assistanceJSON), &assistance)
		}

		t, parseErr := time.Parse(time.RFC3339Nano, createdAtStr)
		if parseErr != nil {
			t, _ = time.Parse(time.RFC3339, createdAtStr)
		}

		isContrast := strings.Contains(title, "[Contrast]")
		for _, a := range assistance {
			if a == domain.AssistanceGuidedContrast {
				isContrast = true
				break
			}
		}

		exposures = append(exposures, mastery.RawExposure{
			SessionID:     sessionID,
			InstanceID:    instanceID,
			TemplateID:    templateID,
			SettingGroup:  settingGroup,
			StageID:       stageID,
			ConceptIDs:    concepts,
			AttemptNumber: attemptNum,
			IsCorrect:     isCorrectInt == 1,
			Assistance:    assistance,
			Timestamp:     t,
			IsContrast:    isContrast,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during attempt rows iteration: %w", err)
	}

	return exposures, nil
}

// SaveMasteryProjection persists an evaluated concept projection to SQLite.
func (s *Store) SaveMasteryProjection(ctx context.Context, proj *mastery.ConceptMastery) error {
	if proj == nil {
		return nil
	}

	projJSON, err := json.Marshal(proj)
	if err != nil {
		return fmt.Errorf("failed to marshal mastery projection json: %w", err)
	}

	var lastTestedStr *string
	if proj.LastTestedAt != nil {
		str := proj.LastTestedAt.UTC().Format(time.RFC3339Nano)
		lastTestedStr = &str
	}
	updatedAtStr := time.Now().UTC().Format(time.RFC3339Nano)

	const query = `
INSERT INTO mastery_projections (
    concept_id, policy_version, evidence_count, independent_count, assisted_count,
    last_tested_at, projection_json, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(concept_id) DO UPDATE SET
    policy_version = excluded.policy_version,
    evidence_count = excluded.evidence_count,
    independent_count = excluded.independent_count,
    assisted_count = excluded.assisted_count,
    last_tested_at = excluded.last_tested_at,
    projection_json = excluded.projection_json,
    updated_at = excluded.updated_at;`

	_, err = s.db.ExecContext(ctx, query,
		proj.ConceptID,
		proj.PolicyVersion,
		proj.TotalEvidenceCount,
		proj.IndependentSuccesses+proj.IndependentErrors,
		proj.AssistedCount,
		lastTestedStr,
		string(projJSON),
		updatedAtStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert mastery projection for %q: %w", proj.ConceptID, err)
	}

	return nil
}

// GetAllMasteryProjections retrieves all stored mastery projections.
func (s *Store) GetAllMasteryProjections(ctx context.Context) (map[string]*mastery.ConceptMastery, error) {
	const query = `SELECT projection_json FROM mastery_projections;`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query mastery projections: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*mastery.ConceptMastery)
	for rows.Next() {
		var jsonBytes string
		if err := rows.Scan(&jsonBytes); err != nil {
			return nil, fmt.Errorf("failed to scan projection row: %w", err)
		}
		var cm mastery.ConceptMastery
		if err := json.Unmarshal([]byte(jsonBytes), &cm); err == nil {
			result[cm.ConceptID] = &cm
		}
	}

	return result, nil
}

// RebuildMastery recalculates and stores concept mastery projections from all historical exposures.
func (s *Store) RebuildMastery(ctx context.Context, ledger *mastery.EvidenceLedger, now time.Time) (map[string]*mastery.ConceptMastery, error) {
	if ledger == nil {
		ledger = mastery.NewEvidenceLedger(mastery.DefaultHalfLifeDays)
	}

	exposures, err := s.GetHistoricalExposures(ctx)
	if err != nil {
		return nil, fmt.Errorf("rebuild mastery failed getting exposures: %w", err)
	}

	projections := ledger.ProcessExposures(exposures, now)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const upsertQuery = `
INSERT INTO mastery_projections (
    concept_id, policy_version, evidence_count, independent_count, assisted_count,
    last_tested_at, projection_json, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(concept_id) DO UPDATE SET
    policy_version = excluded.policy_version,
    evidence_count = excluded.evidence_count,
    independent_count = excluded.independent_count,
    assisted_count = excluded.assisted_count,
    last_tested_at = excluded.last_tested_at,
    projection_json = excluded.projection_json,
    updated_at = excluded.updated_at;`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert query: %w", err)
	}
	defer stmt.Close()

	updatedAtStr := now.UTC().Format(time.RFC3339Nano)

	for _, cm := range projections {
		projJSON, err := json.Marshal(cm)
		if err != nil {
			continue
		}

		var lastTestedStr *string
		if cm.LastTestedAt != nil {
			str := cm.LastTestedAt.UTC().Format(time.RFC3339Nano)
			lastTestedStr = &str
		}

		_, err = stmt.ExecContext(ctx,
			cm.ConceptID,
			cm.PolicyVersion,
			cm.TotalEvidenceCount,
			cm.IndependentSuccesses+cm.IndependentErrors,
			cm.AssistedCount,
			lastTestedStr,
			string(projJSON),
			updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to persist rebuilt projection %q: %w", cm.ConceptID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit rebuilt projections: %w", err)
	}

	return projections, nil
}

// GetMasterySummary calculates the complete curriculum mastery summary with read-time decay.
func (s *Store) GetMasterySummary(ctx context.Context, ledger *mastery.EvidenceLedger, now time.Time) (*mastery.MasterySummary, error) {
	projections, err := s.RebuildMastery(ctx, ledger, now)
	if err != nil {
		return nil, err
	}

	concepts := make([]mastery.ConceptMastery, 0, len(projections))
	for _, cm := range projections {
		concepts = append(concepts, *cm)
	}

	sort.SliceStable(concepts, func(i, j int) bool {
		return concepts[i].ConceptID < concepts[j].ConceptID
	})

	totalMastered := 0
	totalTransferring := 0
	totalLearning := 0
	totalNew := 0
	scoreSum := 0.0

	for _, c := range concepts {
		scoreSum += c.DecayedScore
		switch c.Status {
		case mastery.StatusMastered:
			totalMastered++
		case mastery.StatusTransferring:
			totalTransferring++
		case mastery.StatusLearning:
			totalLearning++
		case mastery.StatusNew:
			totalNew++
		}
	}

	overallScore := 0.0
	if len(concepts) > 0 {
		overallScore = scoreSum / float64(len(concepts))
	}

	return &mastery.MasterySummary{
		PolicyVersion:     mastery.EvidencePolicyVersion,
		OverallScore:      overallScore,
		TotalMastered:     totalMastered,
		TotalTransferring: totalTransferring,
		TotalLearning:     totalLearning,
		TotalNew:          totalNew,
		Concepts:          concepts,
		GeneratedAt:       now,
	}, nil
}
