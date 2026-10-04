# Next hypothesis / 다음 가설 — 원문 토큰 준비 공유

현재 변경은 누적 할당 바이트를 약37% 줄였지만 쌍별 시간 중앙값이 약7.5–7.7% 늘었고 Mallocs가288회 증가했습니다. 다음 실험은 생성자 안에서 요청1개와 후보최대8개의 원문 words/terms를 한 번만 준비해, cardinality와 실제 특징 생성이 공유하도록 하는 것입니다. 아직 구현·실행하지 않았습니다.

검증을 먼저 유지합니다. `ValidatePrepared` 후 고정9슬롯의 생성자 전용 자료를 만들고, 요청의 순서별 FNV prefix도 한 번 계산합니다. terms 생성 후 overlap용 words를 정렬·중복 제거할 수 있지만, cross용 terms의 원래 순서와 반복은 유지해야 합니다. 최종 반환 객체에는 이 임시 자료를 남기지 않습니다. 기존 sort.Slice·중복 누적·상쇄된0·마지막overlap·텍스트clone·offsets·오류 우선순위·Rank를 그대로 유지합니다.

검증용 `NormalizedRequest`를 특징 생성에 대입하지 않습니다. NormalizeText는 camelCase 분리 등을 하지만 기존 words는 다른 규칙을 사용합니다. 실제 raw token/term 수로 용량을 계산하고 overflow 검사를 유지합니다.

다음 실행 전에 별도 코드·입력·가중치·예산을 다시 동결할 제안 기준:

- 모든 feature·score 비트·순위와 입력/소유권 계약 일치, 최종 tight 저장량 동일.
- 현재 scratch 구현 대비 N=1/8 각각6쌍의 모든 구간에서 할당 바이트가 증가하지 않고 Mallocs가 최소144회 감소.
- 각 N의 쌍별 새/현재 시간 비율 중앙값 ≤1.00. 준비·토큰화·스캔 비용 전체 포함.

This unexecuted proposal targets the observed tokenization/allocation tradeoff. Build constructor-local immutable raw words/terms once for one query and up to eight candidates, and reuse ordered query hash prefixes. Both passes still pay cross hashing; speed improvement is a hypothesis. Preserve exact original tokenizer and accumulation semantics. The proposed gates above need a separate freeze and actual comparison before adoption.

새 독립 부모·학습·모델 품질 개선 계획으로 세지 않습니다. 현재 정답 후보5위의 비활성 모델, 불확실한 source-family 편입, 독립2400 평가와 실제 에이전트 비용 검증은 별도 과제입니다. This does not qualify source families, add Golden parents, train weights or activate the failed model.
