// Package pathinput joins reviewed private source evidence to actual numeric
// costs. Hashes verify the reviewed inputs, not a legal opinion or release grant.
package pathinput

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/pathclaim"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

const (
	SealSchema   = "riido-path-claim-input-seal-v1"
	SourceSchema = "riido-private-source-review-47-v1"
	PatchSHA256  = "6bed55ca053705e59b1362f982b86cbd977fa3d216312a018b1657dfb9fd284b"
	localScope   = "conditional local original numeric heads on generic label-free features; no raw redistribution or public-model approval"
)

type File struct{ Path, SHA256 string }
type Files struct{ Source, Costs, Checkpoint File }
type Caches struct{ Roots, Blobs, Subtrees string }

type Conditions struct {
	Candidate   int `json:"conditional_local_numeric_candidate"`
	Unavailable int `json:"checkpoint_root_unavailable"`
	Restrictive int `json:"pending_restrictive_alpha_beta_eula"`
	FileScoped  int `json:"pending_file_scoped_prefect_license"`
	Missing     int `json:"missing_primary_license_candidate"`
}
type SourceRepository struct {
	Role, Repository string
	Tasks            int
	Counts           Conditions
}
type LicenseReference struct {
	Path, GitSHA, SHA256 string
	Bytes                int
}
type Artifact struct {
	Repository, Path, GitSHA, SHA256, ExampleBaseCommit, PrimarySourceURL string
	Bytes, References                                                     int
}
type Snapshot struct {
	Role, ID, Repository, BaseCommit, RootSHA256, Status, GrantClass string
	CatalogAvailable                                                 bool
	LicenseEvidence                                                  []LicenseReference
}
type Source struct {
	Schema                                                                                          string `json:"schema"`
	Scope                                                                                           string
	KeywordScanIsLicenseDecision, HistoricalIssueAuthorPermissionsProven, SourceEligibilityComplete bool
	InventorySHA256, CollectionResultsSHA256, CollectionFailuresSHA256, MembershipSHA256            string
	ArtifactCount                                                                                   int
	Counts                                                                                          Conditions
	Repositories                                                                                    []SourceRepository
	Artifacts                                                                                       []Artifact
	Snapshots                                                                                       []Snapshot
}
type Checkpoint struct {
	Role, Repository, ID, BaseCommit, RootSHA256, TreeID, CatalogSHA256, Status string
	Match                                                                       filelabels.CatalogMatch
}
type Cost struct {
	Role, Repository, ID, Status, Fallback               string
	Features                                             []float64
	FeaturesAvailable, AuxiliaryRequested, AuxiliaryUsed bool
	Work                                                 fileeval.WorkStats
	Baseline, Helper                                     fileeval.Score
	BaselinePages, HelperPages, Gain                     int
}

// Coverage retains all members, including exclusions. ConditionalCandidate is
// the narrow local numeric review status, not unrestricted source clearance.
type Coverage struct {
	Role, Repository                                                                string
	Tasks, ConditionalCandidates, Eligible, PendingSource, Unavailable, InvalidCost int
}
type Prepared struct {
	Rows                                          []pathclaim.Row
	Costs                                         []Cost // private, aligned to Rows; never part of aggregate reports
	Coverage                                      [2]Coverage
	Repositories                                  []Coverage
	ArtifactsVerified, SnapshotReferencesVerified int
	NumericRowsSHA256                             string
}
type Seal struct {
	Schema, PlanSHA256, MembershipSHA256, QueryProjectionSHA256, PatchProjectionSHA256         string
	SourceEvidenceSHA256, CostExamplesSHA256, AvailabilityCheckpointSHA256, NumericRowsSHA256  string
	RunnerRevision                                                                             string
	TrainingCoverage, ValidationCoverage, ProtectedFinal, TrainingEligible, ValidationEligible int
	ValidationRepositories                                                                     [5]pathclaim.RepositoryCount
	Coverage                                                                                   [2]Coverage
	ArtifactsVerified, SnapshotReferencesVerified                                              int
}

