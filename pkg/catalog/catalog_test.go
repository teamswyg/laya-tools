package catalog

import (
	"math"
	"testing"
)

func fixture() (Config, Request) {
	return Config{Models: []Model{
		{ID: "fast", Rank: 0, Quality: .6, Context: 8192, Tools: true, Price: &Price{Input: 1, Output: 2}},
		{ID: "strong", Rank: 1, Quality: .98, Context: 32768, Tools: true, Vision: true, Price: &Price{Input: 10, Output: 30}},
	}, QualityFloors: []float64{.5, .95}, MinConfidence: .9}, Request{Tier: 0, Confidence: .96, InputTokens: 1000, OutputTokens: 1000}
}

func TestSelectionGuards(t *testing.T) {
	for _, tt := range []struct {
		name, want string
		change     func(*Config, *Request)
	}{
		{"cheap capable", "fast", func(c *Config, r *Request) {}},
		{"uncertain", "strong", func(c *Config, r *Request) { r.Confidence = .89 }},
		{"abstention", "strong", func(c *Config, r *Request) { r.Uncertain = true }},
		{"vision", "strong", func(c *Config, r *Request) { r.Vision = true }},
		{"locality blocks all", "", func(c *Config, r *Request) { r.LocalOnly = true }},
		{"context includes output", "strong", func(c *Config, r *Request) { r.InputTokens = 8000 }},
		{"no quality fallback", "", func(c *Config, r *Request) { c.QualityFloors[1] = 1; r.Tier = 1 }},
		{"unknown price", "strong", func(c *Config, r *Request) { c.Models[0].Price = nil }},
		{"explicit free", "fast", func(c *Config, r *Request) { c.Models[0].Price = &Price{} }},
		{"unknown budget usage", "strong", func(c *Config, r *Request) { b := 1.0; c.Models[0].BudgetUSD = &b }},
		{"projected budget", "strong", func(c *Config, r *Request) {
			b := .01
			c.Models[0].BudgetUSD = &b
			r.Usage = map[string]Usage{"fast": {SpentUSD: .009}}
		}},
		{"load penalty", "strong", func(c *Config, r *Request) { c.LoadPenalty = 1; r.Usage = map[string]Usage{"fast": {InFlight: 20}} }},
		{"huge token counts", "", func(c *Config, r *Request) { r.InputTokens = math.MaxInt; r.OutputTokens = math.MaxInt }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, r := fixture()
			tt.change(&c, &r)
			s, err := Select(c, r)
			if err != nil || s.Model != tt.want {
				t.Fatalf("got %+v %v; want %s", s, err, tt.want)
			}
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	for _, change := range []func(*Config, *Request){
		func(c *Config, r *Request) { c.Models[1].ID = "fast" },
		func(c *Config, r *Request) { c.MinConfidence = math.NaN() },
		func(c *Config, r *Request) { c.Models[0].Price.Input = math.Inf(1) },
		func(c *Config, r *Request) { r.Confidence = math.NaN() },
		func(c *Config, r *Request) { r.OutputTokens = 0 },
		func(c *Config, r *Request) { r.Usage = map[string]Usage{"typo": {}} },
		func(c *Config, r *Request) { c.QualityFloors = []float64{.9, .5} },
	} {
		c, r := fixture()
		change(&c, &r)
		if _, err := Select(c, r); err == nil {
			t.Fatal("accepted invalid input")
		}
	}
}

func TestDeterministicTie(t *testing.T) {
	c, r := fixture()
	c.Models[1] = c.Models[0]
	c.Models[1].ID = "aaa"
	for range 2 {
		s, err := Select(c, r)
		if err != nil || s.Model != "aaa" {
			t.Fatal(s, err)
		}
		c.Models[0], c.Models[1] = c.Models[1], c.Models[0]
	}
}

func BenchmarkSelect(b *testing.B) {
	c, r := fixture()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Select(c, r); err != nil {
			b.Fatal(err)
		}
	}
}
