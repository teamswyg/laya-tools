# laya-tools · User Guide / 사용자 안내

**필요한 기능부터 선택해 사용하세요. / Start with the feature you need.**

[결과 파일·외부 검증 연결 검사](https://github.com/teamswyg/laya-tools/blob/ff391d5630c272258a8223a8f00d6e5b894ae9b8/experiments/short-claim/next60-durable-evidence/README.ko.md) · [Result file and outside verification pipeline](https://github.com/teamswyg/laya-tools/blob/ff391d5630c272258a8223a8f00d6e5b894ae9b8/experiments/short-claim/next60-durable-evidence/README.en.md): 고정 배열 기반 준비 코드와 실제 임시 파일/저널의 Sync·응답 연결을 93개 테스트·152개 하위 사례로 확인했습니다. API는 가짜이며 새로운 원본·모델 실행이나 학습 자료 추가는 0입니다. 잘못된 최대 파일의 Go 할당량은 약 91.9% 줄었고 정상 파일의 속도 개선은 입증하지 못했습니다. CI 검토 중입니다. / Fixed-array preparation and actual temporary-file/journal Sync/ACK passed 93 controls and 152 subcases. APIs remain owned fakes; new original/model runs and training additions are zero. Allocated bytes fell about 91.9% for a malformed maximum file, with no established valid-file speed gain. CI review is pending.

[검증 기록을 작게 저장하는 Go 준비 코드](https://github.com/teamswyg/laya-tools/blob/82ee3148de0bc1a7b3716e51c2397789150bed30/experiments/short-claim/next60-compact-evidence-preparation/README.md) · [Compact Go evidence preparation](https://github.com/teamswyg/laya-tools/blob/82ee3148de0bc1a7b3716e51c2397789150bed30/experiments/short-claim/next60-compact-evidence-preparation/README.en.md): 확인되지 않은 값은 그대로 남기면서 검증 결과 저장을 줄이는 작업입니다. 공개 합성 검사 34개와 저장소 전체 race 검사를 통과했습니다. 한 행의 직렬화 실험은 중앙값 12.958µs→1.133µs, 할당량 5,888→586B/op였지만 할당 횟수는 3→7로 늘었습니다. 실제 원본 실행·모델 추론·전체 작업 비용 개선은 아직 검증하지 않았습니다. / This work reduces stored evidence while preserving unknown values. All 34 public synthetic controls and the full repository race checks passed locally. A matched one-row serialization microbenchmark measured median 12.958µs→1.133µs and 5,888→586B/op, with allocation count increasing from 3 to 7. Original execution, model inference and whole-work savings remain unverified. [변경과 CI / Change and CI](https://github.com/teamswyg/laya-tools/pull/130).

[현재 개발 자료 35개·실제 HF 게시](Next60-Development-KO) · [35 verified requests and actual HF publication](Next60-Development-EN): 요청 35·라벨 102, 새·변경 84파일과 표 35행 전체 값 일치, 새 학습 0회. / 35 requests, 102 labels; 84 new/changed files and all 35 viewer rows matched; zero new Fits.

[개발용 학습 자료 2개 확정](Native2-Training-KO) · [Two qualified development-training requests](Native2-Training-EN): 후보 정답 5개·기존 입력 제한 통과, 새 학습 0회 / Five candidate labels, existing input limits passed, zero new fits.

[첫 두 과제의 실제 동작 확인](Native2-Observation-KO) · [Actual checks for the first two requests](Native2-Observation-EN): 기존 초안의 요청 2개·입력 9개·관측 23개를 실제 Go 원본으로 확인했습니다. 자식 최대 RSS 약 9.11MiB·전체 시간 약 0.37초입니다. 관측 당시 적격은 0개이며 후속 결과는 위의 학습 자료 안내를 보세요. / Two existing draft requests, nine inputs and 23 observations were checked against original Go behavior. Child maximum RSS was about 9.11MiB and whole-child time about 0.37s. Qualification was zero at observation; subsequent admission is linked above.

[코드 동작에서 학습 정답 만들기](Claim-Behavior-Validation-KO) · [Turning behavior into training labels](Claim-Behavior-Validation-EN): 요청3개 추가 후 같은 조건으로 실제 학습했지만 확인31회로 단순 정렬27회를 넘겨 실패했습니다. 새 모델은 비활성입니다. / Adding3 verified requests under the same recipe yielded31 checks versus the lexical control's27; utility failed and the model remains inactive.

[공개 함수 동작 검증과 학습 자료 확장](Source-Behavior-80-KO) · [Public behavior verification and training data](Source-Behavior-80-EN): 사전 기대값24개 일치, 새 학습 정답0 /24 matching fixed expectations, zero new training labels.

[다음 학습 자료를 만드는 과정](Claim-Data-Preparation-KO) · [Preparing the next training data](Claim-Data-Preparation-EN): 함수 예제와 사용자 요청, 후보의 실제 동작과 설명·학습 적격성을 구분합니다. / Separate function fixtures from user requests, and candidate behavior from caption fidelity and training eligibility.

| 사용 목적 / Goal | 한국어 | English |
|---|---|---|
| 모델 없이 코드 검색·실행 계획 시작 / Start without a model | [처음 시작하기](Getting-Started-KO) | [Getting started](Getting-Started-EN) |
| 설치·검색·라우팅·저장소 선택·에이전트 인터페이스 / Commands and interfaces | [전체 사용법](https://github.com/teamswyg/laya-tools/blob/main/README.md) | [Full usage](https://github.com/teamswyg/laya-tools/blob/main/README.en.md) |
| 모든 후보를 남기는 작은 동작 힌트 / Small hints that retain every candidate | [모델 없는 Go 도구](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.ko.md) | [Model-free Go tool](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.en.md) |
| 라이선스·공개 범위 / Licensing and publication scope | [라이선스 안내](Licensing-and-Project-KO) | [Licensing guide](Licensing-and-Project-EN) |

Codex 연동은 선택 사항입니다. 현재 도구의 추천은 실험이며 실제 사용량·비용 절감을 입증하지 않았습니다. / Codex integration is optional. Recommendations are experimental; actual usage or cost savings have not been established.

작은 주장 모델 연구에서는 확인 순서를 제안하고 독립 검증을 남기는 방식을 시험합니다. 새 데이터 추가 모델과 기존 두 모델은 효용 검사를 실패해 비활성입니다. / Tiny claim research proposes a verification order while retaining independent checks. The new data-addition model and the two earlier models failed utility gates and remain inactive.

- [최근 실제 비교76](Claim-Diagnostic-76-KO) · [Latest actual diagnostic76](Claim-Diagnostic-76-EN)
- [선택형 Go 압축 가중치: 저장 공간과 CPU 비용](Packed-Weights-78-KO) · [Optional packed Go weights: storage and CPU tradeoff](Packed-Weights-78-EN)
- [자료의 동작·설명을 구분하기](Source-Observation-73-KO) · [Separating source behavior and descriptions](Source-Observation-73-EN)
- [저장 형식과 메모리의 차이](Compact-Storage-73-KO) · [Storage size versus memory](Compact-Storage-73-EN)
- [Go 계산 개선: 같은 점수·할당량, 두 예제 약20~24% 시간 감소](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/feature-prefix-74) · [Identical-score Go optimization](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/feature-prefix-74/README.en.md)

이전 안내의 측정값·실패·당시 진행 상태는 원문 그대로 보존했습니다. / Earlier measurements, failures and status-at-the-time remain verbatim:

- [시점별 연구 안내 원문 / Previous chronological guide](Research-History-75)
- [기존 전체 연구 링크 / Previous full research index](Research-Index-75)
- [전체 문서 목록 / Full documentation index](https://github.com/teamswyg/laya-tools/blob/main/docs/README.md)
