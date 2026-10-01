package main

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal("repository unavailable")
	}
	return p
}

func sourceFixture(t *testing.T) string {
	t.Helper()
	repo := repoRoot(t)
	dir := t.TempDir()
	paths := append([]string{sourceInputPath}, requiredPaths[:]...)
	for _, p := range paths {
		raw, err := os.ReadFile(filepath.Join(repo, p))
		if err != nil {
			t.Fatal("fixture source unavailable")
		}
		dest := filepath.Join(dir, p)
		if os.MkdirAll(filepath.Dir(dest), 0700) != nil || os.WriteFile(dest, raw, 0600) != nil {
			t.Fatal("fixture copy failed")
		}
	}
	return dir
}

func readPreparation(t *testing.T, path string) preparation {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("preparation output unavailable")
	}
	var r preparation
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("preparation output invalid")
	}
	return r
}

func assertPreparationOnly(t *testing.T, r preparation) {
	t.Helper()
	if r.State != "proposals_only_captions_pending" || r.ProposalRows != 12 || r.CandidateDescriptions != 36 || r.Properties != 3 || r.LiteralDefinitions != 9 || r.CaptionReview != "pending" || len(r.CompiledSources) != 19 {
		t.Fatal("unexpected proposal shape")
	}
	if r.TrainingReady || r.OfficialExecutionFrozen || r.ProtectedFinalRead || r.CandidateEvaluations != 0 || r.TruthLabelsAssigned != 0 || r.GroupsAssigned != 0 || r.RolesAssigned != 0 || r.Fits != 0 || r.ModelCalls != 0 || r.RankingRuns != 0 || r.PerformanceRuns != 0 {
		t.Fatal("preparation granted experimental authority")
	}
	if r.GoVersion != "go1.27.1" || r.CPUThreads != 1 || r.GoHeapSoftLimitBytes != 256<<20 || r.SourceInputSHA256 != sourceInputSHA256 {
		t.Fatal("preparation provenance invalid")
	}
}

