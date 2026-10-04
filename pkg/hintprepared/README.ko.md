# 요청별 특징을 한 번 준비하는 Go API

[English](README.en.md) · [측정 전체](evidence/READBACK.public.v1.json)

같은 요청과 후보에 여러 작은 가중치 모델을 적용하거나 반복 점수 계산할 때,
특징을 한 번만 만들어 재사용하는 선택형 API입니다. 현재 CLI·기본 라우팅
정책은 그대로입니다. 결과는 검증 순서에 대한 힌트이며 확률·정답·행동 승인이 아닙니다.
모든 후보를 남깁니다. 기존 품질 실패 모델을 기본 활성화하지 않습니다.

현재 Prepare는 생성자 안에서 특징 개수를 먼저 세고 임시 배열 하나를 재사용합니다.
[공개 Chi 입력 비교](../../experiments/short-claim/next-cohort-chi-audit/constructor-scratch/README.ko.md)에서
준비+순위 구간의 Go 누적 할당 바이트는 약37% 감소했고, 쌍별 시간 중앙값은
약7.5–7.7% 증가했습니다. 할당 횟수는 늘었고 최종 보유 배열은 동일합니다.
메모리 우선의 사전 기준을 통과한 변경이며 속도나 모델 정확도 개선은 아닙니다.

사용 순서는 다음과 같습니다. `rawWeights`는 호출자가 명시적으로 읽은
RIIDOH01/8192 연구 가중치입니다. Laya 원본 ONNX 모델과 다른 형식이며,
형식이 맞는다는 사실만으로 특징 호환성·품질·라이선스 적합성이 보장되지 않습니다.

```go
current, err := shortclaim.ValidateInput(input) // 기존 제품 입력 검증
if err != nil { return err }
owner, err := hintprepared.Prepare(current.Prepared())
if err != nil { return err }
view, err := hintweights.New(rawWeights)
if err != nil { return err }
ranking, err := owner.Rank(current, view)
if err != nil { return err }
// ranking.Order[:ranking.Count]는 현재 입력의 후보 index입니다.
// ID는 current.Prepared().Candidates[index].ID에서 읽습니다.
```

코드는 함수 내부의 사용 조각이며 위 세 패키지를 import합니다. 다운로드나
Codex 등록은 자동으로 하지 않습니다. 잘못된 입력은 기존 고정 오류 그대로
거절합니다. 요청 원문·후보 수·순서별 텍스트가 바뀌면 `ErrMismatch`로 거절하여
호출자가 명시적으로 다시 준비합니다. 유효한 입력의 baseline fallback도
호출자가 선택합니다. ID/provenance만 바뀌면 재사용 가능하고 모델을 바꾸면
새 점수를 계산합니다. 같은 점수는 현재 입력 순서를 따릅니다.

준비본과 View는 읽기 전용으로 공유하고 결과는 호출마다 고정 배열 값으로
반환합니다. 공유 결과 버퍼·캐시·잠금이 없습니다. 복사한 입력이나 결과를
수정해도 준비본을 바꾸지 않습니다. 내부 Features가 학습 코드와 같은 Go
package에 있어 build 의존성은 남지만 runtime에서 학습 함수는 호출하지 않습니다.

아래 표는 위 scratch 변경 전 Prepare를 사용한 과거 합성 실험의 기록입니다.
현재 생성자의 성능 수치로 읽지 마세요. 당시 Go1.27.1/darwin-arm64,
CGO0, GOMAXPROCS1, 각 항목100ms×3회에서
관측한 중앙값입니다. 공개 합성 문장·FP32 숫자 가중치를 사용했고 첫 준비 비용과
준비 후 재사용을 나눴습니다. 원문45행을 손실 없이 gzip 저장했습니다.

| 후보 | 준비 µs / B/op | 참조 full Rank µs / B/op | 준비 후 Rank µs / B/op | BM25 µs / B/op | 어휘 µs / B/op |
|---|---:|---:|---:|---:|---:|
| 1 | 8.857 / 8,200 | 10.115 / 4,624 | 0.154 / 0 | 0.784 / 0 | 0.814 / 0 |
| 5 | 38.279 / 32,152 | 37.918 / 19,856 | 0.541 / 0 | 1.269 / 0 | 1.425 / 0 |
| 8 | 60.050 / 50,040 | 58.700 / 30,664 | 0.831 / 0 | 1.652 / 0 | 1.884 / 0 |

준비 후 Rank·BM25·어휘는 측정된 할당 횟수도0회/op입니다. 한 번만 쓰면 준비
비용이 필요합니다. 같은 후보5개를 반복할 때 시간만 산술적으로 비교하면 BM25
대비 준비 비용을 상쇄하는 데 약53회가 필요하지만 두 방식의 점수·품질은 다릅니다.
참조는 매번 가중치·원문 단어 수 검사와 특징 생성·결과 변환을 수행합니다.
제품 입력 검증은 각 순위 측정의 준비 단계이므로 순수 배열/SIMD/캐시 효과로
해석할 수 없습니다. B/op는 누적 Go 할당이며 보유량·peak·RSS·GPU 값이
아닙니다. 비교 과정에서 참조 계수와 View·입력을 함께 보유했습니다.

```sh
go test -race -count=1 ./pkg/hintprepared
CGO_ENABLED=0 GOMAXPROCS=1 go test -run '^$' -bench '^BenchmarkPreparedAPI' -benchtime=100ms -count=3 -benchmem ./pkg/hintprepared
gzip -dc pkg/hintprepared/evidence/benchmark.stdout.public.txt.gz
```

race 검사에는 C 컴파일러가 필요하며 런타임 API는 순수 Go입니다. 실제 저장소
race 테스트6개·하위검사18개와 vet가 통과했습니다. 첫 stage 비용 실행은 잘못된
선택 이름으로0항목이었고 성공으로 세지 않았습니다. [실패·수정 기록](evidence/STAGE-READBACK.public.v1.json)을 보존했습니다.

학습된 모델 파일·네트워크·Python·새 학습은 이 합성 검사에 필요 없습니다.
실제 API 호출에는 명시적으로 전달한 가중치 View가 필요합니다. 의미 품질, 독립2400 Golden,
전체 작업/LLM/Codex 절감, GPU/MPS 성능은 이 실험의 결과가 아닙니다.
