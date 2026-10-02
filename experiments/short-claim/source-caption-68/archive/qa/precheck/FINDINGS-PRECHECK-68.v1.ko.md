# 요청·대상68 v4 · 실행 전 읽기 전용 검토

동결한 v4 요청과 좁은 target 범위에서 **현재 중대한 request/target/제안 acceptable 불일치를 찾지 않았다**. 정적 원문 해석과 byte/metadata 확인 결과이며, native Got·새 감독 라벨·다양성/학습 준비 승인이나 실행 결과가 아니다. 원본 worker는 help를 포함해 실행하지 않았다.

검토자는68 English request 초안, v4 계획·contract, native worker/metadata 코드와 pre-existing upstream caption의 작성자가 아니다. 이전62 roleplan과67 loader를 작성했고 관련 개발 자료를 본 nonblind 검토자다. 앞선 semantic reviewer는 이번 English 요청 초안 작성자 노출을 스스로 기록했다. 그 보고서는 참고했지만, 이번 판단은 v4의 실제 요청·contract·source doc/code를 다시 읽고 별도 stdlib Go byte 검증으로 확인했다. 다른 저자의 검토이며 blind 평가·독립 source origin·human-only 저자 인증은 아니다.

## 고정 자료와 기계적 확인

- 계획 `PLAN-68.v4.ko.md`:6,995bytes, SHA `1493d14e9228cb147f8941c98d828bc221e0abdfe0be907e27b3799ae875cdb6`.
- contract `contract-68.v4.json`:27,415bytes, SHA `818eed54f6b14b1f512025804cf7782d4a59d94ed30eeeaa46c56a270afbd0eb`.
- 요청4개, 후보10위치, distinct 완전 원문5개, whole source 가족2개.14 text positions는 요청4+후보10이며14 native API 호출이 아니다.
- 원본 source assets15개/99,035bytes와 native-preparation 복사본의 exact bytes/SHA, MIT 전문2개를 확인했다. model/compile execution이나 source API 검증은 아니다.
- 요청 bytes/normalized words는 순서대로72/14,80/15,137/25,155/28이다. distinct captions는111/17,38/6,60/7,49/5,159/26이다. 모두 기존512bytes/32normalized words 안에 있다.

원문 span의0-base `[start,end)` bytes·원래 line·SHA·본문을 확인했다. `//`, indentation/dash/backticks와 마지막 LF는 원래 포함된 범위대로 보존한다. NewVersion의 package bullet은38bytes 전체이며 함수 전용 주석으로 재작성하지 않았다. mid-component 예문의 `.txt`도 그대로다. 단어 count는 선언된 camel-case/alphanumeric 규칙을 별도 streaming 구현으로 확인했고, Features API를 호출하지 않았다.

## 부모 목표와 후보 Want를 분리한 해석

|원래 요청의 목표|후보별 사전 Want|제안 acceptable, 아직 truth 아님|
|---|---|---|
|고정 `v1.2.3` leading-v 허용 거부|Strict=false, New=true|`[0]`|
|같은 입력의 leading-v 허용|Strict=false, New=true|`[1]`|
|slash whole-name, 공통 src/.go 전제에서 direct+nested 둘 다 match|`[true,false]`, `[true,true]`, `[true,false]`|`[1]`|
|같은 전제에서 direct만 match하고 nested 거부|같은 세 후보 Want|`[0,2]`|

반대 목표의 요청에서 동일 후보의 행동 Want가 같은 것은 맞다. 요청 목표가 달라 acceptable proposal만 달라진다. 검사기의 Want-to-goal 집합 비교는 이 **사전 제안의 내부 정합성**만 확인하며 Want를 Got나 라벨로 승격하지 않는다.

Strict의 complete valid-v2 문장과 digit-only parser, 원래 String 문맥의 “leading v는 specification의 semantic version에 포함되지 않는다”는 설명은 고정 leading-v 거부 해석을 뒷받침한다. String을 실행하거나 String behavior를 학습 목표로 추가하지 않았다.

New caption의 “optional v” package-level bullet만 떼어내면 모든 함수가 v를 허용한다고 할 수 없다. v4는 이를 **고정 NewVersion(v1.2.3)로 구현한 library capability**로 좁혀 attribution을 명시한다. doc.go parsing context `[328,852)`, source의 default CoerceNewVersion=true→coerceNewVersion와 anchored optional-v regex가 그 연결의 근거다. DetailedNewVersionErrors=true도 source 선언과 같으나 기본 coercing branch에서는 detailed-error branch가 사용되지 않는다. 함수명·binding·Want·truth·role·review는 학습 특징에 넣지 않는 기존 계약을 유지한다.

