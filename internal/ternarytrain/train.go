package ternarytrain

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"math"
	"math/rand"
	"time"
)

type Config struct {
	Seed         int64   `json:"seed"`
	Threshold    float64 `json:"threshold"`
	LR           float64 `json:"learning_rate"`
	Epochs       int     `json:"epochs"`
	Batch        int     `json:"batch_size"`
	WeightDecay  float64 `json:"weight_decay"`
	Temperature  float64 `json:"temperature"`
	Distillation float64 `json:"distillation_weight"`
}
type Epoch struct {
	Epoch          int     `json:"epoch"`
	TrainNLL       float64 `json:"train_nll"`
	TrainObjective float64 `json:"train_objective"`
	Validation     Metrics `json:"validation"`
	Bytes          int     `json:"bytes"`
	ZeroFraction   float64 `json:"zero_fraction"`
	ChangedSymbols int     `json:"changed_symbols_since_initial"`
}
type Trial struct {
	Config        Config  `json:"config"`
	SelectedEpoch int     `json:"selected_epoch"`
	Updates       int     `json:"optimizer_updates"`
	BestNLL       float64 `json:"best_validation_nll"`
	Eligible      bool    `json:"eligible_under_byte_budget"`
	Seconds       float64 `json:"training_seconds"`
	Epochs        []Epoch `json:"trace"`
}

func clone(s tinyhead.Source) tinyhead.Source {
	s.Gamma = append([]float64(nil), s.Gamma...)
	s.Beta = append([]float64(nil), s.Beta...)
	s.Weight = append([]float64(nil), s.Weight...)
	return s
}

