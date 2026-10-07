// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintsharedprobe

import (
	"bytes"
	"encoding/binary"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintmlp"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"io"
	"math"
	"reflect"
	"testing"
)

func originalRow() Row {
	var r Row
	for option := range r {
		for feature := 0; feature < 6; feature++ {
			r[option][feature] = float32((option+1)*(feature-2)) * .03
		}
	}
	return r
}
func rowBytes(rows []Row) []byte {
	b := make([]byte, len(rows)*RowBytes)
	offset := 0
	for _, r := range rows {
		for _, option := range r {
			for _, v := range option {
				binary.LittleEndian.PutUint32(b[offset:offset+4], math.Float32bits(v))
				offset += 4
			}
		}
	}
	return b
}
func lossFor(m *Model, r *Row, target int) float64 {
	z := m.logits(r)
	maximum := z[0]
	for _, v := range z {
		maximum = math.Max(maximum, v)
	}
	sum := 0.0
	for _, v := range z {
		sum += math.Exp(v - maximum)
	}
	return maximum - z[target] + math.Log(sum)
}

func TestSharedCEGradientMatchesFiniteDifferenceAndBiasInvariant(t *testing.T) {
	r := originalRow()
	m := NewModel()
	for i := 0; i < 6; i++ {
		m.weights[i] = float32(i-3) * .01
	}
	var g [FeatureDimensions]float64
	loss, e := m.accumulate(&r, 4, &g)
	if e != nil || math.Abs(loss-lossFor(m, &r, 4)) > 1e-12 {
		t.Fatal("shared loss")
	}
	for feature := 0; feature < 6; feature++ {
		before := m.weights[feature]
		epsilon := float32(.001)
		m.weights[feature] = before + epsilon
		high := lossFor(m, &r, 4)
		hi := m.weights[feature]
		m.weights[feature] = before - epsilon
		low := lossFor(m, &r, 4)
		lo := m.weights[feature]
		m.weights[feature] = before
		numeric := (high - low) / float64(hi-lo)
		if math.Abs(numeric-g[feature]) > 1e-5 {
			t.Fatalf("gradient%d numeric%g analytic%g", feature, numeric, g[feature])
		}
	}
	z := m.logits(&r)
	p, ok := softmax(z, 1)
	if !ok {
		t.Fatal("softmax")
	}
	for i := range z {
		z[i] += 17
	}
	shifted, ok := softmax(z, 1)
	if !ok {
		t.Fatal("shifted")
	}
	for i := range p {
		if math.Abs(p[i]-shifted[i]) > 1e-14 {
			t.Fatal("shared bias changed probabilities")
		}
	}
	for i := 6; i < FeatureDimensions; i++ {
		if g[i] != 0 {
			t.Fatal("zero feature derivative")
		}
	}
}

func TestKnownFirstAdamWUpdate(t *testing.T) {
	m := NewModel()
	m.weights[0] = .5
	var a optimizer
	a.gradient[0] = 4
	o := FitOptions{LearningRate: .001, WeightDecay: .01}
	if a.update(m, o, 2, 1) != nil {
		t.Fatal("update")
	}
	want := float32(.5*(1-.001*.01) - .001*2/(2+1e-8))
	if m.weights[0] != want || math.Abs(a.first[0]-.2) > 1e-15 || math.Abs(a.second[0]-.004) > 1e-15 {
		t.Fatal("bias correction/batch mean/decay", m.weights[0], a.first[0], a.second[0])
	}
}

type countingReader struct {
	reader *bytes.Reader
	calls  int
	max    int
}

