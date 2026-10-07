// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintsharedprobe

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"strings"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const (
	ArtifactVersion     = 1
	artifactHeaderBytes = 192
	ArtifactBytes       = artifactHeaderBytes + FeatureDimensions*4 + sha256.Size
)

func orderHash() [32]byte {
	var b bytes.Buffer
	for _, intent := range statehint.Intents() {
		b.WriteString(string(intent))
		b.WriteByte(0)
	}
	return sha256.Sum256(b.Bytes())
}
func contractHash() [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{BaseSHA, InstructionSHA, FeatureSchema}, "\x00")))
}

// RSP is a distinct feature-only head format, incompatible with all text-model
// RSH/RSM loaders. Checksums detect corruption, not publisher authentication.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RSP\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], artifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureDimensions)
	binary.LittleEndian.PutUint16(data[10:12], IntentCount)
	data[12], data[13] = 1, 1
	binary.LittleEndian.PutUint64(data[16:24], m.steps)
	binary.LittleEndian.PutUint64(data[24:32], math.Float64bits(m.temperature))
	binary.LittleEndian.PutUint64(data[32:40], uint64(m.seed))
	copy(data[40:104], FeatureSchema)
	order, contract := orderHash(), contractHash()
	copy(data[104:136], order[:])
	copy(data[136:168], contract[:])
	offset := artifactHeaderBytes
	for _, v := range m.weights {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(v))
		offset += 4
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
	if string(data[:4]) != "RSP\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion || binary.LittleEndian.Uint16(data[6:8]) != artifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureDimensions || binary.LittleEndian.Uint16(data[10:12]) != IntentCount || data[12] != 1 || data[13] != 1 || binary.LittleEndian.Uint16(data[14:16]) != 0 {
		return nil, ErrArtifact
	}
	var schema [64]byte
	copy(schema[:], FeatureSchema)
	order, contract := orderHash(), contractHash()
	var reserved [24]byte
	if !bytes.Equal(data[40:104], schema[:]) || !bytes.Equal(data[104:136], order[:]) || !bytes.Equal(data[136:168], contract[:]) || !bytes.Equal(data[168:192], reserved[:]) {
		return nil, ErrArtifact
	}
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	if !bytes.Equal(data[end:], sum[:]) {
		return nil, ErrArtifact
	}
	m := &Model{steps: binary.LittleEndian.Uint64(data[16:24]), temperature: math.Float64frombits(binary.LittleEndian.Uint64(data[24:32])), seed: int64(binary.LittleEndian.Uint64(data[32:40]))}
	offset := artifactHeaderBytes
	for i := range m.weights {
		m.weights[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}
	if !m.valid() {
		return nil, ErrArtifact
	}
	return m, nil
}
