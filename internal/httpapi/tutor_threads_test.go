package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/drill"
	"github.com/trustdan/quant-methods-practice/internal/storage"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

type mockThreadStore struct {
	threads  map[string]*storage.TutorThread
	messages map[string][]*storage.TutorMessage
}

func newMockThreadStore() *mockThreadStore {
	return &mockThreadStore{
		threads:  make(map[string]*storage.TutorThread),
		messages: make(map[string][]*storage.TutorMessage),
	}
}

func (m *mockThreadStore) CreateTutorThread(ctx context.Context, thread *storage.TutorThread) error {
	m.threads[thread.ID] = thread
	return nil
}

func (m *mockThreadStore) GetTutorThread(ctx context.Context, id string) (*storage.TutorThread, error) {
	th, ok := m.threads[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	msgs := m.messages[id]
	thCopy := *th
	thCopy.Messages = make([]storage.TutorMessage, len(msgs))
	for i, msg := range msgs {
		thCopy.Messages[i] = *msg
	}
	return &thCopy, nil
}

func (m *mockThreadStore) ListTutorThreads(ctx context.Context, instanceID, stageID string) ([]*storage.TutorThread, error) {
	var res []*storage.TutorThread
	for _, th := range m.threads {
		if instanceID != "" && th.InstanceID != instanceID {
			continue
		}
		if stageID != "" && th.StageID != stageID {
			continue
		}
		res = append(res, th)
	}
	return res, nil
}

func (m *mockThreadStore) DeleteTutorThread(ctx context.Context, id string) error {
	if _, ok := m.threads[id]; !ok {
		return storage.ErrNotFound
	}
	delete(m.threads, id)
	delete(m.messages, id)
	return nil
}

func (m *mockThreadStore) AppendTutorMessage(ctx context.Context, msg *storage.TutorMessage) error {
	m.messages[msg.ThreadID] = append(m.messages[msg.ThreadID], msg)
	return nil
}

func (m *mockThreadStore) GetTutorMessages(ctx context.Context, threadID string) ([]*storage.TutorMessage, error) {
	return m.messages[threadID], nil
}

func TestTutorThreadsHTTPLifecycle(t *testing.T) {
	threadStore := newMockThreadStore()
	offlineTutor := tutor.NewOfflineTutor(1*time.Millisecond, time.Now)
	tutorMgr := tutor.NewTutorManager(offlineTutor)

	server, err := NewServer(Config{
		Addr:             "127.0.0.1:0",
		TutorThreadStore: threadStore,
		TutorManager:     tutorMgr,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	handler := server.httpServer.Handler

	// 1. POST /api/tutor/threads
	createBody, _ := json.Marshal(createThreadRequestDTO{
		InstanceID: "inst_test_1",
		StageID:    "stage_calc",
		Provider:   "offline",
		Model:      "offline-curriculum",
		Title:      "Question on Trial Independence",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/threads", bytes.NewReader(createBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdThread storage.TutorThread
	if err := json.NewDecoder(rec.Body).Decode(&createdThread); err != nil {
		t.Fatalf("failed to decode created thread: %v", err)
	}
	if createdThread.ID == "" || createdThread.Title != "Question on Trial Independence" {
		t.Fatalf("unexpected created thread: %+v", createdThread)
	}

	// 2. GET /api/tutor/threads
	req = httptest.NewRequest(http.MethodGet, "/api/tutor/threads?instance_id=inst_test_1", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var threads []storage.TutorThread
	if err := json.NewDecoder(rec.Body).Decode(&threads); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(threads) != 1 || threads[0].ID != createdThread.ID {
		t.Fatalf("unexpected list threads: %+v", threads)
	}

	// 3. POST /api/tutor/threads/{id}/messages
	postMsgBody, _ := json.Marshal(postMessageRequestDTO{
		Role:    "user",
		Content: "Can you explain why independence is required?",
	})
	msgURL := fmt.Sprintf("/api/tutor/threads/%s/messages", createdThread.ID)
	req = httptest.NewRequest(http.MethodPost, msgURL, bytes.NewReader(postMsgBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}

	var postMsgResp postMessageResponseDTO
	if err := json.NewDecoder(rec.Body).Decode(&postMsgResp); err != nil {
		t.Fatalf("failed to decode post message response: %v", err)
	}
	if postMsgResp.UserMessageID == "" || postMsgResp.AssistantMessageID == "" {
		t.Fatalf("unexpected response IDs: %+v", postMsgResp)
	}

	// 4. GET /api/tutor/threads/{id}/messages/{msgId}/events
	eventsURL := fmt.Sprintf("/api/tutor/threads/%s/messages/%s/events", createdThread.ID, postMsgResp.AssistantMessageID)
	req = httptest.NewRequest(http.MethodGet, eventsURL, nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from events, got %d", rec.Code)
	}
	eventsOutput := rec.Body.String()
	if !strings.Contains(eventsOutput, "started") || !strings.Contains(eventsOutput, "complete") {
		t.Fatalf("expected started and complete in stream output, got:\n%s", eventsOutput)
	}

	// Allow goroutine to persist assistant message
	time.Sleep(50 * time.Millisecond)

	// 5. GET /api/tutor/threads/{id}
	getThreadURL := fmt.Sprintf("/api/tutor/threads/%s", createdThread.ID)
	req = httptest.NewRequest(http.MethodGet, getThreadURL, nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from get thread, got %d", rec.Code)
	}
	var threadDetail storage.TutorThread
	if err := json.NewDecoder(rec.Body).Decode(&threadDetail); err != nil {
		t.Fatalf("failed to decode thread detail: %v", err)
	}
	if len(threadDetail.Messages) < 2 {
		t.Fatalf("expected at least 2 messages (user + assistant), got %d", len(threadDetail.Messages))
	}
	if threadDetail.Messages[0].Role != "user" || threadDetail.Messages[1].Role != "assistant" {
		t.Errorf("unexpected message roles: %+v", threadDetail.Messages)
	}

	// 6. DELETE /api/tutor/threads/{id}
	req = httptest.NewRequest(http.MethodDelete, getThreadURL, nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from delete thread, got %d", rec.Code)
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, getThreadURL, nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found after deletion, got %d", rec.Code)
	}
}

func TestTutorPedagogicalGuardrailsInSession(t *testing.T) {
	// Verify that unresolved active stages receive Socratic guardrails
	sessMgr := drill.NewSessionManager(nil)
	threadStore := newMockThreadStore()
	offlineTutor := tutor.NewOfflineTutor(1*time.Millisecond, time.Now)
	tutorMgr := tutor.NewTutorManager(offlineTutor)

	server, err := NewServer(Config{
		Addr:             "127.0.0.1:0",
		SessionManager:   sessMgr,
		TutorThreadStore: threadStore,
		TutorManager:     tutorMgr,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	handler := server.httpServer.Handler

	// Start a drill session
	tmplPath := filepath.Join("..", "..", "curriculum", "approved", "binomial-fair-coin-exactly-two.json")
	tmpl, err := bank.ValidateTemplateFile(tmplPath, nil)
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}
	sess, err := sessMgr.CreateSession(tmpl, 1337)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Stage 0 is active and unresolved
	stageID := sess.Stages[0].Instance.ID
	instanceID := sess.QuestionInstance.ID

	// Create thread tied to this session
	createBody, _ := json.Marshal(createThreadRequestDTO{
		SessionID:  sess.ID,
		InstanceID: instanceID,
		StageID:    stageID,
		Provider:   auth.RouteOffline,
		Title:      "Active Stage Help",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tutor/threads", bytes.NewReader(createBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var thread storage.TutorThread
	_ = json.NewDecoder(rec.Body).Decode(&thread)

	// Post message to unresolved stage
	postMsgBody, _ := json.Marshal(postMessageRequestDTO{
		Role:    "user",
		Content: "What is the answer to this question?",
	})
	msgURL := fmt.Sprintf("/api/tutor/threads/%s/messages", thread.ID)
	req = httptest.NewRequest(http.MethodPost, msgURL, bytes.NewReader(postMsgBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var postMsgResp postMessageResponseDTO
	_ = json.NewDecoder(rec.Body).Decode(&postMsgResp)

	// Stream events
	eventsURL := fmt.Sprintf("/api/tutor/threads/%s/messages/%s/events", thread.ID, postMsgResp.AssistantMessageID)
	req = httptest.NewRequest(http.MethodGet, eventsURL, nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	streamOutput := rec.Body.String()
	// Because stage is unresolved, Socratic guidance must be given rather than complete derivation
	if !strings.Contains(streamOutput, "Socratic Guidance") {
		t.Errorf("expected Socratic guidance on unresolved active stage, got:\n%s", streamOutput)
	}
}
