// SPDX-License-Identifier: Apache-2.0
// Offline audit of the frozen cost trace; never loads a learned model.
package hintprepared

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

var constructorSaved = flag.String("constructor-saved", "", "explicit saved constructor gzip trace; no model calls")

// Fixed arrays keep duplicate-key detection independent of map overwrite.
func constructorSavedKeys(d *json.Decoder, depth int) error {
	if depth > 32 {
		return errors.New("depth")
	}
	t, e := d.Token()
	if e != nil {
		return e
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		var keys [64]string
		n := 0
		for d.More() {
			t, e := d.Token()
			if e != nil {
				return e
			}
			key, ok := t.(string)
			if !ok || n == len(keys) {
				return errors.New("keys")
			}
			for i := 0; i < n; i++ {
				if keys[i] == key {
					return errors.New("duplicate")
				}
			}
			keys[n] = key
			n++
			if e := constructorSavedKeys(d, depth+1); e != nil {
				return e
			}
		}
		t, e := d.Token()
		if e != nil || t != json.Delim('}') {
			return errors.New("object")
		}
	case '[':
		for d.More() {
			if e := constructorSavedKeys(d, depth+1); e != nil {
				return e
			}
		}
		t, e := d.Token()
		if e != nil || t != json.Delim(']') {
			return errors.New("array")
		}
	default:
		return errors.New("delimiter")
	}
	return nil
}

// Every canonical field is required; trial Rows deliberately encodes null.
// Map use is confined to cold audit parsing, never the runtime feature loop.
func constructorSavedShape(raw []byte, typ reflect.Type) error {
	raw = bytes.TrimSpace(raw)
	fail := errors.New("shape")
	if typ.Kind() == reflect.Pointer {
		if bytes.Equal(raw, []byte("null")) {
			return nil
		}
		return constructorSavedShape(raw, typ.Elem())
	}
	if bytes.Equal(raw, []byte("null")) {
		if typ.Kind() == reflect.Slice {
			return nil
		}
		return fail
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || len(fields) != typ.NumField() {
			return fail
		}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			v, ok := fields[f.Name]
			if !ok || constructorSavedShape(v, f.Type) != nil {
				return fail
			}
		}
	case reflect.Array, reflect.Slice:
		var items []json.RawMessage
		if json.Unmarshal(raw, &items) != nil || (typ.Kind() == reflect.Array && len(items) != typ.Len()) {
			return fail
		}
		for _, v := range items {
			if constructorSavedShape(v, typ.Elem()) != nil {
				return fail
			}
		}
	}
	return nil
}

func constructorSavedDecode(raw []byte) (constructorReport, error) {
	var r constructorReport
	fail := errors.New("saved_constructor_audit")
	if len(raw) == 0 || len(raw) > 2<<20 || !utf8.Valid(raw) {
		return r, fail
	}
	k := json.NewDecoder(bytes.NewReader(raw))
	k.UseNumber()
	if constructorSavedKeys(k, 0) != nil {
		return r, fail
	}
	if _, e := k.Token(); e != io.EOF {
		return r, fail
	}
	if constructorSavedShape(raw, reflect.TypeOf(r)) != nil {
		return r, fail
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&r) != nil {
		return constructorReport{}, fail
	}
	var more any
	if d.Decode(&more) != io.EOF || constructorSavedAudit(r) != nil {
		return constructorReport{}, fail
	}
	return r, nil
}

