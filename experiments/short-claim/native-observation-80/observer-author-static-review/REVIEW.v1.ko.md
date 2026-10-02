# Source80 관찰기 v2 작성자 측 정적 검토

현재 동결된 관찰기에서 **실행을 막는 구체적인 문제는 찾지 못했습니다**. 이 검토자는 사전 Want 작성자 checkpoint_cli_45입니다. 관찰기 작성자는 task_expansion_52이며 source-first 작성자는 semantic_review60_prep입니다. 담당은 분리되어 있지만 이전 프로젝트 자료에 노출된 비맹검 검토입니다. 독립 Want 의미 검증·독립 원천·학습 승인이 아닙니다.

읽은 범위는 pure/spec.go·run.go·types.go와 cmd/observe/main.go·save.go 전체, 기존 pure/run_test.go, v2 plan·metadata·저장 buildinfo·마지막 작성자 테스트 로그입니다. 원본 Go import/init/API, observer/main, Go test/build, 모델/학습 호출은 모두 0입니다. shell/jq/SHA로 저장된 자료만 확인했습니다.

동결된 Want 38,513바이트 SHA `c93ac2b7e4c9af530b07db28a42ee3f9d6eadb6319f5ddf3a835c0ccff09ffa8`가 코드 상수·계획·실제 복사본에 연결됩니다. 현재 v2 draft는 8,556바이트 SHA `68ae9d6b8dda1c68813e2ece774f2c5cb58c1134224e2b42a2f92c23c5b36965`, native binary는 4,521,186바이트 SHA `55f662cbd1fd3ab6000aeee7b4159d311a591c15040dcf340ed98b326616c6ca`입니다. plan의 33파일 SHA 확인은 모두 통과했고, 그중 원 UUID15·Ordinal1·module2·license2는 20파일입니다. 코드는 이 source binding과 실제 파일을 검사하고 source/observer 디렉터리의 추가 .go 파일을 거절합니다. 이는 코드 읽기와 저장 핀 연결 검사이지 완전한 hermetic compiler 증명은 아닙니다.

24개 frozen fixture 객체·ID·순서는 Want의 exact 전체 SHA로 고정됩니다. Parse 결과는 반환된 `[16]byte` 전체를 hex로 읽어 늦은 오류의 부분 바이트를 보존합니다. Scan은 fresh nonnil literal receiver를 직접 변환한 pointer에 전달하며 반환/panic 뒤 receiver를 읽습니다. nil interface, typed nil bytes, nonnil empty bytes의 입력 metadata가 구분되고, Ordinal 오류 채널은 null입니다. 호출 전 `Observed.Got`과 `Matches`는 null이며 uncalled 결과를 0값 성공으로 만들지 않습니다.

primary_reserved 증가·동기 checkpoint가 끝난 뒤 원 함수에 dispatch합니다. nonnil 반환 오류는 `%T`로 type만 얻고 Error_reserved·분류·checkpoint 뒤 Error를 한 번 호출합니다. 오류 type이 다르면 unclassified로 기록하며, non-string panic 값의 Error/String 메서드를 추가 호출하지 않습니다. primary panic과 Error-method panic은 구분됩니다. Scan이 panic 전에 receiver를 바꿨다면 그 after 값도 보존합니다. 메서드 panic은 비교 실패이며 worker exit도 실패입니다. 정상 반환의 단순 Want 차이는 completed-with-differences로 남기고 승인이나 재시도 의미를 부여하지 않습니다.

직접 entrypoint는 24개입니다. 기대 Error 관찰은 5개여서 추적 예산은 29개이고, 예상보다 많은 오류를 보존하기 위한 maximum은 24+24=48입니다. Run의 정확한 24행과 행당 primary 1회·Error 최대 1회 구조가 그 상한을 지킵니다. 내부 Scan 재귀·Parse·fmt 호출과 4개 namespace MustParse→Parse 정적 startup site는 동적 계측된 반환 수가 아닙니다. init callback 수/return 수는 null입니다. 외부 root reservation은 package 초기화 전에 이루어져야 하며, worker main 내부 예약만으로 startup을 통제한다고 주장하지 않습니다.

계획은 strict unknown/duplicate-key 거절·frozen=true·Go1.27.1/darwin/arm64/CGO0·CPU1·soft heap256MiB·outside60s·output1MiB·retry0을 확인합니다. 현재 draft는 frozen=false이므로 실제로 거절되어야 합니다. root가 봉인된 frozen plan과 SHA를 전달하고 외부 경계를 담당합니다. 출력을 fresh directory와 O_EXCL로 예약하고 checkpoint를 fsync/rename합니다. 저장 실패 후 counter 보존은 best effort이며 strict 전원 장애/TOCTOU 방지·전체 controller hard return-time 보장을 추가로 주장하지 않습니다. Go soft heap은 OS RSS hard cap이 아닙니다.

저장된 마지막 작성자 pure race 로그는 1package·7 top tests·13 subtests PASS, fail/skip0입니다. 저장 buildinfo는 Go1.27.1·trimpath·CGO0·darwin/arm64와 상대 local replacements를 보고합니다. 이 검토자는 테스트나 binary를 재실행하지 않았고 original init/API 결과는 아직 없습니다. 구체적인 source·runtime 사실은 향후 저장된 실제 결과와 분리해 검증해야 합니다. labels/roles/weights/부모 생성·training_ready·production/qualification/protected-final 승인 모두 0/false를 유지합니다.
