# 공개 작업을 실제로 실행하고 결과를 연결하기

[English](task-execution.en.md)

`riido-taskrun`은 공개된 작은 코딩 작업 하나를 Codex CLI에 맡기고, **그 실행에서
나온 파일·사용량·종료 상태·독립 검사 결과를 연결하는** Go 실행기입니다.
평소 Codex 실행에 자동으로 끼어들지 않으며, `--execute`를 지정해야 실행합니다.
현재는 관리자가 측정 절차를 확인하는 macOS용 개발 도구입니다.

## 왜 별도 실행기가 필요한가

사용량 기록이 있다고 코드가 맞는 것은 아닙니다. 테스트를 통과한 파일이
있다고 어느 실행에서 만들어졌는지 알 수 있는 것도 아닙니다.
[기존 기록 도구](task-outcomes.ko.md)는 사용량을 읽고 파일을 검사하는 두 부분을
제공했습니다. 새 실행기는 직접 시작한 하나의 프로세스와 그 결과를 연결합니다.

고정한 공개 작업과 기준 파일을 새 작업 공간에 복사하고, 명시한 모델·추론
설정으로 한 번 실행합니다. 프로세스와 자식 프로세스를 정리한 뒤 결과 파일을
보존하고, 모델이 작성한 주장이나 테스트와 별개인 검사로 요구사항을 확인합니다.
실패한 실행도 기록하므로 성공한 첫 시도만 골라 비용이 줄었다고 말할 수 없습니다.

이 연결은 **직접 소유한 CLI 실행 한 번**에 대한 근거입니다. 공급자가 실제로
어떤 모델을 사용했는지, CLI 내부에서 모델을 몇 번 호출했는지, 다른 실행이
누락됐는지까지 증명하지는 않습니다.

## 현재 개발 파일럿과 보존한 역사 기록

[실제 외부 비교55](../experiments/task-outcomes/RESULTS-55.ko.md)는 기존 humanize·UUID
논리 요청을 Sol6/Luna low에 각각 한 번 실행했습니다. 세 후보 수용·한 후보 거절이며
네 실행 모두 정상 종료·전체 사용량·예산 영수증·인증 정리를 확인했습니다. 실제 누적은
**기록20개·고유 요청7개·다섯 실제 시도 가족·세 저장소**입니다. 네 개 버전 실행이나
검사 수를 새 요청으로 더하지 않으며 학습·보호 final·절감 근거로 사용하지 않습니다.

[사전 계획 50](../experiments/task-outcomes/plan-50.json)은 배포형 실행기가 Go 도구
경로를 찾지 못해 코딩 모델 실행 전에 중단했습니다.
[그 결과](../experiments/task-outcomes/RESULTS-50.ko.md)를 보존하며, 수정한 실행기를
이전 계획의 성공 결과로 덮어쓰지 않습니다.

[계획 51](../experiments/task-outcomes/plan-51.json)은 한 주석 작업의 두 시도 후
중단했습니다. 정확한 주석 계약 거절과 요청 프로필 지원 오류, 초기 오류 집계의
호환성 공백을 [그 결과](../experiments/task-outcomes/RESULTS-51.ko.md)에 그대로 남깁니다.

수정한 도구를 사용한 [별도 계획 52](../experiments/task-outcomes/plan-52.json)는
커밋 `d202ff042106ad36209ef4cdee6ca1e1c9e64514`에 먼저 고정했습니다. 같은 공개
개발 작업 3개에서 `gpt-6-sol`과 `gpt-6-luna`를 각각 `low`로 요청했고,
계획한 **Codex CLI 시도 6번을 모두 기록**했습니다. 여섯 후보의 정해진 파일
범위는 독립 검사에서 수용했습니다. 다섯 시도는 정상 종료와 완전한 기본
사용량을 확인했지만, `catalog-min-context`의 Sol 시도는 120초 한도에 걸렸고
사용량은 미확정입니다. 후보 수용과 실행 완료를 합쳐 성공 6건으로 세지 않습니다.
[실험 52 결과](../experiments/task-outcomes/RESULTS-52.ko.md)에 항목별 근거가 있습니다.

