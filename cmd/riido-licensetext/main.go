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

	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const planHash = "5704a7da30a2e5367c0ee9de6aa357beba77cfd1e562e0fd0e1c41beef4c24a1"

type inventory struct {
	Repository, Path, GitSHA, SHA256 string
	Bytes, References                int
}
type repoCoverage struct {
	Repository                                                                                                            string
	Snapshots, WithVerifiedTexts, NestedSnapshots, NoCandidates, TreeFailures, BlobFailures, Symlinks, VerifiedReferences int
}

func run() error {
	out := flag.String("out", "", "new private output directory")
	offline := flag.Bool("offline", false, "never fetch uncached metadata or text")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new output directory")
	}
	plan, e := os.ReadFile("experiments/license-texts/plan-31.json")
	if e != nil {
		return e
	}
	p := sha256.Sum256(plan)
	if hex.EncodeToString(p[:]) != planHash {
		return fmt.Errorf("plan hash mismatch")
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
	if group.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 || len(selected) != 2400 {
		return fmt.Errorf("frozen selection mismatch")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	var last time.Time
	fetch := func(ctx context.Context, endpoint string) ([]byte, error) {
		if *offline {
			return nil, fmt.Errorf("offline cache miss")
		}
		delay := 250*time.Millisecond - time.Since(last)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}
		last = time.Now()
		return githubmeta.Fetch(ctx, endpoint)
	}
	noFetch := func(context.Context, string) ([]byte, error) { return nil, fmt.Errorf("root cache missing") }
	var observations []repoCoverage
	var objects []inventory
	for i, s := range selected {
		r := repoCoverage{Repository: s.Repository, Snapshots: 1}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		b, e := githubmeta.Root(ctx, s.Repository, s.BaseCommit, ".cache/source-preflight-29", noFetch)
		cancel()
		if e != nil {
			r.TreeFailures++
			observations = append(observations, r)
			continue
		}
		_, entries, e := sweaudit.TreeEntries(s.Repository, b)
		if e != nil {
			return e
		}
		type candidate struct {
			path  string
			entry sweaudit.TreeEntry
		}
		var candidates []candidate
		for _, entry := range entries {
			if entry.Type == "blob" && sweaudit.LicenseName(entry.Path) {
				candidates = append(candidates, candidate{entry.Path, entry})
			}
		}
		if len(candidates) == 0 {
			r.NestedSnapshots = 1
			for _, entry := range entries {
				directory := entry.Type == "tree" && ((s.Repository == "matplotlib/matplotlib" && entry.Path == "LICENSE") || (s.Repository == "tokio-rs/axum" && (entry.Path == "axum" || strings.HasPrefix(entry.Path, "axum-"))))
				if !directory {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				children, e := githubmeta.Tree(ctx, s.Repository, entry.SHA, ".cache/license-subtrees-31", fetch)
				cancel()
				if e != nil {
					r.TreeFailures++
					continue
				}
				for _, child := range children {
					if child.Type == "blob" && (s.Repository == "matplotlib/matplotlib" || sweaudit.LicenseName(child.Path)) {
						candidates = append(candidates, candidate{entry.Path + "/" + child.Path, child})
					}
				}
			}
		}
		if len(candidates) == 0 {
			r.NoCandidates = 1
		}
		for _, c := range candidates {
			if c.entry.Mode == "120000" {
				r.Symlinks++
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			info, e := githubmeta.Blob(ctx, s.Repository, c.entry.SHA, ".cache/license-blobs-31", fetch)
			cancel()
			if e != nil {
				r.BlobFailures++
				continue
			}
			r.VerifiedReferences++
			objects = append(objects, inventory{s.Repository, c.path, info.GitSHA, info.SHA256, info.Bytes, 1})
		}
		if r.VerifiedReferences > 0 {
			r.WithVerifiedTexts = 1
		}
		observations = append(observations, r)
		if (i+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "checked=%d/2400\n", i+1)
		}
	}
	slices.SortFunc(observations, func(a, b repoCoverage) int { return strings.Compare(a.Repository, b.Repository) })
	var repos []repoCoverage
	for _, r := range observations {
		if len(repos) == 0 || repos[len(repos)-1].Repository != r.Repository {
			repos = append(repos, repoCoverage{Repository: r.Repository})
		}
		a := &repos[len(repos)-1]
		a.Snapshots += r.Snapshots
		a.WithVerifiedTexts += r.WithVerifiedTexts
		a.NestedSnapshots += r.NestedSnapshots
		a.NoCandidates += r.NoCandidates
		a.TreeFailures += r.TreeFailures
		a.BlobFailures += r.BlobFailures
		a.Symlinks += r.Symlinks
		a.VerifiedReferences += r.VerifiedReferences
	}
	slices.SortFunc(objects, func(a, b inventory) int {
		return strings.Compare(a.Repository+"\x00"+a.Path+"\x00"+a.GitSHA, b.Repository+"\x00"+b.Path+"\x00"+b.GitSHA)
	})
	n := 0
	for _, x := range objects {
		if n > 0 && objects[n-1].Repository == x.Repository && objects[n-1].Path == x.Path && objects[n-1].GitSHA == x.GitSHA {
			objects[n-1].References++
		} else {
			objects[n] = x
			n++
		}
	}
	objects = objects[:n]
	report := struct {
		Schema, PlanSHA256, SelectedSHA256                                                                                                                                                    string
		Snapshots, WithVerifiedTexts, NestedSnapshots, NoCandidates, TreeFailures, BlobFailures, Symlinks, VerifiedReferences, UniqueCandidateObjects, DistinctTextObjects, DistinctTextBytes int
		Repositories                                                                                                                                                                          []repoCoverage
		ThirdPartyCoverageComplete, IssueTextLicenseResolved, BlanketTrainingOrRedistributionApproval, ProductionReady                                                                        bool
	}{Schema: "riido-license-text-coverage-v1", PlanSHA256: planHash, SelectedSHA256: group.SelectedSHA256, Repositories: repos, UniqueCandidateObjects: len(objects)}
	for _, r := range repos {
		report.Snapshots += r.Snapshots
		report.WithVerifiedTexts += r.WithVerifiedTexts
		report.NestedSnapshots += r.NestedSnapshots
		report.NoCandidates += r.NoCandidates
		report.TreeFailures += r.TreeFailures
		report.BlobFailures += r.BlobFailures
		report.Symlinks += r.Symlinks
		report.VerifiedReferences += r.VerifiedReferences
	}
	copyObjects := slices.Clone(objects)
	slices.SortFunc(copyObjects, func(a, b inventory) int { return strings.Compare(a.GitSHA, b.GitSHA) })
	previous := ""
	for _, x := range copyObjects {
		if x.GitSHA != previous {
			report.DistinctTextObjects++
			report.DistinctTextBytes += x.Bytes
			previous = x.GitSHA
		}
	}
	for i, obj := range []any{report, objects} {
		b, e := json.MarshalIndent(obj, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, []string{"results.json", "inventory.json"}[i]), append(b, '\n'), 0600); e != nil {
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
