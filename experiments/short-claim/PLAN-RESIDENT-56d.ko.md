# 상주 자원 실험56d 사전 설계: 캐시 없는 JSONL 왕복 비용

[English](PLAN-RESIDENT-56d.en.md) · [기존56a 자원 결과](RESULTS-56a.ko.md) · [56b 감사 결과](RESULTS-56b.ko.md) · [다음 의미·효용 경로56c](NEXT-56c.ko.md)

**이 문서는 실행 전 설계안이다. 기계 실행 계획·입력·바이너리의 최종 봉인, 공식 자원 측정, CI 완료 기록은 아직 없다.** 여기 적힌 횟수와 예산은 제안값이며 측정값이 아니다. 현재56b의 새 성능 실행·모델 호출·fit·새 가중치는 모두0이다.

작은 Go 힌트 도구를 여러 번 쓰는 에이전트를 가상 사용자로 둔다. 매 요청마다 프로세스를 새로 시작하는 대신 하나의 프로세스에 JSON 한 줄을 보내고 추천 검사 순서 한 줄을 받는다. 이번 질문은 이 왕복에 드는 전체 CPU·메모리·시간이다. 순위의 의미 정확도와 실제 LLM 사용량 감소는 별도 효용 검증 대상이다.

## 기존 관측과 이번 범위

[56a](performance-56a.json)는 같은8후보 예제를 기준마다 cold3회·warm1회로 실행했다. 첫 cold는 **370.932333ms**, 나머지11회는2.564958~5.232667ms였으며 최초 실행·OS/파일 cache의 원인은 미확인이다. [별도 프로파일](profile-summary-56a.json)의 `runtime.kevent` flat0.88초·**83.02%**도 귀속 미해결로 남는다. 이를 application 병목이나 공유 lock 경합으로 확정하지 않는다. 이 원문·수치·첫 관측은 보존한다.

현재 [stream](../../cmd/riido-shortclaim/main.go)은 매 줄 JSON 준비·검증·순위·입력 digest·실제 출력 직렬화를 수행하며 결과 캐시가 없다. [benchmark](../../cmd/riido-shortclaim/bench.go)는 `io.Discard` 출력과 단계 반복을 사용하고 `--stream`과 함께 실행할 수 없다. `--cpuprofile`도 benchmark 전용이다. 56a의 warm p95를 상주 pipe 왕복 지연으로 바꾸어 부르지 않는다. 오프라인 `go/types` 원천 감사 비용도 runtime 힌트 비용에 포함하지 않는다.

## 측정 전에 준비하고 봉인할 입력

기존 [48요청](probes-56.json)과 [24요청](probes-56b.json), 합계72개 공개 자체 작성 자료를 준비 원천으로 삼는다. 각각의 `FeatureInputs` 투영으로 **요청과 순서 있는 후보 설명만** 추출한다. prototype·source·계약·정답·역할은 순위 입력에서 제외한다. typed 자료의 원천·언어 검사는 준비 단계에서 수행하고, 자원 측정 driver는 준비된 runtime JSONL만 읽는다.

1. 후보가 원래3개인 투영만 사용한다. 후보를 삭제·추가·복제하거나 설명을 잘라3개로 맞추지 않는다. 요청과 후보 각각 원문·정규화512바이트 이하, 정규화32단어 이하, JSON 한 줄12KiB 이하를 검증한다.
2. 정규화된 요청, 후보 수3, 정규화된 후보 설명3개를 **순서대로 길이를 붙여** 인코딩한 SHA-256을 `feature_sha256`으로 제안한다. 인코딩 규칙과 구현 해시는 실행 전에 고정한다. ID·provenance는 이 해시에서 제외한다. 같은 설명의 ID만 바꾸거나 whitespace만 바꾼 것은 새로운 feature payload가 아니다. 이 해시는 보수적인 중복 선택 기준이며, 모든 기준선의 출력 동등성이나 미래 cache key의 안전성을 증명하지 않는다.
3. 안정적인 원본 순서에서 feature SHA를 중복 제거한다. 선택·제외 목록, 원래 부모와의 연결, 요청/후보의 바이트·단어 수, 서로 다른 payload 수를 공개 준비 manifest에 남긴다. **적어도2개의 서로 다른 feature payload를 확인하기 전에는 공식 기계 실행 계획을 봉인하거나 실행하지 않는다.** 더 많은 payload가 독립 정답·학습 그룹·최종 요청 수를 늘리는 것도 아니다.
4. `same`은 고정한 첫 payload를 반복한다. `distinct`는 중복 제거한 전체 목록을 안정적인 순서로 순환한다. 측정 단계의 첫 목록 위치도 고정한다. 각 workload는 모든 기준선에서 동일한 byte sequence를 사용한다. 두 workload의 크기·단어 분포 차이를 공개하고, 차이를 입력 재사용만의 인과 효과로 해석하지 않는다.
5. runtime envelope의 schema·ID·provenance는 공개 상수·평면 식별자로 준비해 각 JSON 줄의 raw SHA와 기존 입력 digest를 고정한다. ID는 출력 대응 확인용 metadata다. `feature_sha256`, JSON raw SHA, metadata도 포함하는 입력 digest는 각각 다른 값으로 기록한다.

