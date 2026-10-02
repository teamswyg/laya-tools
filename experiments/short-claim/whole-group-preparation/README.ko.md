# 전체 연결 그룹 입력 준비65

[English](README.en.md) · [원래 제안](membership-proposal-65.json) · [독립 대조](QA-65.ko.md)

같은 코드·후보·부모에서 나온 자료를 서로 다른 역할에 넣지 않도록, 기존17그룹을 **멤버와 관계까지 포함한 전체 snapshot**으로 보관합니다. 역할 모듈에 digest와 opaque ID만 전달해서 원래 관계가 증명된 것처럼 보이지 않게 하는 입력 자료입니다. 원래72부모·216후보·known34/no_answer17/unknown21을 유지했습니다. 그룹8에는 closed-window와 clamp-window가 함께 있으며 unknown-only 그룹64도 삭제하지 않습니다.

원래453멤버·1922관계와 2934개 원문 pointer/값 참조를 보존했습니다. 원래 source60 roots·204 component 관계,124 parent-edge의1022이유와 원래 acceptable 집합이 연결되어 있습니다. 동일 raw span을 가리킨다는 이유로 서로 다른 component ID를 합치지 않았습니다. 표준 라이브러리·observer/whole-file provenance는 기존 global policy로 남고 새 grouping edge를 만들지 않았습니다.

원형 제작1회와 별도 독립 검증2회를 구분합니다. 독립 검증 첫 회는 검증기의 map/struct JSON key-order 비교 오류로 실패했고, helper를 좁게 수정한 다음 회는 통과했습니다. 원래 artifact 수정은0입니다. [실패 보고](INDEPENDENT-AUDIT-65.attempt1.json)·[실패 장부](ATTEMPT-1-FAILURE-LEDGER.json)와 [최종 보고](INDEPENDENT-AUDIT-65.json)·[장부](ATTEMPT-2-LEDGER.json)를 함께 보존합니다. 이는 선언된 저장 snapshot의 대조이며 외부 관계 누락·독립 작성 원천·학습 자격을 증명하지 않습니다.

Membership 제안은 raw UTF-8와 u64 big-endian 길이-prefix를 사용하며17digest를 계산했습니다. Seed/order/실제 역할은 아직 없습니다. 후속 실행기는 이 파일 **전체 SHA**, 원문·source·인코딩 코드와 seed/coverage를 먼저 고정하고 Go 모듈에 전달해야 합니다. 결과를 보고 seed나 그룹을 바꾸지 않습니다. 저장된 known 그룹 하한과 실제 loss-eligible 그룹 수는 별도입니다.

전체 proposal JSON은1,986,200bytes이며 encoded membership은245,674bytes입니다. 이는 파일/payload 크기이고 RSS·CPU 개선 측정이 아닙니다. [Archive guard](../../../internal/roleplan/frozen_membership65_test.go)는 파일별2MiB 이하로11개 원형을 exact SHA/크기 대조하며 역할 API나 학습을 호출하지 않습니다. 멤버들을 array/slice로 읽는 후속 Go 실행기의 재료이며 실행 때 쓰는 라우터/점수 모델과 별도입니다.

원형 HANDOFF/장부의 private/게시0 상태는 준비 당시 기록입니다. 이번 보관과 CI는 별도 PR로 진행하며 과거 문구를 성공으로 고치지 않습니다. 현재 역할·fit·새 label·weights·보호 최종 읽기0, training_ready=false입니다. 일반 AI 보조 협업 비용은 측정하지 않았습니다. 새15저장소·매개발fit2400·모든prototype-per-role 조건은 추가하지 않습니다.
