package storage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/worksheets"
)

func TestWorksheetRestartAtomicCommandsAndNoMasteryWrites(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "worksheets.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err = RunMigrations(ctx, db, nil); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db, nil)
	b, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, _ := b.Get("binomial_fair_coin_exactly_two")
	r, err := worksheets.New("worksheet_restart", tmpl, 42, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.CreateWorksheet(ctx, r); err != nil {
		t.Fatal(err)
	}
	key := r.Items[0].Key
	draft := worksheets.Command{ID: "save_draft", Revision: 1, Type: "save_draft", Answers: map[string]domain.SubmittedAnswer{key: {Kind: domain.StageKindChoice, OptionID: r.Items[0].State.Instance.ExpectedAnswer.OptionID}}}
	saved, err := store.CommandWorksheet(ctx, r.ID, draft)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 2 || saved.Items[0].State.DraftAnswer == nil {
		t.Fatal("draft not saved")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	store = NewStore(db, nil)
	replay, err := store.CommandWorksheet(ctx, r.ID, draft)
	if err != nil || replay.Revision != 2 {
		t.Fatalf("restart replay failed: %+v %v", replay, err)
	}
	draft.Answers[key] = domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "forged"}
	if _, err = store.CommandWorksheet(ctx, r.ID, draft); err != worksheets.ErrConflict {
		t.Fatal("changed replay accepted")
	}
	draft.ID = "stale"
	if _, err = store.CommandWorksheet(ctx, r.ID, draft); err != worksheets.ErrConflict {
		t.Fatal("stale revision accepted")
	}
	// Inject failure after UPDATE but before command insertion: no state can commit.
	if _, err = db.Exec(`CREATE TRIGGER fail_worksheet_command BEFORE INSERT ON worksheet_commands BEGIN SELECT RAISE(ABORT,'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	cmd := worksheets.Command{ID: "atomic_submit", Revision: 2, Type: "submit", Answers: map[string]domain.SubmittedAnswer{}}
	for _, item := range r.Items {
		expected := item.State.Instance.ExpectedAnswer
		a := domain.SubmittedAnswer{Kind: item.State.Instance.Kind, OptionID: expected.OptionID}
		if expected.Value != nil {
			a.NumericRaw = "3/8"
		}
		cmd.Answers[item.Key] = a
	}
	if _, err = store.CommandWorksheet(ctx, r.ID, cmd); err == nil {
		t.Fatal("injected failure did not abort")
	}
	latest, err := store.GetWorksheet(ctx, r.ID)
	if err != nil || latest.Revision != 2 || len(latest.Items[0].State.Attempts) != 0 {
		t.Fatalf("partial attempt committed: %+v %v", latest, err)
	}
	if _, err = db.Exec(`DROP TRIGGER fail_worksheet_command`); err != nil {
		t.Fatal(err)
	}
	completed, err := store.CommandWorksheet(ctx, r.ID, cmd)
	if err != nil || completed.Status != "completed" {
		t.Fatalf("%v %v", completed, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	store = NewStore(db, nil)
	again, err := store.CommandWorksheet(ctx, r.ID, cmd)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(completed)
	got, _ := json.Marshal(again)
	if string(want) != string(got) {
		t.Fatal("submission replay changed snapshot or attempts")
	}
	if _, err = store.CommandWorksheet(ctx, r.ID, worksheets.Command{ID: "edit_completed", Revision: 3, Type: "save_draft"}); err == nil {
		t.Fatal("completed worksheet edited")
	}
	list, err := store.ListWorksheets(ctx)
	if err != nil || len(list) != 1 || list[0].Status != "completed" {
		t.Fatal("review missing on restart")
	}
	for _, table := range []string{"attempts", "mastery_projections", "candidate_questions"} {
		var count int
		if err = db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("worksheet wrote to %s", table)
		}
	}
}

func TestWorksheetUpgradeKeepsStageOneDatabaseAndBackup(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "upgrade.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	all, err := LoadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if err = EnsureSchemaMigrationsTable(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(all[0].SQL); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO schema_migrations(version,name,checksum,applied_at) VALUES(?,?,?,?)`, all[0].Version, all[0].Name, all[0].Checksum, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db, nil)
	if err = store.SaveSettings(ctx, "prior_settings", `{"question_count":7}`); err != nil {
		t.Fatal(err)
	}
	backupPath := filepath.Join(t.TempDir(), "before-worksheets.db")
	calls := 0
	if err = RunMigrations(ctx, db, func() error { calls++; return Backup(ctx, db, backupPath) }); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("missing pre-upgrade backup")
	}
	if value, err := store.GetSettings(ctx, "prior_settings"); err != nil || value != `{"question_count":7}` {
		t.Fatal("existing settings changed")
	}
	backup, err := Open(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Close()
	var count int
	if err = backup.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil || count != 1 {
		t.Fatal("backup did not retain original schema")
	}
}
