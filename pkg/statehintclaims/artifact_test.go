// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"testing"
)

func checksum(data []byte) {
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	copy(data[end:], sum[:])
}
func TestTrainedArtifactExactlyRoundtripsScoresPredictionsAndBytes(t *testing.T) {
	m := NewModel()
	var training TrainingWorkspace
	if _, err := m.Fit(ownedSamples(), FitOptions{Epochs: 3, BatchSize: 2}, &training); err != nil {
		t.Fatal(err)
	}
	data := saved(t, m)
	if ArtifactBytes != 73988 || ParameterCount != 18441 {
		t.Fatal("distinct bounded format constants")
	}
	if len(data) != ArtifactBytes {
		t.Fatal("distinct bounded format dimensions", len(data))
	}
	loaded, err := Load(bytes.NewReader(data))
	if err != nil || !bytes.Equal(data, saved(t, loaded)) || loaded.TrainingSteps() != 6 || loaded.InitializationSeed() != DefaultSeed || loaded.Temperature() != 1 {
		t.Fatal("trained artifact/history byte drift", err)
	}
	for _, text := range []string{"Owned inference fixture with a current claim.", "직접 만든 한글 추론 자료입니다.", "..."} {
		var a, b Workspace
		pa, ea := m.Predict(text, &a)
		pb, eb := loaded.Predict(text, &b)
		sa, esa := m.Scores(text, &a)
		sb, esb := loaded.Scores(text, &b)
		if ea != nil || eb != nil || esa != nil || esb != nil || pa != pb || sa != sb {
			t.Fatal("full prediction/score reload parity", text)
		}
	}
}

func TestArtifactRejectsCorruptionUnknownContractsNonfiniteAndTrailingBytes(t *testing.T) {
	original := saved(t, NewModel())
	for _, tc := range []struct {
		name   string
		change func([]byte)
	}{
		{"legacy-rsh", func(b []byte) { copy(b[:4], "RSH\x00") }},
		{"legacy-rsm", func(b []byte) { copy(b[:4], "RSM\x00") }},
		{"legacy-rsp", func(b []byte) { copy(b[:4], "RSP\x00") }},
		{"unknown-version", func(b []byte) { b[4]++ }},
		{"unknown-schema", func(b []byte) { b[32] ^= 1 }},
		{"different-order", func(b []byte) { b[96] ^= 1 }},
		{"different-objective", func(b []byte) { b[128] ^= 1 }},
		{"nonfinite-weight", func(b []byte) { binary.LittleEndian.PutUint32(b[192:196], math.Float32bits(float32(math.NaN()))) }},
		{"nonfinite-bias", func(b []byte) {
			end := len(b) - sha256.Size
			binary.LittleEndian.PutUint32(b[end-4:end], math.Float32bits(float32(math.Inf(1))))
		}},
		{"changed-confidence", func(b []byte) { binary.LittleEndian.PutUint64(b[160:168], math.Float64bits(.8)) }},
		{"changed-temperature", func(b []byte) { binary.LittleEndian.PutUint64(b[176:184], math.Float64bits(.65)) }},
		{"reserved", func(b []byte) { b[184] = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := append([]byte(nil), original...)
			tc.change(bad)
			checksum(bad) // These must fail even with a valid checksum.
			if _, err := Load(bytes.NewReader(bad)); err != ErrArtifact {
				t.Fatal("invalid contract accepted", err)
			}
		})
	}
	corrupt := append([]byte(nil), original...)
	corrupt[192] ^= 1
	for _, bad := range [][]byte{corrupt, original[:len(original)-1], append(append([]byte(nil), original...), 0)} {
		if _, err := Load(bytes.NewReader(bad)); err != ErrArtifact {
			t.Fatal("corrupt/truncated/trailing artifact accepted")
		}
	}
	reader := bytes.NewReader(append(append([]byte(nil), original...), bytes.Repeat([]byte{1}, 100)...))
	if _, err := Load(reader); err != ErrArtifact || reader.Len() != 99 {
		t.Fatal("reader consumed beyond one trailing byte")
	}
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestInvalidAndShortArtifactWriters(t *testing.T) {
	m := NewModel()
	if m.Save(shortWriter{}) != io.ErrShortWrite || m.Save(nil) != ErrModel {
		t.Fatal("invalid/partial artifact write accepted")
	}
	m.weights[0][0][0] = float32(math.Inf(1))
	var b bytes.Buffer
	if m.Save(&b) != ErrModel || b.Len() != 0 {
		t.Fatal("nonfinite model wrote partial output")
	}
	if _, err := Load(nil); err != ErrArtifact {
		t.Fatal("nil artifact reader accepted")
	}
}
