// riido-qat separates ternary checkpoint selection from sealed final checking.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/ternarytrain"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type plan struct {
	Experiment     string    `json:"experiment"`
	DevelopmentSHA string    `json:"development_feature_sha256"`
	FinalSHA       string    `json:"final_families_sha256"`
	Seeds          []int64   `json:"seeds"`
	Thresholds     []float64 `json:"thresholds"`
	Rates          []float64 `json:"learning_rates"`
	Epochs         int       `json:"epochs"`
	Batch          int       `json:"batch_size"`
	Decay          float64   `json:"weight_decay"`
	Temperature    float64   `json:"training_temperature"`
	Temperatures   []float64 `json:"calibration_temperatures"`
	Distillation   float64   `json:"distillation_weight"`
}
type selected struct {
	File        string               `json:"file"`
	SHA         string               `json:"sha256"`
	Seed        int64                `json:"seed"`
	Threshold   float64              `json:"threshold"`
	Rate        float64              `json:"learning_rate"`
	Epoch       int                  `json:"epoch"`
	Temperature float64              `json:"temperature"`
	Bytes       int                  `json:"bytes"`
	SymbolsBits float64              `json:"symbol_bits_per_weight"`
	TotalBits   float64              `json:"artifact_bits_per_weight"`
	Validation  ternarytrain.Metrics `json:"validation"`
	Calibration ternarytrain.Metrics `json:"calibration"`
	PTQ         ternarytrain.Metrics `json:"same_threshold_ptq_validation"`
}
type selection struct {
	Experiment       string               `json:"experiment"`
	PlanSHA          string               `json:"plan_sha256"`
	DevelopmentSHA   string               `json:"development_feature_sha256"`
	FinalFamiliesSHA string               `json:"final_families_sha256"`
	FinalOpened      bool                 `json:"final_opened"`
	Models           []selected           `json:"models"`
	Trials           []ternarytrain.Trial `json:"trials"`
}

func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func train(p plan, planHash string, d ternarytrain.Dataset, hash, out string) error {
	if d.Schema != "riido-tiny-features-v1" || hash != p.DevelopmentSHA {
		return fmt.Errorf("development feature hash mismatch")
	}
	train := ternarytrain.Split(d.Rows, "train")
	val := ternarytrain.Split(d.Rows, "validation")
	cal := ternarytrain.Split(d.Rows, "calibration")
	s := selection{Experiment: p.Experiment, PlanSHA: planHash, DevelopmentSHA: hash, FinalFamiliesSHA: p.FinalSHA}
	for _, seed := range p.Seeds {
		loss := math.Inf(1)
		var best *tinyhead.Model
		var chosen selected
		for _, threshold := range p.Thresholds {
			for _, lr := range p.Rates {
				c := ternarytrain.Config{Seed: seed, Threshold: threshold, LR: lr, Epochs: p.Epochs, Batch: p.Batch, WeightDecay: p.Decay, Temperature: p.Temperature, Distillation: p.Distillation}
				m, tr, e := ternarytrain.Train(d.Source, train, val, c)
				s.Trials = append(s.Trials, tr)
				if e != nil {
					write(filepath.Join(out, "failed-selection.json"), s)
					return e
				}
				fmt.Fprintf(os.Stderr, "seed=%d threshold=%.1f lr=%g epoch=%d validation_nll=%.6f seconds=%.2f\n", seed, threshold, lr, tr.SelectedEpoch, tr.BestNLL, tr.Seconds)
				if m != nil && tr.BestNLL < loss {
					loss = tr.BestNLL
					best = m
					chosen = selected{File: fmt.Sprintf("seed-%d.rdh", seed), Seed: seed, Threshold: threshold, Rate: lr, Epoch: tr.SelectedEpoch}
				}
			}
		}
		if best == nil {
			return fmt.Errorf("no candidate within 600 bytes for seed %d", seed)
		}
		// Temperature selection uses calibration only after choosing weights by validation.
		calLoss := math.Inf(1)
		var calibrated *tinyhead.Model
		for _, temp := range p.Temperatures {
			m, e := best.WithTemperature(temp)
			if e != nil {
				return e
			}
			v, _, e := ternarytrain.Evaluate(m, cal)
			if e != nil {
				return e
			}
			if v.NLL < calLoss {
				calLoss = v.NLL
				calibrated = m
				chosen.Temperature = temp
				chosen.Calibration = v
			}
		}
		if calibrated == nil {
			return fmt.Errorf("empty calibration grid")
		}
		v, _, e := ternarytrain.Evaluate(calibrated, val)
		if e != nil {
			return e
		}
		chosen.Validation = v
		ptq, e := tinyhead.Build(d.Source, tinyhead.Ternary, chosen.Threshold)
		if e != nil {
			return e
		}
		chosen.PTQ, _, e = ternarytrain.Evaluate(ptq, val)
		if e != nil {
			return e
		}
		b := calibrated.Encode()
		chosen.Bytes = len(b)
		chosen.SHA = ternarytrain.Hash(b)
		chosen.TotalBits = 8 * float64(len(b)) / float64(3*calibrated.Width())
		chosen.SymbolsBits = 8 * float64(len(b)-44) / float64(3*calibrated.Width())
		if e = os.WriteFile(filepath.Join(out, chosen.File), b, 0600); e != nil {
			return e
		}
		s.Models = append(s.Models, chosen)
	}
	if e := write(filepath.Join(out, "selection.json"), s); e != nil {
		return e
	}
	fmt.Println("Both seeds selected and sealed; no final dataset was loaded. Run --stage check separately.")
	return nil
}

