# Claim position features preview v1 — EXPERIMENTAL

This optional Go feature module explores a compact position encoding. It adds a
1024-bin signed hash bank to the actual 2048-bin v2 contextual features, for
3072 bins total. It exposes features only. It has no model, predictions,
training, threshold changes, or production integration.

The fixed design was recorded in [DESIGN.lock.json](DESIGN.lock.json) before
extracting the new bank. The base extractor, existing models, CLI and training
remain unchanged. The separate schema is
`utf8-word12-char2345-v2-pos4-fnv1a-signed-logtf-l2-3072-preview-v1`.
This preview does not change or qualify the two matched siblings contract.

## Why this probe exists

The earlier locked engineering probe retained six pairs / twelve original
fictional development reports. Every pair had identical 2048-bin dense
float32 bits, word count and unhashed local-feature multiset. A blind AI reading
in a separate context judged each pair different in at least one claim head; the
author and blind reader differed on 0/36 head states. Those are same-provider AI
hypotheses from separate contexts; exact underlying-model identity and independence
are unestablished. They are not human gold or an accuracy
evaluation. Their agreement does not remove correlated interpretation risk.

All six original sparse first-seen orders differed. Equal dense bits therefore
do not establish identical actual predictions: floating-point accumulation
order can matter. Neither probe called a model. The original private artifacts
are immutable; this public projection pins them by logical artifact name and
digest without publishing private locations.

## Fixed encoding

1. Call `statehintwide.ExtractContextual` for the base. Expose its values,
   float32 bits, activated zero bins, sparse encounter order and word count
   exactly as returned.
2. Lowercase the complete rune sequence with `unicode.ToLower`. For `R > 0`,
   an event starting at rune `s` receives region `floor(4*s/R)` in `[0,3]`.
   Every rune, including whitespace and punctuation, contributes to `R`.
3. Enumerate the same local features as v2: Unicode letter/number word
   unigrams, successive-word bigrams, and contiguous character windows of
   lengths 2–5. A bigram starts at its first word. A window crossing a region
   boundary uses its starting rune. No event is emitted for `R = 0`.
4. FNV-1a-32 hashes ASCII `p`, a raw region byte `0..3`, the original feature
   prefix (`w`, `b`, `2`, `3`, `4`, `5`), then lowercased UTF-8 rune bytes.
   Bigrams have the same zero-byte token separator as v2. The bank uses
   `h % 1024`, with sign `-1` when `h >> 31 != 0`, otherwise `+1`; public
   indices add 2048. All four tags share one physical 1024-bin bank.
5. Sum signed counts, transform each active count with
   `copysign(log1p(abs(count)), count)`, and independently L2-normalize the
   append bank at fixed scale **1**. Preserve v2's float64 sum, float32 scale
   and multiply rounding order. Do not normalize the concatenation again.
   When both banks have nonzero norms, the combined norm is approximately
   `sqrt(2)`, with the original base still approximately unit norm.

The public v2 API exposes sparse values and word count, not event positions.
The append bank therefore reparses input in a separate fixed rune buffer. No
manual semantic labels, regex negation rules or label-selected features enter
the encoding. The implementation has fixed arrays and separate sparse index
and value storage, with no hot-loop maps, string allocations, global mutation
or locks.

## API and limits

```go
var workspace statehintpositionpreview.Workspace
view, err := statehintpositionpreview.Extract(text, &workspace)
// On success, use view.Len(), view.At(i), and view.WordCount().
```

The zero-value workspace is ready to use. Inputs must be complete valid UTF-8,
at most 4096 bytes, with no NUL rune; no truncation occurs. Empty input succeeds
with an empty view and word count zero. Invalid input returns exactly
`statehintwide.ErrInput` and a zero view. Workspace contents after an error
are unspecified; the next successful extraction clears all active bins.

The view borrows its caller-owned workspace and expires on **any next call**
using that workspace, including an unsuccessful call. `At` returns a value;
retain copies when needed. Do not copy a workspace after first use, reuse it
while a view is read, or share it between concurrent callers. Separate caller
workspaces support concurrency without locks. `At` has the same bounds
precondition as indexing a slice.

## Evidence and cost

[results.json](results.json) retains all six unchanged original pairs,
including any failure. All twelve texts preserve exact base bits and sparse
order. All six pairs have different append-bank dense bits: changed normalized
bin counts are **671, 721, 798, 697, 711, 765**. Zero separation failures were
observed on these six pairs. These counts include possible changes caused by
the bank's shared normalization scale, so they are not counts of newly
understood semantic relations.

