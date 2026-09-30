# 초기 측정 — 2026-09-30

[English](measurements.md) · 한국어

Apple M4 Pro, 통합 메모리 24GiB, macOS 26.6.2 arm64. 실행 경로는 Go 바이너리와 ONNX Runtime 1.30.0입니다. 개발 측정이며 독립 평가 정확도나 실제 Codex 청구액 절감을 입증하지 않습니다.

## 짧은 판단

47토큰 boolean 질문, 워밍업 1회 이후 CPU 설정별 직렬 30회 측정:

| 네이티브 CPU 스레드 | warm 중앙값 | 세션 초기화 |
|---|---:|---:|
| 1 | 40.29ms | 667ms |
| 2 | 32.71ms | 647ms |
| 4 (기본) | 26.28ms | 624ms |
| 8 | 23.56ms | 620ms |

FP32 CPU 4스레드 20회 중앙값은 38.54ms였습니다. INT8이 현재 기본입니다. 짧은 시험에서는 8스레드가 가장 빨랐지만, 4스레드는 코딩 작업에 CPU 여유를 남깁니다. 모든 입력 길이·배터리·발열 조건에서 최적이라는 뜻은 아닙니다.

모델 파일은 약 572MiB, 초기 배포 압축은 약 455MiB였습니다. Go HeapAlloc은 약 12–14MB, 별도 `/usr/bin/time -l`의 최대 RSS는 INT8 약 1.39GiB, FP32 약 1.48GiB였습니다. Go heap을 전체 앱 메모리로 해석하면 안 됩니다. 별도 live RSS는 원본 JSON에 있습니다.

pprof에서 대부분의 CPU 시간은 네이티브 호출(`runtime.cgocall` 또는 이름 없는 프레임)이었습니다. ORT 프로파일링은 오버헤드가 있어 표의 수치는 프로파일 없이 측정했습니다. Core ML MLProgram은 동적 그래프의 axis/rank 오류로 초기화하지 못했습니다. GPU 속도·GPU 메모리·ANE 사용률은 주장하지 않습니다. 현재 CPU 기본값은 시험한 경로 중 동작한 설정이며 모든 Metal/Core ML 구현과의 비교 결과가 아닙니다.

## 코드 검색

공개 corpus는 encode/httpx 0.28.1, 커밋 `26d48e0634e6ee9cdc0533996db289ce4b430177`의 패키지 폴더 284,396바이트입니다. Go 구현 전 작성한 개발 질문 18개(영어 12, 한국어 6), 키워드 후보 8개, 32줄 구간/24줄 간격, 정규화 BM25와 Laya 관련성 50:50 결합, 중복 구간 억제를 사용했습니다. 임베딩 검색은 없습니다.

지표는 반환 구간에 목표 함수의 **정의 줄**이 포함되는지입니다. 같은 함수 안의 유용한 본문 구간도 정의 줄이 없으면 실패로 계산하므로 의미적 정답률이 아닙니다. 무작위 표본이나 독립 평가셋도 아닙니다.

| 방식 | 영어 정의 줄 Hit@3 | 영어 Hit@5 | 영어 중앙값 | 한국어 Hit@3 |
|---|---:|---:|---:|---:|
| BM25 | 3/12 | 8/12 | 0.25ms | 2/6 |
| BM25 + base Laya | 3/12 | 8/12 | 1,760ms | 3/6 |
| BM25 + Laya Code r1 | 5/12 | 8/12 | 1,641ms | 2/6 |

영어 정의 줄 Hit@1은 세 방식 모두 0/12였습니다. 고정 구간과 엄격한 지표를 보완해야 제품 수준의 결론을 낼 수 있습니다. 일부 구간은 512토큰에서 잘렸고, 한국어 질문 2개는 키워드 후보가 없었습니다. 재정렬은 없는 후보를 복구하지 못합니다. 원본 결과에는 인덱싱과 모델 시작 시간을 따로 기록했습니다. 표는 기존 인덱스·상주 모델 조건이며 CLI cold 경로는 그 비용이 추가됩니다.

Code r1의 작은 top-3 개선에는 큰 지연이 추가됐습니다. 식별자나 정확한 문자열은 `--lexical` 또는 grep이 적합합니다. 종단 간 Codex 토큰·비용·작업 시간 비교는 하지 않았습니다.

## 모델 라우터 시험

주석 오타부터 여러 지역에 걸친 마이그레이션까지 직접 작성한 12개 요청에 한국어 1개가 포함됐습니다. 기본 확신도 0.9에서 모두 strong/default를 유지했습니다. 이는 fallback 동작을 보여주지만 **하향 선택 0건, 비용 절감 입증 없음**입니다. 임계값 인하는 실험이지 검증된 운영 권장이 아닙니다. 전용 난이도 라벨·독립 과제·후속 작업 성공 평가가 필요합니다.

## 재현

```sh
git clone --branch 0.28.1 --depth 1 https://github.com/encode/httpx.git /tmp/httpx-eval
go run ./cmd/eval --root /tmp/httpx-eval/httpx
# 설치된 체크포인트를 평가하려면 --model-dir와 --runtime 추가
riidolaya bench --threads 4 --iterations 30
riidolaya serve < benchmarks/router-queries.jsonl
```

정리한 [원본 결과](../benchmarks/results)에는 로컬 사용자명·인증 정보·비공개 코드·pprof를 포함하지 않습니다. 프로파일은 로컬 경로가 들어갈 수 있어 공개하지 않습니다.

추가 실험에서 ORT 가중치 prepacking을 끄면 최대 RSS는 약 1.39→1.24GiB로 줄었으나 중앙값이 약 27→47.5ms로 늘었고 양자화 확률도 바뀌었습니다. 이 설정은 기본값에 반영하지 않았습니다. 별도 warm JSONL 프로세스도 약 1.38GiB RSS를 사용해 모델 메모리가 시작 순간에만 필요한 것은 아니었습니다.

## x86 INT8 이식성

Linux CI에서 positive/negative 의미 검사가 한때 둘 다 약 0.5를 반환하며 실패했습니다. 동일 체크포인트는 이전 작업 및 Apple Silicon에서 통과했고, 체크섬·토크나이저 parity도 유지됐습니다. 이는 [ONNX Runtime 문서](https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html)가 설명하는 VNNI 없는 x86의 U8S8 포화 위험과 일치하는 관찰입니다.

amd64에서는 `session.x64quantprecision=1`을 켜 해당 AVX2/AVX512 CPU에서 정밀도 보존 경로를 사용하도록 했습니다. [원본 옵션 정의](https://github.com/microsoft/onnxruntime/blob/v1.30.0/include/onnxruntime/core/session/onnxruntime_session_options_config_keys.h). 기존 의미 검사는 계속 필수입니다. Apple Silicon 설정을 바꾸거나 모든 x86에서의 정확도를 보장하는 조치는 아닙니다. Linux CI는 후속 차이를 분석할 수 있도록 CPU 정보를 기록합니다.
