# GitHub Actions 성능 재생

한국어 · [English](performance-replay.en.md)

[Performance replay](https://github.com/teamswyg/laya-tools/actions/workflows/performance.yml)에서 **Run workflow**로 실행합니다. 기본 브랜치의 검토된 코드를 권장합니다. 정기 실행이나 PR마다 추가 실행은 하지 않습니다.

현재 base INT8 모델을 표준 Linux x64와 macOS ARM 러너의 CPU에서 실행합니다. 기존 CI에서도 이 두 환경의 실제 추론이 통과했습니다. GitHub 공식 문서상 공개 저장소의 표준 러너 사용은 무료이며, 현재 Linux 16GB·macOS ARM 7GB 메모리는 로컬에서 측정한 약 1.4GiB RSS보다 큽니다. 공유 러너 성능은 변동하므로 실제 실행 보고서를 기준으로 판단하세요. 아티팩트 저장에는 별도 정책이 적용되어 작은 JSON·Markdown만 7일 보관합니다.

- [러너 사양](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [Actions 과금](https://docs.github.com/en/billing/concepts/product-billing/github-actions)

## 무엇을 검사하나요?

| 종류 | 고정 입력과 확인할 결과 |
|---|---|
| 기본 추론 | 초기 로딩, 워밍업 후 5회 추론 시간, Go heap |
| 코드 검색 | 현재 공개 저장소에서 같은 키워드로 lexical과 Laya 비교 |
| 난이도 라우팅 | 쉬움·일반·복잡한 영어 요청 및 한국어 보류 정책 |
| 모델 변경 계획 | 가상 카탈로그로 하향·상향 추천 |
| 저장소 선택 | 기존 가상 저장소 6개·질문 24개의 lexical/Laya 결과 |

각 OS에서 11개 시나리오를 순차 실행합니다. 매 시나리오는 새 프로세스이므로 전체 시간에는 모델 로딩도 포함됩니다. 기본 추론 출력 안의 warm 시간은 로딩을 제외합니다. 저장소 평가의 개별 요청 시간에는 첫 추론이 포함됩니다. 네이티브 추론 실패 후 키워드 fallback만 실행한 결과는 성공으로 처리하지 않습니다. 모델 분류가 기대 난이도를 맞혔다는 정확도 인증은 아닙니다.

여기서 open loop는 **결과를 보고 다음 입력·모델·임계값을 바꾸지 않는 고정 재생**입니다. 요청을 일정 도착률로 밀어 넣는 open-loop 부하 시험은 아닙니다. 자동 튜닝·유료 Codex 호출·무한 재시도는 없습니다. base 체크포인트와 CPU만 포함하며 code 체크포인트·GPU/CoreML·서비스 동시 처리량은 이번 범위 밖입니다.

## 결과 읽기

Actions 실행의 Summary와 `performance-<OS>` 아티팩트를 확인하세요. 표에는 전체 실행 시간, 프로세스 CPU 시간, 최대 RSS가 있습니다. JSON에는 각 명령의 원래 결과(판단 보류·확신도·저장소 평가 집계 등)가 함께 들어갑니다. CPU 시간은 여러 코어의 합이므로 벽시계 시간보다 클 수 있습니다. RSS는 Go와 네이티브 할당을 포함하지만 GPU 메모리 측정치는 아닙니다. pprof 원본은 업로드하지 않습니다.

두 OS는 서로 다른 하드웨어입니다. 절대 수치를 서로의 성능 회귀 기준으로 삼지 마세요. 소수 반복의 개발용 표본이며 실제 사용자 정확도·Codex 절감률·유의미한 p95를 보장하지 않습니다. 현재는 수치를 수집하며 속도 임계값으로 병합을 막지 않습니다.

러너는 한 번에 하나, OS별 최대 20분, 개별 시나리오는 최대 2분입니다. 실패한 시나리오는 기록하고 다음 고정 시나리오를 검사한 뒤 최종 작업을 실패 처리합니다. 전체 시간 초과나 빌드/다운로드 실패라면 보고서가 없을 수도 있으며 Actions 로그를 확인해야 합니다.

## 로컬 재현

저장소 루트에서 실행합니다. Go 1.27.1, C 컴파일러와 CPU 런타임을 사용합니다.

```sh
mkdir -p .cache/performance
go build -trimpath -o bin/riidolaya ./cmd/riidolaya
go build -trimpath -o bin/repoeval ./cmd/repoeval
go build -trimpath -o bin/perfsuite ./cmd/perfsuite
./bin/riidolaya setup
./bin/perfsuite --repeats 1 --output .cache/performance/report.json --summary .cache/performance/summary.md
```

`--repeats`는 1~3이며 모델의 답에 따라 반복 수가 늘어나지 않습니다. 공개·직접 작성한 fixture만 사용하고, 실제 업무 프롬프트와 비공개 저장소를 이 공개 워크플로에 넣지 마세요.
