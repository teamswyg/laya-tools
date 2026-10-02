# Source81: Want 채널과 호출 경로 제안

현재 단계는 **실행 전 관찰 범위 제안**이다. 고정된 source-first 자료의 4개 goal과 32개 입력 객체·순서·null을 유지하고, 이후 Want가 어떤 채널을 구별해야 하는지 정리했다. 최종 literal Want, Got, 후보 만족 정답, 부모·역할·가중치는 만들지 않았다. 원 패키지 compile/import/init/API/tests/go list 및 관찰기·모델 실행은 모두 0이다. Root가 경로와 채널을 확정한 후 별도 literal Want를 봉인할 수 있다.

원 proposal SHA는 `dcd7fbd62cbc53d35608069c1c179aeb97353acf2d7c9aaebc858d8db22f1962`, source-first Handoff SHA는 `14a03815901b4a0b0b47adc0f84c25cc56f00e32c56d854bb3fa97b58d96c0b4`다. fixture 객체에는 독립 ID가 없으므로 원 goal_id와 `/goals/G/input_only_fixtures/I` 포인터를 식별자로 사용한다. 편의상 추가한 ordinal은 원래 fixture ID가 아니다. 32회 fixture를 32개 독립 문제·새 부모로 세지 않는다.

## hook 반환 관찰

두 mapstructure goal은 fixture마다 factory를 한 번 만들고 `DecodeHookExec`를 한 번 호출하는 경로를 제안한다. factory의 anonymous 함수는 이름 있는 `DecodeHookFuncType`으로 직접 단언하지 않는다. decoder 전체, weak hook 또는 cached helper를 호출하지 않는다. 목적 타입은 유효한 zero reflection value로 만들고 진짜 정의 타입을 사용한다. `defined Time`과 `defined []string`은 alias가 아니다. 원 입력의 `source_type=int,input_literal="7"`은 원 객체를 고치지 않고 builtin int 7로 해석하는 명시적 fixture 설정이다. nil source/invalid reflect/named-string 입력은 추가하지 않는다.

반환 채널은 normal return, interface nil, 실제 동적 타입, tagged 값, error, panic으로 나눈다. tagged 값은 string/int/[]string/time.Time 또는 unsupported다. 문자열·정수·슬라이스는 full 값과 nil 여부를 보존한다. time.Time은 성공 여부와 관계없이 Date·Clock·Nanosecond·Zone의 네 getter를 각 한 번 호출해 날짜·시각·나노초·zone 이름/offset을 남기는 제안이다. 오류가 있어도 반환된 zero time을 버리지 않는다. `time.Time.String/Format/MarshalJSON`을 통한 암묵적 formatting은 사용하지 않는다. 직접 hook의 반환을 목적 변수 assignment로 설명하지 않는다.

원문에서 시간 hook은 builtin string이면서 정확한 time.Time 목적일 때만 parse한다. 다른 목적 타입과 int 입력은 원 data를 반환한다. slice hook의 목적 조건은 `t == reflect.SliceOf(f)`이며, 단순 Slice kind 판정이 아니다. 빈 builtin string은 nil 아닌 빈 []string을 반환하고, 순서·중복·중간 빈 항목을 보존한다. 이 내용은 원문 분기 분석이며 실행 결과나 넓은 모든 reflection 입력의 증명이 아니다.

## pflag 상태 관찰

각 fixture는 fresh `NewFlagSet(..., ContinueOnError)`와 `StringSliceVar`를 사용한다. FlagSet 이름 `scope81`, flag 이름 `items`, usage 빈 문자열, shorthand/deprecation/custom normalizer 없음을 **Root 확인 전 제안**으로 명시한다. 이름은 오류 문자열에 들어가므로 literal Want 전에 확정해야 한다. 등록 후 Lookup은 한 번만 하고, 얻은 Flag 포인터의 Changed 필드를 매 단계 직접 읽는다. SliceValue의 GetSlice를 등록 직후와 각 operation 후에 한 번 호출하고, nil 여부를 따로 보존한 owned snapshot으로 복사한다. 원 input/default/Replace 인수는 변경하지 않는다.

`GetStringSlice`, Value.String/Type의 추가 직접 호출이나 추가 alias mutation은 제안하지 않는다. GetStringSlice는 String/CSV 재파싱 경로라 다른 관찰이다. 등록이 내부에서 Value.String을 호출하는 것은 소스 경로로 따로 기록하며 독립 계측한 callback으로 더하지 않는다. GetSlice가 원문상 alias 가능한 view라는 사실과 observer가 그 순간 복사했다는 사실을 구분한다.

원문에서 첫 성공 Set은 대체, 다음 성공 Set은 append이고 CSV 실패는 값과 내부 changed를 바꾸기 전에 반환한다. Replace와 Append는 CSV 파싱과 내부 changed 변경 없이 수행된다. 공개 Flag.Changed는 성공한 FlagSet.Set이 바꾸며, unexported changed를 직접 관측했다고 기록하지 않는다. 오류 후에도 다음 고정 operation과 snapshot은 보존한다. Replace(null)과 Replace([]), 빈 Set, 중복 및 쉼표를 합치거나 정규화하지 않는다.

