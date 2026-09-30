# 초소형 판단 모델 실험 — tinyhead-01

[English](README.en.md) · [진행 이슈 #15](https://github.com/teamswyg/laya-tools/issues/15) · [상위 학습 이슈 #11](https://github.com/teamswyg/laya-tools/issues/11) · [공개 HF 컬렉션](https://huggingface.co/collections/JooYoon/riidolaya-public-research-6abcbd5ddb1917912fc5de38)

목적은 “작업을 어느 수준의 모델에 맡길까?”라는 작은 결정을 더 적은 자원으로 실행하는 것입니다. 긴 답변을 생성하는 LLM을 새로 만드는 실험은 아닙니다. 현재 라우터의 기본 동작은 바꾸지 않습니다.

## 이번에 실제로 한 일

v0.2의 고정된 Laya 인코더와 판단 헤드를 부모로 삼았습니다. LayerNorm의 학습된 곱셈·덧셈을 선형 층에 합치고, 그 헤드를 FP32 / INT8 / 3진으로 변환했습니다. 3진은 가중치를 `-1, 0, +1`과 행별 배율로 표현합니다. 0 여부를 비트맵으로, 0이 아닌 값의 부호만 별도 비트열로 저장합니다. 논문의 아이디어를 참고한 독립 Go 구현이며, 논문의 x86/GPU 커널이나 파일 형식과 호환된다는 뜻은 아닙니다.

- Python 유지: 원래 PyTorch 모델을 불러와 MPS에서 공개 합성 문장의 특징을 한 번 추출하는 변환 도구.
- Go로 구현: 헤드 실행, 압축, 검증 손실에 따른 후보 선택, 정확도·신뢰도 지표, 반복 측정, 모델 파일 검증, 공개 패키지·CI 확인.
- 학습된 원본 헤드를 사후 양자화한 실험입니다. 인코더 재학습이나 ternary QAT는 아직 하지 않았습니다.
- 판단 결과는 `fast / standard / strong`의 확률 세 개입니다. 실제 코드 작업이나 모델 전환을 실행하지 않습니다.

Python을 Go로 바꾼다고 MPS 커널 자체가 빨라지지는 않습니다. 이번 전환은 Python 없이 헤드를 실행하고 반복 실험을 재현할 수 있게 하는 데 의미가 있습니다.

## 측정 결과와 해석

[전체 결과](results.json), [사전 계획](plan.json), [측정 범위](measurement.json). Apple M4 Pro / Go 1.27.1. 아래는 인코더 특징이 이미 메모리에 있을 때 **헤드만** 100,000회 실행한 단일 실행 평균입니다. 속도 우열에 통계적 유의성을 주장하지 않습니다.

| 형태 | 파일 크기 | 헤드 시간 | 판단 변경 / 264건 | 압축 기준 |
|---|---:|---:|---:|---|
| 원래 v0.2 safetensors | 20,772 B | 이번 비교 대상 아님 | 기준 | 부모 |
| 합친 FP32 | 12,332 B | 4.26 µs | 0 | 통과 |
| 행별 INT8 | 3,116 B | 4.21 µs | 0 | 통과 |
| 3진, 임계값 0.5 | 660 B | 5.91 µs | 0 | 통과 |
| 3진, 임계값 0.7 | 616 B | 5.84 µs | 1 | 탈락 |
| 3진, 임계값 1.0 | 571 B | 5.14 µs | 1 | 탈락 |
| 3진, 임계값 1.3 | 533 B | 4.76 µs | 1 | 탈락 |

모두 호출당 할당 0회입니다. 660B 후보는 부모 헤드보다 약 96.8% 작지만 FP32보다 느립니다. 비트를 푸는 비용 때문일 가능성이 있으며, 이번 프로파일만으로 CPU 명령별 원인을 확정하지는 않습니다. SIMD 구현은 아직 없습니다.

검증 36건의 NLL만으로 0.5 후보를 선택했습니다. train 180 / validation 36 / calibration 24 / 이전 test 24는 모두 **이미 사용하거나 확인했던 합성 데이터**입니다. 재사용 test를 새로운 독립 검증이라고 부르지 않습니다. 통과는 이 범위의 압축 회귀 기준 통과이며 실전 성공률·토큰 절약·한국어·새 저장소에 대한 증명이 아닙니다. 기존 confidence 0.9를 유지했습니다. 강한 모델이 필요한 일을 fast로 보낸 새 오류는 0건입니다. 확률은 달라졌으므로 더 넓은 검증이 필요합니다.

**660B만으로 전체 라우터가 돌아가지는 않습니다.** 원래 Laya 인코더와 토크나이저가 여전히 필요합니다. 특징 추출 과정의 최대 RSS는 약 2.60 GiB였습니다. 헤드 비교 프로세스의 RSS에는 JSON 특징셋·Go 런타임·프로파일러까지 포함되며, 전체 Laya 메모리가 아닙니다. MPS 드라이버 메모리와 RSS는 중복될 수 있어 더하지 않습니다.

3진 후보의 0 비율은 약 39.7%이고, 부호·비트맵은 616B, 헤더·배율·편향은 44B입니다. 기호만 약 1.604 bit/weight로 **1.58보다 낮지 않습니다**. 연속 5-trit 포장은 615B로 오히려 1B 작고, 128개 블록마다 패딩하는 5-trit 포장은 624B입니다. 더 희소한 탈락 후보가 작다는 이유만으로 성능 기준을 바꾸지 않습니다.

## 재현과 사용

일반 사용자는 기존 `riidolaya`를 계속 사용합니다. 아래는 연구자용 별도 도구입니다. 로컬 연산에는 HF 연결이 필요 없고, 이미 받은 모델·특징을 재사용합니다.

```sh
# 최초 특징 추출만 기존 maintainer PyTorch/MPS 환경 필요
.cache/mps-training-venv/bin/python scripts/training/export_tiny_features.py \
  --base .cache/training/base \
  --package .cache/training/pdca-06-package-v0.2 \
  --out .cache/tinyhead/features.json

# 이후 Go만으로 압축·평가. out은 존재하지 않는 디렉터리여야 함
go run ./cmd/riido-tinybench \
  --features .cache/tinyhead/features.json \
  --out .cache/tinyhead/my-run --iterations 100000

# 공개 후보 패키지 구조·SHA·부모·라이선스·결과 연결 확인
# 공개 직전에는 반드시 --require-ci 사용
go run ./cmd/riido-hubcheck --package .cache/tinyhead/package --require-ci

go test -race ./internal/tinyhead ./internal/researchbundle
go test ./internal/tinyhead -run '^$' -bench BenchmarkPredict -benchmem
```

특징 캐시는 토크나이저·인코더 리비전, pooling, 입력 해시와 함께 식별해야 합니다. 이 실험은 부모 헤드와 데이터 SHA를 고정한 공개용 264건만 받습니다. 새로운 원문이나 사내 저장소를 자동 수집하지 않습니다. `--cpuprofile .cache/tinyhead/local.pprof`는 선택 사항이며 프로파일은 공개하지 않습니다. 시간 비교 실행에서는 프로파일러를 끕니다.

`RDTINY01`은 실험용 포맷입니다. little-endian 44B 헤더는 magic 8B, width 4B, kind 4B, temperature FP32, bias 3×FP32, scale 3×FP32입니다. payload는 행 우선 3×width 가중치입니다. FP32는 4B, INT8은 signed byte, 3진은 연속 presence bitmap 뒤에 nonzero 순서의 양수 부호 비트를 둡니다. padding은 0입니다. 추론은 FP64 누산·LayerNorm epsilon 1e-5·stable softmax를 사용합니다. 로더는 크기·유한값·끝의 불필요한 바이트를 검사합니다. 아직 공개 안정 API를 약속하지 않습니다.

헤드 모델은 읽기 전용입니다. 호출자별 scratch 배열을 사용하므로 잠금이 필요 없습니다. 뜨거운 경로에는 map이 없고, 세 클래스 결과는 `[3]float64`, 가중치는 연속 byte 배열입니다. 보고서·매니페스트의 map은 느린 준비 단계에만 있습니다. 이 크기의 선형 분류기에 ECS 전체 구조를 도입할 근거는 아직 없습니다.

## 자료 검토에서 반영한 점

출처 확인일 2026-09-30, 코드 라이선스와 리비전은 [sources.json](sources.json). 제3자 코드·가중치·데이터를 이번 구현에 복사하지 않았습니다.

- [Laya](https://huggingface.co/convaiinnovations/laya): 현재 부모 모델. 기존 Apache-2.0 고지와 원본 리비전을 유지합니다. 범용 선택지 헤드와 우리의 고정 난이도 헤드는 다릅니다.
- [Laya-MLX](https://github.com/mizorewww/laya-mlx): Apple Silicon의 다음 인코더 비교 후보. FP16 포트와 검증 자료를 참고하되, 다른 Mac의 수치를 우리 성능으로 사용하지 않습니다. 다음 실험은 동일 문장·pooling·온도·정확도로 MPS와 비교해야 합니다.
- [JEV / TypeSafe](https://typesafe.ai/): 작은 결정을 별도 처리하는 제품 방향을 참고합니다. API 접근과 모델·데이터 재배포 권리는 별개입니다. 이번에 유료 호출·응답 학습·가중치 복제를 하지 않았습니다.
- [Gestalt-Lab/jeff](https://github.com/Gestalt-Lab/jeff): Qwen3-4B 기반 어댑터, 코드·어댑터 Apache-2.0. calibration과 정확도를 별도 보고하는 점을 채택합니다. 최소 사양 후보로 다운로드하지 않았습니다.
- [firelex/jeff](https://github.com/firelex/jeff): 별개 프로젝트입니다. 코드 MIT, 가중치 Apache-2.0 표기. 평가셋과 학습셋의 중복 제거, 실패 결과 유지가 참고점입니다. [학습 자료](https://github.com/firelex/jeff/blob/main/docs/data-sources.md)는 라이선스가 섞여 있으므로 그대로 가져오지 않습니다.
- [BitNet](https://github.com/microsoft/BitNet), [b1.58 논문](https://arxiv.org/abs/2402.17764): 3진 모델과 전용 실행 경로를 참고합니다. 기존 Laya를 파일 포맷만 바꿔 같은 성능의 BitNet으로 만들 수 있다는 근거는 아닙니다.
- [BITCOS 논문](https://arxiv.org/html/2609.16338v1), [소개 글](https://news.hada.io/topic?id=33822): 0이 많은 3진 가중치를 비트맵과 부호로 나누는 저장 아이디어를 실제 구현했습니다. 논문의 Intel 커널 속도를 Apple Silicon에 적용하지 않습니다. 위의 byte 비교처럼 형식·패딩·배율을 포함해 비교합니다.

이는 재배포 범위 점검이며 원 모델의 모든 사전학습 데이터 권리를 검증했다는 보장은 아닙니다. 향후 코드를 가져오면 해당 리비전의 LICENSE/NOTICE를 먼저 포함합니다.

## 다음 경로와 모델 관계

| 경로 | 부모·자식 관계 | 목적 / 다음 통과 기준 |
|---|---|---|
| 현재 고정 헤드 | Laya encoder → v0.2 → FP32/INT8/3진 형제 | 같은 특징·같은 라벨, 압축 손실 확인. 기본 라우터 승격은 별도 |
| 3진 QAT | v0.2 특징 → 새로 학습한 3진 헤드 | 이번 사후 압축보다 더 작은 희소도에서 회귀 줄이기. train만 업데이트, validation만 선택, calibration만 보정 |
| 인코더 없는 학생 | 공개 원문 → word/char hash → 작은 선형 분류기 | 전체 메모리를 줄이는 핵심 가설. LLM으로 부르지 않음. Go 학습 가능. 규칙 baseline과 비교하고 모르는 입력은 유보 |
| 인코더 경량화 | 기존 인코더 → INT8 또는 MLX FP16 비교 | 실제 처음 실행·반복 실행·RSS/MPS와 품질을 같이 측정. 지원 그래프·정밀도 확인 전 승격 금지 |

다음 새 학습은 **서로 다른 출처/작업 계열로 고정한 미사용 final**을 먼저 만들어야 합니다. 예: 문서 수정, 로컬 함수 변경, 테스트 실패 조사, API 계약 변경, 동시성·데이터 정합성 작업. 단어 하나로 난이도가 결정되지 않도록 같은 어휘의 쉬운/어려운 쌍, 한국어·영어, 정보 부족·범위 밖 사례를 넣습니다. 초기 expected tier는 가설이며 실제 테스트 통과·재시도·비용 결과와 분리합니다. 새로운 실전 성공 라벨은 아직 확보하지 않았습니다.

학생 모델은 교사의 틀린 판단까지 배우지 않게 원래 정답과 교사 확률을 구분합니다. confidence를 무조건 신뢰하지 않고 accepted precision, coverage, strong→fast 오류, 보정 오차, 다중 seed, 새 저장소 결과를 함께 봅니다. decomposition·repo routing은 별도 형제 헤드로 두고 난이도 성공을 전용하지 않습니다. 조합 실험은 한 번에 하나만 수행해 24GiB Mac의 메모리와 발열을 관리합니다. “세계 최초/최저 사양”은 검증된 주장이 아닙니다.

## 공개 작업 공간

[HF 모델 패키지](https://huggingface.co/JooYoon/riidolaya-tinyhead-experiment-v0.1)는 실험 코드의 CI 성공 후 게시합니다. GitHub에는 코드·계획·작은 결과·해시만, HF에는 6개 헤드와 공개 합성 특징·확률·라이선스·manifest를 둡니다. 정확한 HF 커밋과 게시 검증은 [#15](https://github.com/teamswyg/laya-tools/issues/15)에 기록합니다. 기존 v0.1/v0.2를 덮어쓰지 않습니다.

로컬 실험 → 공개 파일만 새 staging 폴더 → Go 무결성 검사 → secret scan → 정확한 소스 커밋 CI → HF 게시 → 고정 커밋으로 다시 다운로드·전수 해시 검증 순서입니다. 업로드는 별도 단계라 네트워크 장애가 학습·평가를 막지 않습니다. 패키지 검사는 데이터 권리 검토를 대신하지 않습니다. 원시 pprof·토큰·절대 로컬 경로·사내 이름·실제 사용자 입력은 올리지 않습니다. HF Space 서버나 유료 GPU는 만들지 않았습니다.
