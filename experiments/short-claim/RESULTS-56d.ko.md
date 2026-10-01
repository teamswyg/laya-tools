# 상주 Go 힌트 도구의 실제 왕복 비용: 56d 결과

[English](RESULTS-56d.en.md) · [원본 결과](results-56d.json) · [요약 수치](summary-56d.json) · [사전 설계](PLAN-RESIDENT-56d.ko.md) · [실행 계획](execution-plan-56d.json) · [봉인 기록](freeze-56d.json)

**봉인한 24개 실행 행이 모두 완료됐다. 재시도 없이 25,080개의 반복 요청을 검증했고, 자식 프로세스의 lifetime peak RSS는 9.21875–10.515625MiB였다.** 한 번 켜 둔 작은 Go 도구에 JSON 한 줄을 보내고 검사 순서 한 줄을 받는 실제 비용을 측정했다. Laya 모델 추론·학습·의미 정확도·실제 LLM 사용량 절감을 측정한 결과는 아니다. 이 로컬 기록 작성은 CI 실행에 앞선다. 실제 CI 완료 링크와 자동 병합 확인은 검증 뒤 [이슈19](https://github.com/teamswyg/laya-tools/issues/19)의 별도 완료 댓글에 기록한다.

## 사람과 에이전트가 이 결과를 읽는 방법

가상 사용자는 작은 힌트 도구를 여러 번 호출하는 에이전트다. 요청과 후보 설명 3개를 보내면 `riido-shortclaim --stream`이 모두 남겨 둔 채 검사 순서를 제안한다. 외부의 독립 검사가 실제 의미를 확인해야 하며, 응답은 계속 `unverified_heuristic`이다. 이번 실험은 이러한 호출 경로의 비용과 입력·출력 대응을 확인했다.

실행한 것은 `fixed_order`, BM25, 단어·연속 단어 일치, 좁은 규칙의 **무학습 Go 기준선**이다. 매 요청마다 준비·검증·순위·digest·출력 직렬화를 다시 수행했으며 결과 cache hit는 0이다. GPU·인코더·Laya 가중치·모델 호출은 사용하지 않았다. 따라서 아래 메모리를 “Laya 모델 전체가 쓰는 메모리”로 읽으면 안 된다. Pool의 48개 입력은 모두 좁은 규칙 문법 밖이어서 `narrow_rule`도 전부 `unsupported_rule_request`를 기록하며 BM25로 fallback했다. 성공한 좁은 문법 경로의 성능은 이번 자료에서 관측하지 않았다.

## 반복 관측과 독립 자료를 구분한다

기존 공개 원본 72개에서 후보가 **원래 3개인 48개만 선택**했다. 나머지 24개는 원래 후보 수가 2개 또는 4개라 제외했으며, 후보를 잘라 3개로 맞추지 않았다. 정규화된 텍스트 특징의 중복은 0개여서 pool 크기도 48개다. 원래 부모·후보 ID와 순서는 [corpus](wire-corpus-56d.json)의 72개 origin link에 보존된다. ID와 출처는 순위 특징에 들어가지 않는다.

| 원천 | 원래 부모 | 선택 | 제외 |
|---|---:|---:|---:|
| 기존 `probes-56` | 48 | 24 | 24: 2후보 12개, 4후보 12개 |
| Typed `probes-56b` | 24 | 24 | 0 |
| 합계 | 72 | 48 | 24 |

`same`은 첫 payload 하나를 반복하고, `distinct`는 48개 pool을 원래 순서대로 순환한다. 각 기준선·workload마다 새 자식 프로세스를 3회 시작해 `4 × 2 × 3 = 24`행을 실행했다. 행당 첫 응답 1개 + warmup 20개 + timed 1,024개 = 1,045개다. 전체 첫 응답은 24개, warmup은 480개, timed는 24,576개다.

**25,080회와 24,576회는 반복 처리·타이밍 관측 횟수다. 새로운 독립 요청이나 학습·최종 평가 표본 수가 아니다.** 이번 단계의 새 독립 요청, fit, 모델 호출, 새 모델, protected final 읽기는 모두 0이다. 도메인별 서로 다른 보호된 최종 요청 최소 2,400개 목표는 충족하지 않았다.

## 실제 시간·CPU·메모리

Apple M4 Pro·24GiB·macOS 26.6.2·darwin/arm64·Go 1.27.1에서 native CPU 경로를 실행했다. CGO=0·trimpath 빌드이며, controller와 child의 GOMAXPROCS=1 및 Go heap 256MiB soft limit을 사용했다. 한 번에 child 1개·요청 1개만 진행했다. 이는 OS 스레드 수·CPU affinity·전체 RSS의 hard cap을 보장하는 설정은 아니다.

아래 p95는 각 행의 timed 1,024개 **wire RTT**에서 구한 뒤, 같은 조건의 새 프로세스 3행 사이 최솟값–최댓값을 표시했다. 합친 표본의 p95나 신뢰구간이 아니다. Wire RTT는 요청 쓰기 시작부터 완전한 응답 줄 수신까지이며, 이후 controller 검증 완료까지의 시간은 따로 기록한다. RSS와 CPU/request 열은 표시를 위해 반올림했다. 정확한 값은 원본 JSON에 있다.

| 기준선 | Workload | 3행의 wire RTT p95 범위, µs | Child peak RSS 범위, MiB | Child lifetime CPU/request 범위, µs |
|---|---|---:|---:|---:|
| `fixed_order` | same | 21.875–37.292 | 9.219–9.281 | 15.307–32.073 |
| `fixed_order` | distinct | 39.875–43.375 | 10.094–10.391 | 23.844–24.723 |
| `bm25` | same | 24.125–27.334 | 9.250–9.469 | 17.424–17.767 |
| `bm25` | distinct | 45.833–47.375 | 10.109–10.266 | 27.104–27.762 |
| `lexical_ordered` | same | 23.208–24.708 | 9.281–9.438 | 17.414–17.559 |
| `lexical_ordered` | distinct | 48.709–50.333 | 10.031–10.359 | 28.340–28.801 |
| `narrow_rule` | same | 23.916–25.666 | 9.250–9.453 | 17.111–17.605 |
| `narrow_rule` | distinct | 46.834–49.125 | 10.219–10.516 | 27.447–27.847 |

Child의 user+system CPU 합계는 24개 lifetime을 합쳐 **0.568013초**다. 행별 CPU/request는 첫 응답·warmup을 포함한 **1,045개**로 나눈 값이며, timed 구간만의 CPU나 순수 추론 시간은 아니다. 모든 child의 RSS 원래 단위는 Darwin의 bytes이고, 변환 후도 같은 값이다. Child 전체 범위는 9,666,560–11,026,432 bytes다.

Controller는 별도로 **user 0.2652초 + system 0.240851초 = 0.506051초**의 replay CPU를 썼고, lifetime peak RSS는 **34,717,696 bytes = 33.109375MiB**였다. Controller RSS는 준비를 포함한 전체 lifetime peak이며 행별 RAM이나 replay 중 증가량이 아니다. Child와 controller의 서로 다른 peak를 더해 동시에 쓴 총 peak라고 부르지 않는다. 전체 replay wall은 **1.482910833초**였다. 이 wall/CPU에는 행별 기대 stream 해시 준비와 정리가 포함되며, child 시작 전의 행별 해시 준비는 204.166–427.041µs로 따로 기록됐다.

Timed controller 비용의 행별 p95 범위는 수신부터 검증 완료까지 8.083–17.458µs, 실제 검증 작업 7.125–15.750µs, 수신 후 검증 시작까지의 전달 간격 1.042–2.125µs였다. 쓰기 대기의 p95 범위는 1.375–2.333µs다. 서로 겹치는 시간이나 다른 항목의 p95를 더하거나 빼서 순수 application 시간을 만들지 않는다. **순수 startup·child 내부 처리·pipe 비용은 unobserved**다.

첫 행 `fixed_order/same/repeat 1`의 시작 호출부터 첫 응답까지는 **577.13ms**였다. 나머지 23행은 2.784792–3.267459ms다. 첫 값을 버리지 않았으며 원인은 미확인이다. 새 프로세스를 OS cache까지 비운 cold라고 부르지 않았고 cache flush도 하지 않았다.

## 결과가 온전히 완료됐는가

24행 모두 planned/attempted/fully-written/received/validated가 단계별 예정 수와 같았다. Failed·incomplete·not-attempted·stderr bytes는 모두 0이다. 모든 행에서 입력·출력 실제 stream SHA가 기대 SHA와 같고, EOF·stdout/stderr reader join·watcher join·cleanup complete·reaped를 확인했다. `Wait` 호출은 행당 정확히 1회였으며 종료 코드는 모두 0이었다. 취소·시간 초과·정리 오류도 기록되지 않았다.

시작 호출부터 정리까지 기록한 행 wall은 29.215084–617.744ms로 child 15초 예산 안에 있었고, 전체 replay는 60초 안에 완료됐다. 정리 시간은 3.250–6.708µs로 공유 1,000ms 정리 예산 안이었다. 자동 재시도는 0회다. 공개 원본 결과는 **81,471 bytes**, SHA-256은 `d24a50cf71968bb0336f58812ce02a647facff28cb40595e6cebbcd9587c4eaf`다. JSON 결과 크기는 8MiB 공개 예산 안이며, 반복 전송한 전체 wire bytes의 합계와는 다른 항목이다.

기대 protocol 응답은 준비 단계에서 같은 공개 구현으로 만들었다. 따라서 해시 일치와 완료 상태는 전달·결정적 출력의 정합성을 보여 주며, 독립 의미 정답이나 올바른 검사 순서를 인증하지 않는다.

## 봉인과 해석의 한계

준비 소스는 [60b0350](https://github.com/teamswyg/laya-tools/commit/60b03500fab417470d48d1a096be4ca1f43fafc2), 입력·바이너리·24행 계획을 공식 실행 전에 고정한 matrix commit은 [891e52d](https://github.com/teamswyg/laya-tools/commit/891e52d83e9aaf39c87cb5ef55c6d492724bebf1)이다. [Freeze](freeze-56d.json)에는 실행 당시 **사전 테스트 파일 5개를 포함한 구현 pin 31개**, 두 바이너리 SHA와 build recipe·corpus·plan SHA가 있다. Git blob 31개 확인은 별도 봉인 작업에서 수행했고 CLI는 commit 형식만 확인한다. 결과 수집 후 회귀 guard를 추가했으며, 이를 공식 실행 전에 존재한 검사나 봉인 원천으로 소급하지 않는다. CI 완료와 자동 병합은 [이슈19](https://github.com/teamswyg/laya-tools/issues/19)에 별도 기록한다.

입력 크기도 두 workload에서 같지 않다. `same`의 LF 포함 전송 줄은 385 bytes다. 전체 pool은 385–1,165 bytes, 평균 675.6875 bytes이며 요청 정규화 단어 수는 4–32개, 후보는 4–32개다. 첫 요청은 8단어이고 후보는 5/7/8단어다. 그러므로 same/distinct 차이를 입력 재사용이나 cache만의 인과 효과로 해석하지 않는다. 원문 요청은 36–236 bytes, 후보는 28–252 bytes로, 가능한 모든 512-byte 입력·언어·최악 조건을 측정한 것도 아니다. 정규화가 기호를 제거하므로 feature SHA는 미래 의미·규칙·모델 cache key의 안전성 증거가 아니다.

[56a](RESULTS-56a.ko.md)의 첫 cold **370.932333ms**와 별도 profile의 `runtime.kevent` flat **83.02%**는 원인·귀속 미해결 상태로 보존한다. 56a는 8후보와 `io.Discard`/단계 benchmark, 56d는 선택한 3후보와 실제 상주 pipe 왕복이므로 직접 속도 향상률을 계산할 수 없다. 이번 실험은 cache·SIMD·SoA·lock 변경의 성과나 LLM 절감을 입증하지 않는다.

다음 의미·효용 검토는 [56c 계획](PLAN-56c.ko.md)에서 분리해 진행한다. Open-loop 동시성·matched-byte echo 대조·ready 신호·cache와 무효화는 후속 별도 설계 대상이다. **최종 결정을 내리지 않는 작은 힌트의 호출 비용을 확인했다는 범위**를 유지하며, 의미 정확도·전체 LLM 사용량·저사양 모델의 가중치/실행 메모리는 각자 별도의 증거를 요구한다.
