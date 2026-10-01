# PDCA55: 외부 Go 요청 두 개의 실제 프로필 비교

[English](RESULTS-55.en.md)

humanize의 전체 int64 서수 API와 UUID의 엄격한 소문자 파서, **서로 다른 개발 요청 두 개**를 Sol6 low와 Luna low에 각각 한 번씩 맡겼습니다. humanize는 두 후보 모두 수용됐고, UUID는 Sol6 후보가 수용되고 Luna 후보가 거절됐습니다. 네 CLI는 모두 정상 종료했고 전체 기본 사용량을 확인했습니다. 프로세스가 정상 종료하는 것과 요구를 충족하는 것은 별개입니다.

네 프로필 시도나 humanize 6개·UUID 218개의 테스트를 고유 요청 수에 더하지 않습니다. 두 요청의 한 번씩 관측으로 모델 서열이나 금액 절감을 결론 내리지 않습니다.

## 결과를 보기 전에 고정한 조건

[humanize 계획](plan-55-humanize.json)과 [UUID 계획](plan-55-uuid.json)은 원천 revision이 달라 별도로 만들고, [전체 예산](parent-budget-55.json)이 네 슬롯을 묶었습니다. 계획은 커밋 `fe7f3c68df1b299d04a9115081c52fd4d81afd9b`에 먼저 고정했습니다. 순서는 **humanize/Sol6 → UUID/Luna → humanize/Luna → UUID/Sol6**, 동시 실행 1개, 모델 실행 120초와 별도 독립 검사 45초 한도입니다. 실행기의 재시도·resume·fallback은 없었습니다.

2번과 3번 사이에는 사용자가 요청한 휴식과 읽기 전용 upstream 검토가 있었고, 명시적 재개 후 남은 순서대로 실행했습니다. 휴식은 실패 슬롯을 다시 실행하거나 새 ledger를 만드는 근거로 사용하지 않았습니다. 이 간격과 요청별 관측 하나라는 조건 때문에 시간을 통제된 일반 벤치마크로 해석하지 않습니다.

[최종 receipt](parent-receipts-55.json)는 **예약 4개·durable 시작 marker 4개·종료 receipt 4개·시작 상태 미확인 0개**를 기록합니다. 예산은 소진됐고 활성 실행은 없습니다. 네 실행 모두 자식 프로세스 그룹 정리와 복사 인증의 제거 상태를 확인했습니다. 네 슬롯의 환불은 없습니다. 이 cap은 동일한 private durable ledger 하나에 적용하며 호스트·계정 전체나 공급자 내부 호출·재시도 횟수의 한도는 아닙니다.

