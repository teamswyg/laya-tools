// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package claimsdemo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

// These fixtures test HTTP transport and scheduling only. They do not measure
// research-model quality, and do not train or imitate a native learned model.
type fixturePredictor struct {
	calls      atomic.Int64
	active     atomic.Int64
	peak       atomic.Int64
	started    chan struct{}
	release    <-chan struct{}
	failure    error
	mu         sync.Mutex
	workspaces map[*statehintclaims.Workspace]bool
}

func (p *fixturePredictor) Predict(_ string, workspace *statehintclaims.Workspace) (statehintclaims.Prediction, error) {
	p.calls.Add(1)
	active := p.active.Add(1)
	defer p.active.Add(-1)
	for current := p.peak.Load(); active > current; current = p.peak.Load() {
		if p.peak.CompareAndSwap(current, active) {
			break
		}
	}
	p.mu.Lock()
	if p.workspaces == nil {
		p.workspaces = make(map[*statehintclaims.Workspace]bool)
	}
	p.workspaces[workspace] = true
	p.mu.Unlock()
	if p.started != nil {
		select {
		case p.started <- struct{}{}:
		default:
		}
	}
	if p.release != nil {
		<-p.release
	}
	return fixturePrediction(), p.failure
}

func fixturePrediction() statehintclaims.Prediction {
	p := statehintclaims.Prediction{Source: statehintclaims.Learned, TrainingSteps: 7}
	for i, head := range statehintclaims.Heads() {
		p.Heads[i] = statehintclaims.HeadPrediction{
			Head: head.String(), Winner: statehintclaims.True, State: statehintclaims.Unknown,
			Probabilities: [3]float64{0.6, 0.1, 0.3}, Confidence: 0.6, Margin: 0.3, UnknownReason: "low_confidence",
		}
	}
	return p
}

func fixtureHandler(p *fixturePredictor) *handler {
	return newHandler(p, ModelInfo{Name: "HTTP fixture; not a research model", TrainingSteps: 7, ArtifactSHA256: strings.Repeat("a", 64)})
}

func jsonBody(text string) string {
	b, _ := json.Marshal(map[string]string{"text": text})
	return string(b)
}

func request(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://127.0.0.1:8877"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHintsPreserveCandidateAbstentionAndDoNotEchoText(t *testing.T) {
	p := &fixturePredictor{}
	h := fixtureHandler(p)
	const synthetic = "Please reply to the original sample."
	w := request(h, http.MethodPost, "/api/hints", jsonBody(synthetic))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var result Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Schema != Schema || result.State != "ready" || result.InputBytes != len(synthetic) || result.InferenceUS < 0 || result.Prediction == nil || p.calls.Load() != 1 {
		t.Fatalf("unexpected response: %+v; calls=%d", result, p.calls.Load())
	}
	for i, head := range result.Prediction.Heads {
		if head.Head != statehintclaims.Heads()[i].String() || head.Winner != statehintclaims.True || head.State != statehintclaims.Unknown || head.UnknownReason != "low_confidence" || head.Probabilities != [3]float64{0.6, 0.1, 0.3} {
			t.Fatalf("candidate or abstention was altered: %+v", head)
		}
	}
	if result.Model.SemanticQualityQualified || result.Model.Mode != "research_preview" || strings.Contains(w.Body.String(), synthetic) {
		t.Fatal("response asserted quality or echoed input")
	}
}

func TestEmptyInputResetsWithoutInference(t *testing.T) {
	p := &fixturePredictor{}
	h := fixtureHandler(p)
	for _, text := range []string{"", " \t\r\n", "\u2003\u3000"} {
		w := request(h, http.MethodPost, "/api/hints", jsonBody(text))
		var result Response
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.State != "empty" || result.Prediction != nil || result.InferenceUS != 0 || result.InputBytes != len(text) {
			t.Fatalf("empty reset failed: %d %s", w.Code, w.Body.String())
		}
	}
	if p.calls.Load() != 0 {
		t.Fatal("empty input invoked predictor")
	}
}

