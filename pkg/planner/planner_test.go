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