var sink [3]float64

type checked struct {
	Model          selected                  `json:"model"`
	Metrics        ternarytrain.Metrics      `json:"metrics"`
	Pass           bool                      `json:"gate_pass"`
	Predictions    []ternarytrain.Prediction `json:"predictions"`
	WarmNS         float64                   `json:"warm_head_ns_per_op"`
	Allocations    float64                   `json:"allocations_per_op"`
	MaxReloadError float64                   `json:"max_reload_probability_error"`
}

func check(p plan, planHash string, dev ternarytrain.Dataset, devHash, finalPath, selectionPath, out string) error {
	b, e := os.ReadFile(selectionPath)
	if e != nil {
		return e
	}
	var s selection
	if e = json.Unmarshal(b, &s); e != nil {
		return e
	}
	if s.Experiment != p.Experiment || s.PlanSHA != planHash || s.DevelopmentSHA != devHash || devHash != p.DevelopmentSHA || s.FinalFamiliesSHA != p.FinalSHA || s.FinalOpened || len(s.Models) != 2 {
		return fmt.Errorf("selection seal mismatch")
	}
	// Load final only after selection manifest and every checkpoint hash are verified.
	models := make([]*tinyhead.Model, 2)
	for i, m := range s.Models {
		if m.Seed != p.Seeds[i] || m.File != fmt.Sprintf("seed-%d.rdh", m.Seed) {
			return fmt.Errorf("unexpected seed/model path")
		}
		b, e := os.ReadFile(filepath.Join(filepath.Dir(selectionPath), m.File))
		if e != nil {
			return e
		}
		if ternarytrain.Hash(b) != m.SHA || len(b) != m.Bytes {
			return fmt.Errorf("checkpoint seal mismatch")
		}
		models[i], e = tinyhead.Decode(b)
		if e != nil {
			return e
		}
	}
	final, finalHash, e := ternarytrain.Load(finalPath)
	if e != nil {
		return e
	}
	if final.Schema != "riido-qat-final-v1" {
		return fmt.Errorf("wrong final schema")
	}
	var provenance struct {
		Hash string `json:"final_families_sha256"`
	}
	if e = json.Unmarshal(final.Provenance, &provenance); e != nil {
		return e
	}
	if provenance.Hash != p.FinalSHA {
		return fmt.Errorf("wrong final provenance")
	}
	parent, e := tinyhead.Build(dev.Source, tinyhead.Float32, 0)
	if e != nil {
		return e
	}
	parentMetrics, parentPred, e := ternarytrain.Evaluate(parent, final.Rows)
	if e != nil {
		return e
	}
	ptq, e := tinyhead.Build(dev.Source, tinyhead.Ternary, .5)
	if e != nil {
		return e
	}
	ptqMetrics, _, e := ternarytrain.Evaluate(ptq, final.Rows)
	if e != nil {
		return e
	}
	allPass := true
	results := []checked{}
	for i, m := range models {
		v, pred, e := ternarytrain.Evaluate(m, final.Rows)
		if e != nil {
			return e
		}
		c := checked{Model: s.Models[i], Metrics: v, Pass: ternarytrain.Pass(v, parentMetrics, m.Bytes()), Predictions: pred}
		allPass = allPass && c.Pass
		reload, e := tinyhead.Decode(m.Encode())
		if e != nil {
			return e
		}
		_, again, e := ternarytrain.Evaluate(reload, final.Rows)
		if e != nil {
			return e
		}
		for j := range pred {
			for k := 0; k < 3; k++ {
				c.MaxReloadError = math.Max(c.MaxReloadError, math.Abs(pred[j].Probability[k]-again[j].Probability[k]))
			}
		}
		if c.MaxReloadError != 0 {
			return fmt.Errorf("binary reload parity failure")
		}
		scratch := make([]float64, m.Width())
		for j := 0; j < 100; j++ {
			sink, _ = m.Predict(final.Rows[j%len(final.Rows)].Feature, scratch)
		}
		c.Allocations = testing.AllocsPerRun(100, func() { sink, _ = m.Predict(final.Rows[0].Feature, scratch) })
		start := time.Now()
		for j := 0; j < 100000; j++ {
			sink, _ = m.Predict(final.Rows[j%len(final.Rows)].Feature, scratch)
		}
		c.WarmNS = float64(time.Since(start).Nanoseconds()) / 100000
		results = append(results, c)
	}
	report := struct {
		Experiment   string                    `json:"experiment"`
		PlanSHA      string                    `json:"plan_sha256"`
		SelectionSHA string                    `json:"selection_sha256"`
		FinalSHA     string                    `json:"final_features_sha256"`
		BothPass     bool                      `json:"both_seeds_pass"`
		Parent       ternarytrain.Metrics      `json:"parent_fp32"`
		ParentPred   []ternarytrain.Prediction `json:"parent_predictions"`
		PTQ          ternarytrain.Metrics      `json:"previous_ptq_660b"`
		Results      []checked                 `json:"results"`
		Scope        string                    `json:"scope"`
	}{p.Experiment, planHash, ternarytrain.Hash(b), finalHash, allPass, parentMetrics, parentPred, ptqMetrics, results, "36 newly authored English scope-labelled synthetic cases; not observed coding outcomes; head only, encoder remains FP32"}
	if e = write(filepath.Join(out, "results.json"), report); e != nil {
		return e
	}
	fmt.Printf("final cases=%d parent=%d/%d both_seeds_pass=%v\n", len(final.Rows), parentMetrics.Correct, parentMetrics.Cases, allPass)
	for _, c := range results {
		fmt.Printf("seed=%d bytes=%d bits/weight=%.4f correct=%d/%d accepted=%d/%d strong_to_fast=%d gate=%v\n", c.Model.Seed, c.Model.Bytes, c.Model.TotalBits, c.Metrics.Correct, c.Metrics.Cases, c.Metrics.AcceptedCorrect, c.Metrics.Accepted, c.Metrics.StrongToFast, c.Pass)
	}
	if !allPass {
		return fmt.Errorf("final quality gate failed; results preserved")
	}
	return nil
}
func run() error {
	stage := flag.String("stage", "train", "train, check or bench; train never loads final")
	planPath := flag.String("plan", "experiments/ternary-qat/plan.json", "frozen plan")
	devPath := flag.String("development", "", "public development features")
	finalPath := flag.String("final", "", "final features, check stage only")
	selectionPath := flag.String("selection", "", "sealed selection.json, check stage only")
	out := flag.String("out", "", "new output directory")
	flag.Parse()
	if *stage != "train" && *stage != "check" && *stage != "bench" {
		return fmt.Errorf("unknown stage")
	}
	if *stage == "train" && (*finalPath != "" || *selectionPath != "") {
		return fmt.Errorf("train rejects final/selection arguments")
	}
	b, e := os.ReadFile(*planPath)
	if e != nil {
		return e
	}
	var p plan
	if e = json.Unmarshal(b, &p); e != nil {
		return e
	}
	if (p.Experiment != "ternary-qat-01" && p.Experiment != "ternary-qat-02") || len(p.Seeds) != 2 || p.Seeds[0] == p.Seeds[1] || len(p.Thresholds) == 0 || len(p.Rates) == 0 || len(p.Temperatures) == 0 {
		return fmt.Errorf("unsupported or incomplete plan")
	}
	d, hash, e := ternarytrain.Load(*devPath)
	if e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if *stage == "train" {
		return train(p, ternarytrain.Hash(b), d, hash, *out)
	}
	if *stage == "bench" {
		return bench(d, *finalPath, *selectionPath, *out)
	}
	return check(p, ternarytrain.Hash(b), d, hash, *finalPath, *selectionPath, *out)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
