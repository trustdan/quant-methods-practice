package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/exam"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func TestExamsHTTPAPIAndSecurityBoundaries(t *testing.T) {
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatal(err)
	}

	b, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil {
		t.Fatal(err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{Bank: b, Store: store})
	if err != nil {
		t.Fatal(err)
	}

	call := func(method, path, body, origin string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1"+path, bytes.NewBufferString(body))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(w, r)
		return w
	}

	// 1. Unauthenticated access rejected with 401
	if w := call("GET", "/api/exams", "", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated request, got %d", w.Code)
	}

	// Authenticate session via bootstrap token
	w := call("POST", "/api/local-session", `{"bootstrap_token":"`+srv.BootstrapToken()+`"}`, "http://127.0.0.1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to authenticate session: %d", w.Code)
	}
	cookie := w.Result().Cookies()[0]

	// 2. Cross-origin mutation rejected with 403
	for _, badOrigin := range []string{"http://evil.example", "http://127.0.0.1:9090"} {
		if w := call("POST", "/api/exams", `{}`, badOrigin, cookie); w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for bad origin %s, got %d", badOrigin, w.Code)
		}
	}

	// 3. Unknown fields or malformed JSON rejected with 400
	if w := call("POST", "/api/exams", `{"unknown_prop": 123}`, "http://127.0.0.1", cookie); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown fields, got %d", w.Code)
	}

	// 4. Create an assessment session (POST /api/exams)
	w = call("POST", "/api/exams", `{"title":"Midterm Test","question_count":2,"duration_minutes":30,"seed":42}`, "http://127.0.0.1", cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create exam: %d %s", w.Code, w.Body.String())
	}

	// Verify key and hint withholding in returned JSON
	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "expected_answer") || strings.Contains(bodyStr, "hint_markdown") || strings.Contains(bodyStr, "explanation_markdown") {
		t.Fatalf("active exam leaked expected answers or explanations: %s", bodyStr)
	}

	var pubView exam.PublicExamView
	if err := json.Unmarshal(w.Body.Bytes(), &pubView); err != nil {
		t.Fatalf("failed to unmarshal public exam view: %v", err)
	}
	if pubView.Status != exam.StatusActive || pubView.TotalQuestions != 2 || pubView.Report != nil {
		t.Fatalf("unexpected active exam view: %+v", pubView)
	}

	examID := pubView.ID

	// 5. List exams (GET /api/exams)
	w = call("GET", "/api/exams", "", "", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to list exams: %d", w.Code)
	}
	var summaries []exam.ExamSummary
	if err := json.Unmarshal(w.Body.Bytes(), &summaries); err != nil || len(summaries) == 0 {
		t.Fatalf("expected exam summaries: %v", err)
	}

	// 6. Get exam by ID (GET /api/exams/{id})
	w = call("GET", "/api/exams/"+examID, "", "", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to get exam by ID: %d", w.Code)
	}

	// 7. Requesting report while active rejected with 403 (GET /api/exams/{id}/report)
	w = call("GET", "/api/exams/"+examID+"/report", "", "", cookie)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 fetching report for active exam, got %d", w.Code)
	}

	// 8. Submit answer command (POST /api/exams/{id}/commands)
	firstQ := pubView.Questions[0]
	firstStage := firstQ.Stages[0]

	cmdJSON := fmt.Sprintf(`{"command_id":"cmd_save_1","expected_revision":%d,"type":"save_answer","stage_id":"%s","answer":{"kind":"%s","option_id":"%s"}}`,
		pubView.Revision, firstStage.ID, firstStage.Kind, firstStage.Options[0].ID)

	w = call("POST", "/api/exams/"+examID+"/commands", cmdJSON, "http://127.0.0.1", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("save_answer command failed: %d %s", w.Code, w.Body.String())
	}

	var updatedView exam.PublicExamView
	if err := json.Unmarshal(w.Body.Bytes(), &updatedView); err != nil || updatedView.Revision != pubView.Revision+1 {
		t.Fatalf("unexpected updated exam view: %+v", updatedView)
	}

	// 9. Submit exam (POST /api/exams/{id}/commands)
	cmdSubmitJSON := fmt.Sprintf(`{"command_id":"cmd_submit_exam","expected_revision":%d,"type":"submit_exam"}`, updatedView.Revision)
	w = call("POST", "/api/exams/"+examID+"/commands", cmdSubmitJSON, "http://127.0.0.1", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("submit_exam command failed: %d %s", w.Code, w.Body.String())
	}

	var completedView exam.PublicExamView
	if err := json.Unmarshal(w.Body.Bytes(), &completedView); err != nil {
		t.Fatalf("failed to unmarshal completed view: %v", err)
	}
	if completedView.Status != exam.StatusCompleted || completedView.Report == nil {
		t.Fatalf("expected completed status with report, got %+v", completedView)
	}

	// 10. Fetch report after completion (GET /api/exams/{id}/report)
	w = call("GET", "/api/exams/"+examID+"/report", "", "", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to fetch exam report after completion: %d", w.Code)
	}

	// 11. Fetch report as markdown export (GET /api/exams/{id}/report?format=markdown)
	w = call("GET", "/api/exams/"+examID+"/report?format=markdown", "", "", cookie)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "# Assessment Report:") {
		t.Fatalf("failed to fetch markdown export: %d %s", w.Code, w.Body.String())
	}

	// 12. History endpoint (GET /api/history)
	w = call("GET", "/api/history?limit=10&offset=0", "", "", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to fetch history: %d", w.Code)
	}
	var history []storage.HistoryItem
	if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil || len(history) == 0 {
		t.Fatalf("expected history items: %v", err)
	}
	if history[0].ID != examID || history[0].Type != "exam" || history[0].Status != "completed" {
		t.Fatalf("unexpected history item: %+v", history[0])
	}
}
