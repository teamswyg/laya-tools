package taskverify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const maxOutputBytes = 1 << 20

type testOutcome struct {
	passed, isolated, unknown bool
	tests                     int
	code                      string
}

type boundedOutput struct {
	b        bytes.Buffer
	overflow bool
	cancel   context.CancelFunc
}

func (w *boundedOutput) Write(p []byte) (int, error) {
	// Every command shares the same comparable *boundedOutput for Stdout and
	// Stderr. os/exec guarantees at most one Write goroutine in that case; Run
	// waits for its I/O copier before callers inspect the buffer or overflow.
	// This writer has no separate concurrent-Write contract.
	if len(p) > maxOutputBytes-w.b.Len() {
		w.overflow = true
		w.cancel()
		return 0, io.ErrShortWrite
	}
	return w.b.Write(p)
}

func isolatedTests(parent context.Context, files []BaseFile, task string, timeout time.Duration, goRoot string) testOutcome {
	r := testOutcome{code: "independent_tests_failed"}
	definition, versioned := TaskDefinition(task)
	module, err := trustedEvaluationModule(task, files)
	if err != nil {
		r.unknown, r.code = true, err.Error()
		return r
	}
	if !sandboxSupported() {
		r.unknown, r.code = true, "isolation_unavailable"
		return r
	}
	dir, err := os.MkdirTemp("", "riido-task-check-")
	if err != nil {
		r.unknown, r.code = true, "check_directory_unavailable"
		return r
	}
	defer os.RemoveAll(dir)
	// Resolve the host's temporary-directory symlink before sandbox path rules.
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		r.unknown, r.code = true, "check_directory_unavailable"
		return r
	}
	for _, f := range files {
		if f.Path == "go.mod" {
			continue
		}
		p := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			r.unknown, r.code = true, "check_setup_failed"
			return r
		}
		if err = os.WriteFile(p, f.Data, 0600); err != nil {
			r.unknown, r.code = true, "check_setup_failed"
			return r
		}
	}
	// Legacy recipes retain the same minimal offline module. The two opt-in v2
	// recipes use only internally pinned upstream bytes, never candidate config.
	if err = os.WriteFile(filepath.Join(dir, "go.mod"), module, 0600); err != nil {
		r.unknown, r.code = true, "check_setup_failed"
		return r
	}
	if task == "catalog-min-context" {
		if err = os.WriteFile(filepath.Join(dir, "pkg/catalog/taskverify_contract_test.go"), []byte(contractTests), 0600); err != nil {
			r.unknown, r.code = true, "check_setup_failed"
			return r
		}
	}
	if versioned {
		contract, available := definitionContractSource(task)
		if !available || digest([]byte(contract)) != definition.ContractSHA256 {
			r.unknown, r.code = true, "independent_contract_unavailable"
			return r
		}
		if err = os.WriteFile(filepath.Join(dir, filepath.FromSlash(definition.ContractPath)), []byte(contract), 0600); err != nil {
			r.unknown, r.code = true, "check_setup_failed"
			return r
		}
	}
	cache := filepath.Join(dir, "cache")
	work := filepath.Join(dir, "work")
	for _, p := range []string{cache, work} {
		if os.Mkdir(p, 0700) != nil {
			r.unknown, r.code = true, "check_setup_failed"
			return r
		}
	}
	if goRoot == "" {
		goRoot = runtime.GOROOT()
	}
	if goRoot == "" || !filepath.IsAbs(goRoot) || strings.ContainsAny(goRoot, "\x00\r\n") {
		r.unknown, r.code = true, "trusted_go_unavailable"
		return r
	}
	toolchain, err := filepath.EvalSymlinks(goRoot)
	if err != nil {
		r.unknown, r.code = true, "trusted_go_unavailable"
		return r
	}
	goBinary := filepath.Join(toolchain, "bin", "go")
	if st, e := os.Lstat(goBinary); e != nil || !st.Mode().IsRegular() || st.Mode()&0111 == 0 {
		r.unknown, r.code = true, "trusted_go_unavailable"
		return r
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	env := isolatedEnv(toolchain, dir, cache, work)
	// Probe trusted runtime startup independently. An unavailable sandbox/runtime
	// must not be mistaken for an incorrect candidate's behavior.
	probe, err := sandboxCommand(ctx, goBinary, []string{"version"}, toolchain, dir)
	if err != nil {
		r.unknown, r.code = true, "isolation_unavailable"
		return r
	}
	probe.Dir, probe.Env = dir, env
	probeOut := &boundedOutput{cancel: cancel}
	probe.Stdout, probe.Stderr = probeOut, probeOut
	probe.WaitDelay = time.Second
	if probe.Run() != nil || probeOut.overflow || strings.TrimSpace(probeOut.b.String()) != "go version go1.27.1 "+runtime.GOOS+"/"+runtime.GOARCH {
		r.unknown, r.code = true, "isolation_execution_unavailable"
		return r
	}
	args := []string{"test", "-json", "-count=1", "-timeout=10s", "./pkg/catalog"}
	if task == "catalog-min-context" {
		args = append(args, "./pkg/planner")
	}
	if versioned {
		args = []string{"test", "-json", "-count=1", "-timeout=10s"}
		args = append(args, definition.Packages...)
	}
	cmd, err := sandboxCommand(ctx, goBinary, args, toolchain, dir)
	if err != nil {
		r.unknown, r.code = true, "isolation_unavailable"
		return r
	}
	cmd.Dir = dir
	// No inherited authentication, HOME, CODEX_HOME, PATH, GOFLAGS or proxy state.
	// Direct GOROOT execution and offline/local toolchain settings prevent fetching.
	cmd.Env = env
	out := &boundedOutput{cancel: cancel}
	cmd.Stdout, cmd.Stderr = out, out
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	r.isolated = true
	if task == humanizeOrdinalTaskV2 || task == uuidCanonicalTaskV2 {
		actual, readErr := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readErr != nil || !bytes.Equal(actual, module) {
			r.unknown, r.code = true, "evaluation_module_changed"
			return r
		}
	}
	if out.overflow {
		// Resource exhaustion interrupts verification; it is not an independent
		// assertion that the requested behavior is incorrect.
		r.unknown, r.code = true, "test_output_limit"
		return r
	}
	if ctx.Err() != nil {
		r.unknown, r.code = true, "test_timeout_or_cancelled"
		return r
	}
	if err != nil {
		// A sandbox setup failure is not evidence that candidate behavior is wrong.
		if bytes.Contains(out.b.Bytes(), []byte("sandbox-exec:")) || bytes.Contains(out.b.Bytes(), []byte("operation not permitted")) {
			r.unknown, r.code = true, "isolation_execution_unavailable"
		}
		return r
	}
	if versioned {
		return versionedTerminalPasses(definition, out.b.Bytes())
	}
	contractPassed, catalogPassed, plannerPassed := false, false, false
	for _, line := range bytes.Split(out.b.Bytes(), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var event struct {
			Action  string
			Package string
			Test    string
		}
		if json.Unmarshal(line, &event) != nil {
			r.code = "invalid_test_output"
			return r
		}
		if event.Action != "pass" {
			continue
		}
		if event.Test != "" {
			r.tests++
		}
		if event.Package == "github.com/teamswyg/laya-tools/pkg/catalog" {
			if event.Test == "" {
				catalogPassed = true
			} else if event.Test == "TestTaskVerifyMinimumContext" {
				contractPassed = true
			}
		}
		if event.Package == "github.com/teamswyg/laya-tools/pkg/planner" && event.Test == "" {
			plannerPassed = true
		}
	}
	r.passed = catalogPassed && r.tests > 0 && (task != "catalog-min-context" || (contractPassed && plannerPassed))
	if r.passed {
		r.code = "independent_pinned_tests_passed"
	}
	return r
}

func isolatedEnv(toolchain, dir, cache, work string) []string {
	return []string{
		"GOROOT=" + toolchain, "GOCACHE=" + cache, "GOTMPDIR=" + work,
		"GOPATH=" + filepath.Join(dir, "gopath"), "GOTOOLCHAIN=local", "GOWORK=off",
		"GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOMAXPROCS=2",
		"GOENV=off", "GOFLAGS=", "TMPDIR=" + work,
	}
}
