// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"math"
	"sync"
	"testing"
)

func TestPredictionGuardCommonContractAndAllocationFreeWorkspace(t *testing.T) {
	m := NewModel()
	var w Workspace
	for _, text := range []string{"", "!!!", "🙂"} {
		p, e := m.Predict(text, &w)
		if e != nil || p.Intent != Unclear || p.GuardReason != "no_word_content" || p.Source != Untrained || p.Margin != 0 {
			t.Fatal("guard contract changed", p, e)
		}
		for _, v := range p.Probabilities {
			if v != .125 {
				t.Fatal("guard posterior")
			}
		}
	}
	p, e := m.Predict("Owned original arithmetic fixture", &w)
	if e != nil || p.Source != Untrained || p.TrainingSteps != 0 {
		t.Fatal(e)
	}
	var sum float64
	for _, v := range p.Probabilities {
		if !finite(v) || v < 0 || v > 1 {
			t.Fatal("invalid posterior")
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Fatal("softmax sum")
	}
	if allocations := testing.AllocsPerRun(10, func() {
		if _, e := m.Predict("Owned original arithmetic fixture", &w); e != nil {
			panic(e)
		}
	}); allocations != 0 {
		t.Fatalf("caller workspace prediction allocated%g", allocations)
	}
	copy := m.Clone()
	if copy == m || copy == nil || copy.SetTemperature(2) != nil || m.Temperature() != 1 {
		t.Fatal("clone shares mutable state")
	}
	if m.SetTemperature(math.NaN()) == nil || m.SetTemperature(.01) == nil {
		t.Fatal("invalid temperature")
	}
	var absent *Model
	if absent.Clone() != nil || absent.SetTemperature(1) == nil {
		t.Fatal("nil model contract")
	}
}
func TestReadOnlyPredictionWithSeparateWorkspaces(t *testing.T) {
	m := NewModel()
	ss := toySamples()
	if _, e := m.Fit(ss, FitOptions{Epochs: 2, BatchSize: 8}); e != nil {
		t.Fatal(e)
	}
	want, e := m.Predict(ss[0].Text, nil)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	var failures [8]bool
	for i := range failures {
		wg.Go(func() {
			var w Workspace
			for n := 0; n < 20; n++ {
				got, e := m.Predict(ss[0].Text, &w)
				if e != nil || got != want {
					failures[i] = true
				}
			}
		})
	}
	wg.Wait()
	for _, failed := range failures {
		if failed {
			t.Fatal("immutable model prediction drifted")
		}
	}
}
