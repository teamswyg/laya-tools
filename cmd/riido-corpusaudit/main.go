// riido-corpusaudit emits aggregate source statistics without publishing text.
// This is a data audit, not model training or a quality benchmark.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

type row struct {
	Code       string          `json:"code"`
	Query      string          `json:"doc"`
	ID         string          `json:"idx"`
	Label      *int            `json:"label"`
	CodeTokens json.RawMessage `json:"code_tokens"`
	DocTokens  json.RawMessage `json:"docstring_tokens"`
}
type key struct {
	Value string
	Row   int
}

func digest(s string) string     { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func normalized(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
func wordCount(s string) int {
	return len(strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }))
}

type report struct {
	Schema, SourceSHA256                                                                                                                                       string
	Rows, Positive, Negative, UniqueQueries, UniqueCodes, ConnectedGroups, LargestGroup, DuplicatePairs, ConflictingPairs, WithinCurrentInputScope, MissingIDs int
	HasRepositoryMetadata                                                                                                                                      bool
}

func audit(in io.Reader) (report, error) {
	r := report{Schema: "riido-cosqa-audit-v1"}
	h := sha256.New()
	d := json.NewDecoder(io.TeeReader(io.LimitReader(in, (64<<20)+1), h))
	d.DisallowUnknownFields()
	tok, e := d.Token()
	if e != nil || tok != json.Delim('[') {
		return r, fmt.Errorf("expected JSON array")
	}
	var keys []key
	var pairs []key
	labels := []int{}
	for d.More() {
		if r.Rows >= 100000 {
			return r, fmt.Errorf("too many rows")
		}
		var v row
		if e = d.Decode(&v); e != nil {
			return r, e
		}
		if strings.TrimSpace(v.Query) == "" || strings.TrimSpace(v.Code) == "" || (v.Label == nil || (*v.Label != 0 && *v.Label != 1)) {
			return r, fmt.Errorf("invalid row")
		}
		q := digest(normalized(v.Query))
		c := digest(strings.TrimSpace(v.Code))
		keys = append(keys, key{"q" + q, r.Rows}, key{"c" + c, r.Rows})
		pairs = append(pairs, key{q + c, r.Rows})
		labels = append(labels, *v.Label)
		if *v.Label == 1 {
			r.Positive++
		} else {
			r.Negative++
		}
		if v.ID == "" {
			r.MissingIDs++
		}
		if len(v.Query) <= 4096 && len(v.Code) <= 4096 && wordCount(v.Query) <= 64 && wordCount(v.Code) <= 64 {
			r.WithinCurrentInputScope++
		}
		r.Rows++
	}
	if _, e = d.Token(); e != nil {
		return r, e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return r, fmt.Errorf("unexpected trailing data")
	}
	r.SourceSHA256 = hex.EncodeToString(h.Sum(nil))
	parents := make([]int, r.Rows)
	for i := range parents {
		parents[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		for parents[i] != i {
			parents[i] = parents[parents[i]]
			i = parents[i]
		}
		return i
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Value < keys[j].Value })
	for i, k := range keys {
		if i > 0 && keys[i-1].Value == k.Value {
			a, b := find(keys[i-1].Row), find(k.Row)
			parents[a] = b
		} else if k.Value[0] == 'q' {
			r.UniqueQueries++
		} else {
			r.UniqueCodes++
		}
	}
	sizes := make([]int, r.Rows)
	for i := range parents {
		sizes[find(i)]++
	}
	for _, n := range sizes {
		if n > 0 {
			r.ConnectedGroups++
			r.LargestGroup = max(r.LargestGroup, n)
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Value < pairs[j].Value })
	for i := 0; i < len(pairs); {
		j := i + 1
		conflict := false
		for j < len(pairs) && pairs[j].Value == pairs[i].Value {
			if labels[pairs[j].Row] != labels[pairs[i].Row] {
				conflict = true
			}
			j++
		}
		r.DuplicatePairs += j - i - 1
		if conflict {
			r.ConflictingPairs++
		}
		i = j
	}
	return r, nil
}
func main() {
	path := flag.String("input", "", "local source JSON; content never printed")
	expected := flag.String("sha256", "", "required pinned source SHA-256")
	flag.Parse()
	if len(*expected) != 64 {
		fmt.Fprintln(os.Stderr, "pinned source sha256 required")
		os.Exit(1)
	}
	f, e := os.Open(*path)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || info.Size() > 64<<20 {
		fmt.Fprintln(os.Stderr, "source exceeds limit")
		os.Exit(1)
	}
	r, e := audit(f)
	if e == nil && r.SourceSHA256 != *expected {
		e = fmt.Errorf("source SHA mismatch")
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if e = enc.Encode(r); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