Feature separation does **not** fix semantics or show accuracy. Four coarse
regions still lose within-region order and may collide. Moving an unrelated
prefix, suffix or punctuation can move region boundaries. Position can encode
formatting shortcuts, and these engineered swaps do not estimate natural
comment prevalence or performance. No new annotations, predictions, learned
weights or references were produced.

The measured workspace is **54,312 bytes**, versus v2 **30,744 bytes**, on the
recorded darwin/arm64 Go toolchain. This is a struct `sizeof`, including fixed
buffers; it is not process RSS or model memory. A hypothetical nine-logit
float32 linear weight array would be `9*3072*4 = 110,592` bytes, versus
`9*2048*4 = 73,728` bytes (**+36,864 bytes / +50%**). This is arithmetic only;
no actual model payload or GPU allocation was measured or created.

The bounded benchmark cycles through exactly these original twelve texts,
one extraction per operation, with CPU setting 2, 200 ms per sub-benchmark and
three repeats. Actual timing/allocation results and source pins are in
[BENCHMARK.json](BENCHMARK.json); the raw Go benchmark output is
[benchmark.txt](benchmark.txt). The extra bank and separate parsing add memory
and latency to an inexpensive baseline. Larger representation capacity needs
to justify those costs in a separate future semantic evaluation; this preview
provides no such qualification.

| Extractor | Median ns/op (three runs) | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Actual v2 | 21,652 | 0 | 0 |
| Position preview | 44,942 | 0 | 0 |

The preview's median extraction latency is about **2.08×** v2 on this bounded
Apple M4 Pro run. The workspace adds **23,568 bytes** per caller.

## Reproduction and public projection

From the repository root with its required Go 1.27.1 or a compatible local
toolchain, and network disabled:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2 go test -race ./pkg/statehintpositionpreview ./experiments/claim-position-features-preview-v1/replay -count=1 -cpu=2
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2 go vet ./pkg/statehintpositionpreview ./experiments/claim-position-features-preview-v1/replay
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2 go run ./experiments/claim-position-features-preview-v1/replay --check
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2 go run ./experiments/claim-position-features-preview-v1/replay
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2 go test ./pkg/statehintpositionpreview -run '^$' -bench '^BenchmarkExtractOriginalTwelve$' -cpu=2 -benchtime=200ms -count=3 -benchmem
```

Replay writes JSON to stdout by default; `--output -` also selects stdout.
Writing a file requires an explicit `--output path`. The `--check` command
recomputes and compares the frozen `results.json` without writing any file or
JSON output, and rejects a simultaneous file-output option. It checks feature
receipts, schemas, aggregates, input/design digests, encoding-source pins and
size arithmetic. Execution-host metadata and the reporting tool's historical
source pin may differ; current tool pins are in CHECKS/PROVENANCE/MANIFEST.
The recorded feature and benchmark results retain their original source pins.

The replay reads only the public cases/design and permitted source pins. It
checks all twelve exact text digests and the locked base source before feature
extraction. Results contain no model operation and no semantic labels. Tests
cover exact base parity, maximum ASCII and multibyte UTF-8, empty and invalid
input, a hand-enumerated rune/hash fixture, signed collisions, log-TF and
independent norms, zero allocations, deterministic reuse/error recovery, and
separate-workspace concurrency.
The replay test also verifies that default stdout, explicit stdout and successful
or failed read-only checks leave recorded evidence byte for byte unchanged;
only an explicit file-output option creates a result file.

The initial reuse test accidentally supplied a 4140-byte mechanical clearing
fixture and correctly hit the input limit. Its count was reduced to 3910 bytes;
the design and twelve original texts were not edited. Focused package tests,
race and vet then passed. [CHECKS.json](CHECKS.json) records the checks.

[cases.json](cases.json) is the exact feature-only public projection: pair/case
IDs, scenario families, original texts and exact digests, with semantic labels
and rationales omitted. [PROVENANCE.json](PROVENANCE.json) pins immutable source
evidence; [EXCLUSIONS.json](EXCLUSIONS.json) retains permanent exclusions.
Every original and same-scenario paraphrase, translation, clause permutation,
negation, quotation/adoption or temporal variant is excluded from final400,
final1200, final2400, **all future training sets and all future selectors**.
These manifests do not mutate a selector or registry. The bundle's
[MANIFEST.json](MANIFEST.json) pins the public files and its bounded size.

Korean: [README.ko.md](README.ko.md).
