// riido-pairtrain trains external binary relevance probes without reading
// holdout features or using holdout labels for weight/epoch selection.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
)

type plan struct {
	Schema, SourceSHA256, Training, Final, Preserve string
	Seeds                                           []uint64
	LearningRates                                   []float64
	Modes                                           []string
	L2                                              float64
	Epochs, Batch                                   int
	Promotion                                       bool
}
type artifact struct {
	Mode                string
	Seed                uint64
	File, SHA256        string
	Bytes               int
	Fit                 pairlearn.Result
	ExportValidationNLL float64
	ExportValidationAUC *float64
}
type manifest struct {
	TrainingPositiveRate, PriorValidationNLL                                                        float64
	Schema, SourceSHA256, PlanSHA256, MembershipSHA256                                              string
	Counts, GroupCounts                                                                             map[string]int
	TrainingEligible, TrainingExcluded, ValidationEligible, ValidationExcluded, FeaturePayloadBytes int
	Trials                                                                                          []pairlearn.Result
	Models                                                                                          []artifact
	Seconds                                                                                         float64
	Promotion                                                                                       bool
}

func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned local source")
	planPath := flag.String("plan", "experiments/semantic-scale/train-plan-02.json", "frozen training plan")
	out := flag.String("out", "", "new local output directory (required)")
	cpuProfile := flag.String("cpu-profile", "", "optional local CPU profile")
	heapProfile := flag.String("heap-profile", "", "optional local Go heap profile")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("--out required")
	}
	pb, e := os.ReadFile(*planPath)
	if e != nil {
		return e
	}
	var p plan
	if e = json.Unmarshal(pb, &p); e != nil {
		return e
	}
	if p.Schema != "riido-pair-training-plan-v1" || p.SourceSHA256 != paireval.SourceSHA || p.Training != "development" || p.Final != "reserve1" || p.Preserve != "reserve2" || p.Promotion || !reflect.DeepEqual(p.Seeds, []uint64{1729, 2718}) || !reflect.DeepEqual(p.LearningRates, []float64{.05, .2}) || !reflect.DeepEqual(p.Modes, []string{"fp32", "ternary_ste"}) || p.L2 != .0001 || p.Epochs != 20 || p.Batch != 128 {
		return fmt.Errorf("unsupported plan; revise experiment contract explicitly")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if *cpuProfile != "" {
		f, e := os.Create(*cpuProfile)
		if e != nil {
			return e
		}
		defer f.Close()
		if e = pprof.StartCPUProfile(f); e != nil {
			return e
		}
		defer pprof.StopCPUProfile()
	}
	start := time.Now()
	rows, e := paireval.Load(*input)
	if e != nil {
		return e
	}
	split := paireval.Partition(rows)
	if split.Counts["development"] != 12189 || split.Counts["validation"] != 607 || split.Counts["reserve1"] != 2403 {
		return fmt.Errorf("unexpected frozen partition")
	}
	// Persist the source/config/split commitment before any training or scoring.
	if e = write(filepath.Join(*out, "split.json"), split); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "plan.json"), pb, 0600); e != nil {
		return e
	}
	train, e := pairlearn.Prepare(rows, split, "development")
	if e != nil {
		return e
	}
	validation, e := pairlearn.Prepare(rows, split, "validation")
	if e != nil {
		return e
	}
	m := manifest{Schema: "riido-pair-trained-v1", SourceSHA256: paireval.SourceSHA, PlanSHA256: paireval.Hash(pb), MembershipSHA256: split.MembershipSHA256, Counts: split.Counts, GroupCounts: split.GroupCounts, TrainingEligible: len(train.Labels), TrainingExcluded: train.Excluded, ValidationEligible: len(validation.Labels), ValidationExcluded: validation.Excluded, FeaturePayloadBytes: (len(train.Values) + len(validation.Values)) * 10}
	for _, seed := range p.Seeds {
		for _, mode := range p.Modes {
			var best pairlearn.Result
			for _, lr := range p.LearningRates {
				r, e := pairlearn.Fit(train, validation, pairlearn.Config{Seed: seed, Mode: mode, LearningRate: lr, L2: p.L2, Epochs: p.Epochs, Batch: p.Batch})
				if e != nil {
					return e
				}
				m.Trials = append(m.Trials, r)
				if best.Weights == nil || r.ValidationNLL < best.ValidationNLL {
					best = r
				}
				fmt.Fprintf(os.Stderr, "seed=%d mode=%s lr=%g epoch=%d validation_nll=%.6f\n", seed, mode, lr, r.Epoch, r.ValidationNLL)
			}
			exports := []string{mode}
			if mode == "fp32" {
				exports = []string{"fp32", "int8", "ternary_ptq"}
			}
			for _, kind := range exports {
				weights := best.Weights
				if kind != mode {
					weights = hintlearn.Quantize(weights, kind)
				}
				b, e := hintlearn.Encode(weights, kind)
				if e != nil {
					return e
				}
				decoded, e := hintlearn.Decode(b)
				if e != nil {
					return e
				}
				file := fmt.Sprintf("%s-%d.hbin", kind, seed)
				if e = os.WriteFile(filepath.Join(*out, file), b, 0600); e != nil {
					return e
				}
				m.Models = append(m.Models, artifact{kind, seed, file, hintlearn.Hash(b), len(b), best, pairlearn.NLL(validation, decoded), pairlearn.AUC(validation, decoded)})
			}
		}
	}
	prior := 0.
	for _, y := range train.Labels {
		prior += y
	}
	prior /= float64(len(train.Labels))
	priorLoss := 0.
	for _, y := range validation.Labels {
		priorLoss -= y*math.Log(prior) + (1-y)*math.Log1p(-prior)
	}
	m.TrainingPositiveRate = prior
	m.PriorValidationNLL = priorLoss / float64(len(validation.Labels))
	if *heapProfile != "" {
		runtime.GC()
		f, e := os.Create(*heapProfile)
		if e != nil {
			return e
		}
		e = pprof.WriteHeapProfile(f)
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		runtime.KeepAlive(train)
		runtime.KeepAlive(validation)
	}
	m.Seconds = time.Since(start).Seconds()
	if e = write(filepath.Join(*out, "selection.json"), m); e != nil {
		return e
	}
	summary := m
	summary.Trials = nil
	summary.Models = nil
	return json.NewEncoder(os.Stdout).Encode(summary)
}
