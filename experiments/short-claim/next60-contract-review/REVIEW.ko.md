# next60 첫 계약의 독립 검토

첫 20개 초안은 원문에 근거한 설계 자료이지만, **원본 실행·라벨·학습 준비가 완료된 계약은 0개**입니다. 요청 20개, 유한 입력 96개, 후보 설명 59개를 읽기 전용으로 검토했고, 기존 Source53의 ID·문구·저장소·리비전 연결 18개와 공개 원문/메타데이터의 바이트 핀 38개를 확인했습니다. 이 검토에서 원본 실행, 후보 빌드, 실험/모델 호출, 새 라벨과 외부 쓰기는 모두 **0회**입니다. 계약을 읽은 횟수와 실험한 횟수는 다릅니다.

가장 강한 반론은 이 묶음이 기존 목록과 쉽게 알아볼 수 있는 API 표현을 확장하는 데 그칠 수 있다는 것입니다. 모델이 저렴한 주장 우선순위를 잘 정하는 대신 함수 이름이나 우리가 작성한 설명을 외울 수 있습니다. native 2개는 새로 작성한 요청이지만 새 소스 계열을 추가하지 않습니다. 입력 사례가 많으면 한 계약 안의 잘못된 구현을 더 잘 구분할 수 있으나, 독립 요청 수나 일반화 근거가 늘어나는 것은 아닙니다. 따라서 현재의 적격 0개 상태가 타당합니다.

불변 입력은 `DRAFT-CONTRACTS.v1.json`, 209,341 B, SHA-256 `c7bbf7bcc7912300fabdd6e671b0922df295299f00e5e7201df1ab5e824a2a9b`입니다. 공개 최종 핀 파일은 3,245 B, SHA-256 `0c3df33513224e0bcafa4687c49a255167eca4c536061f6ba93677c83c6db070`입니다. 작성자의 첫 메시지에 다른 길이 추정이 있었지만 작성자가 정정했고 독립 실측도 209,341 B와 일치했습니다. 이 검토는 기존 봉인을 수정하지 않습니다. `SOURCE-READ-RECEIPT.v1.json`, `METADATA-QA.v1.json`, `OLD79-COMPARISON.v1.json`에 원문이나 비공개 실험 출력 복사 없이 근거를 남겼습니다.

## 수치의 의미

| 항목 | 결과 |
|---|---:|
| 막힘을 포함한 원문 계약 검토 완료 | 20개 |
| 기존 Source53 ID·문구·저장소·리비전 일치 | 18개 |
| 새로 작성한 native 논리 요청 초안 | 2개 |
| 유한 입력 / 후보 설명 | 96개 / 59개 |
| 저장소 / 잠정 연결 계열 | 7개 / 6개 |
| 독립성이 증명된 새 계열 | 0개 |
| 적격 요청 / 실제 새 부모 / 라벨 | 0개 / 0개 / 0개 |
| 이 검토의 baseline·수정·잘못된 후보 실행 | 0회 / 0회 / 0회 |
| native 실행·Fit·Project·Features·Score·모델 Decode/Encode | 모두 0회 |

이 검토는 일반적인 AI 협업으로 작성했습니다. 모델 호출 0이라는 표현은 별도의 실험 worker, 판정/API 시험, 모델 연산이 없었다는 뜻입니다. 협업 자체에서 AI를 사용하지 않았거나 비용이 없었다는 주장이 아닙니다.

## 계약별 발견

모든 행은 관측기·실제 구현·유한 호환성 검사·권리 및 소스 범위 검토가 남아 있습니다. 아래는 각 계약의 구분 조건과 남은 경계이며, 정답 후보 인덱스나 truth를 부여하지 않습니다.

