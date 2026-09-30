package researchbundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const claimResultsSHA = "11da37b39f1c6e4a507df2b83decf5f539ad2aae148466c3ea60a98d627d5359"
const claimPlanSHA = "ae516f0bf1019453da32de70f63a72cfc7f6291efb45db811a222d0650f77725"
const pageClaimResultsSHA = "bfaa3a78cf231c1c85814a5570f95f38c640ca58a2a62ed01f71d13bcb2b55cc"

var claimRepo = regexp.MustCompile(`^JooYoon/riidolaya-search-claim-research-v[0-9]+\.[0-9]+$`)
var pageClaimRepo = regexp.MustCompile(`^JooYoon/riidolaya-page-claim-research-v[0-9]+\.[0-9]+$`)
var claimRepos = []string{"etcd-io/etcd", "kubernetes/test-infra", "lxc/lxd"}

func verifyClaimHead(b []byte, a retrievalbench.ClaimModelResult) error {
	return verifyClaimHeadContract(b, a, searchclaim.Schema, claimPlanSHA)
}

func verifyClaimHeadContract(b []byte, a retrievalbench.ClaimModelResult, schema, plan string) error {
	var h retrievalbench.ClaimHead
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&h); e != nil {
		return e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fmt.Errorf("extra head content")
	}
	fold := -1
	for i, r := range claimRepos {
		if r == a.EvaluationRepository {
			fold = i
		}
	}
	if fold < 0 {
		return fmt.Errorf("unknown fold")
	}
	if h.Schema != schema || h.Mode != a.Mode || h.Seed != a.Seed || h.PlanSHA256 != plan || h.EvaluationRepository != a.EvaluationRepository || h.TrainingRepository != claimRepos[(fold+1)%3] || h.ValidationRepository != claimRepos[(fold+2)%3] || h.Epoch != a.Epoch || h.ValidationNLL != a.ValidationNLL || len(h.Weights) != searchclaim.Dimension {
		return fmt.Errorf("head contract mismatch")
	}
	if h.Epoch < 1 || h.Epoch > 100 || math.IsNaN(h.ValidationNLL) || math.IsInf(h.ValidationNLL, 0) {
		return fmt.Errorf("invalid training metadata")
	}
	scale := 0.
	for _, w := range h.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || math.Abs(w) > 1e6 {
			return fmt.Errorf("invalid coefficient")
		}
		if strings.HasPrefix(h.Mode, "ternary_") && w != 0 {
			if scale == 0 {
				scale = math.Abs(w)
			}
			if math.Abs(w) != scale {
				return fmt.Errorf("nonternary coefficient")
			}
		}
	}
	return nil
}

// VerifySearchClaim binds numerical exports to reviewed immutable results. It
// checks publication integrity, not corpus metrics or production usefulness.
func VerifySearchClaim(dir string, m Manifest) (Manifest, error) {
	return verifyClaimBundle(dir, m, false)
}

func VerifyPageClaim(dir string, m Manifest) (Manifest, error) {
	return verifyClaimBundle(dir, m, true)
}

func verifyClaimBundle(dir string, m Manifest, page bool) (Manifest, error) {
	schema, origin, headSchema := "riido-search-claim-bundle-v1", "public_codesearchnet_runtime_claim_09", searchclaim.Schema
	planHash, resultHash, allowedRepo := claimPlanSHA, claimResultsSHA, claimRepo
	if page {
		schema, origin, headSchema = "riido-page-claim-bundle-v1", "public_codesearchnet_page_claim_14", "riido-page-claim-v1"
		planHash, resultHash, allowedRepo = retrievalbench.PageClaimPlanSHA256, pageClaimResultsSHA, pageClaimRepo
	}
	if m.Schema != schema || !allowedRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "none" || m.Origin != origin || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != 30 {
		return m, fmt.Errorf("invalid claim manifest")
	}
	b, e := read(filepath.Join(dir, "results.json"), 1<<20)
	if e != nil {
		return m, e
	}
	if paireval.Hash(b) != resultHash {
		return m, fmt.Errorf("unreviewed results")
	}
	var r retrievalbench.ClaimReport
	if page {
		var p retrievalbench.PageClaimReport
		if e = json.Unmarshal(b, &p); e != nil {
			return m, e
		}
		if p.PageSize != 20 || p.ReplicationPassesExploratoryGate {
			return m, fmt.Errorf("unexpected page experiment state")
		}
		r.Questions, r.Dimension, r.PlanSHA256 = p.Questions, p.Dimension, p.PlanSHA256
		r.PrimaryPassesExploratoryGate, r.ProductionReady = p.PrimaryPassesExploratoryGate, p.ProductionReady
		for _, a := range p.Models {
			r.Models = append(r.Models, retrievalbench.ClaimModelResult{File: a.File, SHA256: a.SHA256, Mode: a.Mode, EvaluationRepository: a.EvaluationRepository, Bytes: a.Bytes, Seed: a.Seed, Epoch: a.Epoch, ValidationNLL: a.ValidationWeightedNLL})
		}
	} else {
		if e = json.Unmarshal(b, &r); e != nil {
			return m, e
		}
	}
	if len(r.Models) != 24 || r.Questions != 2948 || r.Dimension != 12 || r.PlanSHA256 != planHash || r.PrimaryPassesExploratoryGate || r.ProductionReady {
		return m, fmt.Errorf("unexpected experiment state")
	}
	expected := []string{"LICENSE", "NOTICE", "README.md", "README.ko.md", "results.json", "plan.json"}
	seen := map[string]bool{}
	for _, a := range r.Models {
		fold := -1
		for i, repo := range claimRepos {
			if repo == a.EvaluationRepository {
				fold = i
			}
		}
		if fold < 0 || (a.Seed != 1729 && a.Seed != 2718) {
			return m, fmt.Errorf("invalid fold/seed")
		}
		switch a.Mode {
		case "fp32", "int8", "ternary_ptq", "ternary_ste":
		default:
			return m, fmt.Errorf("invalid mode")
		}
		name := fmt.Sprintf("fold%d-%s-%d.claim.json", fold, a.Mode, a.Seed)
		if name != a.File || seen[name] {
			return m, fmt.Errorf("invalid model name")
		}
		seen[name] = true
		expected = append(expected, name)
		b, e := read(filepath.Join(dir, name), 1<<16)
		if e != nil {
			return m, e
		}
		if len(b) != a.Bytes || paireval.Hash(b) != a.SHA256 {
			return m, fmt.Errorf("head does not match reviewed export")
		}
		if e = verifyClaimHeadContract(b, a, headSchema, planHash); e != nil {
			return m, e
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return m, e
	}
	if len(entries) != len(expected)+1 {
		return m, fmt.Errorf("extra release files")
	}
	for _, name := range expected {
		b, e := read(filepath.Join(dir, name), 1<<20)
		if e != nil {
			return m, e
		}
		if m.Files[name] != paireval.Hash(b) {
			return m, fmt.Errorf("manifest hash mismatch")
		}
		switch name {
		case "LICENSE":
			if m.Files[name] != "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9" {
				return m, fmt.Errorf("license mismatch")
			}
		case "plan.json":
			if m.Files[name] != planHash {
				return m, fmt.Errorf("plan mismatch")
			}
		case "NOTICE":
			for _, repo := range claimRepos {
				if !strings.Contains(string(b), repo) {
					return m, fmt.Errorf("missing source attribution")
				}
			}
		case "README.md":
			if !strings.Contains(string(b), "Not production ready") || !strings.Contains(string(b), "No LLM savings established") {
				return m, fmt.Errorf("missing limitations")
			}
		}
	}
	return m, nil
}
