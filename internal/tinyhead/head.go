// Package tinyhead is an experimental fixed-class head over frozen Laya features.
// It does not tokenize text or execute the encoder. Models are immutable; callers
// own scratch space, so independent callers need neither locks nor shared caches.
package tinyhead

import (
	"encoding/binary"
	"errors"
	"math"
)

const MaxWidth = 65536
const header = 44
const magic = "RDTINY01"

type Kind uint32

const (
	Float32 Kind = iota
	Int8
	Ternary
)

type Source struct {
	Gamma       []float64  `json:"gamma"`
	Beta        []float64  `json:"beta"`
	Weight      []float64  `json:"weight"`
	Bias        [3]float64 `json:"bias"`
	Temperature float64    `json:"temperature"`
}
type Model struct {
	width       int
	kind        Kind
	temperature float64
	bias, scale [3]float64
	payload     []byte
}

func finite(x float64) bool   { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func f32(x float64) float64   { return float64(float32(x)) }
func put(b []byte, x float64) { binary.LittleEndian.PutUint32(b, math.Float32bits(float32(x))) }
func get(b []byte) float64    { return float64(math.Float32frombits(binary.LittleEndian.Uint32(b))) }
func (m *Model) Width() int   { return m.width }
func (m *Model) Bytes() int   { return header + len(m.payload) }
func (m *Model) ZeroFraction() float64 {
	if m.kind != Ternary {
		return 0
	}
	n := 3 * m.width
	ones := 0
	for i := 0; i < n; i++ {
		if m.payload[i/8]&(1<<uint(i%8)) != 0 {
			ones++
		}
	}
	return float64(n-ones) / float64(n)
}

// Build folds LayerNorm's learned affine transform into the linear layer.
// Normalization itself remains at inference. Ternary scale is row mean magnitude
// of selected weights; threshold is a multiple of the row mean absolute weight.
func Build(s Source, kind Kind, threshold float64) (*Model, error) {
	n := len(s.Gamma)
	if n < 1 || n > MaxWidth || len(s.Beta) != n || len(s.Weight) != 3*n || kind > Ternary || !finite(s.Temperature) || s.Temperature <= 0 || !finite(threshold) || threshold < 0 {
		return nil, errors.New("invalid dimensions, kind or calibration")
	}
	for _, a := range [][]float64{s.Gamma, s.Beta, s.Weight, s.Bias[:]} {
		for _, v := range a {
			if !finite(f32(v)) {
				return nil, errors.New("non-finite source")
			}
		}
	}
	m := &Model{width: n, kind: kind, temperature: f32(s.Temperature)}
	if m.temperature <= 0 || !finite(m.temperature) {
		return nil, errors.New("temperature outside FP32 range")
	}
	w := make([]float64, 3*n)
	for r := 0; r < 3; r++ {
		m.bias[r] = s.Bias[r]
		for i := 0; i < n; i++ {
			w[r*n+i] = f32(s.Weight[r*n+i] * s.Gamma[i])
			m.bias[r] += s.Weight[r*n+i] * s.Beta[i]
		}
		m.bias[r] = f32(m.bias[r])
	}
	switch kind {
	case Float32:
		m.payload = make([]byte, 12*n)
		for i, v := range w {
			put(m.payload[i*4:], v)
		}
	case Int8:
		m.payload = make([]byte, 3*n)
		for r := 0; r < 3; r++ {
			mx := 0.
			for _, v := range w[r*n : (r+1)*n] {
				mx = math.Max(mx, math.Abs(v))
			}
			m.scale[r] = f32(mx / 127)
			if m.scale[r] == 0 {
				m.scale[r] = 1
			}
			for i := 0; i < n; i++ {
				q := math.Round(w[r*n+i] / m.scale[r])
				q = math.Max(-127, math.Min(127, q))
				m.payload[r*n+i] = byte(int8(q))
			}
		}
	case Ternary:
		present := make([]byte, (3*n+7)/8)
		sign := make([]byte, (3*n+7)/8)
		nz := 0
		for r := 0; r < 3; r++ {
			mean := 0.
			for _, v := range w[r*n : (r+1)*n] {
				mean += math.Abs(v)
			}
			mean /= float64(n)
			sum := 0.
			count := 0
			for i := 0; i < n; i++ {
				j := r*n + i
				v := w[j]
				if math.Abs(v) > threshold*mean {
					present[j/8] |= 1 << uint(j%8)
					if v > 0 {
						sign[nz/8] |= 1 << uint(nz%8)
					}
					nz++
					count++
					sum += math.Abs(v)
				}
			}
			m.scale[r] = 1
			if count > 0 {
				m.scale[r] = f32(sum / float64(count))
			}
		}
		m.payload = append(present, sign[:(nz+7)/8]...)
	}
	// Exercise the same strict checks as loading an external artifact.
	return Decode(m.Encode())
}
func (m *Model) Encode() []byte {
	b := make([]byte, m.Bytes())
	copy(b, magic)
	binary.LittleEndian.PutUint32(b[8:], uint32(m.width))
	binary.LittleEndian.PutUint32(b[12:], uint32(m.kind))
	put(b[16:], m.temperature)
	for i := 0; i < 3; i++ {
		put(b[20+4*i:], m.bias[i])
		put(b[32+4*i:], m.scale[i])
	}
	copy(b[header:], m.payload)
	return b
}
func Decode(b []byte) (*Model, error) {
	if len(b) < header || string(b[:8]) != magic {
		return nil, errors.New("invalid tinyhead header")
	}
	m := &Model{width: int(binary.LittleEndian.Uint32(b[8:])), kind: Kind(binary.LittleEndian.Uint32(b[12:])), temperature: get(b[16:])}
	if m.width < 1 || m.width > MaxWidth || m.kind > Ternary || !finite(m.temperature) || m.temperature <= 0 {
		return nil, errors.New("invalid tinyhead dimensions/calibration")
	}
	for i := 0; i < 3; i++ {
		m.bias[i] = get(b[20+4*i:])
		m.scale[i] = get(b[32+4*i:])
		if !finite(m.bias[i]) || !finite(m.scale[i]) || (m.kind != Float32 && m.scale[i] <= 0) {
			return nil, errors.New("invalid bias/scale")
		}
	}
	n := 3 * m.width
	size := 12 * m.width
	if m.kind == Int8 {
		size = n
	}
	if m.kind == Ternary {
		size = (n + 7) / 8
		if len(b) < header+size {
			return nil, errors.New("truncated bitmap")
		}
		nz := 0
		for i := 0; i < size*8; i++ {
			if b[header+i/8]&(1<<uint(i%8)) != 0 {
				if i >= n {
					return nil, errors.New("nonzero bitmap padding")
				}
				nz++
			}
		}
		size += (nz + 7) / 8
		if len(b) == header+size && nz%8 != 0 && b[len(b)-1]>>uint(nz%8) != 0 {
			return nil, errors.New("nonzero sign padding")
		}
	}
	if len(b) != header+size {
		return nil, errors.New("wrong payload length")
	}
	if m.kind == Float32 {
		for i := 0; i < n; i++ {
			if !finite(get(b[header+4*i:])) {
				return nil, errors.New("non-finite weight")
			}
		}
	}
	m.payload = append([]byte(nil), b[header:]...)
	return m, nil
}

// Predict returns fast/standard/strong probabilities. scratch must be width long,
// exclusively owned for this call and not overlap x. Arithmetic uses float64 with
// FP32-stored weights. This portable scalar baseline makes no SIMD/GPU claim.
func (m *Model) Predict(x, scratch []float64) ([3]float64, error) {
	var p [3]float64
	if len(x) != m.width || len(scratch) != m.width {
		return p, errors.New("feature/scratch width mismatch")
	}
	mean := 0.
	for _, v := range x {
		if !finite(v) {
			return p, errors.New("non-finite feature")
		}
		mean += v
	}
	mean /= float64(m.width)
	variance := 0.
	for _, v := range x {
		d := v - mean
		variance += d * d
	}
	inv := 1 / math.Sqrt(variance/float64(m.width)+1e-5)
	if !finite(mean) || !finite(inv) || inv == 0 {
		return p, errors.New("feature normalization overflow")
	}
	for i, v := range x {
		scratch[i] = (v - mean) * inv
	}
	nz := 0
	bitmapBytes := (3*m.width + 7) / 8
	for r := 0; r < 3; r++ {
		sum := 0.
		switch m.kind {
		case Float32:
			row := m.payload[4*r*m.width : 4*(r+1)*m.width]
			for i, v := range scratch {
				sum += get(row[4*i:]) * v
			}
		case Int8:
			row := m.payload[r*m.width : (r+1)*m.width]
			for i, v := range scratch {
				sum += float64(int8(row[i])) * v
			}
		case Ternary:
			for i, v := range scratch {
				j := r*m.width + i
				if m.payload[j/8]&(1<<uint(j%8)) != 0 {
					if m.payload[bitmapBytes+nz/8]&(1<<uint(nz%8)) != 0 {
						sum += v
					} else {
						sum -= v
					}
					nz++
				}
			}
		}
		if m.kind != Float32 {
			sum *= m.scale[r]
		}
		p[r] = (sum + m.bias[r]) / m.temperature
		if !finite(p[r]) {
			return [3]float64{}, errors.New("logit overflow")
		}
	}
	mx := math.Max(p[0], math.Max(p[1], p[2]))
	den := 0.
	for i := range p {
		p[i] = math.Exp(p[i] - mx)
		den += p[i]
	}
	for i := range p {
		p[i] /= den
	}
	return p, nil
}

// FoldSource returns the equivalent FP32 folded parameters with identity affine
// normalization. It is useful for training shadow weights without the old affine.
func FoldSource(s Source) (Source, error) {
	m, e := Build(s, Float32, 0)
	if e != nil {
		return Source{}, e
	}
	out := Source{Gamma: make([]float64, m.width), Beta: make([]float64, m.width), Weight: make([]float64, 3*m.width), Bias: m.bias, Temperature: m.temperature}
	for i := range out.Gamma {
		out.Gamma[i] = 1
	}
	for i := range out.Weight {
		out.Weight[i] = get(m.payload[4*i:])
	}
	return out, nil
}

// WithTemperature creates an independent immutable calibrated model.
func (m *Model) WithTemperature(t float64) (*Model, error) {
	if !finite(t) || t <= 0 || !finite(f32(t)) || f32(t) <= 0 {
		return nil, errors.New("invalid temperature")
	}
	b := m.Encode()
	put(b[16:], t)
	return Decode(b)
}
