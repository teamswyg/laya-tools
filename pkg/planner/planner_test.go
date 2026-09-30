package planner_test

import (
	"encoding/json"
	"github.com/teamswyg/laya-tools/pkg/planner"
	"os"
	"testing"
)

func fixtures(t testing.TB) (planner.Config, planner.Request) {
	t.Helper()
	var c planner.Config
	var r planner.Request
	for p, v := range map[string]any{"config.json": &c, "request.json": &r} {
		b, err := os.ReadFile("../../examples/planner/" + p)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, v); err != nil {
			t.Fatal(err)
		}
	}
	return c, r
}
func TestComposition(t *testing.T) {
	c, r := fixtures(t)
	p, err := planner.Build(c, r)
	if err != nil || p.Status != "recommend" || p.RecommendedModel != "example-fast" {
		t.Fatal(p, err)
	}
	r.Pinned = true
	p, err = planner.Build(c, r)
	if err != nil || p.Status != "hold" || p.RecommendedModel != r.Current {
		t.Fatal(p, err)
	}
	// Manual pins never authorize violating privacy constraints.
	r.Assessment.LocalOnly = true
	p, err = planner.Build(c, r)
	if err != nil || p.Status != "blocked" || p.RecommendedModel != "" {
		t.Fatal(p, err)
	}
}
func TestUncertainCannotDowngrade(t *testing.T) {
	c, r := fixtures(t)
	r.Assessment.Uncertain = true
	p, err := planner.Build(c, r)
	if err != nil || p.Status != "hold" || p.RecommendedModel != "example-strong" {
		t.Fatal(p, err)
	}
}
func TestIneligibleCurrentCannotBeHeld(t *testing.T) {
	c, r := fixtures(t)
	r.Current = "example-fast"
	r.Assessment.Tier = 2
	r.Assessment.Confidence = .1
	p, err := planner.Build(c, r)
	if err != nil || p.Status != "blocked" || p.RecommendedModel != "" {
		t.Fatal(p, err)
	}
}
func TestUnknownCurrent(t *testing.T) {
	c, r := fixtures(t)
	r.Current = "unregistered"
	if _, err := planner.Build(c, r); err == nil {
		t.Fatal("unknown current accepted")
	}
}
func BenchmarkBuild(b *testing.B) {
	c, r := fixtures(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := planner.Build(c, r); err != nil {
			b.Fatal(err)
		}
	}
}

func TestBidirectionalPlans(t *testing.T) {
	for _, tc := range []struct {
		name, status, direction, model string
		change                         func(*planner.Config, *planner.Request)
	}{
		{"downgrade", "recommend", "downgrade", "example-fast", func(c *planner.Config, r *planner.Request) {}},
		{"upgrade", "recommend", "upgrade", "example-strong", func(c *planner.Config, r *planner.Request) {
			r.Current = "example-fast"
			r.Assessment.Tier = 2
			n := 0
			r.PromptsSinceSwitch = &n
		}},
		{"pinned upgrade blocked", "blocked", "", "", func(c *planner.Config, r *planner.Request) {
			r.Current = "example-fast"
			r.Assessment.Tier = 2
			r.Pinned = true
		}},
		{"uncertain upgrade blocked", "blocked", "", "", func(c *planner.Config, r *planner.Request) {
			r.Current = "example-fast"
			r.Assessment.Tier = 2
			r.Assessment.Uncertain = true
		}},
		{"upgrade budget blocked", "blocked", "", "", func(c *planner.Config, r *planner.Request) {
			r.Current = "example-fast"
			r.Assessment.Tier = 2
			zero := 0.0
			c.Catalog.Models[2].BudgetUSD = &zero
		}},
		{"hold", "hold", "unchanged", "example-strong", func(c *planner.Config, r *planner.Request) { r.Pinned = true }},
		{"initial", "recommend", "initial", "example-fast", func(c *planner.Config, r *planner.Request) { r.Current = ""; r.ContextTokens = 0 }},
		{"lateral", "recommend", "lateral", "example-fast", func(c *planner.Config, r *planner.Request) { c.Catalog.Models[0].Rank = c.Catalog.Models[2].Rank }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, r := fixtures(t)
			tc.change(&c, &r)
			p, err := planner.Build(c, r)
			if err != nil || p.Status != tc.status || p.Direction != tc.direction || p.RecommendedModel != tc.model {
				t.Fatalf("plan=%+v err=%v", p, err)
			}
			if tc.name == "upgrade" && p.Reason != "quality_upgrade" {
				t.Fatal(p)
			}
		})
	}
}
