# 배열 중심 검색과 난이도 골든셋: 2026-09-30

한국어 · [English](layout-and-goldens.en.md)

이번 작업은 검색의 반복 연산을 줄이고, 실제 개발 작업을 Laya가 어떤 난이도로 보는지 기록하는 실험입니다. Actions 파일 캐시는 변경하지 않았습니다.

## 실제 변경과 비용

기존 검색은 각 코드 조각의 문자열 map을 순회하고, 질문 단어의 중복 제거와 BM25 로그·정규화 계산을 매 검색마다 반복했습니다. 새 구조는 인덱스 생성 시 단어를 정수 ID로 바꾸고 각 단어에 `chunks []int`와 `weights []float64`를 저장합니다. BM25 기여도를 미리 계산하여 질문 시 해당 단어가 있는 조각에만 더합니다. 정렬도 큰 본문 구조체 대신 작은 ID·점수 쌍을 사용하고 상위 후보만 본문으로 복원합니다.

숫자 배열과 코드 본문을 분리한 SoA입니다. ECS 전체 프레임워크를 추가하지 않았습니다. 캐시 친화성을 의도한 변경이며 하드웨어 cache-miss counter를 측정한 것은 아닙니다. 문자열→ID 사전, 인덱스 구성 중 빈도 집계, 토크나이저 어휘/BPE 사전처럼 키 조회가 필요한 map은 남겼습니다. 모든 map을 배열로 바꾸면 오히려 선형 탐색이나 큰 희소 배열 비용이 생깁니다.

Apple M4 Pro, Go 1.27.1, 256개의 합성 소스 파일과 기존 2개 fixture 파일, 같은 질문, 5회 benchmark 결과의 중앙값입니다. 생성·파일 입출력은 검색 측정에서 제외하며 인덱스 생성은 따로 측정했습니다. 비교 기준은 변경 전 main `2467752`의 검색 구현입니다.

| 측정 | 이전 | 이후 | 해석 |
|---|---:|---:|---|
| 반복 검색 | 51,459 ns/op | 25,221 ns/op | 약 51% 감소, 약 2.04배 처리속도 |
| 검색 할당 바이트 | 66,802 B/op | 23,250 B/op | 약 65% 감소 |
| 검색 할당 횟수 | 22 | 25 | 3회 증가; 총 할당량과 별도 지표 |
| 인덱스 생성 | 66.42 ms/op | 71.53 ms/op | 약 7.7% 증가 |
| 인덱스 생성 할당 | 15.81 MB/op | 15.91 MB/op | 약 0.6% 증가; 상주 메모리가 아님 |

반복 검색의 개선이며 단발 CLI나 Laya 추론 전체가 2배 빨라졌다는 뜻이 아닙니다. 공개 원본 수치는 [results](../benchmarks/results/layout-goldens-20260930)에 있습니다. pprof에서 문자열 map 조회·삽입과 검색 경로가 관찰됐지만 macOS 런타임·GC 비용도 컸습니다. pprof 원본은 공개하지 않습니다. 프로파일은 참고 증거이며 단일 원인을 확정하지 않습니다.

## SIMD와 lock 분석

- 검색 점수 합산은 연속 weights를 읽지만 문서 ID를 따라 scores에 흩어 써야 합니다. 현재는 사전 계산·해시 조회 제거가 검증된 개선입니다. 수제 ARM64/x86 어셈블리 SIMD는 추가하지 않았습니다. 다음 단계는 숫자 커널이 실제 병목인지 별도 프로파일한 뒤 scalar/벡터 결과·경계·dispatch를 비교하는 것입니다.
- 모델의 INT8 행렬 연산은 ONNX Runtime이 수행합니다. x86의 양자화 정밀도 보호 설정도 유지했습니다. [ORT 양자화 안내](https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html)는 아키텍처별 명령과 포화 문제를 설명합니다. 이번 작업에서 특정 SIMD 명령이 실행됐다고 계측한 것은 아닙니다.
- `Engine.mu`는 변경 가능한 토크나이저 캐시와 `Predict`/`Close` 사이 네이티브 세션 수명을 함께 보호합니다. 읽기처럼 보이는 Predict도 캐시를 쓰므로 RWMutex로 바꾸지 않았습니다. lock을 줄이려면 worker별 encoder와 안전한 session 수명 관리가 먼저이며, worker별 모델 복제는 RAM 비용이 있습니다.
- `sync.Once`는 프로세스 전역 ORT 환경 초기화용입니다. 반복 추론 병목으로 보지 않았습니다. 한 프로세스에서 서로 다른 runtime 경로를 바꾸는 구조도 아닙니다.
- `App`은 현재 순차 CLI/JSONL/MCP 처리 소유 객체입니다. Engine의 lock이 App 초기화까지 보호하지 않으므로 concurrent-safe로 오해하지 않게 계약을 명시했습니다. 병렬 사용자는 별도 소유 또는 외부 직렬화가 필요합니다.
- 검색 Index는 생성 이후 불변으로 사용하고 요청별 점수 배열을 따로 만듭니다. 공유 scratch buffer나 새 전역 lock을 추가하지 않았습니다. 공개 Chunks도 호출자가 수정하면 안 됩니다.

