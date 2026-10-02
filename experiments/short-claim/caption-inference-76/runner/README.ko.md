이 실행기는 공개된 실패 모델71/72를 고정한 채 소스 캡션 A/B의 순서 효과를 보는 개발용 도구다. 새 학습·Laya/생성 모델·GPU 실행기는 아니다. 모델·입력·정답의 변경, 자동 재시도, production 기본 활성화를 지원하지 않는다. root가 봉인한 탐색용 source rubric은 기존 후보의 실제 null 필드나 훈련 라벨을 덮어쓰지 않는다.

`pure/`는 표준 라이브러리만 사용하는 작은 검증·순서·카운터 모듈이다. 합성 테스트는 가짜 Prepare/Decode/Features/Score/Baselines를 주입한다. saved PLAN/typed rubric을 읽는 테스트는 메타데이터 결속만 확인한다. 실제 모델 파일이나 runtime 함수는 실행하지 않는다. `main.go`만 기존 merge98 source의 `shortclaim.Validate`, `shortclaim.Baselines`, `hintlearn.Decode/Features/Score`에 연결한다. 특징 입력은 raw request와 raw caption뿐이다.

실행 명령은 후속 root의 단일 공식 호출에만 사용한다. 다음 변수는 root가 검증한 로컬 파일을 가리킨다. 준비 단계에서 이 명령이나 바이너리의 help를 실행하지 않는다.

```sh
"$RUNNER" --execute-frozen-diagnostic \
  --plan "$PLAN" --plan-sha "$PLAN_SHA" \
  --truth "$TRUTH" --truth-sha "$TRUTH_SHA" \
  --model71 "$MODEL71" --model72 "$MODEL72" \
  --source-root "$SOURCE_ROOT" --attempt-dir "$NEW_ATTEMPT"
```

PLAN SHA `a25c94b6…`, 별도 typed truth SHA `1cec5c10…`, request75 SHA `e1290b64…`, peer receipt SHA `e897a1ff…`와 source commit `7cea4909…`가 코드에 고정되어 있다. schema/scope와 `frozen_for_diagnostic:true`, 명시적 `qualification/production_ready/training_ready/protected_final/independent_heldout/source_truth_is_universal:false`, `fit_calls:0`을 검증한다. 같은3부모·3후보 순서·영어 요청·bool/acceptable 배열을 대조하고 null/unknown을 거짓 음성으로 바꾸지 않는다. 이번 전용 schema는 봉인된3 answerable 부모만 허용한다. A/B에서 평가9개 후보를 모두 유지한다.

모델 파일은 모든 메타데이터/source 검증과 exclusive attempt reservation 뒤에만 읽는다. 각각32,792 B와 고정 SHA를 확인한 뒤 Decode한다. decode된 FP64 계수는 메모리에만 남고 결과에 직렬화하지 않는다. 고정 카운터는6 Prepare,6 Baselines,2 Decode,36 Features,36 Score이며 각4 비교 방식·12 모델 부모 순서·36 개별 점수를 출력한다. 동률은 원래 후보 순서, 최소 checks 비교자 동률은 fixed_order/BM25/lexical_ordered/narrow_rule 배열 순서다. 일반 영어 요청의 narrow_rule BM25 fallback도 결과에 남긴다. 같은3후보의 Top3는 정보가 없는 지표다.

새 attempt 디렉터리는 기존 것이 있으면 거부한다. 각 호출 전 reserved 카운터와 반환 후 prefix를 fsync한 임시 파일의 atomic rename으로 `partial.json`에 저장한다. reserved는 dispatch intent이며 종료 직전 실제 함수 진입을 증명하지 않는다. 반환된 점수는 행마다 저장하고 panic/NaN/오류는 실패로 남긴다. `results.json`은 exclusive final이고 stdout도 같은 구조화 JSON이다. 성공하려면 exit0·completed state·정확한 카운터·qualification false가 모두 필요하다. 부분 실패의 exit1 또는 timeout124를 성공으로 취급하지 않는다. save/출력 실패가 있으면 마지막 durable prefix만 남을 수 있다.

CPU1/Go soft heap256 MiB/60초 watchdog/1 MiB 결과 상한을 사용한다. 자체 watchdog은 Exit124를 요청하며 root는 외부에서 프로세스를 회수하고 OS time/RSS/exit와 stdout·final·partial을 확인해야 한다. Go heap/HeapSys/TotalAlloc/누적 할당 차이·GC와 worker 구간 wall은 결과에 넣고, 외부 whole-process CPU/RSS와 구분한다. worker wall snapshot은 final 파일/출력 완료 전 시점이다. fsync와 검증의 비용도 전체 프로세스 측정에 포함된다. heap soft limit은 OS RSS hard cap이 아니다.

개인 로컬 replace 경로가 있는 maintainer module이다. 이 소스·module 파일을 그대로 public runtime port로 게시하지 않는다. source pin과 buildinfo의 일치는 검증 가능하지만 hermetic compiler proof는 아니다. 캡션 B 작성자와 과거 결과에 노출된 작성자가 만든 nonblind 도구이며, 독립 검토 후 root가 단일 실행을 맡는다. 보호된2,400개 final·새 role·훈련 라벨·Fit·추가 threshold/model search는0이다.
