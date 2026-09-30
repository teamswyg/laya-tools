package researchbundle

import (
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/claimtree"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"testing"
)

func TestShallowContractAndTopology(t *testing.T) {
	m := claimtree.Model{Nodes: 1}
	m.Feature[0] = -1
	a := retrievalbench.ShallowCandidate{EvaluationRepository: claimRepos[0], Depth: 1, MinLeaf: 16, Nodes: 1}
	h := retrievalbench.ShallowHead{Schema: "riido-shallow-claim-v1", PlanSHA256: retrievalbench.ShallowClaimPlanSHA256, TrainingRepository: claimRepos[1], ValidationRepository: claimRepos[2], EvaluationRepository: claimRepos[0], Depth: 1, MinLeaf: 16, Tree: m}
	b, _ := json.Marshal(h)
	if e := verifyShallowHead(b, a); e != nil {
		t.Fatal(e)
	}
	h.Tree.Feature[0] = 0
	b, _ = json.Marshal(h)
	if e := verifyShallowHead(b, a); e == nil {
		t.Fatal("accepted cycle")
	}
	h.Tree = m
	h.TrainingRepository = claimRepos[0]
	b, _ = json.Marshal(h)
	if e := verifyShallowHead(b, a); e == nil {
		t.Fatal("accepted heldout as training")
	}
}
