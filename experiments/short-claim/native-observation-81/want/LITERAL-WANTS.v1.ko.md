# Source81: 실행 전에 봉인한 32개 literal Want

[WANTS.v1.json](WANTS.v1.json)은 네 행동 목표의 원본 입력 32개에 대해 소스를 읽고 작성한 유한 관찰 기대값입니다. 원 함수나 동등 동작을 계산하는 대체 함수를 실행해 만든 결과가 아닙니다. 원 입력 객체·순서·`goal_id`·JSON 포인터·`expected: null`을 보존했습니다. 별도 원본 fixture ID는 없으며, `ordinal`은 위치를 설명하는 메타데이터입니다.

관찰 경로는 [Root의 사전 범위 확인](source81-root-scope-confirmation.v1.json)으로 먼저 확정했습니다. 앞선 [채널 제안](CHANNEL-PROPOSAL.v1.json)과 [그 당시 인계서](HANDOFF.v1.json)는 그대로 보존합니다. 앞선 문서의 원천 누락 상태를 소급해 고치지 않았습니다. 새 Source82 공급 인계서 `bb0db6e0…`의 57개 원천·모듈·라이선스 파일은 이번 작성에서 바이트와 SHA를 대조했지만, Source82의 48개 Go 1.27.1 선택 제안은 아직 실제 컴파일·링크 증거가 아닙니다.

Want 작성자는 Source81/82 원천 작성자와 다릅니다. AI의 도움을 받아 소스와 이전 제안을 읽은 비맹검 작업이며, 사람의 블라인드 판단·독립 원천 계보·학습 정답 승인을 주장하지 않습니다. `Got`, 학습 truth·role·weight·acceptable은 모두 null입니다. 새 부모·라벨·역할·학습·모델 호출은 0이고, 모든 readiness/qualification/production/final 플래그는 false입니다.

## 실제로 관찰할 채널

Hook 입력마다 새 factory를 한 번 만들고 `DecodeHookExec`를 한 번 호출합니다. 유효한 zero reflect 목적지를 쓰며, defined Time과 defined slice는 alias가 아닌 새 named type입니다. 원문 `source_type: int, input_literal: "7"`은 내장 `int(7)`로 해석합니다. 전체 Decoder, weak hook, cached hook 경로는 사용하지 않습니다.

Hook 반환은 type과 string/int/ordered string slice를 별도 필드로 기록합니다. 반환이 실제 `time.Time`이면 오류가 있어도 Date·Clock·Nanosecond·Zone을 각 한 번 읽습니다. String·Format·MarshalJSON은 호출하지 않습니다. 해당 반환 채널이 없으면 그 필드는 null입니다.

Pflag 입력마다 `scope81`/ContinueOnError의 새 FlagSet에 `items`를 usage 빈 문자열로 StringSliceVar 등록하고 Lookup을 한 번 합니다. 초기 상태와 각 고정 operation 뒤에 GetSlice를 한 번 읽고, nil 여부를 보존한 owned copy와 공개 `Flag.Changed`를 기록합니다. GetStringSlice·Value.String·Type을 추가로 호출하거나 반환 alias를 변조하지 않습니다. 초기 snapshot에는 operation 반환 채널이 없으므로 `operation_error: null`입니다. 성공한 operation의 오류 채널은 `error_nil: true`이며, 이는 초기 null과 다릅니다.

오류가 반환되면 바깥 Error 한 번, 지원되는 Unwrap 한 번, nonnil 원인 Error 한 번을 관찰합니다. 원인의 공개 `time.ParseError`/`csv.ParseError` 필드와 CSV sentinel identity는 직접 읽습니다. CSV의 `ParseError.Err.Error`를 별도로 호출하지 않습니다. 실제 예상과 다른 오류 type은 원래 type과 미지원 범위를 남겨야 합니다. panic은 primary/getter/error 채널을 구분하며, 문자열 원시 값만 텍스트로 보관하고 다른 값의 Error/String은 호출하지 않습니다.

## 소스에서 예상한 반환값

시간 목표의 layout은 `2006-01-02`입니다. 처음 네 입력만 실제 time 반환으로 예상하며 나머지는 변환하지 않고 원래 string/int를 반환합니다.

| Fixture 위치 | 입력/목적지 | 예상 반환 |
|---|---|---|
| 0 | `2024-02-29` → time.Time | Date 2024/2/29, Clock 0/0/0, Nanosecond 0, UTC/0, 오류 nil |
| 1 | `2024-02-30` → time.Time | Date 1/1/1의 zero time, 바깥 timeParseError + 원인 ParseError |
| 2 | 빈 문자열 → time.Time | 같은 zero time, 바깥 timeParseError + 원인 ParseError |
| 3 | `bad` → time.Time | 같은 zero time, 바깥 timeParseError + 원인 ParseError |
| 4/5/6 | 같은 유효 날짜 → string / *time.Time / defined Time | 원래 날짜 문자열, 오류 nil, 시간 채널 null |
| 7 | int `7` → time.Time | int 7, 오류 nil, 시간 채널 null |

