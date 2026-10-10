package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

type candidateFlags struct {
	generate, list                                       bool
	preview, approve, reject, retire, export, importFile string
	reviewer, notes                                      string
	revision                                             int
	semantic                                             bool
}

func (f candidateFlags) requested() bool {
	return f.generate || f.list || f.preview != "" || f.approve != "" || f.reject != "" || f.retire != "" || f.export != "" || f.importFile != ""
}
func runCandidateCLI(ctx context.Context, store *storage.Store, f candidateFlags, seed int64) error {
	count := 0
	for _, b := range []bool{f.generate, f.list, f.preview != "", f.approve != "", f.reject != "", f.retire != "", f.export != "", f.importFile != ""} {
		if b {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("choose one candidate command at a time")
	}
	printJSON := func(v any) error { e := json.NewEncoder(os.Stdout); e.SetIndent("", "  "); return e.Encode(v) }
	if f.generate || f.importFile != "" {
		p := candidates.Local(seed)
		source := candidates.Source{Mode: "local", Route: "offline", Model: "local-binomial-v1", Seed: seed}
		if f.importFile != "" {
			info, err := os.Stat(f.importFile)
			if err != nil {
				return err
			}
			if info.Size() > 16000 {
				return fmt.Errorf("proposal exceeds 16000 bytes")
			}
			data, err := os.ReadFile(f.importFile)
			if err != nil {
				return err
			}
			p, err = candidates.Decode(data)
			if err != nil {
				return err
			}
			source.Mode = "manual"
			source.Model = "manual"
		}
		id := "candidate_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		rec, err := candidates.NewRecord(id, p, source, time.Now())
		if err != nil {
			return err
		}
		if err = store.CreateCandidate(ctx, rec); err != nil {
			return err
		}
		return printJSON(rec)
	}
	if f.list {
		records, err := store.ListCandidates(ctx)
		if err != nil {
			return err
		}
		return printJSON(records)
	}
	if f.preview != "" {
		rec, err := store.GetCandidate(ctx, f.preview)
		if err != nil {
			return err
		}
		return printJSON(rec)
	}
	if f.export != "" {
		var b *bank.Bank
		for _, dir := range []string{"curriculum/approved", "../curriculum/approved", "../../curriculum/approved"} {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				b, err = bank.LoadActiveBank(dir, nil)
				if err != nil {
					return err
				}
				break
			}
		}
		if b == nil {
			return fmt.Errorf("approved curriculum directory unavailable")
		}
		records, err := store.ListCandidates(ctx)
		if err != nil {
			return err
		}
		for _, rec := range records {
			if rec.Status == "approved" {
				b.Add(rec.Template)
			}
		}
		// Exclusive creation protects existing exports from accidental replacement.
		out, err := os.OpenFile(f.export, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		err = enc.Encode(b.List())
		closeErr := out.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	id, action := f.approve, "approve"
	if f.reject != "" {
		id, action = f.reject, "reject"
	}
	if f.retire != "" {
		id, action = f.retire, "retire"
	}
	rec, err := store.GetCandidate(ctx, id)
	if err != nil {
		return err
	}
	next, err := candidates.PrepareReview(rec, f.revision, candidates.Review{Action: action, Reviewer: f.reviewer, Notes: f.notes, SemanticConfirmed: f.semantic}, time.Now())
	if err != nil {
		return err
	}
	if next.Status == "approved" && next.Template.Status != domain.StatusActive {
		return fmt.Errorf("approval did not produce active content")
	}
	if err = store.ReviewCandidate(ctx, next, f.revision); err != nil {
		return err
	}
	return printJSON(next)
}
