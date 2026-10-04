// SPDX-License-Identifier: Apache-2.0
// Opt-in, one exposed Chi parent. Default controls use synthetic callbacks only.
// Scorer construction costs are not quality, verification savings or RSS/GPU.
package hintprepared

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

var constructorInput = flag.String("constructor-input", "", "explicit pinned public input")
var constructorModel = flag.String("constructor-model", "", "explicit pinned inactive model")
var constructorOutput = flag.String("constructor-output", "", "new caller-owned report file")

const constructorInputPin = "39e0aadbfe91c0b15ad3da257725de8d74c7a82c9d4222f73da8f937e72bf4c2"
const constructorModelPin = "dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004"
const constructorScope = "One exposed development parent; all 24 intervals pay Prepare and N Rank; GC/MemStats and parity conversion outside elapsed interval; Go cumulative allocation, not peak/retained heap/RSS/GPU; no profiles, labels, roles, Fit, qualification, activation or protected evaluation. Instrumented wrappers exclude accessors/runtime/hidden calls."

// Exact prior Prepare body; only its function name changes. Original Features
// remains the independent oracle. Rank is shared unchanged by both owners.
func constructorLegacy(p shortclaim.Prepared) (*Prepared, error) {
	if err := shortclaim.ValidatePrepared(p); err != nil {
		return nil, err
	}
	var rows [shortclaim.MaxCandidates][]hintlearn.Feature
	var offsets [shortclaim.MaxCandidates + 1]uint32
	total := 0
	for i := 0; i < p.Count; i++ {
		fs := hintlearn.Features(p.Request, p.Candidates[i].Text)
		if len(fs) > hintweights.Dimension {
			return nil, ErrFeature
		}
		for _, f := range fs {
			if f.Index < 0 || f.Index >= hintweights.Dimension || math.IsNaN(f.Value) || math.IsInf(f.Value, 0) {
				return nil, ErrFeature
			}
		}
		rows[i] = fs
		total += len(fs)
		offsets[i+1] = uint32(total) // 8*8192 can be 65536; never uint16 offsets.
	}
	out := &Prepared{request: strings.Clone(p.Request), offsets: offsets, count: p.Count, features: make([]hintweights.Feature, total)}
	for i := 0; i < p.Count; i++ {
		out.texts[i] = strings.Clone(p.Candidates[i].Text)
		for j, f := range rows[i] {
			out.features[int(offsets[i])+j] = hintweights.Feature{Index: f.Index, Value: f.Value}
		}
	}
	out.ready = true
	return out, nil
}

type constructorEvent struct {
	API, Code                         string
	Attempted, Returned, Error, Panic bool
}
type constructorCounts struct{ Attempted, Returned, Errors, Panics, Prepares, Ranks int }
type constructorCost struct {
	ElapsedNS                int64
	TotalAllocBytes, Mallocs uint64
}
type constructorBits struct {
	Index     int
	ValueBits string
}
type constructorPayload struct {
	Count           int
	Ready           bool
	Offsets         [9]uint32
	FeatureLenCap   [2]int
	FeatureBytes    int
	ClonedTextBytes int
	OwnerFixedBytes int
	Rows            [][]constructorBits
}
type constructorRanking struct {
	Kind         string
	Count        int
	Order        [8]int
	ScoreBits    [8]string
	CandidateIDs [5]string
	OrderedIDs   [5]string
	Fallback     string
}
type constructorResult struct {
	Event         constructorEvent
	Ranking       *constructorRanking
	MatchesAnchor *bool
}
type constructorTrial struct {
	N              int
	Pair           int
	Position       int
	Method         string
	State          string
	Prepare        constructorEvent
	Owner          *constructorPayload
	OwnerParity    bool
	OwnsInput      bool
	Calls          []constructorResult
	RemainingRanks int
	Cost           constructorCost
	Counts         constructorCounts
}
type constructorStage struct {
	Name  string
	Event constructorEvent
	Cost  constructorCost
}
type constructorReport struct {
	Schema              string
	State               string
	Scope               string
	InputSHA            string
	ModelSHA            string
	ModelRef            string
	Parents             int
	RuntimeGo           string
	GOMAXPROCS          int
	MemoryLimitBytes    int64
	Fit                 int
	NewLabels           int
	NewRoles            int
	Qualified           bool
	Activated           bool
	ProtectedEvaluation bool
	Stages              []constructorStage
	Legacy              *constructorPayload
	Scratch             *constructorPayload
	LegacyAnchor        *constructorRanking
	ScratchAnchor       *constructorRanking
	Trials              []constructorTrial
	Totals              constructorCounts
}

