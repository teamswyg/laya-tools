package taskrun

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

// The authored shell transports a public candidate and a synthetic usage trace.
// These controls run the owned executor, not Codex inference or real CLI sandbox.
func upstreamParentFixture(t *testing.T, body string) [4]Request {
	t.Helper()
	req := fixtureRequest(t, "cat > received-prompt.txt\n"+body)
	root := filepath.Dir(req.PrivateDir)
	toolchain, err := resolveGoRoot(context.Background(), req.GoRoot)
	if err != nil {
		t.Fatal(err)
	}
	var childRequests [2]Request
	for i, fixture := range []struct{ id, dir string }{
		{"go55-humanize-ordinal64-v2", "go53-humanize-ordinal64-v1"},
		{"go55-uuid-canonical-parse-v2", "uuid-canonical-v1"},
	} {
		child := req
		child.TaskID = fixture.id
		child.GoRoot = toolchain.root
		child.BaseDir, err = filepath.Abs(filepath.Join("../taskverify/testdata", fixture.dir))
		if err != nil {
			t.Fatal(err)
		}
		spec, _ := taskverify.TaskSpec(child.TaskID)
		child.ExpectedSpecSHA256 = digestJSON(spec)
		writePlan(t, &child, root)
		raw, _ := os.ReadFile(child.PlanFile)
		var p plan
		if json.Unmarshal(raw, &p) != nil {
			t.Fatal("authored plan unavailable")
		}
		p.GoToolchain, _ = json.Marshal(struct {
			Version string `json:"version"`
			Hash    string `json:"binary_sha256"`
		}{toolchain.version, toolchain.hash})
		recipe, _ := taskverify.TaskEvaluationRecipe(child.TaskID)
		p.EvaluationRecipes = []planEvaluationRecipe{{TaskID: child.TaskID, PromptSHA256: digest([]byte(spec.Prompt)), RecipeSHA256: recipe.RecipeSHA256, ModuleSHA256: recipe.ModuleSHA256, LanguageVersion: recipe.LanguageVersion}}
		secondProfile := p.Profiles[0]
		secondProfile.ID = "fixture-other"
		secondProfile.Model = "fixture-other-model"
		p.Profiles = append(p.Profiles, secondProfile)
		secondAttempt := p.Attempts[0]
		secondAttempt.Ordinal = 2
		secondAttempt.Profile = secondProfile.ID
		p.Attempts = append(p.Attempts, secondAttempt)
		p.Execution.Max = 2
		raw, _ = json.Marshal(p)
		child.PlanFile = filepath.Join(root, fixture.id+"-plan.json")
		child.PlanSHA256 = digest(raw)
		if err := os.WriteFile(child.PlanFile, raw, 0600); err != nil {
			t.Fatal(err)
		}
		childRequests[i] = child
	}
	manifest := ParentBudgetManifest{Schema: ParentBudgetSchema, MaxReservations: 4, Concurrency: 1, NoRefund: true}
	for _, child := range childRequests {
		spec, _ := taskverify.TaskSpec(child.TaskID)
		manifest.ChildPlans = append(manifest.ChildPlans, ParentChildPlan{PlanSHA256: child.PlanSHA256, BaseRevision: spec.BaseRevision})
	}
	var requests [4]Request
	for i, childIndex := range [4]int{0, 1, 1, 0} {
		child := childRequests[childIndex]
		if i >= 2 {
			child.AttemptOrdinal = 2
			child.Model = "fixture-other-model"
		}
		child.GlobalOrdinal = i + 1
		child.PrivateDir = filepath.Join(root, parentSlot(i+1)+"-attempt")
		child.LedgerDir = filepath.Join(root, "shared-ledger")
		child.ParentFile = filepath.Join(root, "parent.json")
		spec, _ := taskverify.TaskSpec(child.TaskID)
		recipe, _ := taskverify.TaskEvaluationRecipe(child.TaskID)
		profile := "fixture-low"
		if child.AttemptOrdinal == 2 {
			profile = "fixture-other"
		}
		manifest.OrderedAttempts = append(manifest.OrderedAttempts, ParentBudgetAttempt{GlobalOrdinal: i + 1, ChildAttemptPins: ChildAttemptPins{PlanSHA256: child.PlanSHA256, AttemptOrdinal: child.AttemptOrdinal, TaskID: child.TaskID, SpecSHA256: child.ExpectedSpecSHA256, PromptSHA256: digest([]byte(spec.Prompt)), RecipeSHA256: recipe.RecipeSHA256, ProfileID: profile, Model: child.Model, Reasoning: child.Reasoning, CLIHash: child.ExpectedCLIHash, GoHash: toolchain.hash}})
		requests[i] = child
	}
	raw, _ := json.Marshal(manifest)
	if err := os.WriteFile(requests[0].ParentFile, raw, 0600); err != nil {
		t.Fatal(err)
	}
	for i := range requests {
		requests[i].ParentSHA256 = digest(raw)
	}
	return requests
}

