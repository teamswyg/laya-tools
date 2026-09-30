// riido-trainingcatalog acquires bounded development metadata, never outcomes.
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
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

const planSHA = "cb5a480bd51f4a7961d89b969cc98f9431d8d83af01e7089362566a29c51d6d1"
const rootCache = ".cache/training-roots-38"
const catalogCache = ".cache/training-catalogs-38"
const treeCache = ".cache/training-subtrees-38"
const blobCache = ".cache/training-license-blobs-38"

type coverage struct {
	Role, Repository                                                                                                                                                                string
	Snapshots, Roots, Catalogs, Files, Symlinks, Gitlinks, MaxFiles, MaxPathBytes                                                                                                   int
	RootFailures, CatalogFailures, LicenseTreeFailures, LicenseBlobFailures, LicenseSymlinks, UnresolvedLicenseDirectories, LicenseReferences, WithLicenseText, NoLicenseCandidates int
}

func add(a *coverage, b coverage) {
	a.Snapshots += b.Snapshots
	a.Roots += b.Roots
	a.Catalogs += b.Catalogs
	a.Files += b.Files
	a.Symlinks += b.Symlinks
	a.Gitlinks += b.Gitlinks
	a.MaxFiles = max(a.MaxFiles, b.MaxFiles)
	a.MaxPathBytes = max(a.MaxPathBytes, b.MaxPathBytes)
	a.RootFailures += b.RootFailures
	a.CatalogFailures += b.CatalogFailures
	a.LicenseTreeFailures += b.LicenseTreeFailures
	a.LicenseBlobFailures += b.LicenseBlobFailures
	a.LicenseSymlinks += b.LicenseSymlinks
	a.UnresolvedLicenseDirectories += b.UnresolvedLicenseDirectories
	a.LicenseReferences += b.LicenseReferences
	a.WithLicenseText += b.WithLicenseText
	a.NoLicenseCandidates += b.NoLicenseCandidates
}

type inventory struct {
	Repository, Path, GitSHA, SHA256 string
	Bytes, References                int
}
type failure struct{ Role, ID, Repository, BaseCommit, Stage string }