func constructorCall[T any](api string, fn func() (T, error)) (v T, e constructorEvent) {
	e = constructorEvent{API: api, Attempted: true}
	done := false
	defer func() {
		if !done {
			_ = recover()
			e.Panic = true
			e.Code = "panic"
		}
	}()
	var err error
	v, err = fn()
	done = true
	e.Returned = true
	if err != nil {
		e.Error = true
		e.Code = "returned_error"
	}
	return
}
func constructorOK(e constructorEvent) bool { return e.Attempted && e.Returned && !e.Error && !e.Panic }
func constructorAdd(c *constructorCounts, e constructorEvent) {
	if e.Attempted {
		c.Attempted++
		if strings.HasSuffix(e.API, "_prepare") {
			c.Prepares++
		}
		if strings.HasSuffix(e.API, "_rank") {
			c.Ranks++
		}
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
func constructorMeasure(fn func()) constructorCost {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	fn()
	elapsed := time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&after)
	return constructorCost{elapsed, after.TotalAlloc - before.TotalAlloc, after.Mallocs - before.Mallocs}
}
func constructorHex(v float64) string {
	var b [8]byte
	u := math.Float64bits(v)
	for i := 7; i >= 0; i-- {
		b[i] = byte(u)
		u >>= 8
	}
	return hex.EncodeToString(b[:])
}
func constructorSnap(r shortclaim.Ranking) constructorRanking {
	s := constructorRanking{Kind: r.Kind, Count: r.Count, Order: r.Order, Fallback: r.FallbackReason, CandidateIDs: [5]string{"post_path", "post_query", "pre_path", "twice_path", "unchanged"}}
	for i, v := range r.Scores {
		s.ScoreBits[i] = constructorHex(v)
	}
	for i := 0; i < 5 && i < r.Count; i++ {
		x := r.Order[i]
		if x >= 0 && x < 5 {
			s.OrderedIDs[i] = s.CandidateIDs[x]
		}
	}
	return s
}
func constructorFrozen() constructorRanking {
	r := shortclaim.Ranking{Kind: Kind, Count: 5, Order: [8]int{2, 4, 1, 3, 0}}
	for i, b := range [5]uint64{0xbfc4dd509998bba7, 0xbfc40874752fd85e, 0xbfb69c0f88bf8dcb, 0xbfc45cd2f7e3a36f, 0xbfbc36f53a763f34} {
		r.Scores[i] = math.Float64frombits(b)
	}
	return constructorSnap(r)
}
func constructorOwner(p *Prepared, rows bool) *constructorPayload {
	if p == nil {
		return nil
	}
	o := &constructorPayload{Count: p.count, Ready: p.ready, Offsets: p.offsets, FeatureLenCap: [2]int{len(p.features), cap(p.features)}, FeatureBytes: len(p.features) * int(unsafe.Sizeof(hintweights.Feature{})), OwnerFixedBytes: int(unsafe.Sizeof(Prepared{})), ClonedTextBytes: len(p.request)}
	for _, s := range p.texts {
		o.ClonedTextBytes += len(s)
	}
	if rows && p.count >= 1 && p.count <= 8 {
		o.Rows = make([][]constructorBits, p.count)
		for i := 0; i < p.count; i++ {
			a, b := int(p.offsets[i]), int(p.offsets[i+1])
			if a > b || b > len(p.features) {
				o.Rows = nil
				break
			}
			o.Rows[i] = make([]constructorBits, b-a)
			for j, f := range p.features[a:b] {
				o.Rows[i][j] = constructorBits{f.Index, constructorHex(f.Value)}
			}
		}
	}
	return o
}
func constructorSameOwner(p, w *Prepared) bool {
	if p == nil || w == nil || !p.ready || p.count != w.count || p.request != w.request || p.texts != w.texts || p.offsets != w.offsets || len(p.features) != len(w.features) || cap(p.features) != len(p.features) {
		return false
	}
	for i, f := range p.features {
		g := w.features[i]
		if f.Index != g.Index || math.Float64bits(f.Value) != math.Float64bits(g.Value) {
			return false
		}
	}
	return true
}
func constructorOwns(p *Prepared, in shortclaim.Prepared) bool {
	if p == nil || p.request != in.Request || p.count != in.Count || unsafe.StringData(p.request) == unsafe.StringData(in.Request) {
		return false
	}
	for i := 0; i < p.count; i++ {
		if p.texts[i] != in.Candidates[i].Text || unsafe.StringData(p.texts[i]) == unsafe.StringData(in.Candidates[i].Text) {
			return false
		}
	}
	return true
}

// Fixed raw arrays retain every actual returned Ranking inside the timer.
// Owner/feature/rank conversion and parity are deliberately outside it.
func constructorInterval(n, pair, position int, method string, in shortclaim.Prepared, w *Prepared, anchor constructorRanking, prepare func() (*Prepared, error), rank func(*Prepared) (shortclaim.Ranking, error), measure func(func()) constructorCost) constructorTrial {
	x := constructorTrial{N: n, Pair: pair, Position: position, Method: method, State: "unavailable", Calls: []constructorResult{}, RemainingRanks: n}
	if n < 1 || n > 8 || (method != "legacy" && method != "scratch") {
		return x
	}
	var owner *Prepared
	var events [8]constructorEvent
	var ranks [8]shortclaim.Ranking
	calls := 0
	x.Cost = measure(func() {
		owner, x.Prepare = constructorCall(method+"_prepare", prepare)
		if !constructorOK(x.Prepare) || owner == nil {
			return
		}
		for i := 0; i < n; i++ {
			ranks[calls], events[calls] = constructorCall(method+"_rank", func() (shortclaim.Ranking, error) { return rank(owner) })
			calls++
			if !constructorOK(events[calls-1]) {
				return
			}
		}
	})
	constructorAdd(&x.Counts, x.Prepare)
	x.RemainingRanks = n - calls
	x.Owner = constructorOwner(owner, false)
	x.OwnerParity = constructorSameOwner(owner, w)
	x.OwnsInput = constructorOwns(owner, in)
	complete := constructorOK(x.Prepare) && owner != nil && calls == n
	parity := x.OwnerParity && x.OwnsInput
	for i := 0; i < calls; i++ {
		r := constructorResult{Event: events[i]}
		if events[i].Returned {
			s := constructorSnap(ranks[i])
			r.Ranking = &s
			m := s == anchor
			r.MatchesAnchor = &m
			parity = parity && m
		}
		constructorAdd(&x.Counts, events[i])
		x.Calls = append(x.Calls, r)
		complete = complete && constructorOK(events[i])
	}
	if complete {
		x.State = "completed"
		if !parity {
			x.State = "parity_failure"
		}
	}
	runtime.KeepAlive(owner)
	return x
}
func constructorRead(path string, size int, pin string) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, errors.New("read")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(size+1)))
	closeErr := f.Close()
	h := sha256.Sum256(b)
	if e != nil || closeErr != nil || len(b) != size || hex.EncodeToString(h[:]) != pin {
		return nil, errors.New("pin")
	}
	return b, nil
}
func constructorStageCall[T any](r *constructorReport, name string, fn func() (T, error)) (out T, ok bool) {
	var e constructorEvent
	cost := constructorMeasure(func() { out, e = constructorCall(name, fn) })
	r.Stages = append(r.Stages, constructorStage{name, e, cost})
	constructorAdd(&r.Totals, e)
	return out, constructorOK(e)
}
func constructorRun(inPath, modelPath string) constructorReport {
	r := constructorReport{Schema: "riido-chi-constructor-scratch-v1", State: "unavailable", Scope: constructorScope, InputSHA: constructorInputPin, ModelSHA: constructorModelPin, ModelRef: "JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46", Parents: 1, RuntimeGo: runtime.Version(), GOMAXPROCS: 1, MemoryLimitBytes: 96 << 20, Stages: []constructorStage{}, Trials: []constructorTrial{}}
	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)
	oldLimit := debug.SetMemoryLimit(96 << 20)
	defer debug.SetMemoryLimit(oldLimit)
	input, ok := constructorStageCall(&r, "input_read", func() ([]byte, error) { return constructorRead(inPath, 1417, constructorInputPin) })
	if !ok {
		return r
	}
	raw, ok := constructorStageCall(&r, "model_read", func() ([]byte, error) { return constructorRead(modelPath, 32792, constructorModelPin) })
	if !ok {
		return r
	}
	current, ok := constructorStageCall(&r, "load_validated", func() (shortclaim.ValidatedInput, error) { return shortclaim.LoadValidated(bytes.NewReader(input)) })
	if !ok {
		return r
	}
	view, ok := constructorStageCall(&r, "view_new", func() (*hintweights.View, error) { return hintweights.New(raw) })
	if !ok {
		return r
	}
	p := current.Prepared()
	legacy, ok := constructorStageCall(&r, "legacy_prepare", func() (*Prepared, error) { return constructorLegacy(p) })
	r.Legacy = constructorOwner(legacy, true)
	if !ok || legacy == nil {
		return r
	}
	scratch, ok := constructorStageCall(&r, "scratch_prepare", func() (*Prepared, error) { return Prepare(p) })
	r.Scratch = constructorOwner(scratch, true)
	if !ok || scratch == nil {
		return r
	}
	a, ok := constructorStageCall(&r, "legacy_rank", func() (shortclaim.Ranking, error) { return legacy.Rank(current, view) })
	if r.Stages[len(r.Stages)-1].Event.Returned {
		s := constructorSnap(a)
		r.LegacyAnchor = &s
	}
	if !ok {
		return r
	}
	b, ok := constructorStageCall(&r, "scratch_rank", func() (shortclaim.Ranking, error) { return scratch.Rank(current, view) })
	if r.Stages[len(r.Stages)-1].Event.Returned {
		s := constructorSnap(b)
		r.ScratchAnchor = &s
	}
	if !ok {
		return r
	}
	want := constructorFrozen()
	if legacy == nil || scratch == nil || !constructorSameOwner(scratch, legacy) || !constructorOwns(legacy, p) || !constructorOwns(scratch, p) || *r.LegacyAnchor != want || *r.ScratchAnchor != want {
		r.State = "parity_failure"
		return r
	}
	r.State = "completed"
	for _, n := range [2]int{1, 8} {
		for pair := 0; pair < 6; pair++ {
			order := [2]string{"legacy", "scratch"}
			if pair%2 == 1 {
				order = [2]string{"scratch", "legacy"}
			}
			for pos, method := range order {
				makeOwner := func() (*Prepared, error) { return constructorLegacy(p) }
				if method == "scratch" {
					makeOwner = func() (*Prepared, error) { return Prepare(p) }
				}
				x := constructorInterval(n, pair, pos, method, p, legacy, want, makeOwner, func(o *Prepared) (shortclaim.Ranking, error) { return o.Rank(current, view) }, constructorMeasure)
				r.Trials = append(r.Trials, x)
				constructorAdd(&r.Totals, x.Prepare)
				for _, c := range x.Calls {
					constructorAdd(&r.Totals, c.Event)
				}
				if x.State != "completed" {
					r.State = "unavailable"
				}
			}
		}
	}
	runtime.KeepAlive(input)
	runtime.KeepAlive(raw)
	runtime.KeepAlive(view)
	runtime.KeepAlive(current)
	runtime.KeepAlive(legacy)
	runtime.KeepAlive(scratch)
	return r
}

