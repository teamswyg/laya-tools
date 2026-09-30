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

## 이번 개발 파일럿의 범위

[사전 계획 50](../experiments/task-outcomes/plan-50.json)은 기존 공개 개발 작업
3개를 두 요청 프로필로 비교합니다. `luna-low`는 `gpt-6-luna`와 `low`,
`sol-low`는 `gpt-6.1-sol`과 `low`를 명시합니다. 순서는 계획에 고정되어 있고,
최대 **Codex CLI 실행 6번**, 한 번에 하나, 각 작업 실행 120초와 별도 검사
45초를 한도로 합니다.

이는 내부 LLM 호출이 6번이라는 뜻이 아닙니다. 한 CLI 실행 안에 여러 모델
요청이나 공급자 재시도가 있을 수 있으며, 현재 그 수는 `unknown`입니다.
실행기는 자체 재시도·resume·fallback을 하지 않습니다. 시작한 실패·시간 초과도
파일럿의 실행 수에 포함합니다. 아래 명령은 한 시도만 실행하므로, 관리자가
계획 전체의 횟수·순서와 미실행 항목까지 관리해야 합니다.

| 작업 ID | 독립 검사할 요구사항 | 검사하는 파일 범위 |
| --- | --- | --- |
| `comment-preview-authority` | 저장소 제안은 실행 권한을 주지 않는다는 정확한 주석 변경 | `pkg/reporouter/router.go` 한 파일, 정확한 변경·gofmt 확인 |
| `comment-budget-period` | 예산 기간을 설명하는 주석을 지정 문장으로 변경 | 기준 파일 3개, 정확한 변경·gofmt·고정 테스트 |
| `catalog-min-context` | 선택적 `MinContext` 추가, 음수 거절, 최소 문맥 조건 적용, 기존 동작 유지 | 기준 파일 9개, 경계·JSON 계약과 고정 catalog/planner 테스트 |

공통 기준은 공개 리비전 `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`의
`internal/taskverify/testdata/base`입니다. 파일과 LICENSE·NOTICE, 작업 명세와
독립 검사 코드의 해시를 확인합니다. 파일 범위 밖의 추가 파일이나 런타임
캐시는 `unassessed`입니다. `accepted`는 **해당 요구사항과 범위**의 통과이며
저장소 전체나 임의의 코드에 대한 승인으로 읽으면 안 됩니다.

3개 작업은 두 코드 가족에서 나온 개발 사례입니다. 모델별로 두 번 실행해도
서로 다른 작업은 3개이며, 성공률 일반화나 학습 정답을 만들기에는 부족합니다.
[2,400개 골든셋 설계](golden-set-scale.ko.md)의 최종 평가를 대체하지 않습니다.

## 실제 작업 결과를 보기 전 Laya가 예측한 것

[미리 기록한 예측](../experiments/task-outcomes/routing-predictions-50.json)은
기존 고정 Laya INT8 모델을 CPU, 한 스레드, 기본 임계값 `0.9`로 실행한 결과입니다.
임계값을 바꾸거나 이 세 작업으로 학습하지 않았습니다.

| 작업 | 작성자의 예상 난이도 | Laya의 제안 | 가장 큰 확률 | 적용 tier |
| --- | --- | --- | ---: | --- |
| `comment-preview-authority` | fast | standard | 0.480 | strong, 보류 |
| `comment-budget-period` | fast | fast | 0.486 | strong, 보류 |
| `catalog-min-context` | standard | standard | 0.591 | strong, 보류 |

세 요청 모두 잘리지 않았지만 신뢰도가 `0.9`보다 낮아 보류했습니다. 제안과
작성자의 예상이 일치하는지보다 **실제 작업을 어떤 프로필이 성공시키는지**가
중요합니다. 파일럿은 미리 정한 luna/sol 비교이며, 보류 뒤의 strong 모델을
실행하는 라우터 실사용 시험은 아닙니다. `gpt-6-astra`는 이 계획에서 실행하지
않습니다. 이 확률을 보정된 성공 확률로 해석하지 않습니다.

세 번의 차가운 프로세스에서 Laya 라우터의 최대 RSS는 약 **1.48 GB
(1.38 GiB)**였습니다. 모델·native session을 새로 불러오는 비용이 포함된
각 한 번의 관측이며, 반복 요청의 지연이나 Go heap·GPU 메모리가 아닙니다.
현재 수치는 아주 작은 메모리라는 목표를 달성했다는 근거가 아닙니다.

