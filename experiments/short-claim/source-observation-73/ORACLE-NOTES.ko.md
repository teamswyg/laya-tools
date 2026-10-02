# 세 원천의 독립 oracle 메모 — 원본 실행 전

관찰기 작성과 독립된 source/text 검토다. pinned upstream 세 파일과 source-build handoff만 읽었으며, 관찰기의 `pure/spec.go`·Want·실행 코드는 아직 읽지 않았다. Handoff의 넓은 목표 초안은 이미 보았으므로 그 부분은 blind가 아니다. 아래 값은 원문에서 도출한 **독립 예상/범위 제안**이며, 동결된 공식 Want·관측 Got·새 학습 라벨·훈련 준비 승인에 해당하지 않는다. 원래 API/import-init/원본 테스트/바이너리/모델/fit 실행은 0이다.

## datasize: 숫자뿐 아니라 오류 종류와 receiver를 함께 판정

원문 `UnmarshalText` 116–217행은 ASCII 정수 prefix 뒤의 unit을 처리한다. KB/MB 등은 1000이 아니라 2의 거듭제곱이다(13–19행). unit 주변 공백은 제거하지만 숫자 앞 공백·분수·부호는 지원하지 않는다. 정확히 `Kb/Mb/Gb/Tb/Pb/Eb`를 lower-case 변환 전에 bits 오류로 분기한다. 따라서 `1Mb`와 `1mb`를 동일 취급하면 안 된다. 빈 입력은 loop를 건너뛰어 0으로 성공한다.

| raw input 예시 | 예상 receiver decimal | 예상 오류 |
|---|---:|---|
| `2 KB` | 2048 | 없음 |
| `1Mb` | 0 | ErrBits |
| `1mb` | 1048576 | 없음 |
| 빈 문자열 | 0 | 없음 |
| ` 1KB` 또는 `1.5MB` | 0 | ErrSyntax |
| `18446744073709551615` | 18446744073709551615 | 없음 |
| `18446744073709551616` | 18446744073709551615 | ErrRange |
| `18014398509481983KB` | 18446744073709550592 | 없음 |
| `18014398509481984KB` | 18446744073709551615 | ErrRange |

정상 oracle은 별도 정수 산술(정수 n×unit, 0..2^64−1 범위)로 확인하고, suffix/오류 우선순위는 해당 source 분기와 대조해야 한다. decimal 큰 수를 float64로 변환하면 정밀도를 잃으므로 uint64 또는 decimal string을 사용한다. 오류 oracle은 `*strconv.NumError`의 `Func="UnmarshalText"`, `Num=원래 전체 raw input`, `Err`의 실제 sentinel 동일성을 분리한다. 오류 문구만 비교하지 않는다. Syntax/bits는 receiver0, overflow는 receiver 최댓값으로 바꾼다(206–216행). 모든 오류에서 이전 receiver를 보존한다는 atomic 계약은 **원문과 충돌**한다. 초깃값을 7 등 비영 값으로 두면 이 차이를 관찰할 수 있다. input bytes before/after와 receiver before/after는 서로 다른 항목이다. `uint64(receiver)` 변환만으로 숫자를 읽을 수 있으며 다른 원본 helper 호출은 불필요하다. nil receiver/unsafe aliasing/추가 Parse·MustParse·String API는 범위 밖이다.

## querystring: map key 존재와 value slice 순서를 분리

`Values`는 `url.Values`를 반환한다. oracle은 **key의 존재/부재, 각 value slice의 정확한 순서·중복·빈 문자열**을 비교한다. 출력 편의를 위한 key 정렬은 가능하지만 value slice 정렬·중복 제거·Get의 첫 값만 비교하면 의미를 잃는다. URL 문자열 encoding 순서는 이 API가 직접 정하지 않으므로 primary oracle은 map이며, 문서의 encoded 예시만 그대로 답으로 삼지 않는다.

고정된 exported primitive/value-struct fixture에서 다음 구분이 필요하다: `url:"-"` 및 zero+omitempty는 생략; omitempty 없는 empty string은 `[""]`로 남음; **empty slice/array는 omitempty 없이도 생략**(216–220행); 기본 slice는 반복 값이고 source 순서대로 Add(248–253행); 같은 key를 가진 여러 필드도 declaration 순서의 반복 값으로 합쳐짐; bool 기본은 `false/true`, int option은 `0/1`; nested `user`의 `name`은 `user[name]`이며 flatten하지 않는다. nil scalar pointer는 omitempty가 없으면 empty string 값, nonnil pointer-to-zero는 omitempty여도 pointer 자체가 비어 있지 않으므로 zero 값이 남는다(185–186, 208–214, 283–290, 321–335행).

