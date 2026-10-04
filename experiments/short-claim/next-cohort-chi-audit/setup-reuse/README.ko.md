# 같은 입력의 준비 비용을 재사용하기

작은 주장 모델을 여러 번 호출하는 시스템에서는 매번 텍스트 특징을 만들지 않고 준비된 입력을 재사용할 수 있습니다. 기존 `hintprepared` API의 이 동작을 실제 Go 실행으로 측정했습니다. 코드는 Go 1.27.1이고 Codex 연동 없이도 실행합니다.

| 동일 입력 호출 횟수 | 준비 1회 + N회 점수 계산 중앙값 | 매번 준비 중앙값 | BM25 중앙값 | 준비 재사용 / 재생성 누적 Go 할당 |
|---:|---:|---:|---:|---:|
| 1 | 1.845ms | 2.059ms | 0.0156ms | 445,848 / 445,848B |
| 2 | 1.370ms | 2.805ms | 0.0154ms | 445,848 / 891,696B |
| 4 | 1.140ms | 4.525ms | 0.0244ms | 445,848 / 1,783,392B |
| 8 | 1.099ms | 8.158ms | 0.0434ms | 445,848 / 3,566,784B |

N=8의 같은 순서 블록 비교에서 재생성/재사용 시간 비율은 7.219~7.578배였습니다. N=1은 같은 동작인데도 0.843~1.581배로 흔들립니다. 한 로컬 프로세스의 노출된 개발 사례이며, N이 늘수록 표의 재사용 시간이 감소하는 현상을 보편적 특성으로 해석하면 안 됩니다. N별 여섯 관측의 범위는 [전체 요약](SUMMARY.actual.public.v1.json)에 있습니다.

준비를 재사용하면 반복 비용이 줄지만, 여기서는 BM25가 더 빠릅니다. 기존 모델은 정답 후보를 여전히 마지막에 놓습니다. 이전 [실제 전체 작업](../cost-pilot/README.ko.md)에서는 모델의 검증 호출이198회, BM25는107회였습니다. 이번 준비 비용 개선으로 후보 순서 품질이나 LLM 토큰 절감이 증명되지는 않습니다. 기존 모델은 계속 비활성 상태입니다.

## 직접 사용

저장된 결과의 계약 검사는 모델 다운로드 없이 가능합니다. 압축을 푼 임시 파일에 대해 실행하고 삭제합니다.

```sh
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
gzip -dc experiments/short-claim/next-cohort-chi-audit/setup-reuse/OBSERVATIONS.actual.public.v1.json.gz > "$tmp"
go run ./experiments/short-claim/next-cohort-chi-audit/setup-reuse -saved "$tmp"
```

새 측정은 명시적인 모델 경로가 필요합니다. [동결 명세](FREEZE.public.v1.json)에 기록된 [HF revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46)의 기존 32,792B RIIDOH01 FP32 모델을 사용해야 합니다. `model.hbin`은 예시 파일명입니다.

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/setup-reuse \
  -input experiments/short-claim/next-cohort-chi-audit/preview/INPUT.public.v3.json \
  -model model.hbin
```

이 측정기는 Laya 인코더가 아닌 Go 해시 특징 점수 모델을 호출합니다. 새 실행은 기존 결과를 덮어쓰지 말고 별도 경로에 보관합니다. 사람과 에이전트 모두 `-saved`와 실제 측정을 구분해서 사용할 수 있습니다. API 사용자는 같은 요청·후보 텍스트에 `Prepare`를 한 번 호출한 뒤 `Rank`를 반복합니다. ID·출처만 변경할 수 있으며 요청·후보 텍스트가 바뀌면 새로 준비해야 합니다. 점수는 확률이나 사실 판정이 아닙니다.

## 프로파일에서 확인한 다음 과제

프로파일은 시간 측정과 분리한 네 프로세스에서 수집했습니다. 준비는128회, 재사용 점수 계산은262,144회씩 반복했습니다. Go 누적 할당의 대부분은 `Features` 임시 특징과 `Prepare`의 최종 배열에 잡혔습니다. 점수 계산 CPU 표본은 가중치 읽기·변환에 집중됐습니다. [공개 집계](PROFILES-AGGREGATE.actual.public.v1.json)만 올리고 원본 프로파일은 로컬에 보관했습니다.

다음 실험은 임시 특징 배열과 최종 배열의 동시 생존을 줄이거나, 가중치를 미리 해독하는 선택 옵션을 비교합니다. 배열/SoA 또는 SIMD를 채택하려면 점수 bit·순서·오류 계약을 그대로 보존하고 준비 비용·유지 메모리를 함께 측정해야 합니다. 현재 준비 API는 호출자 소유의 불변 배열이며 공용 캐시와 잠금을 추가하지 않습니다.

72개 구간과524,548회 프로파일 점수 호출은 같은 사례의 반복입니다. 독립 Golden 사례가 늘어난 것이 아닙니다. 학습·라벨·역할 변경은0이며 보호된2400개 평가는 읽지 않았습니다. `TotalAlloc`과 pprof `alloc_space`는 누적 Go 할당량이고 RSS/GPU 메모리가 아닙니다. 96MiB는 Go GC의 소프트 목표입니다. 짧은 준비 CPU 표본은 정밀한 병목 비율을 판단하기에 부족합니다.

원시 입력·모델 바이트, 검증된 입력, 바이트를 소유하는 View가 BM25 구간에서도 함께 살아 있습니다. 최초 사용 비용 비교가 아닙니다. CPU 프로파일은 초기 준비를 제외하고, 할당 프로파일은 프로세스 준비까지 포함하며 차감하지 않았습니다. 반복문 안의 결과 동일성 검사도 프로파일에 포함됩니다.

[영문](README.en.md) · [전체 원자료](OBSERVATIONS.actual.public.v1.json.gz) · [실행 기록](EXECUTION.actual.public.v1.json) · [문구 품질 검토](QUALITY-CAPTION-REVIEW.public.v1.md) · [기존 소스·라이선스 고지](../NOTICE)
