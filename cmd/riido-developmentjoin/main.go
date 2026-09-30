// riido-developmentjoin checks target coverage without ranking or network access.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

const projectionSHA = "6bed55ca053705e59b1362f982b86cbd977fa3d216312a018b1657dfb9fd284b"
const planSHA = "fdd3ebe4c3bc117c01cc01834b52ffd3726cba71b82e965ff3036f86090db8ce"

var errUnavailable = errors.New("verified catalog cache unavailable; network disabled")

func offline(context.Context, string) ([]byte, error) { return nil, errUnavailable }

type coverage struct {
	Role, Repository                                                                   string
	Tasks, Roots, Catalogs, UnavailableRoots, UnavailableCatalogs                      int
	LabelParseFailures, UnsupportedLabelTasks, NoOldPathTasks, SupportedTargetTasks    int
	MatchedTasks, AllFileTargetsMapped, TasksWithMissingPaths, TasksWithNonFileTargets int
	OldTargets, RegularFiles, Symlinks, Submodules, Directories, MissingPaths          int
}
type evidence struct {
	Role, Repository, ID, BaseCommit  string
	RootSHA256, TreeID, CatalogSHA256 string
	Status                            string
	Match                             filelabels.CatalogMatch
}
type report struct {
	Schema, PlanSHA256, MembershipSHA256, ProjectionSHA256, EvidenceSHA256                  string
	Total                                                                                   coverage
	Repositories                                                                            []coverage
	AllCatalogsAvailable, AllSupportedTargetsVerified                                       bool
	HistoricalLicensesReviewed, IssueTextLicenseResolved, TrainingApproved, ProductionReady bool
}

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	out := flag.String("out", "", "new private output directory")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output")
	}
	b, e := os.ReadFile("experiments/development-join/plan-40.json")
	if e != nil {
		return e
	}
	if digest(b) != planSHA {
		return fmt.Errorf("plan mismatch")
	}
	members, e := trainingdata.ReadDevelopment(".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	labels, e := filelabels.ReadDevelopment(".cache/development-patches-39.jsonl", projectionSHA, members)
	if e != nil {
		return e
	}
	var all []evidence
	var counts []coverage
	for _, m := range members {
		if m.Role != "train" && m.Role != "validation" {
			return fmt.Errorf("final join forbidden")
		}
		i, ok := slices.BinarySearchFunc(labels, m.Task.ID, func(a filelabels.DevelopmentLabel, b string) int { return strings.Compare(a.ID, b) })
		if !ok {
			return fmt.Errorf("missing fixed target")
		}
		c, v, e := join(m, labels[i], ".cache/training-roots-38", ".cache/training-catalogs-38")
		if e != nil {
			return e
		}
		all = append(all, v)
		counts = append(counts, c)
	}
	counts = aggregate(counts)
	r := report{Schema: "riido-development-join-report-v1", PlanSHA256: planSHA, MembershipSHA256: trainingdata.MembershipSHA256, ProjectionSHA256: projectionSHA, Repositories: counts}
	for _, c := range counts {
		add(&r.Total, c)
	}
	r.AllCatalogsAvailable = r.Total.Catalogs == len(members)
	r.AllSupportedTargetsVerified = r.AllCatalogsAvailable && r.Total.AllFileTargetsMapped == r.Total.SupportedTargetTasks && r.Total.SupportedTargetTasks > 0
	raw, e := json.MarshalIndent(all, "", "  ")
	if e != nil {
		return e
	}
	raw = append(raw, '\n')
	r.EvidenceSHA256 = digest(raw)
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "evidence.json"), raw, 0600); e != nil {
		return e
	}
	b, e = json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600)
}
func join(m sweaudit.TaskRole, l filelabels.DevelopmentLabel, rootCache, catalogCache string) (coverage, evidence, error) {
	c := coverage{Role: m.Role, Repository: m.Task.Repository, Tasks: 1}
	v := evidence{Role: m.Role, Repository: m.Task.Repository, ID: m.Task.ID, BaseCommit: m.Task.BaseCommit}
	if (m.Role != "train" && m.Role != "validation") || m.Task.ID != l.ID {
		return c, v, fmt.Errorf("join role or ID mismatch")
	}
	supported := false
	switch {
	case l.ParseError:
		c.LabelParseFailures = 1
	case l.Result.UnsupportedBlocks > 0:
		c.UnsupportedLabelTasks = 1
	case len(l.Result.OldPaths) == 0:
		c.NoOldPathTasks = 1
	default:
		c.SupportedTargetTasks = 1
		supported = true
	}
	b, e := githubmeta.Root(context.Background(), m.Task.Repository, m.Task.BaseCommit, rootCache, offline)
	if e != nil {
		if !errors.Is(e, errUnavailable) {
			return c, v, e
		}
		c.UnavailableRoots = 1
		c.UnavailableCatalogs = 1
		v.Status = "root_unavailable"
		return c, v, nil
	}
	c.Roots = 1
	v.RootSHA256 = digest(b)
	tree, _, e := sweaudit.TreeEntries(m.Task.Repository, b)
	if e != nil {
		return c, v, e
	}
	v.TreeID = tree
	// Complete validated catalog cache only. Do not reconstruct or fetch anything.
	catalog, e := githubmeta.RecursiveCatalog(context.Background(), m.Task.Repository, tree, catalogCache, offline)
	if e != nil {
		if !errors.Is(e, errUnavailable) {
			return c, v, e
		}
		c.UnavailableCatalogs = 1
		v.Status = "catalog_unavailable"
		return c, v, nil
	}
	c.Catalogs = 1
	raw, e := json.Marshal(catalog)
	if e != nil {
		return c, v, e
	}
	v.CatalogSHA256 = digest(raw)
	if !supported {
		v.Status = "unusable_target"
		return c, v, nil
	}
	match, e := filelabels.MatchCatalog(l.Result.OldPaths, catalog.Tree)
	if e != nil {
		return c, v, e
	}
	v.Match = match
	v.Status = "matched"
	c.MatchedTasks = 1
	c.OldTargets = match.Targets
	c.RegularFiles = match.RegularFiles
	c.Symlinks = match.Symlinks
	c.Submodules = match.Submodules
	c.Directories = match.Directories
	c.MissingPaths = match.Missing
	if match.Missing > 0 {
		c.TasksWithMissingPaths = 1
	}
	if match.Submodules+match.Directories > 0 {
		c.TasksWithNonFileTargets = 1
	}
	if match.Targets > 0 && match.RegularFiles+match.Symlinks == match.Targets {
		c.AllFileTargetsMapped = 1
	}
	return c, v, nil
}
func aggregate(rows []coverage) []coverage {
	slices.SortFunc(rows, func(a, b coverage) int {
		if a.Role != b.Role {
			return strings.Compare(a.Role, b.Role)
		}
		return strings.Compare(a.Repository, b.Repository)
	})
	n := 0
	for _, r := range rows {
		if n == 0 || rows[n-1].Role != r.Role || rows[n-1].Repository != r.Repository {
			rows[n] = r
			n++
		} else {
			add(&rows[n-1], r)
		}
	}
	return rows[:n]
}
func add(a *coverage, b coverage) {
	a.Tasks += b.Tasks
	a.Roots += b.Roots
	a.Catalogs += b.Catalogs
	a.UnavailableRoots += b.UnavailableRoots
	a.UnavailableCatalogs += b.UnavailableCatalogs
	a.LabelParseFailures += b.LabelParseFailures
	a.UnsupportedLabelTasks += b.UnsupportedLabelTasks
	a.NoOldPathTasks += b.NoOldPathTasks
	a.SupportedTargetTasks += b.SupportedTargetTasks
	a.MatchedTasks += b.MatchedTasks
	a.AllFileTargetsMapped += b.AllFileTargetsMapped
	a.TasksWithMissingPaths += b.TasksWithMissingPaths
	a.TasksWithNonFileTargets += b.TasksWithNonFileTargets
	a.OldTargets += b.OldTargets
	a.RegularFiles += b.RegularFiles
	a.Symlinks += b.Symlinks
	a.Submodules += b.Submodules
	a.Directories += b.Directories
	a.MissingPaths += b.MissingPaths
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
