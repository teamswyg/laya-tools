package taskrun

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

// All plans and records below are authored public controls. No CLI, model,
// credential source, or platform sandbox is invoked by these budget tests.
type parentBudgetFixture struct {
	req      ParentBudgetRequest
	manifest ParentBudgetManifest
	recipe   taskverify.EvaluationRecipe
}

func newParentBudgetFixture(t *testing.T) parentBudgetFixture {
	t.Helper()
	dir := t.TempDir()
	recipe := taskverify.EvaluationRecipe{Schema: "riido-task-evaluation-recipe-v1", Kind: "pinned_upstream_module_v1", ModulePath: "example.com/public-control", ModuleSHA256: digest([]byte("module example.com/public-control\n")), LanguageVersion: "1.16", ToolchainVersion: "go1.27.1"}
	recipe.RecipeSHA256 = digestJSON(recipe)
	pins := func(child string, ordinal int) ChildAttemptPins {
		return ChildAttemptPins{PlanSHA256: digest([]byte("child-" + child)), AttemptOrdinal: ordinal, TaskID: "public-task-" + child, SpecSHA256: digest([]byte("spec-" + child)), PromptSHA256: digest([]byte("actual-stdin-" + child)), RecipeSHA256: recipe.RecipeSHA256, ProfileID: []string{"profile-sol", "profile-luna"}[ordinal-1], Model: []string{"gpt-6-sol", "gpt-6-luna"}[ordinal-1], Reasoning: "low", CLIHash: digest([]byte("public-cli-control")), GoHash: digest([]byte("public-go-control"))}
	}
	m := ParentBudgetManifest{Schema: ParentBudgetSchema, MaxReservations: 4, Concurrency: 1, NoRefund: true, ChildPlans: []ParentChildPlan{{PlanSHA256: digest([]byte("child-a")), BaseRevision: strings.Repeat("a", 40)}, {PlanSHA256: digest([]byte("child-b")), BaseRevision: strings.Repeat("b", 40)}}, OrderedAttempts: []ParentBudgetAttempt{{GlobalOrdinal: 1, ChildAttemptPins: pins("a", 1)}, {GlobalOrdinal: 2, ChildAttemptPins: pins("b", 1)}, {GlobalOrdinal: 3, ChildAttemptPins: pins("b", 2)}, {GlobalOrdinal: 4, ChildAttemptPins: pins("a", 2)}}}
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(dir, "parent.json")
	if e = os.WriteFile(file, b, 0600); e != nil {
		t.Fatal(e)
	}
	return parentBudgetFixture{req: ParentBudgetRequest{ParentFile: file, ParentSHA256: digest(b), LedgerDir: filepath.Join(dir, "ledger"), GlobalOrdinal: 1}, manifest: m, recipe: recipe}
}

func (f parentBudgetFixture) request(i int) ParentBudgetRequest {
	r := f.req
	r.GlobalOrdinal = i
	return r
}

func (f parentBudgetFixture) record(i int) Record {
	p := f.manifest.OrderedAttempts[i-1].ChildAttemptPins
	base := ""
	for _, child := range f.manifest.ChildPlans {
		if child.PlanSHA256 == p.PlanSHA256 {
			base = child.BaseRevision
		}
	}
	recipe := f.recipe
	return Record{Schema: "riido-owned-task-attempt-v1", PlanSHA256: p.PlanSHA256, AttemptOrdinal: p.AttemptOrdinal, ProfileID: p.ProfileID, TaskID: p.TaskID, TaskSpecSHA256: p.SpecSHA256, PromptSHA256: p.PromptSHA256, BaseRevision: base, ExecutableSHA256: p.CLIHash, GoBinarySHA256: p.GoHash, GoVersion: "go1.27.1", Applied: AppliedRequest{Model: p.Model, Reasoning: p.Reasoning}, Started: true, DurableStart: true, ProcessStatus: "running", GroupCleanup: "pending", AuthCleanup: "pending", ParentSHA256: f.req.ParentSHA256, GlobalOrdinal: i, RecipeSHA256: p.RecipeSHA256, EvaluationRecipe: &recipe}
}

