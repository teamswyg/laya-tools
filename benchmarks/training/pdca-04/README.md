# PDCA-04 / choice-order consistency

PDCA-03 improved accuracy to 22/24 and 21/24 but failed the original combined
criteria due to 14/120 and 19/120 alternative-order flips, plus 5/24 coverage in
one seed. Keep that failure.

Average supervised losses across three orders and add their Jensen–Shannon
divergence after mapping logits back to canonical class positions. Weight is 1.
Validation selection uses mean NLL plus mean JS over all six orders. This changes
training/selection, not the one-pass inference rule. It does not achieve order
invariance by averaging six inference calls.

156 train / 36 validation / 24 new calibration / 24 new test. All previous final
tests are development training. Prior calibration becomes validation. New
calibration and final families are disjoint from all previous families.
Same temperature grid and unchanged 0.9 confidence gate for baseline and tuned.
LR 3e-5, 12 epochs, two seeds, fresh original weights per seed.

[Fixed plan](plan.json) · [Fixtures](families.json).
Reproduce with the [PDCA-02 commands](../pdca-02/README.md), replacing paths
with pdca-04. Resource guards remain active; simultaneous head graphs may use
more memory. Tests cover consistency-loss gradients and relabeling invariants.

한국어: 정확도만 개선된 사이클 3의 실패를 보존합니다. 같은 작업의 선택지
순서를 바꿨을 때 확률이 달라지면 손실을 추가하고, 검증에서도 이를 반영합니다.
추론을 여섯 번 합산하는 방식이 아닙니다. 보정과 최종 평가는 새 가족을 쓰며
판정 기준을 낮추지 않습니다. 반복적인 소규모 합성 평가의 선택 편향, 같은
작성자의 문체, 실제 저장소/한국어/비용 검증 부재라는 한계는 남습니다.
