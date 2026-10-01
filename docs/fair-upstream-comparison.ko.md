# 외부 Go 작업의 공정한 비교 준비와 개발 관측

[English](fair-upstream-comparison.en.md) · [두 원천 계약](public-go-contracts-54.ko.md) · [규모와 분할 원칙](golden-set-scale.ko.md)

이 문서는 PDCA55의 **평가 framework 준비와 후속 개발 관측**을 설명합니다. 원본 코드에 어떤 변경을 요구하는지 모델에 충분히 알리고, 모델의 자체 검사와 독립 검사가 같은 Go 언어 조건을 사용하도록 합니다. 준비 시점의 외부 모델 실행은 0회였으며, 이후 기존 두 논리 요청에 실제 CLI 시도 4회를 기록했습니다. 로컬 수락기 검사와 가짜 실행기 예산 검사는 실제 모델 관측과 구분합니다. 현재 누적 실제 개발 범위는 **고유 요청 7개·기록 20개·시도한 코드 가족 5개·저장소 3개**입니다. [결과 55](../experiments/task-outcomes/RESULTS-55.ko.md), [기계 판독 결과](../experiments/task-outcomes/results-55.json)

## 수를 먼저 구분합니다

24개 합성 의미 검색 probe는 초기 동작을 확인한 개발 자료입니다. 다음 성능 주장을 위한 목표는 **평가 도메인마다 서로 다른 보호된 최종 요청 최소 2,400개**입니다. 검색, 실제 모델 라우팅, 저장소 선택, 작업 분할은 각자 다른 정답과 결과가 필요합니다. 이미 개발에 읽은 요청을 보호된 최종 자료로 이름만 바꾸지 않습니다.

| 범위 | 준비 시점과 현재의 수·의미 |
| --- | --- |
| 준비 시점까지의 실제 코딩 개발 관측 | 고유 요청 5개, CLI 기록 16개, 세 코드 가족, laya-tools 한 저장소 |
| 컴파일된 논리 검사 작업 | 서로 다른 논리 작업 7개, 원천 저장소 3개. v2 버전 추가로 늘지 않으며 실행 기록과 별도 범위 |
| 보존한 외부 원천 inventory | 원본 후보 120개. 정답이나 120개의 실행 가능한 계약이 아님 |
| 이번 v2 | 기존 외부 요청 2개의 설명·평가 버전. 새 고유 요청 2개를 추가한 것이 아님 |
| 외부 두 요청의 실제 모델 시도 | 준비 시점 0회 → 55에서 CLI 시도 4회, 서로 다른 논리 요청 2개 |
| 현재 누적 실제 코딩 개발 관측 | 고유 요청 7개·기록 20개·시도한 코드 가족 5개·저장소 3개 |
| 이번 외부 두 요청의 학습 실행·학습 라벨·최종 적격 요청 | 각각 0 |
| 이번 준비에서 새로 학습하거나 배포한 가중치 | 0 |

모델을 바꾸거나 같은 요청을 반복하고, 테스트 입력이나 번역을 늘려도 고유 요청 수는 늘지 않습니다. v1과 v2는 `logical_task_id`와 `previous_task_id`로 같은 원천 요청에 연결합니다. registry의 버전 ID가 두 개 늘어도 기존 7개 논리 작업의 수는 그대로입니다. [실제 후보 기록](../benchmarks/training/public-task-candidates.json)과 [원천 120개 목록](../benchmarks/training/public-go-acquisition-53.json)은 서로 합산하지 않습니다.

## 55의 실제 개발 관측

사전에 봉인한 두 child plan과 네 슬롯 순서를 그대로 실행했습니다. 사용자 요청으로 슬롯 2와 3 사이에 upstream 검토를 위한 일시정지가 있었고, 명시적 재개 후 미실행 슬롯 3·4만 진행했습니다. 슬롯 1·2의 실패나 결과를 재시도로 교체하지 않았습니다.