func parentBudgetRecordDir(t *testing.T, start Record, terminal Record) string {
	t.Helper()
	dir := t.TempDir()
	if e := os.Chmod(dir, 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(dir, "codex-home"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := writeRecord(dir, "started-record.json", start); e != nil {
		t.Fatal(e)
	}
	if e := writeRecord(dir, "record.json", terminal); e != nil {
		t.Fatal(e)
	}
	return dir
}

func parentBudgetTerminal(r Record, status string) Record {
	code := 0
	if status != "exited_zero" {
		code = -1
	}
	r.ExitCode = &code
	r.ProcessStatus = status
	r.GroupCleanup = "group_terminated"
	r.AuthCleanup = "removed_or_not_present"
	r.VerificationStatus = "rejected" // Safe failure still consumes and completes.
	return r
}

func TestParentBudgetFourOrderedReservationsNoRefund(t *testing.T) {
	f := newParentBudgetFixture(t)
	for i := 1; i <= 4; i++ {
		l, e := ReserveParentBudget(f.request(i), f.manifest.OrderedAttempts[i-1].ChildAttemptPins)
		if e != nil {
			t.Fatalf("reserve %d: %v", i, e)
		}
		s, e := InspectParentBudget(f.req)
		if e != nil || s.Reservations != i || s.Started != i-1 || s.Completed != i-1 || s.LaunchStateUnknown != 1 || !s.Active {
			t.Fatalf("reservation proof %d: %+v %v", i, s, e)
		}
		start := f.record(i)
		if e = l.MarkStarted(start); e != nil {
			t.Fatal(e)
		}
		status := []string{"exited_zero", "nonzero_exit", "timeout_or_canceled", "capture_write_failed"}[i-1]
		terminal := parentBudgetTerminal(start, status)
		dir := parentBudgetRecordDir(t, start, terminal)
		if e = l.Complete(dir, terminal); e != nil {
			t.Fatalf("complete %d: %v", i, e)
		}
		if e = l.Stop(); e != nil {
			t.Fatal(e)
		}
		s, e = InspectParentBudget(f.req)
		if e != nil || s.Reservations != i || s.Started != i || s.Completed != i || s.LaunchStateUnknown != 0 || s.Active || s.Stopped || s.Exhausted != (i == 4) {
			t.Fatalf("terminal proof %d: %+v %v", i, s, e)
		}
		// Every terminal receipt independently binds the executor's exact bytes.
		var receipt parentTerminal
		b, e := os.ReadFile(filepath.Join(f.req.LedgerDir, parentSlot(i), "terminal.json"))
		if e != nil || json.Unmarshal(b, &receipt) != nil {
			t.Fatal("missing terminal receipt")
		}
		actual, e := os.ReadFile(filepath.Join(dir, "record.json"))
		if e != nil || receipt.RecordSHA256 != digest(actual) {
			t.Fatal("terminal hash mismatch")
		}
		if i < 4 {
			if _, e = ReserveParentBudget(f.request(i), f.manifest.OrderedAttempts[i-1].ChildAttemptPins); e != Error("parent_order_mismatch") {
				t.Fatalf("duplicate: %v", e)
			}
		}
	}
	if _, e := ReserveParentBudget(f.request(5), f.manifest.OrderedAttempts[0].ChildAttemptPins); e != Error("parent_global_limit") {
		t.Fatalf("fifth: %v", e)
	}
	if _, e := ReserveParentBudget(f.request(4), f.manifest.OrderedAttempts[3].ChildAttemptPins); e != Error("parent_global_limit") {
		t.Fatalf("no refund: %v", e)
	}
}

func TestParentBudgetRequestMappingEveryPin(t *testing.T) {
	mutants := []struct {
		name   string
		change func(*ChildAttemptPins)
	}{
		{"child-plan", func(p *ChildAttemptPins) { p.PlanSHA256 = digest([]byte("other-plan")) }},
		{"child-ordinal", func(p *ChildAttemptPins) { p.AttemptOrdinal = 2 }},
		{"task", func(p *ChildAttemptPins) { p.TaskID = "other-public-task" }},
		{"spec", func(p *ChildAttemptPins) { p.SpecSHA256 = digest([]byte("other-spec")) }},
		{"actual-stdin", func(p *ChildAttemptPins) { p.PromptSHA256 = digest([]byte("original-not-actual")) }},
		{"recipe", func(p *ChildAttemptPins) { p.RecipeSHA256 = digest([]byte("other-recipe")) }},
		{"profile", func(p *ChildAttemptPins) { p.ProfileID = "other-profile" }},
		{"model", func(p *ChildAttemptPins) { p.Model = "other-model" }},
		{"reasoning", func(p *ChildAttemptPins) { p.Reasoning = "high" }},
		{"cli", func(p *ChildAttemptPins) { p.CLIHash = digest([]byte("other-cli")) }},
		{"go", func(p *ChildAttemptPins) { p.GoHash = digest([]byte("other-go")) }},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			p := f.manifest.OrderedAttempts[0].ChildAttemptPins
			m.change(&p)
			if _, e := ReserveParentBudget(f.req, p); e != Error("parent_attempt_pin_mismatch") {
				t.Fatalf("accepted mutation: %v", e)
			}
			if _, e := os.Lstat(f.req.LedgerDir); !os.IsNotExist(e) {
				t.Fatal("wrong pins created ledger")
			}
		})
	}
	t.Run("wrong-global-order", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		if _, e := ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e != Error("parent_order_mismatch") {
			t.Fatalf("wrong order: %v", e)
		}
		if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != nil {
			t.Fatal(e)
		}
	})
}

