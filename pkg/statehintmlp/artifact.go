// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
)

const (
	ArtifactVersion     = 3
	artifactHeaderBytes = 160
	ArtifactBytes       = artifactHeaderBytes + ParameterCount*4 + sha256.Size
)

func intentOrderHash() [32]byte {
	var b [256]byte
	n := 0
	for _, i := range Intents() {
		n += copy(b[n:], i)
		b[n] = 0
		n++
	}
	return sha256.Sum256(b[:n])
}

// Save writes a distinct RSM v3 artifact with fixed architecture, initialization
// seed, feature/class schemas, temperature, steps and corruption checksum.
// It is neither an authenticated asset nor compatible with v1/v2 loaders.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RSM\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], artifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureBins)
	binary.LittleEndian.PutUint16(data[10:12], HiddenUnits)
	binary.LittleEndian.PutUint16(data[12:14], IntentCount)
	data[14], data[15] = 1, 1 // ReLU, Glorot-PCG-v3
	binary.LittleEndian.PutUint64(data[16:24], m.steps)
	binary.LittleEndian.PutUint64(data[24:32], math.Float64bits(m.temperature))
	binary.LittleEndian.PutUint64(data[32:40], uint64(m.seed))
	copy(data[40:104], FeatureSchema)
	order := intentOrderHash()
	copy(data[104:136], order[:])
	offset := artifactHeaderBytes
	put := func(v float32) {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(v))
		offset += 4
	}
	for _, row := range m.input {
		for _, v := range row {
			put(v)
		}
	}
	for _, v := range m.hiddenBias {
		put(v)
	}
	for _, row := range m.output {
		for _, v := range row {
			put(v)
		}
	}
	for _, v := range m.outputBias {
		put(v)
	}
	sum := sha256.Sum256(data[:offset])
	copy(data[offset:], sum[:])
	n, e := writer.Write(data[:])
	if e != nil {
		return e
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
	if _, e := io.ReadFull(reader, data[:]); e != nil {
		return nil, ErrArtifact
	}
	var trailing [1]byte
	if n, e := io.ReadFull(reader, trailing[:]); n != 0 || e != io.EOF {
		return nil, ErrArtifact
	}
	if string(data[:4]) != "RSM\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion || binary.LittleEndian.Uint16(data[6:8]) != artifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureBins || binary.LittleEndian.Uint16(data[10:12]) != HiddenUnits || binary.LittleEndian.Uint16(data[12:14]) != IntentCount || data[14] != 1 || data[15] != 1 {
		return nil, ErrArtifact
	}
	var schema [64]byte
	copy(schema[:], FeatureSchema)
	order := intentOrderHash()
	var reserved [24]byte
	if !bytes.Equal(data[40:104], schema[:]) || !bytes.Equal(data[104:136], order[:]) || !bytes.Equal(data[136:160], reserved[:]) {
		return nil, ErrArtifact
	}
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	if !bytes.Equal(data[end:], sum[:]) {
		return nil, ErrArtifact
	}
	m := &Model{steps: binary.LittleEndian.Uint64(data[16:24]), temperature: math.Float64frombits(binary.LittleEndian.Uint64(data[24:32])), seed: int64(binary.LittleEndian.Uint64(data[32:40]))}
	offset := artifactHeaderBytes
	get := func() float32 {
		v := math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
		return v
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
		for c := range m.output[h] {
			m.output[h][c] = get()
		}
	}
	for c := range m.outputBias {
		m.outputBias[c] = get()
	}
	if !m.valid() {
		return nil, ErrArtifact
	}
	return m, nil
}