// Check array shape before decoding fixed arrays: encoding/json otherwise
// silently fills short arrays or discards excess repository/coverage entries.
func (s *Seal) UnmarshalJSON(b []byte) error {
	var shape struct{ ValidationRepositories, Coverage []json.RawMessage }
	if e := json.Unmarshal(b, &shape); e != nil || len(shape.ValidationRepositories) != 5 || len(shape.Coverage) != 2 {
		return fmt.Errorf("input seal array shape mismatch")
	}
	type plain Seal
	var next plain
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&next); e != nil {
		return fmt.Errorf("input seal schema mismatch")
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("extra input seal content")
	}
	*s = Seal(next)
	return nil
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func digest(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func read(file File, maxBytes int64, out any) error {
	if !digest(file.SHA256, 64) {
		return fmt.Errorf("input digest syntax")
	}
	f, e := os.Open(file.Path)
	if e != nil {
		return fmt.Errorf("pinned input unavailable")
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if e != nil || int64(len(b)) > maxBytes || hash(b) != file.SHA256 {
		return fmt.Errorf("input size or digest mismatch")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(out) != nil || d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("input schema mismatch")
	}
	return nil
}
func offline(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("reviewed input cache missing; network disabled")
}

// Load returns only fixed development representatives. It does not collect,
// train, score final tasks, approve public weights or interpret license keywords.
func Load(files Files, caches Caches, membershipPath string) (Prepared, error) {
	var empty Prepared
	members, e := trainingdata.ReadDevelopment(membershipPath)
	if e != nil {
		return empty, e
	}
	var source Source
	var costs []Cost
	var checkpoint []Checkpoint
	if e = read(files.Source, 32<<20, &source); e != nil {
		return empty, e
	}
	if e = read(files.Costs, 32<<20, &costs); e != nil {
		return empty, e
	}
	if e = read(files.Checkpoint, 16<<20, &checkpoint); e != nil {
		return empty, e
	}
	if source.Schema != SourceSchema || source.Scope != localScope || source.MembershipSHA256 != trainingdata.MembershipSHA256 || source.KeywordScanIsLicenseDecision || source.HistoricalIssueAuthorPermissionsProven || source.SourceEligibilityComplete {
		return empty, fmt.Errorf("source review scope or provenance mismatch")
	}
	for _, s := range []string{source.InventorySHA256, source.CollectionResultsSHA256, source.CollectionFailuresSHA256} {
		if !digest(s, 64) {
			return empty, fmt.Errorf("source collection digest missing")
		}
	}
	if source.ArtifactCount != len(source.Artifacts) || source.ArtifactCount < 1 || source.ArtifactCount > 4096 {
		return empty, fmt.Errorf("source artifact bound")
	}
	verified, e := verifyArtifacts(source.Artifacts, caches.Blobs)
	if e != nil {
		return empty, e
	}
	prepared, e := join(members, source.Snapshots, costs, checkpoint, caches, verified)
	if e != nil {
		return empty, e
	}
	prepared.ArtifactsVerified = len(verified)
	var got Conditions
	for _, s := range source.Snapshots {
		if e = increment(&got, s.Status); e != nil {
			return empty, e
		}
	}
	if got != source.Counts {
		return empty, fmt.Errorf("source condition totals mismatch")
	}
	if e = verifyRepositoryCounts(source); e != nil {
		return empty, e
	}
	return prepared, nil
}

func verifyRepositoryCounts(source Source) error {
	var got []SourceRepository
	for _, s := range source.Snapshots {
		key := s.Role + "\x00" + s.Repository
		i, found := slices.BinarySearchFunc(got, key, func(a SourceRepository, k string) int { return strings.Compare(a.Role+"\x00"+a.Repository, k) })
		if !found {
			got = slices.Insert(got, i, SourceRepository{Role: s.Role, Repository: s.Repository})
		}
		got[i].Tasks++
		if e := increment(&got[i].Counts, s.Status); e != nil {
			return e
		}
	}
	want := slices.Clone(source.Repositories)
	slices.SortFunc(want, func(a, b SourceRepository) int {
		return strings.Compare(a.Role+"\x00"+a.Repository, b.Role+"\x00"+b.Repository)
	})
	if len(want) != len(got) || !slices.Equal(want, got) {
		return fmt.Errorf("source repository condition counts mismatch")
	}
	return nil
}

func artifactKey(a Artifact) string { return a.Repository + "\x00" + a.Path + "\x00" + a.GitSHA }
func verifyArtifacts(input []Artifact, cache string) ([]Artifact, error) {
	artifacts := slices.Clone(input)
	slices.SortFunc(artifacts, func(a, b Artifact) int { return strings.Compare(artifactKey(a), artifactKey(b)) })
	for i, a := range artifacts {
		if !digest(a.GitSHA, 40) || !digest(a.SHA256, 64) || !digest(a.ExampleBaseCommit, 40) || a.Bytes < 1 || a.Bytes > githubmeta.MaxLicenseBytes || a.References < 1 || a.References > 13021 || !cleanPath(a.Path) || (i > 0 && artifactKey(artifacts[i-1]) == artifactKey(a)) {
			return nil, fmt.Errorf("invalid reviewed artifact identity")
		}
		wantURL := "https://raw.githubusercontent.com/" + a.Repository + "/" + a.ExampleBaseCommit + "/" + a.Path
		if a.PrimarySourceURL != wantURL {
			return nil, fmt.Errorf("artifact source reference mismatch")
		}
		info, e := githubmeta.Blob(context.Background(), a.Repository, a.GitSHA, cache, offline)
		if e != nil || info.SHA256 != a.SHA256 || info.Bytes != a.Bytes {
			return nil, fmt.Errorf("reviewed license blob integrity failure")
		}
	}
	return artifacts, nil
}

func cleanPath(p string) bool {
	return utf8.ValidString(p) && len(p) > 0 && len(p) <= 4096 && p != "." && p != ".." && !strings.ContainsAny(p, "\x00\\") && !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "../") && path.Clean(p) == p
}
func licenseName(p string) bool {
	p = strings.ToUpper(p)
	return sweaudit.LicenseName(p) || p == "LICENCE" || strings.HasPrefix(p, "LICENCE.") || strings.HasPrefix(p, "LICENCE-")
}

func increment(c *Conditions, status string) error {
	switch status {
	case "conditional_local_numeric_candidate":
		c.Candidate++
	case "checkpoint_root_unavailable":
		c.Unavailable++
	case "pending_restrictive_alpha_beta_eula":
		c.Restrictive++
	case "pending_file_scoped_prefect_license":
		c.FileScoped++
	case "missing_primary_license_candidate":
		c.Missing++
	default:
		return fmt.Errorf("unknown source review state")
	}
	return nil
}

func validScore(s fileeval.Score, pages int) bool {
	return s.Targets > 0 && s.Targets <= 100000 && s.Mapped == s.Targets && s.FirstRank >= 1 && s.FirstRank <= 100000 && s.Targets+s.FirstRank-1 <= 100000 && pages == (s.FirstRank+19)/20 && s.Hit1 == (s.FirstRank <= 1) && s.Hit10 == (s.FirstRank <= 10) && s.Hit100 == (s.FirstRank <= 100) && (!s.All10 || s.Hit10 && s.Targets+s.FirstRank-1 <= 10) && (s.Targets != 1 || s.All10 == s.Hit10) && s.ReciprocalRank == 1/float64(s.FirstRank)
}
func validateCost(c Cost) error {
	if len(c.Features) != searchclaim.PathDimension {
		return fmt.Errorf("cost feature shape")
	}
	for _, v := range c.Features {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return fmt.Errorf("cost feature bound")
		}
	}
	switch c.Status {
	case "scored":
		if !c.FeaturesAvailable || !c.AuxiliaryRequested || !c.AuxiliaryUsed || c.Fallback != "" || c.Work != (fileeval.WorkStats{BaselineRankAttempts: 1, AuxiliaryBuildAttempts: 1, AuxiliaryRankAttempts: 1}) || !validScore(c.Baseline, c.BaselinePages) || !validScore(c.Helper, c.HelperPages) || c.Gain != c.BaselinePages-c.HelperPages || c.Baseline.Targets != c.Helper.Targets || c.Baseline.Hit1 != c.Helper.Hit1 {
			return fmt.Errorf("inconsistent paired successful cost")
		}
	case "checkpoint_unavailable", "parse_failure", "no_old_target", "unsupported_target", "unmapped_target", "ranking_error":
		if c.BaselinePages != 0 || c.HelperPages != 0 || c.Gain != 0 || c.Baseline != (fileeval.Score{}) || c.Helper != (fileeval.Score{}) {
			return fmt.Errorf("unscoreable cost has page labels")
		}
	default:
		return fmt.Errorf("unknown cost status")
	}
	return nil
}

