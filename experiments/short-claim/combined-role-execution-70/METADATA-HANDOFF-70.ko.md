# 전체 가족 단위 76개 요청 연결 준비

이 자료는 **private metadata 준비 결과**입니다. 원래 72개 요청과 원천68의 네 요청을 한 번에 역할 배정할 수 있게 연결했습니다. 새 전체 역할 배정, 입력 검증, 특징 생성, Project, 학습과 모델 호출은 아직 0회입니다. 공유 저장소는 변경하지 않았습니다.

## 부모가 바로 이어받을 파일

| 파일 | 용도 |
|---|---|
| `v2/metadata-attempt-1/combined-family-proposal-70.json` | 전체 76개 요청·226개 후보 위치, 원문 pointer, 원래 17개 그룹과 새 두 whole-family 그룹 |
| `v2/preparation-plan.json` | 실제 metadata 연결 전에 고정한 입력 12개·worker·binary·role source 핀 |
| `v2/metadata-attempt-1/invocation-receipt.json` | 준비 시도 전에 남긴 고정 receipt |
| `v2/metadata-attempt-1/attempt-result.json` | 성공과 실제 호출 범위 |
| `v2/future-role-plan-70.draft.json` | 다음 전체 역할 실행의 미실행 초안; runtime worker·binary·실제 실행 계획 고정은 부모 담당 |
| `ATTEMPT-LEDGER-70.json` | 실패를 포함한 전체 준비 장부 |

결과 schema는 `riido-combined-whole-family-proposal-70-v1`, 상태는 `private_whole_metadata_proposal_no_combined_roles_or_projection`입니다. 결과는 2,694,968 bytes이며 SHA256은 `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21`입니다. 4 MiB 상한은 maintainer metadata 구현 한도입니다. 새로운 과학적 자격 기준이나 runtime 메모리 사용량을 뜻하지 않습니다. 기존 65 원문 1,986,200 bytes는 수정하지 않았습니다.

## 그대로 유지한 데이터와 두 원천 가족

기존 72개 요청·216개 후보 위치의 request, candidate 순서, 함수 정답, 허용 후보 집합, 코드·bundle 핀, 마스크와 미확정 상태를 보존했습니다. 원래 17개 그룹의 전체 metadata는 `original65_group_unchanged`에 같은 JSON 값으로 담고 원래 raw subtree SHA와 canonical JSON pointer SHA를 따로 기록했습니다. 출력의 들여쓰기는 달라질 수 있으므로 원래 파일 바이트 동일성은 입력 전체 SHA로 확인합니다. 그룹 8의 closed/clamp 공유와 그룹 64의 unknown-only 상태도 유지했습니다.

추가 순서는 계획68v4의 원래 순서 그대로입니다.

| 전체 index | 원래 ID | 그룹 | 후보 수 | 제한된 제안 허용 위치 |
|---:|---|---:|---:|---|
| 72 | `source68-prefix-reject` | 72 | 2 | `[0]` |
| 73 | `source68-prefix-allow` | 72 | 2 | `[1]` |
| 74 | `source68-nested-allow` | 74 | 3 | `[1]` |
| 75 | `source68-nested-reject` | 74 | 3 | `[0,2]` |

마지막 ID의 실제 의미는 direct-only 목표이며 ID를 바꾸지 않았습니다. semver 그룹 72는 같은 upstream package·helper·타입·전역·재사용 문장을 공유하는 두 요청 전체입니다. doublestar 그룹 74도 동일 원칙으로 묶었습니다. 후보나 문장만 따로 그룹화하지 않았습니다. 원천별 원래 작성 흐름은 별개지만 새 네 request와 연결 도구는 AI 도움을 받아 작성했습니다. 작성 다양성 문제가 모두 해결됐다는 뜻은 아닙니다.

새 두 그룹은 보수적인 whole-package 원문 폐쇄 범위를 사용합니다. semver는 고정 revision `61fc460d28283a91c53be65c2e0f20b494ac8ad9`의 6개 assets, doublestar는 `8b690afa33319b0a1869367f594e53977e38bc99`의 9개 assets입니다. 각각 Go 원문·원본 go.mod·전체 MIT LICENSE를 유지했습니다. 15개 파일의 SHA/크기와 후보 문장 10개 위치의 raw quote 범위를 읽어서 대조했습니다. Go 파일은 shared semantic node이며 라이선스와 module metadata는 폐쇄 범위 증거입니다. MIT 문구나 표준 라이브러리를 공유한다는 이유로 두 가족을 합치지는 않았습니다. source/helper AST 재탐색이나 실행 가능한 새 원천 폐쇄 범위 승인은 수행하지 않았습니다.

## 실제 metadata 수치와 학습 범위