func constructorSavedEvent(e constructorEvent, api string) bool {
	return e.API == api && e.Code == "" && e.Attempted && e.Returned && !e.Error && !e.Panic
}
func constructorSavedPayload(p *constructorPayload, full bool) bool {
	if p == nil || p.Count != 5 || !p.Ready || p.FeatureLenCap[0] < 1 || p.FeatureLenCap[0] != p.FeatureLenCap[1] || p.FeatureLenCap[0] > 5*8192 || p.FeatureBytes != 16*p.FeatureLenCap[0] || p.ClonedTextBytes < 1 || p.OwnerFixedBytes < 1 || p.Offsets[0] != 0 || int(p.Offsets[5]) != p.FeatureLenCap[0] {
		return false
	}
	for i := 6; i < 9; i++ {
		if p.Offsets[i] != 0 {
			return false
		}
	}
	if !full {
		return p.Rows == nil
	}
	if len(p.Rows) != 5 {
		return false
	}
	for i, row := range p.Rows {
		if p.Offsets[i] > p.Offsets[i+1] || p.Offsets[i+1]-p.Offsets[i] > 8192 || len(row) != int(p.Offsets[i+1]-p.Offsets[i]) || len(row) < 1 {
			return false
		}
		previous := 0
		for j, f := range row {
			if f.Index < 0 || f.Index >= 8192 || len(f.ValueBits) != 16 {
				return false
			}
			for _, c := range f.ValueBits {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					return false
				}
			}
			bits, e := strconv.ParseUint(f.ValueBits, 16, 64)
			v := math.Float64frombits(bits)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return false
			}
			if j == len(row)-1 {
				if f.Index != 0 || v < 0 || v > 1 {
					return false
				}
			} else {
				if f.Index <= previous {
					return false
				}
				previous = f.Index
			}
		}
	}
	return true
}
func constructorSavedAudit(r constructorReport) error {
	fail := errors.New("saved_constructor_audit")
	if r.Schema != "riido-chi-constructor-scratch-v1" || r.State != "completed" || r.Scope != constructorScope || r.InputSHA != constructorInputPin || r.ModelSHA != constructorModelPin || r.ModelRef != "JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46" || r.Parents != 1 || r.RuntimeGo != "go1.27.1" || r.GOMAXPROCS != 1 || r.MemoryLimitBytes != 96<<20 || r.Fit != 0 || r.NewLabels != 0 || r.NewRoles != 0 || r.Qualified || r.Activated || r.ProtectedEvaluation {
		return fail
	}
	want := constructorFrozen()
	if r.LegacyAnchor == nil || r.ScratchAnchor == nil || *r.LegacyAnchor != want || *r.ScratchAnchor != want || !constructorSavedPayload(r.Legacy, true) || !constructorSavedPayload(r.Scratch, true) || !reflect.DeepEqual(r.Legacy, r.Scratch) {
		return fail
	}
	names := [8]string{"input_read", "model_read", "load_validated", "view_new", "legacy_prepare", "scratch_prepare", "legacy_rank", "scratch_rank"}
	if len(r.Stages) != len(names) || len(r.Trials) != 24 {
		return fail
	}
	var counts constructorCounts
	for i, s := range r.Stages {
		if s.Name != names[i] || !constructorSavedEvent(s.Event, names[i]) || s.Cost.ElapsedNS <= 0 {
			return fail
		}
		constructorAdd(&counts, s.Event)
	}
	summary := *r.Legacy
	summary.Rows = nil
	index := 0
	for _, n := range [2]int{1, 8} {
		for pair := 0; pair < 6; pair++ {
			order := [2]string{"legacy", "scratch"}
			if pair%2 == 1 {
				order = [2]string{"scratch", "legacy"}
			}
			for pos, method := range order {
				x := r.Trials[index]
				index++
				if x.N != n || x.Pair != pair || x.Position != pos || x.Method != method || x.State != "completed" || !constructorSavedEvent(x.Prepare, method+"_prepare") || x.Owner == nil || !constructorSavedPayload(x.Owner, false) || !reflect.DeepEqual(*x.Owner, summary) || !x.OwnerParity || !x.OwnsInput || x.RemainingRanks != 0 || len(x.Calls) != n || x.Cost.ElapsedNS <= 0 || x.Cost.TotalAllocBytes == 0 || x.Cost.Mallocs == 0 {
					return fail
				}
				var local constructorCounts
				constructorAdd(&local, x.Prepare)
				for _, c := range x.Calls {
					if !constructorSavedEvent(c.Event, method+"_rank") || c.Ranking == nil || *c.Ranking != want || c.MatchesAnchor == nil || !*c.MatchesAnchor {
						return fail
					}
					constructorAdd(&local, c.Event)
				}
				if local != x.Counts {
					return fail
				}
				constructorAdd(&counts, x.Prepare)
				for _, c := range x.Calls {
					constructorAdd(&counts, c.Event)
				}
			}
		}
	}
	if counts != r.Totals || counts != (constructorCounts{Attempted: 140, Returned: 140, Prepares: 26, Ranks: 110}) {
		return fail
	}
	return nil
}