계획 52는 한 번에 하나, 작업 실행 120초와 별도 검사 45초를 한도로 했습니다.
완료한 계획 52·53·54와 중단한 계획 50·51은 다시 실행하지 않습니다. 새 실제 실행에는
현재 프로필과 도구를 확인한 별도 계획이 필요합니다.

52의 여섯 시도, 53·54 각각의 네 시도가 내부 LLM 호출 수라는 뜻은 아닙니다. 한 CLI 실행 안에 여러 모델
요청이나 공급자 재시도가 있을 수 있으며, 현재 그 수는 `unknown`입니다.
실행기는 자체 재시도·resume·fallback을 하지 않습니다. 시작한 실패·시간 초과도
파일럿의 실행 수에 포함합니다. 기존 v1 계획은 관리자가 전체 횟수·순서와
미실행 항목을 관리합니다. 새 외부 v2 작업은 [공유 parent ledger](fair-upstream-comparison.ko.md)가
두 child plan의 네 예약과 순서를 함께 제한합니다. 하나의 고정 ledger에 대한
한도이며 호스트·계정 전체 또는 provider 내부 호출의 전역 한도는 아닙니다.

| 작업 ID | 독립 검사할 요구사항 | 검사하는 파일 범위 |
| --- | --- | --- |
| `comment-preview-authority` | 저장소 제안은 실행 권한을 주지 않는다는 정확한 주석 변경 | `pkg/reporouter/router.go` 한 파일, 정확한 변경·gofmt 확인 |
| `comment-budget-period` | 예산 기간을 설명하는 주석을 지정 문장으로 변경 | 기준 파일 3개, 정확한 변경·gofmt·고정 테스트 |
| `catalog-min-context` | 선택적 `MinContext` 추가, 음수 거절, 최소 문맥 조건 적용, 기존 동작 유지 | 기준 파일 9개, 경계·JSON 계약과 고정 catalog/planner 테스트 |

위 세 작업의 공통 기준은 공개 리비전 `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`의
`internal/taskverify/testdata/base`입니다. 파일과 LICENSE·NOTICE, 작업 명세와
독립 검사 코드의 해시를 확인합니다. 파일 범위 밖의 추가 파일이나 런타임
캐시는 `unassessed`입니다. `accepted`는 **해당 요구사항과 범위**의 통과이며
저장소 전체나 임의의 코드에 대한 승인으로 읽으면 안 됩니다.

행동 작업 `repo-keyword-language-guard`는 별도 버전 기준
`146b02b9c37e6d90a11386050ccfdffc036c113e`의
`internal/taskverify/testdata/repo-keyword-language-guard-v1`을 사용합니다.
선택된 저장소 후보의 모든 키워드를 Judge 호출 전에 검사하는 계약입니다.
소스는 `go.mod`, `pkg/reporouter/router.go`, `pkg/reporouter/router_test.go`의 3개
파일이며 LICENSE·NOTICE도 별도로 고정합니다. 기존 세 작업의 명세와 해시는
그대로 보존합니다.

[실험 53](../experiments/task-outcomes/RESULTS-53.ko.md)은 **이 요청 하나를 Luna/Sol
low로 각각 두 번, 총 네 CLI 시도**했습니다. 네 후보 모두 독립 범위 수용·정상
종료·완전한 전체 기본 사용량을 확인했습니다. 네 반복은 고유 요청 네 개가
아닙니다. 네 시도에서 같은 mutable 소스 변경이 관측됐고, 첫 완료·수용 후보의
키워드 보호 로직을 runtime에 반영해 고정 독립 계약으로 회귀 검사했습니다.
후보가 작성한 테스트는 수용 근거로 가져오지 않았습니다. 참고 구현·오답으로
검증기를 검사하는 일은 실제 모델 결과와 구분합니다.

51·52만의 역사적 합계는 8기록·3요청이고, 53까지는 12기록·4요청·두 실제 시도 가족입니다.
52의 수용 후보 여섯 개와 53의 네 개, 51의 거절·지원 오류를 원문 그대로 보존합니다.

