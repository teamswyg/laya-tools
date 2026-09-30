# PDCA-02 / 순서 편향 보완

[사전 고정 계획](plan.json)은 PDCA-01의 판정 기준을 유지합니다.
이전 final 24건은 이제 validation입니다. 새로운 8가족 24건만 이번 final입니다.
전체 156건 = train 84 / validation 36 / calibration 12 / test 24.
가족은 분할을 넘지 않지만, 같은 작성자의 문체와 개념은 반복될 수 있습니다.
실제 저장소 성공률·비용 절감의 증거로 해석하지 않습니다.

한 작업을 세 가지 선택지 순서로 학습해 손실을 평균 냅니다. 다음 epoch에는
나머지 순서군을 사용합니다. validation의 여섯 순서 평균 NLL로만 후보를 고르고,
두 seed의 파일을 고정한 후 final을 평가합니다. 216개 validation 판단은
36개 작업의 반복이며, 독립 작업 216건이 아닙니다.

영문 난이도 분류에 국한한 실험입니다. 분할 모델, 한글, 실제 코드 실행,
모델별 비용/성공률, Go ONNX 배포는 별도 검증입니다.

## Reproduce locally

Use the pinned maintainer MPS environment described in
[the pilot report](../../../docs/mps-pilot.en.md). Run each seed sequentially to
limit unified memory use. Output directories must be new.

```sh
python scripts/training/pdca_train.py train --base .cache/training/base --data benchmarks/training/pdca-02 --out .cache/training/pdca-02-seed1729 --seed 1729
python scripts/training/pdca_train.py train --base .cache/training/base --data benchmarks/training/pdca-02 --out .cache/training/pdca-02-seed2718 --seed 2718
# Freeze both selection.json / head hashes before either final check:
python scripts/training/pdca_train.py check --base .cache/training/base --data benchmarks/training/pdca-02 --out .cache/training/pdca-02-seed1729 --seed 1729
python scripts/training/pdca_train.py check --base .cache/training/base --data benchmarks/training/pdca-02 --out .cache/training/pdca-02-seed2718 --seed 2718
```

English: Keep the same acceptance criteria as failed PDCA-01. Its viewed test
becomes validation; eight new families form the final test. Three-order gradient
averaging and all-six-order validation selection target position bias.
Train/validation/calibration/test counts are 84/36/12/24, including 24 legacy
training rows. Authored synthetic labels, shared writing style and small samples
limit generalization. Two seeds are repeatability checks, not independent datasets.
