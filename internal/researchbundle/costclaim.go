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

const spreadClaimResultsSHA = "ddb7b2546ec1c9740489bf29168e7d7b0b34692ebeb47251d34649b483195a1a"

var spreadClaimRepo = regexp.MustCompile(`^JooYoon/riidolaya-score-spread-research-v[0-9]+\.[0-9]+$`)

const coverageClaimResultsSHA = "580f822349881d8cf278937df40844d1de66a5270d3ad0909d9eac5b94fbe9fb"

var coverageClaimRepo = regexp.MustCompile(`^JooYoon/riidolaya-coverage-claim-research-v[0-9]+\.[0-9]+$`)

type claimBundleKind uint8

const (
	costBundle claimBundleKind = iota
	spreadBundle
	coverageBundle
)

func verifyCostHead(b []byte, a retrievalbench.CostCandidate) error {
	return verifyCostHeadShape(b, a, "riido-cost-claim-v1", retrievalbench.CostClaimPlanSHA256, 12)
}
func verifyCostHeadShape(b []byte, a retrievalbench.CostCandidate, schema, plan string, dimension int) error {
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
	return verifyClaimHeadDimension(raw, retrievalbench.ClaimModelResult{Mode: "fp32", Seed: a.Seed, Epoch: a.Epoch, EvaluationRepository: a.EvaluationRepository, ValidationNLL: a.ValidationWeightedNLL}, schema, plan, dimension)
}

// VerifyCostClaim checks the exact reviewed research exports, not their utility.
func VerifyCostClaim(dir string, m Manifest) (Manifest, error) {
	return verifyCostBundle(dir, m, costBundle)
}
func VerifySpreadClaim(dir string, m Manifest) (Manifest, error) {
	return verifyCostBundle(dir, m, spreadBundle)
}
func VerifyCoverageClaim(dir string, m Manifest) (Manifest, error) {
	return verifyCostBundle(dir, m, coverageBundle)
}
func verifyCostBundle(dir string, m Manifest, kind claimBundleKind) (Manifest, error) {
	schema, origin, headSchema, reportSchema := "riido-cost-claim-bundle-v1", "public_codesearchnet_cost_claim_20", "riido-cost-claim-v1", "riido-cost-claim-report-v1"
	plan, resultHash, allowedRepo, dimension := retrievalbench.CostClaimPlanSHA256, costClaimResultsSHA, costClaimRepo, 12
	switch kind {
	case costBundle:
	case spreadBundle:
		schema = "riido-spread-claim-bundle-v1"
		origin = "public_codesearchnet_score_spread_21"
		headSchema = "riido-spread-claim-v1"
		reportSchema = "riido-spread-claim-report-v1"
		plan = retrievalbench.SpreadClaimPlanSHA256
		resultHash = spreadClaimResultsSHA
		allowedRepo = spreadClaimRepo
		dimension = 16
	case coverageBundle:
		schema = "riido-coverage-claim-bundle-v1"
		origin = "public_codesearchnet_coverage_claim_25"
		headSchema = "riido-coverage-claim-v1"
		reportSchema = "riido-coverage-claim-report-v1"
		plan = retrievalbench.CoverageClaimPlanSHA256
		resultHash = coverageClaimResultsSHA
		allowedRepo = coverageClaimRepo
		dimension = 20
	default:
		return m, fmt.Errorf("unknown claim bundle kind")
	}

	if m.Schema != schema || !allowedRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "none" || m.Origin != origin || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != 36 {
		return m, fmt.Errorf("invalid cost claim manifest")
	}
	b, err := read(filepath.Join(dir, "results.json"), 1<<20)
	if err != nil {
		return m, err
	}
	if paireval.Hash(b) != resultHash {
		return m, fmt.Errorf("unreviewed results")
	}
	var r retrievalbench.CostClaimReport
	if err = json.Unmarshal(b, &r); err != nil {
		return m, err
	}
	if r.Schema != reportSchema || r.PlanSHA256 != plan || r.Questions != 2948 || r.Dimension != dimension || r.PageSize != 20 || len(r.Candidates) != 30 || r.PrimaryPassesExploratoryGate || r.ReplicationPassesExploratoryGate || r.ProductionReady {
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
		if err = verifyCostHeadShape(data, a, headSchema, plan, dimension); err != nil {
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
			if m.Files[name] != plan {
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