func join(members []sweaudit.TaskRole, source []Snapshot, costs []Cost, points []Checkpoint, caches Caches, artifacts []Artifact) (Prepared, error) {
	var out Prepared
	if len(members) == 0 || len(members) > pathclaim.MaxRows || len(source) != len(members) || len(costs) != len(members) || len(points) != len(members) {
		return out, fmt.Errorf("source/cost coverage mismatch")
	}
	ms := slices.Clone(members)
	ss := slices.Clone(source)
	cs := slices.Clone(costs)
	ps := slices.Clone(points)
	slices.SortFunc(ms, func(a, b sweaudit.TaskRole) int { return strings.Compare(a.Task.ID, b.Task.ID) })
	slices.SortFunc(ss, func(a, b Snapshot) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(cs, func(a, b Cost) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(ps, func(a, b Checkpoint) int { return strings.Compare(a.ID, b.ID) })
	components := make([]string, len(ms))
	for i, m := range ms {
		components[i] = m.Task.ComponentSHA256
	}
	slices.Sort(components)
	if len(slices.Compact(slices.Clone(components))) != len(ms) {
		return out, fmt.Errorf("duplicate development component")
	}
	out.Coverage[0].Role = "train"
	out.Coverage[1].Role = "validation"
	refCounts := make([]int, len(artifacts))
	exampleBound := make([]bool, len(artifacts))
	for i, m := range ms {
		s, c, p := ss[i], cs[i], ps[i]
		t := m.Task
		if t.Source != "train" || !digest(t.ComponentSHA256, 64) || (m.Role != "train" && m.Role != "validation") || s.ID != t.ID || c.ID != t.ID || p.ID != t.ID || s.Role != m.Role || c.Role != m.Role || p.Role != m.Role || s.Repository != t.Repository || c.Repository != t.Repository || p.Repository != t.Repository || s.BaseCommit != t.BaseCommit || p.BaseCommit != t.BaseCommit || (i > 0 && ms[i-1].Task.ID == t.ID) {
			return Prepared{}, fmt.Errorf("source/cost development identity mismatch")
		}
		var state Conditions
		if e := increment(&state, s.Status); e != nil {
			return Prepared{}, e
		}
		if e := validateCost(c); e != nil {
			return Prepared{}, e
		}
		if s.RootSHA256 != p.RootSHA256 || s.CatalogAvailable != (p.CatalogSHA256 != "") {
			return Prepared{}, fmt.Errorf("source/checkpoint availability mismatch")
		}
		if p.RootSHA256 == "" {
			if p.TreeID != "" || p.CatalogSHA256 != "" || p.Status != "root_unavailable" || s.Status != "checkpoint_root_unavailable" || len(s.LicenseEvidence) != 0 || c.Status != "checkpoint_unavailable" {
				return Prepared{}, fmt.Errorf("unexpected unavailable source evidence")
			}
		} else {
			if !digest(p.RootSHA256, 64) || !digest(p.TreeID, 40) || s.Status == "checkpoint_root_unavailable" {
				return Prepared{}, fmt.Errorf("invalid available root evidence")
			}
			b, e := githubmeta.Root(context.Background(), t.Repository, t.BaseCommit, caches.Roots, offline)
			if e != nil || hash(b) != p.RootSHA256 {
				return Prepared{}, fmt.Errorf("source root integrity failure")
			}
			tree, entries, e := sweaudit.TreeEntries(t.Repository, b)
			if e != nil || tree != p.TreeID {
				return Prepared{}, fmt.Errorf("source tree identity mismatch")
			}
			if p.CatalogSHA256 == "" {
				if p.Status != "catalog_unavailable" || c.Status != "checkpoint_unavailable" {
					return Prepared{}, fmt.Errorf("root-only availability mismatch")
				}
			} else if !digest(p.CatalogSHA256, 64) || (p.Status != "matched" && p.Status != "unusable_target") || c.Status == "checkpoint_unavailable" {
				return Prepared{}, fmt.Errorf("catalog evidence mismatch")
			}
			if e := verifySnapshotReferences(t.Repository, entries, s.LicenseEvidence, caches.Subtrees); e != nil {
				return Prepared{}, e
			}
			if c.Status == "scored" && (p.Status != "matched" || p.Match.Targets != c.Baseline.Targets || p.Match.RegularFiles != p.Match.Targets || p.Match.Symlinks != 0 || p.Match.Submodules != 0 || p.Match.Directories != 0 || p.Match.Missing != 0 || len(p.Match.MissingPaths) != 0) {
				return Prepared{}, fmt.Errorf("scored targets disagree with checkpoint join")
			}
		}
		refKeys := make([]string, len(s.LicenseEvidence))
		for k, r := range s.LicenseEvidence {
			key := t.Repository + "\x00" + r.Path + "\x00" + r.GitSHA
			j, found := slices.BinarySearchFunc(artifacts, key, func(a Artifact, k string) int { return strings.Compare(artifactKey(a), k) })
			if !found || artifacts[j].SHA256 != r.SHA256 || artifacts[j].Bytes != r.Bytes {
				return Prepared{}, fmt.Errorf("snapshot license reference mismatch")
			}
			out.SnapshotReferencesVerified++
			refCounts[j]++
			if artifacts[j].ExampleBaseCommit == t.BaseCommit {
				exampleBound[j] = true
			}
			refKeys[k] = key
		}
		slices.Sort(refKeys)
		if len(slices.Compact(refKeys)) != len(s.LicenseEvidence) {
			return Prepared{}, fmt.Errorf("duplicate snapshot document reference")
		}
		if s.Status == "conditional_local_numeric_candidate" && (s.GrantClass == "" || len(s.LicenseEvidence) == 0) {
			return Prepared{}, fmt.Errorf("candidate lacks reviewed document evidence")
		}
		role := 0
		if m.Role == "validation" {
			role = 1
		}
		status := 0 // unavailable, pending source, invalid cost, eligible
		if s.Status != "checkpoint_root_unavailable" {
			status = 1
			if s.Status == "conditional_local_numeric_candidate" {
				status = 2
				if s.CatalogAvailable && c.Status == "scored" {
					status = 3
				}
			}
		}
		update := func(n *Coverage) {
			n.Tasks++
			if s.Status == "conditional_local_numeric_candidate" {
				n.ConditionalCandidates++
			}
			switch status {
			case 0:
				n.Unavailable++
			case 1:
				n.PendingSource++
			case 2:
				n.InvalidCost++
			case 3:
				n.Eligible++
			}
		}
		update(&out.Coverage[role])
		j, found := slices.BinarySearchFunc(out.Repositories, m.Role+"\x00"+t.Repository, func(a Coverage, k string) int { return strings.Compare(a.Role+"\x00"+a.Repository, k) })
		if !found {
			out.Repositories = slices.Insert(out.Repositories, j, Coverage{Role: m.Role, Repository: t.Repository})
		}
		update(&out.Repositories[j])
		if status == 3 {
			group, _ := slices.BinarySearch(components, t.ComponentSHA256)
			row := pathclaim.Row{Role: m.Role, Repository: t.Repository, Group: group, BaselinePages: c.BaselinePages, HelperPages: c.HelperPages}
			copy(row.Features[:], c.Features)
			out.Rows = append(out.Rows, row)
			c.Features = slices.Clone(c.Features)
			out.Costs = append(out.Costs, c)
		}
	}
	for i, a := range artifacts {
		if refCounts[i] != a.References || !exampleBound[i] {
			return Prepared{}, fmt.Errorf("artifact reference count mismatch")
		}
	}
	if len(out.Rows) > 0 {
		h, e := pathclaim.RowsSHA256(out.Rows)
		if e != nil {
			return Prepared{}, e
		}
		out.NumericRowsSHA256 = h
	}
	return out, nil
}

func verifySnapshotReferences(repo string, entries []sweaudit.TreeEntry, refs []LicenseReference, subtrees string) error {
	var candidates []sweaudit.TreeEntry
	for _, entry := range entries {
		if entry.Type == "blob" && licenseName(entry.Path) {
			candidates = append(candidates, entry)
		}
		name := strings.ToUpper(entry.Path)
		if entry.Type == "tree" && (name == "LICENSE" || name == "LICENSES" || name == "LICENCE" || name == "LICENCES") {
			children, e := githubmeta.Tree(context.Background(), repo, entry.SHA, subtrees, offline)
			if e != nil {
				return fmt.Errorf("required license subtree unavailable")
			}
			for _, child := range children {
				if child.Type == "tree" {
					return fmt.Errorf("reviewed license subtree is incomplete")
				}
				if child.Type == "blob" {
					child.Path = entry.Path + "/" + child.Path
					candidates = append(candidates, child)
				}
			}
		}
	}
	slices.SortFunc(candidates, func(a, b sweaudit.TreeEntry) int { return strings.Compare(a.Path, b.Path) })
	for _, r := range refs {
		i, found := slices.BinarySearchFunc(candidates, r.Path, func(a sweaudit.TreeEntry, p string) int { return strings.Compare(a.Path, p) })
		if !found || candidates[i].SHA != r.GitSHA || candidates[i].Mode == "120000" {
			return fmt.Errorf("document is not bound to source snapshot")
		}
	}
	return nil
}

func MakeSeal(p Prepared, files Files, runner string) (Seal, error) {
	s := Seal{Schema: SealSchema, PlanSHA256: pathclaim.PlanSHA256, MembershipSHA256: trainingdata.MembershipSHA256, QueryProjectionSHA256: trainingdata.QueryProjectionSHA256, PatchProjectionSHA256: PatchSHA256, SourceEvidenceSHA256: files.Source.SHA256, CostExamplesSHA256: files.Costs.SHA256, AvailabilityCheckpointSHA256: files.Checkpoint.SHA256, NumericRowsSHA256: p.NumericRowsSHA256, RunnerRevision: runner, TrainingCoverage: p.Coverage[0].Tasks, ValidationCoverage: p.Coverage[1].Tasks, ProtectedFinal: 2402, TrainingEligible: p.Coverage[0].Eligible, ValidationEligible: p.Coverage[1].Eligible, Coverage: p.Coverage, ArtifactsVerified: p.ArtifactsVerified, SnapshotReferencesVerified: p.SnapshotReferencesVerified}
	for i, name := range pathclaim.ValidationRepositories() {
		s.ValidationRepositories[i].Repository = name
		for _, r := range p.Repositories {
			if r.Role == "validation" && r.Repository == name {
				s.ValidationRepositories[i].Eligible = r.Eligible
			}
		}
	}
	if s.TrainingCoverage != 7335 || s.ValidationCoverage != 5686 || !digest(runner, 40) || !digest(s.NumericRowsSHA256, 64) {
		return Seal{}, fmt.Errorf("input seal coverage or revision mismatch")
	}
	return s, nil
}
func (s Seal) Readiness(sealSHA256 string) (pathclaim.Readiness, error) {
	if s.Schema != SealSchema || s.PlanSHA256 != pathclaim.PlanSHA256 || s.MembershipSHA256 != trainingdata.MembershipSHA256 || s.QueryProjectionSHA256 != trainingdata.QueryProjectionSHA256 || s.PatchProjectionSHA256 != PatchSHA256 || !digest(s.CostExamplesSHA256, 64) {
		return pathclaim.Readiness{}, fmt.Errorf("seal fixed scope mismatch")
	}
	r := pathclaim.Readiness{Schema: pathclaim.ReadinessSchema, PlanSHA256: s.PlanSHA256, MembershipSHA256: s.MembershipSHA256, NumericRowsSHA256: s.NumericRowsSHA256, AvailabilityCheckpointSHA256: s.AvailabilityCheckpointSHA256, SourceEvidenceSHA256: s.SourceEvidenceSHA256, InputSealSHA256: sealSHA256, RunnerRevision: s.RunnerRevision, TrainingCoverage: s.TrainingCoverage, ValidationCoverage: s.ValidationCoverage, ProtectedFinal: s.ProtectedFinal, TrainingEligible: s.TrainingEligible, ValidationEligible: s.ValidationEligible, ValidationRepositories: s.ValidationRepositories}
	return r, r.Validate()
}

// ReadSeal bounds and pins the complete file. The caller must additionally prove
// that its exact bytes are committed before fitting, and verify runner identity.
func ReadSeal(file File) (Seal, error) { var s Seal; e := read(file, 64<<10, &s); return s, e }