Glob 요청은 slash/whole-name matching, `src/`와 `.go`를 공통 전제로 명시했다. normal star와 mid-component `**.go`는 non-separator wildcard이고, standalone `/**/`는0개 이상의 directory에 대응한다. 원문/고정 source 구조는 두 names에 대한 각각의 Want와 제안 acceptable을 지지한다. `.txt` 원문 예제를 `.go`로 편집하지 않고 구조적으로 같은 fixed instance임을 옆자료로 명시했다. 모든 pattern placement·Windows·filesystem·error/panic·성능을 승인한 것은 아니다.

## 남는 범위 한계

이 contract에서 Semver 모델 target은 **v1.2.3 accepted bit뿐**이다. `1.2.3`, `1.2`의 accepted Wants와 NewVersion.String 세 값은 auxiliary 관측이다. missing-patch/정확한 정규화 문자열, nil/error/panic behavior 전체는 caption이 약속한 추가 목표가 아니다. v3 generic-attempts fullfinitecoverage는 계속 held/unsupported이며 v4로 소급 승인하지 않는다.

New package-capability attribution은 context-dependent이며 snippet 하나의 보편 계약이 아니다. StrictNewVersion 이름은 원문 안에 있어 lexical shortcut 가능성이 있다. 반대 목표 요청은 같은 후보5개를 재사용하고 같은 source의 helper/type/regex/init/flags/observer/validation/sentinel을 공유한다.4요청을4독립 source group으로 늘리거나 재사용 literal을 새로운 독립 사례로 세지 않는다. 새 NewVersion 호출 범위는57의 Strict-only 관측 확대라고 주장하지 않는다.

원래72요청/216후보/unknown21/17그룹·정답과67 mask는 변경하지 않았다. 실제 union·role·Features·fit을 수행하지 않았고 source diversity/training_ready는 false다. 준비한 pre-existing upstream wording assets는 origin 근거의 일부지만 실제 모델 입력 사용·일반화·protected-final2400·효용/절감이나 전체 fit 자격을 증명하지 않는다. 새 수치 gate나 인간 재승인 절차를 추가하지 않았다.

## 실제 Got 이후의 별도 확인 범위

부모가 원래 controller receipt·동결 worker/compile closure/binary·결과의 exact pins를 제공하면 바이너리를 재실행하지 않고 저장 runtime trace를 읽는다. accepted는 returned && !panicked && !error_present && version_present로 다시 도출하고, panic/미반환을 유효 reject로 바꾸지 않는다. Glob은 returned/no panic/no error일 때 Got_defined를 확인한다. Want와 비교 결과를 실제 Got에서 다시 계산하며, 사전 Want로 Got를 채운 흔적이 없는지 확인한다.

planned entrypoint12회(Strict3/New3/Match6)와 NewVersion.String observer3회, flags 전후값을 별도로 확인한다. 개별 init/helper 호출수는 계측했다고 주장하지 않는다. String은 accepted New에만 호출하며 Strict.String은0이어야 한다.6version+6glob+3String의15비교는 target과 auxiliary가 섞인 관측 집계이며 모델 정확도나4요청 평가 성공으로 바꾸지 않는다.

그 다음 실제 Got vector와 원래 요청 목표에서 acceptable을 다시 도출해 사전 `[0]/[1]/[1]/[0,2]`와 비교하고, 이번 좁은 source/text target만 별도로 확인한다. mismatches/partial failures는 보존하며 Want·초안·old labels를 바꾸거나 retry하지 않는다. 이 목록은 원래68 계약의 검증 범위이고 새 gate가 아니다. 지금은 Got 확인을 완료했다고 보고하지 않는다.

## 실행 장부

stdlib-only metadata helper1회 성공/실패0, bounded file reads36회/476,472payload bytes다. 이는 helper의 읽기 집계이며 native latency/RSS가 아니다. word/span toy tests2개를 race1회로 검증해2 pass events/실패0, package1.756초였고 vet1회 통과했다. 포맷 대상은 새 private helper/test만이다. upstream package를 import하지 않으므로 원본 init/API 실행0이다.

보고 전 읽기 명령19회 중 탐색2회가 실패했다. 없는 private ledger glob을 먼저 추측했고, 존재하지 않는 textnorm 디렉터리를 추측한 검색이 실패했다. 실제 필수 artifact는 존재했고 핀 검사 실패0이다. 긴 출력2회는 잘려 필요한 projection/source 범위를 재읽었다. 잘못 추측한 JSON key의 null은 누락이나 정답 근거로 해석하지 않고 실제 parent keys를 확인했다. 실패 원문은 private logs에 남겼다.

원본 바이너리/help/API/init, 원본 AST/formatter/Bind/Generate, Features·roles/order·fit/models/judge/paid·새 관측 라벨·protected-final, 공유 edit/Git/외부 게시 모두0이다. 별도 프로세스0은 일반 AI-assisted Codex 협업 비용0을 뜻하지 않으며 그 비용은 측정하지 않았다.
