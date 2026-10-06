package statehint

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"
)

func serializedFixture(t *testing.T) []byte {
	t.Helper()
	m := NewModel()
	m.weights[123][2] = .75 // visibly synthetic serialization fixture
	m.bias[7] = -.1
	m.steps = 23
	if err := m.SetTemperature(1.7); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := m.Save(&b); err != nil {
		t.Fatal(err)
	}
	if b.Len() != ArtifactBytes {
		t.Fatal("artifact length drift")
	}
	return b.Bytes()
}

func TestArtifactExactRoundTripAndCorruption(t *testing.T) {
	data := serializedFixture(t)
	m, err := Load(bytes.NewReader(data))
	if err != nil || m.weights[123][2] != .75 || m.TrainingSteps() != 23 || m.Temperature() != 1.7 {
		t.Fatal(m, err)
	}
	var saved bytes.Buffer
	if err := m.Save(&saved); err != nil || !bytes.Equal(data, saved.Bytes()) {
		t.Fatal("artifact round trip drift", err)
	}
	bad := [][]byte{data[:len(data)-1], append(append([]byte(nil), data...), 0), {}, append([]byte(nil), data...)}
	bad[3][150] ^= 1
	for _, value := range bad {
		if _, err := Load(bytes.NewReader(value)); !errors.Is(err, ErrArtifact) {
			t.Fatal("corrupt artifact accepted", err)
		}
	}
}

func TestArtifactSemanticValidationWithCorrectChecksum(t *testing.T) {
	for _, mutate := range []func([]byte){
		func(b []byte) { b[4] = 2 },
		func(b []byte) { b[8] = 1 },
		func(b []byte) { b[12] = 1 },
		func(b []byte) { b[32] ^= 1 },
		func(b []byte) { b[96] ^= 1 },
		func(b []byte) { binary.LittleEndian.PutUint64(b[24:32], math.Float64bits(math.Inf(1))) },
		func(b []byte) { binary.LittleEndian.PutUint32(b[128:132], math.Float32bits(float32(math.NaN()))) },
		func(b []byte) {
			binary.LittleEndian.PutUint32(b[len(b)-36:len(b)-32], math.Float32bits(float32(math.Inf(-1))))
		},
	} {
		data := append([]byte(nil), serializedFixture(t)...)
		mutate(data)
		digest := sha256.Sum256(data[:len(data)-sha256.Size])
		copy(data[len(data)-sha256.Size:], digest[:])
		if _, err := Load(bytes.NewReader(data)); !errors.Is(err, ErrArtifact) {
			t.Fatal("valid checksum bypassed semantic validation")
		}
	}
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestArtifactInvalidWriterAndModel(t *testing.T) {
	if err := NewModel().Save(shortWriter{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
	m := NewModel()
	m.weights[0][0] = float32(math.Inf(1))
	if err := m.Save(&bytes.Buffer{}); !errors.Is(err, ErrModel) {
		t.Fatal("nonfinite weights saved")
	}
	if _, err := Load(nil); !errors.Is(err, ErrArtifact) {
		t.Fatal(err)
	}
}
