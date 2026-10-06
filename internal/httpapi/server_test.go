package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func newTestServer(t *testing.T) *Server {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Test UI</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('test');"),
		},
	}

	srv, err := NewServer(Config{
		Addr:     "127.0.0.1:0",
		AssetsFS: mockFS,
		Version:  "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var data map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if data["status"] != "ok" || data["version"] != "0.1.0-test" {
		t.Fatalf("unexpected health data: %v", data)
	}

	// Verify security headers
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing or incorrect X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("missing or incorrect X-Frame-Options header")
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Errorf("missing Content-Security-Policy header")
	}
}

func TestLocalSessionExchange(t *testing.T) {
	srv := newTestServer(t)
	bootstrapToken := srv.BootstrapToken()

	// 1. Valid exchange
	body, _ := json.Marshal(map[string]string{"bootstrap_token": bootstrapToken})
	req := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on first exchange, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "quant_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || !sessionCookie.HttpOnly {
		t.Fatalf("expected HttpOnly quant_session cookie")
	}

	// 2. Second exchange with same bootstrap token must fail (single-use)
	req2 := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req2.Host = "127.0.0.1"
	rec2 := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reused bootstrap token, got %d", rec2.Code)
	}
}

func TestHostHeaderRejection(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Host = "attacker.example.com"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for non-loopback host, got %d", rec.Code)
	}
}

func TestCrossOriginMutationRejection(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"bootstrap_token": "foo"})
	req := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://malicious-site.com")
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for disallowed origin, got %d", rec.Code)
	}
}

func TestStaticAndSPAFallback(t *testing.T) {
	srv := newTestServer(t)

	// Static file request
	req1 := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	req1.Host = "127.0.0.1"
	rec1 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 for static asset, got %d", rec1.Code)
	}

	// SPA route fallback (e.g. /practice/session-1) should serve index.html
	req2 := httptest.NewRequest(http.MethodGet, "/practice/session-1", nil)
	req2.Host = "127.0.0.1"
	rec2 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route, got %d", rec2.Code)
	}
	if !bytes.Contains(rec2.Body.Bytes(), []byte("Test UI")) {
		t.Fatalf("expected index.html body on SPA fallback")
	}
}

