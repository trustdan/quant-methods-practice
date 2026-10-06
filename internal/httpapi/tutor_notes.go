package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
	"github.com/trustdan/quant-methods-practice/internal/storage"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

// NoteStore defines the persistence interface required for saved notes and drafts.
type NoteStore interface {
	SaveExplanation(ctx context.Context, note *storage.SavedExplanation) error
	GetExplanation(ctx context.Context, id string) (*storage.SavedExplanation, error)
	ListExplanations(ctx context.Context, queryFilter, topicFilter string) ([]*storage.SavedExplanation, error)
	DeleteExplanation(ctx context.Context, id string) error
	SaveTutorDraft(ctx context.Context, draft *storage.TutorDraft) error
	GetTutorDraft(ctx context.Context, id string) (*storage.TutorDraft, error)
	ClearTutorDraft(ctx context.Context, id string) error
}

type createTutorRequestDTO struct {
	SessionID       string                  `json:"session_id,omitempty"`
	InstanceID      string                  `json:"instance_id,omitempty"`
	StageID         string                  `json:"stage_id,omitempty"`
	Action          tutor.TutorAction       `json:"action"`
	FollowUpKind    tutor.FollowUpKind      `json:"follow_up_kind,omitempty"`
	CustomPrompt    string                  `json:"custom_prompt,omitempty"`
	Provider        string                  `json:"provider,omitempty"`
	SubmittedAnswer *domain.SubmittedAnswer `json:"submitted_answer,omitempty"`
}

type tutorRequestResponseDTO struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

func (s *Server) handleTutorRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var dto createTutorRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if dto.Action == "" {
		dto.Action = tutor.ActionExplain
	}

	tutorReq := tutor.TutorRequest{
		RequestID:       tutor.GenerateRequestID(),
		SessionID:       dto.SessionID,
		InstanceID:      dto.InstanceID,
		StageID:         dto.StageID,
		Action:          dto.Action,
		FollowUpKind:    dto.FollowUpKind,
		CustomPrompt:    dto.CustomPrompt,
		Provider:        dto.Provider,
		SubmittedAnswer: dto.SubmittedAnswer,
	}

	// Server-side context derivation: populate instance and stage snapshots from verified session
	if dto.SessionID != "" && s.sessionManager != nil {
		sess, ok := s.sessionManager.GetSession(dto.SessionID)
		if ok && sess != nil {
			// Record tutor assistance exposure on active/unresolved stage (strict read-only regarding grades)
			if dto.StageID != "" {
				_ = sess.RecordTutorAssistance(dto.StageID, nil)
			}

			// Find matching question instance
			if len(sess.Questions) > 0 {
				var matchedQ *drill.QuestionState
				for i := range sess.Questions {
					q := &sess.Questions[i]
					if q.QuestionInstance.ID == dto.InstanceID || q.QuestionInstance.TemplateID == dto.InstanceID || dto.InstanceID == "" {
						matchedQ = q
						break
					}
				}
				if matchedQ == nil && sess.CurrentQuestionIndex >= 0 && sess.CurrentQuestionIndex < len(sess.Questions) {
					matchedQ = &sess.Questions[sess.CurrentQuestionIndex]
				}
				if matchedQ != nil {
					tutorReq.Instance = &matchedQ.QuestionInstance
					for _, st := range matchedQ.Stages {
						if st.Instance.ID == dto.StageID || dto.StageID == "" {
							tutorReq.Stage = &st.Instance
							break
						}
					}
				}
			} else {
				tutorReq.Instance = &sess.QuestionInstance
				for _, st := range sess.Stages {
					if st.Instance.ID == dto.StageID || dto.StageID == "" {
						tutorReq.Stage = &st.Instance
						break
					}
				}
			}
		}
	}

	reqID, _, err := s.tutorManager.StartRequest(context.Background(), tutorReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start tutor request: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tutorRequestResponseDTO{
		RequestID: reqID,
		Status:    "started",
	})
}

