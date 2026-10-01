package taskrun

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

func TestOwnedVersionedBehavioralClosure(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace)
	req.TaskID = "repo-keyword-language-guard"
	baseDir, err := filepath.Abs("../taskverify/testdata/repo-keyword-language-guard-v1")
	if err != nil {
		t.Fatal(err)
	}
	req.BaseDir = baseDir
	spec, err := taskverify.TaskSpec(req.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	req.ExpectedSpecSHA256 = digestJSON(spec)
	source, err := os.ReadFile(filepath.Join(baseDir, "pkg/reporouter/router.go"))
	if err != nil {
		t.Fatal(err)
	}
	oldGuard := `if nonLatin(repo.Summary) {`
	newGuard := `unsupported := nonLatin(repo.Summary)
		for _, keyword := range repo.Keywords {
			unsupported = unsupported || nonLatin(keyword)
		}
		if unsupported {`
	if strings.Count(string(source), oldGuard) != 1 {
		t.Fatal("authored behavioral fixture no longer matches its pinned source")
	}
	// This fake CLI only transports an authored candidate and usage fixture.
	// It performs no inference and proves no real CLI sandbox enforcement.
	correctFile := filepath.Join(filepath.Dir(req.PrivateDir), "authored-correct.go")
	if err := os.WriteFile(correctFile, []byte(strings.Replace(string(source), oldGuard, newGuard, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(req.CodexBinary)
	if err != nil {
		t.Fatal(err)
	}
	quotedPath := "'" + strings.ReplaceAll(correctFile, "'", "'\\''") + "'"
	script = []byte(strings.Replace(string(script), completeTrace, "/bin/cp "+quotedPath+" pkg/reporouter/router.go\n"+completeTrace, 1))
	if err := os.WriteFile(req.CodexBinary, script, 0700); err != nil {
		t.Fatal(err)
	}
	req.ExpectedCLIHash = digest(script)
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	r, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Started || r.BaseRevision != spec.BaseRevision || r.TaskSpecSHA256 != req.ExpectedSpecSHA256 || r.VerificationStatus != "accepted" || r.Verification.BaseRevision != spec.BaseRevision || len(r.Verification.AttributionFiles) != 2 || r.CandidateFileCount != 3 || !r.WholeAttemptUsageComplete {
		t.Fatalf("new snapshot or attribution lost across owned execution: %+v", r)
	}
}

// These checks exercise version and provenance boundaries without inference.
func TestVersionedTaskBaseAndAttribution(t *testing.T) {
	for _, tc := range []struct {
		id, fixture, attribution string
		sourceCount, stagedCount int
	}{
		{"repo-keyword-language-guard", "repo-keyword-language-guard-v1", "NOTICE", 3, 5},
		{"taskoutcome-event-key-bounds", "taskoutcome-event-key-bounds-v1", "NOTICE", 3, 5},
		{"go53-humanize-ordinal64", "go53-humanize-ordinal64-v1", "LICENSE", 4, 5},
		{"go53-uuid-canonical-parse", "uuid-canonical-v1", "LICENSE", 24, 25},
	} {
		id := tc.id
		t.Run(id, func(t *testing.T) {
			spec, err := taskverify.TaskSpec(id)
			if err != nil || spec.BaseRevision == taskverify.BaseRevision {
				t.Fatal("new task did not retain its separate public snapshot")
			}
			base, all, err := LoadBase(filepath.Join("../taskverify/testdata", tc.fixture), id)
			if err != nil || len(base) != tc.sourceCount || len(all) != tc.stagedCount {
				t.Fatalf("versioned public closure unavailable: %v", err)
			}
			copyDir := t.TempDir()
			if err := writeFiles(copyDir, all); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(copyDir, tc.attribution), []byte("authored incorrect attribution"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := LoadBase(copyDir, id); err != Error("attribution_pin_mismatch") {
				t.Fatalf("changed attribution accepted: %v", err)
			}
			if _, _, err := LoadBase("../taskverify/testdata/base", id); err == nil {
				t.Fatal("old incomplete closure accepted as a new snapshot")
			}
		})
	}
}

func TestVersionedPlanRejectsWrongSnapshotBeforeInference(t *testing.T) {
	for _, id := range []string{"repo-keyword-language-guard", "taskoutcome-event-key-bounds", "go53-humanize-ordinal64", "go53-uuid-canonical-parse"} {
		t.Run(id, func(t *testing.T) {
			req := fixtureRequest(t, completeTrace)
			req.TaskID = id
			spec, err := taskverify.TaskSpec(req.TaskID)
			if err != nil {
				t.Fatal(err)
			}
			req.ExpectedSpecSHA256 = digestJSON(spec)
			writePlan(t, &req, filepath.Dir(req.PrivateDir))
			if profile, err := validatePlan(req, trustedToolchain{}); err != nil || profile != "fixture-low" {
				t.Fatalf("valid new snapshot refused: %q %v", profile, err)
			}
			bytes, err := os.ReadFile(req.PlanFile)
			if err != nil {
				t.Fatal(err)
			}
			var p plan
			if err := json.Unmarshal(bytes, &p); err != nil {
				t.Fatal(err)
			}
			p.PublicBase = taskverify.BaseRevision
			bytes, err = json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(req.PlanFile, bytes, 0600); err != nil {
				t.Fatal(err)
			}
			req.PlanSHA256 = digest(bytes)
			if _, err := validatePlan(req, trustedToolchain{}); err != Error("plan_task_spec_mismatch") {
				t.Fatalf("old revision accepted for separately versioned task: %v", err)
			}
		})
	}
}
