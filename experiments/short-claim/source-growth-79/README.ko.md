# 다음 데이터 확대: 6개 새 원천 가족, 12개 논리 요청

기존 3개 문제를 다시 학습하는 대신, **서로 다른 동작 조건을 구별해야 하는 12개 요청**을 준비한다. 6개 라이브러리를 우선 대상으로, 2개를 예비 대상으로 골랐다. 현재는 원천 선택·취득 제안이다. 실제 부모, 후보 caption, 정답, 역할, 가중치는 생성하거나 배정하지 않았다. 기계 자료는 [SOURCE-TARGETS.v1.json](SOURCE-TARGETS.v1.json)에 있다.

| 우선 대상 | 준비할 두 요청 | 지금 확인한 범위 |
|---|---|---|
| google/uuid | 오류를 반환하며 이미 해독한 앞부분을 보존하는 파서 / nil·빈 입력에서 기존 receiver를 유지하는 scanner | 보존 원문·도움 함수·라이선스를 읽고 원래 Git blob과 대조 |
| dustin/go-humanize | 11·12·13 예외가 있는 서수 / 마지막 접속사 앞의 쉼표를 넣는 단어 목록 | 원래 Ordinal·라이선스만 읽음. 단어 연결 함수와 다른 후보는 원문 필요 |
| go-viper/mapstructure | time 목적 타입에만 작동하는 변환 hook / 빈 문자열과 중복 순서를 구별하는 slice hook | revision·파일·선언 위치만 확보. 타입 guard·도움 함수 본문 미확인 |
| spf13/pflag | 첫 Set과 반복 Set의 상태 차이 / CSV 해석 없이 덮어쓰는 Replace | 원문 취득 후 malformed CSV와 상태 변경 순서 확인 필요 |
| go-ini/ini | 잘못된 정수 항목에서 오류를 반환하는 변환 / 잘못된 항목 자리에 0을 유지하는 변환 | 원하는 정책을 제안한 단계. 실제 Ints가 0을 넣는다고 아직 주장하지 않음 |
| spf13/afero | 실패한 Seek의 위치 보존 / 쓰기·truncate를 거부하는 읽기 전용 wrapper | 원문 취득 후 음수 offset·flags·backing 상태 변경 순서 확인 필요 |

**UUID와 작은 비음수 Ordinal부터 원천 기반 유한 계약을 만드는 것이 가장 빠르다.** 원문이 이미 있고 입력도 짧다. 그다음 mapstructure hook과 pflag CSV 상태를 읽으면, 단어가 비슷해도 조건·오류·순서가 다른 후보를 만들 수 있는지 확인할 수 있다. INI와 Afero는 조건을 만족하지 않는 구현일 수도 있다. 그 경우 결과를 본 뒤 요구를 고치지 않고, 해당 초안을 unknown으로 보존하고 별도 요청 버전을 검토한다.

예비 Cobra는 인자 검사 함수만 검토한다. 이미 pflag와 의존 관계가 있어 별도 독립 그룹으로 세면 안 된다. retryablehttp는 고정된 context·response에 대한 정책 함수만 후보로 두고 실제 HTTP 요청을 제외한다. MPL 고지, 의존성, 시간·난수·global 경로 때문에 첫 실행 대상으로 추천하지 않는다.

## 먼저 필요한 원문

