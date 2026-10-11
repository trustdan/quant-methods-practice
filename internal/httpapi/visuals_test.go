package httpapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVisualsNormalEndpoint(t *testing.T) {
	srv := newTestServer(t)

	// 1. Standard Normal lower tail at x=0
	req := httptest.NewRequest(http.MethodGet, "/api/visuals/normal?mu=0&sigma=1&tail=lower&x=0", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var normResp NormalVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &normResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if normResp.Family != "normal" || normResp.Mu != 0 || normResp.Sigma != 1 {
		t.Errorf("unexpected distribution metadata: %+v", normResp)
	}
	if normResp.ShadedProbability != 0.5 {
		t.Errorf("expected shaded probability 0.5, got %v", normResp.ShadedProbability)
	}
	if len(normResp.Points) == 0 {
		t.Errorf("expected curve points, got none")
	}

	// 2. Interval between -1.96 and 1.96
	reqInt := httptest.NewRequest(http.MethodGet, "/api/visuals/normal?mu=0&sigma=1&tail=interval&x1=-1.96&x2=1.96", nil)
	reqInt.Host = "127.0.0.1"
	wInt := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wInt, reqInt)

	if wInt.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wInt.Code)
	}
	var intResp NormalVisualResponse
	_ = json.Unmarshal(wInt.Body.Bytes(), &intResp)
	if intResp.ShadedProbability < 0.949 || intResp.ShadedProbability > 0.951 {
		t.Errorf("expected interval probability approx 0.95, got %v", intResp.ShadedProbability)
	}

	// 3. Invalid sigma
	reqBad := httptest.NewRequest(http.MethodGet, "/api/visuals/normal?mu=0&sigma=-2", nil)
	reqBad.Host = "127.0.0.1"
	wBad := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative sigma, got %d", wBad.Code)
	}
}

func TestVisualsUniformEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/uniform?a=10&b=20&tail=interval&x1=12&x2=17", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var uniResp UniformVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &uniResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if math.Abs(uniResp.ShadedProbability-0.5) > 1e-9 {
		t.Errorf("expected shaded probability 0.5, got %v", uniResp.ShadedProbability)
	}
	if math.Abs(uniResp.HeightDensity-0.1) > 1e-9 {
		t.Errorf("expected height density 0.1, got %v", uniResp.HeightDensity)
	}

	// Invalid bounds: a >= b
	reqBad := httptest.NewRequest(http.MethodGet, "/api/visuals/uniform?a=20&b=10", nil)
	reqBad.Host = "127.0.0.1"
	wBad := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid bounds, got %d", wBad.Code)
	}
}

func TestVisualsExponentialEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/exponential?lambda=2&tail=upper&x=1", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var expResp ExponentialVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &expResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if expResp.Lambda != 2 || expResp.Mean != 0.5 {
		t.Errorf("unexpected exponential params: lambda=%v, mean=%v", expResp.Lambda, expResp.Mean)
	}
	if expResp.MemorylessNote == "" {
		t.Errorf("expected memoryless note in response")
	}

	// Invalid lambda <= 0
	reqBad := httptest.NewRequest(http.MethodGet, "/api/visuals/exponential?lambda=0", nil)
	reqBad.Host = "127.0.0.1"
	wBad := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero lambda, got %d", wBad.Code)
	}
}

