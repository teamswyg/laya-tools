# PDCA 실측 기록 — 모든 시도

[English](pdca-results.en.md) · [방법과 한계](pdca-tuning.ko.md)

수용은 보정된 확률 0.9 이상인 판단입니다. 각 final은 24건이며, 두 seed는 같은 데이터에서 학습을 반복한 것입니다.

**PDCA-06은 구조가 다른 난이도 전용 분류기입니다.** 원본 choice head를 고친 것이 아닙니다. 이 행의 0/120은 고정 출력 레이블을 다른 순서로 직렬화했다가 복원하는 검사이며, 원본의 선택지 순서별 재추론과 같은 지표로 해석하면 안 됩니다. 정확도·수용률·위험 분류 기준의 수치는 유지했습니다.

| Cycle / seed | 정답: 원본 → 학습 | strong 정답 | 수용 정답/수용 수 | 순서 뒤집힘 | 전체 판정 |
|---|---:|---:|---:|---:|---|
| [pdca-01 / 1729](../benchmarks/results/pdca-01/seed1729-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/9 | 34/120 | 실패 |
| [pdca-01 / 2718](../benchmarks/results/pdca-01/seed2718-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/10 | 29/120 | 실패 |
| [pdca-02 / 1729](../benchmarks/results/pdca-02/seed1729-evaluation.json) | 16/24 → 16/24 | 2/8 | 0/0 | 25/120 | 실패 |
| [pdca-02 / 2718](../benchmarks/results/pdca-02/seed2718-evaluation.json) | 16/24 → 18/24 | 4/8 | 3/4 | 30/120 | 실패 |
| [pdca-03 / 1729](../benchmarks/results/pdca-03/seed1729-evaluation.json) | 16/24 → 22/24 | 7/8 | 5/5 | 14/120 | 실패 |
| [pdca-03 / 2718](../benchmarks/results/pdca-03/seed2718-evaluation.json) | 16/24 → 21/24 | 8/8 | 6/6 | 19/120 | 실패 |
| [pdca-05 / 1729](../benchmarks/results/pdca-05/seed1729-evaluation.json) | 17/24 → 21/24 | 7/8 | 12/12 | 23/120 | 실패 |
| [pdca-05 / 2718](../benchmarks/results/pdca-05/seed2718-evaluation.json) | 17/24 → 21/24 | 7/8 | 12/12 | 18/120 | 실패 |
| [pdca-06 / 1729](../benchmarks/results/pdca-06/seed1729-evaluation.json) | 15/24 → 24/24 | 8/8 | 24/24 | 0/120 | 통과 (고정 분류) |
| [pdca-06 / 2718](../benchmarks/results/pdca-06/seed2718-evaluation.json) | 15/24 → 24/24 | 8/8 | 24/24 | 0/120 | 통과 (고정 분류) |

순서 뒤집힘은 같은 24건에 다섯 가지 다른 선택지 순서를 적용한 120번 비교입니다. 독립 작업 120건이 아닙니다. 수용 건수를 24로 나누면 수용률입니다.

PDCA-01의 원본 보고서는 9/10 precision을 float32 0.899999976으로 기록했습니다. 이후 정수 건수 나눗셈으로 수정했고 0.9 기준은 유지했습니다. 다른 실패 조건 때문에 PDCA-01 전체 실패 판정은 동일합니다.

PDCA-04는 validation만 보고 조기 중단했습니다. 두 seed 완료나 final 통과를 주장하지 않습니다. [중단 시점의 기록](../benchmarks/results/pdca-04/stopped.json)을 보존했습니다. PDCA-05는 아직 평가하지 않은 같은 final/calibration을 사용하며, 이미 평가한 PDCA-01/02/03 final을 다시 숨겨진 시험으로 취급하지 않습니다.

단순 키워드 참조는 PDCA-02/03에서 19/24였습니다. 확률 보정이나 수용 정책이 없는 비교이며, 학습 모델을 써야 한다는 당위의 증거로 생략하지 않습니다. 반복 적응 실험, 작은 표본, 자체 작성 레이블의 한계 때문에 실제 저장소/비용으로 일반화하지 않습니다.

## 1차 개선의 범위

난이도 전용 5,123-parameter head는 두 seed 모두 새 final 24/24(원본 15/24), strong 8/8, 수용 24/24 및 수용 정답 24/24를 기록했습니다. 키워드 참조는 18/24였습니다. 기존 영문 회귀 집합은 원본 20/32에서 29/32·28/32입니다. 배포 후보는 final 점수가 아니라 validation NLL로 선택한 seed 2718입니다.

학습 약 14.30초·16.75초, head 약 20KiB, 최고 RSS 약 2.75GiB, 단계 종료 driver 샘플 최대 약 2.02GiB입니다. CPU/MPS 24건 판단 차이 0, 최대 확률 차이 2.384e-7. 서로 다른 학습 방법·횟수이므로 학습 시간 차이를 동일 조건 처리량 비교로 주장하지 않습니다. 원본 인코더 메모리는 여전히 필요합니다.

이 결과로 현재 소규모 영어 합성 난이도 실험의 1차 개선을 마무리합니다. 24건 모두 맞았다는 것이 일반 정확도 100%라는 뜻은 아닙니다. 분할, 한국어, 실제 코딩 성공과 비용, Go ONNX 실행 검증은 남아 있습니다. 배포 상태와 불변 revision은 이슈 #11에 기록합니다.