// Surrogate accumulates an identity straight-through gradient. Quantization and
// row-scale derivatives are detached. This is not the derivative of rounding.
func Surrogate(p [3]float64, label int, normalized []float64, temperature float64, g []float64, b *[3]float64) {
	for c := 0; c < 3; c++ {
		d := p[c]
		if c == label {
			d--
		}
		d /= temperature
		b[c] += d
		for i, x := range normalized {
			g[c*len(normalized)+i] += d * x
		}
	}
}
func Symbols(s tinyhead.Source, threshold float64) []int8 {
	n := len(s.Gamma)
	out := make([]int8, len(s.Weight))
	for r := 0; r < 3; r++ {
		mean := 0.
		for _, v := range s.Weight[r*n : (r+1)*n] {
			mean += math.Abs(float64(float32(v)))
		}
		mean /= float64(n)
		for i := 0; i < n; i++ {
			v := float64(float32(s.Weight[r*n+i]))
			if math.Abs(v) > threshold*mean {
				out[r*n+i] = 1
				if v < 0 {
					out[r*n+i] = -1
				}
			}
		}
	}
	return out
}
func Train(source tinyhead.Source, train, val []Row, c Config) (*tinyhead.Model, Trial, error) {
	start := time.Now()
	tr := Trial{Config: c, BestNLL: 0}
	if len(train) == 0 || len(val) == 0 || c.Epochs < 1 || c.Epochs > 1000 || c.Batch < 1 || c.LR <= 0 || c.LR > 1 || c.Temperature <= 0 || c.Threshold < 0 || c.Threshold > 10 || !finite(c.LR) || !finite(c.Threshold) || !finite(c.Temperature) || c.WeightDecay < 0 || !finite(c.WeightDecay) || c.Distillation < 0 || !finite(c.Distillation) {
		return nil, tr, fmt.Errorf("invalid training configuration")
	}
	for _, r := range train {
		if r.Split != "train" || r.Label < 0 || r.Label > 2 || len(r.Feature) != len(source.Gamma) {
			return nil, tr, fmt.Errorf("gradient rows must be train only")
		}
	}
	for _, r := range val {
		if r.Split != "validation" || r.Label < 0 || r.Label > 2 || len(r.Feature) != len(source.Gamma) {
			return nil, tr, fmt.Errorf("selection rows must be validation only")
		}
	}
	s, e := tinyhead.FoldSource(source)
	if e != nil {
		return nil, tr, e
	}
	s = clone(s)
	s.Temperature = c.Temperature
	n := len(s.Gamma)
	initial := Symbols(s, c.Threshold)
	teacherSource := clone(source)
	teacherSource.Temperature = 1
	teacher, err := tinyhead.Build(teacherSource, tinyhead.Float32, 0)
	if err != nil {
		return nil, tr, err
	}
	teacherTargets := make([][3]float64, len(train))
	teacherScratch := make([]float64, n)
	if c.Distillation > 0 {
		for i, r := range train {
			p, e := teacher.Predict(r.Feature, teacherScratch)
			if e != nil {
				return nil, tr, e
			}
			teacherTargets[i] = CenteredLogits(p, 1)
		}
	}
	g := make([]float64, len(s.Weight))
	scratch := make([]float64, n)
	var gb [3]float64
	order := make([]int, len(train))
	for i := range order {
		order[i] = i
	}
	rng := rand.New(rand.NewSource(c.Seed))
	var best *tinyhead.Model
	bestLoss := math.Inf(1)
	for epoch := 1; epoch <= c.Epochs; epoch++ {
		if time.Since(start) > 120*time.Second {
			return nil, tr, fmt.Errorf("candidate exceeded 120 seconds")
		}
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		loss := 0.
		objective := 0.
		for from := 0; from < len(order); from += c.Batch {
			to := min(from+c.Batch, len(order))
			clear(g)
			gb = [3]float64{}
			m, e := tinyhead.Build(s, tinyhead.Ternary, c.Threshold)
			if e != nil {
				return nil, tr, e
			}
			for _, idx := range order[from:to] {
				r := train[idx]
				p, e := m.Predict(r.Feature, scratch)
				if e != nil {
					return nil, tr, e
				}
				loss -= math.Log(math.Max(p[r.Label], 1e-300))
				Surrogate(p, r.Label, scratch, c.Temperature, g, &gb)
				if c.Distillation > 0 {
					objective += DistillSurrogate(p, teacherTargets[idx], scratch, c.Temperature, c.Distillation, g, &gb)
				}
			}
			count := float64(to - from)
			for i := range g {
				g[i] = g[i]/count + c.WeightDecay*s.Weight[i]
			}
			norm := 0.
			for _, v := range g {
				norm += v * v
			}
			for i := range gb {
				gb[i] /= count
				norm += gb[i] * gb[i]
			}
			factor := 1.
			if math.Sqrt(norm) > 5 {
				factor = 5 / math.Sqrt(norm)
			}
			for i := range s.Weight {
				s.Weight[i] -= c.LR * factor * g[i]
			}
			for i := range s.Bias {
				s.Bias[i] -= c.LR * factor * gb[i]
			}
			tr.Updates++
		}
		m, e := tinyhead.Build(s, tinyhead.Ternary, c.Threshold)
		if e != nil {
			return nil, tr, e
		}
		v, _, e := Evaluate(m, val)
		if e != nil {
			return nil, tr, e
		}
		symbols := Symbols(s, c.Threshold)
		changed := 0
		for i := range symbols {
			if symbols[i] != initial[i] {
				changed++
			}
		}
		tr.Epochs = append(tr.Epochs, Epoch{epoch, loss / float64(len(train)), (loss + objective) / float64(len(train)), v, m.Bytes(), m.ZeroFraction(), changed})
		if m.Bytes() <= 600 && v.NLL < bestLoss {
			bestLoss = v.NLL
			best = m
			tr.SelectedEpoch = epoch
			tr.BestNLL = v.NLL
			tr.Eligible = true
		}
	}
	tr.Seconds = time.Since(start).Seconds()
	return best, tr, nil
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// CenteredLogits removes the softmax-invariant shared offset.
func CenteredLogits(p [3]float64, temperature float64) [3]float64 {
	var z [3]float64
	mean := 0.
	for i, v := range p {
		z[i] = math.Log(math.Max(v, 1e-300)) * temperature
		mean += z[i] / 3
	}
	for i := range z {
		z[i] -= mean
	}
	return z
}

// DistillSurrogate preserves teacher margins on train rows, with detached
// quantizer/scales. Targets are centered, so the residual gradient sums to zero.
func DistillSurrogate(p, teacher [3]float64, x []float64, temperature, weight float64, g []float64, b *[3]float64) float64 {
	z := CenteredLogits(p, temperature)
	loss := 0.
	for c := 0; c < 3; c++ {
		d := z[c] - teacher[c]
		loss += .5 * weight * d * d / 3
		gradient := weight * d / 3
		b[c] += gradient
		for i, v := range x {
			g[c*len(x)+i] += gradient * v
		}
	}
	return loss
}
