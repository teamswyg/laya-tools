// riido-propertyaudit is a bounded offline truth audit, not a model or router.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/finiteproperty"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/scopedproperty"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

//go:embed main.go
var compiledMain string

const maxBytes = 1 << 20

type artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type plan struct {
	Schema        string     `json:"schema"`
	SourceCommit  string     `json:"source_commit"`
	GoVersion     string     `json:"go_version"`
	GOOS          string     `json:"goos"`
	GOARCH        string     `json:"goarch"`
	BinarySHA256  string     `json:"binary_sha256"`
	BinaryBytes   int        `json:"binary_bytes"`
	DatasetSHA256 string     `json:"dataset_sha256"`
	DatasetBytes  int        `json:"dataset_bytes"`
	Sources       []artifact `json:"compiled_sources"`
	Support       []artifact `json:"support_artifacts"`
	ReviewID      string     `json:"caption_review_id"`
	Attempts      int        `json:"official_attempts"`
	Retries       int        `json:"retries"`
	CPUThreads    int        `json:"go_max_procs"`
	HeapSoftLimit int64      `json:"go_heap_soft_limit_bytes"`
	Scope         string     `json:"scope"`
}

type record struct {
	Schema           string                 `json:"schema"`
	State            string                 `json:"state"`
	PlanSHA256       string                 `json:"plan_sha256"`
	SourceCommit     string                 `json:"source_commit"`
	InputCommit      string                 `json:"input_freeze_commit"`
	GoVersion        string                 `json:"go_version"`
	GOOS             string                 `json:"goos"`
	GOARCH           string                 `json:"goarch"`
	BinarySHA256     string                 `json:"binary_sha256"`
	GitBlobsVerified int                    `json:"git_blobs_verified_before_observation"`
	Observation      finiteproperty.Results `json:"observation"`
	Limitations      []string               `json:"limitations"`
}

var requiredSources = [...]string{
	"cmd/riido-propertyaudit/main.go",
	"internal/behaviorprobe/audit_provenance.go",
	"internal/behaviorprobe/data.go",
	"internal/finiteproperty/audit.go",
	"internal/finiteproperty/dataset.go",
	"internal/lexicalhint/audit_provenance.go",
	"internal/lexicalhint/features.go",
	"internal/scopedproperty/dataset.go",
	"internal/scopedproperty/spec.go",
	"internal/typedbehavior/audit.go",
	"internal/typedbehavior/dataset.go",
	"internal/typedbehavior/fixtures.go",
	"internal/typedbehavior/flow.go",
	"internal/typedbehavior/groups.go",
	"internal/typedbehavior/property_observation.go",
	"internal/typedbehavior/source.go",
	"internal/typedbehavior/spec.go",
	"internal/typedbehavior/state.go",
	"pkg/shortclaim/audit_provenance.go",
	"pkg/shortclaim/baseline.go",
	"pkg/shortclaim/input.go",
}

var supportPaths = [...]string{
	"cmd/riido-propertyaudit/main_test.go",
	"experiments/short-claim/CAPTION-REVIEW-56e.en.md",
	"experiments/short-claim/CAPTION-REVIEW-56e.ko.md",
	"experiments/short-claim/PLAN-56e.en.md",
	"experiments/short-claim/PLAN-56e.ko.md",
	"experiments/short-claim/build-recipe-56e.json",
	"experiments/short-claim/preparation-plan-56c.json",
	"experiments/short-claim/probes-56b.json",
	"experiments/short-claim/probes-56c.json",
	"experiments/short-claim/preparation-56c.json",
	"go.mod",
	"go.sum",
	"internal/finiteproperty/dataset_test.go",
}

