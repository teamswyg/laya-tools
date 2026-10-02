# 78: The cost of smaller weight representations

The smaller weight table reduced stored payload, but repeated scoring in this experiment was slower than the reference FP64 array. v1 took 4.08–5.10 times as long; v2, which moved format selection outside the feature loop, still took 1.21–2.68 times as long. The [public Go package hintweights](../../../pkg/hintweights/README.en.md) is therefore an option developers explicitly call. Existing models and default Decode/Score paths remain unchanged.

The target is a **small 8,192-dimensional head in RIIDOH01 format**. It reads FP32, INT8, or ternary coefficients and scores already-prepared features. This experiment did not measure Laya's complete text encoder, a newly trained model, or GPU inference. Ternary storage does not imply native 1.58-bit arithmetic.

## What was measured

The original benchmark ran once per version. Four formats of public, project-authored synthetic inputs, construction/scoring, two implementations, and five repetitions produced **80 rows per invocation**. Scoring used the same 266 ordered feature entries, preserving duplicates; construction and preparation were outside score timing. Neither 80 rows nor repetitions count as independent training problems.

The target was Apple M4 Pro, Go 1.27.1 darwin/arm64, CGO 0, and trimpath. CPU was set to 1 and the Go heap soft limit to 256 MiB, which is not a hard RSS limit. Saved records show Start 1, joined Wait, exit 0, retries 0, and no timeout or overflow for each invocation. They used fixed legacy→packed order and repeated warm loops. Startup state and background load were not separately controlled.

**There have been 0 benchmark executions after porting to the public pkg/hintweights namespace.** The numbers below belong to the private prototype kernel. The port ran public synthetic race/vet/build and bit/ownership checks. This archive does not itself claim that the parent repository's full CI has completed.

## Scoring is still slower

Each median uses five samples within that version's invocation. The ratio is packed/reference time; values above 1 mean slower. All score samples report 0 B/op and 0 allocs/op for both implementations.

| Format | v1 reference → packed ns/op | v1 ratio | v2 reference → packed ns/op | v2 ratio |
|---|---:|---:|---:|---:|
| FP32 | 155.8 → 635.9 | 4.08× | 155.9 → 213.1 | 1.37× |
| INT8 | 156.7 → 680.6 | 4.34× | 156.4 → 189.6 | 1.21× |
| Ternary, 13 nonzeros | 155.8 → 794.1 | 5.10× | 155.9 → 248.9 | 1.60× |
| Ternary, 8,192 nonzeros | 155.7 → 776.2 | 4.99× | 156.1 → 418.8 | 2.68× |

v2 replaces per-feature format dispatch with three serial Score loops. v1 and v2 ran in separate invocations, so their difference is not causal proof that the change alone produced the improvement or that it transfers to another CPU. Every sample and original hash is preserved in [v1 metrics](v1/saved-qa/METRICS.v1.json) and [v2 metrics](v2/saved-qa/METRICS.v2.json).

## Payload, construction allocations, and process memory differ

The reference table owns 8,192 FP64 values: 65,536 B of payload. Packed payload is the length of its owned wire bytes plus ternary prefix data. Constructor B/op is the **observed cumulative allocated bytes per operation**, not live heap or RSS. The following B/op and allocation counts were the same in both versions.

| Format | Packed owned payload B | Constructor reference → packed B/op | Constructor reference → packed allocs/op |
|---|---:|---:|---:|
| FP32 | 32,792 | 65,536 → 41,024 | 1 → 2 |
| INT8 | 8,216 | 65,536 → 9,536 | 1 → 2 |
| Ternary, 13 nonzeros | 1,308 | 65,536 → 1,504 | 1 → 3 |
| Ternary, 8,192 nonzeros | 2,330 | 65,536 → 2,656 | 1 → 3 |

Ternary payload is `24 + 1,024 + ceil(nonzero/8) + 258` B. Counts exclude struct/slice headers, allocator padding, original input, and feature arrays. Original input and New's owned copy may coexist. Construction was faster on these inputs, but that is not a benefit for repeated scoring of an already-prepared table.

| Version | Whole-process peak RSS B | Peak footprint B | OS real / user / system seconds |
|---|---:|---:|---:|
| v1 | 11,173,888 | 8,520,184 | 19.76 / 18.73 / 0.24 |
| v2 | 11,288,576 | 8,684,000 | 19.47 / 18.55 / 0.25 |

