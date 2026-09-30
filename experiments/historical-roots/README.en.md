# Historical root access for all2,400 selected tasks — experiment30

**All2,400 frozen pre-fix snapshots have accessible root-tree metadata.** This extends the previous53-repository sample to every selected snapshot.2,230 contain root license-name candidates;170 require nested-path inspection. This is not completed license-text review or model evaluation.

## Scope and results

Following the [precommitted plan](plan-30.json), recompute experiment29's selection from pinned projections. Verify2,400 entries and membership SHA-256 `fa58c4e51ae307963eb587813e49adf70d987dcb4893f6ee1aa6050bc6b80a5a`, then query each exact `base_commit` root tree. Retain every selected request, including the53 long inputs.

| Item | Result |
|---|---:|
| Target / accessible snapshots | 2,400 / 2,400 |
| Unavailable snapshots | 0 |
| Snapshots with root license candidates | 2,230 |
| Snapshots needing nested-path inspection | 170 |
| Metadata response bytes | 14,752,218 |
| Deduplicated candidate objects | 143 |

The170 are163 `matplotlib/matplotlib` snapshots and7 `tokio-rs/axum` snapshots. Further inspection of previously sampled roots shows a `LICENSE` directory in Matplotlib and a multi-package layout in Axum. **Absence of a matching root file does not mean absence of a license.** Nested lookup is a subsequent step.

The LICENSE/COPYING/NOTICE filename rule is unchanged from experiment29.143 counts distinct `(repository, path, blob ID)` tuples, not143 license types. Object references and occurrence counts stay in a private local manifest. The [public repository aggregates](results-30.json) contain no individual task or commit IDs.

## Execution and caching

Two workers own separate input/result slots. Bound requests to30 seconds, responses to2MiB and error capture to4KiB; space each worker's request starts by at least250ms. Existing GitHub CLI authentication handles credentials. Credentials and raw errors never enter public results.

Only validated successful responses enter the cache through a temporary file and atomic rename. Reject incomplete trees, invalid object types/modes/IDs, duplicate paths and oversized responses. Do not replace failures with easier tasks. Reuse the original53 sample responses. This cache is source metadata, not license clearance.

An `--offline` replay makes no GitHub requests. Missing cache entries remain unavailable observations. All2,400 entries were present here; the aggregate and private candidate manifest are byte-identical to the collection run.

```sh
go run ./cmd/riido-rootcoverage --out .cache/historical-roots-new
# After collection, reproduce into a new directory without network requests.
go run ./cmd/riido-rootcoverage --offline --out .cache/historical-roots-offline-new
```

Pinned projections are required. Output directories must be new. Keep `candidates.json` and the source cache local; do not upload them to Git/HF.

## Verification, resources and interpretation

Injected-response tests cover cache reuse, failure non-persistence, malformed identities and bounded output. Test candidate deduplication and per-repository success/failure/missing-candidate accounting. Full Go race tests, vet and formatting passed; the original53-sample report still reproduces. An actual2,400-entry offline run under the race detector also matches the original aggregate.

One cache-only M4 Pro CPU replay took0.93s wall,0.44s user,0.10s sys and32,440,320 bytes peak RSS (about30.9MiB). **This does not measure the entire initial network collection**, model inference or GPU costs.14.8MB of metadata does not estimate full source or container storage.

The result establishes **root-list access at every selected commit**. No code or license contents were downloaded and recursive catalogs were not checked. Candidate files still require text and third-party-condition review. `LicensesReviewed` and `ProductionReady` remainfalse.

Next inspect nested license paths for170 snapshots, verify deduplicated candidate contents and document source conditions. Long-input handling and actual2,400-task evaluation remain pending. No training, inference, new HF release, default activation or use of existing final reserves occurred.
