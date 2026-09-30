// Package catalog selects the least expensive model that meets explicit constraints.
// It performs no inference, network requests, accounting, or state mutation.
// Selection structure adapted from mmornati/system-one-router internal/router/score.go
// at a437d00bca33a4c10038b37a5efe04dc0e3d40bb (Apache-2.0).
package catalog

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Price is USD per million tokens. Nil means unknown; a non-nil zero price
// explicitly means free. Cache rates are optional, never inferred from input rates.
type Price struct {
	Input      float64  `json:"input"`
	Output     float64  `json:"output"`
	CacheRead  *float64 `json:"cache_read,omitempty"`
	CacheWrite *float64 `json:"cache_write,omitempty"`
}

func (p *Price) Validate() error {
	if p == nil {
		return nil
	}
	for _, v := range []float64{p.Input, p.Output} {
		if !finite(v) || v < 0 {
			return fmt.Errorf("prices must be finite and nonnegative")
		}
	}
	for _, v := range []*float64{p.CacheRead, p.CacheWrite} {
		if v != nil && (!finite(*v) || *v < 0) {
			return fmt.Errorf("cache prices must be finite and nonnegative")
		}
	}
	return nil
}

type Model struct {
	ID string `json:"id"`
	// Rank is an operator-supplied capability ordering, not an independently measured score.
	Rank      int      `json:"rank"`
	Quality   float64  `json:"quality"`
	Context   int      `json:"context"`
	Tools     bool     `json:"tools"`
	Vision    bool     `json:"vision"`
	Local     bool     `json:"local"`
	Price     *Price   `json:"price,omitempty"`
	BudgetUSD *float64 `json:"budget_usd,omitempty"`
}

type Config struct {
	Models []Model `json:"models"`
	// Low-confidence assessments use the largest configured quality floor.
	QualityFloors []float64 `json:"quality_floors"`
	MinConfidence float64   `json:"min_confidence"`
	LoadPenalty   float64   `json:"load_penalty"`
}

type Usage struct {
	SpentUSD float64 `json:"spent_usd"`
	InFlight int     `json:"in_flight"`
}