| 항목 | 수치 |
|---|---:|
| 전체 요청 / 후보 위치 | 76 / 226 |
| 기존 answerable / no_answer / unknown 요청 | 34 / 17 / 21 |
| 새 제한된 answerable 제안 요청 | 4 |
| 전체 그룹 / 기존 known-containing / 새 proposed-known | 19 / 16 / 2 |
| 기존 함수 정답 positive / negative | 46 / 107 |
| 새 finite 제안 positive / negative | 5 / 5 |
| loss-scope positive / negative | 40 / 100 |
| known이지만 loss mask가 false인 위치 | 23 |
| unknown null 후보 위치 | 63 |
| any-loss-eligible / positive-bearing 그룹 | 17 / 14 |

마지막 두 그룹 수는 설명용입니다. 새 positive 또는 calibration 최소 기준을 만들지 않습니다. 원래 unknown 21개 요청·63개 후보 위치는 label null·mask false 그대로이며, 원천68의 제한된 제안 label은 `proposed_source68_finite_label`로 분리했습니다. 기존 label을 수정하거나 새로운 10개 위치의 실제 학습 가중치를 계산하지 않았습니다. 모든 `actual_training_loss_weight`는 null입니다.

원래 함수 정답은 후보 구현이 요청을 충족하는지에 대한 정답입니다. caption fidelity mask가 false여도 이 함수 정답 자체가 사라지지 않습니다. **A**는 기존 51개와 새 제한된 네 요청의 전체 known/no_answer 목표를 보존하고 mask는 loss에만 적용하는 주 분석입니다. **B**는 완전한 caption 범위의 별도 진단이며 A의 분모나 gate를 대체하지 않습니다. hard negative와 no_answer 후보를 삭제하지 않습니다. 새 네 요청은 finite scope 제안 상태이므로 실제 fit DTO 채택은 부모의 동결 결속과 QA 후 명시적으로 진행합니다.

## 다음 역할 실행은 하나의 전체 배정

`whole_groups`의 원래 17개 순서를 유지하고 그룹 72/74를 뒤에 붙입니다. canonical members와 From/Kind/To 관계를 그대로 읽고 고정 62의 LP64BE membership domain·no-LF recipe로 연결합니다. 이번 도구는 표준 라이브러리의 독립 proposal encoder로 19개 metadata digest를 재현했으며 public `MembershipDigest`, `OrderDigest`, `AllocateCounts`, `Assign`은 호출하지 않았습니다. 해시와 저장 원문 대조는 기록된 관계의 결속을 보여줄 뿐 미기록 외부 공유 관계가 없다는 증명은 아닙니다.

부모는 검토와 source/input/worker/binary/실행 계획 동결 후 ASCII seed `1729`로 전체 19개 그룹을 **한 번** 배정해야 합니다. 기존 3:1:1 largest-remainder와 train→validation→calibration tie order를 유지합니다. 원래 66의 역할은 `historical_original72_role_provenance_only`에만 있으며 새 76개 역할에 복사하지 않습니다. 그룹 64는 기존 unknown-only provenance train 정책을 유지합니다. 실제 배정 전 `11/4/3` 등을 결과로 주장하지 않습니다.

coverage는 기존 최소 기준만 적용합니다. 원래 answerable·no_answer 모두 있는 16개 그룹과 새 answerable 두 그룹을 metadata에서 연결하여, stored_truth answerable 18개 / no_answer 16개 component 집합에 기존 `9/3/3`을 사용합니다. mask-qualified coverage는 배정 이후 별도 보고하며 positive-bearing 수는 설명용입니다. coverage가 실패하면 결과를 그대로 남기고 seed·그룹·membership·분모·기준을 고치거나 유리한 결과를 찾기 위해 재시도하지 않습니다. 새로운 모든 prototype×모든 role, 15개 repo, 2,400개 매-fit 조건은 없습니다. ≥2,400 protected-final/domain 목표는 별도이며 이번에는 해당 자료를 읽지 않았습니다.

## 숨기지 않은 첫 실패

첫 준비 시도는 원래 66의 `parent_roles` 배열이 parent index 순서라고 잘못 가정하여 `combined_family_prep_prior_role_index`로 실패했습니다. 원문은 그룹 순서이며 예를 들어 index 0..11 뒤에 36..39가 들어 있습니다. 원래 worker·binary·plan·receipt·실패 장부는 이 폴더의 v1에 그대로 남겼습니다. v2는 `original_parent_index`를 원래 JSON row index로 연결하는 배열 lookup만 추가했습니다. 원래 role 배열·parent/candidate 순서·seed·정답·mask·그룹은 고치지 않았습니다. 별도 v2 계획을 동결하고 metadata 연결 1회를 성공시켰습니다. 총 준비 시도는 **2회, 성공 1회, 실패 1회**이며 실제 역할 배정·순위·특징·학습·원천 API·모델 호출은 0회입니다.

합성 테스트는 초기 전체 race 9개, build-info 범위 수정 후 2개, v2 순서 결속 수정 후 2개로 총 3회/13 named-test 실행을 통과했습니다. vet 3회, binary build 2회, plan seal 2회도 통과했습니다. 이는 준비 도구의 기계적 계약 검사이며 원래 데이터의 새 의미 승인이나 모델 성능 증명이 아닙니다.
