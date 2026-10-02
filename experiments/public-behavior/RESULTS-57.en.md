# Public behavior verification57 results

**Collected62 actual original Go API observations:59 matched prewritten finite expectations and3 disagreed.** Two cover the same large-prerelease ordering failure in opposite directions. The other contrasts early glob success with a stronger validation policy. These are not model accuracy, general function correctness or2400 independent task results.

## Findings

| Observation | Prewritten expectation | Original result | Meaning |
|---|---|---|---|
| `1.0.0-99999999999999999999` vs `1.0.0-100000000000000000000` | SemVer numeric precedence: smaller(-1) | Larger(+1) | Both strict parses succeed; uint64 overflow enters the comparison's string branch. Preserve this finite normative discrepancy. |
| Reversed pair | +1 | -1 | Confirms the same pattern, without adding an independent bug or task. |
| `Match("{a,[}", "a")` | Stronger whole-pattern policy: false + ErrBadPattern | true + nil | The first alternative succeeds before the malformed later alternative is checked. Do not generalize this into a universal upstream API obligation. |
| `ValidatePattern("{a,[}")` | false | false | A separate observation. No validation gate was inserted before Match and no result was synthesized. |

These findings show why broad candidate claims such as numeric version comparison or pattern-error handling need **small claims with explicit input scope and failure conditions** connected to real verification. Tiny-model semantic understanding and cost savings remain separate questions.

## Counts and independence

- Strict parse16, metadata comparison5, prerelease comparison9, path matching9, escape/character matching8, error matching6 and pattern validation9:62 operation observations.
- Requested entry APIs90: StrictNewVersion44, Compare14, Match23, ValidatePattern9. Getter/String/Original calls224 are separate. Upstream internal helper calls are not counted.
- Normative expectations14:12 matched/2 disagreed; descriptive API46:all matched; stronger policy2:1 matched/1 disagreed. Known library errors17 were accurately observed; unknown/panic/unsupported0.
- Verified28 Git blobs before observing. Official collection1, official retries0, exit0. A plan preparation attempt rejected the symbolic `HEAD` value with zero observations; it is not an official retry.
- Two source families and six finite property contracts. Inputs, aliases, reversed pairs and contract variants do not add independent final tasks. New model inputs/captions, ranking evaluations, fits, paid calls, weights and protected-final tasks0.

Preserved15 original MIT files totaling99,035 bytes and both full licenses. Original module declaration bytes live under `metadata/upstream.go.mod.txt`; parent `go.mod`/`go.sum` are unchanged. These upstream behaviors already exist and are not six newly completed coding tasks. See the [source manifest](../../internal/publicbehavior/upstream-manifest.json), [prewritten plan](PLAN-57.en.md) and [authoring/review record](oracle-review-57.json).

## Freezes and reproduction

| Evidence | Bytes and SHA256 |
|---|---|
| [Probes](probes-57.json) |218,192 / `c4deb9d9e33dca8ff4498531d898ea82ee05559727cbc4f00d99455910722574` |
| [Execution plan](execution-plan-57.json) |5,472 / `dfc17915b37d4ce007f48a3a8d4e6fcc693f926f6a789562f6a8fcc68c347d7b` |
| [Original observations](results-57.json) |262,559 / `850c60266eceb613565390430adca12cea9b1a56207b803b06cb681f563978c5` |
| Actual Darwin/arm64 Go1.27.1 CGO0 binary, excluded from Git |5,096,818 / `09e5f6079b8a436c432d313a884376a5bde2267e7afdde88c5d422f1777cc5d7` |

Source freeze `9e99914f6d1e35aa9d97413f80fc6e03768e97ce`; input freeze `1462c705087f24666049f8bd31d317490c05dbeb`. The three JSON files total486,223 bytes. Verified19 embedded source/upstream artifacts and7 support artifacts before observing. Post-collection `publication_test.go` and `audit_test.go` are not retroactive pre-collection pins. Regression replay adds no official observations or independent samples. [Usage](USAGE-57.en.md) explains platform boundaries.

Scheduling1 and heap soft limit256MiB are settings, not measured RSS/CPU/GPU/latency. This run executed no Laya/GPU and adds no dependencies or locks to the production hint path. Preserve earlier56a/d measurements and pending/unknowns. Cache and score-metadata suggestions in the [Laya retrospective](../../docs/laya-source-retrospective-57.en.md) are not achievements of this run.

Next acquire accurate requests/captions and candidate relationships across more sources, then freeze and assess model-free ranking utility while preserving every candidate. Do not claim240 acquired sources, minimum15-group/5% necessary utility or2400 distinct protected-final requests per domain. **Training eligibility remains false**; no new HF weights were released.