type Request struct {
	Tier         int     `json:"tier"`
	Confidence   float64 `json:"confidence"`
	Uncertain    bool    `json:"uncertain"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Tools        bool    `json:"tools"`
	Vision       bool    `json:"vision"`
	LocalOnly    bool    `json:"local_only"`
	// Snapshot for one caller-defined budget period. This is not a reservation ledger.
	Usage map[string]Usage `json:"usage,omitempty"`
}

type Candidate struct {
	ID           string   `json:"id"`
	Eligible     bool     `json:"eligible"`
	Reasons      []string `json:"reasons"`
	EstimatedUSD *float64 `json:"estimated_usd,omitempty"`
	EffectiveUSD *float64 `json:"effective_usd,omitempty"`
}

type Selection struct {
	Model           string      `json:"model,omitempty"`
	RequiredQuality float64     `json:"required_quality"`
	Guarded         bool        `json:"guarded"`
	Candidates      []Candidate `json:"candidates"`
}

func finite(v float64) bool      { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func probability(v float64) bool { return finite(v) && v >= 0 && v <= 1 }

func (c Config) Validate() error {
	if len(c.Models) == 0 || len(c.Models) > 1024 || len(c.QualityFloors) == 0 || len(c.QualityFloors) > 32 {
		return fmt.Errorf("require 1..1024 models and 1..32 quality floors")
	}
	if !probability(c.MinConfidence) || c.MinConfidence < .5 || !finite(c.LoadPenalty) || c.LoadPenalty < 0 {
		return fmt.Errorf("invalid confidence threshold or load penalty")
	}
	last := -1.0
	for _, q := range c.QualityFloors {
		if !probability(q) || q < last {
			return fmt.Errorf("quality floors must be ascending in [0,1]")
		}
		last = q
	}
	seen := map[string]bool{}
	for _, m := range c.Models {
		if strings.TrimSpace(m.ID) == "" || seen[m.ID] || m.Rank < 0 || !probability(m.Quality) || m.Context <= 0 {
			return fmt.Errorf("invalid or duplicate model %q", m.ID)
		}
		seen[m.ID] = true
		if err := m.Price.Validate(); err != nil {
			return fmt.Errorf("model %s: %w", m.ID, err)
		}
		if m.BudgetUSD != nil && (!finite(*m.BudgetUSD) || *m.BudgetUSD < 0) {
			return fmt.Errorf("invalid model budget")
		}
	}
	return nil
}

func Select(c Config, r Request) (Selection, error) {
	if err := c.Validate(); err != nil {
		return Selection{}, err
	}
	if r.Tier < 0 || r.Tier >= len(c.QualityFloors) || !probability(r.Confidence) || r.InputTokens < 0 || r.OutputTokens <= 0 {
		return Selection{}, fmt.Errorf("invalid assessment or token estimate (output must be positive)")
	}
	known := map[string]bool{}
	for _, m := range c.Models {
		known[m.ID] = true
	}
	for id, u := range r.Usage {
		if !known[id] || !finite(u.SpentUSD) || u.SpentUSD < 0 || u.InFlight < 0 {
			return Selection{}, fmt.Errorf("invalid usage for %q", id)
		}
	}
	s := Selection{RequiredQuality: c.QualityFloors[r.Tier], Guarded: r.Uncertain || r.Confidence < c.MinConfidence, Candidates: make([]Candidate, 0, len(c.Models))}
	if s.Guarded {
		s.RequiredQuality = c.QualityFloors[len(c.QualityFloors)-1]
	}
	bestCost := math.Inf(1)
	bestQuality := -1.0
	for _, m := range c.Models {
		v := Candidate{ID: m.ID, Reasons: []string{}}
		reject := func(reason string) { v.Reasons = append(v.Reasons, reason) }
		if m.Quality < s.RequiredQuality {
			reject("below_quality_floor")
		}
		// Subtraction avoids integer overflow for untrusted token counts.
		if r.InputTokens > m.Context || r.OutputTokens > m.Context-r.InputTokens {
			reject("context_limit")
		}
		if r.Tools && !m.Tools {
			reject("tools_required")
		}
		if r.Vision && !m.Vision {
			reject("vision_required")
		}
		if r.LocalOnly && !m.Local {
			reject("local_required")
		}
		u, usageKnown := r.Usage[m.ID]
		if m.BudgetUSD != nil && !usageKnown {
			reject("budget_usage_unknown")
		}
		if m.Price == nil {
			reject("price_unknown")
		} else {
			cost := (float64(r.InputTokens)*m.Price.Input + float64(r.OutputTokens)*m.Price.Output) / 1e6
			effective := cost * (1 + c.LoadPenalty*float64(u.InFlight))
			if !finite(cost) || !finite(effective) {
				return Selection{}, fmt.Errorf("cost overflow for %s", m.ID)
			}
			v.EstimatedUSD, v.EffectiveUSD = &cost, &effective
			if m.BudgetUSD != nil && (u.SpentUSD > *m.BudgetUSD || cost > *m.BudgetUSD-u.SpentUSD) {
				reject("projected_budget_exceeded")
			}
		}
		v.Eligible = len(v.Reasons) == 0
		if v.Eligible && (*v.EffectiveUSD < bestCost || (*v.EffectiveUSD == bestCost && (m.Quality > bestQuality || (m.Quality == bestQuality && m.ID < s.Model)))) {
			s.Model, bestCost, bestQuality = m.ID, *v.EffectiveUSD, m.Quality
		}
		s.Candidates = append(s.Candidates, v)
	}
	sort.Slice(s.Candidates, func(i, j int) bool { return s.Candidates[i].ID < s.Candidates[j].ID })
	return s, nil
}
