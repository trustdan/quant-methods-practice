package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
)

type CandidateStore interface {
	CreateCandidate(context.Context, *candidates.Record) error
	ListCandidates(context.Context) ([]*candidates.Record, error)
	GetCandidate(context.Context, string) (*candidates.Record, error)
	ReviewCandidate(context.Context, *candidates.Record, int) error
}

// candidateMu serializes in-memory bank changes with durable reviews.
// Generation does not hold it, so a slow provider cannot block practice or review.
type candidateState struct {
	sync.Mutex
	generating bool
}

func decodeCandidateBody(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16000))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected one JSON document")
	}
	return nil
}
func candidateJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleCandidates(w http.ResponseWriter, r *http.Request) {
	if !s.requireAppSession(w, r) {
		return
	}

	if s.candidateStore == nil {
		http.Error(w, "candidate storage unavailable", http.StatusServiceUnavailable)
		return
	}
	if r.URL.Path == "/api/candidates/export" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		templates := s.bank.List()
		w.Header().Set("Content-Disposition", `attachment; filename="approved-bank.json"`)
		candidateJSON(w, templates)
		return
	}
	if r.URL.Path == "/api/candidates" {
		switch r.Method {
		case http.MethodGet:
			records, err := s.candidateStore.ListCandidates(r.Context())
			if err != nil {
				http.Error(w, "cannot load candidates", 500)
				return
			}
			candidateJSON(w, records)
		case http.MethodPost:
			s.createCandidate(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/candidates/"), "/")
	if len(parts) != 2 || parts[1] != "review" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var dto struct {
		Revision          int    `json:"expected_revision"`
		Action            string `json:"action"`
		Reviewer          string `json:"reviewer"`
		Notes             string `json:"notes"`
		SemanticConfirmed bool   `json:"semantic_confirmed"`
	}
	if err := decodeCandidateBody(w, r, &dto); err != nil {
		http.Error(w, "invalid review JSON", 400)
		return
	}
	s.candidateState.Lock()
	defer s.candidateState.Unlock()
	rec, err := s.candidateStore.GetCandidate(r.Context(), parts[0])
	if err != nil {
		http.Error(w, "candidate not found", 404)
		return
	}
	next, err := candidates.PrepareReview(rec, dto.Revision, candidates.Review{Action: dto.Action, Reviewer: dto.Reviewer, Notes: dto.Notes, SemanticConfirmed: dto.SemanticConfirmed}, time.Now())
	if err != nil {
		code := 400
		if errors.Is(err, candidates.ErrConflict) {
			code = 409
		}
		http.Error(w, err.Error(), code)
		return
	}
	if err = s.candidateStore.ReviewCandidate(r.Context(), next, dto.Revision); err != nil {
		code := 500
		if errors.Is(err, candidates.ErrConflict) {
			code = 409
		}
		http.Error(w, "review could not be saved; reload before retrying", code)
		return
	}
	if next.Status == "approved" {
		s.bank.Add(next.Template)
	} else if next.Status == "retired" {
		s.bank.Remove(next.ID)
	}
	candidateJSON(w, next)
}

func (s *Server) createCandidate(w http.ResponseWriter, r *http.Request) {
	var dto struct {
		Mode     string               `json:"mode"`
		Seed     int64                `json:"seed"`
		Proposal *candidates.Proposal `json:"proposal,omitempty"`
	}
	if err := decodeCandidateBody(w, r, &dto); err != nil {
		http.Error(w, "invalid candidate JSON (unknown fields and multiple documents are rejected)", 400)
		return
	}
	if dto.Mode != "local" && dto.Mode != "manual" && dto.Mode != "ai" {
		http.Error(w, "choose local, manual, or ai generation", 400)
		return
	}
	if dto.Mode != "manual" && dto.Proposal != nil {
		http.Error(w, "proposal is only accepted for manual generation", 400)
		return
	}
	p := candidates.Local(dto.Seed)
	source := candidates.Source{Mode: dto.Mode, Route: "offline", Model: "local-binomial-v1", Seed: dto.Seed}
	if dto.Mode == "manual" {
		if dto.Proposal == nil {
			http.Error(w, "manual generation requires a proposal", 400)
			return
		}
		p = *dto.Proposal
		source.Model = "manual"
	}
	if dto.Mode == "ai" {
		s.candidateState.Lock()
		if s.candidateState.generating {
			s.candidateState.Unlock()
			http.Error(w, "candidate generation already running", 409)
			return
		}
		s.candidateState.generating = true
		s.candidateState.Unlock()
		defer func() { s.candidateState.Lock(); s.candidateState.generating = false; s.candidateState.Unlock() }()
		source.Route = s.providerManager.GetActiveRoute()
		source.Model = s.providerManager.GetActiveModel()
		if source.Route == "offline" {
			http.Error(w, "select a connected AI provider and model in Settings first, or use local variation", 400)
			return
		}
		svc, err := s.providerManager.GetActiveProvider()
		if err != nil {
			http.Error(w, "selected provider unavailable", 400)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		p, err = generateCandidateWording(ctx, svc, p)
		if err != nil {
			http.Error(w, "AI candidate was not saved: "+err.Error(), 422)
			return
		}
	}
	id := "candidate_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	rec, err := candidates.NewRecord(id, p, source, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err = s.candidateStore.CreateCandidate(r.Context(), rec); err != nil {
		http.Error(w, "candidate could not be saved", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	candidateJSON(w, rec)
}
