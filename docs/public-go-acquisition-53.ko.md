# 공개 Go 작업 후보 120개: 실행 계약을 준비하기 위한 원천 목록

[English](public-go-acquisition-53.en.md) · [정규 JSON](../benchmarks/training/public-go-acquisition-53.json) · [2,400개 확보 계획](golden-set-acquisition.ko.md)

공개 Go 저장소 **8개에서 서로 다른 동작을 요청하는 원본 후보 120개**를 작성했다. 각 저장소의 revision, 실제 파일, 함수·타입 선언, 루트 LICENSE, 파일별 출처 신호와 NOTICE 존재 여부를 읽었다. **현재 실행 가능한 후보는 0개**다. 120은 요청 초안 수이며 모델 실행 수·난이도 정답 수·학습 정답 수·독립 통계 표본 수가 아니다.

작은 개발 시험만으로 성능을 판단하지 않기 위해 원천을 넓히는 단계다. 이 목록은 개발 과정에서 이미 읽었으므로 보호된 final에 넣지 않는다. 최종 목표인 영역별 별도 요청 2,400개와 학습·선택·보정 자료를 확보한 것으로도 세지 않는다. Go 코딩의 실제 완료 가능성을 검사할 원천이고, 저장소 선택이나 작업 분할의 정답을 대신하지 않는다.

## 무엇을 확보했는가

