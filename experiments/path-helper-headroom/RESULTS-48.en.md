# Experiment 48: training-only helper headroom screen stopped before fitting

All three fixed new helpers failed the unchanged **5% page-reduction necessary condition** on **4,456 eligible training requests across 20 repositories**. Even an oracle that knew every target and chose only beneficial helper outcomes could not reach 5%. Additional claim fitting for these helpers stops here. This screen did not score new validation or protected final outcomes, invoke models, fit coefficients, produce weights or activate a policy. Experiment 46 remains a failed experiment.

The [precommitted plan](plan-48.json) fixes the three rules and training-only scope. The [aggregate result](results-48.json) records all policies and repository totals; the [resource observations](resources-48.json) describe the two actual offline runs.

Pages are **pages of 20 candidates until the first mapped target file**, a file-search cost proxy. Net reduction is `(baseline pages − helper pages) / baseline pages`. Oracle headroom is `sum(max(0, baseline pages − helper pages)) / baseline pages`: it ignores every harmful switch and knows gold labels. It is an optimistic diagnostic, not an executable routing policy, and ignores the other quality, call and repository constraints. With **43,914 baseline pages**, the 5% gate needs at least **2,196 saved pages**. Passing this training screen would only justify a separately precommitted next experiment; it would not establish validation usefulness.

| Fixed policy | Pages | Net saved pages | Net reduction | Oracle saved pages | Oracle headroom | Wins / losses / ties |
|---|---:|---:|---:|---:|---:|---:|
| Baseline | 43,914 | 0 | 0% | 0 | 0% | 0 / 0 / 4,456 |
| `normalized-positive64` | 42,816 | 1,098 | 2.5003% | 1,357 | 3.0901% | 125 / 255 / 4,076 |
| `joined-fields-rrf64` | 42,823 | 1,091 | 2.4844% | 1,615 | 3.6776% | 217 / 514 / 3,725 |
| `explicit-path64` | 43,365 | 549 | 1.2502% | 598 | 1.3618% | 86 / 37 / 4,333 |
| Legacy control, `legacy-normalized-full` | 41,894 | 2,020 | 4.5999% | 3,030 | 6.8998% | 184 / 550 / 3,722 |

The legacy control's 6.8998% **training oracle** does not overturn [experiment 46's validation failure](../path-cost-claim/RESULTS-46.en.md). Its actual training net reduction is 4.5999%, and an oracle upper bound alone proves no usable claim selector. Experiment 46 already observed a validation oracle bound of 3.113% and failed its fixed usefulness gates. This screen does not retune those gates or select a new validation winner.

The rules rank paths without looking at target labels; labels enter only the later cost calculation. They supply small lists of file hints rather than generated sentences:

- `normalized-positive64` uses the existing identifier splitting and full-path BM25 ranking, keeps at most 64 paths with strictly positive scores, and alternates baseline and auxiliary candidates with baseline first.
- `joined-fields-rrf64` retains lower-case joined identifiers alongside split identifiers, ranks full paths and basenames separately, then adds fixed reciprocal-rank scores `1/(60 + one-based rank)` from positive results in each field. It keeps at most 64 fused hints with deterministic ties and the same baseline-first interleave.
- `explicit-path64` looks for complete query tokens containing a slash or dot that exactly match a case-sensitive catalog path or a suffix at a slash boundary. More matched path segments come first, followed by longer tokens and catalog order; no exact evidence leaves the baseline unchanged. It keeps at most 64 hints and uses no BM25 index.

Quality does not rescue the failed headroom screen. Equal-repository Hit10 is **42.1142%** for baseline, **42.2723%** for positive64, **41.4535%** for joined RRF and **44.2129%** for explicit paths. Explicit paths improve this training metric while still lacking enough page headroom; joined RRF lowers it. These are development aggregates, not task completion or LLM usage measurements.

