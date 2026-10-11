package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/exam"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

// ExamStore provides persistence for assessment sessions and history.
type ExamStore interface {
	CreateExam(context.Context, *exam.ExamRecord) error
	GetExam(context.Context, string) (*exam.ExamRecord, error)
	ListExams(context.Context) ([]exam.ExamSummary, error)
	CommandExam(context.Context, string, exam.ExamCommand) (*exam.ExamRecord, error)
	GetExamReport(context.Context, string) (*exam.ExamReport, error)
	GetHistory(context.Context, int, int) ([]storage.HistoryItem, error)
}

func decodeExamBody(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected single JSON document")
	}
	return nil
}

func (s *Server) handleExams(w http.ResponseWriter, r *http.Request) {
	if !s.requireAppSession(w, r) {
		return
	}
	if s.examStore == nil {
		http.Error(w, "exam storage unavailable", http.StatusServiceUnavailable)
		return
	}

	if r.URL.Path == "/api/exams" {
		switch r.Method {
		case http.MethodGet:
			list, err := s.examStore.ListExams(r.Context())
			if err != nil {
				http.Error(w, "failed to load exams", http.StatusInternalServerError)
				return
			}
			if list == nil {
				list = []exam.ExamSummary{}
			}
			candidateJSON(w, list)
		case http.MethodPost:
			s.createExam(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/exams/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	examID := parts[0]

	// POST /api/exams/{id}/commands
	if len(parts) == 2 && parts[1] == "commands" && r.Method == http.MethodPost {
		var cmd exam.ExamCommand
		if err := decodeExamBody(w, r, &cmd); err != nil {
			http.Error(w, fmt.Sprintf("invalid command body: %v", err), http.StatusBadRequest)
			return
		}

		next, err := s.examStore.CommandExam(r.Context(), examID, cmd)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			if errors.Is(err, exam.ErrConflict) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			if errors.Is(err, exam.ErrInvalidInput) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, fmt.Sprintf("failed to execute exam command: %v", err), http.StatusInternalServerError)
			return
		}

		candidateJSON(w, exam.Public(next, time.Now()))
		return
	}

	// GET /api/exams/{id}/report
	if len(parts) == 2 && parts[1] == "report" && r.Method == http.MethodGet {
		rec, err := s.examStore.GetExam(r.Context(), examID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rec.Status == exam.StatusActive {
			http.Error(w, "assessment is currently active; report and answers are withheld", http.StatusForbidden)
			return
		}

		// Check if export markdown was requested
		if r.URL.Query().Get("format") == "markdown" {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="assessment-report-%s.md"`, examID))
			_, _ = io.WriteString(w, exam.FormatMarkdownReport(rec))
			return
		}

		rep, err := s.examStore.GetExamReport(r.Context(), examID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		candidateJSON(w, rep)
		return
	}

	// GET /api/exams/{id}
	if len(parts) == 1 && r.Method == http.MethodGet {
		rec, err := s.examStore.GetExam(r.Context(), examID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		candidateJSON(w, exam.Public(rec, time.Now()))
		return
	}

	http.NotFound(w, r)
}

func (s *Server) createExam(w http.ResponseWriter, r *http.Request) {
	var cfg exam.ExamConfig
	if err := decodeExamBody(w, r, &cfg); err != nil {
		http.Error(w, fmt.Sprintf("invalid exam configuration: %v", err), http.StatusBadRequest)
		return
	}

	if s.bank == nil || s.bank.Count() == 0 {
		http.Error(w, "no approved questions available in curriculum bank", http.StatusServiceUnavailable)
		return
	}

	allTmpls := s.bank.List()
	id := storage.GenerateExamID()
	rec, err := exam.New(id, cfg, allTmpls, time.Now())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to initialize exam: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.examStore.CreateExam(r.Context(), rec); err != nil {
		http.Error(w, fmt.Sprintf("failed to persist exam: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	candidateJSON(w, exam.Public(rec, time.Now()))
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if !s.requireAppSession(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 20
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if s.examStore == nil {
		candidateJSON(w, []storage.HistoryItem{})
		return
	}

	items, err := s.examStore.GetHistory(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch history: %v", err), http.StatusInternalServerError)
		return
	}

	candidateJSON(w, items)
}
