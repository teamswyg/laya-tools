# 두 원본 관측기의 독립 검토

원문은 **논리적 요청2개, 입력 벡터9개, 후보×입력 관측23회**의 제한된 관측 설계를 뒷받침합니다. IPNet는 후보3개×벡터5개, hook 합성은 후보2개×벡터4개입니다. 이 검토의 **원본 패키지 초기화, 원본 API 호출, 후보 빌드, 모델 연산, HTTP 요청, 라벨 지정은 모두0**입니다. 원문 읽기·해시·메타데이터 비교는 준비 작업입니다. 먼저 완료한20개 초안 검토는 별도 봉인했고 변경하지 않았습니다.

프로젝트가 작성한 Want9개는 이 유한 범위에서 선택 원문과 모순이 없습니다. adapter는 이를 관측하는 데 필요한 상태·오류 구분을 유지합니다. 이는 정적 결론이며 관측된 정답, 학습 적격성, 일반화 결과 또는 worker 실행 허가가 아닙니다. Root가 바깥쪽 시작 제어, 권리, 원문·빌드, CI 조건을 별도로 완료해야 합니다.

메타데이터 helper는 closure 핀67개와 원문→복사본 binding60쌍을 검증했습니다. 그중57쌍은 Source81 원문·고지를 재사용하고,3쌍은 Source82 고지와 실제 Go 도구 체인의 라이선스·버전입니다. 최초 중간 JSON은60개를 모두 “Source81”이라고 잘못 이름 붙였습니다. 봉인 전 분류를 수정했고 그 중간 JSON은 비공개로 보존했습니다. 저장된 표준 패키지118개와 Root 현재118개의 import path·선택 파일명·imports를 정렬 비교한 결과 전부 동일합니다. Root 현재 GoList 메타데이터는 패키지123개(표준118, 원본3, 관측기·protocol2)를 나타냅니다. pflag42개, mapstructure 루트4개, internal/errors2개 파일을 선택하고 Go1.20 이전 대체 파일2개를 제외합니다. 이 검토자는 GoList를 실행하지 않았습니다.

| 경계 | 정적 발견 | Root 실행 전 조건 |
| --- | --- | --- |
| 시작 | 인수나 readiness를 거부해도 import 패키지는 `main` 전에 초기화됩니다. `pflag.CommandLine = NewFlagSet(...)`은 `func init` 선언 없이도 실행되는 초기화입니다. 해당 선언이 없다는 사실만으로 안전한 시작을 증명할 수 없습니다. | controller는 **Start 이전**에 불변 원문·binary·tool·plan·readiness 핀을 검증하고 독점 outside 예약을 확정해야 합니다. 초기화·핀 검사·관측·출력 저장을 전체 자식 wall/RSS에 포함합니다. |
| flag 상태 | `native.go:83`은 후보×벡터마다 새 FlagSet을 만들고 실제 exported typed constructor와 실제 `Value.Set`을 호출하도록 작성돼 있습니다. `net.IPNet`, `net.IP`, `net.IPMask`를 구분하고 실제 바이트 길이를 복사해 snapshot합니다. | typed 상태를 직접 비교합니다. IP·mask를 network로 재해석하거나 문자열 하나로 줄이거나 실패한 parse가 상태를 바꿨다고 추정하면 안 됩니다. 이 경로는 `FlagSet.Set`을 호출하지 않으므로 `Flag.Changed`는 진단값입니다. |
| hook | `native.go:144`는 원본 OrCompose/Compose constructor와 `DecodeHookExec`에 실제 Value-signature callback을 넘깁니다. trace는 hook 순서·횟수·유효성·kind·정수 입력을 유지합니다. Compose의 nil 출력 후 invalid source Value도 fixture가 안전하게 받습니다. | 고정 int/null 출력과 허용한 callback 동작5개를 유지합니다. `return`에 별도 dynamic-type 필드가 없어 이 범위 밖 임의 Go 타입까지 JSON 값만으로 인증할 수는 없습니다. |
| 오류 | `ErrorChannel`은 presence와 message를 분리하므로 nil과 비-nil 빈 메시지 오류를 구분합니다. type과 `errors.Is(E1/E2)`도 진단으로 기록합니다. | 정확한 `E1\nE2\n` 결합과 빈 hook 목록의 비-nil 오류를 유지합니다. OrCompose가 새로 만든 결합 오류는 E1/E2 동일성을 보존하지 않고 Compose의 첫 오류는 보존합니다. 이 요청에 errors.Join 요구를 추가하면 안 됩니다. |
| panic | recovery는 `Got.panic=true`를 기록하고 worker 자체를 자동 실패시키지는 않습니다. | 결과 반환은 관측 근거이지 의미론적 PASS가 아닙니다. 후속 qualification은 panic·오류·typed 상태·callback 채널을 명시적으로 확인해야 합니다. 자식 exit0를 후보 truth로 바꾸면 안 됩니다. |
| 경로 | worker 핀 검사는 전달된 source root 아래의 symlink 구성 요소를 거부합니다. 절대 source/output 경로와 최종 output 디렉터리도 확인합니다. | outside plan이 canonical root, ancestor/root symlink 정책, source/output/controller 디렉터리 분리와 독점 소유를 동결해야 합니다. `main`이 이를 전부 강제하지는 않습니다. |
| 자원 | worker Go heap·관측 timer는 시작·핀 검사를 제외하며 timer는 호출 사이에서 검사됩니다. | 전체 자식의 Darwin 바이트 단위 OS MaxRSS를 사용합니다. Go heap이나 GPU 메모리가 아닙니다. soft Go heap 제한은 OS RSS 상한이 아닙니다. 바깥에서300초 제한·프로세스 그룹 종료·Wait를 수행하고 정상 반환 직후의 초과도 기록하며 자동 재시도하지 않습니다. |

