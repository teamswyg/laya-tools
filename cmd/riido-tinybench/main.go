// riido-tinybench measures experimental head compression, not full-model latency.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/internal/tinyhead"
)

type row struct {
	ID        string     `json:"id"`
	Split     string     `json:"split"`
	Label     int        `json:"label"`
	Feature   []float64  `json:"feature"`
	Reference [3]float64 `json:"reference"`
}
type input struct {
	Schema     string          `json:"schema"`
	Origin     string          `json:"origin"`
	Source     tinyhead.Source `json:"source"`
	Rows       []row           `json:"rows"`
	Provenance json.RawMessage `json:"provenance"`
}
type metrics struct {
	Cases             int     `json:"cases"`
	Correct           int     `json:"correct"`
	Accepted          int     `json:"accepted"`
	AcceptedCorrect   int     `json:"accepted_correct"`
	StrongToFast      int     `json:"strong_to_fast"`
	Accuracy          float64 `json:"accuracy"`
	AcceptedPrecision float64 `json:"accepted_precision"`
	NLL               float64 `json:"nll"`
	Brier             float64 `json:"brier"`
	Flips             int     `json:"reference_winner_flips"`
	MaxError          float64 `json:"max_probability_error"`
}
type result struct {
	Name           string             `json:"name"`
	Threshold      float64            `json:"threshold"`
	Bytes          int                `json:"artifact_bytes"`
	SHA256         string             `json:"sha256"`
	ZeroFraction   float64            `json:"zero_fraction"`
	FiveTritBytes  int                `json:"five_trit_symbol_bytes"`
	TwoBitBytes    int                `json:"two_bit_symbol_bytes"`
	Splits         map[string]metrics `json:"splits"`
	Nanoseconds    float64            `json:"warm_head_ns_per_op"`
	Allocations    float64            `json:"allocations_per_op"`
	RegressionPass bool               `json:"compression_gate_pass"`
}

func winner(p [3]float64) int {
	k := 0
	for i := 1; i < 3; i++ {
		if p[i] > p[k] {
			k = i
		}
	}
	return k
}
func score(m *tinyhead.Model, rows []row, split string) metrics {
	var out metrics
	scratch := make([]float64, m.Width())
	for _, r := range rows {
		if r.Split != split {
			continue
		}
		p, e := m.Predict(r.Feature, scratch)
		if e != nil {
			panic(e)
		}
		k := winner(p)
		out.Cases++
		if k == r.Label {
			out.Correct++
		}
		if p[k] >= .9 {
			out.Accepted++
			if k == r.Label {
				out.AcceptedCorrect++
			}
		}
		if r.Label == 2 && k == 0 {
			out.StrongToFast++
		}
		if k != winner(r.Reference) {
			out.Flips++
		}
		out.NLL -= math.Log(math.Max(p[r.Label], 1e-300))
		for i := range p {
			out.MaxError = math.Max(out.MaxError, math.Abs(p[i]-r.Reference[i]))
			target := 0.
			if i == r.Label {
				target = 1
			}
			d := p[i] - target
			out.Brier += d * d
		}
	}
	if out.Cases > 0 {
		out.Accuracy = float64(out.Correct) / float64(out.Cases)
		out.NLL /= float64(out.Cases)
		out.Brier /= float64(out.Cases)
	}
	if out.Accepted > 0 {
		out.AcceptedPrecision = float64(out.AcceptedCorrect) / float64(out.Accepted)
	}
	return out
}
func check(in input) error {
	if in.Schema != "riido-tiny-features-v1" || in.Origin != "original_synthetic_pdca06_development" || len(in.Rows) == 0 {
		return fmt.Errorf("unsupported public fixture schema/origin")
	}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, r := range in.Rows {
		if r.ID == "" || seen[r.ID] || r.Label < 0 || r.Label > 2 || len(r.Feature) != len(in.Source.Gamma) {
			return fmt.Errorf("invalid or duplicate row")
		}
		seen[r.ID] = true
		switch r.Split {
		case "train", "validation", "calibration", "test":
		default:
			return fmt.Errorf("invalid split")
		}
		counts[r.Split]++
		total := 0.
		for _, p := range r.Reference {
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
				return fmt.Errorf("invalid reference probability")
			}
			total += p
		}
		if math.Abs(total-1) > 1e-5 {
			return fmt.Errorf("reference probabilities do not sum to one")
		}
		for _, v := range r.Feature {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("invalid feature")
			}
		}
	}
	for _, s := range []string{"train", "validation", "calibration", "test"} {
		if counts[s] == 0 {
			return fmt.Errorf("missing split %s", s)
		}
	}
	return nil
}

var sink [3]float64