All **7,335 fixed training members** remain in coverage: **4,456 eligible**, **221 pending source**, **2,630 unavailable** and **28 invalid cost**. Excluded members were not replaced or silently removed from the denominator of coverage. The 4,456 requests, rather than repeats or policy variants, define the evaluated sample. No new validation rankings or scores were computed, and the **2,402 protected final members remained unscored**. The sealed input retains earlier validation metadata; that metadata is not a new validation result.

Every new helper was attempted on all 4,456 requests with **zero fallbacks**. Positive64 emitted 274,050 hints, built 3,568 indexes and made 4,456 index searches. Joined RRF emitted 274,332 hints, built 7,136 field indexes and made 8,912 field searches. Explicit paths emitted 6,088 hints and performed 134,445,953 anchor comparisons, with no index build or search. The shared catalog cache recorded **3,568 builds and 888 hits**. Fewer hints or higher Hit10 alone does not measure avoided helper work or downstream savings.

The original Go implementation uses caller-owned arrays and immutable indexes without query locks. Bounded construction maps remain; this is not a claim that the whole implementation is map-free. Complete UTF-8 queries are limited to 128 KiB; unique canonical path catalogs to 100,000 paths and 16 MiB; each aligned text field to 16 MiB, 262,144 lexical tokens and a 131,072-term vocabulary. Explicit matching checks the unique-anchor × path count bound of 8,000,000 before comparisons. Auxiliary bounds or errors fall back to the unchanged baseline with a recorded reason, rather than truncating input or dropping tasks. No SIMD performance benefit was measured or claimed.

| Actual full offline observation | Wall seconds | User CPU seconds | System CPU seconds | Peak RSS bytes | Peak RSS MiB |
|---|---:|---:|---:|---:|---:|
| First run | 82.84 | 89.67 | 2.64 | 111,902,720 | 106.71875 |
| Exact replay | 82.30 | 90.25 | 2.46 | 112,492,544 | 107.28125 |

Both observations include the **whole offline source/cache verification, index construction and 4,456-request screen** and meet the **256 MiB peak RSS target**. They are two full-run observations, not single-request inference latency, p95 estimates or Go heap pprof measurements. No GPU was used. No LLM calls, tokens, retries, actual helper skipping or money savings were measured.

The actual replay comparisons for **private case output and public aggregate output both returned exit 0**, establishing byte-identical files in those two comparisons. Repeated execution is a reproducibility check, not another independent sample. Private case contents, source bodies, paths, queries and target labels are not included in this report.

| Evidence pin | Value |
|---|---|
| Clean runner revision | `828d2c8dec00d144de7ddec96c27b6e20fe74311` |
| Runner binary SHA256 | `c211570fbf7aa64e58488105b6013848ba66eac158cdaf19fad861599d220da8` |
| Plan SHA256 | `f7e740f0ff0da3706592f79d30d240779d8e1261504ef4f4f4dec937904e61fd` |
| Input seal SHA256 | `3e8793cb0f3c5d23473fedfd3d4faed58c05dbd3949e77308ca1f8ee56f39187` |
| Aggregate result SHA256 | `e23d6a76cdee999f119471882d31ea8170ea281c36fa60f566d847efbb7c7b4a` |
| Rankings SHA256 | `63f0f8ad72f3753736f34755054422d160b16eaf8205259c2534d799fa14ef5d` |

The design was informed by identifier and retrieval concepts in Apache-2.0 [laya-codex at the fixed revision `580bc73c2ed95fd319db93ef725f30bf35047428`](https://github.com/pilotspace/laya-codex/tree/580bc73c2ed95fd319db93ef725f30bf35047428) and the [original 2009 reciprocal-rank-fusion paper](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf). The Go implementation was independently authored; no upstream code or GPL Spiral code or tables were copied. Numeric-source evidence does not grant permission to release source text, training data or models. Ternary compression remains a separate later lane and cannot repair the failed helper usefulness screen.
