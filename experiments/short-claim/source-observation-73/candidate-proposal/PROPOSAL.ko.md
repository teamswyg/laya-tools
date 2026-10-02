# 관찰에서 후보 문제로: 세 부모의 draft

현재는 **부모 draft 3개·후보 위치 9개**만 준비했다. 실제 새 parent/label/role/fit/feature 실행은0이며 truth는 모두 `null/unknown`, acceptable set·role·loss weight도 null이다. Root의 새 native Got을 읽거나 답으로 사용하지 않은 상태에서 고정했다. `CANDIDATE-DRAFTS.v2.json`이 최종 설계안이고 v1은 역사적 초안으로 보존한다.

사람과 에이전트가 찾을 코드를 다음 세 요청으로 표현한다. 영문 요청이 모델 입력이며 한국어는 같은 부모의 표시용 번역이다. 후보 caption은 번역·재작성하지 않은 공개 upstream 문장/서명/짧은 함수 원문이다.

| 목표 | 사용자가 찾을 코드 | 선언 순서로 고정한 후보 |
|---|---|---|
| datasize | 정수·이진 단위를 읽고 오류 종류별 receiver 상태와 오류를 반환하는 바이트 크기 메서드 | UnmarshalText, Parse, MustParse |
| querystring | 중복 값 순서·빠진 값과 빈 값·중첩 대괄호 이름을 보존하는 구조체 쿼리 encoder | Encoder, Values, valueString |
| shlex | 빈 quote token과 잘못 닫힌 quote/마지막 escape 앞의 완료 token을 보존하는 문자열 splitter | Lexer.Next, Tokenizer.Next, Split |

각 요청은22/22/21개의 budget words, 후보는8–21개다. 최대 후보 수는각3개이며 caption은 최대122 B다. 이는 stdlib 기반의 별도 byte/word 확인이며 원래 Normalize/Validate/Features API는 실행하지 않았다. JSON은 정확한 source line·byte range·raw SHA·revision·license를 연결한다. 선언의 전체 코드와 의존 helper는 **검토용 metadata**이지 모델에 몰래 넣는 입력이 아니다. 모델 입력은 오직 **request+caption**이다. 파일명·source ID·SHA·license·role·truth·Want/Got·private 경로는 feature에 들어가지 않는다.

## 현재 가장 큰 한계: caption에 필요한 조건이 없다

모든 request/candidate의 coverage를 `pending`으로 적었다. 단지 복사한 byte가 원문과 같다는 것과, 그 caption으로 요청의 조건을 알 수 있다는 것은 다르다.

- datasize의 UnmarshalText에는 doc comment가 없어 **signature 한 줄**이다. 정수 문법·이진 단위·오류별 상태는 caption에 없다. Parse는 helper를 호출하는 전체 짧은 함수지만 helper 동작이 없고 receiver method도 아니다. MustParse의 panic branch는 오류 반환 요청과 잠재적으로 충돌한다. 이 점을 기록했지만 정답/오답 label은 부여하지 않았다.
- query.Values의 짧은 comment는 URL value encoding만 말한다. 순서·빈 값·중첩 이름은 나타내지 않는다. Encoder comment는 custom interface이며 현재 callback-free 관측 범위 밖이다. valueString comment는 단일 값 표현만 다룬다. 각각의 누락을 JSON에 따로 적었다.
- shlex.Split comment는 일반적인 string→slice 분리만 말한다. empty words와 완료 prefix가 없다. 두 Next comment는 word/token stream으로 반환 형태가 다르고 오류 중간 토큰 정책도 생략한다.

**뒤의 full code나 Got으로 답을 설명할 수 있어도 현재 request+caption만 보고 그 조건을 신뢰성 있게 구분할 수 있다는 뜻은 아니다.** 새 데이터 수만 늘려도 빠진 caption 조건은 생기지 않는다. 이 draft를 올바르게 보이는 답 caption으로 덮지 않는다. 필요하면 별도 버전의 source-derived finite-behavior caption을 작성하고 AI-assisted 저작 경로, fidelity/coverage QA와 사전 고정 ablation으로 비교해야 한다. v2와 Want 기준은 보존한다.

## 24개 관찰 계획에서 3개 문제로 넘어가는 조건

24개는 datasize10/query6/shlex8의 **finite probe 수**이고, 목표와 관련 원천 family는3개다. 이번 설계에서 새 native 관측 결과는 아직 사용하지 않았다. 기존 독립 검토는 frozen Want가 source와 일치한다는 예상 검토이며 실제 Got이 아니다.

- **양성:** native 결과를 고정 source/Want/observer에 결속한 뒤, 요청의 범위를 명시하고 모든 후보의 원문·helper·caption fidelity와 해당 조건을 검토해야 한다. 직접 호출하지 않은 Parse/MustParse/Next/interface/helper를 같은 행동이라고 자동 승격하지 않는다. 필요한 새로운 관찰은 별도 contract를 먼저 고정한다. 만족 후보가 여러 개면 하나만 정답으로 강제하지 않는다.
- **무응답(no_answer):** 제공된 모든 후보가 그 supported contract를 만족하지 않는다는 근거가 있어야 한다. 짧은 caption에 정보가 없다는 이유만으로 모두 false 처리하지 않는다.
- **unknown:** native 차이/오류 의미 미확정, caption 누락, 후보 scope 미확정이 남으면 유지한다. 현재3 draft는 전부 여기에 해당하며 실제 label은0개다. 일부 probe가 맞았다고 broad truth·candidate 만족·training-ready로 승격하지 않는다.

## 기존76과 함께 쓸 때의 family/누출 경계

UnmarshalText와 Parse/MustParse wrappers는 한 family, Values와 reflectValue/valueString/tag helpers는 한 family, Split와 Next/scanStream/classifier는 한 family다. source alias·receiver/typed adapter·helper·상수/sentinel·파라미터 인스턴스·derived caption 계보를 기존76의 저장된 원문 refs에 대조하고 공유 연결이 있으면 함께 묶어야 한다. 현재 whole-group 숫자 배정·전체 graph union·role 실행은 하지 않았다. 기존76의 역할을 새 membership에 복사하지 않고, 이후 전체 membership을 동결한 다음 같은 family가 train/validation 양쪽에 섞이지 않게 계획한다. 기존76 및 semver/glob 이력은 수정하지 않는다.

KR/EN·정상/오류 변형은 같은 parent/wholefamily다. **3을6으로 늘리지 않는다.** 같은 표준 라이브러리나 Google 조직명만으로 관계없는 source를 한 그룹으로 합치지도 않는다. 반면 동일한 요청 작성·caption 선택·observer·검토자 흐름은 상관된 저작 provenance로 기록하며 독립성/human-only 증명이 아니다.

MIT/BSD-3-Clause/Apache-2.0 원문과 shlex copyright header는 `licenses/`에 byte 그대로 포함했다. attribution과 source URL/SHA를 JSON에 연결했으며 새 학습 자산의 권리·qualification 승인을 대신하지 않는다. 생성 command3회 중 최초 unused-import compile 실패1을 보존했고 성공 metadata 생성은2회(v1와 선언순서 v2)다. 원본 API/init/test/observer·모델/학습/feature/role/실제labels·보호final·공유 수정·GitHub/HF0이다. AI-assisted 작업 비용은 측정하지 않았다.
