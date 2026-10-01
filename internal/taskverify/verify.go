// Package taskverify independently checks a bounded public task closure.
// It never launches a model or treats candidate-authored tests as acceptance.
package taskverify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

const BaseRevision = "6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3"
const Schema = "riido-task-verification-v1"
const MaxFiles = 32
const MaxFileBytes = 1 << 20
const MaxTotalBytes = 8 << 20
const DefaultTimeout = 45 * time.Second
const MaxTimeout = 60 * time.Second

type BaseFile struct {
	Path   string
	SHA256 string
	Data   []byte
}

type Request struct {
	TaskID       string
	BaseRevision string
	BaseFiles    []BaseFile
	// AttributionFiles pins the original public LICENSE/NOTICE staged with a
	// versioned definition. Frozen original tasks retain their existing contract.
	AttributionFiles []BaseFile
	CandidateDir     string
	Timeout          time.Duration
	// GoRoot optionally supplies the trusted offline installation for packaged
	// trimpath callers whose runtime has no compiled-in GOROOT. It never changes
	// process-global environment and is used only for isolated behavioral checks.
	GoRoot string
}

type FileHash struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Code   string `json:"code"`
}

type Report struct {
	Schema             string     `json:"schema"`
	TaskID             string     `json:"task_id"`
	SpecVersion        int        `json:"spec_version"`
	SpecSHA256         string     `json:"spec_sha256"`
	BaseRevision       string     `json:"base_revision"`
	BaseSHA256         string     `json:"base_sha256"`
	CandidateSHA256    string     `json:"candidate_sha256,omitempty"`
	CandidateScope     string     `json:"candidate_scope"`
	Status             string     `json:"status"`
	Accepted           bool       `json:"accepted"`
	Checks             []Check    `json:"checks"`
	BaseFiles          []FileHash `json:"base_files"`
	CandidateFiles     []FileHash `json:"candidate_files,omitempty"`
	IndependentTests   int        `json:"independent_tests"`
	ExecutionIsolated  bool       `json:"execution_isolated"`
	CandidateTestsUsed bool       `json:"candidate_tests_used"`
	AttributionFiles   []FileHash `json:"attribution_files,omitempty"`
	AttributionSHA256  string     `json:"attribution_sha256,omitempty"`
}

// Spec fixes the task contract before future model runs. It is not an outcome label.
type Spec struct {
	ID                     string   `json:"id"`
	Version                int      `json:"version"`
	BaseRevision           string   `json:"base_revision"`
	SourcePath             string   `json:"source_path"`
	Prompt                 string   `json:"prompt"`
	Acceptance             []string `json:"acceptance"`
	AcceptanceSourceSHA256 string   `json:"acceptance_source_sha256,omitempty"`
	DefinitionSHA256       string   `json:"definition_sha256,omitempty"`
}

func TaskSpec(id string) (Spec, error) {
	if d, ok := TaskDefinition(id); ok {
		return definitionSpec(d), nil
	}
	s := Spec{ID: id, Version: 2, BaseRevision: BaseRevision}
	switch id {
	case "comment-budget-period":
		s.SourcePath = "pkg/catalog/catalog.go"
		s.Prompt = "Replace only the Usage snapshot comment with: Snapshot supplied by the caller for a single budget period; no reservation is made."
		s.Acceptance = []string{"Exact supplied comment replacement across the task closure", "gofmt clean", "Unmodified pinned catalog tests pass in isolation"}
	case "comment-preview-authority":
		s.SourcePath = "pkg/reporouter/router.go"
		s.Prompt = "Change the Result.Status comment to: candidate, suggest, or abstain; a suggestion never authorizes execution."
		s.Acceptance = []string{"Exact supplied comment replacement across the task closure", "gofmt clean"}
	case "catalog-min-context":
		s.SourcePath = "pkg/catalog/catalog.go"
		s.Prompt = "Add optional MinContext int to catalog.Request with JSON min_context,omitempty. Reject negative MinContext values. Mark models with Context below MinContext ineligible with reason min_context_required while preserving existing zero-value behavior and token-capacity checks."
		s.Acceptance = []string{"Fixed independent zero, negative, boundary, insufficient, JSON and overflow checks pass", "Unmodified pinned catalog and planner tests pass", "Only task-closure source is assessed; candidate-owned test assertions are not acceptance evidence", "Changed imports, init functions or compiler directives exceed the execution shape and produce verifier_unknown"}
		s.AcceptanceSourceSHA256 = digest([]byte(contractTests))
	default:
		return Spec{}, fmt.Errorf("unknown_task")
	}
	return s, nil
}

