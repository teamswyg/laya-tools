// Maintainer-only, feature extraction only. All 24 input strings are newly
// authored, unlabelled development probes, excluded from final 2400 and selectors.
// No corpus, model artifact, classifier call, training, or selection is used.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

type pair struct {
	ID, Focus, A, B string
}

type side struct {
	Text        string  `json:"text"`
	InputDigest string  `json:"input_utf8_sha256"`
	ValidUTF8   bool    `json:"valid_utf8"`
	Bytes       int     `json:"bytes"`
	Runes       int     `json:"runes"`
	Words       int     `json:"words"`
	Active      int     `json:"active_bins"`
	Nonzero     int     `json:"nonzero_bins"`
	Zero        int     `json:"active_zero_bins"`
	Norm        float64 `json:"l2_norm"`
	Digest      string  `json:"dense_float32_sha256"`
	Error       string  `json:"error,omitempty"`
	dense       [statehintwide.FeatureBins]float32
	seen        [statehintwide.FeatureBins]bool
	order       []uint16
}

type result struct {
	ID                   string  `json:"id"`
	Focus                string  `json:"focus"`
	A                    side    `json:"a"`
	B                    side    `json:"b"`
	Comparable           bool    `json:"comparable"`
	DenseBitIdentical    bool    `json:"dense_bit_identical"`
	WordCountIdentical   bool    `json:"word_count_identical"`
	SparseOrderIdentical bool    `json:"sparse_order_identical"`
	ChangedBins          int     `json:"changed_bins"`
	AddedActive          int     `json:"added_active_bins"`
	RemovedActive        int     `json:"removed_active_bins"`
	MaxAbsoluteDelta     float64 `json:"max_absolute_delta"`
	L2Delta              float64 `json:"l2_delta"`
	Cosine               float64 `json:"cosine"`
	FirstChanged         []int   `json:"first_changed_bins,omitempty"`
}

func extract(text string) side {
	s := side{Text: text, Bytes: len(text), Runes: utf8.RuneCountInString(text)}
	inputDigest := sha256.Sum256([]byte(text))
	s.InputDigest, s.ValidUTF8 = hex.EncodeToString(inputDigest[:]), utf8.ValidString(text)
	var workspace statehintwide.Workspace
	view, err := statehintwide.ExtractContextual(text, &workspace)
	if err != nil {
		s.Error = err.Error()
		return s
	}
	s.Words, s.Active = view.WordCount(), view.Len()
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		s.dense[f.Index], s.seen[f.Index] = f.Value, true
		s.order = append(s.order, f.Index)
		if f.Value == 0 {
			s.Zero++
		} else {
			s.Nonzero++
		}
		s.Norm += float64(f.Value) * float64(f.Value)
	}
	s.Norm = math.Sqrt(s.Norm)
	var encoded [statehintwide.FeatureBins * 4]byte
	for i, v := range s.dense {
		binary.LittleEndian.PutUint32(encoded[i*4:i*4+4], math.Float32bits(v))
	}
	digest := sha256.Sum256(encoded[:])
	s.Digest = hex.EncodeToString(digest[:])
	return s
}

func compare(p pair) result {
	r := result{ID: p.ID, Focus: p.Focus, A: extract(p.A), B: extract(p.B)}
	r.Comparable = r.A.Error == "" && r.B.Error == ""
	if !r.Comparable {
		return r
	}
	r.DenseBitIdentical, r.WordCountIdentical, r.SparseOrderIdentical = true, r.A.Words == r.B.Words, len(r.A.order) == len(r.B.order)
	if r.SparseOrderIdentical {
		for i := range r.A.order {
			if r.A.order[i] != r.B.order[i] {
				r.SparseOrderIdentical = false
				break
			}
		}
	}
	var dot float64
	for i, a := range r.A.dense {
		b := r.B.dense[i]
		if math.Float32bits(a) != math.Float32bits(b) {
			r.DenseBitIdentical = false
			r.ChangedBins++
			if len(r.FirstChanged) < 16 {
				r.FirstChanged = append(r.FirstChanged, i)
			}
		}
		if !r.A.seen[i] && r.B.seen[i] {
			r.AddedActive++
		}
		if r.A.seen[i] && !r.B.seen[i] {
			r.RemovedActive++
		}
		d := float64(b) - float64(a)
		r.MaxAbsoluteDelta = math.Max(r.MaxAbsoluteDelta, math.Abs(d))
		r.L2Delta += d * d
		dot += float64(a) * float64(b)
	}
	r.L2Delta = math.Sqrt(r.L2Delta)
	if r.A.Norm*r.B.Norm > 0 {
		r.Cosine = dot / (r.A.Norm * r.B.Norm)
	}
	return r
}