오류 oracle은 top-level scalar 등 non-struct input에서 **nil map+오류**다(140–141행). nil interface/typed nil pointer는 **nonnil empty map+nil error**(126–137행)와 구분한다. 이 함수에는 receiver가 없다. 고정 primitive fixture의 input snapshot과 output을 구분하면 충분하며 일반적인 custom object 불변성까지 확장하지 않는다.

`Values`는 `reflectValue` 오류 시 이미 채운 map도 반환한다(144–145행). 그러나 오류 발생점은 custom Encoder와 그 nested/embedded 전파에 연결된다(201–204, 264–277행). 이번 callback-free fixture로 그 **부분 map+오류 경로를 관찰한 것처럼 표시하지 않는다**. custom Encoder/IsZero/Stringer, cyclic objects, arbitrary reflection types, time option은 현재 bounded 초안 밖이다. `omitempty`의 단순 설명을 empty slice 포함 전체 key-presence 보장으로 과장하지 않는다.

## shlex: 미완성 단어가 아닌 완료 prefix만 반환

원문은 고정 ASCII separator(`space/tab/CR/LF`), 두 quote 모드, escape, comment 상태 머신이다. POSIX shell 실행/변수 치환/완전한 shell escape 규격이 oracle이 아니다. 같은 구현을 그대로 복제한 reference보다 **손으로 추적한 상태 경계와 고정 literal token 배열**을 사용한다.

`one "two three" four`는 `one`, `two three`, `four`; `'' ""`는 빈 token 두 개; `a""b`는 `ab` 한 개다. start-state의 `#`는 comment를 시작하지만 `a#b`의 `#`는 in-word default로 그대로 남는다. single quote 안의 backslash는 literal이며 double quote/outside의 backslash는 다음 rune를 literal로 처리한다. 따라서 범용 shell 라이브러리나 `strings.Fields`를 정답 oracle로 쓰면 안 된다.

나중 잘못 닫힌 quote 또는 dangling escape가 있으면 **이미 완료된 token 배열+오류**가 반환된다. 예를 들어 `ok "unfinished`는 `["ok"]`+quote-EOF 오류이고, `ok` 뒤 즉시 마지막 backslash가 붙은 입력은 아직 완료된 token이 없으므로 `[]`+escape-EOF 오류다. 한 token이 whitespace/정상 EOF까지 도달했는지와 불완전 단어의 문자 수를 혼동하지 않는다. Tokenizer는 오류 때 partial token을 만들지만 Lexer.Next가 버리고(154–156행), Split은 이전 배열만 반환한다(403–415행). 오류까지 읽은 partial word를 추가하거나 이전 배열을 모두 지우는 계약은 둘 다 충돌한다.

quote EOF/escape EOF는 각각 원문의 두 고정 error message로 구분한다(284/302, 320/345행). export된 sentinel이 아니므로 새로 만든 같은 문구의 error와 identity 비교하지 않는다. 내부 `io.EOF` 종료는 Split이 nil error로 바꾸며, empty/whitespace/comment-only는 **nonnil empty slice+nil error**다. receiver API가 아니고 입력은 string이다. Reader 오류 주입·Tokenizer 직접 호출·nil Reader·Unicode whitespace 확대는 범위 밖이다.

## 다음 frozen Want와 비교할 최소 항목

정상값·오류 class/identity·nil-vs-empty·완료 prefix·receiver 변화는 독립 field로 유지한다. 관측된 finite literal을 source 전체 보장이나 후보 만족 라벨로 바로 확대하지 않는다. 같은 source의 정상/오류 여러 probe는 독립 source family 수를 늘리지 않는다. 다음은 author handoff가 동결된 뒤 이 메모를 바꾸지 않고 spec/Want와 대조하는 단계다. 차이가 있으면 위치와 원문 근거를 보고하고 원본 실행 전에 해결한다.

Source pin: datasize `51293273…583d9`(aa82cc1e…); query `644ee90c…83ff7`(965d79f2…); shlex `f34d676e…220b`(e7afc7fb…). Full SHA/line evidence와 작업 범위는 `SOURCE-REVIEW-RECEIPT.json`에 있다. 원문/hand-off 저작자임이나 human-only 다양성은 주장하지 않는다. AI-assisted 읽기 비용은 측정하지 않았다.
