// riido-taskverify verifies public task behavior without invoking a model.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

type config struct {
	task, base, candidate, goRoot string
	spec                          bool
	timeout                       time.Duration
}

func parse(args []string) (config, error) {
	var c config
	f := flag.NewFlagSet("riido-taskverify", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.StringVar(&c.task, "task", "", "public task ID; use --spec to inspect its exact acceptance contract")
	f.StringVar(&c.base, "base-dir", "", "private local copy of the pinned public base closure")
	f.StringVar(&c.candidate, "candidate-dir", "", "candidate checkout; only task-closure files are assessed")
	f.StringVar(&c.goRoot, "go-root", "", "optional absolute trusted Go1.27.1 installation for packaged trimpath behavioral checks")
	f.BoolVar(&c.spec, "spec", false, "print the public task specification without executing checks")
	f.DurationVar(&c.timeout, "timeout", taskverify.DefaultTimeout, "isolated verification timeout, at most60s")
	if e := f.Parse(args); e != nil {
		return c, fmt.Errorf("invalid_arguments")
	}
	if _, e := taskverify.TaskSpec(c.task); e != nil || f.NArg() != 0 || c.timeout <= 0 || c.timeout > taskverify.MaxTimeout {
		return c, fmt.Errorf("invalid_task_or_limits")
	}
	if !c.spec && (strings.TrimSpace(c.base) == "" || strings.TrimSpace(c.candidate) == "") {
		return c, fmt.Errorf("require_base_and_candidate")
	}
	return c, nil
}
func readBase(dir, task string) ([]taskverify.BaseFile, error) {
	paths, e := taskverify.BasePaths(task)
	if e != nil {
		return nil, e
	}
	r, e := os.OpenRoot(dir)
	if e != nil {
		return nil, fmt.Errorf("base_root_unavailable")
	}
	defer r.Close()
	files := make([]taskverify.BaseFile, 0, len(paths))
	total := 0
	for _, p := range paths {
		parts := strings.Split(p, "/")
		for i := range parts {
			info, e := r.Lstat(strings.Join(parts[:i+1], "/"))
			if e != nil || info.Mode()&os.ModeSymlink != 0 || (i+1 < len(parts) && !info.IsDir()) {
				return nil, fmt.Errorf("base_file_unavailable")
			}
		}
		f, e := r.Open(p)
		if e != nil {
			return nil, fmt.Errorf("base_file_unavailable")
		}
		info, e := f.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > taskverify.MaxFileBytes {
			f.Close()
			return nil, fmt.Errorf("base_file_limit")
		}
		b, e := io.ReadAll(io.LimitReader(f, taskverify.MaxFileBytes+1))
		f.Close()
		if e != nil || len(b) > taskverify.MaxFileBytes {
			return nil, fmt.Errorf("base_file_limit")
		}
		total += len(b)
		if total > taskverify.MaxTotalBytes {
			return nil, fmt.Errorf("base_total_limit")
		}
		h := sha256.Sum256(b)
		files = append(files, taskverify.BaseFile{Path: p, SHA256: hex.EncodeToString(h[:]), Data: b})
	}
	return files, nil
}
func execute(args []string, out, diagnostics io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(out, "riido-taskverify --task catalog-min-context --spec\nriido-taskverify --task catalog-min-context --base-dir BASE --candidate-dir CANDIDATE [--go-root GOROOT]\nVerifies a bounded public task closure; no model is launched. Packaged trimpath builds require an explicit trusted Go1.27.1 installation for behavioral checks when no default root is available. Exit0=accepted,1=rejected,2=invalid,3=verification unavailable. Candidate-owned tests are not acceptance evidence.")
		return 0
	}
	c, e := parse(args)
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
	}
	if c.spec {
		s, _ := taskverify.TaskSpec(c.task)
		if json.NewEncoder(out).Encode(s) != nil {
			fmt.Fprintln(diagnostics, "output_failed")
			return 2
		}
		return 0
	}
	base, e := readBase(c.base, c.task)
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
	}
	r, e := taskverify.Verify(context.Background(), taskverify.Request{TaskID: c.task, BaseRevision: taskverify.BaseRevision, BaseFiles: base, CandidateDir: c.candidate, Timeout: c.timeout, GoRoot: c.goRoot})
	if e != nil {
		fmt.Fprintln(diagnostics, e)
		return 2
	}
	if json.NewEncoder(out).Encode(r) != nil {
		fmt.Fprintln(diagnostics, "output_failed")
		return 2
	}
	if r.Accepted {
		return 0
	}
	if r.Status == "verifier_unknown" {
		return 3
	}
	return 1
}
func main() { os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr)) }
