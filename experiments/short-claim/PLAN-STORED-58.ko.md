# 저장 정답 기반 비학습 효용58 계획

[English](PLAN-STORED-58.en.md) · [이전 그룹 감사](RESULTS-56b.ko.md) · [공개 동작57](../public-behavior/RESULTS-57.ko.md)

**관측 전 개발 진단 계획이다. 새로운 정답·독립 요청·역할·학습·모델·가중치·유료 호출·보호 최종 접근0이다.** 기존72부모의 저장 정답으로 순서를 바꿀 여지를 확인한다. 후보 코드를 다시 실행하거나57의62동작을 학습 라벨에 합치지 않는다. 기존 사용자 승인 범위에서 진행하며 추가 사람 승인 절차를 만들지 않는다.

## 확인할 질문

짧은 힌트 모델을 학습하기 전에, 좋은 순서가 독립 확인 횟수를 줄일 수 있는지 확인한다. 요청과 후보 설명만 사용하는 네 가지 고정 기준을 비교하고, 정답을 미리 안다고 가정한 상한을 계산한다. 이 상한은 모델이 달성한 효용·실제 후보 실행·CPU/RSS/시간·LLM 토큰·요금 절감이 아니다.

| 봉인 입력 | SHA256 |
|---|---|
| [기존48부모](probes-56.json) | `0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0` |
| [추가24부모](probes-56b.json) | `0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e` |
| [기존 저장 정답·그룹](results-56b.json) | `4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4` |

72부모/216후보 설명은 답 있음34·답 없음17·판단 보류21이다. 후보2개인 요청12·3개48·4개12를 모두 유지한다. 전체17/라벨16연결 그룹은15운영 하한을 이미 통과했다.15개 repository 조건을 추가하지 않는다. 같은 합성 영어 작성 흐름의 한계와 기존 unknown/pending을 유지한다.56a에서 이미 본48개 순위는 새로운72개 진단과 구분한다.

## 입력 연결과 특징 경계

Go `storedaudit.Bind`는 정확한3raw SHA/바이트를 먼저 확인하고 부모·후보 ID/순서/수·허용 인덱스·상태·벡터 기록·그룹 membership을 연결한다. raw input SHA와 canonical typed-dataset SHA는 다르며 혼동하지 않는다. 중복/범위 밖·모순·불완전 입력은 고정 오류로 거절한다. unknown을 false나 no-answer로 바꾸지 않는다.

`Text`는 원문 request와 순서가 있는 candidate 설명만 가진다. 정답·source/prototype/contract/group/role·문서 파일명은 별도 metadata다. scorer에는 predictable local candidate ID와 고정 provenance만 전달하며 이 값들도 특징으로 사용하지 않는다. 원문512바이트/유효UTF-8와 기존 정규화32단어 계약을 유지하고 자르거나 다시 쓰지 않는다. 준비 테스트의 저장 연결과 literal 순열 검사는 새 source truth·실제72개 순위 관측이 아니다.

## 비용과 대조군

네 기준 순서는 `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule`이다. 좁은 문법이 지원하지 않는 일반 문장은 기존 BM25로 복귀하고 fallback 분모는unknown을 포함한72다. 요청을 관측 뒤 `behavior-v1:` 문법으로 바꾸지 않는다. 후보를 삭제하지 않고 동률은 원래 순서다.

- 답 있음: 첫 허용 후보까지 확인하는 횟수. 허용 집합이 여러 개일 수 있다. oracle도1회다.
- 답 없음: 기준과oracle 모두 해당 후보 전체를 확인한다.
- unknown21: 원래72 분모·관계·이유에 남기되 비용/Top1/Top3에서 제외한다. 제외된0은 실제 확인0의 증거가 아니다.
- known51와answerable34의 비용을 별도로 보고하고, 작은2~4후보에서Top3가 쉽게 포화되는 한계를 유지한다.
- 개별 부모와그룹별 비용, known이 있는16그룹의macro를 추가 진단으로 보고한다. macro는global5% 기준을 대체하지 않는다.

