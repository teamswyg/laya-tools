// SPDX-License-Identifier: Apache-2.0
// Owned numeric/synthetic API contracts only: not new independent Golden tasks,
// semantic model/source qualification, training, performance or RSS evidence.
// Encode/Decode/New are test-only, in-memory; no model assets/files/network.
package hintprepared

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func apiInput(n int) shortclaim.Input {
	docs := [8]string{"alpha alpha β", "한글 오류 반환", "beta!!!", "unrelated δ", "Foo_bar keeps item", "beta; ALPHA!!!", "repeat repeat repeat", "zeta delta"}
	in := shortclaim.Input{Schema: shortclaim.Schema, Request: "alpha β 한글 alpha", Provenance: "owned-synthetic-v1", Candidates: make([]shortclaim.Candidate, n)}
	for i := range in.Candidates {
		in.Candidates[i] = shortclaim.Candidate{ID: fmt.Sprintf("c%d", i), Text: docs[i%len(docs)]}
	}
	return in
}

func copyAPIInput(in shortclaim.Input) shortclaim.Input {
	in.Candidates = append([]shortclaim.Candidate(nil), in.Candidates...)
	return in
}

func validated(t *testing.T, in shortclaim.Input) shortclaim.ValidatedInput {
	t.Helper()
	v, err := shortclaim.ValidateInput(in)
	if err != nil {
		t.Fatal("owned input", err)
	}
	return v
}

func prepared(t *testing.T, v shortclaim.ValidatedInput) *Prepared {
	t.Helper()
	p, err := Prepare(v.Prepared())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func stored(t *testing.T, w []float64, mode string) (*hintweights.View, []float64) {
	t.Helper()
	b, err := hintlearn.Encode(w, mode)
	if err != nil {
		t.Fatal(err)
	}
	v, err := hintweights.New(b)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := hintlearn.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	return v, decoded
}

func numbers(t *testing.T, mode string, zero bool) (*hintweights.View, []float64) {
	t.Helper()
	w := make([]float64, hintweights.Dimension)
	if !zero {
		for i := range w {
			switch mode {
			case "fp32":
				w[i] = float64(i%7-3) * .125
			case "int8":
				w[i] = float64(i%255-127) * .125
			case "ternary_ptq":
				if i%3 != 0 {
					w[i] = .5
					if i%2 == 0 {
						w[i] = -.5
					}
				}
			default:
				t.Fatal("owned mode")
			}
		}
	}
	return stored(t, w, mode)
}

func wantRank(v shortclaim.ValidatedInput, w []float64) shortclaim.Ranking {
	p := v.Prepared()
	r := shortclaim.Ranking{Kind: Kind, Count: p.Count}
	for i := 0; i < p.Count; i++ {
		r.Order[i] = i
		r.Scores[i] = hintlearn.Score(w, hintlearn.Features(p.Request, p.Candidates[i].Text))
	}
	// Independent reference ordering uses sort rather than SDK insertion sort.
	sort.Slice(r.Order[:r.Count], func(i, j int) bool {
		a, b := r.Order[i], r.Order[j]
		if r.Scores[a] == r.Scores[b] {
			return a < b
		}
		return r.Scores[a] > r.Scores[b]
	})
	return r
}

func sameRanking(a, b shortclaim.Ranking) bool {
	if a.Kind != b.Kind || a.Count != b.Count || a.Order != b.Order || a.FallbackReason != b.FallbackReason {
		return false
	}
	for i, x := range a.Scores {
		if math.Float64bits(x) != math.Float64bits(b.Scores[i]) {
			return false
		}
	}
	return true
}

func TestPrepareRequiresPublicValidation(t *testing.T) {
	base := validated(t, apiInput(1)).Prepared()
	tests := []struct {
		name   string
		mutate func(*shortclaim.Prepared)
		want   error
	}{
		{"zero_count", func(p *shortclaim.Prepared) { p.Count = 0 }, shortclaim.ErrCandidateCount},
		{"nine_count", func(p *shortclaim.Prepared) { p.Count = 9 }, shortclaim.ErrCandidateCount},
		{"request_normalization", func(p *shortclaim.Prepared) { p.NormalizedRequest = "forged" }, shortclaim.ErrPrepared},
		{"text_normalization", func(p *shortclaim.Prepared) { p.Candidates[0].NormalizedText = "forged" }, shortclaim.ErrPrepared},
		{"inactive_id", func(p *shortclaim.Prepared) { p.Candidates[1].ID = "inactive" }, shortclaim.ErrPrepared},
		{"inactive_text", func(p *shortclaim.Prepared) { p.Candidates[1].Text = "inactive" }, shortclaim.ErrPrepared},
		{"inactive_normalization", func(p *shortclaim.Prepared) { p.Candidates[1].NormalizedText = "inactive" }, shortclaim.ErrPrepared},
		{"empty_request", func(p *shortclaim.Prepared) { p.Request = "" }, shortclaim.ErrEmptyText},
		{"empty_text", func(p *shortclaim.Prepared) { p.Candidates[0].Text = "" }, shortclaim.ErrEmptyText},
		{"schema", func(p *shortclaim.Prepared) { p.Schema = "other" }, shortclaim.ErrSchema},
		{"id_ascii", func(p *shortclaim.Prepared) { p.Candidates[0].ID = "한글" }, shortclaim.ErrIdentifier},
		{"empty_id", func(p *shortclaim.Prepared) { p.Candidates[0].ID = "" }, shortclaim.ErrIdentifier},
		{"id_65", func(p *shortclaim.Prepared) { p.Candidates[0].ID = strings.Repeat("i", 65) }, shortclaim.ErrIdentifier},
		{"provenance", func(p *shortclaim.Prepared) { p.Provenance = "private/path" }, shortclaim.ErrIdentifier},
		{"raw_513", func(p *shortclaim.Prepared) { p.Request = strings.Repeat("x", 513) }, shortclaim.ErrTextBounds},
		{"text_513", func(p *shortclaim.Prepared) { p.Candidates[0].Text = strings.Repeat("x", 513) }, shortclaim.ErrTextBounds},
		{"request_words_33", func(p *shortclaim.Prepared) { p.Request = strings.Repeat("a ", 32) + "a" }, shortclaim.ErrNormalizedBounds},
		{"text_words_33", func(p *shortclaim.Prepared) { p.Candidates[0].Text = strings.Repeat("a ", 32) + "a" }, shortclaim.ErrNormalizedBounds},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			bad := base
			c.mutate(&bad)
			if err := shortclaim.ValidatePrepared(bad); err != c.want {
				t.Fatal("negative public diagnostic", err)
			}
			p, err := Prepare(bad)
			if p != nil || err != c.want {
				t.Fatal("public validation bypass", err)
			}
		})
	}
	bad := validated(t, apiInput(5)).Prepared()
	bad.Candidates[1].ID = bad.Candidates[0].ID
	if p, err := Prepare(bad); p != nil || err != shortclaim.ErrDuplicateID {
		t.Fatal("duplicate ID diagnostic")
	}
	if p, err := Prepare(shortclaim.Prepared{}); p != nil || err != shortclaim.ErrCandidateCount {
		t.Fatal("zero input diagnostic")
	}
}