func TestPracticeSessionsAPI(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("UI")},
	}

	tmplPath := filepath.Join("..", "..", "curriculum", "approved", "binomial-fair-coin-exactly-two.json")
	tmpl, err := bank.ValidateTemplateFile(tmplPath, nil)
	if err != nil {
		t.Fatalf("failed to load approved template: %v", err)
	}

	b := bank.NewBank()
	b.Add(tmpl)

	srv, err := NewServer(Config{
		Addr:     "127.0.0.1:0",
		AssetsFS: mockFS,
		Version:  "0.1.0-test",
		Bank:     b,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// 1. Create session via POST /api/practice/sessions
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/practice/sessions", bytes.NewReader([]byte("{}")))
	reqCreate.Host = "127.0.0.1"
	recCreate := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recCreate, reqCreate)

	if recCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", recCreate.Code, recCreate.Body.String())
	}

	var sessionView drill.PublicSessionView
	if err := json.NewDecoder(recCreate.Body).Decode(&sessionView); err != nil {
		t.Fatalf("failed to decode session view: %v", err)
	}

	if sessionView.ID == "" || len(sessionView.Stages) != 7 {
		t.Fatalf("expected 7 stages in session, got %d", len(sessionView.Stages))
	}

	// 2. Retrieve session via GET /api/practice/sessions/{id}
	reqGet := httptest.NewRequest(http.MethodGet, "/api/practice/sessions/"+sessionView.ID, nil)
	reqGet.Host = "127.0.0.1"
	recGet := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on get session, got %d", recGet.Code)
	}

	// 3. Submit choice answer on Stage 1
	subBody, _ := json.Marshal(drill.SessionCommand{
		CommandID:        "cmd_http_1",
		ExpectedRevision: sessionView.Revision,
		Type:             drill.CmdSubmitAnswer,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "count_heads",
		},
	})
	reqCmd := httptest.NewRequest(http.MethodPost, "/api/practice/sessions/"+sessionView.ID+"/commands", bytes.NewReader(subBody))
	reqCmd.Host = "127.0.0.1"
	recCmd := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recCmd, reqCmd)

	if recCmd.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on command, got %d: %s", recCmd.Code, recCmd.Body.String())
	}

	var cmdRes drill.CommandResult
	if err := json.NewDecoder(recCmd.Body).Decode(&cmdRes); err != nil {
		t.Fatalf("failed to decode command result: %v", err)
	}
	if !cmdRes.Success {
		t.Fatalf("expected command success, got %s", cmdRes.ErrorMessage)
	}

	// 4. Test numeric grading on stage 6 via command
	// Jump / navigate to stage index 5 (Stage 6: Calculate)
	// First let's advance session to stage 5 directly
	sess, _ := srv.sessionManager.GetSession(sessionView.ID)
	sess.CurrentStageIndex = 5
	sess.Stages[5].Status = drill.StageStatusActive

	for _, numStr := range []string{"0.375", "37.5%", "3/8"} {
		// Reset stage 5 to active for each test variant
		sess.CurrentStageIndex = 5
		sess.Stages[5].Status = drill.StageStatusActive
		sess.Stages[5].Attempts = nil

		numCmdBody, _ := json.Marshal(drill.SessionCommand{
			CommandID: "cmd_num_" + numStr,
			Type:      drill.CmdSubmitAnswer,
			Answer: &domain.SubmittedAnswer{
				Kind:       domain.StageKindNumeric,
				NumericRaw: numStr,
			},
		})
		reqNum := httptest.NewRequest(http.MethodPost, "/api/practice/sessions/"+sessionView.ID+"/commands", bytes.NewReader(numCmdBody))
		reqNum.Host = "127.0.0.1"
		recNum := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(recNum, reqNum)

		if recNum.Code != http.StatusOK {
			t.Fatalf("num %q expected 200 OK, got %d: %s", numStr, recNum.Code, recNum.Body.String())
		}
		var numRes drill.CommandResult
		_ = json.NewDecoder(recNum.Body).Decode(&numRes)
		if !numRes.Success {
			t.Fatalf("num %q expected success, got %s", numStr, numRes.ErrorMessage)
		}
	}
}

func TestPracticeSessionPersistentRestart(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "http_restart.db")

	db1, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ctx := context.Background()
	if err := storage.RunMigrations(ctx, db1, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store1 := storage.NewStore(db1, nil)

	approvedBank, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil || approvedBank.Count() == 0 {
		approvedBank, _ = bank.LoadActiveBank("curriculum/approved", nil)
	}

	// Server 1
	srv1, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Bank:    approvedBank,
		Store:   store1,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer 1 failed: %v", err)
	}

	// 1. Create session on Server 1
	reqPost := httptest.NewRequest(http.MethodPost, "/api/practice/sessions", bytes.NewReader([]byte(`{"template_id":"binomial_fair_coin_exactly_two"}`)))
	reqPost.Host = "127.0.0.1"
	recPost := httptest.NewRecorder()
	srv1.httpServer.Handler.ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", recPost.Code)
	}

	var sessionView drill.PublicSessionView
	if err := json.NewDecoder(recPost.Body).Decode(&sessionView); err != nil {
		t.Fatalf("failed to decode created session: %v", err)
	}

	// 2. Submit answer to Stage 1 on Server 1
	cmdBody, _ := json.Marshal(drill.SessionCommand{
		CommandID:        "cmd_http_st1",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          sessionView.Stages[0].ID,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "count_heads",
		},
	})
	reqCmd := httptest.NewRequest(http.MethodPost, "/api/practice/sessions/"+sessionView.ID+"/commands", bytes.NewReader(cmdBody))
	reqCmd.Host = "127.0.0.1"
	recCmd := httptest.NewRecorder()
	srv1.httpServer.Handler.ServeHTTP(recCmd, reqCmd)

	if recCmd.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on submit, got %d", recCmd.Code)
	}

	// 3. Shutdown Server 1 and close db1
	_ = srv1.Close()
	_ = db1.Close()

	// 4. Start Server 2 with fresh connection to same database
	db2, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Open 2 failed: %v", err)
	}
	defer db2.Close()

	store2 := storage.NewStore(db2, nil)
	srv2, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Bank:    approvedBank,
		Store:   store2,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer 2 failed: %v", err)
	}
	defer srv2.Close()

	// 5. Query GET /api/practice/sessions on Server 2 (no ID provided in path)
	reqGet := httptest.NewRequest(http.MethodGet, "/api/practice/sessions", nil)
	reqGet.Host = "127.0.0.1"
	recGet := httptest.NewRecorder()
	srv2.httpServer.Handler.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on resume, got %d: %s", recGet.Code, recGet.Body.String())
	}

	var resumedView drill.PublicSessionView
	if err := json.NewDecoder(recGet.Body).Decode(&resumedView); err != nil {
		t.Fatalf("failed to decode resumed session: %v", err)
	}

	if resumedView.ID != sessionView.ID {
		t.Errorf("resumed session ID %q != original %q", resumedView.ID, sessionView.ID)
	}
	if resumedView.CurrentStageIndex != 1 {
		t.Errorf("expected resumed stage index 1, got %d", resumedView.CurrentStageIndex)
	}
	if resumedView.Stages[0].Status != drill.StageStatusCompleted {
		t.Errorf("expected Stage 1 completed, got %v", resumedView.Stages[0].Status)
	}
	if len(resumedView.Stages[0].Attempts) != 1 {
		t.Errorf("expected 1 attempt on Stage 1, got %d", len(resumedView.Stages[0].Attempts))
	}
}

func TestSettingsEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tmpDir, "settings_test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Store:   store,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	// 1. GET default settings
	reqGet := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	reqGet.Host = "127.0.0.1"
	recGet := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET /api/settings, got %d", recGet.Code)
	}

	var s1 map[string]any
	_ = json.NewDecoder(recGet.Body).Decode(&s1)
	if s1["question_count"].(float64) != 10 {
		t.Errorf("expected default question_count=10, got %v", s1["question_count"])
	}

	// 2. POST updated settings
	payload := `{"question_count":5,"module_ids":["module_01"],"intensity":"intensive"}`
	reqPost := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader([]byte(payload)))
	reqPost.Host = "127.0.0.1"
	recPost := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusOK {
		t.Fatalf("expected 200 on POST /api/settings, got %d", recPost.Code)
	}

	// 3. GET saved settings
	reqGet2 := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	reqGet2.Host = "127.0.0.1"
	recGet2 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recGet2, reqGet2)

	var s2 map[string]any
	_ = json.NewDecoder(recGet2.Body).Decode(&s2)
	if s2["question_count"].(float64) != 5 {
		t.Errorf("expected saved question_count=5, got %v", s2["question_count"])
	}
	if s2["intensity"] != "intensive" {
		t.Errorf("expected saved intensity='intensive', got %v", s2["intensity"])
	}
}

func TestBankEndpoint(t *testing.T) {
	approvedBank, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil || approvedBank.Count() == 0 {
		approvedBank, _ = bank.LoadActiveBank("curriculum/approved", nil)
	}

	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Bank:    approvedBank,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/bank", nil)
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET /api/bank, got %d", rec.Code)
	}

	var list []templateSummary
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode template summaries: %v", err)
	}

	if len(list) != 10 {
		t.Errorf("expected 10 approved templates in bank, got %d", len(list))
	}
}

