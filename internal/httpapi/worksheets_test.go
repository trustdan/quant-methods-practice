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
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/storage"
	"github.com/trustdan/quant-methods-practice/internal/worksheets"
)

func TestWorksheetsHTTPBoundaryAndDatasetReview(t *testing.T) {
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
	srv, err := NewServer(Config{Bank: b, Store: storage.NewStore(db, nil)})
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
	if w := call("GET", "/api/worksheets", "", "", nil); w.Code != 401 {
		t.Fatal("unauthenticated worksheet access")
	}
	w := call("POST", "/api/local-session", `{"bootstrap_token":"`+srv.BootstrapToken()+`"}`, "http://127.0.0.1", nil)
	cookie := w.Result().Cookies()[0]
	for _, origin := range []string{"", "http://127.0.0.1:9999", "https://evil.example"} {
		if w = call("POST", "/api/worksheets", `{}`, origin, cookie); w.Code != 403 {
			t.Fatalf("origin %s: %d", origin, w.Code)
		}
	}
	for _, body := range []string{`{"mode":"full_solution","template_id":"binomial_fair_coin_exactly_two","expected_answer":1}`, `{} {}`, strings.Repeat(" ", 128*1024) + `{}`} {
		if w = call("POST", "/api/worksheets", body, "http://127.0.0.1", cookie); w.Code != 400 {
			t.Fatal("invalid or unbounded JSON accepted")
		}
	}
	w = call("POST", "/api/worksheets", `{"mode":"full_solution","template_id":"binomial_fair_coin_exactly_two","seed":42}`, "http://127.0.0.1", cookie)
	if w.Code != 201 || strings.Contains(w.Body.String(), "expected_answer") || strings.Contains(w.Body.String(), "source_template") || strings.Contains(w.Body.String(), "hint_markdown") {
		t.Fatalf("bad public form %d: %s", w.Code, w.Body.String())
	}
	var view worksheets.View
	if err = json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	private, err := srv.worksheetStore.GetWorksheet(context.Background(), view.ID)
	if err != nil {
		t.Fatal(err)
	}
	answers := map[string]domain.SubmittedAnswer{}
	for _, item := range private.Items {
		expected := item.State.Instance.ExpectedAnswer
		answer := domain.SubmittedAnswer{Kind: item.State.Instance.Kind, OptionID: expected.OptionID}
		if expected.Value != nil {
			answer.NumericRaw = fmt.Sprintf("%.17g", *expected.Value)
		}
		answers[item.Key] = answer
	}
	command, _ := json.Marshal(worksheets.Command{ID: "http_atomic_submit", Revision: 1, Type: "submit", Answers: answers})
	w = call("GET", "/api/worksheets/"+view.ID+"/export", "", "", cookie)
	if w.Code != 200 || strings.Contains(w.Body.String(), "3/8") {
		t.Fatal("unearned export key")
	}
	if _, err = db.Exec(`CREATE TRIGGER fail_http_command BEFORE INSERT ON worksheet_commands BEGIN SELECT RAISE(ABORT,'private storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	commandPath := "/api/worksheets/" + view.ID + "/commands"
	w = call("POST", commandPath, string(command), "http://127.0.0.1", cookie)
	if w.Code != 500 || strings.Contains(w.Body.String(), "private storage failure") || !strings.Contains(w.Body.String(), "Retry") {
		t.Fatal("storage failure must retain the pending command and hide backend details")
	}
	if _, err = db.Exec(`DROP TRIGGER fail_http_command`); err != nil {
		t.Fatal(err)
	}
	w = call("POST", commandPath, string(command), "http://127.0.0.1", cookie)
	if w.Code != 200 {
		t.Fatalf("pending retry failed: %d %s", w.Code, w.Body.String())
	}
	firstResult := w.Body.String()
	w = call("POST", commandPath, string(command), "http://127.0.0.1", cookie)
	if w.Code != 200 || w.Body.String() != firstResult {
		t.Fatal("HTTP command replay changed result")
	}
	body := `{"mode":"dataset","csv":"experiment_id,successes\na,2\nb,0\n","reviewer":"human","source_note":"synthetic rows = four tosses"}`
	if w = call("POST", "/api/worksheets", body, "http://127.0.0.1", cookie); w.Code != 400 {
		t.Fatal("review bypassed")
	}
	w = call("POST", "/api/worksheets/dataset-preview", `{"csv":"experiment_id,successes\na,2\nb,0\n"}`, "http://127.0.0.1", cookie)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "expected_answer") {
		t.Fatal("author preview must expose keys for semantic review")
	}
	w = call("POST", "/api/worksheets", strings.TrimSuffix(body, "}")+`,"reviewed":true}`, "http://127.0.0.1", cookie)
	if w.Code != 201 || strings.Contains(w.Body.String(), "expected_answer") || !strings.Contains(w.Body.String(), "answer keys were shown") {
		t.Fatalf("reviewed case not created safely: %d %s", w.Code, w.Body.String())
	}
	if b.Count() != 10 {
		t.Fatal("case generation modified active bank")
	}
}
