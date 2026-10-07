// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package claimsdemo serves a loopback-only preview of three attributed claim hints.
package claimsdemo

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

const (
	Schema          = "riidolaya-live-claims-v1"
	MaxRequestBytes = 32768
	MaxTextBytes    = statehintclaims.MaxTextBytes
)

//go:embed web/*
var webFiles embed.FS

type ModelInfo struct {
	Name                     string `json:"name"`
	TrainingSteps            uint64 `json:"training_steps"`
	ArtifactSHA256           string `json:"artifact_sha256"`
	SemanticQualityQualified bool   `json:"semantic_quality_qualified"`
	Mode                     string `json:"mode"`
}

type Limits struct {
	MaxTextBytes    int `json:"max_text_bytes"`
	MaxRequestBytes int `json:"max_request_bytes"`
	Workers         int `json:"workers"`
}

type Thresholds struct {
	ConfidenceFloor float64 `json:"confidence_floor"`
	MarginFloor     float64 `json:"margin_floor"`
	Temperature     float64 `json:"temperature"`
}

type Status struct {
	Schema     string     `json:"schema"`
	State      string     `json:"state"`
	Model      ModelInfo  `json:"model"`
	Limits     Limits     `json:"limits"`
	Thresholds Thresholds `json:"thresholds"`
}

type Response struct {
	Schema      string                      `json:"schema"`
	State       string                      `json:"state"`
	Prediction  *statehintclaims.Prediction `json:"prediction"`
	InferenceUS int64                       `json:"inference_us"`
	InputBytes  int                         `json:"input_bytes"`
	Model       ModelInfo                   `json:"model"`
}

type predictor interface {
	Predict(string, *statehintclaims.Workspace) (statehintclaims.Prediction, error)
}

type handler struct {
	predictor predictor
	model     ModelInfo
	mux       *http.ServeMux
	mu        sync.Mutex
	workspace statehintclaims.Workspace
}

// NewHandler uses an immutable, already-loaded model. The one reusable workspace
// is protected by a mutex so concurrent requests never share active feature state.
// The server never trains or changes the model's fixed numeric thresholds.
func NewHandler(model *statehintclaims.Model, artifactSHA256 string) http.Handler {
	return newHandler(model, ModelInfo{
		Name:           "three-claim research model",
		TrainingSteps:  model.TrainingSteps(),
		ArtifactSHA256: artifactSHA256,
	})
}

func newHandler(p predictor, model ModelInfo) *handler {
	model.SemanticQualityQualified = false
	model.Mode = "research_preview"
	h := &handler{predictor: p, model: model, mux: http.NewServeMux()}
	h.mux.HandleFunc("/api/hints", h.hints)
	h.mux.HandleFunc("/api/status", h.status)
	h.mux.HandleFunc("/", h.static)
	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	// Validate Host even without Origin to exclude DNS rebinding to this local port.
	if !loopbackAuthority(r.Host) || !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "origin_rejected", "Use this demo from its local URL.")
		return
	}
	h.mux.ServeHTTP(w, r)
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	return host == "127.0.0.1" || host == "::1"
}

func loopbackAuthority(authority string) bool {
	if host, port, err := net.SplitHostPort(authority); err == nil {
		return loopbackHost(host) && validPort(port)
	}
	return loopbackHost(authority) && !strings.Contains(authority, ":")
}

func sameOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}
	if len(origins) != 1 {
		return false
	}
	u, err := url.Parse(origins[0])
	return err == nil && u.Scheme == "http" && u.User == nil && u.Path == "" && u.RawQuery == "" && u.Fragment == "" && loopbackAuthority(u.Host) && strings.EqualFold(u.Host, r.Host)
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Use GET for model status.")
		return
	}
	writeJSON(w, http.StatusOK, Status{
		Schema: Schema, State: "ready", Model: h.model,
		Limits:     Limits{MaxTextBytes: MaxTextBytes, MaxRequestBytes: MaxRequestBytes, Workers: 1},
		Thresholds: Thresholds{ConfidenceFloor: statehintclaims.ConfidenceFloor, MarginFloor: statehintclaims.MarginFloor, Temperature: 1},
	})
}