| 슬롯 | 요청 프로필·작업 | 독립 검증 | 본 실행 / 독립 검사 시간 |
| --- | --- | --- | --- |
| 1 | Sol low · humanize | accepted, 종료 pass 6개 | 76.404초 / 4.170초 |
| 2 | Luna low · UUID | rejected, pass 수 미집계 | 98.471초 / 5.852초 |
| 3 | Luna low · humanize | accepted, 종료 pass 6개 | 20.836초 / 4.646초 |
| 4 | Sol low · UUID | accepted, 종료 pass 218개 | 43.474초 / 5.421초 |

네 실행 모두 정상 종료·전체 사용량·인증 제거·프로세스 그룹 정리와 terminal receipt를 확인했습니다. 요청 프로필은 실행기에 명시한 설정이며 제공자가 확인한 모델 정체는 모두 `unknown`입니다. UUID 실패 후보는 뒤쪽의 잘못된 입력에서 이미 채운 UUID를 반환해 오류 시 전체를 0으로 반환해야 하는 요구를 어깁니다. 이는 봉인한 후보 소스의 정적 검토이며 재실행하지 않았습니다. 실패 기록의 pass 수 0은 검사 미실행이나 218개 전부 실패라는 뜻이 아닙니다.

[사전 native 예측](../experiments/task-outcomes/routing-predictions-55.json)은 전체 Prompt를 그대로 입력했고 두 건 모두 입력이 잘려 보류했습니다. 실제 라우팅 효용이나 보정된 난이도 정답이 아닙니다. 이번에 처음 실제 관측한 기존 두 요청 때문에 관측 범위가 5개에서 7개로 늘었지만, 컴파일된 registry의 새 논리 요청은 **0개**입니다. 요청마다 프로필별 한 관측, 일시정지와 서로 다른 캐시 사용량으로 모델 순위·금액·구독 소모율 절감을 입증하지 않습니다. 학습·protected final 채점·가중치 배포·기본 라우팅 변경은 수행하지 않았습니다. [원본 결과와 한계](../experiments/task-outcomes/results-55.json), [공개 영수증](../experiments/task-outcomes/parent-receipts-55.json)

## 왜 v2가 필요한가요

기존 humanize의 Prompt는 한 문장이어서 legacy 보존, 변경 가능 파일과 지원하는 코드 형태를 충분히 전달하지 않았습니다. UUID에는 많은 제약이 있었지만 정확한 허용 함수 목록까지 공개하지 않았습니다. 독립 검사기가 요구하는 조건을 모델이 실제로 받지 못하면 공정한 비교가 어렵습니다.

새 v2는 전체 요구를 `Spec.Prompt`에 넣습니다. 모델에 보내는 stdin은 이 Prompt 그대로이며 별도 Acceptance 항목이나 GitHub 문서를 모델이 읽었다고 가정하지 않습니다. 실제 Prompt bytes의 SHA, Spec, Definition, 독립 Contract와 실행 recipe를 계획 및 기록에 연결합니다. 이전 7개 작업의 Spec·Definition·Prompt·Contract와 원본 fixture 해시는 그대로 보존합니다. 과거 계획과 실행 기록도 다시 쓰지 않습니다.

| 요청 | v2 ID | 기존 논리요청 | 새 행동 요구의 예 |
| --- | --- | --- | --- |
| 전체 int64 서수 | `go55-humanize-ordinal64-v2` | `go53-humanize-ordinal64` | 새 API의 `-21st`와 기존 `Ordinal(-21)`의 `-21th`를 함께 보존 |
| 소문자 UUID 파서 | `go55-uuid-canonical-parse-v2` | `go53-uuid-canonical-parse` | 새 API는 엄격한 소문자 36바이트만 허용하고 기존 broad Parse/ParseBytes는 유지 |

humanize 계약은 호출 가능한 타입 `func(int64) string`을 확인합니다. 선언된 함수와 함수 값 변수의 차이를 보장하는 계약은 아닙니다. UUID는 기존 prefix 뒤에 정확히 `func ParseCanonical(s string) (UUID, error)` 하나를 추가하는 제한된 형태만 지원합니다.

