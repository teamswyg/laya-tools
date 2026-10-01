# 작업 기록과 요구사항 충족 여부 확인하기

[English](task-outcomes.en.md)

라우터가 작은 모델을 추천했다는 사실만으로 비용 절감이나 작업 성공을 알 수는
없습니다. 실제 사용량과 결과 파일이 요구사항을 충족했는지를 각각 확인해야
합니다. 이 문서는 이미 있는 실행 기록을 읽는 `riido-taskoutcome`과 결과 파일을
별도로 검사하는 `riido-taskverify`의 사용법을 설명합니다. 두 도구는 모델을
실행하지 않습니다. Codex 연동도 선택 사항입니다.

현재 [공개 작업 후보](../benchmarks/training/public-task-candidates.json)는 **8개·세 코드 가족**입니다.
실제 시도한 범위는 **서로 다른 요청 4개·두 코드 가족·한 저장소, 직접 소유한 기록
누적 12건**입니다. [실험 53](../experiments/task-outcomes/RESULTS-53.ko.md)은 키워드
보호 요청 하나를 Luna/Sol low로 각각 두 번 반복했고, 네 후보 모두 독립 범위
수용·정상 종료·완전한 전체 기본 사용량을 확인했습니다. 네 반복은 네 요청이 아닙니다.
첫 완료·수용 후보의 키워드 보호를 runtime에 반영하고 고정 계약으로 회귀 검사했습니다.

51의 정확한 주석 거절·프로필 지원 오류와 52의 여섯 후보 수용·다섯 실행 완료·
한 시간 초과/사용량 미확정은 역사적 원문 기록으로 보존합니다. 기록 수·후보 수용·
실행 완료는 다른 수치입니다. 공급자가 확인한 실제 모델과 금액·구독 잔량은 미확정입니다.
이 개발 사례들은 현재 학습 라벨과 final에서 제외합니다.

새 `taskoutcome-event-key-bounds`는 versioned 검증기를 준비한 별도 parser 후보이며
**모델 시도 0개**입니다. [공개 Go 원천 후보](public-go-acquisition-53.ko.md) 120개·
8저장소도 **실행 자격 0개**이며 실제 실행 수나 라벨 수로 더하지 않습니다.
아래 예제는 직접 작성한 파서 입력 또는 검증 방법이고 절감이나 영역별 2,400건
최종 평가를 대신하지 않습니다.

## 기존 기록 집계하기

저장소 루트에서 Go 바이너리를 만듭니다. 런타임에 Python은 필요하지 않습니다.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskoutcome ./cmd/riido-taskoutcome
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
```

먼저 공개된 직접 작성 예제로 출력 형식을 확인할 수 있습니다.

```sh
.cache/bin/riido-taskoutcome --input examples/task-outcomes/complete-authored.jsonl
.cache/bin/riido-taskoutcome --input examples/task-outcomes/partial-authored.jsonl
```

| 직접 작성한 입력 | 관측 입력 | 관측 캐시 입력 | 관측 출력 | 전체 사용량 |
| --- | ---: | ---: | ---: | --- |
| `complete-authored.jsonl` | 250 | 110 | 50 | 입력·캐시·출력 합계가 알려져 있음 |
| `partial-authored.jsonl` | 100 | 미확정 | 20 | 두 번째 turn 실패로 미확정 |

첫 예제에는 완료된 turn이 두 개 있습니다. reasoning 출력은 한 turn에서만
5가 관측되어, 관측 합계는 5이고 전체 reasoning 합계는 미확정입니다. 선택
필드인 reasoning의 누락은 알려진 입력·캐시·출력 합계를 지우지 않습니다.

캐시 입력 110은 입력 250에 이미 포함되어 있습니다. reasoning 출력도 출력에
포함됩니다. 이들을 다시 더하지 않으며, 캐시를 뺀 값을 청구 토큰이라고
표시하지도 않습니다. 토큰을 요금·구독 사용량·남은 사용량으로 환산하지 않습니다.

실제 기록은 공개 디렉터리 대신 로컬 비공개 위치에 보관합니다. 다음 경로는
사용법을 위한 예시입니다.

```sh
.cache/bin/riido-taskoutcome --input .cache/private-task-trace.jsonl
.cache/bin/riido-taskoutcome --input - < .cache/private-task-trace.jsonl
```

선택적으로 요청했던 모델과 작업의 짧은 공개 식별자를 붙일 수 있습니다.

```sh
.cache/bin/riido-taskoutcome \
  --input examples/task-outcomes/complete-authored.jsonl \
  --task-label authored-example \
  --requested-model example-model \
  --requested-reasoning low
