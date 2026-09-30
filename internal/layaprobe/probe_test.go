package layaprobe

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"math"
	"strings"
	"testing"
)

type fakePredictor struct{ calls int }

func (p *fakePredictor) Predict(state, kind, instruction string, options []string) (inference.Prediction, error) {
	if strings.Contains(state, "RESERVE") || kind != "choice" || instruction != Instruction {
		return inference.Prediction{}, fmt.Errorf("wrong input scope")
	}
	index := p.calls % 2
	if len(options) != 2 || options[0] != choices[index] || options[1] != choices[1-index] {
		return inference.Prediction{}, fmt.Errorf("wrong option order")
	}
	probabilities := [][2]float64{{.8, .2}, {.4, .6}, {.1, .9}, {.8, .2}, {.8, .2}, {.4, .6}}
	v := probabilities[p.calls]
	p.calls++
	return inference.Prediction{Probabilities: []float64{v[0], v[1]}, Truncated: p.calls >= 5, MS: 1}, nil
}
func TestValidationOnlyOrderAndMatchedExclusions(t *testing.T) {
	pos, neg := 1, 0
	rows := []paireval.Row{{Code: "read file", Query: "dev"}, {Code: "read file", Query: "read", Label: &pos}, {Code: "write", Query: "read", Label: &neg}, {Code: "x", Query: "RESERVE"}, {Code: "truncated", Query: "read", Label: &pos}, {Code: strings.Repeat("x", 4097), Query: "large", Label: &pos}}
	split := paireval.Split{Rows: []paireval.Membership{{Split: "development"}, {Split: "validation"}, {Split: "validation"}, {Split: "reserve1"}, {Split: "validation"}, {Split: "validation"}}}
	fake := &fakePredictor{}
	r, err := Run(rows, split, fake, func() error { return nil }, func(int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	if r.ValidationCases != 4 || r.ScopeExcluded != 1 || r.TokenizerExcluded != 1 || r.Eligible != 2 || r.NativeCalls != 6 || r.PrimaryAUC == nil || *r.PrimaryAUC != 1 || math.Abs(r.MeanAbsoluteOrderGap-.15) > 1e-9 {
		t.Fatalf("wrong result: %+v", r)
	}
}
func TestGuardAndProbabilityFailures(t *testing.T) {
	pos := 1
	rows := []paireval.Row{{Code: "x", Query: "y", Label: &pos}}
	split := paireval.Split{Rows: []paireval.Membership{{Split: "validation"}}}
	fake := &fakePredictor{}
	if _, err := Run(rows, split, fake, func() error { return fmt.Errorf("limit") }, func(int, int) {}); err == nil || fake.calls != 0 {
		t.Fatal("guard ignored")
	}
	for _, values := range [][]float64{{}, {1}, {.5, .6}, {math.NaN(), .5}, {-.1, 1.1}} {
		if _, err := probability(inference.Prediction{Probabilities: values}, 0); err == nil {
			t.Fatal("invalid probability accepted")
		}
	}
}

type documentedFake struct{ calls int }

func (p *documentedFake) Predict(state, kind, instruction string, options []string) (inference.Prediction, error) {
	if kind != "noul" || !strings.HasPrefix(state, "File: [unavailable]\n") || !strings.Contains(instruction, "Is this source code relevant to the software change:") {
		return inference.Prediction{}, fmt.Errorf("wrong documented protocol")
	}
	if p.calls%2 == 0 && options[0] != "false: no, the statement does not hold" {
		return inference.Prediction{}, fmt.Errorf("wrong documented order")
	}
	probs := [][2]float64{{.1, .9}, {.1, .9}, {.8, .2}, {.8, .2}}
	v := probs[p.calls]
	p.calls++
	return inference.Prediction{Probabilities: []float64{v[0], v[1]}}, nil
}
func TestDocumentedPrimaryIsForwardNotAverage(t *testing.T) {
	yes, no := 1, 0
	rows := []paireval.Row{{Code: "a", Query: "a", Label: &yes}, {Code: "b", Query: "b", Label: &no}}
	split := paireval.Split{Rows: []paireval.Membership{{Split: "validation"}, {Split: "validation"}}}
	r, err := RunProtocol(rows, split, &documentedFake{}, func() error { return nil }, func(int, int) {}, "noul")
	if err != nil {
		t.Fatal(err)
	}
	if r.PlanSHA256 != DocumentedPlanSHA256 || r.PrimaryAUC == nil || *r.PrimaryAUC != 1 || r.ReverseAUC == nil || *r.ReverseAUC != 0 || r.AveragedAUC == nil || *r.AveragedAUC != .5 {
		t.Fatal("documented primary mapping wrong", r)
	}
}
