// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"math"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintfamily"
)

func TestIndependentPredictionsAndExplicitUnknownVersusAbstention(t *testing.T) {
	if ConfidenceFloor != statehintfamily.ConfidenceFloor || MarginFloor != statehintfamily.MarginFloor {
		t.Fatal("existing numeric floors changed")
	}
	m := NewModel()
	m.steps = 7 // Owned numeric fixture, not a claimed training or quality result.
	m.bias[0][0], m.bias[1][0], m.bias[2][2] = 8, 8, 8
	var w Workspace
	p, err := m.Predict("Original owned claim fixture.", &w)
	if err != nil || p.Source != Learned || p.TrainingSteps != 7 || p.Heads[0].State != True || p.Heads[1].State != True || p.Heads[2].State != Unknown || p.Heads[2].UnknownReason != "semantic_unknown" {
		t.Fatal("heads are exclusive or unknown target lost", p, err)
	}
	before := p
	m.bias[0][0], m.bias[0][1] = 0, 8
	after, err := m.Predict("Original owned claim fixture.", &w)
	if err != nil || after.Heads[0].State != False || after.Heads[1] != before.Heads[1] || after.Heads[2] != before.Heads[2] {
		t.Fatal("one head changed another head's probabilities", after)
	}
	m.bias[0] = [3]float32{1, 0, 0}
	after, err = m.Predict("Original owned claim fixture.", &w)
	if err != nil || after.Heads[0].Winner != True || after.Heads[0].State != Unknown || after.Heads[0].UnknownReason != "low_confidence" {
		t.Fatal("low-confidence abstention is not distinct", after)
	}
	u := NewModel()
	untrained, err := u.Predict("Owned word content.", &w)
	if err != nil || untrained.Source != Untrained || untrained.TrainingSteps != 0 {
		t.Fatal("untrained model invented history")
	}
	for _, h := range untrained.Heads {
		if h.State != Unknown || h.Winner != Unknown || h.UnknownReason != "untrained" {
			t.Fatal("untrained/tied numeric output", h)
		}
	}
	empty, err := m.Predict("...", &w)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range empty.Heads {
		if h.State != Unknown || h.Winner != Unknown || h.UnknownReason != "no_word_content" || h.Margin != 0 {
			t.Fatal("punctuation became a claim", h)
		}
	}
}

func TestStableSoftmaxAndBoundedText(t *testing.T) {
	p, ok := softmax([3]float64{10000, 9999, 9998})
	if !ok || math.Abs(p[0]+p[1]+p[2]-1) > 1e-14 || math.Abs(p[0]-.6652409557748218) > 1e-14 {
		t.Fatal("softmax was unstable", p)
	}
	if _, ok := softmax([3]float64{0, math.Inf(1), 0}); ok {
		t.Fatal("nonfinite logits accepted")
	}
	m := NewModel()
	var w Workspace
	for _, text := range []string{strings.Repeat("a", MaxTextBytes+1), "bad\x00input", string([]byte{0xff})} {
		if _, err := m.Predict(text, &w); err != ErrInput {
			t.Fatal("bounded UTF8 input was not reused", err)
		}
	}
	if _, err := m.Predict("owned text", nil); err != ErrInput {
		t.Fatal("inference workspace was not caller-owned")
	}
	for _, s := range States() {
		parsed, ok := ParseState(s.String())
		if !ok || parsed != s {
			t.Fatal("explicit state roundtrip")
		}
	}
	for _, s := range []string{"", "TRUE", "unclear", " true"} {
		if _, ok := ParseState(s); ok {
			t.Fatal("unknown state wire accepted", s)
		}
	}
}

func TestFrozenInferenceUsesIndependentWorkspaces(t *testing.T) {
	m := NewModel()
	var training TrainingWorkspace
	if _, err := m.Fit(ownedSamples(), FitOptions{Epochs: 2, BatchSize: 2}, &training); err != nil {
		t.Fatal(err)
	}
	var w Workspace
	baseline, err := m.Predict("Original owned shared inference fixture.", &w)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan bool, 8)
	for i := 0; i < cap(done); i++ {
		go func() {
			var caller Workspace
			for j := 0; j < 10; j++ {
				p, err := m.Predict("Original owned shared inference fixture.", &caller)
				if err != nil || p != baseline {
					done <- false
					return
				}
			}
			done <- true
		}()
	}
	for i := 0; i < cap(done); i++ {
		if !<-done {
			t.Fatal("caller workspaces coupled frozen predictions")
		}
	}
}
