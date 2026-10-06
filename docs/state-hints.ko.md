# 라벨·이모지·상태 힌트

[English](state-hints.en.md) · [학습 방법과 고정 계획](../experiments/state-hints/TRAINING.md) · [Go 결과](../experiments/state-hints/results/go-v1.json) · [Laya 결과](../experiments/state-hints/results/laya-v1.json)

내용을 작은 비용으로 분류해 뤼이도의 라벨, 이모지 표시, 작업 상태에 대한 제안을 만드는 개발용 기능입니다. 의미는 `question`, `blocker`, `reference`, `progress`, `completion_report`, `cancel_request`, `planned`, `unclear`의 여덟 가지로 고정했습니다. 모델 선택이나 추론 수준 조절은 이 기능의 범위에 포함하지 않습니다.

Go 분류기 학습과 실제 Laya 파인튜닝을 모두 실행했습니다. 현재 결과는 실험용이며 운영 도입 조건을 충족하지 못했습니다. 출력은 항상 shadow 제안이고 `mutation_executed:false`입니다. 실제 뤼이도 API, 반응 추가, 알림 발송에 연결하지 않았습니다.

## 사용법과 적용 조건

```sh
riidolaya state-hint --text "이 작업을 취소해 주세요." --json
riidolaya state-hint --model ./statehint.rsh --jsonl < requests.jsonl
```

첫 번째 명령은 별도 모델 없이 **학습하지 않은 규칙 기준선**을 사용합니다. 그 one-hot 점수는 규칙 선택을 나타내며 학습 모델의 보정된 확률과 같은 의미가 아닙니다. 학습 모델은 명시적으로 제공한 로컬 `.rsh` 파일에서 읽습니다. 자동 모델 다운로드는 없습니다. JSONL의 기본 입력은 `{"text":"분류할 내용"}`이며 선택적으로 `context`를 제공할 수 있습니다. CLI는 원문을 출력에 되풀이하지 않습니다.

텍스트는 유효한 UTF-8이고 4,096바이트 이하여야 합니다. JSONL 요청은 16 KiB 한도와 중첩 깊이 12의 경계를 둡니다. 모호한 중복 키, 알 수 없는 필드, 뒤따르는 JSON 값, 잘못된 context를 거절합니다. 학습 모델은 공백·기호·이모지만 있는 입력을 `no_word_content` 사유의 `unclear`로 처리합니다.

제안에 쓰는 라벨 ID는 호출자가 현재 active catalog에서 의미별로 제공합니다. 이모지는 제공한 후보 목록의 canonical lowercase code만 표시용으로 제안합니다. 기존 라벨·이모지와 같으면 추가 제안이 없습니다. `unclear`, 미학습 모델, 기준 미달 결과도 제안하지 않습니다. 계획의 confidence 하한은 0.9이며 CLI의 기본 margin은 0.05입니다.

작업 상태는 일반적인 `todo`, `active`, `done`, `cancelled`로 표현합니다. `progress`와 `started`, `completion_report`와 `completed`, `cancel_request`와 `cancelled`가 각각 일치하고, 이벤트의 작업 ID와 버전도 현재 입력과 일치해야 상태 변경을 제안합니다. 완료·취소된 작업은 일반 내용 힌트로 다시 열지 않습니다. 같은 상태를 다시 선택하면 상태 변경은 no-op입니다. 충돌하는 이벤트나 근거가 없는 문장만으로 상태를 바꾸지 않습니다.

`expected_version`과 `command_id`가 필요하지만, **JSON에 적은 `trusted:true`가 실제 권한이나 이벤트 증명을 만들어 주지는 않습니다.** 향후 신뢰할 수 있는 adapter가 이벤트 출처, 현재 catalog, 권한, 현재 버전, 명령 중복을 실제 적용 시점에 확인해야 합니다. 현재 CLI는 이 검증을 대신하거나 쓰기를 실행하지 않습니다.

## 두 학습 방법

| 방법 | 실제 수행한 작업 | 사용 범위 |
|---|---|---|
| 작은 Go 모델 | Unicode 소문자 단어와 문자 2/3그램을 1,024개 구간으로 해시하고, 8개 의미의 linear softmax를 cross-entropy + AdamW로 학습 | Go에서 내용 분류와 shadow 제안 |
| 기존 Laya 모델 | 원본 checkpoint의 앞부분을 동결하고 최종 `scorer.3.weight/bias` 1,025개만 cross-entropy + AdamW로 파인튜닝 | 유지보수용 MPS 비교 실험 |

Go 모델은 고정 배열과 호출자 소유 workspace를 사용하며 lock이나 공용 가변 cache를 두지 않습니다. weights를 복제해 warm start 학습할 수 있습니다. 이때 optimizer 상태는 새로 시작하고 temperature는 별도 calibration 집합으로 결정합니다. 8,200개 parameter의 파일은 32,960바이트입니다. `.rsh`에는 feature schema와 의미 순서, temperature, SHA-256을 저장하며 손상·비정상 수치·후행 데이터를 거절합니다. 모델 바이너리는 Git에 넣지 않습니다.

