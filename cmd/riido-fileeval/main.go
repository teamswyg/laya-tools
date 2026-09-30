// riido-fileeval compares frozen path retrieval policies; no training or network.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const planHash = "b363e79a1e46123ed1d4fa302ff369b679cb6eb09330f60df0899b1a7a7f821e"

type policy struct {
	Name                                                    string
	Hit1, Hit10, Hit100, AllTargets10, FirstTargetPages     int
	ReciprocalRankSum, Hit1Rate, Hit10Rate, Hit100Rate, MRR float64
}
type aggregate struct {
	Repository                                                                                                                                                                                     string
	Rows, ApplicableCostRows, NoOldLabelRows, LabelParseFailures, UnsupportedLabelBlocks, RankingFailures, AuxiliaryFallbacks, OldLabels, MappedOldLabels, ImprovedPages, WorsenedPages, TiedPages int
	Policies                                                                                                                                                                                       [3]policy
}

func newAggregate(repo string) aggregate {
	r := aggregate{Repository: repo}
	for i, n := range []string{"path_bm25", "identifier_bm25", "baseline_first"} {
		r.Policies[i].Name = n
	}
	return r
}
func add(r *aggregate, scores [3]fileeval.Score) {
	for i, s := range scores {
		p := &r.Policies[i]
		if s.Hit1 {
			p.Hit1++
		}
		if s.Hit10 {
			p.Hit10++
		}
		if s.Hit100 {
			p.Hit100++
		}
		if s.All10 {
			p.AllTargets10++
		}
		p.ReciprocalRankSum += s.ReciprocalRank
		if s.FirstRank > 0 {
			p.FirstTargetPages += (s.FirstRank + 19) / 20
		}
	}
	r.MappedOldLabels += scores[0].Mapped
	if scores[0].FirstRank > 0 {
		r.ApplicableCostRows++
		a, b := (scores[0].FirstRank+19)/20, (scores[2].FirstRank+19)/20
		switch {
		case b < a:
			r.ImprovedPages++
		case b > a:
			r.WorsenedPages++
		default:
			r.TiedPages++
		}
	}
}
func rates(r *aggregate) {
	if r.Rows == 0 {
		return
	}
	for i := range r.Policies {
		p := &r.Policies[i]
		p.Hit1Rate = float64(p.Hit1) / float64(r.Rows)
		p.Hit10Rate = float64(p.Hit10) / float64(r.Rows)
		p.Hit100Rate = float64(p.Hit100) / float64(r.Rows)
		p.MRR = p.ReciprocalRankSum / float64(r.Rows)
	}
}

type prepared struct {
	Selected      sweaudit.Selection
	TreeID, Query string
	Label         filelabels.Label
}

