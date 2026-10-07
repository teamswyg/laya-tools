// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import "testing"

func TestParameterSnapshotOwnsItsStorage(t *testing.T) {
	m := NewModel()
	m.weights[17][1][2] = .25
	m.bias[2][1] = -.5
	p, e := m.Parameters()
	if e != nil || p.Weights[17][1][2] != .25 || p.Bias[2][1] != -.5 {
		t.Fatal("snapshot differs")
	}
	p.Weights[17][1][2] = 99
	p.Bias[2][1] = 99
	if m.weights[17][1][2] != .25 || m.bias[2][1] != -.5 {
		t.Fatal("snapshot mutated model")
	}
	var absent *Model
	if _, e := absent.Parameters(); e == nil {
		t.Fatal("nil model")
	}
}
