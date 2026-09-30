# PDCA-01: family-separated difficulty experiment

[한국어] 초기 난이도 판단 개선을 검증하는 **자체 작성 영어 평가**입니다. 실제 코딩 모델의 성공/비용, 독립 외부 저장소 작업 성능을 측정한 자료가 아닙니다. 원본 Laya·기존 판단 head와 관련된 라이선스는 유지하며 새 텍스트는 저장소 Apache-2.0으로 배포합니다. 비공개 코드·고객 요청·외부 데이터셋은 포함하지 않습니다.

Each family has three authored capability labels in `cases`: fast, standard, strong. Only case text enters the model. Family IDs and labels never enter its input. Families do not cross train/validation/calibration/test. Conceptual skills and stylistic cues can still recur across families: family-ID separation is not proof of independent real-world generalization or label correctness.

| Partition | New families | New cases | Additional legacy cases |
|---|---:|---:|---:|
| Train | 12 | 36 | 24 from pilot-v1 train only |
| Validation | 4 | 12 | 0 |
| Calibration | 4 | 12 | 0 |
| Final authored evaluation | 8 | 24 | 0 |

The existing 36 routing goldens stay out of training; their 32 English rows are an additional regression check. Their four Korean rows retain the existing runtime abstention policy and are not claimed as supported.

`plan.json` fixes learning rates, epochs, two seeds and acceptance before training. Training shuffles choice positions; target indices follow the actual position. Validation NLL alone chooses checkpoints. Candidate hashes are saved before a separate `check` command opens the final evaluation. Baseline and candidate temperatures use the same calibration set and grid for a fair probability comparison. Final evaluation repeats every alternative choice ordering and checks the legacy corpus.

**All registered conditions must pass for both seeds.** A first improvement means evidence on this limited authored evaluation only. It does not authorize default model replacement or downstream automation. Small accepted subsets have wide uncertainty even at 100% observed precision. If a cycle fails, retain results, do not lower gates, and use a new held-out family set for another cycle. Published cases eventually become regression data.

계획의 기준을 통과해도 자동 실행·기본 모델 승격은 별도 검증입니다. 테스트 정답과 실행 결과는 학습 입력에 넣지 않습니다. 같은 레이블의 전형적인 표현을 학습했을 가능성, 작은 표본과 주관적인 난이도 기준의 한계를 함께 기록합니다.
