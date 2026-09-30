# Changed helper calls and concentrated losses — experiment 26

**Keeping all 2,948 distinct questions, one question accounts for 144 of the primary comparison's 150 additional pages.** Moving from the frozen 16-feature policy to the 20-feature policy mainly lost a valuable helper call. We do not remove this case or relax the acceptance criteria.

## Method

The [precommitted plan](plan-26.json) verifies both experiment21/25 release archives and replays their frozen heads and thresholds without fitting or recalibration. Every published repository-level and aggregate metric is reproduced exactly. Questions are partitioned into neither calls, new-only calls, old-only calls, and both call.

Model-file counts such as24 are not question counts. This audit uses2,948 questions and3,009 candidates. Two seeds repeat the same questions; they are not independent datasets. The three repositories have already informed development, so this is not a fresh final evaluation.

## Results

| Seed | Old→new pages | Old→new calls | Worsened questions / added pages | Improved questions / saved pages | Largest one / total additions |
|---|---:|---:|---:|---:|---:|
| 1729 | 6,326→6,439 | 2,408→2,363 | 4 / 150 | 12 / 37 | 144 / 150 (96%) |
| 2718 | 6,331→6,440 | 2,317→2,328 | 3 / 147 | 13 / 38 | 144 / 147 (97.96%) |

Concentration divides by the sum of **positive increases**, not the net113-page change. Each seed contains a144-page loss; aggregate evidence alone does not establish that it is the same question across seeds.

For the primary seed,127 new-only calls save31 pages net. The172 old-only calls cost144 additional pages net when skipped by the new policy. Captured beneficial calls increase315→320, yet missed benefits increase53→170 pages. Capturing more beneficial cases can still worsen total cost when a large benefit is missed.

## Next hypothesis: allow a skipped helper to return

A separate experiment should consider running the helper once when the client requests another page after an initial skip. Runtime must react to actual continuation, without knowing the answer. Evaluation stopping at a known answer only simulates a verifier and is not real-user validation.

No benefit for this policy has been measured yet. Precommit controls, ordering, deduplication, calls, completion quality and resource costs separately. Retain the large-loss case, the static policy's failure record and existing acceptance criteria.

## Verification and limits

The entire aggregate report is byte-identical on replay. Synthetic tests exercise all four transitions, signs, conservation, concentration denominator, loss buckets and invalid ranks. Full Go race tests and vet passed.

On M4 Pro CPU, source verification, feature preparation, both policy replays and aggregation together took1.83s wall,1.42s user and98,238,464 bytes peak RSS (about93.7MiB). These are whole-process measurements, not single-inference or GPU-memory measurements.

```sh
go run ./cmd/riido-callchange --out .cache/call-change-new
```

[Aggregate results](results-26.json). No new model or HF weights were produced. Git excludes raw questions, individual labels, weights and profiles. No LLM savings established. Not production ready. A final evaluation with **at least2,400 fresh real requests separated from training and tuning** remains necessary. Retrieval results do not establish model-routing, repository-routing or task-decomposition quality.
