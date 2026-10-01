# 작은 동작 힌트 도구 사용하기

[English](USAGE-56.en.md) · [이전 준비 계획](PLAN-56.ko.md) · [실행 계획](execution-plan-56a.json)

`riido-shortclaim`은 짧은 요구와 후보 설명을 받아 **어느 후보부터 독립적으로 확인할지** 제안하는 Go 도구다. 모든 후보를 남기며, 작업 완료·정답·실행 권한을 확정하지 않는다. Laya 인코더나 학습된 가중치를 사용하지 않는 별도 실험이다.

사람은 “만료 항목을 제거하면서 순서를 보존한다”는 요구와 후보들의 짧은 설명을 넣을 수 있다. 에이전트는 반환된 후보 ID 순서대로 실제 코드와 검사를 확인할 수 있다. 설명이 실제 코드를 충실히 나타내는지는 이 도구가 보장하지 않는다. 전체 코드를 몰래 요약하거나 LLM을 호출하지 않는다.

## 처음 실행

저장소에서 Go1.27.1로 다음을 실행한다. Python, 모델 다운로드, 외부 API 키 없이 빌드·실행된다.

```sh
CGO_ENABLED=0 go build -trimpath -o bin/riido-shortclaim ./cmd/riido-shortclaim
./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
```

출력의 `verification_order`는 후보 ID와 점수다. `status`는 항상 `unverified_heuristic`이며 점수는 성공 확률이 아니다. 동작 반전이나 부정을 이해하지 못한 어휘 기준이 잘못된 순서를 제안할 수 있으므로 실제 확인이 필요하다. `input_sha256`는 요청·후보·순서·메타데이터의 길이를 구분한 입력 식별용 digest다. 원문과 provenance를 결과에 되풀이하지 않지만 **후보 ID는 출력하므로 공개할 수 있는 ID를 사용**한다.

요청과 각 후보는 UTF-8 최대512바이트, 정규화 후 최대512바이트·32단어이며 후보는1~8개다. JSON 한 건은12KiB 이하, ID는1~64바이트, provenance는1~128바이트의 한정된 ASCII 식별자다. schema는 `riido-short-behavior-claim-v1`이다. 중복 키·알 수 없는 키·잘못된 유니코드·길이 초과·알 수 없는 schema는 고정 오류 코드와 실패 종료로 거절한다. 입력을 자르지 않는다. 이는 준비 계획의 fallback 제안을 구체화한 실제 입력 계약이다.

## 여러 번 쓰기

각 JSON 객체를 한 줄로 전송하고 `--stream`을 사용한다. 하나의 프로세스가 여러 요청을 처리하므로 매번 프로세스를 시작할 필요가 없다. 잘못된 레코드를 만나면 종료하고 그 뒤 레코드는 처리하지 않는다. 입력을 닫으면 정상 종료한다. 실행 결과 캐시는 없으며 매번 실제 계산한다.

```sh
./bin/riido-shortclaim --stream --baseline bm25 < public-requests.jsonl
```

`public-requests.jsonl`은 같은 입력 schema로 직접 준비할 파일 이름이다. 사람이 선택해서 실행하는 preview이며 기존 `riidolaya` 라우팅이나 Codex 기본 설정을 바꾸지 않는다. Go 애플리케이션은 `pkg/shortclaim.Load`, `Validate`, `Rank`를 사용한다. 반환된 `Prepared`를 수정하거나 직접 구성할 수 있어 `Rank`는 다시 검증한다. 모델별 승인이나 자동 작업 실행 인터페이스는 없다.

## 네 가지 비교 방식

| 옵션 | 의미와 범위 |
|---|---|
| `fixed_order` | 원래 순서 그대로 |
| `bm25` | 후보 내부 단어 빈도·문서 빈도로 일치도를 계산 |
| `lexical_ordered` | 단어와 순서를 유지한 연속 두 단어의 일치 |
| `narrow_rule` | 양쪽이 `behavior-v1:`의 공개된 좁은 문법을 완전히 따를 때만 규칙 비교. 한 후보라도 문법 밖이면 요청 전체를 BM25로 처리 |

```sh
./bin/riido-shortclaim --baseline narrow_rule < examples/shortclaim/rule.json
```

규칙은1~4개 `if [not] 이름 [비교연산 이름] then [not] 동작 [대상] else [not] 동작 [대상]` 절을 처리한다. 비교연산 `< <= > >= == !=`에는 주변 공백이 필요하다. 동작은 keep/remove/accept/reject/preserve/reverse다. 조건 부정은 분기 교환으로 비교하고 동작 부정은 정확한 flag로 보존한다. 절 순서·피연산자 순서를 유지한다. 일반 문장·코드의 의미를 추론하거나 동의어·반대말·논리적 함의를 일반화하지 않는다. 실제 자연어 개발 자료를 규칙에 맞춰 바꾸지 않았다.

## 구조와 측정

요청당 고정 후보 배열8개와 단어 배열32개를 사용한다. 토큰은 문자열의 일부를 가리키며, 작은 점수 배열과 안정 정렬로 처리한다. 공유 캐시·공유 가변 상태·공유 lock·runtime map이 없다. 메타데이터와 정답은 특징에서 분리한다. 전체 시스템이 ECS/SoA나 SIMD라는 주장은 하지 않는다. 이 크기에서는 고정 배열로 검사할 수 있는 비용을 먼저 측정한다.

`--benchmark`는 공개된8후보 예제에서 입력 준비, 순위 계산, 입력 digest, 직렬화, 전체 Go 호출을 측정한다. 순위 단계는 재검증·특징·점수·정렬이 합쳐진 값이며 각각의 독립 타이밍으로 오인하지 않는다. `cmd/riido-shortperf`는 사전 고정 계획을 확인한 뒤 OS 전체 프로세스 RSS·CPU·cold 시작을 기록한다. 측정 명령과 결과는 후속 결과 문서에서 제공한다.

`--cpuprofile`은 유지보수용 **로컬 전용**이다. 원시 프로파일은 개인 경로와 stack 정보가 포함될 수 있으므로 Git이나 Hub에 올리지 않는다. Go HeapAlloc은 아직 회수하지 않은 객체를 포함하며 OS peak RSS와 다르다. 이 실험은 CPU만 사용하고 GPU·ORT·인코더는 사용하지 않는다.

48개 개발 요청은48개 독립 표본이 아니다. 공유 후보·원형·소스·핵심 템플릿을 전이적으로 묶고 자연어 충실성과 유한 Go truth table을 따로 확인한다. 이 도구의 빠른 실행은 의미 정확도·LLM 사용량 절감을 입증하지 않으며 도메인별 최종2,400요청 목표도 대신하지 않는다.
