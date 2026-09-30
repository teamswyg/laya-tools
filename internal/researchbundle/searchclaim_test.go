package researchbundle

import (
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaimHeadRejectsTextAndWrongFold(t *testing.T) {
	a := retrievalbench.ClaimModelResult{Mode: "ternary_ste", EvaluationRepository: claimRepos[0], Seed: 1729, Epoch: 5, ValidationNLL: .6}
	h := retrievalbench.ClaimHead{Schema: searchclaim.Schema, Mode: a.Mode, Seed: a.Seed, Epoch: 5, ValidationNLL: .6, EvaluationRepository: claimRepos[0], TrainingRepository: claimRepos[1], ValidationRepository: claimRepos[2], PlanSHA256: claimPlanSHA, Weights: make([]float64, 12)}
	h.Weights[0] = .1
	b, _ := json.Marshal(h)
	if e := verifyClaimHead(b, a); e != nil {
		t.Fatal(e)
	}
	var fields map[string]any
	json.Unmarshal(b, &fields)
	fields["Query"] = "raw corpus text"
	bad, _ := json.Marshal(fields)
	if e := verifyClaimHead(bad, a); e == nil {
		t.Fatal("accepted raw text")
	}
	h.TrainingRepository = h.EvaluationRepository
	bad, _ = json.Marshal(h)
	if e := verifyClaimHead(bad, a); e == nil {
		t.Fatal("accepted wrong fold")
	}
	h.TrainingRepository = claimRepos[1]
	h.Weights[1] = .2
	bad, _ = json.Marshal(h)
	if e := verifyClaimHead(bad, a); e == nil {
		t.Fatal("accepted nonternary head")
	}
}

func TestManifestRejectsUnknownContent(t *testing.T) {
	dir := t.TempDir()
	if e := os.WriteFile(filepath.Join(dir, "MANIFEST.json"), []byte(`{"raw_query":"unexpected source text"}`), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Verify(dir); e == nil || !strings.Contains(e.Error(), "unknown field") {
		t.Fatalf("unexpected result %v", e)
	}
}
