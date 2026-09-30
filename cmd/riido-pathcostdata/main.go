// riido-pathcostdata prepares private numeric paired-search examples offline.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

const planSHA = "68b7f4f3ded46a9c7d3a96ada0824d046e193f454febf4d65a6095c19baf10de"
const checkpointSHA = "2cfeff79d68a652181b6129914d25062ddce3c33f698106178502fcd709fa6fa"
const patchSHA = "6bed55ca053705e59b1362f982b86cbd977fa3d216312a018b1657dfb9fd284b"

type config struct {
	out, planPath, planSHA, checkpointPath, checkpointSHA string
}

type plan struct {
	Schema                 string `json:"schema"`
	MembershipSHA256       string `json:"membership_sha256"`
	QueryProjectionSHA256  string `json:"query_projection_sha256"`
	PatchProjectionSHA256  string `json:"patch_projection_sha256"`
	AvailabilityCheckpoint string `json:"availability_checkpoint"`
	QueryReader            string `json:"query_reader"`
	FeatureOrder           string `json:"feature_order"`
	Costs                  string `json:"costs"`
	Denominators           string `json:"denominators"`
	Integrity              string `json:"integrity"`
	Command                string `json:"command"`
	Validation             string `json:"validation"`
	ProductionReady        *bool  `json:"production_ready"`
}

type checkpoint struct {
	Role, Repository, ID, BaseCommit  string
	RootSHA256, TreeID, CatalogSHA256 string
	Status                            string
	Match                             filelabels.CatalogMatch
}
type example struct {
	Role, Repository, ID, Status, Fallback               string
	Features                                             [searchclaim.PathDimension]float64
	FeaturesAvailable, AuxiliaryRequested, AuxiliaryUsed bool
	Work                                                 fileeval.WorkStats
	Baseline, Helper                                     fileeval.Score
	BaselinePages, HelperPages, Gain                     int
}
type totals struct {
	Role, Repository                                                                              string
	Tasks, Unavailable, ParseFailures, Unsupported, NoOldTargets, Unmapped, RankingErrors, Scored int
	FeaturesAvailable, AuxiliaryRequested, AuxiliaryUsed, Fallbacks                               int
	Work                                                                                          fileeval.WorkStats
	BaselinePages, HelperPages, Wins, Losses, Ties                                                int
	BaselineHit1, BaselineHit10, BaselineHit100, BaselineAll10                                    int
	HelperHit1, HelperHit10, HelperHit100, HelperAll10                                            int
}
type report struct {
	Schema, PlanSHA256, MembershipSHA256, QuerySHA256, PatchSHA256, CheckpointSHA256, ExamplesSHA256 string
	FeatureSchema                                                                                    string
	PageSize                                                                                         int
	Total                                                                                            totals
	Roles, Repositories                                                                              []totals
	TrainingPerformed, SourceEligibilityComplete, FinalScored, ProductionReady                       bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		return err
	}
	planHash, err := readPlan(cfg.planPath, cfg.planSHA)
	if err != nil {
		return err
	}
	members, err := trainingdata.ReadDevelopment(".cache/training-partition-37/membership.json")
	if err != nil {
		return err
	}
	points, checkpointHash, err := readCheckpoint(cfg.checkpointPath, cfg.checkpointSHA, members)
	if err != nil {
		return err
	}
	labels, err := filelabels.ReadDevelopment(".cache/development-patches-39.jsonl", patchSHA, members)
	if err != nil {
		return err
	}
	rows := make([]example, 0, len(members))
	err = trainingdata.VisitQueries(".cache/training-identity-36.jsonl", members, func(m sweaudit.TaskRole, q sweaudit.Row) error {
		i, ok := slices.BinarySearchFunc(points, m.Task.ID, func(a checkpoint, id string) int { return strings.Compare(a.ID, id) })
		if !ok {
			return fmt.Errorf("missing checkpoint member")
		}
		p := points[i]
		v := example{Role: m.Role, Repository: m.Task.Repository, ID: m.Task.ID}
		// Required roots remain pinned even when a complete catalog is unavailable.
		paths, err := cachedPaths(p, ".cache/training-roots-38", ".cache/training-catalogs-38")
		if err != nil {
			return err
		}
		if p.CatalogSHA256 == "" {
			v.Status = "checkpoint_unavailable"
			rows = append(rows, v)
			return nil
		}
		// Gold is not an input to ranking, features or the always-on selector.
		selection, rankErr := fileeval.RankSelected(q.Request, paths, func([searchclaim.PathDimension]float64) (bool, error) { return true, nil })
		v.Features, v.FeaturesAvailable = selection.Features, selection.FeaturesAvailable
		v.AuxiliaryRequested, v.AuxiliaryUsed = selection.AuxiliaryRequested, selection.AuxiliaryUsed
		v.Work, v.Fallback = selection.Work, selection.Fallback
		if rankErr != nil {
			v.Status = "ranking_error"
		} else {
			j, found := slices.BinarySearchFunc(labels, m.Task.ID, func(a filelabels.DevelopmentLabel, id string) int { return strings.Compare(a.ID, id) })
			if !found {
				return fmt.Errorf("missing target")
			}
			if err := score(&v, paths, selection, labels[j]); err != nil {
				return err
			}
		}
		rows = append(rows, v)
		return nil
	})
	if err != nil {
		return err
	}
	if len(rows) != len(members) {
		return fmt.Errorf("output denominator mismatch")
	}
	slices.SortFunc(rows, func(a, b example) int { return strings.Compare(a.ID, b.ID) })
	r := report{Schema: "riido-path-cost-data-report-v1", PlanSHA256: planHash, MembershipSHA256: trainingdata.MembershipSHA256, QuerySHA256: trainingdata.QueryProjectionSHA256, PatchSHA256: patchSHA, CheckpointSHA256: checkpointHash, FeatureSchema: searchclaim.PathSchema, PageSize: 20}
	r.Total, r.Roles, r.Repositories, err = summarize(rows)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	r.ExamplesSHA256 = digest(raw)
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	// No publication before every source, callback and aggregate has succeeded.
	if err = os.Mkdir(cfg.out, 0700); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(cfg.out, "examples.json"), raw, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cfg.out, "results.json"), append(b, '\n'), 0600)
}

