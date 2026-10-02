# 새 공개 사례를 포함한76개 전체 역할70

원래72개 요청에 upstream68의 좁은 유한 입력 요청4개를 추가하고, **전체76개를 고정 seed1729로 처음 다시 배정했다**. 과거66 결과에 새 사례를 train으로 붙인 것이 아니다. 원래17가족과72개 원문·정답·unknown·mask는 그대로 보존하고 관련된 새 요청을 두 source 가족으로 묶었다.

[전체 제안](combined-family-proposal-70.json)은76부모·226후보·19whole-group이며 known38/no_answer17/unknown21을 가진다. 새4개 정답은 실제 source68 관측을 근거로 한 **유한 범위의 proposed supervision**이다. 이를 함수 전체의 모든 동작 정답이나 독립 과제4가족으로 확대하지 않는다. [메타데이터 준비](METADATA-HANDOFF-70.ko.md)와 [실행기 준비](ROLE-EXECUTION-HANDOFF-70.ko.md)는 실행 전 snapshot으로 유지한다.

| 실제 역할 | train | validation | calibration |
|---|---:|---:|---:|
| 전체 provenance 가족 |12|4|3|
| known-containing 가족 |11|4|3|
| 부모 요청 |44|20|12|
| 후보 위치 |130|60|36|
| unknown 후보, nullable 유지 |39|15|9|
| loss 적격 양성/음성 |22/51|13/31|5/18|
| 알려진 masked 후보 |18|1|4|
| loss 적격 후보가 있는 가족, 설명용 |10|4|3|

source identity1회 → membership 검증1회 → Assign1회가 반환됐다. 재시도0, exit0이다. 저장된 “답할 후보 있음/없음”의 역할별 여섯 coverage가 기존9/3/3 기준을 모두 통과했다. 양성 후보가 있는 가족 수를 새 성공 기준으로 만들지 않았다. unknown만 있는 가족은 provenance train으로 배정됐지만 후보 정답이 생긴 것은 아니다.

두 새 upstream 가족72/74가 실제로 train에 배정됐다. 따라서 이번 validation 성능은 이 두 라이브러리로의 전이·일반화 검증이 아니다. membership·seed·분모를 결과에 맞춰 다시 고르지 않았다. 이전72개 [실제 역할66](../role-execution-66/README.ko.md)과 [첫 투영69](../projection-execution-69/README.ko.md)는 별도 역사 결과로 보존한다.

[원본 실제 결과](results-70.json), [호출 장부](role-attempt-70.json), [외부 자원 장부](ROOT-ACTUAL-LEDGER-70.v1.json), [실행 전 예약](ROOT-INVOCATION-70.v1.json), [동결 계획](execution-plan-70.frozen-1.json)을 공개한다. 실행 전 [독립 정적·저장 메타데이터 검토](FINDINGS-PRECHECK-70.v1.ko.md)는1078개 참조와 입력·source·upstream·binary를 확인했고 실제 역할이나 순서 해시를 미리 계산하지 않았다. 검토자는69/71 작성 및 과거68 검토에 참여한 비맹검 AI 에이전트이며70 실행기 작성자와는 다르다.

전체 역할 준비 프로세스의 외부 wall은0.477674625초, peak RSS53,673,984B(약51.19MiB)였다. 입력 읽기·metadata join·배정·JSON 기록을 포함한 개발 관측이다. CPU1/soft heap256MiB/외부300초 한도를 사용했고 GPU·Go heap·모델 추론 지연·LLM 절감은 측정하지 않았다. Features/Project/fit/model/paid/final read0회다. 출처 다양성·학습 준비 flags는 false로 유지되며 bounded development 학습은 별도의 입력·배열·근거 동결 뒤 진행한다.

처음 metadata join에서66 DTO의 row 순서와 original index를 같다고 가정한 거절을 [준비 장부](METADATA-ATTEMPT-LEDGER-70.json)에 남겼다. DTO를 정렬하거나 재배정하지 않고 lookup만 수정했다. 실행 전 검토가 찾은 controller의 partial JSON 처리도 공식 실행 전에 고쳤고 옛 버전은 private에 보존했다. [복사 장부](COPY-LEDGER-93.v1.json)의20개 원본은 동일 바이트이며 private source·binary·raw log는 제외했다. 동결 계획의 `limits` 문자열에는 원래 초안 설명이 남아 있으며 실제 상태/실행 허가는 별도 필드와 예약 장부가 명시한다. [English](README.en.md).