func licenseName(s string) bool {
	p := strings.ToUpper(s)
	return sweaudit.LicenseName(s) || p == "LICENCE" || strings.HasPrefix(p, "LICENCE.") || strings.HasPrefix(p, "LICENCE-")
}
func licenseDir(s string) bool {
	switch strings.ToUpper(s) {
	case "LICENSE", "LICENSES", "LICENCE", "LICENCES":
		return true
	}
	return false
}
func writeJSON(path string, x any) error {
	b, e := json.MarshalIndent(x, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func run() error {
	out := flag.String("out", "", "new private result directory")
	mode := flag.String("mode", "sample", "sample or all development tasks")
	offline := flag.Bool("offline", false, "never fetch an uncached object")
	budget := flag.Int("request-budget", 256, "maximum new requests,1..2000")
	flag.Parse()
	if *out == "" || (*mode != "sample" && *mode != "all") || *budget < 1 || *budget > 2000 {
		return fmt.Errorf("invalid output, mode or request budget")
	}
	plan, e := os.ReadFile("experiments/training-catalogs/plan-38.json")
	if e != nil {
		return e
	}
	ph := sha256.Sum256(plan)
	if hex.EncodeToString(ph[:]) != planSHA {
		return fmt.Errorf("plan changed")
	}
	rows, e := trainingdata.ReadDevelopment(".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	if len(rows) != 13021 {
		return fmt.Errorf("development membership changed")
	}
	if *mode == "sample" {
		rows, e = trainingdata.Samples(rows)
		if e != nil || len(rows) != 25 {
			return fmt.Errorf("sample membership changed")
		}
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	var last time.Time
	calls := 0
	fetch := func(ctx context.Context, endpoint string) ([]byte, error) {
		if *offline || calls >= *budget {
			return nil, fmt.Errorf("offline or request budget exhausted")
		}
		if wait := 300*time.Millisecond - time.Since(last); wait > 0 {
			time.Sleep(wait)
		}
		last = time.Now()
		calls++
		child, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if strings.HasSuffix(endpoint, "?recursive=1") {
			return githubmeta.FetchCatalog(child, endpoint)
		}
		return githubmeta.Fetch(child, endpoint)
	}
	var observations []coverage
	var objects []inventory
	var failures []failure
	failed := func(r sweaudit.TaskRole, stage string) {
		failures = append(failures, failure{r.Role, r.Task.ID, r.Task.Repository, r.Task.BaseCommit, stage})
	}
	for n, row := range rows {
		if row.Role != "train" && row.Role != "validation" {
			return fmt.Errorf("final collection forbidden")
		}
		s := row.Task
		c := coverage{Role: row.Role, Repository: s.Repository, Snapshots: 1}
		b, e := githubmeta.Root(context.Background(), s.Repository, s.BaseCommit, rootCache, fetch)
		if e != nil {
			c.RootFailures++
			c.CatalogFailures++
			failed(row, "root")
			observations = append(observations, c)
			continue
		}
		c.Roots++
		treeID, entries, e := sweaudit.TreeEntries(s.Repository, b)
		if e != nil {
			return e
		}
		catalog, e := githubmeta.CompleteCatalog(context.Background(), s.Repository, treeID, catalogCache, treeCache, fetch)
		if e != nil {
			c.CatalogFailures++
			failed(row, "catalog")
		} else {
			c.Catalogs++
			paths := 0
			for _, entry := range catalog.Tree {
				if entry.Type == "blob" {
					paths += len(entry.Path)
					c.Files++
					if entry.Mode == "120000" {
						c.Symlinks++
					}
				}
				if entry.Type == "commit" {
					c.Gitlinks++
				}
			}
			c.MaxFiles = c.Files
			c.MaxPathBytes = paths
		}
		type candidate struct {
			path  string
			entry sweaudit.TreeEntry
		}
		var candidates []candidate
		for _, entry := range entries {
			if entry.Type == "blob" && licenseName(entry.Path) {
				candidates = append(candidates, candidate{entry.Path, entry})
			}
			if entry.Type == "tree" && licenseDir(entry.Path) {
				children, e := githubmeta.Tree(context.Background(), s.Repository, entry.SHA, treeCache, fetch)
				if e != nil {
					c.LicenseTreeFailures++
					failed(row, "license-tree")
					continue
				}
				for _, child := range children {
					if child.Type == "tree" {
						c.UnresolvedLicenseDirectories++
					}
					if child.Type == "blob" {
						candidates = append(candidates, candidate{entry.Path + "/" + child.Path, child})
					}
				}
			}
		}
		if len(candidates) == 0 {
			c.NoLicenseCandidates++
		}
		for _, x := range candidates {
			if x.entry.Mode == "120000" {
				c.LicenseSymlinks++
				continue
			}
			blob, e := githubmeta.Blob(context.Background(), s.Repository, x.entry.SHA, blobCache, fetch)
			if e != nil {
				c.LicenseBlobFailures++
				failed(row, "license-blob")
				continue
			}
			c.LicenseReferences++
			objects = append(objects, inventory{s.Repository, x.path, blob.GitSHA, blob.SHA256, blob.Bytes, 1})
		}
		if c.LicenseReferences > 0 {
			c.WithLicenseText++
		}
		observations = append(observations, c)
		if *mode == "sample" || (n+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "visited=%d/%d requests=%d\n", n+1, len(rows), calls)
		}
	}
	slices.SortFunc(observations, func(a, b coverage) int {
		if a.Role != b.Role {
			return strings.Compare(a.Role, b.Role)
		}
		return strings.Compare(a.Repository, b.Repository)
	})
	var repositories []coverage
	var total coverage
	for _, c := range observations {
		add(&total, c)
		if len(repositories) == 0 || repositories[len(repositories)-1].Role != c.Role || repositories[len(repositories)-1].Repository != c.Repository {
			repositories = append(repositories, coverage{Role: c.Role, Repository: c.Repository})
		}
		add(&repositories[len(repositories)-1], c)
	}
	slices.SortFunc(objects, func(a, b inventory) int {
		return strings.Compare(a.Repository+"\x00"+a.Path+"\x00"+a.GitSHA, b.Repository+"\x00"+b.Path+"\x00"+b.GitSHA)
	})
	compact := objects[:0]
	for _, x := range objects {
		if len(compact) > 0 && compact[len(compact)-1].Repository == x.Repository && compact[len(compact)-1].Path == x.Path && compact[len(compact)-1].GitSHA == x.GitSHA {
			compact[len(compact)-1].References++
		} else {
			compact = append(compact, x)
		}
	}
	objects = compact
	report := struct {
		Schema, PlanSHA256, MembershipFileSHA256, PartitionSHA256, Mode                                                          string
		Total                                                                                                                    coverage
		Repositories                                                                                                             []coverage
		DevelopmentTasks, ExcludedFinalTasks, UniqueLicenseReferences                                                            int
		AllDevelopmentCatalogsAvailable, HistoricalLicensesReviewed, IssueTextLicenseResolved, TrainingApproved, ProductionReady bool
	}{"riido-training-catalog-v1", planSHA, trainingdata.MembershipSHA256, trainingdata.PartitionSHA256, *mode, total, repositories, 13021, 2402, len(objects), *mode == "all" && total.Catalogs == 13021, false, false, false, false}
	for i, x := range []any{report, objects, failures} {
		if e = writeJSON(filepath.Join(*out, []string{"results.json", "inventory.json", "failures.json"}[i]), x); e != nil {
			return e
		}
	}
	fmt.Fprintf(os.Stderr, "complete mode=%s snapshots=%d catalogs=%d verified_license_snapshots=%d requests=%d\n", *mode, total.Snapshots, total.Catalogs, total.WithLicenseText, calls)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
