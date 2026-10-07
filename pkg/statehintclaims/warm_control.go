// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"math"
	"math/rand/v2"
)

// WarmFitReport separates inherited model history from new optimizer updates.
type WarmFitReport struct {
	FitReport
	BaseTrainingSteps uint64 `json:"base_training_steps"`
	NewOptimizerSteps uint64 `json:"new_optimizer_steps"`
}

// WarmFit is a fixed, matched float continuation control for the ternary study.
// It copies the supplied parent, resets Adam moments and optimizer step, and
// adds40epochs with the same order/recipe as ternary QAT. It changes neither
// parent nor the existing fresh Fit implementation. Exclusive receiver and
// workspace ownership are required; errors leave receiver/parent unchanged.
func (m *Model) WarmFit(parent *Model, samples []Sample, workspace *TrainingWorkspace) (WarmFitReport, error) {
	if m == nil || !m.valid() || parent == nil || m == parent || !parent.valid() || workspace == nil || len(samples) == 0 || len(samples) > 1_000_000 {
		return WarmFitReport{}, ErrTraining
	}
	o := FitOptions{Epochs: 40, BatchSize: 32, LearningRate: .001, WeightDecay: .01, Seed: 1729}
	for _, s := range samples {
		for _, t := range s.Targets {
			if _, ok := StateIndex(t); !ok {
				return WarmFitReport{}, ErrTraining
			}
		}
		view, e := statehintwide.ExtractContextual(s.Text, &workspace.features)
		if e != nil || view.WordCount() == 0 {
			return WarmFitReport{}, ErrTraining
		}
	}
	updates := uint64(40 * ((len(samples) + 31) / 32))
	base := parent.steps
	if updates > math.MaxUint64-base {
		return WarmFitReport{}, ErrTraining
	}
	workspace.working = *parent
	workspace.optimizer = adamWorkspace{}
	if cap(workspace.order) < len(samples) {
		workspace.order = make([]int, len(samples))
	} else {
		workspace.order = workspace.order[:len(samples)]
	}
	for i := range workspace.order {
		workspace.order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(o.Seed), uint64(o.Seed)^0x9e3779b97f4a7c15))
	r := WarmFitReport{FitReport: FitReport{Samples: len(samples), Epochs: 40, Initialization: "parent_float_copy_fresh_adamw", Seed: o.Seed}, BaseTrainingSteps: base}
	var sum float64
	var step uint64
	for epoch := 0; epoch < 40; epoch++ {
		random.Shuffle(len(workspace.order), func(i, j int) { workspace.order[i], workspace.order[j] = workspace.order[j], workspace.order[i] })
		for start := 0; start < len(samples); start += 32 {
			end := min(start+32, len(samples))
			workspace.optimizer.gradient = gradient{}
			for _, i := range workspace.order[start:end] {
				s := samples[i]
				view, e := statehintwide.ExtractContextual(s.Text, &workspace.features)
				if e != nil {
					return WarmFitReport{}, ErrTraining
				}
				loss, e := workspace.working.accumulate(view, s.Targets, &workspace.optimizer.gradient)
				if e != nil {
					return WarmFitReport{}, e
				}
				sum += loss
			}
			step++
			if e := workspace.optimizer.update(&workspace.working, o, end-start, step); e != nil {
				return WarmFitReport{}, e
			}
			workspace.working.steps = base + step
			r.Batches++
		}
	}
	r.NewOptimizerSteps = step
	r.TrainingSteps = workspace.working.steps
	r.MeanLoss = sum / (float64(len(samples)) * 40)
	if !finite(r.MeanLoss) || !workspace.working.valid() {
		return WarmFitReport{}, ErrTraining
	}
	*m = workspace.working
	return r, nil
}
