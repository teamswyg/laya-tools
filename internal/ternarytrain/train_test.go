package ternarytrain

import (
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"math"
	"testing"
)

func fixture() (tinyhead.Source, []Row, []Row) {
	n := 12
	s := tinyhead.Source{Gamma: make([]float64, n), Beta: make([]float64, n), Weight: make([]float64, 3*n), Temperature: .5}
	for i := range s.Gamma {
		s.Gamma[i] = 1
	}
	for i := range s.Weight {
		s.Weight[i] = .01 * math.Sin(float64(i))
	}
	train := []Row{}
	val := []Row{}
	for label := 0; label < 3; label++ {
		x := make([]float64, n)
		x[label*4] = 2
		x[label*4+1] = 1
		train = append(train, Row{ID: string(rune('a' + label)), Split: "train", Label: label, Feature: x})
		val = append(val, Row{ID: string(rune('d' + label)), Split: "validation", Label: label, Feature: x})
	}
	return s, train, val
}
func TestSTEUpdatesAndDeterminism(t *testing.T) {
	s, train, val := fixture()
	initial, _ := tinyhead.Build(s, tinyhead.Ternary, 1)
	before, _, _ := Evaluate(initial, val)
	c := Config{Seed: 1729, Threshold: 1, LR: .01, Epochs: 40, Batch: 2, Temperature: .5}
	m, tr, e := Train(s, train, val, c)
	if e != nil {
		t.Fatal(e)
	}
	after, _, e := Evaluate(m, val)
	if e != nil || after.NLL >= before.NLL || tr.Updates != 80 || tr.SelectedEpoch < 1 {
		t.Fatalf("no learning before=%v after=%v updates=%d err=%v", before, after, tr.Updates, e)
	}
	again, tr2, e := Train(s, train, val, c)
	if e != nil || Hash(m.Encode()) != Hash(again.Encode()) || tr.SelectedEpoch != tr2.SelectedEpoch {
		t.Fatal("nondeterministic selection")
	}
	if s.Weight[0] != 0 {
		t.Fatal("modified source")
	}
	if tr.Epochs[len(tr.Epochs)-1].ChangedSymbols == 0 {
		t.Fatal("no ternary symbols changed")
	}
}
func TestNoFinalGradientsOrSelection(t *testing.T) {
	s, train, val := fixture()
	c := Config{Seed: 1, Threshold: 1, LR: .01, Epochs: 1, Batch: 2, Temperature: .5}
	train[0].Split = "final"
	if _, _, e := Train(s, train, val, c); e == nil {
		t.Fatal("final gradients accepted")
	}
	train[0].Split = "train"
	val[0].Split = "final"
	if _, _, e := Train(s, train, val, c); e == nil {
		t.Fatal("final selection accepted")
	}
}
func TestSurrogateFiniteDifferenceWithDetachedForwardOffset(t *testing.T) {
	x := []float64{.3, -.2}
	w := []float64{.1, .2, -.1, .3, .2, -.4}
	temp := .75
	label := 1
	prob := func(w []float64) [3]float64 {
		var p [3]float64
		sum := 0.
		for c := 0; c < 3; c++ {
			p[c] = math.Exp((w[2*c]*x[0] + w[2*c+1]*x[1]) / temp)
			sum += p[c]
		}
		for c := range p {
			p[c] /= sum
		}
		return p
	}
	p := prob(w)
	g := make([]float64, 6)
	var b [3]float64
	Surrogate(p, label, x, temp, g, &b)
	epsilon := 1e-6
	for i := range w {
		w[i] += epsilon
		a := -math.Log(prob(w)[label])
		w[i] -= 2 * epsilon
		z := -math.Log(prob(w)[label])
		w[i] += epsilon
		if math.Abs((a-z)/(2*epsilon)-g[i]) > 1e-8 {
			t.Fatal("surrogate chain rule")
		}
	}
}
func TestGateDoesNotRewardAbstainingEverything(t *testing.T) {
	parent := Metrics{Accuracy: 1}
	v := Metrics{Accuracy: 1, Precision: 1, Recall: 1, Coverage: 0}
	if Pass(v, parent, 500) {
		t.Fatal("zero coverage accepted")
	}
	v.Coverage = .5
	if !Pass(v, parent, 500) {
		t.Fatal("valid gate rejected")
	}
	v.StrongToFast = 1
	if Pass(v, parent, 500) {
		t.Fatal("unsafe downgrade accepted")
	}
}

func TestDistillationGradient(t *testing.T) {
	x := []float64{.3, -.2}
	w := []float64{.1, .2, -.1, .3, .2, -.4}
	teacher := [3]float64{.3, -.1, -.2}
	temp := 1.5
	weight := .1
	prob := func(w []float64) [3]float64 {
		var p [3]float64
		sum := 0.
		for c := 0; c < 3; c++ {
			p[c] = math.Exp((w[2*c]*x[0] + w[2*c+1]*x[1]) / temp)
			sum += p[c]
		}
		for c := range p {
			p[c] /= sum
		}
		return p
	}
	g := make([]float64, 6)
	var b [3]float64
	DistillSurrogate(prob(w), teacher, x, temp, weight, g, &b)
	objective := func() float64 {
		dummy := make([]float64, 6)
		var bias [3]float64
		return DistillSurrogate(prob(w), teacher, x, temp, weight, dummy, &bias)
	}
	epsilon := 1e-6
	for i := range w {
		w[i] += epsilon
		a := objective()
		w[i] -= 2 * epsilon
		z := objective()
		w[i] += epsilon
		if math.Abs((a-z)/(2*epsilon)-g[i]) > 1e-9 {
			t.Fatal("distillation gradient mismatch")
		}
	}
}
