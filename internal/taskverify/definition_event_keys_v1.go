package taskverify

import "encoding/json"

const eventKeyBoundsTask = "taskoutcome-event-key-bounds"
const eventKeyBoundsRevision = "ee72334166e2962b0821d5198a50fdd13f92ab29"
const eventKeyBoundsPrompt = "In pkg/taskoutcome/summary.go, make every JSONL event's top-level keys unambiguous and bounded: reject any duplicate decoded key, including unknown keys and escape-equivalent names, with ErrDuplicate. Accept at most 64 top-level keys including control keys, and at most 128 UTF-8 bytes per decoded key; reject larger count or key size with ErrJSON. Limits and duplicate tracking reset for each event. Keep key comparison case-sensitive. Preserve existing lifecycle, usage validation/totals, startup diagnostic handling, fixed redacted errors, and opaque active-turn item contents. Apply these new limits only to the event envelope, not recursively to nested item payloads; existing nested usage/startup control rules stay unchanged. Do not retain unknown key names or values in Summary."

func eventKeyBoundsDefinition() Definition {
	const pkg = "github.com/teamswyg/laya-tools/pkg/taskoutcome"
	return Definition{
		Schema: DefinitionSchema, ID: eventKeyBoundsTask, Version: 1,
		BaseRevision: eventKeyBoundsRevision,
		SourceURL:    "https://github.com/teamswyg/laya-tools/tree/" + eventKeyBoundsRevision,
		License:      "Apache-2.0", SourcePath: "pkg/taskoutcome/summary.go",
		MutablePaths: []string{"pkg/taskoutcome/summary.go"},
		Files: []FileHash{
			{"go.mod", "d3dac31b277d3945ff0333c5099e5e5c47807bd8310b0c9d71d29dbdeff8e9ec"},
			{"pkg/taskoutcome/summary.go", "688ee64441b9c70c5ec88f9713ae61fb2e0526239a4b2aa781174418278ad7d8"},
			{"pkg/taskoutcome/summary_test.go", "59bc3741a8e72a62ac82a65c4480d4629b549fd76cda775dc5a662b5ebc5b366"},
			{"LICENSE", "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9"},
			{"NOTICE", "00b0aa3ee756ede902b6c354550ba59e9f9ab54764ff78957dbbc3e0ab791bae"},
		},
		Imports:     []string{"bufio", "bytes", "crypto/sha256", "encoding/hex", "encoding/json", "io", "math", "strconv", "strings", "unicode/utf8"},
		PackageName: "taskoutcome", Packages: []string{pkg},
		RequiredTests: []TerminalTest{
			{pkg, "TestTaskVerifyEnvelopeDuplicateKeys"},
			{pkg, "TestTaskVerifyEnvelopeKeyCountBoundaries"},
			{pkg, "TestTaskVerifyDecodedKeyByteBoundaries"},
			{pkg, "TestTaskVerifyPerEventResetAndUsage"},
			{pkg, "TestTaskVerifyOpaqueItemsAndRedaction"},
			{pkg, "TestTaskVerifyExistingNestedControls"},
		},
		ContractPath:   "pkg/taskoutcome/taskverify_contract_test.go",
		ContractSHA256: digest([]byte(eventKeyBoundsContractTests)),
		PromptSHA256:   digest([]byte(eventKeyBoundsPrompt)),
	}
}

func eventKeyBoundsSpec(d Definition) Spec {
	b, _ := json.Marshal(d)
	return Spec{
		ID: d.ID, Version: d.Version, BaseRevision: d.BaseRevision,
		SourcePath: d.SourcePath, Prompt: eventKeyBoundsPrompt,
		Acceptance: []string{
			"All decoded envelope keys, including unknown and escape-equivalent keys, reject duplicates with ErrDuplicate; comparison remains case-sensitive",
			"64 total envelope keys and 128 decoded UTF-8 key bytes are inclusive limits; larger counts/keys reject with ErrJSON; tracking resets per event",
			"Valid lifecycle and partial/complete usage aggregates, exact trace hashes, startup controls and opaque nested item contents retain their existing meanings without raw retention",
			"Immutable public base tests and every named independent envelope contract terminate with pass in the exact package under offline isolation",
			"Only declared source closure is assessed; original LICENSE/NOTICE staging bytes are pinned separately; candidate-authored tests do not determine acceptance",
			"Changed imports, init functions or compiler directives exceed the supported execution shape and produce verifier_unknown",
		},
		AcceptanceSourceSHA256: d.ContractSHA256, DefinitionSHA256: digest(b),
	}
}