func parseConfig(args []string) (config, error) {
	c := config{planPath: "experiments/path-cost-data/plan-43.json", planSHA: planSHA, checkpointPath: ".cache/development-join-40-resume2/evidence.json", checkpointSHA: checkpointSHA}
	f := flag.NewFlagSet("riido-pathcostdata", flag.ContinueOnError)
	f.StringVar(&c.out, "out", "", "new private output directory")
	f.StringVar(&c.planPath, "plan", c.planPath, "pinned preparation plan; requires all four overrides")
	f.StringVar(&c.planSHA, "plan-sha256", c.planSHA, "SHA256 of preparation plan; requires all four overrides")
	f.StringVar(&c.checkpointPath, "checkpoint", c.checkpointPath, "pinned availability checkpoint; requires all four overrides")
	f.StringVar(&c.checkpointSHA, "checkpoint-sha256", c.checkpointSHA, "SHA256 of availability checkpoint; requires all four overrides")
	if err := f.Parse(args); err != nil {
		return config{}, err
	}
	if c.out == "" || f.NArg() != 0 {
		return config{}, fmt.Errorf("require output and no positional arguments")
	}
	count := 0
	f.Visit(func(v *flag.Flag) {
		if v.Name == "plan" || v.Name == "plan-sha256" || v.Name == "checkpoint" || v.Name == "checkpoint-sha256" {
			count++
		}
	})
	if count != 0 && count != 4 {
		return config{}, fmt.Errorf("require all four plan and checkpoint overrides together")
	}
	if c.planPath == "" || c.checkpointPath == "" || !hexDigest(c.planSHA, 32) || !hexDigest(c.checkpointSHA, 32) {
		return config{}, fmt.Errorf("invalid plan or checkpoint path/hash")
	}
	return c, nil
}

func readPlan(path, expected string) (string, error) {
	b, actual, err := readPinned(path, expected, 1<<20, "plan")
	if err != nil {
		return "", err
	}
	var p plan
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(&p) != nil || d.Decode(new(any)) != io.EOF {
		return "", fmt.Errorf("plan schema invalid")
	}
	if p.Schema != "riido-path-cost-data-plan-v1" || p.MembershipSHA256 != trainingdata.MembershipSHA256 || p.QueryProjectionSHA256 != trainingdata.QueryProjectionSHA256 || p.PatchProjectionSHA256 != patchSHA || p.ProductionReady == nil || *p.ProductionReady {
		return "", fmt.Errorf("plan changes fixed preparation scope")
	}
	for _, text := range []string{p.AvailabilityCheckpoint, p.QueryReader, p.FeatureOrder, p.Costs, p.Denominators, p.Integrity, p.Validation} {
		if strings.TrimSpace(text) == "" {
			return "", fmt.Errorf("plan policy missing")
		}
	}
	// Policy text documents the experiment; it never configures source, features,
	// page size, fitting, final scoring or production activation in this command.
	return actual, nil
}

