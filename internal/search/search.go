// Package search retrieves small, line-addressable code excerpts before optional Laya reranking.
package search

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/inference"
)

type Chunk struct {
	Path      string   `json:"path"`
	Start     int      `json:"start"`
	End       int      `json:"end"`
	Text      string   `json:"text"`
	Score     float64  `json:"score"`
	Lexical   float64  `json:"lexical_score"`
	Relevance *float64 `json:"relevance,omitempty"`
	Truncated bool     `json:"truncated,omitempty"`
}

// posting separates hot numeric arrays from cold source text (SoA).
type posting struct {
	chunks  []int
	weights []float64
}

// Index is immutable after Load. Concurrent searches use request-local scores.
// Callers must not modify Chunks, including while a Search is running.
type Index struct {
	Chunks     []Chunk
	vocabulary map[string]int
	postings   []posting
	lengths    []int
	avg        float64
	Bytes      int
	Files      int
}
type Scorer interface {
	Predict(string, string, string, []string) (inference.Prediction, error)
}
type Result struct {
	Results    []Chunk  `json:"results"`
	Candidates int      `json:"candidates"`
	Files      int      `json:"files"`
	Chunks     int      `json:"chunks"`
	Warnings   []string `json:"warnings,omitempty"`
	Engine     string   `json:"engine"`
}

var word = regexp.MustCompile(`[\pL\pN]+`)
var camel = regexp.MustCompile(`([a-z0-9])([A-Z])`)
var allowed = map[string]bool{".go": true, ".py": true, ".rs": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".java": true, ".c": true, ".h": true, ".cpp": true, ".swift": true, ".kt": true, ".rb": true, ".sh": true, ".md": true, ".json": true, ".yaml": true, ".yml": true, ".toml": true}
var excluded = map[string]bool{"node_modules": true, "vendor": true, "dist": true, "build": true, "target": true, "__pycache__": true}

func Terms(s string) []string {
	return word.FindAllString(strings.ToLower(camel.ReplaceAllString(s, "$1 $2")), -1)
}
func safePath(path string) bool {
	if filepath.IsAbs(path) || strings.Contains(path, "\\") {
		return false
	}
	for _, p := range strings.Split(filepath.ToSlash(path), "/") {
		if strings.HasPrefix(p, ".") || excluded[p] || p == ".." {
			return false
		}
	}
	base := strings.ToLower(filepath.Base(path))
	if strings.Contains(base, "secret") || strings.Contains(base, "credential") || strings.HasPrefix(base, "id_rsa") || strings.HasSuffix(base, "lock.json") {
		return false
	}
	return allowed[strings.ToLower(filepath.Ext(path))]
}