func run() error {
	out := flag.String("out", "", "new aggregate output directory")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output")
	}
	b, e := os.ReadFile("experiments/file-localization/plan-34.json")
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != planHash {
		return fmt.Errorf("plan mismatch")
	}
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	group, selected, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil {
		return e
	}
	if len(selected) != 2400 || group.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 {
		return fmt.Errorf("selection mismatch")
	}
	labelsFull, e := filelabels.Read(".cache/patch-labels-full-33.jsonl", filelabels.FullProjectionSHA256)
	if e != nil {
		return e
	}
	labelsMulti, e := filelabels.Read(".cache/patch-labels-multilingual-33.jsonl", filelabels.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	for _, rows := range [][]sweaudit.Row{full, multi} {
		slices.SortFunc(rows, func(a, b sweaudit.Row) int { return strings.Compare(a.ID, b.ID) })
	}
	noFetch := func(context.Context, string) ([]byte, error) { return nil, fmt.Errorf("offline catalog missing") }
	var tasks []prepared
	// Require every catalog before the first task is ranked. Partial acquisition
	// cannot quietly become a smaller, easier evaluation set.
	for _, s := range selected {
		rows, labels := full, labelsFull
		if s.Source == "multilingual" {
			rows, labels = multi, labelsMulti
		}
		i, ok := slices.BinarySearchFunc(rows, s.ID, func(a sweaudit.Row, b string) int { return strings.Compare(a.ID, b) })
		if !ok {
			return fmt.Errorf("missing query")
		}
		j, ok := slices.BinarySearchFunc(labels, s.ID, func(a filelabels.Label, b string) int { return strings.Compare(a.ID, b) })
		if !ok {
			return fmt.Errorf("missing label row")
		}
		root, e := githubmeta.Root(context.Background(), s.Repository, s.BaseCommit, ".cache/source-preflight-29", noFetch)
		if e != nil {
			return fmt.Errorf("root not ready")
		}
		treeID, _, e := sweaudit.TreeEntries(s.Repository, root)
		if e != nil {
			return e
		}
		if _, e = githubmeta.RecursiveCatalog(context.Background(), s.Repository, treeID, ".cache/file-catalogs-33", noFetch); e != nil {
			return fmt.Errorf("all2400 catalogs must be ready before scoring")
		}
		tasks = append(tasks, prepared{s, treeID, rows[i].Request, labels[j]})
	}
	total := newAggregate("")
	repos := make([]aggregate, len(group.SelectedCounts.Repositories))
	for i, r := range group.SelectedCounts.Repositories {
		repos[i] = newAggregate(r.Repository)
	}
	started := time.Now()
	for n, t := range tasks {
		ri, ok := slices.BinarySearchFunc(repos, t.Selected.Repository, func(a aggregate, b string) int { return strings.Compare(a.Repository, b) })
		if !ok {
			return fmt.Errorf("unknown repository")
		}
		targets := []*aggregate{&total, &repos[ri]}
		for _, r := range targets {
			r.Rows++
			r.OldLabels += len(t.Label.Result.OldPaths)
			r.UnsupportedLabelBlocks += t.Label.Result.UnsupportedBlocks
			if t.Label.ParseError {
				r.LabelParseFailures++
			}
			if len(t.Label.Result.OldPaths) == 0 {
				r.NoOldLabelRows++
			}
		}
		catalog, e := githubmeta.RecursiveCatalog(context.Background(), t.Selected.Repository, t.TreeID, ".cache/file-catalogs-33", noFetch)
		if e != nil {
			return e
		}
		var paths []string
		for _, entry := range catalog.Tree {
			if entry.Type == "blob" {
				paths = append(paths, entry.Path)
			}
		}
		orders, fallback, e := fileeval.Rank(t.Query, paths)
		if e != nil {
			for _, r := range targets {
				r.RankingFailures++
			}
			continue
		}
		if fallback {
			for _, r := range targets {
				r.AuxiliaryFallbacks++
			}
		}
		if !t.Label.ParseError && t.Label.Result.UnsupportedBlocks == 0 {
			var scores [3]fileeval.Score
			for i, o := range orders {
				scores[i], e = fileeval.Measure(paths, o, t.Label.Result.OldPaths)
				if e != nil {
					return e
				}
			}
			for _, r := range targets {
				add(r, scores)
			}
		}
		if (n+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "ranked=%d/2400\n", n+1)
		}
	}
	rates(&total)
	for i := range repos {
		rates(&repos[i])
	}
	gate := total.Rows == 2400 && total.RankingFailures == 0 && total.LabelParseFailures == 0 && total.UnsupportedLabelBlocks == 0 && total.ApplicableCostRows > 0 && total.Policies[2].FirstTargetPages <= total.Policies[0].FirstTargetPages && total.Policies[2].Hit10 >= total.Policies[0].Hit10
	report := struct {
		Schema, PlanSHA256, SelectedSHA256                                            string
		Total                                                                         aggregate
		Repositories                                                                  []aggregate
		AllCatalogsChecked, LexicalGatePassed, LLMSavingsEstablished, ProductionReady bool
	}{"riido-file-localization-v1", planHash, group.SelectedSHA256, total, repos, true, gate, false, false}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	b, e = json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600); e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "index_and_rank_elapsed_ns=%d\n", time.Since(started).Nanoseconds())
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
