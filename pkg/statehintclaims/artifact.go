// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

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

func orderHash() [sha256.Size]byte {
	var b bytes.Buffer
	for _, h := range Heads() {
		b.WriteString(h.String())
		b.WriteByte(0)
	}
	for _, s := range States() {
		b.WriteString(string(s))
		b.WriteByte(0)
	}
	return sha256.Sum256(b.Bytes())
}
func numericContractHash() [sha256.Size]byte {
	return sha256.Sum256([]byte("feature-major-nine-columns;mean-three-categorical-ce;fresh-zero-adamw;bias-no-decay;fixed-T1;v1"))
}

// RSC is a distinct three-claim format. It cannot silently load RSH/RSM/RSP
// models. SHA-256 detects corruption, not publisher authenticity. No optimizer
// or task authority is serialized. Float32 order is feature/head/state, then
// head/state biases; prediction temperature and numeric floors are fixed.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RSC\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], artifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureBins)
	data[10], data[11], data[12] = HeadCount, StateCount, 1
	binary.LittleEndian.PutUint64(data[16:24], m.steps)
	binary.LittleEndian.PutUint64(data[24:32], uint64(m.seed))
	copy(data[32:96], FeatureSchema)
	order, contract := orderHash(), numericContractHash()
	copy(data[96:128], order[:])
	copy(data[128:160], contract[:])
	binary.LittleEndian.PutUint64(data[160:168], math.Float64bits(ConfidenceFloor))
	binary.LittleEndian.PutUint64(data[168:176], math.Float64bits(MarginFloor))
	binary.LittleEndian.PutUint64(data[176:184], math.Float64bits(1))
	offset := artifactHeaderBytes
	for _, feature := range m.weights {
		for _, head := range feature {
			for _, value := range head {
				binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(value))
				offset += 4
			}
		}
	}
	for _, head := range m.bias {
		for _, value := range head {
			binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(value))
			offset += 4
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
	var flagsReserved [3]byte
	if string(data[:4]) != "RSC\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion || binary.LittleEndian.Uint16(data[6:8]) != artifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureBins || data[10] != HeadCount || data[11] != StateCount || data[12] != 1 || !bytes.Equal(data[13:16], flagsReserved[:]) {
		return nil, ErrArtifact
	}
	var schema [64]byte
	copy(schema[:], FeatureSchema)
	order, contract := orderHash(), numericContractHash()
	var reserved [8]byte
	if !bytes.Equal(data[32:96], schema[:]) || !bytes.Equal(data[96:128], order[:]) || !bytes.Equal(data[128:160], contract[:]) || binary.LittleEndian.Uint64(data[160:168]) != math.Float64bits(ConfidenceFloor) || binary.LittleEndian.Uint64(data[168:176]) != math.Float64bits(MarginFloor) || binary.LittleEndian.Uint64(data[176:184]) != math.Float64bits(1) || !bytes.Equal(data[184:192], reserved[:]) {
		return nil, ErrArtifact
	}
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	if !bytes.Equal(data[end:], sum[:]) {
		return nil, ErrArtifact
	}
	m := &Model{steps: binary.LittleEndian.Uint64(data[16:24]), seed: int64(binary.LittleEndian.Uint64(data[24:32]))}
	offset := artifactHeaderBytes
	for i := range m.weights {
		for h := range m.weights[i] {
			for c := range m.weights[i][h] {
				m.weights[i][h][c] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
				offset += 4
			}
		}
	}
	for h := range m.bias {
		for c := range m.bias[h] {
			m.bias[h][c] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
			offset += 4
		}
	}
	if !m.valid() {
		return nil, ErrArtifact
	}
	return m, nil
}