// Git supplies ignore semantics; ripgrep is a fallback for non-Git directories.
func fileList(root string) ([]string, error) {
	gitRoot, err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output()
	var out []byte
	if err == nil && strings.TrimSpace(string(gitRoot)) == root {
		out, err = exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard").Output()
	} else {
		cmd := exec.Command("rg", "--files", "-0")
		cmd.Dir = root
		out, err = cmd.Output()
		if x, ok := err.(*exec.ExitError); ok && x.ExitCode() == 1 {
			err = nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("list files: use a Git repository root or install ripgrep: %w", err)
	}
	paths := strings.Split(string(out), "\x00")
	sort.Strings(paths)
	return paths, nil
}
func Load(root string) (*Index, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	paths, err := fileList(root)
	if err != nil {
		return nil, err
	}
	idx := &Index{vocabulary: map[string]int{}}
	for _, path := range paths {
		if !safePath(path) {
			continue
		}
		full := filepath.Join(root, path)
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil || resolved != full {
			continue
		}
		info, err := os.Lstat(full)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 256*1024 {
			continue
		}
		if idx.Bytes+int(info.Size()) > 32*1024*1024 {
			return nil, fmt.Errorf("source exceeds 32 MiB budget; search a smaller root")
		}
		b, err := os.ReadFile(full)
		if err != nil {
			return nil, err
		}
		if bytes.IndexByte(b, 0) >= 0 || !utf8.Valid(b) {
			continue
		}
		idx.Bytes += len(b)
		idx.Files++
		lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
		for start := 0; start < len(lines); start += 24 {
			end := min(start+32, len(lines))
			text := strings.Join(lines[start:end], "\n")
			tf := map[string]int{}
			terms := Terms(path + "\n" + text)
			for _, t := range terms {
				tf[t]++
			}
			for t, count := range tf {
				id, ok := idx.vocabulary[t]
				if !ok {
					id = len(idx.postings)
					idx.vocabulary[t] = id
					idx.postings = append(idx.postings, posting{})
				}
				post := &idx.postings[id]
				post.chunks = append(post.chunks, len(idx.Chunks))
				post.weights = append(post.weights, float64(count))
			}
			idx.lengths = append(idx.lengths, len(terms))
			idx.avg += float64(len(terms))
			idx.Chunks = append(idx.Chunks, Chunk{Path: filepath.ToSlash(path), Start: start + 1, End: end, Text: text})
			if len(idx.Chunks) > 50000 {
				return nil, fmt.Errorf("source exceeds 50000 chunks; search a smaller root")
			}
			if end == len(lines) {
				break
			}
		}
	}
	if len(idx.Chunks) > 0 {
		idx.avg /= float64(len(idx.Chunks))
		for i := range idx.postings {
			p := &idx.postings[i]
			df := len(p.chunks)
			idf := math.Log(1 + (float64(len(idx.Chunks)-df)+.5)/(float64(df)+.5))
			for j, chunk := range p.chunks {
				f := p.weights[j]
				p.weights[j] = idf * f * 2.2 / (f + 1.2*(.25+.75*float64(idx.lengths[chunk])/idx.avg))
			}
		}
		idx.lengths = nil
	}
	return idx, nil
}
func (idx *Index) Search(query, candidateQuery string, k, limit int, scorer Scorer) (Result, error) {
	r := Result{Results: []Chunk{}, Files: idx.Files, Chunks: len(idx.Chunks), Engine: "bm25"}
	if strings.TrimSpace(query) == "" || len(query) > 8192 || k < 1 || k > 64 || limit < 1 || limit > k {
		return r, fmt.Errorf("query required (max 8192 bytes), candidates 1..64, limit 1..candidates")
	}
	if candidateQuery == "" {
		candidateQuery = query
	}
	// Resolve strings once, then score only documents containing each term.
	terms := Terms(candidateQuery)
	ids := make([]int, 0, len(terms))
	scores := make([]float64, len(idx.Chunks))
	for _, term := range terms {
		id, ok := idx.vocabulary[term]
		if !ok || slices.Contains(ids, id) {
			continue
		}
		ids = append(ids, id)
		p := &idx.postings[id]
		for j, chunk := range p.chunks {
			scores[chunk] += p.weights[j]
		}
	}
	type hit struct {
		chunk int
		score float64
	}
	hits := make([]hit, 0)
	for i, score := range scores {
		if score > 0 {
			hits = append(hits, hit{i, score})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		a, b := idx.Chunks[hits[i].chunk], idx.Chunks[hits[j].chunk]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Start < b.Start
	})
	ranked := make([]Chunk, min(k, len(hits)))
	for i := range ranked {
		h := hits[i]
		ranked[i] = idx.Chunks[h.chunk]
		ranked[i].Score = h.score
		ranked[i].Lexical = h.score
	}
	r.Candidates = len(ranked)
	if len(ranked) == 0 {
		r.Warnings = append(r.Warnings, "No lexical candidates. Try code identifiers or --candidate-query with English terms; reranking cannot recover absent candidates.")
	}
	for _, ch := range query {
		if unicode.Is(unicode.Hangul, ch) {
			r.Warnings = append(r.Warnings, "English Laya checkpoint: Korean accuracy is unvalidated; supply English candidate terms.")
			break
		}
	}
	if scorer != nil && len(ranked) > 0 {
		r.Engine = "bm25+laya"
		top := ranked[0].Lexical
		for i := range ranked {
			c := &ranked[i]
			p, err := scorer.Predict("File: "+c.Path+"\n"+c.Text, "noul", "Is this source code relevant to the software change: \""+query+"\"?", []string{"false: no, the statement does not hold", "true: yes, the statement holds"})
			if err != nil {
				return r, err
			}
			v := p.Probabilities[1]
			c.Relevance = &v
			c.Truncated = p.Truncated
			c.Score = .5*c.Lexical/top + .5*v
		}
		sortChunks(ranked)
	}
	// Suppress overlapping windows in the final context, preserving candidate coverage.
	for _, c := range ranked {
		duplicate := false
		for _, picked := range r.Results {
			overlap := min(c.End, picked.End) - max(c.Start, picked.Start) + 1
			if c.Path == picked.Path && overlap > min(c.End-c.Start+1, picked.End-picked.Start+1)/2 {
				duplicate = true
				break
			}
		}
		if !duplicate {
			r.Results = append(r.Results, c)
		}
		if len(r.Results) == limit {
			break
		}
	}
	return r, nil
}
func sortChunks(c []Chunk) {
	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Score != c[j].Score {
			return c[i].Score > c[j].Score
		}
		if c[i].Path != c[j].Path {
			return c[i].Path < c[j].Path
		}
		return c[i].Start < c[j].Start
	})
}
