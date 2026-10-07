// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"bytes"
	"math"
	"math/rand/v2"
	"strings"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"github.com/teamswyg/laya-tools/pkg/tritpack"
)

// checkedPackedLogitsReference freezes the preceding checked decoder exactly.
// It is only a test reference: Model never retains this function, a decoded
// weight array, or views pointing into its own packed storage.
func checkedPackedLogitsReference(m *Model, view statehintwide.ContextualFeatureView) ([HeadCount][StateCount]float64, error) {
	var result [HeadCount][StateCount]float64
	for h := range result {
		for c := range result[h] {
			result[h][c] = float64(m.bias[h][c])
		}
	}
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		for h := range result {
			for c := range result[h] {
				index := int(f.Index)*HeadCount*StateCount + h*StateCount + c
				trit, err := tritpack.At(m.packed[:], index, WeightTritCount)
				if err != nil {
					return result, ErrModel
				}
				result[h][c] += float64(trit) * float64(m.scales[h]) * float64(f.Value)
			}
		}
	}
	return result, nil
}

func checkedPackedScoresReference(m *Model, text string) (ScoreResult, error) {
	var workspace statehintwide.Workspace
	view, err := statehintwide.ExtractContextual(text, &workspace)
	if err != nil {
		return ScoreResult{}, err
	}
	logits, err := checkedPackedLogitsReference(m, view)
	if err != nil {
		return ScoreResult{}, err
	}
	r := ScoreResult{Logits: logits, WordCount: view.WordCount()}
	for h := range r.Probabilities {
		p, ok := softmax(r.Logits[h])
		if !ok {
			return ScoreResult{}, ErrModel
		}
		r.Probabilities[h] = p
	}
	return r, nil
}

func checkedPackedPredictionReference(m *Model, scores ScoreResult) Prediction {
	r := Prediction{Source: Learned, TrainingSteps: m.TrainingSteps(), Mode: m.mode.String(), ParentSHA256: m.parentSHAHex}
	if r.TrainingSteps == 0 {
		r.Source = Untrained
	}
	for h := range r.Heads {
		p := scores.Probabilities[h]
		if scores.WordCount == 0 {
			p = [StateCount]float64{1.0 / StateCount, 1.0 / StateCount, 1.0 / StateCount}
		}
		r.Heads[h] = headPrediction(Head(h), p, r.TrainingSteps, scores.WordCount)
	}
	return r
}

func assertCheckedDecoderParity(t *testing.T, m *Model, texts []string) {
	t.Helper()
	before := *m
	var workspace Workspace
	for _, text := range texts {
		want, err := checkedPackedScoresReference(m, text)
		if err != nil {
			t.Fatal("checked score reference", err)
		}
		got, err := m.Scores(text, &workspace)
		if err != nil || got.WordCount != want.WordCount {
			t.Fatal("lookup score result", err)
		}
		for h := range got.Logits {
			for c := range got.Logits[h] {
				if math.Float64bits(got.Logits[h][c]) != math.Float64bits(want.Logits[h][c]) || math.Float64bits(got.Probabilities[h][c]) != math.Float64bits(want.Probabilities[h][c]) {
					t.Fatalf("decoder changed score bits for %q at head/state %d/%d", text, h, c)
				}
			}
		}
		prediction, err := m.Predict(text, &workspace)
		wantPrediction := checkedPackedPredictionReference(m, want)
		if err != nil || prediction != wantPrediction {
			t.Fatal("decoder changed prediction", err)
		}
		for h, head := range prediction.Heads {
			if math.Float64bits(head.Confidence) != math.Float64bits(wantPrediction.Heads[h].Confidence) || math.Float64bits(head.Margin) != math.Float64bits(wantPrediction.Heads[h].Margin) {
				t.Fatal("decoder changed prediction confidence/margin bits")
			}
			for c, probability := range head.Probabilities {
				if math.Float64bits(probability) != math.Float64bits(wantPrediction.Heads[h].Probabilities[c]) {
					t.Fatal("decoder changed prediction probability bits")
				}
			}
		}
	}
	if *m != before {
		t.Fatal("lookup inference mutated model")
	}
}