| 초안 접미사 / 기존 Source53 ID | 원문 근거와 다음 유한 검사 |
|---|---|
| `afero-sub-validation` / `go53-afero-iofs-sub-validation` | 현재 Sub는 뷰를 만들고 nil 오류를 반환합니다. 빈 이름·상위 경로·절대 경로로 새 사전 검증을 구분할 수 있습니다. `path.Clean` 잘못된 후보는 빈 이름이 `.`으로 바뀌는 경우에는 맞지만, `../a`는 정리해도 여전히 유효하지 않습니다. 다른 실수를 실제로 추가하지 않는 한 그 설명은 부정확합니다. 모든 잘못된 입력의 PathError Op/Path와 오류 동일성, 유효하지만 없는 하위 경로의 나중 Open 실패도 고정해야 합니다. |
| `humanize-strict-comma` / `go53-humanize-strict-comma-parse` | 쉼표를 전부 없애면 `12,34`도 통과하므로 의미 있는 baseline 구분입니다. 선행 0이 있는 묶음이 canonical인지 정하고, +부호·공백·빈 값·비ASCII 경계를 추가해야 합니다. 범위 오류 때 값 0은 새 API 정책입니다. ParseInt의 오류 시 포화 값을 그대로 넘기면 안 됩니다. |
| `ini-quoted-comments` / `go53-ini-quoted-comment` | 기본 주석 탐색은 둘러싼 따옴표 제거보다 먼저이고, UnescapeValueDoubleQuotes는 LastIndex를 쓰는 별도 경로입니다. 5개 Want는 새 strict 모드 명세에 맞습니다. 기존 옵션과 원시 바이트 reader adapter를 고정해야 합니다. 이 API의 문자열 source는 INI 본문이 아니라 파일명입니다. 이스케이프 백슬래시와 옵션 OFF의 실제 기대값도 필요하며 새 옵션 부재만으로 baseline을 거부하면 안 됩니다. |
| `ini-delete-index` / `go53-ini-section-delete-index` | 없는 인덱스는 현재 nil을 반환하고 음수 인덱스는 안전하지 않은 인덱스 조정에 도달합니다. 매 입력에 새 중복 section을 만들면 구분 가능합니다. 모든 잘못된 입력에서 section 이름만 보지 말고 키 값·이름·순서·전체 sectionIndexes도 기록하고 기본 section 및 이름 정규화 호환성을 정해야 합니다. |
| `mapstructure-key-collision` / `go53-mapstructure-normalized-key-collision` | 원본은 정확한 키를 먼저 찾고, 아니면 첫 MatchName 결과를 택합니다. 다른 값뿐 아니라 같은 값의 중복도 새 ambiguity 정책에서 오류여야 합니다. 정확한 대소문자 키와 변형 키가 같이 있는 입력, 정렬된 충돌 이름, 대상 필드 불변을 검증해야 합니다. 소스를 읽은 것으로 map 순서 경쟁을 관측했다고 쓰면 안 됩니다. |
| `pflag-map-snapshot` / `go53-pflag-map-snapshot` | 원본 getter는 map 저장 공간 공유를 명시합니다. 어느 쪽을 변경해도 반대쪽이 유지되는지로 소유한 복사본을 구분합니다. 등록 방식, nil과 비nil 빈 map, getter 타입 오류, 기존 shared-map getter 보존을 고정해야 합니다. map 헤더 복사는 소유권 있는 복사가 아닙니다. |
| `mapstructure-nil-config` / `go53-mapstructure-decoder-nil-config` | NewDecoder(nil)은 config를 역참조하지만 nil Result와 typed-nil Result는 이미 Metadata/default 초기화 전에 거부합니다. **config guard만 추가한 코드가 typed-nil panic을 일으킨다는 설명은 맞지 않습니다.** 오류 sentinel/type과 실제 잘못된 후보를 명시해야 합니다. 예를 들어 Metadata를 너무 일찍 바꾸거나 잘못된 reflection 접근을 직접 넣어야 합니다. 기존 안전 처리를 새 개선으로 기록하지 않습니다. |
| `ini-byte-budget` / `go53-ini-parse-byte-budget` | 4 B source 두 개의 총합 제한은 source별 제한과 구분됩니다. 파일명 문자열이 아닌 []byte/reader를 써야 합니다. 0/무제한, limit+1, reader 오류와 초과 동시 발생의 우선순위, Close 소유권과 예산 유효기간을 정해야 합니다. “기존 동작 보존”이라는 추상 Want는 유한 관측값이 아닙니다. |
| `cobra-required-annotation` / `go53-cobra-required-annotation` | 빈 annotation을 [0]으로 접근하므로 빈 값·알 수 없는 값·여러 원소가 검증 추가를 구분합니다. true+Changed=true, annotation 없음, DisableFlagParsing, 여러 flag 오류의 정렬과 불변도 추가해야 합니다. false가 required 검사를 건너뛰는지도 명시합니다. |
| `humanize-fractional-bytes` / `go53-humanize-exact-fractional-bytes` | 소수 입력은 float64 곱셈을 쓰지만 정수는 이미 별도 검사된 정수 경로입니다. MaxUint64.0과 MaxUint64+0.9는 잘라내기 이전 범위 검사를 구분하는 좋은 입력입니다. 정확한 정수 호환성·문법 오류·음수·유한 자릿수와 단위 한도를 추가해야 합니다. 이미 구현된 정수 개선을 다시 새 작업으로 세지 않습니다. |
| `humanize-bigbytes-precision` / `go53-humanize-bigbytes-precision` | 원본은 입력을 복사하지만 몫/나머지를 float64로 바꾸고 거친 표시 정밀도를 택합니다. 입력 불변과 16자리 요구는 의미 있습니다. 짝수 반올림·반올림 후 단위 승격·정밀도18/19·QB보다 큰 값은 명세되었거나 보류된 경계이므로 literal Want가 필요합니다. BigBytesN이 없는 것만으로 거부 근거가 되지 않습니다. |
| `humanize-precise-ftoa` / `go53-humanize-precise-ftoa` | 원본은 6자리 포맷 후 잘라냅니다. 이진수로 정확한 1.125/1.375 tie 쌍과 9자리 작은 값으로 반올림 및 정밀도를 구분합니다. float bits 또는 정확한 십진→이진 변환, 소수0자리 tie, 잘못된 정밀도와 NaN을 고정해야 합니다. 음수0 정책을 유지하고 기대값은 독립 계산합니다. |
| `afero-bounded-read` / `go53-afero-bounded-read` | 원본 ReadFile은 Stat를 용량 힌트로 사용하고 EOF까지 읽습니다. 잘못된 Stat 크기와 limit+1이 좋은 구분입니다. 0과 limit+1 산술 검사, 읽기 오류/초과 우선순위, Close 오류와 부분 데이터 정책을 정해야 합니다. Close 횟수는 유용하지만 Close 오류 명세를 대신하지 못합니다. |
| `afero-exclusive-write` / `go53-afero-exclusive-safe-write` | Exists→Create 사이에 획득 구간이 있습니다. **경쟁자를 Exists 이후에만 넣으면 Exists를 제거한 수정 경로에서는 경쟁자가 사라질 수 있습니다.** 두 구현 모두 목적지 획득 직전에 동일한 경쟁이 생기도록 고정하고 flags와 기존 바이트 보존을 기록해야 합니다. fs.ErrExist 호환성, copy/close cause, OpenFile/상위 디렉터리 실패도 필요합니다. |
| `cobra-error-writer` / `go53-cobra-error-writer` | PrintErr는 fmt.Fprint의 길이·오류를 버립니다. 이를 고쳐주지 않는 관측 가능한 baseline shim, 새 문자열 메서드 서명, 상속 error writer, nil short-write 정규화 및 cause 보존을 고정해야 합니다. “새 API가 없다”는 컴파일 실패는 의미 있는 baseline 실패가 아닙니다. |
| `cobra-local-finalizers` / `go53-cobra-command-finalizers` | 전역 finalizer는 명령별 격리를 제공하지 않습니다. baseline 대체 경로를 정해야 합니다. 기존 전역 callback을 등록하는 방법은 프로세스 내 시험 범위를 따로 허용하지 않으면 “전역 변경 금지”와 충돌합니다. 검증 오류·두 번의 실행·전역/로컬 순서를 추가해야 매 실행 한 번이라는 주장을 할 수 있습니다. callback panic/오류는 계속 제외합니다. |
| `cobra-context-preflight` / `go53-cobra-context-preflight` | help/version·Runnable 이후, preRun 이전 경계가 새 opt-in 정책을 뒷받침합니다. Cobra 등록 initializer와 Go package initializer는 다릅니다. 유효 context·callback 구성·help/오류 반환 경로·옵션 OFF의 literal 횟수를 고정해야 합니다. initializer Want 0은 package init 0의 증거가 아닙니다. |
| `retryhttp-deadline` / `go53-retryhttp-deadline-backoff` | 원본은 backoff 뒤 timer를 할당합니다. 결정적인 wait==remaining 경계가 유용합니다. 가짜 시계만으로 원본 time.NewTimer를 관측할 수 없으므로 충실한 baseline timing/timer 계측과 cause 우선순위를 정해야 합니다. 현재 MPL/cleanhttp 정책의 권리 보류를 유지합니다. 이 초안으로 실제 전송·sleep 시험을 실행할 근거는 생기지 않습니다. |
| `pflag-native-ipnet` / 새 native 초안 | IPv4 network/오류/상태 Want 5개는 읽은 Set 경로와 일치합니다. IPNet/IP/IPMask 타입이 달라 adapter의 충실성은 별도 막힘입니다. 기존 masking과 Source53의 opt-in host-bit 거부를 구분합니다. 노출된 계열 안의 새 정책 요청입니다. |
| `mapstructure-native-or` / 새 native 초안 | 결과·오류·callback Want 4개는 OrCompose 경로와 일치합니다. nil 결과+nil 오류는 성공, 전부 실패하면 끝 개행을 포함한 새 오류, hook 0개면 nonnil 빈 메시지 오류입니다. hook 서명·순서·reflection target·직접 관측 경로를 고정해야 합니다. |