IPNet의 유한 기대값은 공백·host bits 입력에서 `(192.0.2.0,[255,255,255,0])`, `/32`에서 `(203.0.113.5,[255,255,255,255])`, `/0`에서 주소·mask 모두0으로 성공하는 것입니다. 잘못된 `/33`과 prefix 없는 입력은 비-nil 오류와 초기192.0.2.0/24 상태 보존을 요구합니다. plain-IP와 IP-mask 후보는 자기 타입을 유지하므로 해당 오류·출력을 IPNet 답으로 변환하면 안 됩니다. 이 벡터5개로 IPv6, 빈 입력, 연속 상태 변경, aliasing 또는 mask 문자열 동등성까지 주장할 수 없습니다.

OrCompose의 오류→성공→미사용 hook은9를 반환하고 횟수[1,1,0], 원래 입력[4,4]를 유지해야 합니다. nil 성공은 null·[1,0]으로 멈춥니다. 오류2개는 null·[1,1]·입력[4,4]·`E1\nE2\n`을 요구합니다. hook이 없으면 null과 **메시지가 빈 비-nil 오류**입니다. Compose 후보도 같은 입력을 실제로 관측해야 합니다. 첫 오류에서 멈추고 이전 출력을 전달하며 nil 출력 후 invalid Value를 전달하고, hook이 없으면 초기4·nil 오류를 반환합니다. 이는 나중에 검사할 원문 기반 기대값이며 검토자가 생성한 Got가 아닙니다.

`pure/protocol.go`는 JSON 깊이·배열을 제한하고 중복 member 이름·모르는 typed 필드를 거부합니다.2/3/5와2/2/4 schedule을 고정하고 준비 Got/truth/roles/weights가 null이어야 하며 실제 동결 초안의 요청·revision·후보 설명·source file·Input/Want를 비교합니다. worker는 dispatch 전 draft binding을 사용합니다. 합성 protocol 테스트와 작성자의 순수 source-field/format 검사는 유용한 gate 근거지만 원본 동작 테스트가 아닙니다. 작성자가 보고한62310B dummy 결과는 추정치이며 모든 원본 출력의 상한 증명이 아닙니다. 실제 출력은 별도로 계수·제한·핀 검증해야 합니다.