func TestUpstreamParentRequiredBeforeAnyLaunch(t *testing.T) {
	for _, id := range []string{"go55-humanize-ordinal64-v2", "go55-uuid-canonical-parse-v2"} {
		req := fixtureRequest(t, completeTrace)
		req.TaskID = id
		if err := validateRequest(req); err != Error("parent_budget_required") {
			t.Fatal("unbounded v2 request accepted", err)
		}
		if record, err := Run(context.Background(), req); err != Error("parent_budget_required") || record.Started {
			t.Fatal("v2 request launched before parent validation")
		}
	}
}

func TestUpstreamPlannedRecipeEveryPin(t *testing.T) {
	spec, _ := taskverify.TaskSpec("go55-humanize-ordinal64-v2")
	recipe, _ := taskverify.TaskEvaluationRecipe(spec.ID)
	tc := trustedToolchain{version: "go version go1.27.1 authored/platform", hash: digest([]byte("authored-go"))}
	p := plan{GoToolchain: json.RawMessage(`{"authored":"required-by-helper"}`), EvaluationRecipes: []planEvaluationRecipe{{TaskID: spec.ID, PromptSHA256: digest([]byte(spec.Prompt)), RecipeSHA256: recipe.RecipeSHA256, ModuleSHA256: recipe.ModuleSHA256, LanguageVersion: recipe.LanguageVersion}}}
	if err := validatePlannedRecipe(spec.ID, spec, p, tc); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct {
		name string
		edit func(*planEvaluationRecipe)
	}{
		{"task", func(r *planEvaluationRecipe) { r.TaskID = "different-task" }},
		{"stdin", func(r *planEvaluationRecipe) { r.PromptSHA256 = digest([]byte("incomplete-prompt")) }},
		{"recipe", func(r *planEvaluationRecipe) { r.RecipeSHA256 = digest([]byte("different-recipe")) }},
		{"module", func(r *planEvaluationRecipe) { r.ModuleSHA256 = digest([]byte("go 1.27.1\n")) }},
		{"language", func(r *planEvaluationRecipe) { r.LanguageVersion = "1.27.1" }},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			copy := p
			copy.EvaluationRecipes = append([]planEvaluationRecipe(nil), p.EvaluationRecipes...)
			mutant.edit(&copy.EvaluationRecipes[0])
			if err := validatePlannedRecipe(spec.ID, spec, copy, tc); err == nil {
				t.Fatal("recipe mutation accepted")
			}
		})
	}
	p.GoToolchain = nil
	if err := validatePlannedRecipe(spec.ID, spec, p, tc); err != Error("planned_evaluation_recipe_required") {
		t.Fatal("unpinned Go accepted", err)
	}
}

const authoredOrdinal64 = `
if [ -f ordinals.go ]; then
cat >> ordinals.go <<'GO'
func Ordinal64(x int64) string {
 magnitude := uint64(x)
 if x < 0 { magnitude = uint64(-(x+1)) + 1 }
 suffix := "th"
 residue := magnitude % 100
 if residue < 11 || residue > 13 {
  switch magnitude % 10 { case 1: suffix="st"; case 2: suffix="nd"; case 3: suffix="rd" }
 }
 return strconv.FormatInt(x,10) + suffix
}
GO
gofmt -w ordinals.go
fi
`

