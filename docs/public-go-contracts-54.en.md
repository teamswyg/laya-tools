# PDCA54: preparing acceptance contracts for two public Go candidates

[한국어](public-go-contracts-54.ko.md) · [Machine-readable record](../benchmarks/training/public-go-contracts-54.json)

**Two requests from the 120 public Go source candidates now have locally verified acceptance contracts.** This is not evidence of a model solving them or a ground-truth difficulty label. A verifier rejects the original source, accepts an independently authored correct change, and rejects known incorrect changes.

This document records development evidence at preparation time. CI-verified public execution support and publication evidence must be checked separately. Local registry and offline verification integration are implemented, but this record does not yet contain the corresponding public CI or merge evidence.

| Category | Count in this record |
| --- | ---: |
| Candidates in the source investigation | 120 |
| Locally verified acceptance contracts among them | 2 |
| Actual model attempts for these two requests | 0 |
| Training runs or training labels for these requests | 0 |
| Requests eligible for protected final evaluation | 0 |

The contracts develop requests already among the 120; the candidate count is not 122. The original [PDCA53 investigation](public-go-acquisition-53.en.md) and [120-candidate inventory](../benchmarks/training/public-go-acquisition-53.json) remain unchanged. That inventory's SHA256 is `e79cf90505fff4c636de99f322ab82001afd22516216cebcd8249b9b3d61cf67`. Separate laya-tools tasks and model pilots are not added to these requests' execution counts.

## 1. Signed ordinals across the full int64 range

`go53-humanize-ordinal64` adds `Ordinal64` to [the pinned dustin/go-humanize source](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e). For example, `-21` becomes `-21st` and `-11` becomes `-11th`. The minimum value `-9223372036854775808` must keep every digit.

The original provides only `Ordinal(int)`. Its negative inputs use `th`, so `Ordinal(-21)` must still return `-21th`; only the new API selects suffixes from the magnitude. The contract checks the callable type `func(int64) string`; it does not distinguish a declared function from a function-valued variable.

The five copied original files are `go.mod`, `ordinals.go`, `ordinals_test.go`, the required `common_test.go` helper, and `LICENSE`. This is a closed ordinal subset, not the entire upstream package. Only `ordinals.go` is mutable.

The [original MIT license](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/LICENSE) is retained unchanged. `number.go`, with its separate WTFPL origin, is neither copied nor built. The actual source has no `NOTICE`, and none was invented. This repository's Apache-2.0 notice does not replace upstream MIT.

The oracle uses arbitrary-precision `math/big` arithmetic and an explicit table of 100 suffix residues. The authored correct control uses decimal string tails, so it does not copy the oracle implementation. Checks cover both signs, the `11/12/13` exceptions, every suffix residue, inclusive int64 endpoints, 32-bit and floating-point precision boundaries, and fixed wide inputs.

Actual offline verification rejects the baseline and accepts the correct control with six required terminal passes covering original and independent tests. Thirteen incorrect controls are rejected. They include a comment-only change with no new API and a wrong-signature compilation failure, so not all thirteen are counted as behavioral detections. An independent reviewer accepts a separate safe unsigned-arithmetic implementation and rejects two controls with incorrect large teen or minimum-value suffixes. Deliberately broken candidate-authored tests are never acceptance evidence.

## 2. An exact lowercase UUID text parser

`go53-uuid-canonical-parse` adds `ParseCanonical` to [the pinned google/uuid source](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d). It accepts exactly 36 bytes of lowercase ASCII hex and the four fixed hyphens. Every failure returns an error and a zero UUID, including failure after a valid prefix. Zero, maximum and arbitrary version/variant bit patterns remain valid.

Existing `Parse` and `ParseBytes` retain uppercase, URN, raw 32-digit hex, wrapper compatibility and error categories. The brief request in the source investigation is preserved; the development contract separately refines compatibility and the supported implementation shape in its own prompt.

The frozen closure has 23 original Go files plus `go.mod` and the [BSD-3-Clause LICENSE](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/LICENSE): 25 files. Original Google attribution remains intact, and the actual tree has no `NOTICE`. Source redistribution must preserve the copyright, conditions and disclaimer; binary redistribution must provide the notices required by the license. This is not blanket clearance for other files, dependencies or future model releases.

