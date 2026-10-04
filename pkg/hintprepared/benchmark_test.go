package hintprepared

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// Synthetic opt-in cost probes, executed only with an explicit -bench request.
// Reference includes Rank's weight/raw-word validation, Features, slice output
// and the fixed-value adapter below. Warm uses constructor-validated input and
// prepared features. BM25/lexical use the same validated input's public normalized
// text contract and compute their own nonlearned scores per call. Their scores
// and orders need not equal learned scores. Shortclaim validation is setup for
// all ranking probes; Prepare measures its own validation/cloning/feature work.
// These different valid contracts are not an intrinsic cache/AoS/SoA speed or
// quality comparison, independent evaluation data, or a model activation gate.
// Encoding/New/initial parity are outside every timer; no model file or Fit.
var preparedAPIBenchRankingSink shortclaim.Ranking
var preparedAPIBenchOwnerSink *Prepared

type preparedAPIBenchFixture struct {
	validated shortclaim.ValidatedInput
	input     shortclaim.Prepared
	owner     *Prepared
	view      *hintweights.View
	weights   []float64
}

func preparedAPIBenchReference(w []float64, p shortclaim.Prepared) (shortclaim.Ranking, error) {
	var texts [shortclaim.MaxCandidates]string
	for i := 0; i < p.Count; i++ {
		texts[i] = p.Candidates[i].Text
	}
	order, scores, err := hintlearn.Rank(w, p.Request, texts[:p.Count])
	if err != nil {
		return shortclaim.Ranking{}, err
	}
	if len(order) != p.Count || len(scores) != p.Count {
		return shortclaim.Ranking{}, fmt.Errorf("benchmark reference shape")
	}
	out := shortclaim.Ranking{Kind: "benchmark_reference_full_rank", Count: p.Count}
	copy(out.Order[:], order)
	copy(out.Scores[:], scores)
	return out, nil
}

func preparedAPIBenchSetup(b *testing.B, n int) preparedAPIBenchFixture {
	b.Helper()
	// Original public text: no truth labels, model quality or independent-parent
	// evidence is inferred. The repeated final text also exercises stable ties.
	candidates := [shortclaim.MaxCandidates]shortclaim.Candidate{
		{ID: "c0", Text: "keep active entries and remove expired entries"},
		{ID: "c1", Text: "remove active entries and keep expired entries"},
		{ID: "c2", Text: "keep active entries then remove expired entries"},
		{ID: "c3", Text: "remove expired entries then keep active entries"},
		{ID: "c4", Text: "keep all entries"},
		{ID: "c5", Text: "remove all entries"},
		{ID: "c6", Text: "preserve active entries and remove expired entries"},
		{ID: "c7", Text: "keep active entries and remove expired entries"},
	}
	validated, err := shortclaim.ValidateInput(shortclaim.Input{
		Schema: shortclaim.Schema, Request: "keep active entries and remove expired entries",
		Candidates: candidates[:n], Provenance: "original-synthetic-cost-probe",
	})
	if err != nil {
		b.Fatal(err)
	}
	p := validated.Prepared()
	weights := make([]float64, hintweights.Dimension)
	for i := range weights {
		// Binary fractions are exactly representable before and after FP32 encode.
		weights[i] = float64(i%17-8) / 32
	}
	raw, err := hintlearn.Encode(weights, "fp32")
	if err != nil {
		b.Fatal(err)
	}
	view, err := hintweights.New(raw)
	if err != nil {
		b.Fatal(err)
	}
	owner, err := Prepare(p)
	if err != nil {
		b.Fatal(err)
	}
	reference, err := preparedAPIBenchReference(weights, p)
	if err != nil {
		b.Fatal(err)
	}
	warm, err := owner.Rank(validated, view)
	if err != nil {
		b.Fatal(err)
	}
	if warm.Count != reference.Count || warm.Order != reference.Order || warm.FallbackReason != "" {
		b.Fatal("benchmark order/count parity")
	}
	for i := 0; i < p.Count; i++ {
		fs := hintlearn.Features(p.Request, p.Candidates[i].Text)
		packed := make([]hintweights.Feature, len(fs))
		for j, f := range fs {
			packed[j] = hintweights.Feature{Index: f.Index, Value: f.Value}
		}
		fromView, err := view.Score(packed)
		if err != nil {
			b.Fatal(err)
		}
		bits := math.Float64bits(hintlearn.Score(weights, fs))
		if bits != math.Float64bits(fromView) || bits != math.Float64bits(reference.Scores[i]) || bits != math.Float64bits(warm.Scores[i]) {
			b.Fatal("benchmark FP64 bit parity")
		}
	}
	return preparedAPIBenchFixture{validated, p, owner, view, weights}
}

func BenchmarkPreparedAPIConstruct(b *testing.B) {
	for _, n := range []int{1, 5, 8} {
		b.Run(fmt.Sprintf("candidates_%d", n), func(b *testing.B) {
			b.StopTimer()
			f := preparedAPIBenchSetup(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				owner, err := Prepare(f.input)
				if err != nil {
					b.Fatal(err)
				}
				// Opaque pointer consumption needs no private payload/field access.
				preparedAPIBenchOwnerSink = owner
			}
			b.StopTimer()
			runtime.KeepAlive(f)
		})
	}
}

func BenchmarkPreparedAPIReferenceFullRank(b *testing.B) {
	for _, n := range []int{1, 5, 8} {
		b.Run(fmt.Sprintf("candidates_%d", n), func(b *testing.B) {
			b.StopTimer()
			f := preparedAPIBenchSetup(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				ranking, err := preparedAPIBenchReference(f.weights, f.input)
				if err != nil {
					b.Fatal(err)
				}
				preparedAPIBenchRankingSink = ranking
			}
			b.StopTimer()
			runtime.KeepAlive(f)
		})
	}
}

func BenchmarkPreparedAPIWarmRank(b *testing.B) {
	for _, n := range []int{1, 5, 8} {
		b.Run(fmt.Sprintf("candidates_%d", n), func(b *testing.B) {
			b.StopTimer()
			f := preparedAPIBenchSetup(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				ranking, err := f.owner.Rank(f.validated, f.view)
				if err != nil {
					b.Fatal(err)
				}
				preparedAPIBenchRankingSink = ranking
			}
			b.StopTimer()
			// Both representations coexist as common untimed fixture state. No
			// retained/peak heap, process RSS or whole-task memory claim is made.
			runtime.KeepAlive(f)
		})
	}
}

func preparedAPIBenchBaseline(b *testing.B, kind string) {
	for _, n := range []int{1, 5, 8} {
		b.Run(fmt.Sprintf("candidates_%d", n), func(b *testing.B) {
			b.StopTimer()
			f := preparedAPIBenchSetup(b, n)
			// This untimed call only checks availability of the existing public
			// method. There is no expected quality/order or allocation threshold.
			if _, err := f.validated.Rank(kind); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				ranking, err := f.validated.Rank(kind)
				if err != nil {
					b.Fatal(err)
				}
				preparedAPIBenchRankingSink = ranking
			}
			b.StopTimer()
			runtime.KeepAlive(f)
		})
	}
}

func BenchmarkPreparedAPIBM25(b *testing.B) {
	preparedAPIBenchBaseline(b, shortclaim.BM25Kind)
}

func BenchmarkPreparedAPILexicalOrdered(b *testing.B) {
	preparedAPIBenchBaseline(b, shortclaim.LexicalOrderedKind)
}
