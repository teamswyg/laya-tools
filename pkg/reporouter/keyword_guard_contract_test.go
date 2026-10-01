// Regression coverage copied from the published independent keyword contract.
// Candidate-authored tests are not included. The frozen task contract remains unchanged.
package reporouter_test

import (
	"errors"
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/reporouter"
	"math"
	"reflect"
	"testing"
)

type contractJudge struct {
	calls      int
	candidates []reporouter.Repository
	judgment   reporouter.Judgment
	failure    error
}

func (j *contractJudge) Choose(_ string, candidates []reporouter.Repository) (reporouter.Judgment, error) {
	j.calls++
	j.candidates = append([]reporouter.Repository(nil), candidates...)
	return j.judgment, j.failure
}
func contractIndex(t *testing.T, repos []reporouter.Repository) *reporouter.Index {
	t.Helper()
	idx, err := reporouter.New(repos)
	if err != nil {
		t.Fatal("original public fixture rejected")
	}
	return idx
}
func lexicalResult(t *testing.T, idx *reporouter.Index, c reporouter.Config) reporouter.Result {
	t.Helper()
	r, err := idx.Preview("login", c, nil)
	if err != nil || r.Status != "candidate" || len(r.Candidates) == 0 {
		t.Fatal("lexical fixture unavailable")
	}
	return r
}
func expectMetadataBlock(t *testing.T, idx *reporouter.Index, c reporouter.Config) {
	t.Helper()
	want := lexicalResult(t, idx, c)
	want.Reason = "metadata_language_unvalidated"
	j := &contractJudge{judgment: reporouter.Judgment{Probabilities: []float64{.95, .05}}}
	got, err := idx.Preview("login", c, j)
	if err != nil || j.calls != 0 || !reflect.DeepEqual(got, want) {
		t.Fatal("blocked metadata called Judge or changed lexical result")
	}
}
func TestPreviewContractEverySelectedKeyword(t *testing.T) {
	for _, unsupported := range []string{"로그인", "認証", "авторизация", "Δοκιμή"} {
		for count := 1; count <= 8; count++ {
			c := reporouter.DefaultConfig()
			c.Candidates = count
			for selected := 0; selected < count; selected++ {
				for position := 0; position < 32; position++ {
					repos := make([]reporouter.Repository, 8)
					for i := range repos {
						keywords := make([]string, 32)
						for k := range keywords {
							keywords[k] = "marker"
						}
						if i == selected {
							keywords[position] = unsupported
						}
						repos[i] = reporouter.Repository{Name: fmt.Sprintf("example/auth-%d", i), Summary: "Login sessions", Keywords: keywords}
					}
					idx := contractIndex(t, repos)
					// Replacing one nonquery term preserves token length and query scoring.
					// Verify the intended candidate position instead of letting all blocking
					// keywords silently move their repository to the end of the ranking.
					lexical := lexicalResult(t, idx, c)
					if len(lexical.Candidates) != count || lexical.Candidates[selected].Name != repos[selected].Name {
						t.Fatal("selected-position fixture changed")
					}
					expectMetadataBlock(t, idx, c)
				}
			}
		}
	}
	// A lower-scoring candidate inside the selected set also contributes metadata.
	idx := contractIndex(t, []reporouter.Repository{
		{Name: "example/login", Summary: "Login sessions login", Keywords: []string{"login"}},
		{Name: "example/auth", Summary: "Login", Keywords: []string{"latin", "로그인"}},
	})
	c := reporouter.DefaultConfig()
	c.Candidates = 2
	expectMetadataBlock(t, idx, c)
}
func TestPreviewContractSelectedScope(t *testing.T) {
	for _, other := range []reporouter.Repository{
		{Name: "example/payments", Summary: "Invoices", Keywords: []string{"決済"}},
		{Name: "example/login-disabled", Summary: "Login", Keywords: []string{"로그인"}, Disabled: true},
		{Name: "example/less", Summary: "Login", Keywords: []string{"로그인"}},
	} {
		idx := contractIndex(t, []reporouter.Repository{{Name: "example/login", Summary: "Login login", Keywords: []string{"login"}}, other})
		c := reporouter.DefaultConfig()
		c.Candidates = 1
		j := &contractJudge{judgment: reporouter.Judgment{Probabilities: []float64{.95, .05}}}
		got, err := idx.Preview("login", c, j)
		if err != nil || j.calls != 1 || len(j.candidates) != 1 || j.candidates[0].Name != "example/login" || got.Status != "suggest" || got.Suggested != "example/login" || got.Reason != "experimental_laya_choice" {
			t.Fatal("unselected metadata affected Judge")
		}
	}
	idx := contractIndex(t, []reporouter.Repository{{Name: "example/auth", Summary: "Login", Keywords: []string{"로그인"}}})
	j := &contractJudge{}
	got, err := idx.Preview("payment", reporouter.DefaultConfig(), j)
	if err != nil || j.calls != 0 || got.Status != "abstain" || got.Reason != "no_lexical_evidence" || got.Suggested != "" {
		t.Fatal("no-evidence behavior changed")
	}
}
func TestPreviewContractNilJudge(t *testing.T) {
	idx := contractIndex(t, []reporouter.Repository{{Name: "example/auth", Summary: "Login", Keywords: []string{"로그인"}}})
	r := lexicalResult(t, idx, reporouter.DefaultConfig())
	if r.Reason != "lexical_only_not_calibrated" || r.Suggested != "example/auth" || r.Judgment != nil || r.Confidence != 0 || r.Margin != 0 || !r.Preview {
		t.Fatal("nil Judge acquired a language guard or authority")
	}
}
func TestPreviewContractExistingLanguageGuards(t *testing.T) {
	idx := contractIndex(t, []reporouter.Repository{{Name: "example/auth", Summary: "Login", Keywords: []string{"로그인"}}})
	j := &contractJudge{}
	r, err := idx.Preview("로그인", reporouter.DefaultConfig(), j)
	if err != nil || j.calls != 0 || r.Status != "candidate" || r.Reason != "language_unvalidated" || r.Suggested != "example/auth" {
		t.Fatal("query-language priority changed")
	}
	idx = contractIndex(t, []reporouter.Repository{{Name: "example/auth", Summary: "Login 로그인", Keywords: []string{"latin"}}})
	expectMetadataBlock(t, idx, reporouter.DefaultConfig())
	// Existing summary guards must also cover every selected position. All
	// summaries have two equal-length terms, so the blocking repository does
	// not change its lexical position when the second term is replaced.
	for count := 1; count <= 8; count++ {
		c := reporouter.DefaultConfig()
		c.Candidates = count
		for selected := 0; selected < count; selected++ {
			repos := make([]reporouter.Repository, 8)
			for i := range repos {
				summary := "Login marker"
				if i == selected {
					summary = "Login 로그인"
				}
				repos[i] = reporouter.Repository{Name: fmt.Sprintf("example/auth-%d", i), Summary: summary, Keywords: []string{"latin"}}
			}
			idx := contractIndex(t, repos)
			lexical := lexicalResult(t, idx, c)
			if len(lexical.Candidates) != count || lexical.Candidates[selected].Name != repos[selected].Name {
				t.Fatal("summary-position fixture changed")
			}
			expectMetadataBlock(t, idx, c)
		}
	}
}
func TestPreviewContractEnglishJudgmentGuards(t *testing.T) {
	idx := contractIndex(t, []reporouter.Repository{{Name: "example/auth", Summary: "Login sessions", Keywords: []string{"Résumé", "123", "login", "!!!"}}})
	c := reporouter.DefaultConfig()
	for _, tc := range []struct {
		p              reporouter.Judgment
		failure        error
		status, reason string
	}{
		{reporouter.Judgment{Probabilities: []float64{.95, .05}}, nil, "suggest", "experimental_laya_choice"},
		{reporouter.Judgment{Probabilities: []float64{.89, .11}}, nil, "abstain", "confidence_or_margin"},
		{reporouter.Judgment{Probabilities: []float64{.02, .98}}, nil, "abstain", "none_or_ambiguous"},
		{reporouter.Judgment{Probabilities: []float64{.99, .01}, Truncated: true}, nil, "abstain", "input_truncated"},
		{reporouter.Judgment{Probabilities: []float64{1}}, nil, "abstain", "invalid_judgment"},
		{reporouter.Judgment{Probabilities: []float64{math.NaN(), 0}}, nil, "abstain", "invalid_judgment"},
		{reporouter.Judgment{Probabilities: []float64{.8, .3}}, nil, "abstain", "invalid_judgment"},
		{reporouter.Judgment{}, errors.New("authored unavailable judge"), "candidate", "inference_unavailable"},
	} {
		j := &contractJudge{judgment: tc.p, failure: tc.failure}
		got, err := idx.Preview("login", c, j)
		if err != nil || j.calls != 1 || got.Status != tc.status || got.Reason != tc.reason || !got.Preview {
			t.Fatal("English judgment behavior changed")
		}
		if got.Status == "abstain" && got.Suggested != "" {
			t.Fatal("abstention retained recommendation")
		}
	}
	c.Margin = .95
	j := &contractJudge{judgment: reporouter.Judgment{Probabilities: []float64{.95, .05}}}
	r, err := idx.Preview("login", c, j)
	if err != nil || j.calls != 1 || r.Status != "abstain" || r.Reason != "confidence_or_margin" {
		t.Fatal("margin guard changed")
	}
}
