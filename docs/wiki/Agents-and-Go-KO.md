# 에이전트와 Go 프로그램 연동

[English](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-EN) · [홈](https://github.com/teamswyg/laya-tools/wiki)

필요한 가장 작은 인터페이스부터 선택하세요. CLI 사용에는 등록이나 배경 서비스가 필요 없습니다.

## 한 번 요청: JSON

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
```

사람용 문장 대신 JSON을 해석합니다. 종료 코드뿐 아니라 preview/status/reason을 검사합니다. stderr와 stdout을 섞지 마세요. 실제 카탈로그와 질문에는 비공개 정보가 있을 수 있으므로 로그도 비공개로 관리합니다.

## 여러 번 요청: 상주 JSONL

검색·모델 추천:

```sh
riidolaya serve --root /path/to/repository
```

```json
{"id":1,"op":"search","query":"redirect authorization","lexical":true}
{"id":2,"op":"route","query":"Fix a typo"}
```

저장소 preview는 별도 프로세스·카탈로그를 사용합니다.

```sh
riidolaya repo-serve --catalog examples/repositories/catalog.json --laya
```

```json
{"id":1,"query":"Fix login sessions"}
{"id":2,"query":"refund invoices"}
```

한 줄에 JSON 하나를 보내면 요청 ID가 포함된 응답 하나를 받습니다. 저장소 응답은 `result` 안에 preview를 담고, 오류는 `error`로 표시합니다. 모델 없이 사용하려면 `--laya`를 뺍니다. 저장소 프로세스는 인덱스를 유지하고 필요한 요청에서 모델을 처음 로딩합니다. 현재 구현에서는 프로세스끼리 모델을 공유하지 않으므로 여러 개를 띄우면 RAM도 중복될 수 있습니다. 끝나면 stdin을 닫거나 자식 프로세스를 종료하세요.

`serve`는 plan이나 repo-preview 요청을 받지 않으므로 프로토콜을 섞지 않습니다. 코드 검색은 모델이 상주해도 요청마다 코드 후보를 다시 구성합니다.

## 선택 사항: MCP

```sh
riidolaya mcp --root /path/to/repository
```

도구는 `search_code`, `route_model`입니다. 저장소 preview와 planner는 현재 MCP 도구가 아닙니다. Codex에 선택적으로 등록하려면:

```sh
codex mcp add riidolaya -- /absolute/path/to/riidolaya mcp --root /absolute/path/to/repository
```

현재 Codex 모델을 바꾸거나 권한 설정을 완화하지 않습니다. 도구의 추천이 작업 실행 허가를 뜻하지 않습니다.

## 공개 Go 패키지 import

저장소에는 Go 모듈 하나가 있습니다. 기존 Go 모듈에서:

```sh
go get github.com/teamswyg/laya-tools@v0.3.0
```

모델 없는 저장소 preview 예제:

```go
package main

import (
    "fmt"
    "log"
    "github.com/teamswyg/laya-tools/pkg/reporouter"
)

func main() {
    idx, err := reporouter.New([]reporouter.Repository{
        {Name: "example/billing", Summary: "Invoices and refunds"},
    })
    if err != nil { log.Fatal(err) }
    result, err := idx.Preview("refund invoices", reporouter.DefaultConfig(), nil)
    if err != nil { log.Fatal(err) }
    fmt.Println(result.Status, result.Suggested)
}
```

예상 출력은 `candidate example/billing`이며 작업 배정이 아닙니다.

| 패키지 | 연결하는 프로그램이 제공할 정보 |
|---|---|
| `pkg/reporouter` | 접근 권한에 따라 거른 카탈로그, 선택적 판단 어댑터 |
| `pkg/catalog` | 능력·품질·가격, 예산 사용액 스냅샷 |
| `pkg/switchpolicy` | 현재/대상 프로필, 확신도, 컨텍스트, 전환 이력 |
| `pkg/planner` | 카탈로그·정책·요청. recommend/hold/blocked 결과 확인 |

이 패키지는 CGO·Python·모델 파일이 필요 없습니다. 네이티브 추론은 내부 패키지에 남아 있습니다. 권한·상태 저장·원자적 예산 예약·실행·결과 평가는 연결하는 프로그램의 책임입니다. riido-daemon에 이미 연결된 것은 아닙니다.

[정책 상세](https://github.com/teamswyg/laya-tools/blob/main/docs/ecosystem.ko.md) · [저장소 선택 상세](https://github.com/teamswyg/laya-tools/blob/main/docs/repository-routing-preview.ko.md)
