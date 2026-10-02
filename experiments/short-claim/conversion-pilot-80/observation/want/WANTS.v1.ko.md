# Source80 후보 72개 유한 기대값

기존 입력 24개를 후보 9개에 적용하는 **실행 전 예상값**을 [WANTS.v1.json](WANTS.v1.json)에 봉인했습니다. 후보가 어떻게 실패하는지도 그대로 예상합니다. 이는 정답 라벨·학습 역할·실제 후보 관찰 결과가 아닙니다.

순서는 목표 0..2 → 후보 0..2 → 원 fixture 0..7입니다. `ordinal=g*24+c*8+f`, `original_fixture_ordinal=g*8+f`이며, 원 fixture ID/full input/Want/Got/input-observation 참조를 유지합니다. 기존 원문 20개, 원 Want·관찰 24개와 새 함수 9개의 정확한 byte/line 핀을 메타데이터로 대조했습니다. 모델 입력에는 향후 request+caption만 사용하며 코드·ID·Want·Got·원천 메타데이터를 숨겨 넣지 않습니다.

| 목표와 후보 | 고정 기대값 |
| --- | --- |
| ParseZeroOnError | 오류 3개에서 반환 UUID는 zero, 오류는 그대로입니다. |
| ParsePreserve | 원 Parse의 반환값과 오류를 유지합니다. raw32 늦은 오류의 `…ddfe00`과 standard36의 `…dd0000`을 합치지 않습니다. |
| ParsePanicOnError | 오류 3개는 오류값 panic입니다. 반환 UUID/error 채널은 null이고 panic의 presence/type만 남깁니다. panic_text는 null, 추가 Error/String/Format 호출은 0입니다. 원 MustParse의 문자열 panic을 재현한다고 주장하지 않습니다. |
| ScanPreserve | nil/empty/error에서 ff..00을 유지하고 성공에서 decoded bytes를 반환합니다. |
| ScanClearBefore | nil/empty/error fixture 0,1,2,3,6,7은 zero, 성공 4,5는 decoded bytes입니다. before는 **wrapper 입력 ff..00**이며 내부 Scan 직전의 zero local receiver를 뜻하지 않습니다. after는 wrapper가 반환한 UUID입니다. |
| ScanSuppressError | 반환 state는 유지하며 오류 2개를 nil로 바꿉니다. 해당 error_type/message도 null입니다. |
| OrdinalLastDigit | `0th,1st,2nd,3rd,11st,12nd,13rd,112nd`입니다. |
| OrdinalConstantTH | `0th,1th,2th,3th,11th,12th,13th,112th`입니다. |
| OrdinalPreserve | 원 8개 결과를 그대로 유지합니다. 음수나 whole humanize package를 새로 검증하지 않습니다. |

기존 pure.Channels는 **JSON 11필드**입니다. 이를 바꾸지 않고 오류값 panic의 텍스트 미관측은 별도 `Want_observed.nonstring_panic_text_unobserved`에 기록합니다. Parse panic 3행만 true입니다. 반환 오류의 Error 메서드만 한 번 부르도록 예상하며 panic payload와 억제된 오류는 추가 호출하지 않습니다. nil interface·typed nil bytes·nonnil empty bytes는 입력 메타데이터에서 구분됩니다.

예상 primary 72개 = 정상 반환 69 + panic 3, 명시적 returned Error 10 = UUID URN 2 + stdlib errorString 8입니다. 따라서 **candidate 수준 tracked 예상 82, 최대 120(72+48)**입니다. 원 Parse/Scan/Ordinal의 내부 호출 예상 56(24/24/8)과 owned Ordinal 16은 소스를 읽어 얻은 별도 추정이며 실제 계측값이 아닙니다. `56+48=104`도 다른 분모입니다. UUID import-startup의 네 source site, Scan 재귀와 formatting 내부 호출의 실제 반환 계수는 미계측/null입니다.

작성자 `semantic_review60_prep`은 원 Source80 준비 저자이며 새 후보 코드/관찰기 작성자는 아닙니다. 기존 Got를 이미 본 비맹검 AI 보조 작성입니다. 고정 base Want와 [LITERAL-OVERRIDES.v1.json](LITERAL-OVERRIDES.v1.json)의 수동 literal 변경표를 조립했고, 저장 Got는 원 참조를 대조하는 데만 사용했습니다. 이것은 native 행동 재현기나 독립 학습 정답 판정기가 아닙니다.

stdlib-only 메타데이터 조립 첫 1회 성공/실패 0, 별도 JSON null/count 확인도 통과했습니다. 원 API/import/init/compile/tests/native/observer 재실행, Features/Score/Fit/HF, 공유 수정은 0입니다. Got/truth/role/weight/label/mask/acceptable은 모든 72행에서 null이고 actual parent 생성은 false입니다. 세 예정 요청과 두 upstream family이며 72개 독립 부모·2400 final·training_ready로 승격하지 않습니다. 이 협업의 AI 사용 비용은 측정하지 않았습니다.

라이선스는 UUID full retained Darwin 15파일+원 go.mod/BSD-3-Clause와 humanize Ordinal-only 원문 slice+원 go.mod/MIT를 구별합니다. 원 20핀 및 후보 제안의 전문 고지를 그대로 참조합니다. full humanize·WTFPL 범위나 추가 원문을 얻거나 일괄 허가했다고 주장하지 않습니다. 독립 peer와 이후 별도 동결 실행은 root가 진행합니다.
