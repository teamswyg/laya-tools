# 59: 설명을 원문과 리터럴 소스 위치에 연결하는 사전 계획

[English](PLAN-REFERENCES-59.en.md) · [58 다음 준비](NEXT-STORED-58.ko.md) · [58 결과](RESULTS-STORED-58.ko.md)

**이번 작업은 설명이 맞는지 판정하기 전에, 검토할 원문과 근거 위치를 정확하게 연결하는 준비다.** 기존 공개 요청 72개와 후보 설명 216개를 그대로 읽고, 문장 288개의 JSON 위치·UTF-8 바이트 수·해시와 계약 18개의 리터럴 소스 표현식 위치를 기록한다. 새 요청, 새 정답, 모델 성능이나 비용 절감 결과를 만드는 단계가 아니다.

58의 알려진 요청 51개는 답 있음 34개와 답 없음 17개이며, unknown 21개도 전체 72개에 남는다. 후보 2·3·4개 구성과 원래 순서, 기존 문구·정답·그룹 17개를 바꾸지 않는다. 알려진 부모를 포함한 그룹은 16개이고, quoted-delimiters의 그룹 ID64는 부모 4개가 모두 unknown이다. 이름·해시·그룹·정답·검토 상태는 감사 메타데이터이며 scorer 특징으로 전달하지 않는다.

## 한 번의 메타데이터 생성 범위

`riido-captionref prepare`는 소스·입력·지원 파일과 실행 계획을 준비한다. 이 모드의 Bind와 메타데이터 생성 호출은 0이고, Git blob 검사 대상은 20개다. 이후 소스·입력·실행 계획·바이너리를 봉인한 뒤, `riido-captionref references`로 별도 새 출력 디렉터리의 `results.json`에 공식 메타데이터 생성을 한 번 실행한다. 실행 계획의 고정 위치는 `experiments/short-claim/execution-plan-59.json`이다. 예정 공식 시도는 1회, 재시도는 0회다. 실패하면 부분 기록과 고정된 실패 코드·단계별 카운터를 남기며, 같은 공식 시도를 조용히 다시 실행하지 않는다.

생성기는 고정된 공개 raw 파일 3개를 `storedaudit.Bind`로 연결하고, Go AST를 읽어 legacy 계약의 개별 원소와 typed 리터럴 배열의 표현식 구간을 찾는다. 원본 후보 함수나 원본 패키지 registry를 실행하지 않는다. 원시 표현식 해시는 직렬화된 리터럴 표 해시와 다르다. 표현식 위치를 찾는 일은 실제 Input/Want를 재구성하거나 후보별 Got·실패 위치를 새로 관측하는 일이 아니다.

| 확인할 항목 | 고정 범위 |
|---|---|
| 컴파일된 소스 | 5개: CLI main, captionref references/provenance, storedaudit binding/provenance |
| 실행 입력 | 6개: 원본 raw JSON 3개와 리터럴 소스 data.go/state.go/flow.go 3개 |
| 실행 지원 파일 | 9개: 이 계획 양언어, build recipe, oracle review, CLI/생성기 테스트 2개, go.mod/go.sum/LICENSE |
| 공식 Git blob 검사 | 소스 5 + 지원 9 + 입력 6 + 실행 계획 1 = 21개 |
| 생성할 참조 | 부모 72개, 후보 위치 216개, 원문 참조 288개, 계약 표현식 18개, 리터럴 소스 파일 3개 |

