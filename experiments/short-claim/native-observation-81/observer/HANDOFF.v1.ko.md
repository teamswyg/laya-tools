# Source81 관찰기 준비 인계

여기서 원 패키지 초기화0은 mapstructure/pflag 관찰기 바이너리를 실행하지 않았다는 뜻입니다. 준비 helper·대역 테스트 프로그램의 일반 stdlib 초기화를 전부0이라고 주장하지 않습니다. API/모델0은 이번 실험의 명시 경로이며 AI 보조 대화의 비용은 미측정입니다.

네 행동 목표·32개 입력에 대해 **관찰기 소스와 컴파일만 준비**했습니다. 원천은 mapstructure와 pflag 연결 family 두 개이며, 입력32개를32개 요청이나 독립 그룹으로 세지 않습니다. 원 함수·패키지 initializer·관찰기 main/--help는 아직 실행0입니다. Want32 객체·포인터·순서·null과 기존 감독/역할은 수정하지 않았습니다.

원문57개와 cached Go1.27.1 근거8개를 byte-exact로 복사하고 모듈/소스의 .txt 확장자만 제거했습니다. MIT·pflag BSD-3-Clause·Go BSD 전문과 원문 고지를 그대로 유지했습니다. !go1.20인 mapstructure 두 파일의2022 Go Authors 고지도 보존되며, 컴파일되지 않는다고 원문 권리 고지가 사라지지 않습니다. Go1.23.4 flag 원문은 참고 자료이며 실행 중인 Go1.27.1 stdlib로 바꿔 주장하지 않습니다. private harness go.mod의 require 버전은 로컬 replace를 위한 작성자 설정이고, 실제 원천 revision/bytes는 핀으로 고정합니다.

offline Go1.27.1 go list 메타데이터에서 mapstructure root4·internal/errors2·pflag42, 실제선택48·제외2를 확인했습니다. [BUILD-SELECTION.v1.json](BUILD-SELECTION.v1.json)은 경로를 제거한 별도 투영입니다. 이것은 파일 선택 근거이며 initializer 실행이나 전체 stdlib의 hermetic closure 증명이 아닙니다. [PIN-MANIFEST.v1.json](PIN-MANIFEST.v1.json)은 원문57+cached8+문서4=69개 exactcopy를 기록합니다. 원문/참고 및 자체 소스7개를 합친 선언 closure는76핀입니다.

Hook은 factory가 반환한 실제 any 함수값을 DecodeHookExec에 넘깁니다. source builtin int 문자열7은 int7로, 대상은 유효한 zero reflect 값과 진짜 정의된 Time/[]string 타입으로 준비합니다. 매 입력 fresh pflag FlagSet(scope81, ContinueOnError), items 등록·Lookup1회, 등록 뒤/각 작업 뒤 GetSlice를 호출하고 전체 ordered slice를 소유 복사합니다. nil과 nonnil-empty를 구분하며 public Flag.Changed만 읽습니다. GetStringSlice/Value.String/Type/alias mutation은 추가하지 않습니다.

예상 명시 호출은182, 보수적 최대338입니다. 사전 예상 원천 경로148에 outer Error6·Unwrap6을 더해 original-public160, cause Error6·time getter4×4를 더해 stdlib-explicit22입니다. 성공/오류 여부에 따라 실제 호출 수는 달라질 수 있으므로 예약·시도·정상 반환·panic을 각각 기록하고 차이를 결과에 보존합니다. returned time.Time은 zero-on-error도 Date/Clock/Nanosecond/Zone 각1회만 관찰합니다. 에러는 outer Error→가능한 Unwrap1→nonnull cause Error1 및 공개 ParseError 필드/CSV sentinel identity만 유지합니다. Err.Error나 전체 cause-chain을 추가로 관찰하지 않습니다. 예상과 다른 타입은 실제 타입과 지원되지 않는 채널을 남깁니다.

호출 전 intent/counter를 partial 파일에 저장·동기화하며 체크포인트 실패 뒤에는 해당 호출을 수행하지 않습니다. 단계별 undispatched 필드는null로 유지합니다. primary/getter/error panic을 나누고 모든 panic 사건을 기록합니다. 문자열 panic만 원문 text로 보존하고 비문자열에 Error/String/Format을 호출하지 않습니다. Want/Got 차이는 completed_with_differences로 남기며 기대값을 재작성하거나 재시도하지 않습니다.

원본 import가 포함되지 않은 pure 패키지만 대역 race2회(처음14개, 최종15개) PASS, skip/fail0입니다. vet은3회 중 첫unused-import 실패1을 원문과 함께 보존하고 최종PASS; compile-only build1 PASS입니다. 대역 결과는 원 API 관찰이나 Want 정답 승인으로 바꾸지 않습니다. main의 유효 실행 인자 경로는 아직 실행/테스트하지 않았습니다.

[BUILD-IDENTITY.v1.json](BUILD-IDENTITY.v1.json)은 Go1.27.1/darwin-arm64/CGO0/trimpath/buildvcsfalse 및 바이너리·선언 closure 핀을 기록합니다. -ldflags에 선언 closure SHA를 연결하지만 독립 compiler attestation이나 런타임 성공을 보장하지 않습니다. PLAN.v1.draft.json은 frozen=false이며 물리 host root가 들어 있어 공개하지 않습니다. Root만 사전 peer 검토 후 frozen 토큰false→true로 봉인하고 바깥에서 fresh O_EXCL/fsync 예약을 child Start보다 먼저 해야 합니다. 원문 package init는 main 전에 일어나므로 그 전체 수명은 밖의60초 guard가 담당하며 실제 init callback/return 계수는미계측 null입니다.

Root 실행 인자는 --plan, --plan-sha256, --attempt-dir이고 worker 하위 경로는 fresh여야 합니다. 결과는 results.json/results.partial.json/reservation.json; 성공 final·partial·stdout은 같은 JSON bytes입니다. 환경은 CPU1·GOMEMLIMIT268435456·Go1.27.1이며 전체 외부 수명60초·stdout/workerJSON1MiB·retry0입니다. Go soft heap은 OS RSS hard cap이 아니며 SIGKILL 뒤 joined Wait의 엄격한 반환시간을 보장하지 않습니다. 시간/RSS·함수별 비용·모델 추론 성능은 아직 측정하지 않았습니다.

Want author와 observer author는 다르지만 이 작성자는 이전 프로젝트/Want80 자료에 노출된 AI 보조 비맹검 작성자입니다. 학습 label/역할/weight/parent 생성0, Features/Project/Fit/모델/paid/HF/공유 저장소 편집/외부 게시0입니다. qualification/training/production/protected-final false를 유지합니다. [PREPARATION-LEDGER.v1.json](PREPARATION-LEDGER.v1.json)에 준비 실패와 대역/원문 실행의 경계를 기록했습니다.
