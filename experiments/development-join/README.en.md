# Joining development targets to pre-fix catalogs — experiment 40

**Among 1,273 available catalogs, 1,267 tasks had usable old-file targets; all 4,841 checked paths matched regular files at the corresponding snapshot. Another 11,748 of the 13,021 development tasks lack catalogs and remain unverified.** This checks target/candidate compatibility, not search accuracy, training or model quality.

After committing the [plan](plan-40.json), we waited for experiment 38's first continuation to terminate normally at 2,000 new requests. Including prior caches, the [collection checkpoint](collection-checkpoint-38.json) contains 1,273 roots/complete catalogs and license-text candidates at 1,261 snapshots. Candidate text is not license approval. A cache-only replay reproduced the collection aggregates, private inventory and failure list exactly.

The join visits all fixed 7,335 training and 5,686 validation tasks. Final 2,402 tasks and unselected siblings are excluded. Experiment 39's pinned patch projection is verified before labels are decoded. Missing catalogs and unavailable targets are not silently dropped.

| Join checkpoint | Count |
|---|---:|
| All development tasks | 13,021 |
| Complete pre-fix catalogs available | 1,273 |
| Catalogs unavailable | 11,748 |
| Supported old-path target tasks in the full cohort | 12,669 |
| Tasks with both catalog and usable targets checked | 1,267 |
| Checked tasks with every target mapped to a file candidate | 1,267 |
| Checked old paths / regular-file matches | 4,841 / 4,841 |
| Missing paths, directories, gitlinks, symlinks among checked targets | 0 each |
| Full-cohort parse failures / no-old-path tasks | 294 / 58 |

[Per-repository report](results-40.json). The six cases represented by 1,273−1,267 have catalogs but no usable old-file targets. The counts 294 and 58 refer to **all 13,021 tasks**, not those six or the available 1,273. Acquisition follows a fixed role/repository/ID order; this partial set is neither random nor representative evidence of overall quality.

Go binary-searches validated sorted path arrays. Regular/executable blobs, symlink blobs, gitlinks, directories and absent paths remain distinct. Regular and symlink names count as file candidates consistently with existing search evaluation, but link targets are never opened or followed. Gitlinks and directories are not converted into ordinary-file targets. Parse failures, unsupported target blocks and absent old targets cannot count as complete mapping successes.

The cache-reader callback never accesses the network. Only absent caches become explicit unavailability; corrupt compressed/JSON data and identity mismatches stop execution. Existing bounds remain: 2MiB roots, 32MiB stored catalogs, 100,000 entries/16MiB paths and the existing patch limits. Catalogs are handled one snapshot at a time, without a new global cache, lock or concurrent cache mutation.

Private evidence records the fixed role/ID/pre-fix commit, raw-root digest, tree ID, normalized catalog digest and target matches. The evidence file SHA-256 is `3452e20f925c03a79c5b0527c7decf58efcba00b6bbe78e50fad3e316b32b7c2`. Only code, plans, repository aggregates and digests are public; source text, target paths and per-task evidence are not published.

```sh
# Offline join using pinned inputs and currently available local caches
go run ./cmd/riido-developmentjoin --out .cache/development-join-new
```

With acquisition stopped at the same checkpoint, normal execution, replay, full race execution and an isolated measured execution produced byte-identical aggregates and private evidence. Acquisition subsequently resumed, so future runs may find additional catalogs. The documented 1,273 is a completed checkpoint, not live state. Synthetic checks cover absent/corrupt caches, wrong snapshot reuse, path kinds/order/duplicates, unusable targets and full denominators. Full race/vet/format/redacted scans gate publication.

After acquisition and other checks finished, one isolated M4 Pro CPU run covering all 13,021 inputs, 1,273 catalogs, joins and output took **2.79s wall**, 2.99s user and 0.15s system, with **65,699,840 bytes peak RSS (about 62.7MiB)**. Reading the patch projection is included; Parquet conversion, network, model inference, GPU and training are excluded. This is not single-inference latency or evidence of a speedup.

`AllCatalogsAvailable=false` and `AllSupportedTargetsVerified=false`. Next, extend acquisition using the same caches and verify pending joins. Historical licensing, issue-text permissions and training approval remain separate and unresolved. Targets remain learning/evaluation outputs, never inference features. No model training, final evaluation or new Hugging Face release occurs here.
