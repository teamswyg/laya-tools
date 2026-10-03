# Unknown-preserving Go input and an initial allocation improvement

An opt-in Go reader and conservative existing-projection bridge now support the next five/eight-candidate cohort. Read the [usage and wire contract](../../../pkg/shortclaimdata/COHORT.en.md). The original reader, finite37 labels/results and trainer are unchanged. **New actual cohort observations, labels and training remain0.**

U remains null; ambiguity retains known T/F facts. Partial-U, ambiguity, calibration and wholly ineligible parents remain in the full audit and are withheld from learning columns. Every parent is checked for declared group/family role consistency and common pins before selection. Selected complete known parents retain all candidates and truthful labels at zero weight. Structural acceptance and opaque digest equality establish neither semantic truth, rights, ancestry nor permission to train.

Each synthetic benchmark ran three times for300ms on Go1.27.1/darwin-arm64. Times below are medians from sequential measurements on one environment.

| Operation | Initial time | Revised time | Revised Go allocation |
|---|---:|---:|---:|
| Read five-candidate row | 8.702μs | 8.756μs | 11,736 B /215 allocations |
| Read eight-candidate row | 11.503μs | 11.484μs | 15,250 B /291 allocations |
| Bridge one audit-only parent | 6.018μs | 5.334μs | **0 B /0 allocations** |
| Bridge five-candidate train +eight-candidate validation | 78.753μs | 79.517μs | 63,768 B /592 allocations |

The audit-only path initially allocated one unnecessary4,096-byte acceptable-index array. Allocating it only upon encountering a selected positive yielded0 B/0 allocations in all three revised runs. All-negative parents need no acceptable array either. These sequential timing differences are not a general speedup claim; unchanged-path variation and selected-path costs remain visible.

Zero refers to measured **heap allocation per bridge call**, not zero stack/audit storage, caller-held text, process RSS or native/GPU memory. `project-owned-B` for the13 known rows is13,477 B and counts only Project's returned payload. No model inference, whole-work gain, token or monetary savings was measured.

Both [initial QA](LOCAL-QA.actual.public.v1.json) and [revised QA](LOCAL-QA.actual.public.v2.json) are retained. Each passed36 top-level tests and114 subtests in both default and `panicnil=1` race modes, with no failures/skips. Fifteen test functions are new; the150 records include pre-existing controls. Repository-wide race/vet passed. Five original source/result files match HEAD byte-for-byte, and six new files remained stable during each QA. Synthetic declarations are not new independent Golden samples. Existing synthetic Fit fixtures are distinct from new corpus training.

[Independent source review](REVIEW.public.v1.json) performed no execution. Zero returned feature counters on rejected calls are not an independent feature-call instrumentation result; preselection guard ordering was confirmed from source. The QA files' pending-CI states are historical at-write states; final CI/bot merge is recorded separately in Issue19.

Reproduce with `go test -race ./pkg/shortclaimdata ./internal/claimfit` and `go test -run='^$' -bench='Benchmark(CohortRow|ProjectCohort)$' -benchmem -benchtime=300ms -count=3 ./pkg/shortclaimdata ./internal/claimfit`. Only public code and synthetic declarations are used; the APIs open no input path and train no new model.

[한국어](README.ko.md)
