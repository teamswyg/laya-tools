# PDCA after the 79-request experiment: broaden qualified data first

The recommendation remains to test the tiny model as a claim or ordering hint, and stop rerunning the same 79-request collection with changed seeds, epochs or losses. This fixed experiment did not establish useful improvement. It does not establish that every small model or ternary training approach is impossible.

## Observed results

Three training-only requests and nine labels were appended to the original 76, followed by one fit with the first FP32 recipe. The old validation contains ten answerable and five no-answer requests; five unknown requests are excluded. The strongest deterministic control requires 27 simulated checks with Top-1 5/10. The new model requires 31 with 2/10, matching the first fit 71 totals. Validation BCE moves from 0.6657976915843424 to 0.6662350441005636, slightly worse. Prior pair+BCE fit 72 also failed, with 33 checks and 1/10.

On the three new training requests, the model requires six checks with Top-1 1/3; the lexical control requires four with 2/3. This is training data, not a generalization test. Top-3 3/3 is weak evidence when every request has exactly three candidates. Check counts simulate candidate verification using saved labels. They are not measured Codex calls, tokens or charges.

## What source establishes, and what remains a hypothesis

The current extractor already forms request/document unigram and ordered-bigram crosses, signed-hashes them into 8,191 cells and adds one lexical-overlap cell. Saying it lacks all interaction features would be incorrect. Scores are linear coefficient sums, without a language encoder or sentence generation. The representation does not explicitly preserve condition/action binding, negation scope or distant subject/object relations.

Sparse coverage of new term crosses, condition-relation limitations, hash collisions and mismatch between BCE and verification order are plausible explanations. None is established as the cause of this failure. This analysis does not recompute features, so it does not assert that any actual request aliases another feature vector or fails because of a collision.

Projection excludes unknown labels from fit rows and preserves masked rows with zero weight. Such rows contribute no data gradient or BCE mass. Fit 71 has 91 rows with weight sum 73; fit 79 has 100 with weight sum 82. Both are smaller than batch 128, so both use one batch per epoch. A change in batch count does not explain this result. Weight normalization followed by batch division averages over positive weight mass; the new nine signals also change the relative contribution of existing signals. This is not a causal decomposition of feature, normalization and floating-point accumulation effects.

## Comparing the next paths

| Path | Value now | Main limitation | Recommendation |
|---|---|---|---|
| New semantic requests and independent contracts | Adds learnable signal and source-group-held-out development evidence | Qualification costs time; increasing counts alone does not guarantee utility | Primary path: prepare 60 requests within a 30–100 batch |
| One limited condition/action feature comparison | Tests a representation hypothesis | Crosses already exist; handcrafted grammar may overfit the author | Prepare in parallel; predeclare baseline versus one change on new development groups |
| Objective or quantization changes | Separately studies ordering objectives and storage/execution formats | Pair+BCE already failed here; compressing a failed model does not establish usefulness | Defer old-validation loss/threshold searches |

A first relation-feature hypothesis would preserve bounded condition/action/polarity associations, such as preserving state on error versus on success. Do not feed provenance IDs, labels, roles or repository names into features. First use newly authored development diagnostics to check distinguishability. Any utility test needs new groups, the fixed BCE recipe and a separate preregistered experiment. No such extractor or new fit was executed by this analysis.

## Concrete next batch: a target of 60 new semantic requests

Count distinct behavioral contracts, not function inputs, CSV rows, candidate counts, translations, error-message variants or restatements of one bug. Deduplicate against the 79 requests and the 120 source-survey drafts. An existing semantic contract retains its ID and may add one newly qualified contract, but not a newly invented request. Report new logical requests separately from new qualified records.

Target ten distinct contracts in each of six strata: grammar/parsing; state and ownership after errors; configuration/type conversion; exact arithmetic/formatting; bounded file/I/O behavior; cancellation/retry/lifecycle behavior. These are acquisition targets, not 60 acquired records. The 53 source inventory supplies starting anchors. Source 82's 57 source/license files are preparation assets, not 57 requests. Source 81's 32 input fixtures and four goals do not become 32 new requests.

Acquire at least six connected source groups across multiple repositories. Record existing training exposure for UUID and humanize rather than treating them as unseen-source validation. Cobra/pflag dependencies, copied standard-library code and shared templates may merge apparently different repositories. Freeze complete groups into train, selection-validation and calibration before reading scores. If feasible, allocate 3/2/1 groups with approximately 30/20/10 requests; if provenance merges groups, acquire more sources and freeze revised counts. An intermediate 30 is an acquisition checkpoint, not automatic permission to train.

These 60 inspected development requests are not final data. Each claimed domain still requires a separately acquired protected final collection of at least 2,400 requests, with candidate, truth, permissions and cost-observation eligibility established. Existing file-search final membership 2,402 is not interchangeable with code-claim ordering or actual profile-routing goldens.

Every qualified request needs a public revision, minimal executable closure, applicable licenses and actual NOTICE files, a frozen semantic contract, faithful candidate descriptions, finite input/output/error/state-channel checks and acceptable/no-answer/unknown truth. Function-behavior claims use native original behavior plus independent expectations. Actual coding-change tasks additionally need an unchanged baseline that fails, an independently authored correct implementation that passes and seeded wrong implementations that fail. Do not report an already satisfied original behavior as successful bug fixing. Finite checks support their stated contract scope, not universal input correctness.

## Product and integration scope

This path is a tiny CPU hint model comparing short Go behavior requests with candidate descriptions. Candidate-truth labels do not label which Codex profile succeeds on a task. Actual routing requires matched profile attempts, independent completion and whole costs including failures, retries and escalation. Repository selection separately needs access and acceptable repository sets; decomposition needs the complete merged parent outcome.

Keep integration optional, return candidate order and supported scope, and preserve verification and fallback. Scores are not calibrated success probabilities. The model has no Bloom-filter guarantee of never missing an answer. Do not permanently remove candidates or force model downgrades from this evidence.

Ternary and INT8 formats remain separate experiments. The current Decode expands coefficients to FP64, so smaller disk files do not automatically reduce decoded memory. Direct packed scoring, reused prepared inputs and whole-request measurement are needed to establish benefits for frequent low-resource hints. No GPU execution, GPU-memory savings or SVD benefit was observed here.

## Resource planning for 2,400 requests is separate

The current sparse representation already separates offsets, indices, values, labels, groups and weights into CSR arrays. It is not waiting for an initial SoA conversion. Preserving these columns while reducing large float JSON and repeated copies is a concrete next pipeline option. No SIMD speedup was measured here.

The actual 79-request projection owns 2,207,317 B of sparse payload and its serialized result is **5,369,450 B**. A proportional same-density extrapolation to 2,400 requests gives approximately 67,057,732 B and 163,122,532 B. This is hypothetical arithmetic, not a scale benchmark. The projected payload is about 51,132 B below 64 MiB, leaving only about 0.08% margin with scratch and allocator costs outside that count. Projected JSON exceeds the current 12 MiB result ceiling and the 64 MiB total new-stage ceiling. Do not apply the fixed 79-request guard to all 2,400 requests.

A separate version must plan group-preserving chunks, binary CSR or another format avoiding float text, whole-stage and per-path retained accounting, and repeated-inference batching. Chunk and CSR-row counts do not replace semantic-request counts. Chunking does not reset previous fit, model or storage consumption. First qualify the proposed 60-request batch and prepare capacity; this analysis launches no new whole-corpus Project.

This independent follow-up reads saved results and source only. New Fit, Project, Features, Score, model Decode, native functions and uploads are all zero. The accompanying JSON pins evidence and separates findings from proposed acquisition and ablation work.
