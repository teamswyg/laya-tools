package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

type counts struct {
	Repository                                                                                                             string
	Rows, ParseFailures, UnsupportedBlocks, WithOldPaths, NoOldPaths, NewFiles, TotalOldPaths, MaxOldPaths, DuplicatePaths int
}
type selectedLabel struct {
	Source, Repository string
	Label              filelabels.Label
}

func run() error {
	out := flag.String("out", "", "new private output directory; labels.json never public")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output")
	}
	p, e := os.ReadFile("experiments/file-catalogs/label-plan-33.json")
	if e != nil {
		return e
	}
	plan := sha256.Sum256(p)
	if hex.EncodeToString(plan[:]) != "72a9a203bd852bedd10ab9d9e04e5619496866d00e34bad3514ce73b98419dc9" {
		return fmt.Errorf("label plan mismatch")
	}
	a, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	b, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	group, selected, e := sweaudit.GroupEvaluation(a, b)
	if e != nil {
		return e
	}
	if group.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 || len(selected) != 2400 {
		return fmt.Errorf("selection mismatch")
	}
	full, e := filelabels.Read(".cache/patch-labels-full-33.jsonl", filelabels.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := filelabels.Read(".cache/patch-labels-multilingual-33.jsonl", filelabels.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	if len(full) != 2294 || len(multi) != 300 {
		return fmt.Errorf("source label counts")
	}
	var labels []selectedLabel
	var observations []counts
	for _, s := range selected {
		source := full
		if s.Source == "multilingual" {
			source = multi
		}
		i, ok := slices.BinarySearchFunc(source, s.ID, func(a filelabels.Label, b string) int { return strings.Compare(a.ID, b) })
		if !ok {
			return fmt.Errorf("missing selected label")
		}
		l := source[i]
		labels = append(labels, selectedLabel{s.Source, s.Repository, l})
		c := counts{Repository: s.Repository, Rows: 1, UnsupportedBlocks: l.Result.UnsupportedBlocks, NewFiles: l.Result.NewFiles, TotalOldPaths: len(l.Result.OldPaths), MaxOldPaths: len(l.Result.OldPaths), DuplicatePaths: l.Result.DuplicatePaths}
		if l.ParseError {
			c.ParseFailures = 1
		}
		if len(l.Result.OldPaths) > 0 {
			c.WithOldPaths = 1
		} else {
			c.NoOldPaths = 1
		}
		observations = append(observations, c)
	}
	slices.SortFunc(observations, func(a, b counts) int { return strings.Compare(a.Repository, b.Repository) })
	var repos []counts
	var total counts
	for _, c := range observations {
		if len(repos) == 0 || repos[len(repos)-1].Repository != c.Repository {
			repos = append(repos, counts{Repository: c.Repository})
		}
		for _, d := range []*counts{&repos[len(repos)-1], &total} {
			d.Rows += c.Rows
			d.ParseFailures += c.ParseFailures
			d.UnsupportedBlocks += c.UnsupportedBlocks
			d.WithOldPaths += c.WithOldPaths
			d.NoOldPaths += c.NoOldPaths
			d.NewFiles += c.NewFiles
			d.TotalOldPaths += c.TotalOldPaths
			d.MaxOldPaths = max(d.MaxOldPaths, c.MaxOldPaths)
			d.DuplicatePaths += c.DuplicatePaths
		}
	}
	report := struct {
		Schema, PlanSHA256, SelectedSHA256, FullProjectionSHA256, MultilingualProjectionSHA256 string
		Total                                                                                  counts
		Repositories                                                                           []counts
		EvaluationOnly, QualityEstablished                                                     bool
	}{"riido-file-label-coverage-v1", hex.EncodeToString(plan[:]), group.SelectedSHA256, filelabels.FullProjectionSHA256, filelabels.MultilingualProjectionSHA256, total, repos, true, false}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	for i, obj := range []any{report, labels} {
		data, e := json.MarshalIndent(obj, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, []string{"results.json", "labels.json"}[i]), append(data, '\n'), 0600); e != nil {
			return e
		}
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
