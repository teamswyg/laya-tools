# Historical license text acquisition — experiment 31

**At least one license candidate text was verified for every one of the 2,400 frozen evaluation snapshots.** This establishes acquisition, not complete legal review, model quality, or blanket training/redistribution permission.

## Scope and results

Follow the [precommitted plan](plan-31.json), recompute experiment 29 membership and retain all 2,400 tasks. Only the 170 snapshots without root candidates receive bounded nested lookup: immediate files in Matplotlib's LICENSE directory, or license-name files in Axum's axum and axum-* packages. Do not recursively scan arbitrary code or vendor paths.

| Item | Result |
|---|---:|
| Selected snapshots / snapshots with verified text | 2,400 / 2,400 |
| Snapshots requiring nested lookup | 170 |
| No candidates / tree failures / blob failures / symlinks | 0 each |
| Per-snapshot text references | 4,941 |
| Unique repository/path/object tuples | 169 |
| Distinct text objects / total text bytes | 159 / 820,911 |

Publish [repository aggregates](results-31.json); keep individual tasks, commits, text digests and original bodies local. The 169 tuples are not license types. Finding at least one text does not establish coverage of every file in a snapshot.

## Integrity and caching

Match subtree IDs to parent entries. Verify blob response ID, declared size and base64, then recompute Git blob SHA-1 from content and record a separate SHA-256. Bound text to 256KiB, valid UTF-8 without NUL, and reverify cached bodies. Do not follow symlinks arbitrarily.

Use one request at a time, minimum 250ms start spacing, 30-second timeout and 2MiB API response limit. Persist only verified bodies through a temporary file and atomic rename. Require a new output directory.

```sh
go run ./cmd/riido-licensetext --out .cache/license-texts-new
go run ./cmd/riido-licensetext --offline --out .cache/license-texts-offline-new
```

Pinned source projections and previous root caches are prerequisites. Offline mode never fills missing entries over the network. Both aggregate and private inventory match the original collection byte for byte, including an actual 2,400-snapshot replay under the race detector. Run full Go race/vet, formatting and redacted secret checks.

One M4 Pro CPU cache replay took 0.63s wall, 0.39s user, 0.13s sys and 29,130,752 bytes maximum RSS (about 27.8MiB). This excludes initial network collection, model/GPU execution and full source storage.

## Source conditions

The acquired Matplotlib text permits analysis/testing with notice preservation and a change summary for distributed derivatives. Separate font/image terms are also present, so a single repository-wide assumption is insufficient. Axum package text includes MIT; the selected historical Terraform text is MPL 2.0. Do not substitute a current repository license for historical conditions. These are representative text checks, not completed legal review of all 159 bodies.

- MIT requires preservation of copyright and permission notices in copies or substantial portions. [Official text](https://opensource.org/license/mit)
- Apache 2.0 redistribution includes license-copy, modification-notice, attribution and applicable NOTICE requirements. [Official section 4](https://www.apache.org/licenses/LICENSE-2.0)
- Mozilla distinguishes internal use from external distribution; distribution obligations include source availability for covered files. [Official MPL FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/)

Prepare subsequent evaluation with sources kept local and only aggregate findings published. Repository code licensing does not automatically resolve issue-author text or every third-party file. Raw corpus/source bundles and trained-model publication each require applicable-condition review. This experiment performs neither training nor model publication.

## Next

53 of the 2,400 requests exceed the existing 8,192-byte input bound. Keep these tasks and precommit input-window and cost accounting before constructing pre-fix file catalogs and evaluating helpful/harmful hints. Source preparation must not substitute for quality evaluation. Production readiness remains false; existing final reserves remain unused.
