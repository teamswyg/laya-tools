// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintwide

import (
	"bytes"
	"testing"
)

func TestBorrowedContextualAPIMatchesExistingExtractorExactly(t *testing.T) {
	for _, text := range []string{"I am checking an owned boundary; checking checking.", "가상 작업 단위를 확인하고 있습니다. 완료 여부는 별도입니다.", "İ Aé\nB2?", "", "..."} {
		var public, existing Workspace
		view, err := ExtractContextual(text, &public)
		if err != nil || extract(text, &existing, Contextual) != nil || view.Len() != existing.count || view.WordCount() != existing.wordCount || ContextualFeatureSchema != contextSchema {
			t.Fatal("existing extractor/schema drift")
		}
		for i := 0; i < view.Len(); i++ {
			f := view.At(i)
			if f.Index != existing.indices[i] || f.Value != existing.values[f.Index] {
				t.Fatal("feature/order/float32 drift", i, f)
			}
		}
	}
	if _, err := ExtractContextual("owned text", nil); err != ErrInput {
		t.Fatal("nil caller workspace accepted")
	}
	var w Workspace
	if _, err := ExtractContextual("invalid\x00input", &w); err != ErrInput {
		t.Fatal("existing bounded-input guard changed")
	}
}

func TestContextualAPIBorrowsWorkspaceWithoutArtifactOrAllocationChange(t *testing.T) {
	var w Workspace
	view, err := ExtractContextual("owned borrowed feature text", &w)
	if err != nil || view.workspace != &w {
		t.Fatal("view is not caller-owned")
	}
	if allocs := testing.AllocsPerRun(20, func() {
		if _, err := ExtractContextual("owned borrowed feature text", &w); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatal("feature API copied/allocated", allocs)
	}
	m := NewModel(Contextual)
	m.weights[3][2], m.bias[1] = .75, -.5
	m.steps = 11
	var before, after bytes.Buffer
	if m.Save(&before) != nil {
		t.Fatal("owned v2 artifact")
	}
	if _, err := ExtractContextual("different owned input", &w); err != nil || m.Save(&after) != nil || !bytes.Equal(before.Bytes(), after.Bytes()) {
		t.Fatal("public extraction changed existing v2 model bytes")
	}
}