func TestPrepareProducesPendingTransportAndPreservesExistingOutput(t *testing.T) {
	dir := sourceFixture(t)
	t.Chdir(dir)
	dest := filepath.Join(dir, "new-output")
	var stdout bytes.Buffer
	oldProcs := runtime.GOMAXPROCS(0)
	oldHeap := debug.SetMemoryLimit(-1)
	if err := run([]string{"--stage", "prepare", "--out", dest}, &stdout); err != nil {
		t.Fatal(err)
	}
	if runtime.GOMAXPROCS(0) != oldProcs || debug.SetMemoryLimit(-1) != oldHeap {
		t.Fatal("preparation changed caller runtime settings")
	}
	r := readPreparation(t, filepath.Join(dest, "preparation.json"))
	assertPreparationOnly(t, r)
	raw, err := os.ReadFile(filepath.Join(dest, "probes.json"))
	if err != nil || len(raw) != r.ProposalsBytes || digest(raw) != r.ProposalsSHA256 {
		t.Fatal("proposal bytes are not bound to summary")
	}
	if bytes.Contains(stdout.Bytes(), []byte(dir)) {
		t.Fatal("private output directory was echoed")
	}
	before := append([]byte(nil), raw...)
	stdout.Reset()
	err = run([]string{"--out", dest}, &stdout)
	if err == nil || err.Error() != "scopeprep_output_directory_not_new" || stdout.Len() != 0 {
		t.Fatal("existing output was accepted")
	}
	after, err := os.ReadFile(filepath.Join(dest, "probes.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("existing proposal changed")
	}
}

func TestSourceBindingRejectsEditsExtraFilesAndEscapesBeforePreparation(t *testing.T) {
	for _, mutation := range []string{"changed_source", "extra_initializer", "directory_bounds", "changed_input", "source_escape"} {
		t.Run(mutation, func(t *testing.T) {
			dir := sourceFixture(t)
			path := filepath.Join(dir, "internal/scopedproperty/spec.go")
			want := "scopeprep_stale_compiled_source"
			switch mutation {
			case "changed_source":
				if os.WriteFile(path, []byte("package scopedproperty\n"), 0600) != nil {
					t.Fatal("mutation failed")
				}
			case "extra_initializer":
				path = filepath.Join(dir, "internal/typedbehavior/unlisted.go")
				if os.WriteFile(path, []byte("package typedbehavior\nfunc init() {}\n"), 0600) != nil {
					t.Fatal("mutation failed")
				}
				want = "scopeprep_unlisted_package_source"
			case "directory_bounds":
				for i := 0; i < 65; i++ {
					name := filepath.Join(dir, "internal/lexicalhint", fmt.Sprintf("bounded-fixture-%02d.txt", i))
					if os.WriteFile(name, nil, 0600) != nil {
						t.Fatal("directory fixture failed")
					}
				}
				want = "scopeprep_source_directory_bounds"
			case "changed_input":
				if os.WriteFile(filepath.Join(dir, sourceInputPath), []byte("{}"), 0600) != nil {
					t.Fatal("mutation failed")
				}
				want = "scopeprep_source_input_changed"
			case "source_escape":
				outside := filepath.Join(t.TempDir(), "outside-source.go")
				raw, err := os.ReadFile(path)
				if err != nil || os.WriteFile(outside, raw, 0600) != nil || os.Remove(path) != nil || os.Symlink(outside, path) != nil {
					t.Fatal("escape fixture failed")
				}
				want = "scopeprep_source_unavailable"
			}
			t.Chdir(dir)
			dest := filepath.Join(dir, "refused-output")
			var out bytes.Buffer
			err := run([]string{"--out", dest}, &out)
			if err == nil || err.Error() != want || out.Len() != 0 {
				t.Fatal("source binding did not fail closed")
			}
			if _, err := os.Stat(dest); !os.IsNotExist(err) {
				t.Fatal("source rejection created output")
			}
		})
	}
}

func TestArgumentsAndHelpDoNotEchoSuppliedText(t *testing.T) {
	const marker = "PRIVATE_FIXTURE_SENTINEL"
	for _, args := range [][]string{{}, {"--stage", "audit", "--out", marker}, {"--unknown=" + marker}, {"--out", marker, marker}} {
		var out bytes.Buffer
		err := run(args, &out)
		if err == nil || strings.Contains(err.Error(), marker) || bytes.Contains(out.Bytes(), []byte(marker)) || out.Len() != 0 {
			t.Fatal("invalid arguments were accepted or echoed")
		}
	}
	var out bytes.Buffer
	if err := run([]string{"--help"}, &out); !errors.Is(err, flag.ErrHelp) || !strings.Contains(out.String(), "proposals only") {
		t.Fatal("static help unavailable")
	}
}

func TestPackagedCGOZeroTrimpathPrepareWithoutGOROOTHint(t *testing.T) {
	repo := repoRoot(t)
	if runtime.Version() != "go1.27.1" {
		t.Fatal("packaged smoke requires exact toolchain")
	}
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	bin := filepath.Join(t.TempDir(), "scopeprep")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, goBinary, "build", "-trimpath", "-buildvcs=false", "-o", bin, "./cmd/riido-scopeprep")
	build.Dir = repo
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOMAXPROCS=1")
	if _, err := build.Output(); err != nil {
		t.Fatal("packaged build failed")
	}
	info, err := buildinfo.ReadFile(bin)
	if err != nil || info.GoVersion != "go1.27.1" {
		t.Fatal("packaged toolchain unavailable")
	}
	cgoZero, trimpath := false, false
	for _, setting := range info.Settings {
		if setting.Key == "CGO_ENABLED" && setting.Value == "0" {
			cgoZero = true
		}
		if setting.Key == "-trimpath" && setting.Value == "true" {
			trimpath = true
		}
	}
	if !cgoZero || !trimpath {
		t.Fatal("packaged build settings do not match smoke contract")
	}
	dest := filepath.Join(t.TempDir(), "proposal-output")
	runCtx, runCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer runCancel()
	cmd := exec.CommandContext(runCtx, bin, "--stage", "prepare", "--out", dest)
	cmd.Dir = repo
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GOROOT=") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOMAXPROCS=1")
	if _, err := cmd.Output(); err != nil {
		t.Fatal("packaged preparation without GOROOT failed")
	}
	assertPreparationOnly(t, readPreparation(t, filepath.Join(dest, "preparation.json")))
}
