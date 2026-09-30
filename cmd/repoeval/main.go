// repoeval uses only caller-supplied fixtures and never accesses GitHub or paid models.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/teamswyg/laya-tools/internal/app"
	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/repojudge"
	"github.com/teamswyg/laya-tools/pkg/reporouter"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func read(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
func run() error {
	catalog := flag.String("catalog", "examples/repositories/catalog.json", "original or public fixture catalog")
	queries := flag.String("queries", "benchmarks/repository-queries.json", "original or public fixtures")
	laya := flag.Bool("laya", false, "enable native base checkpoint")
	flag.Parse()
	var repos []reporouter.Repository
	var cases []struct {
		ID       string `json:"id"`
		Language string `json:"language"`
		Query    string `json:"query"`
		Expected string `json:"expected"`
	}
	if err := read(*catalog, &repos); err != nil {
		return err
	}
	if err := read(*queries, &cases); err != nil {
		return err
	}
	if len(cases) == 0 {
		return fmt.Errorf("empty evaluation")
	}
	start := time.Now()
	idx, err := reporouter.New(repos)
	if err != nil {
		return err
	}
	indexMS := float64(time.Since(start).Microseconds()) / 1000
	a := &app.App{Options: inference.Options{ModelDir: assets.ModelDir("base"), Runtime: assets.Runtime(), Provider: "cpu", Threads: 4, MaxTokens: 512}}
	defer a.Close()
	var judge reporouter.Judge
	loadMS := 0.0
	if *laya {
		e, err := a.Engine()
		if err != nil {
			return err
		}
		loadMS = e.LoadMS
		judge = repojudge.Judge{Engine: a.Engine}
	}
	type row struct {
		ID       string            `json:"id"`
		Language string            `json:"language"`
		Expected string            `json:"expected"`
		Result   reporouter.Result `json:"result"`
		MS       float64           `json:"ms"`
	}
	rows := []row{}
	times := []float64{}
	positives, recall, top1, accepted, correct, wrong, negativeSuggestions, negativeCases := 0, 0, 0, 0, 0, 0, 0, 0
	rawDecisions, rawCorrect := 0, 0
	for _, c := range cases {
		start := time.Now()
		r, err := idx.Preview(c.Query, reporouter.DefaultConfig(), judge)
		if err != nil {
			return err
		}
		ms := float64(time.Since(start).Microseconds()) / 1000
		rows = append(rows, row{c.ID, c.Language, c.Expected, r, ms})
		times = append(times, ms)
		if c.Expected != "" {
			positives++
			if len(r.Candidates) > 0 && r.Candidates[0].Name == c.Expected {
				top1++
			}
			for _, v := range r.Candidates {
				if v.Name == c.Expected {
					recall++
					break
				}
			}
		} else {
			negativeCases++
			if r.Suggested != "" {
				negativeSuggestions++
			}
		}
		if r.Status == "suggest" {
			accepted++
			if r.Suggested == c.Expected {
				correct++
			} else {
				wrong++
			}
		}
		if r.Judgment != nil {
			rawDecisions++
			winner := 0
			for i, v := range r.Judgment.Probabilities {
				if v > r.Judgment.Probabilities[winner] {
					winner = i
				}
			}
			name := ""
			if winner < len(r.Candidates) {
				name = r.Candidates[winner].Name
			}
			if name == c.Expected {
				rawCorrect++
			}
		}
	}
	sort.Float64s(times)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"fixture": "original synthetic development set; not Riido production traffic or held-out accuracy", "laya": *laya, "repositories": len(repos), "cases": len(cases), "positive_cases": positives, "negative_cases": negativeCases, "shortlist_recall_count": recall, "lexical_top1_correct": top1, "model_judgments": rawDecisions, "raw_model_correct": rawCorrect, "accepted": accepted, "accepted_correct": correct, "accepted_wrong": wrong, "negative_suggestions": negativeSuggestions, "index_ms": indexMS, "load_ms": loadMS, "request_median_ms": times[len(times)/2], "request_p95_ms": times[(len(times)*95-1)/100], "go_heap_bytes": mem.HeapAlloc, "memory_note": "Go heap excludes native/GPU allocations; collect peak RSS separately. Model load is excluded from request timings; first inference is included.", "rows": rows})
}