// Gates were written before actual execution. A failing cost gate remains a
// valid recorded experiment; it does not authorize adopting the new constructor.
func constructorSavedGates(r constructorReport) (bool, [2]float64) {
	pass := true
	var medians [2]float64
	for group := 0; group < 2; group++ {
		var ratios [6]float64
		for pair := 0; pair < 6; pair++ {
			a, b := r.Trials[group*12+pair*2], r.Trials[group*12+pair*2+1]
			if a.Method != "legacy" {
				a, b = b, a
			}
			ratios[pair] = float64(b.Cost.ElapsedNS) / float64(a.Cost.ElapsedNS)
			if float64(b.Cost.TotalAllocBytes)/float64(a.Cost.TotalAllocBytes) > .8 {
				pass = false
			}
		}
		sort.Float64s(ratios[:])
		medians[group] = (ratios[2] + ratios[3]) / 2
		if medians[group] > 1.25 {
			pass = false
		}
	}
	return pass, medians
}

func constructorSavedInflate(compressed []byte) ([]byte, error) {
	fail := errors.New("saved_gzip_frame")
	if len(compressed) == 0 || len(compressed) > 131072 {
		return nil, fail
	}
	reader := bytes.NewReader(compressed)
	z, e := gzip.NewReader(reader)
	if e != nil {
		return nil, fail
	}
	z.Multistream(false)
	b, e := io.ReadAll(io.LimitReader(z, (2<<20)+1))
	closeErr := z.Close()
	if e != nil || closeErr != nil || len(b) > 2<<20 || reader.Len() != 0 {
		return nil, fail
	}
	return b, nil
}

func constructorSavedWitness() (*constructorPayload, error) {
	// Go test runs in this package directory. This is the public frozen fixture,
	// not a model, private prompt, protected row or new independent task.
	b, e := constructorRead("../../experiments/short-claim/next-cohort-chi-audit/preview/INPUT.public.v3.json", 1417, constructorInputPin)
	if e != nil {
		return nil, e
	}
	in, e := shortclaim.LoadValidated(bytes.NewReader(b))
	if e != nil {
		return nil, errors.New("witness_input")
	}
	p := in.Prepared()
	w := &constructorPayload{Count: p.Count, Ready: true, ClonedTextBytes: len(p.Request), OwnerFixedBytes: int(reflect.TypeOf(Prepared{}).Size()), Rows: make([][]constructorBits, p.Count)}
	for i := 0; i < p.Count; i++ {
		w.ClonedTextBytes += len(p.Candidates[i].Text)
		fs := hintlearn.Features(p.Request, p.Candidates[i].Text) // unchanged independent oracle
		w.Rows[i] = make([]constructorBits, len(fs))
		for j, f := range fs {
			w.Rows[i][j] = constructorBits{Index: f.Index, ValueBits: constructorHex(f.Value)}
		}
		w.Offsets[i+1] = w.Offsets[i] + uint32(len(fs))
	}
	w.FeatureLenCap = [2]int{int(w.Offsets[p.Count]), int(w.Offsets[p.Count])}
	w.FeatureBytes = w.FeatureLenCap[0] * int(reflect.TypeOf(hintlearn.Feature{}).Size())
	return w, nil
}
func constructorSavedMatchesWitness(r constructorReport, w *constructorPayload) bool {
	return w != nil && reflect.DeepEqual(r.Legacy, w) && reflect.DeepEqual(r.Scratch, w)
}