func TestVisualsDiscreteEndpoint(t *testing.T) {
	srv := newTestServer(t)

	// Binomial
	reqBinom := httptest.NewRequest(http.MethodGet, "/api/visuals/discrete?family=binomial&n=4&p=0.5&event=equal&k=2", nil)
	reqBinom.Host = "127.0.0.1"
	wBinom := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBinom, reqBinom)

	if wBinom.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wBinom.Code)
	}
	var binomResp DiscreteVisualResponse
	_ = json.Unmarshal(wBinom.Body.Bytes(), &binomResp)
	if binomResp.TotalProbability != 0.375 {
		t.Errorf("expected total probability 0.375 for binomial n=4, p=0.5, k=2, got %v", binomResp.TotalProbability)
	}

	// Poisson
	reqPoiss := httptest.NewRequest(http.MethodGet, "/api/visuals/discrete?family=poisson&lambda=3&event=at_most&k=2", nil)
	reqPoiss.Host = "127.0.0.1"
	wPoiss := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wPoiss, reqPoiss)

	if wPoiss.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wPoiss.Code)
	}
	var poissResp DiscreteVisualResponse
	_ = json.Unmarshal(wPoiss.Body.Bytes(), &poissResp)
	if len(poissResp.Bars) == 0 {
		t.Errorf("expected bars for poisson")
	}
}

func TestVisualsVennEndpoint(t *testing.T) {
	srv := newTestServer(t)

	// Valid Venn
	req := httptest.NewRequest(http.MethodGet, "/api/visuals/venn?p_a=0.6&p_b=0.5&p_intersection=0.3&region=union", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var vennResp VennVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &vennResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if vennResp.ShadedProbability != 0.8 {
		t.Errorf("expected union probability 0.8, got %v", vennResp.ShadedProbability)
	}

	// Frechet inequality violation
	reqBad := httptest.NewRequest(http.MethodGet, "/api/visuals/venn?p_a=0.7&p_b=0.6&p_intersection=0.1", nil)
	reqBad.Host = "127.0.0.1"
	wBad := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for Frechet violation, got %d", wBad.Code)
	}

	// Method not allowed
	reqPost := httptest.NewRequest(http.MethodPost, "/api/visuals/normal", nil)
	reqPost.Host = "127.0.0.1"
	wPost := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST to visuals endpoint, got %d", wPost.Code)
	}
}

func TestVisualsSamplingEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/sampling?pop_dist=skewed&n=25&trials=300&seed=123", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SamplingVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Family != "sampling" || resp.SampleSize != 25 || resp.Trials != 300 {
		t.Errorf("unexpected metadata: %+v", resp)
	}
	if len(resp.Bars) == 0 || len(resp.NormalCurve) == 0 {
		t.Errorf("expected bars and normal curve points")
	}
	if resp.CLTNote == "" {
		t.Errorf("expected CLT note")
	}
}

func TestVisualsCIEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/confidence_interval?conf_level=0.95&mu=50&sigma=10&n=16&num_intervals=20&seed=42", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ConfidenceIntervalVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Family != "confidence_interval" || resp.ConfidenceLevel != 0.95 || len(resp.Intervals) != 20 {
		t.Errorf("unexpected CI response: %+v", resp)
	}
	if resp.CapturedCount == 0 {
		t.Errorf("expected at least some intervals to capture truth")
	}
}

func TestVisualsHypothesisEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/hypothesis_test?test_type=mean_t&tail=two_tailed&alpha=0.05&null_mu=50&x_bar=54&s=8&n=25", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp HypothesisVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Family != "hypothesis_test" || resp.TestType != "mean_t" {
		t.Errorf("unexpected hypothesis response: %+v", resp)
	}
	if len(resp.Points) == 0 {
		t.Errorf("expected curve points")
	}
	if resp.Decision == "" || resp.RejectionRegion == "" {
		t.Errorf("expected decision and rejection region")
	}
}

func TestVisualsRegressionEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/visuals/regression?dataset=advertising_sales&x_pred=5.0", nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp RegressionVisualResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Family != "regression" || resp.N != 8 {
		t.Errorf("unexpected regression metadata: %+v", resp)
	}
	if resp.PredictedX == nil || *resp.PredictedX != 5.0 || resp.PredictedY == nil {
		t.Errorf("expected prediction at x = 5.0, got: px=%v, py=%v", resp.PredictedX, resp.PredictedY)
	}
	if resp.AssociationNote == "" {
		t.Errorf("expected association note")
	}
}