## 첫 native 2개에서 충실하게 관측할 최소 내용

IPNet의 [ipnet.go:17](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/ipnet.go#L17)은 바깥 공백을 제거하고 net.ParseCIDR에 위임하며 성공 후에만 상태를 씁니다. 매 벡터를 소유한 IP `[192,0,2,0]`, mask `[255,255,255,0]`으로 초기화합니다. 성공 3개의 IP/mask 기대값은 각각 `[192,0,2,0]/[255,255,255,0]`, `[203,0,113,5]/[255,255,255,255]`, `[0,0,0,0]/[0,0,0,0]`입니다. /33과 prefix 없는 값은 nonnil 오류이며 초기 바이트를 유지합니다. 이는 원문에 근거한 literal 기대값이지 관측 결과가 아닙니다.

후속 관측기는 매번 새 명시적 FlagSet에 공개 IPNet/IP/IPMask 값을 등록하고 저장된 Flag.Value.Set을 호출할 수 있습니다. 원래 타입 구분과 소유한 바이트 snapshot, 반환 오류, panic을 따로 기록합니다. IP나 IPMask에 network 결과를 인위적으로 만들지 않습니다. Value.Set 직접 호출은 FlagSet.Set의 Changed 처리와 오류 wrapping을 거치지 않으므로 경로를 명시해야 합니다. 등록 중 String이 암묵 호출될 수 있습니다. 생성자·helper·문자열 변환 호출은 후보×벡터 15개 관측과 따로 집계해야 합니다. String/GetIPNet으로 읽으면 원본/helper 호출과 재파싱이 추가될 수 있어 literal 바이트 snapshot이 낫습니다. IPv6, 오류 메시지 본문, 넓은 parser 정확성은 제외합니다. 나중에 host가 0이 아닌 /0나 /31을 넣으면 masking 구분은 강해지지만 새 봉인 버전이며 요청 수는 늘지 않습니다.

대체 hook의 [decode_hooks.go:140](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go#L140)은 최초 from/to reflect 값을 재사용하고 첫 nil 오류에서 종료하며 실패 메시지를 이어 붙입니다. 정수4, 유효한 정수 목적지 reflect.Value, 정해진 지원 서명과 매 시험 새 불변 E1/E2 cause를 사용합니다. 동일한 literal hook을 OrCompose와 Compose에 넣고 동작을 고쳐주지 않습니다. callback 순서·입력 타입/값·결과 동적 타입/값·오류 nil 여부/문자열·panic을 기록합니다. nil-result 성공, 전부 실패한 null 결과, 빈 목록의 nonnil 오류를 구분합니다. 메시지 같음은 errors.Is 동일성의 근거가 아니므로 기존 연결 문자열 계약에 wrapped cause 보존을 추가 주장하지 않습니다. 생성자·cached hook 정규화·DecodeHookExec 호출은 후보×벡터 8개 관측과 별도입니다.

현재 검토는 빌드나 실행을 허용하지 않습니다. 15+8 관측은 조건부 설계이며 실제 원본 호출 횟수나 프로세스 전체 호출 수가 아닙니다. 계약의 최소 검토 파일 목록은 전체 빌드 소스 범위가 아닙니다. Source81/82 보관 자료를 재사용해 수집을 줄일 수 있지만 새 관측기·정확한 package 선택·의존성·build tag·toolchain/고지·정적 init·자원 계획은 별도 동결 연결이 필요합니다. 이후 봉인 구현은 기존 초안 이력과 분리해 검토할 수 있습니다.

## 중복·계열·권리

공개 saved membership76과 공개 qualified append3로 79개 요청 기록을 읽어 비교했습니다. 초안20개와 정확한 요청 문구가 같은 것은 0개입니다. 문구가 다르다는 사실이 의미 독립성을 증명하지는 않습니다. 구체적인 계약은 기존 정수·slice·event·UUID·ordinal·SemVer·glob 동작과 다르지만 공통 메커니즘은 남깁니다. 바이트 예산→quota-total, map snapshot→owned-snapshot, rollback→atomic-commit, cleanup→cancellation-lifecycle, hook routing→transform-order/error-identity, 따옴표→quoted-delimiters입니다. 검토한 차이는 계약 범위이며 새 원전이나 미노출 계열 인증이 아닙니다. Source53 18개 ID는 기존 논리 요청입니다. 재표현·literal 추가·잘못된 후보 추가·정답 후보 제거로 요청 수를 늘리지 않습니다.

Cobra+pflag는 하나의 연결 계열을 유지합니다. pflag와 mapstructure는 Source81에서, humanize는 old79 학습 추가분에서 이미 노출됐습니다. 이 검토 묶음을 protected final로 옮기지 않습니다. 역할을 정하기 전에 실제 차용 helper·복사 코드·alias/fork·공통 후처리를 graph edge로 기록해야 합니다. 같은 표준 라이브러리 import나 Go Authors 고지만으로 모든 계열을 합치지는 않습니다. Go import와 복사 원전의 lineage는 다른 증거입니다. 새 구체 연결이 기존 역할을 가로지르면 관련 학습 계획을 멈추고 근거를 보존하며 조용히 분할하거나 재배정하지 않습니다.

root license 7개는 바이트를 확인했지만 초안의 release/training clearance는 모두 false입니다. 로컬 근거는 pflag BSD-3-Clause, mapstructure/humanize MIT, afero/Cobra/INI Apache-2.0, retryablehttp MPL-2.0을 기록합니다. 이는 증거와 상태 검토이며 포괄적 법적 허가가 아닙니다. pflag의 복사 Go flag 소스, mapstructure의 복사 Go 고지, INI의 JSON 원전 신호, afero의 Go Authors/여러 저작권자, humanize number.go의 별도 원전은 선택 파일별 provenance가 필요합니다. 기존 작은 Ordinal slice는 number.go를 제외했으므로 향후 whole humanize 빌드를 허가하지 않습니다. 상위 저장소에 NOTICE가 없다는 사실은 원전·의존성 고지까지 면제하지 않습니다. 현재 MPL 권리 보류는 미해결 프로젝트 정책이며 MPL이 이 실험을 전면 금지한다는 판단이 아닙니다.

## 대안과 권고

| 대안 | 현재 적합성 | 이점 / 비용 |
|---|---|---|
| A. 20개 초안을 보존하고 후보 실행 없이 계약을 더 수집 | 보통 | 변화가 가장 작고 기획 범위를 넓힙니다. 적격은 0개이며 관측기·실행 결함은 시험하지 못합니다. |
| B. 초안을 보존하고 native 관측기 2개만 별도 동결·검토한 뒤 모든 조건이 통과하면 Root 계획에서 한 번 실행 | **가장 높음** | 기존 소스를 재사용해 입력9개로 adapter·집계 결함을 드러낼 수 있습니다. 실제 closure/init/권리/자원 검토가 필요하고 계속 노출된 개발 자료입니다. |
| C. 작은 coding 부분집합의 baseline·수정·잘못된 구현 fixture를 먼저 준비 | 다음 단계에 유용 | 코드 변경 난이도 계약을 직접 검증하지만 원문·원전·oracle 일이 더 큽니다. 없는 API를 인공적인 정답 동작으로 대체하지 않고 MPL 후보 보류를 유지합니다. |

다음 준비는 B가 가장 적합하고 이후 작은 C 부분집합이 좋습니다. 관측과 독립 검토 Want로 truth를 먼저 동결하고, 충실한 짧은 설명·순서·한도·계열 역할·가중치는 따로 정합니다. 소스 비교 성공은 LLM 작업 성공 결과가 아닙니다. hint의 효용은 별도 실험에서 같은 lexical 및 no-hint 기준선과 비교하며 전체 검증기/agent 작업·지연·OS 최대 RSS를 포함해야 합니다. 이 검토로 절약 효과를 주장하지 않습니다. 적격30/60 목표는 현재 고정79 loader와 분리하고, 도메인별로 의미상 별개인 protected final 최소2400 요청 목표를 유지합니다.
