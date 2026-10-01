# 작은 Go 주장 도구: 구현·무학습 진단56a

[English](RESULTS-56a.en.md) · [사용법](USAGE-56.ko.md) · [숫자 결과](results-56a.json) · [자원 결과](performance-56a.json) · [프로파일 집계](profile-summary-56a.json)

**큰 인코더 없는 Go preview를 구현하고 측정했다. 48개 자체 작성 요청은11개 연결 그룹으로 묶여 학습 준비 하한15개에 못 미친다. 분할·학습·가중치·모델 호출은 모두0이며, 실제 LLM 절감을 주장하지 않는다.** 이전 [준비 계획56](PLAN-56.ko.md)은 준비 당시 자료0·실측0의 역사적 snapshot으로 유지한다.

## 무엇을 구현했나

`riido-shortclaim`은 한정된 요구와 후보 설명으로 독립 검사 순서를 제안한다. 입력과 정규화의 길이, Unicode·중복 JSON 키·후보 수를 엄격히 검사하며 모든 후보를 남긴다. 한 프로세스로 여러 JSONL 요청을 처리하는 `--stream`, Go 패키지 `pkg/shortclaim`, 네 무학습 기준을 제공한다. 모델 다운로드·Python·외부 API 키가 필요 없다. 기존 Laya 난이도 라우터와 Codex 기본 설정은 바꾸지 않는다.

기준은 fixed order, BM25, 단어/연속 두 단어 일치, 명시된 `behavior-v1:` 좁은 문법이다. 일반 문장에는 규칙을 추측해서 적용하지 않는다. 자연어 개발 자료48개는 전부 규칙 밖이므로 좁은 규칙 결과는 **48/48 BM25 fallback**이다. 유한 문법에서 빠르게 비교할 수 있다는 사실과 일반 문장 의미를 이해한다는 주장은 구분한다.

## 사전에 고정한 범위

