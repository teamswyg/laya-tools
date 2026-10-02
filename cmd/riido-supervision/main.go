// SPDX-License-Identifier: Apache-2.0
// Maintainer metadata inventory only; no source behavior, features or fitting.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `riido-supervision: inspect pinned original supervision metadata

Required flags:
  --input-root PATH     Read-only root containing the pinned public JSON inputs.
  --output PATH         New result file; existing files are never overwritten.
  --plan-sha256 SHA     Exact reviewed67 policy SHA.

This command proposes loss masks while preserving original truth, candidates
and groups. It does not assign roles, extract features or train a model.
`

type envelope struct {
	State   string `json:"state"`
	Failure string `json:"failure_code"`
	Report  Output `json:"report"`
	Ledger  Ledger `json:"ledger"`
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWithConfig(args, stdout, stderr, frozenConfig)
}

// Config injection is private to this package's synthetic tests. Public run
// always uses the original fixed pins/counts and the reviewed policy SHA.
func runWithConfig(args []string, stdout, stderr io.Writer, cfg Config) int {
	fs := flag.NewFlagSet("riido-supervision", flag.ContinueOnError)
	// flag package errors may include supplied values; discard them and return
	// fixed diagnostics. Never echo raw flags, argument values or private paths.
	fs.SetOutput(io.Discard)
	in := fs.String("input-root", "", "read-only repository root")
	out := fs.String("output", "", "new result JSON path")
	plan := fs.String("plan-sha256", "", "exact reviewed67 policy SHA")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprint(stdout, usage)
			return 0
		}
		fmt.Fprintln(stderr, "scope67_invalid_flags")
		return 2
	}
	if fs.NArg() != 0 || *in == "" || *out == "" || *plan != planSHA {
		fmt.Fprintln(stderr, "scope67_explicit_frozen_invocation_required")
		return 2
	}
	f, e := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		fmt.Fprintln(stderr, "scope67_output_reserve_failed")
		return 1
	}
	report, ledger, err := load(*in, cfg)
	result := envelope{State: "passed_metadata_proposed_supervision_only", Report: report, Ledger: ledger}
	if err != nil {
		result.State = "failed"
		result.Failure = err.Error()
	}
	b, e := json.MarshalIndent(result, "", "  ")
	if e == nil {
		_, e = f.Write(append(b, '\n'))
	}
	ce := f.Close()
	if e != nil || ce != nil {
		fmt.Fprintln(stderr, "scope67_output_persist_failed")
		return 1
	}
	if err != nil {
		fmt.Fprintln(stderr, result.Failure)
		return 1
	}
	fmt.Fprintln(stdout, "scope67 metadata inventory persisted; supervision proposed only; training_ready=false")
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
