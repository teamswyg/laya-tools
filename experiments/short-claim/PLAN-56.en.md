# Short claims and a tiny Go model: experiment 56 preparation plan

[한국어](PLAN-56.ko.md) · [Machine-readable plan](plan-56.json) · [Golden-set scale](../../docs/golden-set-scale.en.md)

**The current state is `prepare_contracts_only`: 0 new data items, 0 fits, 0 weights, and no new runtime performance measurements.** Data, partition, and execution-plan readiness gates are incomplete. The user has already authorized continued local training, verification, and public publication. Continue within that scope when readiness gates, plans, and CI are ready; do not create an additional human approval step. This is not a record of new training, model publication, or production activation.

The goal is to test whether many small hints can help without a large encoder. We begin with one claim: **“this candidate may contain evidence that satisfies the requested behavior.”** We do not jointly predict model difficulty, actual task completion, automatic approval, or monetary savings.

## How this differs from existing experiments

[Experiment 55](../task-outcomes/results-55.json) completed four attempts: two external Go requests, each run once with each of two profiles. One observation per profile/request combination is a development comparison, insufficient for general performance or stable success-rate claims. [The preceding Laya observations](../task-outcomes/routing-predictions-55.json) received the full original prompts but abstained after truncation at 512 tokens. Those are cold CPU whole-process observations, not resource results for a tiny claim model.

The short input in 56 is a **new, separate contract**. We preserve 55's original prompts, answers, specs, hashes, token budget, and historical sealed plans. We do not shorten existing requests to rescore 55 or improve its success rate. Initially, we use no generative model, automatic summarizer, Laya encoder, or teacher output.

## A bounded claim and input

For example, a request to “keep active entries and remove expired entries” must be distinguished from a candidate that “keeps expired entries and removes active entries.” Initial domains cover conditional retention/removal, acceptance/rejection, order preservation, and inclusive boundaries: observable behavior that independent checks can verify.

The new input contains a short `request`, short `candidate`, feature-contract `schema`, and source-linking `provenance`. Only text enters the features; target IDs, correct candidate ranks, check outcomes, generator modes, and training roles do not.

| Limit | Proposed contract |
|---|---|
| Candidates | 1–8 per request |
| Text | Each request and candidate is valid UTF-8, at most 512 bytes |
| Normalization | Each text after `lexicalhint.NormalizeText` is at most 512 bytes and 32 words |
| Overflow/failure | Do not truncate. Out-of-scope input, unknown schema, or nonfinite scores explicitly fall back to the declared lexical baseline order. |
| Public output | An `unverified` score, model/feature-contract hashes, and fallback reason. Do not retain original input text in result logs. |

A word bound does not guarantee equal information capacity in Korean and English. Report language and byte/word-length strata separately. Preserve input and option order; use the identical normalization function for training and inference. Seal the concrete new schema name and hashes separately after implementation and verification.

The score is neither a calibrated success probability nor action authorization. It proposes verification order, without removing candidates or certifying full task completion. Preserve all candidates. Record unknown behavior, ambiguous requirements, and no-answer requests separately instead of forcing true/false labels.

## Publishable data and ground truth

First prepare original Go microcontracts, truth tables, and authored short requests/candidates. Independent checks establish whether observable candidate behavior satisfies a request. Test the labeling system itself with an unchanged wrong implementation, a correct implementation, and counterexamples involving condition reversal, negation, `<=/<` boundaries, and operation order. Review natural-language fidelity to the truth table separately. Passing code tests does not automatically certify a natural-language label.

Represent ground truth as a set of acceptable candidates, not a single mandatory ID. Report insufficient input evidence and check errors as `unknown`. Record exclusions before inspecting model scores, keeping original request denominators and exclusion counts. Do not turn no-answer/unknown cases into successes for a failing model.

The development preparation target is **48–120 candidate parent contracts**. Actual acquisition is 0. This is not a target or guarantee of 48–120 independent groups. Variants sharing a prototype, source, or core template may form one connected group. Random names, constants, or translations alone do not create new independent requests.

