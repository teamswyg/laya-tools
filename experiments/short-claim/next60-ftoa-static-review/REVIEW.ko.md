# 봉인된 ftoa 준비의 정적 검토

현재 실행 준비는 **false**다. 22개 파일 63,996바이트와 주요 원문·Want·소스 핀을 확인했다. 코드·Want·기존 seal을 바꾸거나 Go·컴파일·함수·모델을 실행하지 않았다. 나는 정수 oracle 작성자이므로 oracle 독립 QA로 세지 않는다. worker/후보/Want 비저자의 비맹검 소스 검토 범위다.

실행 전 남은 경계는 네 가지다.

- main89–94는 once와 빈 결과 파일을 동기화하지만 호출 중간 기록은 메모리에만 있다. kill/timeout 뒤 original 호출 수는 0이 아니라 **unknown**이다.
- main39/86은 compiler_selection_pending을 자체 거절하지 않는다. 7개 파일 목록은 전체 compiler 선택 증거가 아니므로 Root 외부 고정 증거가 필요하다.
- 새 ftoa 18행 스키마는 기존 native2 23행 controller와 호환되지 않는다. 자료도 이를 BLOCKED로 명시한다.
- nonfinite Want의 null 문장값과 실제 Go 반환 빈 문자열, 원본의 unavailable 오류 채널은 별도 비교 정책이 필요하다. OracleChecks는 후보 만족 판정이 아니다.

원본986바이트는 소수6자리 표시 뒤 잘라내며 문자열만 반환한다. authored precise는 원하는 자리에서 반올림하고, flawed는 6자리 반올림·파싱 후 다시 반올림한다. 함수 signature와 실패 후보 변경이 별도 버전으로 공개돼 있다. 원래 draft의 Want6개는 유지됐다.

9/8과 11/8은 절반에서 짝수 방향을 다르게 확인한다. 1.999의 올림과 9자리 작은 수, 음수0 정규화·Inf 오류 정책은 소스 독해와 맞는다. 이는 아직 관측값이 아니다. 소수 리터럴의 실제 binary64 비트·컴파일·일반 정확성은 검증하지 않았다. source-static 예상표는 [RECEIPT.v1.json](RECEIPT.v1.json)에 있으며 정답/labels가 아니다.

전체 humanize와 compiler closure·CI·권리·자원 증거는 pending이다. 선택 원문과 MIT 고지는 전체 패키지를 승인하지 않는다. 이 자료는 기존 개발 요청 하나·유한 입력6개·예정 후보 관찰18개이며 새 부모/가족/학습 승격이 아니다. [장부](LEDGER.v1.json)에 실제 읽기와 한계를 기록했다.
