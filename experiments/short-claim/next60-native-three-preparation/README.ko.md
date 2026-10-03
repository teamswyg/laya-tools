# 다음 세 과제: 실행 전에 정한 기대값과 후보 설명

현재 실제 개발 자료는 **30요청·88라벨**입니다. 이 폴더는 다음 세 과제의 준비 기록이며 아직 새 정답·관측·모델을 추가하지 않았습니다.

| 과제 | 기대하는 동작 | 고정 입력 | 의도한 계약의 표시 위치 |
|---|---|---:|---:|
| 범위가 명확한 비트 보수 | 지정한 너비의 비트를 뒤집고 입력을 보존 | 4 | 0 |
| 오류를 구분하는 JSON 정수 읽기 | 정수 표현·범위를 확인하고 실패를 반환 | 5 | 1 |
| 상태를 보존하는 바이너리 읽기 | 길이·마지막 비트를 검사하고 실패 시 기존 값을 보존 | 5 | 2 |

합계14입력·9후보·예정42실행입니다. 표시 위치는 계획이며 실제 정답 분포나 모델 정확도가 아닙니다. 큰 헤더는 원본 호출 전에 제외하고, 현재 JSON 입력에는 지수 표현·잘못된 JSON이 없어 해당 범위의 실행 성과를 주장하지 않습니다.

[기대값 동결](Root-prefreeze/LITERAL-WANT-FREEZE.actual.v1.json), [최종 설명·순서 동결](Root-prefreeze/CAPTIONS-ORDER.frozen.actual.v1.json), [가족·역할](Root-prefreeze/FAMILY-ROLE-DECISION.actual.v1.json), [별도 정적 검토](nonauthor-review/REVIEW.v1.json)를 읽어 주세요. Root가 기존 초안을 변경하지 않고 후속 문서로 입력·기대값·설명·순서만 채택했습니다. prospective 그룹82(Bitset),83(GJSON+Match+Pretty)는 확인한 등록부 범위의 개발용 가족이며 기존79/현재30의 역할은 그대로입니다.

JSON 표시 후보와 기존 내부 구현의 대응은 **[2,1,0]**입니다. 원래 Int에는 오류 반환이 없으므로 display2의 해당 조건은 확인 불가(U)로 남깁니다. 오류 반환을 가짜 nil로 만들지 않으며, U만으로 부정 라벨을 만들지 않습니다. 알려진 반례와 U가 같이 있을 때에도 각각 보존합니다.

literal-draft의 DRAFT/미채택 상태와 내부 순서는 당시 기록입니다. 현재 상태는 Root-prefreeze의 후속 동결 문서가 설명합니다. 원래 봉인17파일 중 사적 경로 파일·원 봉인 manifest는 게시하지 않았고, 공개13payload는 바이트를 그대로 복사했습니다. 이 폴더의 [공개 파생 기록](PUBLIC-DERIVATION.v1.json)과 manifest가 공개 범위를 정의합니다.

own Go는 실행되지 않은 .go.txt 초안이며 oracle도 같은 작성자의 준비입니다. native 연결·작업자·외부 실행기·메서드 전후 저장 ACK·실제 ABI/compiler·자원 승인 후 별도 원본 실행을 진행합니다. 코드 검토는 비눈가림 정적 검토이며 독립 실행·보호평가 증거가 아닙니다. 원본문·바이너리·모델·raw journal은 게시하지 않습니다. [출처·라이선스 핀](literal-draft/SOURCE-PINS.v1.json)과 네 원 BSD/MIT 전문 고지를 notices에 보존했으며, 모델 학습/게시 권한의 전체 확인과는 구분합니다.

[English](README.en.md) · [기존30행](../next60-development-thirty/README.ko.md) · [게시 완료 기록](../publication-proof-123/README.ko.md)
