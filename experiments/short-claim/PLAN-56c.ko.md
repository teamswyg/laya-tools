# 짧은 주장 실험56c: 한 속성의 정답과 문장 준비 계획

[English](PLAN-56c.en.md) · [다음 방향](NEXT-56c.ko.md) · [56b 결과](RESULTS-56b.ko.md)

**이번 문서는 사전 준비 설계다. 새 속성의 공식 감사·역할 분할·순위 관측·학습·모델 호출·성능 측정·최종 평가는 아직 실행하지 않는다.** 이 문서를 작성하거나 commit하는 것만으로 실행 입력·원천·관측기가 봉인되지는 않는다. 실제 구현과 문구 검토가 준비되면 별도 실행 manifest를 결과 전에 고정해야 한다. 이 문서 작성 시점에는56b PR79의 CI가 진행 중이며, 검증 완료나 병합을 주장하지 않는다.

## 누가 무엇에 쓰려는가

가상 사용자는 짧은 Go 함수 설명을 보고 코드를 검사할 순서를 정하는 에이전트다. 예를 들어 “quote 안의 comma를 보존하는 후보” 또는 “문법 오류에 부분 토큰을 남기지 않는 후보”를 찾는다. 작은 도구는 후보의 **검사 순서만 제안**하고 외부의 독립 검사가 실제 코드를 확인한다. 모든 후보·원래 순서·fallback을 보존한다. 힌트가 낮은 후보를 삭제하거나 실행 승인을 자동으로 주지 않는다.

한 속성이 맞는 함수도 다른 속성에서는 틀릴 수 있다. 점수는 정답 확률·전체 함수의 정답·실행 승인으로 사용하지 않는다. 같은 작업에 작은 힌트를 여러 번 쓰는 방식의 효용과 실제 LLM 사용량은 별도 실험에서 확인해야 한다.

## 보존할 출발점

56b의 기존 합계72요청·전체17연결 그룹·라벨16연결 그룹은 역사적 개발 기록이다. 두 그룹 수가 준비 하한15를 넘었어도, 한 협업 합성 영어 작성 과정의 자료라는 한계와 역할 계획·다양성 부족은 남아 있다. 통계적 독립성이나 학습 준비를 뜻하지 않는다.

기존 구분자4요청의 unknown, 원문·후보·정답·그룹·봉인 manifest·결과를 그대로 보존한다. 좁은 속성 요청은 별도 ID와 원래 부모 연결을 가진 새 개발 자료로 준비하며, 기존 unknown을 known으로 덮어쓰지 않는다. 공개 원천120개 조사 snapshot과 protected final·CoSQA reserve도 보존한다.

이번 계획의 실행 권한은 준비 설계까지다. 역할0·분할0·fits0·새 가중치0·모델 호출0·새 순위0·새 성능 실행0·유료 호출0·protected final 접근/채점0을 유지한다. 기존 실제 코딩 요청7개·기록20개와 소진된 유료 호출 원장을 바꾸지 않는다.

## 최초 범위: 세 속성과 기존 아홉 literal

기존 `internal/typedbehavior/flow.go`의 고정된 원천 네 개만 재사용한다: `quoted-correct`, `quoted-literal-comma`, `quoted-inside-escape`, `quoted-partial-error`. 새 후보·원격 코드·사용자 콜백·동적 compile 입력을 받지 않는다. 닫힌 registry와 실제 코드/helper closure 해시가 맞는 원천만 지원한다.

