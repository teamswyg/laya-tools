package hintlearn

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Encode stores FP32, INT8, or bitmap + nonzero sign bits. Decode expands to
// float64 coefficients for the reference Go scorer; packing is a disk format,
// not a claim of native ternary arithmetic or reduced decoded coefficient RAM.
func Encode(w []float64, mode string) ([]byte, error) {
	if len(w) != Dimension {
		return nil, fmt.Errorf("invalid width")
	}
	kind := byte(0)
	switch mode {
	case "fp32":
	case "int8":
		kind = 1
	case "ternary_ste", "ternary_ptq":
		kind = 2
	default:
		return nil, fmt.Errorf("invalid mode")
	}
	mx := 0.
	nonzero := 0
	for _, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.IsInf(float64(float32(v)), 0) {
			return nil, fmt.Errorf("nonfinite coefficient")
		}
		mx = math.Max(mx, math.Abs(v))
		if v != 0 {
			nonzero++
		}
	}
	scale := float64(float32(mx))
	if kind == 1 {
		scale = float64(float32(mx / 127))
	}
	size := 24 + Dimension*4
	if kind == 1 {
		size = 24 + Dimension
	}
	if kind == 2 {
		size = 24 + (Dimension+7)/8 + (nonzero+7)/8
	}
	b := make([]byte, size)
	copy(b, "RIIDOH01")
	binary.LittleEndian.PutUint32(b[8:], Dimension)
	b[12] = kind
	binary.LittleEndian.PutUint32(b[16:], math.Float32bits(float32(scale)))
	binary.LittleEndian.PutUint32(b[20:], uint32(nonzero))
	sign := 0
	for i, v := range w {
		switch kind {
		case 0:
			binary.LittleEndian.PutUint32(b[24+4*i:], math.Float32bits(float32(v)))
		case 1:
			if scale != 0 {
				q := math.Round(v / scale)
				if q < -127 || q > 127 || math.Abs(v-q*scale) > 1e-7*math.Max(1, math.Abs(v)) {
					return nil, fmt.Errorf("not int8 quantized")
				}
				b[24+i] = byte(int8(q))
			}
		case 2:
			if v != 0 {
				if math.Abs(math.Abs(v)-scale) > 1e-7*math.Max(1, scale) {
					return nil, fmt.Errorf("not ternary")
				}
				b[24+i/8] |= 1 << uint(i%8)
				if v > 0 {
					b[24+Dimension/8+sign/8] |= 1 << uint(sign%8)
				}
				sign++
			}
		}
	}
	return b, nil
}
func Decode(b []byte) ([]float64, error) {
	if len(b) < 24 || string(b[:8]) != "RIIDOH01" || binary.LittleEndian.Uint32(b[8:]) != Dimension || b[13] != 0 || b[14] != 0 || b[15] != 0 {
		return nil, fmt.Errorf("invalid header")
	}
	kind := b[12]
	scale := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[16:])))
	n := int(binary.LittleEndian.Uint32(b[20:]))
	if math.IsNaN(scale) || math.IsInf(scale, 0) || scale < 0 || n > Dimension {
		return nil, fmt.Errorf("invalid scale/count")
	}
	size := 24 + Dimension*4
	if kind == 1 {
		size = 24 + Dimension
	} else if kind == 2 {
		size = 24 + Dimension/8 + (n+7)/8
	} else if kind != 0 {
		return nil, fmt.Errorf("invalid kind")
	}
	if len(b) != size {
		return nil, fmt.Errorf("invalid length")
	}
	w := make([]float64, Dimension)
	nz := 0
	for i := range w {
		switch kind {
		case 0:
			w[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[24+4*i:])))
		case 1:
			q := int8(b[24+i])
			if q == -128 {
				return nil, fmt.Errorf("invalid int8")
			}
			w[i] = float64(q) * scale
		case 2:
			if b[24+i/8]&(1<<uint(i%8)) != 0 {
				if nz >= n || scale == 0 {
					return nil, fmt.Errorf("invalid bitmap/count")
				}
				w[i] = -scale
				if b[24+Dimension/8+nz/8]&(1<<uint(nz%8)) != 0 {
					w[i] = scale
				}
			}
		}
		if math.IsNaN(w[i]) || math.IsInf(w[i], 0) {
			return nil, fmt.Errorf("nonfinite weight")
		}
		if w[i] != 0 {
			nz++
		}
	}
	if nz != n {
		return nil, fmt.Errorf("count mismatch")
	}
	if kind == 2 && n%8 != 0 && b[len(b)-1]>>uint(n%8) != 0 {
		return nil, fmt.Errorf("nonzero padding")
	}
	return w, nil
}
