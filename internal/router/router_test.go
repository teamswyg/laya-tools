package router

import (
	"errors"
	"github.com/teamswyg/laya-tools/internal/inference"
	"math"
	"strings"
	"testing"
)

type fake struct {
	p     inference.Prediction
	err   error
	calls int
}

func (f *fake) Predict(string, string, string, []string) (inference.Prediction, error) {
	f.calls++
	return f.p, f.err
}
func TestRoutingPolicy(t *testing.T) {
	c := Config{Fast: "fast", Standard: "middle", Strong: "strong", Threshold: .9}
	for _, tc := range []struct {
		name, prompt, override string
		p                      inference.Prediction
		err                    error
		model                  string
		abstain                bool
		calls                  int
	}{
		{"explicit", "complex migration", "chosen", inference.Prediction{}, nil, "chosen", false, 0},
		{"korean", "로그인 수정", "", inference.Prediction{}, nil, "strong", true, 0},
		{"uncertain", "edit", "", inference.Prediction{Probabilities: []float64{.6, .3, .1}}, nil, "strong", true, 1},
		{"confident", "edit", "", inference.Prediction{Probabilities: []float64{.95, .03, .02}}, nil, "fast", false, 1},
		{"truncated", "edit", "", inference.Prediction{Probabilities: []float64{.95, .03, .02}, Truncated: true}, nil, "strong", true, 1},
		{"failed", "edit", "", inference.Prediction{}, errors.New("native failure"), "strong", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &fake{p: tc.p, err: tc.err}
			r, err := Route(tc.prompt, tc.override, c, s)
			if err != nil || r.Model != tc.model || r.Abstained != tc.abstain || s.calls != tc.calls {
				t.Fatalf("%+v %v calls %d", r, err, s.calls)
			}
		})
	}
	if _, err := Route(strings.Repeat("x", 65537), "", c, nil); err == nil {
		t.Fatal("accepted oversized prompt")
	}
}

func TestRejectNonfiniteThreshold(t *testing.T) {
	for _, v := range []float64{math.NaN(), math.Inf(1)} {
		c := Config{Threshold: v}
		if NeedsInference("test", "", c) {
			t.Fatal("nonfinite threshold attempts model load")
		}
		if _, err := Route("test", "", c, nil); err == nil {
			t.Fatal("nonfinite threshold accepted")
		}
	}
}
