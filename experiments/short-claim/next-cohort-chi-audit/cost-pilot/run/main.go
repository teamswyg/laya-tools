// SPDX-License-Identifier: Apache-2.0
// One exposed-parent cost audit. Explicit inactive local asset only; no training.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func uniqueJSON(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var walk func() error
	walk = func() error {
		t, e := d.Token()
		if e != nil {
			return e
		}
		v, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		seen := map[string]bool{}
		for d.More() {
			if v == '{' {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || seen[s] {
					return errors.New("duplicate_json_key")
				}
				seen[s] = true
			}
			if e = walk(); e != nil {
				return e
			}
		}
		_, e = d.Token()
		return e
	}
	if walk() != nil {
		return errors.New("invalid_json")
	}
	if _, e := d.Token(); e != io.EOF {
		return errors.New("trailing_json")
	}
	return nil
}
func required(x any, t reflect.Type) error {
	if t == reflect.TypeOf(json.RawMessage{}) || t == reflect.TypeOf([]byte{}) {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		if x == nil {
			return nil
		}
		return required(x, t.Elem())
	}
	if x == nil && t.Kind() != reflect.Slice && t.Kind() != reflect.Map && t.Kind() != reflect.Interface {
		return errors.New("null_value")
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := x.(map[string]any)
		if !ok {
			return errors.New("object_required")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")
			name := tag[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			v, ok := m[name]
			if !ok {
				if len(tag) > 1 && tag[1] == "omitempty" {
					continue
				}
				return errors.New("field_missing")
			}
			if e := required(v, f.Type); e != nil {
				return e
			}
		}
	case reflect.Array, reflect.Slice:
		if x == nil && t.Kind() == reflect.Slice {
			return nil
		}
		a, ok := x.([]any)
		if !ok || (t.Kind() == reflect.Array && len(a) != t.Len()) {
			return errors.New("array_width")
		}
		for _, v := range a {
			if e := required(v, t.Elem()); e != nil {
				return e
			}
		}
	}
	return nil
}

const rawCap = 2 << 20
const outputCap = 262144
const bundleCap = 8 << 20
const modelBytes = 32792
const modelSHA = "dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004"
const modelRef = "JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46"

var orders = [4][4]string{{"bm25", "fixed", "model", "lexical"}, {"fixed", "lexical", "bm25", "model"}, {"lexical", "model", "fixed", "bm25"}, {"model", "bm25", "lexical", "fixed"}}

//go:embed main.go main_test.go
var ownSource embed.FS

type rankValue struct {
	Kind           string     `json:"kind"`
	Order          [8]int     `json:"order"`
	Scores         [8]float64 `json:"scores"`
	Count          int        `json:"count"`
	FallbackReason string     `json:"fallback_reason,omitempty"`
}

type pin struct {
	name string
	size int
	sha  string
	dest string
}

var pins = [...]pin{
	{"source/chi/chain.go.txt", 1517, "8c22d7bbc23f4b4d46ded5ef721a9c3a173031fa2c2c9a7b48c46c2b07e8bd80", "chi/chain.go"},
	{"source/chi/chi.go.txt", 4812, "47c70ececcbb9d71f973eda3cbadad0a46c8cc2261b285f7049b5261f337d678", "chi/chi.go"},
	{"source/chi/context.go.txt", 5788, "b19edcca252e2fe74e82802c4c7ce1a1c0855728f9ede1288f0224283dfc7e53", "chi/context.go"},
	{"source/chi/mux.go.txt", 17308, "cc44c2d620e6306b16d6d80f5f6c70f02b5814b357a4f5823372818f355ae67d", "chi/mux.go"},
	{"source/chi/tree.go.txt", 22069, "f4b12b63b662fb8e36658172b36b35705cfb24eefae0665635f4fbd52e64fb79", "chi/tree.go"},
	{"source/chi/go.mod.txt", 149, "7eb620f0fce870d93eaba87bf6f8c36284db786108030ca229a5016461797090", "chi/go.mod"},
	{"source/chi/LICENSE.txt", 1123, "a2d51b7515acfaff2f7a88688650f2fc4fd99561383e72bba2305e3db59a1647", "chi/LICENSE"},
	{"source/ARMON-NOTICE.txt", 1079, "831892cd31b9eef0311bb1de9014527ef5d3592eed7add1f9f829510d2065e62", "chi/ARMON-NOTICE"},
	{"route/observer.gofmt.go.txt", 12999, "1cfdc8d9e693f02e27ef03623b5ae5b7b66cc3f8c0a4ddc52bc85c006924ec4a", ""},
	{"route/INPUT.proposed.public.v1.json", 2658, "7c18bf602e1566ecbeef4708676691e5b3ee5d2ee4f9845e8938b33cd1620834", "route.json"},
	{"route/WANTS.public.v1.json", 4367, "a79e528f4c217ee5c162056e05ad0586efda3002b8c3a9a01cadbf43feef508b", "wants.json"},
	{"route/OBSERVATIONS.actual.public.v1.json.gz", 2575, "4a439111e5848974b85efe5eb706b23b2238233df088815e679e9be3c4f21dde", ""},
	{"route/HARNESS.go.mod.public.v1.txt", 133, "f0e8fa63df6716ef6a1f33c3b0b1ce6720cb032975776f01d325a3de3a82590f", ""},
	{"preview/INPUT.public.v3.json", 1417, "39e0aadbfe91c0b15ad3da257725de8d74c7a82c9d4222f73da8f937e72bf4c2", "score.json"},
	{"preview/OBSERVATIONS.actual.public.v1.json", 3167, "c4eaa757587f667488dd475197bade68c79baad7b345390eb65bbb3b33937eef", ""},
}

type candidate struct {
	ID      string `json:"id"`
	Caption string `json:"caption"`
}
type fixture struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Raw  string `json:"raw_path"`
	Opt  bool   `json:"opt_in"`
}
type input struct {
	Schema     string      `json:"schema"`
	Fixtures   []fixture   `json:"fixtures"`
	Candidates []candidate `json:"candidates"`
}
type urlState struct {
	Path string `json:"path"`
	Raw  string `json:"raw_path"`
}
type panicState struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type event struct {
	Phase     string      `json:"phase"`
	API       string      `json:"api"`
	Attempted bool        `json:"attempted"`
	Returned  bool        `json:"returned"`
	Value     *string     `json:"value,omitempty"`
	Error     *string     `json:"error,omitempty"`
	Panic     *panicState `json:"panic,omitempty"`
}
type params struct {
	Chi  *string `json:"chi_url_param"`
	Path *string `json:"request_path_value"`
}
type handler struct {
	Kind      string  `json:"kind"`
	Context   *bool   `json:"context_present"`
	RoutePath *string `json:"route_path"`
	Before    *params `json:"params_before"`
	After     *params `json:"params_after"`
	Returned  bool    `json:"returned"`
}
type trial struct {
	Fixture          string      `json:"fixture_id"`
	Candidate        string      `json:"candidate_id"`
	Gate             bool        `json:"gate"`
	Before           *urlState   `json:"original_url_before"`
	After            *urlState   `json:"original_url_after"`
	Dispatch         *urlState   `json:"dispatch_url"`
	RequestCloned    bool        `json:"request_cloned"`
	URLCloned        bool        `json:"url_cloned"`
	Status           *int        `json:"status"`
	Route            int         `json:"route_handler_calls"`
	NotFound         int         `json:"not_found_handler_calls"`
	Handlers         []*handler  `json:"handlers"`
	DispatchReturned *bool       `json:"dispatch_returned"`
	Normal           bool        `json:"normal_return"`
	Panic            *panicState `json:"panic"`
	ObserverError    *string     `json:"observer_error"`
	Ledger           []event     `json:"public_api_calls"`
}
type counts struct {
	Trials    int `json:"trials"`
	Normal    int `json:"normal_returns"`
	Panics    int `json:"panics"`
	Attempted int `json:"api_attempted"`
	Returned  int `json:"api_returned"`
	Errors    int `json:"api_errors"`
	APIPanics int `json:"api_panics"`
}
type report struct {
	Schema     string      `json:"schema"`
	Status     string      `json:"status"`
	SHA        string      `json:"input_sha256"`
	Parents    int         `json:"parents"`
	Candidates []candidate `json:"candidates"`
	Qualified  bool        `json:"qualified"`
	Default    bool        `json:"default_active"`
	Models     int         `json:"model_calls"`
	Fits       int         `json:"fits"`
	Counts     counts      `json:"counts"`
	Trials     []trial     `json:"trials"`
}
type want struct {
	Status       int    `json:"status"`
	Route        int    `json:"route_handler_count"`
	NotFound     int    `json:"not_found_count"`
	Chi          string `json:"url_param_inside_handler"`
	Path         string `json:"path_value_inside_handler"`
	OriginalPath string `json:"original_Path_after"`
	OriginalRaw  string `json:"original_RawPath_after"`
	Normal       bool   `json:"normal_return"`
	Panic        bool   `json:"panic"`
	Errors       int    `json:"unescape_error_count"`
}
type wants struct {
	Schema   string `json:"schema"`
	State    string `json:"state"`
	Request  string `json:"request"`
	Parents  int    `json:"parent_count"`
	Family   string `json:"source_family"`
	Fixtures []struct {
		ID   string `json:"id"`
		Want want   `json:"Want"`
	} `json:"fixtures"`
	Role   json.RawMessage `json:"role"`
	Labels int             `json:"learning_labels"`
	Fits   int             `json:"Fit"`
}