[실험 54](../experiments/task-outcomes/RESULTS-54.ko.md)는
[사전 계획 54](../experiments/task-outcomes/plan-54.json)에 따라
`taskoutcome-event-key-bounds` **새 요청 하나를 Sol6/Luna low로 각각 두 번** 실행했습니다.
네 시도 모두 독립 검사 75개·정상 종료·완전한 전체 기본 사용량을 확인했습니다.
75개 검사와 네 반복은 고유 요청 하나의 검증·반복 수입니다. 54까지의 역사적 누적은
**직접 소유한 기록 16개·서로 다른 요청 5개·실제 시도한 세 코드 가족·한 저장소**입니다.

로컬 후보 registry는 기존 **8개·세 코드 가족·한 저장소** 그대로입니다. 파서의 별도 기준
`ee72334166e2962b0821d5198a50fdd13f92ab29`에 명세·공개 소스 3파일·LICENSE/NOTICE와
독립 검사를 묶었습니다. event 최상위 키의 중복·64개 개수·128 UTF-8 byte 경계를
검사하며, 후보가 작성한 테스트는 수용 근거로 쓰지 않았습니다. 첫 수용·정상 종료
후보의 고정 배열 구현을 runtime에 반영했습니다. 기존 봉인된 명세와 51~53의 기록은
변경하지 않았습니다.

[공개 Go 원천 목록](public-go-acquisition-53.ko.md)의 **53 시점 역사적 snapshot**은
120개·8저장소·실행 적격 0개입니다. 이 목록은 그대로 보존합니다.
54에서는 그중 외부 두 요청의 [독립 검증 계약](public-go-contracts-54.ko.md)을
로컬에서 준비·검사했지만 **외부 모델 실행·학습 라벨·final 적격은 당시 0개**였습니다.
컴파일된 검사 작업은 기존 실제 시도 작업 다섯 개와 외부 두 개, 합계 **7개 작업·3개 저장소**이며,
로컬 후보 registry 8개와 별도로 셉니다. 계약 준비와 대조군 검사는 실행 수·정답 수나
[2,400개 최종 평가](golden-set-scale.ko.md)를 대신하지 않습니다.

55에서는 외부 두 논리요청을 그대로 연결한 v2 평가 버전을 추가했습니다.
전체 조건을 실제 stdin에 공개하고 actor와 독립 검사의 원본 `go.mod` 언어를
일치시켰습니다. registry ID는 두 개 늘지만 고유 검사 작업은 여전히 7개입니다.
후속 실제 비교는 두 기존 외부 요청의 네 실행을 기록했습니다. 학습 라벨·최종 적격은 계속0개입니다.
[실제 결과55](../experiments/task-outcomes/RESULTS-55.ko.md)와 [공정한 비교 조건](fair-upstream-comparison.ko.md)에
원본 언어와 Go 1.27.1 실행 파일, 범위 밖 형태, 라이선스 검사 범위가 있습니다.

v2 실행에는 기존 인자와 함께 `--parent-file`, `--parent-sha256`,
`--budget-dir`, `--global-ordinal`이 필요합니다. 모델 시작 전에 슬롯을
영구 소비하고 실패에도 돌려주지 않습니다. 시작·종료 파일과 정리 결과를
확인한 뒤에만 다음 슬롯을 열며, 불명확한 시작·종료는 ledger를 중단합니다.
ledger는 모든 실행 디렉터리 밖에 둡니다. 이 안내는 실제 실행 계획이 아니므로,
실행할 계획과 해시를 별도로 먼저 봉인합니다.

```sh
riido-taskrun --budget-status \
  --parent-file /absolute/public/parent.json \
  --parent-sha256 PARENT_SHA256 \
  --budget-dir /absolute/private/shared-ledger
```

