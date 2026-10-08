// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintpositionpreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

type originalCase struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	SHA256 string `json:"exact_utf8_sha256"`
}

type originalPair struct {
	ID string       `json:"id"`
	A  originalCase `json:"a"`
	B  originalCase `json:"b"`
}

func originalPairs(tb testing.TB) []originalPair {
	tb.Helper()
	raw, err := os.ReadFile("../../experiments/claim-position-features-preview-v1/cases.json")
	if err != nil {
		tb.Fatal(err)
	}
	var packet struct {
		Pairs []originalPair `json:"pairs"`
	}
	if err := json.Unmarshal(raw, &packet); err != nil || len(packet.Pairs) != 6 {
		tb.Fatal("expected all six original pairs", err)
	}
	for _, pair := range packet.Pairs {
		for _, c := range []originalCase{pair.A, pair.B} {
			digest := sha256.Sum256([]byte(c.Text))
			if hex.EncodeToString(digest[:]) != c.SHA256 {
				tb.Fatalf("exact original text digest changed: %s", c.ID)
			}
		}
	}
	return packet.Pairs
}

func denseBits(view FeatureView) [FeatureBins]uint32 {
	var bits [FeatureBins]uint32
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		bits[f.Index] = math.Float32bits(f.Value)
	}
	return bits
}

func baseBits(view statehintwide.ContextualFeatureView) [BaseFeatureBins]uint32 {
	var bits [BaseFeatureBins]uint32
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		bits[f.Index] = math.Float32bits(f.Value)
	}
	return bits
}

func assertBaseParity(t *testing.T, text string) {
	t.Helper()
	var baseWorkspace statehintwide.Workspace
	var previewWorkspace Workspace
	base, baseErr := statehintwide.ExtractContextual(text, &baseWorkspace)
	preview, previewErr := Extract(text, &previewWorkspace)
	if baseErr != previewErr {
		t.Fatalf("input guard mismatch: base %v preview %v", baseErr, previewErr)
	}
	if baseErr != nil {
		if preview.Len() != 0 || preview.WordCount() != 0 {
			t.Fatal("error returned a nonempty view")
		}
		return
	}
	if preview.WordCount() != base.WordCount() || preview.Len() < base.Len() {
		t.Fatal("base word count or activated bins changed")
	}
	for i := 0; i < base.Len(); i++ {
		actual, expected := preview.At(i), base.At(i)
		if actual.Index != expected.Index || math.Float32bits(actual.Value) != math.Float32bits(expected.Value) {
			t.Fatalf("base sparse value/order drift at %d", i)
		}
	}
	actualBits, expectedBits := denseBits(preview), baseBits(base)
	for i, expected := range expectedBits {
		if actualBits[i] != expected {
			t.Fatalf("first 2048 dense bits drift at %d", i)
		}
	}
}

func TestActualV2BaseBitParityAndFullInputBounds(t *testing.T) {
	inputs := []string{
		"", "a", "...", "Aé B2", "İ ABC １２３\n가나다🙂!", "Repeated repeated repeated.",
		strings.Repeat("a", MaxTextBytes),
		strings.Repeat("한", 1365) + "a", // 4096 bytes, 1366 runes.
		strings.Repeat("🙂", 1024),       // 4096 bytes, 1024 runes.
		strings.Repeat("x", MaxTextBytes+1), "bad\xff", "bad\x00input",
	}
	for _, text := range inputs {
		assertBaseParity(t, text)
	}
	for _, pair := range originalPairs(t) {
		assertBaseParity(t, pair.A.Text)
		assertBaseParity(t, pair.B.Text)
	}
	if _, err := Extract("input", nil); err != statehintwide.ErrInput {
		t.Fatal("nil workspace was accepted")
	}
	var w Workspace
	view, err := Extract(strings.Repeat("a", MaxTextBytes), &w)
	if err != nil || w.runeCount != MaxTextBytes || view.WordCount() != 1 {
		t.Fatal("maximum ASCII input was truncated", err)
	}
	view, err = Extract(strings.Repeat("한", 1365)+"a", &w)
	if err != nil || w.runeCount != 1366 || view.WordCount() != 1 {
		t.Fatal("maximum multibyte input was truncated", err)
	}
}

func TestPositionHashRuneBoundariesAndIndependentNormalization(t *testing.T) {
	// Hand-enumerated events for lowercased "Aé B2": R=5, each record is
	// ASCII p + raw region byte + original prefix + UTF-8 payload. The
	// bigram starts at rune 0; character windows use their starting rune.
	records := []string{
		"p\x00waé", "p\x02wb2", "p\x00baé\x00b2",
		"p\x002aé", "p\x002é ", "p\x012 b", "p\x022b2",
		"p\x003aé ", "p\x003é b", "p\x013 b2",
		"p\x004aé b", "p\x004é b2", "p\x005aé b2",
	}
	var expected [PositionFeatureBins]float32
	var order []uint16
	var seen [PositionFeatureBins]bool
	for _, record := range records {
		h := fnv.New32a()
		_, _ = h.Write([]byte(record))
		sum := h.Sum32()
		i := sum % PositionFeatureBins
		if !seen[i] {
			order = append(order, uint16(i))
			seen[i] = true
		}
		if sum>>31 != 0 {
			expected[i]--
		} else {
			expected[i]++
		}
	}
	var norm float64
	for _, i := range order {
		v := math.Copysign(math.Log1p(math.Abs(float64(expected[i]))), float64(expected[i]))
		expected[i] = float32(v)
		norm += v * v
	}
	scale := float32(1 / math.Sqrt(norm))
	for _, i := range order {
		expected[i] *= scale
	}
	var w Workspace
	view, err := Extract("Aé B2", &w)
	if err != nil || w.count != len(order) {
		t.Fatal("hand-enumerated event count mismatch", err)
	}
	for n, i := range order {
		if w.indices[n] != i || math.Float32bits(w.values[i]) != math.Float32bits(expected[i]) {
			t.Fatalf("hand-enumerated position event mismatch at %d", n)
		}
	}
	var baseNorm, positionNorm float64
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		if f.Index < BaseFeatureBins {
			baseNorm += float64(f.Value) * float64(f.Value)
		} else {
			positionNorm += float64(f.Value) * float64(f.Value)
		}
	}
	if math.Abs(baseNorm-1) > 1e-6 || math.Abs(positionNorm-1) > 1e-6 {
		t.Fatalf("banks must normalize separately at scale 1: %g %g", baseNorm, positionNorm)
	}
}