source closure는 제외된 대체 파일2개를 포함한 원본 Go 파일50개, upstream module 파일2개, 테스트를 포함한 관측기·protocol 원문, root module 파일을 고정합니다. 테스트 원문 핀 보존은 worker에 선택됐다는 뜻이 아닙니다. Go1.23.4 flag 본문은 실행하지 않는 복사 계보 근거이고, 실제 표준 라이브러리 선택은 Go1.27.1입니다. 저장된 패키지 선택은 Root 현재 메타데이터와 일치하지만 GoList·해시만으로 실제 compiler·assembler·linker·cache object 계보를 증명하지는 않습니다. 준비 기본 tool 핀은 Go·compile·link이며 추가 tool 핀과 Root 실제 빌드/cache receipt의 범위를 구분해야 합니다. Go flags·환경·overlay, CGO0, GOOS/GOARCH, trimpath, VCS 미포함을 동결하고 원본 worker를 시작하지 않은 상태에서 결과 build metadata를 검증해야 합니다.

pflag BSD-3-Clause 루트 고지, mapstructure MIT 고지, 실제 Go1.27.1 BSD 라이선스, 별도 historical Go 고지와 Go1.23.4 복사 원문이 보존되어 있습니다. pflag Go2009 파일 헤더와 mapstructure의 제외된 Go2022 유래 join 파일을 패키지 루트 라이선스로 몰래 바꿀 수 없습니다. 후자의 정확한 원본 복사 계보 동등성은 아직 미확인입니다. 문서는 **비공개 불변 원문의 유한 관측**을 원문·worker 배포, 학습, 모델 배포와 분리해서 검토할 근거입니다. 이 모든 용도에 대한 인증은 아닙니다. native 쌍은 MPL 저장소를 선택하지 않으며 기존 MPL 보류는 유지합니다. 선택 import만으로 복사를 증명하거나 실제 복사 계보 연결을 지우면 안 됩니다.

pflag/Cobra connected family와 mapstructure family는 Source53/Source81에 노출된 상태를 유지합니다. 이 요청2개가 새 독립 family2개가 되는 것은 아닙니다. 입력9개는 요청9개가 아니고 관측23회는 학습 부모23개가 아닙니다. role·weight·acceptable index·모델 변경은 지정하지 않습니다. 후보 설명이 원하는 동작을 직접 설명하므로 향후 ranker가 설명 일치만으로 성공할 수도 있습니다. 의미론적 효용이나 LLM 비용 절감을 주장하기 전 lexical/no-hint 대조가 필요합니다.

대안은 세 가지입니다.

- **A: 원문 준비 상태 유지.** 자원 사용이 가장 적고 시작·CI·권리·빌드 gate가 미해결인 동안 적합합니다. 동작 근거는 늘어나지 않습니다.
- **B: outside plan을 동결한 후 Root가 제한 관측1회.** 명시한 모든 gate 통과 후 가장 적합합니다. 고정 dispatch23회로 유한 Want9개를 확인하고 typed 상태·오류·callback 원래 근거를 비공개 보존합니다. 라벨은 자동으로 생기지 않습니다.
- **C: 지금 관측기 확대·변경.** 이후 범위를 넓힐 수 있지만 고정 실행 전에 하면 실험을 바꾸게 됩니다. 추가 타입·callback·벡터는 B 검토 후 별도 버전 plan으로 분리합니다.

권고는 **Root 전체 gate 완료 후 B, 그전에는 A**입니다. 현재 고정 범위의 유한 Want·adapter 정적 검토에서 모순은 찾지 못했습니다. 시작 예약, 정확한 build/tool/cache 인증, CI, 로컬 권리, canonical 경로·자원 강제는 Root 조건으로 남습니다. 게시·학습 qualification·모델 평가는 각각 후속 독립 판단입니다.