```

이 식별자는 호출자가 제공한 요청값입니다. `--requested-model`은 실제 사용된
모델의 증거가 아니며, `observed_model`은 `unknown`으로 남습니다. 요청값에
프롬프트·개인정보·경로·토큰을 넣지 마세요. 집계 결과도 실제 비공개 작업에
사용했다면 공개 전에 별도로 검토해야 합니다.

## 집계 결과를 읽을 때 구분할 것

- `turn.completed`는 turn이 끝났다는 기록입니다. 코드가 맞거나 작업이
  완료됐다는 판정은 아닙니다.
- `usage.*.observed_total`은 기록에서 실제로 읽은 합계입니다. `total`은
  전체 구간의 해당 필드가 알려져 있을 때만 숫자로 표시합니다.
- `null` 또는 `unknown`은 미확정입니다. 관측한 **0**과 다릅니다. 캐시 필드만
  없으면 입력·출력 관측값은 보존됩니다.
- `usage_complete`는 모든 turn의 입력·캐시·출력과 깨끗한 종료를 확인했을
  때만 참입니다. 실패·초기 오류·열린 turn·알 수 없는 control 이벤트가 있으면 부분
  관측값을 보존하되 전체 합계를 확정하지 않습니다.
  이는 제공된 기록의 범위이며, 다른 시도나 기록 파일이 누락되지 않았다는
  증거는 아닙니다.
- 실제 모델, 프로세스 종료 상태, 실행 시간, 기록의 생산 출처,
  `task_acceptance`는 이 집계기에서 `unknown`입니다.
- `startup_error_items`는 `thread.started` 뒤 첫 `turn.started` 전에 나온
  `item.completed`의 `error` 항목 수입니다. 이 좁은 초기 진단 형식의 개수만
  기록하고 원문·ID는 버립니다. 오류가 있으면 이후 turn이 끝나도 전체 사용량은
  미확정입니다. 이 필드 자체는 프로필 가용성 오류의 원인을 분류하지 않습니다.

입력 SHA-256은 공백과 개행까지 포함한 정확한 bytes를 연결합니다. 누가
작성했는지, 실제 모델이 실행됐는지 또는 작업을 완료했는지는 증명하지 않습니다.
본 집계기가 모델을 실행하거나 결과의 생산 출처를 검증하지는 않습니다.
그 연결은 별도의 [선택적 실행기](task-execution.ko.md)가 제공합니다.

초기 오류 처리 보완은 실험 51의 두 번째 원본 실행 기록을 바꾸지 않습니다.
그 기록의 `summary_status: parse_failed`는 당시 고정한 집계기의 실제 결과로
유지합니다. 새 집계기로 재해석한다면 원본을 참조하는 별도 버전의 기록으로
남겨야 하며, 없던 사용량이나 작업 성공을 만들어내지 않습니다.

집계기는 메시지·reasoning 원문·도구 명령·파일 본문·로컬 경로·thread/item ID를
출력에 보관하지 않습니다. 기본 한도는 전체 64 MiB, 한 줄 1 MiB, 이벤트
100,000개, turn 4,096개입니다. 잘못된 JSON·중복 주요 키·수치 오버플로·한도
초과는 원문을 포함하지 않는 오류 코드로 거절합니다.

## 결과 파일을 별도로 검증하기

아래 네 계약에는 실제 모델 시도가 있습니다. `riido-taskverify`는 해당 **정해진 파일 범위**를 검사하며,
별도 parser의 versioned 검증기를 준비했습니다. 명세·독립 검사는 버전으로 관리하며 모델 시도는 0개입니다.
먼저 명세를 읽습니다. 이 명령은 검사나 모델 실행 없이 JSON 명세를 출력합니다.

```sh
.cache/bin/riido-taskverify --task catalog-min-context --spec
.cache/bin/riido-taskverify --task repo-keyword-language-guard --spec
```

| 작업 ID | 검사하는 요구사항 | 파일 범위와 검사 방식 |
| --- | --- | --- |
| `comment-budget-period` | 예산 기간 설명 주석을 명세 문장으로 정확히 교체 | 고정 `go.mod`, catalog 소스·테스트 3개 파일. 정확한 변경·gofmt와 고정 테스트 실행 |
| `comment-preview-authority` | 저장소 추천이 실행 승인을 뜻하지 않는다는 주석으로 정확히 교체 | `pkg/reporouter/router.go` 1개 파일. 정확한 변경·gofmt 정적 검사 |
| `catalog-min-context` | 선택적 `MinContext` 필드, 음수 거절, 최소 컨텍스트 조건과 기존 동작 보존 | catalog·planner·switchpolicy와 예제를 포함한 고정 9개 파일. 독립 경계값 검사와 고정 catalog·planner 테스트 실행 |
| `repo-keyword-language-guard` | 선택된 후보의 키워드 중 하나라도 비라틴 글자가 있으면 Judge를 부르지 않고 기존 후보 결과와 보류 이유를 보존 | 별도 기준의 고정 `go.mod`, reporouter 소스·테스트 3개 파일. 독립 계약과 고정 테스트 실행; 53의 같은 요청 네 시도 모두 수용 |

파일 범위는 명세의 작업을 검사하기 위한 closure입니다. 범위 밖 파일은
**미검사**이며, `accepted`가 저장소 전체·모든 요구사항의 안전성을 뜻하지
않습니다. 변경 없는 base는 요구한 작업을 수행한 후보로 인정하지 않습니다.
후보가 직접 작성한 테스트의 성공 주장도 승인 근거로 사용하지 않습니다.

원래 세 작업의 공개 기준 파일은 `internal/taskverify/testdata/base`에 있습니다. 기준 revision은
`6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`이며, 검증기는 필요한 각 파일의
SHA-256을 확인합니다. 새 키워드 작업은 공개 리비전
`146b02b9c37e6d90a11386050ccfdffc036c113e`의 별도 기준
`internal/taskverify/testdata/repo-keyword-language-guard-v1`을 사용합니다.
버전이 붙은 작업 정의는 소스·명세·독립 계약을 묶고 LICENSE·NOTICE도 별도로
고정합니다. 기존 세 작업의 명세와 해시를 새 작업에 맞춰 바꾸지 않습니다.
후보 디렉터리는 선택한 작업의 같은 기준에서 작업을 반영한 파일을
준비해야 합니다. 예를 들어 다음 명령은 **이미 준비한** 후보를 검사합니다.

```sh
.cache/bin/riido-taskverify \
  --task catalog-min-context \
  --base-dir internal/taskverify/testdata/base \
  --candidate-dir .cache/task-candidates/catalog-min-context \
  --go-root /absolute/trusted/go-toolchain
