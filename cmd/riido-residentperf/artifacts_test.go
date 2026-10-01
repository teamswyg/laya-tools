package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func residentRepoRoot(t *testing.T) string {
	t.Helper()
	wd, e := os.Getwd()
	if e != nil {
		t.Fatal("test repository unavailable")
	}
	return filepath.Clean(filepath.Join(wd, "../.."))
}

func frozenSourceFixture(t *testing.T) (string, Plan) {
	t.Helper()
	repo, fixture := residentRepoRoot(t), t.TempDir()
	for _, path := range manifestPaths() {
		raw, e := os.ReadFile(filepath.Join(repo, path))
		if e != nil {
			t.Fatal("test frozen source unavailable")
		}
		dest := filepath.Join(fixture, path)
		if os.MkdirAll(filepath.Dir(dest), 0700) != nil || os.WriteFile(dest, raw, 0600) != nil {
			t.Fatal("test fixture creation failed")
		}
	}
	t.Chdir(fixture)
	p := policyPlan()
	pins, e := currentManifest()
	if e != nil {
		t.Fatal("test manifest unavailable")
	}
	p.ImplementationFiles = pins
	recipe, e := recipeBytes()
	if e != nil {
		t.Fatal("test build recipe unavailable")
	}
	if os.MkdirAll(filepath.Dir(recipePath), 0700) != nil || os.WriteFile(recipePath, recipe, 0600) != nil {
		t.Fatal("test recipe creation failed")
	}
	p.BuildRecipeSHA256 = hashBytes(recipe)
	return fixture, p
}

func TestBoundSourcesRejectDriftExtraImplementationAndModule(t *testing.T) {
	for _, mode := range [6]string{"valid", "source-drift", "unlisted-init", "manifest-order", "module-drift", "recipe-drift"} {
		t.Run(mode, func(t *testing.T) {
			_, p := frozenSourceFixture(t)
			switch mode {
			case "source-drift":
				if os.WriteFile("cmd/riido-residentperf/main.go", []byte("changed"), 0600) != nil {
					t.Fatal("test drift failed")
				}
			case "unlisted-init":
				if os.WriteFile("cmd/riido-shortclaim/extra.go", []byte("package main\nfunc init(){}\n"), 0600) != nil {
					t.Fatal("test drift failed")
				}
			case "manifest-order":
				p.ImplementationFiles[0], p.ImplementationFiles[1] = p.ImplementationFiles[1], p.ImplementationFiles[0]
			case "module-drift":
				if os.WriteFile("go.mod", []byte("different module"), 0600) != nil {
					t.Fatal("test drift failed")
				}
			case "recipe-drift":
				if os.WriteFile(recipePath, []byte("{}\n"), 0600) != nil {
					t.Fatal("test drift failed")
				}
			}
			e := verifySources(&p)
			if (e == nil) != (mode == "valid") {
				t.Fatal("source acceptance differs from frozen manifest")
			}
		})
	}
}

func TestSourceBoundsAndRootEscape(t *testing.T) {
	t.Run("directory-bound", func(t *testing.T) {
		_, _ = frozenSourceFixture(t)
		for i := 0; i < 65; i++ {
			if os.WriteFile(fmt.Sprintf("cmd/riido-shortclaim/entry-%02d", i), nil, 0600) != nil {
				t.Fatal("test entries failed")
			}
		}
		if verifySources(nil) == nil {
			t.Fatal("unbounded directory accepted")
		}
	})
	t.Run("outside-symlink", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "private")
		if os.WriteFile(outside, []byte("private"), 0600) != nil {
			t.Fatal("test outside file failed")
		}
		rootDir := t.TempDir()
		if os.Symlink(outside, filepath.Join(rootDir, "escape")) != nil {
			t.Fatal("test symlink failed")
		}
		root, e := os.OpenRoot(rootDir)
		if e != nil {
			t.Fatal("test root unavailable")
		}
		defer root.Close()
		if _, e := rootRead(root, "escape", 32); e == nil || strings.Contains(e.Error(), outside) {
			t.Fatal("root escape accepted or disclosed")
		}
	})
}

func TestCompiledManifestAndPublicRecipe(t *testing.T) {
	pins, e := compiledPins()
	if e != nil || len(pins) != 22 || len(manifestPaths()) != 31 || len(runtimePaths()) != 24 {
		t.Fatal("compiled closure differs")
	}
	for i, p := range pins {
		if !slices.Contains(runtimePaths(), p.Path) || !validSHA(p.SHA256) || (i > 0 && pins[i-1].Path >= p.Path) {
			t.Fatal("compiled pin invalid")
		}
	}
	raw, e := recipeBytes()
	if e != nil {
		t.Fatal("build recipe unavailable")
	}
	if bytes.Contains(raw, []byte(os.TempDir())) || bytes.Contains(raw, []byte("/Users/")) {
		t.Fatal("recipe disclosed private output path")
	}
	var recipe struct {
		Environment []string `json:"environment"`
		Driver      []string `json:"driver_command"`
		Child       []string `json:"child_command"`
	}
	if json.Unmarshal(raw, &recipe) != nil || len(recipe.Environment) != 11 || len(recipe.Driver) != 8 || len(recipe.Child) != 8 || recipe.Driver[4] != "-p=1" || recipe.Driver[3] != "-buildvcs=false" || recipe.Environment[0] != "CGO_ENABLED=0" {
		t.Fatal("build recipe policy differs")
	}
	for _, v := range []string{"", "main", strings.Repeat("A", 40), strings.Repeat("0", 39), strings.Repeat("0", 41)} {
		if validCommit(v) {
			t.Fatal("mutable source revision accepted")
		}
	}
	if !validCommit(strings.Repeat("a", 40)) {
		t.Fatal("immutable source revision refused")
	}
}

// Packaged build controls exercise real buildinfo, not an invented fixture.
// They are not the frozen 24-row resource replay and execute no child request.
func TestPackagedBinaryIdentityAndRequiredBuildSettings(t *testing.T) {
	repo := residentRepoRoot(t)
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	build := func(trim bool) []byte {
		t.Helper()
		dest := filepath.Join(t.TempDir(), "child")
		args := []string{"build", "-buildvcs=false", "-p=1", "-o", dest}
		if trim {
			args = append(args, "-trimpath")
		}
		args = append(args, "./cmd/riido-shortclaim")
		cmd := exec.Command(tool, args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=", "GOEXPERIMENT=", "GOMAXPROCS=1", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
		if _, e := cmd.CombinedOutput(); e != nil {
			t.Fatal("packaged build control failed")
		}
		raw, e := os.ReadFile(dest)
		if e != nil {
			t.Fatal("packaged binary unavailable")
		}
		return raw
	}
	const childPath = "github.com/teamswyg/laya-tools/cmd/riido-shortclaim"
	raw := build(true)
	if e := verifyBinary(raw, hashBytes(raw), childPath); e != nil {
		t.Fatal("actual required build rejected")
	}
	if verifyBinary(raw, strings.Repeat("0", 64), childPath) == nil || verifyBinary(raw, hashBytes(raw), childPath+"wrong") == nil || verifyBinary([]byte("not a binary"), hashBytes([]byte("not a binary")), childPath) == nil {
		t.Fatal("invalid packaged identity accepted")
	}
	bad := build(false)
	if verifyBinary(bad, hashBytes(bad), childPath) == nil {
		t.Fatal("non-trimpath packaged build accepted")
	}
}
