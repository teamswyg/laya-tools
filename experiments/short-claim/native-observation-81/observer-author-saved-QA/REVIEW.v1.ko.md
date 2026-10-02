# Source81 저장 결과 무결성 검사

Root가 한 번 실행한 32개 유한 입력은 모두 사전 Want와 일치했습니다. 최종 결과·마지막 partial·외부 stdout은 각각 240,467바이트이며 SHA256 `8f8a94dd6afb287a63518b16a1245290e9fc1795bec5ed34a9c3dc6c5aff4635`로 같습니다. 재실행 없이 저장된 파일만 읽었습니다.

이 검사는 관찰기 저자 checkpoint_cli_45의 비맹검 사후 기록 검사입니다. Want 저자 task_expansion_52와 원문/Want 동료 검토자는 별도이지만, 이 보고서를 독립 정답 승인이나 독립 원천 검증으로 부르지 않습니다. 이전 role62/projection76/Want80/conversion80 노출도 있습니다.

32개의 전체 원래 record, 입력 객체·pointer·순서·사전 Want를 보존했습니다. 각 Got의 모든 중첩 채널과 nullable 값이 Want와 같습니다. time getter 4묶음, outer Error→Unwrap→cause Error 6묶음, pflag snapshot 42개를 확인했습니다. snapshot에는 nil slice 2개와 nil이 아닌 빈 slice 3개가 구분되어 있습니다. 오류 때 time의 영값과 CSV 오류 필드, 공개 Flag.Changed 및 Replace/Append의 상태도 원래 Want/Got 전체 비교에 포함됩니다.

16종 API의 reserved/attempted/returned 합계가 모두 182이며 panic과 difference는 0입니다. 목표별 입력은 각각 8개, 호출 수는 41/16/65/60입니다. original public dispatch 160회와 stdlib 명시 호출 22회를 구분합니다. 내부 registration/formatting/hook 호출 및 package initializer의 동적 호출 수는 미계측 null을 유지합니다. closure 76개 파일의 바이트·SHA와 manifest·binary의 바인딩을 읽기만으로 확인했으며, 동결 plan은 원래 draft의 frozen false→true 한 토큰만 변경되었습니다.

Root의 외부 실행은 Start 1회, joined Wait, exit 0, retry 0, timeout false입니다. 전체 child `/usr/bin/time`은 real 2.55초/user 0.15초/sys 0.18초, 최대 RSS 19,611,648바이트(18.703125MiB), peak memory footprint 16,679,416바이트입니다. controller wall은 2.5557745초입니다. 초기화·소스 검증·JSON·fsync를 포함한 전체 프로그램 수치이며 함수별 비용, Go heap, GPU/Laya 또는 모델 추론 측정이 아닙니다. 256MiB는 Go soft heap 설정이고 RSS hard cap 보장은 아닙니다.

author helper 1회 성공/실패 0, 사전 계획은 [PLAN.v1.json](PLAN.v1.json), 구체적인 해시·카운터 근거는 [MECHANICS.v1.json](MECHANICS.v1.json)입니다. 원 API·native worker·controller·원 테스트·Features·Score·Fit 재실행은 0입니다. Root의 사전 metadata helper는 2회 중 1회 실패(-ldflags 필드가 trimpath build metadata에 없음을 잘못 필수화), 별도 saved-review helper는 1회 성공으로 원래 기록을 보존합니다. observer 준비의 vet 1회 실패와 patch context 거절 등도 기존 준비 장부에 남습니다. 모든 읽기 오류를 빠짐없이 포착했다는 주장은 하지 않습니다.

32 fixtures는 4개 목표/2개 원천 계열의 관찰이며 새 parent·label·role·weight는 0입니다. 실제 감독 적격성이나 사용자 요청/후보 캡션의 충실성을 관찰한 실험이 아닙니다. qualification/training_ready/production_ready/protected_final은 false입니다. private argument paths·원시 OS 로그·바이너리·private helper는 공개 제외하고 SHA만 인계합니다.
