// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintwide

// ContextualFeatureSchema identifies the unchanged Contextual extractor.
const ContextualFeatureSchema = contextSchema

// SparseFeature is one activated bin, including signed zero-valued collisions.
type SparseFeature struct {
	Index uint16
	Value float32
}

// ContextualFeatureView borrows its caller-owned Workspace. It is valid only
// until that workspace is reused by extraction or prediction. Do not retain it
// across reuse or share the workspace concurrently; no feature copy is made.
type ContextualFeatureView struct{ workspace *Workspace }

func (v ContextualFeatureView) Len() int {
	if v.workspace == nil {
		return 0
	}
	return v.workspace.count
}

// At requires 0 <= index < Len(), as with indexing a slice.
func (v ContextualFeatureView) At(index int) SparseFeature {
	i := v.workspace.indices[:v.workspace.count][index]
	return SparseFeature{Index: i, Value: v.workspace.values[i]}
}

func (v ContextualFeatureView) WordCount() int {
	if v.workspace == nil {
		return 0
	}
	return v.workspace.wordCount
}

// ExtractContextual reuses the existing signed log-TF/L2 extractor exactly.
// The workspace must be supplied and exclusively owned by this call's caller.
func ExtractContextual(text string, workspace *Workspace) (ContextualFeatureView, error) {
	if err := extract(text, workspace, Contextual); err != nil {
		return ContextualFeatureView{}, err
	}
	return ContextualFeatureView{workspace: workspace}, nil
}
