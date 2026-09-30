# Laya 생태계에서 가져온 정책 모듈

한국어 · [English](ecosystem.en.md)

[laya.tools](https://laya.tools/)는 Nielogiczny가 운영하는 독립 커뮤니티 디렉터리입니다. 프로젝트를 찾는 데 도움을 받았습니다. 이 저장소가 해당 사이트나 Laya 제작사의 공식 제품이라는 뜻은 아닙니다.

2026-09-30에 [라우팅](https://laya.tools/laya-for-routing), [코딩 도구](https://laya.tools/laya-for-coding-tools), [검색/RAG](https://laya.tools/laya-for-search-and-rag) 목록을 확인하고, 아래 저장소의 공개 소스와 라이선스를 조사했습니다. 상위 프로젝트의 성능 주장을 우리 도구의 실측 결과로 사용하지 않습니다.

| 프로젝트 | 확인한 내용 | 이번 적용 |
|---|---|---|
| [system-one-router](https://github.com/mmornati/system-one-router/tree/a437d00bca33a4c10038b37a5efe04dc0e3d40bb), Apache-2.0 | Go 게이트웨이. 능력·예상 가격·입력 한도·도구 지원·로컬 제한·부하·예산으로 후보 평가 | `internal/router/score.go`의 선택 구조를 참고·재구현한 `pkg/catalog` |
| [pi-pignon](https://github.com/siiick/pi-pignon/tree/4d97d1a35134b813bac6a8e345a1bf9b0bbf0bdb), MIT | TypeScript Pi 확장. 상향/하향 확신도, 전환 간격, 캐시 비용 회수 계산 | `src/policy.ts`의 정책 일부와 공식을 Go `pkg/switchpolicy`로 이식 |
| [laya_router](https://github.com/glukicov/laya_router/tree/a2278d9690676a6232098f1c0cbaf44e9f727e95), Apache-2.0 | README의 등급별 평가와 질문 문구에 따른 결과 차이 | 평가 설계 참고. 전체 정확도 외에 약한 모델로 잘못 내리는 비율도 따로 측정할 필요. 코드·평가 데이터 미복사 |
| [laya-codex](https://github.com/pilotspace/laya-codex/tree/580bc73c2ed95fd319db93ef725f30bf35047428), Apache-2.0 | Rust 코드 검색. `crates/laya-rank/src/fusion.rs`의 순위 결합(RRF) 구현 | 기존 검색의 비교 실험 후보. 이번에 검색 알고리즘을 바꾸거나 코드 복사하지 않음 |
| [keel](https://github.com/codejunkie99/keel/tree/5cef4569e7f850b3a85dc998cc64c64e944c80a0), MIT | `crates/engine/src/jev_routing.rs`: 허용 후보를 먼저 만들고 새 세션 경계에서 라우팅 | 활성 세션을 임의로 변경하지 않는 경계 설계 참고. 앱/런타임 코드는 미복사 |
| [fast-laya-compaction](https://github.com/ShinyDataTech/fast-laya-compaction) | 압축 관련 후보. 조사 시 GitHub 메타데이터에서 라이선스를 확인하지 못함 | 이식 대상에서 제외. 무손실·절감 주장도 검증하지 않음 |

## 무엇을 가져왔고 무엇을 바꿨나

`catalog`는 먼저 품질·도구·이미지·로컬 실행·컨텍스트 한도를 검사하고, 통과한 모델 중 추정 비용과 부하를 반영한 값이 가장 작은 후보를 고릅니다. 가격은 입력 토큰과 예상 출력 토큰으로 계산합니다. 동점이면 품질, ID 순으로 결정해 결과를 재현할 수 있습니다.

원본 system-one-router의 도메인별 가중 점수 대신, 첫 버전은 사용자가 지정한 단일 품질 점수를 씁니다. 자동으로 측정한 성능이 아닙니다. 원본의 부족한 후보 중 최선 선택과 달리, 품질 기준을 만족하는 후보가 없으면 보류합니다. 낮은 확신도나 분류 보류는 가장 높은 품질 기준을 적용합니다. 입력뿐 아니라 예상 출력도 컨텍스트 한도에 포함합니다.

예산은 **기존 사용액 + 이번 요청 추정액**으로 검사합니다. 예산이 설정되었는데 사용액을 모르면 해당 후보를 제외합니다. 가격 누락도 무료로 간주하지 않습니다. `price: null` 또는 생략은 미상, `price: {"input":0,"output":0}`는 사용자가 명시한 무료 모델입니다. 부하는 비용 비교에만 반영하고 실제 예산 금액에는 가산하지 않습니다.

`switchpolicy`는 pi-pignon의 다음 계산을 옮겼습니다. 각 가격의 단위는 동일한 USD/백만 토큰입니다. 분자와 분모의 공통 단위가 약분되므로 결과는 요청 횟수입니다.

```text
캐시 재구축 추가 비용 = 기존 컨텍스트 토큰 × max(0, max(새 입력 가격, 새 캐시 쓰기 가격) - 새 캐시 읽기 가격)
요청당 절약액 = 기존 컨텍스트 토큰 × (현재 캐시 읽기 가격 - 새 캐시 읽기 가격)
              + 예상 출력 토큰 × (현재 출력 가격 - 새 출력 가격)
비용 회수 요청 수 = 캐시 재구축 추가 비용 / 요청당 절약액
```

원본과 달리 가격이나 전환 이력이 없으면 하향·동급 전환을 보류합니다. 원본의 작은 컨텍스트에서 가격을 몰라도 허용하는 대체 규칙은 가져오지 않았습니다. 동급 전환도 높은 확신도와 절약 근거를 요구합니다. 능력 상향은 확신도 기준을 통과하면 전환 간격과 캐시 경제성 제한을 건너뜁니다. 다만 `planner`가 앞서 검사하는 품질·예산 등 제약까지 무시하지는 않습니다.

새 세션(`current` 없음, `context_tokens: 0`)에는 잃을 캐시가 없으므로 진입 확신도만 확인합니다. 진행 중인 세션의 현재 모델이 없으면 보류합니다. 수동 고정(`pinned`)은 전환을 막지만, 기존 모델이 필수 제약을 위반할 때 그것을 허용하는 근거가 되지는 않습니다.

## Go 모듈과 책임 경계

저장소에는 하나의 `go.mod`를 유지합니다. 그 안의 세 공개 패키지를 외부 Go 프로그램에서 직접 import합니다. 정책 패키지는 표준 라이브러리만 사용하며 CGO, Python, 모델 파일이 필요 없습니다.

- `github.com/teamswyg/laya-tools/pkg/catalog`: 후보 제약 검사와 비용 비교.
- `github.com/teamswyg/laya-tools/pkg/switchpolicy`: 전환 보류와 캐시 비용 회수 판단.
- `github.com/teamswyg/laya-tools/pkg/planner`: 두 정책을 조합한 실행 전 계획.

```go
// cfg와 req는 프로젝트가 관리하는 planner.Config / planner.Request입니다.
plan, err := planner.Build(cfg, req)
if err != nil {
    return err
}
// plan.Status: recommend / hold / blocked
// plan.RecommendedModel: 추천 또는 유지 가능한 모델. blocked일 때 비어 있음.
// plan.Selection.Candidates: 각 후보의 제외 사유와 비용 추정.
```

`recommend`는 추천, `hold`는 현재 모델 유지가 제약을 만족함, `blocked`는 조건을 만족하는 추천을 만들지 못함입니다. 어떤 상태에서도 실행하지 않습니다. 유효한 요청의 `blocked`는 정상 정책 결과이므로 CLI 종료 코드는 0입니다. 입력 오류는 0이 아닙니다. 에이전트는 종료 코드뿐 아니라 `plan.status`를 검사해야 합니다.

`riidolaya plan`은 사람이 입력한 평가값으로도 실행할 수 있습니다. 마지막 인자로 작업 문장을 주면 기존 로컬 Laya 분류기를 사용해 등급·확신도·보류 상태를 덮어씁니다. 필요한 토큰 수, 예산 사용액, 도구 요구 등은 호출자가 제공하며 Laya가 추측하지 않습니다. 한국어·추론 실패·낮은 확신도는 기존 정책대로 보류하며, 보류를 높은 확신도로 바꿔 해석하지 않습니다.

Laya 없는 정책 계획과 Laya 분류를 포함한 계획은 결과 JSON의 `classification` 유무로 구별됩니다. 모델을 실제 실행하는 기존 `codex` 명령에는 이 정책을 자동 연결하지 않았습니다. `serve`/MCP도 현재 검색과 기존 라우터만 제공합니다. riido-daemon은 이 공개 패키지를 import해 상태 저장과 실행 어댑터를 붙일 수 있습니다.

## 사용 예시

저장소 루트에서 실행합니다. 예제 모델 이름·가격·품질·확신도는 모두 **설명용 가상 값**입니다. 실제 공급자 요금표나 정확도 평가가 아닙니다.

```sh
go build -o bin/riidolaya ./cmd/riidolaya

# 모델 파일 없이 순수 정책만 계산
./bin/riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json

# 요청 JSON을 stdin으로 전달
cat examples/planner/request.json | ./bin/riidolaya plan --config examples/planner/config.json --request - --json

# 같은 조건에 실제 로컬 Laya 분류를 추가 (setup으로 모델 설치 필요)
./bin/riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json 'Fix a spelling mistake in a comment'
```

기본 예제는 `example-strong`에서 `example-fast`로 바꾸는 비용 회수 기간을 약 `0.1385`회로 계산합니다. 이는 입력한 가상 가격으로 계산한 값이며 실제 비용 절감 결과가 아닙니다. `prompts_since_switch`를 1로 바꾸면 `cooldown` 사유로 현재 모델을 유지합니다. `local_only: true`를 넣으면 예제의 모든 모델이 원격 후보이므로 `blocked`입니다.

현재 모델의 `id`는 공급자·모델·추론 설정이 같은 실행 프로필을 유일하게 나타내야 합니다. 같은 모델이라도 추론 설정이 다르면 서로 다른 ID를 사용합니다. `rank`는 공급자의 공식 성능 순위가 아니라 사용자가 관리하는 능력 순서입니다.

## 측정과 제한

Apple M4 Pro, Go 1.27, 2026-09-30 로컬 마이크로벤치마크 1회:

| 대상 | 조건 | ns/op | B/op | allocs/op |
|---|---|---:|---:|---:|
| catalog.Select | 가상 후보 2개 | 132.0 | 296 | 6 |
| switchpolicy.Decide | 가상 전환 1개 | 28.92 | 8 | 1 |
| planner.Build | 예제 후보 3개 | 246.9 | 432 | 9 |

```sh
CGO_ENABLED=0 go test ./pkg/...
go test ./pkg/... -run '^$' -bench . -benchmem
```

위 값은 **정책 함수만** 측정했습니다. CLI 시작 시간·JSON 읽기·Laya 로딩/추론·실제 작업 시간·전체 프로세스 RSS를 포함하지 않습니다. Laya 추론을 켜면 기존 약 1.4GiB 규모의 네이티브 모델 메모리가 별도로 필요합니다. 작은 정책 모듈을 추가했다고 모델 메모리가 줄어든 것은 아닙니다.

캐시 계산은 기존 컨텍스트 전체를 재사용하고 같은 크기의 출력을 반복한다는 단순 추정입니다. 캐시 적중률·접두사 변경·신규 입력·지연·재시도·실제 청구량은 포함하지 않습니다. 예산 스냅샷도 동시 요청에 대한 예약이나 누적 장부가 아닙니다. 서비스에서 강제 예산 제한을 하려면 별도 원자적 예약/정산이 필요합니다.

이 API 가격 계산을 ChatGPT/Codex 구독의 포함 사용량이나 잔여 한도로 환산할 수 없습니다. 이식한 정책이 실제 작업 성공률을 유지하며 비용을 줄이는지는 아직 검증하지 않았습니다.

## 출처와 라이선스

이식 대상은 위 표에 고정한 커밋으로 추적합니다. MIT 원문과 저작권은 [pi-pignon.LICENSE](../licenses/pi-pignon.LICENSE), Apache-2.0 원문은 [system-one-router.LICENSE](../licenses/system-one-router.LICENSE)에 보존했습니다. `pkg/switchpolicy`의 이식 부분에는 원저작권과 MIT 고지를 유지하며, 그 외 자체 코드는 저장소의 Apache-2.0을 따릅니다. 변경 내용은 이 문서와 소스 헤더에 기록했습니다.
