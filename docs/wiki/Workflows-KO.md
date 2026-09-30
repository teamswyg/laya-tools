# 기능별 사용법

[English](https://github.com/teamswyg/laya-tools/wiki/Workflows-EN) · [홈](https://github.com/teamswyg/laya-tools/wiki)

`examples/…` 경로가 있는 예제는 laya-tools를 clone한 폴더에서 실행합니다. 옵션은 질문 앞에 둡니다. 예제 가격과 저장소 목록은 실제 사용자 설정이 아닌 가상 자료입니다.

## 특정 동작을 담당하는 코드 찾기

```sh
riidolaya search --root . --lexical --json 'redirect authorization'
riidolaya search --root . --candidates 8 --limit 3 'where are redirect headers removed?'
```

먼저 키워드·식별자를 사용합니다. 두 번째 명령은 설치된 Laya를 추가로 사용하며, 사용할 수 없으면 경고와 함께 키워드로 대체합니다. 경로·줄 범위·코드 일부를 보고 내용을 확인하세요. 후보에 없는 코드를 재정렬로 복구할 수는 없습니다. 재정렬은 지연이 클 수 있으므로 실제 질문에서 `--lexical`과 비교하세요.

코드 관련성 체크포인트를 추가하려면:

```sh
riidolaya setup --checkpoint code
riidolaya search --checkpoint code 'redirect authentication'
```

모델·저장소 선택 실험에는 code 대신 base를 사용합니다. 한국어 검색은 미검증이며 `--candidate-query 'English identifiers'`로 영어 식별자 힌트를 주면 키워드 후보 생성에 도움이 될 수 있습니다.

## 실행 전에 모델 선택 비교

```sh
riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json
```

계산기처럼 생각하면 됩니다. catalog에는 능력·컨텍스트·가격을, request에는 예상 토큰·작업 평가·현재 모델·전환 이력을 제공합니다. 구독 정보를 읽어서 자동으로 채우는 값이 아닙니다.

| `plan.status` | 사용자가 할 일 |
|---|---|
| `recommend` | 제안 모델과 다른 후보의 제외 이유를 확인합니다. 실행된 것은 없습니다. |
| `hold` | 현재 모델이 제약을 만족하고 전환이 보류됐다는 뜻입니다. |
| `blocked` | 허용되는 추천이 없습니다. 능력·가격·예산 사용액·사유를 확인하고 빈 모델 ID로 실행하지 않습니다. |

유효한 blocked 계획은 종료 코드 0입니다. 잘못된 JSON·입력은 0이 아닙니다. 마지막에 작업 문장을 붙이면 로컬 Laya 분류를 추가하며 유료 코딩 모델을 호출하지 않습니다. 이 계획은 `codex`에 자동 적용되지 않습니다.

## 작업을 담당할 저장소 preview

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
riidolaya repo-preview --catalog examples/repositories/catalog.json --laya --json 'refund invoices'
```

호출자가 접근할 수 있는 저장소만 카탈로그에 넣습니다. 전체 소스 대신 짧은 역할 설명과 별칭을 사용하세요. [예제 카탈로그](https://github.com/teamswyg/laya-tools/blob/main/examples/repositories/catalog.json)를 참고하면 됩니다.

| `status` | 의미 |
|---|---|
| `candidate` | 키워드 후보이며 정확도가 보정된 선택이 아닙니다. `reason`을 확인합니다. |
| `suggest` | 실험적 Laya 추천이 기준을 통과했습니다. 실행 권한은 아닙니다. |
| `abstain` | 추천을 보류했습니다. 후보·사유를 확인하거나 사람이 해결합니다. |

모든 결과는 `preview: true`입니다. GitHub 접근·clone·저장소 수정·작업 배정을 하지 않고 권한도 부여하지 않습니다. `--laya`는 선택 사항이며 시험한 구성은 영어용입니다. 현재 측정에서는 정확도 개선을 입증하지 못했습니다.

## 선택적으로 새 Codex 작업 실행

아래 이름을 본인의 Codex에서 사용할 수 있는 모델 ID로 바꿉니다.

```sh
riidolaya route --fast-model YOUR_FAST_MODEL --standard-model YOUR_STANDARD_MODEL --strong-model YOUR_STRONG_MODEL --json 'Fix a spelling mistake'
riidolaya codex --model YOUR_MODEL --dry-run 'Implement the feature'
```

첫 명령은 추천만, 두 번째는 실행할 명령만 보여줍니다. `--dry-run`을 빼면 설치된 Codex를 실행하고 기존 사용량이 소모될 수 있습니다. 로그인·권한·승인 설정은 그대로 적용되며 진행 중인 대화의 모델은 바꾸지 않습니다.

route 출력의 `suggested_tier`는 모델 제안, `tier`는 정책 적용 결과이며 `abstained: true`는 강한 능력을 유지하도록 보류했다는 뜻입니다. 높은 분류 확률이 코딩 성공을 보장하지 않습니다. 직접 지정한 `--model`이 우선합니다.

다음: [에이전트·Go 연동](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-KO) 또는 [문제 해결](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-KO).

## 모델을 올리거나 내리는 양방향 추천

쉬운 작업은 낮추고, 어려운 작업은 현재 모델보다 강한 모델로 올릴 수 있습니다. 현재 모델은 요청 JSON의 `current`, 성능 순서는 설정의 `rank`로 지정합니다. 가격이 비싸다는 이유만으로 상향으로 판단하지 않습니다.

```sh
riidolaya plan --config examples/planner/config.json --request examples/planner/upgrade.json --json
```

이 예제는 `example-fast`에서 `example-strong`으로 `direction: upgrade`, `reason: quality_upgrade`를 반환합니다. 기존 `request.json`은 하향 예제입니다. 위 예제는 직접 제공한 난이도 평가를 사용하며 Laya 추론은 실행하지 않습니다. 명령 끝에 영어 작업 설명을 붙이면 로컬 Laya가 난이도를 평가합니다.

`direction`은 `upgrade`(상향), `downgrade`(하향), `lateral`(같은 등급), `initial`(첫 선택), `unchanged`(유지)이며, `blocked`에는 없습니다. 상향은 비용 회수 기간이나 전환 대기 횟수보다 품질을 우선하지만 예산·기능·컨텍스트 제약과 수동 고정은 무시하지 않습니다. Laya의 불확실한 판단은 상향 근거로 바꾸지 않으며, 현재 모델도 요구 조건을 못 맞추면 `blocked`입니다. 기본 설정에서는 catalog의 확신도 기준 0.9도 통과해야 하므로 switch의 상향 기준 0.5만 넘는다고 추천하지 않습니다.

추천만 반환합니다. 진행 중인 Codex 대화를 자동 전환하거나 실패를 감지해 재실행하지 않습니다. 연동하는 에이전트가 작업 단계마다 현재 모델과 평가를 갱신해 호출해야 합니다.