func TestUpstreamParentFourOwnedAuthoredStarts(t *testing.T) {
	requireDarwin(t)
	requests := upstreamParentFixture(t, authoredOrdinal64+completeTrace)
	for i, req := range requests {
		record, err := Run(context.Background(), req)
		if err != nil || !record.Started || !record.DurableStart || record.ParentBudgetCompletion() != "terminal_record_and_budget_verified" {
			t.Fatalf("owned slot %d failed: %v %+v", i+1, err, record)
		}
		spec, _ := taskverify.TaskSpec(req.TaskID)
		prompt, err := os.ReadFile(filepath.Join(req.PrivateDir, "workspace", "received-prompt.txt"))
		if err != nil || !bytes.Equal(prompt, []byte(spec.Prompt)) || digest(prompt) != record.PromptSHA256 {
			t.Fatal("actual full stdin differs from its public pin")
		}
		module, err := os.ReadFile(filepath.Join(req.PrivateDir, "workspace", "go.mod"))
		if err != nil || record.EvaluationRecipe == nil || digest(module) != record.EvaluationRecipe.ModuleSHA256 || record.RecipeSHA256 != spec.EvaluationRecipeSHA256 {
			t.Fatal("actor module/recipe not bound to actual record")
		}
		if (i == 0 || i == 3) && (record.VerificationStatus != "accepted" || record.Verification.EvaluationRecipe == nil || record.Verification.EvaluationRecipe.RecipeSHA256 != record.RecipeSHA256) {
			t.Fatalf("authored positive candidate was not accepted by pinned independent recipe: %+v", record.Verification)
		}
		status, err := InspectParentBudget(parentRequest(req))
		if err != nil || status.Reservations != i+1 || status.Started != i+1 || status.Completed != i+1 || status.Active || status.Stopped {
			t.Fatal("shared controller count differs from owned records", status, err)
		}
		duplicate := req
		duplicate.PrivateDir += "-duplicate"
		if repeated, err := Run(context.Background(), duplicate); err == nil || repeated.Started {
			t.Fatal("duplicate ordered slot started a main command")
		}
	}
	last := requests[3]
	last.GlobalOrdinal = 5
	last.PrivateDir += "-fifth"
	if extra, err := Run(context.Background(), last); err == nil || extra.Started {
		t.Fatal("fifth owned main start accepted")
	}
}

func TestUpstreamParentPreflightFailuresConsumeNoReservation(t *testing.T) {
	requireDarwin(t)
	requests := upstreamParentFixture(t, completeTrace)
	req := requests[0]
	req.ExpectedCLIHash = strings.Repeat("0", 64)
	if record, err := Run(context.Background(), req); err == nil || record.Started {
		t.Fatal("bad pin launched")
	}
	if _, err := os.Stat(req.LedgerDir); !os.IsNotExist(err) {
		t.Fatal("preflight created or consumed the ledger")
	}
	req = requests[0]
	req.LedgerDir = req.PrivateDir
	if record, err := Run(context.Background(), req); err != Error("parent_ledger_must_be_outside_attempt") || record.Started {
		t.Fatal("actor-writable ledger accepted", err)
	}
}

func TestUpstreamParentMissingTerminalStopsBudget(t *testing.T) {
	requireDarwin(t)
	requests := upstreamParentFixture(t, "printf 'authored existing file' > ../record.json\n"+completeTrace)
	record, err := Run(context.Background(), requests[0])
	if err != nil || !record.Started || record.VerificationStatus != "record_write_failed" || record.ParentBudgetCompletion() != "finalization_failed" {
		t.Fatal("unavailable durable terminal was upgraded to a completed budget", err, record)
	}
	status, err := InspectParentBudget(parentRequest(requests[0]))
	if err != nil || status.Reservations != 1 || status.Started != 1 || status.Completed != 0 || !status.Active || !status.Stopped {
		t.Fatal("unknown terminal ledger was refunded or released", status, err)
	}
	if next, err := Run(context.Background(), requests[1]); err == nil || next.Started {
		t.Fatal("stopped parent proceeded to the second child")
	}
}
