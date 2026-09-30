// Package staticembed implements a pinned static embedding in pure Go.
// This is a distinct experimental family, not a Laya encoder replacement.
package staticembed

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"unicode/utf8"
)

const Rows = 29528
const Dimension = 64
const Revision = "389b9f64be5aa4ae7a6bc6fe95ef20ce485ae5da"
const ModelSHA = "f95ffde02ad06f63ae38eb9d400038cd5ccaf8411ec3cb650c6025113f96cbb8"
const VocabularySHA256 = "e67e803f624fb4d67dea1c730d06e1067e1b14d830e2c2202569e3ef0f70bb50"
const MaxBytes = 4096

// Model is immutable after loading. Calls own their scratch/results, so no
// cache mutation or locks are needed. Quantized tables stay packed in RAM.
type Model struct {
	tokens  tokenizer
	mode    string
	fp      []float32
	int8s   []int8
	ternary []byte
	scales  []float32
}

func readVerified(path, want string, limit int64) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(b)
	if int64(len(b)) > limit || hex.EncodeToString(h[:]) != want {
		return nil, fmt.Errorf("asset hash/size mismatch")
	}
	return b, nil
}
func Load(dir, mode string) (*Model, error) {
	if mode != "fp32" && mode != "int8" && mode != "ternary" {
		return nil, fmt.Errorf("invalid precision")
	}
	tb, e := readVerified(filepath.Join(dir, "tokenizer.json"), VocabularySHA256, 1<<20)
	if e != nil {
		return nil, e
	}
	t, e := parseTokenizer(tb)
	if e != nil {
		return nil, e
	}
	b, e := readVerified(filepath.Join(dir, "model.safetensors"), ModelSHA, 8<<20)
	if e != nil {
		return nil, e
	}
	if len(b) < 8 {
		return nil, fmt.Errorf("short header")
	}
	n := int(binary.LittleEndian.Uint64(b[:8]))
	if n < 1 || n > 4096 || 8+n > len(b) {
		return nil, fmt.Errorf("invalid header")
	}
	var h struct {
		Embeddings struct {
			DType   string `json:"dtype"`
			Shape   []int  `json:"shape"`
			Offsets []int  `json:"data_offsets"`
		}
	}
	if e = json.Unmarshal(b[8:8+n], &h); e != nil {
		return nil, e
	}
	v := h.Embeddings
	if v.DType != "F32" || len(v.Shape) != 2 || v.Shape[0] != Rows || v.Shape[1] != Dimension || len(v.Offsets) != 2 || v.Offsets[0] != 0 || v.Offsets[1] != Rows*Dimension*4 || len(b) != 8+n+v.Offsets[1] {
		return nil, fmt.Errorf("unsupported tensor")
	}
	m := &Model{tokens: t, mode: mode}
	if mode == "fp32" {
		m.fp = make([]float32, Rows*Dimension)
	} else {
		m.scales = make([]float32, Rows)
		if mode == "int8" {
			m.int8s = make([]int8, Rows*Dimension)
		} else {
			m.ternary = make([]byte, Rows*Dimension/4)
		}
	}
	raw := b[8+n:]
	for r := 0; r < Rows; r++ {
		var row [Dimension]float32
		scale := float32(0)
		for j := range row {
			x := math.Float32frombits(binary.LittleEndian.Uint32(raw[(r*Dimension+j)*4:]))
			if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
				return nil, fmt.Errorf("nonfinite embedding")
			}
			row[j] = x
			if mode == "int8" {
				scale = max(scale, float32(math.Abs(float64(x))))
			} else {
				scale += float32(math.Abs(float64(x)))
			}
		}
		switch mode {
		case "fp32":
			copy(m.fp[r*Dimension:], row[:])
		case "int8":
			scale /= 127
			m.scales[r] = scale
			if scale != 0 {
				for j, x := range row {
					m.int8s[r*Dimension+j] = int8(math.Max(-127, math.Min(127, math.Round(float64(x/scale)))))
				}
			}
		case "ternary":
			scale /= Dimension
			m.scales[r] = scale
			if scale != 0 {
				for j, x := range row {
					if float32(math.Abs(float64(x))) >= .7*scale {
						q := byte(1)
						if x < 0 {
							q = 2
						}
						i := r*Dimension + j
						m.ternary[i/4] |= q << uint(2*(i%4))
					}
				}
			}
		}
	}
	return m, nil
}
func (m *Model) TableBytes() int {
	return len(m.fp)*4 + len(m.int8s) + len(m.ternary) + len(m.scales)*4
}
func (m *Model) Tokenize(s string) ([]int, error) {
	if len(s) > MaxBytes || !utf8.ValidString(s) {
		return nil, fmt.Errorf("input outside UTF-8/4096-byte scope")
	}
	return m.tokens.encode(s), nil
}
func (m *Model) Encode(s string) ([Dimension]float32, []int, error) {
	var out [Dimension]float32
	ids, e := m.Tokenize(s)
	if e != nil {
		return out, nil, e
	}
	for _, id := range ids {
		off := id * Dimension
		for j := range out {
			var x float32
			switch m.mode {
			case "fp32":
				x = m.fp[off+j]
			case "int8":
				x = float32(m.int8s[off+j]) * m.scales[id]
			case "ternary":
				i := off + j
				q := (m.ternary[i/4] >> uint(2*(i%4))) & 3
				if q == 1 {
					x = m.scales[id]
				} else if q == 2 {
					x = -m.scales[id]
				}
			}
			out[j] += x
		}
	}
	if len(ids) == 0 {
		return out, ids, nil
	}
	sum := float32(0)
	for j := range out {
		out[j] /= float32(len(ids))
		sum += out[j] * out[j]
	}
	denom := float32(math.Sqrt(float64(sum))) + 1e-32
	for j := range out {
		out[j] /= denom
	}
	return out, ids, nil
}
func Cosine(a, b [Dimension]float32) float64 {
	s := 0.
	for i, v := range a {
		s += float64(v) * float64(b[i])
	}
	return s
}
