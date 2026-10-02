# 두 번째 학습 실패에서 다음으로 할 일

**다음은 새 개발 데이터의 의미·표현 다양성을 늘리는 것을 우선 추천한다.** λ·seed·threshold를 다시 고르거나 더 작은 3진 모델로 압축할 근거는 아직 없다. 현재 모델은 후보를 검증할 순서에 힌트를 주는 주장 모델이며, 정답·실행·승인을 결정하는 모델이 아니다. 이번 두 모델 모두 기본 활성화나 효과 주장에 적합하지 않다.

| 동일한 validation15 요청·45 후보 | 강한 lexical control | 첫 BCE | BCE+pair λ1 |
|---|---:|---:|---:|
| 확인해야 한 후보 수 | 27 | 31 | 33 |
| 첫 후보가 acceptable인 요청 | 5/10 | 2/10 | 1/10 |
| 3위 안에 acceptable | 10/10 | 10/10 | 10/10 |
| validation BCE | 해당 없음 | 0.66580 | 0.68855 |

no_answer5개는 어떤 순서에서도11 checks가 필요하다. 나머지 answerable10개의 비용은 control16→첫 모델20→두 번째22다. Top3 유지가 Top1이나 비용 개선을 뜻하지 않는다. 두 번째는 부모5·8·37에서 비용이 각각1 늘고, 부모60에서1 줄어 총2가 악화됐다. typed9→8 개선은 있었지만 legacy22→25 악화가 더 컸다. 같은 후보·truth·mask·role·control이고 새 추출도 없으므로 이번 비교는 한 가지 objective 변경의 개발 결과다.

훈련 pair NLL은0.69246→0.59361로 내려갔다. 검증은0.69441→0.69273으로, 두 후보를 구별하지 못하는 margin0의 기준 `log(2)=0.69315` 근처다. 저장된26개 nonzero 검증 pair 중 positive가 앞선 것은10개이며, 부모 평균 weight 기준40%다. 평균 margin은 약0.00091이다. 즉 훈련 관계는 학습했지만 네 검증 source group으로 전달된 효과가 작다. 이것만으로 과적합·수렴·구현 버그 중 하나를 확정할 수는 없다. 두 실행 모두50 epoch=50번 업데이트의 상한에 도달했고, 독립 합성 검증은 loss 수식과 nil/default 경로 보존을 확인했다.

**이번 실패를 epoch 선택만으로 설명하기는 어렵다.** validation BCE의 최소도, pair NLL의 최소도 epoch50이었다. Pair loss가 실제 비용과 완전히 같지는 않다: pair loss는 모든 positive가 모든 negative보다 높기를 요구하지만, 실제 비용은 acceptable 하나를 일찍 찾으면 줄어든다. no_answer의 순서는 비용을 바꾸지 않지만 BCE에는 원래 negative supervision으로 남는다. 이런 정합성 한계는 있지만, 이번 결과는 selector만 바꾸면 개선된다는 증거를 주지 않는다.

저장된 배열을 읽은 작은 통계에서 pair gradient에 기여하는 훈련 부모는16개뿐이었다(전체 known train31, 원래 zero weights 보존). 검증은10개다. 요청 고유 단어는 훈련166개/검증82개이고, 검증50개(약61%)는 훈련 요청에 없었다. caption은 검증100개 중67개가 훈련 caption에 없었다. 이 비율은 고유 vocabulary 기준이며 전체 token 빈도 비율은 아니다. 가족을 묶어 분리한 결과 `verified`, `inclusive`, `limit`, `release` 같은 검증 의미가 훈련에서 직접 보이지 않는다. 두 새 upstream 가족도 모두 train이어서 이 run에는 새 upstream 일반화 관측이 없다.

Feature는 encoder 없는 unigram·인접 bigram의 query×caption 곱을8192개 signed hash bin에 넣고, exact token recall 하나를 추가하는 선형 모델이다. 단어·인접 순서는 보지만 괄호/논리 범위, 시간 순서의 의미, 동의어를 명시적으로 다루지 않는다. tokenizer는 `<`, `>` 같은 구두점을 버리므로 문구가 같고 연산자만 다른 입력은 구별 정보를 잃을 수 있다. 다만 현재 실패가 특정 해시 충돌 때문이라고 단정할 수 없다. nonzero-weight validation의6107개 bin 중490개만 train에서 없었고(약8%), 그 L2 mass는5.87%였다. bin이 겹쳐도 같은 의미의 word-pair가 훈련됐다는 뜻은 아니다. 압축·충돌·의미 부족을 분리해서 실험해야 한다.

선택지는 세 가지다.

1. **개발 데이터 확장 — 추천.** 새로운 공개·라이선스 확인된 원문에서 조건 범위, 경계 포함, before/after, negation/exception을 구별하는 대조 후보와 다양한 표현을 수집한다. 단순 템플릿 복제·literal 수를 독립 요청 수로 세지 않는다. source/author/template 가족별로 분리하고, 불명확한 truth는 unknown으로 둔다. 다음 데이터 실험에서는 feature·objective 중 어느 하나도 동시에 바꾸지 않아 원인을 구분한다.
2. **작은 feature 개선 — 다음 순위.** raw request/caption만으로 연산자·부정·시간 관계를 보존하는 적은 수의 명시적 채널을 별도 ablation한다. source ID·role·old correct·acceptable은 feature에 넣지 않는다. 빠른 lexical control을 유지한 opt-in 힌트 실험이며, runtime이나 모델 기본값을 지금 바꾸는 작업은 아니다.
3. **utility 목적 재설계 — 이후 검토.** first-acceptable 비용에 더 가까운 surrogate/epoch 선택을 별도로 사전 고정한다. 이번 validation을 보고 유리한 목표를 골라 같은 validation으로 승리를 선언하지 않는다. 이미 pair λ1이 실패했으므로 다른 objective를 계속 돌리는 선택은 후순위다.

기존76개는 개발 회귀로 보존한다. **별도 ≥2400개 fresh-domain final golden set**을 새로 확보·동결하고, 모델/목표 선택에 소비하지 않아야 효과를 주장할 수 있다. unknown21, no_answer, zero-weight known rows, calibration no-fit 및 모든 후보 보존 정책은 유지한다. 큰 GPU나 압축보다 검증 비용이 실제로 줄어드는지가 우선이다.

근거: 첫 결과 SHA `060eaa602f27f7fd9a67306e2fa2a530f0ef4c5c4548858a7818d9dfdb4bb28a`; 두 번째 `e01d7daf8a5f26cd5578674e87f996a578d747f6622221308da4624a71995220`; cached projection `f68bff5f48747a66f038c17262568c1bd91251a8c81b3f46f7ae5090c8fe4b99`. Source는 CI95 `d506f58ccf9629e2d2b6ca3cd1766ab20789bf64`의 hintlearn/learn.go, pairlearn/learn.go·ranking.go, 원래 utility.go를 읽었다. [작은 saved 통계](SAVED-STATS.json)는 stdlib-only1회 성공이다. jq 출력식 typo1회는 수정했으며 데이터/API 실패가 아니다. 이번 분석의 새 Fit/Features/Project/역할/라벨/추론/benchmark/HF는0이고, AI-assisted 협업 비용은 측정하지 않았다. 이는 backend/runner 작성자의 비맹검 개발 분석이며 독립 효과 승인이나 새 gate가 아니다.
