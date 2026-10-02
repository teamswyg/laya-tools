# logfmt·percent 관찰77 준비 v2

두 공개 Go 원천에서 작은 동작 두 개를 관찰할 준비입니다. logfmt12 입력은 record와 key/value의 순서·중복·빈 값·오류를, dataurl12 입력은 Escape4/Unescape8의 바이트 percent 변환을 다룹니다. 24개 literal은24개 부모나 독립 목표가 아닙니다. 새 부모·후보·역할·학습 정답·모델·Fit·보호 최종2400 사용은0이고, 학습/운영/다양성 자격은 모두false입니다.

원문은 먼저 읽었고, 원 API 실행 없이 Wants를 작성했습니다. source-first77 제안은 `eaec641e…`, 공개 원문은 source-audit74의 고정 commit `0a25491e83df23809c30d3497113e1d4969256b6`과 연결합니다. logfmt revision은 `804e98fff868b206344991c57a8182172e5ba41e`, dataurl은 `d1553a71de50473073e188aa79cebf7f993f20fe`입니다. Go8파일 원문과 원래 module/license/doc16파일 총51,234바이트를 그대로 복사해 해시를 확인했습니다. 두 MIT 전문과 Go BSD-3 전문/NOTICE를 보존하며, MIT가 Go 유래 코드를 재라이선스했다고 해석하지 않습니다. 정확한 Go 복제 릴리스의 입증도 아닙니다.

logfmt에서는 map으로 합치지 않고 성공 pair와 실패 step을 순서대로 남깁니다. 매 getter callback 안에서 nil·길이·hex를 자체 문자열로 만든 뒤 다음 ScanRecord로 넘어갑니다. 원 API의 borrowed slice가 항상 owned라는 뜻은 아닙니다. empty key·unquoted equals·unterminated/invalid quoted value의 유효 접두부와 SyntaxError의 Msg/Line/Pos/실제 Error() 문구도 별개로 남깁니다. Pos는 byte 위치이며 문자 위치로 바꾸지 않습니다. Scanner의 정확한 input 소비량은 주장하지 않습니다.

v1에는 EOF 이후 ScanKeyval을 호출하는 잘못된 관찰 순서가 있었습니다. 별도 source peer가 Scanner token=nil과 기존 dec.pos의 slice 경로에서 panic 가능성을 찾았습니다. native로 재현하지 않았습니다. v1 Want39,675바이트/SHA `b21af96b…`, 소스·바이너리·테스트 기록은 `history-v1/`에 그대로 남습니다. v2는 terminal ScanRecord=false와 Err=nil이면 모든 추가 scan/getter/Err를 생략합니다. sticky 결과는null이고 `skipped_reason:"no_current_record_after_EOF"`입니다. Err가nonnull인 sticky error에서만 기존 guard가 있는 반복 ScanRecord/ScanKeyval/Key/Value/Err를 관찰합니다. 일반적인 무panic 안전성을 승인하지 않습니다.

percent는 Escape/Unescape만 호출합니다. plus를 공백으로 바꾸지 않으며, decoded NUL/high byte는 nil·길이·hex로 보존합니다. 오류 때 반환되는 nil 바이트와 성공한 빈 결과의 nil 여부도 합치지 않습니다. DecodeString/Decode/New/WriteTo와 parser/lexer를 호출하지 않습니다. 원문 lexer의 goroutine 경로, 임의 Reader/Writer, 환경·파일·네트워크 API는 관찰 범위 밖입니다.

v2 사전 Want는38,148바이트/SHA `5cc2498d…`로 봉인했습니다. 예정 직접 callback217회는 생성자11+1, ScanRecord28, ScanKeyval/Key/Value 각31, Err48, Error()24, Escape4, Unescape8입니다. 여기에는 원 package public 호출209회와 stdlib Error()8회가 포함됩니다. 내부 helper/초기화 실행 횟수를 동적으로 계측한 수치는 아닙니다. 실제 호출은 모두0입니다.

root go.mod는 상대 local replacements를 사용합니다. logfmt 원래 go.mod/go.sum은 불변이며 go-cmp는 원 테스트 의존성입니다. dataurl에는 원래 go.mod/go.sum이 없어 별도 maintainer용 go.mod를 명시적으로 추가했습니다. 원 Go 코드는 수정하지 않았습니다. offline Go1.27.1/CGO0/trimpath/buildvcs=false 빌드는 원 package initializer를 실행하지 않습니다. native 프로세스를 나중에 시작하면 main 전에 package/transitive stdlib 초기화가 실행됩니다. 기존 외부 wrapper의 예약은 Start보다 앞서지만 initializer를 직접 계측하지는 않습니다.

실행 예정인 worker는 정확한 plan/Want/source/binary 핀을 확인하고 새 child attempt를 예약합니다. 각 직접 callback 전에 fsync한 partial에 예약 카운터를 남깁니다. 오류·panic·불일치와 유효 접두부를 보존하며 Wants를 고치거나 native 재시도를 하지 않습니다. `completed_with_differences`는 모든24개가 일치했다는 뜻이 아닙니다. observer 오류는 exit1입니다. 결과/stdout은1MiB보다 작게 제한합니다.

원본을 import하지 않는 `pure/` 테스트만 실행했습니다. v2 race12top+2sub/vet/build가 통과했고, v1의 두 race/vet/build 성공과 원문 문제 발견도 별도 장부에 남습니다. root는 기존 generic wrapper를 재사용해 CPU1/Go soft heap256MiB/외부60초 process-group kill 후 Wait 한 번으로 나중에 한 child를 실행할 수 있습니다. OS RSS나 전체 반환 시간의 hard cap이 아니며 아직 측정값도 없습니다. 이 폴더·binary·원시 로그·helper는private이고 GitHub/HF 게시0입니다.
