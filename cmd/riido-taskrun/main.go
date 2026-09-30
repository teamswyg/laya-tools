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
	request taskrun.Request
	spec    bool
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
	f.DurationVar(&c.request.Timeout, "timeout", taskrun.DefaultTimeout, "one main-attempt wall deadline, at most180s; independent verification has its own45s")
	f.BoolVar(&c.spec, "spec", false, "print the public specification; never launch Codex")
	if f.Parse(args) != nil || f.NArg() != 0 {
		return c, fmt.Errorf("invalid_arguments")
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
		fmt.Fprintln(out, "riido-taskrun --task catalog-min-context --spec\nriido-taskrun --execute --task TASK --model MODEL --reasoning low --base-dir BASE --private-dir NEW --codex-bin BIN --codex-sha256 SHA --codex-version 'codex-cli 0.158.0' --plan-file PLAN --plan-sha256 PLAN_SHA --attempt-ordinal N --task-spec-sha256 SPEC_SHA [--auth-source-dir AUTH] [--timeout 120s]\nOwns one explicit bounded public attempt on supported macOS. Raw traces stay private; no retries, resume, fallback or provider-identity claims. Exit0=accepted and process exited0,1=recorded unsuccessful attempt,2=prelaunch refusal,3=acceptance unavailable. Help/spec launch nothing.")
		return 0
	}
	c, e := parse(args)
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
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
