# 전체 원본 코호트 고정

`riido-sourcecohort`는 등록된 전체 원본 코호트의 바이트 고정값, 배분,
전체 결합, 읽기 영수증, 근거 위치와 선언된 검토를 확인합니다. 원본·댓글·라벨·
모델 점수를 만들지 않습니다. 예제와 테스트는 합성 소프트웨어 자료이며 승인된
데이터가 아닙니다. 필드별 상세 계약은 [영문 문서](source-cohort-freeze.en.md)에 있습니다.

```sh
go build -o ./riido-sourcecohort ./cmd/riido-sourcecohort
./riido-sourcecohort --default-recipe > recipe.json
./riido-sourcecohort --help
```

v1은 정확히 400행입니다. 사전 라벨 맥락 층 8개에 각각 50행,
층마다 TRAIN/DEV/CAL/TEST 35/5/5/5행이며 DEV는 short 30행, general 10행입니다.
고정된 recipe도 이 계약과 같아야 합니다. 작은 부분집합을 완료로 처리하지 않습니다.

## 원본 생성 전: 배분 확인

호출자가 독립적으로 고정한 plan, recipe, prospective registry를 제공합니다.
registry의 모든 행은 source_id, stratum, split, dev_style, source:null,
명시적인 dependencies 배열을 갖습니다. DEV 외 dev_style은 빈 문자열입니다.
split은 train/dev/cal/test입니다. ID는 ASCII 영문·숫자·밑줄·하이픈, 최대 128바이트입니다.

plan schema는 `riido-sourcecohort-plan-v1`입니다. creation/reviews/allocation은
이 단계에서 명시적인 null이며 source_schema도 null일 수 있습니다. 아직 없는
자료의 해시나 manifest를 꾸며 넣지 않습니다. 알려진 파일 참조는 path/sha256/bytes의
정확한 값이며 path는 --root 기준 상대 경로입니다. 절대 경로·상위 이동·root 밖으로
나가는 부모 심볼릭 링크는 거부합니다. 0바이트는 빈 파일 해시와만 결합됩니다.
선택 필드 `reviewer_assignments`는 고정 plan에서만 사용합니다. 배분 단계에서는
비어 있지 않은 배정 목록을 거부합니다.

```sh
./riido-sourcecohort --root ./new-cohort \
  --plan allocation-plan.json --sha256 <독립적으로-고정한-plan-해시> \
  --bytes <정확한-plan-바이트수> --allocation-only \
  --out ./new-cohort/allocation-evidence
```

400개 슬롯과 계획된 의존성의 배분을 확인하고 집계 상태 allocation_ready를
출력합니다. ALLOCATION.private.json, SUMMARY.json과 새 읽기 영수증을 남깁니다.
원본 문서는 열지 않습니다. 이 결과는 배분에 관한 결과입니다.

## 생성·검토 후: 전체 고정

실제 creation/reviews/source_schema 고정값과 앞 단계 ALLOCATION.private.json의
고정값을 담은 새 plan을 씁니다. 생성 후 registry는 별도 파일로 보존하며 기존 슬롯
ID·층·분할·스타일·계획 의존성을 유지하고 각 source에 실제 파일 참조를 넣습니다.
이전 배분 자료를 수정하지 않습니다.

```sh
./riido-sourcecohort --root ./new-cohort \
  --plan freeze-plan.json --sha256 <독립적으로-고정한-plan-해시> \
  --bytes <정확한-plan-바이트수> --out ./new-cohort/freeze-evidence
```

registry/creation/reviews는 모든 400개 source_id에 정확히 한 번씩 결합되어야
합니다. 누락·추가·중복은 전체 고정을 막습니다. 생성 기록에는 원본 고정값,
author_id, creator_kind=ai_nonhuman, 선언된 rights_allowed=true, 실제 생성 UTC,
의존성과 생성 영수증이 필요합니다. 참조한 입력이 없다면 consulted_none=true와
빈 consulted 배열을 명시합니다.

검토는 원본·스키마·observation·외부 검사 보고서의 고정값, reviewer_id와 UTC,
complete/structural, source_scope=complete_source_situation,
independent_semantic_verdict=declared_pass, holds, evidence와 읽기 영수증을
보존합니다. author와 reviewer ID는 달라야 하지만 이는 선언된 절차 제한입니다.
실제 신원이나 제공자 독립성을 증명하지 않습니다.

검토자별 새 컨텍스트를 쓰려면 고정 plan에 `reviewer_assignments`를 선언할 수
있습니다. 생략하거나 `[]` 또는 `null`을 쓰면 모든 검토에 단일 `checker_id`를
쓰는 기존 동작을 유지합니다. 각 배정의 `reviewer_id`는 서로 달라야 하며
`author_id`와도 달라야 합니다. 기본 `checker_id`를 반드시 포함하고 각
`source_ids` 목록은 비어 있으면 안 됩니다. 검토자 배정은 최대 400개,
각 원본 ID 목록도 최대 400개이며 목록 전체는
등록된 400개 원본 ID 전체를 정확히 한 번씩 나눠 맡아야 합니다.
누락·추가·중복 ID를 허용하지 않으며 각 검토의 reviewer_id는
해당 원본에 배정된 검토자와 같아야 합니다.

