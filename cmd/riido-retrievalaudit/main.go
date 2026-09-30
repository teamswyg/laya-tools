// riido-retrievalaudit audits local public-corpus projections without scoring models.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

type row struct{ Repository, Path, Name, Code, Query, URL string }
type group struct {
	Rows         int `json:"rows"`
	Repositories int `json:"repositories"`
}
type report struct {
	Schema           string   `json:"schema"`
	ProjectionSHA256 string   `json:"projection_sha256"`
	Rows             int      `json:"rows"`
	Repositories     int      `json:"repositories"`
	UniqueQueries    int      `json:"unique_normalized_queries"`
	UniqueCode       int      `json:"unique_code_token_sequences"`
	CodeScanErrors   int      `json:"code_scan_errors"`
	EmptyQueries     int      `json:"empty_queries"`
	ShortQueries     int      `json:"queries_under_three_words"`
	QueryInCode      int      `json:"normalized_query_verbatim_in_code"`
	NameInQuery      int      `json:"function_name_in_query"`
	VendorPaths      int      `json:"vendor_or_third_party_paths"`
	GeneratedHints   int      `json:"generated_path_or_function_marker_hints"`
	Groups           []group  `json:"repository_query_code_connected_groups"`
	ModelScored      bool     `json:"model_scored"`
	LicenseCleared   bool     `json:"license_cleared"`
	Limitations      []string `json:"limitations"`
}

func normalize(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }

// Ignore comments and layout, but preserve literal contents and token boundaries.
// This is exact token deduplication, not variable-renaming/semantic clone detection.
func codeKey(code string) (string, bool) {
	fset := token.NewFileSet()
	file := fset.AddFile("", -1, len(code))
	var s scanner.Scanner
	valid := true
	s.Init(file, []byte(code), func(token.Position, string) { valid = false }, 0)
	var out strings.Builder
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			lit = ";"
		}
		fmt.Fprintf(&out, "%d:%d:%s;", tok, len(lit), lit)
	}
	if !valid {
		return "raw:" + code, false
	}
	return out.String(), true
}

type unions []int

func (u unions) root(i int) int {
	for u[i] != i {
		u[i] = u[u[i]]
		i = u[i]
	}
	return i
}
func (u unions) join(a, b int) {
	a = u.root(a)
	b = u.root(b)
	if a != b {
		u[b] = a
	}
}

func audit(rows []row) report {
	r := report{Schema: "riido-retrieval-source-audit-v1", Rows: len(rows), Limitations: []string{
		"Documentation queries are proxies, not independently written user requests.",
		"Generated-code hints are incomplete: function excerpts lack file headers.",
		"Root repository licenses do not clear per-file or third-party obligations.",
		"Token deduplication does not identify all semantic clones or pretraining overlap.",
		"Counts are descriptive; no eligibility filter, split, model selection or scoring occurred.",
	}}
	u := make(unions, len(rows))
	for i := range u {
		u[i] = i
	}
	repos := map[string]int{}
	queries := map[string]int{}
	codes := map[string]int{}
	link := func(m map[string]int, key string, i int) {
		if old, ok := m[key]; ok {
			u.join(i, old)
		} else {
			m[key] = i
		}
	}
	for i, v := range rows {
		q := normalize(v.Query)
		c, valid := codeKey(v.Code)
		if !valid {
			r.CodeScanErrors++
		}
		link(repos, v.Repository, i)
		if q != "" {
			link(queries, q, i)
		}
		link(codes, c, i)
		if q == "" {
			r.EmptyQueries++
		} else if strings.Contains(normalize(v.Code), q) {
			r.QueryInCode++
		}
		if len(strings.Fields(q)) < 3 {
			r.ShortQueries++
		}
		name := v.Name
		if dot := strings.LastIndex(name, "."); dot >= 0 {
			name = name[dot+1:]
		}
		for _, word := range strings.FieldsFunc(q, func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' }) {
			if name != "" && word == strings.ToLower(name) {
				r.NameInQuery++
				break
			}
		}
		p := "/" + strings.ToLower(v.Path) + "/"
		if strings.Contains(p, "/vendor/") || strings.Contains(p, "/third_party/") || strings.Contains(p, "/third-party/") {
			r.VendorPaths++
		}
		if strings.Contains(p, ".pb.go/") || strings.Contains(p, "/zz_generated") || strings.Contains(p, "/generated/") || strings.Contains(strings.ToLower(v.Code), "code generated") {
			r.GeneratedHints++
		}
	}
	r.Repositories = len(repos)
	r.UniqueQueries = len(queries)
	r.UniqueCode = len(codes)
	counts := map[int]int{}
	groupRepos := map[int]map[string]bool{}
	for i, v := range rows {
		root := u.root(i)
		counts[root]++
		if groupRepos[root] == nil {
			groupRepos[root] = map[string]bool{}
		}
		groupRepos[root][v.Repository] = true
	}
	for root, n := range counts {
		r.Groups = append(r.Groups, group{n, len(groupRepos[root])})
	}
	sort.Slice(r.Groups, func(i, j int) bool {
		if r.Groups[i].Rows != r.Groups[j].Rows {
			return r.Groups[i].Rows > r.Groups[j].Rows
		}
		return r.Groups[i].Repositories > r.Groups[j].Repositories
	})
	return r
}

func run(input string) error {
	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	limited := &io.LimitedReader{R: f, N: 64*1024*1024 + 1}
	s := bufio.NewScanner(io.TeeReader(limited, h))
	s.Buffer(make([]byte, 65536), 2*1024*1024)
	var rows []row
	for s.Scan() {
		var v row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if err := d.Decode(&v); err != nil {
			return fmt.Errorf("row %d: invalid JSON", len(rows)+1)
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("row %d: trailing JSON", len(rows)+1)
		}
		if v.Repository == "" || v.Code == "" || v.Path == "" {
			return fmt.Errorf("row %d: missing source fields", len(rows)+1)
		}
		rows = append(rows, v)
		if len(rows) > 100000 {
			return fmt.Errorf("row limit exceeded")
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	if limited.N == 0 {
		return fmt.Errorf("input limit exceeded")
	}
	if len(rows) == 0 {
		return fmt.Errorf("empty input")
	}
	r := audit(rows)
	r.ProjectionSHA256 = hex.EncodeToString(h.Sum(nil))
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	return e.Encode(r)
}

func main() {
	input := flag.String("input", "", "local JSONL projection; never publish raw rows")
	flag.Parse()
	if err := run(*input); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
