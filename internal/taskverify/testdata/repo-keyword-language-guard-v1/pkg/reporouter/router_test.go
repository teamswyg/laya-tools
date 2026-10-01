package reporouter

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

type fakeJudge struct {
	p     Judgment
	calls int
	err   error
}

func (j *fakeJudge) Choose(_ string, _ []Repository) (Judgment, error) { j.calls++; return j.p, j.err }
func fixture(t testing.TB) *Index {
	t.Helper()
	idx, err := New([]Repository{{Name: "example/auth", Summary: "Login sessions and access tokens"}, {Name: "example/billing", Summary: "Payment invoices and refunds"}})
	if err != nil {
		t.Fatal(err)
	}
	return idx
}
func TestPreviewGuards(t *testing.T) {
	for _, tt := range []struct {
		name           string
		p              Judgment
		err            error
		status, reason string
	}{
		{"confident", Judgment{Probabilities: []float64{.95, .05}}, nil, "suggest", "experimental_laya_choice"},
		{"uncertain", Judgment{Probabilities: []float64{.6, .4}}, nil, "abstain", "confidence_or_margin"},
		{"none", Judgment{Probabilities: []float64{.02, .98}}, nil, "abstain", "none_or_ambiguous"},
		{"truncated", Judgment{Probabilities: []float64{.99, .01}, Truncated: true}, nil, "abstain", "input_truncated"},
		{"bad shape", Judgment{Probabilities: []float64{1}}, nil, "abstain", "invalid_judgment"},
		{"bad sum", Judgment{Probabilities: []float64{.95, .3}}, nil, "abstain", "invalid_judgment"},
		{"nan", Judgment{Probabilities: []float64{math.NaN(), 0}}, nil, "abstain", "invalid_judgment"},
		{"inference failure", Judgment{}, fmt.Errorf("fixture"), "candidate", "inference_unavailable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &fakeJudge{p: tt.p, err: tt.err}
			r, err := fixture(t).Preview("login", DefaultConfig(), j)
			if err != nil || r.Status != tt.status || r.Reason != tt.reason || !r.Preview {
				t.Fatal(r, err)
			}
			if r.Status == "abstain" && r.Suggested != "" {
				t.Fatal("abstention retained recommendation")
			}
		})
	}
}
func TestNoUnnecessaryInference(t *testing.T) {
	idx, err := New([]Repository{{Name: "example/auth", Summary: "Login sessions", Keywords: []string{"로그인"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"로그인 수정", "make lunch"} {
		j := &fakeJudge{}
		if _, err := idx.Preview(query, DefaultConfig(), j); err != nil {
			t.Fatal(err)
		}
		if j.calls != 0 {
			t.Fatal("unexpected inference")
		}
	}
}
func TestMetadataValidation(t *testing.T) {
	for _, repos := range [][]Repository{nil, {{Name: "../escape"}}, {{Name: "example/auth", Disabled: true}}, {{Name: "example/auth"}, {Name: "EXAMPLE/AUTH"}}, {{Name: "example/auth", Summary: strings.Repeat("x", 2049)}}} {
		if _, err := New(repos); err == nil {
			t.Fatal("accepted invalid catalog")
		}
	}
}
func TestDisabledNeverSelected(t *testing.T) {
	idx, err := New([]Repository{{Name: "example/old", Summary: "login", Disabled: true}, {Name: "example/new", Summary: "login"}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := idx.Preview("login", DefaultConfig(), nil)
	if err != nil || r.Suggested != "example/new" || len(r.Candidates) != 1 {
		t.Fatal(r, err)
	}
}
func TestCatalogSnapshot(t *testing.T) {
	repos := []Repository{{Name: "example/auth", Keywords: []string{"login"}}}
	idx, err := New(repos)
	if err != nil {
		t.Fatal(err)
	}
	repos[0].Name = "example/changed"
	repos[0].Keywords[0] = "payments"
	r, err := idx.Preview("login", DefaultConfig(), nil)
	if err != nil || r.Suggested != "example/auth" {
		t.Fatal(r, err)
	}
}
func BenchmarkPreview(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			repos := make([]Repository, n)
			for i := range repos {
				repos[i] = Repository{Name: fmt.Sprintf("example/service-%d", i), Summary: "Authentication login sessions access tokens"}
			}
			idx, err := New(repos)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err := idx.Preview("login access tokens", DefaultConfig(), nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
