package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

type artifact struct {
	Mode       string
	Seed       uint64
	File, Hash string
}
type frozen struct {
	Schema, PlanHash, SelectionHash string
	Split                           paireval.Split
	Models                          []artifact
}

func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func run() error {
	stage := flag.String("stage", "freeze", "freeze or evaluate")
	input := flag.String("input", "", "pinned local source JSON")
	planPath := flag.String("plan", "experiments/semantic-scale/pair-plan.json", "predefined evaluation plan")
	selPath := flag.String("selection", "", "archived synthetic selection JSON")
	frozenPath := flag.String("frozen", "", "frozen manifest for evaluate")
	out := flag.String("out", "", "new output directory")
	flag.Parse()
	if *out == "" || *selPath == "" || (*stage != "freeze" && *stage != "evaluate") {
		return fmt.Errorf("require stage, selection and new out")
	}
	pb, e := os.ReadFile(*planPath)
	if e != nil {
		return e
	}
	var p struct {
		Schema    string         `json:"schema"`
		Source    string         `json:"source_sha256"`
		Seed      string         `json:"seed"`
		Minimum   map[string]int `json:"minimum_rows"`
		Promotion bool           `json:"promotion"`
	}
	if e = json.Unmarshal(pb, &p); e != nil {
		return e
	}
	if p.Schema != "riido-pair-evaluation-plan-v1" || p.Source != paireval.SourceSHA || p.Seed != "riido-pair-groups-01" || p.Promotion || !reflect.DeepEqual(p.Minimum, map[string]int{"final": 2400, "reserve1": 2400, "reserve2": 2400, "validation": 600, "calibration": 600}) {
		return fmt.Errorf("unexpected plan")
	}
	sb, e := os.ReadFile(*selPath)
	if e != nil {
		return e
	}
	var selected struct {
		Schema   string
		Selected []artifact
	}
	if e = json.Unmarshal(sb, &selected); e != nil {
		return e
	}
	if selected.Schema != "riido-hint-selection-v1" || len(selected.Selected) != 8 {
		return fmt.Errorf("invalid archived models")
	}
	weights := [][]float64{}
	seen := map[string]bool{}
	for _, a := range selected.Selected {
		if a.File != fmt.Sprintf("%s-%d.hbin", a.Mode, a.Seed) || (a.Mode != "fp32" && a.Mode != "int8" && a.Mode != "ternary_ptq" && a.Mode != "ternary_ste") || (a.Seed != 1729 && a.Seed != 2718) || seen[a.File] {
			return fmt.Errorf("invalid model entry")
		}
		seen[a.File] = true
		b, e := os.ReadFile(filepath.Join(filepath.Dir(*selPath), a.File))
		if e != nil {
			return e
		}
		if hintlearn.Hash(b) != a.Hash {
			return fmt.Errorf("model hash mismatch")
		}
		w, e := hintlearn.Decode(b)
		if e != nil {
			return e
		}
		weights = append(weights, w)
	}
	rows, e := paireval.Load(*input)
	if e != nil {
		return e
	}
	split := paireval.Partition(rows)
	for name, n := range p.Minimum {
		if split.Counts[name] < n {
			return fmt.Errorf("insufficient %s cases", name)
		}
	}
	if split.Counts["development"] == 0 {
		return fmt.Errorf("no development groups")
	}
	f := frozen{"riido-pair-frozen-v1", paireval.Hash(pb), paireval.Hash(sb), split, selected.Selected}
	if *stage == "freeze" {
		if e = os.Mkdir(*out, 0700); e != nil {
			return e
		}
		return write(filepath.Join(*out, "frozen.json"), f)
	}
	fb, e := os.ReadFile(*frozenPath)
	if e != nil {
		return e
	}
	var check frozen
	if e = json.Unmarshal(fb, &check); e != nil {
		return e
	}
	if !reflect.DeepEqual(check, f) {
		return fmt.Errorf("frozen manifest changed")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	start := time.Now()
	lex := paireval.NewLexical(rows, split)
	names := []string{"bm25", "token_overlap"}
	for _, m := range f.Models {
		names = append(names, fmt.Sprintf("%s-%d", m.Mode, m.Seed))
	}
	cal := make([][]paireval.Prediction, len(names))
	final := make([][]paireval.Prediction, len(names))
	for _, name := range []string{"calibration", "final"} {
		for _, i := range paireval.Indices(split, name) {
			row := rows[i]
			bm, ov := lex.Score(row.Query, row.Code), paireval.Overlap(row.Query, row.Code)
			eligible := paireval.InScope(row)
			var features []hintlearn.Feature
			if eligible {
				features = hintlearn.Features(row.Query, row.Code)
			}
			for model := range names {
				score := bm
				ok := true
				if model == 1 {
					score = ov
				} else if model >= 2 {
					ok = eligible
					if eligible {
						score = hintlearn.Score(weights[model-2], features)
					}
				}
				v := paireval.Prediction{Row: i, Group: split.Rows[i].Group, Label: *row.Label, Score: score, Eligible: ok}
				if name == "calibration" {
					cal[model] = append(cal[model], v)
				} else {
					final[model] = append(final[model], v)
				}
			}
		}
	}
	thresholds := make([]float64, len(names))
	for i := range names {
		thresholds[i] = paireval.Threshold(cal[i])
		for j := range final[i] {
			p := &final[i][j]
			threshold := thresholds[i]
			if !p.Eligible {
				threshold = thresholds[0]
			}
			p.Positive = p.Score >= threshold
		}
	}
	type result struct {
		Name                    string
		Threshold               float64
		Metrics                 paireval.Metrics
		DeltaVsBM25             paireval.Interval
		PositiveDeltaLowerBound bool
		InScopeMetrics          paireval.Metrics
		FallbackMetrics         paireval.Metrics
	}
	results := []result{}
	for i, name := range names {
		delta := paireval.DeltaInterval(final[i], final[0])
		eligible, fallback := []paireval.Prediction{}, []paireval.Prediction{}
		for _, prediction := range final[i] {
			if paireval.InScope(rows[prediction.Row]) {
				eligible = append(eligible, prediction)
			} else {
				fallback = append(fallback, prediction)
			}
		}
		results = append(results, result{name, thresholds[i], paireval.Measure(final[i]), delta, delta.Low > 0, paireval.Measure(eligible), paireval.Measure(fallback)})
	}
	trace, e := json.Marshal(final)
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "case-trace.json"), trace, 0600); e != nil {
		return e
	}
	summary := struct {
		Schema, PlanHash, SelectionHash, FrozenHash, TraceHash string
		Counts, GroupCounts                                    map[string]int
		Seconds                                                float64
		Results                                                []result
		Promotion                                              bool
	}{"riido-pair-results-v1", f.PlanHash, f.SelectionHash, paireval.Hash(fb), paireval.Hash(trace), split.Counts, split.GroupCounts, time.Since(start).Seconds(), results, false}
	return write(filepath.Join(*out, "results.json"), summary)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
