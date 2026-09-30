# Connected groups and2,400 evaluation-only representatives — experiment29

**Freeze2,400 representative requests from2,455 connected components.** Do not train on this corpus or replace difficult requests after observing outcomes. This prepares evaluation membership; it does not establish model quality.

## Selection

Under the [precommitted plan](plan-29.json), connect rows sharing an instance ID, normalized request string, or repository/pre-fix commit. Connections are transitive: if A/B share a request and B/C share a snapshot, all three form one group. Go uses sorted arrays and an array union-find.

Choose the lexicographically smallest source-tagged ID in each group. Hash its sorted member IDs, sort components by this signature, and take the first2,400. No answers, difficulty, performance or request length enter selection. All source rows remain evaluation-only, including unselected components and siblings; they are not training/calibration data.

| Item | Result |
|---|---:|
| Source rows | 2,594 |
| Connected components | 2,455 |
| Non-singleton components | 128 |
| Largest component | 5 rows |
| Selected representatives | 2,400 |
| Distinct requests / IDs / repository-commits | 2,400 each |
| Selected repositories | 53 |
| Full / Multilingual source | 2,107 / 293 |

Component sizes are2,327 singletons,121 pairs,4 triples,2 groups of4 and1 group of5. Selected representatives do not overlap under these three exact connection rules. **This does not prove independence across different commits of a repository or semantically similar requests.** Repository imbalance remains, including806 Django requests. Evaluation must report repository-level results and account for correlations.

The selected membership SHA-256 in the [report](results-29.json) is `fa58c4e51ae307963eb587813e49adf70d987dcb4893f6ee1aa6050bc6b80a5a`. Individual selections contain no request text and remain in a local0600 file. Git excludes individual IDs/hashes. Future source-eligibility changes require an explicit new plan and explanation, never silent outcome-based replacement.

## Input constraint discovered

Selected raw requests total4,363,135 bytes; the longest is71,136 bytes. **53 exceed the existing search API's8,192-byte limit.** They were neither removed nor silently truncated. A subsequent input contract must distinguish the full request from bounded model/search views and report truncation, multiple windows, costs and failures. No long-input strategy has yet been validated.

## Source access preflight

For each represented repository, choose the lexicographically first selected identity's pre-fix commit and fetch only GitHub root-tree metadata. Use two workers,30-second request timeouts, a2MiB response cap and local caching. Existing GitHub CLI authentication handles credentials; neither credentials nor raw error messages enter public reports.

All53 sampled roots were accessible, totaling331,899 response bytes.51 contain root filenames matching the LICENSE/COPYING/NOTICE candidate rule. `matplotlib/matplotlib` and `tokio-rs/axum` do not. **This does not mean they lack licenses; nested paths and source terms still need review.** No candidate file contents were downloaded or legal conditions determined.53 samples do not verify all2,400 snapshots. [Preflight report](source-preflight-29.json).

These small responses do not estimate full-source, recursive-tree or container storage. This is an access check without downloading or executing code.

## Verification, usage and next steps

Tests cover transitive cross-key connections, cross-source ID links, order invariance, lexical representatives, input preservation, strict2,400 shortfall behavior and retention of long requests. Preflight recomputes and checks the frozen selection, rejecting API errors, incomplete trees and duplicate paths. Aggregate grouping and the private selection file both replay byte-identically; cached preflight does too. Full race tests, vet and formatting passed.

One M4 Pro CPU grouping run took0.53s wall,0.16s user and31,522,816 bytes peak RSS (about30.1MiB), excluding network preflight, model inference and GPU costs. Existing commands, ranking and sessions remain unchanged.

```sh
go run ./cmd/riido-taskgroups --out .cache/evaluation-groups-29
go run ./cmd/riido-sourcepreflight --selected .cache/evaluation-groups-29/selected.json --out .cache/source-preflight-new.json
```

Pinned projections are required and the first output directory must be new. Keep `selected.json` local. Next verify target source conditions/pre-fix file catalogs, define bounded long-request handling and fix task-appropriate evaluation metrics. No new training, inference or HF release; existing final reserves remain unused. **This reduces some evaluation-dependence risks; it does not establish actual agent savings.** Not production ready.
