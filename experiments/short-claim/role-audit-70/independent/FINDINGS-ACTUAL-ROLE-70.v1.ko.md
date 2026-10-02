# 전체76 역할 배정 결과의 독립 확인

저장된 최초 결과29864B의 SHA와 입력을 확인했다. 부모의 실제 실행은 identity1·membership verification1·Assign1, 재시도0·exit0이다. 나는 역할70 실행 코드 작성자와 다른 검토자다. 이전70 정적 검토와68 source/caption 검토에 노출됐고69/71 드라이버를 작성했으므로 blind 검토는 아니다. 이번에는 stdlib Go로 저장된 JSON·byte/hash·배열만 읽었으며 원본 digest·order·Allocate·Verify·Assign·Features·Project·fit을 호출하지 않았다.

14개 입력6192443B, source15개와 upstream asset15개의 exact pin, 기존1078개 JSON pointer/value SHA를 대조했다. 과거17개 whole group와 새2개 upstream 가족 연결이 보존되고76개 부모·226개 후보의 정답·순서·mask가 입력과 일치한다. unknown21개·63개 nullable 후보는 그대로이며 실제 loss weights는 모두null이다. frozen plan은 draft에서 state와 authorization 두 값만 바뀌었다.

| 구분 | Train | Validation | Calibration |
|---|---:|---:|---:|
| 전체 그룹 |12|4|3|
| known 포함 그룹 |11|4|3|
| 부모/후보 |44/130|20/60|12/36|
| unknown 부모/nullable 후보 |13/39|5/15|3/9|
| loss-eligible P/N |22/51|13/31|5/18|
| masked known 후보 |18|1|4|
| any-eligible 그룹 |10|4|3|
| positive-bearing 그룹 |8|4|2|

같은 그룹의 부모들이 다른 역할로 나뉘지 않았다. 새 semver·doublestar 가족72/74는 모두train이고 unknown-only64도train이다. 기존 stored-truth coverage6개를 입력의 component 목록에서 다시 세어 결과와 비교했으며 모두 원래 최소값을 만족한다. 이는 새 역할을 계산하거나 seed를 탐색한 것이 아니라 이미 저장된 역할의 보존 검사다. any-eligible/positive-bearing 값은 설명용이며 새 gate·통계적 독립성·training-ready 승인이 아니다.

초기 invocation receipt의 actual_assign_attempts0은 호출 전 예약 snapshot이다. 최종 result와 외부 actual ledger의 Assign1이 실행 후 카운터이며 서로 다른 시점을 혼동하지 않는다. 외부 제어기의 전체 child wall0.477674625초·peakRSS53673984B를 그대로 참조한다. 알고리즘만의 시간·메모리나 하드 RSS 상한을 측정한 것은 아니다.

자체 checker는2회였다. 첫 시도는 decoded map과 typed struct의 JSON key 순서 차이를 값 차이로 오해하여 실패했다. typed 결과도 같은 decoded canonical value로 정규화한 두 번째 검사와 vet1이 통과했고 실제 artifact는 수정하지 않았다. 읽기1회는 존재하지 않는 types.go를 추측해 실패했으며 runner.go/loader.go로 보완했다. 별도 원본 API·모델·유료 benchmark·fit·새 역할/정답/가중치·shared edit·public write는0이다. AI-assisted 검토와 협업 비용은 측정하지 않았다.

이 확인은 동결된 metadata와 저장된 실제 결과의 정합성 범위다. 기록되지 않은 외부 코드 공유가 없다는 증명이나 human-only authorship·새 가족의 validation 일반화·학습 준비성 승인으로 확대하지 않는다. source68 supervision은 좁은 유한 proposed 상태를 유지하고 training_ready와 source_authoring_diversity_cleared도false다.
