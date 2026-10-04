# Verifying decimal-to-integer claims

[한국어](README.ko.md) · [Optional Go API](../../../pkg/hintprepared/README.en.md)

Tiny hints must distinguish similar descriptions to suggest a useful verification order. This development audit executes pinned public decimal code to compare rejecting fractions first, checking range first, checked truncation, and unchecked IntPart. It is neither generative-model training nor a new model-quality evaluation.

The request is to **return non_integer for a fraction before checking bounds, return out_of_range for integral int64 overflow, otherwise return the exact integer**, preserving the input coefficient and exponent. These error names belong to the owned wrappers, separately from upstream parse errors.

| Mechanism | Literal Wants passed |
|---|---:|
| strict_fraction_first | 8/8 |
| strict_range_first | 7/8 |
| truncate_checked | 5/8 |
| intpart_unchecked | 3/8 |

`9223372036854775808.1` is both fractional and beyond range; the requested error is non_integer. Checking bounds first instead returns out_of_range. `1.00` is an integer: it must return1 while retaining coefficient100 and exponent−2.

All32 actual trials returned normally. All298 explicitly wrapped API calls returned, with zero API errors or panics. The13 deliberate wrapper rejections are separate from API failures. Before/after coefficient and exponent values were preserved. See the [full comparison](evidence/COMPARISON.actual.public.v1.json) and [complete observations](evidence/OBSERVATIONS.actual.public.v1.json.gz):74,491 raw JSON bytes retained as2,253 gzip bytes, with CRC/EOF/UTF-8/full JSON/count checks.

## Reproduce

With Go1.27.1, run from the repository root:

```sh
go run ./experiments/short-claim/next-cohort-decimal-audit/replay
```

The verifier pins source/input/notices, builds actual decimal code in an owned temporary module, compares complete observations and literal Wants, and removes temporary files. Use `-go`/`-root` for another executable or packet location. No model, authentication or training is required. The [actual public Go replay](replay/EXECUTION.actual.public.v1.json) passed locally on macOS and left no owned temporary directories. Success means reproducing all observations, including failing candidates; it does not mean every candidate meets the request or model recommendation quality improved. Check the PR for Linux/macOS CI status.

## Scope and next work

The ledger covers explicitly wrapped decimal/big.Int calls. Package startup can parse ln10 and initialize floating-point constants; hidden/startup calls are excluded from298. Only the eight exact frozen literals are accepted. A short scientific notation can still request expensive scaling, so arbitrary inputs/exponent sweeps are outside this experiment.

[Execution resource reports](evidence/EXECUTION.actual.public.v1.json) cover one complete child, including startup,32 conversions, snapshots and JSON/gzip output. They do not measure individual conversions or model cost. Darwin raw units were not converted; GPU and whole-system memory were not measured. GOMEMLIMIT is a soft Go target, not an RSS limit.

This is one exposed development request. Semantic overlap with earlier integer-conversion tasks and family relationships remain under review: zero added independent Golden parents, labels, roles or Fits. No model/threshold was tuned from these results. Thirty-two trials/four candidates do not become independent samples. Model utility, independent2400 evaluation and whole-task/LLM savings remain unproven.

Next, qualify materially different requests and related-source groups, then compare fixed order, BM25, lexical and model hints using the same actual verification cost. Retain all candidates and record failures/fallbacks. Do not fit confidence or early-stop cutoffs to this exposed case.

## Sources and notices

The source is [shopspring/decimal at the pinned revision](https://github.com/shopspring/decimal/tree/ca4740823783f3bc026235a5ed3515aca626da92). Actual compiler selection records121 packages/843 source files. Require v1.0.0 is a local-replacement graph sentinel, not an upstream-release assertion.

Retain the [complete Decimal/fpd MIT notices](source/decimal/LICENSE.txt) and Go2009 headers. The [complete historical Go BSD notice](source/GO-BSD-NOTICE.txt) is a pinned notice comparator; the actual copied predecessor and extent remain unknown. The owned observer is Apache-2.0; no upstream endorsement is claimed. Finite licensed audit publication and whole-family training admission are separate.

Pending text in the [freeze](evidence/FREEZE.public.v1.json) reflects its creation time; completed outcomes are in EXECUTION/COMPARISON/CLEANUP. Private controllers, host paths, raw stderr and model bodies are withheld. [PACKET](PACKET.public.v1.json) and SHA256SUMS connect public aliases and complete file hashes.
