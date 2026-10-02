# 다음 source-only 배치 준비

원래 20개 초안에서 순서를 바꾸지 않고 네 요청을 다음 준비 대상으로 골랐습니다. 목표 열 개에는 여섯 개가 부족합니다. 권리·정확한 원문 closure·미완성 계약을 보류한 결과이며, 다른 revision의 module cache나 빈 selector로 숫자를 채우지 않았습니다. 현재 라운드의 실제 qualified 요청은 Root가 확인한 두 개 그대로입니다. 이번 자료가 만든 새 요청·정답·역할·학습은 모두 0입니다.

| 원 ordinal | 요청 ID | 고정 원문 범위 | 기존 입력 / 후보 설명 |
|---|---|---|---|
| 1 | next60-humanize-strict-comma | comma.go + ftoa.go + big.go 원문 slice 제안 | 6 / 3 |
| 4 | next60-mapstructure-key-collision | mapstructure Source82 archive, Go 1.27 선택 6개 Go 파일 제안 | 4 / 3 |
| 5 | next60-pflag-map-snapshot | pflag Source82 archive 전체 42개 Go 파일 | 4 / 3 |
| 9 | next60-humanize-fractional-bytes | bytes.go 원문 slice 제안 | 5 / 3 |

19개 유한 입력은 네 요청의 fixture입니다. 후보 12개는 실행 코드가 아닌 기존 설명입니다. 각 요청의 baseline-design → reference-design → wrong-seed-design 순서와 그 안의 fixture 순서를 유지하면 향후 후보 dispatch 제안은 57개지만, 실제 호출 수나 독립 부모 수가 아닙니다. 입력·초안 Want·null Got는 원 JSON SHA와 pointer로 참조하고 복제·수정하지 않습니다. 전체 20개 선택/보류/제외 사유는 [metadata](BATCH-SELECTORS.v1.json)에 있습니다.

현재 로컬 캐시의 humanize v1.0.1, mapstructure v2.5.0, pflag v1.0.10, Cobra v1.10.2, afero v1.15.0은 모두 초안의 고정 revision과 다릅니다. 빌드 가능하다는 이유로 대체할 수 없습니다. Source82의 mapstructure/pflag 원문 및 notice 56핀, 확보된 humanize 원문·license 6핀은 현재 SHA가 기록과 일치합니다. 이것은 파일 대조이며 컴파일·연결·초기화 검증은 아닙니다.

humanize comma.go 전체를 보존하면 CommafWithDigits가 사용하는 ftoa.go의 stripTrailingDigits와 BigComma가 사용하는 big.go의 oom이 필요합니다. ParseComma 단독의 의미 closure와 파일 단위 compiler closure를 구분합니다. bytes.go는 표·보조함수까지 같은 파일에 있고 직접 읽은 import는 표준 라이브러리뿐입니다. 두 slice 제안은 number.go를 포함한 humanize 전체 패키지의 WTFPL 계보를 해결했다는 뜻이 아닙니다. ftoa.go를 helper로 참조한다고 이미 봉인된 ftoa 관찰을 다시 실행하지 않습니다.

mapstructure의 8개 보관 Go 파일 중 reflect_go1_19.go와 internal/errors/join_go1_19.go는 Go 1.27 선택에서 제외할 제안입니다. 실제 compiler selection은 Root가 별도로 고정해야 합니다. pflag의 세 핵심 파일은 전체 compiler closure가 아닙니다. 패키지 전역 FlagSet 초기화와 보조 오류·getter 호출도 adapter 계획에서 분리해야 합니다. 원 module 언어는 humanize 1.21, mapstructure 1.18, pflag 1.12이고 향후 binary toolchain은 Go 1.27.1로 고정할 제안입니다.

ordinal 18·19는 이미 qualified된 pflag/mapstructure 요청이라 제외했고, 11의 ftoa는 별도 진행 중이라 제외했습니다. 나머지 13개는 보류했습니다. INI/afero의 복사 원문 notice와 완전한 의존 파일, Cobra의 정확한 전체 원문, retryablehttp의 MPL/cleanhttp 경계가 아직 충분하지 않습니다. nil-config의 오류 identity/text와 bigbytes의 반올림 뒤 prefix·QB 초과 규칙도 덜 정해졌습니다. 추가로 mapstructure.go:478–485는 typed-nil Result에서 오류를 반환하므로, “config만 검사한 후보는 typed-nil에서 panic”이라는 wrong-seed 설명을 그대로 사용할 수 없습니다. 이것은 source-read finding이며 실행 재현이 아닙니다. 원 초안은 그대로 보존합니다.