[The 120 public source candidates](../../docs/public-go-acquisition-53.en.md) and [the two external contracts](../../docs/public-go-contracts-54.en.md) may be reviewed later as separate development transfer probes. Previously inspected material is not a new final set. Before adopting external material, pin its revision, LICENSE/NOTICE, use/distribution scope, and independent ground truth again. Do not use or publish private code, real user prompts, authentication information, or raw execution traces.

## Connected-group audit and partition

Build the source graph first. Connect shared requests, candidate source, prototypes/core templates, translations/paraphrases, counterexamples, parents/children/siblings, and forks/copied source; take transitive components. Shared source in negative candidates also creates an edge. Do not use validation/calibration candidates as training negatives. Record limitations from a shared author/generator separately.

**No partitioning or training before the actual connected-component audit.** The target allocation is approximately 60/20/20% by group for train/validation/calibration. Proposed preparation minima are 15 actual connected groups, at least 9 in train and 3 each in validation/calibration. These are operational preparation minima, not statistical guarantees or deployment eligibility. If groups are insufficient or too concentrated for a valid partition, stop rather than splitting groups or lowering minima. Seal exact role membership, counts, group graph, and data hashes before considering training.

Report parent requests, pairs, variants, contract families, connected groups, and language counts separately. For example, 20 variants of 120 parents produce 2,400 rows, not 2,400 independent final requests. A **new final set of at least 2,400 requests for semantic retrieval utility remains unacquired and unscored**. Do not read or reassign existing CoSQA reserves or file-search protected final data. Validation in 56 measures development utility only; promotion and production use are prohibited.

## Preparation sequence and stop conditions

1. Prepare input/ground-truth verification contracts and wrong-implementation controls.
2. Audit groups and seal eligible development data, roles, baselines, and budgets.
3. Without training, evaluate fixed order, BM25, word/ordered-word matching, and a disclosed narrow conditional/negation grammar rule on identical candidates, labels, and fallback scope.
4. Use a ground-truth ordering oracle to compute an upper bound on possible verification-count savings. If the baseline is already optimal, or maximum relative gain is below 5%, stop without training. The oracle computes a bound; it is not a deployable policy or model feature.
5. Only after the bound, partition, data eligibility, and resource preparation pass, seal a separate training-execution plan. The training counts below are proposals for that review; execution readiness is incomplete and fits remain 0. Continue within the user's existing authorization when data, plans, and CI are ready. Evidence gates that stop failed preparation are not a request for additional user approval.

Verification count means independent checks until the first acceptable candidate in the fixed set. Fix no-answer/unknown handling in advance. Disallow apparent savings from candidate deletion. Define maximum relative gain as `(best nonlearned baseline checks - oracle checks) / baseline checks` and report the valid denominator. Do not interpret it as actual LLM token, price, or completion-time savings.

## Proposed tiny model comparison

| Comparison | Role |
|---|---|
| Fixed order, BM25, lexical/ordered-word matching, narrow grammar rule | Nonlearned controls. Do not force learning when rules are optimal. |
| 16-coefficient lexical FP32 | A cheap trainable lexical control, not evidence of semantic understanding |
| Existing 8,192-coefficient signed hashed cross features, FP32 | A claim candidate learning request/candidate associations without a large encoder |
| INT8/PTQ children of FP32 | Quantization losses from the same fitted parent |
| Separately trained ternary STE sibling | A separate comparison of ternary-forward training |

Propose **at most 8 fits**: two feature families × seeds 1729/2718 × FP32/STE. Use learning rate 0.1, L2 0.0001, at most 50 epochs, batch128, and the existing ternary method with global mean-absolute scale and 0.7× threshold. Including INT8/PTQ derivatives of FP32 yields at most 16 artifacts. Seeds, files, and epochs are not unique requests. Do not change the primary candidate or gate after seeing the best result; the proposed primary is 8,192-coefficient FP32/seed1729. Select epochs by validation NLL. Calibration separately examines a later sealed abstention policy; it does not automatically establish probability accuracy.

Here FP32 means **coefficients used for scoring are rounded to float32**. Existing `pairlearn` shadow coefficients, gradients, loss accumulation, and reference scoring use float64. It does not mean all arrays, arithmetic, or the process run in FP32. `.hbin` is an 8,192-dimensional format and must not be reused directly for a 16-coefficient model. Existing `.hbin` loading expands to 8,192 float64 coefficients, 64KiB for coefficients alone. Bitmap/sign file compression does not prove runtime RAM savings, native ternary arithmetic, or 1.58-bit execution performance.

