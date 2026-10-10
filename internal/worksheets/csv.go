// Package worksheets provides full-form practice with immutable snapshots.
// It keeps supported worksheet evidence separate from independent mastery.
package worksheets

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

const MaxCSVBytes = 64 * 1024
const MaxRows = 500

var safeID = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_-]{0,79}$`)

type Row struct {
	ExperimentID string `json:"experiment_id"`
	Successes    int    `json:"successes"`
}
type Dataset struct {
	Approval    *domain.ApprovalRecord `json:"approval,omitempty"`
	Rows        []Row                  `json:"rows"`
	SourceNote  string                 `json:"source_note,omitempty"`
	Reviewer    string                 `json:"reviewer,omitempty"`
	RuleVersion int                    `json:"rule_version"`
}

// ParseCSV accepts data only, never paths, formulas or an inferred schema.
func ParseCSV(raw string) (*Dataset, error) {
	if len(raw) > MaxCSVBytes || !utf8.ValidString(raw) {
		return nil, errors.New("CSV must be UTF-8 and at most 64 KiB")
	}
	r := csv.NewReader(strings.NewReader(raw))
	r.FieldsPerRecord = 2
	header, err := r.Read()
	if err != nil || len(header) != 2 || header[0] != "experiment_id" || header[1] != "successes" {
		return nil, errors.New("CSV header must be exactly experiment_id,successes")
	}
	d := &Dataset{Rows: []Row{}, RuleVersion: 1}
	seen := map[string]bool{}
	for {
		fields, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV row %d: %w", len(d.Rows)+2, err)
		}
		if len(d.Rows) >= MaxRows {
			return nil, errors.New("CSV allows at most 500 experiments")
		}
		if !safeID.MatchString(fields[0]) || seen[fields[0]] {
			return nil, errors.New("experiment IDs must be unique, 1–80 ASCII letters/digits/underscores/hyphens, starting with a letter/digit/underscore")
		}
		// Single ASCII digit excludes spreadsheet formulas, whitespace and alternate encodings.
		if len(fields[1]) != 1 || fields[1][0] < '0' || fields[1][0] > '4' {
			return nil, errors.New("successes must be an integer from 0 to 4")
		}
		n, _ := strconv.Atoi(fields[1])
		d.Rows = append(d.Rows, Row{fields[0], n})
		seen[fields[0]] = true
	}
	if len(d.Rows) == 0 {
		return nil, errors.New("CSV needs at least one experiment")
	}
	return d, nil
}

func (d *Dataset) Summary() (int, float64, error) {
	counts := make([]int, len(d.Rows))
	for i, row := range d.Rows {
		counts[i] = row.Successes
	}
	return mathengine.EmpiricalFrequency(counts, 4, 2)
}
