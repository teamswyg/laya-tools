# laya-tools · User Guide / 사용자 안내

[![CI](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/teamswyg/laya-tools)](https://github.com/teamswyg/laya-tools/releases/latest)
[![Project license](https://img.shields.io/badge/project_license-Apache--2.0-blue)](https://github.com/teamswyg/laya-tools/blob/main/LICENSE)
[![Status](https://img.shields.io/badge/status-experimental-orange)](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-EN)

**Choose your language / 언어를 선택하세요:** [English — Start here](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-EN) · [한국어 — 처음 시작하기](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-KO)

`riidolaya` helps you find code, compare model-selection policies, and preview which repository owns a task. Start without a model; add local Laya inference only when you want to measure its benefit. Codex integration is optional.

`riidolaya`는 코드 찾기, 모델 선택 정책 비교, 작업을 담당할 저장소 preview를 돕습니다. 모델 없이 먼저 시작하고, 효과를 측정할 때 로컬 Laya 추론을 추가하세요. Codex 연동은 선택 사항입니다.

> Experimental: actual Codex cost savings have not been demonstrated. Repository results are previews, not automatic work assignments.
>
> 실험 단계입니다. 실제 Codex 비용 절감은 아직 입증하지 않았으며, 저장소 선택 결과도 자동 작업 배정이 아닌 preview입니다.

**Latest owned task results / 최근 실제 작업 결과:** [한국어 결과54](https://github.com/teamswyg/laya-tools/wiki/Task-Results-54-KO) / [English results54](https://github.com/teamswyg/laya-tools/wiki/Task-Results-54-EN). One parser request ran twice per Sol6/Luna low profile: all four passed the independent closure with zero exit and complete core usage. Its bounded array-based key validation is integrated into runtime. Cumulative observations are16 records,5 distinct requests,3 attempted families and1 repository. The8 local candidates are a different count. Laya abstained at0.610; cold whole-process RSS was about1.51GB. These development results do not establish general capability, ultra-small memory or savings. / 파서 요청 하나를 Sol6/Luna low로 각각 두 번 실행해 네 후보 모두 독립 수용·정상 종료·전체 기본 사용량을 확인했고 고정 배열 기반 키 검사를 runtime에 반영했습니다. 누적16기록·5고유 요청·세 실제 시도 가족·한 저장소이며 로컬후보8개는 다른 수치입니다. Laya는0.610에서 보류했고 cold 전체RSS 약1.51GB였습니다. 일반적인 능력·초저메모리·절감을 입증하지 않습니다. [KO source](https://github.com/teamswyg/laya-tools/blob/main/experiments/task-outcomes/RESULTS-54.ko.md) · [EN source](https://github.com/teamswyg/laya-tools/blob/main/experiments/task-outcomes/RESULTS-54.en.md).

**Prior training / 이전 학습:** [학습 4,456개·검증 2,599개 결과](https://github.com/teamswyg/laya-tools/wiki/Path-Claim-Results-46-KO) / [4,456 training and 2,599 validation examples](https://github.com/teamswyg/laya-tools/wiki/Path-Claim-Results-46-EN). All ten tiny-head fits completed; usefulness gates failed. / 작은 모델 10개 학습은 완료했지만 유용성 기준은 실패했습니다.

**Golden sets / 골든셋:** [24개와 2,400개의 의미 및 도메인별 정답](https://github.com/teamswyg/laya-tools/blob/main/docs/golden-set-scale.ko.md) / [Scale, held-out evaluation and domain-specific evidence](https://github.com/teamswyg/laya-tools/blob/main/docs/golden-set-scale.en.md). At least 2,400 distinct final requests per evaluated domain remains the target, with related cases grouped. Search evidence and development attempts do not validate final model routing. / 영역별 구분되는 최종 요청 최소 2,400건을 목표로 하며 관련 사례는 그룹으로 관리합니다. 검색 결과와 개발 시도로 최종 모델 라우팅을 검증했다고 볼 수 없습니다.

**Acquisition and budgets / 확보와 실행 예산:** [한국어 계획](https://github.com/teamswyg/laya-tools/wiki/Golden-Set-Acquisition-KO) / [English plan](https://github.com/teamswyg/laya-tools/wiki/Golden-Set-Acquisition-EN). Follow staged source collection, independently checked task contracts, grouped splits, and bounded comparisons; repeated model executions do not increase unique-request counts. / 원천 확보·독립 작업 계약·그룹 분할·제한된 비교 실행을 단계적으로 진행합니다. 같은 요청을 여러 모델로 실행해도 고유 요청 수는 늘지 않습니다. [검토 원본](https://github.com/teamswyg/laya-tools/blob/main/docs/golden-set-acquisition.ko.md) · [Reviewed source](https://github.com/teamswyg/laya-tools/blob/main/docs/golden-set-acquisition.en.md).

**Public Go source survey / 공개 Go 원천 조사:** [한국어](https://github.com/teamswyg/laya-tools/wiki/Public-Go-Acquisition-53-KO) / [English](https://github.com/teamswyg/laya-tools/wiki/Public-Go-Acquisition-53-EN). 120 authored behavior candidates from eight repositories, with zero execution-eligible contracts, model labels or final-eligible requests. / 8저장소에서 원본 행동 후보 120개를 작성했으며 실행 자격·모델 라벨·final 자격은 0개입니다. [KO source](https://github.com/teamswyg/laya-tools/blob/main/docs/public-go-acquisition-53.ko.md) · [EN source](https://github.com/teamswyg/laya-tools/blob/main/docs/public-go-acquisition-53.en.md).

**External contract preparation54 / 외부 계약 준비54:** [한국어](https://github.com/teamswyg/laya-tools/wiki/Public-Go-Contracts-54-KO) / [English](https://github.com/teamswyg/laya-tools/wiki/Public-Go-Contracts-54-EN). The survey above is the preserved PDCA53 snapshot. Independent contracts are now prepared for2 of its120 candidates, preserving humanize MIT and UUID BSD source/license bytes and module identities. External model attempts, training and final eligibility remain0. / 위 원천 조사는53 당시의 보존 기록입니다. 그120개 중2개에 독립 계약을 준비하고 humanize MIT·UUID BSD 원본/라이선스와 모듈 이름을 보존했습니다. 외부 모델 실행·학습·final 적격은0개입니다.

**Fair comparison preparation 55 / 공정 비교 준비 55:** [한국어 안내](https://github.com/teamswyg/laya-tools/blob/main/docs/fair-upstream-comparison.ko.md) · [English guide](https://github.com/teamswyg/laya-tools/blob/main/docs/fair-upstream-comparison.en.md).

The two new v2 IDs clarify the same two external requests; they do not add distinct tasks. Full model input and original Go language conditions are fixed before comparison. External model attempts remain 0. Two child plans share four ordered reservations in one durable ledger, with no refunds and concurrency 1; this is not a host-wide or provider-wide cap. This prepares development comparisons, while at least 2,400 distinct protected final requests per domain remains a separate target. Passing CI and frozen plans are required before actual model execution.

새 v2 ID 두 개는 기존 외부 요청 두 개의 설명·평가 버전이며 고유 작업 수를 늘리지 않습니다. 전체 모델 입력과 원본 Go 언어 조건을 비교 전에 고정합니다. 외부 모델 시도는 0개입니다. 두 작업 계획이 한 durable ledger의 순서가 정해진 예약 네 개를 공유하며 환불 없이 한 번에 하나씩 실행합니다. 호스트나 공급자 전체의 호출 제한은 아닙니다. 이는 개발 비교를 위한 준비이며, 도메인별 서로 다른 보호된 최종 요청 최소 2,400개는 별도 목표입니다. 실제 모델 실행 전에 CI와 계획 봉인이 필요합니다.

**Usage and task completion / 사용량과 작업 완료:** [한국어 사용법](https://github.com/teamswyg/laya-tools/blob/main/docs/task-outcomes.ko.md) / [English guide](https://github.com/teamswyg/laya-tools/blob/main/docs/task-outcomes.en.md). Existing telemetry and independent acceptance are separate; missing evidence stays unknown. Authored examples are tool checks, not model outcomes. / 기존 기록의 사용량과 독립 작업 검증을 따로 확인하며, 누락은 미확정으로 남깁니다. 직접 작성한 예제는 도구 검사이며 모델 작업 결과가 아닙니다.

**Explicit task execution / 명시적 작업 실행:** [한국어](https://github.com/teamswyg/laya-tools/blob/main/docs/task-execution.ko.md) / [English](https://github.com/teamswyg/laya-tools/blob/main/docs/task-execution.en.md). An opt-in Go executor binds a public task's CLI invocation, usage, candidate and independent checks; failures stay visible. The development pilot is separate from final routing evaluation. / 선택적 Go 실행기로 공개 작업의 CLI 실행·사용량·후보·독립 검사를 연결하고 실패도 기록합니다. 개발 파일럿은 최종 라우팅 평가와 구분합니다.

**Preserved prelaunch stop / 실행 전 중단 기록:** [한국어 결과 50](https://github.com/teamswyg/laya-tools/blob/main/experiments/task-outcomes/RESULTS-50.ko.md) / [English results 50](https://github.com/teamswyg/laya-tools/blob/main/experiments/task-outcomes/RESULTS-50.en.md). The earlier packaging defect stopped before any main coding-model process. This historical result is retained separately. / 이전 패키징 문제는 코딩 모델 본 실행 전에 멈췄으며 그 결과를 별도로 보존합니다.

**Search helper screen / 검색 방식 비교:** [한국어 결과](https://github.com/teamswyg/laya-tools/blob/main/experiments/path-helper-headroom/RESULTS-48.ko.md) / [English results](https://github.com/teamswyg/laya-tools/blob/main/experiments/path-helper-headroom/RESULTS-48.en.md). Three fixed helpers failed the 5% necessary headroom gate on 4,456 training requests; fitting stopped. Full runs used 106.7–107.3MiB peak RSS. / 학습용 4,456개에서 세 후보 모두 5% 개선 여지 기준에 못 미쳐 추가 학습을 중단했습니다. 전체 실행의 최대 RSS는 106.7–107.3MiB였습니다.

## What do you want to do? / 무엇을 하고 싶나요?

| Goal / 목적 | English | 한국어 |
|---|---|---|
| Install and get your first result / 설치·첫 실행 | [Getting started](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-EN) | [처음 시작하기](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-KO) |
| Find code, plan models, preview repositories / 기능별 사용 | [Workflows](https://github.com/teamswyg/laya-tools/wiki/Workflows-EN) | [기능별 사용법](https://github.com/teamswyg/laya-tools/wiki/Workflows-KO) |
| Connect an agent or Go application / 에이전트·Go 연동 | [Agents and Go](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-EN) | [에이전트·Go 연동](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-KO) |
| Understand memory, speed, and errors / 성능·문제 해결 | [Performance and troubleshooting](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-EN) | [성능과 문제 해결](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-KO) |
| Reuse or redistribute / 프로젝트 목적·라이선스 | [Project and licensing](https://github.com/teamswyg/laya-tools/wiki/Licensing-and-Project-EN) | [프로젝트와 라이선스](https://github.com/teamswyg/laya-tools/wiki/Licensing-and-Project-KO) |

## What runs locally? / 어디서 실행되나요?

| Operation / 명령 | Behavior / 동작 |
|---|---|
| `search --lexical`, `plan` without a task, `repo-preview` without `--laya` | Local Go logic; no model download required / 모델 없이 로컬 Go로 처리 |
| `setup` | Downloads public model/runtime assets / 공개 모델·런타임 다운로드 |
| Laya-assisted search, routing, preview | Local inference; roughly 1.4 GiB measured process RSS / 로컬 추론, 측정된 전체 메모리 약 1.4GiB |
| `codex --dry-run` | Shows a proposed command; does not launch Codex / 실행 계획만 표시 |
| `codex` | Starts your installed Codex; its normal provider and usage apply / 기존 Codex 실행, 기존 공급자·요금 정책 적용 |

CI/release badges describe build status and versions, not proven savings or accuracy. The license badge describes project code; upstream notices still apply.

CI·릴리스 배지는 검사 상태와 버전을 나타내며 정확도·절감 효과의 인증이 아닙니다. 라이선스 배지는 자체 코드 기준이며 외부 구성 요소 고지도 적용됩니다.

[Repository](https://github.com/teamswyg/laya-tools) · [Downloads](https://github.com/teamswyg/laya-tools/releases/latest) · [Detailed bilingual documents / 상세 문서](https://github.com/teamswyg/laya-tools/blob/main/docs/README.md) · [Report an issue / 문제 제보](https://github.com/teamswyg/laya-tools/issues)


- [Semantic hints: 초소형 주장·힌트 모델 연구](Semantic-Hints-KO)
- [Semantic hints: target and experimental plan](Semantic-Hints-EN)

- [실제 관련성 학습 02](https://github.com/teamswyg/laya-tools/wiki/Pair-Training-02-KO) / [Real relevance training 02](https://github.com/teamswyg/laya-tools/wiki/Pair-Training-02-EN)

- [저비용 특징 학습 03](https://github.com/teamswyg/laya-tools/wiki/Lexical-Learning-03-KO) / [Low-cost feature learning 03](https://github.com/teamswyg/laya-tools/wiki/Lexical-Learning-03-EN)

- [정적 임베딩 Go 실행·압축](https://github.com/teamswyg/laya-tools/wiki/Static-Embeddings-04-KO) / [Static embeddings in Go](https://github.com/teamswyg/laya-tools/wiki/Static-Embeddings-04-EN)

- [정적 임베딩 학습 05](https://github.com/teamswyg/laya-tools/wiki/Static-Alignment-05-KO) / [Static alignment learning 05](https://github.com/teamswyg/laya-tools/wiki/Static-Alignment-05-EN)

- [검색 자료 점검 06](https://github.com/teamswyg/laya-tools/wiki/Retrieval-Audit-06-KO) / [Retrieval source audit 06](https://github.com/teamswyg/laya-tools/wiki/Retrieval-Audit-06-EN)

- [2,948개 질문 검색 시험](https://github.com/teamswyg/laya-tools/wiki/Retrieval-Baseline-07-KO) / [2,948-query retrieval baseline](https://github.com/teamswyg/laya-tools/wiki/Retrieval-Baseline-07-EN)

- [순위 배열 재사용과 메모리 할당](https://github.com/teamswyg/laya-tools/wiki/Rank-Reuse-08-KO) / [Ranking array reuse and allocation](https://github.com/teamswyg/laya-tools/wiki/Rank-Reuse-08-EN)

- [보조 검색 주장 모델 학습 09](https://github.com/teamswyg/laya-tools/wiki/Search-Claim-09-KO) / [Auxiliary-search claim training 09](https://github.com/teamswyg/laya-tools/wiki/Search-Claim-09-EN)

- [첫 후보를 보존하는 보조 검색](https://github.com/teamswyg/laya-tools/wiki/Baseline-First-10-KO) / [Baseline-first identifier hints](https://github.com/teamswyg/laya-tools/wiki/Baseline-First-10-EN)

- [후보를 조금씩 읽는 페이지 출력](https://github.com/teamswyg/laya-tools/wiki/Hint-Pagination-11-KO) / [Paged candidate output](https://github.com/teamswyg/laya-tools/wiki/Hint-Pagination-11-EN)

## Page cost experiment 12 / 페이지 비용 시험

[한국어](Page-Cost-12-KO) · [English](Page-Cost-12-EN). 2,948 distinct queries: fewer emitted candidates does not establish lower CPU cost. [PR #35](https://github.com/teamswyg/laya-tools/pull/35) tracks CI and merge status.

## Reusable search session / 검색 결과 재사용 세션

[한국어](Hint-Session-13-KO) · [English](Hint-Session-13-EN). Use `--session --limit 20` to compute once and read subsequent pages using a cursor. See the associated PR for CI and merge status.

## Page-saving claim training 14 / 페이지 절감 주장 학습

[한국어](Page-Claim-14-KO) · [English](Page-Claim-14-EN). 24 model exports over 2,948 queries all match always-helper; the fixed gate fails. [PR #37](https://github.com/teamswyg/laya-tools/pull/37) tracks CI and publication readiness.

## Helper-call budget experiment 15 / 보조 호출 예산 시험

[한국어](Page-Budget-15-KO) · [English](Page-Budget-15-EN). Validation-only calibration: the simple score-gap rule beats the learned heads, but all fixed gates fail. Reuses the immutable [experiment-14 models](https://huggingface.co/JooYoon/riidolaya-page-claim-research-v0.1/tree/4e0addaa24d407bb11282f193e039c943a08bd0a).

## CoNaLa distinct-question audit 16 / 실제 질문 수 감사

[한국어](CoNaLa-Audit-16-KO) · [English](CoNaLa-Audit-16-EN). 2,879 curated rows contain 2,074 question IDs / 2,089 normalized original intents, insufficient alone for 2,400 distinct original questions. No model scoring or training.

- Laya relevance 17/17b: [한국어](Laya-Relevance-17-KO) / [English](Laya-Relevance-17-EN)

- Laya implementation / FP32–INT8 diagnostic 18: [한국어](Laya-Parity-18-KO) / [English](Laya-Parity-18-EN)

- Original Laya validation 19: [한국어](Laya-Original-19-KO) / [English](Laya-Original-19-EN)

- Call-cost claim learning 20: [한국어](Cost-Claim-20-KO) / [English](Cost-Claim-20-EN)

- Cheap score-spread features 21: [한국어](Score-Spread-21-KO) / [English](Score-Spread-21-EN)

- Shallow page-gain claims 22: [한국어](Shallow-Claim-22-KO) / [English](Shallow-Claim-22-EN)

- Claim input signal audit 23: [한국어](Claim-Signal-Audit-23-KO) / [English](Claim-Signal-Audit-23-EN)

- Query coverage features 24: [한국어](Query-Coverage-24-KO) / [English](Query-Coverage-24-EN)

- Joint coverage claim training 25: [한국어](Coverage-Claim-25-KO) / [English](Coverage-Claim-25-EN)

- Changed-call loss audit 26: [한국어](Call-Change-Audit-26-KO) / [English](Call-Change-Audit-26-EN)

- Deferred helper on continuation27: [한국어](Deferred-Helper-27-KO) / [English](Deferred-Helper-27-EN)

- Real GitHub task source audit28: [한국어](Real-Task-Audit-28-KO) / [English](Real-Task-Audit-28-EN)

- Grouped evaluation membership29: [한국어](Evaluation-Groups-29-KO) / [English](Evaluation-Groups-29-EN)

- All historical root metadata30: [한국어](Historical-Roots-30-KO) / [English](Historical-Roots-30-EN)

- Historical license text coverage 31: [한국어](License-Texts-31-KO) / [English](License-Texts-31-EN)

- Complete long-query execution 32: [한국어](Long-Query-32-KO) / [English](Long-Query-32-EN)

- Complete pre-fix catalogs and isolated labels 33: [한국어](File-Catalogs-33-KO) / [English](File-Catalogs-33-EN)

- [실제 2,400개 파일 검색 결과와 한계](File-Localization-34-KO) · [Actual 2,400-task localization results and limitations](File-Localization-34-EN)

- [인덱스 재사용의 효과와 한계](Path-Index-Reuse-35-KO) · [Index reuse benefits and limits](Path-Index-Reuse-35-EN)

- [별도 학습 후보 15,423개 출처 검사](Training-Source-36-KO) · [Separate 15,423-group training-source audit](Training-Source-36-EN)

- [학습·검증·최종 2,402개 분리](Training-Partition-37-KO) · [Train/validation and 2,402 final tasks](Training-Partition-37-EN)

- [학습·검증 파일 목록과 과거 라이선스 수집](Training-Catalogs-38-KO) · [Development catalogs and historical license collection](Training-Catalogs-38-EN)

- [개발 자료 13,021개의 정답 준비와 실패 보존](Development-Labels-39-KO) · [Targets and explicit failures across 13,021 development tasks](Development-Labels-39-EN)

- [개발 정답과 수정 전 파일 목록의 연결 검사](Development-Join-40-KO) · [Development targets joined to pre-fix catalogs](Development-Join-40-EN)

- [숫자 모델 학습 범위와 긴 질문 입력 계약](Training-Scope-41-KO) · [Numerical training scope and long-query input contract](Training-Scope-41-EN)

- [선택할 때만 계산하는 보조 파일 검색](Deferred-Path-Search-42-KO) · [Actually deferred auxiliary path search](Deferred-Path-Search-42-EN)

- [작업별 보조 검색 이득·손해 자료](Path-Cost-Data-43-KO) · [Per-task auxiliary search gains and losses](Path-Cost-Data-43-EN)

- [학습·검증을 함께 수집하는 순서](Fair-Catalog-44-KO) · [Acquire training and validation together](Fair-Catalog-44-EN)

- [확장 비용·출처 준비와 검증 부족](Path-Cost-Data-45-KO) · [Expanded costs and source readiness](Path-Cost-Data-45-EN)

- [24개와 2,400개: 작은 주장 모델 학습 기준](Path-Claim-Plan-46-KO) · [24 versus 2,400: tiny claim fitting requirements](Path-Claim-Plan-46-EN)
