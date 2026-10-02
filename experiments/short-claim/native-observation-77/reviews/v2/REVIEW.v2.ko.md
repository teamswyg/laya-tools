# 77 v2: EOF 변경 부분 검토

좁게 검토한 v2 변경에 남은 실행 장애는 없습니다. `pure/run.go` 148–150행은 terminal Err가 nil이면 바로 `no_current_record_after_EOF`로 반환합니다. 뒤의 ScanRecord/ScanKeyval/Key/Value/Err를 예약하지 않습니다. sticky 결과는 null이며, “호출해서 false가 나왔다”와 구분됩니다. nonnil 오류에서는 기존 guard가 지원하는 반복 관찰을 유지합니다.

원 v1 Want와 검토 기록을 고치지 않고 별도 v2를 비교했습니다. 7개 EOF-nil fixture는 각각 ScanRecord 1, ScanKeyval 1, Key 1, Value 1, Err 2를 계획에서 뺐습니다. 오류 fixture 5개의 값·sticky 관찰·예산은 같고, 12개 logfmt의 기존 record 내용/입력과 percent 12개 fixture 전체 필드도 같습니다. schema와 nullable DTO 변경은 명시적 v2입니다.

고정 API 순서의 예상 호출 수는 `[11,1,28,31,31,31,48,24,4,8]`입니다. 합계 **217 = primary callback 193 + 원 logfmt SyntaxError.Error 16 + 표준 라이브러리 ErrorString 8**입니다. 따라서 원 패키지 public 호출은 209, 표준 오류 메시지 호출은 8입니다. 이 숫자는 실행 전 계획으로 실제 호출 수가 아닙니다. 24 fixture는 같은 두 행동 목표의 유한 관찰이며 새 부모 24개나 훈련 라벨이 아닙니다.

Want `5cc2498d…`, draft plan `5ee1d22f…`, HANDOFF `7d6250f5…`, 바이너리 `f3d2bfb8…`의 실제 바이트 길이/SHA를 확인했습니다. 원천/라이선스 16개와 유지보수 소스/recipe 10개 핀이 맞고 plan과 handoff의 Want·binary·원천 목록이 같습니다. 바이너리를 실행하지 않고 build info만 읽어 Go 1.27.1, CGO 0, trimpath, Darwin arm64와 상대 경로 local replacements를 확인했습니다. 이것은 독립적인 compiler 증명이 아닙니다. 원 dataurl의 module 파일 부재와 별도로 추가한 유지보수 go.mod를 구분합니다.

저자는 EOF와 sticky-error의 가짜 callback regression을 실행해 통과했다고 기록했습니다. 검토자는 테스트 본문을 읽었으며 재실행하지 않았습니다. 원 native 초기화/API/관찰기/테스트/모델/Feature/Project/Fit 실행은 0입니다. 실제 실행 시 초기화는 main counter보다 먼저 일어나므로 기존 root 바깥 프로세스 제어와 분리됩니다. 이번 검토는 그 제어기를 다시 검토하거나 새 승인 흐름을 추가하지 않습니다.

v1의 잘못된 범위를 바꾼 소스/메타데이터 검토만 완료했습니다. 실제 24개 결과 일치, 보편적 함수 정답, 훈련 자료 다양성 또는 최종 일반화를 주장하지 않습니다. 새 부모·라벨·역할은 0이며 qualification/training/production/final은 false입니다. [영수증](RECEIPT.v2.json)과 [장부](ATTEMPT-LEDGER.v2.json)에 전체 핀과 한계를 남겼습니다.
