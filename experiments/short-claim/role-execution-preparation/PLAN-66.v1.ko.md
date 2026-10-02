66은 저장된 17개 whole-group에 개발 역할을 한 번 배정하기 위한 실행 전 초안입니다. 현재 `execution-plan-66.draft.json`은 `execution_authorized=false`이며 실제 원본 역할 배정은 0입니다. 이는 새 사람 승인 절차가 아니라, 기존 사용자 승인 범위 안에서 코드·입력·계획·binary를 동결하고 자동 검사를 통과하기 전 실행을 막는 구분입니다.

입력은 65의 전체 membership snapshot 1,986,200 B입니다. 원본 파일은 그대로 두고 SHA가 같은 사본을 `inputs/`에 보존했습니다. 72개 부모·216개 후보 위치, known 34 / no_answer 17 / unknown 21, 원래 17개 그룹·알려진 부모가 들어 있는 16개 그룹을 유지합니다. closed/clamp의 공유 그룹 8과 unknown-only 그룹 64를 분할하거나 합치지 않습니다. 453개 member와 1,922개 관계의 완전성은 저장된 감사 대상과의 일치이며, 누락된 외부 관계가 없다는 증명은 아닙니다. 65의 독립 metadata 감사와 원래 60/61 의미 검토의 미확정 항목을 모두 보존합니다.

동결할 seed는 ASCII `1729`, 정확한 4 byte `31 37 32 39`입니다. 후보 seed를 검색하지 않습니다. membership은 공개62의 `MembershipDigest`와 동일한 LP64BE 형식, domain `riido-whole-component-membership-v1`을 쓰는 제안입니다. UTF-8 byte 순으로 정렬한 member IDs와 From/Kind/To 관계를 모두 포함하고, sideband인 group ID·known flag·원문 SHA들은 전체 snapshot과 별도로 결속합니다. 원본의 OrderDigest, AllocateCounts, Assign는 아직 호출하지 않았습니다.

기존 3:1:1 largest-remainder 규칙과 train / validation / calibration의 9 / 3 / 3 known-group 하한을 그대로 사용합니다. 저장 metadata를 읽어 확인한 16개 known-containing 그룹 모두에 answerable과 no_answer가 함께 들어 있습니다. 따라서 두 `stored_truth` coverage 집합은 정확히 같으며 각 역할에서 기존 9 / 3 / 3 기준을 확인하는 6개 조건은 더 강한 새 기준을 추가하지 않습니다. 이는 미래 loss-eligible 행이나 의미 검토 완료 여부가 아닙니다. 언어·행동·prototype·cohort의 실제 분포는 역할 반환 후 metadata로 보고하며, 모든 prototype을 모든 역할에 넣는 기준은 만들지 않습니다.

unknown-only 그룹 64는 train provenance로만 남고 알려진 그룹 하한이나 라벨에 포함되지 않습니다. known-containing 그룹 안의 unknown도 자신의 whole-group 역할을 따르면서 unknown 상태를 유지합니다. 표준 라이브러리/observer 예외와 authored helper/type/sentinel 관계를 재발명하지 않습니다. 학습용 mask, 새 라벨, feature projection, 합성 표본 증가를 만들지 않습니다.

CLI 초안은 `--stage prepare|execute --plan <path> --plan-sha256 <exact> --input <membership65> --repo-root <source-root> --source-root <implementation-root> --out <new-dir>`를 받습니다. prepare는 계획·SHA·metadata만 검증하고 역할을 호출하지 않습니다. execute에는 추가로 `--ledger <one-fixed-new-ledger>`가 필요합니다. frozen state, 실제 binary SHA, 정확한 Go 1.27.1 / CGO 0 / trimpath build-info를 확인한 후에만, 새 장부와 결과를 먼저 예약하고 VerifyMembershipSet 1회 및 Assign 최대 1회를 호출합니다. 거절·부분 counter를 보존하며 seed 교체, 그룹 변경, 자동 retry나 장부 환불을 하지 않습니다. 이 제한은 호출자가 고정한 한 장부에 적용되며 host 전역의 실행 제한을 증명하지 않습니다. 다른 장부를 만들어 수동 반복하는 것은 계획 밖의 재실행입니다.

모든 raw 입력은 파일당 8 MiB, 계획과 결과는 1 MiB로 제한합니다. 실제 full snapshot은 이 범위 안에 있으므로 억지로 1 MiB shard로 분해하지 않습니다. 원래 데이터가 가진 null/배열/순서와 source pins는 전체 bytes SHA로 보존하며, 해시를 확인한 같은 bytes를 decode합니다. worker는 GOMAXPROCS 1, Go heap soft limit 256 MiB를 설정합니다. 이는 총 RSS hard limit 또는 성능 측정이 아닙니다.

이번 준비의 기준은 기존 role-recipe59, PLAN-REFERENCES59와 NEXT-STORED58입니다. 58 결과와 데이터가 이미 관측됐으므로 seed 고정은 결과에 유리한 배정을 고르는 일을 막을 뿐, blind 검증이나 독립 모집단을 만들지 않습니다. 현재 대상은 짧은 영어 Go 행동 주장과 같은 합성 문장 작성 흐름입니다. source diversity와 source/text 미확정 항목은 해결됐다고 보지 않습니다. 한국어 설명은 한국어 모델 입력 검사가 아닙니다. 2,400개 이상의 별도 protected-final 목표는 유지하며 매 development fit의 최소 표본으로 바꾸지 않습니다.

현재 original graph union / role ordering / role assignment / 후보 API / feature / fit / model / 유료 호출 / protected-final read는 모두 0입니다. 별도 모델 프로세스 0은 이 AI 협업 자체의 비용이 0이라는 뜻이 아닙니다. 준비 코드의 합성 검사 및 binary build만 수행했고, 원본 prepare/execute CLI는 실행하지 않았습니다. 이 문서와 private plan의 작성은 CI 게시·실행 동결·실제 배정 또는 training readiness 완료가 아닙니다.
