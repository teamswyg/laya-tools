# 다음 1회: BCE에 요청 내부 상대 순서 손실 추가

첫 FP32 fit은 실행에 성공했지만 효용은 실패했다. Validation BCE 0.6658, 가장 강한 lexical_ordered 27 checks/Top1 5, learned 31 checks/Top1 2, Top3 둘 다10이다. 검증 요청15개(답 있음10/없음5)의45후보를 모두 평가했다. Lifecycle-acquire의 세 요청에서 checks5→9가 되었고 fully-eligible 진단B도26→30으로 나빠졌다. 한 개의 loss mask만으로 실패를 설명할 수 없다. 원래 결과71849B SHA060eaa602f27f7fd9a67306e2fa2a530f0ef4c5c4548858a7818d9dfdb4bb28a는 수정하지 않았다.

BCE는 후보별0/1 예측을 학습하고, 실제 효용은 같은 요청의 만족 후보를 먼저 찾는 것이다. BCE 감소가 순서 개선을 보장하지는 않지만 BCE로 순위를 배우는 것이 불가능하다는 뜻도 아니다. 단어/bigram 해시가 인과순서를 충분히 표현하는지, 적은 자료로 배울 수 있는지도 미확인이다. 기존91행/batch128은 epoch당1회, 총50회 갱신이었다. 손실 감소만으로 수렴을 단정하지 않는다.

대안A는 비학습 비교군 유지(가장 작으나 학습 가설 미검증), B는 상대 순서 손실 하나 추가(선택), C는 의미 특징·새 원천 자료 확대(다음 단일 실험에는 여러 요인을 바꿈)다. B도 lifecycle 인과관계를 자동으로 습득한다는 보장은 없다.

## 구현된 선택 API

`pairlearn.FitWithRanking` 또는 `FitWithRankingTrace`에 `RankingConfig{Schema: RankingSchema, Lambda: 1, TrainingParentIDs: ..., ValidationParentIDs: ...}`를 전달한다. ParentIDs는 원래 RowRefs의 numeric parent index를 row별로 연결하는 supervision metadata이고 특징에는 들어가지 않는다. nil은 legacy API로 직행한다. λ0는 유효 association을 확인하고 일반 BCE 가중치도 보존하며 pair plan/진단을 계산하지 않는다. λ1은 nil/unit 또는1/0 가중치만 허용한다. 잘못된 schema·association은 오류이고 조용한 fallback이 아니다.

원래 BCE와 모든 물리적 행/라벨/가중치는 그대로 둔다. 같은 parent의 기존 positive×negative 쌍만 추가한다. Endpoint weight의 곱이0인 쌍도 감사 배열에 보존하며 기울기는0이다. 나머지는 parent 안에서 평균한 후 eligible parent를 평균한다. 추가 loss는 `log(1+exp(-(s_positive-s_negative)))`, 계수는1로 고정한다. 이 계수는 최적값이 아니라 다음1회를 위한 가설이다. No_answer는 원래 전부0인 BCE를 유지하며 가짜 긍정이나 비교 쌍을 만들지 않는다.

배치의 positive endpoint가 그 쌍의 gradient를 소유한다. 같은 FP32 scorer 계수로 양 끝을 점수화하고 pair contribution에 전체 행수N을 곱한 뒤 기존 batch 크기b로 나눈다. 고정 계수에서 균등한 size-b 표본의 기대값은 전체 pair gradient와 같다. 기존 random reshuffling은 배치 사이에 계수를 갱신하므로 이전 갱신에 조건부인 독립·불편 표집까지 주장하지 않는다. Negative가 현재 배치 밖에 있어도 양 끝의 sparse features에 반대 gradient가 적용된다.

학습/검증 source-group 중복 및 parent의 cross-role/cross-source-group 연결을 거절한다. Parent당 최대8행, 새 association/pair/diagnostic backing arrays의 보수적64MiB cap를 사전 검사한다. 이는 전체 heap/RSS cap가 아니다. 새 경로는 SoA 배열을 소유하고 map/lock을 사용하지 않는다. Dataset 자체의 큰 특징 배열은 복사하지 않으며 caller가 호출 동안 입력을 불변으로 유지한다.

호출자가 frozen truth/role/source clearance를 확인해야 한다. 이 numeric association API는 unknown이나 출처 적격성을 증명할 수 없다. Unknown은 null 감사에만, calibration은 fit/epoch선택/효용에서 제외하는 기존 driver 책임을 유지한다. 원래 BCE Trace와 validation-BCE earliest-min epoch selector는 바꾸지 않았다. 별도 RankingTrace는 pair loss와 mask0 쌍을 기록하고 모델 계수는 담지 않는다. Audit의 Weights는 supervision weight다.

## 다음 실행 한계

다음 실제 fit은 CI 뒤 별도 입력/소스/바이너리 동결 후1회다. 기존 seed1729/FP328192/50epoch/batch128/LR0.1/L2.0001, 원래91/45행, validation BCE 선택, 모든45후보의A효용, 기존 control tie order 및5%/Top1/Top3 조건을 유지한다. Seed/mask/λ/threshold 탐색은 하지 않고 기존8fits/16assets ceiling 안에서 진행한다. 이번 구현에서는 실제76 Fit/Features/Project/Baselines, 새 실제 모델 asset, HF 게시, runtime/CLI default 활성화 모두0이다. 합성 테스트는 학습기와 특징 추출을 실제로 호출하므로 전체 모델 연산0이라는 의미가 아니다.

같은 validation 실패를 보고 선택한 후속안이다. 개선돼도 개발 신호이고 fresh-domain 최소2400개의 구분되는 요청을 봉인한 최종 평가 요구를 유지한다. Pair·반복·바꿔쓰기를 독립 요청으로 세지 않으며 연결된 source/author/template 그룹을 split 사이에 섞지 않는다. 새 upstream 두 가족이 모두 train이라는 한계도 유지한다. 3진 압축·규모 확대는 효용 근거가 생길 때까지 미룬다. 실제 LLM 호출 절감이나 production 권장 근거가 아니다.

[RankNet 원문](https://www.microsoft.com/en-us/research/wp-content/uploads/2005/08/icml_ranking.pdf)은 같은 요청 후보의 점수 차이에 logistic loss를 적용한다. [LambdaRank 설명](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/MSR-TR-2010-82.pdf)은 pairwise error 감소와 상위 결과 효용이 다를 수 있음을 설명한다. 이번은 plain pairwise 보조항이며 first-positive checks를 직접 최적화하는 LambdaRank가 아니다. [선택 편향 연구](https://jmlr.org/papers/v11/cawley10a.html)는 제한된 평가 자료에 선택 기준을 반복 최적화하는 위험을 뒷받침한다.
