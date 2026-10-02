# Using the caption-reference preparation tool

[한국어](USAGE-REFERENCES-59.ko.md) · [Official result and interpretation](RESULTS-REFERENCES-59.en.md) · [Preregistered plan](PLAN-REFERENCES-59.en.md)

[Collection/failure ledger](collection-ledger-59.json) · [Independent reference review](reference-review-59.json)

`riido-captionref` is a Go tool that helps maintainers and agents **locate original text and evidence for caption review**. It connects original JSON positions, text hashes and Go literal-expression locations. It needs no Python or model weights and runs no candidate implementation, Laya, ranking or training. Building or generating references does not approve caption meaning, licensing or future model-distribution eligibility.

## Read the result and begin review

Most users only need [caption-coverage-59.json](caption-coverage-59.json) and the [result explanation](RESULTS-REFERENCES-59.en.md). There is no need to repeat official generation.

1. Read `generation_counters`: Bind 1/1, AST 3/3, parents 72, captions 216 and contracts 18.
2. Find `request` or candidate `caption` under `references.parents`. Follow its original file, JSON pointer, decoded UTF-8 byte count and SHA. Read the unchanged original text and order.
3. For the same prototype, read historical input scope and `literal_source_reference` under `references.contracts`. Byte offsets start at 0 and exclude the end. Line numbers are physical original lines with inclusive bounds.
4. Follow the [content-review recipe](content-review-recipe-59.json) to separately inspect implementation descriptions, request scope, observable fields and negation/boundaries. Record exact text/source locations and remaining uncertainty. All reviews currently remain pending.

Acceptable indices and candidate checked/failed counts are historical stored truth, not new executions by this tool. Keep all 21 unknowns and unknown-only group 64 unchanged. IDs, groups, truth, literals and review states must not enter scorer features. Faithfully describing an incorrect implementation does not make a caption an acceptable answer.

## Code checks and CI reproduction boundaries

From the repository root, run the tool and stored-reference regression checks with Go 1.27.1.

```sh
GOTOOLCHAIN=go1.27.1 go test -race ./internal/captionref ./cmd/riido-captionref
```

Unit checks use small owned fixtures for AST handling, byte ranges, pins, I/O bounds, exclusive output and failure counters. The [frozen regression test](../../cmd/riido-captionref/frozen_test.go), added after the official result, separately replays and exactly compares the `references` body and `generation_counters` from existing inputs. It performs stored metadata Bind but calls no source API, ranking or model. There are no scores or floating-point tolerances. Historical Darwin envelope/source/input/plan/binary pins are verified separately; current Linux/macOS CI executables need not equal the original Darwin binary.

Local frozen regression and full race/vet passed after the official result. The read-only reviewer in the [independent reference review](reference-review-59.json) made no Generate, Bind or test calls. Repeated checks add no official attempts, independent requests or fresh truth, and remain distinct from collection counts in the [ledger](collection-ledger-59.json). Later regression files are not retroactively part of the nine preregistered support files. Final GitHub CI is pending at this document's writing point and must be checked on its actual run.

## The completed official collection procedure

The following records **the historical procedure that produced the one official attempt**. It does not instruct users to reuse an output path or repeat official59 with a different path. The source commit was `75a9776230f7eaef3294247cc10bb7f13342f102`; the input commit was `69e9ddc51e218da029572e2bb463500dc36cd143`.

```sh
GOTOOLCHAIN=go1.27.1 CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o .cache/captionref59-driver ./cmd/riido-captionref
.cache/captionref59-driver prepare --repo . --source-commit 75a9776230f7eaef3294247cc10bb7f13342f102 --out .cache/captionref59-plan.json
# Freeze the generated plan bytes unchanged at experiments/short-claim/execution-plan-59.json, then create the input commit:
.cache/captionref59-driver references --repo . --input-commit 69e9ddc51e218da029572e2bb463500dc36cd143 --out .cache/captionref59-official/results.json
```

The same Go 1.27.1/CGO 0/trimpath/Darwin arm64 binary served plan preparation and official generation. The existing plan pins its actual executable SHA, so another build or platform cannot replay the historical official CLI run. Executables stay in local private paths, outside Git. A new experiment needs a separate version, preregistered plan, freezes and attempt ledger.

`prepare` makes Bind 0/Generate 0 and checks 20 Git blobs: sources 5, support 9 and inputs 6. `references` first checks 21: sources 5, support 9, inputs 6 and plan 1. It then reserves a new directory and exclusive `results.json` before Bind and AST reference generation. Input/output files are bounded to 1 MiB each, the executable to 64 MiB and Git verification to 5 seconds per file. Output-path reuse protection does not replace a host-wide official-attempt policy.

Failures preserve `state=incomplete`, fixed `failure_code/failure_cause` and operation counters. If result encoding/writing cannot persist the full file, the error diagnostics retain those counters. Do not delete failures or silently retry them as if nothing ran. The [prototype ledger](prototype-ledger-59.json) separately preserves preparation 2/Bind 2 and the first failure alongside the single official attempt.

## Next preparation artifacts

[Source-transfer explanation](SOURCE-TRANSFER-59.en.md) and its [machine-readable references](source-transfer-59.json) contain two scoped semver/glob proposals connected to retained originals and saved experiment 57 observations. New requests, captions, truth and source API calls are 0. The [role recipe](role-recipe-59.json) remains unassigned, retaining train 9/validation 3/calibration 3 floors and permitting transfer 0. Content review, relationship graphs, coverage and membership/seed freezes are next.

Read `training_ready=false`, `no_roles_plan` and `synthetic_single_pipeline` together. This evidence does not demonstrate model utility, token/financial savings, measured memory/CPU/GPU or Korean-input performance. The 2,400 distinct protected-final-request target is separate, not a minimum for every development fit. Existing Laya CI evidence is separate; this tool's own model-call count is 0.
