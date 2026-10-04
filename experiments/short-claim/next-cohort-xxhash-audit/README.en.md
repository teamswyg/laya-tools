# xxhash: one-parent audit for the next cohort

This experiment checks candidate behavior before using it to train a small claim model. It does not train a model or change Codex routing. Similar candidate descriptions can hide different checks on serialized internal state.

There is one parent request: **atomically import bounded xxhash states; require coherent short lanes and zero unused tail, and preserve receiver and input on rejection.** Fifteen inputs × five candidates means 75 trials, not 75 independent training examples.

## Run and verify

Install Go 1.27.1 and run from the repository root:

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/verify.sh
```

No model download or Python is required. The script copies the published upstream sources and observer into temporary workspace, runs Go CPU with `purego` and network dependency access disabled, and compares every output byte with the saved observations. Linux and macOS CI use the same command. Consult the PR's actual checks for CI status.

## Candidate mechanisms

| Candidate | Checks | Actual contract matches |
| --- | --- | --- |
| direct | Explicit size, magic, bound, lane equations and unused tail | 15/15 |
| reconstruct | Public-API short-state reconstruction and full serialization comparison; structural path at total32 | 15/15 |
| legacy | Original Unmarshal size and magic checks | 7/15 |
| roundtrip | Full import/reserialization byte comparison | 10/15 |
| digest | Digest comparison for short states; legacy import at total≥32 | 7/15 |

All75 trials completed without panic or input mutation. The ledger records587 actual API calls and9 API error returns. All34 candidate rejections preserved public serialized state and Sum64. [Separate readback](READBACK.actual.public.v1.json) contains the values and per-mechanism matrix. These are finite contract checks, not learned-model accuracy.

The first attempt emitted no JSON because of its96KiB output cap. Its [failure record](INITIAL-OUTPUT-BOUND-FAILURE.actual.public.v1.json), source and plan remain retained. The second attempt changed output serialization only, keeping candidate bodies, inputs and Wants unchanged; the complete102,643-byte JSON was stored as4,219-byte gzip and fully decoded for verification. Saved-data checks also cover missing statuses, incorrect call totals and deleted observations.

To read the complete observations:

```sh
gzip -dc experiments/short-claim/next-cohort-xxhash-audit/observations.actual.json.gz
```

Every candidate checks a temporary object and commits the imported state. The common observer does not validate lanes, tails or totals on a candidate's behalf. Each trial starts from a fresh seed19 receiver containing `old`. Rejections compare complete before/after Marshal state and Sum64; input mutation is also observed. [Complete observations](observations.actual.json.gz) retain attempted calls, returns, errors and panics by API and phase. Digest values use hex strings to avoid integer precision loss.

Preservation measurements cover public serialized state and digest observations. They do not measure every internal memory byte or hidden unused-tail bytes. `purego` excludes this source's assembly path; it does not imply absence of unsafe code.

## Contract scope

Upstream is [a pinned cespare/xxhash revision](https://github.com/cespare/xxhash/tree/ab37246c889f9db16b606fda1c232d659df9271d). Our experiment requires stronger validation than that API: total≤32, zero unused tail, and lane equations when total<32. Failure to satisfy this additional contract is not an assertion of an upstream defect.

At total32, arbitrary lanes with zero tail are structurally accepted. Original payload/seed recovery and reachability through Write are not established. Seed0, concurrency, aliasing, longer states and independent numerical digest correctness remain out of scope.

## Training separation

Literal inputs and contract Wants are fixed before execution and reviewed by another source-only reader. Authors and reviewers know the mechanisms; this is not blind evidence. [Execution freeze](FREEZE.public.v2.json) and [file hashes](SHA256SUMS) bind the inputs, captions, observer and source closure.

Lineage closure against the old79/37 examples, including helpers and forks, remains unresolved. This parent has no train/validation/calibration assignment or learning labels/weights. Existing37 data and models are unchanged. Structural Reader success cannot establish missing lineage qualifications.

## Licensing and performance interpretation

The three included upstream Go sources retain exact original bytes and their [MIT notice](source/xxhash/LICENSE.txt). The owned observer is Apache-2.0. Permission to run these sources does not settle future model/dataset redistribution rights.

This is a finite behavior audit. It measures no SIMD/GPU execution, learned quality improvement, whole-task acceleration or token savings. Temporary Go compiler artifacts are distinct from retained model binaries; the cumulative research storage limit is preserved.