다음 작은 구현은 네 selector의 concrete 후보 세 개씩과 typed adapter를 만드는 것입니다. 먼저 기존 입력과 오류/상태 채널, 정확한 code pin, 짧고 충실한 caption을 봉인합니다. comma는 grouping/range와 실패 반환값, collision은 같은 값·정확한 대소문자 이름도 포함한 충돌/원 상태, snapshot은 nil/빈 map과 양방향 변이, fractional bytes는 float JSON을 쓰지 않는 uint64 및 range-before-truncate를 읽어야 합니다. 새 범위를 추가해야 하면 실행 전에 별도 버전을 만들고, 관찰 후 Want를 고치거나 seed를 다시 고르지 않습니다.

mechanics는 배치 catalog 하나와 source-family별 child로 나눠 재사용합니다. typed request span·candidate·fixture 배열과 offset을 두고 원 순서로 dispatch합니다. conversion80의 고정 3×3×8 검증기를 그대로 확장할 수는 없으므로 bounded schedule/typed result schema가 최소 새 코드입니다. 기존 pin·bounded read·exclusive/fsync 예약·partial checkpoint·kill/Wait·오류 보존을 재사용하되 실제 code/schema에 맞는 새 봉인이 필요합니다. 후보 및 직접 Error 호출은 사전 예약/반환/panic을 따로 기록하고, 계측하지 않은 inner API/init는 null로 둡니다.

unknown·불완전 closure·harness 실패는 배치를 중단하고 미시작 행을 보존하며 재시도하지 않습니다. 정상적으로 관찰된 불일치는 그대로 남깁니다. whole humanize, mapstructure, Cobra↔pflag 연결과 실제 helper/alias 공유를 유지하며 요청별 무작위 역할 분할을 하지 않습니다. 비슷한 ownership/atomic 의미만으로 새로운 group union을 만들지는 않습니다. baseline과 wrong-seed가 같은 함수를 쓰는 경우도 별도 원천/독립 태스크로 세지 않습니다.

모델 입력은 별도로 검토한 request/caption text만 허용합니다. ID·Want/Got·source refs·역할·counter는 feature가 아닙니다. 유한 관찰, 코드 만족/문장 충실성/weight·whole-family 역할 qualification, Fit은 서로 다른 단계입니다. 원문 읽기·compile·관찰 runtime·archive·CI 비용도 구분합니다. 이번 준비에서 Go/build/test/helper process/native/model/Features/Project/Fit/HTTP/HF/protected-final/공유 쓰기는 모두 0입니다. 속도·RSS·정확도 향상을 측정하지 않았습니다.

30/60 개발 체크포인트와 도메인별 보호된 2,400개 의미 요청 목표를 유지합니다. 이 네 selector와 기존 가족으로 최종 다양성을 충족했다거나 모든 개발 fit에 2,400개가 필요하다고 주장하지 않습니다. 도달 시간을 추정하지 않습니다.

기존 공개 근거는 [20 draft](https://github.com/teamswyg/laya-tools/blob/acf514d3dd25db2e199e43da80ccf4477d8ff310/experiments/short-claim/next60-acquisition-drafts/DRAFT-CONTRACTS.v1.json), [Source82 notice](https://github.com/teamswyg/laya-tools/blob/acf514d3dd25db2e199e43da80ccf4477d8ff310/experiments/short-claim/source-closure-82/NOTICE-82.md)입니다. 원 SHA와 위 immutable repository commit이 참조 경계입니다. [고정 humanize MIT](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/LICENSE), [mapstructure MIT](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/LICENSE), [pflag BSD](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/LICENSE) 원문을 보존하는 기존 archive를 참조합니다. 이 private 준비는 아직 게시되지 않았으며 release/training 법적 승인이나 원문 전체 계보 증명이 아닙니다.
