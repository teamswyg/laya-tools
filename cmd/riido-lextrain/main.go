// riido-lextrain runs a fixed development-only lexical-control experiment.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"os"
	"path/filepath"
	"time"
)

const planSHA = "877eb95845612f5dc5679f79b2de729754e15b812f84457f5a01cc3a6f146dee"

type model struct {
	Schema, SourceSHA256, MembershipSHA256, Variant, Mode string
	Seed                                                  uint64
	Weights                                               []float64
}
type trial struct {
	Variant string
	Fit     pairlearn.Result
}
type artifact struct {
	Variant, Mode, File, SHA256 string
	Seed                        uint64
	Bytes                       int
	Fit                         pairlearn.Result
	ValidationNLL               float64
	ValidationAUC               *float64
}
type report struct {
	Schema, SourceSHA256, PlanSHA256, MembershipSHA256                         string
	TrainingEligible, TrainingExcluded, ValidationEligible, ValidationExcluded int
	BM25ValidationAUC, OverlapValidationAUC                                    *float64
	Trials                                                                     []trial
	Models                                                                     []artifact
	PrimaryReadyForFinal, Promotion                                            bool
	Seconds                                                                    float64
}

func save(path string, v any) ([]byte, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return nil, e
	}
	b = append(b, '\n')
	return b, os.WriteFile(path, b, 0600)
}
func prepare(rows []paireval.Row, s paireval.Split, name string, length bool) (pairlearn.Dataset, error) {
	d := pairlearn.Dataset{Split: name, Offsets: []int{0}}
	if len(rows) != len(s.Rows) {
		return d, fmt.Errorf("membership length mismatch")
	}
	if name != "development" && name != "validation" {
		return d, fmt.Errorf("holdout preparation rejected")
	}
	for i, r := range rows {
		if s.Rows[i].Split != name {
			continue
		}
		if r.Label == nil || (*r.Label != 0 && *r.Label != 1) {
			return d, fmt.Errorf("invalid label")
		}
		if !paireval.InScope(r) {
			d.Excluded++
			continue
		}
		fs := lexicalhint.Features(r.Query, r.Code, length)
		for j, v := range fs {
			if v != 0 {
				d.Indices = append(d.Indices, uint16(j))
				d.Values = append(d.Values, v)
			}
		}
		d.Offsets = append(d.Offsets, len(d.Indices))
		d.Labels = append(d.Labels, float64(*r.Label))
		d.Groups = append(d.Groups, s.Rows[i].Group)
	}
	return d, nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned local source")
	plan := flag.String("plan", "experiments/semantic-scale/lexical-plan-03.json", "exact fixed plan")
	out := flag.String("out", "", "new output directory, required")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("--out required")
	}
	pb, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	if paireval.Hash(pb) != planSHA {
		return fmt.Errorf("plan mismatch")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	start := time.Now()
	rows, e := paireval.Load(*input)
	if e != nil {
		return e
	}
	s := paireval.Partition(rows)
	if s.Counts["development"] != 12189 || s.Counts["validation"] != 607 || s.Counts["reserve1"] != 2403 {
		return fmt.Errorf("partition mismatch")
	}
	if _, e = save(filepath.Join(*out, "split.json"), s); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "plan.json"), pb, 0600); e != nil {
		return e
	}
	r := report{Schema: "riido-lexical-training-v1", SourceSHA256: paireval.SourceSHA, PlanSHA256: planSHA, MembershipSHA256: s.MembershipSHA256}
	lex := paireval.NewLexical(rows, s)
	bp, op := []paireval.Prediction{}, []paireval.Prediction{}
	for i, v := range rows {
		if s.Rows[i].Split != "validation" || !paireval.InScope(v) {
			continue
		}
		p := paireval.Prediction{Label: *v.Label, Eligible: true, Score: lex.Score(v.Query, v.Code)}
		bp = append(bp, p)
		p.Score = paireval.Overlap(v.Query, v.Code)
		op = append(op, p)
	}
	r.BM25ValidationAUC = paireval.AUC(bp)
	r.OverlapValidationAUC = paireval.AUC(op)
	for _, variant := range []string{"matching", "matching_length"} {
		train, e := prepare(rows, s, "development", variant == "matching_length")
		if e != nil {
			return e
		}
		val, e := prepare(rows, s, "validation", variant == "matching_length")
		if e != nil {
			return e
		}
		r.TrainingEligible = len(train.Labels)
		r.TrainingExcluded = train.Excluded
		r.ValidationEligible = len(val.Labels)
		r.ValidationExcluded = val.Excluded
		for _, seed := range []uint64{1729, 2718} {
			for _, mode := range []string{"fp32", "ternary_ste"} {
				var best pairlearn.Result
				for _, lr := range []float64{.05, .2} {
					fit, e := pairlearn.Fit(train, val, pairlearn.Config{Seed: seed, Mode: mode, LearningRate: lr, L2: .0001, Epochs: 100, Batch: 128, Dimension: lexicalhint.Dimension})
					if e != nil {
						return e
					}
					r.Trials = append(r.Trials, trial{variant, fit})
					if best.Weights == nil || fit.ValidationNLL < best.ValidationNLL {
						best = fit
					}
				}
				kinds := []string{mode}
				if mode == "fp32" {
					kinds = []string{"fp32", "int8", "ternary_ptq"}
				}
				for _, kind := range kinds {
					w := best.Weights
					if kind != mode {
						w = hintlearn.Quantize(w, kind)
					}
					file := fmt.Sprintf("%s-%s-%d.lexmodel.json", variant, kind, seed)
					b, e := save(filepath.Join(*out, file), model{lexicalhint.Schema, paireval.SourceSHA, s.MembershipSHA256, variant, kind, seed, w})
					if e != nil {
						return e
					}
					auc := pairlearn.AUC(val, w)
					r.Models = append(r.Models, artifact{variant, kind, file, paireval.Hash(b), seed, len(b), best, pairlearn.NLL(val, w), auc})
					if variant == "matching" && kind == "fp32" && seed == 1729 && auc != nil && r.BM25ValidationAUC != nil {
						r.PrimaryReadyForFinal = *auc >= .6 && *auc-*r.BM25ValidationAUC >= .03
					}
					if auc == nil {
						return fmt.Errorf("validation requires both classes")
					}
					fmt.Fprintf(os.Stderr, "%s %s seed%d validation_auc=%.6f nll=%.6f\n", variant, kind, seed, *auc, pairlearn.NLL(val, w))
				}
			}
		}
	}
	r.Seconds = time.Since(start).Seconds()
	if _, e = save(filepath.Join(*out, "selection.json"), r); e != nil {
		return e
	}
	r.Trials = nil
	r.Models = nil
	return json.NewEncoder(os.Stdout).Encode(r)
}
