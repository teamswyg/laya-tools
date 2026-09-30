# PDCA-05 / warm-started all-six-order consistency

PDCA-04 stopped on validation before final inference. Preserve that partial
history and keep its unexecuted final/calibration. Viewed final sets from
PDCA-01/02/03 remain development training. 156/36/24/24 cases.

Warm-start each seed from its exact PDCA-03 head, pinned in plan.json. Use a fresh
optimizer; this is not checkpoint-resume equivalence. Compare all six orders per
update with mean CE + JS, weight 1. LR 1e-5, six epochs.
Select by validation error rate + winner flip rate + 0.1×JS, never final scores.
Final acceptance gates remain those of PDCA-01.

```sh
python scripts/training/pdca_train.py train --base .cache/training/base --data benchmarks/training/pdca-05 --out .cache/training/pdca-05-seed1729 --seed 1729 --initial-head .cache/training/pdca-03-seed1729/head.safetensors
python scripts/training/pdca_train.py train --base .cache/training/base --data benchmarks/training/pdca-05 --out .cache/training/pdca-05-seed2718 --seed 2718 --initial-head .cache/training/pdca-03-seed2718/head.safetensors
# Both candidates must be frozen before either final evaluation:
python scripts/training/pdca_train.py check --base .cache/training/base --data benchmarks/training/pdca-05 --out .cache/training/pdca-05-seed1729 --seed 1729
python scripts/training/pdca_train.py check --base .cache/training/base --data benchmarks/training/pdca-05 --out .cache/training/pdca-05-seed2718 --seed 2718
```

These exact intermediate head files are retained locally. Re-running prior MPS
training is not a promise of bit-identical hashes. A published replacement head
supports inference reproduction; the training recipe alone does not promise
bitwise reproduction or supply all intermediate warm-start checkpoints.

한국어: 사이클 4의 최종 평가를 실행하지 않은 상태에서 계획을 새로 고정했습니다.
사이클 3 head를 정확한 해시로 검증해 이어 학습하며 optimizer는 새로 만듭니다.
여섯 선택지 순서를 매번 함께 학습합니다. 검증용 선택 기준을 바꿨지만 최종
통과 기준과 신뢰도 0.9는 유지합니다. 합성 영어 사례의 제한된 연구 실험이며
자동 실행·한국어·실제 코딩 비용 절감을 검증한 것은 아닙니다.
