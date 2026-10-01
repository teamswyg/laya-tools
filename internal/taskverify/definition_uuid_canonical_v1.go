package taskverify

import "encoding/json"

const uuidCanonicalTask = "go53-uuid-canonical-parse"
const uuidCanonicalRevision = "2d3c2a9cc518326daf99a383f07c4d3c44317e4d"
const uuidCanonicalModulePath = "github.com/google/uuid"
const uuidCanonicalPrompt = "Append only func ParseCanonical(s string) (UUID, error) to the end of uuid.go, preserving the original code and attribution exactly apart from the trusted Go 1.27.1 gofmt normalization. Accept exactly 36 bytes: lowercase ASCII hexadecimal digits in the 32 digit positions and hyphens only at byte offsets 8, 13, 18 and 23. Return the exact decoded 16 bytes, including the all-zero and all-ones identifiers; do not restrict version or variant bits. Reject uppercase, URN/raw/bracketed/wrapped forms, whitespace, malformed separators, invalid hex and every other byte length with a non-nil error and the zero UUID, even after a valid prefix. Preserve existing Parse and ParseBytes behavior, including their broad formats and error categories. The new parser must be deterministic and must not read entropy or change package-global/random-pool state. This verifier supports only one new function using local values, loops, the original pure Parse/xtob helpers, pure builtins/conversions and selected pure strings/bytes/hex/fmt/errors operations. Calls through function aliases, function literals, pointers/global-array aliases, helpers, concurrency, init functions, imports and compiler directives are outside its supported shape and produce verifier_unknown, even for otherwise correct implementations."

// uuidCanonicalDefinition binds a bounded public development contract.
// Its presence in the registry does not supply a model outcome or difficulty label.
func uuidCanonicalDefinition() Definition {
	const pkg = uuidCanonicalModulePath
	return Definition{
		Schema: DefinitionSchema, ID: uuidCanonicalTask, Version: 1,
		BaseRevision: uuidCanonicalRevision,
		ModulePath:   uuidCanonicalModulePath,
		SourceURL:    "https://github.com/google/uuid/tree/" + uuidCanonicalRevision,
		License:      "BSD-3-Clause", SourcePath: "uuid.go",
		MutablePaths: []string{"uuid.go"},
		Files: []FileHash{
			{"go.mod", "73da47b6338b00a082fd451aa35a3273d3adc09b8e9bba98dab01091e402af6e"},
			{"dce.go", "6a4b89e7f6165a673e76f00bebaabf8d0d7eb4356a672c6cb771d95a45fb2695"},
			{"doc.go", "d0e86d759ab8cd66ab72bb4b2c00b40475a5ceb4a2ff4cf881dcae91f1482491"},
			{"hash.go", "588b314f396d99aa304c9e89b6786e9280ec430223349f2ebb48d9145272f81d"},
			{"json_test.go", "6f7aa3003d72ed5899d269941ae2ad0df0494a66f4b38d5f61fc79aafb9e5b91"},
			{"marshal.go", "933eb226d049268e17809fd37f777c650b9ec8cb3fa988534b792610a64dcc43"},
			{"node.go", "73a93d192ca24078e0f93b8c1675e13adb442421252d3974d697b0821c4832cf"},
			{"node_js.go", "334b6ac9f50339341696319aaf9b912ea8007d0aacf0e909a48a0bb02d982d6d"},
			{"node_net.go", "78e274b465de8da8690f4ab06281f41025b142c1d093e337fda1d6726b35662b"},
			{"null.go", "791638f42d54fcf07ac752df6b2df92593f7da308f944465e2cda443dc3346cf"},
			{"null_test.go", "34e2e5bffa6a6dce2599573e25c576ef476b34fc091eb929da7cc269bc0eccff"},
			{"seq_test.go", "6ad9e9722a9add93984ce7e1d2eeb8dcd72c32134f06bb4b0b7c7ef293338585"},
			{"sql.go", "29281b11e8f97f1ddf73d087854cb401ba42a87c4f59ce45fc69763a4d927f2f"},
			{"sql_test.go", "f89320ca4dd3a7aeb5c69e9216d44c990cc86ab023fed2435c4b09ebec7af78f"},
			{"time.go", "4154b2d7897acfff253c4c2ac0d38d1765cfe3f6ca2e482c71c0677d98380d63"},
			{"time_test.go", "4aadacc3d9ed4b0a281e4bc0aaa030192cb01073e06533dcfb63540078c1b002"},
			{"util.go", "4237eba5a7bc7934b938c9b0bce5fa502d832e5660de71cbef130582e517f89e"},
			{"uuid.go", "00d93a3f65d8b0063266e63ff0e16ac543e30a911fd35119bcf1a66437161360"},
			{"uuid_test.go", "6d23625e6a86f809d76fcabfdddffd16fc438e60c610f71bdb1a2c0ccc7f942e"},
			{"version1.go", "38a2c229a2aad4bb72d45529304f210b1f31e8ef60bbd2605690a10f0d8a8242"},
			{"version4.go", "f252aeb4028659d83cbf7b037d4524f7c9b76cde1fdca1f6f7de310dab6f4dcd"},
			{"version6.go", "0bafc74725a9ee94939d6cde6bad8e13804aa594328e7b20d7cf05fb86315307"},
			{"version6_test.go", "32ae7065267bd8827123c8801511cac781aa4cf45dee47953e07890e7c4b1109"},
			{"version7.go", "1b154d61434747ca823e5c5c0f58dd3ffc47620badefd0b65626be0010fe6c28"},
			{"LICENSE", "0a8d61ed3cbfd5312326e8126c31ce9c627a283adc99131b56896d29ada04b2d"},
		},
		Imports:     []string{"bytes", "crypto/rand", "encoding/hex", "errors", "fmt", "io", "strings", "sync"},
		PackageName: "uuid", Packages: []string{pkg},
		RequiredTests: []TerminalTest{
			{pkg, "TestTaskVerifyCanonicalFixedVectors"},
			{pkg, "TestTaskVerifyCanonicalEveryBytePosition"},
			{pkg, "TestTaskVerifyCanonicalAlternativeLengths"},
			{pkg, "TestTaskVerifyCanonicalFailureIsZero"},
			{pkg, "TestTaskVerifyLegacyParserCompatibility"},
			{pkg, "TestTaskVerifyCanonicalDeterminismAndEntropy"},
		},
		ContractPath:   "taskverify_contract_test.go",
		ContractSHA256: digest([]byte(uuidCanonicalContractTests)),
		PromptSHA256:   digest([]byte(uuidCanonicalPrompt)),
	}
}

