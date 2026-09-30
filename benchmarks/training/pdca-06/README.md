# PDCA-06 / fixed-tier classifier on the frozen Laya encoder

The original arbitrary-choice head repeatedly failed order robustness despite
accuracy gains. This experiment changes architecture: mean-pool state-only
non-special-token embeddings from the frozen Laya encoder, then train
LayerNorm(1024) + Linear(1024, 3). Only **5,123** parameters train.

The fixed fast/standard/strong labels are outputs, not an input option list.
Output serialization invariance is structural. It is NOT repaired robustness of
the original choice head, and it cannot answer arbitrary-choice or decomposition
questions. The existing Go ONNX artifact cannot load this head.

180 train / 36 validation / 24 calibration / 24 new final, two seeds, same numeric
accuracy/safety/coverage/precision/legacy gates and confidence 0.9.
AdamW .001/.003, weight decay .1, batch 32, up to 100 epochs per LR.
Select by validation NLL; freeze both candidates before final evaluation.
Feature extraction caches training/validation embeddings only, not final data.

```sh
python scripts/training/fixed_tier.py train --base .cache/training/base --data benchmarks/training/pdca-06 --out .cache/training/pdca-06-seed1729 --seed 1729
python scripts/training/fixed_tier.py train --base .cache/training/base --data benchmarks/training/pdca-06 --out .cache/training/pdca-06-seed2718 --seed 2718
python scripts/training/fixed_tier.py check --base .cache/training/base --data benchmarks/training/pdca-06 --out .cache/training/pdca-06-seed1729 --seed 1729
python scripts/training/fixed_tier.py check --base .cache/training/base --data benchmarks/training/pdca-06 --out .cache/training/pdca-06-seed2718 --seed 2718
```

[Measurements](../../results/pdca-06) · [All attempts and limitations](../../../docs/pdca-results.en.md)

한국어: 원본 인코더를 동결하고 난이도 세 등급을 직접 분류하는 작은 head를
학습합니다. 약 20KiB 가중치라도 원본 인코더의 메모리/저장은 필요합니다.
두 seed에서 새 final 24/24, 원본 15/24였습니다. 이는 작은 자체 작성 영어
평가의 1차 결과이며 한국어·실제 저장소 성공·비용 절감·범용 선택지 성능의
증거가 아닙니다. 사용자 실행 경로는 Go를 유지하며 새 형식의 ONNX 검증은
별도 단계입니다. Python은 유지보수 학습/참조 전용입니다.