[The existing 24-request synthetic pilot](../semantic-learning/README.en.md) passed its pilot gate only for FP32/INT8; ternary variants failed. [The 16-coefficient CoSQA development result](../semantic-scale/LEXICAL-03.en.md) also missed its prespecified AUC gate. Preserve those failures; do not replace utility with small file size as the new success criterion.

## Proposed development-utility and resource gates

Fix actual measurement conditions in the later training-execution plan. **Every value below is an unmeasured target, not a performance report.**

| Item | Target/handling |
|---|---|
| Development utility | At least 5% fewer mean independent checks than the strongest nonlearned control on the same validation |
| Quality | No Top1/Top3 degradation; separately report denominators and regressions for reversal, negation, boundaries, and language strata |
| Policy verification | If the original first candidate is preserved, report unchanged Top1 as a structural guarantee, not learned improvement |
| Ternary | After FP32 proves useful, compare the same utility/quality gates and actual latency/RSS; do not adopt on file size alone |
| Execution | One CPU thread; no GPU, ORT, or Laya encoder |
| Training budget | Total wall at most 5 minutes, peak process RSS256MiB, prepared sparse payload at most 64MiB |
| Inference target | Process RSS at most 32MiB; warm p95 at most 1ms per request with up to 8 candidates |
| Storage budget | At most 64MiB total new data, models, and numeric records; do not silently increase limits |
| Stop | Group/source/label failures, target leakage, nonfinite values, budget excess, optimal rules, or insufficient utility |

Do not count different seeds as independent user samples. Report point estimates and uncertainty for small groups. Unknown measurements are not zero. Passing a development gate does not replace a new final evaluation or establish general semantic understanding or actual cost savings.

Measure feature extraction, scoring, ordering, serialization, cold process start, and warm requests separately. Distinguish Go heap/allocations/pprof from OS whole-process RSS, user CPU, and wall. Explicitly state GPU nonuse for this CPU-only experiment. Prefer one loaded immutable coefficient array and caller-owned scratch; count cache hits separately from actual scoring. Publish aggregates and hashes, not original inputs, raw pprof/ORT traces, or personal paths.

## Go reuse and publication

Propose a new 56-specific loader/driver reusing `internal/lexicalhint`, `internal/hintlearn`, `internal/pairlearn`, and `FitWithTrace`. Existing historical CLIs pin their datasets/plans; do not bypass those pins with different data. `internal/ternarytrain`, which requires Laya features, is not the main path for this encoder-free experiment. New input/model contracts and reproducibility checks are not implemented yet.

GitHub will contain plans, code, original public fixtures, numeric aggregates, licenses/provenance, and model SHA/ref metadata. **Do not put weights or model binaries in Git.** Consider model artifact publication only after development-utility, reproducibility, format, license, and public-content review, using a separate immutable Hugging Face revision. Cross-reference exact file refs, scope, and failures; neither publication nor CI passage means production promotion. There is currently no HF publication or model ref for 56.

## Primary references and limits

- [Original fastText paper](https://arxiv.org/abs/1607.01759): a reason to compare a simple cheap text classifier, not a guarantee of this experiment's quality or speed.
- [Original Feature Hashing paper](https://arxiv.org/abs/0902.2206): a reference for a bounded feature space; directly inspect collisions and unseen expressions.
- [Original BitNet b1.58 paper](https://arxiv.org/abs/2402.17764): a reference for ternary training comparisons, not evidence that this Go scorer is a BitNet LLM, the smallest possible system, or a first implementation.
- [Pinned upstream Laya code](https://github.com/NandhaKishorM/laya/tree/6d942c92081fbc139e736bbd9ac0023223c29b7f): [exact input/truncation statistics](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/laya/common.py#L135), [bounded length batching](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/laya/onnx_agent.py#L667), and [example result cache](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/examples/hooks/cache.py#L19). These inform input/operational design; upstream execution, model adoption, and numerical reproduction have not occurred. Do not compare T4/MPS figures with the current cold Mac CPU baseline.