실행 전 별도 준비 검사에서 기대 protocol 응답을 만든 뒤 플랫폼별 응답 bytes SHA를 봉인한다. 이는 입력 대응·결정적 출력 검사용이며 실제 상주 처리기는 모든 요청을 다시 계산한다. 실제 응답 SHA도 결과에 기록한다. 서로 다른 플랫폼의 부동소수점 출력 bytes가 같다고 가정하지 않는다. 이 단계의 protocol 준비는 정답 라벨·oracle 효용을 관측하는 실험이 아니다.

## 작은 실행 행렬과 예산

| 사전 제안 | 값과 의미 |
|---|---|
| 기준선 | `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule` |
| workload와 반복 | same/distinct 각각 새 상주 프로세스3회: `4 × 2 × 3 = 24 children` |
| 한 child의 요청 | 첫 응답1 + warmup20 + timed1024 = **1045요청** |
| 전체 예정 요청 | 25,080요청; timed24,576개. 반복이며 새로운 개발·최종 자료가 아님 |
| 처리 방식 | child1개·inflight1개. 이전 응답의 대응 검사를 마친 뒤 다음 요청 전송 |
| CPU·Go heap | child와 controller의 GOMAXPROCS=1, Go heap256MiB soft limit |
| 시간 제한 | child15초, 전체 공식 replay60초; 자동 재시도0 |
| 공개 결과 저장 | 신규 결과 합계8MiB 이내; 매 응답을 무제한으로 누적하지 않음 |

GOMAXPROCS는 Go의 동시 실행 P 수이며 OS 스레드 수·CPU affinity 보장이 아니다. Go heap soft limit은 전체 RSS·native·GPU 메모리의 hard cap이 아니다. CPU-only이고 인코더·GPU·모델·결과 캐시는 사용하지 않는 경로를 고정한다. 프로세스 종료와 pipe 정리 확인을 포함한 제한 정책은 driver 구현·검사 후 봉인한다.

`distinct`도 서로 다른 feature pool을 순환하는 반복이다. Child의 timed1024회가 서로 다른1024요청이라는 뜻이 아니며, timed24,576개는 타이밍 관측의 예정 개수다. 독립 학습·최종 표본 수에 더하지 않는다.

실행 순서24행을 사전에 고정한다. 기존 기준선 순서와 same/distinct 순서를 교대로 배치하는 결정을 측정 전에 남긴다. 첫 child와 각 child의 첫 응답을 별도 기록하며 제외하거나 다른 기준선의 예열로 대체하지 않는다. 새 프로세스를 OS cache까지 비운 cold라고 부르지 않으며 cache flush를 수행하지 않는다. 빌드·바이너리 검증/복사·corpus 준비는 별도 준비 단계다.

## 관측 가능한 시간과 관측 밖의 비용

| 항목 | 고정할 관측 경계 |
|---|---|
| spawn-call wall | controller의 시작 호출 직전부터 그 호출 반환까지. child가 준비됐다는 신호는 아님 |
| 첫 응답까지 | 시작 호출 직전부터 첫 완전한 응답 줄 수신까지: startup·첫 처리·pipe·controller가 합쳐짐 |
| wire RTT | 각 요청 쓰기 시작부터 완전한 응답 줄 수신까지. 쓰기 대기·child 처리·pipe·수신 framing 포함 |
| write blocking | 요청 줄 쓰기 시작부터 모든 byte의 쓰기가 반환될 때까지 |
| controller 검증 | 완전한 응답 줄 수신 후 JSON·digest·후보 보존 검사가 끝날 때까지 |
| warm 분포 | warmup 뒤1024개의 wire RTT p50/p95/max와 완료 수 |

