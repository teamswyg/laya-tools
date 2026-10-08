// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintpositionpreview is an optional EXPERIMENTAL position-feature
// extractor. It provides no model, prediction, training, or production contract.
package statehintpositionpreview

import (
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	BaseFeatureBins     = statehintwide.FeatureBins
	PositionFeatureBins = 1024
	FeatureBins         = BaseFeatureBins + PositionFeatureBins
	PositionRegions     = 4
	MaxTextBytes        = statehintwide.MaxTextBytes
	PositionBankScale   = 1

	// FeatureSchema identifies this preview, never an existing model schema.
	FeatureSchema = "utf8-word12-char2345-v2-pos4-fnv1a-signed-logtf-l2-3072-preview-v1"
)

// Workspace owns all extraction memory. Use its zero value, do not copy it
// after first use, and give each concurrent caller a separate workspace.
type Workspace struct {
	base      statehintwide.Workspace
	values    [PositionFeatureBins]float32
	indices   [PositionFeatureBins]uint16
	seen      [PositionFeatureBins]bool
	count     int
	runes     [MaxTextBytes]rune
	runeCount int
}

// SparseFeature is an activated bin, including zero-valued signed collisions.
type SparseFeature struct {
	Index uint16
	Value float32
}

// FeatureView borrows a Workspace. It is valid only until the next extraction
// using that workspace, including a failed call. Copy values to retain them.
// Reading a view while its workspace is reused is not safe.
type FeatureView struct {
	base      statehintwide.ContextualFeatureView
	workspace *Workspace
}

func (v FeatureView) Len() int {
	if v.workspace == nil {
		return 0
	}
	return v.base.Len() + v.workspace.count
}

// At requires 0 <= index < Len(), as with indexing a slice. Unchanged v2 bins
// appear first in their original encounter order, followed by position bins.
func (v FeatureView) At(index int) SparseFeature {
	if index < v.base.Len() {
		f := v.base.At(index)
		return SparseFeature{Index: f.Index, Value: f.Value}
	}
	i := v.workspace.indices[:v.workspace.count][index-v.base.Len()]
	return SparseFeature{Index: uint16(BaseFeatureBins) + i, Value: v.workspace.values[i]}
}

// WordCount returns the actual v2 extractor's count, without reinterpretation.
func (v FeatureView) WordCount() int { return v.base.WordCount() }

func (w *Workspace) clearPosition() {
	for _, i := range w.indices[:w.count] {
		w.values[i] = 0
		w.seen[i] = false
	}
	w.count, w.runeCount = 0, 0
}

func hashByte(h uint32, b byte) uint32 { return (h ^ uint32(b)) * 16777619 }

func hashRunes(h uint32, rs []rune) uint32 {
	var encoded [utf8.UTFMax]byte
	for _, r := range rs {
		n := utf8.EncodeRune(encoded[:], r)
		for _, b := range encoded[:n] {
			h = hashByte(h, b)
		}
	}
	return h
}

func (w *Workspace) emit(start int, prefix byte, a, b []rune) {
	region := byte(PositionRegions * start / w.runeCount)
	h := hashByte(2166136261, 'p')
	h = hashByte(h, region)
	h = hashByte(h, prefix)
	h = hashRunes(h, a)
	if b != nil {
		h = hashByte(h, 0)
		h = hashRunes(h, b)
	}
	i := h % PositionFeatureBins
	if !w.seen[i] {
		w.seen[i] = true
		w.indices[w.count] = uint16(i)
		w.count++
	}
	value := float32(1)
	if h>>31 != 0 {
		value = -1
	}
	w.values[i] += value
}

func (w *Workspace) normalizePosition() {
	var sum float64
	for _, i := range w.indices[:w.count] {
		v := float64(w.values[i])
		v = math.Copysign(math.Log1p(math.Abs(v)), v)
		w.values[i] = float32(v)
		sum += v * v
	}
	if sum > 0 {
		scale := float32(PositionBankScale / math.Sqrt(sum))
		for _, i := range w.indices[:w.count] {
			w.values[i] *= scale
		}
	}
}

// Extract preserves the actual v2 2048-bin values bit for bit and appends one
// independently normalized 1024-bin signed position bank, at fixed scale 1.
// It accepts complete valid UTF-8 up to MaxTextBytes, rejects NUL, and never
// truncates. Input errors are exactly statehintwide.ErrInput. No memory is
// allocated. The result aliases the exclusively caller-owned workspace.
func Extract(text string, workspace *Workspace) (FeatureView, error) {
	if workspace == nil {
		return FeatureView{}, statehintwide.ErrInput
	}
	base, err := statehintwide.ExtractContextual(text, &workspace.base)
	if err != nil {
		return FeatureView{}, err
	}
	w := workspace
	w.clearPosition()
	for _, r := range text {
		w.runes[w.runeCount] = unicode.ToLower(r)
		w.runeCount++
	}
	rs := w.runes[:w.runeCount]
	start, previousStart, previousEnd := -1, -1, -1
	word := func(end int) {
		w.emit(start, 'w', rs[start:end], nil)
		if previousStart >= 0 {
			w.emit(previousStart, 'b', rs[previousStart:previousEnd], rs[start:end])
		}
		previousStart, previousEnd = start, end
		start = -1
	}
	for i, r := range rs {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if start < 0 {
				start = i
			}
		} else if start >= 0 {
			word(i)
		}
	}
	if start >= 0 {
		word(len(rs))
	}
	for n := 2; n <= 5; n++ {
		for i := 0; i+n <= len(rs); i++ {
			w.emit(i, byte('0'+n), rs[i:i+n], nil)
		}
	}
	w.normalizePosition()
	return FeatureView{base: base, workspace: w}, nil
}