func TestParentBudgetManifestLimitsAndAmbiguity(t *testing.T) {
	mutants := []struct {
		name   string
		change func(*ParentBudgetManifest)
	}{
		{"max-five", func(m *ParentBudgetManifest) { m.MaxReservations = 5 }},
		{"concurrent-two", func(m *ParentBudgetManifest) { m.Concurrency = 2 }},
		{"refund", func(m *ParentBudgetManifest) { m.NoRefund = false }},
		{"same-revision", func(m *ParentBudgetManifest) { m.ChildPlans[1].BaseRevision = m.ChildPlans[0].BaseRevision }},
		{"same-plan", func(m *ParentBudgetManifest) { m.ChildPlans[1].PlanSHA256 = m.ChildPlans[0].PlanSHA256 }},
		{"missing-slot", func(m *ParentBudgetManifest) { m.OrderedAttempts = m.OrderedAttempts[:3] }},
		{"swapped-order", func(m *ParentBudgetManifest) {
			m.OrderedAttempts[0], m.OrderedAttempts[1] = m.OrderedAttempts[1], m.OrderedAttempts[0]
		}},
		{"duplicate-child-ordinal", func(m *ParentBudgetManifest) { m.OrderedAttempts[3].AttemptOrdinal = 1 }},
		{"orphan-child", func(m *ParentBudgetManifest) { m.OrderedAttempts[0].PlanSHA256 = digest([]byte("orphan")) }},
		{"changed-child-prompt", func(m *ParentBudgetManifest) { m.OrderedAttempts[3].PromptSHA256 = digest([]byte("changed")) }},
		{"changed-cli", func(m *ParentBudgetManifest) { m.OrderedAttempts[3].CLIHash = digest([]byte("changed")) }},
		{"private-label", func(m *ParentBudgetManifest) { m.OrderedAttempts[0].ProfileID = "private/path\n" }},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			m.change(&f.manifest)
			b, _ := json.Marshal(f.manifest)
			if os.WriteFile(f.req.ParentFile, b, 0600) != nil {
				t.Fatal("write")
			}
			f.req.ParentSHA256 = digest(b)
			if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e == nil {
				t.Fatal("invalid parent accepted")
			}
		})
	}
	for _, name := range []string{"duplicate-decoded-key", "unknown-key", "trailing-json", "case-folded-alias", "unicode-folded-alias"} {
		t.Run(name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			b, _ := os.ReadFile(f.req.ParentFile)
			switch name {
			case "duplicate-decoded-key":
				b = append([]byte(`{"sch\u0065ma":"conflict",`), b[1:]...)
			case "unknown-key":
				b = append([]byte(`{"unplanned":true,`), b[1:]...)
			case "trailing-json":
				b = append(b, []byte(` {}`)...)
			case "case-folded-alias":
				b = append([]byte(`{"SCHEMA":"`+ParentBudgetSchema+`",`), b[1:]...)
			case "unicode-folded-alias":
				b = append([]byte(`{"ſchema":"`+ParentBudgetSchema+`",`), b[1:]...)
			}
			if os.WriteFile(f.req.ParentFile, b, 0600) != nil {
				t.Fatal("write")
			}
			f.req.ParentSHA256 = digest(b)
			if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e == nil {
				t.Fatal("ambiguous manifest accepted")
			}
		})
	}
	t.Run("raw-hash", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		f.req.ParentSHA256 = digest([]byte("not-parent-bytes"))
		if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != Error("parent_pin_mismatch") {
			t.Fatalf("raw hash: %v", e)
		}
	})
}

