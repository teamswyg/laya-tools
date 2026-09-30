# PDCA 실측 기록 — 모든 시도

[English](pdca-results.en.md) · [방법과 한계](pdca-tuning.ko.md)

수용은 보정된 확률 0.9 이상인 판단입니다. 각 final은 24건이며, 두 seed는 같은 데이터에서 학습을 반복한 것입니다.

| Cycle / seed | 정답: 원본 → 학습 | strong 정답 | 수용 정답/수용 수 | 순서 뒤집힘 | 전체 판정 |
|---|---:|---:|---:|---:|---|
| [pdca-01 / 1729](../benchmarks/results/pdca-01/seed1729-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/9 | 34/120 | 실패 |
| [pdca-01 / 2718](../benchmarks/results/pdca-01/seed2718-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/10 | 29/120 | 실패 |
| [pdca-02 / 1729](../benchmarks/results/pdca-02/seed1729-evaluation.json) | 16/24 → 16/24 | 2/8 | 0/0 | 25/120 | 실패 |
| [pdca-02 / 2718](../benchmarks/results/pdca-02/seed2718-evaluation.json) | 16/24 → 18/24 | 4/8 | 3/4 | 30/120 | 실패 |
| [pdca-03 / 1729](../benchmarks/results/pdca-03/seed1729-evaluation.json) | 16/24 → 22/24 | 7/8 | 5/5 | 14/120 | 실패 |
| [pdca-03 / 2718](../benchmarks/results/pdca-03/seed2718-evaluation.json) | 16/24 → 21/24 | 8/8 | 6/6 | 19/120 | 실패 |

순서 뒤집힘은 같은 24건에 다섯 가지 다른 선택지 순서를 적용한 120번 비교입니다. 독립 작업 120건이 아닙니다. 수용 건수를 24로 나누면 수용률입니다.

PDCA-01의 원본 보고서는 9/10 precision을 float32 0.899999976으로 기록했습니다. 이후 정수 건수 나눗셈으로 수정했고 0.9 기준은 유지했습니다. 다른 실패 조건 때문에 PDCA-01 전체 실패 판정은 동일합니다.

PDCA-04는 validation만 보고 조기 중단했습니다. 두 seed 완료나 final 통과를 주장하지 않습니다. [중단 시점의 기록](../benchmarks/results/pdca-04/stopped.json)을 보존했습니다. PDCA-05는 아직 평가하지 않은 같은 final/calibration을 사용하며, 이미 평가한 PDCA-01/02/03 final을 다시 숨겨진 시험으로 취급하지 않습니다.

단순 키워드 참조는 PDCA-02/03에서 19/24였습니다. 확률 보정이나 수용 정책이 없는 비교이며, 학습 모델을 써야 한다는 당위의 증거로 생략하지 않습니다. 반복 적응 실험, 작은 표본, 자체 작성 레이블의 한계 때문에 실제 저장소/비용으로 일반화하지 않습니다.