func sources() ([]artifact, error) {
	out := []artifact{{"cmd/riido-propertyaudit/main.go", finiteproperty.SHA256([]byte(compiledMain))}}
	add := func(prefix string, entries []typedbehavior.SourceArtifact) {
		for _, e := range entries {
			out = append(out, artifact{prefix + e.Name, e.SHA256})
		}
	}
	add("internal/typedbehavior/", typedbehavior.SourceArtifacts())
	add("internal/typedbehavior/", typedbehavior.PropertyObservationSourceArtifacts())
	add("internal/scopedproperty/", scopedproperty.SourceArtifacts())
	finite := finiteproperty.SourceArtifacts()
	add("internal/finiteproperty/", finite[:])
	for _, e := range behaviorprobe.AuditSourceArtifacts() {
		out = append(out, artifact{"internal/behaviorprobe/" + e.Name, e.SHA256})
	}
	for _, e := range lexicalhint.AuditSourceArtifacts() {
		out = append(out, artifact{"internal/lexicalhint/" + e.Name, e.SHA256})
	}
	for _, e := range shortclaim.AuditSourceArtifacts() {
		out = append(out, artifact{"pkg/shortclaim/" + e.Name, e.SHA256})
	}
	slices.SortFunc(out, func(a, b artifact) int { return strings.Compare(a.Path, b.Path) })
	if len(out) != len(requiredSources) {
		return nil, errors.New("propertyaudit_source_manifest_invalid")
	}
	for i, a := range out {
		if a.Path != requiredSources[i] || !hash(a.SHA256, 64) {
			return nil, errors.New("propertyaudit_source_manifest_invalid")
		}
	}
	return out, nil
}

func hash(s string, n int) bool {
	_, err := hex.DecodeString(s)
	return len(s) == n && strings.ToLower(s) == s && err == nil
}

func bounded(root *os.Root, path string) ([]byte, error) {
	f, err := root.Open(path)
	if err != nil {
		return nil, errors.New("propertyaudit_file_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return nil, errors.New("propertyaudit_file_bounds")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(raw) > maxBytes {
		return nil, errors.New("propertyaudit_file_bounds")
	}
	return raw, nil
}

func support(root *os.Root) ([]artifact, error) {
	out := make([]artifact, len(supportPaths))
	for i, path := range supportPaths {
		raw, err := bounded(root, path)
		if err != nil {
			return nil, err
		}
		out[i] = artifact{path, finiteproperty.SHA256(raw)}
	}
	return out, nil
}

func verifyFiles(root *os.Root, entries []artifact) error {
	for _, a := range entries {
		raw, err := bounded(root, a.Path)
		if err != nil {
			return err
		}
		if finiteproperty.SHA256(raw) != a.SHA256 {
			return errors.New("propertyaudit_source_or_support_changed")
		}
	}
	return nil
}

func verifyDirectories(root *os.Root) error {
	for _, dir := range []string{"cmd/riido-propertyaudit", "internal/finiteproperty", "internal/behaviorprobe", "internal/lexicalhint", "internal/scopedproperty", "internal/typedbehavior", "pkg/shortclaim"} {
		f, err := root.Open(dir)
		if err != nil {
			return errors.New("propertyaudit_package_unavailable")
		}
		entries, readErr := f.ReadDir(65)
		_ = f.Close()
		if readErr != nil && readErr != io.EOF || len(entries) > 64 {
			return errors.New("propertyaudit_package_bounds")
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") && !slices.Contains(requiredSources[:], dir+"/"+e.Name()) {
				return errors.New("propertyaudit_unlisted_package_source")
			}
		}
	}
	return nil
}

type boundedBuffer struct{ bytes.Buffer }

func (b *boundedBuffer) Write(raw []byte) (int, error) {
	if len(raw) > maxBytes-b.Len() {
		return 0, errors.New("propertyaudit_git_blob_bounds")
	}
	return b.Buffer.Write(raw)
}

// Read fixed paths from the supplied immutable commit, with no shell, network,
// hooks or raw diagnostic publication. A commit's syntax alone is insufficient.
func verifyGit(commit string, entries []artifact) error {
	if !hash(commit, 40) {
		return errors.New("propertyaudit_source_commit_invalid")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, a := range entries {
		cmd := exec.CommandContext(ctx, "git", "--no-pager", "show", commit+":"+a.Path)
		var stdout boundedBuffer
		cmd.Stdout = &stdout
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil || finiteproperty.SHA256(stdout.Bytes()) != a.SHA256 {
			return errors.New("propertyaudit_git_blob_mismatch")
		}
	}
	return nil
}

func binary() (string, int, error) {
	path, err := os.Executable()
	if err != nil {
		return "", 0, errors.New("propertyaudit_binary_unavailable")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", 0, errors.New("propertyaudit_binary_unavailable")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil || len(raw) > 16<<20 {
		return "", 0, errors.New("propertyaudit_binary_bounds")
	}
	return finiteproperty.SHA256(raw), len(raw), nil
}

func writeNew(path string, raw []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("propertyaudit_output_create_failed")
	}
	_, w := f.Write(raw)
	c := f.Close()
	if w != nil || c != nil {
		return errors.New("propertyaudit_output_write_failed")
	}
	return nil
}

