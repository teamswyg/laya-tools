// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"unsafe"
)

func repairDigest(data []byte) {
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	copy(data[end:], sum[:])
}

func TestTrainedArtifactRoundTripExactBytesAndProbabilities(t *testing.T) {
	parent, m := syntheticPTQ(t)
	parentHexBacking := unsafe.StringData(m.parentSHAHex)
	var workspace TrainingWorkspace
	if _, err := m.WarmFit(parent, syntheticSamples(), &workspace); err != nil {
		t.Fatal(err)
	}
	if unsafe.StringData(m.parentSHAHex) != parentHexBacking || unsafe.StringData(workspace.forward.parentSHAHex) != parentHexBacking {
		t.Fatal("continuation recreated immutable parent digest text")
	}
	data := modelBytes(t, m)
	if len(data) != 4023 || ArtifactBytes != 4023 || PackedWeightBytes != 3687 || WeightTritCount != 18432 || string(data[:4]) != "RQT\x00" || binary.LittleEndian.Uint16(data[6:8]) != 256 {
		t.Fatalf("layout: %d", len(data))
	}
	loaded, err := Load(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if *m != *loaded || !bytes.Equal(data, modelBytes(t, loaded)) || m.Metadata() != loaded.Metadata() {
		t.Fatal("roundtrip changed packed parameters, metadata or bytes")
	}
	for _, text := range []string{"amber kite waits", "violet boat rests", "...", "새 구름 쉼"} {
		want, err := m.Scores(text, &Workspace{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := loaded.Scores(text, &Workspace{})
		if err != nil || got != want {
			t.Fatalf("score parity %q: %+v %+v %v", text, want, got, err)
		}
		wantPrediction, err := m.Predict(text, &Workspace{})
		if err != nil {
			t.Fatal(err)
		}
		gotPrediction, err := loaded.Predict(text, &Workspace{})
		if err != nil || gotPrediction != wantPrediction {
			t.Fatalf("predict parity %q: %+v %+v %v", text, wantPrediction, gotPrediction, err)
		}
	}
}

func TestParentSHAHexCacheIsDerivedAndNotSerialized(t *testing.T) {
	_, m := syntheticPTQ(t)
	data := modelBytes(t, m)
	copy := m.Clone()
	copy.parentSHAHex = strings.Clone(m.parentSHAHex)
	if !bytes.Equal(data, modelBytes(t, copy)) {
		t.Fatal("digest text backing altered artifact bytes")
	}
	loaded, err := Load(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.parentSHAHex != m.parentSHAHex || !loaded.parentSHAHexMatches() || !bytes.Equal(data[168:200], loaded.parentSHA[:]) {
		t.Fatal("loader did not derive digest text from the serialized raw digest")
	}
	for _, mutate := range []func(*Model){
		func(model *Model) { model.parentSHAHex = "" },
		func(model *Model) { model.parentSHAHex = strings.ToUpper(model.parentSHAHex) },
		func(model *Model) { model.parentSHA[0] ^= 1 },
	} {
		invalid := m.Clone()
		mutate(invalid)
		var writer bytes.Buffer
		if invalid.valid() || invalid.Save(&writer) != ErrModel || writer.Len() != 0 {
			t.Fatal("accepted inconsistent private digest cache")
		}
	}
}

func TestArtifactRejectsNoncanonicalOrCorruptedData(t *testing.T) {
	_, m := syntheticPTQ(t)
	data := modelBytes(t, m)
	for _, tt := range []struct {
		name   string
		mutate func([]byte)
	}{
		{"magic", func(d []byte) { d[0] = 'S' }},
		{"version", func(d []byte) { binary.LittleEndian.PutUint16(d[4:6], 2) }},
		{"header_size", func(d []byte) { binary.LittleEndian.PutUint16(d[6:8], 192) }},
		{"bins", func(d []byte) { binary.LittleEndian.PutUint16(d[8:10], 1024) }},
		{"heads", func(d []byte) { d[10] = 2 }},
		{"states", func(d []byte) { d[11] = 2 }},
		{"mode", func(d []byte) { d[12] = 3 }},
		{"reserved_flags", func(d []byte) { d[13] = 1 }},
		{"ptq_optimizer_steps", func(d []byte) { binary.LittleEndian.PutUint64(d[24:32], 1) }},
		{"qat_without_steps", func(d []byte) { d[12] = byte(QAT) }},
		{"qat_nonrecipe_steps", func(d []byte) { d[12] = byte(QAT); binary.LittleEndian.PutUint64(d[24:32], 1) }},
		{"step_overflow", func(d []byte) {
			d[12] = byte(QAT)
			binary.LittleEndian.PutUint64(d[16:24], math.MaxUint64)
			binary.LittleEndian.PutUint64(d[24:32], 40)
		}},
		{"feature_schema", func(d []byte) { d[40] ^= 1 }},
		{"feature_padding", func(d []byte) { d[103] = 1 }},
		{"head_state_order", func(d []byte) { d[104] ^= 1 }},
		{"numeric_contract", func(d []byte) { d[136] ^= 1 }},
		{"confidence_floor", func(d []byte) { binary.LittleEndian.PutUint64(d[200:208], math.Float64bits(.8)) }},
		{"margin_floor", func(d []byte) { binary.LittleEndian.PutUint64(d[208:216], math.Float64bits(.04)) }},
		{"temperature", func(d []byte) { binary.LittleEndian.PutUint64(d[216:224], math.Float64bits(2)) }},
		{"scale_floor", func(d []byte) { binary.LittleEndian.PutUint64(d[224:232], math.Float64bits(1e-6)) }},
		{"reserved_tail", func(d []byte) { d[255] = 1 }},
		{"scale_nan", func(d []byte) { binary.LittleEndian.PutUint32(d[256:260], math.Float32bits(float32(math.NaN()))) }},
		{"scale_inf", func(d []byte) { binary.LittleEndian.PutUint32(d[260:264], math.Float32bits(float32(math.Inf(1)))) }},
		{"scale_zero", func(d []byte) { binary.LittleEndian.PutUint32(d[264:268], 0) }},
		{"scale_below_floor", func(d []byte) { binary.LittleEndian.PutUint32(d[256:260], math.Float32bits(float32(ScaleFloor/2))) }},
		{"bias_nan", func(d []byte) { binary.LittleEndian.PutUint32(d[268:272], math.Float32bits(float32(math.NaN()))) }},
		{"bias_inf", func(d []byte) { binary.LittleEndian.PutUint32(d[300:304], math.Float32bits(float32(math.Inf(-1)))) }},
		{"invalid_packed_byte", func(d []byte) { d[304] = 243 }},
		{"noncanonical_trit_padding", func(d []byte) { d[3990] -= 9 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bad := bytes.Clone(data)
			tt.mutate(bad)
			repairDigest(bad) // Must reject contract violation even with valid SHA.
			if _, err := Load(bytes.NewReader(bad)); err == nil {
				t.Fatal("accepted invalid canonical data")
			}
		})
	}
	corrupted := bytes.Clone(data)
	corrupted[304] ^= 1
	if _, err := Load(bytes.NewReader(corrupted)); err == nil {
		t.Fatal("accepted damaged digest")
	}
	for _, bad := range [][]byte{data[:0], data[:255], data[:len(data)-1], append(bytes.Clone(data), 0)} {
		if _, err := Load(bytes.NewReader(bad)); err == nil {
			t.Fatal("accepted truncation or trailing bytes")
		}
	}
	if _, err := Load(nil); err == nil {
		t.Fatal("accepted nil reader")
	}
}

type failingWriter struct{ short bool }

func (w failingWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, errors.New("owned writer failure")
}

func TestSaveWriterFailureAndInvalidModel(t *testing.T) {
	_, m := syntheticPTQ(t)
	if err := m.Save(failingWriter{short: true}); err != io.ErrShortWrite {
		t.Fatalf("short write: %v", err)
	}
	if err := m.Save(failingWriter{}); err == nil || err == ErrModel {
		t.Fatalf("writer error: %v", err)
	}
	if err := m.Save(nil); err == nil {
		t.Fatal("accepted nil writer")
	}
	var nilModel *Model
	if err := nilModel.Save(&bytes.Buffer{}); err == nil {
		t.Fatal("accepted nil model")
	}
	m.bias[0][0] = float32(math.NaN())
	if err := m.Save(&bytes.Buffer{}); err == nil {
		t.Fatal("saved nonfinite model")
	}
}
