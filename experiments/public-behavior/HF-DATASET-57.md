---
license: apache-2.0
size_categories:
- n<1K
tags:
- go
- finite-behavior
- semver
- glob
- development-only
configs:
- config_name: finite_observations
  data_files:
  - split: development
    path: data/observations.jsonl
---

# riidolaya public behavior57 / 공개 동작 관측57

62 finite operation observations from two pinned public Go source families.59 match prewritten expectations and3 disagree: two reversed observations of one large-prerelease numeric ordering discrepancy and one stronger glob-validation policy difference. These are **not62 independent coding tasks, model accuracy, a held-out final set or LLM savings**. New ranking, fitting, model calls, weights and protected-final reads0; training eligibility remains false.

공개 Go 원천 가족2개의 유한 동작62건입니다. 사전 기대값59건과 일치하고3건이 다릅니다. 큰 버전 숫자 비교의 같은 문제를 양방향으로 확인한 두 건과, 더 엄격한 패턴 검사 정책의 차이 한 건입니다. **62개의 독립 코딩 작업·모델 정확도·보호 최종 평가·LLM 절감 결과가 아닙니다.** 새 순위 평가·학습·모델 호출·가중치·최종 접근0이며 학습 적격false입니다.

## Read the rows / 행 읽기

`case_id`, `property_id`, `source_family` and `expectation_kind` are audit metadata. `operation`, `left` and `right` are exact raw UTF-8 API inputs. `expected_json` preserves the frozen checks; `observed_json` preserves every typed original observation as JSON text, retaining uint64 precision. `matches_finite_expectation` compares those checks. Source-reading hypotheses are separate in `probes-57.json` and never acceptance criteria. Metadata, expected outcomes and source IDs are not automatically supplied as model input.

위 ID·원천·기대 종류는 감사 메타데이터이고 operation/left/right는 정확한 원본 API 입력입니다. 기대값과 관측값은 JSON 문자열로 보존하여 큰 uint64 숫자의 정밀도를 잃지 않습니다. 일치 여부는 해당 유한 검사에만 적용합니다. 정답·source ID를 모델 입력에 자동 주입하는 자료 구조로 해석하지 마세요. 사용하지 않은 observation 필드의0/false는 호출 증거가 아닙니다. `{a,[}` 패턴의 검사false는 별도 `glob_validate` 행에서 확인합니다.

Normative SemVer, descriptive library API and a separately authored stronger policy are distinct expectations. Match is never prevalidated by the adapter. Validating `{a,[}` is a separate operation from matching its first alternative. Primary entry calls90 and getter/String/Original calls224 are separate from62 observations. Related cases, aliases, reversals and six properties do not add source independence.

규범·기술적 API·별도 강한 정책을 구분합니다. Match 앞에 검사를 삽입하지 않았습니다.90회 entry API와224회 getter/String/Original 호출은62관측과 별도입니다. 원천·helper·반전·입력 변형·여섯 속성을 독립 요청 수로 부풀리지 않습니다.

## Provenance and rights / 출처와 권리

- [Masterminds/semver v3.4.0](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9), MIT; copyright2014–2019 Matt Butcher and Matt Farina.
- [bmatcuk/doublestar v4.9.1](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99), MIT; copyright2014 Bob Matcuk.
- [SemVer2.0.0](https://semver.org/spec/v2.0.0.html) supplies normative rules; expectations are separately authored paraphrases/literals, not copied specification text.
- Source freeze `9e99914f6d1e35aa9d97413f80fc6e03768e97ce`; input freeze `1462c705087f24666049f8bd31d317490c05dbeb`.
- Reviewed GitHub publication: [KO results](https://github.com/teamswyg/laya-tools/blob/@REVIEWED_HEAD@/experiments/public-behavior/RESULTS-57.ko.md), [EN results](https://github.com/teamswyg/laya-tools/blob/@REVIEWED_HEAD@/experiments/public-behavior/RESULTS-57.en.md).

Owned curation is Apache-2.0; preserve included full third-party MIT notices. Source bytes/revisions/licenses are mapped in `upstream-manifest.json`; original code stays in the referenced GitHub fixtures. Inputs/oracles were AI-assisted authored and independently source-read reviewed, not human-approved or blind unseen evaluation. No private prompts/source, credentials, absolute local paths, traces or model binaries are included. No teacher labels were imported. Keep at least15 groups/5% necessary utility and2400 distinct protected-final requests/domain requirements.

자체 작성·정리는 Apache-2.0이며 동봉된 MIT 전문·저작권 고지를 보존합니다. 원본 Go 코드는 연결된 GitHub fixture에 있습니다. AI-assisted 작성과 독립 소스 읽기 검토이며 사람 승인·blind 평가로 표시하지 않습니다. 민감한 로컬 내용이나 모델 본체를 포함하지 않았습니다. 최소15그룹/필요 효용5%/도메인별 고유 보호 최종2400개 기준은 그대로입니다.

Publish only after the exact reviewed GitHub head passes CI and automatically merges. Consumers should pin the returned HF commit, rather than assume mutable main is immutable. Any future changed experiment gets a new version; do not replace these recorded observations.

정확한 reviewed head의 CI 통과·자동 병합 뒤에만 게시합니다. 게시 후 반환된 HF commit을 고정해 사용하세요. 후속 실험은 별도 버전으로 만들고 이번 관측을 교체하지 않습니다.