func TestParentBudgetConcurrentControllersAndStarts(t *testing.T) {
	f := newParentBudgetFixture(t)
	// Initialize without taking a slot to isolate simultaneous atomic ownership.
	_, raw, e := loadParent(f.req)
	if e != nil {
		t.Fatal(e)
	}
	r, e := openParentLedger(f.req, raw, true)
	if e != nil {
		t.Fatal(e)
	}
	r.Close()
	const count = 24
	gate := make(chan struct{})
	leases := make(chan *ParentBudgetLease, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
			if e != nil {
				errs <- e
			} else {
				leases <- l
			}
		}()
	}
	close(gate)
	wg.Wait()
	close(leases)
	close(errs)
	if len(leases) != 1 || len(errs) != count-1 {
		t.Fatalf("ownership winners=%d rejects=%d", len(leases), len(errs))
	}
	l := <-leases
	startWinners := make(chan bool, count)
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startWinners <- l.MarkStarted(f.record(1)) == nil
		}()
	}
	wg.Wait()
	close(startWinners)
	startCount := 0
	for ok := range startWinners {
		if ok {
			startCount++
		}
	}
	if startCount != 1 {
		t.Fatalf("start marker winners=%d", startCount)
	}
	s, e := InspectParentBudget(f.req)
	if e != nil || s.Reservations != 1 || s.Started != 1 || s.Completed != 0 || !s.Active {
		t.Fatalf("proof: %+v %v", s, e)
	}
	if e = l.Stop(); e != nil {
		t.Fatal(e)
	}
	if _, e := ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e != Error("parent_budget_stopped") {
		t.Fatalf("active lease resumed: %v", e)
	}
}

func TestParentBudgetSafeFailureAndUnknownUsageCanFinish(t *testing.T) {
	for _, status := range []string{"exited_zero", "nonzero_exit", "timeout_or_canceled", "capture_limit", "inherited_pipe_timeout", "capture_write_failed"} {
		t.Run(status, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
			if e != nil {
				t.Fatal(e)
			}
			start := f.record(1)
			if l.MarkStarted(start) != nil {
				t.Fatal("start")
			}
			terminal := parentBudgetTerminal(start, status)
			terminal.WholeAttemptUsageComplete = false
			terminal.SummaryStatus = "unknown"
			terminal.VerificationStatus = "verifier_unknown"
			dir := parentBudgetRecordDir(t, start, terminal)
			if e = l.Complete(dir, terminal); e != nil {
				t.Fatal(e)
			}
			if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e != nil {
				t.Fatalf("safe failed attempt blocked next: %v", e)
			}
		})
	}
}

func TestParentBudgetTerminalFailuresSealAndDoNotRefund(t *testing.T) {
	mutants := []struct {
		name   string
		change func(*Record)
	}{
		{"group-unknown", func(r *Record) { r.GroupCleanup = "cleanup_failed" }},
		{"auth-unknown", func(r *Record) { r.AuthCleanup = "cleanup_failed" }},
		{"still-running", func(r *Record) { r.ProcessStatus = "running" }},
		{"not-started", func(r *Record) { r.Started = false }},
		{"start-not-durable", func(r *Record) { r.DurableStart = false }},
		{"no-exit-proof", func(r *Record) { r.ExitCode = nil }},
		{"zero-status-nonzero-code", func(r *Record) { code := 1; r.ExitCode = &code }},
		{"nonzero-status-zero-code", func(r *Record) { r.ProcessStatus = "nonzero_exit" }},
		{"different-parent", func(r *Record) { r.ParentSHA256 = digest([]byte("other")) }},
		{"different-global", func(r *Record) { r.GlobalOrdinal = 2 }},
		{"different-plan", func(r *Record) { r.PlanSHA256 = digest([]byte("other")) }},
		{"different-child-ordinal", func(r *Record) { r.AttemptOrdinal = 2 }},
		{"different-task", func(r *Record) { r.TaskID = "other" }},
		{"different-spec", func(r *Record) { r.TaskSpecSHA256 = digest([]byte("other")) }},
		{"different-stdin", func(r *Record) { r.PromptSHA256 = digest([]byte("other")) }},
		{"different-recipe-pin", func(r *Record) { r.RecipeSHA256 = digest([]byte("other")) }},
		{"different-profile", func(r *Record) { r.ProfileID = "other" }},
		{"different-model", func(r *Record) { r.Applied.Model = "other" }},
		{"different-reasoning", func(r *Record) { r.Applied.Reasoning = "high" }},
		{"different-cli", func(r *Record) { r.ExecutableSHA256 = digest([]byte("other")) }},
		{"different-go", func(r *Record) { r.GoBinarySHA256 = digest([]byte("other")) }},
		{"different-base", func(r *Record) { r.BaseRevision = strings.Repeat("c", 40) }},
		{"recipe-absent", func(r *Record) { r.EvaluationRecipe = nil }},
		{"recipe-contents-changed", func(r *Record) { p := *r.EvaluationRecipe; p.LanguageVersion = "1.27.1"; r.EvaluationRecipe = &p }},
		{"unknown-status", func(r *Record) { r.ProcessStatus = "invented" }},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
			if e != nil {
				t.Fatal(e)
			}
			start := f.record(1)
			if l.MarkStarted(start) != nil {
				t.Fatal("start")
			}
			terminal := parentBudgetTerminal(start, "exited_zero")
			m.change(&terminal)
			dir := parentBudgetRecordDir(t, start, terminal)
			if e = l.Complete(dir, terminal); e != Error("parent_terminal_proof_failed") {
				t.Fatalf("accepted invalid terminal: %v", e)
			}
			s, e := InspectParentBudget(f.req)
			if e != nil || s.Reservations != 1 || s.Started != 1 || s.Completed != 0 || !s.Active || !s.Stopped {
				t.Fatalf("sealed proof: %+v %v", s, e)
			}
			if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e == nil {
				t.Fatal("failed slot resumed")
			}
		})
	}
}

