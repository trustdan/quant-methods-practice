package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func TestArcadeScoresEndpoint(t *testing.T) {
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	if err := storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{
		Addr:  "127.0.0.1:0",
		Store: store,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	call := func(method, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "http://127.0.0.1/api/arcade/scores", bytes.NewBufferString(body))
		req.Host = "127.0.0.1"
		req.Header.Set("Origin", "http://127.0.0.1")
		rec := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(rec, req)
		return rec
	}

	// 1. Initial GET should return empty leaderboard and 0 high score
	rec := call(http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp arcadeScoresResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Scores) != 0 || resp.HighScore != 0 {
		t.Fatalf("expected empty scores, got len=%d highScore=%d", len(resp.Scores), resp.HighScore)
	}

	// 2. Reject demo scores
	demoReq := `{"is_demo": true, "score": 9000, "initials": "BOT"}`
	rec = call(http.MethodPost, demoReq)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for demo score, got %d", rec.Code)
	}

	// 3. Reject negative score
	negReq := `{"is_demo": false, "score": -50, "initials": "NEG"}`
	rec = call(http.MethodPost, negReq)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for negative score, got %d", rec.Code)
	}

	// 4. Reject invalid json
	rec = call(http.MethodPost, `not-json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid json, got %d", rec.Code)
	}

	// 5. Submit valid score
	validReq := `{"game_version":"v1.0","seed":1234,"is_demo":false,"duration_seconds":45.5,"sectors_cleared":2,"score":1500,"initials":"ACE"}`
	rec = call(http.MethodPost, validReq)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for valid score, got %d: %s", rec.Code, rec.Body.String())
	}

	// 6. GET should now return the submitted score
	rec = call(http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Scores) != 1 || resp.HighScore != 1500 {
		t.Fatalf("expected 1 score with highScore=1500, got len=%d highScore=%d", len(resp.Scores), resp.HighScore)
	}
	if resp.Scores[0].Initials != "ACE" || resp.Scores[0].Score != 1500 {
		t.Fatalf("unexpected score record: %+v", resp.Scores[0])
	}

	// 7. Method not allowed for DELETE
	rec = call(http.MethodDelete, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405 for DELETE, got %d", rec.Code)
	}
}