func main() {
	capacityBase := strings.Repeat("q ", 2047) + "q"
	koreanSentence := "청록색 고리를 차례로 닦습니다. "
	koreanBase := strings.Repeat(koreanSentence, 4095/len(koreanSentence))
	koreanBase += strings.Repeat(" ", 4095-len(koreanBase)) + "."
	probes := []pair{
		{"p01", "English terminal question cue", "I am braiding the azure cord.", "I am braiding the azure cord?"},
		{"p02", "English local negation", "I have lacquered the brass tile.", "I have not lacquered the brass tile."},
		{"p03", "English quotation punctuation", "I am etching the jade token.", "\"I am etching the jade token.\""},
		{"p04", "English tense/aspect wording", "I am sanding the ivory block.", "I will sand the ivory block."},
		{"p05", "Korean question inflection", "청록색 고리를 닦았습니다.", "청록색 고리를 닦았습니까?"},
		{"p06", "Korean negation inflection", "산호색 봉투를 접었습니다.", "산호색 봉투를 접지 않았습니다."},
		{"p07", "Korean tense/aspect inflection", "자주색 나무패를 다듬고 있습니다.", "자주색 나무패를 다듬을 예정입니다."},
		{"p08", "English quoted versus direct clause scope permutation", "The label reads: \"I am polishing the amber spoon now.\" I am polishing the cobalt spoon now.", "The label reads: \"I am polishing the cobalt spoon now.\" I am polishing the amber spoon now."},
		{"p09", "English adoption/negation scope permutation", "I accept this note as mine: \"I have polished the amber spoon today.\" I do not accept this note as mine: \"I have polished the cobalt spoon today.\"", "I accept this note as mine: \"I have polished the cobalt spoon today.\" I do not accept this note as mine: \"I have polished the amber spoon today.\""},
		{"p10", "Korean quoted versus direct clause scope permutation", "쪽지에는 “나는 지금 청록색 고리를 닦고 있습니다.”라고 적혀 있습니다. 나는 지금 산호색 고리를 닦고 있습니다.", "쪽지에는 “나는 지금 산호색 고리를 닦고 있습니다.”라고 적혀 있습니다. 나는 지금 청록색 고리를 닦고 있습니다."},
		{"p11", "Synthetic maximum-word-count 4096-byte late punctuation cue", capacityBase + ".", capacityBase + "?"},
		{"p12", "Valid Korean UTF-8 byte boundary 4096 versus 4097", koreanBase, koreanBase + "!"},
	}
	if len(probes) > 12 {
		panic("probe budget exceeded")
	}
	results := make([]result, 0, len(probes))
	accepted, collisions := 0, 0
	for _, p := range probes {
		r := compare(p)
		if r.A.Error == "" {
			accepted++
		}
		if r.B.Error == "" {
			accepted++
		}
		if r.Comparable && r.DenseBitIdentical && r.WordCountIdentical {
			collisions++
		}
		results = append(results, r)
	}
	report := struct {
		Purpose    string   `json:"purpose"`
		Exposure   string   `json:"exposure"`
		Schema     string   `json:"feature_schema"`
		Pairs      int      `json:"original_unlabelled_pairs"`
		Accepted   int      `json:"accepted_inputs"`
		Collisions int      `json:"identical_dense_feature_and_word_count_pairs"`
		Results    []result `json:"results"`
	}{"Feature information audit only; no semantic labels, classifier queries, model loads, training or model selection.", "All own inputs are development-exposed and excluded from final 2400 and selectors.", statehintwide.ContextualFeatureSchema, len(probes), accepted, collisions, results}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