## 사용하기

저장소 루트에서 Go 바이너리를 만듭니다. 실행기에 Python은 필요하지 않습니다.

```sh
mkdir -p .cache/bin
go build -o .cache/bin/riido-taskrun ./cmd/riido-taskrun
.cache/bin/riido-taskrun --help
.cache/bin/riido-taskrun --task comment-preview-authority --spec
```

`--help`와 `--spec`은 모델을 실행하지 않습니다. 기본 호출도 실행을 거절합니다.
실제 실행 전에는 사전 계획, 공개 기준 파일, 고정한 Codex 실행 파일과 버전,
기존 로컬 ChatGPT 로그인이 준비되어 있어야 합니다. 현재 지원 버전은
`codex-cli 0.158.0`이며, 실행 파일 해시가 계획과 다르면 추론 전에 거절합니다.
버전이나 계획을 바꾼 경우에는 새 계획을 먼저 정해야 합니다.

다음은 **계획 50의 첫 시도를 실제로 실행하는** 예입니다. `/absolute/...`는
설명을 위한 자리 표시자입니다. 자신의 절대 경로로 지정해야 하며, 비공개
출력 디렉터리는 아직 존재하지 않아야 합니다. 이는 예제를 읽는 것과 달리
계정 사용량을 소비할 수 있는 실행입니다.
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
  --codex-sha256 788a818fbb9596869c7a487554507cb8bdca17584b8671112b23f9e225ba35c8 \
  --codex-version 'codex-cli 0.158.0' \
  --plan-file /absolute/laya-tools/experiments/task-outcomes/plan-50.json \
  --plan-sha256 186a59fd11f09cce4e1e26fe3b79919a7bf1fc65db19caa25a3d2aa7aaa1a967 \
  --attempt-ordinal 1 \
  --task-spec-sha256 a01f1c95e818511d6bcac35a0eb20b0b6c943a73b4fe7652a550bb0036dbd747 \
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
| 0 | CLI가 0으로 종료하고 독립 검사가 대상 범위의 요구사항을 수용. 사용량이 완전하다는 뜻은 아님 |
| 1 | 시작한 시도의 불성공을 기록. JSON의 실행·검사 상태로 원인을 구분해야 함 |
| 2 | 인자·고정값·인증·격리 등의 시작 전 거절, 또는 출력 오류 |
| 3 | 시도는 기록했지만 수용 확인·후보 회수·기록 저장 등의 근거가 미확정 |
| `--help`·`--spec`의 0 | 정보 출력 성공. 모델 실행이나 후보 수용을 한 것은 아님 |

에이전트는 종료 코드만으로 학습 라벨을 만들지 않고 JSON의 실행·검사·사용량을
각각 읽습니다. 계획에 대한 시작 전 거절이나 미실행 항목도 남기고 성공한
결과로만 바꾸지 마세요. 로컬 계획의 해시와 순서 확인은 계획 자체가 결과를
보기 전에 공개되었다는 사실을 독립적으로 증명하지 않습니다.

## 지금 확인하는 것과 다음 단계

이 문서의 예는 실행 방법을 설명하며, 실제 코딩 모델의 결과를 보고하는 것이
아닙니다. [계획 50](../experiments/task-outcomes/plan-50.json)과
[사전 예측](../experiments/task-outcomes/routing-predictions-50.json)은 실행 후
기록과 구분해야 합니다. 이 단계에서는 새 학습, 모델 가중치 공개, 실사용
라우팅 활성화나 최종 2,400건 채점을 하지 않습니다.

적은 개발 작업으로 먼저 성공·실패·미확정 근거를 정확하게 남길 수 있는지
확인합니다. 이후 서로 다른 공개 작업을 늘리고 같은 조건에서 모델별 성공과
모든 시도의 사용량을 비교해야 상향·하향·보류 라우팅의 효과를 판단할 수
있습니다. 자세한 해석은 [작업 기록과 독립 검사](task-outcomes.ko.md), 필요한
규모와 분리는 [골든셋 설계](golden-set-scale.ko.md)에 설명되어 있습니다.