func TestMalformedAndOversizedInputNeverInvokesModel(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
	}{
		{"empty body", "", 400},
		{"missing text", `{}`, 400},
		{"wrong top-level", `[]`, 400},
		{"null text", `{"text":null}`, 400},
		{"numeric text", `{"text":17}`, 400},
		{"nested text", `{"text":{"text":"sample"}}`, 400},
		{"unknown field", `{"text":"sample","extra":true}`, 400},
		{"duplicate field", `{"text":"first","text":"second"}`, 400},
		{"escaped duplicate field", `{"text":"first","te\u0078t":"second"}`, 400},
		{"trailing object", `{"text":"sample"}{}`, 400},
		{"trailing scalar", `{"text":"sample"} true`, 400},
		{"trailing garbage", `{"text":"sample"} junk`, 400},
		{"trailing comma", `{"text":"sample",}`, 400},
		{"NUL", `{"text":"sample\u0000"}`, 400},
		{"invalid UTF-8", string([]byte{'{', '"', 't', 'e', 'x', 't', '"', ':', '"', 0xff, '"', '}'}), 400},
		{"text byte limit", jsonBody(strings.Repeat("x", MaxTextBytes+1)), 413},
		{"multibyte byte limit", jsonBody(strings.Repeat("한", 1366)), 413},
		{"JSON body limit", strings.Repeat(" ", MaxRequestBytes+1), 413},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &fixturePredictor{}
			w := request(fixtureHandler(p), http.MethodPost, "/api/hints", tt.body)
			if w.Code != tt.status || p.calls.Load() != 0 {
				t.Fatalf("status=%d calls=%d body=%s", w.Code, p.calls.Load(), w.Body.String())
			}
			if w.Header().Get("Content-Type") != "application/json; charset=utf-8" || !json.Valid(w.Body.Bytes()) {
				t.Fatal("error did not return bounded JSON")
			}
		})
	}
}

func TestExactUTF8ByteLimitIsAccepted(t *testing.T) {
	p := &fixturePredictor{}
	text := strings.Repeat("한", 1365) + "x"
	if len(text) != MaxTextBytes {
		t.Fatal("fixture is not at the byte boundary")
	}
	w := request(fixtureHandler(p), http.MethodPost, "/api/hints", jsonBody(text))
	if w.Code != http.StatusOK || p.calls.Load() != 1 {
		t.Fatalf("exact byte limit rejected: %d", w.Code)
	}
}

