// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

// Parameters is an owned copy for isolated quantization/adaptation experiments.
// It exposes no optimizer, task authority, or mutable view of a live Model.
type Parameters struct {
	Weights [FeatureBins][HeadCount][StateCount]float32
	Bias    [HeadCount][StateCount]float32
}

// Parameters returns a value copy. Mutating the copy cannot alter the source
// model. The caller must not invoke it concurrently with Fit on this model.
func (m *Model) Parameters() (Parameters, error) {
	if !m.valid() {
		return Parameters{}, ErrModel
	}
	return Parameters{Weights: m.weights, Bias: m.bias}, nil
}