func TestLookupTableMatchesEveryValidCodeAndCanonicalTail(t *testing.T) {
	if decoderTableBytes != 1215 || unsafe.Sizeof(packedTritDigits) != uintptr(decoderTableBytes) {
		t.Fatal("shared static decoder byte accounting changed")
	}
	for code := 0; code < 243; code++ {
		packed := []byte{byte(code)}
		for digit := 0; digit < 5; digit++ {
			want, err := tritpack.At(packed, digit, 5)
			if err != nil || packedTritDigits[code][digit] != want {
				t.Fatalf("table code/digit %d/%d differs from checked decoder", code, digit)
			}
		}
	}
	var packed [PackedWeightBytes]byte
	for i := range packed {
		packed[i] = 121 // Five zero trits.
	}
	// The last group contains two live trits and three zero-trit padding digits.
	// These nine codes cover every live-trit pair, with both signs and zero.
	for pair := 0; pair < 9; pair++ {
		packed[PackedWeightBytes-1] = byte(117 + pair)
		if err := tritpack.Validate(packed[:], WeightTritCount); err != nil {
			t.Fatal("invalid owned tail fixture", err)
		}
		for index := WeightTritCount - 2; index < WeightTritCount; index++ {
			want, err := tritpack.At(packed[:], index, WeightTritCount)
			got := packedTritDigits[packed[index/5]][index%5]
			if err != nil || got != want {
				t.Fatalf("partial-tail pair/index %d/%d differs from checked decoder", pair, index)
			}
		}
	}
}

func ownedDecoderTexts() []string {
	return []string{
		"", "...", "a", "amber kite waits", "violet boat rests",
		"cedar lantern drifts beside an ochre pebble", "새 구름 쉼",
		strings.Repeat("a b c d e f g h i j k l m n o p q r s t u v w x y z ", 70),
	}
}

func TestLookupDecoderExactBitsOnOwnedPackedModelsCloneAndLoad(t *testing.T) {
	tableBefore := packedTritDigits
	random := rand.New(rand.NewPCG(1729, 817))
	scaleChoices := [...]float32{float32(ScaleFloor), .03125, .1, 1, 17, math.MaxFloat32 / 1024}
	for fixture := 0; fixture < 24; fixture++ {
		// Directly constructed private numerical fixtures are not real assets or
		// evidence of training history. Their encodings are fully validated.
		m := &Model{mode: PTQ, baseSteps: 2120, parentSeed: DefaultSeed, parentSHAHex: strings.Repeat("0", 64)}
		for i := 0; i < PackedWeightBytes-1; i++ {
			if fixture == 0 {
				m.packed[i] = byte(i % 243) // Every full-group byte occurs.
			} else {
				m.packed[i] = byte(random.IntN(243))
			}
		}
		m.packed[PackedWeightBytes-1] = byte(117 + fixture%9)
		for h := range m.scales {
			m.scales[h] = scaleChoices[(fixture+h)%len(scaleChoices)]
			for c := range m.bias[h] {
				m.bias[h][c] = float32(random.IntN(17)-8) / 8
			}
		}
		if fixture == 0 {
			m.bias[0][0] = float32(math.Copysign(0, -1))
		}
		if !m.valid() {
			t.Fatal("invalid owned decoder model")
		}
		data := modelBytes(t, m)
		loaded, err := Load(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		for _, candidate := range []*Model{m, m.Clone(), loaded} {
			assertCheckedDecoderParity(t, candidate, ownedDecoderTexts())
			if !bytes.Equal(data, modelBytes(t, candidate)) {
				t.Fatal("lookup decoder changed serialized model bytes")
			}
		}
		// Clone owns its packed array rather than retaining a view into m.
		clone := m.Clone()
		clone.packed[0] = (clone.packed[0] + 1) % 243
		if !bytes.Equal(data, modelBytes(t, m)) {
			t.Fatal("clone decoder storage aliases parent model")
		}
	}
	if packedTritDigits != tableBefore {
		t.Fatal("inference changed shared static decoder data")
	}
}

func TestLookupDecoderExactBitsAfterOwnedToyWarmFit(t *testing.T) {
	parent, m := syntheticPTQ(t)
	beforeParent := parentBytes(t, parent)
	tableBefore := packedTritDigits
	var workspace TrainingWorkspace
	if _, err := m.WarmFit(parent, syntheticSamples(), &workspace); err != nil {
		t.Fatal(err)
	}
	data := modelBytes(t, m)
	loaded, err := Load(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []*Model{m, m.Clone(), loaded, &workspace.forward} {
		assertCheckedDecoderParity(t, candidate, ownedDecoderTexts())
		if !bytes.Equal(data, modelBytes(t, candidate)) {
			t.Fatal("toy continuation or reload changed serialized bytes")
		}
	}
	if !bytes.Equal(beforeParent, parentBytes(t, parent)) || packedTritDigits != tableBefore {
		t.Fatal("toy continuation mutated parent or shared decoder table")
	}
}
