# Direction after rereading Laya57

On2026-10-02, reviewed SDK commit [`4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c`](https://github.com/NandhaKishorM/laya/tree/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c). Bodies of19 public files matched their Git blobs. No model execution or weight download occurred. This review is not a measured performance improvement.

## Our target and the reference implementation

The original [Router](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/router.py#L774) **selects a Laya checkpoint** from explicit model/task and language signals. It is not a trained predictor of Codex completion or usage. The [typed decoder](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/agent.py#L1179) provides choice argmax, an expected ordinal score and noul P(true). This differs from free-form generation; rounding a score need not produce its argmax.

Riido prioritizes repeated **hints about candidates to verify** using extremely little resource. Preserve all candidates and fallback; existing verification tools retain truth and acceptance. The next criterion is reducing checking work beyond BM25 across diverse public behaviors, rather than increasing the number of model variants.

## Three inexpensive reference points

| Reference | Subsequent scope | Verification |
|---|---|---|
| Score meaning and abstention | Use the [confidence implementation](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/confidence.py) to distinguish probability/heuristic, calibration and passed/abstained/unevaluated in small metadata | An unconfigured gate is not passed; nonfinite/missing scores cannot pass |
| Support checks before heavy loading | Review consistent English checkpoint language guards across repository preview and task routing | Unsupported scripts cause zero loader calls; preserve explicit override; Latin script does not certify English support |
| Repeated-result caching | A separate bounded experiment informed by the [cache example](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/examples/hooks/cache.py), keyed by raw text, ordered candidates, implementation/model revision and temperature/budgets | Negation, `<`/`<=`, candidate order and revision changes miss; test mutation, byte limits, eviction and lookup costs |

These are proposals. Do not retroactively modify public evidence or claim cache, speed, memory or fee savings. A normalized feature SHA can collapse punctuation and case semantics. Cache lookup might even exceed the current microsecond Go scoring cost.

`answer_confidence=max(p)` already existed in the earlier SDK pin. It differs from entropy-based choice/score `confidence`; noul uses max(p). Neither is domain success probability without independent checkpoint/option-count/domain calibration. Earlier entropy descriptions apply only to that choice/score field; preserve historical snapshots.

Do not rebuild existing Go choice/noul, truncation checks, resident ORT, BPE cache, none/ambiguous, fallback or JSONL as a new engine. Go `Predict` currently lacks score support, so not every reference primitive is ported. Do not remove locks protecting mutable BPE state and Predict/Close lifetime merely because of the reference Router's lock placement. Its Agent has a separate inference read lock.

English421M/multilingual322M are total model sizes; their backbones are395M/307M. A small ternary head does not remove retained backbone RAM. [Multiple questions](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/agent.py#L1011) reuse state tokenization while retaining question-specific encoder inputs. Do not cache question-dependent hidden states by state string alone.

## Benchmarks and licenses

The original [tier-routing example](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/examples/29_presets_model_router.py) is a policy example. The [application benchmark](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/research/scripts/bench_apps.py#L243) routes399 GSM8K/MBPP/AG News domain labels, not actual Codex completion or costs. [BENCHMARKS](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/BENCHMARKS.md#L168) explicitly records missing committed evidence for a fine-tuned row and weak base results. Timing reported on another GPU is not our Mac/Go performance.

The [SDK LICENSE](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/LICENSE) declares Apache-2.0. Preserve license, applicable notices and modification notices when copying or translating. Model-card declarations were checked separately.

| Card | Reviewed revision | Card declaration |
|---|---|---|
| [English](https://huggingface.co/convaiinnovations/laya/blob/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851/README.md) | `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851` | Apache-2.0 |
| [Multilingual](https://huggingface.co/convaiinnovations/laya-multilingual/blob/e4e9ddf21a7b1903b7acffd8814ad4307bf63a67/README.md) | `e4e9ddf21a7b1903b7acffd8814ad4307bf63a67` | Apache-2.0 |
| [Typed decisions](https://huggingface.co/convaiinnovations/laya-typed-decisions/blob/1a793eb568e6718f15941d08f85432581df534e3/README.md) | `1a793eb568e6718f15941d08f85432581df534e3` | Apache-2.0 |

This verifies public declarations, not all training-data rights. Do not import teacher labels as independently established truth. New paid model calls, fits, weights and final evaluation remain0. Preserve at least15 groups,5% necessary utility and2400 distinct protected-final requests per domain.
