# 공유 은닉층 16개 claim MLP 소스 프로토타입

순수 Go, float32 `2048 contextual 희소 입력 → 공유 ReLU 16개 → 독립적인
3상태 범주형 헤드 3개` 모델이다. `statehintwide.ExtractContextual`을 그대로
사용하고 head/state/source/prediction 타입은 `statehintclaims`의 별칭이다.
각 헤드는 별도의 3방향 softmax를 쓰며 누산·손실·확률은 float64이다.
생성형 출력이나 런타임 등록은 없다. 출력은 발화에 귀속된 claim 힌트이고,
`Metadata`의 `qualified`와 `state_authority`는 항상 false이다.

파라미터는 float32 32,937개(가중치 32,912개, 편향 25개)이다.
`Predict`/`Scores`는 모델을 읽기만 하며 호출자 소유 `Workspace`가 필요하다.
동시 호출은 각각 별도 workspace를 쓴다. `Clone`과 `Parameters`는 값 복사라
원본을 바꿀 수 없다. 모델에는 가변 캐시·map·lock이 없고, 재사용한 workspace의
Predict 할당은 0이다. 기존 .9 confidence/.05 margin/T1 및 unknown reason
우선순위(no_word_content, untrained, semantic_unknown, low_confidence,
low_margin)를 유지한다. 이 숫자 기준은 품질 자격이나 상태 권한을 주지 않는다.

`Fit(samples, trainingWorkspace)`는 40 epoch, batch 32, AdamW 학습률 .001,
가중치 감쇠 .01, seed 1729, hard target, 세 헤드 평균 CE로 고정되어 있다.
두 가중치 블록에만 감쇠를 적용하고 두 편향 블록은 감쇠하지 않는다.
ReLU의 0에서 미분은 0이다. 매 Fit마다 가중치·Adam 두 moment·step을 새로
시작한다. hyperparameter 옵션, smoothing, 온도 조정, warm-start, text/feature
corpus 보관은 없다. 모든 UTF-8·길이·word content·target을 갱신 전에 검증하고
오류 시 receiver를 보존한다. 1,680개를 공급하면 `40 × ceil(1680/32) = 2,120`
갱신이다. 이는 산술 확인이며 실제 데이터 학습 결과가 아니다.

초기화 PCG는 `(1729, 1729 XOR 0x696e697472636d31)`이며 입력 feature/hidden,
이후 hidden/head/state 순서로 `float32((2*u-1)*limit)`를 뽑는다. 입력 limit은
`sqrt(6/(2048+16))`, 각 독립 출력 헤드 limit은 `sqrt(6/(16+3))`이다.
편향은 0이다. shuffle PCG `(1729, 1729 XOR 0x9e3779b97f4a7c15)`는 기존 cold
linear claim recipe와 같다. linear는 0 초기화이고 이 MLP는 Glorot 초기화이므로
초기 모델이 동일한 비교는 아니다. 데이터와 갱신 recipe를 맞춘 비교이다.
기존 8방향 MLP의 .02/.001 기본값이나 가중치를 상속하지 않는다.

고유 `RCM\0` v1은 little-endian header 192바이트 + float32 payload
131,748바이트 + SHA-256 32바이트 = 131,972바이트다. header는 차원,
ReLU/초기화 코드, seed/step, feature-schema hash, head/state 순서 hash,
수치 계약 hash, .9/.05/T1, 학습 sample 수, parameter/weight 수와 0인 예약
영역을 고정한다. payload 순서는 입력·은닉 편향·출력·출력 편향이다.
Load는 정확한 길이·checksum·유한값·seed·sample/step 관계를 검증한다.
미학습은 sample과 step 모두 0이어야 한다. 악의적인 차원과 count는 할당이나
반복 범위를 결정하지 않는다. checksum은 게시자 인증이 아니다. optimizer,
corpus와 부모 모델 가중치는 저장하지 않는다.

직접 만든 합성 테스트만 수치·직렬화·동시성·할당을 검증한다. 실제 corpus를
학습하거나 조회하지 않았고, 품질·속도·저장 이점·보정 성능을 주장하지 않는다.
ternary/BitNet adapter는 이 패키지 범위 밖이다.