존재하지 않는 날짜의 바깥 오류 문자열은 정확히 `parsing time : day out of range`입니다. `time.ParseError.Message`의 선행 콜론 앞에 wrapper가 공백을 더합니다. 원인 문자열은 `parsing time "2024-02-30": day out of range`이며, LayoutElem/ValueElem은 빈 문자열입니다. 빈 입력과 `bad`에서는 LayoutElem=`2006`, Message가 빈 문자열입니다. 바깥 메시지는 두 입력 모두 `parsing time as "2006-01-02": cannot parse as "2006"`이지만, 원인 메시지·Value·ValueElem은 입력을 보존합니다. 전체 문자열과 공개 필드는 JSON에 literal로 명시했습니다.

슬라이스 목표의 comma separator 결과는 순서대로 `[]`(nonnil), `[a]`, `[a,a,b]`, `[a,"",a]`, `["",""]`입니다. `[]int`와 defined `[]string` 목적지는 원래 `a,b` string을 반환하며 int 입력은 int 7을 반환합니다. 중복과 빈 요소를 삭제하지 않습니다.

Pflag의 초기 값은 항상 `[default]`, Changed=false입니다. Set은 첫 성공에서 대체하고 후속 성공에서 append하며, 오류에서는 둘 다 보존합니다. 단독 빈 Set은 nonnil empty slice, Changed=true입니다. `a,"unterminated`는 CSV StartLine=1/Line=1/Column=16/ErrQuote로 예상하며, 바깥 오류는 InvalidValueError입니다. 오류 뒤 첫 성공은 계속 첫 성공처럼 대체합니다. 성공 뒤 오류는 기존 값을 보존합니다. `"a,b",a`는 `["a,b",a]`입니다.

Replace와 Append는 공개 Flag.Changed를 바꾸지 않습니다. Replace null은 nil, Replace `[]`는 nonnil empty, Append `x,x`는 문자열 하나를 그대로 추가합니다. Replace/Append 뒤 첫 Set은 이전 값을 대체하고, 이미 Set이 성공한 뒤 Replace하면 후속 Set은 대체한 값 뒤에 append합니다. 이 차이와 42개 snapshot 전체를 literal 표로 명시했습니다. alias 변조를 하지 않으므로 이 관찰로 일반적인 ownership 보장을 증명하지 않습니다.

## 예상 예산과 아직 측정하지 않은 것

| 명시적 dispatch | 예상 횟수 |
|---|---:|
| Hook factory / DecodeHookExec | 16 / 16 |
| NewFlagSet / StringSliceVar / Lookup | 각 16 |
| GetSlice / Set / Replace / Append | 42 / 18 / 6 / 2 |
| 바깥 Error / Unwrap / 원인 Error | 각 6 |
| Date / Clock / Nanosecond / Zone | 각 4 |
| 합계 | 182 |

네 목표의 fixture별 합계는 `[6,9,9,9,2,2,2,2]`, `[2,2,2,2,2,2,2,2]`, `[6,8,6,8,9,11,11,6]`, `[6,6,6,6,8,8,10,10]`입니다. 고정 Root 범위의 보수적 최대는 338입니다. 148개 원래 route + 12개 upstream 오류 메서드 = 160개 original public dispatch와 22개 stdlib 명시적 dispatch로 예상합니다. 이 수치는 실제 실적이 아닙니다.

HookExec 내부 hook, registration의 Value.String, wrapper formatting 안의 cause.Error, CSV/time 내부 helper는 명시적 dispatch 예산의 동적 계측 대상이 아닙니다. package initialization도 읽은 소스 근거만 있으며 실제 init 호출/반환은 null입니다. 원 compile/import/init/API/tests/go list/worker/controller는 0입니다.

## 검증·실패·권리 범위

[메타데이터 검사](LITERAL-CHECKS.v1.json)는 32개 원본 객체의 포인터/순서/raw SHA/canonical SHA/null, Source82의 57개 파일, 기존 소스 근거 23개 범위와 신규 stdlib 근거 17개 범위를 연결합니다. cached Go 1.27.1 원문 6개와 VERSION/LICENSE의 8개 SHA를 기록했습니다. 메타데이터 작성 도구 컴파일은 2회 중 첫 1회 타입 표기 오류로 실패했고, 실패 소스를 보존한 뒤 표기만 고쳐 main 1회가 성공했습니다. 행동 정답을 계산하는 함수·원 테스트는 호출하지 않았습니다. 자세한 장부는 [LITERAL-LEDGER.v1.json](LITERAL-LEDGER.v1.json)입니다.

원천은 [mapstructure의 고정 MIT LICENSE](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/LICENSE)와 [pflag의 고정 BSD-3 LICENSE](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/LICENSE), Source81/82의 실제 Go attribution을 그대로 참조합니다. root MIT/BSD 및 복사된 Go 코드의 BSD 고지를 하나로 재허가하지 않습니다. cached Go 1.27.1 LICENSE SHA도 별도 기록했고 원문 재배포를 새로 하지 않았습니다. 계열/alias/helper/저자 흐름의 다양성은 아직 해결됐다고 하지 않습니다. 32개 유한 fixture는 32개 독립 부모나 보호된 최종 2,400개/domain을 충족한 자료가 아닙니다.