func TestReferenceBitsFeatureOrderAndTightStorage(t *testing.T) {
	cases := []shortclaim.Input{apiInput(1), apiInput(5), apiInput(8)}
	bounded := apiInput(1)
	bounded.Request = strings.Repeat("x", 512)
	bounded.Candidates[0].Text = strings.Repeat("y", 512)
	bounded.Candidates[0].ID = strings.Repeat("i", 64)
	words := apiInput(1)
	words.Request = strings.Repeat("a ", 31) + "a"
	words.Candidates[0].Text = strings.Repeat("β ", 31) + "β"
	cases = append(cases, bounded, words)
	for _, in := range cases {
		v := validated(t, in)
		p := prepared(t, v)
		snapshot := v.Prepared()
		if !p.ready || cap(p.features) != len(p.features) || p.offsets[0] != 0 || int(p.offsets[p.count]) != len(p.features) {
			t.Fatal("tight shape")
		}
		for i := 0; i < snapshot.Count; i++ {
			fs := hintlearn.Features(snapshot.Request, snapshot.Candidates[i].Text)
			row := p.features[p.offsets[i]:p.offsets[i+1]]
			if len(row) != len(fs) {
				t.Fatal("row length")
			}
			for j, f := range fs {
				if row[j].Index != f.Index || math.Float64bits(row[j].Value) != math.Float64bits(f.Value) {
					t.Fatal("feature order/value drift")
				}
			}
			if len(fs) > 0 && fs[len(fs)-1].Index != 0 {
				t.Fatal("final overlap feature")
			}
		}
		for _, mode := range [3]string{"fp32", "int8", "ternary_ptq"} {
			view, w := numbers(t, mode, false)
			got, err := p.Rank(v, view)
			if err != nil || !sameRanking(got, wantRank(v, w)) {
				t.Fatal("reference bits/order", mode, err)
			}
			for i := got.Count; i < shortclaim.MaxCandidates; i++ {
				if got.Order[i] != 0 || math.Float64bits(got.Scores[i]) != 0 {
					t.Fatal("inactive result tail")
				}
			}
		}
		// Declared fixed header/array size is separate from tight feature payload
		// and cloned text lengths. No allocation/RSS assertion follows arithmetic.
		fixed := int(unsafe.Sizeof(Prepared{}))
		payload := len(p.features) * int(unsafe.Sizeof(hintweights.Feature{}))
		text := len(p.request)
		for i := 0; i < p.count; i++ {
			text += len(p.texts[i])
		}
		if fixed <= 0 || payload < 0 || text != len(snapshot.Request)+sumTexts(snapshot) {
			t.Fatal("logical storage components")
		}
		for i := p.count; i < shortclaim.MaxCandidates; i++ {
			if p.texts[i] != "" || p.offsets[i+1] != 0 {
				t.Fatal("inactive owner slot")
			}
		}
	}
}

