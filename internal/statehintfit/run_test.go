// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfit

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintfamily"
)

func TestSelectionRejectsUnqualifiedAndKeepsPredeclaredTie(t *testing.T) {
	trials := []Trial{
		{Validation: statehintfamily.Report{Eligible: false, SeverityCost: 0, EightNLL: 0}},
		{Validation: statehintfamily.Report{Eligible: true, SeverityCost: .2, EightNLL: 1}},
		{Validation: statehintfamily.Report{Eligible: true, SeverityCost: .2, EightNLL: .5}},
		{Validation: statehintfamily.Report{Eligible: true, SeverityCost: .2, EightNLL: .5}},
	}
	if choose(trials) != 2 {
		t.Fatal("cost/NLL/predeclared arm order")
	}
	trials[1].Validation.SeverityCost = .1
	if choose(trials) != 1 {
		t.Fatal("NLL bypassed higher priority cost")
	}
	for i := range trials {
		trials[i].Validation.Eligible = false
	}
	if choose(trials) != -1 {
		t.Fatal("unqualified arm selected")
	}
}

func TestFinalReaderSealAcrossFailuresAndNonqualification(t *testing.T) {
	yes, no := statehintfamily.Report{Eligible: true}, statehintfamily.Report{}
	for _, pair := range [][2]statehintfamily.Report{{no, yes}, {yes, no}, {no, no}} {
		lock, test := 0, 0
		opened, err := finalStage(pair[0], pair[1], func() error { lock++; return nil }, func() error { test++; return nil })
		if err != nil || opened || lock != 0 || test != 0 {
			t.Fatal("unqualified stage touched final reader/lock")
		}
	}
	called := 0
	opened, err := finalStage(yes, yes, func() error { return ErrStudy }, func() error { called++; return nil })
	if err != ErrStudy || opened || called != 0 {
		t.Fatal("failed lock touched final reader")
	}
	order := []string{}
	opened, err = finalStage(yes, yes, func() error { order = append(order, "durable-readback-lock"); return nil }, func() error { order = append(order, "one-final-reader"); return nil })
	if err != nil || !opened || strings.Join(order, ",") != "durable-readback-lock,one-final-reader" {
		t.Fatal("final reader order")
	}
}