func run() error {
	path := flag.String("features", "", "local reviewed export JSON; never committed")
	out := flag.String("out", "", "new local artifact directory")
	iterations := flag.Int("iterations", 10000, "warm predictions per variant")
	profile := flag.String("cpuprofile", "", "optional LOCAL-ONLY CPU pprof file")
	flag.Parse()
	if *path == "" || *out == "" || *iterations < 1 || *iterations > 10000000 {
		return fmt.Errorf("features, new out directory and 1..10000000 iterations required")
	}
	f, e := os.Open(*path)
	if e != nil {
		return e
	}
	data, e := io.ReadAll(io.LimitReader(f, 64<<20+1))
	f.Close()
	if e != nil {
		return e
	}
	if len(data) > 64<<20 {
		return fmt.Errorf("feature export exceeds 64 MiB")
	}
	var in input
	if e = json.Unmarshal(data, &in); e != nil {
		return e
	}
	if e = check(in); e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if *profile != "" {
		f, e := os.Create(*profile)
		if e != nil {
			return e
		}
		defer f.Close()
		if e = pprof.StartCPUProfile(f); e != nil {
			return e
		}
		defer pprof.StopCPUProfile()
	}
	specs := []struct {
		name string
		kind tinyhead.Kind
		t    float64
	}{{"folded-f32", tinyhead.Float32, 0}, {"row-int8", tinyhead.Int8, 0}, {"ternary-0.5", tinyhead.Ternary, .5}, {"ternary-0.7", tinyhead.Ternary, .7}, {"ternary-1", tinyhead.Ternary, 1}, {"ternary-1.3", tinyhead.Ternary, 1.3}}
	var results []result
	best := ""
	bestLoss := math.Inf(1)
	// Validation-only selection occurs before reporting any development test score.
	models := make([]*tinyhead.Model, len(specs))
	for i, s := range specs {
		m, e := tinyhead.Build(in.Source, s.kind, s.t)
		if e != nil {
			return e
		}
		models[i] = m
		if s.kind == tinyhead.Ternary {
			v := score(m, in.Rows, "validation")
			if v.NLL < bestLoss {
				bestLoss = v.NLL
				best = s.name
			}
		}
	}
	parityPass := true
	for i, s := range specs {
		m := models[i]
		b := m.Encode()
		hash := sha256.Sum256(b)
		if e = os.WriteFile(filepath.Join(*out, s.name+".rdh"), b, 0600); e != nil {
			return e
		}
		r := result{Name: s.name, Threshold: s.t, Bytes: len(b), SHA256: hex.EncodeToString(hash[:]), ZeroFraction: m.ZeroFraction(), Splits: map[string]metrics{}, RegressionPass: true}
		if s.kind == tinyhead.Ternary {
			r.FiveTritBytes = (3*m.Width() + 4) / 5
			r.TwoBitBytes = (3*m.Width() + 3) / 4
		}
		for _, split := range []string{"train", "validation", "calibration", "test"} {
			v := score(m, in.Rows, split)
			r.Splits[split] = v
			if i == 0 {
				if v.MaxError > 1e-5 || v.Flips != 0 {
					parityPass = false
				}
			} else {
				base := results[0].Splits[split]
				if v.Accuracy+0.01 < base.Accuracy || v.AcceptedPrecision+0.01 < base.AcceptedPrecision || v.StrongToFast > base.StrongToFast {
					r.RegressionPass = false
				}
			}
		}
		scratch := make([]float64, m.Width())
		x := in.Rows[0].Feature
		for j := 0; j < 100; j++ {
			sink, _ = m.Predict(x, scratch)
		}
		r.Allocations = testing.AllocsPerRun(100, func() { sink, _ = m.Predict(x, scratch) })
		start := time.Now()
		for j := 0; j < *iterations; j++ {
			sink, _ = m.Predict(in.Rows[j%len(in.Rows)].Feature, scratch)
		}
		r.Nanoseconds = float64(time.Since(start).Nanoseconds()) / float64(*iterations)
		results = append(results, r)
	}
	hash := sha256.Sum256(data)
	report := struct {
		Experiment string          `json:"experiment"`
		Scope      string          `json:"scope"`
		FeatureSHA string          `json:"feature_sha256"`
		Go         string          `json:"go"`
		Arch       string          `json:"arch"`
		Iterations int             `json:"iterations"`
		Profiled   bool            `json:"cpu_profile_enabled"`
		Provenance json.RawMessage `json:"encoder_export"`
		Parity     bool            `json:"fp32_reference_parity_pass"`
		Selected   string          `json:"validation_selected_ternary"`
		Results    []result        `json:"results"`
	}{"tinyhead-01", "previously viewed synthetic development; warm head only; no encoder speedup claim", hex.EncodeToString(hash[:]), runtime.Version(), runtime.GOARCH, *iterations, *profile != "", in.Provenance, parityPass, best, results}
	b, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600); e != nil {
		return e
	}
	fmt.Println(string(b))
	if !parityPass {
		return fmt.Errorf("FP32 reference parity failed; do not publish")
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