| 속성 초안 | 기존 literal 입력 | 고정할 관측과 제외 범위 |
|---|---|---|
| `quoted-comma-content-v1` | `"a,b",c`, `ab"c,d"ef` —2개 | 두 유효 ASCII 입력에서 정상 반환, Count와 decoded token을 literal 기대값과 비교. Quote를 제거하고 quote 밖 comma는 분리. 이 최초 범위는 backslash가 없는 입력이며 오류 동작을 인증하지 않음 |
| `outside-escaped-comma-v1` | `a\,b,c`, `a\\,b` —2개 | Quote 없는 유효 ASCII 입력에서 정상 반환, Count와 decoded token 비교. Backslash가 다음 byte를 내용으로 소비함. 두 backslash 다음의 comma는 이스케이프되지 않아 분리됨. Quote 안 escape와 오류 동작은 범위 밖 |
| `syntax-zero-output-v1` | `"a,b`, `a\`, `a,b\`, `"a\`, `a,"b` —5개 | 기존 literal이 문법 오류로 정한 입력에서 반드시 syntax 오류, Count=0, 전체 `[8]string{}`. 원문128bytes·8tokens·토큰당 decoded8bytes 이내의 ASCII 문법 오류만 지원. ASCII/용량 오류는 인증하지 않음 |

합계9개는 기존 literal을 좁은 계약에 배치한 수이며 새 독립 요청9개·독립 원천9개·새 관측9개라는 뜻이 아니다. 고정 literal 기대값은 모델 출력이 아니며, 독립적으로 확보한 공개 원천 요청도 아니다. 각 속성은 finite literal 관측만 주장한다. 테이블에 없는 입력의 일반적인 함수 정확성을 증명하지 않는다. quote 안 escaped comma까지 확장하려면 별도 literal·관측·문구·계획을 먼저 고정한다.

기대 Count·token·오류는 후보 실행으로 만들지 않는다. 기존 독립 작성 literal에서 그대로 검토해 옮기고 새 표의 입력·관측 필드·기대값·버전을 함께 해시한다. 실제 공식 관측 전까지 새로운 허용 집합이나 assertion 수를 결과로 기입하지 않는다.

## 전체 함수 라벨과 속성 라벨은 다르다

다음은 **현재 코드 독해로 예상한 동작이며 새 실행 결과가 아니다.** 공식 감사에서는 속성별 literal 관측으로 다시 확인한다.

| 기존 source control | 기존 전체 계약 표시 | quoted comma 속성 예측 | quote 밖 escaped comma 예측 | syntax zero-output 예측 |
|---|---|---|---|---|
| `quoted-correct` | correct | 부합 | 부합 | 부합 |
| `quoted-literal-comma` | wrong | 불일치 | 불일치 | 불일치: 문법 오류를 반환하지 않음 |
| `quoted-inside-escape` | wrong | 부합 | 불일치 | 불일치: quote 밖 trailing escape를 오류로 반환하지 않음 |
| `quoted-partial-error` | wrong | 부합 | 부합 | 불일치: 문법 오류에 부분 출력을 유지 |

`quoted-partial-error`는 정상 파싱 경로가 정상 구현과 같지만 후반 오류 처리가 다르다. 초기 nonASCII 또는 raw 입력128bytes 초과에서는 먼저 오류로 반환하여 출력을 비운다. 따라서 “초기 오류에서 비움”을 “모든 오류에서 비움”으로 확대하지 않는다.

기존 감사의 “whole-contract wrong 소스는 최소1벡터 실패해야 한다” 조건을 속성 감사에 적용하지 않는다. Whole-contract control 표시·기존 부모의 known/no-answer/unknown·기존 acceptable 배열은 속성 라벨로 재사용하지 않는다. 같은 원천이 속성에는 부합할 수 있고, 전체 계약에서 오답만 있는 후보 세트도 한 속성에서는 답이 있을 수 있다. 기존56b 감사는 수정하지 않고 별도 속성 관측을 준비한다.

`syntax-zero-output-v1`은 non-vacuous 계약이다. `got.Error!=OK`일 때만 출력을 검사하는 방식은 쓰지 않는다. 고정된 다섯 입력 모두 지정 syntax 오류를 반환해야 하므로, 오류를 아예 생략한 후보도 불일치다. Count만0이면 충분하지 않으며 사용하지 않은 슬롯을 포함한 전체 토큰 배열이 비어야 한다. 정확한 오류 종류까지 검사하지 않는 더 약한 속성은 다른 계약으로 선언해야 하며 이번 속성과 혼용하지 않는다.

## 짧은 영어 문장의 완전성

실제 순위 입력은 요청과 후보의 영어 text뿐이다. 각 문구는 유효 UTF-8이며 raw·normalized 각각 최대512bytes, normalized 최대32words를 지켜야 한다. 공백으로 대충 센 단어 수 대신 기존 `shortclaim.Validate`와 동일한 정규화 규칙을 적용한다. 자동 요약·잘림·숨은 registry 정보로 빠진 뜻을 보충하지 않는다.

다음은 **후속 문구 검토용 초안**이다. 실제 길이 검사와 독립 의미 검토는 아직 수행하지 않았으며, 이 문서로 READY를 부여하지 않는다.

| 속성 | 영어 요청 초안 |
|---|---|
| quoted comma | ASCII without backslashes; ≤128 input bytes, ≤8 tokens, ≤8 decoded bytes/token. Preserve commas inside balanced double quotes, strip quotes, split outside commas; never panic. |
| outside escape | ASCII, no quotes or trailing escape; ≤128 input bytes, ≤8 tokens, ≤8 decoded bytes/token. Backslash consumes next byte as content; split unescaped commas; never panic. |
| syntax zero-output | ASCII ≤128 input bytes; ≤8 tokens, ≤8 decoded bytes/token. Unclosed quotes or trailing escapes must return syntax error, zero Count and eight empty Tokens; never panic. |

짧아진 범위가 실제로32words 안에서 충실히 전달되는지는 코드 검토와 별도로 확인한다. 후보 문구도 원래 소스와 이 속성의 조건·관측을 정확히 표현해야 한다. 원문 후보를 그대로 재사용하더라도 새 속성에 대한 충실성 검토를 생략하지 않는다. 새 설명이 필요하면 별도 버전·해시를 두고 기존56b 문구는 바꾸지 않는다. 한글·영문 문서 번역은 새 평가 입력이나 독립 요청을 추가하지 않는다.

## 새 계약·원천·특징·그룹

새 wire/result schema는 기존 v2와 분리한다. 제안 이름은 `riido-scoped-property-probes-v1`, `riido-scoped-property-truth-v1`, `riido-scoped-property-preparation-plan-v1`이며, 실제 구조·버전·제한은 후속 실행 manifest에서 고정한다.

각 부모는 새 ID, 원래 부모 ID, 속성 ID·버전, 요청, 원래 후보 순서와 후보 source/code/bundle pins를 가진다. 속성 정의는 지원 범위, literal 입력·기대값, 비교 필드, 오류·panic·unknown 정책, 문구 검토 상태와 표 해시를 가진다. 관측기는 실제 컴파일된 adapter 원천과 기존 네 source의 closure를 모두 고정해야 한다. 기존 `SourceArtifacts()`가 새 adapter 파일을 자동으로 포함한다고 가정하지 않는다. 새 compiled provenance manifest에 adapter·literal·schema·문구·관측 구현을 명시하며, 디스크 원천과 실행 파일의 일치를 검증한다.

순위 특징은 request/candidate text만 투영한다. property ID·원래 부모·source ID·함수 타입·코드/bundle 해시·literal·ExpectedControl·실패 수·허용 집합·역할·정답은 특징이 아니다. Metadata를 바꿔도 text가 같으면 투영은 같아야 한다. 독립 검사와 oracle은 감사 경로에만 있으며 순위 경로에서 정답을 조회하지 않는다.

같은 원래 부모·모든 positive/negative source·공유 `scanQuoted`와 자체 helper/type/global·원형·코드/문구 복사 관계를 전이적으로 연결한다. 기존 unknown 부모도 포함한다. 속성 ID를 새 prototype/core처럼 사용해 관계를 쪼개지 않는다. 속성·바꿔 쓴 문장·번역·seed·형제·INT8/PTQ child는 독립 그룹을 늘리지 않는다. 새 그룹 수나 라벨 그룹 수는 공식 감사 전에는 예상치도 성과로 기록하지 않는다.

## 준비 READY 체크리스트와 중단 조건

아래 항목은 **앞으로 충족해야 할 조건**이며 현재 통과 기록이 아니다.

- [ ] 세 속성의 범위·제외 조건·9literal·관측 필드·독립 기대값·오류/panic 정책을 고정한다.
- [ ] 기존 네 source/code/helper closure와 새 컴파일 관측기 원천을 함께 고정하고 Go1.27.1의 닫힌 지원 범위를 확인한다.
- [ ] 요청과 모든 후보를 실제 raw/normalized512bytes·normalized32words 규칙으로 검사하고 독립 의미 검토를 마친다. 초과·불완전 문구는 자르지 않고 보류한다.
- [ ] 새 schema가 unknown·known·no-answer와 이유를 구분한다. 지원 밖·불완전 설명·관측기 문제는 unknown이며 부정 라벨이 아니다.
- [ ] 필요한 오류를 생략한 후보, 잔여 슬롯, 안전하게 관측한 no-panic 위반을 속성 불일치로 처리한다. 관측 불가능한 상태를 통과로 처리하지 않는다.
- [ ] Metadata 특징 제외, 원래 후보·fallback 보존, 원래 부모와 모든 소스/helper 관계를 검증한다.
- [ ] 기존56b 네 구분자 unknown·전체 입력·결과·봉인 파일을 정확히 보존한다.
- [ ] 실제 구현·입력·문구 검토·원천·literal·실행 manifest를 commit으로 먼저 봉인한 뒤, 별도 계획으로 공식 감사1회와 CI replay를 진행한다.

지원 밖 입력이나 불완전 문구는 unknown으로 남긴다. 지원된 no-panic 계약에서 안전하게 포착한 후보 panic과 잘못된 반환은 known mismatch다. 손상된 JSON·해시·source binding, 풀리지 않는 closure, 신뢰할 수 없는 관측기, literal/문구 불일치, metadata 유출 또는 관계 누락은 준비나 감사를 중단한다. 결과에 맞춰 기대값·문구·후보·주 기준을 조정하지 않는다. 수정이 필요하면 원인과 새 계획·봉인 경계를 별도로 남긴다. 사람 승인 대신 CI 검증을 사용하되, CI 성공을 의미 검토나 원천 권리 검토의 자동 대체로 해석하지 않는다.

세 속성이 준비되어도 학습 READY는 아니다. 다양성과 학습/validation/calibration 역할·그룹 분할·새 무학습 효용이 고정되지 않으면 fits0을 유지한다. 최소15그룹 및 최소5% 효용 필요조건은 낮추지 않는다. 기존 protected final·CoSQA reserve를 읽거나 라벨로 쓰지 않는다. 도메인별 서로 다른 보호 최종 요청 최소2,400개는 별도 확보 목표다.

## 이후 효용·자원 경로

별도 효용 계획에서는 answerable/no-answer/unknown을 분리하고 고정순서·BM25·어휘·좁은 규칙·oracle을 동일 비용 분모에서 비교한다. 후보를 보존하고 unknown을 실패 라벨로 바꾸지 않는다. Oracle 개선 상한은 실제 모델 이득·LLM 토큰 절감이 아니다. 공개 원천의 서로 다른 개발 요청120→240, 언어·조건·작성 분포와 권리·revision·닫힌 compile·독립 정답을 준비한 뒤 역할과 선택 규칙을 결과 전에 고정한다.

별도 자원 계획에서는 캐시 없는 resident JSONL의 동일 요청과 서로 다른 공개 요청을 구분한다. Startup·warmup·처리·pipe·controller와 전체 child RSS/CPU를 나누고, queue 한도를 가진 open-loop는 그 뒤 검토한다. 캐시·lock·SIMD·GPU의 이득이나 작은 artifact bit 수를 전체 메모리·LLM 비용 개선으로 주장하지 않는다.

FP32 부모의 유용성이 먼저 확인된 뒤 동일 부모 INT8/PTQ child와 같은 예산의 ternary STE 형제를 별도 계획으로 비교한다. 기존 최대8fits·16파생 산출물 제안은 아직 집행하지 않는다. 새 유료 시도는 별도 실행 계획·예산 기록이 필요하다. 유용성·실제 원천 권리·CI 검증을 통과한 새 모델만 Hugging Face에 불변 버전으로 게시하고, Git에는 코드·계획·공개 수치·모델 참조만 둔다. 이번 문서는 새 모델 배포나 그러한 검증의 완료 기록이 아니다.
