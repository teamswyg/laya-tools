# 60단계: 소스 선언과 과거 해시 연결 준비 — 계획 초안

60단계는 공개 소스 선언의 과거 digest 연결과 후속 문장 검토 절차를 준비한다. **설명 승인·새 정답·학습이 아니다.** 코드·입력·바이너리·프로토콜 봉인은 pending, 실제 inventory 실행은0회다. 이 문서 자체는 봉인이나 검토 완료가 아니다.

작은 비결정적 힌트 모델에 근거를 공급하려는 준비다. “설명이 구현을 충실하게 묘사하는가”와 “구현이 요청을 만족하는가”는 다른 질문이다. `odd-positive`는 요청에 맞지 않아도 “양수 홀수만 남긴다”라는 설명은 충실할 수 있다. 기존 correct/wrong·acceptable/rejected·랭킹 점수를 설명 충실도의 답으로 복사하지 않는다.

## 기존 범위와 고정 입력

source root는 legacy36개·typed24개다. **typed** 구성 목록은 관계204개·고유 component ID70개·고유 formatted SHA53개이며 legacy36개를 포함하지 않는다. raw 선언의 고유 SHA 수는 미계산이며53개로 가정하지 않는다. 선언·참조 수는 독립 작업·학습 표본 수가 아니다.

아래 4개 원본 Go 파일과 결과 파일의 바이트·SHA를 고정 입력으로 사용한다. 원본 후보·helper·타입·sentinel을 편집하지 않는다.

| 파일 | 바이트 | SHA256 |
|---|---:|---|
| `internal/behaviorprobe/data.go` | 28420 | `b377343a64c019c205f69865286a21ce5d91dda3051c31aa5c69b75a40300559` |
| `internal/typedbehavior/spec.go` | 1228 | `013a486ecc1d28913c2dda8cf74b3068dcb193d8b77e534dc62e2f87b9ee6b80` |
| `internal/typedbehavior/state.go` | 19402 | `35324ed595bcf384c484f33d98c6866164bebb2b636986ee8ea51bb979bdad7c` |
| `internal/typedbehavior/flow.go` | 18242 | `ea7f09d31e15e561173932ac0bb61db17746cdce929e0f976d97994d282d898b` |
| `experiments/short-claim/results-56b.json` | 266818 | `4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4` |

## AST inventory가 확인할 것

파일 pin 검증 후 AST의 최상위 선언을 찾는다. source ID·root·component ID·관계 순서와 함수/method/type/global/const/sentinel의 node 종류를 유지한다. const/iota·여러 값 선언도 과거 formatting recipe의 block 범위를 보존한다.

범위는 0-based/end-exclusive byte와 물리 줄이며 `//line` 가상 위치를 쓰지 않는다. raw·formatted·scanner-normalized·serialized literal-table SHA는 구분한다. 과거 code/component/bundle recipe와 표준 import·Go 기록을 대조하며 해시 일치는 의미 승인이 아니다.

AST만 사용하며 새 타입 검사·object closure·`SourcePins`·importer·registry/callback/API 호출은 없다. 원본 package initializer·후보도 실행하지 않는다. Bind/Generate·feature/baseline/ranking·모델/fit은 범위 밖이다. 저장된 helper/type/sentinel 연결의 **재결합**은 새 semantic closure 승인이 아니다.

[Go 모듈](../../internal/sourceinventory/README.md)의 registry/dispatch 검사는 AST 식별자 이름과 선언 descriptor의 대응이다. 일반적인 이름 가림 해결이나 실제 런타임 함수 결속을 증명하지 않는다. 입력의 GoVersion 문자열과 실제 컴파일러 확인도 별개이며 실행기가 build-info·바이너리·코드 핀을 결속해야 한다.

이번 계산 정책은 **캐시 없이** 원래 root와 component 관계마다 계산하는 것이다. 파일 SHA·AST parse·포맷·scanner 정규화·bundle의 attempted/completed counters와 논리적 관계 비교를 구분한다. 물리 completed는 호출이 반환했음을 뜻하며 오류 반환도 기록한다. parse 성공은 별도 ParsesSucceeded, expected SHA 일치와 관계 완료도 별도 카운터다. 같은 enum block을 여러 ID가 공유해도 모든 관계204개를 유지한다. 후속 캐시 실험은 별도 버전에서 raw identity+recipe·hit/miss·실제 호출 수와 결과 보존을 검증한다.

[준비 recipe](source-inventory-recipe-60.json)에 저장된 metadata로부터 예상한 성공 작업량은 parse4·root60·component 관계204·포맷264·정규화112·bundle24다. **원본 실행에서 관측한 값은 아니다.** typed 함수/method 관계52와 root60이 정규화112를 구성한다. 고유 ID70/해시53을 실제 계산 횟수로 대체하지 않는다. raw 고유 수는 아직 미계산이며 작업량을 속도·CPU·RSS·GPU 측정으로 부르지 않는다.

회귀 경계는 pin 변경·누락/중복 선언·잘못된 root/component/node·local 선언·물리 위치·반복 관계·recipe 차이다. 자동 연결 검사와 문장 의미 승인은 별개다.

## 내용 검토 축을 유지한다

