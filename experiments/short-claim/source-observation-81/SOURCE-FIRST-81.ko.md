# 새 원천 두 가족: 실행 전 읽기

이번에는 `mapstructure`의 타입 조건과 `pflag`의 문자열 목록 상태를 읽었다. 원천 12개 파일을 고정 커밋과 Git blob 바이트로 확인했으며, 실제 라이브러리 호출·초기화·테스트·관찰기·모델 실행은 모두 0회다. 네 행동 목표가 두 소스 가족에 속한다. 입력 변형이나 번역을 독립 가족으로 세지 않는다. 정답·후보 만족·학습 역할·가중치는 아직 null이다.

`StringToTimeHookFunc`는 문자열이며 목적 타입이 정확히 `time.Time`일 때만 `time.Parse`를 호출한다. `*time.Time`이나 별도 정의 타입을 같은 대상으로 취급하면 안 된다. 실패 반환값도 버리지 말고 반환 타입·값·오류·unwrap을 각각 관찰해야 한다. 오류는 이 버전의 비공개 `timeParseError`로 감싸질 수 있어, 원래 `*time.ParseError`와 최상위 오류 타입을 혼동하지 않는다. 정확한 날짜 실패 값과 문자열은 후속 사전 Want 작성자가 Go 표준 파서 근거와 함께 정해야 한다.

`StringToSliceHookFunc`도 아무 슬라이스에나 적용되지 않는다. 실제 조건은 `t == reflect.SliceOf(f)`다. 기본 `string`→정확한 `[]string` 범위에서 빈 문자열은 nil이 아닌 빈 슬라이스이고, 쉼표 분할은 순서·중복·중간 빈 항목을 제거하지 않는다. 약한 타입 변환용 별도 함수와 구분한다. 이름 있는 문자열 타입은 `Kind` 검사를 통과해도 내부 `data.(string)` 단언과 충돌할 수 있다. 첫 관찰 제안은 기본 타입만 다루며, 임의 reflection 값·nil receiver·범용 panic 안전성은 약속하지 않는다.

두 hook 생성기는 `any`에 익명 함수를 넣어 반환한다. 이름 있는 `DecodeHookFuncType`으로 무조건 타입 단언하는 adapter는 피해야 한다. 공개 `DecodeHookExec`의 타입 변환 경로와 유효한 reflection 인자를 사용하도록 후속 계획에서 명시한다. 직접 hook의 반환을 전체 Decoder의 목적 변수 변경과 합치지 않는다.

`pflag` 문자열 목록의 첫 성공 `Set`은 기본값을 대체하고, 이후 성공 `Set`은 CSV로 읽은 항목을 덧붙인다. CSV 오류가 나면 값과 내부 `changed`를 바꾸기 전에 반환한다. `FlagSet.Set` 경로에서는 외부 `InvalidValueError`가 추가되며 공개 `Flag.Changed`도 성공 후에 바뀐다. 값 자체의 `Set`을 직접 부르는 경로와 같은 오류·상태 관측으로 취급하지 않는다.

`Replace`와 `Append`는 CSV를 해석하지 않고 내부 `changed`를 켜지 않는다. 따라서 첫 `Set` 전의 `Replace` 이후에도 첫 성공 `Set`은 대체한다. 이미 성공한 `Set` 뒤의 `Replace`는 다음 `Set`이 덧붙이는 상태를 유지한다. `Replace(nil)`과 비nil 빈 슬라이스, 쉼표를 포함한 단일 문자열, 중복을 구분하는 입력이 유용하다. 기본값·Replace 입력·GetSlice는 원문상 backing slice를 공유할 수 있다. 소유된 복사라고 설명하지 않는다. `GetStringSlice`는 String/CSV 재파싱을 거치는 별도 공개 경로다.

네 목표의 입력 초안은 SOURCE-PROPOSAL의 각 최대 8개로 제한했다. Want나 기대 오류 문자열은 넣지 않았다. CSV helper는 첫 record만 읽으므로 전체 여러 줄 입력을 모두 검증한다는 요구는 현재 범위 밖이다. 상태는 매 단계 owned snapshot으로 남길 것을 제안하지만 adapter는 아직 작성하지 않았다.

현재 취득 목록은 컴파일 closure가 아니다. mapstructure는 reflection 버전별 파일 2개와 내부 errors 파일 3개가 원문 미취득이다. pflag는 runtime 후보 42개 중 4개 Go 파일만 취득해 나머지 38개와 그 고지·초기화 검토가 남아 있다. 선택된 pflag 원문에서 `CommandLine = NewFlagSet(os.Args[0], ExitOnError)` 초기화 호출 지점 1개가 보인다. 이는 정적 지점 수이며 실제 반환 호출 측정이 아니다. 후속 관찰은 새 FlagSet을 사용해도 전체 package import 초기화를 별도 예약하고 전체 child 수명에 포함해야 한다. 주석 속 예제 `func init`는 실제 초기화로 세지 않았다.

mapstructure는 MIT, pflag는 BSD-3-Clause다. pflag `flag.go`의 Go Authors 2009 헤더와 `text.go`의 Go 1.23.4 복사 문구를 그대로 보존했다. 고정 Go 1.23.4 LICENSE 전문도 추가 취득했다. 과거 2012 Go LICENSE와 본문 길이는 같아도 저작권 연도가 달라 바이트는 다르므로 두 파일을 유지한다. Go 1.23.4 `src/flag/flag.go` 본문은 아직 취득하지 않아 복사 코드 동일성은 unknown이다. [NOTICE](NOTICE-81.md)와 모든 원문 라이선스 전문을 함께 유지하며 전체 배포·학습 권리의 일괄 승인으로 해석하지 않는다.

HTTP 읽기는 19회 모두 성공, 재시도 0회, 응답 본문 합계 177,173B였다. 초기 11개 제안은 보존하고 고지용 Go LICENSE 한 파일에 한해 상한을 12개로 늘렸다. 추가 private 메타 helper는 첫 컴파일에서 JSON용 중첩 map 표기를 Go 타입으로 쓰지 않아 실패했다. HTTP 전에 실패했으며 실패 소스를 보존한 뒤 두 번째 helper 실행은 성공했다. 원문 compile/import/API 실행 실패가 아니다.

이 리뷰는 AI-assisted source/text reading이다. 이 Codex 협업의 비용은 측정하지 않았다. 별도 모델/API 추론이나 유료 trial 0은 협업 자체가 무료라는 뜻이 아니다. 과거 source53·source79·다른 source/caption 작업에 노출된 비블라인드 작성자다. pflag↔Cobra와 Mitchell/wordwrap 저자 흐름, Go 차용 관계는 분리 승인되지 않았다. 모델 입력에 코드·ID·소스 메타데이터·관찰 답을 숨겨 넣지 않으며, 이후 새 source-derived caption·정답·역할 사용은 별도의 고정 범위와 독립 검토가 필요하다. 아직 2,400개 final, 독립 일반화 또는 학습 준비를 입증하지 않는다.
