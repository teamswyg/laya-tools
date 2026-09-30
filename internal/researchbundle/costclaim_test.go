package researchbundle

import (
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"testing"
)

func TestCostHeadPolicyAndFoldIntegrity(t *testing.T) {
	a := retrievalbench.CostCandidate{EvaluationRepository: claimRepos[0], Seed: 1729, Epoch: 1, Penalty: .25, ValidationWeightedNLL: .5, Threshold: retrievalbench.BudgetThreshold{BudgetPercent: 90, ValidationQuestions: 10, ValidationCalls: 5, ValidationPages: 20}}
	h := retrievalbench.CostClaimHead{Head: retrievalbench.ClaimHead{Schema: "riido-cost-claim-v1", Mode: "fp32", TrainingRepository: claimRepos[1], ValidationRepository: claimRepos[2], EvaluationRepository: claimRepos[0], PlanSHA256: retrievalbench.CostClaimPlanSHA256, Seed: 1729, Epoch: 1, ValidationNLL: .5, Weights: make([]float64, 12)}, Penalty: .25, Threshold: a.Threshold}
	b, _ := json.Marshal(h)
	if err := verifyCostHead(b, a); err != nil {
		t.Fatal(err)
	}
	h.Penalty = 1
	b, _ = json.Marshal(h)
	if verifyCostHead(b, a) == nil {
		t.Fatal("accepted changed penalty")
	}
	h.Penalty = .25
	h.Threshold.MinimumScore = 1
	b, _ = json.Marshal(h)
	if verifyCostHead(b, a) == nil {
		t.Fatal("accepted changed threshold")
	}
	h.Threshold = a.Threshold
	h.Head.TrainingRepository = claimRepos[0]
	b, _ = json.Marshal(h)
	if verifyCostHead(b, a) == nil {
		t.Fatal("accepted training on held-out repo")
	}
	h.Head.TrainingRepository = claimRepos[1]
	b, _ = json.Marshal(h)
	b = append(b[:len(b)-1], []byte(",\"private_text\":\"unexpected\"}")...)
	if verifyCostHead(b, a) == nil {
		t.Fatal("accepted extra field")
	}
}
