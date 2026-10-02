# 기존 whole-group의 membership 입력 준비

저장된 56b 감사를 역할 입력으로 사용할 수 있게 정리했다. `membership-proposal-65.json`은 private 제안이며 실제 metadata·인코딩·seed·coverage 실행계획이 봉인된 상태는 아니다. 새 역할이나 정답을 배정하지 않았다.

원래 그룹 17개, known/no_answer 부모가 있는 그룹 16개, 부모 72개·후보 위치 216개를 그대로 썼다. truth는 known 34개·no_answer 17개·unknown 21개다. closed-window와 clamp-window는 원래 그룹 8에 함께 있고, unknown-only 그룹 64의 부모 네 명은 모두 unknown으로 남는다. unknown이나 부정 후보 때문에 member·관계를 제거하지 않았다.

구성원 이름은 parent 72개, candidate-position 216개, source 60개, prototype 18개, core 17개, authored semantic component 70개를 구별하는 namespace로 표현했다. 총 453개다. 원래 parent/candidate/source/prototype/core 필드 연결, source→component 204개 및 기존 parent-edge 124개의 이유 1,022개를 구조적으로 옮겨 총 1,922개 관계를 보존한다. 새 group union을 계산하거나 Go 표준 라이브러리·observer를 새 grouping edge로 추가하지 않았다. 표준 import와 공통 observer/whole-source provenance는 원래 source 기록 및 고정 artifact의 policy에 남는다. 같은 raw enum 범위를 공유하는 component ID도 합치지 않았다.

모든 구성원·관계는 고정 원문 artifact SHA와 JSON pointer, 해당 subtree의 canonical Go JSON SHA로 대조할 수 있다. saved outcome 전체의 참조를 보존하여 acceptable-index 배열과 unknown·null·빈 배열의 원래 의미를 유지한다. 60의 원래 60 roots·204 component 관계를 저장 metadata로만 대조했다. 60/61 review overlay도 원래 해시로 묶었으며 fidelity·coverage 판정을 truth로 승격하지 않는다. 원문 함수 실행, AST·formatter·SourcePins·Bind·Generate는 하지 않았다. source text의 다시 승인된 정확도나 closed-world 완전성을 증명하지 않는다.

정확한 encoding proposal은 공개62 source SHA `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`의 `MembershipDigest`다. `riido-whole-component-membership-v1` domain과 u64 big-endian 길이-prefix, UTF-8 byte 순서의 member, From/Kind/To 순서의 관계를 사용한다. 제안 digest만 그룹마다 한 번 계산했다. seed, OrderDigest, AllocateCounts, Assign은 호출하지 않았다. 구성원 ID·관계 digest만으로 소스 metadata 내용의 완전성은 입증되지 않는다. 이후 실행은 **proposal 파일 전체 SHA·입력 SHA·정확한 encoding recipe를 모두** 고정해야 한다.

실제 metadata 연결은 1회 성공·실패 0회다. 원천 JSON 8개 해시, exact reference 2,934개, 최종 원천 안정성 8개를 확인했다. 제안 인코딩 총 245,674바이트와 JSON 1,986,200바이트는 저장 payload 크기이며 메모리·속도 측정이 아니다. 준비 합성 race 테스트 6개와 vet가 통과했다. 장부는 `ledger-65.json`, 자체 대조 receipt는 `binding-receipt-65.json`이다. receipt는 이 helper 작성자의 자체 검증이며 별도 독립 verifier를 뜻하지 않는다.

다음 실제 역할 실행에서는 이 private 입력과 encoding을 검토하여 고정하고, 사전 seed·coverage 계획을 봉인한 후 기존9/3/3 floor와 unknown 정책을 그대로 적용하면 된다. 유리한 seed 검색·그룹 분할·unknown 승격·결과 기반 그룹 선택을 하지 않는다. 실제 그룹·seed·roles·fit·모델·새 라벨·protected final·공유 레포 변경·게시 모두 0이며 training_ready는 false다. 2400은 별도 최종 일반화 목표이고 여기서 모든 development fit에 새 minimum gate로 붙이지 않았다.