func TestMultiQuestionPracticeSessionAPI(t *testing.T) {
	approvedBank, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil || approvedBank.Count() == 0 {
		approvedBank, _ = bank.LoadActiveBank("curriculum/approved", nil)
	}

	tmpDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tmpDir, "multi_api_test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Bank:    approvedBank,
		Store:   store,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	// 1. Create 10-question session via POST /api/practice/sessions
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/practice/sessions", bytes.NewReader([]byte(`{"question_count":10,"intensity":"standard"}`)))
	reqCreate.Host = "127.0.0.1"
	recCreate := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recCreate, reqCreate)

	if recCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", recCreate.Code, recCreate.Body.String())
	}

	var sessionView drill.PublicSessionView
	if err := json.NewDecoder(recCreate.Body).Decode(&sessionView); err != nil {
		t.Fatalf("failed to decode session view: %v", err)
	}

	if sessionView.TotalQuestions != 10 {
		t.Errorf("expected TotalQuestions=10, got %d", sessionView.TotalQuestions)
	}
	if len(sessionView.Questions) != 10 {
		t.Errorf("expected 10 question summaries, got %d", len(sessionView.Questions))
	}
	if sessionView.CurrentQuestionIndex != 0 {
		t.Errorf("expected CurrentQuestionIndex=0, got %d", sessionView.CurrentQuestionIndex)
	}

	// 2. Dispatch navigate_question command to Question 3
	targetIdx := 3
	cmdBody, _ := json.Marshal(drill.SessionCommand{
		CommandID:           "cmd_nav_q3",
		ExpectedRevision:    sessionView.Revision,
		Type:                drill.CmdNavigateQuestion,
		TargetQuestionIndex: &targetIdx,
	})
	reqCmd := httptest.NewRequest(http.MethodPost, "/api/practice/sessions/"+sessionView.ID+"/commands", bytes.NewReader(cmdBody))
	reqCmd.Host = "127.0.0.1"
	recCmd := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recCmd, reqCmd)

	if recCmd.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on navigate question, got %d: %s", recCmd.Code, recCmd.Body.String())
	}

	var cmdRes drill.CommandResult
	if err := json.NewDecoder(recCmd.Body).Decode(&cmdRes); err != nil {
		t.Fatalf("failed to decode command result: %v", err)
	}
	if !cmdRes.Success {
		t.Fatalf("navigate command failed: %s", cmdRes.ErrorMessage)
	}
	if cmdRes.SessionState.CurrentQuestionIndex != 3 {
		t.Errorf("expected CurrentQuestionIndex=3 after navigation, got %d", cmdRes.SessionState.CurrentQuestionIndex)
	}
}

func TestMasteryEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tmpDir, "mastery_api_test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Store:   store,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	// GET /api/mastery
	reqGet := httptest.NewRequest(http.MethodGet, "/api/mastery", nil)
	reqGet.Host = "127.0.0.1"
	recGet := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/mastery, got %d: %s", recGet.Code, recGet.Body.String())
	}

	var summary map[string]any
	if err := json.NewDecoder(recGet.Body).Decode(&summary); err != nil {
		t.Fatalf("failed to decode mastery summary: %v", err)
	}

	if _, ok := summary["policy_version"]; !ok {
		t.Errorf("expected policy_version in mastery response")
	}
	if _, ok := summary["concepts"]; !ok {
		t.Errorf("expected concepts in mastery response")
	}
}

func TestTutorRequestAndCancellationEndpoint(t *testing.T) {
	srv := newTestServer(t)

	// 1. POST /api/tutor/requests
	reqBody, _ := json.Marshal(map[string]any{
		"action":   "explain",
		"provider": "offline",
	})
	postReq := httptest.NewRequest(http.MethodPost, "/api/tutor/requests", bytes.NewReader(reqBody))
	postReq.Host = "127.0.0.1"
	postRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/tutor/requests, got %d: %s", postRec.Code, postRec.Body.String())
	}

	var postData map[string]string
	if err := json.NewDecoder(postRec.Body).Decode(&postData); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	reqID := postData["request_id"]
	if reqID == "" {
		t.Fatalf("expected non-empty request_id")
	}

	// 2. DELETE /api/tutor/requests/{id} (cancellation)
	delReq := httptest.NewRequest(http.MethodDelete, "/api/tutor/requests/"+reqID, nil)
	delReq.Host = "127.0.0.1"
	delRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on cancel, got %d", delRec.Code)
	}

	var delData map[string]string
	_ = json.NewDecoder(delRec.Body).Decode(&delData)
	if delData["status"] != "cancelled" {
		t.Errorf("expected status cancelled, got %v", delData["status"])
	}
}

func TestNotesCRUDAndExportEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tmpDir, "notes_api_test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Store:   store,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	// 1. POST /api/notes
	notePayload, _ := json.Marshal(map[string]any{
		"id":           "note_api_01",
		"topic":        "Binomial Distribution",
		"raw_markdown": "Derivation: $P(X=2) = 0.375$",
		"provider_info": map[string]any{
			"title":    "Fair Coin Calculation",
			"concepts": []string{"binomial_pmf"},
		},
	})
	saveReq := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewReader(notePayload))
	saveReq.Host = "127.0.0.1"
	saveRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(saveRec, saveReq)

	if saveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on POST /api/notes, got %d: %s", saveRec.Code, saveRec.Body.String())
	}

	// 2. GET /api/notes
	listReq := httptest.NewRequest(http.MethodGet, "/api/notes?topic=binomial", nil)
	listReq.Host = "127.0.0.1"
	listRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/notes, got %d", listRec.Code)
	}

	var notesList []map[string]any
	if err := json.NewDecoder(listRec.Body).Decode(&notesList); err != nil {
		t.Fatalf("failed to decode notes list: %v", err)
	}
	if len(notesList) != 1 {
		t.Fatalf("expected 1 note in list, got %d", len(notesList))
	}

	// 3. GET /api/notes/note_api_01
	getReq := httptest.NewRequest(http.MethodGet, "/api/notes/note_api_01", nil)
	getReq.Host = "127.0.0.1"
	getRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET note by ID, got %d", getRec.Code)
	}

	// 4. GET /api/notes/note_api_01/export
	exportReq := httptest.NewRequest(http.MethodGet, "/api/notes/note_api_01/export", nil)
	exportReq.Host = "127.0.0.1"
	exportRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(exportRec, exportReq)

	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on export note, got %d", exportRec.Code)
	}
	mdExport := exportRec.Body.String()
	if !strings.Contains(mdExport, "Fair Coin Calculation") || !strings.Contains(mdExport, "$P(X=2) = 0.375$") {
		t.Errorf("export missing expected content: %s", mdExport)
	}

	// 5. POST /api/exports
	batchReqBody, _ := json.Marshal(map[string]any{
		"type":     "notes",
		"note_ids": []string{"note_api_01"},
	})
	batchReq := httptest.NewRequest(http.MethodPost, "/api/exports", bytes.NewReader(batchReqBody))
	batchReq.Host = "127.0.0.1"
	batchRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(batchRec, batchReq)

	if batchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on POST /api/exports, got %d", batchRec.Code)
	}

	// 6. DELETE /api/notes/note_api_01
	delReq := httptest.NewRequest(http.MethodDelete, "/api/notes/note_api_01", nil)
	delReq.Host = "127.0.0.1"
	delRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on DELETE note, got %d", delRec.Code)
	}

	// Verify gone
	getReq2 := httptest.NewRequest(http.MethodGet, "/api/notes/note_api_01", nil)
	getReq2.Host = "127.0.0.1"
	getRec2 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(getRec2, getReq2)
	if getRec2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", getRec2.Code)
	}
}

func TestTutorDraftRecoveryEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tmpDir, "drafts_api_test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:    "127.0.0.1:0",
		Store:   store,
		Version: "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer srv.Close()

	// 1. POST /api/tutor/drafts/test_draft_01
	payload, _ := json.Marshal(map[string]any{
		"context_json":  `{"stage":"prob"}`,
		"recovery_text": "Draft content in progress",
	})
	postReq := httptest.NewRequest(http.MethodPost, "/api/tutor/drafts/test_draft_01", bytes.NewReader(payload))
	postReq.Host = "127.0.0.1"
	postRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on save draft, got %d", postRec.Code)
	}

	// 2. GET /api/tutor/drafts/test_draft_01
	getReq := httptest.NewRequest(http.MethodGet, "/api/tutor/drafts/test_draft_01", nil)
	getReq.Host = "127.0.0.1"
	getRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on get draft, got %d", getRec.Code)
	}
	var draftData map[string]any
	_ = json.NewDecoder(getRec.Body).Decode(&draftData)
	if draftData["recovery_text"] != "Draft content in progress" {
		t.Errorf("unexpected draft content: %v", draftData["recovery_text"])
	}

	// 3. DELETE /api/tutor/drafts/test_draft_01
	delReq := httptest.NewRequest(http.MethodDelete, "/api/tutor/drafts/test_draft_01", nil)
	delReq.Host = "127.0.0.1"
	delRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on delete draft, got %d", delRec.Code)
	}
}
