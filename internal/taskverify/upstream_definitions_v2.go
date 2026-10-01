package taskverify

import "encoding/json"

const humanizeOrdinalTaskV2 = "go55-humanize-ordinal64-v2"
const uuidCanonicalTaskV2 = "go55-uuid-canonical-parse-v2"

const humanizeOrdinalPromptV2 = `Implement the existing public request: add Ordinal64 choosing the English suffix from the magnitude while preserving the sign across the full int64 range.

Required behavior:
- Expose Ordinal64 with the exact callable type func(int64) string. For every input from -9223372036854775808 through 9223372036854775807 inclusive, return its exact signed base-10 integer digits followed by the suffix. Preserve the minus sign, avoid overflow at MinInt64, and do not truncate through int, int32 or floating point.
- Select the suffix from the magnitude: residues 11, 12 and 13 modulo 100 use th; otherwise final digits 1, 2 and 3 use st, nd and rd; all other final digits, including zero, use th. Examples: 0th, 1st, 2nd, 3rd, 11th, 112th, -1st, -12th, -123rd, -9223372036854775808th, 9223372036854775807th.
- Preserve the existing Ordinal(int) API and its original behavior. In particular, its negative inputs continue to use th. The new API must produce deterministic results across repeated and varying inputs.

Source and supported execution scope:
- Only ordinals.go is mutable. The original go.mod, common_test.go, ordinals_test.go and MIT LICENSE bytes are pinned to public revision a1b4e66b9a6d890e9e15e7091cf16c8032367d6e of github.com/dustin/go-humanize. Keep the module and attribution unchanged. Candidate-authored test changes never determine acceptance; the verifier uses the original tests and its own independent contract. Files outside this declared closure are unassessed.
- The verifier validates the original LICENSE separately and copies those fixed original bytes into its isolated test directory. Candidate LICENSE contents are outside candidate source acceptance and are not assessed by an accepted label; acceptance does not certify candidate or downstream distribution attribution compliance.
- Keep package humanize and the original single, unaliased import strconv. Changed imports, an init function, or added go: directive comments are outside the supported shape and yield verifier_unknown rather than a behavioral failure. This task does not impose the UUID task's append-only or pure-call checker.
- Format the complete ordinals.go with the trusted Go 1.27.1 gofmt. It must compile using the original module's Go language version 1.21. The compiler/toolchain executable is go1.27.1; its version does not raise the module language version. Do not change go.mod or add dependencies, module configuration or a toolchain directive. Integer range syntax introduced in Go 1.22 is therefore unavailable.
- Under offline isolation, the original TestOrdinals and these independent tests must terminate with pass in exactly github.com/dustin/go-humanize: TestTaskVerifyOrdinal64SignatureAndExamples, TestTaskVerifyOrdinal64EverySuffixResidue, TestTaskVerifyOrdinal64FullRangeDigits, TestTaskVerifyOrdinal64DeterministicRange, TestTaskVerifyOriginalIntOrdinal.

This is a clarified evaluation version of go53-humanize-ordinal64, not a new logical request. Only this public development closure is assessed; broader APIs, arbitrary implementation shapes, model success, training eligibility and protected-final eligibility are not implied.`