**최선 기준은known51 합계가 가장 작은 하나의 고정 기준**이며 부모마다 유리한 기준을 고르는 합성 정책이 아니다. 동률은 위 배열 순서다. 필요 여지는 `(best_known_checks-oracle_known_checks)/best_known_checks`이고 정확한 정수 비교로5%이상을 확인한다.5%미만/이미최적이면 이 자료의 학습을 중단한다. 통과하더라도 학습 가능한 모델·학습 적격을 인증하지 않는다.

## 실행 봉인과 제한

실제 source freeze의9개 non-test Go 소스와13개 support,3입력/실행계획·binary SHA/바이트·Go1.27.1/CGO0/플랫폼을 고정한다. source/input freeze는 별도40hex Git commit이며 각 실제 blob을 실행 전에 대조한다. 추가 Go 소스·unknown/duplicate/extra-array/noncanonical 계획·다른binary를 거절한다.

공식 진단은1회·재시도0이다. new directory와exclusive 결과 파일은 실행 전에 예약한다. 이 경로 제한은global one-shot 보장이 아니며 외부 공식 원장이 필요하다. dispatch/검증/시작한Baselines 호출/완료 control ranking/완료 row를 구분한다. 실패·부분 결과와 호출 후 저장 실패를 숨기지 않는다. 한 request의Baselines 호출은 네 control ranking을 반환하므로 성공 목표는72호출/288ranking/72row이며 새 독립 표본이 아니다.

각 입력·소스·결과는1MiB, binary는64MiB를 상한으로 둔다. Git blob은파일별5초와bounded stdout이다. 평가10초 cooperative deadline은bounded row 사이 및 완료 시 검사하고 detached worker를 만들지 않는다. GOMAXPROCS1·Go heap256MiB soft limit은 설정으로서RSS/CPU/GPU/지연 실측이 아니다. 새 resource/performance 주장이나 캐시·SIMD·lock 최적화는 이번 범위에 포함하지 않는다.

공식 관측 전에 CI 재현 기준도 고정한다. 순위, 정수 확인 수, 상태, 이유, 분모와 나머지 평가 기록은 정확히 일치해야 한다. Linux/macOS의 부동소수점 반올림 차이에 한해 비검증 점수는 양쪽 모두 유한하고 `abs(actual-saved) <= 1e-12*(1+abs(saved))`여야 한다. 점수를 비교한 뒤 그 필드만 제거해 나머지를 정확 비교한다. 원본 결과 파일의 SHA는 별도로 고정하며 점수 허용치를 보고서 수정이나 순위 차이 허용에 사용하지 않는다. CI 반복은 공식 진단 횟수나 독립 요청 수를 늘리지 않는다.

## 다음 판단

`no_roles_plan`과`synthetic_single_pipeline`을 보존한다. 필요 효용 뒤에도 문장 충실성·행동/언어 coverage·작성 흐름·공개 원천 전이 관계를 검토하고, 연결 요소 전체의train/validation/calibration membership과 별도 학습 실행 계획을 봉인해야 한다. 역할이 유리하도록 결과를 보고 그룹을 고르지 않는다.58은역할0·fit0·training_ready=false이며production 활성화를 승인하지 않는다.

도메인별2400개의 고유 보호 최종 요청은 추후 최종 효용·일반화 주장용 별도 목표다. 모든development fit의 선행 최소 수로 재해석하지 않는다. 변형·반전·번역·리터럴·반복·seed/모델 수로2400을 채우지 않고 기존 protected final/CoSQA reserve를 읽거나 재배정하지 않는다. FP32 주 후보·INT8/PTQ child·별도 ternary STE sibling은 이후 별도 자격을 갖춘 비교로 유지한다.

새 모델 본체는Git에 넣지 않는다. 이번에는 자체 공개 Apache-2.0 자료를 재사용하고 LICENSE를 봉인한다. 기존56a/b/c/d/e/57 결과는 고치지 않는다. Go 평가기와계획을먼저 봉인하고결과는별도문서와CI검사로추가하며, CI 통과·자동 병합을확인한뒤Wiki/이슈에근거를연결한다.
