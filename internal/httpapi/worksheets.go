package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/worksheets"
)

type WorksheetStore interface {
	CreateWorksheet(context.Context, *worksheets.Record) error
	GetWorksheet(context.Context, string) (*worksheets.Record, error)
	ListWorksheets(context.Context) ([]worksheets.View, error)
	CommandWorksheet(context.Context, string, worksheets.Command) (*worksheets.Record, error)
}

// requireAppSession authenticates the browser and pins mutations to the exact
// app origin, including its port. Host validation also runs in the middleware.
func (s *Server) requireAppSession(w http.ResponseWriter, r *http.Request) bool {
	s.mu.RLock()
	token := s.sessionToken
	s.mu.RUnlock()
	cookie, err := r.Cookie("quant_session")
	if err != nil || token == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
		http.Error(w, "open the app through the launcher to authenticate", 401)
		return false
	}
	if r.Method != http.MethodGet {
		origin := r.Header.Get("Origin")
		allowed := origin == "http://"+r.Host
		for _, dev := range s.config.AllowedDevOrigins {
			if origin == dev {
				allowed = true
			}
		}
		if !allowed {
			http.Error(w, "mutations require the app origin", 403)
			return false
		}
	}
	return true
}
func decodeWorksheetBody(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected one JSON document")
	}
	return nil
}
func (s *Server) handleWorksheets(w http.ResponseWriter, r *http.Request) {
	if !s.requireAppSession(w, r) {
		return
	}
	if s.worksheetStore == nil {
		http.Error(w, "worksheet storage unavailable", 503)
		return
	}
	if r.URL.Path == "/api/worksheets/dataset-preview" {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var dto struct {
			CSV string `json:"csv"`
		}
		if err := decodeWorksheetBody(w, r, &dto); err != nil {
			http.Error(w, "invalid preview JSON", 400)
			return
		}
		dataset, err := worksheets.ParseCSV(dto.CSV)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		tmpl, ok := s.bank.Get("binomial_fair_coin_exactly_two")
		if !ok {
			http.Error(w, "reviewed model reference unavailable", 503)
			return
		}
		model, err := worksheets.New("dataset_preview", tmpl, 42, nil, time.Now())
		if err != nil {
			http.Error(w, "cannot build model preview", 500)
			return
		}
		observed, err := worksheets.PreviewDatasetQuestion(dataset, 42)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		candidateJSON(w, struct {
			Dataset   *worksheets.Dataset       `json:"dataset"`
			Questions []domain.QuestionInstance `json:"questions"`
		}{dataset, append([]domain.QuestionInstance{observed}, model.Questions...)})
		return
	}
	if r.URL.Path == "/api/worksheets" {
		switch r.Method {
		case "GET":
			list, err := s.worksheetStore.ListWorksheets(r.Context())
			if err != nil {
				http.Error(w, "cannot load worksheets", 500)
				return
			}
			candidateJSON(w, list)
		case "POST":
			s.createWorksheet(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/worksheets/"), "/")
	if len(parts) == 2 && parts[1] == "commands" && r.Method == "POST" {
		var cmd worksheets.Command
		if err := decodeWorksheetBody(w, r, &cmd); err != nil {
			http.Error(w, "invalid command JSON", 400)
			return
		}
		_, err := s.worksheetStore.GetWorksheet(r.Context(), parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		rec, err := s.worksheetStore.CommandWorksheet(r.Context(), parts[0], cmd)
		if err != nil {
			code := 500
			message := "Worksheet could not be saved. Retry the pending command."
			if errors.Is(err, worksheets.ErrInvalidInput) {
				code = 400
				message = err.Error()
			}
			if errors.Is(err, worksheets.ErrConflict) {
				code = 409
				message = err.Error()
			}
			http.Error(w, message, code)
			return
		}
		candidateJSON(w, worksheets.Public(rec))
		return
	}
	if (len(parts) != 1 && !(len(parts) == 2 && parts[1] == "export")) || r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	rec, err := s.worksheetStore.GetWorksheet(r.Context(), parts[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 2 {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="practice-worksheet.md"`)
		_, _ = io.WriteString(w, worksheets.Markdown(rec))
		return
	}
	candidateJSON(w, worksheets.Public(rec))
}
func (s *Server) createWorksheet(w http.ResponseWriter, r *http.Request) {
	var dto struct {
		TemplateID string `json:"template_id"`
		Seed       int64  `json:"seed"`
		Mode       string `json:"mode"`
		CSV        string `json:"csv,omitempty"`
		Reviewer   string `json:"reviewer,omitempty"`
		SourceNote string `json:"source_note,omitempty"`
		Reviewed   bool   `json:"reviewed,omitempty"`
	}
	if err := decodeWorksheetBody(w, r, &dto); err != nil {
		http.Error(w, "invalid worksheet JSON", 400)
		return
	}
	if dto.Mode != "full_solution" && dto.Mode != "dataset" {
		http.Error(w, "choose full_solution or dataset", 400)
		return
	}
	var dataset *worksheets.Dataset
	var err error
	if dto.Mode == "dataset" {
		if !dto.Reviewed || strings.TrimSpace(dto.Reviewer) == "" || strings.TrimSpace(dto.SourceNote) == "" || len(dto.Reviewer) > 120 || len(dto.SourceNote) > 2000 {
			http.Error(w, "review the case wording, assumptions, hints and keys; confirm row meaning and provide your name and source note", 400)
			return
		}
		dataset, err = worksheets.ParseCSV(dto.CSV)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		dataset.Reviewer = strings.TrimSpace(dto.Reviewer)
		dataset.SourceNote = strings.TrimSpace(dto.SourceNote)
		dto.TemplateID = "binomial_fair_coin_exactly_two"
	} else if dto.CSV != "" || dto.Reviewed || dto.Reviewer != "" || dto.SourceNote != "" {
		http.Error(w, "dataset fields require dataset mode", 400)
		return
	}
	tmpl, ok := s.bank.Get(dto.TemplateID)
	if !ok {
		http.Error(w, "approved question not found", 400)
		return
	}
	rec, err := worksheets.New("worksheet_"+uuid.NewString(), tmpl, dto.Seed, dataset, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err = s.worksheetStore.CreateWorksheet(r.Context(), rec); err != nil {
		http.Error(w, "worksheet could not be saved", 500)
		return
	}
	w.WriteHeader(201)
	candidateJSON(w, worksheets.Public(rec))
}