func (h *handler) hints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Use POST for claim hints.")
		return
	}
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		writeError(w, http.StatusUnsupportedMediaType, "json_required", "Send UTF-8 application/json.")
		return
	}
	if r.Header.Get("Content-Encoding") != "" {
		writeError(w, http.StatusUnsupportedMediaType, "encoding_rejected", "Send an uncompressed JSON body.")
		return
	}
	if r.Context().Err() != nil {
		writeError(w, http.StatusRequestTimeout, "request_canceled", "The request was canceled.")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "The JSON body exceeds 32768 bytes.")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_json", "Cannot read the JSON body.")
		}
		return
	}
	input, err := decodeText(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Send exactly one JSON object with one string field named text.")
		return
	}
	if !utf8.ValidString(input) || strings.ContainsRune(input, 0) {
		writeError(w, http.StatusBadRequest, "invalid_text", "Text must be valid UTF-8 without NUL characters.")
		return
	}
	if len(input) > MaxTextBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "text_too_large", "Text exceeds the 4096-byte UTF-8 limit; shorten it to continue.")
		return
	}
	response := Response{Schema: Schema, State: "empty", InputBytes: len(input), Model: h.model}
	if strings.TrimSpace(input) == "" {
		writeJSON(w, http.StatusOK, response)
		return
	}
	h.mu.Lock()
	// A canceled request waiting for the workspace must not run an inference.
	if r.Context().Err() != nil {
		h.mu.Unlock()
		writeError(w, http.StatusRequestTimeout, "request_canceled", "The request was canceled.")
		return
	}
	start := time.Now()
	prediction, err := h.predictor.Predict(input, &h.workspace)
	response.InferenceUS = time.Since(start).Microseconds()
	h.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "inference_failed", "The model could not produce claim hints.")
		return
	}
	response.State = "ready"
	response.Prediction = &prediction
	writeJSON(w, http.StatusOK, response)
}

func decodeText(body []byte) (string, error) {
	if !utf8.Valid(body) {
		return "", errors.New("invalid UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(body))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return "", errors.New("expected object")
	}
	var text string
	seen := false
	for d.More() {
		key, err := d.Token()
		if err != nil || key != "text" || seen {
			return "", errors.New("unexpected or duplicate field")
		}
		value, err := d.Token()
		if err != nil {
			return "", errors.New("invalid text value")
		}
		var ok bool
		text, ok = value.(string)
		if !ok {
			return "", errors.New("text must be a string")
		}
		seen = true
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') || !seen {
		return "", errors.New("incomplete object")
	}
	if _, err := d.Token(); err != io.EOF {
		return "", errors.New("trailing JSON value")
	}
	return text, nil
}

func (h *handler) static(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Use GET or HEAD for demo files.")
		return
	}
	var name, contentType string
	switch r.URL.Path {
	case "/":
		name, contentType = "index.html", "text/html; charset=utf-8"
	case "/app.js":
		name, contentType = "app.js", "text/javascript; charset=utf-8"
	case "/style.css":
		name, contentType = "style.css", "text/css; charset=utf-8"
	default:
		writeError(w, http.StatusNotFound, "not_found", "This demo has no resource at that path.")
		return
	}
	data, err := webFiles.ReadFile("web/" + name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "asset_unavailable", "The demo page is unavailable.")
		return
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Schema string `json:"schema"`
		Error  struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{Schema: Schema, Error: struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		status = http.StatusInternalServerError
		data = []byte(`{"schema":"riidolaya-live-claims-v1","error":{"code":"response_failed","message":"The response could not be encoded."}}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(append(data, '\n'))
}