var pinned = []FileHash{
	{"go.mod", "d3dac31b277d3945ff0333c5099e5e5c47807bd8310b0c9d71d29dbdeff8e9ec"},
	{"pkg/catalog/catalog.go", "b425fbe634d2a945c703e58497de61db0b85d2aa3b098bc8d7565868b233307a"},
	{"pkg/catalog/catalog_test.go", "f305a4810566fe7607d7a1b88bced06a4f3b8283efde4feffa4bcb960018cc31"},
	{"pkg/planner/planner.go", "43a1b4850fbcc000827f42a336e38f30cae14a6cf32e2d7cfa899844d295c43a"},
	{"pkg/planner/planner_test.go", "8b730928a75d96a13b218af405d88833faeacda096ee946730919477a8718f65"},
	{"pkg/switchpolicy/policy.go", "b5ce347968ca572a13ab8eb645f5d4de7faba6e6eaa455394b44133e77d1aaf0"},
	{"pkg/switchpolicy/policy_test.go", "34cc45d094115053df6c80ffd027302ba973121953c1cfaa5b56be8b22077a95"},
	{"examples/planner/config.json", "288ecaac052b95d6a8d04384f59b0a224cdf0fbe115711dd52d346b3446192bd"},
	{"examples/planner/request.json", "c0cc45f17f2d52ff6aa4d5eaa446ce0ecf8abfec1dd600a58067f1370af2956e"},
	{"pkg/reporouter/router.go", "3b39a732f230782caff03fa4fbb1a7ed4ca146802716dd7c737122a84676d1f0"},
}

// BasePaths returns the closed file manifest the caller must read from BaseRevision.
// The verifier does not invoke Git or trust a supplied revision string by itself.
func BasePaths(id string) ([]string, error) {
	if d, ok := TaskDefinition(id); ok {
		return definitionPaths(d), nil
	}
	if _, err := TaskSpec(id); err != nil {
		return nil, err
	}
	switch id {
	case "comment-preview-authority":
		return []string{"pkg/reporouter/router.go"}, nil
	case "comment-budget-period":
		return []string{"go.mod", "pkg/catalog/catalog.go", "pkg/catalog/catalog_test.go"}, nil
	default:
		return []string{"go.mod", "pkg/catalog/catalog.go", "pkg/catalog/catalog_test.go", "pkg/planner/planner.go", "pkg/planner/planner_test.go", "pkg/switchpolicy/policy.go", "pkg/switchpolicy/policy_test.go", "examples/planner/config.json", "examples/planner/request.json"}, nil
	}
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hashList(files []BaseFile) []FileHash {
	out := make([]FileHash, len(files))
	for i, f := range files {
		out[i] = FileHash{f.Path, digest(f.Data)}
	}
	slices.SortFunc(out, func(a, b FileHash) int { return strings.Compare(a.Path, b.Path) })
	return out
}

func listDigest(files []FileHash) string {
	b, _ := json.Marshal(files)
	return digest(b)
}

func fileAt(files []BaseFile, p string) []byte {
	for _, f := range files {
		if f.Path == p {
			return f.Data
		}
	}
	return nil
}

func validPath(p string) bool {
	return p != "" && p != "." && path.Clean(p) == p && !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "../") && !strings.ContainsAny(p, "\\\x00")
}

func validateBase(id, revision string, supplied []BaseFile) ([]BaseFile, error) {
	paths, err := BasePaths(id)
	spec, specErr := TaskSpec(id)
	definition, versioned := TaskDefinition(id)
	if err != nil || specErr != nil || revision != spec.BaseRevision || len(supplied) != len(paths) || len(supplied) > MaxFiles {
		return nil, fmt.Errorf("invalid_base_manifest")
	}
	owned := make([]BaseFile, 0, len(paths))
	total := 0
	for _, p := range paths {
		found := 0
		for _, f := range supplied {
			if !validPath(f.Path) || len(f.Data) > MaxFileBytes {
				return nil, fmt.Errorf("invalid_base_file")
			}
			if f.Path != p {
				continue
			}
			found++
			want := ""
			if versioned {
				want = definitionPin(definition, p)
			} else {
				for _, pin := range pinned {
					if pin.Path == p {
						want = pin.SHA256
					}
				}
			}
			if want == "" || f.SHA256 != want || digest(f.Data) != want {
				return nil, fmt.Errorf("base_pin_mismatch")
			}
			total += len(f.Data)
			if total > MaxTotalBytes {
				return nil, fmt.Errorf("base_byte_limit")
			}
			owned = append(owned, BaseFile{f.Path, f.SHA256, bytes.Clone(f.Data)})
		}
		if found != 1 {
			return nil, fmt.Errorf("invalid_base_manifest")
		}
	}
	return owned, nil
}

