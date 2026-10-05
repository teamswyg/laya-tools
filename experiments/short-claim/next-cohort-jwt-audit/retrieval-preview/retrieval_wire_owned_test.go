// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestRetrievalCompleteWireByteBoundary(t *testing.T) {
	// Two quotes and one encoder newline are charged in the exact cap.
	raw, code := marshalRetrievalBounded(strings.Repeat("x", retrievalStdoutLimit-3))
	if code != retrievalWireOK || len(raw) != retrievalStdoutLimit || raw[len(raw)-1] != '\n' {
		t.Fatal("exact complete wire boundary rejected")
	}
	raw, code = marshalRetrievalBounded(strings.Repeat("x", retrievalStdoutLimit-2))
	if code != retrievalWireBoundUnknown || raw != nil {
		t.Fatal("oversized output partially returned")
	}
	raw, code = marshalRetrievalBounded(math.Inf(1))
	if code != retrievalWireMarshalUnknown || raw != nil {
		t.Fatal("nonfinite JSON accepted or leaked error")
	}
}
func TestRetrievalGoEscapingAndFiniteScalarBudget(t *testing.T) {
	// Go's default JSON encoder escapes < as six bytes. Pure Node sizing is not
	// an exact Go marshal proof; runtime guards charge actual escaped bytes.
	n := (retrievalStdoutLimit - 3) / 6
	raw, code := marshalRetrievalBounded(strings.Repeat("<", n))
	if code != retrievalWireOK || len(raw) != 6*n+3 {
		t.Fatal("HTML escaping not charged")
	}
	if raw, code := marshalRetrievalBounded(strings.Repeat("<", n+1)); code != retrievalWireBoundUnknown || raw != nil {
		t.Fatal("escaped oversized output accepted")
	}
	for _, v := range []float64{math.MaxFloat64, -math.MaxFloat64, math.SmallestNonzeroFloat64, 1.0000000000000002e-6, -1.0000000000000002e-6, 0, 1e20} {
		raw, err := json.Marshal(v)
		if err != nil || len(raw) > 32 {
			t.Fatal("finite scalar exceeded prospective conservative budget")
		}
	}
}