func sumTexts(p shortclaim.Prepared) int {
	n := 0
	for i := 0; i < p.Count; i++ {
		n += len(p.Candidates[i].Text)
	}
	return n
}

func TestExactMismatchAndMetadataOnlyChanges(t *testing.T) {
	in := apiInput(5)
	current := validated(t, in)
	p := prepared(t, current)
	view, w := numbers(t, "fp32", false)
	base, err := p.Rank(current, view)
	if err != nil {
		t.Fatal(err)
	}
	edits := [6]shortclaim.Input{copyAPIInput(in), copyAPIInput(in), copyAPIInput(in), copyAPIInput(in), copyAPIInput(in), copyAPIInput(in)}
	edits[0].Request = strings.ToUpper(in.Request)
	edits[1].Request += " "
	edits[2].Request += "!!!"
	edits[3].Candidates[0].Text += " revised"
	edits[4].Candidates[0], edits[4].Candidates[1] = edits[4].Candidates[1], edits[4].Candidates[0]
	edits[5].Candidates = edits[5].Candidates[:4]
	for _, edit := range edits {
		got, err := p.Rank(validated(t, edit), view)
		if err != ErrMismatch || got != (shortclaim.Ranking{}) {
			t.Fatal("changed exact input accepted", err)
		}
	}
	meta := copyAPIInput(in)
	meta.Provenance = "owned-synthetic-v2"
	for i := range meta.Candidates {
		meta.Candidates[i].ID = fmt.Sprintf("new%d", len(meta.Candidates)-i)
	}
	changed := validated(t, meta)
	got, err := p.Rank(changed, view)
	if err != nil || !sameRanking(got, base) || !sameRanking(got, wantRank(changed, w)) {
		t.Fatal("metadata entered features", err)
	}
	now := changed.Prepared()
	for _, index := range got.Order[:got.Count] {
		if now.Candidates[index].ID != meta.Candidates[index].ID {
			t.Fatal("current ID mapping")
		}
	}
	// Mutating exported copies/caller slices neither alters opaque current nor p.
	exported := current.Prepared()
	exported.Request = "changed"
	exported.Candidates[0].Text = "changed"
	in.Request = "changed"
	in.Candidates[0].Text = "changed"
	again, err := p.Rank(current, view)
	if err != nil || !sameRanking(again, base) {
		t.Fatal("caller mutation leaked")
	}
	if p.request == in.Request || p.texts[0] == in.Candidates[0].Text {
		t.Fatal("owner raw binding changed")
	}
}

