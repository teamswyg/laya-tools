// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Private maintainer measurement. Fresh original inputs are not evaluation data.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

const (
	modelPath       = ".cache/live-claims-demo/model/claims.rsc"
	modelSHA        = "cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b"
	modelBytes      = 73988
	modelSteps      = 4240
	serverURL       = "http://127.0.0.1:8877"
	schema          = "riidolaya-live-claims-v1"
	checksumInitial = uint64(14695981039346656037)
)

// Authored only for this fixed resource protocol; no semantic target labels.
var inputs = [...]string{
	"Please list the next two checks for the cedar draft.",
	"The violet page is still being revised this afternoon.",
	"The paper lantern note has reached its final line.",
	"I have not started arranging the marble tokens yet.",
	"Does the quiet notebook need one more blue heading?",
	"지금 작은 은빛 메모의 문장을 정리하고 있어요.",
	"초록 봉투의 짧은 설명을 여기서 마무리했어요.",
	"다음에 읽을 노란 메모의 제목을 하나 제안해 주세요.",
}

type modelInfo struct {
	TrainingSteps            uint64 `json:"training_steps"`
	ArtifactSHA256           string `json:"artifact_sha256"`
	SemanticQualityQualified bool   `json:"semantic_quality_qualified"`
	Mode                     string `json:"mode"`
}
type status struct {
	Schema string    `json:"schema"`
	State  string    `json:"state"`
	Model  modelInfo `json:"model"`
	Limits struct {
		Workers int `json:"workers"`
	} `json:"limits"`
	Thresholds struct {
		ConfidenceFloor float64 `json:"confidence_floor"`
		MarginFloor     float64 `json:"margin_floor"`
		Temperature     float64 `json:"temperature"`
	} `json:"thresholds"`
}
type hintResponse struct {
	Schema      string                      `json:"schema"`
	State       string                      `json:"state"`
	Prediction  *statehintclaims.Prediction `json:"prediction"`
	InferenceUS int64                       `json:"inference_us"`
	InputBytes  int                         `json:"input_bytes"`
	Model       modelInfo                   `json:"model"`
}
type memory struct {
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	Sys          uint64 `json:"go_sys_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	NumGC        uint32 `json:"num_gc"`
	PauseTotalNS uint64 `json:"pause_total_ns"`
}
type distribution struct {
	MeanNS float64 `json:"mean_ns"`
	P50NS  int64   `json:"p50_ns"`
	P95NS  int64   `json:"p95_ns"`
	P99NS  int64   `json:"p99_ns"`
	MinNS  int64   `json:"min_ns"`
	MaxNS  int64   `json:"max_ns"`
}
type inputMeta struct {
	ID            int `json:"id"`
	UTF8Bytes     int `json:"utf8_bytes"`
	JSONBodyBytes int `json:"json_body_bytes"`
	MeasuredCalls int `json:"measured_calls"`
}
type report struct {
	Protocol                    string        `json:"protocol"`
	Arm                         string        `json:"arm"`
	Repeat                      int           `json:"repeat"`
	StartedUTC                  string        `json:"started_utc"`
	GoVersion                   string        `json:"go_version"`
	GOOS                        string        `json:"goos"`
	GOARCH                      string        `json:"goarch"`
	CGOEnabled                  string        `json:"cgo_enabled"`
	GOMAXPROCS                  int           `json:"gomaxprocs"`
	GOMEMLIMIT                  string        `json:"gomemlimit"`
	ModelSHA256                 string        `json:"model_sha256"`
	ModelBytes                  int           `json:"model_bytes"`
	TrainingSteps               uint64        `json:"training_steps"`
	ModelHashVerifiedBeforeLoad bool          `json:"model_hash_verified_before_load"`
	InputCorpusSHA256           string        `json:"input_corpus_sha256"`
	Inputs                      []inputMeta   `json:"inputs"`
	WarmupCalls                 int           `json:"warmup_calls"`
	MeasuredCalls               int           `json:"measured_calls"`
	TotalHintCalls              int           `json:"total_hint_calls"`
	CompletedWarmupCalls        int           `json:"completed_warmup_calls"`
	CompletedMeasuredCalls      int           `json:"completed_measured_calls"`
	HTTPStatusCalls             int           `json:"http_status_calls"`
	LoadNS                      int64         `json:"load_ns"`
	ElapsedNS                   int64         `json:"elapsed_ns"`
	CallsPerSecond              float64       `json:"calls_per_second"`
	Latency                     distribution  `json:"per_call_latency"`
	ServerInferenceReportedUS   *distribution `json:"server_inference_reported_us,omitempty"`
	PredictionChecksumHex       string        `json:"prediction_checksum_hex"`
	FirstCycleChecksumHex       string        `json:"first_cycle_checksum_hex"`
	ChecksumScope               string        `json:"checksum_scope"`
	BeforeLoad                  memory        `json:"before_load"`
	AfterLoadGC                 memory        `json:"after_load_gc"`
	WarmReadyGC                 memory        `json:"warm_ready_gc"`
	AfterCalls                  memory        `json:"after_calls"`
	RetainedAfterCallsGC        memory        `json:"retained_after_calls_gc"`
	TimedAllocatedBytes         uint64        `json:"timed_allocated_bytes"`
	TimedMallocs                uint64        `json:"timed_mallocs"`
	AllocatedBytesPerCall       float64       `json:"allocated_bytes_per_call"`
	MallocsPerCall              float64       `json:"mallocs_per_call"`
	TimedGCs                    uint32        `json:"timed_gcs"`
	TimedGCPauseNS              uint64        `json:"timed_gc_pause_ns"`
	ServerBefore                *status       `json:"server_before,omitempty"`
	ServerAfter                 *status       `json:"server_after,omitempty"`
	Error                       string        `json:"error,omitempty"`
}

func mem() memory {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return memory{m.HeapAlloc, m.HeapInuse, m.HeapSys, m.Sys, m.TotalAlloc, m.Mallocs, m.Frees, m.NumGC, m.PauseTotalNs}
}
func describe(xs []int64) distribution {
	var sum float64
	for _, x := range xs {
		sum += float64(x)
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	q := func(p float64) int64 { return xs[int(math.Ceil(float64(len(xs))*p))-1] }
	return distribution{sum / float64(len(xs)), q(.5), q(.95), q(.99), xs[0], xs[len(xs)-1]}
}
func mixByte(c uint64, b byte) uint64 { return (c ^ uint64(b)) * 1099511628211 }
func mixU64(c, v uint64) uint64 {
	for range 8 {
		c = mixByte(c, byte(v))
		v >>= 8
	}
	return c
}
func mixString(c uint64, s string) uint64 {
	c = mixU64(c, uint64(len(s)))
	for i := range len(s) {
		c = mixByte(c, s[i])
	}
	return c
}
func predictionChecksum(c uint64, p statehintclaims.Prediction) uint64 {
	c = mixU64(c, p.TrainingSteps)
	c = mixString(c, string(p.Source))
	for _, h := range p.Heads {
		c = mixString(c, h.Head)
		c = mixString(c, string(h.Winner))
		c = mixString(c, string(h.State))
		c = mixString(c, h.UnknownReason)
		for _, v := range h.Probabilities {
			c = mixU64(c, math.Float64bits(v))
		}
		c = mixU64(c, math.Float64bits(h.Confidence))
		c = mixU64(c, math.Float64bits(h.Margin))
	}
	return c
}
func verifiedModel() (*statehintclaims.Model, error) {
	f, err := os.Open(modelPath)
	if err != nil {
		return nil, errors.New("cannot open explicit model")
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil || !s.Mode().IsRegular() || s.Size() != modelBytes {
		return nil, errors.New("model size/type mismatch")
	}
	b, err := io.ReadAll(io.LimitReader(f, modelBytes+1))
	if err != nil || len(b) != modelBytes {
		return nil, errors.New("model read mismatch")
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != modelSHA {
		return nil, errors.New("model SHA mismatch before Load")
	}
	m, err := statehintclaims.Load(bytes.NewReader(b))
	if err != nil {
		return nil, errors.New("model Load failed")
	}
	if m.TrainingSteps() != modelSteps {
		return nil, errors.New("model training step mismatch")
	}
	return m, nil
}
func validateInfo(m modelInfo) bool {
	return m.ArtifactSHA256 == modelSHA && m.TrainingSteps == modelSteps && !m.SemanticQualityQualified && m.Mode == "research_preview"
}
func requestJSON(ctx context.Context, client *http.Client, method, path string, body []byte, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, serverURL+path, bytes.NewReader(body))
	if err != nil {
		return errors.New("cannot construct loopback request")
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.New("live HTTP instance unavailable or request failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("live HTTP status %d", res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 16385))
	if err != nil || len(b) > 16384 {
		return errors.New("live HTTP response budget exceeded")
	}
	if err := json.Unmarshal(b, result); err != nil {
		return errors.New("live HTTP JSON decode failed")
	}
	return nil
}
func checkedStatus(ctx context.Context, c *http.Client) (*status, error) {
	var s status
	if err := requestJSON(ctx, c, http.MethodGet, "/api/status", nil, &s); err != nil {
		return nil, err
	}
	if s.Schema != schema || s.State != "ready" || !validateInfo(s.Model) || s.Limits.Workers != 1 || s.Thresholds.ConfidenceFloor != .9 || s.Thresholds.MarginFloor != .05 || s.Thresholds.Temperature != 1 {
		return nil, errors.New("live server/model/protocol pin mismatch")
	}
	return &s, nil
}
func run(r *report) error {
	if runtime.Version() != "go1.27.1" || runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" || runtime.GOMAXPROCS(0) != 2 || os.Getenv("GOMEMLIMIT") != "128MiB" {
		return errors.New("runtime/environment pin mismatch")
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("missing Go build info")
	}
	for _, setting := range bi.Settings {
		if setting.Key == "CGO_ENABLED" {
			r.CGOEnabled = setting.Value
		}
	}
	if r.CGOEnabled != "0" {
		return errors.New("CPU-only pure-Go build pin mismatch")
	}
	n, warm := 30000, 128
	if r.Arm == "http" {
		n, warm = 600, 32
	} else if r.Arm != "inprocess" {
		return errors.New("unknown arm")
	}
	if r.Repeat < 1 || r.Repeat > 3 {
		return errors.New("repeat must be 1..3")
	}
	r.WarmupCalls = warm
	r.MeasuredCalls = n
	r.TotalHintCalls = n + warm
	body := make([][]byte, len(inputs))
	hash := sha256.New()
	r.Inputs = make([]inputMeta, len(inputs))
	for i, s := range inputs {
		var count [8]byte
		binary.LittleEndian.PutUint64(count[:], uint64(len(s)))
		hash.Write(count[:])
		hash.Write([]byte(s))
		b, err := json.Marshal(struct {
			Text string `json:"text"`
		}{s})
		if err != nil {
			return errors.New("original fixture JSON error")
		}
		body[i] = b
		r.Inputs[i] = inputMeta{i + 1, len(s), len(b), n / len(inputs)}
	}
	r.InputCorpusSHA256 = hex.EncodeToString(hash.Sum(nil))
	latencies := make([]int64, n)
	serverTimes := make([]int64, n)
	runtime.GC()
	r.BeforeLoad = mem()
	loadStart := time.Now()
	m, err := verifiedModel()
	if err != nil {
		return err
	}
	r.LoadNS = time.Since(loadStart).Nanoseconds()
	r.ModelHashVerifiedBeforeLoad = true
	runtime.GC()
	r.AfterLoadGC = mem()
	var workspace statehintclaims.Workspace
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	var c *http.Client
	if r.Arm == "http" {
		transport := &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext, MaxConnsPerHost: 1, MaxIdleConns: 1, MaxIdleConnsPerHost: 1, IdleConnTimeout: 10 * time.Second, DisableCompression: true, ForceAttemptHTTP2: false}
		defer transport.CloseIdleConnections()
		c = &http.Client{Transport: transport, Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect prohibited") }}
		r.ServerBefore, err = checkedStatus(ctx, c)
		r.HTTPStatusCalls++
		if err != nil {
			return err
		}
	}
	call := func(i int) (statehintclaims.Prediction, int64, error) {
		if r.Arm == "inprocess" {
			p, err := m.Predict(inputs[i%len(inputs)], &workspace)
			return p, 0, err
		}
		var out hintResponse
		if err := requestJSON(ctx, c, http.MethodPost, "/api/hints", body[i%len(inputs)], &out); err != nil {
			return statehintclaims.Prediction{}, 0, err
		}
		if out.Schema != schema || out.State != "ready" || out.Prediction == nil || out.InputBytes != len(inputs[i%len(inputs)]) || !validateInfo(out.Model) || out.Prediction.TrainingSteps != modelSteps || out.Prediction.Source != statehintclaims.Learned {
			return statehintclaims.Prediction{}, 0, errors.New("live hint output pin mismatch")
		}
		return *out.Prediction, out.InferenceUS, nil
	}
	for i := range warm {
		if ctx.Err() != nil {
			return errors.New("55-second arm budget exhausted during warm-up")
		}
		if _, _, err := call(i); err != nil {
			return err
		}
		r.CompletedWarmupCalls++
	}
	runtime.GC()
	r.WarmReadyGC = mem()
	checksum := checksumInitial
	firstCycleChecksum := checksumInitial
	var firstPredictions [len(inputs)]uint64
	start := time.Now()
	for i := range n {
		if ctx.Err() != nil {
			return errors.New("55-second arm budget exhausted during measured calls")
		}
		callStart := time.Now()
		p, serverUS, err := call(i)
		latencies[i] = time.Since(callStart).Nanoseconds()
		if err != nil {
			return err
		}
		r.CompletedMeasuredCalls++
		individual := predictionChecksum(checksumInitial, p)
		if i < len(inputs) {
			firstPredictions[i] = individual
			firstCycleChecksum = predictionChecksum(firstCycleChecksum, p)
		} else if firstPredictions[i%len(inputs)] != individual {
			return errors.New("repeated prediction checksum changed")
		}
		checksum = predictionChecksum(checksum, p)
		serverTimes[i] = serverUS
	}
	r.ElapsedNS = time.Since(start).Nanoseconds()
	r.AfterCalls = mem()
	r.TimedAllocatedBytes = r.AfterCalls.TotalAlloc - r.WarmReadyGC.TotalAlloc
	r.TimedMallocs = r.AfterCalls.Mallocs - r.WarmReadyGC.Mallocs
	r.TimedGCs = r.AfterCalls.NumGC - r.WarmReadyGC.NumGC
	r.TimedGCPauseNS = r.AfterCalls.PauseTotalNS - r.WarmReadyGC.PauseTotalNS
	r.AllocatedBytesPerCall = float64(r.TimedAllocatedBytes) / float64(n)
	r.MallocsPerCall = float64(r.TimedMallocs) / float64(n)
	r.CallsPerSecond = float64(n) * 1e9 / float64(r.ElapsedNS)
	r.PredictionChecksumHex = fmt.Sprintf("%016x", checksum)
	r.FirstCycleChecksumHex = fmt.Sprintf("%016x", firstCycleChecksum)
	if r.Arm == "http" {
		r.ServerAfter, err = checkedStatus(ctx, c)
		r.HTTPStatusCalls++
		if err != nil {
			return err
		}
	}
	runtime.GC()
	r.RetainedAfterCallsGC = mem()
	runtime.KeepAlive(m)
	runtime.KeepAlive(&workspace)
	runtime.KeepAlive(body)
	runtime.KeepAlive(latencies)
	runtime.KeepAlive(serverTimes)
	r.Latency = describe(latencies)
	if r.Arm == "http" {
		d := describe(serverTimes)
		r.ServerInferenceReportedUS = &d
	}
	return nil
}
func main() {
	arm := flag.String("arm", "", "inprocess or http")
	repeat := flag.Int("repeat", 0, "fixed repeat number 1..3")
	flag.Parse()
	r := report{Protocol: "live-claims-resource-probe-v1", Arm: *arm, Repeat: *repeat, StartedUTC: time.Now().UTC().Format(time.RFC3339Nano), GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GOMAXPROCS: runtime.GOMAXPROCS(0), GOMEMLIMIT: os.Getenv("GOMEMLIMIT"), ModelSHA256: modelSHA, ModelBytes: modelBytes, TrainingSteps: modelSteps, ChecksumScope: "FNV-1a framed numeric and decision prediction fields, measured calls only; not semantic labels"}
	err := run(&r)
	if err != nil {
		r.Error = err.Error()
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if enc.Encode(r) != nil {
		os.Exit(3)
	}
	if err != nil {
		os.Exit(2)
	}
}
