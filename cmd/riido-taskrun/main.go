// riido-taskrun is an explicit maintainer executor, never the default router.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/teamswyg/laya-tools/internal/taskrun"
	"github.com/teamswyg/laya-tools/internal/taskverify"
)

type config struct {
	request      taskrun.Request
	spec         bool
	budgetStatus bool
}

func parse(args []string) (config, error) {
	var c config
	f := flag.NewFlagSet("riido-taskrun", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.BoolVar(&c.request.Execute, "execute", false, "explicitly own one model attempt; no retries or fallback")
	f.StringVar(&c.request.TaskID, "task", "", "pinned public task ID")
	f.StringVar(&c.request.Model, "model", "", "explicit requested Codex model")
	f.StringVar(&c.request.Reasoning, "reasoning", "", "explicit requested reasoning effort")
	f.StringVar(&c.request.BaseDir, "base-dir", "", "absolute pinned public closure directory")
	f.StringVar(&c.request.PrivateDir, "private-dir", "", "new absolute private directory, created exclusively with mode0700")
	f.StringVar(&c.request.CodexBinary, "codex-bin", "", "absolute regular trusted Codex executable")
	f.StringVar(&c.request.AuthSourceDir, "auth-source-dir", "", "optional absolute source of ChatGPT auth.json; no config or other files copied")
	f.StringVar(&c.request.PlanSHA256, "plan-sha256", "", "SHA-256 of the precommitted evaluation plan")
	f.StringVar(&c.request.PlanFile, "plan-file", "", "absolute local precommitted public plan, verified against plan-sha256")
	f.IntVar(&c.request.AttemptOrdinal, "attempt-ordinal", 0, "explicit ordinal in the plan's ordered attempts")
	f.StringVar(&c.request.ExpectedSpecSHA256, "task-spec-sha256", "", "expected TaskSpec SHA-256, also checked against the plan")
	f.StringVar(&c.request.ExpectedCLIHash, "codex-sha256", "", "expected trusted executable SHA-256")
	f.StringVar(&c.request.ExpectedCLIVersion, "codex-version", "", "exact expected version string; currently codex-cli0.158.0")
	f.StringVar(&c.request.GoRoot, "go-root", "", "optional absolute trusted Go1.27.1 installation; required when packaged trimpath binary has no defaultGOROOT")
	f.StringVar(&c.request.ParentFile, "parent-file", "", "optional pinned parent manifest; required for version2 upstream tasks")
	f.StringVar(&c.request.ParentSHA256, "parent-sha256", "", "SHA-256 of the precommitted parent manifest")
	f.StringVar(&c.request.LedgerDir, "budget-dir", "", "private durable four-slot parent ledger; never reuse for another manifest")
	f.IntVar(&c.request.GlobalOrdinal, "global-ordinal", 0, "one-based ordinal across both child plans")
	f.DurationVar(&c.request.Timeout, "timeout", taskrun.DefaultTimeout, "one main-attempt wall deadline, at most180s; independent verification has its own45s")
	f.BoolVar(&c.spec, "spec", false, "print the public specification; never launch Codex")
	f.BoolVar(&c.budgetStatus, "budget-status", false, "read one existing parent ledger without launching or repairing anything")
	if f.Parse(args) != nil || f.NArg() != 0 {
		return c, fmt.Errorf("invalid_arguments")
	}
	if c.budgetStatus {
		if c.spec || c.request.Execute {
			return c, fmt.Errorf("static_mode_conflict")
		}
		return c, nil
	}
	if _, e := taskverify.TaskSpec(c.request.TaskID); e != nil {
		return c, fmt.Errorf("invalid_task")
	}
	if !c.spec && !c.request.Execute {
		return c, fmt.Errorf("execute_opt_in_required")
	}
	return c, nil
}

func execute(args []string, out, diagnostics io.Writer) int {
	return executeContext(context.Background(), args, out, diagnostics)
}

func executeContext(ctx context.Context, args []string, out, diagnostics io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(out, "riido-taskrun --task catalog-min-context --spec\nriido-taskrun --budget-status --parent-file PARENT --parent-sha256 PARENT_SHA --budget-dir LEDGER\nriido-taskrun --execute --task TASK --model MODEL --reasoning low --base-dir BASE --private-dir NEW --codex-bin BIN --codex-sha256 SHA --codex-version 'codex-cli 0.158.0' --plan-file PLAN --plan-sha256 PLAN_SHA --attempt-ordinal N --task-spec-sha256 SPEC_SHA [--go-root GOROOT] [--auth-source-dir AUTH] [--timeout 120s] [--parent-file PARENT --parent-sha256 PARENT_SHA --budget-dir LEDGER --global-ordinal N]\nOwns one explicit bounded public attempt on supported macOS. Version2 upstream tasks require pinned evaluation recipes and one shared four-slot parent ledger. Reservations are consumed before launch and never refunded; unresolved launches stop the ledger. This bounds owned main starts in that ledger, not host-wide or backend calls. Keep the ledger outside all attempt directories. Packaged trimpath builds need explicit --go-root when no compiled-in Go installation is available. Raw traces stay private; no retries, resume, fallback or provider-identity claims. Exit0=accepted and process exited0,1=recorded unsuccessful attempt,2=prelaunch refusal,3=acceptance or parent finalization unavailable. Help/spec/budget-status launch nothing.")
		return 0
	}
	c, e := parse(args)
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
	}
	if c.budgetStatus {
		s, err := taskrun.InspectParentBudget(taskrun.ParentBudgetRequest{ParentFile: c.request.ParentFile, ParentSHA256: c.request.ParentSHA256, LedgerDir: c.request.LedgerDir})
		if err != nil {
			fmt.Fprintln(diagnostics, err)
			return 2
		}
		if json.NewEncoder(out).Encode(s) != nil {
			fmt.Fprintln(diagnostics, "output_failed")
			return 2
		}
		return 0
	}
	if c.spec {
		s, _ := taskverify.TaskSpec(c.request.TaskID)
		if json.NewEncoder(out).Encode(s) != nil {
			fmt.Fprintln(diagnostics, "output_failed")
			return 2
		}
		return 0
	}
	r, e := taskrun.Run(ctx, c.request)
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
	}
	if json.NewEncoder(out).Encode(r) != nil {
		fmt.Fprintln(diagnostics, "output_failed")
		return 2
	}
	if r.ParentBudgetCompletion() == "finalization_failed" {
		fmt.Fprintln(diagnostics, "parent_budget_finalization_failed")
		return 3
	}
	if r.VerificationStatus == "accepted" && r.ProcessStatus == "exited_zero" && r.AuthCleanup == "removed_or_not_present" {
		return 0
	}
	if r.AuthCleanup == "cleanup_failed" || r.VerificationStatus == "unknown" || r.VerificationStatus == "verifier_unknown" || r.VerificationStatus == "verification_failed" || r.VerificationStatus == "record_write_failed" {
		return 3
	}
	return 1
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(executeContext(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
