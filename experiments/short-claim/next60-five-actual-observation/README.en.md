# Actual observations and finite checks for five tasks

[한국어](README.ko.md)

Five distinct requests cover nil decoder configuration, checked diagnostic writes, required annotations, context preflight cancellation and big-integer precision. We froze 23 inputs and Wants before the first native trial, then saved [69 observations](OBSERVATIONS.v1.json), one per input and candidate. Inputs, candidates and observations do not become additional semantic requests.

| Stage | Actual result |
|---|---|
| First native trial | Start 1, Wait 1, exit 0, all 69 dispatches completed |
| Candidate returns / panics | 67 / 2; panics retained as observations |
| Explicitly tracked error methods | 23 normal returns, no panic; internal/startup counts uninstrumented null |
| Saved comparison | 69 rows: 42 satisfied, 21 unsatisfied, 6 unknown |
| Required predicates | 228: 185 true, 32 false, 11 unknown |
| Worker resource measurement | OS max RSS 24,641,536 bytes (23.5 MiB), about 1.20 seconds including start/wait |

The resource measurement covers the local Go worker, startup and durable recording. It is not model inference, GPU memory, Go heap or token savings. The original candidates were not replayed.

All five reference candidates satisfy every frozen predicate with known observations. Nine other candidates have known counterexamples. **The diagnostic writer baseline has unavailable return channels and remains unknown; it is excluded from training.** The BigBytes baseline retains both known text counterexamples and separate unknown inputs. A negative label does not erase those unknown predicates. Read the [independent nonblind comparison](semantic-review/README.en.md) separately from [Root's finite adoption](root/ROOT-QUALIFICATION.v2.json).

The five requests add 14 labels: five positive and nine negative. Combined with the preceding sixteen, the adopted pool has 21 requests and 61 labels. Its evidence covers 103 inputs and 305 original observations. Removing four observations belonging to the excluded baseline leaves 301 selected-candidate observations. [JSONL generation and actual reader validation](../next60-development-twentyone/README.en.md) are complete; adoption does not imply a new fit.

Excluding the writer baseline moves its reference to selected index 0. Always choosing index 1 therefore achieves the arithmetic control **18/21, about 85.7%**. Always calling candidates negative achieves **40/61, about 65.6% per label**. These use different denominators and are not measured model accuracy. Existing connected groups 76, 77 and 78 are reused; this adds no unseen source family. No new corpus fit starts before thirty qualified requests. The sixty-request checkpoint, protected 2,400-per-domain evaluation and 5% utility gate remain.

The first controller invocation failed before Start or Wait because its decoder expected one terminal LF while the frozen fixture had two. [That failed receipt](PRESTART-FAILED.v1.json) has Start/Wait zero and no worker output. Owned decoding was corrected while fixture and Wanted bytes remained unchanged. A prior parser/type name collision was also a preparation compile error. Failures were preserved, without changing the task or labels.

CI mode `five` runs owned synthetic controls and compares **saved** results. It does not launch the original observation worker or a model. Existing repository native Laya CI has its own scope. [Outside process evidence](OUTSIDE-RESULT.v2.json)—reaping, output equality and post-wait pins—is separate from finite semantic comparison.

The [source-selection proposal](SOURCE-SELECTION.v1.json) preserves its pre-execution state. Later source and compiler reviews are in `root/`. `source/` contains owned code; upstream bodies, executables, raw journals and private paths are excluded. `notices/` preserves full Cobra, exact pflag 1.0.9, mapstructure, humanize and selected Go notices. Own material uses repository Apache-2.0; upstream licenses remain unchanged. The private path-bearing execution plan is excluded, so this is not a claim of unchanged path/binary replay.
