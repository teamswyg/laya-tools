package claimtree

import (
	"math"
	"testing"
)

func TestFitStepAndReplay(t *testing.T) {
	x := make([]Vector, 64)
	y := make([]float64, 64)
	for i := range x {
		x[i][3] = float64(i)
		if i >= 32 {
			y[i] = 8
		} else {
			y[i] = -2
		}
	}
	m, e := Fit(x, y, 4, 16)
	if e != nil {
		t.Fatal(e)
	}
	if m.Nodes != 3 || m.Feature[0] != 3 {
		t.Fatalf("unexpected tree %+v", m)
	}
	for i := range x {
		if m.Score(x[i]) != y[i] {
			t.Fatal("wrong prediction")
		}
	}
	copyModel, e := Fit(x, y, 4, 16)
	if e != nil || m != copyModel {
		t.Fatal("not deterministic")
	}
	if n := testing.AllocsPerRun(100, func() { _ = m.Score(x[0]) }); n != 0 {
		t.Fatalf("allocations %v", n)
	}
}
func TestTiesAndMinimumLeaf(t *testing.T) {
	x := make([]Vector, 32)
	y := make([]float64, 32)
	for i := range x {
		x[i][0] = float64(i)
		x[i][1] = float64(i)
		if i >= 16 {
			y[i] = 4
		}
	}
	m, e := Fit(x, y, 1, 16)
	if e != nil {
		t.Fatal(e)
	}
	if m.Feature[0] != 0 {
		t.Fatal("feature tie ordering")
	}
	m, e = Fit(x, y, 4, 17)
	if e != nil || m.Nodes != 1 || m.Score(x[0]) != 2 {
		t.Fatal("minimum leaf ignored")
	}
	for i := range x {
		x[i] = Vector{}
	}
	m, e = Fit(x, y, 4, 1)
	if e != nil || m.Nodes != 1 {
		t.Fatal("split identical features")
	}
}
func TestMalformed(t *testing.T) {
	x := []Vector{{}}
	y := []float64{0}
	x[0][0] = math.NaN()
	if _, e := Fit(x, y, 1, 1); e == nil {
		t.Fatal("accepted nan")
	}
	x[0][0] = 0
	y[0] = math.Inf(1)
	if _, e := Fit(x, y, 1, 1); e == nil {
		t.Fatal("accepted inf")
	}
	m := Model{Nodes: 1}
	m.Feature[0] = -1
	if e := m.Validate(); e != nil {
		t.Fatal(e)
	}
	m.Feature[0] = 0
	if e := m.Validate(); e == nil {
		t.Fatal("accepted cycle")
	}
	m.Feature[0] = -1
	m.Nodes = 2
	if e := m.Validate(); e == nil {
		t.Fatal("accepted unreachable node")
	}
}

var benchScore float64

func BenchmarkScoreDepth4(b *testing.B) {
	var m Model
	m.Nodes = 9
	for i := 0; i < 4; i++ {
		m.Feature[i] = int8(i)
		m.Left[i] = uint8(i + 1)
		m.Right[i] = uint8(5 + i)
		m.Cut[i] = .5
		m.Feature[5+i] = -1
	}
	m.Feature[4] = -1
	m.Value[4] = 2
	if e := m.Validate(); e != nil {
		b.Fatal(e)
	}
	x := Vector{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchScore = m.Score(x)
	}
}
