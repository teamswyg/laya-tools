package statehint

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
)

const (
	ArtifactVersion     = 1
	artifactHeaderBytes = 128
	ArtifactBytes       = artifactHeaderBytes + (FeatureBins*IntentCount+IntentCount)*4 + sha256.Size
)

func intentOrderHash() [sha256.Size]byte {
	var encoded [256]byte
	count := 0
	for _, intent := range Intents() {
		count += copy(encoded[count:], intent)
		encoded[count] = 0
		count++
	}
	return sha256.Sum256(encoded[:count])
}

// Save writes one bounded .rsh artifact, including feature schema, class order,
// temperature and SHA-256 corruption detection. Model binaries are local
// generated artifacts and must not be committed to the source repository.
func (m *Model) Save(writer io.Writer) error {
	if writer == nil || !m.valid() {
		return ErrModel
	}
	var data [ArtifactBytes]byte
	copy(data[:4], "RSH\x00")
	binary.LittleEndian.PutUint16(data[4:6], ArtifactVersion)
	binary.LittleEndian.PutUint16(data[6:8], artifactHeaderBytes)
	binary.LittleEndian.PutUint16(data[8:10], FeatureBins)
	binary.LittleEndian.PutUint16(data[10:12], IntentCount)
	binary.LittleEndian.PutUint64(data[16:24], m.steps)
	binary.LittleEndian.PutUint64(data[24:32], math.Float64bits(m.temperature))
	copy(data[32:96], FeatureSchema)
	orderHash := intentOrderHash()
	copy(data[96:128], orderHash[:])
	offset := artifactHeaderBytes
	for _, row := range m.weights {
		for _, weight := range row {
			binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(weight))
			offset += 4
		}
	}
	for _, bias := range m.bias {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(bias))
		offset += 4
	}
	digest := sha256.Sum256(data[:offset])
	copy(data[offset:], digest[:])
	n, err := writer.Write(data[:])
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

// Load consumes exactly one artifact. Truncation, trailing bytes, schema/order
// drift, nonfinite weights, invalid temperature and corrupt checksums fail
// closed. The checksum detects corruption; it does not authenticate a model.
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
	if string(data[:4]) != "RSH\x00" || binary.LittleEndian.Uint16(data[4:6]) != ArtifactVersion ||
		binary.LittleEndian.Uint16(data[6:8]) != artifactHeaderBytes || binary.LittleEndian.Uint16(data[8:10]) != FeatureBins ||
		binary.LittleEndian.Uint16(data[10:12]) != IntentCount || binary.LittleEndian.Uint32(data[12:16]) != 0 {
		return nil, ErrArtifact
	}
	var featureSchema [64]byte
	copy(featureSchema[:], FeatureSchema)
	orderHash := intentOrderHash()
	if !bytes.Equal(data[32:96], featureSchema[:]) || !bytes.Equal(data[96:128], orderHash[:]) {
		return nil, ErrArtifact
	}
	payloadEnd := len(data) - sha256.Size
	digest := sha256.Sum256(data[:payloadEnd])
	if !bytes.Equal(data[payloadEnd:], digest[:]) {
		return nil, ErrArtifact
	}
	model := &Model{temperature: math.Float64frombits(binary.LittleEndian.Uint64(data[24:32])), steps: binary.LittleEndian.Uint64(data[16:24])}
	offset := artifactHeaderBytes
	for index := range model.weights {
		for class := range model.weights[index] {
			model.weights[index][class] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
			offset += 4
		}
	}
	for class := range model.bias {
		model.bias[class] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}
	if !model.valid() {
		return nil, ErrArtifact
	}
	return model, nil
}
