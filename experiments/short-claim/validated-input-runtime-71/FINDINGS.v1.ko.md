# 역사 소스 보존과 현재 회귀 재현

최적화 뒤 부모의 전체 검사에서 과거 소스 해시를 현재 파일과 대조하던 56a·56b·56e 회귀 3개가 실패했다. 과거 계획·보고서·해시를 수정하지 않고, `9d204c2c700505658c108297d5fa769a835a60c3`의 정확한 소스 5개를 `testdata/shortclaim-source-9d204`의 `.txt`로 보존했다. 원래 4 runtime 파일 외에 56b 계획이 자체 핀으로 포함한 `replay_test.go`도 필요했다. 이 archive는 테스트 데이터이며 runtime fallback이나 학습 feature가 아니다.

56a는 원래 소스 해시를 archive로 검증하고 현재 compiled package-source 해시를 실제 파일과 별도로 대조한다. 현재 평가의 네 baseline 모든 점수·순위·fallback·정답·그룹·분모를 기존 보고서와 비교하며, 원래 floating 허용 오차를 유지한다. 56e는 원래 manifest 경로·각 원본 해시와 support를 검증한 뒤 현재 compiled manifest 전체를 별도로 검증한다. finiteAudit의 전체 관측·정답·카운터·scope 비교는 그대로다. 56b는 원본 plan/result의 고정 SHA와 모든 원본 source pins를 검증한다. 테스트 임시 계획의 source digest만 현재 파일로 바꾸고 strict compiled-source 검증 및 CLI를 사용한다. 각 보고서가 자기 계획의 SHA/ImplementationFiles와 일치함을 먼저 확인하고, 그 두 provenance 필드만 명시적으로 분리한 뒤 전체 정답·그룹·수치·scope·카운터를 비교한다.

실제 behavior/typed CLI에 원래 계획을 주면 source mismatch로 거절되고 출력도 예약되지 않는다. property의 실제 source verifier 역시 원래 source pins를 거절한다. Runtime 소스·검증기 수정, 과거 기록 수정, 신규 skip·오차 완화는 0이다.

대상 3패키지 race는 한 번에 모두 통과했다(1.504/2.079/4.844초). vet 1회와 diff 검사를 통과했다. 준비 중 shell glob 검색 실패 1회, 첫 gofmt의 미닫힌 brace 실패 1회를 [장부](ATTEMPT-LEDGER.v1.json)에 보존했다. 잘못 위치한 digest assertion도 compile/test 전에 발견·수정했다. 실제 대상 test/vet 실패는 0이며 새 benchmark·profile·모델·fit·유료 호출은 없다. 테스트의 기존 Go 관측/회귀 재현은 별도 CI 실행이며 공식 실험이나 첫71 횟수를 덮어쓰지 않는다. 내부 API 횟수는 새로 동적 계측하지 않았다.

이 결과는 구현 작성자의 대상 회귀 확인이다. 독립 검토 및 전체 CI는 부모와 별도 검토자가 담당한다. 부모가 준비한 7개 문서는 이 테스트 수정과 별도로 보존한다. Raw 로그와 호스트 경로는 공개 자료에 포함하지 않는다.
