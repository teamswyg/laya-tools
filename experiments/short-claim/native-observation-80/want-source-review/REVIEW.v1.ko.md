# 실행 전 Want80: 원문 기반 독립 검토

동결된 24개 Want에서 **실행 전 수정이 필요한 의미상 blocker는 발견하지 않았다.** 이 결론은 선택된 입력과 채널의 원문 기반 기대값 검토다. 실제 Got 일치, 관찰기 구현·컴파일 closure, 넓은 정답, 학습 label 적격성을 승인하지 않는다.

검토자는 `semantic_review60_prep`이다. source-first80 작성자이므로 원천 설명 작성의 독립성은 없지만, Want 작성자 `checkpoint_cli_45` 및 관찰기 작성자 `task_expansion_52`와는 다르다. 이전 source·caption·failed-fit 자료에 노출된 AI-assisted, nonblind 개발 검토다. 새 Want가 원문과 맞는지 직접 읽었으며 작성자 CHECKS를 독립 의미 검증으로 재사용하지 않았다.

| 범위 | 원문에서 확인한 기대값 |
|---|---|
| Parse 0·2·4·6·7 | 정상 16바이트 반환과 nil 오류. 길이45는 EqualFold prefix 후 standard 경로, 길이38은 첫 바이트만 건너뛰므로 `[…]`도 선택 Parse에서는 허용된다. 별도 Validate의 계약으로 바꾸지 않는다. |
| Parse 1·3 | 같은 늦은 `ge` 실패라도 raw32는 값 대입 후 ok 검사라 `00112233445566778899aabbccddfe00`, standard36은 ok 검사 후 대입이라 `00112233445566778899aabbccdd0000`이다. 뒤 바이트와 실패 slot을 구분한다. |
| Parse 5 | 잘못된 nine-byte URN prefix는 decode 전에 반환하므로 UUID 전체 zero, value type `uuid.URNPrefixError`, 메시지 `invalid urn prefix: "bad:uuid:"`다. |
| Scan 8–11 | nil interface·빈 string·typed nil bytes·nonnil empty bytes는 각각 입력 객체에 유지하며, 같은 기존 receiver를 변경하지 않는다. ‘null UUID’ 주석을 초기화 대입으로 읽지 않는다. |
| Scan 12·13 | 정상 text는 Parse 성공 뒤 commit하고 raw16 bytes는 copy하므로 정상 receiver 값으로 바뀐다. 매 fixture를 새 nonnil literal receiver로 시작해야 한다. |
| Scan 14·15 | Parse의 지역 부분값은 오류 뒤 commit되지 않는다. text와 non16 bytes의 재귀 경로 모두 receiver 유지, `Scan: invalid UUID format`이며 오류 identity/Unwrap은 여기서 관찰하지 않는다. |
| Ordinal 16–23 | 원래 `int` 함수의 modulo10·modulo100 조건과 Itoa 반환으로 `0th,1st,2nd,3rd,11th,12th,13th,112th`다. 오류 반환 채널이 없어 `error_nil`도 null이다. 비음수 입력 선택은 음수 동작 수정이나 새로운 API가 아니다. |

핵심 근거는 고정 UUID 원문의 `uuid.go`(type20, sentinel51–52, URN error55–59, Parse95–145), `util.go`(xvalues20–40, xtob43–47), `sql.go`(Scan15–52), 고정 humanize `ordinals.go`(Ordinal8–25)다. 원문의 byte span은 source-first `/source_clauses`를 재확인했고, 그 밖의 sentinel·URN Error body도 전체 같은 SHA 파일에서 읽었다. `g`의 table 값255와 `e`의 값14를 byte로 shift/or하면 fe라는 계산은 원문을 읽은 정적 추론이며 원 함수를 실행한 Got이 아니다.

구체 오류 type은 Go1.27.1의 읽은 `errors.New`/`errorString.Error`와 `fmt.Errorf`에 한정한다. 두 format sentinel 실패와 두 Scan 오류는 `*errors.errorString`이다. Scan은 `%v`를 사용하며 `%w` wrapping을 약속하지 않는다. `%T` type 수집은 Error 호출과 분리하고, 반환 nonnil 오류 5개에만 직접 Error를 1번씩 호출하는 계획이 맞다: 원 UUID URN method1 + stdlib4. 직접 entrypoint24 + Error5 = 추적 dispatch29다. Scan 내부 Parse3·재귀1·formatting과 기타 stdlib 호출은 개별 계측된 반환 수가 아니다.

UUID hash.go15–18의 namespace initializer4는 **정적 call site**다. actual init callback/return count는 null, 계측false를 유지해야 한다. main의 token 확인 전 실행될 수 있으므로 root 외부 예약은 child 시작 전에 해야 한다. 전체 child CPU/RSS에는 startup·설정·관찰·저장이 포함되며 soft heap256MiB는 RSS hard cap이 아니다. 이번 검토에서는 원문 import/init·compile·tests·API·worker·모델/Fit을 실행하지 않았다.

입력24개 전체 객체·ID·원래 순서·JSON pointer, 20개 원문/module/license SHA, 16개 source clause와 4개 startup span, stdlib 원문2개 SHA를 metadata로 확인했다. UUID BSD-3-Clause 전문과 Ordinal-only slice MIT 전문도 같은 pin이다. 전체 humanize/WTFPL 원전, 추가 API와 upstream tests는 범위 밖이다. 모든 Got/truth/role/weight/acceptable null과 qualification·training·production·protected-final false를 보존했다. 원천2가 독립 저자2·학습 그룹2 또는 2,400개 미노출 final 증거라는 뜻은 아니다. 관찰기 자체의 후속 검토와 실제 mismatch 기록은 별도다.
