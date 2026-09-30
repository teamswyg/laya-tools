// Package switchpolicy evaluates a proposed model change without executing it.
// Cache-payback formula and asymmetric switch guards adapted from pi-pignon
// src/policy.ts at 4d97d1a35134b813bac6a8e345a1bf9b0bbf0bdb.
// Copyright (c) 2026 Nicolas Chaintron. MIT; see licenses/pi-pignon.LICENSE.
// Modified for Go, explicit unknown data, validation, and conservative holds.
package switchpolicy

import (
	"fmt"
	"math"

	"github.com/teamswyg/laya-tools/pkg/catalog"
)

type Config struct {
	UpgradeConfidence   float64 `json:"upgrade_confidence"`
	DowngradeConfidence float64 `json:"downgrade_confidence"`
	Cooldown            int     `json:"cooldown"`
	MaxPaybackRequests  float64 `json:"max_payback_requests"`
}

func DefaultConfig() Config { return Config{.5, .9, 2, 3} }

type Request struct {
	// IDs identify a full execution profile, including provider/model/reasoning settings.
	Current            *catalog.Model `json:"current,omitempty"`
	Target             catalog.Model  `json:"target"`
	Confidence         float64        `json:"confidence"`
	ContextTokens      int            `json:"context_tokens"`
	OutputTokens       int            `json:"output_tokens"`
	PromptsSinceSwitch *int           `json:"prompts_since_switch,omitempty"`
	Pinned             bool           `json:"pinned"`
}

type Decision struct {
	Action          string   `json:"action"` // recommend or hold; never an executed change
	Reason          string   `json:"reason"`
	PaybackRequests *float64 `json:"payback_requests,omitempty"`
}

func finite(v float64) bool     { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func confidence(v float64) bool { return finite(v) && v >= 0 && v <= 1 }

// Payback returns a finite request count and true only when savings are positive
// and the required cache rates are explicitly known. It assumes reusable full context.
func Payback(current, target *catalog.Price, contextTokens, outputTokens int) (float64, bool) {
	if current == nil || target == nil || current.Validate() != nil || target.Validate() != nil || current.CacheRead == nil || target.CacheRead == nil || target.CacheWrite == nil || contextTokens < 0 || outputTokens <= 0 {
		return 0, false
	}
	premium := float64(contextTokens) * math.Max(0, math.Max(target.Input, *target.CacheWrite)-*target.CacheRead)
	saving := float64(contextTokens)*(*current.CacheRead-*target.CacheRead) + float64(outputTokens)*(current.Output-target.Output)
	if saving <= 0 || !finite(premium) || !finite(saving) {
		return 0, false
	}
	n := premium / saving
	return n, finite(n)
}

func Decide(c Config, r Request) (Decision, error) {
	hold := func(reason string) (Decision, error) { return Decision{Action: "hold", Reason: reason}, nil }
	if !confidence(c.UpgradeConfidence) || c.UpgradeConfidence < .5 || !confidence(c.DowngradeConfidence) || c.DowngradeConfidence < c.UpgradeConfidence || c.Cooldown < 0 || !finite(c.MaxPaybackRequests) || c.MaxPaybackRequests < 0 || !confidence(r.Confidence) || r.ContextTokens < 0 || r.OutputTokens <= 0 || (r.PromptsSinceSwitch != nil && *r.PromptsSinceSwitch < 0) {
		return Decision{}, fmt.Errorf("invalid switch policy input")
	}
	for _, m := range []*catalog.Model{r.Current, &r.Target} {
		if m != nil && (m.ID == "" || m.Rank < 0 || m.Price.Validate() != nil) {
			return Decision{}, fmt.Errorf("invalid model profile")
		}
	}
	if r.Pinned {
		return hold("manual_pin")
	}
	if r.Current == nil {
		if r.ContextTokens != 0 {
			return hold("current_model_unknown")
		}
		if r.Confidence < c.DowngradeConfidence {
			return hold("entry_confidence")
		}
		return Decision{Action: "recommend", Reason: "new_session"}, nil
	}
	if r.Current.ID == r.Target.ID {
		return hold("already_selected")
	}
	if r.Target.Rank > r.Current.Rank {
		if r.Confidence < c.UpgradeConfidence {
			return hold("upgrade_confidence")
		}
		return Decision{Action: "recommend", Reason: "quality_upgrade"}, nil
	}
	// Lateral and downward changes both require strong confidence in this port.
	if r.Confidence < c.DowngradeConfidence {
		return hold("downgrade_confidence")
	}
	if r.PromptsSinceSwitch == nil {
		return hold("switch_history_unknown")
	}
	if *r.PromptsSinceSwitch < c.Cooldown {
		return hold("cooldown")
	}
	n, ok := Payback(r.Current.Price, r.Target.Price, r.ContextTokens, r.OutputTokens)
	if !ok {
		return hold("prices_unknown_or_no_savings")
	}
	d := Decision{Action: "hold", Reason: "cache_payback_too_long", PaybackRequests: &n}
	if n <= c.MaxPaybackRequests {
		d.Action, d.Reason = "recommend", "cache_payback_acceptable"
	}
	return d, nil
}
