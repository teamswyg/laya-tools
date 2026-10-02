# 세 공개 원천의 좁은 관측 준비

이 자료는 **세 동작 목표와 24개 고정 리터럴 probe의 준비**입니다. 아직 실제 관측값은 없습니다. 실제 원본 API·관찰기 binary·원본 package 초기화 실행은 모두 0회입니다. 세 원천을 128개 문제나 2,400개 최종 평가로 세지 않습니다. 새 parent·caption·label·모델·학습·최종 자료 읽기·유료 API도 0입니다. `training_ready`, 전체 의미·caption·parent label 적격성은 false입니다.

| 목표 | API | 고정 probe | 보존하는 관측 |
|---|---|---:|---|
| datasize-unit-parse-state | ByteSize.UnmarshalText | 10 | uint64 정확값, 초기 receiver 7 → 최종 receiver, NumError 함수·입력·range/syntax/ErrBits identity와 문자열 |
| query-tagged-multivalue | query.Values | 6 | nil map과 비nil 빈 map의 구분, tag omission, 값 배열 순서, 중첩 key, 오류 |
| shlex-quoted-argv | shlex.Split | 8 | 빈 quoted word, nil과 빈 slice의 구분, malformed EOF 때 이미 완성된 prefix와 오류 |

기대값 Want는 고정 소스와 API 문서를 읽고 작성한 **source-based 유한 기대값**입니다. 원본 실행 결과를 보고 작성한 oracle이나 독립 원천의 광범위한 정답은 아닙니다. `wants.v1.json` 전체 bytes·SHA와 compiled literal specification을 호출 전에 대조합니다. 불일치 결과가 나와도 Want를 바꾸거나 자동 재시도하지 않습니다. 각각의 Want/Got과 exact 정수·오류·slice/map 상태를 보존합니다. 같은 query key의 값 순서는 그대로이고 key만 정렬해 map 순서의 잡음을 제거합니다.

`pure` package에는 upstream import가 없습니다. Stub observer·big.Int 경계값·checkpoint·panic·ownership·null/empty 차이만 테스트합니다. 원본 세 import는 `cmd/observe`에만 있고 준비 단계에서는 compile/link만 합니다. 기존 source-build 원문 10파일·39,013B, 원본 모듈·라이선스는 변경하거나 재복사하지 않습니다. private observer module의 replace만 그 원문을 참조합니다.

실제 child 실행은 root 담당입니다. 이 import binary가 시작되면 main 이전에 Go package 초기화가 발생합니다. 소스에서 datasize의 `errors.New`, query의 `reflect.TypeOf` 초기화가 보입니다. 초기화 종류·건수는 동적 계측하지 않으며 24개 직접 API 호출 카운터와 구분합니다. Root는 child 시작 **전에** 별도 외부 attempt를 O_EXCL로 예약하고 plan/source/binary/Want pin을 확인해야 합니다.

Root는 draft에서 `Frozen`만 false→true로 봉인하고 새 전체 plan SHA를 인자로 줍니다. Child는 source closure·Go1.27.1/CGO0/trimpath build identity·CPU1·Go soft heap 256MiB·Want를 확인한 뒤 **새 child output directory만** 예약합니다. 이미 존재하는 directory는 거절합니다. 예약 receipt와 empty final output을 fsync한 뒤 매 probe **호출 전에** DispatchReserved 카운터를 partial 결과에 atomic rename+fsync로 저장합니다. 호출 뒤 반환·panic·일치/불일치와 prefix를 다시 저장합니다. 저장 실패는 추가 호출을 멈춥니다. 정상 반환의 불일치는 완료 결과에 남으며 새 감독 label이나 학습 승인을 만들지 않습니다. 예상하지 못한 오류·panic 텍스트는 고정 문자열로 가리고 stack·host/source 경로는 출력하지 않습니다.

실행 인자는 `--plan`, `--plan-sha256`, `--attempt-dir`뿐입니다. 원본 API의 입력은 compiled 공개 finite literal/primitive struct뿐이며 env·file·network·사용자 Go code를 전달하지 않습니다. 결과와 각 partial JSON은 **1MiB 미만**입니다. stdout은 완료 상태 한 줄, stderr는 고정 실패 문자열이며 실제 원문 입력·profile·절대경로를 공개하지 않습니다. 모델이나 feature 실행은 없습니다.

Root 외부 경계안: GOMAXPROCS=1, GOMEMLIMIT=268435456, `/usr/bin/time -l`로 whole-child CPU/RSS를 관측하고 독립 process group을 **60초** 뒤 SIGKILL합니다. Kill 후 Wait가 있으므로 controller 전체 종료가 정확히 60초 안이라는 보장은 하지 않습니다. Go heap soft target은 hard RSS cap이나 GPU 메모리 측정이 아닙니다. 실패·timeout 때 root는 기존 partial/final bytes·SHA·nullable 카운터·OS 관측을 보존하고 자동 재시도하지 않습니다. 실제 command/host 경로/raw 로그는 private입니다.

저자는 관찰기 제작자이며 이전 프로젝트의 projection/controller 준비에 노출되었습니다. 저자 stub test·byte pin 검증은 독립 runtime 관측이나 caption 의미 승인으로 부르지 않습니다. Source pin+trusted compiler/build metadata는 준비의 provenance이며 compiled instruction이 원문과 같다는 형식적 증명은 아닙니다. Root의 별도 읽기 검토 이후에만 실제 child 1회가 예정됩니다.
