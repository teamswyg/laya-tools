package researchbundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
)

const costClaimResultsSHA = "5393cfc9caeca2580bffe2cae05cb61ffadec068c30a5a8bf422fdeb2ad92ca1"

var costClaimRepo = regexp.MustCompile(`^JooYoon/riidolaya-cost-claim-research-v[0-9]+\.[0-9]+$`)

func verifyCostHead(b []byte, a retrievalbench.CostCandidate) error {
	var h retrievalbench.CostClaimHead
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&h); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("extra head content")
	}
	if h.Penalty != a.Penalty || h.Threshold != a.Threshold {
		return fmt.Errorf("cost policy metadata mismatch")
	}
	raw, err := json.Marshal(h.Head)
	if err != nil {
		return err
	}
	return verifyClaimHeadContract(raw, retrievalbench.ClaimModelResult{Mode: "fp32", Seed: a.Seed, Epoch: a.Epoch, EvaluationRepository: a.EvaluationRepository, ValidationNLL: a.ValidationWeightedNLL}, "riido-cost-claim-v1", retrievalbench.CostClaimPlanSHA256)
}

// VerifyCostClaim checks the exact reviewed research exports, not their utility.
func VerifyCostClaim(dir string, m Manifest) (Manifest, error) {
	if m.Schema != "riido-cost-claim-bundle-v1" || !costClaimRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "none" || m.Origin != "public_codesearchnet_cost_claim_20" || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != 36 {
		return m, fmt.Errorf("invalid cost claim manifest")
	}
	b, err := read(filepath.Join(dir, "results.json"), 1<<20)
	if err != nil {
		return m, err
	}
	if paireval.Hash(b) != costClaimResultsSHA {
		return m, fmt.Errorf("unreviewed results")
	}
	var r retrievalbench.CostClaimReport
	if err = json.Unmarshal(b, &r); err != nil {
		return m, err
	}
	if r.Schema != "riido-cost-claim-report-v1" || r.PlanSHA256 != retrievalbench.CostClaimPlanSHA256 || r.Questions != 2948 || r.Dimension != 12 || r.PageSize != 20 || len(r.Candidates) != 30 || r.PrimaryPassesExploratoryGate || r.ReplicationPassesExploratoryGate || r.ProductionReady {
		return m, fmt.Errorf("unexpected cost experiment state")
	}
	expected := []string{"LICENSE", "NOTICE", "README.md", "README.ko.md", "results.json", "plan.json"}
	for _, a := range r.Candidates {
		fold := slices.Index(claimRepos, a.EvaluationRepository)
		pi := slices.Index([]float64{0, .25, 1, 4, 16}, a.Penalty)
		if fold < 0 || pi < 0 || (a.Seed != 1729 && a.Seed != 2718) {
			return m, fmt.Errorf("invalid candidate metadata")
		}
		name := fmt.Sprintf("fold%d-penalty%d-fp32-%d.claim.json", fold, pi, a.Seed)
		if a.File != name || slices.Contains(expected, name) {
			return m, fmt.Errorf("invalid candidate filename")
		}
		expected = append(expected, name)
		data, err := read(filepath.Join(dir, name), 1<<16)
		if err != nil {
			return m, err
		}
		if len(data) != a.Bytes || paireval.Hash(data) != a.SHA256 {
			return m, fmt.Errorf("head does not match reviewed export")
		}
		if err = verifyCostHead(data, a); err != nil {
			return m, err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return m, err
	}
	if len(entries) != 37 {
		return m, fmt.Errorf("extra release files")
	}
	for _, name := range expected {
		data, err := read(filepath.Join(dir, name), 1<<20)
		if err != nil {
			return m, err
		}
		if m.Files[name] != paireval.Hash(data) {
			return m, fmt.Errorf("manifest hash mismatch")
		}
		switch name {
		case "LICENSE":
			if m.Files[name] != "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9" {
				return m, fmt.Errorf("license mismatch")
			}
		case "plan.json":
			if m.Files[name] != retrievalbench.CostClaimPlanSHA256 {
				return m, fmt.Errorf("plan mismatch")
			}
		case "NOTICE":
			for _, repo := range claimRepos {
				if !strings.Contains(string(data), repo) {
					return m, fmt.Errorf("missing source attribution")
				}
			}
		case "README.md":
			if !strings.Contains(string(data), "Not production ready") || !strings.Contains(string(data), "No LLM savings established") {
				return m, fmt.Errorf("missing limitations")
			}
		}
	}
	return m, nil
}
