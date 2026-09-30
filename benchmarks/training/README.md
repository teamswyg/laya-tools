# Original synthetic development pilot / 자체 작성 합성 개발 사례

`pilot-v1.tsv` is newly authored project material distributed under this repository's Apache-2.0 license. It contains no external dataset, private repository content or customer requests. The labels were authored from the documented rubric, not obtained by running coding models. There are 48 difficulty and 40 candidate-plan examples. Label words are separate columns and never included in the model input.

These small synthetic partitions are **development data**, not independent repository-level goldens or production qualification. Related conceptual patterns appear across partitions; exact-text duplicate checks do not establish family independence. The first round's evaluation was observed before extending training, so subsequent results cannot be called an untouched final test. Keep all losses, regressions and abstentions. Do not silently promote these examples to the planned 400-group collection.

| Task | Train | Validation | Calibration | Development test |
|---|---:|---:|---:|---:|
| Difficulty | 24 | 6 | 9 | 9 |
| Decomposition | 16 | 8 | 8 | 8 |

Learning rates and epoch checkpoints are selected using validation NLL. Calibration uses a separate partition and temperature range 0.5–5.0. Reported baseline probabilities use upstream calibration; tuned probabilities use pilot calibration, so calibration gains are not attributable only to learned weights. Confidence gate remains 0.9. None of these labels proves which actual generative model can complete the request.

자체 작성한 영어 합성 개발 자료 88개이며 외부 데이터·비공개 코드·고객 요청을 포함하지 않습니다. 저장소 Apache-2.0으로 제공합니다. 실제 코딩 모델 실행으로 정답을 얻은 자료가 아닙니다. 작은 분할들 사이에 유사한 개념 패턴이 존재하므로 독립 작업군 평가나 운영 성능의 증거가 아닙니다. 첫 결과를 본 뒤 학습 길이를 늘렸으므로 이후 test도 개발 평가로만 취급합니다. 계획한 400그룹 골든셋을 완성했다는 의미가 아닙니다.
