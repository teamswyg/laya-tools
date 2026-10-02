# Go 역할 분할 모듈62

[English](README.en.md) · [Go 구현](../../../internal/roleplan/role.go) · [테스트](../../../internal/roleplan/role_test.go)

학습·검증·보정 자료를 나누는 유지보수용 Go 모듈입니다. 같은 코드와 helper에서 나온 요청들이 서로 다른 역할에 섞이면 검증 점수가 부풀 수 있습니다. 이 모듈은 호출자가 먼저 고정한 **전체 연결 그룹**을 한꺼번에 배정합니다. 실행 때 쓰는 힌트 점수·모델 라우팅 기능과는 별도입니다.

호출 흐름은 `VerifyAgainstFreeze`로 고정 metadata를 대조하고, `VerifyMembershipSet`으로 선언된 구성원·관계·digest·겹침을 확인한 뒤 `Assign`에 같은 component 목록과 고정 seed·coverage 계획을 전달하는 것입니다. 실제 원본17그룹을 아직 배정하지 않았습니다. 입력·출력의 합성 검사만 완료했고, 이 코드를 추가해도 `training_ready=false`입니다.

`Assign`은 알려진 정답을 포함한 그룹만 SHA 순서로 정렬하여3:1:1 비율로 배정합니다. 나머지 배분 동률은 train→validation→calibration 순서입니다. 기존 하한9/3/3과 사전에 선언한 coverage가 부족하면 **역할을 하나도 반환하지 않습니다.** 결과를 보고 그룹을 쪼개거나 seed를 다시 찾지 않습니다. Unknown-only 그룹은 train의 출처 집계에만 남고 알려진 정답 그룹 수에 포함되지 않습니다. Unknown member의 라벨도 만들지 않으며 transfer 그룹0을 허용합니다.

배열과 연속 슬라이스를 사용하며 map·lock을 추가하지 않았습니다. Coverage는 ID를 매번 전체 component 목록에서 찾는 대신, 정렬된 ID 목록으로 한 번 해석해 원래 index를 연속 컬럼에 저장합니다. 중복도 정렬 후 이웃을 비교합니다. 문자열 검색과 metadata 준비 작업이어서 SIMD 코드를 추가하지 않았습니다. 이 구조 변경의 속도·RSS 개선은 아직 실측하지 않았습니다.

구현 한도는 component4096개, seed256bytes, ID512bytes와 coverage 전체65536참조입니다. 각 coverage ID의 UTF-8·길이를 복사·정렬 전에 검사합니다. Membership의 길이-prefix 인코딩은 정확한 byte 길이를 먼저 계산해64MiB를 넘으면 거절하고 buffer를 한 번 Grow합니다. 전체 membership set의 인코딩 합도64MiB를 제한합니다. 이는 과학적 학습 자격의 새 기준이나 전체 process RSS의64MiB 보장이 아닙니다.

독립 검토에서 coverage 참조 수만 제한하면 긴 공통 prefix 문자열의 비교량을 충분히 제한하지 못한다는 문제가 발견됐습니다. 길이·UTF-8 검사 위치를 정렬 앞으로 옮겼고, 빈/잘못된 UTF-8/513bytes/1MiB 입력 거절과 정확512bytes 허용을 확인했습니다. 이전 [발견](independent-findings-62.v1.json)·[장부](independent-ledger-62.v1.json), [수정 확인](independent-findings-62.v2.json)·[영향 검사 장부](independent-ledger-62.v2.json)를 모두 보존합니다. [작성자 port 장부](PORT-LEDGER-62.v2.json)도 별도입니다. `role-v1.go.txt`·`role-v1-test.go.txt`는 이전 코드의 보관본이며 빌드 대상이 아닙니다.

중요한 한계가 있습니다. Opaque digest와 `HasKnownMembers` flag만으로 전체 관계가 빠짐없는지 또는 정답 정보가 맞는지는 증명할 수 없습니다. Pointer·source·policy·완전한 graph snapshot과 seed·coverage의 동결은 실제 실행기의 책임입니다. Membership [인코딩 명세](encoding-proposal-62.json)도 아직 제안이며, 새 데이터 작성 원천을 확보하거나 blind 평가를 만든 결과가 아닙니다. 오류는 고정 문자열이며 typed sentinel identity를 보장하는 API는 아닙니다. 실제 역할·fit·weights·새 라벨은0입니다.