func TestTieCurrentPositionsAndViewRecomputation(t *testing.T) {
	in := apiInput(8)
	for i := range in.Candidates {
		in.Candidates[i].ID = fmt.Sprintf("z%d", 8-i)
	}
	current := validated(t, in)
	p := prepared(t, current)
	zero, _ := numbers(t, "fp32", true)
	first, err := p.Rank(current, zero)
	if err != nil {
		t.Fatal(err)
	}
	kept := first
	for i := 0; i < first.Count; i++ {
		if first.Order[i] != i || math.Float64bits(first.Scores[i]) != 0 {
			t.Fatal("tie ordered by ID")
		}
	}
	permuted := copyAPIInput(in)
	order := [8]int{7, 2, 0, 5, 1, 6, 4, 3}
	for i, j := range order {
		permuted.Candidates[i] = in.Candidates[j]
	}
	next := validated(t, permuted)
	q := prepared(t, next)
	rank, err := q.Rank(next, zero)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < rank.Count; i++ {
		if rank.Order[i] != i {
			t.Fatal("permutation lost tie positions")
		}
	}
	// Query alpha gives literal final overlap values [1,0] for alpha,beta.
	change := shortclaim.Input{Schema: shortclaim.Schema, Request: "alpha", Provenance: "owned-v1", Candidates: []shortclaim.Candidate{{ID: "z", Text: "alpha"}, {ID: "a", Text: "beta"}}}
	c := validated(t, change)
	a := prepared(t, c)
	for i := 0; i < 2; i++ {
		last := a.features[a.offsets[i+1]-1]
		if last.Index != 0 || last.Value != float64(1-i) {
			t.Fatal("literal overlap Want")
		}
	}
	var previous shortclaim.Ranking
	for _, coefficient := range [2]float64{2, -2} {
		w := make([]float64, hintweights.Dimension)
		w[0] = coefficient
		view, _ := stored(t, w, "fp32")
		x, err := a.Rank(c, view)
		wantOrder := [8]int{0, 1}
		if coefficient < 0 {
			wantOrder = [8]int{1, 0}
		}
		if err != nil || x.Order != wantOrder || math.Float64bits(x.Scores[0]) != math.Float64bits(coefficient) || math.Float64bits(x.Scores[1]) != 0 {
			t.Fatal("stale model score/order", err)
		}
		if coefficient > 0 {
			previous = x
		} else if previous.Scores[0] != 2 || previous.Order[0] != 0 {
			t.Fatal("previous value changed")
		}
	}
	if !sameRanking(first, kept) {
		t.Fatal("older independent result changed")
	}
	// Returned arrays are values; caller edits cannot become persistent cache state.
	first.Scores[0] = 999
	first.Order[0] = 7
	again, err := p.Rank(current, zero)
	if err != nil || !sameRanking(again, kept) {
		t.Fatal("result mutation leaked")
	}
}

func TestFixedZeroErrors(t *testing.T) {
	current := validated(t, apiInput(1))
	p := prepared(t, current)
	view, _ := numbers(t, "fp32", false)
	tests := []struct {
		p     *Prepared
		input shortclaim.ValidatedInput
		view  *hintweights.View
		want  error
	}{
		{nil, current, view, ErrPrepared}, {&Prepared{}, current, view, ErrPrepared},
		{p, shortclaim.ValidatedInput{}, view, ErrInput}, {p, current, nil, ErrView}, {p, current, &hintweights.View{}, ErrView},
	}
	for _, c := range tests {
		got, err := c.p.Rank(c.input, c.view)
		if err != c.want || got != (shortclaim.Ranking{}) {
			t.Fatal("nonzero failure result", err)
		}
	}
	// Private corruption control forces nonfinite arithmetic with finite weights;
	// this is not an externally forgeable owner or a semantic training example.
	corrupt := prepared(t, current)
	corrupt.features[0] = hintweights.Feature{Index: 0, Value: math.MaxFloat64}
	w := make([]float64, hintweights.Dimension)
	w[0] = 2
	large, _ := stored(t, w, "fp32")
	got, err := corrupt.Rank(current, large)
	if err != ErrScore || got != (shortclaim.Ranking{}) {
		t.Fatal("nonfinite error", err)
	}
	corrupt.features[0].Index = hintweights.Dimension
	got, err = corrupt.Rank(current, view)
	if err != ErrFeature || got != (shortclaim.Ranking{}) {
		t.Fatal("feature error", err)
	}
}

func TestConcurrentReadonlyOwnerAndValueOutputs(t *testing.T) {
	current := validated(t, apiInput(8))
	p := prepared(t, current)
	view, w := numbers(t, "fp32", false)
	want := wantRank(current, w)
	before := append([]hintweights.Feature(nil), p.features...)
	failures := make(chan string, 8)
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 16; j++ {
				got, err := p.Rank(current, view)
				if err != nil || !sameRanking(got, want) {
					failures <- "concurrent reference mismatch"
					return
				}
				got.Scores[0] = 123
				got.Order[0] = 7 // Local value mutation, no caller scratch.
			}
		}()
	}
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
	for i, f := range before {
		if p.features[i].Index != f.Index || math.Float64bits(p.features[i].Value) != math.Float64bits(f.Value) {
			t.Fatal("owner mutated")
		}
	}
	got, err := p.Rank(current, view)
	if err != nil || !sameRanking(got, want) {
		t.Fatal("persistent result mutation")
	}
}
