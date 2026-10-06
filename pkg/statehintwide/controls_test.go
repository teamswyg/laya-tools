package statehintwide

import (
	"bytes"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"testing"
)

func TestArtifactModesAndCrossVersion(t *testing.T) {
	for _, mode := range []FeatureMode{Simple, Contextual} {
		m := NewModel(mode)
		var b bytes.Buffer
		if e := m.Save(&b); e != nil {
			t.Fatal(e)
		}
		if b.Len() != 65728 {
			t.Fatal("unexpected capacity")
		}
		loaded, e := Load(bytes.NewReader(b.Bytes()))
		if e != nil || loaded.mode != mode {
			t.Fatal("lost feature semantics", e)
		}
		if _, e = statehint.Load(bytes.NewReader(b.Bytes())); e == nil {
			t.Fatal("v1 accepted v2")
		}
		corrupt := append([]byte(nil), b.Bytes()...)
		corrupt[len(corrupt)-1] ^= 1
		if _, e = Load(bytes.NewReader(corrupt)); e == nil {
			t.Fatal("corrupted artifact accepted")
		}
	}
	var old bytes.Buffer
	if e := statehint.NewModel().Save(&old); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(&old); e == nil {
		t.Fatal("v2 accepted v1")
	}
}
func TestOrderedContextAndWorkspaceReuse(t *testing.T) {
	var a, b Workspace
	if e := extract("not yet done", &a, Contextual); e != nil {
		t.Fatal(e)
	}
	if e := extract("done not yet", &b, Contextual); e != nil {
		t.Fatal(e)
	}
	if a.values == b.values {
		t.Fatal("lost word order")
	}
	if e := extract("a completed report", &a, Contextual); e != nil {
		t.Fatal(e)
	}
	var clean Workspace
	if e := extract("a completed report", &clean, Contextual); e != nil {
		t.Fatal(e)
	}
	if a.values != clean.values {
		t.Fatal("workspace leaked previous input")
	}
	for _, v := range a.values {
		if !finite(float64(v)) {
			t.Fatal("nonfinite feature")
		}
	}
}
func TestKnownCapacityCollisionSplit(t *testing.T) {
	wordhash := func(s string) uint32 {
		h := uint32(2166136261)
		h = (h ^ uint32('w')) * 16777619
		return hashRunes(h, []rune(s))
	}
	a, b := wordhash("started"), wordhash("submitted")
	if a%1024 != b%1024 || a%2048 == b%2048 {
		t.Fatal("known capacity ablation identity changed")
	}
}
func TestActualOwnedToyFitAndReload(t *testing.T) {
	m := NewModel(Contextual)
	rows := []Sample{{Text: "please explain the answer", Label: Question}, {Text: "please explain the detail", Label: Question}, {Text: "work has been completed", Label: CompletionReport}, {Text: "work is now completed", Label: CompletionReport}}
	if _, e := m.Fit(rows, FitOptions{Epochs: 30, LearningRate: .02, BatchSize: 4, Seed: 1729}); e != nil {
		t.Fatal(e)
	}
	var buf bytes.Buffer
	if e := m.Save(&buf); e != nil {
		t.Fatal(e)
	}
	copy, e := Load(&buf)
	if e != nil {
		t.Fatal(e)
	}
	var a, b Workspace
	for _, r := range rows {
		p, e := m.Predict(r.Text, &a)
		if e != nil {
			t.Fatal(e)
		}
		q, e := copy.Predict(r.Text, &b)
		if e != nil || p != q {
			t.Fatal("reload parity", e)
		}
	}
}
