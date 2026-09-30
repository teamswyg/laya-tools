package inference

import (
	"encoding/json"
	"fmt"
	"github.com/dlclark/regexp2"
	"golang.org/x/text/unicode/norm"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type pair struct{ a, b string }
type added struct {
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	SingleWord bool   `json:"single_word"`
	LStrip     bool   `json:"lstrip"`
	RStrip     bool   `json:"rstrip"`
}
type Encoder struct {
	vocab          map[string]int64
	ranks          map[pair]int
	added          []added
	pattern        *regexp2.Regexp
	byteMap        [256]string
	cache          map[string][]int64
	cls, sep, mask int64
}
type Sequence struct {
	IDs       []int64 `json:"ids"`
	Markers   []int64 `json:"markers"`
	Truncated bool    `json:"truncated"`
}

// LoadEncoder implements the pinned English Laya tokenizer: NFC, added tokens,
// GPT-2 byte-level pretokenization, and rank-ordered BPE. Other formats are rejected.
func LoadEncoder(path string) (*Encoder, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c struct {
		Model struct {
			Type         string            `json:"type"`
			Vocab        map[string]int64  `json:"vocab"`
			Merges       []json.RawMessage `json:"merges"`
			ByteFallback bool              `json:"byte_fallback"`
			IgnoreMerges bool              `json:"ignore_merges"`
		} `json:"model"`
		Normalizer struct {
			Type string `json:"type"`
		} `json:"normalizer"`
		Pre struct {
			Type   string `json:"type"`
			Prefix bool   `json:"add_prefix_space"`
			Regex  bool   `json:"use_regex"`
		} `json:"pre_tokenizer"`
		Added []added `json:"added_tokens"`
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.Model.Type != "BPE" || c.Model.ByteFallback || c.Model.IgnoreMerges || c.Normalizer.Type != "NFC" || c.Pre.Type != "ByteLevel" || c.Pre.Prefix || !c.Pre.Regex {
		return nil, fmt.Errorf("unsupported tokenizer: need pinned English Laya NFC/ByteLevel/BPE")
	}
	pattern := regexp2.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`, 0)
	pattern.MatchTimeout = time.Second
	e := &Encoder{vocab: c.Model.Vocab, ranks: map[pair]int{}, cache: map[string][]int64{}, added: c.Added, pattern: pattern}
	for i, raw := range c.Model.Merges {
		var p []string
		if json.Unmarshal(raw, &p) != nil {
			var s string
			if err = json.Unmarshal(raw, &s); err != nil {
				return nil, err
			}
			p = strings.Split(s, " ")
		}
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid merge")
		}
		e.ranks[pair{p[0], p[1]}] = i
	}
	for _, a := range e.added {
		e.vocab[a.Content] = a.ID
		if a.Content == "" || a.SingleWord || (a.LStrip && a.Content != "[MASK]") || a.RStrip {
			return nil, fmt.Errorf("unsupported added token flags")
		}
	}
	sort.Slice(e.added, func(i, j int) bool { return len(e.added[i].Content) > len(e.added[j].Content) })
	n := 0
	for i := 0; i < 256; i++ {
		if i >= 33 && i <= 126 || i >= 161 && i <= 172 || i >= 174 {
			e.byteMap[i] = string(rune(i))
		} else {
			e.byteMap[i] = string(rune(256 + n))
			n++
		}
	}
	for i, s := range []string{"[CLS]", "[SEP]", "[MASK]"} {
		v, ok := e.vocab[s]
		if !ok {
			return nil, fmt.Errorf("missing %s", s)
		}
		switch i {
		case 0:
			e.cls = v
		case 1:
			e.sep = v
		case 2:
			e.mask = v
		}
	}
	return e, nil
}
func (e *Encoder) bpe(s string) ([]int64, error) {
	if ids, ok := e.cache[s]; ok {
		return ids, nil
	}
	parts := make([]string, len(s))
	for i := range len(s) {
		parts[i] = e.byteMap[s[i]]
	}
	for len(parts) > 1 {
		best, at := int(^uint(0)>>1), -1
		for i := 0; i < len(parts)-1; i++ {
			if rank, ok := e.ranks[pair{parts[i], parts[i+1]}]; ok && rank < best {
				best = rank
				at = i
			}
		}
		if at < 0 {
			break
		}
		a, b := parts[at], parts[at+1]
		merged := parts[:0]
		for i := 0; i < len(parts); i++ {
			if i+1 < len(parts) && parts[i] == a && parts[i+1] == b {
				merged = append(merged, a+b)
				i++
			} else {
				merged = append(merged, parts[i])
			}
		}
		parts = merged
	}
	ids := make([]int64, len(parts))
	for i, p := range parts {
		id, ok := e.vocab[p]
		if !ok {
			return nil, fmt.Errorf("BPE token missing from vocabulary")
		}
		ids[i] = id
	}
	if len(e.cache) < 4096 {
		e.cache[s] = ids
	}
	return ids, nil
}
func (e *Encoder) Encode(s string) ([]int64, error) {
	s = norm.NFC.String(strings.ReplaceAll(s, "[MASK]", " "))
	var ids []int64
	regular := func(text string) error {
		m, err := e.pattern.FindStringMatch(text)
		for m != nil && err == nil {
			p, e2 := e.bpe(m.String())
			if e2 != nil {
				return e2
			}
			ids = append(ids, p...)
			m, err = e.pattern.FindNextMatch(m)
		}
		return err
	}
	start := 0
	for i := 0; i < len(s); {
		found := false
		for _, a := range e.added {
			if strings.HasPrefix(s[i:], a.Content) {
				if err := regular(s[start:i]); err != nil {
					return nil, err
				}
				ids = append(ids, a.ID)
				i += len(a.Content)
				start = i
				found = true
				break
			}
		}
		if !found {
			_, n := utf8.DecodeRuneInString(s[i:])
			i += n
		}
	}
	if err := regular(s[start:]); err != nil {
		return nil, err
	}
	return ids, nil
}

// Build matches laya.common.build_sequence for choice and noul, including head budgets.
func (e *Encoder) Build(state, kind, instruction string, options []string, maxLen int) (Sequence, error) {
	if len(options) < 2 || len(options) > 10 || maxLen < 256 || maxLen > 512 {
		return Sequence{}, fmt.Errorf("need 2..10 options and 256..512 tokens")
	}
	head, err := e.Encode(kind + " question: " + instruction)
	if err != nil {
		return Sequence{}, err
	}
	opts := make([][]int64, len(options))
	budget := 192
	for i, s := range options {
		ids, err := e.Encode(" " + s)
		if err != nil {
			return Sequence{}, err
		}
		opts[i] = append([]int64{e.mask}, ids[:min(48, len(ids))]...)
		budget -= len(opts[i])
	}
	if budget < 16 {
		per := max(4, (192-16)/len(options))
		budget = 192
		for i, o := range opts {
			opts[i] = o[:min(per, len(o))]
			budget -= len(opts[i])
		}
	}
	ids := append([]int64{e.cls}, head[:min(len(head), max(8, budget))]...)
	ids = append(ids, e.sep)
	var markers []int64
	for _, o := range opts {
		markers = append(markers, int64(len(ids)))
		ids = append(ids, o...)
	}
	ids = append(ids, e.sep)
	st, err := e.Encode(state)
	if err != nil {
		return Sequence{}, err
	}
	room := maxLen - len(ids) - 1
	truncated := len(st) > room || len(head) > max(8, budget)
	ids = append(ids, st[:min(room, len(st))]...)
	ids = append(ids, e.sep)
	return Sequence{IDs: ids, Markers: markers, Truncated: truncated}, nil
}
