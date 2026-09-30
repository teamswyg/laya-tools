package router

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/inference"
	"math"
	"strings"
	"unicode"
)

type Scorer interface {
	Predict(string, string, string, []string) (inference.Prediction, error)
}
type Config struct {
	Fast      string  `json:"fast"`
	Standard  string  `json:"standard"`
	Strong    string  `json:"strong"`
	Threshold float64 `json:"threshold"`
}
type Result struct {
	Tier          string    `json:"tier"`
	SuggestedTier string    `json:"suggested_tier,omitempty"`
	Model         string    `json:"model"`
	Reason        string    `json:"reason"`
	Confidence    float64   `json:"confidence,omitempty"`
	Probabilities []float64 `json:"probabilities,omitempty"`
	Abstained     bool      `json:"abstained"`
	Truncated     bool      `json:"truncated"`
}

// NeedsInference avoids loading native weights when policy can decide up front.
func NeedsInference(prompt, override string, c Config) bool {
	if override != "" || strings.TrimSpace(prompt) == "" || len(prompt) > 65536 || math.IsNaN(c.Threshold) || math.IsInf(c.Threshold, 0) || c.Threshold < .5 || c.Threshold > 1 {
		return false
	}
	for _, r := range prompt {
		if unicode.Is(unicode.Hangul, r) {
			return false
		}
	}
	return true
}

func Route(prompt, override string, c Config, s Scorer) (Result, error) {
	if strings.TrimSpace(prompt) == "" || len(prompt) > 65536 {
		return Result{}, fmt.Errorf("prompt required (max 64 KiB)")
	}
	if override != "" {
		return Result{Tier: "explicit", Model: override, Reason: "explicit model override"}, nil
	}
	if math.IsNaN(c.Threshold) || math.IsInf(c.Threshold, 0) || c.Threshold < .5 || c.Threshold > 1 {
		return Result{}, fmt.Errorf("threshold must be .5..1")
	}
	fallback := func(reason string) Result {
		return Result{Tier: "strong", Model: c.Strong, Reason: reason, Abstained: true}
	}
	// This English checkpoint has no Korean routing calibration. Preserve capability.
	for _, r := range prompt {
		if unicode.Is(unicode.Hangul, r) {
			return fallback("Korean routing is unvalidated"), nil
		}
	}
	if s == nil {
		return fallback("Laya unavailable; preserve strong model"), nil
	}
	p, err := s.Predict(prompt, "choice", "Select the minimum coding capability needed to complete the user's request correctly.", []string{"fast: a small, obvious edit, typo, formatting, or straightforward factual question", "standard: a bounded feature or bug fix with ordinary implementation and tests", "strong: difficult debugging, architecture, security, concurrency, migrations, or ambiguous multi-file work"})
	if err != nil {
		return fallback("Laya inference failed; preserve strong model"), nil
	}
	tiers := []string{"fast", "standard", "strong"}
	models := []string{c.Fast, c.Standard, c.Strong}
	r := Result{Tier: tiers[p.Winner], SuggestedTier: tiers[p.Winner], Model: models[p.Winner], Reason: "experimental Laya complexity classification; probabilities are not correctness guarantees", Confidence: p.Probabilities[p.Winner], Probabilities: p.Probabilities, Truncated: p.Truncated}
	if p.Truncated || r.Confidence < c.Threshold {
		r.Tier = "strong"
		r.Model = c.Strong
		r.Abstained = true
		r.Reason = "low confidence or truncated input; preserve strong model"
	}
	if r.Model == "" {
		r = fallback("selected tier has no configured model")
	}
	return r, nil
}
