# 다음 5개 작업의 실제 저장 결과 비교

Root가 저장한 첫 실행 결과를 봉인된 표준 라이브러리 Go 비교기로 한 번 비교했습니다. 고정 입력 23개와 후보 3개씩, 총 69행에서 만족 42·불만족 21·판단 불가 6입니다. 228개 조건은 알려진 일치 185·알려진 불일치 32·판단 불가 11입니다. 참조 후보 5개는 해당 23개 입력의 모든 Wanted 조건이 알려진 일치입니다. 반례가 있는 후보는 9개이고, 오류 writer 원본 후보 1개는 판단 불가입니다.

표의 위치는 0부터 시작합니다. T는 만족, F는 알려진 반례로 불만족, U는 판단 불가입니다. 참조 후보는 전부 T입니다. 상세 보고서는 각 조건의 Wanted·관측값·존재 여부·규칙, 실제 오류 타입·논리 kind·Error 메서드 반환·텍스트·후보 반환·panic을 보존합니다.

| 작업 | 원본 입력별 결과 | 원본 반례 위치 | 오류 후보 결과 | 오류 후보 반례 위치 |
| --- | --- | --- | --- | --- |
| mapstructure nil config | F F F T | 0, 1, 2 | T F F T | 1, 2 |
| Cobra error writer | U U U U | 없음 | T F F T | 1, 2 |
| Cobra required annotation | F F F T T | 0, 1, 2 | F F F T T | 0, 1, 2 |
| Cobra context preflight | F T T T | 0 | F T T T | 0 |
| Humanize BigBytes precision | F F F T U U | 0, 1, 2 | F F F T T T | 0, 1, 2 |

판단 불가 6행은 오류 writer 원본의 4개 입력과 BigBytes 원본의 음수·잘못된 정밀도 입력 2개입니다. `PrintErr`에는 반환 n/error가 없으므로 writer가 실제로 쓴 글자 수를 반환값처럼 대체하지 않았습니다. BigBytes 원본의 반환 오류도 관측 불가입니다. 따라서 오류 writer 원본은 false 라벨이 아닙니다. BigBytes 원본은 별도의 알려진 문자열 반례가 있으면서 두 행은 U로 남아 있습니다.

nil config와 빈 required annotation에서 원본 후보 panic 2개가 관측됐습니다. 해당 행은 Wanted의 panic=false와 알려진 불일치이며, 반환 오류는 U로 함께 기록했습니다. nil Result의 원본은 선언된 invalid_result 대신 unclassified_error 논리 kind를 반환했습니다. 오류 후보는 metadata의 세 nil 슬라이스를 빈 슬라이스로 바꿔 unchanged 조건을 위반했습니다. 오류 writer의 오류 후보는 단축 쓰기와 실패를 n=3·nil error로 보고했습니다. 취소 context의 비교 후보들은 원하지 않은 initializer/pre-run/run을 각 1회 실행했습니다. BigBytes 원본은 1.50/1.501/1.000000000000001 대신 1.5/1.5/1.0을 출력했고, 오류 후보는 출력이 맞아도 입력을 변경했습니다.

`missing_required` 사례는 후보 3개 모두 tracked Error 정상 반환과 정확한 텍스트 `required flag(s) "item" not set`을 충족했습니다. 실제 Go 타입은 `*errors.errorString`입니다. 논리 kind가 missing_required라는 사실은 새 Go sentinel이나 typed error가 존재한다는 뜻이 아닙니다. 비교에서는 errors.Is, Error, callback, 원본 후보를 다시 호출하지 않았습니다.

비교기 실행은 정확히 1회, 종료 코드 0, stderr 0바이트입니다. 입력 결과는 142,918바이트, SHA-256 `bfaa50f313a851348069e7d2fec4cd6680dd6dab963683285fe307f8b384af1e`입니다. 보고서는 86,140바이트, SHA-256 `c4db2ae8f33b29161af66fba5468edfd507a2b1c579c65c811a0c746478bd0bc`입니다. 실제 worker의 67회 정상 반환·2회 후보 panic·23회 Error 반환 및 마지막 sequence 184와 저장 행의 일관성을 확인했습니다. 외부 프로세스 실행이나 journal Sync 성공은 Root의 별도 증거이며 이 비교기가 증명하지 않습니다. 내부 API·원본 startup 호출 수는 null로 유지됩니다.

검토자는 Wanted·후보·observer의 작성자가 아니지만 이전 소스 검토와 수정 의견을 알고 있어 눈가림 검토가 아닙니다. 검토자가 작성한 비교기는 사전에 합성 검사 20개와 vet를 통과해 봉인됐고, Root가 json.go/compare.go 전체를 독립 검토한 다음 실제 비교를 승인했습니다. 그 준비 파일은 그대로 보존했습니다. 합성 통과는 실제 후보·모델 증거와 구분합니다.

원 worker 재실행, 원본 import·init·API, 모델·학습·projection·score, 보호 자료 읽기, Git·HF·네트워크는 이 비교에서 모두 0입니다. 비교기와 worker의 실제 메모리·GPU 성능 측정은 이 기록에 없습니다. labels_assigned=false와 qualified=false입니다. 알려진 양성과 반례에 따른 후보 채택·라벨·학습 반영은 Root가 별도로 결정합니다. 고정 입력 밖의 일반 정확성, 모델 품질, 비용 절감, 라이선스 보증 또는 새 공개 CI 통과를 주장하지 않습니다.

[상세 비교](FINITE-COMPARISON.v1.json) · [검증 기록](RECEIPT.v1.json)