func readPinned(path, expected string, limit int64, name string) ([]byte, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, "", err
	}
	actual := digest(b)
	if int64(len(b)) > limit || !hexDigest(expected, 32) || actual != expected {
		return nil, "", fmt.Errorf("%s size or hash mismatch", name)
	}
	return b, actual, nil
}

func readCheckpoint(path, expected string, members []sweaudit.TaskRole) ([]checkpoint, string, error) {
	b, actual, err := readPinned(path, expected, 16<<20, "checkpoint")
	if err != nil {
		return nil, "", err
	}
	var rows []checkpoint
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(&rows) != nil || d.Decode(new(any)) != io.EOF {
		return nil, "", fmt.Errorf("checkpoint schema invalid")
	}
	if len(rows) != len(members) || len(rows) == 0 {
		return nil, "", fmt.Errorf("checkpoint count mismatch")
	}
	slices.SortFunc(rows, func(a, b checkpoint) int { return strings.Compare(a.ID, b.ID) })
	ms := slices.Clone(members)
	slices.SortFunc(ms, func(a, b sweaudit.TaskRole) int { return strings.Compare(a.Task.ID, b.Task.ID) })
	for i, p := range rows {
		m := ms[i]
		if (m.Role != "train" && m.Role != "validation") || p.Role != m.Role || p.Repository != m.Task.Repository || p.ID != m.Task.ID || p.BaseCommit != m.Task.BaseCommit || (i > 0 && rows[i-1].ID == p.ID) {
			return nil, "", fmt.Errorf("checkpoint membership mismatch")
		}
		switch p.Status {
		case "root_unavailable":
			if p.RootSHA256 != "" || p.TreeID != "" || p.CatalogSHA256 != "" {
				return nil, "", fmt.Errorf("unexpected unavailable root digest")
			}
		case "catalog_unavailable":
			if !hexDigest(p.RootSHA256, 32) || !hexDigest(p.TreeID, 20) || p.CatalogSHA256 != "" {
				return nil, "", fmt.Errorf("invalid unavailable catalog identity")
			}
		case "matched", "unusable_target":
			if !hexDigest(p.RootSHA256, 32) || !hexDigest(p.TreeID, 20) || !hexDigest(p.CatalogSHA256, 32) {
				return nil, "", fmt.Errorf("invalid available catalog identity")
			}
		default:
			return nil, "", fmt.Errorf("unknown checkpoint status")
		}
	}
	return rows, actual, nil
}
func hexDigest(s string, n int) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == n && strings.ToLower(s) == s
}
func offline(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("checkpoint-required cache missing; network disabled")
}
func cachedPaths(p checkpoint, roots, catalogs string) ([]string, error) {
	if p.RootSHA256 == "" {
		if p.TreeID != "" || p.CatalogSHA256 != "" {
			return nil, fmt.Errorf("catalog lacks required root identity")
		}
		return nil, nil
	}
	b, err := githubmeta.Root(context.Background(), p.Repository, p.BaseCommit, roots, offline)
	if err != nil {
		return nil, err
	}
	if digest(b) != p.RootSHA256 {
		return nil, fmt.Errorf("checkpoint root digest mismatch")
	}
	tree, _, err := sweaudit.TreeEntries(p.Repository, b)
	if err != nil {
		return nil, err
	}
	if tree != p.TreeID {
		return nil, fmt.Errorf("checkpoint tree mismatch")
	}
	if p.CatalogSHA256 == "" {
		// A root-only checkpoint is verified without consulting newer catalogs.
		return nil, nil
	}
	c, err := githubmeta.RecursiveCatalog(context.Background(), p.Repository, tree, catalogs, offline)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if digest(raw) != p.CatalogSHA256 {
		return nil, fmt.Errorf("checkpoint catalog digest mismatch")
	}
	var paths []string
	for _, entry := range c.Tree {
		if entry.Type == "blob" {
			paths = append(paths, entry.Path)
		}
	}
	return paths, nil
}
func score(v *example, paths []string, s fileeval.Selection, l filelabels.DevelopmentLabel) error {
	if v.ID != l.ID {
		return fmt.Errorf("scoring identity mismatch")
	}
	switch {
	case l.ParseError:
		v.Status = "parse_failure"
		return nil
	case l.Result.UnsupportedBlocks > 0:
		v.Status = "unsupported_target"
		return nil
	case len(l.Result.OldPaths) == 0:
		v.Status = "no_old_target"
		return nil
	}
	a, err := fileeval.Measure(paths, s.BaselineOrder, l.Result.OldPaths)
	if err != nil {
		return err
	}
	b, err := fileeval.Measure(paths, s.Order, l.Result.OldPaths)
	if err != nil {
		return err
	}
	if a.Mapped != a.Targets || b.Mapped != b.Targets {
		v.Status = "unmapped_target"
		return nil
	}
	v.Baseline, v.Helper = a, b
	v.BaselinePages = (a.FirstRank + 19) / 20
	v.HelperPages = (b.FirstRank + 19) / 20
	v.Gain = v.BaselinePages - v.HelperPages
	v.Status = "scored"
	return nil
}
func summarize(rows []example) (totals, []totals, []totals, error) {
	var total totals
	var roles, repos []totals
	ordered := slices.Clone(rows)
	slices.SortFunc(ordered, func(a, b example) int {
		if a.Role != b.Role {
			return strings.Compare(a.Role, b.Role)
		}
		if a.Repository != b.Repository {
			return strings.Compare(a.Repository, b.Repository)
		}
		return strings.Compare(a.ID, b.ID)
	})
	for _, v := range ordered {
		if len(roles) == 0 || roles[len(roles)-1].Role != v.Role {
			roles = append(roles, totals{Role: v.Role})
		}
		if len(repos) == 0 || repos[len(repos)-1].Role != v.Role || repos[len(repos)-1].Repository != v.Repository {
			repos = append(repos, totals{Role: v.Role, Repository: v.Repository})
		}
		for _, t := range []*totals{&total, &roles[len(roles)-1], &repos[len(repos)-1]} {
			if err := t.add(v); err != nil {
				return totals{}, nil, nil, err
			}
		}
	}
	return total, roles, repos, nil
}
func (t *totals) add(v example) error {
	t.Tasks++
	t.Work.BaselineRankAttempts += v.Work.BaselineRankAttempts
	t.Work.AuxiliaryBuildAttempts += v.Work.AuxiliaryBuildAttempts
	t.Work.AuxiliaryRankAttempts += v.Work.AuxiliaryRankAttempts
	if v.FeaturesAvailable {
		t.FeaturesAvailable++
	}
	if v.AuxiliaryRequested {
		t.AuxiliaryRequested++
	}
	if v.AuxiliaryUsed {
		t.AuxiliaryUsed++
	}
	if v.Fallback != "" {
		t.Fallbacks++
	}
	switch v.Status {
	case "checkpoint_unavailable":
		t.Unavailable++
	case "parse_failure":
		t.ParseFailures++
	case "unsupported_target":
		t.Unsupported++
	case "no_old_target":
		t.NoOldTargets++
	case "unmapped_target":
		t.Unmapped++
	case "ranking_error":
		t.RankingErrors++
	case "scored":
		t.Scored++
		t.BaselinePages += v.BaselinePages
		t.HelperPages += v.HelperPages
		if v.Gain > 0 {
			t.Wins++
		} else if v.Gain < 0 {
			t.Losses++
		} else {
			t.Ties++
		}
		if v.Baseline.Hit1 {
			t.BaselineHit1++
		}
		if v.Baseline.Hit10 {
			t.BaselineHit10++
		}
		if v.Baseline.Hit100 {
			t.BaselineHit100++
		}
		if v.Baseline.All10 {
			t.BaselineAll10++
		}
		if v.Helper.Hit1 {
			t.HelperHit1++
		}
		if v.Helper.Hit10 {
			t.HelperHit10++
		}
		if v.Helper.Hit100 {
			t.HelperHit100++
		}
		if v.Helper.All10 {
			t.HelperAll10++
		}
	default:
		return fmt.Errorf("unknown example status")
	}
	return nil
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