이 자료는 [기존 공개 원천 목록53](https://github.com/teamswyg/laya-tools/blob/main/benchmarks/training/public-go-acquisition-53.json)의 고정 revision·Git blob SHA-1·라이선스 기록을 재사용했다. 라이브러리 업데이트를 가정하거나 새 다운로드를 하지 않았다. 각 대상의 `minimal_source_read_acquisition_list`에 공식 upstream의 **고정 revision 링크**, 기존 blob SHA-1과 원문이 있는 경우 raw SHA-256을 기록했다. 미확인 SHA는 null이다.

UUID는 `uuid.go`, `util.go`, `marshal.go`, `sql.go`; humanize는 `ordinals.go`, `comma.go`, `english/words.go`; mapstructure는 `decode_hooks.go`, `mapstructure.go`, `errors.go`; pflag는 `string_slice.go`, `flag.go`, `errors.go`; INI는 `key.go`, `section.go`, `file.go`, `ini.go`, `error.go`; Afero는 `mem/file.go`, `mem/dir.go`, `memmap.go`, `readonlyfs.go`, `afero.go`를 우선 읽는다. 각각 원래 `go.mod`와 전체 루트 라이선스도 보존한다. **이 목록은 의미 검토의 시작점이며 실제 Go 패키지 컴파일에 필요한 전체 파일 목록은 아니다.** 생성자와 추가 helper가 필요하면 해당 revision에서 함께 묶는다.

라이선스를 루트 이름 하나로 단순화하지 않는다. humanize의 `number.go`에는 별도 WTFPL 원전이 있고, pflag `text.go`·INI `struct.go`·Afero의 일부 파일에는 Go 유래 복사 신호가 있다. 선택 함수가 그 파일을 호출하지 않더라도 전체 패키지나 원문을 배포할 때 적용 범위를 확인해야 한다. 기존 원문 헤더·전체 라이선스·필요한 원전 고지는 유지하며, 현재 자료가 학습 권리나 배포를 일괄 승인하지 않는다.

## 무엇을 새로운 문제로 세는가

이전 76개 development 데이터는 **17개의 연결된 자체 작성 그룹 + semver·doublestar 2개 원천 가족**, 총 19개 whole group이다. 실제 역할은 train 12·validation 4·calibration 3그룹이므로 19개 전부를 학습했다고 표현하지 않는다. 자체 작성 prototype 이름은 18개지만 일부가 연결되어 그룹은 17개다. 이번 6개 대상은 그 저장된 prototype·upstream 가족 이름과 겹치지 않는다. 이는 직접 이름 대조이며, 외부 의존성과 모든 복사 원전까지 독립성을 증명한 것은 아니다.

datasize·querystring·shlex, 그리고 godotenv·wordwrap·logfmt·dataurl은 이미 원문·caption·관찰 준비에 노출된 가족이므로 이번 “새 원천” 목록에서 제외한다. 앞선 71/72 학습에 포함되지 않았다는 사실과 새 독립 자료라는 주장은 다르다. UUID·humanize도 과거 개발용 taskverify와 목록53에서 노출됐으므로 보호된 final에 넣을 수 있는 미노출 자료로 부르지 않는다.

같은 repository의 두 요청, wrapper, 공통 helper, 번역과 반복 fixture는 연결해 둔다. 한국어·영어 번역은 같은 논리 요청이다. 12개 요청의 목표가 모두 성립한다는 보장도 없다. 8개 저장소나 12개 문장을 독립 저자·독립 그룹 수로 바꾸지 않는다. mapstructure와 이전 wordwrap의 Mitchell 계열, pflag·Afero·Cobra의 spf13 계열과 복사 원전 흐름도 남아 있다. 원문의 저자 흐름과 AI가 작성한 요청·caption의 흐름을 따로 기록한다.

## 작게 관찰하고, 설명에 조건이 있는지 확인한다

각 요청은 20~25개 ASCII 공백 단어, 141~165바이트다. 실제 입력 Normalize는 실행하지 않았다. 우선 후보 3개, 상한 8개를 제안하며 모델 입력은 **request + caption**뿐이다. 원문 코드, 함수 ID, 정답, 관찰값, 역할과 라이선스 metadata를 몰래 feature로 추가하지 않는다. 원천에 조건이 있어도 caption에 없으면, 자료 수만 늘려 해결된다고 기대하지 않는다.

다음 유한 관찰은 짧은 입력·작은 목록·새 receiver·메모리 파일만 사용하고, 반환값·오류·panic·nil/empty·부분 결과·변경 전후 상태를 나눠 남긴다. 예를 들어 UUID의 32바이트 Parse는 실패한 hex 쌍의 계산값을 먼저 대입하지만 36바이트 경로는 성공 여부를 먼저 확인한다. 따라서 “부분 결과”를 모두 같은 배열이나 실패 위치 0으로 가정하면 안 된다. Scan의 빈 입력 comment보다 실제 대입 전 return이 현재 동작을 설명한다. 기존53의 “빈 입력을 초기화하도록 바꾸기” 요청을 현재 정답으로 쓰지 않는다.

상한으로 요청당 8개 fixture, 입력 512바이트·목록 8항목·메모리 파일 256바이트·단일 worker CPU 1·soft heap 256MiB·60초·출력 1MiB를 제안했다. 실제 API 호출 수는 adapter·생성자를 읽은 뒤 따로 고정해야 한다. 최대 96개 fixture 슬롯은 96개 독립 문제나 96번 API 호출이 아니다. 현재 실행 권한·실행 계획으로 해석하지 않는다.

원문에 근거한 후보 만족과 모델에 보이는 caption의 조건 충실성을 구분한 뒤, 기존 절차에 따라 독립 유한 기대값과 관찰 범위를 고정한다. 만족하는 후보가 하나도 없다는 근거가 충분한 경우에만 no-answer를 검토하고, 누락·모호한 조건은 unknown으로 남긴다. 원래 76개 truth·mask·role·failed model은 변경하지 않는다. 이후 학습과 default 활성화는 별도 결과에 달려 있다. **이번 12개 초안은 2,400개 미노출 도메인 final 검증을 대신하지 않는다.**

현재 원천 취득·원래 API/init/test·관찰·Normalize/Features/Project·모델/학습·HF 게시·공유 수정은 모두 0회다. 이것은 AI가 참여한 문서 검토 비용까지 0이라는 뜻은 아니다.
