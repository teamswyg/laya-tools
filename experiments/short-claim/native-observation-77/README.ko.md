# 두 공개 Go 원천의 유한 관찰77

Root가 v2 관찰기를 한 번 실행했고, logfmt12개와 dataurl percent12개, 총24개 fixture가 사전에 봉인한 Want와 모두 일치했습니다. 관찰한 목표는 레코드 디코딩과 percent 바이트 처리의2개입니다. 24개 fixture를24개의 독립 부모·목표·학습 사례로 세지 않습니다. 새 부모·라벨·역할·모델·학습 및 dataset qualification은0이며, training/production/source-diversity 승인은 여전히 없습니다.

[저장된 수치](saved-qa/NUMERIC-RESULT.v1.json)의 추적 callback217개는 원 패키지 공개 호출209개와 표준 오류 문자열 호출8개입니다. 원 호출209개에는193개 primary 호출과16개 `SyntaxError.Error`가 포함됩니다. Callback의 함수별 reserved/returned/panicked와 원래 input/Want/Got, nil/길이/hex, 중복 키·순서·오류 전 부분 결과를 [원래 결과](native-attempt/worker/results.json)에 유지했습니다. 결과·마지막 partial·stdout JSON은 동일43604 B/SHA입니다.

전체 관찰기는 real5.20초/user0.17초/sys0.46초, maximum RSS28.453125MiB(29,835,264 B)였습니다. 이 값에는 시작·GC·관찰 기록·호출 전 checkpoint와 fsync 저장이 포함됩니다. 순수 API 시간, 모델 추론 시간, API당 메모리나 RAM 절감을 측정한 것이 아닙니다. CPU1/soft Go heap256MiB/바깥60초의 한 native child가 exit0·retry0으로 끝났습니다. Soft heap은 OS RSS hard cap이 아닙니다.

[v1 기록](preparation/history-v1/HISTORY.v1.json)은 EOF 후 존재하지 않는 현재 record를 읽을 때 발생 가능한 source-visible slice panic 지적을 보존합니다. 그 v1 Want·source·pure-test 기록은 사전 준비 역사이며 v1 native 실행은 없었습니다. [v2 Want](preparation/WANTS24.v2.json)는 terminal `ScanRecord=false`/`Err=nil` 뒤 추가 scans/getters/Err를 생략하고 해당 값을 null 및 skipped reason으로 남깁니다. Nonnull sticky error 뒤의 제한된 반복은 유지했습니다. 현재24match는 모든 입력의 무panic·보편적 의미나 caption/훈련 정답을 입증하지 않습니다.

[Root preflight](native-attempt/ROOT-PREFLIGHT.v2.json)는 기존76의 generic outside controller를 재사용했다고 명시합니다. [실제 ledger](native-attempt/ROOT-ACTUAL-LEDGER.v1.json)의76 schema 이름은 재사용된 외부 실행 봉투입니다. Worker v2 binary SHA·frozen77 plan SHA·Want SHA로 이번77 실행에 묶였으며, 새76 모델 결과가 아닙니다. Native 초기화는 main의 계수기 전에 실행되므로 동적 계측하지 않았습니다.

이 아카이브는 **실행 가능한 도구가 아닌 정확한 선택-copy**입니다. [Manifest](manifest.v1.json)는53개 원본 bytes/SHA와23개 제외 자료의 해시를 기록합니다. Current/history Go 파일은 inert `.go.txt`로만 복사했습니다. 원본 frozen plan/preflight가 host paths를 포함하지 않아 그대로 복사했고 sanitization0입니다. Binary·private Go modules/helpers·raw OS logs·host-argument BEFORE-START는 게시하지 않습니다. Stage 제작자는 관찰기 작성자여서 자신이 만든 선택-copy 확인을 독립 의미 검토라고 부르지 않습니다. [기존 peer 검토](reviews/v2/RECEIPT.v2.json)와 [saved-record QA](saved-qa/RECEIPT.v1.json)는 각각 원래 범위로 보존했습니다.

원천은 [동결 source-audit74](https://github.com/teamswyg/laya-tools/tree/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74)에 연결됩니다. [logfmt MIT 전문](upstream/logfmt/LICENSE), [dataurl MIT 전문](upstream/dataurl/LICENSE), [Go BSD 전문](attribution/Go-BSD-LICENSE), [조건부 차용 고지](attribution/NOTICE-Go.txt)를 원문 그대로 보존했습니다. Root MIT가 차용 Go 코드의 BSD 조건을 대체한다는 뜻이 아니며, 해당 source/binary redistribution 조건·copyright·disclaimer를 유지해야 합니다. 원천 metadata와 명칭은 runtime feature나 새 권리·학습 승인 scalar가 아닙니다.

[선택-copy 장부](SAFE-COPY-LEDGER.v1.json)와 [SHA256SUMS](SHA256SUMS)가 별도로 있습니다. 이번 staging은 원 native/테스트/모델을 재실행하거나 공유 Git·HF·외부 페이지를 변경하지 않았습니다.