func (s *Server) handleTutorRequestByID(w http.ResponseWriter, r *http.Request) {
	// Path format: /api/tutor/requests/{id} or /api/tutor/requests/{id}/events
	subPath := strings.TrimPrefix(r.URL.Path, "/api/tutor/requests/")
	parts := strings.Split(subPath, "/")
	reqID := parts[0]

	if reqID == "" {
		http.Error(w, "missing request ID", http.StatusBadRequest)
		return
	}

	// 1. Streaming events: GET /api/tutor/requests/{id}/events
	if len(parts) >= 2 && parts[1] == "events" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ch, err := s.tutorManager.GetEventsChannel(reqID)
		if err != nil {
			http.Error(w, "request not found or already completed", http.StatusNotFound)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		for {
			select {
			case <-r.Context().Done():
				s.tutorManager.CancelRequest(reqID)
				return
			case ev, open := <-ch:
				if !open {
					return
				}
				dataBytes, _ := json.Marshal(ev)
				fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
				flusher.Flush()

				if ev.Type == tutor.EventComplete || ev.Type == tutor.EventCancelled || ev.Type == tutor.EventError {
					return
				}
			}
		}
	}

	// 2. Cancellation: DELETE /api/tutor/requests/{id}
	if r.Method == http.MethodDelete {
		s.tutorManager.CancelRequest(reqID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":     "cancelled",
			"request_id": reqID,
		})
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	if s.noteStore == nil {
		http.Error(w, "note storage unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query().Get("q")
		topic := r.URL.Query().Get("topic")

		notes, err := s.noteStore.ListExplanations(r.Context(), q, topic)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to list notes: %v", err), http.StatusInternalServerError)
			return
		}
		if notes == nil {
			notes = []*storage.SavedExplanation{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(notes)

	case http.MethodPost:
		var note storage.SavedExplanation
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(note.ID) == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			note.ID = fmt.Sprintf("note_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
		}

		if err := s.noteStore.SaveExplanation(r.Context(), &note); err != nil {
			http.Error(w, fmt.Sprintf("failed to save note: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(note)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNoteByID(w http.ResponseWriter, r *http.Request) {
	if s.noteStore == nil {
		http.Error(w, "note storage unavailable", http.StatusServiceUnavailable)
		return
	}

	subPath := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	parts := strings.Split(subPath, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, "missing note ID", http.StatusBadRequest)
		return
	}

	// Export endpoint: GET /api/notes/{id}/export
	if len(parts) >= 2 && parts[1] == "export" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		note, err := s.noteStore.GetExplanation(r.Context(), id)
		if err != nil {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}

		scenario := ""
		if note.OriginInstanceID != "" && s.sessionManager != nil {
			if sess, ok := s.sessionManager.GetSession(s.lastSessionID); ok && sess != nil {
				scenario = sess.QuestionInstance.ScenarioMarkdown
			}
		}

		md := tutor.ExportNoteToMarkdown(note, scenario)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("note-%s.md", note.ID)))
		_, _ = w.Write([]byte(md))
		return
	}

	switch r.Method {
	case http.MethodGet:
		note, err := s.noteStore.GetExplanation(r.Context(), id)
		if err != nil {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(note)

	case http.MethodDelete:
		if err := s.noteStore.DeleteExplanation(r.Context(), id); err != nil {
			http.Error(w, fmt.Sprintf("failed to delete note: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "deleted",
			"id":     id,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type exportRequestDTO struct {
	Type    string   `json:"type"`
	NoteIDs []string `json:"note_ids,omitempty"`
}

func (s *Server) handleExports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req exportRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if s.noteStore == nil {
		http.Error(w, "note store unavailable", http.StatusServiceUnavailable)
		return
	}

	notes, err := s.noteStore.ListExplanations(r.Context(), "", "")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve notes for export: %v", err), http.StatusInternalServerError)
		return
	}

	var exported []map[string]string
	idFilter := make(map[string]bool)
	for _, id := range req.NoteIDs {
		idFilter[id] = true
	}

	for _, n := range notes {
		if len(idFilter) > 0 && !idFilter[n.ID] {
			continue
		}
		md := tutor.ExportNoteToMarkdown(n, "")
		exported = append(exported, map[string]string{
			"id":       n.ID,
			"title":    n.ProviderInfo.Title,
			"markdown": md,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(exported)
}

func (s *Server) handleTutorDrafts(w http.ResponseWriter, r *http.Request) {
	if s.noteStore == nil {
		http.Error(w, "draft storage unavailable", http.StatusServiceUnavailable)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/tutor/drafts/")
	if id == "" {
		http.Error(w, "missing draft ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		draft, err := s.noteStore.GetTutorDraft(r.Context(), id)
		if err != nil {
			http.Error(w, "draft not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(draft)

	case http.MethodPost:
		var draft storage.TutorDraft
		if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		draft.ID = id
		if err := s.noteStore.SaveTutorDraft(r.Context(), &draft); err != nil {
			http.Error(w, fmt.Sprintf("failed to save tutor draft: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(draft)

	case http.MethodDelete:
		_ = s.noteStore.ClearTutorDraft(r.Context(), id)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "cleared",
			"id":     id,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
