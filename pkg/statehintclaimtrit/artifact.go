// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math"

	"github.com/teamswyg/laya-tools/pkg/tritpack"
)

const (
	ArtifactVersion     = 1
	ArtifactHeaderBytes = 256
	ArtifactBytes       = ArtifactHeaderBytes + ScaleCount*4 + BiasCount*4 + PackedWeightBytes + sha256.Size
)

func orderHash() [sha256.Size]byte {
	var b bytes.Buffer
	for _, h := range Heads() {
		b.WriteString(h.String())
		b.WriteByte(0)
	}
	for _, s := range States() {
		b.WriteString(s.String())
		b.WriteByte(0)
	}
	return sha256.Sum256(b.Bytes())
}

func numericContractHash() [sha256.Size]byte {
	return sha256.Sum256([]byte("rqt-v1;feature-major-head-state;per-head-f64-absmean-stored-f32-floor1e-5;round-to-even-clip1;little-base3-trit-plus1-padding-zero;float-features-bias-scales-f64-accumulation;mean-three-head-categorical-ce-unknown-explicit;parent-float-shadow-fresh-adamw-step0;lr.001-decay.01-weights-only-beta.9-.999-eps1e-8;identity-ste-no-scale-mask-or-scale-gradient;requantize-once-per-batch-and-final;epochs40-batch32-pcg-seed1729;fixed-T1-confidence.9-margin.05"))
}

// RQT v1 is distinct from RSC/RSH/RSM/RSP and serializes no float master weights
// or optimizer. All integers and floats are little-endian. Header byte ranges:
//
//	0:4 RQT\0; 4:6 version; 6:8 header size; 8:10 bins; 10 heads; 11 states;
//	12 mode (PTQ=1, QAT=2); 13:16 zero; 16:24 inherited steps;
//	24:32 new optimizer steps; 32:40 parent initialization seed;
//	40:104 zero-padded feature schema; 104:136 head/state order SHA-256;
//	136:168 numeric contract SHA-256; 168:200 verified parent artifact SHA-256;
//	200:208 confidence .9; 208:216 margin .05; 216:224 temperature 1;
//	224:232 scale floor 1e-5; 232:256 zero.
//
// Payload: 256:268 three float32 scales; 268:304 nine float32 biases ordered
// head/state; 304:3991 feature/head/state packed trits; 3991:4023 SHA-256 of all
// preceding bytes. The last packed byte holds two trits and three zero-trit
// padding digits. QAT steps must be a positive multiple of the fixed 40 epochs.
// Nonfinite fields, invalid mode/steps, bad hashes/order/version,
// noncanonical header/padding, corruption, truncation and trailing bytes reject.
// SHA-256 detects corruption and records parent identity, not authenticity.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RQT\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], ArtifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureBins)
	data[10], data[11], data[12] = HeadCount, StateCount, byte(m.mode)
	binary.LittleEndian.PutUint64(data[16:24], m.baseSteps)
	binary.LittleEndian.PutUint64(data[24:32], m.newSteps)
	binary.LittleEndian.PutUint64(data[32:40], uint64(m.parentSeed))
	copy(data[40:104], FeatureSchema)
	order, contract := orderHash(), numericContractHash()
	copy(data[104:136], order[:])
	copy(data[136:168], contract[:])
	copy(data[168:200], m.parentSHA[:])
	binary.LittleEndian.PutUint64(data[200:208], math.Float64bits(ConfidenceFloor))
	binary.LittleEndian.PutUint64(data[208:216], math.Float64bits(MarginFloor))
	binary.LittleEndian.PutUint64(data[216:224], math.Float64bits(1))
	binary.LittleEndian.PutUint64(data[224:232], math.Float64bits(ScaleFloor))
	offset := ArtifactHeaderBytes
	for _, s := range m.scales {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(s))
		offset += 4
	}
	for _, head := range m.bias {
		for _, b := range head {
			binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(b))
			offset += 4
		}
	}
	copy(data[offset:offset+PackedWeightBytes], m.packed[:])
	offset += PackedWeightBytes
	sum := sha256.Sum256(data[:offset])
	copy(data[offset:], sum[:])
	n, err := writer.Write(data[:])
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func Load(reader io.Reader) (*Model, error) {
	if reader == nil {
		return nil, ErrArtifact
	}
	var data [ArtifactBytes]byte
	if _, err := io.ReadFull(reader, data[:]); err != nil {
		return nil, ErrArtifact
	}
	var trailing [1]byte
	if n, err := io.ReadFull(reader, trailing[:]); n != 0 || err != io.EOF {
		return nil, ErrArtifact
	}
	var schema [64]byte
	copy(schema[:], FeatureSchema)
	order, contract := orderHash(), numericContractHash()
	var reserved [24]byte
	if string(data[:4]) != "RQT\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion ||
		binary.LittleEndian.Uint16(data[6:8]) != ArtifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureBins ||
		data[10] != HeadCount || data[11] != StateCount || !bytes.Equal(data[13:16], reserved[:3]) ||
		!bytes.Equal(data[40:104], schema[:]) || !bytes.Equal(data[104:136], order[:]) || !bytes.Equal(data[136:168], contract[:]) ||
		binary.LittleEndian.Uint64(data[200:208]) != math.Float64bits(ConfidenceFloor) ||
		binary.LittleEndian.Uint64(data[208:216]) != math.Float64bits(MarginFloor) ||
		binary.LittleEndian.Uint64(data[216:224]) != math.Float64bits(1) ||
		binary.LittleEndian.Uint64(data[224:232]) != math.Float64bits(ScaleFloor) || !bytes.Equal(data[232:256], reserved[:]) {
		return nil, ErrArtifact
	}
	end := ArtifactBytes - sha256.Size
	sum := sha256.Sum256(data[:end])
	if !bytes.Equal(data[end:], sum[:]) {
		return nil, ErrArtifact
	}
	m := &Model{mode: Mode(data[12]), baseSteps: binary.LittleEndian.Uint64(data[16:24]), newSteps: binary.LittleEndian.Uint64(data[24:32]), parentSeed: int64(binary.LittleEndian.Uint64(data[32:40]))}
	copy(m.parentSHA[:], data[168:200])
	offset := ArtifactHeaderBytes
	for h := range m.scales {
		m.scales[h] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}
	for h := range m.bias {
		for c := range m.bias[h] {
			m.bias[h][c] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
			offset += 4
		}
	}
	copy(m.packed[:], data[offset:offset+PackedWeightBytes])
	// Validate without unpacking: the resident model remains packed.
	if !m.usable() || tritpack.Validate(m.packed[:], WeightTritCount) != nil {
		return nil, ErrArtifact
	}
	m.parentSHAHex = hex.EncodeToString(m.parentSHA[:])
	return m, nil
}
