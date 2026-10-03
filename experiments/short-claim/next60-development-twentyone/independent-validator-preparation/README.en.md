# Independent preparation for validating twenty-one development rows

This package contains validator source and synthetic controls only. It has not loaded actual data21 or executed the public Go reader. Root will materialize and pin the prospective data, review the source, then separately run the bridge.

Every byte of the previous sixteen-row file, including its final LF, must be an exact prefix. Five appended rows independently bind frozen Wanted request/caption text, fixture Wanted/IDs, Root adoption and the saved comparison report. Actual reader getters are compared with independently parsed request/candidate IDs/raw text/labels/weights/metadata/counts and unused slots. Valid() alone is insufficient. Prepared() inspects already loaded values; no Features, Score, Project or Fit method is called.

The five candidate counts are 3/2/3/3/3. The unavailable error-writer baseline is excluded. Its remaining reference has selected position zero but original source position one, preserving metadata_id=reference-design. A negative may retain unknown predicates alongside its known counterexample; unknown positions must remain identical. A positive requires every required predicate known true. Root has already frozen adoption; this validator assigns no new labels.

Expected success counts are 21 requests, 61 labels, 21 positive, 40 negative and 103 inputs. Row finite_scope counts **301 observations of selected candidates**. Full first-execution evidence contains **305 observations**, including the excluded baseline. The new five contribute 65 selected versus 69 original observations. These units are kept separate.

An always-select-position-one parent control yields 18/21, approximately 85.7%, not 19/21 because the error-writer positive moves to selected position zero. An always-negative label control yields 40/61. Neither is model accuracy. baseline/reference/wrong-seed IDs and caption style may also reveal the target. This preparation changes no candidate order or text. Permutation, ID exclusion and style controls require separate future experiments.

Each input is bounded to 1MiB, each JSONL row to 16KiB and output to 64KiB. Inputs must be regular files with exact size/hash bindings. Output uses a fresh absolute path, O_EXCL, and checked Write/Sync/Close. Reports contain no supplied paths or error bodies. Filesystem Sync calls establish no power-loss or storage-hardware guarantee.

Exactly ten unique `--name value` options are accepted; equals syntax, duplicate flags and positional arguments fail closed:

```
--data DATA21 --data-sha256 ROOT_FROZEN_DATA21_SHA
--previous-data DATA16 --fixtures FROZEN_FIXTURES
--wants FROZEN_WANTS --adoption ROOT_QUALIFICATION_V2
--adoption-sha256 3c1468970a99bc494f3674c6fcbaf476f485e3d491a850ab7aebb09abadd3bfd
--finite-comparison SAVED_FINITE_REPORT --saved-results SAVED_WORKER_RESULTS
--output FRESH_ABSOLUTE_REPORT_PATH
```

Place module/validate.go and source/reader-main.go.txt in a separate reviewed directory inside the project module. Preparation tests use only the standard library and a declared fake reader; they never compile the bridge. Synthetic records are not actual original/candidate/Go-reader/model evidence.

Initial formatter path errors and a test build failure from an unused variable are preserved. After correction, Go 1.27.1 race testing passed twelve top-level plus ten subtests, and vet passed. A later control comparing exact expected-success JSON bytes produced a final passing run of thirteen top-level plus ten subtests and another passing vet run. Both successful runs and the initial failed build are retained. Go race controls use CGO=1; vet uses CGO=0. Neither imports original packages. Bridge build/execution, actual LoadDevelopmentRow, model/training/features/projection/scoring, Git/HF/network operations all remain zero.

The reviewer did not author Wanted, candidates, observer, Root adoption or the data generator, but knows earlier source review and actual saved comparison, so this is nonblind. The reviewer authored the validator. No general correctness, new model quality, token savings, license guarantee or new CI success is claimed.

[Plan](PLAN.v1.json) · [Expected successful JSON](EXPECTED-OUTPUT.v1.json) · [Position/style control review](LEAKAGE-REVIEW.v1.json)