func (c *countingReader) ReadAt(p []byte, offset int64) (int, error) {
	c.calls++
	if len(p) > c.max {
		c.max = len(p)
	}
	return c.reader.ReadAt(p, offset)
}
func TestActualToyFitFreshTransactionAndTrainedArtifactParity(t *testing.T) {
	rows := []Row{originalRow(), originalRow()}
	rows[1][3][0] = 2
	data := rowBytes(rows)
	reader := &countingReader{reader: bytes.NewReader(data)}
	store, e := NewStore(reader, 2)
	if e != nil {
		t.Fatal(e)
	}
	samples := []Sample{{0, statehint.CompletionReport}, {1, statehint.Progress}}
	options := FitOptions{Epochs: 3, BatchSize: 1, LearningRate: .001, WeightDecay: .01, Seed: 1729}
	m := NewModel()
	r, e := m.Fit(store, samples, options)
	if e != nil || r.TrainingSteps != 6 || r.Batches != 6 || r.Samples != 2 || reader.max != RowBytes || reader.calls != 8 {
		t.Fatal("actual toy streamed fit", r, e, reader)
	}
	var saved bytes.Buffer
	if m.Save(&saved) != nil || saved.Len() != ArtifactBytes {
		t.Fatal("save")
	}
	before := append([]byte(nil), saved.Bytes()...)
	loaded, e := Load(bytes.NewReader(before))
	if e != nil {
		t.Fatal(e)
	}
	var workspace Workspace
	for i := range rows {
		a, e := m.ScoreFeatures(&rows[i], &workspace)
		if e != nil {
			t.Fatal(e)
		}
		b, e := loaded.ScoreFeatures(&rows[i], nil)
		if e != nil || a != b || a.Source != statehint.Learned || a.TrainingSteps != 6 {
			t.Fatal("trained feature parity")
		}
	}
	var again bytes.Buffer
	if loaded.Save(&again) != nil || !bytes.Equal(before, again.Bytes()) {
		t.Fatal("byte roundtrip")
	}
	if _, e := m.Fit(store, []Sample{{9, statehint.Question}}, options); e == nil {
		t.Fatal("bad row accepted")
	}
	var unchanged bytes.Buffer
	if m.Save(&unchanged) != nil || !bytes.Equal(before, unchanged.Bytes()) {
		t.Fatal("failed Fit changed model")
	}
	if _, e := m.Fit(store, samples, options); e != nil {
		t.Fatal(e)
	}
	var fresh bytes.Buffer
	m.Save(&fresh)
	if !bytes.Equal(before, fresh.Bytes()) {
		t.Fatal("Fit was a warm start")
	}
	if _, e := statehintwide.Load(bytes.NewReader(before)); e == nil {
		t.Fatal("feature head accepted by text v2")
	}
	if _, e := statehintmlp.Load(bytes.NewReader(before)); e == nil {
		t.Fatal("feature head accepted by MLP v3")
	}
}

func TestRowsArtifactsAndNumericsFailClosed(t *testing.T) {
	r := originalRow()
	data := rowBytes([]Row{r})
	store, e := NewStore(bytes.NewReader(data), 1)
	if e != nil {
		t.Fatal(e)
	}
	var w Workspace
	if store.ReadRow(0, &w) != nil || w.Row != r {
		t.Fatal("row endian/layout")
	}
	for _, index := range []int{-1, 1} {
		if store.ReadRow(index, &w) == nil {
			t.Fatal("row bound")
		}
	}
	short, _ := NewStore(bytes.NewReader(data[:len(data)-1]), 1)
	if short.ReadRow(0, &w) == nil {
		t.Fatal("short row")
	}
	binary.LittleEndian.PutUint32(data[:4], math.Float32bits(float32(math.NaN())))
	bad, _ := NewStore(bytes.NewReader(data), 1)
	if bad.ReadRow(0, &w) == nil {
		t.Fatal("nonfinite feature")
	}
	if _, e := NewStore(nil, 1); e == nil {
		t.Fatal("nil store")
	}
	m := NewModel()
	p, e := m.ScoreFeatures(&r, nil)
	if e != nil || p.Intent != statehint.Unclear || p.Source != statehint.Untrained || p.Confidence != .125 || p.Margin != 0 {
		t.Fatal("fresh tie")
	}
	if m.SetTemperature(math.NaN()) == nil || m.SetTemperature(.01) == nil {
		t.Fatal("temperature")
	}
	r[0][0] = float32(math.Inf(1))
	if _, e := m.ScoreFeatures(&r, nil); e == nil {
		t.Fatal("nonfinite score row")
	}
	var b bytes.Buffer
	m.Save(&b)
	saved := b.Bytes()
	for _, candidate := range [][]byte{saved[:len(saved)-1], append(append([]byte(nil), saved...), 0), make([]byte, len(saved))} {
		if _, e := Load(bytes.NewReader(candidate)); e == nil {
			t.Fatal("bad artifact")
		}
	}
	if _, e := Load(nil); e == nil {
		t.Fatal("nil artifact")
	}
	if e := m.Save(shortWriter{}); e != io.ErrShortWrite {
		t.Fatal("short writer")
	}
	clone := m.Clone()
	if clone == m || !reflect.DeepEqual(clone, m) {
		t.Fatal("clone ownership")
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
