# 첫 실제 Go 학습 입력 준비69

원래72개 요청에서 학습·validation 배열을 처음으로 만들었다. **Project1회·재시도0·학습0회**다. Laya 추론이나 새 모델 성능 측정은 아니다. 이전 [입력 한도 검사](../input-boundary-69/README.ko.md), [mask와 역할 연결](../mask-role-69/README.ko.md), [비학습 효용](../stored-role-utility-69/STORED-ROLE-UTILITY.ko.md) 다음의 실제 준비 실행이다.

| 실제 저장 결과 | train | validation | calibration |
|---|---:|---:|---:|
| 부모 요청 |44|16|12|
| 후보 위치 |132|48|36|
| 학습용 행 |90|36|0|
| 위 행 중 가중치0 |18|1|0|
| 미확정 후보, 감사 자료에만 보존 |42|12|9|

모든 알려진 train/validation 후보를 남기고 설명의 학습 적격성이 없는 행에는0 가중치를 사용한다. 정답 자체를 바꾸거나 그 후보를 순위 비교에서 제거하지 않는다. 미확정 정답과 calibration은 fitting에 들어가지 않는다. calibration의 mask 적격 그룹3개와 **투영된 fit 행에서의** 양수 가중치 그룹0개는 서로 다른 분모다.

직접 입력 검증72회와 Project1회가 반환됐고, Project 내부의 두 번 순회에서 feature scan252회가 기록됐다. 내부 정규화 횟수는 개별 계측하지 않았다. [실제 외부 실행 장부](ROOT-ACTUAL-LEDGER-69.v1.json)와 [실행 전 예약](ROOT-INVOCATION-69.v1.json)을 보존했다.

| 자원 관측 | 값 | 의미 |
|---|---:|---|
| 반환된 소유 payload |1,774,364B, 약1.69MiB|배열·보존 문자열 등의 기존 Go ABI 계산; 임시 할당 제외|
| 전체 자식 peak RSS |56,573,952B, 약53.95MiB|입력 읽기·검증·준비·JSON 기록을 포함하는 OS 관측|
| 외부 wall |0.465949875초|위 전체 준비 과정의 단일 개발 관측|
| 원본 JSON |7,699,810B|들여쓰기와 반복 진단 배열 포함; 메모리 크기가 아님|
| 공개 요약 |97,255B|원본 참조와 행별 SHA를 가진 파생 자료; 전체 feature 생략|

CPU1과 Go heap soft256MiB를 설정하고 외부300초 종료 한도를 사용했다. RSS를 hard cap으로 강제했다는 뜻은 아니다. Go heap, GPU 메모리, 단일 추론 지연, LLM 절감은 측정하지 않았다. AI 보조 준비·검토 비용도 미측정이다.

[공개 요약](COMPACT-ACTUAL-RESULT-69.v1.json)은 원본 SHA `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`와 원래 감독 정보·행 참조·feature 수·행별 해시를 남긴다. 자체만으로 학습을 재현할 수 없고 원본 배열의 대체 자산이 아니다. 원본 배열과 private 실행 계획은 Git 밖에 보존돼 있으며 이 복사 시점에 HF 원본 배열 게시도 미완료다. 공개 모델 계수는 없다. `SampleWeights`는 학습 손실 가중치다.

[작성자 배열 감사](FINDINGS-ACTUAL-69.v1.ko.md)는 별도 표준 Go checker로 저장된 결과만 대조했다. 이 검토자는69/71 드라이버 작성자이며 독립 검토로 부르지 않는다. [실행 전 별도 검토](FINDINGS-PRECHECK-69.v3.ko.md)는 진단 `Excluded` 분모와 원래 역할 DTO 순서 문제, 이전 거절과 수정 이력을 보존한다. 원래 역할 DTO의 순서는 그대로 두고 original index별 배열 lookup을 사용했다. 사전 검토는 실행 후 배열 검토와 구분한다.

[별도 실행 후 배열 검토](FINDINGS-RUNTIME-PROJECTION-69.v1.ko.md)도 통과했다. 원본 실행을 반복하지 않고30실행 핀·원래 감독·288개 저장 정규화 문자열·희소 열과 진단 부분집합을 확인했다. 처음 검사 도구가 모든 부모의 후보 수를3개로 가정해 거절된 기록을 보존했고, 원래 입력의2/3/4개에 맞춰 checker만 수정해 통과했다. [별도 검토 복사 장부](COPY-RUNTIME-QA-LEDGER-93.v1.json)는8개 원본을 보존하며 검토자는 과거62/67 작성 참여가 있다는 한계도 공개한다. feature를 재추출하거나 학습·출처 다양성을 승인한 결과가 아니다.

[검토 장부의 추산 정정v2](LEDGER-RUNTIME-PROJECTION-69.v2.json)도 원문 그대로 보존했다. 첫 거절 전에 확인한 문자열 수는 원래 v1의192가 아니라 source control-flow에서 재구성한4개다. 성공한288개 검사, 원본 실행과 핵심 성공 receipt는 바뀌지 않았다. v1을 덮어쓰지 않았으며 추산값을 실측 counter로 표현하지 않는다.

[복사 장부](COPY-LEDGER-93.v1.json)는 원본13개를 byte-identical로 복사·재읽은 기록이다. 새 학습·출처 다양성·효용·일반화 승인을 뜻하지 않는다. 다음76개는 upstream68의 네 요청을 추가한 별도 코퍼스이며 원래72개 결과를 소급 수정하지 않는다. [English](README.en.md).
