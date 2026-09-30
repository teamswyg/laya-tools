package tinyhead

import (
	"encoding/binary"
	"math"
	"sync"
	"testing"
)

func fixture(n int) Source {
	s := Source{Gamma: make([]float64, n), Beta: make([]float64, n), Weight: make([]float64, 3*n), Temperature: .5}
	for i := 0; i < n; i++ {
		s.Gamma[i] = 1 + float64(i%3)/10
		s.Beta[i] = float64(i%5) / 20
	}
	for i := range s.Weight {
		s.Weight[i] = float64(i%9-4) / 8
	}
	return s
}
func TestFoldMatchesUnfolded(t *testing.T) {
	s := fixture(17)
	m, e := Build(s, Float32, 0)
	if e != nil {
		t.Fatal(e)
	}
	x := make([]float64, 17)
	for i := range x {
		x[i] = math.Sin(float64(i))
	}
	p, e := m.Predict(x, make([]float64, 17))
	if e != nil {
		t.Fatal(e)
	}
	mean := 0.
	for _, v := range x {
		mean += v
	}
	mean /= 17
	variance := 0.
	for _, v := range x {
		variance += (v - mean) * (v - mean)
	}
	var expected [3]float64
	den := 0.
	for r := 0; r < 3; r++ {
		z := s.Bias[r]
		for i, v := range x {
			z += s.Weight[r*17+i] * ((v-mean)/math.Sqrt(variance/17+1e-5)*s.Gamma[i] + s.Beta[i])
		}
		expected[r] = math.Exp(z / s.Temperature)
		den += expected[r]
	}
	for i := range p {
		if math.Abs(p[i]-expected[i]/den) > 1e-6 {
			t.Fatalf("parity %v %v", p, expected)
		}
	}
}
func TestRoundTripAndConcurrent(t *testing.T) {
	for _, kind := range []Kind{Float32, Int8, Ternary} {
		for _, n := range []int{1, 2, 7, 8, 9, 17, 1024} {
			m, e := Build(fixture(n), kind, .7)
			if e != nil {
				t.Fatal(e)
			}
			b := m.Encode()
			copyModel, e := Decode(b)
			if e != nil {
				t.Fatal(e)
			}
			x := make([]float64, n)
			for i := range x {
				x[i] = float64(i % 3)
			}
			want, _ := m.Predict(x, make([]float64, n))
			var wg sync.WaitGroup
			for j := 0; j < 4; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					p, e := copyModel.Predict(x, make([]float64, n))
					if e != nil || p != want {
						t.Errorf("roundtrip %v %v %v", p, want, e)
					}
				}()
			}
			wg.Wait()
			for k := 0; k < len(b); k++ {
				if _, e := Decode(b[:k]); e == nil {
					t.Fatalf("accepted truncation %d", k)
				}
			}
			if _, e := Decode(append(b, 0)); e == nil {
				t.Fatal("accepted trailing bytes")
			}
		}
	}
}
func TestTernaryDenseEquivalence(t *testing.T) {
	s := fixture(17)
	m, _ := Build(s, Ternary, .7)
	dense := &Model{width: 17, kind: Float32, temperature: m.temperature, bias: m.bias, payload: make([]byte, 12*17)}
	nz := 0
	for j := 0; j < 51; j++ {
		v := 0.
		if m.payload[j/8]&(1<<uint(j%8)) != 0 {
			v = -m.scale[j/17]
			if m.payload[(51+7)/8+nz/8]&(1<<uint(nz%8)) != 0 {
				v = -v
			}
			nz++
		}
		put(dense.payload[4*j:], v)
	}
	x := make([]float64, 17)
	for i := range x {
		x[i] = math.Cos(float64(i))
	}
	p, _ := m.Predict(x, make([]float64, 17))
	q, _ := dense.Predict(x, make([]float64, 17))
	for i := range p {
		if math.Abs(p[i]-q[i]) > 1e-12 {
			t.Fatalf("sign packing %v %v", p, q)
		}
	}
}
func TestRejectInvalid(t *testing.T) {
	s := fixture(8)
	s.Gamma[0] = math.NaN()
	if _, e := Build(s, Float32, 0); e == nil {
		t.Fatal("NaN")
	}
	m, _ := Build(fixture(8), Int8, 0)
	b := m.Encode()
	binary.LittleEndian.PutUint32(b[8:], MaxWidth+1)
	if _, e := Decode(b); e == nil {
		t.Fatal("large width")
	}
	if _, e := m.Predict([]float64{1}, nil); e == nil {
		t.Fatal("width")
	}
	x := make([]float64, 8)
	x[0] = math.Inf(1)
	if _, e := m.Predict(x, make([]float64, 8)); e == nil {
		t.Fatal("non-finite")
	}
}
func BenchmarkPredict(b *testing.B) {
	for _, kind := range []Kind{Float32, Int8, Ternary} {
		name := []string{"fp32", "int8", "ternary"}[kind]
		b.Run(name, func(b *testing.B) {
			m, _ := Build(fixture(1024), kind, .7)
			x := make([]float64, 1024)
			scratch := make([]float64, 1024)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, e := m.Predict(x, scratch); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
func FuzzDecode(f *testing.F) {
	for _, k := range []Kind{Float32, Int8, Ternary} {
		m, _ := Build(fixture(7), k, .7)
		f.Add(m.Encode())
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		m, e := Decode(b)
		if e != nil {
			return
		}
		p, e := m.Predict(make([]float64, m.Width()), make([]float64, m.Width()))
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range p {
			if !finite(v) {
				t.Fatal("non-finite result")
			}
		}
	})
}
