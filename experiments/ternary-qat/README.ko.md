# 3진 의사결정 헤드의 실제 학습과 PDCA

[English](README.en.md) · [이슈 #17](https://github.com/teamswyg/laya-tools/issues/17) · [이전 압축 실험](../tinyhead/README.ko.md) · [공개 HF 컬렉션](https://huggingface.co/collections/JooYoon/riidolaya-public-research-6abcbd5ddb1917912fc5de38)

이번에는 학습된 가중치를 마지막에 줄이기만 하는 방식에서 나아가, **실제로 3진 가중치로 예측하면서 학습**했습니다. 두 번의 PDCA와 총 32개 학습 후보를 실행했습니다. 선택된 헤드는 572B이며 3,072개 선형 가중치를 기준으로 헤더·배율·편향을 포함해 1.490bit/weight입니다. 전체 Laya가 이 크기가 된 것은 아닙니다.

## 왜 생성 능력보다 판단 능력에 집중하는가

우리에게 필요한 출력은 문장이 아니라 fast / standard / strong 확률 세 개입니다. Laya도 이미 인코더 기반 결정 모델이라, 생성 디코더를 새로 제거해서 메모리를 줄이는 단계는 없습니다. 입력 문장을 이해하는 인코더는 여전히 필요합니다. 이번에는 그 뒤의 작은 판단 헤드에 3진 학습이 유효한지 확인했습니다. 인코더 전체나 활성값은 3진화하지 않았습니다.

## 실제로 달라진 학습 방법

1. 부모 v0.2의 LayerNorm affine을 선형 층에 합친 가중치에서 시작합니다. 모든 후보의 초기 부모는 같습니다.
2. 학습 중 보관하는 실수 가중치(shadow weights)는 Go의 FP64 배열입니다. 예측할 때마다 행별 평균 절댓값에 임계값을 곱해 작은 가중치를 0으로 만들고 나머지를 -1/+1과 배율로 바꿉니다.
3. 실제 압축 모델의 예측 오차를 이용해 실수 가중치를 갱신합니다. 반올림은 미분하기 어려워 identity STE라는 근사 기울기를 사용합니다. 배율·임계값의 미분은 계산하지 않습니다. 이는 정확한 반올림 미분이나 BitNet 원논문의 재현이 아닙니다.
4. 각 epoch에서 검증 NLL만으로 600B 이하 후보를 비교합니다. final은 읽지 않습니다. 두 seed의 가중치를 선택한 후 calibration 24건만으로 온도를 고릅니다. 기존 confidence 기준 0.9는 유지합니다.
5. 선택 모델과 계획·데이터의 SHA를 기록한 뒤 별도 check 명령에서 처음 final을 평가합니다.

학습·평가·게시 검증은 Go CPU만 사용합니다. 이미 캐시한 학습 특징을 재사용하며, 새로운 final 문장의 특징을 추출할 때만 기존 Python/PyTorch MPS를 사용했습니다. 두 seed는 다른 minibatch 순서를 뜻하며 독립적인 무작위 초기 모델 두 개는 아닙니다.

## PDCA 기록

| 반복 | 계획·실행 | 확인 | 후속 조치 |
|---|---|---|---|
| 01: 정답만 학습 | 2 seeds × 4 임계값 × 2 학습률, 각 80 epochs | 507B가 선택됐지만 이미 쉬운 train을 거의 확실히 맞혀 gradient가 작았음. 검증 35/36, 같은 압축의 학습 전후 차이가 매우 작음 | 성공적인 학습 개선으로 주장하지 않음. final 미개봉 상태에서 다음 계획 고정 |
| 02: 점수 차이도 학습 | 같은 16개 조합, 각 120 epochs. 정답 CE + 부모의 중심화 logit MSE × 0.1 | 572B, 검증 36/36. 같은 임계값의 학습 전 PTQ는 35/36. 3진 기호 46개 변경 | 두 seed와 온도를 고정하고 final 처음 확인 |

두 번째 학습은 부모가 세 클래스에 부여한 **점수 사이의 차이**도 따라가게 합니다. 이미 정답을 맞힌 사례에서도 압축으로 손실된 정보를 학습 신호로 쓸 수 있습니다. 교사는 고정된 우리 공개 헤드이며 외부 API나 유료 모델을 호출하지 않았습니다. 교사 신호는 train 180건에만 적용했습니다. 정확한 보조 손실은 `0.1 / (2×3) × Σ(학생 중심화 logit − 교사 중심화 logit)²`입니다.

[첫 계획](plan-01.json), [두 번째 계획](plan.json), [32개 후보 요약](development-cycles.json), [최종 결과](results.json), [반복 시간 측정](benchmark.json), [자원·재현 기록](measurement.json). 모든 epoch의 원본 기록은 HF의 `development-cycle-01.json`과 `selection.json`에 있습니다. 같은 실험에서 재학습한 모델이 부모를 대체하지는 않습니다.

## 최종 결과: 장점과 손해를 함께 보기

학습 전에 새로 작성해 고정한 영어 합성 final 36건, 클래스당 12건입니다. 기대 난이도는 명시된 작업 범위를 기반으로 작성한 가설이며 실제 코드 작업 성공/실패에서 얻은 정답이 아닙니다. 기존 train 180 / validation 36 / calibration 24는 재사용했습니다. 기존 test 24는 학습·선택·보정에 사용하지 않았습니다. final은 이번 확인 이후부터 이미 본 데이터입니다.

| 항목 | 부모 FP32 | 이전 660B PTQ | 새 572B QAT, 두 seed |
|---|---:|---:|---:|
| 정답 | 33/36 | 33/36 | 34/36 |
| confidence ≥0.9 판단 중 정답 | 32/33 | 31/31 | 32/33 |
| 판단 수락률 | 91.7% | 86.1% | 91.7% |
| strong 재현율 | 12/12 | 12/12 | 12/12 |
| strong→fast 오류 | 0 | 0 | 0 |
| NLL, 낮을수록 좋음 | 0.171 | 0.128 | 0.179 |
| Brier, 낮을수록 좋음 | 0.111 | 0.088 | 0.109 |

정답 수와 크기는 개선됐지만 **새 모델이 모든 품질 지표에서 더 좋은 것은 아닙니다**. 이전 PTQ보다 확률 품질과 accepted precision이 나빠졌습니다. 사전 기준은 두 seed 모두 통과했지만 36건에서 1건의 차이는 통계적으로 일반적인 우월성을 증명하지 않습니다. 34/36의 Wilson 95% 구간은 약 81.9~98.5%입니다. 자동 라우터로 승격하지 않습니다.

남은 오답은 쉬운 문서 변경을 standard로 보는 과대 분류 2건입니다. rate-limit 단위 오타는 확신이 약 0.852여서 보류되지만, watcher 문서 glob 수정은 약 0.953으로 틀리게 수락됐습니다. 안전한 하향 라우팅뿐 아니라 불필요한 상향 라우팅 비용도 다음 실제 작업 평가에서 확인해야 합니다. 이번에는 이 final을 보고 온도나 가중치를 다시 바꾸지 않았습니다.

## 1.58비트보다 작은 이유와 메모리 범위

선택된 헤드의 0 비율은 약 62.63%입니다. 세 기호가 같은 빈도로 나오지 않기 때문에 고정적인 1.58비트 표현만이 가능한 것은 아닙니다. 파일 구성은 presence bitmap 384B + nonzero sign 144B + 헤더·배율·편향 44B = 572B입니다.

- 기호 부분: 528×8÷3,072 = 1.375bit/weight.
- 전체 파일: 572×8÷3,072 = 1.4896bit/weight.
- 분모는 **선형 층의 3,072개 계수**입니다. 전체 encoder, float 활성값, tokenizer나 작업 메모리의 비트 수가 아닙니다.
- float32 배율·편향과 float64 누산·normalization은 그대로입니다. 학습은 실수 shadow weights와 gradient도 사용합니다. 추론 파일 크기와 학습 메모리를 혼동하지 않습니다.

M4 Pro, Go 1.27.1에서 동일 36개 특징을 사용하고 순서를 순환하며 30,000회씩 5번 측정한 중앙값:

| 헤드 | 크기 | 헤드만 실행 |
|---|---:|---:|
| 합친 FP32 | 12,332B | 4.23µs |
| 이전 3진 PTQ | 660B | 5.73µs |
| QAT seed 1729 | 572B | 5.01µs |
| QAT seed 2718 | 572B | 5.03µs |

이전 3진보다 약 12% 짧은 시간과 13.3% 작은 파일이지만 FP32보다 느립니다. 5회 표본은 시스템 전체 성능 보증이 아닙니다. 모두 호출당 할당 0회입니다. SIMD/GPU 커널은 사용하지 않았습니다. 전체 인코더를 포함한 지연 개선으로 주장하지 않습니다.

두 번째 후보 16개 학습 전체는 약 3.25초, 프로세스 최대 RSS 약 24.1MiB였습니다. 이는 **미리 만든 특징을 쓰는 작은 Go 헤드 학습**입니다. 새 특징을 만든 MPS 프로세스는 약 2.51GiB RSS를 사용했습니다. 두 프로세스는 순서대로 실행됐고 RSS와 MPS 드라이버 메모리를 합산하지 않습니다.

## 재현

HF에서 파일을 받을 때는 이슈에 기록된 immutable commit을 사용합니다. GitHub에는 모델/특징을 올리지 않습니다. 기존 환경에서 다음처럼 실행할 수 있습니다.

```sh
# 새로운 문장의 특징 추출에만 PyTorch/MPS 필요
.cache/mps-training-venv/bin/python scripts/training/export_qat_features.py \
  --base .cache/training/base \
  --package .cache/training/pdca-06-package-v0.2 \
  --out .cache/ternary-qat/final-features.json

# train 명령은 final 인자를 거부합니다.
go run ./cmd/riido-qat --stage train \
  --development .cache/tinyhead/features.json \
  --out .cache/ternary-qat/new-training

# 위 단계가 두 seed를 선택·고정한 뒤 별도로 평가합니다.
go run ./cmd/riido-qat --stage check \
  --development .cache/tinyhead/features.json \
  --final .cache/ternary-qat/final-features.json \
  --selection .cache/ternary-qat/new-training/selection.json \
  --out .cache/ternary-qat/new-check

# 품질 수치를 재계산하고, 공개 직전 정확한 소스 CI까지 확인합니다.
go run ./cmd/riido-hubcheck --package .cache/ternary-qat/package --require-ci
```

출력 경로는 새 디렉터리여야 합니다. HF 패키지에서 받은 `development-features.json`을 `--development`로 직접 사용하면 Python 없이 학습부터 재현할 수 있습니다. 01 반복은 `--plan experiments/ternary-qat/plan-01.json`으로 재현합니다. `--stage bench`는 동일한 입력·선택 파일 인자로 시간만 반복 측정합니다. HF 업로드는 별도 단계이므로 네트워크 상태가 학습을 막지 않습니다.

공개 패키지 검사는 해시만 확인하지 않습니다. 다운로드한 헤드와 final 특징으로 36개 확률·정답 수·수락 수·품질 기준을 Go에서 다시 계산합니다. 원시 프로파일·사내 이름·실제 사용자 프롬프트·인증정보·절대 로컬 경로는 게시하지 않습니다. LICENSE와 부모 NOTICE를 유지합니다.

## 다음 PDCA

- 이 36건을 더 이상 새로운 final로 사용하지 않습니다. 다음 실험 전에 새 작업 계열·출처의 final을 먼저 고정합니다.
- 짧은 문서 수정에 복잡한 용어가 들어간 경우처럼, 표면 어휘와 실제 작업 범위가 다른 train/validation/calibration 사례를 늘립니다. 한 사람이 쓴 반복 템플릿을 벗어나 실제 공개 작업의 결과 라벨도 수집해야 합니다.
- 보정 데이터가 지나치게 쉬워 최소 온도 0.5가 선택되는 문제를 검토합니다. confidence 기준을 내려서 수락률을 맞추지 않습니다.
- 한 번에 바꿀 요인을 제한해 3진 임계값, 학습률, teacher 비중, 보정의 기여를 분리합니다. 새 실험은 NLL·accepted precision의 악화도 명시적으로 제한하는 더 엄격한 사전 기준을 검토합니다.
- 전체 메모리 절감은 encoder-free 학생 또는 인코더 자체의 저비트 학습/실행으로 별도 검증합니다. 현재 헤드 성공을 근거로 전체 Laya가 3진에서도 같은 성능일 것이라고 가정하지 않습니다.
- 새 모델은 버전별 HF 저장소로 게시하고 부모·데이터·코드·CI·해시를 서로 연결합니다. 이전 모델과 실패 기록은 보존합니다. 지속 게시가 검증되지 않은 모델의 자동 승격을 뜻하지는 않습니다.

## 근거와 라이선스

수학적 아이디어를 참고해 Go로 독립 구현했습니다. 외부 학습 데이터나 제3자 구현을 복사하지 않았으며 기존 Apache-2.0 부모 자산의 LICENSE/NOTICE를 유지합니다.

- [Ternary Weight Networks](https://arxiv.org/abs/1605.04711): 임계값과 배율을 이용한 3진 가중치의 선행 연구.
- [STE 연구](https://arxiv.org/abs/1308.3432): 미분하기 어려운 연산에 근사 기울기를 사용하는 배경. 우리의 identity/stop-gradient 선택 자체는 실험 설계입니다.
- [Knowledge distillation](https://arxiv.org/abs/1503.02531): 교사의 추가 정보를 학생에게 전달하는 배경. 이번 centered-logit MSE 혼합은 우리 설정이며 원논문의 결과를 재현했다는 주장이 아닙니다.
- [BitNet b1.58](https://arxiv.org/abs/2402.17764): 저비트 학습의 방향을 참고하지만 전체 Transformer의 BitNet 구현이 아닙니다.
- [BITCOS](https://arxiv.org/html/2609.16338v1): 비대칭 기호 빈도를 활용한 bitmap/sign 표현. 해당 논문의 Intel 성능을 우리 Mac에 전용하지 않습니다.