func uuidCanonicalSpec(d Definition) Spec {
	b, _ := json.Marshal(d)
	return Spec{ID: d.ID, Version: d.Version, BaseRevision: d.BaseRevision, SourcePath: d.SourcePath,
		Prompt: uuidCanonicalPrompt,
		Acceptance: []string{
			"Only 36-byte lowercase ASCII hex with the four fixed hyphens is accepted; decoded UUID bytes match independent fixed and standard-library hex vectors",
			"Every rejected input returns a non-nil error and zero UUID, including late malformed bytes; all-zero/all-ones and arbitrary version/variant bit patterns are accepted",
			"Legacy Parse/ParseBytes broad encodings, fixed error categories and existing serialization behavior remain compatible",
			"New parsing is deterministic with no entropy reads, including repeated and concurrent inputs",
			"Each named independently authored contract terminates with pass in the exact pinned module package under offline isolation; candidate assertions are not used",
			"The full pinned upstream Go source/tests plus go.mod and original BSD LICENSE form the 25-file closure; original Google headers stay unchanged, and the upstream tree has no NOTICE",
			"Original uuid.go remains either the exact pinned raw prefix or the exact trusted Go 1.27.1 formatted prefix (only formatter whitespace differences), followed by only ParseCanonical with local-value mutation and trusted pure calls; attribution or legacy edits, calls through function aliases, function literals, pointers/global-array aliases, helpers, concurrency, imports, init and directives are unsupported and remain verifier_unknown even if behavior would be correct",
		}, AcceptanceSourceSHA256: d.ContractSHA256, DefinitionSHA256: digest(b)}
}