```

위 후보 디렉터리를 만들어 주거나 작업을 대신 수행하는 명령은 아닙니다.
다른 두 작업도 `--task`와 후보 디렉터리를 해당 작업으로 바꾸어 사용할 수
있습니다. `--help`에서 간단한 사용법을 확인할 수 있습니다.
새 키워드 작업에는 `--task repo-keyword-language-guard`와 함께 위의 별도
`--base-dir` 및 그 기준에서 준비한 후보를 사용해야 합니다. `--spec`은 현재
작업의 기준 리비전과 고정값을 확인하는 방법이며 모델을 실행하지 않습니다.
`--go-root`에는 `bin/go`를 포함한 신뢰할 Go 1.27.1 설치의 절대 경로를 지정합니다.
`-trimpath`로 만든 바이너리에 기본 Go 설치 경로가 없으면, 동작 검사가 필요한
작업에는 이 옵션이 필요합니다. 설치 경로나 격리된 Go 실행을 확인할 수 없으면
작업 실패로 단정하지 않고 `verifier_unknown`으로 남깁니다. 후보 코드를 실행하지
않는 주석 정적 검사와는 구분됩니다.

동작 테스트가 필요한 세 작업은 macOS의 `sandbox-exec` 격리를 사용할 수 있을
때만 테스트합니다. 정해진 공개 소스와 고정 테스트를 임시 디렉터리에 넣고,
네트워크와 호스트의 인증 환경을 사용하지 않도록 구성합니다. 격리를 사용할 수
없거나 전체 검증의 시간 제한·중단·출력 한도로 검증을 마치지 못하면
`verifier_unknown`입니다. 실제 독립 테스트의 요구사항 실패는 후보 거절로
처리합니다.
현재 다른 OS에서는 이 세 작업의 동작 검증이 미확정입니다.
`comment-preview-authority`는 코드를 실행하지 않는 정적 검사라 이 제한과
구분됩니다. 이 도구는 임의의 저장소나 임의의 코드 실행을 승인하는 범용
보안 검증기가 아닙니다.

검증 JSON에는 작업 명세, base, 후보 파일의 hash와 검사 결과가 포함됩니다.
`independent_tests`는 통과한 Go 테스트·하위 테스트 이벤트 수이며, 독립된
작업 요청의 표본 수가 아닙니다.
`accepted: false`만 보고 실패를 단정하지 말고 `status`와 각 검사 코드를 함께
읽어야 합니다. `verifier_unknown`은 검증 환경의 미확정을 뜻할 수 있습니다.
지원하는 후보 코드 구조를 벗어나는 경우도 `unsupported_candidate_shape`로
검증 미확정을 표시합니다. 예를 들어 허용한 기본 import 범위를 넘어선 구현은
이 좁은 검증기의 범위 밖일 수 있으며, 그 자체로 구현이 요구사항을 충족하지
못한다는 판정은 아닙니다.

새 키워드 작업은 import 변경·`init` 함수·컴파일러 지시문도 지원 범위 밖으로
처리합니다. 후보가 작성한 테스트 대신 고정한 독립 계약의 지정 테스트들이
정확한 패키지에서 실제 통과했는지 확인합니다. 소스 수용 범위와 LICENSE·NOTICE의
배치 해시도 구분하며, 이것이 후보의 범위 밖 파일까지 검사했다는 뜻은 아닙니다.

| 명령 | 종료 코드 | 뜻 |
| --- | ---: | --- |
| `riido-taskoutcome` | 0 | 집계 또는 도움말 출력 성공. 작업 성공을 뜻하지 않음 |
| `riido-taskoutcome` | 1 | 입력·메타데이터 검증·파싱·한도·출력 오류 |
| `riido-taskoutcome` | 2 | 알 수 없는 옵션 등 명령 인자 문법 오류 |
| `riido-taskverify` 검증 실행 | 0 | 정해진 closure의 요구사항 충족 |
| `riido-taskverify` 검증 실행 | 1 | 후보 거절 |
| `riido-taskverify` 검증 실행 | 2 | 잘못된 인자·기준 파일·출력 등 호출 오류 |
| `riido-taskverify` 검증 실행 | 3 | 격리 또는 동작 검증 미확정 |
| `riido-taskverify --spec` 또는 `--help` | 0 | 명세 또는 도움말 출력 성공. 후보 검증 결과가 아님 |

에이전트는 JSON 필드와 종료 코드를 함께 읽어야 합니다. 파싱이 성공했다는
이유로 작업 완료를 기록하거나, `--spec`의 종료 코드로 후보를 승인하지 마세요.

## 실제 실행과 역사적 예제를 구분하기

[실험 50](../experiments/task-outcomes/RESULTS-50.ko.md)은 실행기 패키징 문제로
코딩 모델 본 실행 전에 중단했습니다.
[실험 51](../experiments/task-outcomes/RESULTS-51.ko.md)은 한 주석 작업의 직접
소유한 CLI 시도 두 건을 연결했지만, 첫 후보 거절과 두 번째 프로필 지원 오류·
초기 오류 집계의 호환성 공백으로 나머지 네 항목을 중단했습니다. 첫 시도의
사용량은 완전히 관측했고 두 번째는 미확정입니다. 비교 가능한 두 프로필의
능력 결과나 절감 효과는 확보하지 못했습니다.

[실험 52](../experiments/task-outcomes/RESULTS-52.ko.md)는 사용 가능한 요청 프로필
`gpt-6-sol`·`gpt-6-luna`를 각각 `low`로 지정한 별도 계획의 여섯 시도를 기록했습니다.
세 요청의 양쪽 후보 모두 독립 범위 검사에서 수용했지만, 정상 종료와 전체 기본
사용량은 다섯 시도에서만 확인했습니다. `catalog-min-context`의 Sol 시도는 120초
한도 후에도 후보가 계약을 충족했으므로 구현 실패로 바꾸지 않으며, 사용량은
`unknown`으로 남깁니다. 두 주석 요청에서만 양쪽의 완전한 사용량을 비교할 수
있습니다. 큰 프로필이 언제나 더 좋다는 가정이나 일반적인 절감률을 만들지 않습니다.

51·52만의 역사적 합계는 8기록·3요청입니다. 현재 53까지는 **12기록·4요청**이고
실제 시도한 가족은 두 개·저장소는 하나입니다. 53의 키워드 계약 네 시도 모두
수용·정상 종료·전체 기본 사용량을 확인했습니다. 현재 수용 후보는 52의 여섯 개와
53의 네 개이고, 완전한 기본 사용량 10기록에는 51의 거절된 첫 시도도 포함됩니다.
지원 오류 1건과 시간 초과 1건을 누락하지 않습니다. 공급자가 확인한 실제 모델은
여전히 미확정이며 parser 후보는 모델이 아직 시도하지 않았습니다.

53의 합계는 입력 339,805·그중 캐시 280,576·출력 4,890·그중 reasoning 76입니다.
Luna의 총 입력은 적었지만 캐시를 제외하면 Luna 31,425·Sol 27,804였습니다.
부분집합을 중복 합산하거나 이 차이를 실제 비용 절감으로 해석하지 않습니다.

실제 라우터 평가에는 같은 작업 명세·기준 파일에서 실행할 수 있는 모델들의
독립 수용 결과, 모든 시도·재시도의 사용량과 요청·관측 근거가 더 필요합니다.
완료율과 전체 작업당 사용량을 함께 비교해야 상향·하향 라우팅의 유용성을
판단할 수 있습니다. [골든셋 설계](golden-set-scale.ko.md)의 영역별 최소
2,400개 최종 요청 목표는 이 개발 기록으로 대체하지 않습니다.
구체적인 원천 확보·그룹 분할·비교 실행 예산은
[골든셋 확보 계획](golden-set-acquisition.ko.md)에 설명되어 있습니다.

현재 직접 작성한 JSONL 예제와 공개 코드 검증 예제는 도구의 동작을 설명하는
자료입니다. 실제 모델 작업 결과가 아닙니다. 이 초기 도구 구현의 범위와 검증
계획은 역사적 [계획 49](../experiments/task-outcomes/plan-49.json)에 있습니다.

[직접 작성한 예제의 실제 도구 실행 기록](../experiments/task-outcomes/fixtures-49.json)은
집계 두 건, 정확한 주석 변경의 수용, 변경 없는 기준 파일의 거절을 담습니다.
여기서 도구 종료 코드는 모델 프로세스의 종료 코드가 아니며, **그 역사적 예제
기록의 모델 결과는 0건**입니다. 이후 실험 51·52·53의 실제 시도 기록과 구분합니다.

원천 수집과 실행 계약은 [공개 Go 후보 120개 조사](public-go-acquisition-53.ko.md)에 별도로 설명합니다.
