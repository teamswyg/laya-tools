package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/pathinput"
)

func configArgs(phase string) []string {
	args := []string{"--phase", phase, "--out", ".cache/owned-run", "--source", "owned-source.json", "--source-sha256", strings.Repeat("1", 64), "--costs", "owned-costs.json", "--costs-sha256", strings.Repeat("2", 64), "--checkpoint", "owned-checkpoint.json", "--checkpoint-sha256", strings.Repeat("3", 64)}
	if phase == "fit" {
		args = append(args, "--seal", "experiments/path-cost-claim/input-46.json", "--seal-sha256", strings.Repeat("4", 64))
	}
	return args
}
func TestPinnedConfigurationAndPrivateOutput(t *testing.T) {
	for _, phase := range []string{"prepare", "fit"} {
		if _, e := parse(configArgs(phase)); e != nil {
			t.Fatal(e)
		}
	}
	for _, p := range []string{"", ".", ".cache", ".cache/", ".cache/../public", ".cache//double", "/tmp/private", "experiments/private"} {
		args := configArgs("prepare")
		args[3] = p
		if _, e := parse(args); e == nil {
			t.Fatal("invalid output allowed", p)
		}
	}
	for _, name := range []string{"--source", "--source-sha256", "--costs", "--costs-sha256", "--checkpoint", "--checkpoint-sha256"} {
		args := configArgs("prepare")
		for i := range args {
			if args[i] == name {
				args[i+1] = ""
			}
		}
		if _, e := parse(args); e == nil {
			t.Fatal("missing pin allowed", name)
		}
	}
	for _, phase := range []string{"final", "release", "train", ""} {
		args := configArgs("prepare")
		args[1] = phase
		if _, e := parse(args); e == nil {
			t.Fatal("unregistered phase", phase)
		}
	}
	args := configArgs("fit")
	for i := range args {
		if args[i] == "--seal" {
			args[i+1] = ".cache/uncommitted-seal.json"
		}
	}
	if _, e := parse(args); e == nil {
		t.Fatal("uncommitted seal path accepted")
	}
	args = configArgs("prepare")
	args = append(args, "--seal", "experiments/path-cost-claim/input-46.json")
	if _, e := parse(args); e == nil {
		t.Fatal("prepare consumed fitting seal")
	}
}

func TestPrivateOutputRejectsSymlinkAndPreservesExistingResults(t *testing.T) {
	t.Chdir(t.TempDir())
	if e := os.Mkdir(".cache", 0700); e != nil {
		t.Fatal(e)
	}
	target := t.TempDir()
	if e := os.Symlink(target, ".cache/redirect"); e != nil {
		t.Fatal(e)
	}
	if e := writeOutput(".cache/redirect/result", report{}, nil); e == nil {
		t.Fatal("symlink published private result")
	}
	if _, e := os.Stat(filepath.Join(target, "result")); !os.IsNotExist(e) {
		t.Fatal("result followed symlink")
	}
	if e := writeOutput(".cache/results", report{Status: "owned-fixture"}, []artifact{{"model-fixture.json", []byte("{}\n")}}); e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{".cache/results", ".cache/results/results.json", ".cache/results/model-fixture.json"} {
		st, e := os.Stat(p)
		if e != nil {
			t.Fatal(e)
		}
		if st.Mode().Perm()&0077 != 0 {
			t.Fatal("public permissions", p)
		}
	}
	before, e := os.ReadFile(".cache/results/results.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = writeOutput(".cache/results", report{Status: "replace"}, nil); e == nil {
		t.Fatal("existing results replaced")
	}
	after, e := os.ReadFile(".cache/results/results.json")
	if e != nil || string(before) != string(after) {
		t.Fatal("existing results changed")
	}
	if e = writeOutput(".cache/invalid", report{}, []artifact{{"../escape", []byte("{}")}}); e == nil {
		t.Fatal("artifact escaped output")
	}
	if _, e = os.Stat(".cache/invalid"); !os.IsNotExist(e) {
		t.Fatal("invalid artifact created output")
	}
}

func TestFitSealMustMatchCommittedBytes(t *testing.T) {
	t.Chdir(t.TempDir())
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git fixture failure: %v %s", e, b)
		}
	}
	git("init", "-q")
	if e := os.MkdirAll("experiments/path-cost-claim", 0700); e != nil {
		t.Fatal(e)
	}
	path := "experiments/path-cost-claim/input-46.json"
	b := []byte("{\"owned\":true}\n")
	if e := os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	file := pathinput.File{Path: path, SHA256: hash(b)}
	if e := committedSeal(file); e == nil {
		t.Fatal("uncommitted input accepted")
	}
	git("add", path)
	git("-c", "user.name=Owned Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "owned fixture")
	if e := committedSeal(file); e != nil {
		t.Fatal(e)
	}
	changed := []byte("{\"owned\":false}\n")
	if e := os.WriteFile(path, changed, 0600); e != nil {
		t.Fatal(e)
	}
	if e := committedSeal(pathinput.File{Path: path, SHA256: hash(changed)}); e == nil {
		t.Fatal("working-tree seal used as committed")
	}
	oversized := []byte(strings.Repeat("x", (64<<10)+1))
	if e := os.WriteFile(path, oversized, 0600); e != nil {
		t.Fatal(e)
	}
	git("add", path)
	git("-c", "user.name=Owned Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "owned oversized fixture")
	if e := committedSeal(pathinput.File{Path: path, SHA256: hash(oversized)}); e == nil {
		t.Fatal("oversized committed input accepted")
	}
}

func TestBoundedCommittedInputDoesNotAllocateOverflow(t *testing.T) {
	b := boundedBuffer{limit: 4}
	if n, e := b.Write([]byte("four")); n != 4 || e != nil {
		t.Fatal("exact limit rejected")
	}
	if n, e := b.Write([]byte("overflow")); n != 0 || e == nil || b.Len() != 4 {
		t.Fatal("overflow buffered before rejection")
	}
}