func TestConstructorCostProbe(t *testing.T) {
	if *constructorInput == "" && *constructorModel == "" && *constructorOutput == "" {
		t.Skip("explicit constructor flags required; no model calls")
	}
	if *constructorInput == "" || *constructorModel == "" || *constructorOutput == "" {
		t.Fatal("constructor_arguments")
	}
	f, e := os.OpenFile(*constructorOutput, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal("constructor_output")
	}
	r := constructorRun(*constructorInput, *constructorModel)
	e = json.NewEncoder(f).Encode(r)
	syncErr := f.Sync()
	closeErr := f.Close()
	if e != nil || syncErr != nil || closeErr != nil {
		t.Fatal("constructor_retention")
	}
	if r.State != "completed" {
		t.Fatal("constructor_unavailable_report_retained")
	}
}

func TestConstructorProbeControls(t *testing.T) {
	p := shortclaim.Prepared{Request: "x", Count: 1}
	p.Candidates[0].Text = "y"
	owner := &Prepared{request: strings.Clone(p.Request), texts: [8]string{strings.Clone("y")}, count: 1, ready: true}
	rank := shortclaim.Ranking{Kind: Kind, Count: 1, Scores: [8]float64{1}}
	want := constructorSnap(rank)
	measure := func(fn func()) constructorCost { fn(); return constructorCost{} }
	good := func() (*Prepared, error) { return owner, nil }
	goodRank := func(*Prepared) (shortclaim.Ranking, error) { return rank, nil }
	for _, n := range [2]int{1, 8} {
		x := constructorInterval(n, 0, 0, "legacy", p, owner, want, good, goodRank, measure)
		if x.State != "completed" || len(x.Calls) != n || x.RemainingRanks != 0 || x.Counts.Attempted != n+1 || x.Counts.Returned != n+1 {
			t.Fatal("synthetic complete trace")
		}
	}
	wrong := rank
	wrong.Scores[0] = 2
	x := constructorInterval(1, 0, 0, "scratch", p, owner, want, good, func(*Prepared) (shortclaim.Ranking, error) { return wrong, nil }, measure)
	if x.State != "parity_failure" || x.Calls[0].Ranking == nil || *x.Calls[0].MatchesAnchor || x.Counts.Errors != 0 {
		t.Fatal("successful wrong ranking is preserved")
	}
	tests := []struct {
		name                  string
		prepare               func() (*Prepared, error)
		rank                  func(*Prepared) (shortclaim.Ranking, error)
		calls, errors, panics int
	}{
		{"prepare_error", func() (*Prepared, error) { return owner, errors.New("private error must not be serialized") }, goodRank, 0, 1, 0},
		{"prepare_panic_nil", func() (*Prepared, error) { panic(nil) }, goodRank, 0, 0, 1},
		{"nil_owner", func() (*Prepared, error) { return nil, nil }, goodRank, 0, 0, 0},
		{"rank_error", good, func(*Prepared) (shortclaim.Ranking, error) { return rank, errors.New("private error") }, 1, 1, 0},
		{"rank_panic", good, func(*Prepared) (shortclaim.Ranking, error) { panic("private value") }, 1, 0, 1},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			x := constructorInterval(8, 0, 0, "scratch", p, owner, want, c.prepare, c.rank, measure)
			if x.State != "unavailable" || len(x.Calls) != c.calls || x.RemainingRanks != 8-c.calls || x.Counts.Errors != c.errors || x.Counts.Panics != c.panics {
				t.Fatal("failure trace")
			}
			if c.name == "rank_error" && (x.Calls[0].Ranking == nil || *x.Calls[0].Ranking != want || !x.Calls[0].Event.Returned) {
				t.Fatal("returned error ranking lost")
			}
			if c.name == "rank_panic" && (x.Calls[0].Ranking != nil || x.Calls[0].Event.Returned) {
				t.Fatal("panic padded as return")
			}
			b, e := json.Marshal(x)
			if e != nil || bytes.Contains(b, []byte("private")) {
				t.Fatal("diagnostic leak")
			}
		})
	}
}