| 공개 저장소 | 후보 수 | 읽은 고정 revision | 실제 루트 라이선스 |
|---|---:|---|---|
| [spf13/pflag](https://github.com/spf13/pflag/commit/c966cfef47379dcb01e7929504d66d94b540945b) | 15 | `c966cfef4737` | [BSD-3-Clause](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/LICENSE) |
| [spf13/afero](https://github.com/spf13/afero/commit/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e) | 15 | `eb6a92826ea5` | [Apache-2.0](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/LICENSE.txt) |
| [spf13/cobra](https://github.com/spf13/cobra/commit/adbc8813901bba65827259daa8e22ff94ec1f30e) | 15 | `adbc8813901b` | [Apache-2.0](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/LICENSE.txt) |
| [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp/commit/fd004584a46724fae09e2f21d7c382e15c893f42) | 15 | `fd004584a467` | [MPL-2.0](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/LICENSE) |
| [go-viper/mapstructure](https://github.com/go-viper/mapstructure/commit/52aa5c6dc1d27226460807054ca2107b2d54fb2d) | 15 | `52aa5c6dc1d2` | [MIT](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/LICENSE) |
| [go-ini/ini](https://github.com/go-ini/ini/commit/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08) | 15 | `e2db55b0e088` | [Apache-2.0](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/LICENSE) |
| [dustin/go-humanize](https://github.com/dustin/go-humanize/commit/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e) | 15 | `a1b4e66b9a6d` | [MIT](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/LICENSE) + number.go WTFPL v2 |
| [google/uuid](https://github.com/google/uuid/commit/2d3c2a9cc518326daf99a383f07c4d3c44317e4d) | 15 | `2d3c2a9cc518` | [BSD-3-Clause](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/LICENSE) |

검토 근거로 등록한 원문 파일은 97개다. 이 수에는 LICENSE와 go.mod도 포함된다. 후보의 변경 앵커는 CLI 인자·명령 트리·가상 파일 시스템·HTTP 재시도·구조체 디코딩·INI 설정·숫자/시간 표시·UUID로 나누었다. 저장소당 15개라는 배치는 초기 수집 편의를 위한 것으로 실제 사용자 작업의 발생 비율을 나타내지 않는다.

모든 후보의 상태는 정확히 `source_candidate_only`다. `contract.status=pending`, `execution_eligible=false`, `training_label_eligible=false`, `final_eligible=false`, `model_executed=false`를 개별 기록했다. 라이선스 검토 pending은 별도 필드다. 계약을 작성 중이라는 표현을 실행 자격으로 해석하면 안 된다.

원본 요청문은 우리가 새로 작성한 **영문 동작 요청**이다. 아래 한글 표는 같은 요청의 짧은 행동 요약이고, 정확한 원문과 제안 완료 조건은 JSON에서 ID로 찾는다. upstream 이슈·댓글·테스트 본문을 복제하거나 같은 버그에 이름만 바꾸어 개수를 늘리지 않았다. 코드 본체·모델 본체·비공개 자료·실제 사용자 요청·인증 정보·원본 실행 추적은 이 목록에 넣지 않았다. 기존 구현의 결함을 입증한 목록은 아니다.

## 후보에서 실행 계약으로 넘어가는 조건

아래 조건을 통과하기 전에는 모델을 실행할 고정 작업으로 등록하지 않는다.

1. 요청의 입력·출력·실패·호환성 정책을 모호함 없이 고정한다. 이미 충족된 요청, 큰 구현 범위, 중복되는 해결 패턴은 수정하거나 제외한다.
2. 변경 가능한 파일뿐 아니라 필요한 지원 코드·플랫폼 파일·의존성·도구 버전·테스트 범위를 최소 closure로 만든다. 현재 파일 목록은 **읽은 변경 앵커**이며 컴파일을 확인한 closure가 아니다.
3. 그 실제 묶음의 SHA-256, revision, LICENSE/NOTICE와 복사·파생 원전 고지를 보존한다. JSON의 Git blob SHA-1은 원천 추적용이며 전체 closure 해시가 아니다.
4. 독립 완료 검사가 변경하지 않은 기준을 거절하고, 별도로 작성한 올바른 참고 구현을 수용하며, 알려진 오답을 거절하는지 검증한다. 기존 테스트 통과나 모델이 만든 테스트만으로 완료를 선언하지 않는다.
5. 필요한 검사와 공개 가능성 검토를 통과한 뒤 별도 실행 계획에 프로필·시도 수·시간·동시 실행·중단 조건을 봉인한다. 참고 구현/오답 검사와 실제 모델 실행 결과를 따로 기록한다.

JSON의 완료 조건은 **제안**이다. 아직 계약을 구현·봉인하지 않았고, 기준 거절·참고 구현 수용·오답 거절도 실행하지 않았다. 예상 fast/standard/strong은 부여하지 않았다. 실제 능력 라벨은 같은 작업에서 실행한 프로필의 독립 완료 결과·실패/종료 상태·전체 사용량에서 얻어야 한다.

## 라이선스에서 남아 있는 일

각 저장소의 고정 LICENSE 원문은 위 표에 연결했다. 검사한 8개 비잘린 트리에는 NOTICE 파일을 찾지 못했다. 이는 지원 파일·의존성·복사된 원전의 NOTICE도 없다는 뜻이 아니다. 프로젝트의 원본 요청·설명은 Apache-2.0이며 upstream 코드의 라이선스를 바꾸지 않는다. 전체 저장소, 파생 데이터, 모델 학습 권리까지 일괄 승인한 기록도 아니다.

- **go-retryablehttp:** 루트와 Go SPDX가 MPL-2.0이다. 변경 파일의 소스/실행물 배포 방식, 소스 제공·고지 범위와 의존성을 실제 closure에 맞춰 결정해야 한다. [MPL 원문](https://www.mozilla.org/en-US/MPL/2.0/)과 [공식 FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/)를 기록했다. 공개 소스라는 이유로 Apache-2.0 단독으로 반입하지 않는다.
- **go-humanize number.go:** 루트 MIT와 다른 출처를 기록한다. gorhill의 [고정 gist revision](https://gist.github.com/gorhill/5285193/cc36e402754549e658cf1d7255fd84f3c284f063)에 [WTFPL v2 원문](https://gist.githubusercontent.com/gorhill/5285193/raw/f45156c00aace6db928e0d89660606e6bb500231/%7ELICENSE)이 실제 포함된다. 원전이 확인되었다는 사실과 두 근거를 보존한 배포 manifest가 아직 없다는 사실을 구분한다.
- **표준 라이브러리 파생 신호:** pflag text.go는 Go 1.23.4 flag.go 복사를 명시하고, afero match.go는 filepath 파생을 명시한다. afero ioutil.go의 Go Authors 저작권과 INI struct.go의 encoding/json 복사 문구도 기록했다. 실제 사용할 부분의 원전 revision·고지 범위를 더 확인해야 하며, 이것만으로 위반을 단정하지 않는다.
- **여러 저작권자·의존성:** afero util.go/mem/file.go의 여러 기여자와 Cobra의 pflag·문서 도구 의존성을 보존·검토해야 한다. [Apache-2.0 원문](https://www.apache.org/licenses/LICENSE-2.0)의 변경·재배포·고지 조건은 실제 배포 범위에 적용한다.

특별한 파일 신호는 그 파일의 후보에 연결했다. 하지만 실제 closure가 넓어지면 모든 후보의 지원 코드·의존성·테스트 라이선스를 다시 검토해야 한다. 권리 검토와 계약 검토 모두 현재 pending이다.

## 연결 그룹과 기존 7개 후보

우선 같은 저장소의 후보를 같은 그룹에 묶는다. Cobra는 고정 [go.mod](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/go.mod)에서 pflag v1.0.9에 의존하므로 두 저장소를 보수적으로 연결했다. 이 목록의 pflag HEAD가 그 의존성 버전과 같다는 뜻은 아니다. 초기 저장소/의존성 그룹은 **7개**이며 독립성을 증명한 최종 그룹 수가 아니다.

같은 파일, 복사된 원전, fork, 공통 지원 코드, 해결 템플릿과 부모·자식·형제 관계를 더 검토하면 그룹이 합쳐질 수 있다. 예를 들어 원자성·자원 상한·소유권·시계 주입을 다른 저장소에서 요청해도 의미론적으로 관련될 수 있다. JSON의 mechanism 태그는 이 검토를 돕는 표시이고 고유 난이도 정답이 아니다. 저장소 일반화 평가에서는 저장소 전체를 나누고 연결 그룹을 학습·선택·보정·final에 흩뿌리지 않는다.

조사 시점의 역사적 7개는 `catalog-budget` 3개와 `repo-preview` 4개다. [현재 registry](../benchmarks/training/public-task-candidates.json)는 별도 `taskoutcome-parser` 후보 하나를 추가해 **8개·세 가족**이며 새 parser는 versioned 검증기 준비·모델 0시도다. 키워드 작업의 실제 네 반복은 [53 결과](../experiments/task-outcomes/RESULTS-53.ko.md)에 별도로 기록한다. 조사 JSON의 역사적 7개 관계와 신규 원천 120개 수는 이 추가로 덮어쓰지 않는다. 새 원천은 기존 laya-tools의 파일과 다르지만 budget·snapshot·filter·상태 처리라는 개념은 관련될 수 있어 중복/그룹 검토를 남겼다. 과거의 봉인된 요청·실행 기록은 수정하지 않고, 신규 120개를 과거 실행 수나 모델 라벨 수에 합산하지 않는다.

## 먼저 계약을 만들 후보 10개

아래는 작고 구체적인 독립 검사를 설계하기 위한 **개발 순서 제안**이다. 실제 모델 난이도 순서·학습 정답·추론 실행 승인이 아니다. 각 후보의 실제 gap과 지원 범위를 확인한 뒤 순서가 바뀌거나 제외될 수 있다.

| 순서 | ID | 동작 | 먼저 살펴볼 이유 |
|---:|---|---|---|
| 1 | `go53-humanize-ordinal64` | 전체 int64 범위의 순위 표시 | 단일 파일·정수 경계의 새 API라 독립 산술 기준을 설계하기 좋다. |
| 2 | `go53-pflag-map-snapshot` | 문자열 맵의 독립 조회 결과 | 복사 소유권을 외부 변경으로 검사할 수 있고 기존 조회와 분리하기 쉽다. |
| 3 | `go53-mapstructure-decoder-nil-config` | 디코더 설정의 nil 진단 | 입력 사전조건·오류를 작은 공개 호출로 검사할 수 있다. |
| 4 | `go53-cobra-active-help-lines` | ActiveHelp의 줄 경계 보호 | 프로토콜의 줄 경계를 고정 문자열로 검사할 수 있다. |
| 5 | `go53-afero-iofs-sub-validation` | io/fs 하위 경로 검증 | 표준 fs 경로 문법과 오류형을 독립 입력으로 대조할 수 있다. |
| 6 | `go53-ini-section-delete-index` | 중복 섹션 삭제의 인덱스 계약 | 음수·범위 밖·정상 인덱스와 상태 보존을 직접 검사할 수 있다. |
| 7 | `go53-uuid-canonical-parse` | 정규 UUID 텍스트 정책 | 허용/거절 문법과 기존 Parse 보존을 고정 byte 벡터로 검사할 수 있다. |
| 8 | `go53-uuid-sql-null-reset` | SQL UUID의 null 초기화 계약 | 미리 채운 receiver에 대한 값·오류의 상태 전이를 직접 검사할 수 있다. |
| 9 | `go53-humanize-strict-comma-parse` | 엄격한 정수 구분자 문법 | 정수 문법과 범위의 반례를 독립적으로 만들기 좋다. |
| 10 | `go53-pflag-canonical-network` | 네트워크 주소의 호스트 비트 검사 | 네트워크/호스트 비트의 명시적 정책을 고정 주소로 검사할 수 있다. |

보다 큰 상태 전이·오버레이·동시성·문서 의존성 후보는 뒤에서 closure를 넓혀야 한다. MPL 소스의 반입/배포 및 복사된 원전 범위를 해결하기 전에는 해당 후보를 실행 자격으로 올리지 않는다. 이 10개를 만드는 동안 이후 120개→240개 이상 개발 신호와 별도의 2,400개 final 원천을 계속 확보하되, 후보 수를 실측 라벨 수로 바꾸어 표시하지 않는다.

## 120개 전체 목록

모든 행은 실행 자격이 없는 개발 원천 후보다. 행동 가족 이름이 다르다는 것만으로 독립성을 증명하지 않는다. 함수·타입의 고정 line 앵커, Git blob SHA-1, 영문 원본 요청과 제안 완료 조건은 [JSON](../benchmarks/training/public-go-acquisition-53.json)에 있다.

### spf13/pflag

루트 BSD-3-Clause 및 Go Authors 헤더를 확인했다. text.go는 Go 1.23.4 flag.go 복사 신호가 있어 원전 고지와 closure 범위를 별도로 확정해야 한다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-pflag-normalization-collision` | 이름 정규화 충돌의 원자적 거절 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.flag_identity` |
| `go53-pflag-unknown-token-report` | 모르는 인자의 위치 보고 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.argument_provenance` |
| `go53-pflag-single-assignment` | 한 번만 지정할 스칼라 플래그 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.assignment_policy` |
| `go53-pflag-parse-stop-callback` | 콜백이 요청하는 정상 파싱 중단 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.parser_control` |
| `go53-pflag-help-display-width` | 유니코드 도움말 정렬 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.display_columns` |
| `go53-pflag-sensitive-default` | 민감 플래그 기본값 가리기 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.help_disclosure` |
| `go53-pflag-annotation-ownership` | 플래그 메타데이터 복사 소유 | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.metadata_ownership` |
| `go53-pflag-bounded-count` | 횟수 플래그의 상한 | [count.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/count.go) | `cli.count_accumulation` |
| `go53-pflag-bounded-string-slice` | CSV 목록의 총 항목 예산 | [string_slice.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/string_slice.go) | `cli.collection_budget` |
| `go53-pflag-map-snapshot` | 문자열 맵의 독립 조회 결과 | [string_to_string.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/string_to_string.go) | `cli.map_snapshot` |
| `go53-pflag-deferred-function` | 파싱 후 함수 플래그 실행 | [func.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/func.go), [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.side_effect_staging` |
| `go53-pflag-go-bridge-conflict` | 표준 flag 연결의 충돌 진단 | [golangflag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/golangflag.go) | `cli.stdlib_bridge` |
| `go53-pflag-text-error-value` | Text 값의 실패 진단 보존 | [text.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/text.go) | `cli.custom_value_error` |
| `go53-pflag-local-time` | 날짜 플래그의 지정 시간대 | [time.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/time.go) | `cli.time_location` |
| `go53-pflag-canonical-network` | 네트워크 주소의 호스트 비트 검사 | [ipnet.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/ipnet.go) | `cli.network_canonicality` |

### spf13/afero

루트 Apache-2.0과 여러 기여자의 헤더를 확인했다. match.go의 filepath 파생 신호 및 ioutil.go의 Go Authors 출처는 원전 추적·고지가 pending이다. util.go와 mem/file.go의 여러 저작권도 보존 대상이다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-afero-iofs-sub-validation` | io/fs 하위 경로 검증 | [iofs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/iofs.go) | `fs.subdirectory_contract` |
| `go53-afero-exclusive-safe-write` | 안전 쓰기의 원자적 새 파일 생성 | [util.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/util.go) | `fs.exclusive_creation` |
| `go53-afero-atomic-replacement` | 실패 시 원본을 지키는 파일 교체 | [ioutil.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/ioutil.go) | `fs.atomic_replacement` |
| `go53-afero-bounded-read` | 파일 전체 읽기의 크기 제한 | [ioutil.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/ioutil.go) | `fs.read_budget` |
| `go53-afero-union-pagination` | 오버레이 디렉터리의 일관된 커서 | [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.overlay_enumeration` |
| `go53-afero-copy-metadata` | 오버레이 복사의 파일 모드 | [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.copy_metadata` |
| `go53-afero-cache-invalidation` | 읽기 캐시의 명시적 무효화 | [cacheOnReadFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/cacheOnReadFs.go) | `fs.cache_invalidation` |
| `go53-afero-cache-clock` | 읽기 캐시의 주입 가능한 시계 | [cacheOnReadFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/cacheOnReadFs.go) | `fs.time_dependency` |
| `go53-afero-copy-on-write-tombstone` | 기본 계층 삭제의 숨김 표식 | [copyOnWriteFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/copyOnWriteFs.go), [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.overlay_deletion` |
| `go53-afero-readonly-truncate` | 읽기 전용 열기 플래그의 금지 범위 | [readonlyfs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/readonlyfs.go) | `fs.open_capability` |
| `go53-afero-regexp-rename` | 정규식 파일 필터의 이름 변경 | [regexpfs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/regexpfs.go) | `fs.filtered_rename` |
| `go53-afero-mem-seek-contract` | 메모리 파일의 seek 경계 | [mem/file.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/mem/file.go) | `fs.file_position` |
| `go53-afero-mem-subtree-rename` | 메모리 디렉터리 이동의 순환 거절 | [memmap.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/memmap.go) | `fs.tree_relocation` |
| `go53-afero-glob-limit` | glob 결과 수의 제한 | [match.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/match.go) | `fs.pattern_expansion` |
| `go53-afero-symlink-lexical-base` | BasePath의 링크 정책 선택 | [basepath.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/basepath.go) | `fs.symlink_policy` |

### spf13/cobra

검사한 소스의 Apache-2.0 헤더를 확인했다. cobra.go의 다른 도구에서 영감을 받았다는 문구는 기록하지만 복사 사실로 판정하지 않는다. pflag 및 문서 도구 의존성 고지는 따로 검토해야 한다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-cobra-checked-tree-insertion` | 명령 트리의 순환 삽입 거절 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.command_graph` |
| `go53-cobra-command-name-conflict` | 명령 이름과 별칭의 충돌 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.command_namespace` |
| `go53-cobra-ranked-suggestions` | 추천 명령의 순위와 개수 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go), [cobra.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/cobra.go) | `cli.diagnostic_ranking` |
| `go53-cobra-completion-state` | 자동 완성의 상태 변경 회수 | [flag_groups.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/flag_groups.go), [completions.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/completions.go) | `cli.completion_isolation` |
| `go53-cobra-context-preflight` | 취소된 명령의 훅 실행 방지 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.execution_cancellation` |
| `go53-cobra-error-writer` | 진단 쓰기의 오류 반환 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.output_failure` |
| `go53-cobra-validator-all-errors` | 위치 인자 검사의 전체 오류 | [args.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/args.go) | `cli.validation_composition` |
| `go53-cobra-required-annotation` | 필수 플래그 메타데이터의 잘못된 값 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.annotation_schema` |
| `go53-cobra-command-finalizers` | 명령별 종료 훅의 범위 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go), [cobra.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/cobra.go) | `cli.lifecycle_cleanup` |
| `go53-cobra-group-registration` | 명령 그룹 ID의 중복 등록 | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.group_registry` |
| `go53-cobra-cancel-completion` | 취소된 자동 완성의 콜백 범위 | [completions.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/completions.go) | `cli.completion_cancellation` |
| `go53-cobra-active-help-lines` | ActiveHelp의 줄 경계 보호 | [active_help.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/active_help.go) | `cli.completion_protocol` |
| `go53-cobra-markdown-filename` | 문서 파일 이름의 충돌 검사 | [doc/md_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/md_docs.go) | `cli.documentation_paths` |
| `go53-cobra-man-fixed-date` | man 문서의 재현 가능한 날짜 | [doc/man_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/man_docs.go) | `cli.documentation_clock` |
| `go53-cobra-yaml-command-metadata` | YAML 명령 문서의 별칭 정보 | [doc/yaml_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/yaml_docs.go) | `cli.documentation_schema` |

### hashicorp/go-retryablehttp

루트 MPL-2.0과 Go 파일의 SPDX를 확인했다. 소스·실행물 배포 방식, 변경 파일·고지·소스 제공 의무 및 의존성 범위를 확정하기 전에는 반입/배포 자격을 주지 않는다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-retryhttp-idempotent-policy` | 메서드와 멱등키 기반 재시도 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.retry_safety` |
| `go53-retryhttp-server-delay-cap` | 서버 Retry-After의 최대 대기 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.server_backpressure` |
| `go53-retryhttp-duration-overflow` | Retry-After 초의 오버플로 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.duration_representation` |
| `go53-retryhttp-header-selection` | 여러 Retry-After 값의 선택 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.header_resolution` |
| `go53-retryhttp-deadline-backoff` | 남은 기한과 대기 시간의 비교 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.deadline_budget` |
| `go53-retryhttp-exhaustion-error` | 소진된 재시도의 구조화 오류 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.failure_telemetry` |
| `go53-retryhttp-request-body-snapshot` | 가변 바이트 요청 본문의 소유 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.body_ownership` |
| `go53-retryhttp-bounded-reader-body` | 스트리밍 요청 본문의 읽기 예산 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.body_buffer_budget` |
| `go53-retryhttp-prepare-error-close` | 재시도 준비 실패 시 본문 정리 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.retry_resource_lifetime` |
| `go53-retryhttp-response-hook-attempts` | 응답 훅의 시도 메타데이터 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.hook_observability` |
| `go53-retryhttp-redacted-query` | 로그 URL의 선택적 쿼리 가리기 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.log_projection` |
| `go53-retryhttp-clock-backoff` | 독립적인 backoff 시계 | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.clock_injection` |
| `go53-retryhttp-deterministic-jitter` | 주입 난수에 따른 jitter | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.randomness_injection` |
| `go53-retryhttp-roundtripper-default` | 라운드트리퍼의 기본 클라이언트 계약 | [roundtripper.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/roundtripper.go) | `http.adapter_initialization` |
| `go53-retryhttp-cert-wrapped-errors` | 랩핑된 인증서 오류의 분류 | [cert_error_go120.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/cert_error_go120.go), [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.error_type_resolution` |

### go-viper/mapstructure

루트 MIT를 확인했다. 검사한 주요 Go 파일에 다른 라이선스 신호를 보지 못했지만 전체 저장소·파생 데이터·학습 권리의 일괄 승인으로 해석하지 않는다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-mapstructure-decoder-nil-config` | 디코더 설정의 nil 진단 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.configuration_precondition` |
| `go53-mapstructure-depth-budget` | 재귀 디코딩 깊이의 상한 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.recursion_budget` |
| `go53-mapstructure-cycle-input` | 순환 입력의 명시적 거절 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.graph_cycle` |
| `go53-mapstructure-collection-budget` | 컨테이너 디코딩 항목 예산 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.container_budget` |
| `go53-mapstructure-normalized-key-collision` | 정규화한 맵 키의 충돌 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.key_identity` |
| `go53-mapstructure-unused-order` | 사용하지 않은 키의 안정된 순서 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.metadata_order` |
| `go53-mapstructure-partial-result-policy` | 실패한 디코딩의 결과 정책 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.transactional_output` |
| `go53-mapstructure-hooks-errors-is` | 대안 훅의 원인 오류 보존 | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.alternative_hooks` |
| `go53-mapstructure-typed-nil-hook` | 합성 훅의 typed-nil 처리 | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.hook_value_lifetime` |
| `go53-mapstructure-delimited-escape` | 이스케이프 가능한 목록 훅 | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.delimited_text` |
| `go53-mapstructure-lossless-numeric` | 손실 없는 약한 숫자 변환 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.numeric_exactness` |
| `go53-mapstructure-hook-location` | 시간 훅의 명시적 지역 | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.temporal_context` |
| `go53-mapstructure-remain-conflict` | remain 필드의 중복 선언 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.struct_schema` |
| `go53-mapstructure-tag-precedence` | 태그 목록의 빈 값 처리 | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.tag_resolution` |
| `go53-mapstructure-unmarshal-ownership` | 텍스트 훅의 새 값 소유 | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.custom_unmarshal` |

### go-ini/ini

루트와 검사한 소스의 Apache-2.0 고지를 확인했다. struct.go에는 encoding/json/encode.go에서 수정해 복사했다는 신호가 있어 원전 revision·고지는 별도로 pending이다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-ini-parse-byte-budget` | INI 입력 바이트 상한 | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.input_budget` |
| `go53-ini-quoted-comment` | 따옴표 값의 주석 경계 | [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.lexical_comments` |
| `go53-ini-section-duplicate-policy` | 중복 섹션의 정책 선택 | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go), [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.section_identity` |
| `go53-ini-interpolation-cycle` | 값 치환의 순환 진단 | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.interpolation_graph` |
| `go53-ini-shadow-snapshot` | 중복 키 값의 빈 값 보존 | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.shadow_history` |
| `go53-ini-reader-reload-policy` | 한 번 쓰는 reader의 재로드 | [data_source.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/data_source.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.source_lifetime` |
| `go53-ini-append-transaction` | 설정 원천 추가의 원자성 | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.source_transaction` |
| `go53-ini-section-delete-index` | 중복 섹션 삭제의 인덱스 계약 | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.indexed_removal` |
| `go53-ini-key-rename` | 키 이름 변경과 조회 인덱스 | [section.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/section.go), [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.key_relocation` |
| `go53-ini-effective-parent-keys` | 부모 섹션의 유효 키 | [section.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/section.go) | `config.inheritance_resolution` |
| `go53-ini-save-atomic` | INI 저장의 파일 교체 계약 | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.persistence_atomicity` |
| `go53-ini-write-style-instance` | 설정 인스턴스별 출력 스타일 | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.format_state` |
| `go53-ini-struct-required-tag` | 구조체 필드의 필수 키 태그 | [struct.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/struct.go) | `config.struct_presence` |
| `go53-ini-struct-nil-target` | 구조체 매핑의 typed-nil 대상 | [struct.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/struct.go) | `config.reflection_precondition` |
| `go53-ini-strict-list-index-errors` | 목록 변환의 실패 위치 | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.typed_list_diagnostics` |

### dustin/go-humanize

API 분류는 NOASSERTION이지만 실제 루트는 MIT다. number.go의 gorhill 원전은 고정 gist revision에서 WTFPL v2를 확인했다. MIT 단독 표시 없이 원전 헤더·별도 라이선스 근거·배포 범위를 보존/확정해야 한다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-humanize-exact-fractional-bytes` | 소수 바이트의 정확한 산술 | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go) | `format.exact_fractional_bytes` |
| `go53-humanize-byte-rates` | 정확한 유리수 바이트 속도 | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go), [bigbytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bigbytes.go) | `format.rational_byte_rates` |
| `go53-humanize-explicit-byte-unit` | 명시한 단위로 바이트 표시 | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go) | `format.explicit_byte_unit` |
| `go53-humanize-bigbytes-precision` | 임의 정밀도 바이트 표시 | [bigbytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bigbytes.go), [big.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/big.go) | `format.arbitrary_precision_bytes` |
| `go53-humanize-strict-comma-parse` | 엄격한 정수 구분자 문법 | [comma.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/comma.go) | `format.grouped_integer_grammar` |
| `go53-humanize-append-comma` | 버퍼에 정수 표시 추가 | [comma.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/comma.go), [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go), [big.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/big.go) | `format.append_number_buffer` |
| `go53-humanize-int64-template` | 정밀 손실 없는 정수 서식 | [number.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/number.go) | `format.exact_integer_template` |
| `go53-humanize-precise-ftoa` | 고정밀 실수 반올림 표시 | [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go) | `format.float_precision_rounding` |
| `go53-humanize-scientific-si` | 지수 표기의 SI 숫자 파싱 | [si.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/si.go) | `format.scientific_si_grammar` |
| `go53-humanize-rounded-si-promotion` | 반올림 이후 SI 단위 승격 | [si.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/si.go), [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go) | `format.rounded_si_promotion` |
| `go53-humanize-relative-table-validation` | 상대 시간 표의 사전 검증 | [times.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/times.go) | `time.checked_magnitude_table` |
| `go53-humanize-duration-parts` | 기간의 제한된 단위 분해 | [times.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/times.go) | `time.bounded_duration_decomposition` |
| `go53-humanize-ordinal64` | 전체 int64 범위의 순위 표시 | [ordinals.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ordinals.go) | `format.signed_ordinal` |
| `go53-humanize-case-plural` | 대소문자를 보존하는 복수형 | [english/words.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/english/words.go) | `text.case_preserving_plural` |
| `go53-humanize-quoted-word-series` | 이스케이프된 단어 나열 | [english/words.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/english/words.go) | `text.escaped_series` |

### google/uuid

루트 BSD-3-Clause 및 검사한 Go 파일의 Google 저작권·LICENSE 연결 헤더를 확인했다. 실제 closure의 플랫폼 파일·테스트·고지 범위를 아직 고정하지 않았다.

| ID | 제안하는 동작 | 읽은 변경 앵커 | 검토용 행동 가족 |
|---|---|---|---|
| `go53-uuid-canonical-parse` | 정규 UUID 텍스트 정책 | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.canonical_text` |
| `go53-uuid-atomic-batch-parse` | 원자적인 UUID 목록 파싱 | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.transactional_batch` |
| `go53-uuid-append-text` | UUID 텍스트 버퍼 추가 | [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.append_serialization` |
| `go53-uuid-sorted-unique` | 원본을 보존하는 UUID 정렬·중복 제거 | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.sorted_deduplication` |
| `go53-uuid-bounded-random-batch` | 제한된 UUID 난수 배치 | [version4.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version4.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.bounded_entropy_batch` |
| `go53-uuid-write-text` | 짧은 쓰기를 검사하는 UUID 출력 | [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.short_write_io` |
| `go53-uuid-sql-null-reset` | SQL UUID의 null 초기화 계약 | [sql.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/sql.go) | `identifier.sql_scan_state` |
| `go53-uuid-owned-binary-value` | 소유권이 분리된 SQL 바이너리 값 | [sql.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/sql.go), [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go) | `identifier.sql_binary_ownership` |
| `go53-uuid-nullable-binary` | nullable UUID의 바이너리 대칭성 | [null.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/null.go) | `identifier.nullable_encoding` |
| `go53-uuid-checked-hash` | 해시 UUID의 명시적 오류 계약 | [hash.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/hash.go) | `identifier.checked_hash_io` |
| `go53-uuid-v1-v6-conversion` | UUID v1·v6 비트 배치 변환 | [version1.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version1.go), [version6.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version6.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.reversible_time_layout` |
| `go53-uuid-checked-timestamp` | UUID 시각의 검증된 추출 | [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go) | `identifier.checked_timestamp` |
| `go53-uuid-explicit-v7-time` | 전역 상태 없는 명시적 UUID v7 | [version7.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go), [version4.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version4.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go) | `identifier.explicit_time_generation` |
| `go53-uuid-instance-v7-generator` | 독립적인 단조 증가 UUID v7 생성기 | [version7.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.isolated_monotonic_generator` |
| `go53-uuid-checked-dce-domain` | 부작용 이전 DCE domain 검사 | [dce.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/dce.go), [version1.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version1.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [node.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/node.go) | `identifier.dce_domain_policy` |

UUID의 독립 bit-layout 기준으로는 [RFC 9562 원문](https://www.rfc-editor.org/rfc/rfc9562.html)을 사용할 수 있다. 인스턴스의 단조 증가·실패 후 상태 정책은 이 프로젝트가 선택할 계약이며 모든 UUID 구현의 필수 성질로 주장하지 않는다.
