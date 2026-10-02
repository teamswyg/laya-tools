# Wrap/Marshal75 제어기 읽기 전용 검토

고정 제어기 source main62d62f…/test9be290…/binary5add1e…/handoff d75a4165…를 검토했다. 검토 범위에 material blocker는0이다. stdlib helper1회가 handoff1+제어기 자산7+관찰기 바인딩5+runtime closure14의27개 byte/SHA 핀과 Go1.27.1·CGO0·darwin/arm64·trimpath·stdlib-only 바이너리 메타데이터를 확인했다. 핀은 컴파일 출처 바인딩이며 trusted compiler의 독립 증명은 아니다.

Start 전의 순서는 plan/handoff/Want/WantSeal·14closure·binary 검증 → 단일 Frozen:false 토큰의 메모리 변경 → 없는 outside directory 생성 → 새 plan·invocation·before-start ledger O_EXCL 저장·file/dir sync → 없는 child directory 생성 → outside-reservation O_EXCL/file sync·child/parent dir sync → raw log O_EXCL → Start이다. 관찰기가 final results를 직접 독점 예약하며 제어기는 이를 선점하지 않는다. 기존 디렉터리/핀 차이/잘못된 인자/중복 JSON key를 거부하고 진단에는 사용자 인자나 경로를 출력하지 않는다.

한 child만 시작하고 CPU1·Go soft heap256MiB·최소 환경을 전달한다. 별도 process group에60초 종료를 걸며 SIGKILL 뒤 Wait를 정확히 한 번 호출하고 기다린다. Wait가 끝나는 총시간까지60초 hard bound라고 주장하지 않는다. 실패/kill/overflow 뒤 자동 재시도는 없다. 기록의 OS time/RSS/footprint는 nullable이며 전체 child init/preflight/원API/checkpoint를 포함하고 controller는 제외한다. soft heap은 전체 RSS hard cap이 아니다.

typed 결과는 고정 PlanSHA/WantSHA,24개 Probe/Want 순서, Got와 Matches의 정확 비교, dispatch/return/panic 및 matches/differences 합계를 검사한다. null과 빈 map entries를 DeepEqual로 구분한다. in-flight 최대1건과 active ID를 확인하고 반환되지 않은 결과를 채우지 않는다. final이 손상되고 partial만 유효하면 final_result_invalid가 남는다. 차이가 있어도 완료된 유한 관측은 completed_with_differences로 보존하며 All24WantsMatched와 별도다. 이후 학습·caption·부모 정답·broad truth 승인은 false이다.

작성자14개 pure/stub+metadata tests를 실행하지 않고 읽었다. final main의 CPU1/softheap 두 설정은 작성자 race 뒤 추가되고 최종 vet/build에 포함되었다. root는 이 exact final source의 별도 pure metadata/stub race1회14PASS/skip0/fail0를 보고했다. 이는 root 보고이고 본 검토자의 추가 실행은 아니다. 이 검토자는 관찰기75 작성자이고 prior source/Want exposure가 있으므로 비맹검이다. 제어기 작성자는 다른 agent이지만 새 원천 의미나 Want 정확성의 독립 판단이라고 부르지 않는다. 원본 init/API/테스트/관찰기·제어기/모델/fit 실행, 코드 변경, 공유 수정, 외부 게시0이며 source-only 검토로 실제 프로세스 종료 성공을 관측했다고 주장하지 않는다.
