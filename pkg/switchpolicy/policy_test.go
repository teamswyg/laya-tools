package switchpolicy

import (
	"github.com/teamswyg/laya-tools/pkg/catalog"
	"math"
	"testing"
)

func ptr[T any](v T) *T { return &v }
func fixture() Request {
	return Request{Current: &catalog.Model{ID: "strong", Rank: 2, Price: &catalog.Price{Input: 10, Output: 30, CacheRead: ptr(1.0), CacheWrite: ptr(10.0)}}, Target: catalog.Model{ID: "fast", Rank: 0, Price: &catalog.Price{Input: 1, Output: 2, CacheRead: ptr(.1), CacheWrite: ptr(1.0)}}, Confidence: .96, ContextTokens: 5000, OutputTokens: 1000, PromptsSinceSwitch: ptr(3)}
}
func TestSwitchGuards(t *testing.T) {
	for _, tt := range []struct {
		name, action, reason string
		change               func(*Request)
	}{
		{"cheap downgrade", "recommend", "cache_payback_acceptable", func(r *Request) {}},
		{"manual pin", "hold", "manual_pin", func(r *Request) { r.Pinned = true }},
		{"low confidence", "hold", "downgrade_confidence", func(r *Request) { r.Confidence = .89 }},
		{"cooldown", "hold", "cooldown", func(r *Request) { r.PromptsSinceSwitch = ptr(1) }},
		{"history missing", "hold", "switch_history_unknown", func(r *Request) { r.PromptsSinceSwitch = nil }},
		{"cache rate missing", "hold", "prices_unknown_or_no_savings", func(r *Request) { r.Target.Price.CacheRead = nil }},
		{"price missing", "hold", "prices_unknown_or_no_savings", func(r *Request) { r.Target.Price = nil }},
		{"no savings", "hold", "prices_unknown_or_no_savings", func(r *Request) { r.Target.Price = r.Current.Price }},
		{"expensive cache miss", "hold", "cache_payback_too_long", func(r *Request) { r.Target.Price.CacheWrite = ptr(100.0) }},
		{"upgrade bypasses cooldown price", "recommend", "quality_upgrade", func(r *Request) {
			r.Target.Rank = 3
			r.Confidence = .6
			r.PromptsSinceSwitch = nil
			r.Target.Price = nil
		}},
		{"uncertain upgrade", "hold", "upgrade_confidence", func(r *Request) { r.Target.Rank = 3; r.Confidence = .4 }},
		{"new session", "recommend", "new_session", func(r *Request) { r.Current = nil; r.ContextTokens = 0 }},
		{"unknown active model", "hold", "current_model_unknown", func(r *Request) { r.Current = nil }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := fixture()
			tt.change(&r)
			d, err := Decide(DefaultConfig(), r)
			if err != nil || d.Action != tt.action || d.Reason != tt.reason {
				t.Fatal(d, err)
			}
		})
	}
}
func TestPaybackReference(t *testing.T) {
	r := fixture()
	n, ok := Payback(r.Current.Price, r.Target.Price, 5000, 1000)
	// Upstream formula: premium 4500 / recurring savings (4500 + 28000).
	if !ok || math.Abs(n-4500.0/32500) > 1e-12 {
		t.Fatal(n, ok)
	}
	r.Target.Price.Input = math.NaN()
	if _, ok := Payback(r.Current.Price, r.Target.Price, 5000, 1000); ok {
		t.Fatal("accepted NaN")
	}
}
func TestInvalidPolicy(t *testing.T) {
	c := DefaultConfig()
	c.MaxPaybackRequests = math.Inf(1)
	if _, err := Decide(c, fixture()); err == nil {
		t.Fatal("accepted infinity")
	}
}
func BenchmarkDecide(b *testing.B) {
	r := fixture()
	c := DefaultConfig()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Decide(c, r); err != nil {
			b.Fatal(err)
		}
	}
}
