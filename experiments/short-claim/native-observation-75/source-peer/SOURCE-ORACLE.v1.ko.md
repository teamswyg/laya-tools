# Wrap / Marshal 원문 우선 의미 메모

관찰기 작성자의 Want·dispatch·handoff를 보기 전에 고정한 소스 해석이다. 라이브러리 import, init, API, 테스트, 바이너리를 실행하지 않았다. 검토자 `semantic_review60_prep`는 이번 관찰기 작성자가 아니다. AI 보조 검토이며 사람이 단독 작성한 oracle이나 실행 관측값으로 주장하지 않는다.

## WrapString

`go-wordwrap/wordwrap.go:16–82`는 문자열을 rune으로 순회하고 `wordBufLen`/`spaceBufLen`을 하나씩 증가시킨다. UTF-8 byte 수, grapheme 수, 화면 폭을 세지 않는다. 한글·한 rune인 Unicode 글자·탭도 각각 1로 센다. 일반 Unicode whitespace는 공백 버퍼에 넣지만 NBSP U+00A0은 단어의 일부로 취급한다(:47–62). `\n`은 별도 분기에서 항상 그대로 쓰고 줄 길이를 0으로 재설정한다(:26–46).

새 줄을 삽입하는 정확한 조건은 `current + wordBufLen + spaceBufLen > lim && wordBufLen < lim`(:64–69)이다. 조건이 참이면 현재 단어 앞에 newline을 쓰고 대기 공백을 버린다. 단어를 쪼개지 않는다. 매우 긴 단어가 폭을 초과할 수 있다는 원문 설명(:10–15)이 구현과 맞는다. 폭 0에서는 `wordBufLen < lim`이 성립하지 않아 자동 newline을 넣지 않는다. 폭 1에서도 비어 있지 않은 단어는 그 조건을 만족하지 않는다. 따라서 '폭 이내로 항상 맞춘다'는 oracle은 부정확하다.

공백을 항상 보존하거나 항상 trim하지도 않는다. 다음 단어가 시작할 때 줄바꿈하면 앞 공백이 사라진다. newline 직전 단어가 없으면 공백이 현재 폭에 들어갈 때만 쓴다(:27–35). 종료 시 단어가 없으면 후행 공백도 현재 폭에 들어갈 때만 쓴다(:73–76). 단어가 있으면 대기 공백과 단어를 모두 쓴다(:77–79). 선행 공백은 삽입 newline의 원인이 될 수 있다. 빈 입력은 빈 문자열이다. 반환형은 string뿐이므로 error나 부분 결과를 별도로 반환하지 않는다. 원본 string·외부 receiver를 수정하는 경로는 없으며 결과 버퍼는 함수 내부다.

원문으로 정당화할 유한 기대값은 입력 rune 순서, 공백/단어 버퍼, 위 조건을 그대로 따라 계산해야 한다. 'Unicode 친화적 화면 줄바꿈', 모든 폭에 대한 최대 길이 보장, invalid UTF-8의 원래 byte 보존은 이 구현의 약속으로 승격하지 않는다. 유한 fixture 결과도 광범위 계약이나 학습 라벨 승인이 아니다.

## Marshal

`godotenv/godotenv.go:183–193`는 map에서 만든 전체 출력 행을 `sort.Strings`로 정렬한 후 `\n`으로 연결한다. map 반복 순서가 아니라 완성된 행의 사전식 순서다. 보통의 key 순서라고 일반화하지 않는다. 마지막 newline은 추가하지 않는다. nil map과 빈 map은 모두 `"", nil`이다. map 값을 읽고 새 행 slice를 만들며 원본 map에 쓰거나 프로세스 환경을 읽고 수정하지 않는다. 이 함수의 명시적 반환은 항상 nil error다.

`isInt(:164–178)`는 선두 `-`를 한 번만 제거한 뒤 적어도 한 글자가 있고 전부 ASCII `0`부터 `9`일 때 true다. `+`·공백·분수·Unicode 숫자는 포함하지 않는다. 선행 0과 `-0`, 길이가 긴 숫자는 허용한다. 수치 변환이나 overflow 판정은 없다. 이 경우 값은 따옴표 없이 출력된다. 그 외에는 double quote 안에 `doubleQuoteEscape` 결과를 넣는다. 따라서 '모든 값은 따옴표로 감싼다'는 comment(:181–182)만을 oracle로 쓰면 구현과 충돌한다.

이스케이프의 실제 순서는 backslash, LF, CR, double quote, `!`, `$`, backtick이다(:26, :235–246). LF/CR은 각각 literal `\n`/`\r`로, 나머지는 앞에 backslash를 붙인다. 먼저 입력 backslash를 두 배로 만들므로 원래 literal escape와 실제 줄바꿈은 구별된다. 모든 whitespace나 single quote를 이스케이프한다고 확대하지 않는다. key는 검증하거나 escape하지 않으므로 임의 key의 dotenv 파싱 가능성·roundtrip을 보장하지 않는다.

`parser.go`는 같은 패키지의 escape/변수 확장 문맥을 설명한다. 하지만 Marshal은 `parseBytes`, `expandEscapes`, `expandVariables`를 호출하지 않는다. Unmarshal/Write/Load를 실제 관측 없이 Marshal의 직접 호출 사실로 대체하지 않는다. parser의 regex 초기화도 이 검토에서는 실행되지 않았다.

## 원문/라이선스 보존

실제 `go-wordwrap/LICENSE.md`와 `godotenv/LICENCE` 전체를 읽었고 두 파일은 MIT 허가·copyright·면책문을 포함한다. source 파생 자료를 배포할 때 원문 notice 전체와 정확한 ref를 보존하는 범위다. 소스 작성자 다양성, 사람만의 작성, 독립성·학습 준비를 승인하는 근거가 아니다. 정확한 파일 hash/byte는 SOURCE-RECEIPT에 고정한다.