func TestBoundedPinnedReadsAndDurableExclusiveJournal(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := os.WriteFile(filepath.Join(dir, "original.json"), []byte("original fictional record"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readPin(root, Pin{"original.json", strings.Repeat("a", 64)}, 1024); err != ErrStudy {
		t.Fatal("wrong digest accepted")
	}
	if _, err := readPin(root, Pin{"original.json", digest([]byte("original fictional record"))}, 10); err != ErrStudy {
		t.Fatal("oversize read accepted")
	}
	for _, name := range []string{"../outside", "/absolute", "."} {
		if _, err := readBounded(root, name, 1024); err != ErrStudy {
			t.Fatal("unsafe path accepted")
		}
	}
	j := &journal{root: root}
	value := []byte("original development lock\n")
	if j.write("LOCK.before-test.json", value) != nil {
		t.Fatal("durable journal write")
	}
	if j.write("LOCK.before-test.json", value) != ErrStudy {
		t.Fatal("overwritten immutable result")
	}
	got, err := os.ReadFile(filepath.Join(dir, "LOCK.before-test.json"))
	info, statErr := os.Stat(filepath.Join(dir, "LOCK.before-test.json"))
	if err != nil || statErr != nil || !bytes.Equal(got, value) || info.Mode().Perm() != 0600 {
		t.Fatal("private journal bytes/mode")
	}
	j.used = 8 << 20
	if j.write("past-budget.json", []byte("x")) != ErrStudy {
		t.Fatal("journal exceeded budget")
	}
	if _, err := os.Stat(filepath.Join(dir, "past-budget.json")); !os.IsNotExist(err) {
		t.Fatal("budget failure created a file")
	}
}

func TestNoImplicitFitOrPrivateOptionEcho(t *testing.T) {
	for _, args := range [][]string{nil, {"--PRIVATE-option"}, {"--plan", "PRIVATE-argument"}, {"--plan", "PRIVATE-argument", "--fit", "--out", "../PRIVATE-path"}, {"--plan", "PRIVATE-argument", "--check", "--fit"}} {
		var out, errOut bytes.Buffer
		err := Run(args, &out, &errOut)
		if err != ErrStudy || out.Len() != 0 || errOut.Len() != 0 || strings.Contains(err.Error(), "PRIVATE") {
			t.Fatal("implicit fit or private error output")
		}
	}
}

func TestBuildRevisionAndCleanStampRequired(t *testing.T) {
	commit := strings.Repeat("a", 40)
	info := &debug.BuildInfo{GoVersion: "go1.27.1", Path: "github.com/teamswyg/laya-tools/cmd/riido-statehint-tune-v4", Main: debug.Module{Path: "github.com/teamswyg/laya-tools"}, Settings: []debug.BuildSetting{{Key: "vcs", Value: "git"}, {Key: "vcs.revision", Value: commit}, {Key: "vcs.modified", Value: "false"}}}
	if !validBuild(info, commit) {
		t.Fatal("matching build stamp")
	}
	if validBuild(info, strings.Repeat("b", 40)) {
		t.Fatal("unbound source commit")
	}
	info.Settings[2].Value = "true"
	if validBuild(info, commit) {
		t.Fatal("dirty source build accepted")
	}
	if validBuild(nil, commit) {
		t.Fatal("missing stamp accepted")
	}
}

func TestPrivateOutputRejectsInsideWorkspaceSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".cache", "statehint"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "docs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "docs"), filepath.Join(dir, ".cache", "statehint", "v4")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := privateOutput(root, ".cache/statehint/v4/new-run"); err != ErrStudy {
		t.Fatal("inside-root symlink bypassed private anchor")
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "new-run")); !os.IsNotExist(err) {
		t.Fatal("symlink target received artifacts")
	}
	if err := os.Remove(filepath.Join(dir, ".cache", "statehint", "v4")); err != nil {
		t.Fatal(err)
	}
	private, err := privateOutput(root, ".cache/statehint/v4/new-run")
	if err != nil {
		t.Fatal(err)
	}
	private.Close()
	for _, name := range []string{".cache/statehint/v4", ".cache/statehint/v4/new-run"} {
		info, err := root.Lstat(name)
		if err != nil || info.Mode().Perm() != 0700 {
			t.Fatal("private anchor/run mode")
		}
	}
	if _, err := privateOutput(root, ".cache/statehint/v4/new-run"); err != ErrStudy {
		t.Fatal("existing run reused")
	}
}

func TestDeclaredTestInputAliasesRejectedBeforeDataReads(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	p := planFixture()
	names := []string{"PLAN.json", p.Rubric.Path, p.Definitions[0].Path, p.Definitions[1].Path, p.Manifest.Path, p.Audit.Path, p.Parent.Path}
	for _, pin := range p.Partitions {
		names = append(names, pin.Path)
	}
	for _, pin := range p.Sources {
		names = append(names, pin.Path)
	}
	// Original metadata fixtures, never corpus/model/plan payloads.
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if inputIdentities(root, "PLAN.json", p) != nil {
		t.Fatal("distinct inode fixture rejected")
	}
	testName := filepath.Join(dir, p.Partitions[3].Path)
	if err := os.Remove(testName); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(filepath.Join(dir, p.Partitions[0].Path), testName); err != nil {
		t.Fatal(err)
	}
	if inputIdentities(root, "PLAN.json", p) != ErrStudy {
		t.Fatal("hardlinked train/test accepted")
	}
	if err := os.Remove(testName); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, p.Manifest.Path), testName); err != nil {
		t.Fatal(err)
	}
	if inputIdentities(root, "PLAN.json", p) != ErrStudy {
		t.Fatal("symlinked test/manifest accepted")
	}
}
