package taskverify

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
)

const DefinitionSchema = "riido-public-task-definition-v1"
const keywordGuardTask = "repo-keyword-language-guard"
const keywordGuardRevision = "146b02b9c37e6d90a11386050ccfdffc036c113e"

// TerminalTest names an independently supplied assertion that must actually run
// and terminate with pass in its exact package. A package-only pass is insufficient.
type TerminalTest struct {
	Package string `json:"package"`
	Test    string `json:"test"`
}

// Definition binds one versioned public source closure and its independent
// acceptance runner. It never supplies candidate assertions or model outcomes.
// TaskDefinition returns owned slices, so callers cannot mutate registry state.
type Definition struct {
	Schema                 string         `json:"schema"`
	ID                     string         `json:"id"`
	Version                int            `json:"version"`
	BaseRevision           string         `json:"base_revision"`
	SourceURL              string         `json:"source_url"`
	License                string         `json:"license"`
	ModulePath             string         `json:"module_path,omitempty"`
	SourcePath             string         `json:"source_path"`
	MutablePaths           []string       `json:"mutable_paths"`
	Files                  []FileHash     `json:"files"`
	Imports                []string       `json:"imports"`
	PackageName            string         `json:"package_name"`
	Packages               []string       `json:"packages"`
	RequiredTests          []TerminalTest `json:"required_tests"`
	ContractPath           string         `json:"contract_path"`
	ContractSHA256         string         `json:"contract_sha256"`
	PromptSHA256           string         `json:"prompt_sha256"`
	CandidateTestsUsed     bool           `json:"candidate_tests_used"`
	LogicalTaskID          string         `json:"logical_task_id,omitempty"`
	PreviousTaskID         string         `json:"previous_task_id,omitempty"`
	EvaluationRecipeSHA256 string         `json:"evaluation_recipe_sha256,omitempty"`
}

const keywordGuardPrompt = "Extend repository preview's existing metadata-language guard to inspect every keyword of each selected lexical candidate before Judge. If any selected candidate summary or keyword contains a non-Latin letter, preserve the lexical candidate result and return reason metadata_language_unvalidated without calling Judge. Inspect only the selected candidates. Preserve nil-Judge lexical behavior, query-language guards, and all existing English judgment/threshold/margin/truncation behavior."

// TaskDefinition exposes only the new versioned registry. The frozen original
// three TaskSpecs, BaseRevision, BasePaths and source pins remain independent.
func TaskDefinition(id string) (Definition, bool) {
	if id == humanizeOrdinalTaskV2 || id == uuidCanonicalTaskV2 {
		return upstreamDefinitionV2(id), true
	}
	if id == humanizeOrdinalTask {
		return humanizeOrdinalDefinition(), true
	}
	if id == uuidCanonicalTask {
		return uuidCanonicalDefinition(), true
	}
	if id == eventKeyBoundsTask {
		return eventKeyBoundsDefinition(), true
	}
	if id != keywordGuardTask {
		return Definition{}, false
	}
	const pkg = "github.com/teamswyg/laya-tools/pkg/reporouter"
	d := Definition{
		Schema: DefinitionSchema, ID: keywordGuardTask, Version: 1,
		BaseRevision: keywordGuardRevision,
		SourceURL:    "https://github.com/teamswyg/laya-tools/tree/" + keywordGuardRevision,
		License:      "Apache-2.0", SourcePath: "pkg/reporouter/router.go",
		MutablePaths: []string{"pkg/reporouter/router.go"},
		Files: []FileHash{
			{"go.mod", "d3dac31b277d3945ff0333c5099e5e5c47807bd8310b0c9d71d29dbdeff8e9ec"},
			{"pkg/reporouter/router.go", "3b39a732f230782caff03fa4fbb1a7ed4ca146802716dd7c737122a84676d1f0"},
			{"pkg/reporouter/router_test.go", "e13f210b9719838ce53554d0ef848976a2bf7f458ebd4eadac519a76fc27cd29"},
			{"LICENSE", "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9"},
			{"NOTICE", "00b0aa3ee756ede902b6c354550ba59e9f9ab54764ff78957dbbc3e0ab791bae"},
		},
		Imports:     []string{"fmt", "math", "regexp", "sort", "strings", "unicode"},
		PackageName: "reporouter", Packages: []string{pkg},
		RequiredTests: []TerminalTest{
			{pkg, "TestTaskVerifyEverySelectedKeyword"},
			{pkg, "TestTaskVerifySelectedScope"},
			{pkg, "TestTaskVerifyNilJudge"},
			{pkg, "TestTaskVerifyExistingLanguageGuards"},
			{pkg, "TestTaskVerifyEnglishJudgmentGuards"},
		},
		ContractPath:   "pkg/reporouter/taskverify_contract_test.go",
		ContractSHA256: digest([]byte(keywordGuardContractTests)),
		PromptSHA256:   digest([]byte(keywordGuardPrompt)),
	}
	return d, true
}