## Go 실행 파일과 언어 버전은 다릅니다

Go 1.27.1 실행 파일을 쓰더라도 `go.mod`의 `go` 줄이 컴파일 언어 조건을 결정합니다. `go` 줄이 없으면 기본 언어 버전은 1.16입니다. [Go 공식 toolchain 설명](https://go.dev/doc/toolchain), [go.mod 설명](https://go.dev/doc/modules/gomod-ref)

| 원천 | 고정 revision | 원본 모듈 언어 | 실행 파일 |
| --- | --- | --- | --- |
| dustin/go-humanize | `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e` | 명시된 Go 1.21 | `go1.27.1` |
| google/uuid | `2d3c2a9cc518326daf99a383f07c4d3c44317e4d` | go directive가 없어 암묵적 Go 1.16 | `go1.27.1` |

v2 독립 검사기는 내부에 고정한 원본 `go.mod` bytes를 사용합니다. 원본 파일 SHA, 모듈 이름, 의존성이 없는 module/go directive 범위를 확인하며 후보의 모듈 설정을 사용하지 않습니다. 후보가 `go.mod`를 바꾸면 검사 대상 밖 소스 변경으로 거절합니다. 후보 테스트·`go.work` 설정도 독립 compiler 입력이 아닙니다. 실행 후 임시 원본 모듈이 그대로인지 확인합니다.

기존 v1의 임시 minimal module은 언어 1.27.1을 유지합니다. v2에만 원본 언어 recipe를 적용하며, 모듈 SHA·언어 버전·실행 파일 버전·recipe SHA를 각각 기록합니다. actor에 새 `-modfile` 경로를 추가하지 않습니다.

실제 로컬 대조 검사에서 두 v2 정답은 통과했습니다. humanize는 6개, UUID는 원본 suite를 포함한 218개 terminal pass를 기록했습니다. 작업별 오답 한 종류는 거절됐고, UUID의 난수 호출·원본 헤더 변경·추가 helper와 humanize의 import 변경은 실행 전에 unknown이 됐습니다. 같은 올바른 구현에 Go 1.22의 정수-range 문법을 넣으면 기존 v1에서는 통과하고 새 v2에서는 컴파일 단계에서 거절됐습니다. 이는 언어 조건 차이를 확인하는 개발 대조군이며 모델 실행이 아닙니다.

## accepted와 unknown의 범위

독립 검사기는 후보가 만든 테스트를 성공 근거로 사용하지 않습니다. 원본 테스트와 별도 고정 계약이 실행되어 요구한 이름의 테스트가 정확한 패키지에서 종료해야 합니다. 완전한 파일은 trusted Go 1.27.1 gofmt에 맞아야 합니다.

UUID checker는 정확한 원본 prefix 또는 고정 formatter 결과 뒤의 함수 하나, 지역 값 쓰기와 명시된 순수 호출 목록만 지원합니다. 함수 별칭을 통한 호출, 포인터, 추가 helper, 동시성, 전역 변경 등은 올바른 결과를 내더라도 `verifier_unknown`입니다. humanize에 이 UUID 전용 제한을 적용하지는 않습니다. 지원 밖 형태, 검증 환경 오류, 서비스 오류와 사용량 누락을 모델의 행동 실패나 사용량 0으로 바꾸지 않습니다.

원본 MIT·BSD-3-Clause LICENSE는 별도로 SHA를 확인한 뒤 검증 디렉터리에 고정 복사합니다. **후보 LICENSE의 내용은 성공 라벨이 평가하는 대상이 아닙니다.** accepted는 후보 또는 후속 배포의 attribution 준수를 인증하지 않습니다. UUID의 원본 copyright header는 평가 대상 소스 prefix의 일부이므로 strict shape 검사로 보존합니다. 이 범위의 소스 라이선스 확인을 모델 학습·가중치·데이터 배포 허가로 확대하지 않습니다.

## 두 작업 계획과 전체 네 슬롯

두 원천의 revision이 다르므로 작업마다 별도의 child plan을 둡니다. 각 child plan에는 두 프로필의 시도 하나씩을 넣고 parent manifest가 두 child plan SHA와 총 네 슬롯의 순서를 고정합니다. 여기서 parent는 실행 예산을 묶는 계획이며 새 부모·자식 모델을 학습한다는 뜻이 아닙니다. parent가 자식 SHA를 참조하므로 순환 해시를 만들지 않습니다.

하나의 고정 manifest와 **하나의 private durable ledger**에서 예약은 최대 4개이고 동시에 활성 controller는 1개입니다. 모델 시작 전에 슬롯을 예약하며 중복, 순서 오류, 다른 task/profile/Prompt/recipe 연결과 다섯 번째 예약을 거절합니다. 실패한 슬롯도 환불하거나 자동 재시도하지 않습니다. crash나 시작 증거 누락 등 불명확한 상태는 추측해 복구하지 않고 fail closed로 중단합니다. 예약 수, durable 시작 marker, terminal receipt와 시작 상태 미확인은 서로 다른 수입니다.

**이 제한은 해당 ledger의 실행에만 적용합니다.** 다른 ledger를 새로 만드는 것을 막는 호스트 전체·계정 전체 예산이 아니며, provider 내부 요청·재시도 횟수도 제한하거나 측정하지 않습니다. 한 실험을 다른 ledger로 재실행해 cap을 우회하면 안 됩니다. 예약·lease·stop 파일을 삭제하거나 ledger를 수동 reset해 재개하는 안전하지 않은 복구 경로는 제공하지 않습니다.

## 실제 모델 실행 전 gate

1. 새 source의 공개 CI와 독립 검토를 확인하고 clean executor·Go·CLI hashes를 고정합니다. 준비 문서만으로 통과를 주장하지 않습니다.
2. 원본 closure·라이선스와 v1 보존 회귀를 확인하고 v2 Spec·전체 stdin·Contract·recipe를 봉인합니다. 모델 자체 검사도 원본 언어로 실행되고 `go.mod`가 바뀌지 않는지 확인합니다.
3. 두 child plans와 parent manifest를 결과를 보기 전에 고정합니다. ledger의 중복·동시 실행·순서·환불 금지를 모델 없는 검사로 확인하고 실제 CLI 연결도 검증합니다.
4. 고정 입력의 사전 라우터 관측과 요청 프로필을 기록한 뒤에만 제한된 실행을 시작합니다. 요청 설정과 실제 제공된 모델, 전체 사용량·시간·실패 상태·검증 상태를 구분합니다.

이 gate 이후 두 요청의 실제 네 관측을 확보했지만, 일반적인 비교 성능, 절감 금액, 구독 소모율이나 초저자원 목표 달성은 확인하지 못했습니다. 유한한 개발 대조군과 네 관측으로 도메인별 2,400개의 독립 최종 요청을 대신할 수 없습니다. 비용·성공 가능 모델 집합의 학습 라벨은 별도 적격성 검토 이후에만 고려합니다.

## 모델을 호출하지 않고 내용 확인하기

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
.cache/bin/riido-taskverify --task go55-humanize-ordinal64-v2 --spec
.cache/bin/riido-taskverify --task go55-uuid-canonical-parse-v2 --spec
```

전체 Prompt와 버전 연결은 [v2 definitions](../internal/taskverify/upstream_definitions_v2.go), recipe는 [evaluation recipe](../internal/taskverify/evaluation_recipe.go), 고정 해시와 로컬 대조 검사는 [v2 tests](../internal/taskverify/upstream_v2_test.go)에 있습니다. ledger 연결은 [parent budget](../internal/taskrun/parent_budget.go)과 executor의 별도 검증 범위입니다. 이 문서에는 모델 원문 trace·인증·개인 코드·가중치를 포함하지 않습니다.