공식 점수·성능 관측 전에 [commit45b2ad3](https://github.com/teamswyg/laya-tools/commit/45b2ad354ce2e56a11dff8f8407b09ba85ac2a70)에 입력·실행기·원천·무학습 [실행 계획](execution-plan-56a.json)을 봉인했다.

| 대상 | SHA-256 |
|---|---|
| 실행 계획 | `9c8461b9fe0a80a0f0976c8f07a68f3a47f0c534800a0a303e1cbd8a613caaeb` |
| 48요청 원본 | `0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0` |
| Go 함수·독립 literal 표 원천 | `b377343a64c019c205f69865286a21ce5d91dda3051c31aa5c69b75a40300559` |
| 8후보 성능 입력 파일 | `bc41dc423139be4adebc0687ffd30237f22458ea58fccf58771e2003276b5155` |
| Mac 측정 바이너리 | `5f5463005f9dd482247a5f028af9c5c6e6d7cec2c04aa18e4605e8a45a278fb2` |

계획에 runtime 원천8개 해시와 Go1.27.1을 함께 고정했다. Go 빌드에는 CGO0·trimpath·buildvcs=false를 사용한다. 실행 파일 바이트·Go 빌드 정보·패키지 ID를 확인한 뒤 검증한 바이트의 별도 private copy를 실행한다. 원본 입력 해시를 확인한 그 바이트를 평가하며, 결과 원문에 개인 경로·환경 전체·원시 trace를 넣지 않는다. 출력 크기와 전체60초·child15초 제한도 적용한다.

## 자료와 정답

자료는 자체 작성한 **48부모·144후보 설명·12동작 원형·11연결 그룹**이다. 영어만 포함하고 같은 작성자의 한계를 갖는다. 조건 유지, 권한의 부정, 포함 경계, 합계 비교, 안정 필터, 첫 항목 제거, 인접 중복 축약, 회전, 비감소 순서, clamp, prefix balance와 산술 동작 순서를 다룬다. 변수·숫자·번역·모델 설정 수를 독립 요청 수로 세지 않는다.

Go 정답 함수12개와 오답 함수24개는 후보가 생성하지 않은 literal 벡터68개로 검사한다. 원형별 벡터를 세 함수에 적용한 기본 대조 검사204개와 실제 후보집합 검사를 별도 기록한다. 코드 hash·이름 정규화 hash·핵심 template·동일 요청·공유 음성 후보를 전이적으로 합친다. closed-window와 clamp-window는 같은 closed-boundary core여서 함께 묶였고, 12원형이11그룹이 됐다.

자연어 충실성은 Go 테스트와 별도로 독립적으로 읽었다. “밖의 값은 받지 않는다”처럼 안쪽 값을 반드시 받는다는 뜻이 빠진 요구와 인접 축약의 필수 동작이 빠진 표현을 점수 관측 전에 보강했다. 유한 벡터와 한 번의 독립 독해는 일반 의미 정답을 인증하지 않는다. 원천·라벨·동작 mode·역할·ID는 scorer 특징에 들어가지 않으며, 요청과 후보 설명만 사용한다.

정답 있는 요청24개, 검사된 정답 없는 요청12개, 정책이 모호한 unknown12개다. Unknown은 원래48개 분모에 남지만 비용·Top1/3에서 제외한다. No-answer는 기준과 oracle 모두 후보를 전부 검사한다. 정답 집합은 여러 후보를 허용할 수 있다.

## 무학습 결과

역할 분할이 없는 개발 진단이다. 수치는 이 자체 작성 집합의 관측이며 holdout이나 실제 Codex 작업 결과가 아니다.

| 기준 | known36개 전체 검사 수 | answerable24개 검사 수 | Top1 /24 | Top3 /24 |
|---|---:|---:|---:|---:|
| 원래 순서 | 72 | 48 | 4 | 24 |
| BM25 | 59 | 35 | 15 | 24 |
| 단어·연속 단어 | 60 | 36 | 14 | 24 |
| 좁은 문법→BM25 fallback | 59 | 35 | 15 | 24 |
| 정답을 미리 아는 oracle | 48 | 24 | — | — |

최선 무학습 기준은 BM25이며 no-answer 포함 평균 검사 수는59/36≈1.639다. Oracle은48/36≈1.333이다. 가능한 최대 상대 이득은 `(59−48)/59≈18.64%`이고 answerable만 보면 `(35−24)/35≈31.43%`다. **Oracle은 달성한 모델 성능·실행 정책·학습 특징이 아니다.** 5% 필요 상한은 넘지만 학습 가능성을 보장하지 않는다. 작성한 후보 순서와 짧은 대조군 구성이 결과에 영향을 준다. Top3 전부 통과도 작은3~4후보 집합과 중복된 허용 후보의 특성을 포함하므로 별도 의미 이해 성과로 확대하지 않는다.

실제 그룹11<15이므로 역할을 배정하거나 그룹을 쪼개지 않았다. `insufficient_connected_groups`와 `no_partition_or_training_execution_plan`으로 중단한다. FP32·INT8·STE/PTQ fit, HF 가중치 게시, final 채점, production 활성화는0이다. 기존 protected final과 CoSQA reserve도 읽지 않았다.

## Mac 자원 측정

Apple M4 Pro·24GiB·darwin/arm64·Go1.27.1이다. GOMAXPROCS=1로 Go의 동시 실행 P를 제한했으며 OS 스레드 수나 affinity를 측정한 것은 아니다. GPU·ORT·Laya 인코더·결과 캐시는 사용하지 않는다.

같은 원본8후보 예제를 기준마다 새 프로세스 cold3회와 warm1회로 실행했다. Warm의 다섯 단계는 각각20번 준비, AllocsPerRun의101번 allocation probe, 실제2,000번 타이밍 표본이다. 반복 횟수는 데이터 표본 수가 아니다. 예제의 요청은 정규화32단어이고 후보는 최대32단어이며 **모든 가능한512바이트·언어·입력 모양의 최악값은 아니다**.

| 기준 | 전체 Go 호출 warm p95 | 반복 측정 child peak RSS | 전체 호출 Go 할당/op |
|---|---:|---:|---:|
| 원래 순서 | 0.063417ms | 10.5625MiB | 291 |
| BM25 | 0.069667ms | 10.2500MiB | 291 |
| 단어·연속 단어 | 0.077042ms | 10.421875MiB | 291 |
| 좁은 문법 | 0.079250ms | 10.468750MiB | 291 |

전체 호출은 JSON 준비·입력 검증·순위·digest·출력 직렬화를 포함하고 출력은 io.Discard에 보낸다. 순위 단계에는 재검증·특징·점수·정렬이 합쳐져 있다. HeapAlloc은 미수거 객체를 포함한 Go heap snapshot이며 RSS와 다르다. Peak RSS는 **반복 측정 프로세스 전체**로 단일 순간 추론 RAM이 아니다. 32MiB·1ms 목표는 이 예제에서 충족했으며 의미 품질 gate를 대신하지 않는다.

12개 cold process 중 첫 시작은370.932333ms, 나머지는2.564958~5.232667ms였다. 첫 값을 버리거나 원인을 확정하지 않는다. OS·파일 cache·최초 실행 효과는 별도 미확인이다. Cold RSS는 약4.53~4.69MiB였다. 성능 replay wall은1.269162916초다. 네 warm child의 user CPU는0.184671/0.200684/0.214853/0.232318초, system CPU는0.003933/0.003496/0.003057/0.003607초였다. 실제 JSON에는 모든 cold/warm 관측을 보존한다. Open-loop 도착률 부하·동시 다중 요청·stream 지연은 아직 측정하지 않았다.

## 프로파일과 최적화 판단

고정 계획에 따라 좁은 규칙만 별도1회, 단계당10,000회로 CPU 프로파일을 수집했다. Headline 지연/RSS에는 섞지 않았다. 프로파일1.34초·samples1.06초에서 runtime.kevent flat0.88초(83.02%), runtime.madvise0.07초, BM250.02초, lexical token화 cumulative0.05초를 관측했다. **kevent 비중을 애플리케이션의 CPU 병목이나 lock 경합으로 해석하지 않는다.** Darwin syscall stack 귀속의 [Go 원천 이슈80921](https://github.com/golang/go/issues/80921)는 go1.28-devel의 별도 보고이며 현재 go1.27.1 관측의 원인으로 재현하거나 확정하지 않았다.

입력은 고정 배열8개, 단어는 배열32개, 빈도는 작은 연속 배열을 쓴다. Runtime map·공유 가변 상태·공유 lock이 없다. Rank만 선택하면 다른 기준의 점수를 계산하지 않는다. 그러나 정규화·재검증·JSON은 할당을 사용하므로 zero-allocation 도구라고 부르지 않는다. 현재 자료로 SIMD·SoA 재배열의 이득이나 application hotspot은 입증하지 않았으며, 이를 강행하기 전에 프로파일 귀속과 stage별 자원 비용을 검증한다.

## 재현과 다음 작업

아래 성능 계획은 고정된 **Mac arm64 바이너리**용이다. 다른 플랫폼용 실행 파일의 SHA를 같은 계획에 넣어 바꾸지 않는다. 무학습 결과 재생은 Linux/macOS CI에서 수행하고 정수·그룹·순서는 정확하게, 부동 소수점 점수만1e-12 상대 허용오차로 비교한다. 품질 검사가 통과해야 자동 병합되며 별도 사람 승인을 요구하지 않는다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o .cache/riido-shortclaim-56 ./cmd/riido-shortclaim
go run ./cmd/riido-behavioraudit --plan-sha256 9c8461b9fe0a80a0f0976c8f07a68f3a47f0c534800a0a303e1cbd8a613caaeb --out .cache/my-short-audit
go run ./cmd/riido-shortperf --plan-sha256 9c8461b9fe0a80a0f0976c8f07a68f3a47f0c534800a0a303e1cbd8a613caaeb --out .cache/my-short-performance.json
```

새 출력만 허용한다. 원시 pprof·개인 경로·실제 업무 입력은 공개하지 않고 원본 fixture·숫자·hash·함수 집계만 게시한다. 이전 준비 snapshot을 제외한 새 Git JSON 자료·숫자 합계는 약260.6KiB이며 원천/문서까지도64MiB 예산보다 작다. 모델 본체는 Git에 올리지 않는다. [다음 확보 계획56b](NEXT-56b.ko.md)는 원자적 반영·오류 원인·snapshot·수명·escape·그래프처럼 다른 실패 양상을 다루되, 새로운 typed truth 계약과 전이 그룹 검사를 먼저 요구한다. 도메인별 서로 다른 최종2,400요청 목표는 그대로이며 이번 final 적격은0이다.
