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

// TutorThreadStore defines the storage capabilities needed for conversational threads.
type TutorThreadStore interface {
	CreateTutorThread(ctx context.Context, thread *storage.TutorThread) error
	GetTutorThread(ctx context.Context, id string) (*storage.TutorThread, error)
	ListTutorThreads(ctx context.Context, instanceID, stageID string) ([]*storage.TutorThread, error)
	DeleteTutorThread(ctx context.Context, id string) error
	AppendTutorMessage(ctx context.Context, msg *storage.TutorMessage) error
	GetTutorMessages(ctx context.Context, threadID string) ([]*storage.TutorMessage, error)
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(b))
}

type createThreadRequestDTO struct {
	SessionID  string `json:"session_id,omitempty"`
	InstanceID string `json:"instance_id"`
	StageID    string `json:"stage_id"`
	Provider   string `json:"provider,omitempty"`
	Model      string `json:"model,omitempty"`
	Title      string `json:"title,omitempty"`
}

type postMessageRequestDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type postMessageResponseDTO struct {
	UserMessageID      string `json:"user_message_id"`
	AssistantMessageID string `json:"assistant_message_id"`
	RequestID          string `json:"request_id"`
	Status             string `json:"status"`
}

// handleTutorThreads routes /api/tutor/threads and /api/tutor/threads/*
func (s *Server) handleTutorThreads(w http.ResponseWriter, r *http.Request) {
	if s.tutorThreadStore == nil {
		http.Error(w, "thread storage unavailable", http.StatusServiceUnavailable)
		return
	}

	subPath := strings.TrimPrefix(r.URL.Path, "/api/tutor/threads")
	subPath = strings.TrimPrefix(subPath, "/")

	// 1. Root /api/tutor/threads
	if subPath == "" {
		switch r.Method {
		case http.MethodGet:
			instanceID := r.URL.Query().Get("instance_id")
			stageID := r.URL.Query().Get("stage_id")
			threads, err := s.tutorThreadStore.ListTutorThreads(r.Context(), instanceID, stageID)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to list threads: %v", err), http.StatusInternalServerError)
				return
			}
			if threads == nil {
				threads = []*storage.TutorThread{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(threads)
			return

		case http.MethodPost:
			var dto createThreadRequestDTO
			if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			if dto.InstanceID == "" || dto.StageID == "" {
				http.Error(w, "instance_id and stage_id are required", http.StatusBadRequest)
				return
			}

			provider := dto.Provider
			if provider == "" || provider == "active" {
				if s.providerManager != nil {
					provider = s.providerManager.GetActiveRoute()
				}
			}
			if provider == "" {
				provider = "offline"
			}

			model := dto.Model
			if model == "" && s.providerManager != nil {
				model = s.providerManager.GetActiveModel()
			}
			if model == "" {
				model = "offline-curriculum"
			}

			title := strings.TrimSpace(dto.Title)
			if title == "" {
				title = fmt.Sprintf("Discussion on %s", dto.StageID)
			}

			thread := &storage.TutorThread{
				ID:         generateID("th"),
				SessionID:  dto.SessionID,
				InstanceID: dto.InstanceID,
				StageID:    dto.StageID,
				Provider:   provider,
				Model:      model,
				Title:      title,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			}

			if err := s.tutorThreadStore.CreateTutorThread(r.Context(), thread); err != nil {
				http.Error(w, fmt.Sprintf("failed to create thread: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(thread)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	// 2. Thread subroutes: {threadID}, {threadID}/messages, {threadID}/messages/{msgID}/events
	parts := strings.Split(subPath, "/")
	threadID := parts[0]
	if threadID == "" {
		http.Error(w, "missing thread ID", http.StatusBadRequest)
		return
	}

	// GET or DELETE /api/tutor/threads/{threadID}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			thread, err := s.tutorThreadStore.GetTutorThread(r.Context(), threadID)
			if err != nil {
				if err == storage.ErrNotFound {
					http.Error(w, "thread not found", http.StatusNotFound)
					return
				}
				http.Error(w, fmt.Sprintf("failed to get thread: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(thread)
			return

		case http.MethodDelete:
			if err := s.tutorThreadStore.DeleteTutorThread(r.Context(), threadID); err != nil {
				if err == storage.ErrNotFound {
					http.Error(w, "thread not found", http.StatusNotFound)
					return
				}
				http.Error(w, fmt.Sprintf("failed to delete thread: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": threadID})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	// Subroute: /api/tutor/threads/{threadID}/messages
	if parts[1] == "messages" {
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				msgs, err := s.tutorThreadStore.GetTutorMessages(r.Context(), threadID)
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to get messages: %v", err), http.StatusInternalServerError)
					return
				}
				if msgs == nil {
					msgs = []*storage.TutorMessage{}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(msgs)
				return

			case http.MethodPost:
				s.handlePostThreadMessage(w, r, threadID)
				return

			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// Subroute: /api/tutor/threads/{threadID}/messages/{msgID}/events
		if len(parts) >= 4 && parts[3] == "events" {
			msgID := parts[2]
			s.handleThreadMessageEvents(w, r, threadID, msgID)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handlePostThreadMessage(w http.ResponseWriter, r *http.Request, threadID string) {
	thread, err := s.tutorThreadStore.GetTutorThread(r.Context(), threadID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "thread not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to retrieve thread: %v", err), http.StatusInternalServerError)
		return
	}

	var dto postMessageRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	content := strings.TrimSpace(dto.Content)
	if content == "" {
		http.Error(w, "content cannot be empty", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	userMsgID := generateID("msg_u")
	userMsg := &storage.TutorMessage{
		ID:        userMsgID,
		ThreadID:  thread.ID,
		Role:      "user",
		Content:   content,
		CreatedAt: now,
	}

	if err := s.tutorThreadStore.AppendTutorMessage(r.Context(), userMsg); err != nil {
		http.Error(w, fmt.Sprintf("failed to save message: %v", err), http.StatusInternalServerError)
		return
	}

	// Derive stage resolution status & problem snapshots from session
	isResolved := false
	var instanceSnapshot *domain.QuestionInstance
	var stageSnapshot *domain.StageInstance

	if thread.SessionID != "" && s.sessionManager != nil {
		sess, ok := s.sessionManager.GetSession(thread.SessionID)
		if ok && sess != nil {
			_ = sess.RecordTutorAssistance(thread.StageID, nil)

			// Locate stage to determine if completed
			if len(sess.Questions) > 0 {
				for i := range sess.Questions {
					q := &sess.Questions[i]
					if q.QuestionInstance.ID == thread.InstanceID || q.QuestionInstance.TemplateID == thread.InstanceID {
						instanceSnapshot = &q.QuestionInstance
						for _, st := range q.Stages {
							if st.Instance.ID == thread.StageID {
								stageSnapshot = &st.Instance
								if st.Status == drill.StageStatusCompleted {
									isResolved = true
								}
								break
							}
						}
						break
					}
				}
			} else {
				instanceSnapshot = &sess.QuestionInstance
				for _, st := range sess.Stages {
					if st.Instance.ID == thread.StageID {
						stageSnapshot = &st.Instance
						if st.Status == drill.StageStatusCompleted {
							isResolved = true
						}
						break
					}
				}
			}
		}
	}

	// Gather prior history
	allMsgs, _ := s.tutorThreadStore.GetTutorMessages(r.Context(), thread.ID)
	var history []tutor.ChatMessage
	for _, m := range allMsgs {
		history = append(history, tutor.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	assistantMsgID := generateID("msg_a")

	provider := thread.Provider
	if provider == "" || provider == "active" {
		if s.providerManager != nil {
			provider = s.providerManager.GetActiveRoute()
		}
	}
	if provider == "" {
		provider = "offline"
	}

	tutorReq := tutor.TutorRequest{
		RequestID:     assistantMsgID,
		SessionID:     thread.SessionID,
		InstanceID:    thread.InstanceID,
		StageID:       thread.StageID,
		ThreadID:      thread.ID,
		StageResolved: isResolved,
		Action:        tutor.ActionChat,
		CustomPrompt:  content,
		History:       history,
		Provider:      provider,
		Instance:      instanceSnapshot,
		Stage:         stageSnapshot,
	}

	_, _, err = s.tutorManager.StartRequest(context.Background(), tutorReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start tutor response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(postMessageResponseDTO{
		UserMessageID:      userMsgID,
		AssistantMessageID: assistantMsgID,
		RequestID:          assistantMsgID,
		Status:             "started",
	})
}

func (s *Server) handleThreadMessageEvents(w http.ResponseWriter, r *http.Request, threadID, msgID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ch, err := s.tutorManager.GetEventsChannel(msgID)
	if err != nil {
		http.Error(w, "stream not active or already completed", http.StatusNotFound)
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

	var fullText strings.Builder
	var fallbackLabel string

	for {
		select {
		case <-r.Context().Done():
			s.tutorManager.CancelRequest(msgID)
			return
		case ev, open := <-ch:
			if !open {
				return
			}

			if ev.Type == tutor.EventTextDelta {
				fullText.WriteString(ev.Delta)
			} else if ev.Type == tutor.EventFallback {
				fallbackLabel = ev.FallbackLabel
			} else if ev.Type == tutor.EventComplete {
				completedText := ev.Text
				if completedText == "" {
					completedText = fullText.String()
				}
				if completedText != "" {
					asstMsg := &storage.TutorMessage{
						ID:            msgID,
						ThreadID:      threadID,
						Role:          "assistant",
						Content:       completedText,
						FallbackLabel: fallbackLabel,
						CreatedAt:     time.Now().UTC(),
					}
					_ = s.tutorThreadStore.AppendTutorMessage(context.Background(), asstMsg)
				}
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