func sha(b []byte) string { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }
func read(path string, limit int) ([]byte, error) {
	st, e := os.Lstat(path)
	if e != nil || !st.Mode().IsRegular() {
		return nil, errors.New("packet_unavailable")
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, errors.New("packet unavailable")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	c := f.Close()
	if e != nil || c != nil || len(b) > limit {
		return nil, errors.New("packet read/bound failure")
	}
	return b, nil
}
func decode(b []byte, v any) error {
	if !utf8.Valid(b) || uniqueJSON(b) != nil {
		return errors.New("invalid UTF8")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return errors.New("invalid JSON shape")
	}
	var x any
	if d.Decode(&x) != io.EOF {
		return errors.New("JSON trailing value")
	}
	var shape any
	q := json.NewDecoder(bytes.NewReader(b))
	q.UseNumber()
	if q.Decode(&shape) != nil || required(shape, reflect.TypeOf(v).Elem()) != nil {
		return errors.New("JSON required shape")
	}
	return nil
}
func unpack(b []byte) ([]byte, error) {
	return unpackLimit(b, rawCap)
}
func unpackLimit(b []byte, limit int) ([]byte, error) {
	r := bytes.NewReader(b)
	z, e := gzip.NewReader(r)
	if e != nil {
		return nil, errors.New("gzip header failure")
	}
	z.Multistream(false)
	raw, e := io.ReadAll(io.LimitReader(z, int64(limit)+1))
	c := z.Close()
	if e != nil || c != nil || len(raw) > limit || r.Len() != 0 {
		return nil, errors.New("gzip CRC/EOF/bound failure")
	}
	return raw, nil
}
func present(p *params) bool { return p != nil && p.Chi != nil && p.Path != nil }
func validate(r report, in input, w wants) error {
	if r.Schema != "riido-chi-route-observation-v1" || r.Status != "development_protocol_audit" || r.SHA != pins[9].sha || r.Parents != 1 || r.Qualified || r.Default || r.Models != 0 || r.Fits != 0 || len(r.Trials) != 45 || !reflect.DeepEqual(r.Candidates, in.Candidates) {
		return errors.New("report_authority")
	}
	var pass [5]int
	for k, t := range r.Trials {
		i, j := k/5, k%5
		v := pilotClassify(t, in.Fixtures[i], w.Fixtures[i].Want, in.Candidates[j])
		if v.State == "unavailable" {
			return errors.New("precheck_unavailable")
		}
		if v.State == "pass" {
			pass[j]++
		}
	}
	if c := pilotChiCounts(r.Trials); c != r.Counts || c != (counts{45, 45, 0, 533, 533, 0, 0}) || pass != [5]int{9, 8, 8, 8, 6} {
		return errors.New("precheck_counts")
	}
	return nil
}

type bounded struct {
	buf      bytes.Buffer
	max      int
	overflow bool
	received int
}

func (b *bounded) Write(p []byte) (int, error) {
	b.received += len(p)
	if len(p) > b.max-b.buf.Len() {
		b.overflow = true
		return 0, errors.New("command output bound")
	}
	return b.buf.Write(p)
}
func (b *bounded) Bytes() []byte { return b.buf.Bytes() }
func environment(tmp string) []string {
	values := []string{"CGO_ENABLED=0", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local", "GOMAXPROCS=2", "GOMEMLIMIT=96MiB", "GOWORK=off", "GOENV=off", "GOFLAGS=", "TMPDIR=" + filepath.Join(tmp, "build-temp")}
	var out []string
	for _, e := range os.Environ() {
		keep := true
		for _, v := range values {
			key := v[:strings.IndexByte(v, '=')+1]
			if strings.HasPrefix(e, key) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, e)
		}
	}
	return append(out, values...)
}

type pilotAPIEvent struct {
	Category  string  `json:"category"`
	Phase     string  `json:"phase"`
	Method    string  `json:"method"`
	Attempted bool    `json:"attempted"`
	Returned  bool    `json:"returned"`
	Error     bool    `json:"error"`
	Panic     bool    `json:"panic"`
	ErrorCode *string `json:"error_code"`
	PanicType *string `json:"panic_type"`
}
type pilotAPICounts struct {
	Attempted int `json:"attempted"`
	Returned  int `json:"returned"`
	Errors    int `json:"errors"`
	Panics    int `json:"panics"`
}
type pilotCost struct {
	ElapsedNS       int64  `json:"elapsed_ns"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	Mallocs         uint64 `json:"mallocs"`
}
type pilotVerdict struct {
	State    string   `json:"state"`
	Channels []string `json:"channels"`
}
type pilotPreVerdict struct {
	CandidateID string          `json:"candidate_id"`
	State       string          `json:"state"`
	Passed      int             `json:"passed"`
	Checks      [9]pilotVerdict `json:"checks"`
}
type pilotRankSnapshot struct {
	Kind           string    `json:"kind"`
	Count          int       `json:"count"`
	Order          [8]int    `json:"order"`
	ScoreBits      [8]string `json:"score_bits"`
	CurrentIDs     [5]string `json:"current_ids"`
	OrderedIDs     [5]string `json:"ordered_ids"`
	FallbackReason string    `json:"fallback_reason"`
}
type pilotCheck struct {
	FixtureIndex int          `json:"fixture_index"`
	FixtureID    string       `json:"fixture_id"`
	Verdict      pilotVerdict `json:"verdict"`
	Trial        trial        `json:"trial"`
}
type pilotVisit struct {
	CandidateIndex      int          `json:"candidate_index"`
	CandidateID         string       `json:"candidate_id"`
	State               string       `json:"state"`
	Checks              []pilotCheck `json:"checks"`
	RemainingFixtureIDs []string     `json:"remaining_fixture_ids"`
}
type pilotPipeline struct {
	PlannedPolicy               string             `json:"planned_policy"`
	EffectivePolicy             string             `json:"effective_policy"`
	Block                       int                `json:"block"`
	Position                    int                `json:"position"`
	State                       string             `json:"state"`
	Ranking                     *rankValue         `json:"ranking"`
	RankSnapshot                *pilotRankSnapshot `json:"rank_snapshot"`
	FailedRanking               *pilotRankSnapshot `json:"failed_ranking"`
	FallbackReason              *string            `json:"fallback_reason"`
	CandidateIDs                [5]string          `json:"candidate_ids"`
	Visited                     [5]bool            `json:"visited"`
	SelectedIndex               *int               `json:"selected_index"`
	Visits                      []pilotVisit       `json:"visits"`
	HintEvents                  []pilotAPIEvent    `json:"hint_events"`
	HintCounts                  pilotAPICounts     `json:"hint_counts"`
	IOCounts                    pilotAPICounts     `json:"io_counts"`
	ChiCounts                   counts             `json:"chi_counts"`
	OwnedVerificationMismatches int                `json:"owned_verification_mismatches"`
	UnavailableChecks           int                `json:"unavailable_checks"`
	ModelRankAttempts           int                `json:"model_rank_attempts"`
	ModelRankReturns            int                `json:"model_rank_returns"`
	Cost                        pilotCost          `json:"cost"`
}
type pilotWire struct {
	Schema              string              `json:"schema"`
	Mode                string              `json:"mode"`
	Block               int                 `json:"block"`
	Status              string              `json:"status"`
	RouteSHA            string              `json:"route_sha256"`
	WantsSHA            string              `json:"wants_sha256"`
	ScoreSHA            string              `json:"score_sha256"`
	ModelSHA            string              `json:"model_sha256"`
	ModelRef            string              `json:"model_ref"`
	Precheck            *report             `json:"precheck"`
	PrecheckVerdicts    []pilotPreVerdict   `json:"precheck_verdicts"`
	Anchor              []pilotRankSnapshot `json:"anchor"`
	SetupCost           *pilotCost          `json:"setup_cost"`
	SetupEvents         []pilotAPIEvent     `json:"setup_events"`
	SetupHintCounts     pilotAPICounts      `json:"setup_hint_counts"`
	SetupIOCounts       pilotAPICounts      `json:"setup_io_counts"`
	Pipelines           []pilotPipeline     `json:"pipelines"`
	ModelRankAttempts   int                 `json:"model_rank_attempts"`
	ModelRankReturns    int                 `json:"model_rank_returns"`
	Fit                 int                 `json:"fit"`
	NewLabels           int                 `json:"new_labels"`
	NewRoles            int                 `json:"new_roles"`
	Qualified           bool                `json:"qualified"`
	DefaultActivated    bool                `json:"default_activated"`
	GPU                 bool                `json:"gpu"`
	LayaEncoder         bool                `json:"laya_encoder"`
	ProtectedEvaluation bool                `json:"protected_evaluation"`
	Scope               string              `json:"scope"`
}

func pilotEventCounts(events []pilotAPIEvent, category string) (c pilotAPICounts, attempts, returns int) {
	for _, e := range events {
		if e.Category == category {
			if e.Attempted {
				c.Attempted++
			}
			if e.Returned {
				c.Returned++
			}
			if e.Error {
				c.Errors++
			}
			if e.Panic {
				c.Panics++
			}
		}
		if e.Method == "model_rank" {
			if e.Attempted {
				attempts++
			}
			if e.Returned {
				returns++
			}
		}
	}
	return
}

func pilotLedgerComplete(t trial, c candidate) bool {
	i := 0
	take := func(phase, api string, value, mayError bool) (event, bool) {
		if i >= len(t.Ledger) {
			return event{}, false
		}
		e := t.Ledger[i]
		i++
		return e, e.Phase == phase && e.API == api && e.Attempted && e.Returned && e.Panic == nil && (e.Value != nil) == value && (mayError || e.Error == nil)
	}
	need := func(phase, api string, value bool) bool { _, ok := take(phase, api, value, false); return ok }
	if !need("setup", "http.NewRequest", true) || !need("setup", "chi.NewRouter", true) || !need("setup", "Mux.Get", false) || !need("setup", "Mux.NotFound", false) {
		return false
	}
	if c.ID == "pre_path" && t.Gate {
		e, ok := take("pre_match", "url.PathUnescape", true, true)
		if !ok || e.Error != nil || !need("pre_match", "Request.Clone", true) {
			return false
		}
	}
	if !need("dispatch", "Mux.ServeHTTP", false) || !need("handler_before", "chi.RouteContext", true) || !need("handler_before", "chi.URLParam", true) || !need("handler_before", "Request.PathValue", true) {
		return false
	}
	if len(t.Handlers) != 1 || t.Handlers[0] == nil {
		return false
	}
	if t.Handlers[0].Kind == "route" && t.Gate && (c.ID == "post_path" || c.ID == "post_query" || c.ID == "twice_path") {
		api := "url.PathUnescape"
		if c.ID == "post_query" {
			api = "url.QueryUnescape"
		}
		e, ok := take("post_match", api, true, true)
		if !ok {
			return false
		}
		if e.Error == nil && c.ID == "twice_path" {
			e, ok = take("post_match", "url.PathUnescape", true, true)
			if !ok {
				return false
			}
		}
		if e.Error == nil && !need("post_match", "Request.SetPathValue", false) {
			return false
		}
	}
	return need("handler_after", "chi.URLParam", true) && need("handler_after", "Request.PathValue", true) && need("handler_response", "ResponseWriter.WriteHeader", false) && i == len(t.Ledger)
}
func pilotClassify(t trial, f fixture, w want, c candidate) pilotVerdict {
	v := pilotVerdict{State: "unavailable", Channels: []string{}}
	if t.Panic != nil {
		v.Channels = append(v.Channels, "panic_unavailable")
	}
	if t.ObserverError != nil {
		v.Channels = append(v.Channels, "observer_error_unavailable")
	}
	if t.Status == nil || t.Before == nil || t.After == nil || t.Dispatch == nil || t.DispatchReturned == nil || len(t.Handlers) == 0 {
		v.Channels = append(v.Channels, "capture_unavailable")
	}
	for _, h := range t.Handlers {
		if h == nil || h.Context == nil || h.RoutePath == nil || !present(h.Before) || !present(h.After) {
			v.Channels = append(v.Channels, "handler_unavailable")
			break
		}
	}
	if !pilotLedgerComplete(t, c) {
		v.Channels = append(v.Channels, "ledger_unavailable")
	}
	if len(v.Channels) != 0 {
		return v
	}
	v.State = "pass"
	add := func(ok bool, name string) {
		if !ok {
			v.Channels = append(v.Channels, name)
		}
	}
	h := t.Handlers[0]
	add(t.Fixture == f.ID && t.Candidate == c.ID, "identity_invariant")
	add(t.Gate == (f.Opt && f.Raw != ""), "gate_invariant")
	add(t.Before.Path == f.Path && t.Before.Raw == f.Raw, "original_before_invariant")
	add(len(t.Handlers) == 1 && t.Route+t.NotFound == 1 && ((h.Kind == "route" && t.Route == 1) || (h.Kind == "not_found" && t.NotFound == 1)), "handler_count_invariant")
	add(*h.Context && *h.RoutePath == "" && h.Returned && *t.DispatchReturned, "live_return_invariant")
	pre := c.ID == "pre_path" && t.Gate
	raw := f.Raw
	if pre {
		raw = ""
	}
	add(t.RequestCloned == pre && t.URLCloned == pre && t.Dispatch.Path == f.Path && t.Dispatch.Raw == raw, "dispatch_invariant")
	add(*t.Status == w.Status, "status")
	add(t.Route == w.Route, "route_handler_count")
	add(t.NotFound == w.NotFound, "not_found_count")
	add(*h.After.Chi == w.Chi, "chi_url_param")
	add(*h.After.Path == w.Path, "path_value")
	add(t.After.Path == w.OriginalPath, "original_path")
	add(t.After.Raw == w.OriginalRaw, "original_raw")
	add(t.Normal == w.Normal, "normal_return")
	add((t.Panic != nil) == w.Panic, "panic")
	errorsSeen := 0
	for _, e := range t.Ledger {
		if e.Error != nil && (e.API == "url.PathUnescape" || e.API == "url.QueryUnescape") {
			errorsSeen++
		}
	}
	add(errorsSeen == w.Errors, "unescape_error_count")
	if len(v.Channels) != 0 {
		v.State = "mismatch"
	}
	return v
}
func pilotChiCounts(trials []trial) (c counts) {
	for _, t := range trials {
		c.Trials++
		if t.Normal {
			c.Normal++
		}
		if t.Panic != nil {
			c.Panics++
		}
		for _, e := range t.Ledger {
			if e.Attempted {
				c.Attempted++
			}
			if e.Returned {
				c.Returned++
			}
			if e.Error != nil {
				c.Errors++
			}
			if e.Panic != nil {
				c.APIPanics++
			}
		}
	}
	return
}
func pilotValidRank(r rankValue) bool {
	if r.Count != 5 || r.Kind == "" {
		return false
	}
	var seen [5]bool
	for i := 0; i < 5; i++ {
		x := r.Order[i]
		if x < 0 || x >= 5 || seen[x] || math.IsNaN(r.Scores[i]) || math.IsInf(r.Scores[i], 0) {
			return false
		}
		seen[x] = true
		if i > 0 {
			a := r.Order[i-1]
			if r.Scores[a] < r.Scores[x] || (r.Scores[a] == r.Scores[x] && a > x) {
				return false
			}
		}
	}
	for i := 5; i < 8; i++ {
		if r.Order[i] != 0 || math.Float64bits(r.Scores[i]) != 0 {
			return false
		}
	}
	return true
}
func pilotSnap(r rankValue, ids [5]string) pilotRankSnapshot {
	s := pilotRankSnapshot{Kind: r.Kind, Count: r.Count, Order: r.Order, CurrentIDs: ids, FallbackReason: r.FallbackReason}
	for i := 0; i < 8; i++ {
		s.ScoreBits[i] = fmt.Sprintf("%016x", math.Float64bits(r.Scores[i]))
	}
	for i := 0; i < 5 && i < r.Count; i++ {
		x := r.Order[i]
		if x >= 0 && x < 5 {
			s.OrderedIDs[i] = ids[x]
		}
	}
	return s
}

type binding struct {
	Bytes int    `json:"bytes"`
	SHA   string `json:"sha256"`
}
type sourcePin struct {
	Name  string `json:"name"`
	Bytes int    `json:"bytes"`
	SHA   string `json:"sha256"`
}
type sourceManifest struct {
	Schema string      `json:"schema"`
	Files  []sourcePin `json:"files"`
}

const freezeTemplate = `{"schema":"riido-chi-cost-freeze-v1","state":"sealed_before_first_pilot_build_or_model_execution","source_pins":null,"parents":1,"labels":0,"roles":0,"Fit":0,"protected_reads":0,"default_activation":false,"gpu_execution":false,"laya_encoder_execution":false,"model":{"bytes":32792,"sha256":"dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004","reference":"JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46"},"go_version":"go1.27.1","CGO_ENABLED":"0","GOMAXPROCS":"2","GOMEMLIMIT":"96MiB","build_timeout_seconds":60,"worker_timeout_seconds":20,"stdout_cap_bytes":262144,"stderr_cap_bytes":8192,"raw_cap_bytes":2097152,"bundle_cap_bytes":8388608,"cold_workers":16,"warm_workers":4,"pipelines":32,"orders":[["bm25","fixed","model","lexical"],["fixed","lexical","bm25","model"],["lexical","model","fixed","bm25"],["model","bm25","lexical","fixed"]],"precheck_before_timed":true,"anchor_before_timed":true,"recompute_trials":true,"warm_setup_separate":true,"cold_lifetime":"one_fresh_worker_per_pipeline","warm_lifetime":"one_fresh_worker_per_block"}`

func checkFreeze(b []byte) (binding, error) {
	var got, want map[string]any
	var ref binding
	if decode(b, &got) != nil || decode([]byte(freezeTemplate), &want) != nil {
		return ref, errors.New("freeze_invalid")
	}
	raw, e := json.Marshal(got["source_pins"])
	if e != nil || decode(raw, &ref) != nil || ref.Bytes < 1 || ref.Bytes > 16384 || !validSHA(ref.SHA) {
		return ref, errors.New("freeze_binding")
	}
	var fields map[string]any
	if decode(raw, &fields) != nil || len(fields) != 2 {
		return ref, errors.New("freeze_binding")
	}
	want["source_pins"] = got["source_pins"]
	if !reflect.DeepEqual(got, want) {
		return ref, errors.New("freeze_mismatch")
	}
	return ref, nil
}
func validSHA(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && hex.EncodeToString(b) == s
}

var repoNames = []string{"internal/hintlearn/learn.go", "internal/hintlearn/model.go", "internal/lexicalhint/audit_provenance.go", "internal/lexicalhint/features.go", "pkg/hintweights/doc.go", "pkg/hintweights/view.go", "pkg/shortclaim/audit_provenance.go", "pkg/shortclaim/baseline.go", "pkg/shortclaim/input.go", "pkg/hintprepared/prepared.go", "LICENSE"}
var pilotNames = []string{"cost-pilot/observer.derived.go.txt", "cost-pilot/pilot.go.txt", "cost-pilot/pilot_test.go.txt", "cost-pilot/run/main.go", "cost-pilot/run/main_test.go"}
var controlNames = strings.Fields("TestPilotFirstMismatchAndNoPruning TestPilotUnavailableContinues TestPilotUnresolvedAndExhausted TestPilotNineChecksAreNotNinePasses TestPilotInvalidRankingAndFallback TestPilotFallbackFailureNoVerification TestPilotTypedKnownAndUnavailableChannels TestPilotPanicNilAttempts TestPilotIncompleteLedgerWithholdsVerdict TestPilotPreparationFailureFallsBackWithoutModelRank TestPilotWilliamsBalance")

func controlsPass(b []byte) bool {
	seen := map[string]int{}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) >= 3 && f[0] == "---" && f[1] == "PASS:" {
			seen[f[2]]++
		}
	}
	for _, name := range controlNames {
		if seen[name] != 1 {
			return false
		}
	}
	return true
}

const repoModule = "module github.com/teamswyg/laya-tools\n\ngo 1.27.1\n"
const workerModule = "module github.com/teamswyg/laya-tools/internal/chicostpilot\n\ngo 1.27.1\n\nrequire (\n github.com/go-chi/chi/v5 v5.0.0\n github.com/teamswyg/laya-tools v0.0.0\n)\n\nreplace github.com/go-chi/chi/v5 => ./chi\nreplace github.com/teamswyg/laya-tools => ./repo\n"

type processRecord struct {
	Stage          string `json:"stage"`
	Block          int    `json:"block"`
	Policy         string `json:"policy"`
	ElapsedNS      int64  `json:"elapsed_ns"`
	Exit           int    `json:"exit"`
	Complete       bool   `json:"complete"`
	Stdout         []byte `json:"stdout"`
	StdoutSHA      string `json:"stdout_sha256"`
	StdoutReceived int    `json:"stdout_received_bytes"`
	StdoutComplete bool   `json:"stdout_complete"`
	StderrBytes    int    `json:"stderr_retained_bytes"`
	StderrReceived int    `json:"stderr_received_bytes"`
	StderrSHA      string `json:"stderr_sha256"`
	StderrComplete bool   `json:"stderr_complete"`
}
type bundle struct {
	Schema         string          `json:"schema"`
	Status         string          `json:"status"`
	Failure        string          `json:"failure_code"`
	FreezeSHA      string          `json:"freeze_sha256"`
	SourcePinsSHA  string          `json:"source_pins_sha256"`
	Parents        int             `json:"parents"`
	Pipelines      int             `json:"pipelines"`
	ModelAttempts  int             `json:"model_rank_attempts"`
	ModelReturns   int             `json:"model_rank_returns"`
	Labels         int             `json:"labels"`
	Roles          int             `json:"roles"`
	Fits           int             `json:"Fit"`
	Qualified      bool            `json:"qualified"`
	Activated      bool            `json:"default_activated"`
	GPU            bool            `json:"gpu"`
	Encoder        bool            `json:"laya_encoder"`
	Protected      int             `json:"protected_reads"`
	Cleanup        bool            `json:"temporary_cleanup_success"`
	CleanupFailure string          `json:"cleanup_failure_code"`
	Processes      []processRecord `json:"processes"`
	Scope          string          `json:"scope"`
}

func command(exe string, args []string, dir, stage, policy string, block, seconds int, b *bundle) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, exe, args...)
	c.Dir = dir
	c.Env = environment(dir)
	c.WaitDelay = time.Second
	limit := outputCap
	if stage == "go_version" {
		limit = 8192
	} else if stage == "synthetic_controls" || stage == "build" {
		limit = 65536
	}
	out, errout := &bounded{max: limit}, &bounded{max: 8192}
	c.Stdout = out
	c.Stderr = errout
	start := time.Now()
	err := c.Run()
	elapsed := time.Since(start).Nanoseconds()
	exit := -1
	if c.ProcessState != nil {
		exit = c.ProcessState.ExitCode()
	}
	ok := err == nil && ctx.Err() == nil && !out.overflow && !errout.overflow
	b.Processes = append(b.Processes, processRecord{stage, block, policy, elapsed, exit, ok, append([]byte(nil), out.Bytes()...), sha(out.Bytes()), out.received, !out.overflow, errout.buf.Len(), errout.received, sha(errout.Bytes()), !errout.overflow})
	if !ok {
		return nil, errors.New("process_failure")
	}
	return out.Bytes(), nil
}

func eventCounts(es []pilotAPIEvent) (pilotAPICounts, pilotAPICounts, int, int, error) {
	for _, e := range es {
		if (e.Category != "io" && e.Category != "hint") || e.Method == "" || e.Phase == "" || !e.Attempted || (!e.Returned && !e.Panic) || (e.Error && !e.Returned) || (e.Panic && e.Returned) || (e.Error != (e.ErrorCode != nil)) || (e.Panic != (e.PanicType != nil)) {
			return pilotAPICounts{}, pilotAPICounts{}, 0, 0, errors.New("event_shape")
		}
	}
	h, a, r := pilotEventCounts(es, "hint")
	i, _, _ := pilotEventCounts(es, "io")
	return h, i, a, r, nil
}
func eventIs(e pilotAPIEvent, category, phase, method string) bool {
	return e.Category == category && e.Phase == phase && e.Method == method && e.Attempted && (e.Returned || e.Panic)
}
func checkSetup(es []pilotAPIEvent, phase string, model bool) (bool, error) {
	methods := []string{"route_load", "wants_read", "score_read", "load_validated"}
	categories := []string{"io", "io", "io", "hint"}
	if model {
		methods = append(methods, "model_read", "view_new", "prepare")
		categories = append(categories, "io", "hint", "hint")
	}
	if len(es) < 1 || len(es) > len(methods) {
		return false, errors.New("setup_length")
	}
	for j, e := range es {
		if !eventIs(e, categories[j], phase, methods[j]) {
			return false, errors.New("setup_sequence")
		}
		if e.Error || e.Panic {
			if j != len(es)-1 {
				return false, errors.New("after_setup_error")
			}
			return true, nil
		}
	}
	if len(es) != len(methods) {
		return false, errors.New("truncated_setup")
	}
	return false, nil
}
func checkTrace(p pilotPipeline, mode string, sharedFailure bool) error {
	es := p.HintEvents
	failed := sharedFailure
	if mode == "cold" {
		n := 0
		for n < len(es) && es[n].Phase == "cold_setup" {
			n++
		}
		var e error
		failed, e = checkSetup(es[:n], "cold_setup", p.PlannedPolicy == "model")
		if e != nil {
			return e
		}
		es = es[n:]
		if p.State == "setup_unavailable" {
			if !failed || len(es) != 0 {
				return errors.New("setup_terminal")
			}
			return nil
		}
		if failed {
			last := p.HintEvents[n-1]
			if last.Category != "hint" || (last.Method != "view_new" && last.Method != "prepare") {
				return errors.New("setup_io_terminal")
			}
		}
	}
	method := "baseline_rank"
	if p.PlannedPolicy == "model" {
		method = "model_rank"
	}
	if p.FallbackReason != nil && *p.FallbackReason == "model_setup_error" {
		if p.PlannedPolicy != "model" || !failed || len(es) != 1 || !eventIs(es[0], "hint", "fallback", "bm25_fallback") {
			return errors.New("preparation_fallback")
		}
		return nil
	}
	if mode == "cold" && failed {
		return errors.New("lost_setup_error")
	}
	if len(es) < 1 || !eventIs(es[0], "hint", "rank", method) {
		return errors.New("missing_rank_attempt")
	}
	if p.FallbackReason == nil {
		if len(es) != 1 || es[0].Error || es[0].Panic {
			return errors.New("rank_success_trace")
		}
	} else {
		if len(es) != 2 || !eventIs(es[1], "hint", "fallback", "bm25_fallback") || (*p.FallbackReason == "rank_error" && !es[0].Error && !es[0].Panic) || (*p.FallbackReason == "invalid_ranking" && (es[0].Error || es[0].Panic)) {
			return errors.New("rank_fallback_trace")
		}
	}
	return nil
}
func unavailableWitness(t trial) bool {
	if t.Panic != nil || t.ObserverError != nil || !t.Normal || t.Status == nil || t.Before == nil || t.After == nil || t.Dispatch == nil || t.DispatchReturned == nil || !*t.DispatchReturned || len(t.Handlers) == 0 {
		return true
	}
	for _, h := range t.Handlers {
		if h == nil || h.Context == nil || h.RoutePath == nil || !present(h.Before) || !present(h.After) {
			return true
		}
	}
	for _, e := range t.Ledger {
		if e.Panic != nil || e.Error != nil || !e.Returned {
			return true
		}
	}
	return false
}
func checkPipeline(p pilotPipeline, block, pos int, policy, mode string, in input, w wants, saved report, expected map[string]pilotRankSnapshot, setupFailed bool) error {
	ids := expected["model"].CurrentIDs
	if p.Block != block || p.Position != pos || p.PlannedPolicy != policy || p.Cost.ElapsedNS < 0 {
		return errors.New("pipeline_schedule")
	}
	h, ioCount, a, r, e := eventCounts(p.HintEvents)
	if e != nil || h != p.HintCounts || ioCount != p.IOCounts || a != p.ModelRankAttempts || r != p.ModelRankReturns {
		return errors.New("pipeline_hint_counts")
	}
	if checkTrace(p, mode, setupFailed) != nil {
		return errors.New("pipeline_event_sequence")
	}
	if p.State == "setup_unavailable" {
		if p.Ranking != nil || p.RankSnapshot != nil || p.FailedRanking != nil || p.SelectedIndex != nil || len(p.Visits) != 0 || p.Visited != [5]bool{} || p.ChiCounts != (counts{}) || p.OwnedVerificationMismatches != 0 || p.UnavailableChecks != 0 {
			return errors.New("setup_shape")
		}
		return nil
	}
	if p.CandidateIDs != ids {
		return errors.New("candidate_retention")
	}
	effective := policy
	if p.FallbackReason != nil {
		effective = "bm25"
		if p.FailedRanking == nil || (*p.FallbackReason != "rank_error" && *p.FallbackReason != "invalid_ranking" && *p.FallbackReason != "model_setup_error") || p.FailedRanking.CurrentIDs != ids {
			return errors.New("fallback_shape")
		}
		for _, s := range p.FailedRanking.ScoreBits {
			if len(s) != 16 {
				return errors.New("failed_bits")
			}
			if _, e := strconv.ParseUint(s, 16, 64); e != nil {
				return errors.New("failed_bits")
			}
		}
	} else if p.FailedRanking != nil {
		return errors.New("failed_rank_without_fallback")
	}
	if p.EffectivePolicy != effective {
		return errors.New("effective_policy")
	}
	if p.Ranking == nil {
		if p.FallbackReason == nil || p.RankSnapshot != nil || p.State != "unresolved" || p.SelectedIndex != nil || len(p.Visits) != 0 || p.Visited != [5]bool{} || p.ChiCounts != (counts{}) || p.OwnedVerificationMismatches != 0 || p.UnavailableChecks != 0 {
			return errors.New("rank_unavailable_shape")
		}
		return nil
	}
	if len(p.HintEvents) == 0 {
		return errors.New("ranking_without_successful_call")
	}
	last := p.HintEvents[len(p.HintEvents)-1]
	if !last.Returned || last.Error || last.Panic {
		return errors.New("ranking_without_successful_call")
	}
	if !pilotValidRank(*p.Ranking) || p.RankSnapshot == nil || *p.RankSnapshot != pilotSnap(*p.Ranking, ids) || *p.RankSnapshot != expected[effective] {
		return errors.New("pipeline_ranking")
	}
	var visited [5]bool
	var ts []trial
	unknown := false
	selected := -1
	mis, unavail := 0, 0
	for vi, v := range p.Visits {
		if vi >= 5 || selected >= 0 || v.CandidateIndex != p.Ranking.Order[vi] || v.CandidateID != ids[v.CandidateIndex] || len(v.Checks) < 1 || len(v.Checks) > 9 {
			return errors.New("visit_prefix")
		}
		visited[v.CandidateIndex] = true
		state := "verified"
		remaining := []string{}
		for fi, c := range v.Checks {
			if c.FixtureIndex != fi || c.FixtureID != in.Fixtures[fi].ID {
				return errors.New("fixture_order")
			}
			got := pilotClassify(c.Trial, in.Fixtures[fi], w.Fixtures[fi].Want, in.Candidates[v.CandidateIndex])
			if c.Trial.Panic != nil && c.Trial.Panic.Value == "callback panic (value withheld)" {
				got = pilotVerdict{"unavailable", []string{"callback_panic"}}
			}
			if !reflect.DeepEqual(got, c.Verdict) {
				return errors.New("verdict_reconstruction")
			}
			ts = append(ts, c.Trial)
			if got.State == "pass" || got.State == "mismatch" {
				if !reflect.DeepEqual(c.Trial, saved.Trials[fi*5+v.CandidateIndex]) {
					return errors.New("complete_trial_changed")
				}
			} else if !unavailableWitness(c.Trial) {
				return errors.New("unavailable_without_witness")
			}
			if got.State != "pass" {
				if fi != len(v.Checks)-1 {
					return errors.New("early_stop")
				}
				state = got.State
				for rest := fi + 1; rest < 9; rest++ {
					remaining = append(remaining, in.Fixtures[rest].ID)
				}
				if state == "mismatch" {
					mis++
				} else if state == "unavailable" {
					unavail++
					unknown = true
				} else {
					return errors.New("verdict_state")
				}
			}
		}
		if state == "verified" && len(v.Checks) != 9 {
			return errors.New("incomplete_success")
		}
		if v.State != state || !reflect.DeepEqual(v.RemainingFixtureIDs, remaining) {
			return errors.New("visit_state")
		}
		if state == "verified" {
			selected = v.CandidateIndex
		}
	}
	state := "exhausted_mismatches"
	if unknown {
		state = "unresolved"
	}
	if selected >= 0 {
		state = "verified"
	} else if len(p.Visits) != 5 {
		return errors.New("incomplete_candidates")
	}
	if p.State != state || p.Visited != visited || p.ChiCounts != pilotChiCounts(ts) || p.OwnedVerificationMismatches != mis || p.UnavailableChecks != unavail || (selected >= 0) != (p.SelectedIndex != nil) || (p.SelectedIndex != nil && *p.SelectedIndex != selected) {
		return errors.New("pipeline_counts_state")
	}
	return nil
}

type priorRanking struct {
	Kind     string     `json:"kind"`
	Count    int        `json:"count"`
	Order    [5]int     `json:"order"`
	IDs      [5]string  `json:"ordered_ids"`
	Scores   [5]float64 `json:"scores"`
	Bits     [5]string  `json:"score_bits"`
	Fallback string     `json:"fallback,omitempty"`
}

func anchorRanks(raw []byte, in input) (map[string]pilotRankSnapshot, error) {
	var doc map[string]json.RawMessage
	var prepared priorRanking
	var baselines [4]priorRanking
	if decode(raw, &doc) != nil || decode(doc["prepared"], &prepared) != nil || decode(doc["baselines"], &baselines) != nil {
		return nil, errors.New("anchor_reference")
	}
	var ids [5]string
	for i, c := range in.Candidates {
		ids[i] = c.ID
	}
	out := map[string]pilotRankSnapshot{}
	names := []string{"model", "fixed", "bm25", "lexical", "narrow"}
	for i, p := range append([]priorRanking{prepared}, baselines[:]...) {
		var value rankValue
		value.Kind = p.Kind
		value.Count = p.Count
		value.FallbackReason = p.Fallback
		copy(value.Order[:], p.Order[:])
		copy(value.Scores[:], p.Scores[:])
		s := pilotSnap(value, ids)
		if !pilotValidRank(value) {
			return nil, errors.New("anchor_reference")
		}
		for j := 0; j < 5; j++ {
			if s.ScoreBits[j] != p.Bits[j] || s.OrderedIDs[j] != p.IDs[j] {
				return nil, errors.New("anchor_reference_bits")
			}
		}
		out[names[i]] = s
	}
	return out, nil
}
func checkWire(x pilotWire, mode, policy string, block int, in input, w wants, saved report, expected map[string]pilotRankSnapshot) error {
	if x.Schema != "riido-chi-paired-cost-worker-v1" || x.Mode != mode || x.Block != block || x.Status != "completed" || x.RouteSHA != pins[9].sha || x.WantsSHA != pins[10].sha || x.ScoreSHA != pins[13].sha || x.ModelSHA != modelSHA || x.ModelRef != modelRef || x.Qualified || x.DefaultActivated || x.GPU || x.LayaEncoder || x.ProtectedEvaluation || x.Fit != 0 || x.NewLabels != 0 || x.NewRoles != 0 {
		return errors.New("worker_authority_status")
	}
	h, i, a, r, e := eventCounts(x.SetupEvents)
	if e != nil || h != x.SetupHintCounts || i != x.SetupIOCounts || (x.SetupCost != nil && x.SetupCost.ElapsedNS < 0) {
		return errors.New("setup_counts")
	}
	setup := x.SetupEvents
	if mode == "anchor" {
		if len(setup) < 2 {
			return errors.New("anchor_events")
		}
		tail := setup[len(setup)-2:]
		if !eventIs(tail[0], "hint", "anchor", "model_rank") || !eventIs(tail[1], "hint", "anchor", "baselines_anchor") || tail[0].Error || tail[0].Panic || tail[1].Error || tail[1].Panic {
			return errors.New("anchor_events")
		}
		setup = setup[:len(setup)-2]
	}
	setupFailed := false
	if mode == "cold" {
		if len(setup) != 0 {
			return errors.New("cold_outer_setup")
		}
	} else {
		var err error
		setupFailed, err = checkSetup(setup, "setup", mode != "precheck")
		if err != nil || (setupFailed && mode != "warm") {
			return errors.New("setup_event_sequence")
		}
		if setupFailed {
			last := setup[len(setup)-1]
			if last.Category != "hint" || (last.Method != "view_new" && last.Method != "prepare") {
				return errors.New("warm_io_setup_terminal")
			}
		}
	}
	switch mode {
	case "precheck":
		if x.Precheck == nil || validate(*x.Precheck, in, w) != nil || !reflect.DeepEqual(*x.Precheck, saved) || len(x.PrecheckVerdicts) != 5 || len(x.Pipelines) != 0 || len(x.Anchor) != 0 || x.SetupCost == nil {
			return errors.New("precheck_mismatch")
		}
		for ci, p := range x.PrecheckVerdicts {
			state := "verified"
			pass := 0
			for fi := 0; fi < 9; fi++ {
				v := pilotClassify(x.Precheck.Trials[fi*5+ci], in.Fixtures[fi], w.Fixtures[fi].Want, in.Candidates[ci])
				if !reflect.DeepEqual(v, p.Checks[fi]) {
					return errors.New("precheck_verdict")
				}
				if v.State == "pass" {
					pass++
				} else if v.State == "unavailable" {
					state = "unavailable"
				} else if state != "unavailable" {
					state = "mismatch"
				}
			}
			if p.CandidateID != in.Candidates[ci].ID || p.State != state || p.Passed != pass {
				return errors.New("precheck_verdict_counts")
			}
		}
	case "anchor":
		if x.Precheck != nil || len(x.PrecheckVerdicts) != 0 || len(x.Pipelines) != 0 || len(x.Anchor) != 5 || x.SetupCost == nil {
			return errors.New("anchor_shape")
		}
		for j, name := range []string{"model", "fixed", "bm25", "lexical", "narrow"} {
			if x.Anchor[j] != expected[name] {
				return errors.New("anchor_bits_order")
			}
		}
	case "cold", "warm":
		n := 1
		if mode == "warm" {
			n = 4
		}
		if len(x.Pipelines) != n || x.Precheck != nil || len(x.PrecheckVerdicts) != 0 || len(x.Anchor) != 0 || (mode == "warm") != (x.SetupCost != nil) {
			return errors.New("timed_shape")
		}
		for j, p := range x.Pipelines {
			pos := j
			wantPolicy := orders[block][j]
			if mode == "cold" {
				wantPolicy = policy
				pos = -1
				for k, s := range orders[block] {
					if s == policy {
						pos = k
					}
				}
			}
			if checkPipeline(p, block, pos, wantPolicy, mode, in, w, saved, expected, setupFailed) != nil {
				return errors.New("timed_validation")
			}
			a += p.ModelRankAttempts
			r += p.ModelRankReturns
		}
	default:
		return errors.New("mode_invalid")
	}
	if a != x.ModelRankAttempts || r != x.ModelRankReturns {
		return errors.New("worker_model_counts")
	}
	return nil
}

type packet struct {
	Data                 [len(pins)][]byte
	Sources              map[string][]byte
	In                   input
	Wants                wants
	Saved                report
	Expected             map[string]pilotRankSnapshot
	FreezeSHA, SourceSHA string
}

func loadPacket(root, repo string) (p packet, err error) {
	freeze, e := read(filepath.Join(root, "cost-pilot/FREEZE.public.v1.json"), 16384)
	if e != nil {
		return p, errors.New("freeze_read")
	}
	ref, e := checkFreeze(freeze)
	if e != nil {
		return p, e
	}
	p.FreezeSHA = sha(freeze)
	raw, e := read(filepath.Join(root, "cost-pilot/SOURCE-PINS.public.v1.json"), ref.Bytes)
	if e != nil || len(raw) != ref.Bytes || sha(raw) != ref.SHA {
		return p, errors.New("source_manifest_binding")
	}
	p.SourceSHA = sha(raw)
	var m sourceManifest
	if decode(raw, &m) != nil || m.Schema != "riido-chi-cost-source-pins-v1" || len(m.Files) != len(repoNames)+len(pilotNames) {
		return p, errors.New("source_manifest_shape")
	}
	allowed := map[string]bool{}
	for _, n := range repoNames {
		allowed[n] = true
	}
	for _, n := range pilotNames {
		allowed[n] = false
	}
	p.Sources = map[string][]byte{}
	for _, f := range m.Files {
		isRepo, ok := allowed[f.Name]
		if !ok || p.Sources[f.Name] != nil || f.Bytes < 1 || f.Bytes > 65536 || !validSHA(f.SHA) {
			return p, errors.New("source_selection")
		}
		base := root
		if isRepo {
			base = repo
		}
		data, e := read(filepath.Join(base, f.Name), f.Bytes)
		if e != nil || len(data) != f.Bytes || sha(data) != f.SHA {
			return p, errors.New("source_pin")
		}
		p.Sources[f.Name] = data
		if strings.HasPrefix(f.Name, "cost-pilot/run/") {
			compiled, e := ownSource.ReadFile(filepath.Base(f.Name))
			if e != nil || !bytes.Equal(compiled, data) {
				return p, errors.New("compiled_controller_binding")
			}
		}
	}
	for i, f := range pins {
		p.Data[i], err = read(filepath.Join(root, f.name), f.size)
		if err != nil || len(p.Data[i]) != f.size || sha(p.Data[i]) != f.sha {
			return p, errors.New("packet_pin")
		}
	}
	derived := bytes.Replace(p.Data[8], []byte("func main()"), []byte("func observerMain()"), 1)
	if !bytes.Equal(derived, p.Sources[pilotNames[0]]) {
		return p, errors.New("observer_derivation")
	}
	for name, want := range map[string]string{"cost-pilot/WORKER.go.mod.txt": workerModule, "cost-pilot/REPO.go.mod.txt": repoModule} {
		data, e := read(filepath.Join(root, name), len(want))
		if e != nil || string(data) != want {
			return p, errors.New("module_binding")
		}
	}
	if decode(p.Data[9], &p.In) != nil || decode(p.Data[10], &p.Wants) != nil || len(p.In.Fixtures) != 9 || len(p.In.Candidates) != 5 || len(p.Wants.Fixtures) != 9 || p.In.Schema != "riido-chi-route-frozen-v1" || p.Wants.Schema != "riido-chi-route-literal-wants-v1" || p.Wants.Parents != 1 || p.Wants.Labels != 0 || p.Wants.Fits != 0 || string(p.Wants.Role) != "null" {
		return p, errors.New("packet_authority")
	}
	for i, f := range p.In.Fixtures {
		if f.ID != p.Wants.Fixtures[i].ID {
			return p, errors.New("fixture_binding")
		}
	}
	raw, e = unpack(p.Data[11])
	if e != nil || len(raw) != 80431 || sha(raw) != "8b7600df77c31787af3efe929166126f8c4c06422ef7d894b687e1a19dec71f1" || decode(raw, &p.Saved) != nil || validate(p.Saved, p.In, p.Wants) != nil {
		return p, errors.New("saved_precheck")
	}
	p.Expected, err = anchorRanks(p.Data[14], p.In)
	return
}
func checkBundle(b bundle, p packet) error {
	if b.Schema != "riido-chi-paired-cost-bundle-v1" || b.Status != "completed_development_cost_audit" || b.Failure != "" || b.CleanupFailure != "" || b.FreezeSHA != p.FreezeSHA || b.SourcePinsSHA != p.SourceSHA || b.Parents != 1 || b.Labels != 0 || b.Roles != 0 || b.Fits != 0 || b.Qualified || b.Activated || b.GPU || b.Encoder || b.Protected != 0 || !b.Cleanup || len(b.Processes) != 25 {
		return errors.New("bundle_authority_schedule")
	}
	index, total, a, r := 0, 0, 0, 0
	check := func(stage, policy string, block int) error {
		q := b.Processes[index]
		index++
		limit := outputCap
		if stage == "go_version" {
			limit = 8192
		} else if stage == "build" || stage == "synthetic_controls" {
			limit = 65536
		}
		if q.Stage != stage || q.Policy != policy || q.Block != block || q.Exit != 0 || !q.Complete || !q.StdoutComplete || !q.StderrComplete || q.ElapsedNS < 0 || len(q.Stdout) > limit || q.StdoutReceived != len(q.Stdout) || q.StdoutSHA != sha(q.Stdout) || q.StderrBytes < 0 || q.StderrBytes > 8192 || q.StderrReceived != q.StderrBytes || !validSHA(q.StderrSHA) {
			return errors.New("process_receipt")
		}
		if stage == "go_version" {
			f := strings.Fields(string(q.Stdout))
			if len(f) != 4 || f[0] != "go" || f[1] != "version" || f[2] != "go1.27.1" {
				return errors.New("saved_go_version")
			}
		} else if stage == "synthetic_controls" {
			if !controlsPass(q.Stdout) {
				return errors.New("saved_controls_missing")
			}
		} else if stage != "build" {
			raw, e := unpack(q.Stdout)
			var x pilotWire
			if e != nil || decode(raw, &x) != nil || checkWire(x, stage, policy, block, p.In, p.Wants, p.Saved, p.Expected) != nil {
				return errors.New("saved_worker_protocol")
			}
			left := q.ElapsedNS
			if x.SetupCost != nil {
				if x.SetupCost.ElapsedNS > left {
					return errors.New("impossible_setup_time")
				}
				left -= x.SetupCost.ElapsedNS
			}
			for _, v := range x.Pipelines {
				if v.Cost.ElapsedNS > left {
					return errors.New("impossible_pipeline_time")
				}
				left -= v.Cost.ElapsedNS
			}
			total += len(x.Pipelines)
			a += x.ModelRankAttempts
			r += x.ModelRankReturns
		}
		return nil
	}
	for _, s := range []string{"go_version", "synthetic_controls", "build", "precheck", "anchor"} {
		if e := check(s, "", 0); e != nil {
			return e
		}
	}
	for block, row := range orders {
		for _, policy := range row {
			if e := check("cold", policy, block); e != nil {
				return e
			}
		}
	}
	for block := range orders {
		if e := check("warm", "", block); e != nil {
			return e
		}
	}
	if index != 25 || total != 32 || b.Pipelines != total || b.ModelAttempts != a || b.ModelReturns != r {
		return errors.New("bundle_totals")
	}
	return nil
}

type options struct{ Root, Repo, Model, Go, Out, Saved string }

func execute(o options, p packet, b *bundle) (err error) {
	model, e := read(o.Model, modelBytes)
	if e != nil || len(model) != modelBytes || sha(model) != modelSHA {
		return errors.New("model_binding")
	}
	tool, e := exec.LookPath(o.Go)
	if e != nil {
		return errors.New("go_unavailable")
	}
	tool, e = filepath.Abs(tool)
	if e != nil {
		return errors.New("go_locator")
	}
	asset, e := filepath.Abs(o.Model)
	if e != nil {
		return errors.New("model_locator")
	}
	tmp, e := os.MkdirTemp("", "riido-chi-cost-")
	if e != nil {
		return errors.New("temporary_setup")
	}
	b.Cleanup = false
	defer func() { err = cleanupResult(b, err, func() error { return os.RemoveAll(tmp) }) }()
	put := func(name string, data []byte) error {
		path := filepath.Join(tmp, name)
		if os.MkdirAll(filepath.Dir(path), 0700) != nil || os.WriteFile(path, data, 0600) != nil {
			return errors.New("temporary_copy")
		}
		return nil
	}
	if os.Chmod(tmp, 0700) != nil || os.Mkdir(filepath.Join(tmp, "build-temp"), 0700) != nil {
		return errors.New("temporary_setup")
	}
	for i, f := range pins {
		if f.dest != "" {
			if e = put(f.dest, p.Data[i]); e != nil {
				return e
			}
		}
	}
	for _, n := range repoNames {
		if e = put(filepath.Join("repo", n), p.Sources[n]); e != nil {
			return e
		}
	}
	for i, n := range []string{"observer.go", "pilot.go", "pilot_test.go"} {
		if e = put(n, p.Sources[pilotNames[i]]); e != nil {
			return e
		}
	}
	if put("go.mod", []byte(workerModule)) != nil || put("repo/go.mod", []byte(repoModule)) != nil {
		return errors.New("temporary_modules")
	}
	version, e := command(tool, []string{"version"}, tmp, "go_version", "", 0, 5, b)
	if e != nil {
		return errors.New("version_failure")
	}
	fields := strings.Fields(string(version))
	if len(fields) != 4 || fields[0] != "go" || fields[1] != "version" || fields[2] != "go1.27.1" {
		return errors.New("go_version")
	}
	controls, e := command(tool, []string{"test", "-count=1", "-v", "-run", "^TestPilot", "-p=1", "-trimpath", "-buildvcs=false", "."}, tmp, "synthetic_controls", "", 0, 60, b)
	if e != nil || !controlsPass(controls) {
		return errors.New("controls_failure")
	}
	worker := filepath.Join(tmp, "worker")
	if _, e = command(tool, []string{"build", "-p=1", "-trimpath", "-buildvcs=false", "-o", worker, "."}, tmp, "build", "", 0, 60, b); e != nil {
		return errors.New("build_failure")
	}
	call := func(mode, policy string, block int) error {
		args := []string{"-mode", mode, "-route", filepath.Join(tmp, "route.json"), "-wants", filepath.Join(tmp, "wants.json"), "-score", filepath.Join(tmp, "score.json"), "-model", asset, "-block", strconv.Itoa(block)}
		if policy != "" {
			args = append(args, "-policy", policy)
		}
		out, e := command(worker, args, tmp, mode, policy, block, 20, b)
		if e != nil {
			return errors.New("worker_process_failure")
		}
		raw, e := unpack(out)
		var x pilotWire
		if e != nil || decode(raw, &x) != nil || checkWire(x, mode, policy, block, p.In, p.Wants, p.Saved, p.Expected) != nil {
			return errors.New("worker_protocol_failure")
		}
		b.Pipelines += len(x.Pipelines)
		b.ModelAttempts += x.ModelRankAttempts
		b.ModelReturns += x.ModelRankReturns
		return nil
	}
	if e = call("precheck", "", 0); e != nil {
		return e
	}
	if e = call("anchor", "", 0); e != nil {
		return e
	}
	for block, row := range orders {
		for _, policy := range row {
			if e = call("cold", policy, block); e != nil {
				return e
			}
		}
	}
	for block := range orders {
		if e = call("warm", "", block); e != nil {
			return e
		}
	}
	return nil
}
func guardedExecute(o options, p packet, b *bundle) (err error) {
	complete := false
	defer func() {
		if !complete {
			_ = recover()
			err = errors.New("controller_panic")
		}
	}()
	err = execute(o, p, b)
	complete = true
	return
}
func cleanupResult(b *bundle, primary error, remove func() error) error {
	b.Cleanup = remove() == nil
	if !b.Cleanup {
		b.CleanupFailure = "cleanup_failure"
		if primary == nil {
			return errors.New("cleanup_failure")
		}
	}
	return primary
}
func writeBundle(path string, b bundle) error {
	raw, e := json.Marshal(b)
	if e != nil || len(raw) > bundleCap {
		return errors.New("bundle_bound")
	}
	var out bounded
	out.max = bundleCap
	z := gzip.NewWriter(&out)
	if _, e = z.Write(raw); e != nil || z.Close() != nil {
		return errors.New("bundle_encode")
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errors.New("result_create")
	}
	_, e = f.Write(out.Bytes())
	s, c := f.Sync(), f.Close()
	if e != nil || s != nil || c != nil {
		return errors.New("result_write")
	}
	return nil
}
func run() error {
	fs := flag.NewFlagSet("chi-cost-audit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o options
	fs.StringVar(&o.Root, "root", "experiments/short-claim/next-cohort-chi-audit", "public Chi packet")
	fs.StringVar(&o.Repo, "repo", ".", "repository source")
	fs.StringVar(&o.Model, "model", "", "explicit existing inactive asset")
	fs.StringVar(&o.Go, "go", "go", "local Go1.27.1")
	fs.StringVar(&o.Out, "out", "", "new result gzip file")
	fs.StringVar(&o.Saved, "saved", "", "model-free saved result audit")
	if fs.Parse(os.Args[1:]) != nil || fs.NArg() != 0 || (o.Saved == "" && (o.Model == "" || o.Out == "")) || (o.Saved != "" && (o.Model != "" || o.Out != "")) {
		return errors.New("arguments_invalid")
	}
	p, e := loadPacket(o.Root, o.Repo)
	if e != nil {
		return errors.New("packet_binding_failure")
	}
	if o.Saved != "" {
		compressed, e := read(o.Saved, bundleCap)
		if e != nil {
			return errors.New("saved_read")
		}
		raw, e := unpackLimit(compressed, bundleCap)
		var b bundle
		if e != nil || decode(raw, &b) != nil || checkBundle(b, p) != nil {
			return errors.New("saved_audit_failure")
		}
		if _, e = fmt.Fprintln(os.Stdout, "Chi saved cost audit passed: parent=1 pipelines=32; no new worker or model calls"); e != nil {
			return errors.New("summary_write")
		}
		return nil
	}
	if _, e = os.Lstat(o.Out); !os.IsNotExist(e) {
		return errors.New("result_already_exists")
	}
	b := bundle{Schema: "riido-chi-paired-cost-bundle-v1", Status: "failed", FreezeSHA: p.FreezeSHA, SourcePinsSHA: p.SourceSHA, Parents: 1, Cleanup: true, Scope: "one exposed development parent; fresh Chi verification per pipeline; precheck equality audits trace integrity only; no cached verdict/scoring substitute; process elapsed separate from worker allocation/timing; no model quality, RSS/GPU, LLM savings or new independent parents"}
	e = guardedExecute(o, p, &b)
	if e == nil {
		b.Status = "completed_development_cost_audit"
		if checkBundle(b, p) != nil {
			e = errors.New("final_bundle_validation")
		}
	}
	if e != nil {
		b.Status = "failed"
		b.Failure = e.Error()
	}
	if w := writeBundle(o.Out, b); w != nil {
		return w
	}
	if e != nil {
		return e
	}
	if _, e = fmt.Fprintln(os.Stdout, "Chi development cost audit completed: parent=1 pipelines=32 Fit=0; result is not a quality improvement claim"); e != nil {
		return errors.New("summary_write")
	}
	return nil
}
func guardedRun() (err error) {
	complete := false
	defer func() {
		if !complete {
			_ = recover()
			err = errors.New("controller_panic")
		}
	}()
	err = run()
	complete = true
	return
}
func main() {
	if e := guardedRun(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
