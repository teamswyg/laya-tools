// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

// Parameters is an owned float32 copy for isolated future adapters. It exposes
// no optimizer and cannot mutate a live model. There is no parameter importer.
type Parameters struct {
	Input      [FeatureBins][HiddenUnits]float32
	HiddenBias [HiddenUnits]float32
	Output     [HiddenUnits][HeadCount][StateCount]float32
	OutputBias [HeadCount][StateCount]float32
}

func (m *Model) Parameters() (Parameters, error) {
	if !m.valid() {
		return Parameters{}, ErrModel
	}
	return Parameters{Input: m.input, HiddenBias: m.hiddenBias, Output: m.output, OutputBias: m.outputBias}, nil
}
