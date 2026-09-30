# Scope and input contract for the next numerical head — experiment 41

**We fixed 16 runtime features for file-search hints and added an opt-in Go path for long queries and large candidate catalogs.** Existing default API limits remain unchanged. This defines inputs and operational scope; it is not new model training, blanket legal clearance or an HF release.

The [scope plan](plan-41.json) and [feature plan](path-features-41.json) were committed before implementation. The intended head offers fallible hints about auxiliary search using runtime numbers; it does not authorize decisions or discard fallback candidates.

### Provenance facts and remaining conditions

The pinned official SWE-bench [README](https://github.com/SWE-bench/SWE-bench/blob/02e7a74ffd0b707aab73d203fe87bdc7c76afc8e/README.md) describes training on preprocessed datasets and identifies MIT licensing. We verified the [LICENSE](https://github.com/SWE-bench/SWE-bench/blob/02e7a74ffd0b707aab73d203fe87bdc7c76afc8e/LICENSE) and rehashed the fixed training card, Parquet and membership. [Evidence inventory](evidence-41.json).

Current [GitHub D.6](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service#6-contributions-under-repository-license) addresses repository-license contributions and superseding agreements. GitHub's own AI-training grant is not our grant; current terms do not retroactively prove every historical issue author's permissions.

A missing card license field alone therefore does not establish a training prohibition. These facts do not substitute for every upstream condition. The historical Prefect sample's [README](https://raw.githubusercontent.com/PrefectHQ/prefect/c36f1a2bc58e1e3b18e2a63f8a6ac1d92e045f10/README.md) also yielded no license grant; today's Apache classification does not resolve that gap. We record references/identities without copying or publishing that README.

### Operational scope

The [machine-readable policy](policy-41.json) separates operational decisions from rights verification.

| Material/action | Current decision |
|---|---|
| Fixed development audits and numeric input implementation | Proceed |
| Local fitting of original numerical heads | Defined as permitted scope only for source-eligible inputs under a precommitted experiment; no new fit yet |
| Unknown/conflicting historical conditions | Retain pending status, never silently mark eligible |
| Redistribution of raw issues/code/patches | Not approved |
| Raw-text generation or upstream encoder fine-tuning | Outside this narrow review |
| New HF model publication or default activation | Requires separate source/output/quality review |

Membership remains 7,335 training, 5,686 validation and 2,402 final tasks. Pending permissions/catalogs/targets remain in coverage denominators and cannot move roles. Any fitting exclusions require a reason recorded before model results. CI is a technical gate, not a warranty covering all rights.

### Go input contract

`searchclaim.PathFeatures` uses the new `riido-path-claim-v1` schema. It accepts valid UTF-8 queries up to 128KiB and baseline rankings up to 100,000 path candidates. Existing `Features` keeps its 8KiB/4,096 limits and `ScoreSpread` keeps its 4,096 limit.

The 16 values combine the existing 12 numeric features and four top-20 score-distribution features: bias, word/byte counts, uppercase/digit/underscore fractions, camel boundaries per word, unique-word fraction, squashed top score/gap, top-ten/full-catalog positive-score fractions, and relative top-20 mean/stddev/first-score share/last-to-first ratio. All are finite in [0,1]. The byte feature saturates at the legacy 8KiB scale for long queries; the query itself is not truncated.

Gold paths/ranks, auxiliary rankings and repository IDs are not function inputs. Token/path identities do not become model features. Rankings are expected to come from the baseline index; the function does not certify arbitrary external permutations. Existing small-domain weights are not automatically attached to the new API.

Synthetic checks establish exact short-input equivalence to the original two functions, long Korean queries/10,000 scores, 128KiB/100,000-item boundaries, invalid UTF-8/scores/bounds and preservation of caller arrays. Full race/vet/format/redacted scans gate publication. Acquisition runs separately, so no performance, memory improvement or SIMD claim is made.

Before fitting, fix source eligibility, cost objectives, controls, seeds, resource budgets and validation selection. Establish useful FP32 signal before independently measuring ternary variants. Do not repeatedly tune against final evaluation. Actual savings and model utility remain unproven.