읽기는 bounded framing으로 응답 줄을 계속 drain하고, 수신 타임스탬프를 잡은 뒤 대응 검사를 한다. 현재 stream에는 ready 신호나 child 내부 타임스탬프가 없다. 따라서 **순수 startup, child 내부 처리만의 지연, 순수 pipe 비용은 unobserved**로 둔다. 시간 항목들은 서로 겹칠 수 있으며 더하거나 차감해 application 시간을 만들지 않는다. 기존 `io.Discard` benchmark도 pipe 시간을 단순 차감하는 근거로 쓰지 않는다.

## CPU·메모리와 protocol 확인

Child 종료 시 user/system CPU와 **그 child 전체 lifetime peak RSS**를 기록한다. Controller는 `RUSAGE_SELF`의 user/system CPU 증가와 별도의 lifetime peak RSS를 기록한다. Controller RSS를 행별로 비교하려면 회별 새 controller에서 측정하고 orchestration 비용을 따로 둔다. 관측기가 child CPU를 controller CPU에 다시 더하지 않도록 검사한다. Peak RSS는 전후 차이·순간 warm RAM·Go heap이 아니며 두 프로세스 peak를 합쳐 동시 총 peak로 주장하지 않는다. Darwin RSS bytes와 Linux KiB의 단위를 구분해 bytes로 변환하고 원래 단위도 기록한다.

완료된 한 child의 CPU/request 분모는 **1045**로 첫 요청·warmup을 포함한다. Warm latency의 예정 분모는 **1024**다. Whole-process CPU를 warm phase만의 CPU라고 부르지 않는다. 각 행에 예정·시도·완료·실패·미완료 수와 입출력 byte 수를 함께 남긴다. 실패·timeout·protocol 불일치를 분모에서 숨기거나 성공으로 채우지 않으며, 불완료 행을 정상적인1024표본 결과로 게시하지 않는다. 불완료 행은 raw CPU와 실제 단계별 수를 보존하며, 완료된1045요청의 CPU/request 값으로 채우지 않는다.

각 응답은 하나의 bounded JSON 줄이어야 한다. schema와 `unverified_heuristic` 상태, 기준선, 고정한 입력 digest, 응답 raw SHA, 유한 점수, 후보3개의 ID가 입력의 정확한 permutation인지 확인한다. `narrow_rule`의 범위 밖 입력은 고정한 BM25 fallback과 `unsupported_rule_request`/`unsupported_rule_candidate` 이유를 확인한다. Malformed 입력은 정상 fallback으로 바꾸지 않는다. Stderr·reader 오류·호출 인자는 공개하지 않고 고정 enum과 종료/정리 상태만 남긴다. Protocol 검사는 후보 함수의 의미 정답을 인증하는 것이 아니다.

## 실행 gate와 이후 경로

새 driver의 원천·검사·manifest, 입력 투영/feature 해시 규칙, 플랫폼별 protocol 기대값, 24행 순서와 종료 정책을 검토·검사한다. **Go1.27.1·CGO0·trimpath binary의 buildinfo, 바이너리 bytes SHA, source/input/output pins, 신규 driver SHA와 기계 실행 계획을 공식 수집 전에 봉인**한다. 정확히 확인한 bytes로 실행해 재읽기 사이 변경을 막고 새 출력만 허용한다. CI는 실제 실행된 뒤 해당 상태·링크를 기록한다. 이 문서에는 실행할 준비가 끝났다는 뱃지나 성공 링크를 붙이지 않는다.

Matched-byte echo/framing **6대조**는 귀속 검토용 **후속 별도 계획**이다. 이번24행에 몰래 추가하거나 그 지연을 실제 왕복에서 차감하지 않는다. 정확한 startup 분리가 필요하면 버전이 있는 ready 신호를 별도 설계한다. Bounded queue의 open-loop 도착률·동시성, 캐시와 무효화, lock·SIMD·SoA 변경은 이후 독립 계획 대상이다. 기존 kevent 귀속 미해결과 첫 cold 관측을 보존하며 현재 cache·SIMD·속도 개선을 주장하지 않는다.

공개 산출물은 원본 공개 fixture, 계획·수치·해시·범위 설명이다. 실제 업무 입력·개인 절대 경로·인증·바이너리·원시 pprof/trace는 게시하지 않는다. 이번 설계의 학습·모델 호출·가중치·운영 활성화·protected final/CoSQA 접근은0이며, **도메인별 서로 다른 보호된 최종 평가 최소2400요청** 목표는 별도다.