func TestParentBudgetIndependentFileBindingAndAuthAbsence(t *testing.T) {
	for _, name := range []string{"terminal-byte-mismatch", "terminal-partial", "terminal-symlink", "start-byte-mismatch", "start-missing", "start-symlink", "auth-remains", "home-symlink", "no-start-marker", "modified-ledger-marker"} {
		t.Run(name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
			if e != nil {
				t.Fatal(e)
			}
			start := f.record(1)
			if name != "no-start-marker" && l.MarkStarted(start) != nil {
				t.Fatal("start")
			}
			terminal := parentBudgetTerminal(start, "exited_zero")
			dir := parentBudgetRecordDir(t, start, terminal)
			write := func(path string, b []byte) {
				t.Helper()
				if os.WriteFile(path, b, 0600) != nil {
					t.Fatal("write")
				}
			}
			switch name {
			case "terminal-byte-mismatch":
				p := filepath.Join(dir, "record.json")
				b, _ := os.ReadFile(p)
				write(p, append(b, ' '))
			case "terminal-partial":
				write(filepath.Join(dir, "record.json"), []byte(`{"schema":`))
			case "start-byte-mismatch":
				p := filepath.Join(dir, "started-record.json")
				b, _ := os.ReadFile(p)
				write(p, append(b, ' '))
			case "start-missing":
				if os.Remove(filepath.Join(dir, "started-record.json")) != nil {
					t.Fatal("remove")
				}
			case "terminal-symlink", "start-symlink":
				nameFile := "record.json"
				if name == "start-symlink" {
					nameFile = "started-record.json"
				}
				p := filepath.Join(dir, nameFile)
				b, _ := os.ReadFile(p)
				target := filepath.Join(t.TempDir(), "public-control.json")
				write(target, b)
				if os.Remove(p) != nil || os.Symlink(target, p) != nil {
					t.Fatal("symlink")
				}
			case "auth-remains":
				write(filepath.Join(dir, "codex-home", "auth.json"), []byte(`{"authored_public_control":true}`))
			case "home-symlink":
				p := filepath.Join(dir, "codex-home")
				if os.Remove(p) != nil || os.Symlink(t.TempDir(), p) != nil {
					t.Fatal("symlink")
				}
			case "modified-ledger-marker":
				write(filepath.Join(f.req.LedgerDir, "slot-01", "started.json"), []byte(`{}`))
			}
			if e = l.Complete(dir, terminal); e != Error("parent_terminal_proof_failed") {
				t.Fatalf("invalid proof accepted: %v", e)
			}
			if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e == nil {
				t.Fatal("uncertain slot resumed")
			}
		})
	}
}