Laya는 원본 English checkpoint `convaiinnovations/laya`의 고정 revision과 파일 hash를 확인해 사용했습니다. 앞부분을 evaluation mode로 고정하고 각 문장의 최종 linear 입력을 한 번 추출한 뒤 학습 epoch에서 재사용했습니다. 이는 최종 linear 학습 비용을 줄이지만 추론할 때 필요한 전체 Laya backbone을 없애지는 않습니다. 작은 Go 모델은 이 backbone의 대체 export나 증류 모델이 아닙니다.

두 방법 모두 validation NLL로 후보를 선택하고 별도 calibration 자료로 temperature를 정했습니다. weights와 temperature를 잠근 후 마지막 test를 평가했습니다. 자세한 recipe와 자원 한도는 [학습 문서](../experiments/state-hints/TRAINING.md)에 있습니다.

## 데이터와 실제 결과

자료는 공개를 위해 직접 작성한 합성 문장 2,400개입니다. 96개 언어별 template family와 48개 한국어·영어 개념 쌍에 각각 이름·번호 변형을 적용했습니다. 실제 사용자 내용은 포함하지 않습니다. train 1,200개/48 families, validation 400개/16 families, calibration 400개/16 families, test 400개/16 families로 나눴습니다. 모든 구간은 합성 개발 자료이고 **2,400개의 독립적인 제품 정답은 아닙니다.**

아래 정확도는 마지막 test의 여덟 의미 중 가장 높은 점수를 선택한 결과입니다. 선택 건수는 confidence 0.9 이상이며 `unclear`가 아닌 결과이고, 실제 쓰기 수가 아닙니다.

| 방법 | 전체 정답/400 | 한국어 정답/200 | 영어 정답/200 | 선택 건수/400 | 선택 중 정답 |
|---|---:|---:|---:|---:|---:|
| 규칙 기준선 | 200 (50%) | 100 (50%) | 100 (50%) | 150 (37.5%) | 150/150 |
| Go 학습 모델, temperature 0.7 | 317 (79.25%) | 170 (85%) | 147 (73.5%) | 136 (34%) | 136/136 |
| Laya 원본, raw | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 93 (23.25%) | 72/93 (77.4%) |
| Laya 파인튜닝, raw | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 50 (12.5%) | 50/50 |
| Laya 파인튜닝, temperature 0.65 | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 74 (18.5%) | 50/74 (67.6%) |

Go 모델의 전체 정확도는 규칙보다 높았지만 선택 범위는 더 작았습니다. 136/136은 이 16개 test family 안의 관측이며 운영 정확도 100%를 의미하지 않습니다.

Laya의 raw test NLL은 1.8307에서 1.6071로 낮아졌지만 맞힌 수는 늘지 않았습니다. calibration에서 선택한 temperature를 적용한 test NLL은 1.7281로 raw tuned보다 나빴고, 선택한 74건 중 24건이 틀렸습니다. NLL이나 calibration 변화가 더 많은 정답을 보장하지 않는 결과를 그대로 보존합니다.

Laya MPS 실행은 약 181.77초였고 최종 parameter delta는 4,252바이트였습니다. 동결 parameter hash는 같았으며 저장한 최종 layer의 cached reload 오차는 0입니다. 전체 MPS 모델로 validation 한 건을 확인한 reload 오차는 약 `9.54e-7`입니다. Go 파일 재로딩은 전체 2,400개 입력의 예측 일치를 확인했습니다.

Go의 짧은 문장 한 개를 같은 프로세스에서 2,000회 측정한 p50은 1.833µs, p95는 1.875µs이고 호출당 allocation은 0입니다. 이는 feature extraction과 scoring만 포함합니다. JSON 처리·계획 생성·파일 읽기·프로세스 시작·전체 RSS는 포함하지 않습니다. 모델 struct 32,816바이트와 workspace 22,552바이트 역시 전체 메모리 사용량이 아닙니다.

## 확인된 실패와 다음 단계

Go 모델은 `completion_report` test 50개를 전부 `progress`로 분류해 recall이 **0/50**이었습니다. 영어 `question` 25개도 전부 `completion_report`로 분류했습니다. 현재 모델로 완료 상태 판단이 검증됐다고 볼 수 없습니다. Laya 역시 한국어 test 정확도가 12.5%이며 파인튜닝 후 전체 정확도 개선이 없어 현재 checkpoint를 운영용으로 채택하지 않습니다.

다음 작업은 완료 보고와 진행 보고, 질문과 인용·가정 표현을 구별하는 새 사례를 모으고, 기존 test를 다시 선택 기준으로 쓰지 않는 새 평가 구간을 만드는 것입니다. 별도 multilingual checkpoint 또는 더 넓은 adapter 학습은 새로운 provenance와 언어별 평가를 갖춘 비교 실험으로 진행합니다. 실제 Riido 연결은 신뢰할 수 있는 adapter의 읽기·shadow 관측부터 시작해 잘못된 라벨·상태 제안과 실제 처리 비용을 측정한 뒤 판단합니다.
