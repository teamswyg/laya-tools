66 public port v2는 기존 공개 테스트 파일 두 개 때문에 준비 검증이 멈추던 문제만 해결합니다. `internal/roleplan`에서 기존 `role_test.go`, `audit_source_test.go`와 함께 `archives_test.go`, `frozen_membership65_test.go`를 정확한 파일명 예외로 허용합니다. 실행 소스는 계속 CLI의 `main.go`/`loader.go`, 역할 패키지의 `role.go`/`audit_source.go` 네 개만 허용하고, 임의 비테스트 Go·미등록 테스트 Go·심볼릭 링크는 거절합니다. 오류는 고정 enum이며 인자나 파일명·절대 경로를 stderr에 넣지 않습니다.

v1 소스·계획·장부는 원래 SHA 그대로 보존했습니다. v2 plan/result/attempt ledger에는 별도 `66-v2` schema를 사용합니다. 새 계획 `public-port-plan-66.v2.draft.json`은 실행 false / binary SHA 빈값인 초안입니다. seed ASCII `1729`와 기존 전체65 입력, 17/16 그룹, 72/216 부모·후보 위치, known 34 / no_answer 17 / unknown 21은 동일합니다. 역할·순서·정답·mask·source relationships를 다시 만들지 않았습니다. 두 stored_truth coverage 집합의 기존 9/3/3 조건도 바꾸지 않았습니다.

최종 role.go SHA `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`, additive helper SHA `f1fb50eb693baa06cd1e9df7ad57275ee557a6bcc17439befbaf9617a6082224`를 유지합니다. worker/source/input/encoding SHA, 실제 Go 1.27.1 / CGO 0 / trimpath binary를 결속하는 절차도 같습니다. Go embed는 source 일치를 확인하는 도구이며 신뢰할 compiler의 독립 증명은 아닙니다. Git source/input commit/blob 동결은 controller가 수행해야 합니다.

prepare는 고정 input/source metadata만 검증하며 역할 함수를 호출하지 않습니다. execute는 frozen 계획과 실제 binary SHA가 없으면 input을 열기 전에 거절합니다. future 실행에서도 하나의 고정 장부와 새 결과를 먼저 예약하고 최대 한 번만 Assign합니다. 장부 reset·seed search·재시도·그룹 분할은 허용하지 않습니다. 장부 제한은 host 전체 전역 제한이 아닙니다.

합성 named tests 10개를 private v2와 공유 레포에서 각각 race로 검사해 모두 통과했고, 각각 vet도 통과했습니다. 추가된 검사는 기존 공개 테스트 파일의 허용, 두 패키지의 임의 실행 파일·미등록 테스트 파일 거절, 알려진 이름의 symlink 거절, 고정 오류의 파일명 비노출, v1 계획의 v2 거절을 확인합니다. 테스트는 무관한 toy membership/seed만 사용하며 원본 prepare / membership 검증 / OrderDigest / AllocateCounts / Assign / 후보 API / 모델 / fit / protected-final은 모두 0입니다. 테스트 진단 시간은 성능 관측이 아닙니다.

허용받은 `cmd/riido-roleplan/{main.go,loader.go,main_test.go}` 세 파일만 공유 레포에 추가했습니다. 다른 공유 파일 수정·commit·push·게시·공식 배정은 하지 않았습니다. 60/61 source/text 미확정, 영어 합성 작성 흐름, source diversity 미충족, training_ready=false를 유지합니다. 새 사람 승인 흐름 없이 기존 사용자 승인 범위와 CI merge gate를 사용합니다.
