// content-audit derives rule-v2 fixed-reference choices and emits a complete
// Markdown preview. It never grants approval. -derive only updates draft files.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func run() error {
	derive := flag.Bool("derive", false, "fill calculated choice values in unapproved drafts")
	dir := flag.String("dir", "curriculum/drafts", "directory to audit")
	preview := flag.String("preview", "docs/REMAINING-CONTENT-REVIEW.md", "complete review preview path")
	reviewer := flag.String("reviewer", "", "only include templates approved by this reviewer")
	reviews := flag.Bool("reviews", false, "write matching review records for already-approved templates")
	flag.Parse()
	paths, err := filepath.Glob(filepath.Join(*dir, "*.json"))
	if err != nil {
		return err
	}
	var out strings.Builder
	count := 0
	out.WriteString("# Final Wave 4 and Wave 5 content review\n\nThese 20 fixed-reference questions were authored and semantically reviewed by the assistant. Numerical choices are checked against the Go engine at bank admission. Course-specific quartile, pooled-test and dummy-variable conventions remain explicitly unconfirmed. Semantic review does not itself authorize activation. Approval status is recorded in each JSON and review file.\n\n")
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		t, err := bank.StrictDecodeTemplate(data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if *reviewer != "" && (t.Approval == nil || t.Approval.Reviewer != *reviewer) {
			continue
		}
		count++
		var quantities map[string]float64
		if t.RuleVersion == 2 {
			quantities, err = mathengine.CurriculumCalculations(t.FamilyID, t.Parameters)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
		if *derive {
			if t.Status != domain.StatusDraft || t.Approval != nil {
				return fmt.Errorf("refusing to change approved content: %s", path)
			}
			for i := range t.Stages {
				s := &t.Stages[i]
				if s.Calculation == "" {
					continue
				}
				v, ok := quantities[s.Calculation]
				if !ok {
					return fmt.Errorf("unknown quantity %s", s.Calculation)
				}
				for j := range s.Options {
					o := &s.Options[j]
					if o.ID == s.ExpectedAnswer.OptionID {
						o.NumericValue = &v
						o.TextMarkdown = fmt.Sprintf("$%.4f$", v)
					}
				}
			}
			for k, v := range quantities {
				t.ScenarioMarkdown = strings.ReplaceAll(t.ScenarioMarkdown, "{{"+k+"}}", fmt.Sprintf("%.4f", v))
			}
			data, err = json.MarshalIndent(t, "", "  ")
			if err != nil {
				return err
			}
			if err = os.WriteFile(path, append(data, '\n'), 0644); err != nil {
				return err
			}
		}
		if err = bank.ValidateTemplate(t, nil); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if *reviews {
			if t.Status != domain.StatusActive || t.Approval == nil {
				return fmt.Errorf("review export requires already approved template %s", t.ID)
			}
			record := map[string]interface{}{
				"schema_version": 1, "template_id": t.ID, "template_version": t.Version, "review_status": "approved",
				"reviewer": t.Approval.Reviewer, "reviewed_at": t.Approval.ReviewedAt, "delegation_basis": t.Approval.DelegationBasis,
				"mathematical_audit": map[string]interface{}{"family_id": t.FamilyID, "rule_version": t.RuleVersion, "parameters": t.Parameters, "engine_quantities": quantities, "four_decimal_choice_collision_check": "passed for calculation-bound stages"},
				"pedagogical_audit":  map[string]interface{}{"stage_count": len(t.Stages), "concept_count": len(t.ConceptIDs), "one_to_one_evidence_mapping": true, "full_preview": "docs/REMAINING-CONTENT-REVIEW.md", "sources": t.SourceRefs, "assumptions": t.Assumptions, "review_scope": "Complete scenario, options, diagnostic hints and explanations reviewed by assistant under explicit user delegation; no user personal reading claimed."},
				"notes":              t.Approval.Notes,
			}
			b, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return err
			}
			reviewPath := filepath.Join("curriculum", "reviews", filepath.Base(strings.TrimSuffix(path, ".json"))+".review.json")
			// A Windows checkout may use CRLF; that is not a changed review.
			if existing, err := os.ReadFile(reviewPath); err == nil && !bytes.Equal(bytes.ReplaceAll(existing, []byte("\r\n"), []byte("\n")), append(b, '\n')) {
				return fmt.Errorf("refusing to overwrite different review %s", reviewPath)
			}
			if err = os.WriteFile(reviewPath, append(b, '\n'), 0644); err != nil {
				return err
			}
		}
		fmt.Fprintf(&out, "## %s\n\nID: `%s`; family `%s`, rule %d; status **%s**.\n\n%s\n\nAssumptions:\n\n", t.Title, t.ID, t.FamilyID, t.RuleVersion, t.Status, t.ScenarioMarkdown)
		for _, a := range t.Assumptions {
			fmt.Fprintf(&out, "- %s\n", a)
		}
		for i, s := range t.Stages {
			fmt.Fprintf(&out, "\n### %d. %s\n\n%s\n\n", i+1, s.ID, s.PromptMarkdown)
			for j, o := range s.Options {
				mark := ""
				if o.ID == s.ExpectedAnswer.OptionID {
					mark = " **[correct]**"
				}
				fmt.Fprintf(&out, "%d. %s%s\n", j+1, o.TextMarkdown, mark)
				if o.HintMarkdown != nil {
					fmt.Fprintf(&out, "   Hint: %s\n", *o.HintMarkdown)
				}
			}
			fmt.Fprintf(&out, "\nExplanation: %s\n", s.ExplanationMarkdown)
			if s.Calculation != "" {
				fmt.Fprintf(&out, "\nEngine quantity `%s`: %.12g. Four-decimal choices; exact Go value retained for audit.\n", s.Calculation, quantities[s.Calculation])
			}
		}
		fmt.Printf("validated %s (%d stages)\n", t.ID, len(t.Stages))
	}
	if count == 0 {
		return fmt.Errorf("no matching templates; preview unchanged")
	}
	if err = os.WriteFile(*preview, []byte(out.String()), 0644); err != nil {
		return err
	}
	fmt.Printf("Validated %d templates; preview: %s\n", count, *preview)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
