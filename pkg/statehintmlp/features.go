// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintmlp provides an isolated fixed-width representation experiment.
package statehintmlp

import (
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const (
	IntentCount    = 8
	FeatureBins    = 2048
	HiddenUnits    = 16
	MaxTextBytes   = 4096
	ParameterCount = FeatureBins*HiddenUnits + HiddenUnits + HiddenUnits*IntentCount + IntentCount
	FeatureSchema  = "utf8-word12-char2345-fnv1a-signed-logtf-l2-2048-v2"
)

type Intent = statehint.Intent
type Source = statehint.Source
type Prediction = statehint.Prediction

const (
	Learned          = statehint.Learned
	Untrained        = statehint.Untrained
	Question         = statehint.Question
	Blocker          = statehint.Blocker
	Reference        = statehint.Reference
	Progress         = statehint.Progress
	CompletionReport = statehint.CompletionReport
	CancelRequest    = statehint.CancelRequest
	Planned          = statehint.Planned
	Unclear          = statehint.Unclear
)

var (
	ErrInput    = statehint.ErrInput
	ErrModel    = statehint.ErrModel
	ErrTraining = statehint.ErrTraining
	ErrArtifact = statehint.ErrArtifact
)

func Intents() [IntentCount]Intent     { return statehint.Intents() }
func IntentIndex(i Intent) (int, bool) { return statehint.IntentIndex(i) }
func finite(v float64) bool            { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Workspace is caller-owned. Reuse it serially; simultaneous predictions need
// separate workspaces. The model has no mutable prediction cache or locks.
type Workspace struct {
	values                [FeatureBins]float32
	indices               [FeatureBins]uint16
	seen                  [FeatureBins]bool
	count                 int
	runes                 [MaxTextBytes]rune
	runeCount, wordCount  int
	preactivation, hidden [HiddenUnits]float64
}

// Bounded private copy of pkg/statehintwide's Contextual extraction, under the
// same Apache-2.0 license. Keeping the copy isolates v3 from existing v1/v2 APIs.
// Gold tests recover every contextual feature through owned v2 linear probes.
func (w *Workspace) clear() {
	for _, i := range w.indices[:w.count] {
		w.values[i] = 0
		w.seen[i] = false
	}
	w.count, w.runeCount, w.wordCount = 0, 0, 0
	w.preactivation = [HiddenUnits]float64{}
	w.hidden = [HiddenUnits]float64{}
}
func hashRunes(h uint32, rs []rune) uint32 {
	var b [utf8.UTFMax]byte
	for _, r := range rs {
		n := utf8.EncodeRune(b[:], r)
		for _, c := range b[:n] {
			h = (h ^ uint32(c)) * 16777619
		}
	}
	return h
}
func (w *Workspace) emit(prefix byte, a, b []rune) {
	h := (uint32(2166136261) ^ uint32(prefix)) * 16777619
	h = hashRunes(h, a)
	if b != nil {
		h = h * 16777619
		h = hashRunes(h, b)
	}
	i := h % FeatureBins
	if !w.seen[i] {
		w.seen[i] = true
		w.indices[w.count] = uint16(i)
		w.count++
	}
	v := float32(1)
	if h>>31 != 0 {
		v = -1
	}
	w.values[i] += v
}
func extract(text string, w *Workspace) error {
	if w == nil || len(text) > MaxTextBytes || !utf8.ValidString(text) {
		return ErrInput
	}
	w.clear()
	for _, r := range text {
		if r == 0 {
			return ErrInput
		}
		w.runes[w.runeCount] = unicode.ToLower(r)
		w.runeCount++
	}
	rs := w.runes[:w.runeCount]
	start, previousStart, previousEnd := -1, -1, -1
	word := func(end int) {
		w.emit('w', rs[start:end], nil)
		if previousStart >= 0 {
			w.emit('b', rs[previousStart:previousEnd], rs[start:end])
		}
		previousStart, previousEnd = start, end
		w.wordCount++
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
			w.emit(byte('0'+n), rs[i:i+n], nil)
		}
	}
	var sum float64
	for _, i := range w.indices[:w.count] {
		v := float64(w.values[i])
		v = math.Copysign(math.Log1p(math.Abs(v)), v)
		w.values[i] = float32(v)
		sum += v * v
	}
	if sum > 0 {
		scale := float32(1 / math.Sqrt(sum))
		for _, i := range w.indices[:w.count] {
			w.values[i] *= scale
		}
	}
	return nil
}
