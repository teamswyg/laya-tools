# 77 v1 Want의 실행 전 원천 비교

실행 전에 수정할 항목 하나를 찾았습니다. **logfmt-03..07의 EOF 이후 추가 ScanKeyval 검사는 현재 레코드가 없는데 호출하며, 소스상 panic 경로가 있습니다.** 해당 v1 Want는 정상 반환·panic 없음으로 기대하므로 이 상태로 관찰을 시작하기에는 맞지 않습니다. 실제 panic을 재현하거나 원 API를 실행하지 않았습니다.

근거는 고정 [logfmt decode.go](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/decode.go#L54) 54–87행과 v1 `pure/run.go` 148–158행입니다. ScanRecord가 EOF에서 false를 반환할 때 `dec.pos`를 초기화하지 않습니다. 고정 Go 1.27.1 Scanner 소스는 마지막 EOF split의 token을 nil로 저장합니다. 이어서 ScanKeyval은 그 nil line을 이전의 양수 pos에서 잘라 읽습니다. logfmt-01의 빈 입력과 logfmt-02의 빈 레코드는 pos가 0이고, logfmt-08..12의 오류 경로에는 nonnil Err guard가 있어 같은 경로와 구분됩니다.

Root는 별도 v2에서 EOF+nil Err 이후에는 현재 레코드의 ScanKeyval/Key/Value를 관찰하지 않고 null/skipped로 남기는 좁은 사용 범위를 선택했습니다. nonnil 오류의 sticky guard는 별도로 유지할 수 있습니다. 원 v1 Want, 소스 및 이전 SOURCE-FIRST-77 문서는 그대로 보존합니다. v2 검토가 완료됐다고 아직 주장하지 않습니다.

그 외 19개 fixture의 전체 Want와, 해당 5개에서 EOF 이전의 값/오류 부분은 읽은 원천과 부합합니다. 이것은 유한한 AI 원천 판독이며 실제 실행 일치나 새 훈련 정답 승인이 아닙니다.

| 범위 | 원천 판독 내용 |
|---|---|
| logfmt-01..02 | 빈 입력/빈 레코드, nil 값, EOF의 nil 오류 |
| logfmt-03..07 | 중복 키 순서, bare/equal-empty/quoted-empty의 nil, 무개행 마지막 줄, 인용 escape/한글/Unicode 쌍은 EOF 이전까지 일치; EOF 추가 호출은 모순 |
| logfmt-08..11 | 성공한 앞 항목과 실패한 Key/Value를 구분; SyntaxError Msg/Line/Pos 및 메시지는 line 1의 byte 위치 6/11/12/14 |
| logfmt-12 | 작은 Scanner buffer의 token-too-long 오류와 오류 guard |
| percent-01..04 | nil 입력/ASCII unreserved/예약 문자/00·7f·80·ff byte, 대문자 percent escape, 원 입력 불변 |
| percent-05..12 | 빈 성공 결과의 nil, plus 보존, 대소문자 hex/00·ff, 잘린 escape/잘못된 hex/원 non-ASCII의 nil+오류 |

[rfc2396.go](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/rfc2396.go#L75)는 늦은 오류에서도 누적한 prefix를 반환하지 않고 nil bytes를 반환합니다. 잘못된 hex의 오류 메시지는 잘못된 byte 대신 현재 percent rune을 써서 `0x25`를 표시합니다. 이 동작을 개선한 메시지로 바꾸지 않았습니다. 오류는 공용 sentinel이 아닌 새 errorString입니다.

Key/Value는 다음 ScanRecord까지 빌린 버퍼일 수 있습니다. 관찰기가 즉시 nil/길이/hex로 복사하는 소스와 원 API의 소유권은 다릅니다. 현재 fixture에는 실제 버퍼 alias 변경이나 모든 UTF-8/오류/Reader 경로를 검증한 증거가 없습니다. arbitrary Reader, 파일·네트워크, full DecodeString/goroutine, 원 invalid UTF-8는 범위 밖입니다.

WANTS24.v1은 39,675 B/SHA `b21af96b142b023d2d85c8448cfcba6fa29775f2454f5749898b2db18c808af0`입니다. 12 logfmt + 4 Escape + 8 Unescape = 24개 유한 fixture이고 행동 목표는 2개입니다. v1의 기계적 call budget 합계는 259이며 실제 실행 수가 아닙니다. 24개 부모/독립 문제/훈련 라벨로 세지 않습니다.

검사자는 Want를 보기 전에 봉인한 SOURCE-FIRST-77의 작성자이며 이번 observer/Want 작성자는 아닙니다. 이전 원천 판독·예정 fixture 노출이 있어 사람의 블라인드 검토가 아닙니다. 이전 한영 notes/제안 SHA 및 이번 Want SHA는 [영수증](RECEIPT.v1.json)에 함께 보존합니다. 해시 자체가 독립 시각 증명은 아닙니다. 원천 16개 SHA와 두 MIT/Go BSD NOTICE를 다시 대조했고, 원 import/init/API/observer/테스트·모델·Feature/Project/Fit·역할·공개 변경은 모두 0입니다. 학습·운영·최종 검증 자격은 false입니다.
