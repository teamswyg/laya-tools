package pathinput

import (
	"context"
	"crypto/sha1" // Reproduce Git's object identity for owned fixture text.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/pathclaim"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

// These are original compact bookkeeping fixtures, not a licensed real corpus
// or evidence that the full 2400+2400 readiness/fitting gates have succeeded.
type inputFixture struct {
	members   []sweaudit.TaskRole
	source    []Snapshot
	costs     []Cost
	points    []Checkpoint
	artifacts []Artifact
	caches    Caches
}

func gitBlobSHA(b []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d%c", len(b), 0)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func treeCachePath(cache, repo, revision string) string {
	return filepath.Join(cache, hash([]byte(repo+"\x00"+revision))+".json")
}

func writeFixtureTree(t *testing.T, cache, repo, revision, treeID string, entries []sweaudit.TreeEntry) []byte {
	t.Helper()
	raw, err := json.Marshal(githubmeta.Catalog{SHA: treeID, Tree: entries})
	if err != nil {
		t.Fatal("owned tree serialization failed")
	}
	if err = os.WriteFile(treeCachePath(cache, repo, revision), raw, 0600); err != nil {
		t.Fatal("owned tree cache write failed")
	}
	return raw
}

func scoreAt(rank int) fileeval.Score {
	return fileeval.Score{Targets: 1, Mapped: 1, FirstRank: rank, Hit1: rank == 1, Hit10: rank <= 10, Hit100: rank <= 100, All10: rank <= 10, ReciprocalRank: 1 / float64(rank)}
}

func newInputFixture(t *testing.T) inputFixture {
	t.Helper()
	f := inputFixture{caches: Caches{Roots: t.TempDir(), Blobs: t.TempDir(), Subtrees: t.TempDir()}}
	text := []byte("Original permission fixture for integrity tests only.\n")
	blob := gitBlobSHA(text)
	if err := os.WriteFile(filepath.Join(f.caches.Blobs, blob+".txt"), text, 0600); err != nil {
		t.Fatal("owned blob cache write failed")
	}
	for i, role := range []string{"train", "validation"} {
		repo := "pypa/pip"
		if role == "validation" {
			repo = "numpy/numpy"
		}
		id, commit, tree := fmt.Sprintf("owned-%d", i), fmt.Sprintf("%040x", i+1), fmt.Sprintf("%040x", i+100)
		raw := writeFixtureTree(t, f.caches.Roots, repo, commit, tree, []sweaudit.TreeEntry{{Path: "LICENSE", Mode: "100644", Type: "blob", SHA: blob}, {Path: "file.go", Mode: "100644", Type: "blob", SHA: strings.Repeat("a", 40)}})
		f.members = append(f.members, sweaudit.TaskRole{Role: role, Task: sweaudit.Selection{Source: "train", ID: id, Repository: repo, BaseCommit: commit, ComponentSHA256: fmt.Sprintf("%064x", i+1)}})
		f.source = append(f.source, Snapshot{Role: role, ID: id, Repository: repo, BaseCommit: commit, RootSHA256: hash(raw), Status: "conditional_local_numeric_candidate", GrantClass: "owned-original-fixture", CatalogAvailable: true, LicenseEvidence: []LicenseReference{{Path: "LICENSE", GitSHA: blob, SHA256: hash(text), Bytes: len(text)}}})
		f.points = append(f.points, Checkpoint{Role: role, Repository: repo, ID: id, BaseCommit: commit, RootSHA256: hash(raw), TreeID: tree, CatalogSHA256: hash(raw), Status: "matched", Match: filelabels.CatalogMatch{Targets: 1, RegularFiles: 1}})
		features := make([]float64, searchclaim.PathDimension)
		features[0], features[15] = .25, .75
		f.costs = append(f.costs, Cost{Role: role, Repository: repo, ID: id, Status: "scored", Features: features, FeaturesAvailable: true, AuxiliaryRequested: true, AuxiliaryUsed: true, Work: fileeval.WorkStats{BaselineRankAttempts: 1, AuxiliaryBuildAttempts: 1, AuxiliaryRankAttempts: 1}, Baseline: scoreAt(21), Helper: scoreAt(2), BaselinePages: 2, HelperPages: 1, Gain: 1})
		f.artifacts = append(f.artifacts, Artifact{Repository: repo, Path: "LICENSE", GitSHA: blob, SHA256: hash(text), Bytes: len(text), References: 1, ExampleBaseCommit: commit, PrimarySourceURL: "https://raw.githubusercontent.com/" + repo + "/" + commit + "/LICENSE"})
	}
	slices.SortFunc(f.artifacts, func(a, b Artifact) int { return strings.Compare(artifactKey(a), artifactKey(b)) })
	return f
}

func (f inputFixture) joined() (Prepared, error) {
	return join(f.members, f.source, f.costs, f.points, f.caches, f.artifacts)
}

func (f inputFixture) replaceRoot(t *testing.T, i int, entries []sweaudit.TreeEntry) {
	t.Helper()
	m := f.members[i].Task
	raw := writeFixtureTree(t, f.caches.Roots, m.Repository, m.BaseCommit, f.points[i].TreeID, entries)
	f.source[i].RootSHA256, f.points[i].RootSHA256 = hash(raw), hash(raw)
}

func (f inputFixture) removeArtifactFor(repo string) []Artifact {
	return slices.DeleteFunc(slices.Clone(f.artifacts), func(a Artifact) bool { return a.Repository == repo })
}

func unscored(c Cost, status string) Cost {
	return Cost{Role: c.Role, Repository: c.Repository, ID: c.ID, Status: status, Features: make([]float64, searchclaim.PathDimension)}
}

func TestOwnedJoinActualCacheRoundTripAndReorderedInput(t *testing.T) {
	f := newInputFixture(t)
	for _, m := range f.members {
		if _, err := githubmeta.Root(context.Background(), m.Task.Repository, m.Task.BaseCommit, f.caches.Roots, offline); err != nil {
			t.Fatal("fixture did not match actual root cache contract")
		}
	}
	input := slices.Clone(f.artifacts)
	slices.Reverse(input)
	before := slices.Clone(input)
	verified, err := verifyArtifacts(input, f.caches.Blobs)
	if err != nil || !reflect.DeepEqual(input, before) || !reflect.DeepEqual(verified, f.artifacts) {
		t.Fatal("artifact verification changed input or failed owned blob roundtrip")
	}
	got, err := f.joined()
	if err != nil || len(got.Rows) != 2 || len(got.Costs) != 2 || got.SnapshotReferencesVerified != 2 || got.Coverage[0].Tasks != 1 || got.Coverage[1].Tasks != 1 || got.Coverage[0].Eligible != 1 || got.Coverage[1].Eligible != 1 || !digest(got.NumericRowsSHA256, 64) {
		t.Fatal("owned join did not preserve full compact denominator")
	}
	if got.Rows[0].Group == got.Rows[1].Group || got.Rows[0].Features[15] != .75 || got.Rows[0].BaselinePages != 2 || got.Rows[0].HelperPages != 1 {
		t.Fatal("global membership or numeric projection changed")
	}
	slices.Reverse(f.members)
	slices.Reverse(f.source)
	slices.Reverse(f.costs)
	slices.Reverse(f.points)
	replay, err := f.joined()
	if err != nil || !reflect.DeepEqual(got, replay) {
		t.Fatal("input ordering altered canonical numeric join")
	}
	if _, err := MakeSeal(got, Files{}, strings.Repeat("a", 40)); err == nil {
		t.Fatal("compact fixture was promoted to actual cohort readiness")
	}
}

func TestPreparedRowsAndAlignedCostsOwnFeatureArrays(t *testing.T) {
	f := newInputFixture(t)
	first, err := f.joined()
	if err != nil {
		t.Fatal("owned numeric join failed")
	}
	second, err := f.joined()
	if err != nil {
		t.Fatal("owned numeric replay failed")
	}
	f.costs[0].Features[0] = .9
	f.costs[1].Features[15] = .1
	if first.Rows[0].Features[0] != .25 || first.Costs[0].Features[0] != .25 || first.Rows[1].Features[15] != .75 || first.Costs[1].Features[15] != .75 || second.Costs[0].Features[0] != .25 {
		t.Fatal("returned numeric rows or private aligned costs alias raw input features")
	}
	first.Costs[0].Features[0] = .5
	if first.Rows[0].Features[0] != .25 || second.Rows[0].Features[0] != .25 || second.Costs[0].Features[0] != .25 {
		t.Fatal("private cost mutation changed numeric rows or another preparation")
	}
}

func TestJoinRejectsDevelopmentIdentityAndProvenanceMisbindings(t *testing.T) {
	for _, mode := range []string{"final", "source_role", "cost_role", "checkpoint_role", "source_id", "cost_id", "checkpoint_id", "source_repo", "cost_repo", "checkpoint_repo", "source_commit", "checkpoint_commit", "member_source", "component_syntax", "component_duplicate", "member_duplicate", "source_missing", "cost_missing", "checkpoint_extra", "source_root", "root_digest", "tree", "catalog_available", "catalog_digest", "catalog_status", "checkpoint_target_count", "checkpoint_symlink"} {
		t.Run(mode, func(t *testing.T) {
			f := newInputFixture(t)
			switch mode {
			case "final":
				f.members[0].Role, f.source[0].Role, f.costs[0].Role, f.points[0].Role = "final", "final", "final", "final"
			case "source_role":
				f.source[0].Role = "validation"
			case "cost_role":
				f.costs[0].Role = "validation"
			case "checkpoint_role":
				f.points[0].Role = "validation"
			case "source_id":
				f.source[0].ID = "owned-other"
			case "cost_id":
				f.costs[0].ID = "owned-other"
			case "checkpoint_id":
				f.points[0].ID = "owned-other"
			case "source_repo":
				f.source[0].Repository = "DataDog/integrations-core"
			case "cost_repo":
				f.costs[0].Repository = "DataDog/integrations-core"
			case "checkpoint_repo":
				f.points[0].Repository = "DataDog/integrations-core"
			case "source_commit":
				f.source[0].BaseCommit = strings.Repeat("b", 40)
			case "checkpoint_commit":
				f.points[0].BaseCommit = strings.Repeat("b", 40)
			case "member_source":
				f.members[0].Task.Source = "final"
			case "component_syntax":
				f.members[0].Task.ComponentSHA256 = "not-a-component"
			case "component_duplicate":
				f.members[1].Task.ComponentSHA256 = f.members[0].Task.ComponentSHA256
			case "member_duplicate":
				f.members[1].Task.ID = f.members[0].Task.ID
			case "source_missing":
				f.source = f.source[1:]
			case "cost_missing":
				f.costs = f.costs[1:]
			case "checkpoint_extra":
				f.points = append(f.points, f.points[0])
			case "source_root":
				f.source[0].RootSHA256 = strings.Repeat("b", 64)
			case "root_digest":
				f.source[0].RootSHA256, f.points[0].RootSHA256 = strings.Repeat("b", 64), strings.Repeat("b", 64)
			case "tree":
				f.points[0].TreeID = strings.Repeat("b", 40)
			case "catalog_available":
				f.source[0].CatalogAvailable = false
			case "catalog_digest":
				f.points[0].CatalogSHA256 = strings.Repeat("x", 64)
			case "catalog_status":
				f.points[0].Status = "unusable_target"
			case "checkpoint_target_count":
				f.points[0].Match.Targets = 2
			case "checkpoint_symlink":
				f.points[0].Match.RegularFiles, f.points[0].Match.Symlinks = 0, 1
			}
			if _, err := f.joined(); err == nil {
				t.Fatal("invalid development/provenance binding accepted")
			}
		})
	}
}

func TestPairedCostConsistencyCannotManufactureTrainingLabels(t *testing.T) {
	for _, mode := range []string{"nan", "inf", "negative_feature", "over_feature", "missing_feature", "extra_feature", "unavailable_feature", "no_aux_request", "no_aux_use", "fallback", "work", "gain", "page", "rank", "mapped", "reciprocal", "hit10", "all10", "pair_targets", "pair_hit1", "too_many_all10", "impossible_all10_capacity", "impossible_target_capacity", "unknown_status", "unscored_pages", "unscored_score"} {
		t.Run(mode, func(t *testing.T) {
			f := newInputFixture(t)
			c := &f.costs[0]
			switch mode {
			case "nan":
				c.Features[0] = math.NaN()
			case "inf":
				c.Features[0] = math.Inf(1)
			case "negative_feature":
				c.Features[0] = -.1
			case "over_feature":
				c.Features[0] = 1.1
			case "missing_feature":
				c.Features = c.Features[:15]
			case "extra_feature":
				c.Features = append(c.Features, 0)
			case "unavailable_feature":
				c.FeaturesAvailable = false
			case "no_aux_request":
				c.AuxiliaryRequested = false
			case "no_aux_use":
				c.AuxiliaryUsed = false
			case "fallback":
				c.Fallback = "auxiliary_error"
			case "work":
				c.Work.AuxiliaryRankAttempts = 0
			case "gain":
				c.Gain = 0
			case "page":
				c.BaselinePages = 3
			case "rank":
				c.Helper.FirstRank = 0
			case "mapped":
				c.Helper.Mapped = 0
			case "reciprocal":
				c.Helper.ReciprocalRank = .75
			case "hit10":
				c.Helper.Hit10 = false
			case "all10":
				c.Helper.All10 = false
			case "pair_targets":
				c.Helper.Targets, c.Helper.Mapped = 2, 2
			case "pair_hit1":
				c.Helper = scoreAt(1)
			case "too_many_all10":
				c.Baseline.Targets, c.Baseline.Mapped, c.Helper.Targets, c.Helper.Mapped = 11, 11, 11, 11
				f.points[0].Match.Targets, f.points[0].Match.RegularFiles = 11, 11
			case "impossible_all10_capacity":
				c.Helper = scoreAt(10)
				c.Baseline.Targets, c.Baseline.Mapped, c.Helper.Targets, c.Helper.Mapped = 2, 2, 2, 2
				f.points[0].Match.Targets, f.points[0].Match.RegularFiles = 2, 2
			case "impossible_target_capacity":
				c.Baseline = scoreAt(100000)
				c.Baseline.Targets, c.Baseline.Mapped, c.Helper.Targets, c.Helper.Mapped = 2, 2, 2, 2
				c.BaselinePages, c.Gain = 5000, 4999
				f.points[0].Match.Targets, f.points[0].Match.RegularFiles = 2, 2
			case "unknown_status":
				c.Status = "no_old_targets"
			case "unscored_pages":
				c.Status = "parse_failure"
			case "unscored_score":
				*c = unscored(*c, "parse_failure")
				c.Helper = scoreAt(1)
			}
			if _, err := f.joined(); err == nil {
				t.Fatal("inconsistent numeric cost accepted")
			}
		})
	}
}

func TestExclusionsStayInCoverageWithoutNegativeReplacement(t *testing.T) {
	for _, status := range []string{"parse_failure", "no_old_target", "unsupported_target", "unmapped_target", "ranking_error"} {
		t.Run(status, func(t *testing.T) {
			f := newInputFixture(t)
			f.costs[0] = unscored(f.costs[0], status)
			got, err := f.joined()
			if err != nil || len(got.Rows) != 1 || len(got.Costs) != 1 || got.Coverage[0].Tasks != 1 || got.Coverage[0].InvalidCost != 1 || got.Coverage[0].Eligible != 0 || got.Coverage[1].Eligible != 1 {
				t.Fatal("invalid target disappeared or became a training negative")
			}
		})
	}
	f := newInputFixture(t)
	f.source[0].Status = "pending_restrictive_alpha_beta_eula"
	got, err := f.joined()
	if err != nil || len(got.Rows) != 1 || got.Coverage[0].Tasks != 1 || got.Coverage[0].PendingSource != 1 || got.Coverage[0].Eligible != 0 {
		t.Fatal("pending source disappeared or became a training negative")
	}
	f = newInputFixture(t)
	f.source[0].CatalogAvailable = false
	f.points[0].CatalogSHA256, f.points[0].Status, f.points[0].Match = "", "catalog_unavailable", filelabels.CatalogMatch{}
	f.costs[0] = unscored(f.costs[0], "checkpoint_unavailable")
	got, err = f.joined()
	if err != nil || len(got.Rows) != 1 || got.Coverage[0].Tasks != 1 || got.Coverage[0].InvalidCost != 1 {
		t.Fatal("root-only source was substituted with a negative")
	}
	if err := os.Remove(treeCachePath(f.caches.Roots, f.members[0].Task.Repository, f.members[0].Task.BaseCommit)); err != nil {
		t.Fatal("owned required-root removal failed")
	}
	if _, err := f.joined(); err == nil {
		t.Fatal("root-only checkpoint skipped required root integrity")
	}
	f = newInputFixture(t)
	f.source[0].Status, f.source[0].GrantClass, f.source[0].RootSHA256, f.source[0].CatalogAvailable, f.source[0].LicenseEvidence = "checkpoint_root_unavailable", "", "", false, nil
	f.points[0].RootSHA256, f.points[0].TreeID, f.points[0].CatalogSHA256, f.points[0].Status, f.points[0].Match = "", "", "", "root_unavailable", filelabels.CatalogMatch{}
	f.costs[0] = unscored(f.costs[0], "checkpoint_unavailable")
	f.artifacts = f.removeArtifactFor(f.members[0].Task.Repository)
	got, err = f.joined()
	if err != nil || len(got.Rows) != 1 || got.Coverage[0].Unavailable != 1 || got.Coverage[0].Tasks != 1 || got.Coverage[0].Eligible != 0 {
		t.Fatal("later cached root changed frozen unavailable membership")
	}
}

func TestRequiredRootAndLicenseArtifactCorruptionAreFatal(t *testing.T) {
	for _, mode := range []string{"root_missing", "root_corrupt", "blob_missing", "blob_changed", "artifact_digest", "artifact_size", "artifact_url", "artifact_unreferenced_commit", "artifact_duplicate", "artifact_reference_count", "snapshot_reference_duplicate", "snapshot_reference_digest", "candidate_document_missing"} {
		t.Run(mode, func(t *testing.T) {
			f := newInputFixture(t)
			switch mode {
			case "root_missing":
				if err := os.Remove(treeCachePath(f.caches.Roots, f.members[0].Task.Repository, f.members[0].Task.BaseCommit)); err != nil {
					t.Fatal("owned cache removal failed")
				}
			case "root_corrupt":
				if err := os.WriteFile(treeCachePath(f.caches.Roots, f.members[0].Task.Repository, f.members[0].Task.BaseCommit), []byte("corrupt"), 0600); err != nil {
					t.Fatal("owned cache mutation failed")
				}
			case "blob_missing":
				if err := os.Remove(filepath.Join(f.caches.Blobs, f.artifacts[0].GitSHA+".txt")); err != nil {
					t.Fatal("owned cache removal failed")
				}
			case "blob_changed":
				if err := os.WriteFile(filepath.Join(f.caches.Blobs, f.artifacts[0].GitSHA+".txt"), []byte("Changed original fixture\n"), 0600); err != nil {
					t.Fatal("owned cache mutation failed")
				}
			case "artifact_digest":
				f.artifacts[0].SHA256 = strings.Repeat("b", 64)
			case "artifact_size":
				f.artifacts[0].Bytes++
			case "artifact_url":
				f.artifacts[0].PrimarySourceURL = "https://example.invalid/fixture"
			case "artifact_unreferenced_commit":
				f.artifacts[0].ExampleBaseCommit = strings.Repeat("d", 40)
				f.artifacts[0].PrimarySourceURL = "https://raw.githubusercontent.com/" + f.artifacts[0].Repository + "/" + f.artifacts[0].ExampleBaseCommit + "/" + f.artifacts[0].Path
			case "artifact_duplicate":
				f.artifacts = append(f.artifacts, f.artifacts[0])
			case "artifact_reference_count":
				f.artifacts[0].References++
			case "snapshot_reference_duplicate":
				f.source[0].LicenseEvidence = append(f.source[0].LicenseEvidence, f.source[0].LicenseEvidence[0])
			case "snapshot_reference_digest":
				f.source[0].LicenseEvidence[0].SHA256 = strings.Repeat("b", 64)
			case "candidate_document_missing":
				f.source[0].LicenseEvidence = nil
			}
			verified, err := verifyArtifacts(f.artifacts, f.caches.Blobs)
			if err == nil {
				_, err = join(f.members, f.source, f.costs, f.points, f.caches, verified)
			}
			if err == nil {
				t.Fatal("required reviewed artifact/root integrity failure accepted")
			}
		})
	}
}

func TestLicenseBlobActual256KiBBound(t *testing.T) {
	for _, size := range []int{githubmeta.MaxLicenseBytes, githubmeta.MaxLicenseBytes + 1} {
		t.Run(fmt.Sprintf("bytes_%d", size), func(t *testing.T) {
			text := []byte(strings.Repeat("x", size))
			blob := gitBlobSHA(text)
			cache := t.TempDir()
			if err := os.WriteFile(filepath.Join(cache, blob+".txt"), text, 0600); err != nil {
				t.Fatal("owned bounded blob write failed")
			}
			commit := strings.Repeat("a", 40)
			a := Artifact{Repository: "pypa/pip", Path: "LICENSE", GitSHA: blob, SHA256: hash(text), Bytes: size, References: 1, ExampleBaseCommit: commit, PrimarySourceURL: "https://raw.githubusercontent.com/pypa/pip/" + commit + "/LICENSE"}
			_, err := verifyArtifacts([]Artifact{a}, cache)
			if (size == githubmeta.MaxLicenseBytes) != (err == nil) {
				t.Fatal("actual license byte bound incorrect")
			}
		})
	}
}

func TestLicenseDocumentMustBindToSnapshotAndDirectSubtree(t *testing.T) {
	for _, mode := range []string{"different_blob", "not_present", "symlink"} {
		t.Run(mode, func(t *testing.T) {
			f := newInputFixture(t)
			entry := sweaudit.TreeEntry{Path: "LICENSE", Mode: "100644", Type: "blob", SHA: f.source[0].LicenseEvidence[0].GitSHA}
			switch mode {
			case "different_blob":
				entry.SHA = strings.Repeat("b", 40)
			case "not_present":
				entry.Path = "README.md"
			case "symlink":
				entry.Mode = "120000"
			}
			f.replaceRoot(t, 0, []sweaudit.TreeEntry{entry})
			if _, err := f.joined(); err == nil {
				t.Fatal("unbound snapshot document accepted")
			}
		})
	}
	for _, mode := range []string{"valid", "missing", "corrupt", "wrong_blob", "nested"} {
		t.Run("subtree_"+mode, func(t *testing.T) {
			f := newInputFixture(t)
			m := f.members[0].Task
			subtree := strings.Repeat("d", 40)
			f.replaceRoot(t, 0, []sweaudit.TreeEntry{{Path: "LICENSES", Mode: "040000", Type: "tree", SHA: subtree}})
			ref := &f.source[0].LicenseEvidence[0]
			ref.Path = "LICENSES/fixture.txt"
			for i := range f.artifacts {
				if f.artifacts[i].Repository == m.Repository {
					f.artifacts[i].Path = ref.Path
					f.artifacts[i].PrimarySourceURL = "https://raw.githubusercontent.com/" + m.Repository + "/" + m.BaseCommit + "/" + ref.Path
				}
			}
			slices.SortFunc(f.artifacts, func(a, b Artifact) int { return strings.Compare(artifactKey(a), artifactKey(b)) })
			entry := sweaudit.TreeEntry{Path: "fixture.txt", Mode: "100644", Type: "blob", SHA: ref.GitSHA}
			if mode == "wrong_blob" {
				entry.SHA = strings.Repeat("e", 40)
			} else if mode == "nested" {
				entry.Path, entry.Mode, entry.Type = "nested", "040000", "tree"
			}
			if mode != "missing" {
				writeFixtureTree(t, f.caches.Subtrees, m.Repository, subtree, subtree, []sweaudit.TreeEntry{entry})
			}
			if mode == "corrupt" {
				if err := os.WriteFile(treeCachePath(f.caches.Subtrees, m.Repository, subtree), []byte("corrupt"), 0600); err != nil {
					t.Fatal("owned subtree mutation failed")
				}
			}
			_, err := f.joined()
			if (mode == "valid") != (err == nil) {
				t.Fatal("direct subtree document verification result incorrect")
			}
		})
	}
}

func TestPinnedReadBoundsStrictJSONAndSealReadinessStayClosed(t *testing.T) {
	for _, mode := range []string{"valid", "hash", "digest_syntax", "unknown", "trailing", "bound", "missing"} {
		t.Run(mode, func(t *testing.T) {
			raw := []byte(`{"Path":"owned","SHA256":"fixture"}`)
			if mode == "unknown" {
				raw = []byte(`{"Path":"owned","SHA256":"fixture","Extra":1}`)
			} else if mode == "trailing" {
				raw = append(raw, []byte(" {}")...)
			}
			name := filepath.Join(t.TempDir(), "owned.json")
			if mode != "missing" {
				if err := os.WriteFile(name, raw, 0600); err != nil {
					t.Fatal("owned input write failed")
				}
			}
			file, limit := File{Path: name, SHA256: hash(raw)}, int64(len(raw))
			if mode == "hash" {
				file.SHA256 = strings.Repeat("b", 64)
			} else if mode == "digest_syntax" {
				file.SHA256 = strings.Repeat("B", 64)
			} else if mode == "bound" {
				limit--
			}
			var got File
			if err := read(file, limit, &got); (mode == "valid") != (err == nil) {
				t.Fatal("pinned strict input result incorrect")
			}
		})
	}
	// Synthetic seal syntax is not proof of actual numeric rows or permissions.
	s := Seal{Schema: SealSchema, PlanSHA256: pathclaim.PlanSHA256, MembershipSHA256: trainingdata.MembershipSHA256, QueryProjectionSHA256: trainingdata.QueryProjectionSHA256, PatchProjectionSHA256: PatchSHA256, SourceEvidenceSHA256: strings.Repeat("a", 64), CostExamplesSHA256: strings.Repeat("b", 64), AvailabilityCheckpointSHA256: strings.Repeat("c", 64), NumericRowsSHA256: strings.Repeat("d", 64), RunnerRevision: strings.Repeat("e", 40), TrainingCoverage: 7335, ValidationCoverage: 5686, ProtectedFinal: 2402, TrainingEligible: 1, ValidationEligible: 1}
	for i, repo := range pathclaim.ValidationRepositories() {
		s.ValidationRepositories[i] = pathclaim.RepositoryCount{Repository: repo, Eligible: 1}
	}
	if _, err := s.Readiness(strings.Repeat("f", 64)); err == nil {
		t.Fatal("syntactic hashes or compact counts bypassed 2400+2400 gate")
	}
	f := newInputFixture(t)
	p, err := f.joined()
	if err != nil {
		t.Fatal("owned compact join failed")
	}
	p.Coverage[0].Tasks, p.Coverage[1].Tasks = 7335, 5686
	seal, err := MakeSeal(p, Files{Source: File{SHA256: strings.Repeat("a", 64)}, Costs: File{SHA256: strings.Repeat("b", 64)}, Checkpoint: File{SHA256: strings.Repeat("c", 64)}}, strings.Repeat("e", 40))
	if err == nil {
		if _, err := seal.Readiness(strings.Repeat("f", 64)); err == nil {
			t.Fatal("edited denominator converted actual compact rows into readiness")
		}
	}
}

func TestReviewedRepositoryCountsAndStatusesUseActualSnapshots(t *testing.T) {
	f := newInputFixture(t)
	s := Source{Snapshots: f.source, Repositories: []SourceRepository{{Role: "train", Repository: "pypa/pip", Tasks: 1, Counts: Conditions{Candidate: 1}}, {Role: "validation", Repository: "numpy/numpy", Tasks: 1, Counts: Conditions{Candidate: 1}}}}
	if err := verifyRepositoryCounts(s); err != nil {
		t.Fatal("owned repository totals incorrect")
	}
	s.Repositories[0].Counts.Candidate++
	if err := verifyRepositoryCounts(s); err == nil {
		t.Fatal("repository totals disagreed with actual snapshots")
	}
	var c Conditions
	for _, status := range []string{"conditional_local_numeric_candidate", "checkpoint_root_unavailable", "pending_restrictive_alpha_beta_eula", "pending_file_scoped_prefect_license", "missing_primary_license_candidate"} {
		if err := increment(&c, status); err != nil {
			t.Fatal("registered source state rejected")
		}
	}
	if c != (Conditions{Candidate: 1, Unavailable: 1, Restrictive: 1, FileScoped: 1, Missing: 1}) || increment(&c, "unknown") == nil {
		t.Fatal("source condition coverage changed")
	}
}

func TestPinnedSealRejectsShortAndExpandedFixedArrays(t *testing.T) {
	b, e := json.Marshal(Seal{})
	if e != nil {
		t.Fatal(e)
	}
	var shape map[string]json.RawMessage
	if e = json.Unmarshal(b, &shape); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"ValidationRepositories", "Coverage"} {
		var values []json.RawMessage
		if e = json.Unmarshal(shape[field], &values); e != nil {
			t.Fatal(e)
		}
		for _, size := range []int{len(values) - 1, len(values) + 1} {
			t.Run(fmt.Sprintf("%s_%d", field, size), func(t *testing.T) {
				copyShape := map[string]json.RawMessage{}
				for k, v := range shape {
					copyShape[k] = v
				}
				next := append([]json.RawMessage(nil), values...)
				if size < len(values) {
					next = next[:size]
				} else {
					next = append(next, values[0])
				}
				copyShape[field], e = json.Marshal(next)
				if e != nil {
					t.Fatal(e)
				}
				raw, e := json.Marshal(copyShape)
				if e != nil {
					t.Fatal(e)
				}
				p := filepath.Join(t.TempDir(), "owned-seal.json")
				if e = os.WriteFile(p, raw, 0600); e != nil {
					t.Fatal(e)
				}
				if _, e = ReadSeal(File{Path: p, SHA256: hash(raw)}); e == nil {
					t.Fatal("invalid fixed array accepted")
				}
			})
		}
	}
}
