// Package statehint classifies bounded original text into generic intent hints.
// It plans shadow changes only; it has no runtime adapter or mutation API.
package statehint

import (
	"errors"
	"math"
	"unicode"
	"unicode/utf8"
)

const (
	IntentCount  = 8
	FeatureBins  = 1024
	MaxTextBytes = 4096
	// FeatureSchema fixes lowercasing, word and character n-grams, FNV-1a
	// namespaces, unsigned bucket selection, and L2 count normalization.
	FeatureSchema = "utf8-lower-word-char23-fnv1a-l2-1024-v1"
)

type Intent string

const (
	Question         Intent = "question"
	Blocker          Intent = "blocker"
	Reference        Intent = "reference"
	Progress         Intent = "progress"
	CompletionReport Intent = "completion_report"
	CancelRequest    Intent = "cancel_request"
	Planned          Intent = "planned"
	Unclear          Intent = "unclear"
)

// Intents returns a value copy in the persisted classifier column order.
func Intents() [IntentCount]Intent {
	return [IntentCount]Intent{Question, Blocker, Reference, Progress, CompletionReport, CancelRequest, Planned, Unclear}
}

func IntentIndex(intent Intent) (int, bool) {
	for index, candidate := range Intents() {
		if candidate == intent {
			return index, true
		}
	}
	return 0, false
}

var (
	ErrInput    = errors.New("statehint: invalid or oversized input")
	ErrModel    = errors.New("statehint: invalid model")
	ErrTraining = errors.New("statehint: invalid training configuration or sample")
	ErrArtifact = errors.New("statehint: invalid or corrupted model artifact")
	ErrPlan     = errors.New("statehint: invalid shadow plan input")
)

// Workspace belongs to the caller. Reuse it serially; concurrent predictions
// must use separate workspaces. The classifier itself is read-only during use.
type Workspace struct {
	values    [FeatureBins]float32
	indices   [FeatureBins]uint16
	count     int
	runes     [MaxTextBytes]rune
	runeCount int
	wordCount int
}

func (w *Workspace) clear() {
	for _, index := range w.indices[:w.count] {
		w.values[index] = 0
	}
	w.count, w.runeCount, w.wordCount = 0, 0, 0
}

func (w *Workspace) emit(prefix byte, runes []rune) {
	hash := uint32(2166136261)
	hash = (hash ^ uint32(prefix)) * 16777619
	var buf [utf8.UTFMax]byte
	for _, r := range runes {
		n := utf8.EncodeRune(buf[:], r)
		for _, b := range buf[:n] {
			hash = (hash ^ uint32(b)) * 16777619
		}
	}
	index := hash % FeatureBins
	if w.values[index] == 0 {
		w.indices[w.count] = uint16(index)
		w.count++
	}
	w.values[index]++
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
	runes := w.runes[:w.runeCount]
	start := -1
	for index, r := range runes {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if start < 0 {
				start = index
			}
		} else if start >= 0 {
			w.emit('w', runes[start:index])
			w.wordCount++
			start = -1
		}
	}
	if start >= 0 {
		w.emit('w', runes[start:])
		w.wordCount++
	}
	for _, length := range [2]int{2, 3} {
		for index := 0; index+length <= len(runes); index++ {
			w.emit(byte('0'+length), runes[index:index+length])
		}
	}
	var sum float64
	for _, index := range w.indices[:w.count] {
		value := float64(w.values[index])
		sum += value * value
	}
	if sum > 0 {
		scale := float32(1 / math.Sqrt(sum))
		for _, index := range w.indices[:w.count] {
			w.values[index] *= scale
		}
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
