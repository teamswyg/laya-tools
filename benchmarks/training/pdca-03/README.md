# PDCA-03 / lower learning rate and expanded development training

PDCA-02 failed: both seeds downgraded difficult work, and confidence-gated
coverage was insufficient. Its new test and PDCA-01's old test are now explicitly
training data. This cycle uses **132 train / 24 validation / 12 calibration /
24 new test** cases. All final families are new. Prior calibration is reused;
this is still adaptive research with the same author and rubric, not an
independent real-repository benchmark.

Reduce learning rate to 3e-5 for ten epochs, retain three-order loss averaging,
and select by six-order validation NLL. Freeze both seeds before the final check.
The acceptance criteria are identical to PDCA-01/02. No threshold reduction.

재현은 [PDCA-02 절차](../pdca-02/README.md)의 데이터/출력 경로를
pdca-03으로 바꿉니다. 이전 평가 사례는 학습용으로 전환했음을 명시하고,
새로운 최종 평가를 사용합니다. 영어 합성 난이도 분류의 제한된 실험이며
실제 저장소 성공률, 비용 절감, 한국어, 분할 모델 개선을 뜻하지 않습니다.