func TestSignedCollisionZeroBinsAndLogTF(t *testing.T) {
	// Independently computed FNV values for region-0 word events:
	// "ah": 0x9667c2bf (-1); "aaa": 0x6f68babf (+1), both bin 703.
	var w Workspace
	w.runeCount = 1
	w.emit(0, 'w', []rune{'a', 'h'}, nil)
	w.emit(0, 'w', []rune{'a', 'a', 'a'}, nil)
	if w.count != 1 || w.indices[0] != 703 || w.values[703] != 0 {
		t.Fatal("signed cancellation must retain one activated zero bin")
	}
	w.normalizePosition()
	if w.count != 1 || math.Float32bits(w.values[703]) != 0 {
		t.Fatal("all-zero collision normalization changed activation")
	}
	w.clearPosition()
	w.count, w.indices[0], w.indices[1] = 2, 1, 2
	w.seen[1], w.seen[2] = true, true
	w.values[1], w.values[2] = -3, 1
	w.normalizePosition()
	negative, positive := -math.Log1p(3), math.Log1p(1)
	scale := float32(1 / math.Sqrt(negative*negative+positive*positive))
	if math.Float32bits(w.values[1]) != math.Float32bits(float32(negative)*scale) || math.Float32bits(w.values[2]) != math.Float32bits(float32(positive)*scale) {
		t.Fatal("signed log-TF/float32 independent normalization drift")
	}
}

func TestDeterministicWorkspaceReuseErrorRecoveryAndBorrowedView(t *testing.T) {
	var reused, fresh Workspace
	first, err := Extract("Aé B2", &reused)
	if err != nil || first.workspace != &reused {
		t.Fatal("view must borrow the caller workspace", err)
	}
	expected := denseBits(first)
	if _, err := Extract(strings.Repeat("long clearing pattern! ", 170), &reused); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"", "z", "bad\xff", "bad\x00input", strings.Repeat("x", MaxTextBytes+1), "Aé B2"} {
		actual, actualErr := Extract(text, &reused)
		independent, independentErr := Extract(text, &fresh)
		if actualErr != independentErr || actual.WordCount() != independent.WordCount() || actual.Len() != independent.Len() || denseBits(actual) != denseBits(independent) {
			t.Fatal("reuse leaked old values or changed error behaviour")
		}
		if actualErr != nil && (actual.Len() != 0 || actual.WordCount() != 0) {
			t.Fatal("failed extraction must return a zero view")
		}
	}
	last, err := Extract("Aé B2", &reused)
	if err != nil || denseBits(last) != expected {
		t.Fatal("deterministic output changed after reuse or rejected inputs")
	}
	if allocs := testing.AllocsPerRun(30, func() {
		if _, err := Extract("Aé B2", &reused); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatal("warmed extraction allocated", allocs)
	}
}

func TestSeparateWorkspaceConcurrency(t *testing.T) {
	texts := []string{"Aé B2", "가나다🙂 DEF 123", "... punctuation !", ""}
	var expected [4][FeatureBins]uint32
	for i, text := range texts {
		var w Workspace
		view, err := Extract(text, &w)
		if err != nil {
			t.Fatal(err)
		}
		expected[i] = denseBits(view)
	}
	var wg sync.WaitGroup
	for i, text := range texts {
		wg.Go(func() {
			var w Workspace
			for n := 0; n < 20; n++ {
				view, err := Extract(text, &w)
				if err != nil || denseBits(view) != expected[i] {
					t.Error("separate caller workspaces interfered")
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestAllSixOriginalPairsRetained(t *testing.T) {
	for _, pair := range originalPairs(t) {
		t.Run(pair.ID, func(t *testing.T) {
			var wa, wb Workspace
			a, errA := Extract(pair.A.Text, &wa)
			b, errB := Extract(pair.B.Text, &wb)
			if errA != nil || errB != nil {
				t.Fatal(errA, errB)
			}
			bitsA, bitsB := denseBits(a), denseBits(b)
			changed := 0
			for i := range bitsA {
				if bitsA[i] != bitsB[i] {
					if i < BaseFeatureBins {
						t.Fatal("original equal base pair changed")
					}
					changed++
				}
			}
			t.Logf("retained feature-only pair: position changed bins=%d", changed)
			// The fixed preview is allowed to fail to separate an engineered
			// pair. Retain and report every outcome; this is not accuracy.
		})
	}
}

func TestWorkspaceSizesAndDistinctContract(t *testing.T) {
	if FeatureBins != 3072 || PositionFeatureBins != 1024 || PositionRegions != 4 || PositionBankScale != 1 || FeatureSchema == statehintwide.ContextualFeatureSchema {
		t.Fatal("fixed preview contract changed")
	}
	t.Logf("Workspace sizeof: v2=%d preview=%d bytes", unsafe.Sizeof(statehintwide.Workspace{}), unsafe.Sizeof(Workspace{}))
}