func definitionSpec(d Definition) Spec {
	if d.ID == humanizeOrdinalTaskV2 || d.ID == uuidCanonicalTaskV2 {
		return upstreamSpecV2(d)
	}
	if d.ID == humanizeOrdinalTask {
		return humanizeOrdinalSpec(d)
	}
	if d.ID == uuidCanonicalTask {
		return uuidCanonicalSpec(d)
	}
	if d.ID == eventKeyBoundsTask {
		return eventKeyBoundsSpec(d)
	}
	b, _ := json.Marshal(d)
	return Spec{
		ID: d.ID, Version: d.Version, BaseRevision: d.BaseRevision,
		SourcePath: d.SourcePath, Prompt: keywordGuardPrompt,
		Acceptance: []string{
			"Every keyword of every selected candidate is checked before Judge; blocked metadata changes only the lexical result reason",
			"Unselected/disabled repositories do not affect the metadata language guard; nil-Judge lexical behavior and existing language/judgment guards are preserved",
			"Immutable public base tests and all named independently supplied contract tests pass in the exact package under offline isolation",
			"Only the declared source closure is assessed; candidate-owned tests never determine acceptance; original LICENSE and NOTICE staging bytes are pinned separately",
			"Changed imports, init functions or compiler directives exceed the supported execution shape and produce verifier_unknown",
		},
		AcceptanceSourceSHA256: d.ContractSHA256, DefinitionSHA256: digest(b),
	}
}

func definitionContractSource(id string) (string, bool) {
	switch id {
	case humanizeOrdinalTask, humanizeOrdinalTaskV2:
		return humanizeOrdinalContractTests, true
	case uuidCanonicalTask, uuidCanonicalTaskV2:
		return uuidCanonicalContractTests, true
	case keywordGuardTask:
		return keywordGuardContractTests, true
	case eventKeyBoundsTask:
		return eventKeyBoundsContractTests, true
	default:
		return "", false
	}
}

// The module identity belongs to the trusted, pinned definition. It is never
// inferred from a candidate go.mod or an inherited repository configuration.
func definitionModulePath(d Definition) (string, error) {
	module := d.ModulePath
	if module == "" {
		module = "github.com/teamswyg/laya-tools"
	}
	if !strings.Contains(module, "/") {
		return "", fmt.Errorf("invalid_module_identity")
	}
	if !validDefinitionImportPath(module) {
		return "", fmt.Errorf("invalid_module_identity")
	}
	if len(d.Packages) == 0 {
		return "", fmt.Errorf("invalid_module_identity")
	}
	for _, pkg := range d.Packages {
		if !validDefinitionImportPath(pkg) || pkg != module && !strings.HasPrefix(pkg, module+"/") {
			return "", fmt.Errorf("invalid_module_identity")
		}
	}
	return module, nil
}

func validDefinitionImportPath(p string) bool {
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
		for _, ch := range part {
			if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' || ch == '.' {
				continue
			}
			return false
		}
	}
	return true
}

func definitionPaths(d Definition) []string {
	paths := make([]string, 0, len(d.Files))
	for _, f := range d.Files {
		if f.Path != "LICENSE" && f.Path != "NOTICE" {
			paths = append(paths, f.Path)
		}
	}
	return paths
}

// AttributionPins returns owned immutable public license/notice identities for
// staging. Attribution does not extend the candidate source-acceptance scope.
func AttributionPins(id string) ([]FileHash, error) {
	if d, ok := TaskDefinition(id); ok {
		license := definitionPin(d, "LICENSE")
		if license == "" {
			return nil, fmt.Errorf("invalid_attribution_manifest")
		}
		pins := []FileHash{{"LICENSE", license}}
		if notice := definitionPin(d, "NOTICE"); notice != "" {
			pins = append(pins, FileHash{"NOTICE", notice})
		}
		return pins, nil
	}
	if _, err := TaskSpec(id); err != nil {
		return nil, fmt.Errorf("unknown_task")
	}
	return []FileHash{
		{"LICENSE", "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9"},
		{"NOTICE", "00b0aa3ee756ede902b6c354550ba59e9f9ab54764ff78957dbbc3e0ab791bae"},
	}, nil
}

func validateAttribution(id string, supplied []BaseFile) ([]BaseFile, error) {
	pins, err := AttributionPins(id)
	if err != nil || len(supplied) != len(pins) {
		return nil, fmt.Errorf("invalid_attribution_manifest")
	}
	owned := make([]BaseFile, 0, len(pins))
	for _, pin := range pins {
		count := 0
		for _, f := range supplied {
			if f.Path != pin.Path {
				continue
			}
			count++
			if len(f.Data) > MaxFileBytes || f.SHA256 != pin.SHA256 || digest(f.Data) != pin.SHA256 {
				return nil, fmt.Errorf("attribution_pin_mismatch")
			}
			owned = append(owned, BaseFile{Path: f.Path, SHA256: f.SHA256, Data: slices.Clone(f.Data)})
		}
		if count != 1 {
			return nil, fmt.Errorf("invalid_attribution_manifest")
		}
	}
	return owned, nil
}

func definitionPin(d Definition, p string) string {
	for _, f := range d.Files {
		if f.Path == p {
			return f.SHA256
		}
	}
	return ""
}

func definitionMutable(d Definition, p string) bool { return slices.Contains(d.MutablePaths, p) }

func supportedDefinitionSource(d Definition, source []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), d.SourcePath, source, parser.ParseComments)
	if err != nil || f.Name.Name != d.PackageName {
		return false
	}
	imports := make([]string, 0, len(f.Imports))
	for _, im := range f.Imports {
		p, err := strconv.Unquote(im.Path.Value)
		if err != nil || im.Name != nil {
			return false
		}
		imports = append(imports, p)
	}
	slices.Sort(imports)
	if !slices.Equal(imports, d.Imports) {
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
	if d.ID == uuidCanonicalTask || d.ID == uuidCanonicalTaskV2 {
		return uuidCanonicalSupportedSource(source)
	}
	return true
}
