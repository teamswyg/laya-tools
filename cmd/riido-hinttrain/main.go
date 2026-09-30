package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type plan struct {
	Schema    string            `json:"schema"`
	Dimension int               `json:"dimension"`
	Epochs    int               `json:"epochs"`
	Seeds     []uint64          `json:"seeds"`
	Rates     []float64         `json:"learning_rates"`
	Modes     []string          `json:"modes"`
	Threshold float64           `json:"threshold"`
	Files     map[string]string `json:"files"`
}
type selected struct {
	Mode       string
	Seed       uint64
	File, Hash string
	Trial      hintlearn.Trial
}
type selection struct {
	Schema, PlanHash string
	Seconds          float64
	Trials           []hintlearn.Trial
	Selected         []selected
}

func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func run() error {
	stage := flag.String("stage", "train", "train or check")
	root := flag.String("data", "experiments/semantic-learning", "frozen fixture directory")
	out := flag.String("out", "", "new local output directory")
	selPath := flag.String("selection", "", "sealed selection.json for check")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("out required")
	}
	pb, e := os.ReadFile(filepath.Join(*root, "plan.json"))
	if e != nil {
		return e
	}
	var p plan
	if e = json.Unmarshal(pb, &p); e != nil {
		return e
	}
	if p.Schema != "riido-hint-learning-plan-v1" || p.Dimension != hintlearn.Dimension || p.Threshold != .7 || p.Epochs != 60 || !slices.Equal(p.Seeds, []uint64{1729, 2718}) || !slices.Equal(p.Rates, []float64{.05, .2}) || !slices.Equal(p.Modes, []string{"fp32", "ternary_ste"}) {
		return fmt.Errorf("unexpected frozen plan")
	}
	load := func(split string) (hintlearn.Dataset, error) {
		d, h, e := hintlearn.Load(filepath.Join(*root, split+".json"), split)
		if e == nil && h != p.Files[split+".json"] {
			e = fmt.Errorf("dataset hash mismatch")
		}
		return d, e
	}
	train, e := load("train")
	if e != nil {
		return e
	}
	val, e := load("validation")
	if e != nil {
		return e
	}
	if e = hintlearn.Disjoint(train, val); e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if *stage == "train" {
		if *selPath != "" {
			return fmt.Errorf("train does not accept selection")
		}
		start := time.Now()
		s := selection{Schema: "riido-hint-selection-v1", PlanHash: hintlearn.Hash(pb)}
		for _, seed := range p.Seeds {
			for _, mode := range p.Modes {
				best := hintlearn.Trial{}
				best.Validation.MeanChecks = math.Inf(1)
				for _, lr := range p.Rates {
					t, e := hintlearn.Fit(train, val, hintlearn.Config{Seed: seed, LearningRate: lr, Mode: mode, Epochs: p.Epochs})
					if e != nil {
						return e
					}
					s.Trials = append(s.Trials, t)
					if t.Validation.MeanChecks < best.Validation.MeanChecks || (t.Validation.MeanChecks == best.Validation.MeanChecks && t.Validation.NLL < best.Validation.NLL) {
						best = t
					}
				}
				modes := []string{mode}
				if mode == "fp32" {
					modes = append(modes, "int8", "ternary_ptq")
				}
				for _, exportMode := range modes {
					w := best.Weights
					if exportMode != mode {
						w = hintlearn.Quantize(w, exportMode)
					}
					b, e := hintlearn.Encode(w, exportMode)
					if e != nil {
						return e
					}
					file := fmt.Sprintf("%s-%d.hbin", exportMode, seed)
					if e = os.WriteFile(filepath.Join(*out, file), b, 0600); e != nil {
						return e
					}
					decoded, e := hintlearn.Decode(b)
					if e != nil {
						return e
					}
					if !slices.Equal(decoded, w) {
						return fmt.Errorf("export changed coefficients")
					}
					s.Selected = append(s.Selected, selected{exportMode, seed, file, hintlearn.Hash(b), best})
				}
			}
		}
		s.Seconds = time.Since(start).Seconds()
		return write(filepath.Join(*out, "selection.json"), s)
	}
	if *stage != "check" || *selPath == "" {
		return fmt.Errorf("require stage check and selection")
	}
	sb, e := os.ReadFile(*selPath)
	if e != nil {
		return e
	}
	var s selection
	if e = json.Unmarshal(sb, &s); e != nil {
		return e
	}
	if s.Schema != "riido-hint-selection-v1" || s.PlanHash != hintlearn.Hash(pb) || len(s.Selected) != 8 {
		return fmt.Errorf("invalid sealed selection")
	}
	weights := [][]float64{}
	sizes := []int{}
	for _, v := range s.Selected {
		if filepath.Base(v.File) != v.File {
			return fmt.Errorf("invalid model filename")
		}
		b, e := os.ReadFile(filepath.Join(filepath.Dir(*selPath), v.File))
		if e != nil {
			return e
		}
		if hintlearn.Hash(b) != v.Hash {
			return fmt.Errorf("model hash mismatch")
		}
		w, e := hintlearn.Decode(b)
		if e != nil {
			return e
		}
		weights = append(weights, w)
		sizes = append(sizes, len(b))
	}
	// Final query content is read only after the entire selected bundle validates.
	final, e := load("final")
	if e != nil {
		return e
	}
	if e = hintlearn.Disjoint(train, final); e != nil {
		return e
	}
	if e = hintlearn.Disjoint(val, final); e != nil {
		return e
	}
	texts := []string{}
	for _, d := range final.Documents {
		texts = append(texts, d.Text)
	}
	idx, e := hintsearch.New(texts)
	if e != nil {
		return e
	}
	baseRanks := []int{}
	base := 0.
	for _, q := range final.Queries {
		r, e := idx.Rank(q.Text)
		if e != nil {
			return e
		}
		for rank, d := range r.Order {
			if final.Documents[d].ID == q.Target {
				baseRanks = append(baseRanks, rank+1)
				base += float64(rank + 1)
				break
			}
		}
	}
	base /= float64(len(final.Queries))
	type result struct {
		Mode              string
		Seed              uint64
		Bytes             int
		Hash              string
		Metrics           hintlearn.Metrics
		ChecksRatio       float64
		Pass              bool
		BoundedMeanChecks float64
		BoundedRanks      []int
	}
	rows := []result{}
	for i, v := range s.Selected {
		m := hintlearn.Evaluate(weights[i], final)
		bounded := 0.
		boundedRanks := []int{}
		for _, q := range final.Queries {
			b, e := idx.Rank(q.Text)
			if e != nil {
				return e
			}
			h, _, e := hintlearn.Rank(weights[i], q.Text, texts)
			if e != nil {
				return e
			}
			order, e := hintsearch.Interleave(b.Order, h)
			if e != nil {
				return e
			}
			for rank, d := range order {
				if final.Documents[d].ID == q.Target {
					bounded += float64(rank + 1)
					boundedRanks = append(boundedRanks, rank+1)
					break
				}
			}
		}
		rows = append(rows, result{v.Mode, v.Seed, sizes[i], v.Hash, m, m.MeanChecks / base, m.Recall4 >= .95 && m.MeanChecks <= .8*base, bounded / float64(len(final.Queries)), boundedRanks})
	}
	return write(filepath.Join(*out, "results.json"), map[string]any{
		"Schema": "riido-hint-final-v1", "PlanHash": hintlearn.Hash(pb), "SelectionHash": hintlearn.Hash(sb), "BM25MeanChecks": base, "BM25Ranks": baseRanks, "Results": rows,
		"DiagnosticRuleRanks": hintlearn.RuleRanks(final), "DiagnosticRuleScope": "post-evaluation nonlearned control with explicit retain->keeps/discard->removes substitutions; not used for training or selection",
	})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