// readCandidate reads only the declared closure. Files outside it are unassessed.
// Root-constrained handles stop path escape; symlink components and special files fail.
func readCandidate(dir string, base []BaseFile) ([]BaseFile, error) {
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("candidate_root_unavailable")
	}
	defer r.Close()
	out := make([]BaseFile, 0, len(base))
	total := 0
	for _, bf := range base {
		parts := strings.Split(bf.Path, "/")
		for i := range parts {
			info, e := r.Lstat(strings.Join(parts[:i+1], "/"))
			if e != nil || info.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("candidate_file_unavailable")
			}
			if i+1 < len(parts) && !info.IsDir() {
				return nil, fmt.Errorf("candidate_file_unavailable")
			}
		}
		f, e := r.Open(bf.Path)
		if e != nil {
			return nil, fmt.Errorf("candidate_file_unavailable")
		}
		info, e := f.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > MaxFileBytes {
			f.Close()
			return nil, fmt.Errorf("candidate_file_limit")
		}
		data, e := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
		f.Close()
		if e != nil || len(data) > MaxFileBytes {
			return nil, fmt.Errorf("candidate_file_limit")
		}
		total += len(data)
		if total > MaxTotalBytes {
			return nil, fmt.Errorf("candidate_byte_limit")
		}
		out = append(out, BaseFile{bf.Path, digest(data), data})
	}
	return out, nil
}

func Verify(ctx context.Context, req Request) (Report, error) {
	r := Report{Schema: Schema, CandidateScope: "task_closure", Status: "rejected", Checks: []Check{}}
	spec, err := TaskSpec(req.TaskID)
	if err != nil {
		return r, err
	}
	r.TaskID, r.BaseRevision = spec.ID, spec.BaseRevision
	r.SpecVersion = spec.Version
	b, _ := json.Marshal(spec)
	r.SpecSHA256 = digest(b)
	if req.Timeout == 0 {
		req.Timeout = DefaultTimeout
	}
	if req.Timeout <= 0 || req.Timeout > MaxTimeout || ctx == nil {
		return r, fmt.Errorf("invalid_verifier_limits")
	}
	base, err := validateBase(req.TaskID, req.BaseRevision, req.BaseFiles)
	if err != nil {
		return r, err
	}
	r.BaseFiles, r.BaseSHA256 = hashList(base), listDigest(hashList(base))
	r.Checks = append(r.Checks, Check{"base_pin", true, "pinned_public_bytes_verified"})
	definition, versioned := TaskDefinition(req.TaskID)
	if versioned {
		attribution, err := validateAttribution(req.TaskID, req.AttributionFiles)
		if err != nil {
			return r, err
		}
		r.AttributionFiles = hashList(attribution)
		r.AttributionSHA256 = listDigest(r.AttributionFiles)
		r.Checks = append(r.Checks, Check{"attribution_pin", true, "pinned_public_attribution_verified"})
	}
	candidate, err := readCandidate(req.CandidateDir, base)
	if err != nil {
		r.Checks = append(r.Checks, Check{"candidate_read", false, err.Error()})
		return r, nil
	}
	r.CandidateFiles = hashList(candidate)
	r.CandidateSHA256 = listDigest(r.CandidateFiles)
	if r.BaseSHA256 == r.CandidateSHA256 {
		r.Checks = append(r.Checks, Check{"requested_change", false, "unchanged_base"})
		return r, nil
	}
	if versioned {
		for _, bf := range base {
			if !definitionMutable(definition, bf.Path) && !strings.HasSuffix(bf.Path, "_test.go") && !bytes.Equal(bf.Data, fileAt(candidate, bf.Path)) {
				r.Checks = append(r.Checks, Check{"task_scope", false, "unrelated_closure_source_changed"})
				return r, nil
			}
		}
		if !supportedDefinitionSource(definition, fileAt(candidate, spec.SourcePath)) {
			r.Status = "verifier_unknown"
			r.Checks = append(r.Checks, Check{"execution_scope", false, "unsupported_candidate_shape"})
			return r, nil
		}
		r.Checks = append(r.Checks, Check{"execution_scope", true, "versioned_source_shape_verified"})
	} else if req.TaskID != "catalog-min-context" {
		if !exactComment(req.TaskID, base, candidate) {
			r.Checks = append(r.Checks, Check{"requested_change", false, "not_exact_comment_change"})
			return r, nil
		}
		r.Checks = append(r.Checks, Check{"requested_change", true, "exact_comment_change"})
	} else {
		if !minimumField(fileAt(candidate, spec.SourcePath)) {
			r.Checks = append(r.Checks, Check{"request_contract", false, "invalid_min_context_contract"})
			return r, nil
		}
		for _, bf := range base {
			if bf.Path != spec.SourcePath && !strings.HasSuffix(bf.Path, "_test.go") && !bytes.Equal(bf.Data, fileAt(candidate, bf.Path)) {
				r.Checks = append(r.Checks, Check{"task_scope", false, "unrelated_closure_source_changed"})
				return r, nil
			}
		}
		r.Checks = append(r.Checks, Check{"request_contract", true, "min_context_field_verified"})
		if !supportedCatalogSource(fileAt(candidate, spec.SourcePath)) {
			r.Status = "verifier_unknown"
			r.Checks = append(r.Checks, Check{"execution_scope", false, "unsupported_candidate_shape"})
			return r, nil
		}
	}
	source := fileAt(candidate, spec.SourcePath)
	formatted, err := format.Source(source)
	if err != nil || !bytes.Equal(source, formatted) {
		r.Checks = append(r.Checks, Check{"format", false, "gofmt_not_clean"})
		return r, nil
	}
	r.Checks = append(r.Checks, Check{"format", true, "gofmt_clean"})
	if req.TaskID == "comment-preview-authority" {
		r.Status, r.Accepted = "accepted", true
		return r, nil
	}
	// Only candidate catalog.go enters the test module. All existing tests and
	// dependency source are immutable pinned base bytes, not candidate assertions.
	execution := slices.Clone(base)
	for i := range execution {
		if execution[i].Path == spec.SourcePath {
			execution[i].Data = bytes.Clone(source)
		}
	}
	outcome := isolatedTests(ctx, execution, req.TaskID, req.Timeout, req.GoRoot)
	r.ExecutionIsolated, r.IndependentTests = outcome.isolated, outcome.tests
	r.Checks = append(r.Checks, Check{"independent_tests", outcome.passed, outcome.code})
	if outcome.unknown {
		r.Status = "verifier_unknown"
	} else if outcome.passed {
		r.Status, r.Accepted = "accepted", true
	}
	return r, nil
}