This RSS covers the **combined process containing both representations, setup, and the Go testing runtime**. There were no representation-specific processes, so it does not establish real-use RAM savings. Controller wall time was 19.764369208 seconds for v1 and 19.47491875 seconds for v2. These are not individual constructor memory or cold-start measurements.

## The calling contract

Callers must keep input bytes stable while New reads them. After return, changing the original input cannot change the View's owned copy. Value copies of a View share immutable storage and support concurrent readers. Feature arrays must also remain stable while a call reads them.

Score preserves feature order and duplicates during FP64 coefficient conversion, multiplication, and addition. It does not move scale after accumulation or skip zero multiplication, preserving `0*Inf`, `0*NaN`, and signed-zero behavior. Synthetic checks found coefficient/score bits equal to the reference. This is not a universal proof across every input or compiler. Port checks also cover nil/zero View, invalid indices/model bytes, and sign-byte boundaries 7/8/9. Read the [API guide](../../../pkg/hintweights/README.en.md) first.

This experiment establishes neither SIMD/SoA acceleration, model quality, training eligibility, production/default qualification, nor lower LLM/Codex usage or charges. Whether the storage/CPU tradeoff is useful depends on the actual caller and workload.

## Reading originals and failure history

- The [v1 preparation ledger](v1/preparation/ATTEMPT-LEDGER.v1.json) preserves the first race failure from an invalid synthetic fixture, the vet failure from unkeyed test literals, and their corrections. Production v1 View was not changed to fix them.
- The [v1 source review](v1/independent-review/RECEIPT.v1.json), [v1 saved-result review](v1/saved-qa/RECEIPT.v1.json), and [v2 saved-result review](v2/saved-qa/RECEIPT.v2.json) describe another reader's scope. These are nonblind AI-assisted reviews with prior exposure, not human blind evaluations or new benchmark executions.
- The [v2 preparation ledger](v2/preparation/ATTEMPT-LEDGER.v2.json) preserves a filename assumption and correction of a wrong v1 handoff pin before sealing. `history/ATTEMPT-LEDGER.pre-handoff-pin-fix.v2.json` is **superseded history**, not the current pin authority. Helper compile/edit failures remain in the [v2 QA ledger](v2/saved-qa/ATTEMPT-LEDGER.v2.json).
- The [public-port handoff](public-port/HANDOFF.v1.json) and [ledger](public-port/ATTEMPT-LEDGER.v1.json) are separate. They record 5 top tests/46 subtests, successful race/vet/build, and 0 benchmarks after porting.

Preparation records showing `benchmark=0` describe **their authoring time**. The subsequent single root invocations are separately preserved in the [v1 plan](v1/root/plan.v1.json)/[actual ledger](v1/root/ROOT-ACTUAL-LEDGER.v1.json) and [v2 plan](v2/root/plan.v2.json)/[actual ledger](v2/root/ROOT-ACTUAL-LEDGER.v1.json). Root ledger schemas retain `76` because a generic controller was reused; their bound plan, binary, and raw hashes identify these experiment 78 invocations.

[COPY-LEDGER](COPY-LEDGER.v1.json) records byte-exact copies of 40 safe originals and hashes/exclusion reasons for 17 private files. Go source is stored inert as `.go.txt`. Private go.mod, binaries, raw benchmark/OS/test logs, host-path helpers, and disassembly are omitted. These new READMEs and the [archive manifest](ARCHIVE-MANIFEST.v1.json) are separately authored; copied originals were not rewritten. Archiving adds no benchmark, model, or training execution.

The archive verifier's first metadata execution failed because it checked the manifest link before creating that file. The manifest records the failed source/ledger hashes and the correction to check only its own links after creation. Copying succeeded once; final verification had 2 attempts, with 1 failure and 1 success. The original experiment was not rerun.

The reference is our public Apache-2.0 [hintlearn model](https://github.com/teamswyg/laya-tools/blob/7cea49090727555924bf22679ffeccc4f38c9865/internal/hintlearn/model.go) and [scorer](https://github.com/teamswyg/laya-tools/blob/7cea49090727555924bf22679ffeccc4f38c9865/internal/hintlearn/learn.go). The new library and our records follow the repository [LICENSE](../../../LICENSE). This archive contains no model body, trained weights, or full external documents.
