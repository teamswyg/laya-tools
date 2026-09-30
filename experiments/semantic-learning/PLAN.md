# Encoder-free learning cycle 01 / 인코더 없는 학습 01

Freeze this plan and file hashes before fitting. No external data, weights or private text.

가설: 단어 순서 특징과 요청/문서의 단어 조합을 학습하는 작은 선형 관련성 모델이, 동일한 동사 규칙을 새로운 도메인 명사에 적용할 때 BM25보다 검증 횟수를 줄인다. 이 실험은 새로운 동작·새로운 문법·한국어·실제 코드에 대한 일반화를 증명하지 않는다.

Hypothesis: a small linear ranker over hashed query/document word and bigram cross-features can transfer a shared retain/discard rule to held-out domain nouns, reducing inspections versus BM25. This does not test unseen operations, new grammar, Korean or real code.

32 original fictional domains: 20 train, six validation, six final. Each has two opposing documents and four queries. Documents contain the same words in different order, exposing the limitation of bag-of-words scoring. Queries use retain/discard while documents use keeps/removes. Final documents never enter training negatives or validation. This deliberately narrow, template-generated task is a controlled learnability probe; not an independent natural-language benchmark. Separate families share grammar and verbs intentionally.

- 8192 hashed coefficients, sparse cross-features plus token overlap; no encoder or pretrained vocabulary.
- Train only train rows against the 40 train documents. Validation uses its own 12 documents; final its own 12. Candidate-count differences must not be interpreted as equivalent split difficulty.
- Two seeds × two learning rates × FP32/ternary straight-through modes = eight candidates. 60 epochs. Select epoch/config using validation mean target rank then NLL. Save selection and models before reading final query content. Record per-seed results; do not select on final.
- Ternary forward uses {-scale,0,+scale} at threshold 0.7×mean absolute shadow weight; identity STE backward, scale detached. FP shadow training remains FP. This is a full encoder-free linear scorer, not ternary Laya or an LLM.
- Compare BM25, FP32, row INT8 PTQ, ternary PTQ and ternary STE where selected. Record per-case ranks, NLL, pure ranking and bounded interleave costs. No probability calibration claim.
- Provisional final gates: recall@4 ≥95%, mean checks ≤80% BM25. Keep failures; do not change gates after inspecting final. Even passing only advances this probe, not the overall product goal.
- Persist public code/data/plans/results in GitHub, weights locally pending strict HF publication validation. No model binaries in GitHub. Training and deployment are Go-only for this path. GPU concurrency is unnecessary for this small sparse learner.
