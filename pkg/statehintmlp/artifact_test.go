// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

func checksum(data []byte) {
	sum := sha256.Sum256(data[:len(data)-32])
	copy(data[len(data)-32:], sum[:])
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }

type countedReader struct {
	r *bytes.Reader
	n int
}

func (c *countedReader) Read(b []byte) (int, error) { n, e := c.r.Read(b); c.n += n; return n, e }
func TestTrainedArtifactExactPredictionAndByteParity(t *testing.T) {
	m := NewModel()
	ss := toySamples()
	if _, e := m.Fit(ss, FitOptions{Epochs: 4, BatchSize: 8, Seed: 73}); e != nil {
		t.Fatal(e)
	}
	if m.SetTemperature(1.25) != nil {
		t.Fatal("temperature")
	}
	data := saved(t, m)
	if len(data) != ArtifactBytes || ArtifactBytes != 131872 || ParameterCount != 32920 {
		t.Fatal("fixed capacity changed")
	}
	q, e := Load(bytes.NewReader(data))
	if e != nil || q.Temperature() != m.Temperature() || q.InitializationSeed() != 73 || q.TrainingSteps() != 12 || !bytes.Equal(saved(t, q), data) {
		t.Fatal("trained artifact state or byte parity", e)
	}
	var a, b Workspace
	for _, s := range append(ss, Sample{Text: "Original fresh 한국어 fixture", Label: Reference}) {
		p, e := m.Predict(s.Text, &a)
		if e != nil {
			t.Fatal(e)
		}
		r, e := q.Predict(s.Text, &b)
		if e != nil || p != r || r.Source != Learned {
			t.Fatal("trained predict reload drift")
		}
	}
	if e := m.Save(shortWriter{}); e != io.ErrShortWrite {
		t.Fatal("short artifact write ignored", e)
	}
	if _, e := statehintwide.Load(bytes.NewReader(data)); e == nil {
		t.Fatal("v2 accepted v3")
	}
	if _, e := statehint.Load(bytes.NewReader(data)); e == nil {
		t.Fatal("v1 accepted v3")
	}
	var old bytes.Buffer
	if statehintwide.NewModel(statehintwide.Contextual).Save(&old) != nil {
		t.Fatal("v2 fixture")
	}
	if _, e := Load(&old); e == nil {
		t.Fatal("v3 silently accepted v2")
	}
}
func TestArtifactMalformedFiniteSchemaVersionAndBoundChecks(t *testing.T) {
	data := saved(t, NewModel())
	for _, bad := range [][]byte{data[:len(data)-1], append(append([]byte(nil), data...), 0)} {
		if _, e := Load(bytes.NewReader(bad)); e == nil {
			t.Fatal("truncated/trailing artifact accepted")
		}
	}
	changes := []func([]byte){func(b []byte) { b[0] ^= 1 }, func(b []byte) { binary.LittleEndian.PutUint16(b[4:6], 2) }, func(b []byte) { binary.LittleEndian.PutUint16(b[10:12], 32) }, func(b []byte) { b[14] = 2 }, func(b []byte) { b[15] = 2 }, func(b []byte) { b[40] ^= 1 }, func(b []byte) { b[104] ^= 1 }, func(b []byte) { b[136] = 1 }, func(b []byte) { binary.LittleEndian.PutUint64(b[24:32], math.Float64bits(math.NaN())) }, func(b []byte) { binary.LittleEndian.PutUint64(b[24:32], 0) }, func(b []byte) {
		binary.LittleEndian.PutUint32(b[artifactHeaderBytes:artifactHeaderBytes+4], math.Float32bits(float32(math.Inf(1))))
	}}
	for i, change := range changes {
		bad := append([]byte(nil), data...)
		change(bad)
		checksum(bad)
		if _, e := Load(bytes.NewReader(bad)); e == nil {
			t.Fatalf("malformed rechecksummed artifact%d accepted", i)
		}
	}
	bad := append([]byte(nil), data...)
	bad[len(bad)-1] ^= 1
	if _, e := Load(bytes.NewReader(bad)); e == nil {
		t.Fatal("corruption accepted")
	}
	reader := &countedReader{r: bytes.NewReader(append(append([]byte(nil), data...), make([]byte, 1<<20)...))}
	if _, e := Load(reader); e == nil || reader.n != ArtifactBytes+1 {
		t.Fatal("artifact read beyond fixed budget", reader.n)
	}
	if _, e := Load(nil); e == nil {
		t.Fatal("nil reader accepted")
	}
	m := NewModel()
	m.input[0][0] = float32(math.NaN())
	if m.Save(io.Discard) == nil {
		t.Fatal("nonfinite model saved")
	}
}
