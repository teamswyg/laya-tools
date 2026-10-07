// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
)

const (
	ArtifactVersion     = 1
	artifactHeaderBytes = 192
	ArtifactBytes       = artifactHeaderBytes + ParameterCount*4 + sha256.Size
)

func featureHash() [sha256.Size]byte { return sha256.Sum256([]byte(FeatureSchema)) }
func orderHash() [sha256.Size]byte {
	var data [256]byte
	offset := 0
	for _, head := range Heads() {
		offset += copy(data[offset:], head.String())
		offset++ // Zero byte delimits each name.
	}
	for _, state := range States() {
		offset += copy(data[offset:], state)
		offset++
	}
	return sha256.Sum256(data[:offset])
}
func numericContractHash() [sha256.Size]byte {
	return sha256.Sum256([]byte("input-hidden-head-state-f32;shared16-relu-zero-derivative;three-independent-categorical-heads;mean-three-hard-ce;fresh-glorot-pcg-initxor696e697472636d31;fan-in2048-out16;head-fan-in16-out3;shuffle-xor9e3779b97f4a7c15;epochs40-batch32-lr.001-adamw-decay.01;beta.9-.999-eps1e-8;bias-no-decay;seed1729;fixed-T1;source-only-unqualified-no-authority;rcm-v1"))
}

// Save writes the distinct RCM v1 contract, never RSC/RSM/RQT parent weights.
// Parameters serialize in input/hidden-bias/output/output-bias order. SHA-256
// detects corruption, not publisher authenticity. Optimizer state, corpus and
// application authority are absent. The header pins fresh initialization and
// the fixed recipe, including the training-sample/step relation.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RCM\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], artifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureBins)
	binary.LittleEndian.PutUint16(data[10:12], HiddenUnits)
	data[12], data[13], data[14], data[15] = HeadCount, StateCount, 1, 1 // ReLU, RCM Glorot-PCG-v1.
	binary.LittleEndian.PutUint64(data[16:24], m.steps)
	binary.LittleEndian.PutUint64(data[24:32], uint64(m.seed))
	feature, order, contract := featureHash(), orderHash(), numericContractHash()
	copy(data[32:64], feature[:])
	copy(data[64:96], order[:])
	copy(data[96:128], contract[:])
	binary.LittleEndian.PutUint64(data[128:136], math.Float64bits(ConfidenceFloor))
	binary.LittleEndian.PutUint64(data[136:144], math.Float64bits(MarginFloor))
	binary.LittleEndian.PutUint64(data[144:152], math.Float64bits(1))
	binary.LittleEndian.PutUint32(data[152:156], m.samples)
	binary.LittleEndian.PutUint32(data[156:160], ParameterCount)
	binary.LittleEndian.PutUint32(data[160:164], WeightCount)
	offset := artifactHeaderBytes
	put := func(value float32) {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(value))
		offset += 4
	}
	for _, row := range m.input {
		for _, value := range row {
			put(value)
		}
	}
	for _, value := range m.hiddenBias {
		put(value)
	}
	for _, row := range m.output {
		for _, head := range row {
			for _, value := range head {
				put(value)
			}
		}
	}
	for _, head := range m.outputBias {
		for _, value := range head {
			put(value)
		}
	}
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

// Load reads at most ArtifactBytes+1 bytes; malicious dimensions or counts
// never control allocations, payload length or iteration bounds.
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
	if string(data[:4]) != "RCM\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion || binary.LittleEndian.Uint16(data[6:8]) != artifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureBins || binary.LittleEndian.Uint16(data[10:12]) != HiddenUnits || data[12] != HeadCount || data[13] != StateCount || data[14] != 1 || data[15] != 1 {
		return nil, ErrArtifact
	}
	feature, order, contract := featureHash(), orderHash(), numericContractHash()
	var reserved [28]byte
	if !bytes.Equal(data[32:64], feature[:]) || !bytes.Equal(data[64:96], order[:]) || !bytes.Equal(data[96:128], contract[:]) || binary.LittleEndian.Uint64(data[128:136]) != math.Float64bits(ConfidenceFloor) || binary.LittleEndian.Uint64(data[136:144]) != math.Float64bits(MarginFloor) || binary.LittleEndian.Uint64(data[144:152]) != math.Float64bits(1) || binary.LittleEndian.Uint32(data[156:160]) != ParameterCount || binary.LittleEndian.Uint32(data[160:164]) != WeightCount || !bytes.Equal(data[164:192], reserved[:]) {
		return nil, ErrArtifact
	}
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	if !bytes.Equal(data[end:], sum[:]) {
		return nil, ErrArtifact
	}
	steps := binary.LittleEndian.Uint64(data[16:24])
	seed := int64(binary.LittleEndian.Uint64(data[24:32]))
	samples := binary.LittleEndian.Uint32(data[152:156])
	if !validHistory(steps, samples, seed) {
		return nil, ErrArtifact
	}
	m := &Model{steps: steps, seed: seed, samples: samples}
	offset := artifactHeaderBytes
	get := func() float32 {
		value := math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
		return value
	}
	for i := range m.input {
		for h := range m.input[i] {
			m.input[i][h] = get()
		}
	}
	for h := range m.hiddenBias {
		m.hiddenBias[h] = get()
	}
	for h := range m.output {
		for head := range m.output[h] {
			for c := range m.output[h][head] {
				m.output[h][head][c] = get()
			}
		}
	}
	for head := range m.outputBias {
		for c := range m.outputBias[head] {
			m.outputBias[head][c] = get()
		}
	}
	if !m.valid() {
		return nil, ErrArtifact
	}
	return m, nil
}