func TestParentBudgetCrashPartialLedgerAndNoAutomaticRecovery(t *testing.T) {
	for _, name := range []string{"empty-existing-ledger", "partial-init", "missing-ready", "unknown-active", "partial-reservation", "partial-start", "partial-terminal", "unknown-file", "slot-symlink", "manifest-changed"} {
		t.Run(name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			_, raw, e := loadParent(f.req)
			if e != nil {
				t.Fatal(e)
			}
			if name == "empty-existing-ledger" || name == "partial-init" {
				if os.Mkdir(f.req.LedgerDir, 0700) != nil {
					t.Fatal("mkdir")
				}
				if name == "partial-init" && os.WriteFile(filepath.Join(f.req.LedgerDir, "manifest.json"), raw[:len(raw)/2], 0600) != nil {
					t.Fatal("write")
				}
			} else {
				r, e := openParentLedger(f.req, raw, true)
				if e != nil {
					t.Fatal(e)
				}
				r.Close()
				switch name {
				case "missing-ready":
					e = os.Remove(filepath.Join(f.req.LedgerDir, "ready"))
				case "unknown-active":
					e = os.Mkdir(filepath.Join(f.req.LedgerDir, "active"), 0700)
				case "unknown-file":
					e = os.WriteFile(filepath.Join(f.req.LedgerDir, "unrecognized"), []byte("public control"), 0600)
				case "slot-symlink":
					e = os.Symlink(t.TempDir(), filepath.Join(f.req.LedgerDir, "slot-01"))
				case "manifest-changed":
					e = os.WriteFile(filepath.Join(f.req.LedgerDir, "manifest.json"), append(raw, ' '), 0600)
				default:
					l, err := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
					if err != nil {
						t.Fatal(err)
					}
					if name != "partial-reservation" && l.MarkStarted(f.record(1)) != nil {
						t.Fatal("start")
					}
					file := "reservation.json"
					if name == "partial-start" {
						file = "started.json"
					}
					if name == "partial-terminal" {
						file = "terminal.json"
					}
					e = os.WriteFile(filepath.Join(f.req.LedgerDir, "slot-01", file), []byte(`{"partial":`), 0600)
				}
				if e != nil {
					t.Fatal(e)
				}
			}
			if _, e = ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e == nil {
				t.Fatal("uncertain ledger recovered")
			}
			if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e == nil {
				t.Fatal("uncertain ledger resumed next")
			}
		})
	}
	t.Run("reserved-not-started-is-unknown", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
		if e != nil {
			t.Fatal(e)
		}
		if e = l.Stop(); e != nil {
			t.Fatal(e)
		}
		s, e := InspectParentBudget(f.req)
		if e != nil || s.Reservations != 1 || s.Started != 0 || s.Completed != 0 || s.LaunchStateUnknown != 1 || !s.Stopped || !s.Active {
			t.Fatalf("unknown launch proof: %+v %v", s, e)
		}
	})
}

func TestParentBudgetCompletionAndStopRaceNeverDoubleRelease(t *testing.T) {
	f := newParentBudgetFixture(t)
	l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
	if e != nil {
		t.Fatal(e)
	}
	start := f.record(1)
	if l.MarkStarted(start) != nil {
		t.Fatal("start")
	}
	terminal := parentBudgetTerminal(start, "exited_zero")
	dir := parentBudgetRecordDir(t, start, terminal)
	const count = 16
	gate := make(chan struct{})
	success := make(chan bool, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() { defer wg.Done(); <-gate; success <- l.Complete(dir, terminal) == nil }()
	}
	close(gate)
	wg.Wait()
	close(success)
	winners := 0
	for ok := range success {
		if ok {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("completion winners=%d", winners)
	}
	if e = l.Stop(); e != nil {
		t.Fatal(e)
	}
	s, e := InspectParentBudget(f.req)
	if e != nil || s.Reservations != 1 || s.Completed != 1 || s.Active || s.Stopped {
		t.Fatalf("duplicate release: %+v %v", s, e)
	}
}

func TestParentBudgetUnsafeLedgerAndChangedParentRefuse(t *testing.T) {
	t.Run("ledger-symlink", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		if os.Symlink(t.TempDir(), f.req.LedgerDir) != nil {
			t.Fatal("symlink")
		}
		if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != Error("parent_unsafe_ledger") {
			t.Fatalf("symlink accepted: %v", e)
		}
	})
	t.Run("public-project-ledger", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		if os.WriteFile(filepath.Join(filepath.Dir(f.req.LedgerDir), "AGENTS.md"), []byte("authored public control"), 0600) != nil {
			t.Fatal("write")
		}
		if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != Error("parent_unsafe_ledger") {
			t.Fatalf("public project ledger accepted: %v", e)
		}
	})
	t.Run("other-parent-pins", func(t *testing.T) {
		f := newParentBudgetFixture(t)
		l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
		if e != nil {
			t.Fatal(e)
		}
		if l.Stop() != nil {
			t.Fatal("stop")
		}
		f.manifest.OrderedAttempts[0].ProfileID = "other-profile"
		b, _ := json.Marshal(f.manifest)
		if os.WriteFile(f.req.ParentFile, b, 0600) != nil {
			t.Fatal("write")
		}
		f.req.ParentSHA256 = digest(b)
		if _, e = ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != Error("parent_ledger_pin_mismatch") {
			t.Fatalf("other parent accepted: %v", e)
		}
	})
	// No budget method mutates the caller's immutable attempt JSON, and adding
	// opt-in fields leaves a legacy record's serialization exactly unchanged.
	t.Run("legacy-record-has-no-parent-fields", func(t *testing.T) {
		b, _ := json.Marshal(Record{})
		for _, key := range []string{"parent_sha256", "global_ordinal", "recipe_sha256", "evaluation_recipe", "parentBudgetCompletion"} {
			if strings.Contains(string(b), `"`+key+`"`) {
				t.Fatalf("legacy JSON acquired %s", key)
			}
		}
	})
}