## 오류·panic의 채널

각 반환 error에는 nil·실제 `%T` 타입·outer Error 문자열을 구별한다. Unwrap 인터페이스를 지원하면 **한 번만** 명시적으로 호출하고, 반환 cause의 nil·타입·Error 문자열을 따로 보존하는 제안이다. time.ParseError의 Layout/Value/LayoutElem/ValueElem/Message와 csv.ParseError의 StartLine/Line/Column 및 ErrQuote/ErrBareQuote/ErrFieldCount identity 비교는 공개 필드를 직접 읽고, 추가 Error/Unwrap 호출을 하지 않는다. pflag GetFlag/GetValue getters도 추가하지 않는다.

시간 실패의 outer private timeParseError와 cause time.ParseError, FlagSet.Set의 InvalidValueError와 cause csv.ParseError는 서로 다른 채널이다. 예상과 다른 타입은 실제 타입·unsupported detail로 남기고 known으로 강제 분류하지 않는다. `%T`는 pinned fmt source상 Error/String을 호출하지 않는다. 그러나 pflag InvalidValueError.Error가 내부 fmt를 통해 csv.ParseError.Error를, csv.ParseError.Error가 sentinel.Error를 호출할 수 있다. 이 내부 source-path 호출은 별도 uninstrumented/null이며 명시 호출 합계를 모든 내부 호출 총수로 표현하지 않는다.

Panic은 정상 오류와 분리해 타입 및 primitive string만 기록하고, 임의 Error/String 호출을 만들지 않는다. 호출 전 채널은 null, 실패한 getter/error 관찰은 해당 채널 null+단계/이유로 보존한다. 예상값을 보고 원 입력·fixture·분기 조건을 바꾸거나, actual outcome 이후 Want를 수정하는 정책은 없다.

## 제안 호출 예산

고정 입력의 operation 수는 Set goal 12개, mutator goal 14개다. 등록 snapshot 16개와 transition snapshot 26개로 GetSlice는 42회다.

| 명시적 upstream 경로 | 제안 횟수 |
|---|---:|
| hook factory / DecodeHookExec | 16 / 16 |
| NewFlagSet / StringSliceVar / Lookup | 16 / 16 / 16 |
| GetSlice / FlagSet.Set / Replace / Append | 42 / 18 / 6 / 2 |
| 위 경로 합계 | 148 |
| 예상 outer Error / outer Unwrap | 6 / 6 |
| 예상 cause Error (stdlib) | 6 |
| 예상 time getter (stdlib) | 16 |
| 명시 observer dispatch 기대 합계 | 182 |

원문 기준 예상 오류는 time 3회와 CSV 3회이며 아직 실측이 아니다. 반환 오류 가능 action 42개에 대해 outer Error·Unwrap·cause Error를 최대 한 번씩, 16개 hook 반환 모두에 네 time getter를 허용하는 보수적 상한은 **338회**다(148+126+64). 이는 Root 확정 전 안전 예산 제안이며 달성·실측 값이 아니다. nil cause/비지원 Unwrap/타입 불일치/panic은 호출을 생략한 이유와 nullable 채널을 유지한다. reflection·입력 decode·formatting·등록 내부 helper를 포함한 총 CPU 또는 총 함수 호출 수를 증명하지 않는다.

## 남은 경계

원 source Handoff의 새 원문 12개와 과거 Go LICENSE 1개를 합쳐 보존 자료 13개가 있다. Go body는 mapstructure 3개와 pflag 4개다. 현재 compiler closure가 아니며 mapstructure 5개, pflag 38개 body의 취득·선택·고지·전체 init 검토가 남아 있다. 별도 source82 작업을 대신하거나 그 완료를 주장하지 않는다. 선택 원문에는 pflag CommandLine NewFlagSet 초기화 site와 mapstructure reflect TypeOf/Elem initializer expression이 보이지만 실제 init callback·return은 null이다.

Go 1.27.1의 time/format.go, time/time.go, time/zoneinfo.go, encoding/csv/reader.go, fmt/print.go와 LICENSE/VERSION을 읽고 바이트 SHA를 붙였다. 날짜-only Parse의 UTC 경로와 zero time의 nil Location 경로는 Local timezone 로딩과 구분한다. cached 표준 라이브러리의 고정 근거이지 새로운 official Git revision 또는 전체 stdlib closure 증명은 아니다. 전체 MIT·BSD 및 Go 차용 고지는 기존 source archive를 참조하고 재허가하지 않는다. mapstructure/wordwrap, pflag/Cobra 및 Go 저자 흐름 독립성은 미확정이다.

별도 Want 작성자이지만 source-first 설명과 과거 결과를 이미 읽은 AI-assisted/nonblind 협업이다. 인간 blind, 원천 독립성, 일반화 또는 학습 적격성을 주장하지 않는다. 현재 literal Want/Got·정답·역할·가중치는 null, 신규 actual parents·Features·Project·Fit·paid/model/native executions는 0이다. [채널 제안](CHANNEL-PROPOSAL.v1.json)과 [인계 기록](HANDOFF.v1.json)이 정확한 원본 포인터·핀·범위를 묶는다.
