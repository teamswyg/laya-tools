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

const shallowClaimResultsSHA = "366c1da0e50cdc043f6c0c698cb8f9ce9eeb481c60249033900392d309bdcdba"

var shallowClaimRepo = regexp.MustCompile(`^JooYoon/riidolaya-shallow-claim-research-v[0-9]+\.[0-9]+$`)

func verifyShallowHead(b []byte, a retrievalbench.ShallowCandidate) error {
	var h retrievalbench.ShallowHead
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&h); e != nil {
		return e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fmt.Errorf("extra head content")
	}
	fold := slices.Index(claimRepos, a.EvaluationRepository)
	if fold < 0 || h.Schema != "riido-shallow-claim-v1" || h.PlanSHA256 != retrievalbench.ShallowClaimPlanSHA256 || h.TrainingRepository != claimRepos[(fold+1)%3] || h.ValidationRepository != claimRepos[(fold+2)%3] || h.EvaluationRepository != a.EvaluationRepository || h.Depth != a.Depth || h.MinLeaf != a.MinLeaf || h.Tree.Nodes != a.Nodes || h.Threshold != a.Threshold {
		return fmt.Errorf("tree contract mismatch")
	}
	return h.Tree.Validate()
}

// VerifyShallowClaim verifies exact research exports, not practical utility.
func VerifyShallowClaim(dir string, m Manifest) (Manifest, error) {
	if m.Schema != "riido-shallow-claim-bundle-v1" || !shallowClaimRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "none" || m.Origin != "public_codesearchnet_shallow_claim_22" || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != 30 {
		return m, fmt.Errorf("invalid shallow manifest")
	}
	b, err := read(filepath.Join(dir, "results.json"), 1<<20)
	if err != nil {
		return m, err
	}
	if paireval.Hash(b) != shallowClaimResultsSHA {
		return m, fmt.Errorf("unreviewed results")
	}
	var r retrievalbench.ShallowReport
	if err = json.Unmarshal(b, &r); err != nil {
		return m, err
	}
	if r.Schema != "riido-shallow-claim-report-v1" || r.PlanSHA256 != retrievalbench.ShallowClaimPlanSHA256 || r.ControlManifestSHA256 != retrievalbench.ShallowControlManifestSHA256 || r.Questions != 2948 || len(r.Candidates) != 24 || r.PrimaryPassesExploratoryGate || r.ProductionReady {
		return m, fmt.Errorf("unexpected experiment state")
	}
	expected := []string{"LICENSE", "NOTICE", "README.md", "README.ko.md", "results.json", "plan.json"}
	for _, a := range r.Candidates {
		fold := slices.Index(claimRepos, a.EvaluationRepository)
		if fold < 0 || a.Depth < 1 || a.Depth > 4 || (a.MinLeaf != 16 && a.MinLeaf != 64) {
			return m, fmt.Errorf("invalid candidate metadata")
		}
		name := fmt.Sprintf("fold%d-depth%d-leaf%d.tree.json", fold, a.Depth, a.MinLeaf)
		if a.File != name || slices.Contains(expected, name) {
			return m, fmt.Errorf("invalid filename")
		}
		expected = append(expected, name)
		data, err := read(filepath.Join(dir, name), 1<<16)
		if err != nil {
			return m, err
		}
		if len(data) != a.Bytes || paireval.Hash(data) != a.SHA256 {
			return m, fmt.Errorf("unreviewed tree")
		}
		if err = verifyShallowHead(data, a); err != nil {
			return m, err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return m, err
	}
	if len(entries) != 31 {
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
			if m.Files[name] != retrievalbench.ShallowClaimPlanSHA256 {
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