실행기는 CI를 통과한 검토 head `c8d1cf66413b696b7c0872c01df8155e3cca29d4`와 tree가 같은 merge `e30da75c5d1ea381924ac5d54e8bcbb9a2f462c8`에서 깨끗하게 빌드했습니다. [당시 CI](https://github.com/teamswyg/laya-tools/actions/runs/36816995319)는 framework 검사입니다. 네 실제 코딩 실행은 로컬 Mac의 Codex CLI를 통해 진행했고, 기존 ChatGPT 로그인만 사용했습니다. 새 API 키·유료 cloud 작업·endpoint는 만들지 않았습니다.

## 모델에 전달한 요구와 Go 조건

이번 v2의 실제 stdin은 고정된 전체 `Spec.Prompt`입니다. 변경 가능 파일, legacy 보존, 허용 형식과 검증기의 지원 범위를 전달했습니다. humanize는 3,150 bytes, UUID는 5,745 bytes이며 계획의 Prompt SHA와 연결합니다. v2는 이미 준비한 두 논리 요청의 설명·평가 버전입니다. 버전 ID 두 개가 새 논리 요청 두 개를 만든 것은 아니며, 이번에 두 요청을 처음 실제로 시도했습니다.

| 원천 요청 | 고정 revision | 원본 언어 | Go 실행 파일 | 계약의 범위 |
| --- | --- | --- | --- | --- |
| humanize 전체 int64 서수 | `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e` | Go 1.21 | Go 1.27.1 | 새 API의 음수·끝값·suffix와 기존 Ordinal 보존 |
| UUID 소문자 파서 | `2d3c2a9cc518326daf99a383f07c4d3c44317e4d` | directive가 없어 암묵적 Go 1.16 | Go 1.27.1 | 소문자 36-byte 문법, 오류 시 zero UUID, 기존 API와 상태 보존 |

독립 검사기는 원본 `go.mod` bytes와 모듈·recipe SHA를 사용했습니다. 후보 모듈 설정과 후보가 작성한 테스트는 성공 근거로 사용하지 않았습니다. UUID는 고정 원본 또는 trusted formatter prefix 뒤에 함수 하나를 추가하는 제한된 형태를 지원합니다. 추가 helper·별칭을 통한 호출·동시성·전역 변경 등 지원 밖 형태는 올바른 출력이라도 `verifier_unknown`이며 행동 실패로 바꾸지 않습니다. [평가 조건 설명](../../docs/fair-upstream-comparison.ko.md)

원본 MIT·BSD-3 LICENSE는 별도로 검증해 고정 복사했습니다. 후보 LICENSE 내용은 성공 라벨의 평가 대상이 아니며 accepted가 후속 배포의 attribution 준수를 인증하지 않습니다. 유한한 계약 통과를 전체 upstream 정확성이나 모델 학습·가중치 배포 권리로 확대하지 않습니다.

## 실제 관측

| 순서 | 요청 | 요청 프로필 | 모델 실행 ms | 별도 검사 ms | 입력 | 입력 중 캐시 | 출력 | 출력 중 reasoning | 독립 결과 |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | humanize | Sol6 low | 76,404 | 4,170 | 59,824 | 44,928 | 1,005 | 83 | 수용·terminal pass 6개 |
| 2 | UUID | Luna low | 98,471 | 5,852 | 152,941 | 129,536 | 2,126 | 0 | 거절·pass 수 미집계 |
| 3 | humanize | Luna low | 20,836 | 4,646 | 45,866 | 39,936 | 649 | 0 | 수용·terminal pass 6개 |
| 4 | UUID | Sol6 low | 43,474 | 5,421 | 82,301 | 66,304 | 1,347 | 231 | 수용·terminal pass 218개 |

합계 입력 **340,932**에 캐시 **280,704**가 포함되고 출력 **5,127**에 reasoning **314**가 포함됩니다. 각각 부분집합이므로 다시 더하지 않습니다. 모델 실행 합계 **239,185 ms**와 별도 검사 **20,089 ms**는 부모 준비·사전 Laya 관측·휴식 시간을 포함하지 않습니다.

humanize에서는 Luna 시도의 입력·출력과 실행 시간이 더 작으면서 같은 계약을 통과했습니다. UUID에서는 Luna 후보가 거절되고 Sol6 후보가 통과했으며 관측 실행 시간도 달랐습니다. 이 두 관측은 요청에 따라 결과가 달라질 수 있음을 보여주며 일반 모델 순위나 성공 확률의 라벨은 아닙니다. 실패 후보의 수리·상위 프로필로 재시도하는 비용은 실행하지 않았습니다. 구독 소모율·청구 단가·공급자 내부 호출과 재시도 횟수는 미확정이라 금액 절감을 주장하지 않습니다.

[공개 수치와 해시](results-55.json)는 네 [시도 기록](attempt-55-01.json)의 raw bytes와 receipt에 연결합니다. 나머지 기록은 [2번](attempt-55-02.json), [3번](attempt-55-03.json), [4번](attempt-55-04.json)입니다. 요청 프로필은 명시되지만 공급자가 증명한 실제 모델 정체는 네 실행 모두 `unknown`입니다. 모델 원문 trace·인증·개인 코드·로컬 경로·가중치는 공개하지 않습니다.

### UUID Luna 후보가 거절된 이유

종료 후 소스만 검토한 결과, 올바른 nonzero prefix 뒤에 잘못된 hex나 뒤쪽 separator가 나오면 오류와 함께 **일부 채워진 UUID**를 반환합니다. 모델에 공개한 요구는 모든 오류에서 zero UUID를 반환하는 것이므로 위반입니다. 이 분석을 위해 모델을 다시 실행하거나 후보를 고쳐 결과를 덮어쓰지 않았습니다.

원본 verifier의 `independent_tests: 0`은 독립 `go test`가 비정상 종료했을 때 terminal pass 집계 전에 반환하는 카운터입니다. **검사가 실행되지 않았거나 218개가 모두 실패했다는 뜻이 아닙니다.** 공개 집계에서는 pass 수를 `null`로 구분합니다. 격리 검사·지원 형태·포맷 gate를 통과한 뒤 행동 검사가 거절한 결과입니다.

## 코딩 전에 관측한 Laya

[사전 예측](routing-predictions-55.json)은 커밋 `1b87612d7848ada919356c9dbe5f1e04f2993ea8`에 먼저 고정했습니다. 전체 Prompt bytes를 그대로 입력하고 기존 INT8 base, CPU 한 스레드, 임계값 0.9, 512-token 한도를 유지했습니다.

| 요청 | 전체 입력 bytes | standard 점수 | 입력 잘림·보류 | cold 프로세스 ms | 앱 관측 ms | 최대 RSS bytes |
| --- | ---: | ---: | --- | ---: | ---: | ---: |
| humanize | 3,150 | 0.502708097 | 모두 true | 1,595 | 1,551.182 | 1,544,044,544 |
| UUID | 5,745 | 0.477323446 | 모두 true | 1,337 | 1,304.806 | 1,581,891,584 |

두 입력 모두 질문·헤더·선택지까지 포함한 전체 512-token 한도에서 잘렸고 Laya는 보류했습니다. 설정상 strong/Astra를 유지하지만 이번 실제 비교는 별도로 고정한 Sol6/Luna 순서입니다. Astra를 실행하거나 실제 라우팅 정책을 활성화하지 않았습니다. 입력 SHA가 전체 Prompt와 같아도 encoder가 전체 내용을 읽었다는 뜻은 아닙니다. 이 점수는 보정된 코딩 성공 확률이 아닙니다.

약 1.54/1.58 GB의 최대 RSS와 시간은 native 로딩을 포함한 **cold CPU 전체 프로세스** 관측입니다. Go heap·GPU 메모리·remote 코딩 모델 자원·warm 지연과 구분합니다. 현재 관측은 초저자원 목표 달성을 보여주지 않습니다.

## upstream 검토와 다음 방향

고정 Laya의 [router preset](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/laya/presets.py#L177)은 난이도·도메인·도구 필요·민감성을 질문합니다. [라우팅 예제](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/examples/29_presets_model_router.py)는 수작업 입력 3개에 threshold 정책을 적용합니다. [model-routing benchmark](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/research/scripts/bench_apps.py#L243)는 399개 입력의 도메인 분류이며 특정 Codex 프로필의 완료·사용량 정답이 아닙니다.

따라서 짧은 주장과 힌트를 점수화하는 구조는 참고할 수 있지만, 모델 선택의 유용성은 실제 완료 결과와 실패 비용으로 별도로 검증해야 한다는 판단입니다. 이번 full Prompt는 잘려서 보류됐으므로, **짧은 주장 입력과 encoder 없이 실행하는 Go 규칙·작은 계수 기준선**은 별도 PDCA56의 계획으로 비교합니다. 이번 입력을 나중에 짧게 바꿔 같은 실험의 성공으로 처리하지 않습니다. 새 실험의 성능은 아직 알 수 없습니다.

## 규모와 적격성

실제 누적 개발 관측은 **고유 요청 7개·CLI 기록 20개·코드 가족 5개·저장소 3개**입니다. 준비된 논리 요청 수를 늘린 것은 아니고, 처음 시도한 외부 두 요청으로 실제 관측 범위가 5개에서 7개로 늘었습니다. 과거 실패·시간 초과·지원 오류와 [원본 120개 후보](../../docs/public-go-acquisition-53.ko.md)는 그대로 보존합니다. 120개 후보는 120개 정답이나 실행 가능한 계약이 아닙니다.

이번 두 요청은 개발에 이미 읽은 자료입니다. **학습 실행·학습 적격 라벨·보호된 final 적격 요청·새 가중치 배포·생산 정책 변경은 모두 0**입니다. 테스트 수나 반복·번역·형제 과제를 별개 최종 요청으로 세지 않습니다.

24개 초기 개발 시험 다음에는 여러 저장소에서 독립 검사를 갖춘 개발 요청 **120개 이상**, 학습 신호 검토 **240개 이상**으로 확장합니다. 별도로 확보한 **주장 영역마다 서로 다른 보호된 최종 요청 2,400개 이상**이 최종 목표입니다. 이 단계들은 아직 달성하지 않았습니다. [규모 설명](../../docs/golden-set-scale.ko.md)과 [확보 계획](../../docs/golden-set-acquisition.ko.md)에 따라 자료와 그룹을 나눕니다. 이번 주기에 학습·보정·새 모델 배포·protected final 채점은 하지 않았습니다.
