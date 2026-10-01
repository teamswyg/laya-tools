package main

import (
	"bytes"
	"debug/buildinfo"
	"errors"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"
)

func validCommit(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' && c < 'a' || c > 'f' {
			return false
		}
	}
	return true
}

func runtimePaths() []string {
	return []string{
		"cmd/riido-residentperf/artifacts.go", "cmd/riido-residentperf/main.go", "cmd/riido-residentperf/prepare.go", "cmd/riido-residentperf/process.go", "cmd/riido-residentperf/protocol.go", "cmd/riido-residentperf/types.go",
		"cmd/riido-shortclaim/bench.go", "cmd/riido-shortclaim/main.go",
		"internal/behaviorprobe/audit_provenance.go", "internal/behaviorprobe/data.go",
		"internal/lexicalhint/audit_provenance.go", "internal/lexicalhint/features.go",
		"internal/typedbehavior/audit.go", "internal/typedbehavior/dataset.go", "internal/typedbehavior/fixtures.go", "internal/typedbehavior/flow.go", "internal/typedbehavior/groups.go", "internal/typedbehavior/property_observation.go", "internal/typedbehavior/source.go", "internal/typedbehavior/spec.go", "internal/typedbehavior/state.go",
		"pkg/shortclaim/audit_provenance.go", "pkg/shortclaim/baseline.go", "pkg/shortclaim/input.go",
	}
}

func manifestPaths() []string {
	out := append(runtimePaths(), "go.mod", "go.sum", "cmd/riido-residentperf/main_test.go", "cmd/riido-residentperf/artifacts_test.go", "cmd/riido-residentperf/prepare_test.go", "cmd/riido-residentperf/protocol_test.go", "cmd/riido-residentperf/process_test.go")
	slices.Sort(out)
	return out
}

func rootRead(root *os.Root, path string, cap int64) ([]byte, error) {
	f, e := root.Open(path)
	if e != nil {
		return nil, errors.New("source_file_unavailable")
	}
	defer f.Close()
	return readRegular(f, cap)
}

func currentManifest() ([]FilePin, error) {
	root, e := os.OpenRoot(".")
	if e != nil {
		return nil, errors.New("repository_root_unavailable")
	}
	defer root.Close()
	paths := manifestPaths()
	out := make([]FilePin, len(paths))
	for i, p := range paths {
		raw, e := rootRead(root, p, 2<<20)
		if e != nil {
			return nil, e
		}
		out[i] = FilePin{Path: p, SHA256: hashBytes(raw)}
	}
	return out, nil
}

func verifySources(plan *Plan) error {
	root, e := os.OpenRoot(".")
	if e != nil {
		return errors.New("repository_root_unavailable")
	}
	defer root.Close()
	compiled, e := compiledPins()
	if e != nil {
		return e
	}
	if len(compiled) != 22 {
		return errors.New("compiled_source_manifest_invalid")
	}
	paths := runtimePaths()
	for i, c := range compiled {
		if !slices.Contains(paths, c.Path) || !validSHA(c.SHA256) || (i > 0 && compiled[i-1].Path == c.Path) {
			return errors.New("compiled_source_manifest_invalid")
		}
		raw, e := rootRead(root, c.Path, 2<<20)
		if e != nil || hashBytes(raw) != c.SHA256 {
			return errors.New("compiled_source_mismatch")
		}
	}
	for _, dir := range []string{"cmd/riido-residentperf", "cmd/riido-shortclaim", "internal/behaviorprobe", "internal/lexicalhint", "internal/typedbehavior", "pkg/shortclaim"} {
		f, e := root.Open(dir)
		if e != nil {
			return errors.New("source_directory_unavailable")
		}
		entries, e := f.ReadDir(64)
		if e == nil {
			more, next := f.ReadDir(1)
			if len(more) != 0 || next != io.EOF {
				f.Close()
				return errors.New("source_directory_bounds")
			}
		}
		f.Close()
		if e != nil && e != io.EOF {
			return errors.New("source_directory_unavailable")
		}
		for _, entry := range entries {
			n := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			if !slices.Contains(paths, dir+"/"+n) {
				return errors.New("unlisted_implementation_source")
			}
		}
	}
	if plan == nil {
		return nil
	}
	want := manifestPaths()
	if len(plan.ImplementationFiles) != len(want) {
		return errors.New("source_manifest_incomplete")
	}
	for i, p := range plan.ImplementationFiles {
		if p.Path != want[i] || !validSHA(p.SHA256) {
			return errors.New("source_manifest_invalid")
		}
		raw, e := rootRead(root, p.Path, 2<<20)
		if e != nil || hashBytes(raw) != p.SHA256 {
			return errors.New("source_hash_mismatch")
		}
	}
	recipe, e := rootRead(root, recipePath, 64<<10)
	if e != nil || hashBytes(recipe) != plan.BuildRecipeSHA256 {
		return errors.New("build_recipe_hash_mismatch")
	}
	canonical, e := recipeBytes()
	if e != nil || !bytes.Equal(recipe, canonical) {
		return errors.New("build_recipe_policy_invalid")
	}
	return nil
}

func recipeBytes() ([]byte, error) {
	v := struct {
		Schema      string     `json:"schema"`
		Go          string     `json:"go_version"`
		GOOS        string     `json:"goos"`
		GOARCH      string     `json:"goarch"`
		Environment [11]string `json:"environment"`
		Driver      [8]string  `json:"driver_command"`
		Child       [8]string  `json:"child_command"`
		Scope       string     `json:"scope"`
	}{"riido-resident-build-recipe-v1", goVersion, runtime.GOOS, runtime.GOARCH, [11]string{"CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=", "GOEXPERIMENT=", "GOMAXPROCS=1", "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, [8]string{"go", "build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "PRIVATE_DRIVER_PATH", "./cmd/riido-residentperf"}, [8]string{"go", "build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "PRIVATE_CHILD_PATH", "./cmd/riido-shortclaim"}, "PRIVATE_*_PATH are private output placeholders. Sources/modules are independently frozen in the plan. Build and protocol preparation precede measured replay."}
	return encodePublic(v, 64<<10)
}

func verifyBinary(raw []byte, expected, path string) error {
	if len(raw) == 0 || len(raw) > 32<<20 || !validSHA(expected) || hashBytes(raw) != expected {
		return errors.New("binary_hash_or_bounds_invalid")
	}
	info, e := buildinfo.Read(bytes.NewReader(raw))
	if e != nil || info.GoVersion != goVersion || info.Path != path {
		return errors.New("binary_build_identity_invalid")
	}
	cgo, trim, osOK, archOK, compiler := false, false, false, false, false
	for _, s := range info.Settings {
		switch s.Key {
		case "CGO_ENABLED":
			cgo = s.Value == "0"
		case "-trimpath":
			trim = s.Value == "true"
		case "GOOS":
			osOK = s.Value == runtime.GOOS
		case "GOARCH":
			archOK = s.Value == runtime.GOARCH
		case "-compiler":
			compiler = s.Value == "gc"
		case "vcs.revision":
			return errors.New("binary_buildvcs_not_disabled")
		case "-tags", "GOEXPERIMENT":
			if s.Value != "" {
				return errors.New("binary_extra_build_configuration")
			}
		}
	}
	if !cgo || !trim || !osOK || !archOK || !compiler {
		return errors.New("binary_build_settings_invalid")
	}
	return nil
}