아래 합성 plan 조각은 형태만 보여 줍니다. 완전한 plan에는 등록된 400개 ID를
모두 나열해야 합니다.

```json
"checker_id": "fake-checker-a",
"reviewer_assignments": [
  {"reviewer_id": "fake-checker-a", "source_ids": ["fake-source-001"]},
  {"reviewer_id": "fake-checker-b", "source_ids": ["fake-source-002"]}
]
```

배정은 검토자별 새 컨텍스트를 선언할 수 있게 하지만 실제 신원이나 제공자
독립성을 증명하지 않습니다. 기존 고정 조건은 모두 유지합니다. 배정이 있으면
각 검토자의 비어 있지 않은 ID는 담당 원본 start 읽기 영수증의 actor와 정확히
일치해야 합니다. result는 start 해시로 계속 결합됩니다.
각 원본 검토자의 원래 reviewpacket start/result는 실제 바이트 해시로 결합되고,
원본의 해시·크기, 읽기 1회, 오류 없음과 생성→시작→완료→결정→검토의 UTC 순서를
확인합니다. result 경로는 start 경로에 .result를 붙인 값입니다. 영수증은 해당
도구의 기록된 읽기만 설명하며 최초의 모든 읽기나 옛 시각을 복원하지 않습니다.

근거 위치는 UTF-8 바이트 start/end이며 end는 제외됩니다. 문자 경계의 비어 있지
않은 범위만 허용하고 source_scope 근거는 전체 원본을 덮어야 합니다.
proposition/opposition/event_anchor/communication_time/absence/not_applicable
근거도 보존합니다. 위치가 유효하다는 사실로 명제·범위·극성·시간의 의미가 맞다고
판단하지 않습니다. 공식 observation 바이트는 private 결합에 그대로 남습니다.
외부 검사 보고서와 실행 binding의 원본·observation·스키마 고정값을 확인하지만
이 도구가 외부 검사를 다시 실행하거나 그 결과를 의미적 진실로 취급하지 않습니다.

registry·생성 행·생성 영수증·검토의 모든 필수 의존성을 합쳐 확인합니다.
없는 ID, 한 목록 안의 중복, 자기 의존성, 분할을 넘는 의존성을 거부합니다.
서로 다른 기록에 반복된 같은 선언은 합칩니다. 정렬된 결합으로 실제 의존성
연결 성분을 계산합니다.

## 결과와 한계

stdout은 집계만 출력합니다. 성공 상태는 structural_provenance_freeze_ready이며,
private FREEZE.private.json에 전체 400행의 원본, observation, 생성·검토 기록과
검사 binding을 보존합니다. 원본 본문·보류 이유·비공개 경로·사례 ID를 stdout에
출력하지 않습니다. 검토 manifest의 보류는 개별 원본 읽기 전에 모두 집계하고
보류가 있으면 고정을 막습니다. 실패해도 전체 분모와 미확인 행을 유지하며 부분집합을
승인하거나 보류를 숨기지 않습니다. 기술 실패 후 확인하지 못한 행은 미확인입니다.

JSON의 누락·알 수 없는 키·중복·대소문자 별칭·잘못된 형식·잘못된 UTF-8·짝 없는
Unicode surrogate·뒤따르는 값·과도한 중첩/배열은 거부합니다. 메타데이터 입력은
최대 8 MiB, 원본은 16384바이트, 고정/출력 한도는 최대 64 MiB입니다. plan은 고정
한도와 다음 영수증 쌍 4096바이트, 종결 메타데이터 65536바이트를 미리 확보해야
하며 출력 최소치는 1 MiB입니다. 초기 plan 읽기에는 별도 64 MiB bootstrap 한도를
씁니다. 매 작업과 최종 직렬화의 공간을 확인합니다. 새 비공개 디렉터리와 배타적인
0600 파일만 쓰고 동기화·원자적 게시를 수행합니다. 기존 출력은 덮어쓰지 않습니다.
저장장치 실패로 종결 기록이 불가능하면 미완성 자료를 남깁니다.

이 결과는 원본 의미, 완전한 명제 목록의 진실, 실제 저자, 법적 권리,
독립 제공자, human Gold를 증명하지 않습니다. 권리·AI 생성·의미 검토는 선언 및
별도의 검토 책임입니다. 학습 자료를 승인하거나 S3–S11, frame/reference, 스타일,
지원 셀, 정확도, unsafe-positive, 자원·유용성 조건을 충족시키지 않습니다.