종료된 Engine의 Predict가 nil 세션을 사용하던 경계를 오류 반환으로 수정했고, 실제 모델에서 Predict/Close 동시 호출을 race detector와 함께 검사했습니다. 검색은 기존 map 기반 독립 계산과 점수·순서를 비교하고 병렬 읽기도 검사했습니다. 메모리 안전성을 속도와 교환하지 않았습니다.

## 골든셋과 단계적 실제 사용

[36개 고정 작업](../benchmarks/routing-golden.json)에 예상 난이도와 이유를 추론 전에 붙였습니다. 레이블은 작성자 판단이며 독립 전문가 합의나 실제 모델별 성공률이 아닙니다. 12→24→36개 누적 단계로 실행하며 각 단계에서 엔진을 한 번 로딩해 재사용합니다. 앞 단계 결과를 보고 레이블·임계값·입력을 변경하지 않았습니다. 기본 확신도 0.9를 유지했습니다.

| 누적 작업 | 실제 분류된 영어 | 기대 난이도 일치 | 정책상 승인 | 한국어 보류 |
|---|---:|---:|---:|---:|
| 12 | 9 | 5/9 | 0 | 3 |
| 24 | 21 | 14/21 | 0 | 3 |
| 36 | 32 | 20/32 (62.5%) | 0 | 4 |

마지막 단계에서 expected strong 11개 중 raw strong은 2개, standard는 8개, fast는 1개였습니다. 강한 작업의 과소평가가 중요합니다. 승인된 하향 오분류가 0이어도 추천 자체가 0이므로 성능 우수로 해석하면 안 됩니다. 한국어 정책 보류를 분류 정답으로 세지 않았습니다. 모든 행에 기대값·원시 제안·실제 정책 결과·확신도·보류·시간을 남겼습니다.

현재 개발에 직접 적용한 작업도 포함했습니다. Laya는 난이도를 분류했고 코딩은 이 작업의 코딩 에이전트가 수행했습니다. 선택된 하위 모델에 작업을 보내 성공률을 비교한 실험은 아닙니다.

| 실제 작업 ID | expected | Laya raw | 수행 결과 |
|---|---|---|---|
| closed-engine | standard | standard | 종료 후 호출 오류와 lifecycle/race 검사 |
| golden-loader | standard | standard | 입력·레이블 검증과 fixture 테스트 |
| json-report | standard | standard | confusion matrix와 단계별 결과 파일 |
| search-soa | strong | standard | SoA 검색, 독립 점수·순서 회귀 검사, 전후 benchmark |
| lock-lifetime | strong | standard | lock 소유·수명 분석, 실제 동시 종료 검사 |
| simd-dispatch | strong | standard | 타당성 검토만 수행; SIMD 구현은 미수행 |

36개 전부에 코딩 작업을 수행한 것은 아닙니다. 실제 구현 과제와 분류 전용 과제를 구분해야 합니다. 다음 표본은 레이블을 먼저 고정하고 모호한 작업·짧지만 위험한 수정·다국어를 추가하며, 변경된 평가셋은 새 버전으로 남기는 방식이 적합합니다.

## 재현과 CI

```sh
go test ./internal/search -run '^$' -bench 'Benchmark(Search|Index)Layout' -benchmem -count 5
go build -o bin/routeeval ./cmd/routeeval
./bin/riidolaya setup
./bin/routeeval --stage 1
./bin/routeeval --stage 2
./bin/routeeval --stage 3
```

저장소 루트에서 실행합니다. setup용 riidolaya는 기존 설치본 또는 소스 빌드가 필요합니다. Performance replay Actions도 세 단계 JSON과 검색 benchmark를 아티팩트에 기록합니다. 출력 확률은 플랫폼에 따라 다를 수 있으므로 로컬 숫자를 CI의 정확한 정답으로 강제하지 않습니다. 일반 CI는 fixture 유효성·동작 보존·race를 검사하고, 성능 재생은 실제 추론과 판단 지표를 기록합니다.
