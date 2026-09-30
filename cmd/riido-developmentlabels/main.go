// riido-developmentlabels prepares targets only; it performs no ranking/training.
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
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

type coverage struct {
	Role, Repository                                                                                                                                                                           string
	Tasks, BlankPatchTasks, SupportedOldPathTasks, NonGitPreambleTasks, OversizedTasks, ParseErrors, UnsupportedTasks, UnsupportedBlocks, NoOldPaths, NewFiles, OldPathReferences, MaxOldPaths int
}
type report struct {
	Schema, PlanSHA256, AmendmentSHA256, MembershipSHA256, ProjectionSHA256                                            string
	Total                                                                                                              coverage
	Repositories                                                                                                       []coverage
	HistoricalLicensesReviewed, IssueTextLicenseResolved, TrainingApproved, CatalogMembershipVerified, ProductionReady bool
}

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	out := flag.String("out", "", "new private output directory")
	projection := flag.String("projection", "", "development-only patch projection")
	const projectionSHA = "6bed55ca053705e59b1362f982b86cbd977fa3d216312a018b1657dfb9fd284b"
	expected := flag.String("sha256", projectionSHA, "expected complete projection SHA256")
	flag.Parse()
	if *out == "" || *projection == "" || *expected != projectionSHA {
		return fmt.Errorf("out, projection and sha256 required")
	}
	plan, e := os.ReadFile("experiments/development-labels/plan-39.json")
	if e != nil {
		return e
	}
	ph := sha256.Sum256(plan)
	const planSHA = "3d144254337f57d3dee1630cdd4786b1baf9a1ba811e40c6050983766b5de9db"
	if hex.EncodeToString(ph[:]) != planSHA {
		return fmt.Errorf("plan mismatch")
	}
	amendment, e := os.ReadFile("experiments/development-labels/oversize-amendment-39.json")
	if e != nil {
		return e
	}
	ah := sha256.Sum256(amendment)
	const amendmentSHA = "fffbaafe84559f81bdfa8dde01eeb1b670dc164ac318dc67ff535bceff5790f6"
	if hex.EncodeToString(ah[:]) != amendmentSHA {
		return fmt.Errorf("amendment mismatch")
	}
	members, e := trainingdata.ReadDevelopment(".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	labels, e := filelabels.ReadDevelopment(*projection, *expected, members)
	if e != nil {
		return e
	}
	reps, e := summarize(members, labels)
	if e != nil {
		return e
	}
	r := report{Schema: "riido-development-label-report-v1", PlanSHA256: planSHA, AmendmentSHA256: amendmentSHA, MembershipSHA256: trainingdata.MembershipSHA256, ProjectionSHA256: *expected, Repositories: reps}
	for _, c := range reps {
		add(&r.Total, c)
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if e = write(filepath.Join(*out, "results.json"), r); e != nil {
		return e
	}
	return write(filepath.Join(*out, "labels.json"), labels)
}
func summarize(members []sweaudit.TaskRole, labels []filelabels.DevelopmentLabel) ([]coverage, error) {
	rows := slices.Clone(members)
	slices.SortFunc(rows, func(a, b sweaudit.TaskRole) int { return strings.Compare(a.Task.ID, b.Task.ID) })
	if len(rows) != len(labels) {
		return nil, fmt.Errorf("label member count mismatch")
	}
	var out []coverage
	for i, m := range rows {
		l := labels[i]
		if m.Task.ID != l.ID || (m.Role != "train" && m.Role != "validation") {
			return nil, fmt.Errorf("label role or ID mismatch")
		}
		c := coverage{Role: m.Role, Repository: m.Task.Repository, Tasks: 1}
		if l.FailureKind == "blank_patch" {
			c.BlankPatchTasks = 1
		}
		if l.FailureKind == "non_git_preamble" {
			c.NonGitPreambleTasks = 1
		}
		if l.Oversized {
			c.OversizedTasks = 1
		}
		if l.ParseError {
			c.ParseErrors = 1
		} else {
			c.UnsupportedBlocks = l.Result.UnsupportedBlocks
			if c.UnsupportedBlocks > 0 {
				c.UnsupportedTasks = 1
			}
			c.NewFiles = l.Result.NewFiles
			c.OldPathReferences = len(l.Result.OldPaths)
			c.MaxOldPaths = c.OldPathReferences
			if c.OldPathReferences > 0 && c.UnsupportedBlocks == 0 {
				c.SupportedOldPathTasks = 1
			}
			if c.OldPathReferences == 0 {
				c.NoOldPaths = 1
			}
		}
		out = append(out, c)
	}
	slices.SortFunc(out, func(a, b coverage) int {
		if a.Role != b.Role {
			return strings.Compare(a.Role, b.Role)
		}
		return strings.Compare(a.Repository, b.Repository)
	})
	n := 0
	for _, c := range out {
		if n == 0 || out[n-1].Role != c.Role || out[n-1].Repository != c.Repository {
			out[n] = c
			n++
		} else {
			add(&out[n-1], c)
		}
	}
	return out[:n], nil
}
func add(a *coverage, b coverage) {
	a.Tasks += b.Tasks
	a.BlankPatchTasks += b.BlankPatchTasks
	a.SupportedOldPathTasks += b.SupportedOldPathTasks
	a.NonGitPreambleTasks += b.NonGitPreambleTasks
	a.OversizedTasks += b.OversizedTasks
	a.ParseErrors += b.ParseErrors
	a.UnsupportedTasks += b.UnsupportedTasks
	a.UnsupportedBlocks += b.UnsupportedBlocks
	a.NoOldPaths += b.NoOldPaths
	a.NewFiles += b.NewFiles
	a.OldPathReferences += b.OldPathReferences
	a.MaxOldPaths = max(a.MaxOldPaths, b.MaxOldPaths)
}
func write(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	return os.WriteFile(p, b, 0600)
}