// Canonical byte equality additionally rejects duplicate keys, extra array
// elements, null substitution and values silently dropped by encoding/json.
func decode(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return errors.New("propertyaudit_invalid_artifact")
	}
	canonical, err := finiteproperty.JSON(v)
	if err != nil || !bytes.Equal(raw, canonical) {
		return errors.New("propertyaudit_noncanonical_artifact")
	}
	return nil
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("riido-propertyaudit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	stage := fs.String("stage", "", "prepare or audit")
	out := fs.String("out", "", "new output directory")
	in := fs.String("in", "", "frozen preparation directory")
	commit := fs.String("source-commit", "", "actual immutable source commit")
	inputCommit := fs.String("input-commit", "", "actual immutable dataset/plan commit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprintln(stdout, "Usage: riido-propertyaudit --stage prepare --source-commit COMMIT --out NEW_DIRECTORY\nFreeze dataset.json as experiments/short-claim/probes-56e.json and plan.json as execution-plan-56e.json. Then: --stage audit --input-commit INPUT_COMMIT --in PREPARED_DIRECTORY --out NEW_DIRECTORY\nOffline finite truth only. Go1.27.1; run from the public repository. Keep compiled binary unchanged.")
		}
		return err
	}
	if fs.NArg() != 0 || *out == "" || *stage != "prepare" && *stage != "audit" || *stage == "prepare" && (*in != "" || !hash(*commit, 40) || *inputCommit != "") || *stage == "audit" && (*in == "" || *commit != "" || !hash(*inputCommit, 40)) {
		return errors.New("propertyaudit_invalid_arguments")
	}
	if runtime.Version() != "go1.27.1" {
		return errors.New("propertyaudit_toolchain_not_pinned")
	}
	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)
	oldHeap := debug.SetMemoryLimit(256 << 20)
	defer debug.SetMemoryLimit(oldHeap)
	root, err := os.OpenRoot(".")
	if err != nil {
		return errors.New("propertyaudit_source_unavailable")
	}
	defer root.Close()
	compiled, err := sources()
	if err != nil {
		return err
	}
	if err := verifyFiles(root, compiled); err != nil {
		return err
	}
	if err := verifyDirectories(root); err != nil {
		return err
	}
	supports, err := support(root)
	if err != nil {
		return err
	}
	binaryHash, binaryBytes, err := binary()
	if err != nil {
		return err
	}
	var outputs [2][]byte
	var names [2]string
	var auditErr error
	var calls, completed int
	if *stage == "prepare" {
		if err := verifyGit(*commit, append(slices.Clone(compiled), supports...)); err != nil {
			return err
		}
		dataset, err := finiteproperty.Prepare()
		if err != nil {
			return err
		}
		dataRaw, err := finiteproperty.JSON(dataset)
		if err != nil {
			return err
		}
		p := plan{Schema: "riido-finite-property-plan-56e-v2", SourceCommit: *commit, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
			BinarySHA256: binaryHash, BinaryBytes: binaryBytes, DatasetSHA256: finiteproperty.SHA256(dataRaw), DatasetBytes: len(dataRaw), Sources: compiled, Support: supports,
			ReviewID: dataset.ReviewID, Attempts: 1, CPUThreads: 1, HeapSoftLimit: 256 << 20,
			Scope: "One official truth audit; four connected sources, nine inherited finite inputs, twelve derived rows. Caption review precedes observation. No ranking, learned inference, paid calls, fitting, split or final access."}
		planRaw, err := finiteproperty.JSON(p)
		if err != nil {
			return err
		}
		outputs, names = [2][]byte{dataRaw, planRaw}, [2]string{"dataset.json", "plan.json"}
	} else {
		inputRoot, err := os.OpenRoot(*in)
		if err != nil {
			return errors.New("propertyaudit_input_unavailable")
		}
		defer inputRoot.Close()
		planRaw, err := bounded(inputRoot, "plan.json")
		if err != nil {
			return err
		}
		dataRaw, err := bounded(inputRoot, "dataset.json")
		if err != nil {
			return err
		}
		var p plan
		var data finiteproperty.Dataset
		if err := decode(planRaw, &p); err != nil {
			return err
		}
		if err := decode(dataRaw, &data); err != nil {
			return err
		}
		if p.Schema != "riido-finite-property-plan-56e-v2" || p.GoVersion != runtime.Version() || p.GOOS != runtime.GOOS || p.GOARCH != runtime.GOARCH || p.BinarySHA256 != binaryHash || p.BinaryBytes != binaryBytes || p.DatasetSHA256 != finiteproperty.SHA256(dataRaw) || p.DatasetBytes != len(dataRaw) || !slices.Equal(p.Sources, compiled) || !slices.Equal(p.Support, supports) || p.ReviewID != data.ReviewID || p.Attempts != 1 || p.Retries != 0 || p.CPUThreads != 1 || p.HeapSoftLimit != 256<<20 || p.Scope != "One official truth audit; four connected sources, nine inherited finite inputs, twelve derived rows. Caption review precedes observation. No ranking, learned inference, paid calls, fitting, split or final access." {
			return errors.New("propertyaudit_plan_binding_invalid")
		}
		if err := verifyGit(p.SourceCommit, append(slices.Clone(compiled), supports...)); err != nil {
			return err
		}
		inputs := []artifact{{"experiments/short-claim/probes-56e.json", finiteproperty.SHA256(dataRaw)}, {"experiments/short-claim/execution-plan-56e.json", finiteproperty.SHA256(planRaw)}}
		if err := verifyGit(*inputCommit, inputs); err != nil {
			return err
		}
		// Claim the new output directory before observations. Reusing an existing
		// output fails before any candidate runs; separate CI replay is labeled.
		if err := os.Mkdir(*out, 0700); err != nil {
			return errors.New("propertyaudit_output_directory_exists")
		}
		got, observationErr := finiteproperty.Audit(data)
		auditErr = observationErr
		calls, completed = got.ActualCandidateCalls, got.CompletedObservations
		state := "complete"
		if auditErr != nil {
			state = "audit_failed"
		}
		r := record{Schema: "riido-finite-property-record-56e-v2", State: state, PlanSHA256: finiteproperty.SHA256(planRaw), SourceCommit: p.SourceCommit, InputCommit: *inputCommit, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
			BinarySHA256: binaryHash, GitBlobsVerified: len(compiled) + len(supports) + len(inputs), Observation: got,
			Limitations: []string{"Finite literal agreement is not a general function proof or caption-generation score.", "Twelve derived rows, thirty-six candidate labels and thirty-six observed source-input pairs add zero independent parents; all four sources share a helper family.", "Preflight reconstructs typed24, pending12 and finite12 rows:192 text normalizations enforce bounds only; no scoring features or ranking are computed.", "Original56c pending captions and56b unknowns remain unchanged. Truth is inherited authored literals, never candidate tests or whole-function control roles.", "No utility, latency, CPU/RSS/GPU, model, fit, protected-final, production activation or LLM-savings result in this audit."}}
		raw, err := finiteproperty.JSON(r)
		if err != nil {
			return err
		}
		outputs, names = [2][]byte{raw, nil}, [2]string{"results.json", ""}
	}
	if *stage == "prepare" {
		if err := os.Mkdir(*out, 0700); err != nil {
			return errors.New("propertyaudit_output_directory_exists")
		}
	}
	for i, raw := range outputs {
		if raw == nil {
			continue
		}
		if len(raw) > maxBytes {
			return errors.New("propertyaudit_output_bounds")
		}
		if err := writeNew(*out+string(os.PathSeparator)+names[i], raw); err != nil {
			_, _ = fmt.Fprintf(stdout, "propertyaudit_output_failed actual_candidate_calls=%d completed_observations=%d\n", calls, completed)
			return err
		}
	}
	if auditErr != nil {
		_, _ = fmt.Fprintf(stdout, "propertyaudit_audit_failed actual_candidate_calls=%d completed_observations=%d\n", calls, completed)
		return errors.New("propertyaudit_observation_failed")
	}
	_, err = fmt.Fprintf(stdout, "propertyaudit_%s_complete\n", *stage)
	return err
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		_, _ = fmt.Fprintln(os.Stderr, "propertyaudit_failed")
		os.Exit(1)
	}
}
