# Actual file localization — experiment 34, prepared execution

Freeze the [evaluation plan](plan-34.json) before inspecting ranking outcomes. This documents the execution contract; actual task rankings have not yet been evaluated.

Use the same 2,400 complete requests and complete pre-fix file paths. Report original-path BM25, identifier-normalized BM25 using existing `NormalizeText`, and baseline-first interleaving. The primary comparison is original versus interleaved. No training, threshold tuning or best-result selection occurs.

`fileeval.Rank` accepts only query and paths. Labels enter `Measure` after ranking. Different file identities remain distinct even when normalized text matches. Auxiliary input-limit failures fall back to baseline and are counted.

Compute Hit@1/10/100 and MRR over all 2,400 tasks and report repository aggregates. Keep the two new-file-only tasks. Separately count total/mapped old-file labels and whether every target appears in the first ten. Compare 20-candidate pages to the first target, with improved/worsened/tied counts, only for tasks with at least one mapped old target; expose that applicable denominator.

The precommitted primary gate requires no increase in applicable total pages and no decrease in all-task Hit@10. Passing establishes only a lexical file-localization result, not task success, difficulty, decomposition, actual LLM savings or production readiness. Page costs assume perfect equal-cost target recognition. Changed files are not exhaustive relevance labels.

```sh
go run ./cmd/riido-fileeval --out .cache/file-eval-new
```

Pinned query/label projections and all 2,400 local catalogs are required. No network requests occur. Any missing catalog aborts before the first ranking; this guard was verified against the incomplete cache. Do not turn the currently available small subset into a final evaluation. Final scoring follows completed collection.