func exactComment(id string, base, candidate []BaseFile) bool {
	p := "pkg/catalog/catalog.go"
	old := "// Snapshot for one caller-defined budget period. This is not a reservation ledger."
	newText := "// Snapshot supplied by the caller for a single budget period; no reservation is made."
	if id == "comment-preview-authority" {
		p = "pkg/reporouter/router.go"
		old = "// candidate, suggest, abstain; never execution authority"
		newText = "// candidate, suggest, or abstain; a suggestion never authorizes execution."
	}
	for _, f := range base {
		want := f.Data
		if f.Path == p {
			if bytes.Count(want, []byte(old)) != 1 {
				return false
			}
			want = bytes.Replace(want, []byte(old), []byte(newText), 1)
		}
		if !bytes.Equal(want, fileAt(candidate, f.Path)) {
			return false
		}
	}
	return true
}

func minimumField(source []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), "catalog.go", source, parser.ParseComments)
	if err != nil || f.Name.Name != "catalog" {
		return false
	}
	count := 0
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, item := range gen.Specs {
			ts := item.(*ast.TypeSpec)
			if ts.Name.Name != "Request" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return false
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if name.Name != "MinContext" {
						continue
					}
					count++
					ident, ok := field.Type.(*ast.Ident)
					if !ok || ident.Name != "int" || len(field.Names) != 1 || field.Tag == nil {
						return false
					}
					tag, e := strconv.Unquote(field.Tag.Value)
					if e != nil || tag != `json:"min_context,omitempty"` {
						return false
					}
				}
			}
		}
	}
	return count == 1
}

// The offline execution module deliberately retains the pinned import closure.
// Other source shapes are unsupported, not evidence of a failed task outcome.
func supportedCatalogSource(source []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), "catalog.go", source, parser.ParseComments)
	if err != nil {
		return false
	}
	imports := []string{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil || im.Name != nil {
			return false
		}
		imports = append(imports, p)
	}
	slices.Sort(imports)
	if !slices.Equal(imports, []string{"fmt", "math", "sort", "strings"}) {
		return false
	}
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "go:") {
				return false
			}
		}
	}
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "init" {
			return false
		}
	}
	return true
}
