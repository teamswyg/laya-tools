// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"github.com/teamswyg/laya-tools/pkg/tritpack"
)

// All text and supervision in this package's tests are tiny original synthetic
// fixtures. No real corpus, private input, evaluation set or model asset is read.
func syntheticSamples() []Sample {
	return []Sample{
		{Text: "amber kite waits", Targets: [HeadCount]State{True, False, Unknown}},
		{Text: "violet boat rests", Targets: [HeadCount]State{False, Unknown, True}},
	}
}

func parentBytes(t *testing.T, parent *statehintclaims.Model) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := parent.Save(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func modelBytes(t *testing.T, m *Model) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := m.Save(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func syntheticParent(t *testing.T) (*statehintclaims.Model, string) {
	t.Helper()
	p := statehintclaims.NewModel()
	// A small owned fixture only. Parent training API is never used by WarmFit.
	if _, err := p.Fit(syntheticSamples(), statehintclaims.FitOptions{Epochs: 2, BatchSize: 1, LearningRate: .02}, &statehintclaims.TrainingWorkspace{}); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(parentBytes(t, p))
	return p, hex.EncodeToString(sum[:])
}

func syntheticPTQ(t *testing.T) (*statehintclaims.Model, *Model) {
	t.Helper()
	parent, digest := syntheticParent(t)
	m, err := FromFloat(parent, digest)
	if err != nil {
		t.Fatal(err)
	}
	return parent, m
}

func TestRoundToEvenClippedTrits(t *testing.T) {
	for _, tt := range []struct {
		w    float32
		want int8
	}{
		{-100, -1}, {-1.5, -1}, {-1, -1}, {-.5, 0}, {0, 0}, {.5, 0}, {1, 1}, {1.5, 1}, {100, 1},
	} {
		if got := quantizedTrit(tt.w, 1); got != tt.want {
			t.Errorf("Q(%g): got %d want %d", tt.w, got, tt.want)
		}
	}
	if quantizedTrit(math.Nextafter32(.5, 1), 1) != 1 || quantizedTrit(math.Nextafter32(-.5, -1), 1) != -1 {
		t.Fatal("rounding did not distinguish adjacent representable values")
	}
}

func TestPerHeadAbsMeanScaleFloorAndOrder(t *testing.T) {
	var p statehintclaims.Parameters
	for i := range p.Weights {
		for c := range p.Weights[i][0] {
			p.Weights[i][0][c] = 1
			p.Weights[i][1][c] = -2
		}
	}
	m := &Model{mode: PTQ}
	if err := m.quantize(&p); err != nil {
		t.Fatal(err)
	}
	if m.scales != [HeadCount]float32{1, 2, float32(ScaleFloor)} {
		t.Fatalf("scales %v", m.scales)
	}
	for _, feature := range []int{0, 1, FeatureBins - 1} {
		for h, want := range [HeadCount]int8{1, -1, 0} {
			for c := 0; c < StateCount; c++ {
				got, err := tritpack.At(m.packed[:], feature*HeadCount*StateCount+h*StateCount+c, WeightTritCount)
				if err != nil || got != want {
					t.Fatalf("feature/head/state %d/%d/%d: %d %v", feature, h, c, got, err)
				}
			}
		}
	}
	// Sparse values affect a mean across every one of the head's 6144 weights.
	p = statehintclaims.Parameters{}
	p.Weights[0][0][0] = 1
	if err := m.quantize(&p); err != nil {
		t.Fatal(err)
	}
	if m.scales[0] != float32(1.0/(FeatureBins*StateCount)) || m.scales[1] != float32(ScaleFloor) {
		t.Fatalf("sparse absmean %v", m.scales)
	}
	p.Weights[0][0][0] = float32(math.Inf(1))
	if err := m.quantize(&p); err == nil {
		t.Fatal("accepted nonfinite shadow weight")
	}
}

func TestFromFloatVerifiesActualParentAndOwnsCopy(t *testing.T) {
	parent, digest := syntheticParent(t)
	before := parentBytes(t, parent)
	for _, bad := range []string{"", "xyz", strings.Repeat("0", 64), digest[:62]} {
		if _, err := FromFloat(parent, bad); err == nil {
			t.Fatalf("accepted digest %q", bad)
		}
	}
	if _, err := FromFloat(nil, digest); err == nil {
		t.Fatal("accepted nil parent")
	}
	m, err := FromFloat(parent, strings.ToUpper(digest))
	if err != nil {
		t.Fatal(err)
	}
	parameters, err := parent.Parameters()
	if err != nil {
		t.Fatal(err)
	}
	reference := &Model{}
	if err := reference.quantize(&parameters); err != nil {
		t.Fatal(err)
	}
	if m.packed != reference.packed || m.scales != reference.scales || m.bias != parameters.Bias {
		t.Fatal("PTQ does not project the parent's parameters")
	}
	parameters.Weights[0][0][0] = 999
	if !bytes.Equal(before, parentBytes(t, parent)) {
		t.Fatal("conversion or copy altered parent")
	}
	meta := m.Metadata()
	if meta.Mode != "ptq" || meta.BaseTrainingSteps != parent.TrainingSteps() || meta.NewOptimizerSteps != 0 || meta.TrainingSteps != parent.TrainingSteps() || meta.ParentSHA256 != digest || meta.ParentInitializationSeed != parent.InitializationSeed() {
		t.Fatalf("provenance %+v", meta)
	}
	if meta.BitNetLLM || meta.W158A8 || meta.QualityQualified || meta.StateAuthority || meta.PhysicalBitsPerTrit != 1.6 || meta.InformationBitsPerTrit != math.Log2(3) {
		t.Fatalf("overstated metadata %+v", meta)
	}
	if unsafe.Sizeof(*m) > 4096 {
		t.Fatalf("resident model too large: %d", unsafe.Sizeof(*m))
	}
}

func TestPredictGuardsAndInheritedLearnedSource(t *testing.T) {
	var p statehintclaims.Parameters
	p.Bias = [HeadCount][StateCount]float32{{12, 0, 0}, {0, 12, 0}, {0, 0, 12}}
	m := &Model{mode: PTQ, baseSteps: 2120, parentSHAHex: strings.Repeat("0", sha256.Size*2)}
	if err := m.quantize(&p); err != nil {
		t.Fatal(err)
	}
	r, err := m.Predict("synthetic words", &Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != Learned || r.TrainingSteps != 2120 || r.Mode != "ptq" || r.Heads[0].State != True || r.Heads[1].State != False || r.Heads[2].State != Unknown || r.Heads[2].UnknownReason != "semantic_unknown" {
		t.Fatalf("prediction %+v", r)
	}
	empty, err := m.Predict("...", &Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	for _, head := range empty.Heads {
		if head.State != Unknown || head.UnknownReason != "no_word_content" || head.Probabilities != [StateCount]float64{1.0 / 3, 1.0 / 3, 1.0 / 3} {
			t.Fatalf("word guard %+v", head)
		}
	}
	m.baseSteps = 0
	r, err = m.Predict("synthetic words", &Workspace{})
	if err != nil || r.Source != Untrained || r.Heads[0].UnknownReason != "untrained" {
		t.Fatalf("untrained %+v %v", r, err)
	}
	if got := headPrediction(ResponseRequested, [StateCount]float64{.6, .3, .1}, 1, 1); got.State != Unknown || got.UnknownReason != "low_confidence" {
		t.Fatalf("confidence guard %+v", got)
	}
	if got := headPrediction(ResponseRequested, [StateCount]float64{1.0 / 3, 1.0 / 3, 1.0 / 3}, 1, 1); got.Winner != Unknown {
		t.Fatalf("tie %+v", got)
	}
}

func TestScoresOnlyDecodesActiveBinsAndValidatesInputs(t *testing.T) {
	_, m := syntheticPTQ(t)
	w := &Workspace{}
	want, err := m.Scores("a", w)
	if err != nil {
		t.Fatal(err)
	}
	var features statehintwide.Workspace
	view, err := statehintwide.ExtractContextual("a", &features)
	if err != nil {
		t.Fatal(err)
	}
	var addressed [PackedWeightBytes]bool
	for i := 0; i < view.Len(); i++ {
		for j := 0; j < HeadCount*StateCount; j++ {
			addressed[(int(view.At(i).Index)*HeadCount*StateCount+j)/5] = true
		}
	}
	// Internal fault injection detects an accidental whole-weight validation or
	// unpack during scoring. Public constructors never create this invalid state.
	for i, used := range addressed {
		if !used {
			m.packed[i] = 255
			break
		}
	}
	got, err := m.Scores("a", w)
	if err != nil || got != want {
		t.Fatalf("score read inactive bins: %+v %v", got, err)
	}
	if m.valid() {
		t.Fatal("whole-artifact validation missed the injected corruption")
	}
	for _, s := range []string{string([]byte{0xff}), "a\x00b", strings.Repeat("a", MaxTextBytes+1)} {
		if _, err := m.Scores(s, w); err == nil {
			t.Fatal("accepted invalid text")
		}
	}
	if _, err := m.Scores("a", nil); err == nil {
		t.Fatal("accepted nil workspace")
	}
	var nilModel *Model
	if _, err := nilModel.Scores("a", w); err == nil {
		t.Fatal("accepted nil model")
	}
}

func TestConcurrentInferenceOwnsSeparateWorkspaces(t *testing.T) {
	_, m := syntheticPTQ(t)
	before := *m
	want, err := m.Scores("amber kite waits", &Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	errors := make(chan error, 8)
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			workspace := &Workspace{}
			for j := 0; j < 16; j++ {
				got, err := m.Scores("amber kite waits", workspace)
				if err != nil {
					errors <- err
					return
				}
				if got != want {
					errors <- ErrModel
					return
				}
			}
		}()
	}
	group.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	if *m != before {
		t.Fatal("inference mutated model")
	}
}

func TestWarmPredictHasNoAllocationsAndPreservesParentSHA(t *testing.T) {
	parent, m := syntheticPTQ(t)
	sum := sha256.Sum256(parentBytes(t, parent))
	wantSHA := hex.EncodeToString(sum[:])
	loaded, err := Load(bytes.NewReader(modelBytes(t, m)))
	if err != nil {
		t.Fatal(err)
	}
	for name, model := range map[string]*Model{"constructed": m, "loaded": loaded, "clone": m.Clone()} {
		t.Run(name, func(t *testing.T) {
			workspace := &Workspace{}
			var last Prediction
			allocations := testing.AllocsPerRun(1000, func() {
				var err error
				last, err = model.Predict("amber kite waits", workspace)
				if err != nil {
					t.Fatal(err)
				}
			})
			if allocations != 0 {
				t.Fatalf("warm Predict allocated %g times with caller workspace", allocations)
			}
			if last.ParentSHA256 != wantSHA || model.Metadata().ParentSHA256 != wantSHA || !model.parentSHAHexMatches() {
				t.Fatal("prediction or metadata changed the parent digest")
			}
			if unsafe.StringData(last.ParentSHA256) != unsafe.StringData(model.parentSHAHex) || unsafe.StringData(model.Metadata().ParentSHA256) != unsafe.StringData(model.parentSHAHex) {
				t.Fatal("prediction or metadata recreated immutable digest text")
			}
		})
	}
	clone := m.Clone()
	if unsafe.StringData(clone.parentSHAHex) != unsafe.StringData(m.parentSHAHex) || len(m.parentSHAHex) != 64 {
		t.Fatal("clone did not share the immutable 64-byte digest backing")
	}
	t.Logf("resident Model struct: %d bytes; additional immutable parent digest backing: %d bytes (shared by clones)", unsafe.Sizeof(*m), len(m.parentSHAHex))
}
