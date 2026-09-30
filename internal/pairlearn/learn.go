// Package pairlearn trains a bounded binary relevance probe. It shares the
// runtime feature contract with hintlearn but not its synthetic ranking loss.
package pairlearn

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
)

// Dataset stores sparse features in contiguous columns. No shared mutable
// state or locks; one owner per training run. Values retain scorer precision.
type Dataset struct {
	Split    string
	Offsets  []int
	Indices  []uint16
	Values   []float64
	Labels   []float64
	Groups   []int
	Excluded int
}

func Prepare(rows []paireval.Row, split paireval.Split, name string) (Dataset, error) {
	d := Dataset{Split: name, Offsets: []int{0}}
	if name != "development" && name != "validation" {
		return d, fmt.Errorf("training preparation rejects evaluation split")
	}
	if len(rows) != len(split.Rows) {
		return d, fmt.Errorf("membership length mismatch")
	}
	for i, r := range rows {
		if split.Rows[i].Split != name {
			continue
		}
		if r.Label == nil || (*r.Label != 0 && *r.Label != 1) {
			return d, fmt.Errorf("invalid label")
		}
		if !paireval.InScope(r) {
			d.Excluded++
			continue
		}
		fs := hintlearn.Features(r.Query, r.Code)
		if len(d.Indices)+len(fs) > 32_000_000 {
			return d, fmt.Errorf("feature cache exceeds 320 MB payload budget")
		}
		for _, f := range fs {
			d.Indices = append(d.Indices, uint16(f.Index))
			d.Values = append(d.Values, f.Value)
		}
		d.Offsets = append(d.Offsets, len(d.Indices))
		d.Labels = append(d.Labels, float64(*r.Label))
		d.Groups = append(d.Groups, split.Rows[i].Group)
	}
	if len(d.Labels) == 0 {
		return d, fmt.Errorf("empty dataset")
	}
	return d, nil
}
func (d Dataset) score(w []float64, i int) float64 {
	s := 0.
	for k := d.Offsets[i]; k < d.Offsets[i+1]; k++ {
		s += w[d.Indices[k]] * d.Values[k]
	}
	return s
}
func logistic(s float64) float64 {
	if s >= 0 {
		return 1 / (1 + math.Exp(-s))
	}
	e := math.Exp(s)
	return e / (1 + e)
}
func loss(s, y float64) float64 { return math.Max(s, 0) - y*s + math.Log1p(math.Exp(-math.Abs(s))) }
func NLL(d Dataset, w []float64) float64 {
	v := 0.
	for i, y := range d.Labels {
		v += loss(d.score(w, i), y)
	}
	return v / float64(len(d.Labels))
}

type Config struct {
	Seed             uint64
	Mode             string
	LearningRate, L2 float64
	Epochs, Batch    int
}
type Result struct {
	Config        Config
	Epoch         int
	ValidationNLL float64
	TrainingNLL   float64
	Weights       []float64 `json:"-"`
}

func validate(d Dataset) error {
	if len(d.Labels) == 0 || len(d.Groups) != len(d.Labels) || len(d.Offsets) != len(d.Labels)+1 || len(d.Indices) != len(d.Values) || d.Offsets[0] != 0 || d.Offsets[len(d.Labels)] != len(d.Values) {
		return fmt.Errorf("invalid sparse columns")
	}
	for i, y := range d.Labels {
		if (y != 0 && y != 1) || d.Offsets[i] > d.Offsets[i+1] || d.Offsets[i] < 0 {
			return fmt.Errorf("invalid row")
		}
	}
	for i, v := range d.Values {
		if int(d.Indices[i]) >= hintlearn.Dimension || math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("invalid feature")
		}
	}
	return nil
}
func Fit(train, validation Dataset, c Config) (Result, error) {
	var best Result
	if train.Split != "development" || validation.Split != "validation" {
		return best, fmt.Errorf("invalid training split")
	}
	for _, d := range []Dataset{train, validation} {
		if e := validate(d); e != nil {
			return best, e
		}
	}
	gs := slices.Clone(train.Groups)
	slices.Sort(gs)
	for _, g := range validation.Groups {
		if _, found := slices.BinarySearch(gs, g); found {
			return best, fmt.Errorf("group leakage")
		}
	}
	if (c.Mode != "fp32" && c.Mode != "ternary_ste") || !(c.LearningRate > 0 && c.LearningRate <= 1) || !(c.L2 >= 0 && c.L2 <= 1) || c.Epochs < 1 || c.Epochs > 100 || c.Batch < 1 || c.Batch > 1024 {
		return best, fmt.Errorf("invalid configuration")
	}
	rng := rand.New(rand.NewPCG(c.Seed, c.Seed+1))
	w := make([]float64, hintlearn.Dimension)
	for i := range w {
		w[i] = rng.NormFloat64() * .01
	}
	order := make([]int, len(train.Labels))
	for i := range order {
		order[i] = i
	}
	grad := make([]float64, len(w))
	best = Result{Config: c, ValidationNLL: math.Inf(1)}
	for epoch := 1; epoch <= c.Epochs; epoch++ {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for start := 0; start < len(order); start += c.Batch {
			end := min(start+c.Batch, len(order))
			clear(grad)
			// Identity STE, detached global ternary scale; shadow weights stay FP64.
			qw := hintlearn.Quantize(w, c.Mode)
			for _, i := range order[start:end] {
				delta := logistic(train.score(qw, i)) - train.Labels[i]
				for k := train.Offsets[i]; k < train.Offsets[i+1]; k++ {
					grad[train.Indices[k]] += delta * train.Values[k]
				}
			}
			for i := range w {
				w[i] -= c.LearningRate * (grad[i]/float64(end-start) + c.L2*w[i])
			}
		}
		qw := hintlearn.Quantize(w, c.Mode)
		nll := NLL(validation, qw)
		if math.IsNaN(nll) || math.IsInf(nll, 0) {
			return best, fmt.Errorf("nonfinite training")
		}
		if nll < best.ValidationNLL {
			best.Epoch = epoch
			best.ValidationNLL = nll
			best.Weights = qw
		}
	}
	best.TrainingNLL = NLL(train, best.Weights)
	return best, nil
}

// AUC measures threshold-free ordering on the same eligible validation rows.
func AUC(d Dataset, w []float64) *float64 {
	ps := make([]paireval.Prediction, len(d.Labels))
	for i, y := range d.Labels {
		ps[i] = paireval.Prediction{Label: int(y), Score: d.score(w, i), Eligible: true}
	}
	return paireval.AUC(ps)
}