func TestOriginMethodsMediaTypeAndSecurity(t *testing.T) {
	tests := []struct {
		name, method, path, host, origin, contentType, encoding string
		status                                                  int
	}{
		{name: "same origin", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", origin: "http://127.0.0.1:8877", contentType: "application/json", status: 200},
		{name: "foreign origin", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", origin: "https://example.invalid", contentType: "application/json", status: 403},
		{name: "foreign local port", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", origin: "http://127.0.0.1:9000", contentType: "application/json", status: 403},
		{name: "opaque origin", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", origin: "null", contentType: "application/json", status: 403},
		{name: "rebound host", method: "POST", path: "/api/hints", host: "example.invalid:8877", contentType: "application/json", status: 403},
		{name: "GET hints", method: "GET", path: "/api/hints", host: "127.0.0.1:8877", status: 405},
		{name: "preflight", method: "OPTIONS", path: "/api/hints", host: "127.0.0.1:8877", status: 405},
		{name: "POST status", method: "POST", path: "/api/status", host: "127.0.0.1:8877", contentType: "application/json", status: 405},
		{name: "text form", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", contentType: "text/plain", status: 415},
		{name: "missing media type", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", status: 415},
		{name: "UTF-8 media type", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", contentType: "application/json; charset=utf-8", status: 200},
		{name: "wrong charset", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", contentType: "application/json; charset=latin1", status: 415},
		{name: "compressed request", method: "POST", path: "/api/hints", host: "127.0.0.1:8877", contentType: "application/json", encoding: "gzip", status: 415},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &fixturePredictor{}
			r := httptest.NewRequest(tt.method, "http://127.0.0.1:8877"+tt.path, strings.NewReader(jsonBody("Original transport fixture.")))
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}
			if tt.encoding != "" {
				r.Header.Set("Content-Encoding", tt.encoding)
			}
			w := httptest.NewRecorder()
			fixtureHandler(p).ServeHTTP(w, r)
			if w.Code != tt.status || (tt.status != 200 && p.calls.Load() != 0) {
				t.Fatalf("status=%d calls=%d", w.Code, p.calls.Load())
			}
			for key, expected := range map[string]string{"Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "Referrer-Policy": "no-referrer", "X-Frame-Options": "DENY"} {
				if w.Header().Get(key) != expected {
					t.Fatalf("missing %s", key)
				}
			}
			if w.Header().Get("Access-Control-Allow-Origin") != "" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src 'self'") || strings.Contains(w.Header().Get("Content-Security-Policy"), "unsafe-inline") {
				t.Fatal("unexpected CORS or content security policy")
			}
		})
	}
}

func TestStatusAndEmbeddedAssets(t *testing.T) {
	p := &fixturePredictor{}
	h := fixtureHandler(p)
	w := request(h, http.MethodGet, "/api/status", "")
	var status Status
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &status) != nil || status.Schema != Schema || status.State != "ready" || status.Limits.MaxTextBytes != 4096 || status.Limits.MaxRequestBytes != 32768 || status.Limits.Workers != 1 || status.Thresholds.ConfidenceFloor != .9 || status.Thresholds.MarginFloor != .05 || status.Thresholds.Temperature != 1 || status.Model.Mode != "research_preview" || status.Model.SemanticQualityQualified {
		t.Fatalf("unexpected status: %s", w.Body.String())
	}
	for path, contentType := range map[string]string{"/": "text/html; charset=utf-8", "/app.js": "text/javascript; charset=utf-8", "/style.css": "text/css; charset=utf-8"} {
		get := request(h, http.MethodGet, path, "")
		head := request(h, http.MethodHead, path, "")
		if get.Code != 200 || get.Body.Len() == 0 || get.Header().Get("Content-Type") != contentType || head.Code != 200 || head.Body.Len() != 0 {
			t.Fatalf("asset or HEAD failed: %s", path)
		}
	}
	if request(h, http.MethodGet, "/missing-file", "").Code != 404 || p.calls.Load() != 0 {
		t.Fatal("static/status routing invoked model or wrong unknown route")
	}
}

func TestConcurrentRequestsSerializeOneWorkspace(t *testing.T) {
	release := make(chan struct{})
	p := &fixturePredictor{started: make(chan struct{}, 1), release: release}
	h := fixtureHandler(p)
	const count = 12
	results := make(chan int, count)
	var wg sync.WaitGroup
	for range count {
		wg.Go(func() {
			results <- request(h, http.MethodPost, "/api/hints", jsonBody("Concurrent original fixture.")).Code
		})
	}
	select {
	case <-p.started:
	case <-time.After(3 * time.Second):
		t.Fatal("first inference did not start")
	}
	close(release)
	wg.Wait()
	close(results)
	for status := range results {
		if status != 200 {
			t.Fatalf("concurrent status %d", status)
		}
	}
	if p.calls.Load() != count || p.peak.Load() != 1 || len(p.workspaces) != 1 {
		t.Fatalf("inference was not serialized: calls=%d active_peak=%d workspaces=%d", p.calls.Load(), p.peak.Load(), len(p.workspaces))
	}
}

func TestCanceledWaitingRequestDoesNotInvokeModel(t *testing.T) {
	release := make(chan struct{})
	p := &fixturePredictor{started: make(chan struct{}, 1), release: release}
	h := fixtureHandler(p)
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		request(h, http.MethodPost, "/api/hints", jsonBody("First original fixture."))
	}()
	select {
	case <-p.started:
	case <-time.After(3 * time.Second):
		t.Fatal("first inference did not start")
	}
	ctx, cancel := context.WithCancel(context.Background())
	reader := &readNotification{Reader: strings.NewReader(jsonBody("Canceled original fixture.")), done: make(chan struct{})}
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8877/api/hints", reader).WithContext(ctx)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	secondDone := make(chan struct{})
	go func() { defer close(secondDone); h.ServeHTTP(w, r) }()
	select {
	case <-reader.done:
	case <-time.After(3 * time.Second):
		t.Fatal("second request body was not read")
	}
	cancel()
	close(release)
	<-firstDone
	<-secondDone
	if p.calls.Load() != 1 || w.Code != http.StatusRequestTimeout {
		t.Fatalf("canceled waiter invoked model: calls=%d status=%d", p.calls.Load(), w.Code)
	}
}

type readNotification struct {
	*strings.Reader
	done chan struct{}
	once sync.Once
}

func (r *readNotification) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if r.Reader.Len() == 0 {
		r.once.Do(func() { close(r.done) })
	}
	return n, err
}

func TestInferenceErrorsDoNotLeakTextOrUnderlyingDetails(t *testing.T) {
	const privateLikeSynthetic = "Original private-like fixture for transport only."
	p := &fixturePredictor{failure: errors.New(privateLikeSynthetic)}
	w := request(fixtureHandler(p), http.MethodPost, "/api/hints", jsonBody(privateLikeSynthetic))
	if w.Code != 500 || bytes.Contains(w.Body.Bytes(), []byte(privateLikeSynthetic)) || !strings.Contains(w.Body.String(), "inference_failed") {
		t.Fatal("inference error exposed details or returned wrong status")
	}
}