이 조회는 기존 ledger만 읽으며 모델 실행·복구·예약을 하지 않습니다.
`reservations`, `durable_start_markers`, `terminal_receipts`,
`launch_state_unknown`을 따로 표시합니다. CLI는 시작 뒤 ledger 종료 증거를
확인하지 못하면 진단 `parent_budget_finalization_failed`와 exit 3을 반환합니다.
모델 실행 JSON은 먼저 쓴 불변 자료이고 별도 terminal receipt가 그 바이트를
연결하므로 JSON의 accepted만으로 parent 완료를 판단하지 않습니다.

## 실제 작업 결과를 보기 전 Laya가 예측한 것

[미리 기록한 예측](../experiments/task-outcomes/routing-predictions-50.json)은
기존 고정 Laya INT8 모델을 CPU, 한 스레드, 기본 임계값 `0.9`로 실행한 결과입니다.
임계값을 바꾸거나 이 세 작업으로 학습하지 않았습니다. 계획 52는 이 예측을
변경 없이 참조하며, 새 예측이나 새로운 독립 작업 세 개로 세지 않습니다.

| 작업 | 작성자의 예상 난이도 | Laya의 제안 | 가장 큰 확률 | 적용 tier |
| --- | --- | --- | ---: | --- |
| `comment-preview-authority` | fast | standard | 0.480 | strong, 보류 |
| `comment-budget-period` | fast | fast | 0.486 | strong, 보류 |
| `catalog-min-context` | standard | standard | 0.591 | strong, 보류 |

세 요청 모두 잘리지 않았지만 신뢰도가 `0.9`보다 낮아 보류했습니다. 제안과
작성자의 예상이 일치하는지보다 **실제 작업을 어떤 프로필이 성공시키는지**가
중요합니다. 파일럿은 미리 정한 luna/sol 비교이며, 보류 뒤의 strong 모델을
실행하는 라우터 실사용 시험은 아닙니다. `gpt-6-astra`는 실행하지 않았습니다.
과거 예측의 standard 설정은 `gpt-6.1-sol`이었으므로, 이를 이번 `gpt-6-sol`의
예측으로 바꾸어 읽지도 않습니다. 이 확률은 보정된 성공 확률이 아닙니다.

50에서 관측하고 52가 참조한 역사 기록에서, 세 번의 차가운 프로세스의 최대 RSS는 약 **1.48 GB
(1.38 GiB)**였습니다. 모델·native session을 새로 불러오는 비용이 포함된
각 한 번의 관측이며, 반복 요청의 지연이나 Go heap·GPU 메모리가 아닙니다.
현재 수치는 아주 작은 메모리라는 목표를 달성했다는 근거가 아닙니다.

53에서는 코딩 결과 전에 [별도 예측](../experiments/task-outcomes/routing-predictions-53.json)을
고정했습니다. INT8 base·CPU 한 스레드·임계값 0.9에서 standard를 제안했지만,
최대 점수 약 **0.669**로 보류했습니다. 설정상 fallback은 strong/Astra이고 실제
Astra 코딩 실행은 없습니다. 새 standard/Sol6 매핑은 이 새 예측에만 적용합니다.
한 번의 cold 최대 RSS는 약 **1.50 GB/1.40 GiB**였으며 warm 지연·Go heap·GPU
메모리나 보정된 코딩 성공 확률을 측정한 것이 아닙니다.

54의 [사전 예측](../experiments/task-outcomes/routing-predictions-54.json)은
같은 INT8 base·CPU 한 스레드·임계값 0.9에서 standard를 제안했지만
**0.610351**로 보류했습니다. strong/Astra fallback은 설정이며 실제 Astra 실행은 없습니다.
한 번의 cold 전체 시간은 **1,183ms**, 최대 RSS는 **1,508,032,512 bytes**
(약 1.508 GB/1.40 GiB)였습니다. native 로딩을 포함한 프로세스 관측이고,
warm 지연·Go heap·GPU 자원이나 보정된 성공 확률이 아닙니다. 초저자원 목표 달성 근거가 아닙니다.

## 사용하기