func TestParentBudgetBoundedDecodedJSON(t *testing.T) {
	// These values are decoded through the production parser without any field
	// semantics. Limits apply before reflection/manifest validation can begin.
	for _, test := range []struct {
		name, raw string
		valid     bool
	}{
		{"depth-eight", strings.Repeat("[", 8) + `0` + strings.Repeat("]", 8), true},
		{"depth-nine", strings.Repeat("[", 9) + `0` + strings.Repeat("]", 9), false},
		{"decoded-key-128", `{"` + strings.Repeat(`\u0061`, 128) + `":0}`, true},
		{"decoded-key-129", `{"` + strings.Repeat(`\u0061`, 129) + `":0}`, false},
		{"multibyte-key-128", `{"` + strings.Repeat("가", 42) + `ab":0}`, true},
		{"multibyte-key-129", `{"` + strings.Repeat("가", 43) + `":0}`, false},
		{"decoded-duplicate", `{"a":0,"\u0061":1}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var v any
			ok := parentJSON([]byte(test.raw), &v) == nil
			if ok != test.valid {
				t.Fatalf("valid=%t want=%t", ok, test.valid)
			}
		})
	}
	for _, n := range []int{32, 33} {
		t.Run("object-"+parentSlot(n), func(t *testing.T) {
			var b strings.Builder
			b.WriteByte('{')
			for i := 0; i < n; i++ {
				if i > 0 {
					b.WriteByte(',')
				}
				key, _ := json.Marshal(parentSlot(i))
				b.Write(key)
				b.WriteString(`:0`)
			}
			b.WriteByte('}')
			var v any
			if (parentJSON([]byte(b.String()), &v) == nil) != (n == 32) {
				t.Fatal("object boundary")
			}
		})
		t.Run("array-"+parentSlot(n), func(t *testing.T) {
			b := `[` + strings.TrimSuffix(strings.Repeat(`0,`, n), ",") + `]`
			var v any
			if (parentJSON([]byte(b), &v) == nil) != (n == 32) {
				t.Fatal("array boundary")
			}
		})
	}
}

func TestParentBudgetStartFailuresSealLease(t *testing.T) {
	for _, name := range []string{"wrong-schema", "wrong-model", "wrong-stdin", "wrong-recipe", "not-started", "not-durable", "already-terminal", "exit-present", "group-unknown", "partial-marker"} {
		t.Run(name, func(t *testing.T) {
			f := newParentBudgetFixture(t)
			l, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins)
			if e != nil {
				t.Fatal(e)
			}
			r := f.record(1)
			switch name {
			case "wrong-schema":
				r.Schema = "other"
			case "wrong-model":
				r.Applied.Model = "other"
			case "wrong-stdin":
				r.PromptSHA256 = digest([]byte("other"))
			case "wrong-recipe":
				p := *r.EvaluationRecipe
				p.ModuleSHA256 = digest([]byte("other"))
				r.EvaluationRecipe = &p
			case "not-started":
				r.Started = false
			case "not-durable":
				r.DurableStart = false
			case "already-terminal":
				r.ProcessStatus = "exited_zero"
			case "exit-present":
				code := 0
				r.ExitCode = &code
			case "group-unknown":
				r.GroupCleanup = "unknown"
			case "partial-marker":
				if os.WriteFile(filepath.Join(f.req.LedgerDir, "slot-01", "started.json"), []byte(`{"partial":`), 0600) != nil {
					t.Fatal("write")
				}
			}
			if e = l.MarkStarted(r); e != Error("parent_start_proof_failed") {
				t.Fatalf("invalid start: %v", e)
			}
			if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e == nil {
				t.Fatal("failed start refunded")
			}
			if _, e = os.Stat(filepath.Join(f.req.LedgerDir, "active")); e != nil {
				t.Fatal("failed start released owner")
			}
		})
	}
}

func TestParentBudgetPartialReservationStillCounted(t *testing.T) {
	f := newParentBudgetFixture(t)
	if _, e := ReserveParentBudget(f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins); e != nil {
		t.Fatal(e)
	}
	if os.WriteFile(filepath.Join(f.req.LedgerDir, "slot-01", "reservation.json"), []byte(`{"partial":`), 0600) != nil {
		t.Fatal("write")
	}
	s, e := InspectParentBudget(f.req)
	if e == nil || s.Reservations != 1 || s.Started != 0 || s.Completed != 0 || s.LaunchStateUnknown != 1 || !s.Active {
		t.Fatalf("partial proof: %+v %v", s, e)
	}
}

func TestParentBudgetCrossProcessExclusive(t *testing.T) {
	f := newParentBudgetFixture(t)
	_, raw, e := loadParent(f.req)
	if e != nil {
		t.Fatal(e)
	}
	r, e := openParentLedger(f.req, raw, true)
	if e != nil {
		t.Fatal(e)
	}
	r.Close()
	control := struct {
		Request ParentBudgetRequest
		Pins    ChildAttemptPins
	}{f.req, f.manifest.OrderedAttempts[0].ChildAttemptPins}
	b, _ := json.Marshal(control)
	file := filepath.Join(filepath.Dir(f.req.ParentFile), "public-subprocess-control.json")
	if os.WriteFile(file, b, 0600) != nil {
		t.Fatal("write")
	}
	gate := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			cmd := exec.Command(os.Args[0], "-test.run=^TestParentBudgetCrossProcessHelper$")
			cmd.Env = []string{"RIIDO_PUBLIC_PARENT_BUDGET_HELPER=" + file, "TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir()}
			e := cmd.Run()
			if e == nil {
				results <- 0
				return
			}
			var ex *exec.ExitError
			if errors.As(e, &ex) {
				results <- ex.ExitCode()
			} else {
				results <- 99
			}
		}()
	}
	close(gate)
	wg.Wait()
	close(results)
	winners, rejected := 0, 0
	for code := range results {
		if code == 0 {
			winners++
		} else if code == 10 {
			rejected++
		} else {
			t.Fatalf("helper exit=%d", code)
		}
	}
	if winners != 1 || rejected != 1 {
		t.Fatalf("winners=%d rejected=%d", winners, rejected)
	}
	// The winning fake controller exits immediately without a start marker. Its
	// durable unknown lease cannot be reclaimed by a new process or next slot.
	s, e := InspectParentBudget(f.req)
	if e != nil || s.Reservations != 1 || s.Started != 0 || s.LaunchStateUnknown != 1 || !s.Active {
		t.Fatalf("crashed-owner proof: %+v %v", s, e)
	}
	if _, e = ReserveParentBudget(f.request(2), f.manifest.OrderedAttempts[1].ChildAttemptPins); e == nil {
		t.Fatal("cross-process lease restarted")
	}
}

func TestParentBudgetCrossProcessHelper(t *testing.T) {
	file := os.Getenv("RIIDO_PUBLIC_PARENT_BUDGET_HELPER")
	if file == "" {
		t.Skip("authored control subprocess only")
	}
	b, e := os.ReadFile(file)
	if e != nil {
		os.Exit(11)
	}
	var c struct {
		Request ParentBudgetRequest
		Pins    ChildAttemptPins
	}
	if json.Unmarshal(b, &c) != nil {
		os.Exit(11)
	}
	if _, e = ReserveParentBudget(c.Request, c.Pins); e != nil {
		os.Exit(10)
	}
	os.Exit(0)
}
