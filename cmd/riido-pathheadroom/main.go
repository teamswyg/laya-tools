// riido-pathheadroom screens fixed helpers on reviewed training data only.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/pathinput"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/internal/trainingdata"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const planSHA = "f7e740f0ff0da3706592f79d30d240779d8e1261504ef4f4f4dec937904e61fd"
const sealSHA = "3e8793cb0f3c5d23473fedfd3d4faed58c05dbd3949e77308ca1f8ee56f39187"
const sourceSHA = "1da77c56769b6e60c8dd9ca947ac60265737984e6129223a424787354644a115"
const costSHA = "fb8ce49e0111f7da0a2070381bd7109a556ed6e3a36aba7ab2e1d58852165084"
const pointSHA = "98c28f1c858fbd2609c4e0e1088677f0a77873d21a0c5d70bfd5b970c0607cf4"

type config struct{ out string }
type metrics struct {
	Policy, Repository                                           string
	Tasks, Pages, Wins, Losses, Ties, PositiveGain, NegativeGain int
	Hit1, Hit10, Hit100, All10, Hints, Attempts, Fallbacks       int
	IndexBuildAttempts, IndexSearchAttempts, AnchorComparisons   int
	OraclePageReduction, MacroHit10                              float64
	NecessaryFivePercentHeadroom                                 bool
}
type privateCase struct {
	ID, Repository   string
	Baseline, Legacy fileeval.Score
	Helpers          [3]fileeval.Score
	HintCounts       [3]int
	Fallbacks        [3]string
	Work             [3]hintsearch.HelperWork
}
type report struct {
	Schema, Status, PlanSHA256, InputSealSHA256, RunnerRevision                                                string
	Input                                                                                                      pathinput.Seal
	TrainingCoverage                                                                                           pathinput.Coverage
	Policies, Repositories                                                                                     []metrics
	RankingsSHA256, PrivateCasesSHA256                                                                         string
	Cache                                                                                                      fileeval.CacheStats
	TrainingPerformed, ValidationScored, FinalScored, GPUUsed, ActualModelEvaluationPerformed, ProductionReady bool
}

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func parse(args []string) (config, error) {
	var c config
	f := flag.NewFlagSet("riido-pathheadroom", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.StringVar(&c.out, "out", "", "new private output directory under .cache; training-only offline screen")
	if e := f.Parse(args); e != nil {
		return c, fmt.Errorf("invalid_arguments")
	}
	if f.NArg() != 0 || filepath.IsAbs(c.out) || filepath.Clean(c.out) != c.out || !strings.HasPrefix(c.out, ".cache/") || c.out == ".cache/" {
		return c, fmt.Errorf("require a new private output under .cache")
	}
	return c, nil
}

func privateAncestors(out string) error {
	if filepath.IsAbs(out) || filepath.Clean(out) != out || !strings.HasPrefix(out, ".cache/") {
		return fmt.Errorf("require a canonical private output")
	}
	parent := filepath.Dir(out)
	for parent != "." {
		info, e := os.Lstat(parent)
		if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("require existing private directory ancestors without symlinks")
		}
		parent = filepath.Dir(parent)
	}
	return nil
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func pinned(name, expected string, limit int64, v any) error {
	f, e := os.Open(name)
	if e != nil {
		return fmt.Errorf("pinned input unavailable")
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != expected {
		return fmt.Errorf("pinned input size or hash mismatch")
	}
	if v == nil {
		return nil
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("pinned input schema mismatch")
	}
	return nil
}
func runnerRevision() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("require committed Go runner")
	}
	var rev string
	dirty := true
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			rev = setting.Value
		case "vcs.modified":
			dirty = setting.Value != "false"
		}
	}
	if dirty || len(rev) != 40 {
		return "", fmt.Errorf("require a clean committed Go build")
	}
	return rev, nil
}
func files() pathinput.Files {
	return pathinput.Files{
		Source:     pathinput.File{Path: ".cache/source-review-47/evidence.json", SHA256: sourceSHA},
		Costs:      pathinput.File{Path: ".cache/path-cost-data-47/examples.json", SHA256: costSHA},
		Checkpoint: pathinput.File{Path: ".cache/development-join-47/evidence.json", SHA256: pointSHA},
	}
}
func verifySeal(p pathinput.Prepared, f pathinput.Files) (pathinput.Seal, error) {
	s, e := pathinput.ReadSeal(pathinput.File{Path: "experiments/path-cost-claim/input-46.json", SHA256: sealSHA})
	if e != nil {
		return s, e
	}
	// Preserve the old input runner identity: this command does not fit its model.
	actual, e := pathinput.MakeSeal(p, f, s.RunnerRevision)
	if e != nil || !reflect.DeepEqual(actual, s) {
		return s, fmt.Errorf("actual inputs differ from frozen experiment46 seal")
	}
	if s.TrainingEligible != 4456 || s.ValidationEligible != 2599 || s.ProtectedFinal != 2402 {
		return s, fmt.Errorf("frozen role counts changed")
	}
	return s, nil
}
func offline(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("reviewed catalog cache missing; network disabled")
}
func pathsFor(p pathinput.Checkpoint) ([]string, error) {
	b, e := githubmeta.Root(context.Background(), p.Repository, p.BaseCommit, ".cache/training-roots-38", offline)
	if e != nil || digest(b) != p.RootSHA256 {
		return nil, fmt.Errorf("reviewed root identity mismatch")
	}
	tree, _, e := sweaudit.TreeEntries(p.Repository, b)
	if e != nil || tree != p.TreeID {
		return nil, fmt.Errorf("reviewed tree identity mismatch")
	}
	c, e := githubmeta.RecursiveCatalog(context.Background(), p.Repository, tree, ".cache/training-catalogs-38", offline)
	if e != nil {
		return nil, fmt.Errorf("reviewed catalog unavailable")
	}
	raw, e := json.Marshal(c)
	if e != nil || digest(raw) != p.CatalogSHA256 {
		return nil, fmt.Errorf("reviewed catalog identity mismatch")
	}
	var paths []string
	for _, entry := range c.Tree {
		if entry.Type == "blob" {
			paths = append(paths, entry.Path)
		}
	}
	return paths, nil
}
func frame(h hash.Hash, s string) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(len(s)))
	h.Write(b[:])
	h.Write([]byte(s))
}
func orderHash(h hash.Hash, id string, order []int) {
	frame(h, id)
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(len(order)))
	h.Write(b[:])
	for _, position := range order {
		binary.LittleEndian.PutUint64(b[:], uint64(position))
		h.Write(b[:])
	}
}
func pages(s fileeval.Score) int { return (s.FirstRank + 19) / 20 }
func add(m *metrics, base, score fileeval.Score, hints, attempts int, fallback string, work hintsearch.HelperWork) {
	m.Tasks++
	m.Pages += pages(score)
	gain := pages(base) - pages(score)
	switch {
	case gain > 0:
		m.Wins++
		m.PositiveGain += gain
	case gain < 0:
		m.Losses++
		m.NegativeGain -= gain
	default:
		m.Ties++
	}
	if score.Hit1 {
		m.Hit1++
	}
	if score.Hit10 {
		m.Hit10++
	}
	if score.Hit100 {
		m.Hit100++
	}
	if score.All10 {
		m.All10++
	}
	m.Hints += hints
	m.Attempts += attempts
	if fallback != "" {
		m.Fallbacks++
	}
	m.IndexSearchAttempts += work.IndexSearchAttempts
	m.IndexBuildAttempts += work.IndexBuildAttempts
	m.AnchorComparisons += work.AnchorComparisons
}
func finish(m *metrics, basePages int) {
	if basePages > 0 {
		m.OraclePageReduction = float64(m.PositiveGain) / float64(basePages)
	}
	m.NecessaryFivePercentHeadroom = basePages > 0 && m.PositiveGain*100 >= basePages*5
}
func summarize(cases []privateCase, builds [3]hintsearch.HelperWork) ([]metrics, []metrics, error) {
	policies := []metrics{{Policy: "baseline"}, {Policy: "legacy-normalized-full"}, {Policy: "normalized-positive64"}, {Policy: "joined-fields-rrf64"}, {Policy: "explicit-path64"}}
	var repos []metrics
	for _, c := range cases {
		if c.Baseline.Targets < 1 || c.Baseline.Mapped != c.Baseline.Targets || c.Baseline.FirstRank < 1 {
			return nil, nil, fmt.Errorf("invalid complete training score")
		}
		for i := range policies {
			score, hints, attempts, fallback, work := c.Baseline, 0, 0, "", hintsearch.HelperWork{}
			if i == 1 {
				score, attempts = c.Legacy, 1
			}
			if i >= 2 {
				j := i - 2
				score, hints, attempts, fallback, work = c.Helpers[j], c.HintCounts[j], 1, c.Fallbacks[j], c.Work[j]
			}
			if score.Targets != c.Baseline.Targets || score.Mapped != score.Targets || score.FirstRank < 1 {
				return nil, nil, fmt.Errorf("incomplete helper training score")
			}
			add(&policies[i], c.Baseline, score, hints, attempts, fallback, work)
			key := policies[i].Policy + "\x00" + c.Repository
			pos, found := slices.BinarySearchFunc(repos, key, func(m metrics, k string) int { return strings.Compare(m.Policy+"\x00"+m.Repository, k) })
			if !found {
				repos = slices.Insert(repos, pos, metrics{Policy: policies[i].Policy, Repository: c.Repository})
			}
			add(&repos[pos], c.Baseline, score, hints, attempts, fallback, work)
		}
	}
	for i := range policies {
		finish(&policies[i], policies[0].Pages)
		if i >= 2 {
			if policies[i].IndexBuildAttempts != builds[i-2].IndexBuildAttempts {
				return nil, nil, fmt.Errorf("helper construction accounting mismatch")
			}
		}
		count := 0
		for _, r := range repos {
			if r.Policy == policies[i].Policy {
				policies[i].MacroHit10 += float64(r.Hit10) / float64(r.Tasks)
				count++
			}
		}
		if count > 0 {
			policies[i].MacroHit10 /= float64(count)
		}
	}
	for i := range repos {
		key := "baseline\x00" + repos[i].Repository
		pos, found := slices.BinarySearchFunc(repos, key, func(m metrics, k string) int { return strings.Compare(m.Policy+"\x00"+m.Repository, k) })
		if !found {
			return nil, nil, fmt.Errorf("missing repository baseline")
		}
		finish(&repos[i], repos[pos].Pages)
	}
	return policies, repos, nil
}
func run(args []string) error {
	cfg, e := parse(args)
	if e != nil {
		return e
	}
	if _, e = os.Lstat(cfg.out); !os.IsNotExist(e) {
		return fmt.Errorf("output must be new")
	}
	if e = privateAncestors(cfg.out); e != nil {
		return e
	}
	if e = pinned("experiments/path-helper-headroom/plan-48.json", planSHA, 64<<10, nil); e != nil {
		return e
	}
	rev, e := runnerRevision()
	if e != nil {
		return e
	}
	f := files()
	prepared, e := pathinput.Load(f, pathinput.Caches{Roots: ".cache/training-roots-38", Blobs: ".cache/training-license-blobs-38", Subtrees: ".cache/training-subtrees-38"}, ".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	seal, e := verifySeal(prepared, f)
	if e != nil {
		return e
	}
	members, e := trainingdata.ReadDevelopment(".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	labels, e := filelabels.ReadDevelopment(".cache/development-patches-39.jsonl", pathinput.PatchSHA256, members)
	if e != nil {
		return e
	}
	var points []pathinput.Checkpoint
	if e = pinned(f.Checkpoint.Path, pointSHA, 16<<20, &points); e != nil {
		return e
	}
	slices.SortFunc(points, func(a, b pathinput.Checkpoint) int { return strings.Compare(a.ID, b.ID) })
	var costs []pathinput.Cost
	for _, c := range prepared.Costs {
		if c.Role == "train" {
			costs = append(costs, c)
		}
	}
	slices.SortFunc(costs, func(a, b pathinput.Cost) int { return strings.Compare(a.ID, b.ID) })
	if len(costs) != 4456 {
		return fmt.Errorf("training eligibility changed")
	}
	var cases []privateCase
	var ranker fileeval.Ranker
	var previous []string
	var helper *hintsearch.PathHelpers
	var builds [3]hintsearch.HelperWork
	rankingHash := sha256.New()
	e = trainingdata.VisitQueries(".cache/training-identity-36.jsonl", members, func(m sweaudit.TaskRole, q sweaudit.Row) error {
		// Validation and protected-final requests never reach a new ranker.
		if m.Role != "train" {
			return nil
		}
		i, ok := slices.BinarySearchFunc(costs, m.Task.ID, func(c pathinput.Cost, id string) int { return strings.Compare(c.ID, id) })
		if !ok {
			return nil
		}
		cost := costs[i]
		j, ok := slices.BinarySearchFunc(points, m.Task.ID, func(p pathinput.Checkpoint, id string) int { return strings.Compare(p.ID, id) })
		if !ok {
			return fmt.Errorf("missing reviewed training checkpoint")
		}
		paths, e := pathsFor(points[j])
		if e != nil {
			return e
		}
		orders, fallback, e := ranker.Rank(q.Request, paths)
		if e != nil || fallback {
			return fmt.Errorf("legacy training rank replay failed")
		}
		var requestBuilds [3]hintsearch.HelperWork
		if helper == nil || !slices.Equal(previous, paths) {
			helper, e = hintsearch.NewPathHelpers(paths)
			if e != nil {
				return fmt.Errorf("helper path catalog failed")
			}
			previous = slices.Clone(paths)
			work := helper.BuildWork()
			for k := range builds {
				builds[k].IndexBuildAttempts += work[k].IndexBuildAttempts
				requestBuilds[k] = work[k]
			}
		}
		hints, e := helper.Rank(q.Request, orders[fileeval.Baseline])
		if e != nil || len(hints) != 3 {
			return fmt.Errorf("helper ranking failed")
		}
		// Inspect training labels only after every label-free ranking finishes.
		l, ok := slices.BinarySearchFunc(labels, m.Task.ID, func(l filelabels.DevelopmentLabel, id string) int { return strings.Compare(l.ID, id) })
		if !ok {
			return fmt.Errorf("missing reviewed training label")
		}
		gold := labels[l].Result.OldPaths
		c := privateCase{ID: m.Task.ID, Repository: m.Task.Repository}
		c.Baseline, e = fileeval.Measure(paths, orders[fileeval.Baseline], gold)
		if e != nil {
			return e
		}
		c.Legacy, e = fileeval.Measure(paths, orders[fileeval.Interleaved], gold)
		if e != nil {
			return e
		}
		if c.Baseline != cost.Baseline || c.Legacy != cost.Helper || pages(c.Baseline) != cost.BaselinePages || pages(c.Legacy) != cost.HelperPages {
			return fmt.Errorf("legacy training costs differ from frozen input")
		}
		frame(rankingHash, m.Task.ID)
		orderHash(rankingHash, "baseline", orders[fileeval.Baseline])
		orderHash(rankingHash, "legacy-normalized-full", orders[fileeval.Interleaved])
		for k, h := range hints {
			if h.ID != hintsearch.PathHelperIDs()[k] || !h.Attempted {
				return fmt.Errorf("fixed helper identity or attempt accounting changed")
			}
			c.Helpers[k], e = fileeval.Measure(paths, h.Order, gold)
			if e != nil {
				return e
			}
			c.HintCounts[k], c.Fallbacks[k], c.Work[k] = h.HintCount, h.Fallback, h.Work
			c.Work[k].IndexBuildAttempts += requestBuilds[k].IndexBuildAttempts
			orderHash(rankingHash, h.ID, h.Order)
		}
		cases = append(cases, c)
		return nil
	})
	if e != nil {
		return e
	}
	if len(cases) != seal.TrainingEligible {
		return fmt.Errorf("training screen denominator mismatch")
	}
	slices.SortFunc(cases, func(a, b privateCase) int { return strings.Compare(a.ID, b.ID) })
	policies, repos, e := summarize(cases, builds)
	if e != nil {
		return e
	}
	private, e := json.MarshalIndent(cases, "", "  ")
	if e != nil {
		return e
	}
	private = append(private, '\n')
	r := report{Schema: "riido-path-helper-headroom-result-v1", Status: "training_only_screen_complete_no_model_selection", PlanSHA256: planSHA, InputSealSHA256: sealSHA, RunnerRevision: rev, Input: seal, TrainingCoverage: prepared.Coverage[0], Policies: policies, Repositories: repos, RankingsSHA256: hex.EncodeToString(rankingHash.Sum(nil)), PrivateCasesSHA256: digest(private), Cache: ranker.Stats()}
	public, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	if e = os.Mkdir(cfg.out, 0700); e != nil {
		return fmt.Errorf("new private output creation failed")
	}
	if e = os.WriteFile(filepath.Join(cfg.out, "cases.json"), private, 0600); e != nil {
		return fmt.Errorf("private numeric case write failed")
	}
	if e = os.WriteFile(filepath.Join(cfg.out, "results.json"), append(public, '\n'), 0600); e != nil {
		return fmt.Errorf("aggregate write failed")
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Status string
		Cases  int
	}{r.Status, len(cases)})
}
