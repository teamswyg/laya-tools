# PDCA54: UUID source and independent verification contract

[한국어](UPSTREAM-54-UUID.ko.md)

This is development evidence for turning the public candidate `go53-uuid-canonical-parse` into a bounded, independently checkable code task. It does not record a model solving the task or establish its difficulty. The original candidate in the [120-source investigation](public-go-acquisition-53.en.md) is not reclassified as a training answer.

## Original behavior and the new task

The source is [google/uuid at the pinned revision](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d). The original [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) accepts hyphenated text, uppercase, raw 32-digit hexadecimal and URNs through `Parse` and `ParseBytes`. Its 38-byte case does not validate the surrounding characters, so a wrapper such as `[...]` is accepted. Some late failures can return a partially decoded UUID with an error.

The new task preserves those APIs and adds only `ParseCanonical(s string) (UUID, error)`. It accepts exactly 36 bytes of lowercase ASCII hexadecimal with the four fixed hyphens. Every rejected input returns an error and the zero UUID, regardless of the failure position. Version and variant bits are unrestricted, including the all-zero and all-ones identifiers. The task does not tighten the existing APIs.

## Source and license scope

| Item | Pinned scope |
|---|---|
| Source revision | `2d3c2a9cc518326daf99a383f07c4d3c44317e4d` |
| Files | 23 original Go files, including all seven original test files, plus `go.mod` and `LICENSE`: 25 files |
| Integrity | Every file checked against SHA-256 and its public Git blob identity |
| External dependencies | This code package imports only the standard library |
| License | Original [BSD-3-Clause LICENSE](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/LICENSE) |
| NOTICE | Absent from the complete inspected revision tree; no invented NOTICE |

The copied source preserves the original LICENSE and existing Google attribution. The LICENSE states 2009 and 2014; `uuid.go` states 2018. `time_test.go` and `version6_test.go` have no separate file header. No separate license or copied-origin signal was observed in the selected files. This is not blanket legal clearance for the entire repository, future dependencies or future model releases. Source redistribution must preserve the copyright, conditions and disclaimer; binary redistribution also needs the notices required by those conditions.

The original [go.mod](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/go.mod) contains only the module declaration. Its bytes remain pinned. An isolated execution copy uses a minimal trusted recipe naming `github.com/google/uuid` and Go 1.27.1. The upstream repository is not modified and no new dependency is downloaded.

## Supported implementation shape

This verifier requires either the exact 10,254-byte original `uuid.go` prefix or its separately pinned 10,251-byte normalization by the trusted Go 1.27.1 `gofmt`, followed by only the new function. The upstream file differs from the current formatter in struct whitespace and documentation line breaks/indentation. Only that fixed normalization is allowed; legacy semantics and attribution are unchanged. The function may use local values and loops, the original pure `Parse`/`xtob` helpers, and selected pure standard functions and conversions. Randomness, global writes, global-array aliases, calls through function aliases, function literals, pointers, concurrency, added helpers, and edits to legacy APIs or attribution are outside the supported shape.

Unsupported code remains `verifier_unknown`. A correct general implementation outside this shape is not counted as a model failure. This is a bounded task verifier, not a general effect checker for Go. Local targets are bound to their declaration objects rather than identifier names, rejecting attempts to hide a global write behind an inner declaration with the same name.

## Independent assertions and development controls

Candidate-authored tests are not acceptance evidence. The trusted external test package adds six named assertions alongside the original tests:

1. Exact 16-byte results for zero, maximum, ordinary and unrestricted-bit fixed vectors.
2. Each of 256 byte values at each of 36 positions: 9,216 checks, using authored character rules and standard `encoding/hex` expectations.
3. Rejection of uppercase, URN, raw, wrapped, whitespace, Unicode and alternative lengths.
4. Error plus zero UUID even for late invalid bytes.
5. Existing `Parse`/`ParseBytes` formats and error-category compatibility.
6. Repeated and concurrent results, preservation of the configured randomness reader, and random-pool contents, cursor and enabled state.

Reader instrumentation alone is not claimed to prove absence of every possible entropy path. Direct `crypto/rand.Read` calls are also rejected by the constrained source gate. These checks cover the supported shape and the exercised inputs.

The following development evidence was actually checked with Go 1.27.1 on macOS arm64 in offline isolation:

- The complete pinned original Go test suite recorded 212 terminal passes. This is the build scope selected on the current host, not proof of every platform or external integration.
- The unchanged baseline lacks the new API and fails the independent contract.
- An authored positive control passes the original 212 tests and six independent tests.
- Fourteen wrong implementations within the supported shape each terminate all six independent tests and are rejected for behavior. Compilation failure is not counted as a behavioral detection.
- Twenty-one unsupported controls are separately checked and refused: attribution deletion, direct entropy access, global writes, alias calls, extra helpers and inner-shadow attempts to write global state. They are not difficulty or success-rate labels.
- Integration checks through the central verifier accept the positive control while ignoring broken candidate-owned assertions. A separate reviewer also ran five independent CLI controls: an independently implemented parser was accepted, uppercase-permissive behavior was rejected, and direct entropy access, an inner-shadow global write and copyright-header deletion remained unknown.
- The complete race checks for this task, static checks and frozen prior-spec identity checks passed locally. Published CI and actual model attempts require separate records.

The 9,216 input checks and multiple controls are **not distinct model tasks**. This contract represents one semantic request, with zero model executions. It does not make all 120 candidates execution-, training- or final-evaluation-eligible. Actual model attempts require separately frozen source, contract and execution plans, with separate results.
