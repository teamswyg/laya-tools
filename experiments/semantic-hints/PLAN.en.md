# Semantic hint experiment 01 — cheap claims that help search

Research only. Independent of difficulty routing and ternary QAT; no default behavior changes.

## Target and hypothetical user

**Help an agent order expensive evidence inspections across repositories it is already allowed to access.** A hypothetical Harbor workspace has 64 repositories with summaries and evidence IDs. For “find the implementation that prevents duplicate charges,” the model suggests that repo-17/evidence-4 may help. A subsequent verifier inspects it. The model does not edit payments code, authorize actions, or make the final repository/LLM selection.

The initial executable experiment has 16 fictional evidence items, not 64 repositories. Scaling to 64/256/1024 candidates requires separate measurements.

## A claim is a scoped, unverified hint

Use a finite predicate such as `may_contain_evidence`, candidate/evidence IDs, catalog snapshot ID, ranking score, `unverified` status and reason code. See the JSON design example in [the Korean plan](PLAN.ko.md); it is a proposed interface, not this benchmark CLI's output schema. A score is not a truth probability. Invalidate reuse when query, snapshot, permissions or model changes. Missing hints do not establish falsehood; evidence references do not prove relevance.

Start with repository/code evidence retrieval. Later predicates `may_need_joint_change`, `may_be_decomposable`, and `may_need_stronger_model` require independent labels, verifiers and experiments. Truth adjudication, permission decisions and execution authority are outside the target.

Request → cheap ranking → inspect promising candidates → inspect remaining candidates → exhaustive search or original LLM path. Exhausted budgets return incomplete, not “no answer.” Exhausting a finite catalog does not establish absence in the real world. Wrong hints cost inspections; delayed correct hints increase latency. Hard latency limits need a separate omission policy.

The Bloom-filter analogy is architectural: a cheap front end helps an expensive back end. Learned semantic scores have no no-false-negative guarantee. Learned Bloom filters use backup structures for known set members; that guarantee does not transfer to open semantic retrieval. Preserving all candidates only protects completeness within the supplied finite catalog and an accurate verifier.

## Separate experiment tracks

| Track | Question | Current stage |
|---|---|---|
| A | Can exact token overlap reduce inspections? | Executed Go baseline |
| B | Are compact hashed bits a useful speed/collision tradeoff? | 256/4096-bit comparison; no training |
| C | Can an encoder-free ternary learner improve semantic generalization? | Future QAT after freezing new split families |
| D | Does selectively invoking Laya/previous QAT justify its cost? | Future independent comparison including ~2.5GiB encoder cost |
| E | Does a real agent use fewer LLM resources at equal task quality? | Future paired public-task evaluation |

A/B results are neither C training results nor E token savings. Hashed terms are not a semantic encoder. Reject ternary when its size advantage does not justify quality/speed versus FP32 or INT8.

## Frozen measurement contract

Development fixtures: 16 authored English domains × direct, paraphrase and contrasting/negative request, plus four absent-answer requests. One author, no independent final evaluation. Contrast requests only provide a narrow negation probe. All methods use identical labels; ranking receives text and catalog only, never the target ID.

Compare catalog order, fixed shuffled order, exact token intersection, and 256/4096-bit hashed intersection. Record target rank, recall@1/@4, checks until first correct candidate, absent-answer full-scan cost and worst case. The oracle reads an authored target ID; it is not a real verifier, LLM or coding evaluator. Checks have equal assumed costs. Compare against strong nonlearned baselines, not only arbitrary ordering.

Warm latency includes query tokenization, scoring and sorting; catalog construction is excluded. Record allocations and separate OS process maximum RSS. Go pprof does not measure native/GPU memory. Identical inputs should repeat identical rankings, not count as independent votes. Meaningful repeated inference must evaluate new evidence/candidates/hypotheses.

Future C targets, **not achievements**: model ≤64KiB, prediction state ≤1MiB, process RSS ≤64MiB, warm p95 ≤1ms for 64 candidates, independent final recall@4 ≥95%, and ≥20% fewer checks than exact tokens. Do not lower gates during tuning. Report language/family regressions and fallback costs.

Total cost = feature preparation + all hint calls + all verification + fallback + final LLM. Faster hints fail the system objective if total cost does not fall. Actual LLM cost remains unmeasured until paired deployment evaluation.

## Learning plan and model family

C takes bounded word/character features for the query/candidate pair and predicts relevance ranking. Positives require evidence satisfying the request. Hard negatives share vocabulary but have opposite conditions, different APIs, stale versions or wrong repositories. Track synonyms, Korean, English, mixed code identifiers, absent answers and multiple required evidence items separately. Teacher assertions alone do not establish labels: require evidence and independent verification.

Split train/validation/calibration/final by repository and semantic family; audit duplicates and paraphrase leakage. Design features on development data, then seal a new final set. Once final data informs a revision it becomes development data. Use at least two seeds and equal splits/budgets for FP32/INT8/PTQ/QAT; preserve individual failures.

Start with **one relevance learner** plus nonlearned baselines. Laya is an optional teacher/comparator; a distilled child need not inherit its encoder. The existing difficulty model and new hint model are sibling experiments with different targets. Teacher scores are auxiliary signals, not factual evidence. Add domain-specific children/siblings only after demonstrating a single-model failure and benefit from specialization.

## Repetition, concurrency and resources

A hundred identical calls repeat the same errors. Cache using request + catalog version + model version + permission scope; deduplicate verification. Evaluate total batch cost for 1/10/100/1000 hints, not just a single call. Never publish private queries as cache keys or logs.

Begin encoder-free experiments on CPU. Do not duplicate GPU encoders for tiny heads. For future GPU learning, pilot at most two workers and reduce to one if memory pressure, swap or throughput worsens. Never benchmark final latency alongside training. Apparent GPU availability is not a memory/throughput measurement.

## Publication and sources

Publish code/plans/public synthetic results to GitHub. Publish newly trained models to immutable Hugging Face versions after CI/license/hash gates. A/B creates no trained model, so it produces no model release. Existing QAT releases keep their original scope. Never release private repositories or actual user prompts.

- [Learned Bloom Filters, Mitzenmacher](https://papers.neurips.cc/paper_files/paper/2018/hash/0f49c89d1e7298bb9930789c8ed59d48-Abstract.html): scope of guarantees and backup structure, not semantic retrieval guarantees.
- [Algorithms with Predictions](https://www.cs.toronto.edu/~bor/2421s21/papers/mitzenmacher-survey.pdf): predictions assisting established algorithms.
- [Previous ternary QAT](../ternary-qat/README.en.md): 572B is the head alone, not the full encoder.