const uuidCanonicalPromptV2 = `Implement the existing public request by appending exactly func ParseCanonical(s string) (UUID, error) to uuid.go. Keep the parameter name s, unnamed results, no receiver and no type parameters.

Required behavior:
- Accept exactly 36 bytes: 32 lowercase ASCII hexadecimal digits and hyphens only at byte offsets 8, 13, 18 and 23. Return the exact decoded 16 bytes. Accept the all-zero and all-ones identifiers and arbitrary version and variant bits.
- Every other input, including uppercase, URN, raw, bracketed/wrapped forms, surrounding whitespace, malformed separators, invalid hex, non-ASCII bytes and every other byte length, returns a non-nil error AND the zero UUID. A valid prefix followed by an invalid byte must not leak a partial result.
- Preserve existing Parse and ParseBytes behavior, their broad accepted formats and error categories, and existing serialization behavior. The new parser must be deterministic for repeated and concurrent inputs and must not read entropy or change package-global/random-pool state.

Source and exact supported shape:
- Only uuid.go is mutable. All other files in the 25-file public closure, including original Go tests, go.mod and BSD-3-Clause LICENSE, remain pinned to revision 2d3c2a9cc518326daf99a383f07c4d3c44317e4d of github.com/google/uuid. Preserve the original Google copyright headers. Upstream has no NOTICE; do not invent one. Candidate-authored tests never determine acceptance. Files outside the declared closure are unassessed.
- The verifier validates the original LICENSE separately and copies those fixed original bytes into its isolated test directory. Candidate LICENSE contents are outside candidate source acceptance and are not assessed by an accepted label; acceptance does not certify candidate or downstream distribution attribution compliance. The original copyright headers inside the assessed uuid.go prefix remain subject to the strict source-shape check.
- Preserve the original uuid.go as the exact 10254-byte prefix with SHA-256 00d93a3f65d8b0063266e63ff0e16ac543e30a911fd35119bcf1a66437161360, or its exact trusted Go 1.27.1 gofmt-normalized 10251-byte prefix with SHA-256 b19f8aad57fbe17219758f0142742fb3e1879d26506f7fce99169b3ec770c7e8. Append only the one ParseCanonical function; comments and whitespace are allowed. Format the complete file with trusted Go 1.27.1 gofmt.
- Keep package uuid and exactly these existing unaliased imports: bytes, crypto/rand, encoding/hex, errors, fmt, io, strings, sync. The new function may use local values, local arrays/slices, ordinary loops and conditionals. Writes, indexing/slicing targets, copy/append destinations and hex.Decode destinations must resolve to function-local values (or blank assignment). Do not alias or mutate a global array such as Nil. A copied global value may be inspected without mutation.
- The only allowed direct calls are Parse, xtob, len, cap, copy, append, make and these type conversions: UUID, byte, rune, int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64, string, bool.
- The complete allowed package-call list is strings.ToLower, strings.ToUpper, strings.TrimSpace, strings.TrimPrefix, strings.TrimSuffix, strings.Trim, strings.ReplaceAll, strings.Replace, strings.EqualFold, strings.HasPrefix, strings.HasSuffix, strings.Index, strings.IndexByte, strings.Contains, strings.Count, strings.Repeat; hex.Decode, hex.DecodeString, hex.EncodeToString, hex.DecodedLen; bytes.Equal, bytes.Index, bytes.IndexByte, bytes.Contains, bytes.Count, bytes.HasPrefix, bytes.HasSuffix; fmt.Errorf, fmt.Sprintf; errors.New. The only allowed methods on a local identifier, Nil or Max are String and Variant. This list describes support, not permission to normalize and accept invalid input.
- Do not declare local names shadowing these reserved names: nil, true, false, UUID, error, byte, rune, int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64, string, bool, Nil, Max, ErrInvalidUUIDFormat, RFC4122, Parse, xtob, len, cap, copy, append, make, strings, bytes, hex, fmt, errors. The blank identifier is allowed. Only locals and these reserved identifiers are supported.
- Calls through function aliases (including otherwise pure helpers), function literals/types, extra helper declarations, local type declarations, pointers/address-of/dereference, type assertions/interfaces, goroutines, defer, channels/sends/receives, global-state writes, changed imports, init functions and added go: directive comments are unsupported. They yield verifier_unknown even if the output would otherwise be correct. An unused alias declaration is not itself a prohibited call. This is a deliberately bounded checker, not a general effect proof.
- Compile with the unchanged original go.mod. It has no go directive, so the module language version is implicitly 1.16; the trusted compiler executable is go1.27.1. Do not add a directive, dependencies, module configuration or a toolchain directive. Newer language syntax, including generics, predeclared any/min/max and integer range, is unavailable.
- Under offline isolation, the complete pinned original Go test suite and all six independent tests must terminate with pass in exactly github.com/google/uuid: TestTaskVerifyCanonicalFixedVectors, TestTaskVerifyCanonicalEveryBytePosition, TestTaskVerifyCanonicalAlternativeLengths, TestTaskVerifyCanonicalFailureIsZero, TestTaskVerifyLegacyParserCompatibility, TestTaskVerifyCanonicalDeterminismAndEntropy.

This is a clarified evaluation version of go53-uuid-canonical-parse, not a new logical request. Unsupported shapes are unknown rather than model-failure labels. Finite development controls do not establish general model ability, training eligibility or protected-final eligibility.`

func upstreamDefinitionV2(id string) Definition {
	var d Definition
	var prompt string
	if id == humanizeOrdinalTaskV2 {
		d, prompt = humanizeOrdinalDefinition(), humanizeOrdinalPromptV2
	} else {
		d, prompt = uuidCanonicalDefinition(), uuidCanonicalPromptV2
	}
	d.LogicalTaskID, d.PreviousTaskID = d.ID, d.ID
	d.ID, d.Version = id, 2
	d.PromptSHA256 = digest([]byte(prompt))
	recipe, err := TaskEvaluationRecipe(id)
	if err == nil {
		d.EvaluationRecipeSHA256 = recipe.RecipeSHA256
	}
	return d
}

func upstreamSpecV2(d Definition) Spec {
	var s Spec
	if d.ID == humanizeOrdinalTaskV2 {
		s = humanizeOrdinalSpec(d)
		s.Prompt = humanizeOrdinalPromptV2
	} else {
		s = uuidCanonicalSpec(d)
		s.Prompt = uuidCanonicalPromptV2
	}
	s.LogicalTaskID, s.PreviousTaskID = d.LogicalTaskID, d.PreviousTaskID
	s.EvaluationRecipeSHA256 = d.EvaluationRecipeSHA256
	s.Acceptance = append(s.Acceptance, "Full requirements and supported shape are disclosed in Prompt; evaluation uses the pinned dependency-free upstream go.mod language with the trusted go1.27.1 toolchain")
	b, _ := json.Marshal(d)
	s.DefinitionSHA256 = digest(b)
	return s
}
