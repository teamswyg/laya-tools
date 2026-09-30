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