Go 1.27.1, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-p=1`로 빌드한다. 읽을 입력과 결과의 상한은 1 MiB, 바이너리 상한은 64 MiB다. 파일을 읽기 전과 읽는 중에 크기·일반 파일·안전한 경로를 검사하고, 디스크와 컴파일된 소스의 바이트·해시를 비교한다. Git blob 검사는 파일마다 최대 5초다. `GOMAXPROCS=1`, Go heap soft limit 256 MiB는 설정이며 실제 CPU·전체 RSS·GPU·지연 측정값이 아니다.

출력은 provenance envelope, `generation_counters`, `references` 본문을 포함한다. 성공 상태는 `references_generated_content_review_pending`, 실패는 `incomplete`와 고정 `failure_code`다. 출력은 `O_EXCL`로 새로 생성하여 이전 결과를 덮어쓰지 않는다. 공개할 경우 원본 envelope를 바꾸지 않고 `caption-coverage-59.json`으로 복사한다. 계획 파일을 작성한 것만으로 실행 봉인이나 결과 생성이 완료된 것은 아니다.

공식 결과 이후 별도 `frozen_test.go`로 CI 회귀 재현을 준비한다. `references` Report 전체와 `generation_counters`는 바이트 또는 구조체를 **정확히 비교**하며 부동소수점 허용 오차를 쓰지 않는다. 원래 소스·입력·실행 계획·바이너리와 native Darwin envelope 핀은 따로 확인한다. Linux/macOS CI 바이너리가 원래 바이너리와 같다고 주장하지 않는다. 회귀 반복은 공식 시도나 독립 요청을 늘리지 않으며, 관측 뒤 추가한 테스트를 원래 지원 파일 9개에 소급해서 넣지 않는다.

## 참조가 정확해도 설명 검토는 남는다

모든 source-fidelity, request-contract coverage, observation-fields 검토는 `pending`으로 유지하고 `literal_payload_reified=false`, `training_ready=false`를 기록한다. 실제 함수·helper·type·sentinel의 전체 source closure와 기존 code/bundle 해시 계산식의 재연결도 이번 표현식 참조만으로 승인하지 않는다.

원본 정답과 문장 충실도는 다른 문제다. 잘못 동작하는 구현을 문장이 정확히 설명할 수도 있다. [내용 검토 recipe](content-review-recipe-59.json)는 구현 설명의 충실도, 요청 계약의 범위, 관측 필드, 명시적 부정·경계의 네 검토 축을 나눈다. 원문 구간과 근거를 독립적으로 읽고 누락·모순·지원하지 않는 범위를 기록한다. 해시·위치 검증 통과나 단어 발견을 내용 승인으로 바꾸지 않는다. 설명을 잘라서 정답처럼 만들거나 unknown을 오답/답 없음으로 바꾸지 않는다.

[역할 recipe](role-recipe-59.json)는 아직 `unassigned`다. 기존 whole-component 하한 train 9 / validation 3 / calibration 3을 유지하고 transfer 0을 허용한다. 알려진 부모가 있는 그룹 16개는 숫자상 하한을 만족시킬 수 있지만, 실제 배정·coverage·통계적 독립성이 검증된 것은 아니다. 미래 membership·seed 봉인과 고정 알고리즘으로 component 전체를 배정하며, 지금 seed나 역할 소속을 정하지 않는다. 이미 본 58의 자료를 블라인드 검증으로 되돌릴 수 없고, 그룹별 점수에 맞춰 배정하지 않는다. 새 gate나 원형 18개를 모든 역할에 요구하는 조건은 추가하지 않는다.

NEXT-STORED-58에서 제안한 세 산출물 중 이번 결과는 caption coverage의 **참조 준비 부분**이다. 의미 검토, source-transfer의 두 scoped proposal과 역할의 실제 배정은 남는다. 별도 내용·역할 recipe와 prototype ledger는 이번 실행 지원 파일 9개에 포함되지 않는 후속 준비 기록이며, 실행 중 읽지 않는다. `no_roles_plan`, `synthetic_single_pipeline`을 유지한다.

## 실패도 함께 기록한다

[prototype ledger](prototype-ledger-59.json)는 비공개 준비의 2회 실행과 Bind 2회를 보존한다. 첫 실행은 수동으로 적은 원형 이름 세 개가 원본과 달라 출력 전에 실패했다. 출력은 0개이고 당시 바이너리 SHA는 기록되지 않았다. 이름을 원본과 일치시킨 두 번째 실행은 참조 1개를 만들었다. 이 이력과 독립 참조 검토는 공식59 시도나 의미 승인, 새 독립 요청이 아니다.

새 정답·원천 API·순위·모델/유료 호출·역할 배정·fit·가중치·보호 최종 접근·운영 활성화는 모두 0이다. 자동 검사·독립 내용 검토·기존 CI 기준을 따르며 새 사람 승인 절차를 추가하지 않는다. 원문/소스 핀·순서·참조가 맞지 않거나 예산·지원 범위를 벗어나면 중단한다. 의미 검토의 부족함은 그대로 남기고 정답·그룹·기준을 바꾸지 않는다.

도메인별 서로 다른 보호 최종 요청 2,400개는 여전히 별도의 일반화 목표다. 기존 72개, 문장 288개, 리터럴 표현식 18개, 재현 반복을 최종 요청 수로 더하지 않으며 기존 보호 최종/CoSQA reserve를 읽지 않는다.
