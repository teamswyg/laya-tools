# Short claim hints: what did changing captions teach us?

[한국어](README.ko.md) · [Saved result](observations/results.json) · [Independent saved-result checks](saved-qa/SAVED-QA.en.md) · [Detailed analysis](analysis/ANALYSIS.v1.en.md)

More detailed descriptions did not make the small learned models rank better in this experiment. Source-derived behavior captions helped the existing word-matching controls, while both fixed models worsened. The current user path is a **model-free Go tool that suggests candidate verification order**. These failed models are not activated by default.

The three requests concerned parsing size errors and receiver state, query-value order and omission rules for structs, and completed tokens before an unterminated quote or trailing escape error. Each request kept the same three candidates: A used the original captions; B used authored, source-grounded behavior captions. Requests, candidate IDs/order, development rubric and models stayed fixed. Only request and caption text entered model features; code, rubric and observed answers were not hidden inputs. Two FP32 models × two caption versions × nine candidate positions produced 36 scores, without retraining.

`checks` sums **the first acceptable candidate's zero-based rank + 1** under the fixed development rubric. An acceptable candidate in third position costs three simulated checks. This is not a count of actual verifier or LLM calls. Smaller values mean the ordering reaches an acceptable candidate sooner.

| Fixed method | A → B checks | A → B Top1 |
|---|---:|---:|
| Model 71, pointwise supervision | 3 → 8 | 3/3 → 0/3 |
| Model 72, pointwise + relative-order supervision | 5 → 6 | 1/3 → 1/3 |
| Original order | 6 → 6 | 1/3 → 1/3 |
| BM25 / lexical_ordered | 8 → 3 | 0/3 → 3/3 |

`narrow_rule` also scored 8 → 3, but every ordinary-language request fell back to BM25. This does not show rule-based semantic understanding. Every method achieved Top3=3/3 because each request has only three candidates; the metric cannot distinguish ordering quality here. B changed length and sentence structure as well as content, so this is not a causal isolation of added information. Differences from training phrasing and missing discriminative conditions remain hypotheses.

Both models had already failed utility evaluation on their earlier validation data. The immutable [model 71 HF archive](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33) and [model 72 HF archive](https://huggingface.co/JooYoon/riidolaya-shortclaim-rank-bce-failed-72/tree/d090b00e9a5dab00d5372dfd6412c9aee0c60b7b) preserve research failures. Each FP32 file is 32,792 B; exact artifact hashes are in the [frozen plan](plan/PLAN.v1.json). Public archiving and this small diagnostic do not qualify the models. `qualification`, `production_ready`, `training_ready` and `protected_final` remain false. Original null labels, roles and weights were not overwritten by the separate development source rubric.

**To try the current tool**, build from the repository with Go 1.27.1. No Python, model download or API key is needed.

```sh
CGO_ENABLED=0 go build -trimpath -o bin/riido-shortclaim ./cmd/riido-shortclaim
./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
```

A person supplies a short request and candidate descriptions; an agent follows `verification_order` to inspect actual code and checks. Every candidate stays present. Status is `unverified_heuristic`; scores are not success probabilities or authorization to act. Each request/caption is limited to 512 B and 32 normalized words, with 1–8 candidates. Use `--stream` for repeated requests. The [usage guide](../USAGE-56.en.md) covers the input schema, examples and four controls. This command loads neither HF model and does not automatically change Codex routing.

On macOS, `/usr/bin/time -l ./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json` can measure your own whole model-free process. That is a separate user run. The research worker's [saved resource sample](RESOURCES.v1.json) recorded maximum RSS **10.0625 MiB**, real 2.23 s, user 0.05 s and system 0.17 s. It includes startup, source/file hashes, model decode, controls, checkpoints and fsync; it differs from Go heap. Do not divide 2.23 seconds by 36 scores to claim inference latency or LLM savings. It used neither GPU execution nor the Laya encoder.

Research source is archived as `runner/source/*.go.txt`, not shipped as a directly executable public CLI. The maintainer `go.mod` containing private local paths is omitted. Raw logs and model bodies are not committed to Git.

Next, prepare new comparison problems rather than repeatedly fitting these three requests. B's word controls already reached the minimum three checks, leaving no ordering headroom on this slice. Before viewing scores, freeze requests for new public behaviors that distinguish returned errors from panic, conditional state changes, negation and exceptions; write faithful source-grounded candidate descriptions. Group shared sources, helpers and aliases together to prevent train/validation leakage, and freeze the rubric and existing controls before scoring. Repeated failures could then motivate a separate feature ablation for semantic conditions. These three development requests are not independent final data, and authors were exposed to the material and earlier results. The target of 2,400 fresh protected final requests per domain remains unmet.