func TestConstructorSaved(t *testing.T) {
	if *constructorSaved == "" {
		t.Skip("explicit saved trace required; no model calls")
	}
	if *constructorInput != "" || *constructorModel != "" || *constructorOutput != "" {
		t.Fatal("saved_arguments")
	}
	f, e := os.Open(*constructorSaved)
	if e != nil {
		t.Fatal("saved_read")
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > 131072 {
		t.Fatal("saved_compressed_bounds")
	}
	compressed, e := io.ReadAll(io.LimitReader(f, 131073))
	if e != nil || len(compressed) > 131072 {
		t.Fatal("saved_read_bound")
	}
	b, e := constructorSavedInflate(compressed)
	if e != nil {
		t.Fatal(e)
	}
	r, e := constructorSavedDecode(b)
	if e != nil {
		t.Fatal(e)
	}
	witness, e := constructorSavedWitness()
	if e != nil || !constructorSavedMatchesWitness(r, witness) {
		t.Fatal("independent_public_feature_witness")
	}
	pass, medians := constructorSavedGates(r)
	t.Logf("full traces and independent public feature witness checked; adoption gates=%v paired scratch/legacy medians=%v; no model calls", pass, medians)
}

func TestConstructorSavedRejectsMalformed(t *testing.T) {
	// These do not need or inspect the real model or cost fixture.
	for _, b := range [][]byte{nil, []byte(`{}`), []byte(`{"Schema":"x","Schema":"x"}`), []byte(`{"Schema":"x"} {}`), []byte(`null`), []byte(`{"Schema":null}`)} {
		if _, e := constructorSavedDecode(b); e == nil {
			t.Fatal("malformed audit accepted")
		}
	}
}

func constructorSyntheticReport() constructorReport {
	p := &constructorPayload{Count: 5, Ready: true, FeatureLenCap: [2]int{10, 10}, FeatureBytes: 160, ClonedTextBytes: 6, OwnerFixedBytes: 1, Rows: make([][]constructorBits, 5)}
	for i := 0; i < 5; i++ {
		p.Offsets[i+1] = uint32((i + 1) * 2)
		p.Rows[i] = []constructorBits{{Index: i + 1, ValueBits: "0000000000000000"}, {Index: 0, ValueBits: "0000000000000000"}}
	}
	a := constructorFrozen()
	r := constructorReport{Schema: "riido-chi-constructor-scratch-v1", State: "completed", Scope: constructorScope, InputSHA: constructorInputPin, ModelSHA: constructorModelPin, ModelRef: "JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46", Parents: 1, RuntimeGo: "go1.27.1", GOMAXPROCS: 1, MemoryLimitBytes: 96 << 20, Legacy: p, Scratch: p, LegacyAnchor: &a, ScratchAnchor: &a}
	event := func(api string) constructorEvent { return constructorEvent{API: api, Attempted: true, Returned: true} }
	for _, name := range [8]string{"input_read", "model_read", "load_validated", "view_new", "legacy_prepare", "scratch_prepare", "legacy_rank", "scratch_rank"} {
		e := event(name)
		r.Stages = append(r.Stages, constructorStage{Name: name, Event: e, Cost: constructorCost{ElapsedNS: 1}})
		constructorAdd(&r.Totals, e)
	}
	for _, n := range [2]int{1, 8} {
		for pair := 0; pair < 6; pair++ {
			order := [2]string{"legacy", "scratch"}
			if pair%2 == 1 {
				order = [2]string{"scratch", "legacy"}
			}
			for pos, method := range order {
				owner := *p
				owner.Rows = nil
				x := constructorTrial{N: n, Pair: pair, Position: pos, Method: method, State: "completed", Prepare: event(method + "_prepare"), Owner: &owner, OwnerParity: true, OwnsInput: true, Cost: constructorCost{ElapsedNS: 1000, TotalAllocBytes: 1000, Mallocs: 1}}
				if method == "scratch" {
					x.Cost.TotalAllocBytes = 800
				}
				constructorAdd(&x.Counts, x.Prepare)
				constructorAdd(&r.Totals, x.Prepare)
				for i := 0; i < n; i++ {
					m := true
					e := event(method + "_rank")
					x.Calls = append(x.Calls, constructorResult{Event: e, Ranking: &a, MatchesAnchor: &m})
					constructorAdd(&x.Counts, e)
					constructorAdd(&r.Totals, e)
				}
				r.Trials = append(r.Trials, x)
			}
		}
	}
	return r
}

func TestConstructorSavedSyntheticValidAndCounterexamples(t *testing.T) {
	// A fabricated protocol fixture, not actual model/ranking/performance evidence.
	r := constructorSyntheticReport()
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := constructorSavedDecode(b); e != nil {
		t.Fatal("valid synthetic protocol rejected")
	}
	for _, bad := range [][]byte{append([]byte(`{"Schema":"duplicate",`), b[1:]...), append([]byte(`{"Unexpected":1,`), b[1:]...), bytes.Replace(b, []byte(`"Schema":`), []byte(`"schema":`), 1), bytes.Replace(b, []byte(`"Parents":1`), []byte(`"Parents":null`), 1)} {
		if _, e := constructorSavedDecode(bad); e == nil {
			t.Fatal("strict shape counterexample accepted")
		}
	}
	for _, bad := range [][]byte{bytes.Replace(b, []byte(`,"Fit":0`), nil, 1), bytes.Replace(b, []byte(`"Error":false`), []byte(`"Error":null`), 1), bytes.Replace(b, []byte(`"Order":[2,4,1,3,0,0,0,0]`), []byte(`"Order":[2,4,1,3,0,0,0]`), 1)} {
		if _, e := constructorSavedDecode(bad); e == nil {
			t.Fatal("missing zero/nested null/short array accepted")
		}
	}
	var coherent constructorReport
	if json.Unmarshal(b, &coherent) != nil {
		t.Fatal("fixture")
	}
	coherent.Legacy.Rows[0][0].ValueBits = "3fe0000000000000"
	coherent.Scratch.Rows[0][0].ValueBits = "3fe0000000000000"
	if constructorSavedMatchesWitness(coherent, r.Legacy) {
		t.Fatal("mutual agreement substituted for independent witness")
	}
	mutations := []func(*constructorReport){
		func(x *constructorReport) { x.Trials = x.Trials[:23] },
		func(x *constructorReport) { x.Trials[0].Calls[0].Ranking = nil },
		func(x *constructorReport) { x.Trials[0].Calls[0].Event.Returned = false },
		func(x *constructorReport) { x.Trials[0].Owner.Offsets[1]++ },
		func(x *constructorReport) { x.Trials[0].Counts.Ranks++ },
		func(x *constructorReport) { x.Totals.Attempted++ },
		func(x *constructorReport) { x.Legacy.Rows[0][0].ValueBits = "7ff0000000000000" },
		func(x *constructorReport) { x.LegacyAnchor.Order[0] = 0 },
	}
	for _, mutate := range mutations {
		var x constructorReport
		if json.Unmarshal(b, &x) != nil {
			t.Fatal("fixture")
		}
		mutate(&x)
		raw, e := json.Marshal(x)
		if e != nil {
			t.Fatal(e)
		}
		if _, e := constructorSavedDecode(raw); e == nil {
			t.Fatal("trace/parity counterexample accepted")
		}
	}
	pass, medians := constructorSavedGates(r)
	if !pass || medians != ([2]float64{1, 1}) {
		t.Fatal("frozen gate boundary")
	}
	r.Trials[1].Cost.TotalAllocBytes = 801
	if pass, _ := constructorSavedGates(r); pass {
		t.Fatal("allocation gate relaxed")
	}
	r = constructorSyntheticReport()
	for i := range r.Trials {
		if r.Trials[i].Method == "scratch" {
			r.Trials[i].Cost.ElapsedNS = 1250
		}
	}
	if pass, _ := constructorSavedGates(r); !pass {
		t.Fatal("time gate boundary rejected")
	}
	for i := range r.Trials {
		if r.Trials[i].Method == "scratch" {
			r.Trials[i].Cost.ElapsedNS = 1251
		}
	}
	if pass, _ := constructorSavedGates(r); pass {
		t.Fatal("time gate relaxed")
	}
	failedGate, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := constructorSavedDecode(failedGate); e != nil {
		t.Fatal("failed cost gate discarded valid observation")
	}
}

func TestConstructorSavedGzipFrameControls(t *testing.T) {
	compress := func(b []byte) []byte {
		var out bytes.Buffer
		z := gzip.NewWriter(&out)
		if _, e := z.Write(b); e != nil {
			t.Fatal(e)
		}
		if e := z.Close(); e != nil {
			t.Fatal(e)
		}
		return out.Bytes()
	}
	b, e := json.Marshal(constructorSyntheticReport())
	if e != nil {
		t.Fatal(e)
	}
	gz := compress(b)
	plain, e := constructorSavedInflate(gz)
	if e != nil || !bytes.Equal(plain, b) {
		t.Fatal("synthetic frame rejected")
	}
	crc := append([]byte(nil), gz...)
	crc[len(crc)-8] ^= 1
	bad := [][]byte{nil, gz[:len(gz)-1], crc, append(append([]byte(nil), gz...), ' '), append(append([]byte(nil), gz...), compress([]byte(" \n"))...), compress(bytes.Repeat([]byte("x"), (2<<20)+1)), make([]byte, 131073)}
	for _, raw := range bad {
		if _, e := constructorSavedInflate(raw); e == nil {
			t.Fatal("CRC/EOF/frame/bound counterexample accepted")
		}
	}
}