저장소 루트에서 Go 바이너리를 만듭니다. 실행기에 Python은 필요하지 않습니다.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskrun ./cmd/riido-taskrun
.cache/bin/riido-taskrun --help
.cache/bin/riido-taskrun --task comment-preview-authority --spec
.cache/bin/riido-taskrun --task repo-keyword-language-guard --spec
.cache/bin/riido-taskrun --task taskoutcome-event-key-bounds --spec
```

`--help`와 `--spec`은 모델을 실행하지 않습니다. 기본 호출도 실행을 거절합니다.
실제 실행 전에는 사전 계획, 공개 기준 파일, 고정한 Codex 실행 파일과 버전,
신뢰할 Go 설치, 기존 로컬 ChatGPT 로그인이 준비되어 있어야 합니다. 현재 지원 버전은
`codex-cli 0.158.0`이며, 실행 파일 해시가 계획과 다르면 추론 전에 거절합니다.
버전이나 계획을 바꾼 경우에는 새 계획을 먼저 정해야 합니다.

`-trimpath`로 만든 배포형 바이너리는 기본 Go 설치 경로가 없을 수 있으므로
`--go-root`로 신뢰할 Go 1.27.1 설치 디렉터리를 명시합니다. 실행기는 `bin/go`의
실제 경로·버전·bytes의 SHA-256을 검사하고 새 계획의 고정값과 비교합니다.
기본 경로가 없는데 이 옵션도 주지 않으면 `trusted_toolchain_unavailable`로
실행 전에 거절합니다. 호스트의 PATH를 임의로 따라가거나 Go를 새로 내려받지 않습니다.

다음은 **새 사전 계획을 준비할 때 사용할 인자 형식**입니다. `NEW_*`, `VERIFIED_*`,
`PLANNED_ORDINAL`, `/absolute/...`는 자리 표시자이며 실행 가능한 고정 계획이
아닙니다. 명세와 기준 파일은 선택한 작업에 맞추고, 해시는 실제로 확인한 값으로
채워 새 계획을 먼저 고정하세요. 아래 `gpt-6-luna` 요청도 새 계획의 항목과
일치해야 합니다. 비공개 출력 디렉터리는 아직 존재하지 않아야 합니다. 실제
실행은 계정 사용량을 소비할 수 있습니다.
출력 디렉터리는 OS 임시 디렉터리 아래에 두어야 하며, 상위에 저장소나
`AGENTS.md`·`.codex`·`.agents`가 있으면 실행 전에 거절합니다.

```sh
.cache/bin/riido-taskrun \
  --execute \
  --task comment-preview-authority \
  --model gpt-6-luna \
  --reasoning low \
  --base-dir /absolute/laya-tools/internal/taskverify/testdata/base \
  --private-dir /private/tmp/riido-NEW-ATTEMPT \
  --codex-bin /absolute/trusted/codex \
  --codex-sha256 VERIFIED_CODEX_SHA256 \
  --codex-version 'codex-cli 0.158.0' \
  --go-root /absolute/trusted/go-toolchain \
  --plan-file /absolute/private/NEW-FROZEN-PLAN.json \
  --plan-sha256 NEW_PLAN_SHA256 \
  --attempt-ordinal PLANNED_ORDINAL \
  --task-spec-sha256 VERIFIED_TASK_SPEC_SHA256 \
  --auth-source-dir /absolute/private/codex-login \
  --timeout 120s
