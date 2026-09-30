package staticembed

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	text string
	id   int
}
type tokenizer struct {
	vocab   []token
	special []token
	unknown int
}

func parseTokenizer(b []byte) (tokenizer, error) {
	var t tokenizer
	var c struct {
		Model struct {
			Type  string
			Vocab map[string]int
		}
		Added []struct {
			ID      int
			Content string
		} `json:"added_tokens"`
	}
	if e := json.Unmarshal(b, &c); e != nil {
		return t, e
	}
	if c.Model.Type != "WordPiece" || len(c.Model.Vocab) != Rows {
		return t, fmt.Errorf("unsupported vocabulary")
	}
	// The caller verifies the exact source hash; normalization, WordPiece limits
	// and added-token flags below implement this immutable artifact only.
	t.vocab = make([]token, 0, len(c.Model.Vocab))
	for s, id := range c.Model.Vocab {
		if id < 0 || id >= Rows {
			return t, fmt.Errorf("invalid token ID")
		}
		t.vocab = append(t.vocab, token{s, id})
	}
	slices.SortFunc(t.vocab, func(a, b token) int { return strings.Compare(a.text, b.text) })
	for _, a := range c.Added {
		t.special = append(t.special, token{a.Content, a.ID})
	}
	var ok bool
	t.unknown, ok = t.lookup("[UNK]")
	if !ok {
		return t, fmt.Errorf("missing unknown token")
	}
	return t, nil
}
func (t *tokenizer) lookup(s string) (int, bool) {
	i, ok := slices.BinarySearchFunc(t.vocab, s, func(a token, b string) int { return strings.Compare(a.text, b) })
	if !ok {
		return 0, false
	}
	return t.vocab[i].id, true
}
func chinese(r rune) bool {
	return r >= 0x4e00 && r <= 0x9fff || r >= 0x3400 && r <= 0x4dbf || r >= 0x20000 && r <= 0x2a6df || r >= 0x2a700 && r <= 0x2b73f || r >= 0x2b740 && r <= 0x2b81f || r >= 0x2b820 && r <= 0x2ceaf || r >= 0xf900 && r <= 0xfaff || r >= 0x2f800 && r <= 0x2fa1f
}
func punctuation(r rune) bool {
	return r >= 33 && r <= 47 || r >= 58 && r <= 64 || r >= 91 && r <= 96 || r >= 123 && r <= 126 || unicode.IsPunct(r)
}
func normalize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == 0 || r == utf8.RuneError || ((unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r)) && r != '\t' && r != '\n' && r != '\r') {
			continue
		}
		if unicode.IsSpace(r) {
			b.WriteByte(' ')
		} else if chinese(r) {
			b.WriteByte(' ')
			b.WriteRune(r)
			b.WriteByte(' ')
		} else {
			b.WriteRune(r)
		}
	}
	s = norm.NFD.String(strings.ToLower(b.String()))
	b.Reset()
	for _, r := range s {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func (t *tokenizer) segment(s string, ids []int) []int {
	rs := []rune(normalize(s))
	start := 0
	emit := func(word string) {
		wr := []rune(word)
		if len(wr) == 0 || len(wr) > 100 {
			return
		}
		saved := len(ids)
		for begin := 0; begin < len(wr); {
			end := len(wr)
			found := false
			for end > begin {
				sub := string(wr[begin:end])
				if begin > 0 {
					sub = "##" + sub
				}
				id, ok := t.lookup(sub)
				if ok {
					if id != t.unknown {
						ids = append(ids, id)
					}
					found = true
					break
				}
				end--
			}
			if !found {
				ids = ids[:saved]
				return
			}
			begin = end
		}
	}
	for i, r := range rs {
		if unicode.IsSpace(r) || punctuation(r) {
			emit(string(rs[start:i]))
			if punctuation(r) {
				emit(string(r))
			}
			start = i + 1
		}
	}
	emit(string(rs[start:]))
	return ids
}
func (t *tokenizer) encode(s string) []int {
	ids := []int{}
	start := 0
	for i := 0; i < len(s); {
		matched := false
		for _, a := range t.special {
			if strings.HasPrefix(s[i:], a.text) {
				ids = t.segment(s[start:i], ids)
				if a.id != t.unknown {
					ids = append(ids, a.id)
				}
				i += len(a.text)
				start = i
				matched = true
				break
			}
		}
		if !matched {
			_, n := utf8.DecodeRuneInString(s[i:])
			i += n
		}
	}
	return t.segment(s[start:], ids)
}
