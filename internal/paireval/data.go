// Package paireval provides a grouped external relevance evaluation. It does
// not redistribute source text and does not equate pair labels with retrieval.
package paireval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

const SourceSHA = "4d1a227fe3652992d1e98f43fa9dc1b8f54a6c27a1a24b9d182bdfb8cbc3274c"

type Row struct {
	Code       string          `json:"code"`
	Query      string          `json:"doc"`
	ID         string          `json:"idx"`
	Label      *int            `json:"label"`
	CodeTokens json.RawMessage `json:"code_tokens"`
	DocTokens  json.RawMessage `json:"docstring_tokens"`
}

func Hash(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func Load(path string) ([]Row, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if e != nil {
		return nil, e
	}
	if Hash(b) != SourceSHA {
		return nil, fmt.Errorf("source hash mismatch")
	}
	var rows []Row
	if e = json.Unmarshal(b, &rows); e != nil {
		return nil, e
	}
	for i := range rows {
		v := &rows[i]
		if v.Label == nil || (*v.Label != 0 && *v.Label != 1) || v.Code == "" || v.Query == "" {
			return nil, fmt.Errorf("invalid row")
		}
		v.CodeTokens = nil
		v.DocTokens = nil
	}
	return rows, nil
}
func Words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
func InScope(r Row) bool {
	return len(r.Query) <= 4096 && len(r.Code) <= 4096 && len(Words(r.Query)) <= 64 && len(Words(r.Code)) <= 64
}
func NormalizeQuery(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
func NormalizeCode(s string) string  { return strings.Join(strings.Fields(s), " ") }

type Membership struct {
	Group int
	Split string
}
type Split struct {
	Schema, SourceSHA256, Seed, MembershipSHA256 string
	Groups                                       int
	Rows                                         []Membership
	Counts                                       map[string]int
	GroupCounts                                  map[string]int
}
type key struct {
	text string
	row  int
}

func Partition(rows []Row) Split {
	keys := make([]key, 0, 2*len(rows))
	parent := make([]int, len(rows))
	for i, r := range rows {
		parent[i] = i
		keys = append(keys, key{"q" + Hash([]byte(NormalizeQuery(r.Query))), i}, key{"c" + Hash([]byte(NormalizeCode(r.Code))), i})
	}
	find := func(i int) int { return i }
	find = func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].text == keys[j].text {
			return keys[i].row < keys[j].row
		}
		return keys[i].text < keys[j].text
	})
	for i := 1; i < len(keys); i++ {
		if keys[i].text == keys[i-1].text {
			a, b := find(keys[i].row), find(keys[i-1].row)
			parent[a] = b
		}
	}
	canonical := make([]string, len(rows))
	sizes := make([]int, len(rows))
	for _, k := range keys {
		root := find(k.row)
		if canonical[root] == "" || k.text < canonical[root] {
			canonical[root] = k.text
		}
	}
	for i := range rows {
		sizes[find(i)]++
	}
	type group struct {
		root, size int
		key        string
	}
	gs := []group{}
	seed := "riido-pair-groups-01"
	for i, n := range sizes {
		if n > 0 {
			gs = append(gs, group{i, n, Hash([]byte(seed + canonical[i]))})
		}
	}
	sort.Slice(gs, func(i, j int) bool { return gs[i].key < gs[j].key })
	names := []string{"final", "reserve1", "reserve2", "validation", "calibration", "development"}
	minimum := []int{2400, 2400, 2400, 600, 600}
	s := Split{Schema: "riido-pair-split-v1", SourceSHA256: SourceSHA, Seed: seed, Groups: len(gs), Rows: make([]Membership, len(rows)), Counts: map[string]int{}, GroupCounts: map[string]int{}}
	assigned := make([]Membership, len(rows))
	stage := 0
	for i, g := range gs {
		for stage < len(minimum) && s.Counts[names[stage]] >= minimum[stage] {
			stage++
		}
		name := names[stage]
		assigned[g.root] = Membership{i, name}
		s.Counts[name] += g.size
		s.GroupCounts[name]++
	}
	for i := range rows {
		s.Rows[i] = assigned[find(i)]
	}
	b, _ := json.Marshal(s.Rows)
	s.MembershipSHA256 = Hash(b)
	return s
}
func Indices(s Split, name string) []int {
	out := []int{}
	for i, r := range s.Rows {
		if r.Split == name {
			out = append(out, i)
		}
	}
	return out
}

// Lexical uses development-only document statistics; scores are not calibrated
// probabilities and unknown query terms carry no collection evidence.
type Lexical struct {
	terms     []string
	df        []int
	documents int
	average   float64
}

func NewLexical(rows []Row, s Split) *Lexical {
	l := &Lexical{}
	type doc struct{ k, text string }
	docs := []doc{}
	for i, r := range rows {
		if s.Rows[i].Split == "development" {
			docs = append(docs, doc{Hash([]byte(NormalizeCode(r.Code))), r.Code})
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].k < docs[j].k })
	all := []string{}
	for i, d := range docs {
		if i > 0 && docs[i-1].k == d.k {
			continue
		}
		ts := Words(d.text)
		l.average += float64(len(ts))
		l.documents++
		sort.Strings(ts)
		for j, t := range ts {
			if j == 0 || t != ts[j-1] {
				all = append(all, t)
			}
		}
	}
	sort.Strings(all)
	for _, t := range all {
		n := len(l.terms)
		if n > 0 && l.terms[n-1] == t {
			l.df[n-1]++
		} else {
			l.terms = append(l.terms, t)
			l.df = append(l.df, 1)
		}
	}
	if l.documents > 0 {
		l.average /= float64(l.documents)
	}
	if l.average == 0 {
		l.average = 1
	}
	return l
}