```

`--auth-source-dir`은 기존 ChatGPT `auth.json`이 있는 디렉터리입니다.
토큰을 인자나 채팅에 붙여 넣을 필요는 없습니다. API 키 인증은 거절하고,
필요한 인증 파일만 복사합니다. 설정·rules·skills·MCP·기록이나 관련 없는
환경 변수는 물려받지 않습니다. 복사한 인증은 모델의 작업 공간 밖에 두고
종료 시 정리합니다.
공유 백그라운드 서버, 앱·hooks·다른 에이전트·memories·자동 스킬 설치·웹 검색도
이 측정 실행에서는 사용하지 않도록 고정합니다.

새 작업 공간에는 대상 공개 파일만 배치합니다. macOS 권한 프로필로 작업
공간 쓰기와 필요한 도구 읽기를 허용하며, 모델이 만드는 명령의 네트워크와
형제 디렉터리 접근은 거절합니다. 실제 사전 확인 실행에서 이 제한을 검증할
수 없으면 모델을 시작하지 않습니다. 이는 Codex CLI 자체의 공식 서비스
통신과 별개의 제한입니다. 평소 쓰는 체크아웃을 수정하지 않습니다.
명령 실행에는 확인한 같은 Go 설치와 제한된 환경을 사용합니다. Go 도구·모듈의
추가 다운로드와 CGO는 끄고, 독립 검사에도 같은 설치 경로를 전달합니다.

stdout JSONL·stderr·후보 파일·`record.json`은 로컬 비공개 디렉터리에 남습니다.
GitHub나 Hugging Face로 자동 전송하지 않습니다. 공개 기록에는 이 공개 작업의
점검된 수치·열거값·해시만 넣습니다. 원문 기록, 인증, 로컬 절대 경로를
커밋하지 마세요.

## 사람과 에이전트가 결과를 읽을 때

| 확인할 항목 | 의미 |
| --- | --- |
| `provenance` | `executor_owned_single_attempt`는 이 실행기가 직접 시작하고 회수한 한 번의 기록 |
| `applied_request` | 실제로 자식 CLI에 전달한 요청. 공급자가 그 모델을 사용했다는 증거와는 별개 |
| `observed_model` | 현재는 `unknown`. 요청 모델과 같은 값으로 채우지 않음 |
| `process_status`·`exit_code` | CLI 종료 상태. 코드가 맞는지와는 별개 |
| `go_version`·`go_binary_sha256` | 실행 전에 확인하고 계획과 비교한 로컬 Go 도구의 버전과 실행 파일 해시 |
| `verification_status` | `accepted`, `rejected`, `verifier_unknown` 등의 독립 검사 결과 |
| `usage_summary` | 유효하게 읽은 사용량. 실패 시 읽은 부분을 보존하고, 전체 미확정을 0으로 만들지 않음 |
| `whole_attempt_usage_complete` | 정상 종료와 프로세스 그룹 정리, 두 출력의 완전한 회수, 모든 turn의 기본 사용량이 갖춰졌는지 |

후보가 요구사항을 충족해도 사용량은 불명일 수 있습니다. 그 후보를 구현 실패로
분류하지 않습니다. 반대로 CLI가 0으로 종료하고 사용량이 갖춰져도 독립 검사에서
실패했다면 작업 성공으로 보지 않습니다.

입력·캐시 입력·출력 사용량은 따로 표시합니다. 캐시 입력은 입력에, reasoning은
출력에 포함되므로 다시 더하지 않습니다. 토큰을 금액, 계약 사용량이나 잔량으로
환산하지 않습니다. 기록의 wall time은 로컬 CLI 프로세스 시간이며 별도 검사의
시간이나 원격 모델의 GPU 사용량이 아닙니다.

stdout은 64 MiB, stderr는 1 MiB로 제한합니다. 시간·회수 한도·파싱 실패·
검증 환경의 부족도 숨기지 않고 기록합니다. 예를 들어 `timeout_or_canceled`나
`capture_limit`은 실행 상태이고, `unsupported_candidate_shape`는 좁은 검증기의
지원 범위를 벗어났다는 뜻입니다. 이들을 모두 모델 능력 부족으로 셀 수는 없습니다.

| `riido-taskrun` 종료 코드 | 의미 |
| ---: | --- |
| 0 | CLI가 0으로 종료하고 독립 검사가 대상 범위의 요구사항을 수용하며 복사한 인증을 정리함. 사용량이 완전하다는 뜻은 아님 |
| 1 | 시작한 시도의 불성공을 기록. JSON의 실행·검사 상태로 원인을 구분해야 함 |
| 2 | 인자·고정값·인증·격리 등의 시작 전 거절, 또는 출력 오류 |
| 3 | 시도는 기록했지만 수용 확인·후보 회수·기록 저장 등의 근거가 미확정이거나 인증 정리가 실패 |
| `--help`·`--spec`의 0 | 정보 출력 성공. 모델 실행이나 후보 수용을 한 것은 아님 |

에이전트는 종료 코드만으로 학습 라벨을 만들지 않고 JSON의 실행·검사·사용량을
각각 읽습니다. 계획에 대한 시작 전 거절이나 미실행 항목도 남기고 성공한
결과로만 바꾸지 마세요. 로컬 계획의 해시와 순서 확인은 계획 자체가 결과를
보기 전에 공개되었다는 사실을 독립적으로 증명하지 않습니다.

## 지금 확인하는 것과 다음 단계

위 명령 예제는 실행 방법을 설명하며 자체로 모델 결과를 만들지 않습니다.
[실험 52](../experiments/task-outcomes/RESULTS-52.ko.md)는 별도로 고정한 실제 실행
기록입니다. 두 주석 작업에서는 양쪽 프로필의 수용과 사용량을 비교할 수 있지만,
행동 작업의 Sol 시도에는 전체 사용량이 없어 완전한 비용 비교를 할 수 없습니다.
52 시점의 표본은 세 요청뿐이므로 프로필 절감이나 실사용 라우팅 성공을 입증하지 않습니다.
새 학습, 모델 가중치 공개, 실사용 라우팅 활성화나 최종 2,400건 채점은 하지 않았습니다.

[실험 50](../experiments/task-outcomes/RESULTS-50.ko.md)과
[실험 51](../experiments/task-outcomes/RESULTS-51.ko.md)의 중단 기록은 그대로 보존합니다.
새 실제 실행은 사용 가능한 프로필을 확인하고 별도 계획·새 출력 디렉터리로
고정해야 합니다.

적은 개발 작업으로 먼저 성공·실패·미확정 근거를 정확하게 남길 수 있는지
확인합니다. 이후 서로 다른 공개 작업을 늘리고 같은 조건에서 모델별 성공과
모든 시도의 사용량을 비교해야 상향·하향·보류 라우팅의 효과를 판단할 수
있습니다. 자세한 해석은 [작업 기록과 독립 검사](task-outcomes.ko.md), 필요한
규모와 분리는 [골든셋 설계](golden-set-scale.ko.md), 확보 순서와 실행 예산은
[골든셋 확보 계획](golden-set-acquisition.ko.md)에 설명되어 있습니다.

[53의 역사적 결과](../experiments/task-outcomes/RESULTS-53.ko.md)는 한 행동 요청의
네 반복입니다. 입력 339,805·그중 캐시 280,576·출력 4,890·그중 reasoning 76을
기록했습니다. Luna의 총 입력은 적었지만 uncached 입력은 Luna 31,425·Sol 27,804로
순서가 바뀌었습니다. 이 수치를 54의 기록과 혼합하거나 실제 금액 절감으로 바꾸지 않습니다.

역사적 [54 결과](../experiments/task-outcomes/RESULTS-54.ko.md)는 파서 요청 하나의
네 반복이며 입력 485,609·그중 캐시 423,552·출력 5,282·그중 reasoning 157입니다.
이 요청에서 Luna의 입력·출력과 관측 실행 시간은 두 반복 모두 작았지만,
실제 모델 정체·구독 사용량·금액·공급자 내부 호출 수는 미확정입니다.
54당시 누적 다섯 요청은 일반 성능·실사용 라우팅을 검증하지 못합니다.
키워드 보호와 파서 경계 처리는 runtime에 반영했지만 새 학습·보정·가중치 공개·
Codex 정책 활성화·protected final 채점은 하지 않았습니다.

후속 [55 결과](../experiments/task-outcomes/RESULTS-55.ko.md)는 기존 외부 두 요청을
처음 실행해 현재 관측 범위를 일곱 요청·스무 기록으로 확장했습니다. 새 final이나
학습 라벨은 아니며, 프로필별 한 관측과 중간 휴식의 한계도 함께 읽으세요.
