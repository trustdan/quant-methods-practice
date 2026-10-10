package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/candidates"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

func TestCandidateAuditRestartAndSnapshotIsolation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "candidates.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = RunMigrations(ctx, db, nil); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db, nil)
	rec, err := candidates.NewRecord("candidate_original", candidates.Local(0), candidates.Source{Mode: "local"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.CreateCandidate(ctx, rec); err != nil {
		t.Fatal(err)
	}
	review := candidates.Review{Action: "approve", Reviewer: "test human", Notes: "all conditions checked", SemanticConfirmed: true}
	approved, err := candidates.PrepareReview(rec, 1, review, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.ReviewCandidate(ctx, approved, 1); err != nil {
		t.Fatal(err)
	}
	if err = store.ReviewCandidate(ctx, approved, 1); err != candidates.ErrConflict {
		t.Fatal("duplicate review accepted")
	}
	sm := drill.NewSessionManagerWithStore(store, nil)
	sess, err := sm.CreateSession(approved.Template, 42)
	if err != nil {
		t.Fatal(err)
	}
	// A wording edit is a separate unapproved candidate; old instance remains untouched.
	p := rec.Proposal
	p.Scenario = "Edited candidate scenario"
	edited, err := candidates.NewRecord("candidate_edited", p, rec.Source, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.CreateCandidate(ctx, edited); err != nil {
		t.Fatal(err)
	}
	review.Action = "retire"
	retired, err := candidates.PrepareReview(approved, 2, review, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.ReviewCandidate(ctx, retired, 2); err != nil {
		t.Fatal(err)
	}
	var audits, attempts, projections int
	_ = db.QueryRow("SELECT count(*) FROM content_approval_events").Scan(&audits)
	_ = db.QueryRow("SELECT count(*) FROM attempts").Scan(&attempts)
	_ = db.QueryRow("SELECT count(*) FROM mastery_projections").Scan(&projections)
	if audits != 2 || attempts != 0 || projections != 0 {
		t.Fatalf("audit/learning counts: %d %d %d", audits, attempts, projections)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store = NewStore(db, nil)
	saved, err := store.GetCandidate(ctx, rec.ID)
	if err != nil || saved.Status != "retired" || len(saved.Reviews) != 2 {
		t.Fatalf("review restart: %+v %v", saved, err)
	}
	restored, err := store.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.QuestionInstance.ScenarioMarkdown != rec.Proposal.Scenario || *restored.Stages[5].Instance.ExpectedAnswer.Value != .375 {
		t.Fatal("retirement/edit changed saved snapshot")
	}
}
