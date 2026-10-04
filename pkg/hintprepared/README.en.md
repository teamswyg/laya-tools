# Prepare per-request features once

[한국어](README.ko.md) · [Complete observations](evidence/READBACK.public.v1.json)

This optional Go API reuses features when several small weight models score the
same request/candidates or when scores are recomputed repeatedly. Existing CLI
and default routing policy remain unchanged. Rankings are unverified suggestions
for verification order, never probabilities, verdicts or action authorization.
Every candidate remains. Existing failed models are not activated by default.

Current Prepare shares constructor-local raw token plans and ordered query hash
prefixes between cardinality counting and feature generation.
The [follow-up token-plan comparison](../../experiments/short-claim/next-cohort-chi-audit/constructor-token-plan/README.en.md)
reduced cumulative allocation bytes by about12.9% and allocation counts by421
against the merged scratch constructor. Paired time medians were about1–2.2%
lower, with some slower pairs; general speed improvement remains unproved.
Final owner storage and every feature/score bit remain unchanged, with no pool,
global cache or lock added.

The earlier constructor change counted features first and reused one local buffer.
The [public Chi comparison](../../experiments/short-claim/next-cohort-chi-audit/constructor-scratch/README.en.md)
saved about37% of cumulative Go allocation bytes in paid Prepare+Rank intervals,
while median paired time ratios increased about7.5–7.7%. Allocation counts rose;
final owner storage is unchanged. This passed prewritten memory-first gates and
does not improve speed or model accuracy.

The usage fragment belongs inside a function with the three packages imported.
`rawWeights` is an explicitly loaded RIIDOH01/8192 research weight asset, not the
original Laya ONNX model. Format/dimension validity does not establish feature
compatibility, quality, source clearance or license suitability.

```go
current, err := shortclaim.ValidateInput(input)
if err != nil { return err }
owner, err := hintprepared.Prepare(current.Prepared())
if err != nil { return err }
view, err := hintweights.New(rawWeights)
if err != nil { return err }
ranking, err := owner.Rank(current, view)
if err != nil { return err }
// ranking.Order[:ranking.Count] contains current candidate indices.
// Read IDs from current.Prepared().Candidates[index].ID.
```

No download or Codex registration happens automatically. Prepare preserves fixed
product validation errors. A changed raw query/count/ordered text returns
`ErrMismatch`; callers explicitly rebuild. Baseline fallback on valid input is
also caller policy. ID/provenance-only changes reuse features, while a new View
always recomputes scores. Ties follow current input order.

Owners and Views are immutable; each call returns independent fixed value arrays.
There is no shared result buffer, cache or lock. Editing returned input/result
copies cannot modify owners. Features still has a build dependency on the same
internal Go package as training code, but no training function is called at runtime.

The table below preserves a historical synthetic experiment using Prepare before
the scratch change. These are not current-constructor performance figures.
Observed medians then: Go1.27.1/darwin-arm64, CGO0,
GOMAXPROCS1, 100ms × three repeats per item. Public synthetic texts and numeric
FP32 weights separate construction from reuse. All 45 raw rows are stored losslessly.

| Candidates | Prepare µs / B/op | Reference full Rank µs / B/op | Prepared Rank µs / B/op | BM25 µs / B/op | Lexical µs / B/op |
|---|---:|---:|---:|---:|---:|
| 1 | 8.857 / 8,200 | 10.115 / 4,624 | 0.154 / 0 | 0.784 / 0 | 0.814 / 0 |
| 5 | 38.279 / 32,152 | 37.918 / 19,856 | 0.541 / 0 | 1.269 / 0 | 1.425 / 0 |
| 8 | 60.050 / 50,040 | 58.700 / 30,664 | 0.831 / 0 | 1.652 / 0 | 1.884 / 0 |

Prepared Rank, BM25 and lexical also measured zero allocations/op. One-off use
must pay preparation cost. For five unchanged candidates, timing arithmetic alone
requires roughly 53 uses to amortize preparation versus BM25, but the methods
produce different scores and quality. Reference repeats weight/raw-word checks,
feature extraction and result conversion. Product input validation is untimed
setup for all ranking probes, so differences do not isolate layout/SIMD/cache effects. B/op
is cumulative Go allocation, not retained/peak/RSS/GPU memory. Both reference
coefficients and View/input representations remained live during comparison.

```sh
go test -race -count=1 ./pkg/hintprepared
CGO_ENABLED=0 GOMAXPROCS=1 go test -run '^$' -bench '^BenchmarkPreparedAPI' -benchtime=100ms -count=3 -benchmem ./pkg/hintprepared
gzip -dc pkg/hintprepared/evidence/benchmark.stdout.public.txt.gz
```

Race tests need a C compiler; the runtime API is pure Go. Six actual repository
race tests, 18 named subcases and vet passed. The initial stage benchmark selected
zero items with a wrong name and was not counted as success; its [failure/repair
record](evidence/STAGE-READBACK.public.v1.json) remains.

No learned model file, network, Python or new training is required for these
synthetic checks; actual API calls require an explicitly supplied weight View. These results do
not establish semantic quality, independent2400 Golden performance, whole-task/
LLM/Codex savings or GPU/MPS performance.
