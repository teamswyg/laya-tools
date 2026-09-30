// riido-filecatalog collects complete historical path metadata, never source bodies.
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
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const planHash = "e260491f591d927635ce6204c551e027862b4f66aff70ee9d52af2278df4760e"

type observation struct {
	Repository                                                                                                                               string
	Snapshots, Available, Unavailable, RegularFiles, Symlinks, Gitlinks, MaxFiles, MaxFilePathBytes, OverDocumentLimit, OverCatalogByteLimit int
}

func run() error {
	out := flag.String("out", "", "new private output directory")
	offline := flag.Bool("offline", false, "never fetch missing catalogs")
	budget := flag.Int("request-budget", 2000, "maximum new requests, 0..2000; cached catalogs still checked")
	flag.Parse()
	if *out == "" || *budget < 0 || *budget > 2000 {
		return fmt.Errorf("require output and request budget0..2000")
	}
	b, e := os.ReadFile("experiments/file-catalogs/plan-33.json")
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
	if group.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 || len(selected) != 2400 {
		return fmt.Errorf("selection mismatch")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	requests := 0
	var last time.Time
	fetch := func(ctx context.Context, endpoint string) ([]byte, error) {
		if *offline || requests >= *budget {
			return nil, fmt.Errorf("offline or request budget reached")
		}
		delay := 300*time.Millisecond - time.Since(last)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}
		requests++
		last = time.Now()
		return githubmeta.FetchCatalog(ctx, endpoint)
	}
	noFetch := func(context.Context, string) ([]byte, error) { return nil, fmt.Errorf("root cache missing") }
	var observations []observation
	// Diagnostics remain private, including source identities and local errors.
	var failures []struct{ Repository, BaseCommit, Error string }
	available := 0
	for i, s := range selected {
		r := observation{Repository: s.Repository, Snapshots: 1}
		b, e := githubmeta.Root(context.Background(), s.Repository, s.BaseCommit, ".cache/source-preflight-29", noFetch)
		if e == nil {
			var treeID string
			treeID, _, e = sweaudit.TreeEntries(s.Repository, b)
			if e == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				c, err := githubmeta.RecursiveCatalog(ctx, s.Repository, treeID, ".cache/file-catalogs-33", fetch)
				cancel()
				e = err
				if e == nil {
					r.Available = 1
					available++
					for _, entry := range c.Tree {
						switch entry.Type {
						case "commit":
							r.Gitlinks++
						case "blob":
							r.MaxFiles++
							r.MaxFilePathBytes += len(entry.Path)
							if entry.Mode == "120000" {
								r.Symlinks++
							} else {
								r.RegularFiles++
							}
						}
					}
					if r.MaxFiles > hintsearch.MaxDocuments {
						r.OverDocumentLimit = 1
					}
					if r.MaxFilePathBytes > hintsearch.MaxCatalogBytes {
						r.OverCatalogByteLimit = 1
					}
				}
			}
		}
		if e != nil {
			r.Unavailable = 1
			failures = append(failures, struct{ Repository, BaseCommit, Error string }{s.Repository, s.BaseCommit, e.Error()})
		}
		observations = append(observations, r)
		if (i+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "checked=%d/2400 available=%d network_requests=%d\n", i+1, available, requests)
		}
	}
	slices.SortFunc(observations, func(a, b observation) int { return strings.Compare(a.Repository, b.Repository) })
	var repos []observation
	total := observation{Snapshots: 2400}
	for _, r := range observations {
		if len(repos) == 0 || repos[len(repos)-1].Repository != r.Repository {
			repos = append(repos, observation{Repository: r.Repository})
		}
		x := &repos[len(repos)-1]
		x.Snapshots += r.Snapshots
		for _, x := range []*observation{x, &total} {
			x.Available += r.Available
			x.Unavailable += r.Unavailable
			x.RegularFiles += r.RegularFiles
			x.Symlinks += r.Symlinks
			x.Gitlinks += r.Gitlinks
			x.MaxFiles = max(x.MaxFiles, r.MaxFiles)
			x.MaxFilePathBytes = max(x.MaxFilePathBytes, r.MaxFilePathBytes)
			x.OverDocumentLimit += r.OverDocumentLimit
			x.OverCatalogByteLimit += r.OverCatalogByteLimit
		}
	}
	report := struct {
		Schema, PlanSHA256, SelectedSHA256                                                              string
		Total                                                                                           observation
		Repositories                                                                                    []observation
		AllSelectedCatalogsAvailable, SourceBodiesDownloaded, ActualQualityEstablished, ProductionReady bool
	}{"riido-file-catalog-coverage-v1", planHash, group.SelectedSHA256, total, repos, available == 2400, false, false, false}
	b, e = json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600); e != nil {
		return e
	}
	b, e = json.MarshalIndent(failures, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "failures.json"), append(b, '\n'), 0600); e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "completed available=%d/2400 network_requests=%d\n", available, requests)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
