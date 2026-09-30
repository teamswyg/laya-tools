package main

import (
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"strings"
	"testing"
)

type fakeBuilder struct{ calls int }

func (f *fakeBuilder) Build(string, string, string, []string, int) (inference.Sequence, error) {
	f.calls++
	return inference.Sequence{IDs: []int64{1, 2, 3}, Markers: []int64{1, 2}}, nil
}
func TestPrepareOnlyScopedValidation(t *testing.T) {
	rows := []paireval.Row{{Query: "a", Code: "b"}, {Query: "reserve", Code: "hidden"}, {Query: "long", Code: strings.Repeat("x ", 65)}}
	split := paireval.Split{Rows: []paireval.Membership{{Split: "validation"}, {Split: "reserve1"}, {Split: "validation"}}}
	b := &fakeBuilder{}
	out, err := prepare(rows, split, b)
	if err != nil || len(out.Cases) != 2 || b.calls != 2 {
		t.Fatalf("bad isolation %v %d %d", err, len(out.Cases), b.calls)
	}
	if out.Cases[0].Options[0] != out.Cases[1].Options[1] || out.Cases[0].State != "File: [unavailable]\nb" {
		t.Fatal("wrong protocol")
	}
	// Labels are deliberately nil. Neither preparation nor exported requests need them.
}
func TestReplayRejectsWrongIdentity(t *testing.T) {
	req := request{State: "state", Kind: "noul", Instruction: "instruction", Options: []string{"no", "yes"}, IDs: []int64{1, 2}}
	pred := inference.Prediction{Tokens: 2, Probabilities: []float64{.2, .8}}
	r := &replay{requests: []request{req}, scores: []response{{hash(req), pred}}}
	if _, err := r.Predict(req.State, req.Kind, req.Instruction, []string{"yes", "no"}); err == nil {
		t.Fatal("accepted wrong order")
	}
	if r.at != 0 {
		t.Fatal("advanced on rejection")
	}
	if _, err := r.Predict(req.State, req.Kind, req.Instruction, req.Options); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Predict(req.State, req.Kind, req.Instruction, req.Options); err == nil {
		t.Fatal("reused prediction")
	}
	r.at = 0
	r.scores[0].RequestSHA256 = "changed"
	if _, err := r.Predict(req.State, req.Kind, req.Instruction, req.Options); err == nil {
		t.Fatal("accepted changed identity")
	}
	r.scores[0].RequestSHA256 = hash(req)
	r.scores[0].Prediction.Truncated = true
	if _, err := r.Predict(req.State, req.Kind, req.Instruction, req.Options); err == nil {
		t.Fatal("accepted changed truncation")
	}
}
