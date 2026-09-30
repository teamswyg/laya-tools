// Package planner composes catalog constraints and switch guards into a dry plan.
// Callers own classification, state persistence, accounting, and execution.
package planner

import (
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/catalog"
	"github.com/teamswyg/laya-tools/pkg/switchpolicy"
)

type Config struct {
	Catalog catalog.Config      `json:"catalog"`
	Switch  switchpolicy.Config `json:"switch"`
}

type Request struct {
	Assessment         catalog.Request `json:"assessment"`
	Current            string          `json:"current,omitempty"`
	ContextTokens      int             `json:"context_tokens"`
	PromptsSinceSwitch *int            `json:"prompts_since_switch,omitempty"`
	Pinned             bool            `json:"pinned"`
}

type Plan struct {
	Mode             string                 `json:"mode"`
	Status           string                 `json:"status"` // recommend, hold, or blocked
	RecommendedModel string                 `json:"recommended_model,omitempty"`
	Direction        string                 `json:"direction,omitempty"` // upgrade, downgrade, lateral, initial, or unchanged; absent when blocked
	Reason           string                 `json:"reason"`
	Selection        catalog.Selection      `json:"selection"`
	Switch           *switchpolicy.Decision `json:"switch,omitempty"`
}

func Build(c Config, r Request) (Plan, error) {
	p := Plan{Mode: "plan_only", Status: "blocked"}
	if r.ContextTokens < 0 || r.ContextTokens > r.Assessment.InputTokens || (r.PromptsSinceSwitch != nil && *r.PromptsSinceSwitch < 0) {
		return p, fmt.Errorf("invalid context tokens or switch history")
	}
	s, err := catalog.Select(c.Catalog, r.Assessment)
	if err != nil {
		return p, err
	}
	p.Selection = s
	var current, target *catalog.Model
	for i := range c.Catalog.Models {
		m := &c.Catalog.Models[i]
		if m.ID == r.Current {
			current = m
		}
		if m.ID == s.Model {
			target = m
		}
	}
	if r.Current != "" && current == nil {
		return p, fmt.Errorf("current model is not in catalog")
	}
	// Validate switch configuration even when no eligible candidate is found.
	validationTarget := c.Catalog.Models[0]
	if target != nil {
		validationTarget = *target
	}
	confidence := r.Assessment.Confidence
	if s.Guarded {
		confidence = 0
	} // Never reinterpret abstention as strong evidence.
	d, err := switchpolicy.Decide(c.Switch, switchpolicy.Request{Current: current, Target: validationTarget, Confidence: confidence, ContextTokens: r.ContextTokens, OutputTokens: r.Assessment.OutputTokens, PromptsSinceSwitch: r.PromptsSinceSwitch, Pinned: r.Pinned})
	if err != nil {
		return p, err
	}
	if target == nil {
		p.Reason = "no_eligible_model"
		return p, nil
	}
	p.Switch = &d
	p.Reason = d.Reason
	if d.Action == "recommend" {
		p.Status, p.RecommendedModel = "recommend", target.ID
		switch {
		case current == nil:
			p.Direction = "initial"
		case target.Rank > current.Rank:
			p.Direction = "upgrade"
		case target.Rank < current.Rank:
			p.Direction = "downgrade"
		default:
			p.Direction = "lateral"
		}
		return p, nil
	}
	// A switch guard cannot authorize keeping a model that violates constraints.
	for _, v := range s.Candidates {
		if current != nil && v.ID == current.ID && v.Eligible {
			p.Status, p.RecommendedModel = "hold", current.ID
			p.Direction = "unchanged"
			return p, nil
		}
	}
	p.Reason = "switch_held_and_current_ineligible: " + d.Reason
	return p, nil
}
