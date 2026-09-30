# A helper on continuation — experiment27

**Continuation reduces page losses but does not pass the existing gate.** Across all2,948 distinct questions, the primary policy moves from6,326 to6,296 pages and2,408 to2,439 helper calls. It still needs9 more pages than always-helper (6,287). Defaults remain unchanged.

## User-facing hypothesis

If the first20 baseline candidates answer the request, a helper may be unnecessary. If the client asks for another page, invoke the helper once. Preserve the candidates already returned and apply the helper ordering to the remaining candidates. An initially wrong claim can thus be revisited.

This experiment evaluates the ordering offline; it does not integrate with runtime sessions. Existing cursors bind to an immutable ranking digest, so a changing order requires a separate session contract and tests.

## Fixed method

The [plan](plan-27.json) was committed before execution. Keep the first20 baseline candidates, then append the existing baseline-first interleave with all previously emitted IDs removed. Preserve all3,009 candidates without duplicates. Freeze the16/20-feature heads, thresholds, two seeds and gap rule. No training or calibration occurs. A classifier-free deferred policy is a separate control.

The evaluator assumes a perfect verifier stops on the page containing a known answer. **Gold affects simulated stopping and logical call counts only, never features, initial decisions or ordering.** Real clients may miss an answer or seek more evidence. Reported calls are not actual agent usage. Preparation computes every helper ranking for counterfactual comparison, so these are not counts of actual ranking executions either.

## Results

| Policy | Seed | Static→deferred pages | Initial + continuation calls | Recall@10 | Pass |
|---|---:|---:|---:|---:|---|
| Baseline→pure deferred | — | 9,488→6,451 | 0 + 523 | 0.7483 | No |
| Gap rule | — | 6,323→6,296 | 2,548 + 17 | 0.8097 | No |
| 16 features, primary | 1729 | 6,326→6,296 | 2,408 + 31 | 0.8100 | No |
| 16 features, replication | 2718 | 6,331→6,298 | 2,317 + 34 | 0.8094 | No |
| 20 features, secondary | 1729 | 6,439→6,294 | 2,363 + 27 | 0.8111 | No |
| 20 features, secondary | 2718 | 6,440→6,295 | 2,328 + 29 | 0.8104 | No |

Always-helper uses2,948 calls/6,287 pages with Recall@10 0.8138. Recall@1 is0.4179 for every policy. Continuation changes only ranks after20, preserving each static policy's Recall@10.

The primary comparison improves11 cases by43 pages and worsens6 by13, saving30 net. The primary20-feature comparison saves162 pages in6 cases and adds17 in9. Rescue helps aggregate cost, not every question. Aggregates do not establish identity with a particular experiment26 question.

Pure deferred cuts simulated calls82.26% versus always-helper but adds164 pages (2.61%). This is a potentially useful tradeoff, not a reason to relax the precommitted requirement of no additional pages and at least10% fewer calls. Learned variants also retain the existing requirement to weakly dominate the frozen static gap rule on pages/calls, with at least one strict improvement.

## Implementation, reproduction and resources

Go `InterleaveAfterPrefix` uses a fixed bool array and contiguous candidate storage, compacting and shifting the existing interleave output without mutating inputs. It introduces no shared state or locks. It computes ordering only; it does not defer actual helper construction.

Tests exhaust small permutations, every prefix length and partial/empty hints, checking ordering, uniqueness and input preservation. They cover malformed permutations, page20/21 boundaries, single calls and accounting. Every published static head/rule fold and aggregate replays exactly. The full new report is byte-identical on replay. Full race tests, vet and formatting passed.

After other tests finished, an M4 Pro CPU replay of source verification, features, policies and aggregation took1.44s wall,1.46s user and96,878,592 bytes peak RSS (about92.4MiB). This is not per-inference or GPU cost. Three fixed synthetic benchmarks of3,009 candidates, reverse-order hints and a20-candidate prefix took7,766–8,005ns,27,648B and2 allocations per ordering. Feature extraction, index construction, helper search and IPC are excluded. CPU profile samples were dominated by runtime waiting/scheduling; they do not establish a search bottleneck. Raw profiles stay local.

```sh
go run ./cmd/riido-deferred --out .cache/deferred-helper-new
```

[Aggregate results](results-27.json). No new weights or HF release. The16/20-feature heads are our own linear models, not Laya or ternary checkpoints. This is repeatedly observed development data from three repositories; two seeds are not independent data. Not production ready. No LLM savings established. Fresh real-request evaluation of at least2,400 and actual agent completion/cost measurement remain outstanding.

Next audit separate real-task sources for licensing, duplication, answer leakage and feasible evaluation rather than tuning away the9-page gap on this corpus. Establish whether public issue tasks better represent the intended usage; their patches are not model-routing or decomposition ground truth.
