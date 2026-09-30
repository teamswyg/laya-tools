// riido-aligntrain learns small relevance heads over frozen representations.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/alignment"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"os"
	"path/filepath"
	"time"
)

const planSHA = "0ea3776307eafa3012d03a0b969eaa2220247a336e7a4105eac74d2e75b2bd8d"

type trial struct {
	Variant, Encoder string
	Fit              pairlearn.Result
}
type artifact struct {
	Variant, Encoder, Head, File, SHA256 string
	Seed                                 uint64
	Bytes                                int
	Fit                                  pairlearn.Result
	ValidationNLL                        float64
	ValidationAUC                        *float64
}
type model struct {
	Schema, SourceSHA256, PartitionSHA256, EmbeddingRevision, Variant, Encoder, Head string
	Seed                                                                             uint64
	Weights                                                                          []float64
}
type report struct {
	Schema, SourceSHA256, PlanSHA256, PartitionSHA256, EmbeddingRevision       string
	TrainingEligible, TrainingExcluded, ValidationEligible, ValidationExcluded int
	BM25ValidationAUC                                                          *float64
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
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned local pairs")
	plan := flag.String("plan", "experiments/static-alignment/plan-05.json", "fixed development plan")
	dir := flag.String("model-dir", ".cache/potion-base-2M", "pinned static source")
	out := flag.String("out", "", "new local output directory required")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("--out required")
	}
	pb, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	if paireval.Hash(pb) != planSHA {
		return fmt.Errorf("plan hash mismatch")
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
	r := report{Schema: "riido-alignment-training-v1", SourceSHA256: paireval.SourceSHA, PlanSHA256: planSHA, PartitionSHA256: s.MembershipSHA256, EmbeddingRevision: staticembed.Revision}
	lex := paireval.NewLexical(rows, s)
	bp := []paireval.Prediction{}
	for i, row := range rows {
		if s.Rows[i].Split == "validation" && paireval.InScope(row) {
			bp = append(bp, paireval.Prediction{Label: *row.Label, Eligible: true, Score: lex.Score(row.Query, row.Code)})
		}
	}
	r.BM25ValidationAUC = paireval.AUC(bp)
	for _, variant := range []string{"lexical", "normalized_cosine", "normalized_alignment"} {
		encoders := []string{"fp32", "ternary"}
		if variant == "lexical" {
			encoders = []string{"none"}
		}
		for _, encoder := range encoders {
			var m *staticembed.Model
			if encoder != "none" {
				m, e = staticembed.Load(*dir, encoder)
				if e != nil {
					return e
				}
			}
			train, e := alignment.Prepare(rows, s, "development", variant, m)
			if e != nil {
				return e
			}
			val, e := alignment.Prepare(rows, s, "validation", variant, m)
			if e != nil {
				return e
			}
			r.TrainingEligible = len(train.Labels)
			r.TrainingExcluded = train.Excluded
			r.ValidationEligible = len(val.Labels)
			r.ValidationExcluded = val.Excluded
			for _, seed := range []uint64{1729, 2718} {
				for _, head := range []string{"fp32", "ternary_ste"} {
					var best pairlearn.Result
					for _, lr := range []float64{.05, .2} {
						fit, e := pairlearn.Fit(train, val, pairlearn.Config{Seed: seed, Mode: head, LearningRate: lr, L2: .0001, Epochs: 100, Batch: 128, Dimension: alignment.Dimension(variant)})
						if e != nil {
							return e
						}
						r.Trials = append(r.Trials, trial{variant, encoder, fit})
						if best.Weights == nil || fit.ValidationNLL < best.ValidationNLL {
							best = fit
						}
					}
					kinds := []string{head}
					if head == "fp32" {
						kinds = []string{"fp32", "int8", "ternary_ptq"}
					}
					for _, kind := range kinds {
						w := best.Weights
						if kind != head {
							w = hintlearn.Quantize(w, kind)
						}
						file := fmt.Sprintf("%s-%s-%s-%d.align.json", variant, encoder, kind, seed)
						b, e := save(filepath.Join(*out, file), model{alignment.Schema, paireval.SourceSHA, s.MembershipSHA256, staticembed.Revision, variant, encoder, kind, seed, w})
						if e != nil {
							return e
						}
						auc := pairlearn.AUC(val, w)
						if auc == nil || r.BM25ValidationAUC == nil {
							return fmt.Errorf("validation requires both classes")
						}
						nll := pairlearn.NLL(val, w)
						r.Models = append(r.Models, artifact{variant, encoder, kind, file, paireval.Hash(b), seed, len(b), best, nll, auc})
						if variant == "normalized_alignment" && encoder == "fp32" && kind == "fp32" && seed == 1729 {
							r.PrimaryReadyForFinal = *auc >= .60 && *auc-*r.BM25ValidationAUC >= .03
						}
						fmt.Fprintf(os.Stderr, "%s encoder=%s head=%s seed=%d AUC=%.6f NLL=%.6f\n", variant, encoder, kind, seed, *auc, nll)
					}
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
