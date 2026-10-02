# Results59: text and literal-source references prepared

[한국어](RESULTS-REFERENCES-59.ko.md) · [Usage](USAGE-REFERENCES-59.en.md) · [Preregistered plan](PLAN-REFERENCES-59.en.md) · [Official original](caption-coverage-59.json)

[Collection/failure ledger](collection-ledger-59.json) · [Independent reference review](reference-review-59.json)

**One official metadata generation connected existing requests and candidate captions to their exact review locations. Content approval and training readiness remain pending.** A reviewer can now find a caption's original JSON location and text hash alongside its contract's Go literal-expression location. The generator executed no candidate implementation or model.

The official attempt count is **1, with 0 retries and exit code 0**. Its state is `references_generated_content_review_pending`, with `preparation_only=true`, `content_review=pending`, `literal_payload_reified=false` and `training_ready=false`. Public `caption-coverage-59.json` is a byte-preserving copy of the original envelope. “Coverage” in its filename identifies reference preparation, not a semantic coverage approval rate.

## What was connected

| Item | Official result |
|---|---:|
| Stored metadata Bind attempts / completed | 1 / 1 |
| Go AST parse attempts / completed | 3 / 3 |
| Existing parent requests | 72 |
| Existing candidate caption positions | 216 |
| Text references: 72 requests + 216 captions | 288 |
| Raw contract-expression references | 18 |
| Git blobs verified before generation | 21 |

AST parsing reads source syntax. It does not execute literal expressions, reconstruct Input/Want values or acquire new candidate Got values or failure locations. Historical truth, candidate checked/failed counts and acceptable indices remain metadata linked to stored records. They are neither new labels nor scorer features.

Text hashes cover exact decoded UTF-8 bytes. Raw expression hashes cover original Go source ranges `[start_byte,end_byte)`. These differ from historical serialized literal-table hashes, formatted candidate-declaration hashes and bundle hashes covering helpers/types/sentinels. Hash agreement alone does not approve complete source closure or caption meaning.

## Preserved denominators and pending reviews

| Preserved evidence | Count |
|---|---:|
| Answerable / no-answer / unknown | 34 / 17 / 21 |
| Whole connected groups / groups containing known parents | 17 / 16 |
| Parents with 2 / 3 / 4 candidates | 12 / 48 / 12 |
| Prototypes | 18 |

Unknown-only group 64 remains in the complete graph; its four parents were not converted to rejected or no-answer labels. Duplicate candidate positions and original order remain unchanged. Parents, captions, text references and expressions are distinct units; they are not added together as new independent requests or final samples.

| Review still outstanding | Pending |
|---|---:|
| Request-contract review | 72 |
| Candidate source closure | 216 |
| Candidate caption source fidelity | 216 |
| Candidate caption request-contract coverage | 216 |
| Contract observation-fields review | 18 |

A caption may faithfully describe an incorrect implementation. Source fidelity therefore differs from satisfying the request. The [content-review recipe](content-review-recipe-59.json) defines the next inspection of text, contracts, observable fields, negation/boundaries and supporting evidence. This generation created no completed content-review labels.

The [independent reference review](reference-review-59.json) verified 288 JSON text locations totaling 28,787 UTF-8 bytes, 18 raw-expression ranges/physical lines, 148 historical vector expressions, and preservation of 17 groups and 124 relation edges. All 738 review entries above remain pending. The reviewer made no Generate, Bind, source API or test calls; this pass establishes reference integrity only.

## Frozen provenance and attempt history

| Original provenance | Value |
|---|---|
| Source commit | `75a9776230f7eaef3294247cc10bb7f13342f102` |
| Input commit | `69e9ddc51e218da029572e2bb463500dc36cd143` |
| Execution plan | 4,681 B; SHA256 `a321d625e986ec7e621b827a0616b7492291f624bd395bc32b424a0ad1178819` |
| Official executable | 5,202,770 B; SHA256 `03efc5808316128e0c336887907ee61d4aa9bab59cc764b8401e53d4786a21ec` |
| Original result | 360,591 B; SHA256 `7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29` |

The original runtime was Go 1.27.1 / Darwin arm64, CGO 0 and a trimpath build. Before generation, the tool verified Git bytes for five compiled Go sources, nine support files, six inputs and one execution plan. Official plan preparation ran once with Bind 0/Generate 0. A new directory and exclusive `results.json` reserved the output. This prevents accidental path reuse; it is not a host-global single-run lock.

The [prototype ledger](prototype-ledger-59.json) preserves two separate preparation executions and two Bind calls. The first panicked before output because three manually entered prototype names differed from historical names; zero output and the missing historical executable hash remain explicit. After correcting the names, the second produced private references with content review still pending. Neither attempt adds official executions or new truth.

Full race/vet and public-artifact checks passed before the official result. The CLI's initial two `GOTOOLCHAIN=local` race/vet checks stopped at the default Go 1.27.0 version guard before tests ran; installed Go 1.27.1 checks then passed. After the result, the [frozen regression test](../../cmd/riido-captionref/frozen_test.go) and local full race/vet passed. Stored-metadata replays add no official attempts or independent data. The [collection ledger](collection-ledger-59.json) separates three prototype-plus-official collection Bind calls from later replays. The later `frozen_test.go` is not retroactively part of the original nine support files, and this document does not declare final GitHub CI complete at its writing point.

## Practical next use and remaining boundaries

[Source-transfer preparation](SOURCE-TRANSFER-59.en.md) contains two scoped proposals: semver strict parsing and glob slash boundaries. It links 15 retained upstream artifacts and 25 saved experiment 57 observations with 287 Want/Got field references. New downloads, API observations, parents, captions, truth and acceptable sets remain 0. Upstream authorship differs from this project's AI-assisted synthetic caption process. External code or another reviewer alone does not clear `synthetic_single_pipeline`.

The [role recipe](role-recipe-59.json) is `unassigned`. It records a future whole-component allocation rule; membership/seed freezes, coverage and actual assignments are still absent. Existing train 9 / validation 3 / calibration 3 floors remain, and transfer 0 is permitted. Observed experiment 58 scores must not choose favorable groups or seeds. Role names cannot turn existing evidence into unseen or blinded validation.

`no_roles_plan` and `synthetic_single_pipeline` remain. New labels, independent requests, roles, rankings, source APIs, model/paid calls, fits, weights, protected-final access and activation are all 0. This tool itself does not run Laya or change historical Laya CI evidence. Go scheduling 1 and the 256 MiB heap soft limit are settings, not measured RSS/CPU/GPU/latency or LLM token/financial savings.

At least 2,400 distinct protected-final requests per domain remain a separate generalization target, not a prerequisite for every development fit. Reference generation does not fill that target. Subsequent content review and a separate training plan follow the existing automatic checks and CI policy without adding a human-approval flow.
