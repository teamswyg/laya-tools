package taskverify

import "encoding/json"

const humanizeOrdinalTask = "go53-humanize-ordinal64"
const humanizeOrdinalRevision = "a1b4e66b9a6d890e9e15e7091cf16c8032367d6e"

// Preserve the original source-candidate request; the independent contract
// makes its full-range behavior and existing API preservation explicit.
const humanizeOrdinalPrompt = "Add Ordinal64 choosing the English suffix from the magnitude while preserving the sign across the full int64 range."

func humanizeOrdinalDefinition() Definition {
	const pkg = "github.com/dustin/go-humanize"
	return Definition{
		Schema: DefinitionSchema, ID: humanizeOrdinalTask, Version: 1,
		BaseRevision: humanizeOrdinalRevision,
		SourceURL:    "https://github.com/dustin/go-humanize/tree/" + humanizeOrdinalRevision,
		License:      "MIT", SourcePath: "ordinals.go", ModulePath: pkg,
		MutablePaths: []string{"ordinals.go"},
		Files: []FileHash{
			{"go.mod", "73ef52590e41200a113c6f8fbffd10cc5eb90dc0dc11d79bc5d46d85bddf9c0d"},
			{"ordinals.go", "aac3d5ceefd8044baae1f3deb76613470c7eb94fc26af1ea51cec93f8eab075f"},
			{"ordinals_test.go", "c0ea06ca0466e03f24ae61b73001fbe0525fa045cbc29e7154e127fc485627bb"},
			{"common_test.go", "6da7856470f2295db251c715ef253aa1e07eb7a8335a246f08ac48b4b7dedd82"},
			{"LICENSE", "a973b4498c13eb74baa2a8e5c351426a6826f2fcdd909916dbe53ee2e755fd71"},
		},
		Imports:     []string{"strconv"},
		PackageName: "humanize", Packages: []string{pkg},
		RequiredTests: []TerminalTest{
			{pkg, "TestOrdinals"},
			{pkg, "TestTaskVerifyOrdinal64SignatureAndExamples"},
			{pkg, "TestTaskVerifyOrdinal64EverySuffixResidue"},
			{pkg, "TestTaskVerifyOrdinal64FullRangeDigits"},
			{pkg, "TestTaskVerifyOrdinal64DeterministicRange"},
			{pkg, "TestTaskVerifyOriginalIntOrdinal"},
		},
		ContractPath:   "taskverify_contract_test.go",
		ContractSHA256: digest([]byte(humanizeOrdinalContractTests)),
		PromptSHA256:   digest([]byte(humanizeOrdinalPrompt)),
	}
}

func humanizeOrdinalSpec(d Definition) Spec {
	b, _ := json.Marshal(d)
	return Spec{
		ID: d.ID, Version: d.Version, BaseRevision: d.BaseRevision,
		SourcePath: d.SourcePath, Prompt: humanizeOrdinalPrompt,
		Acceptance: []string{
			"Ordinal64 has the exact func(int64) string signature and preserves the signed base-10 digits including both inclusive int64 endpoints",
			"Suffixes use the magnitude: residues 11,12,13 modulo100 are th; other final digits 1,2,3 use st,nd,rd; all remaining residues including zero use th",
			"The existing Ordinal(int) API retains its exact original behavior, including negative inputs using th",
			"Original MIT license and source/helper/test bytes remain pinned; only ordinals.go is mutable and candidate tests never determine acceptance",
			"The original TestOrdinals and every independently supplied named contract terminate with pass in exactly github.com/dustin/go-humanize under offline isolation",
			"Only this minimal ordinal source closure is assessed; broader upstream APIs, other file licenses, model execution and training eligibility are unassessed",
			"Changed imports, init functions or compiler directives exceed the supported execution shape and produce verifier_unknown",
		},
		AcceptanceSourceSHA256: d.ContractSHA256, DefinitionSHA256: digest(b),
	}
}
