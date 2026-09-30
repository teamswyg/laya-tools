// eval measures retrieval definition-line hits on a public, pinned corpus.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/search"
	"os"
	"time"
)

func main() {
	root := flag.String("root", "", "public corpus root")
	model := flag.String("model-dir", "", "empty for lexical baseline")
	lib := flag.String("runtime", "", "ORT library")
	queries := flag.String("queries", "benchmarks/httpx-queries.json", "labeled queries")
	k := flag.Int("candidates", 8, "candidate budget")
	flag.Parse()
	if err := run(*root, *model, *lib, *queries, *k); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(root, model, lib, queries string, k int) error {
	b, err := os.ReadFile(queries)
	if err != nil {
		return err
	}
	var qs []struct {
		ID, Lang, Query string
		Gold            []struct {
			Path string
			Line int
		}
	}
	if err = json.Unmarshal(b, &qs); err != nil {
		return err
	}
	start := time.Now()
	idx, err := search.Load(root)
	if err != nil {
		return err
	}
	indexMS := float64(time.Since(start).Microseconds()) / 1000
	var scorer search.Scorer
	loadMS := 0.0
	if model != "" {
		e, err := inference.New(inference.Options{ModelDir: model, Runtime: lib, Threads: 4, MaxTokens: 512, Provider: "cpu"})
		if err != nil {
			return err
		}
		defer e.Close()
		scorer = e
		loadMS = e.LoadMS
		if _, err = e.Predict("warmup", "noul", "Is this warmup?", []string{"false: no, the statement does not hold", "true: yes, the statement holds"}); err != nil {
			return err
		}
	}
	type row struct {
		ID           string  `json:"id"`
		Lang         string  `json:"lang"`
		Rank         int     `json:"rank"`
		CandidateHit bool    `json:"candidate_hit"`
		MS           float64 `json:"ms"`
		Truncated    int     `json:"truncated"`
	}
	rows := []row{}
	for _, q := range qs {
		start := time.Now()
		r, err := idx.Search(q.Query, "", k, k, scorer)
		if err != nil {
			return err
		}
		row := row{ID: q.ID, Lang: q.Lang, MS: float64(time.Since(start).Microseconds()) / 1000}
		for i, c := range r.Results {
			if c.Truncated {
				row.Truncated++
			}
			for _, g := range q.Gold {
				if c.Path == g.Path && c.Start <= g.Line && c.End >= g.Line {
					row.CandidateHit = true
					if row.Rank == 0 {
						row.Rank = i + 1
					}
				}
			}
		}
		rows = append(rows, row)
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"rows": rows, "candidates": k, "files": idx.Files, "chunks": len(idx.Chunks), "source_bytes": idx.Bytes, "index_ms": indexMS, "load_ms": loadMS, "metric": "rank of retrieved chunk containing target function definition; candidate_hit is after overlap suppression, not raw candidate recall"})
}