The verifier supports the exact 10,254-byte original `uuid.go` prefix or its fixed 10,251-byte Go1.27.1 formatter result, followed by only one new function. Normalization permits only the verified formatter whitespace differences. Supported code uses local values, loops, the original pure `Parse`/`xtob` helpers and selected pure standard calls. Global writes, direct entropy reads, calls through function aliases, pointers/global-array aliases, added helpers, function literals, concurrency, import/init/directive changes and attribution edits are unsupported. Declaration objects distinguish an inner local declaration from a subsequent write to actual global state.

**Unsupported code is `verifier_unknown`, even if its behavior is correct. It must not become a model failure or low-capability label.** This is not a general effect checker for Go programs.

The original test suite records 212 terminal passes on the current host. The baseline lacks the new API and fails the independent contract. The correct control passes the original 212 plus six independent tests: 218 total. Fourteen incorrect implementations within the supported shape each terminate all six independent tests and are rejected for behavior. Twenty-one unsupported controls are assessed separately by the shape gate.

Independent CLI review accepts a separate `hex.DecodeString` implementation with 218 passes and rejects an uppercase-permissive implementation. Direct entropy access, hidden global pool writes and deleted Google attribution become unknown before execution. Fixed vectors, the 36-position × 256-byte matrix, repeated/concurrent calls, randomness-reader and pool-state checks are controls for one request. The 9,216 input checks and 218 terminal passes are not separate tasks. Reader instrumentation alone does not prove absence of every entropy path; direct calls are also blocked by the constrained source gate.

See the [separate UUID source and contract record](UPSTREAM-54-UUID.en.md) for details.

## Checking the contracts and their identities

In the development checkout, build the executable once, then read the specs without launching a model:

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
.cache/bin/riido-taskverify --task go53-humanize-ordinal64 --spec
.cache/bin/riido-taskverify --task go53-uuid-canonical-parse --spec
```

To verify a candidate, supply the pinned public source and its candidate copy. A trusted, working Go1.27.1 installation is required. Recorded executions use macOS arm64; they do not establish every platform.

```sh
.cache/bin/riido-taskverify --task TASK --base-dir PUBLIC_PINNED_BASE --candidate-dir PUBLIC_CANDIDATE_COPY --go-root TRUSTED_GO_1_27_1
```

The executable above returns exit 0 for accepted, 1 for rejected, 2 for invalid input and 3 for verification unavailable. Original module bytes remain pinned; only the temporary compiler module uses the original module identity and Go1.27.1. Execution fetches no external dependency and inherits no authentication, user home or network access. Test elapsed time is not model latency or cost.

| Frozen identity | Humanize | UUID |
| --- | --- | --- |
| Spec SHA256 | `5c6fd6058a48bd0e6441a68c351e7b1ed491aec20baff78c6942e377ae79f5d5` | `c8e3a2edcc2024ce68ba7ea7b5fed00554b2e1c36a12174f942f32b5a5a5eec3` |
| Definition SHA256 | `99a7c0723e3c9f40a069003ccf3f9c2e290c3cbe8e67be3a4ced02d4a31a29b8` | `9413ed54a10ee9a1576f71a68cf664134570cf84e735f8daeb88c66c1f4ad7c8` |
| Contract SHA256 | `3b1351162b746fa61f7837af2923644eb2f4df100de643114c9d9f0eb064fd06` | `259f3fcd56ac9bf0d4d307cc1b8ed77968b6b3bf21327160e09efd440f71ffd5` |
| Prompt SHA256 | `107ae4d6da023114cc615ebdef6709053be8471bb0ee37f9763af73f279a94a9` | `5c6456db8da12a5e9510cf9f39a448298df4dd73bead5706a41288f567f30cb2` |

Per-file SHA256, Git blobs and implementation references are in the separate JSON record. Contract tests are finite development evidence, not formal proofs over all int64 values, programs or environments. License review of these copied source scopes is not extended to model training, weights or dataset distribution eligibility.

## Work remaining for the 2,400-request target

The target is at least 2,400 **distinct protected-final requests per domain**. The 120 source candidates and two development contracts are neither an answer set nor achieved final coverage. Material already read in development is excluded from protected final evaluation.

Public CI evidence and a fresh model-attempt plan must first be frozen. Actual whole-request model execution, independent completion checks, usage and failure status are then needed before considering capability labels. Cost and the best routing profile remain unknown. Repositories, copied origins, shared code and parent/child/sibling requests belong to connected groups; train, selection, calibration and final material require separate acquisition. Translations, repeated attempts and residue/byte controls stay with their original request group.

Model upgrades/downgrades, repository selection, task decomposition and small nonbinding hints each need their own requests and oracles. These two Go acceptance contracts do not substitute for those domains' training labels.