[59단계 recipe](content-review-recipe-59.json)의 네 축을 그대로 쓴다. 별도 축 이름이나 상태 schema로 조용히 바꾸지 않는다.

| 기존 축 | 검토 질문 |
|---|---|
| `source_fidelity` | 오동작까지 포함해 설명이 연결 구현을 충실히 묘사하는가? |
| `request_contract_coverage` | 요청의 범위와 설명의 포함·누락·모순은 무엇인가? |
| `observation_fields` | 반환·오류·상태·순서·소유권·무패닉의 관측 범위가 표현되는가? |
| `explicit_negative_boundaries` | 부정·포함/제외·잘못된 입력·미지원 범위가 명확한가? |

기존 상태 어휘는 `pending`, `consistent_with_scoped_evidence`, `omits_required_scope`, `contradicts_scoped_evidence`, `unsupported_or_uncertain`이다. **이번 단계의 실제 내용 상태는 모두 pending이다.** 이후 별도 봉인 절차에서만 scoped 판단을 기록할 수 있으며 그때도 새 truth label이나 전역 승인으로 승격하지 않는다. unknown truth와 pending review는 다르다.

후속 행은 원래 parent/candidate 위치·ID·source, request/caption pointer·UTF-8 SHA/bytes·문구 byte 범위, contract/source pins, helper/type/sentinel·observer·literal, 입력/관측 필드, 부정 경계·미해결 해석·reviewer/record SHA를 연결한다. metadata는 feature가 아니다. 원문512바이트·정규화32단어를 유지하며 잘라 통과시키지 않는다. aggregate checked/failed는 개별 vector Got가 아니다.

## 고정된 세 계열의 프로토콜 예시

선택은 사전 고정이며 이전 랭킹에 따른 caption 선택이 아니다. 아래는 검토 준비 범위이지 실제 review records가 아니다.

| 계열 | 기존 부모 / 후보 위치 / source ID / 다른 문장 / finite 사례 | 원본 위치 |
|---|---|---|
| stable-odd | 4 / 12 / 3 / 4 / 5 | `probes-56.json` parents 16–19 |
| atomic-commit | 4 / 12 / 4 / 4 / 18 | `probes-56b.json` parents 0–3 |
| error-identity | 4 / 12 / 4 / 4 / 11 | `probes-56b.json` parents 4–7 |
| 합계 | 12 / 36 / 11 / 12 / 34 | 원래 순서·관계 유지 |

`probe-05-ambiguous`, `typed56b-p01-4`, `typed56b-p02-4`는 unknown을 유지한다. 34는 기존 literal 사례이며 반복 위치를 곱한 독립 요청 수가 아니다.

- stable-odd: `odd-correct/positive/reversed`의 음수·중복·순서를 분리한다. `check`의 slices.Equal은 nil/빈 슬라이스를 구별하지 않는다. alias·capacity·할당량은 미관측이며 legacy panic/입력 변경은 unknown이다.
- atomic-commit: root 네 개와 `decodeAtomic/atomicConfig/ErrSyntax`, observer를 연결한다. Returned와 After, 정확한 sentinel enum, Panicked를 따로 본다. partial-result의 “or zero if invalid”가 prefix인지 전체 문법인지 남겨둔다. nonnil 목적지 전제를 nil 입력의 무패닉 주장으로 넓히지 않는다.
- error-identity: root 네 개와 `identityFacts/identityCause/Error`, fixture·observer를 연결한다. 반환 facts와 Cause.Code의 사후 변경을 구분한다. “Cause pointers are nonnil”의 의미와 “error chain 보존”의 범위를 남겨둔다. typed-nil·custom Is/As/Unwrap·전체 chain 불변성은 검증된 범위가 아니다.

## 실행 전과 완료의 경계

첫 실제 inventory는 자체 코드·지원 프로토콜·입력·build recipe·바이너리·계획 봉인 후에만 진행한다. 현재 원천 실행0회다. 성공은 선언·과거 digest 연결 확인이며 전체216개 내용 검토·동작 재관측·효용 검증이 아니다. 내용 판단은 후속 범위다.

원본72/216, 모든21 unknown, 그룹17/라벨 포함16, 연결 관계·순서·기존 정답을 보존한다. 핀·범위·연결이 틀리면 고정 오류/미완료를 기록한다. 해석·관측 범위가 부족하면 pending을 유지한다. 새15개 repo나 모든 fit 이전2400개 gate를 추가하지 않는다. 기존17/16은 기존 최소를 이미 통과했으며 서로 다른 보호 최종 요청2400개 목표는 별도 일반화 검증이다. protected final/CoSQA reserve를 읽거나 재배정하지 않는다.

이번 준비는 Go 모듈·합성 검사·계획 초안을 코드 검토와 CI 대상으로 기록한다. 원본 inventory/내용 records 생성·새 SourcePins/Bind/Generate/후보/모델/유료 호출·labels/roles/fits/weights/승인/활성화는0이다. 59의738개 내용/closure 검토는 pending이다. 기존 저장소의 회귀·실제 모델 통합 검사는 유지되며 그 재생을 새60 수집이나 모델 성능 실험으로 세지 않는다. 사람 승인 절차를 추가하지 않는다.
