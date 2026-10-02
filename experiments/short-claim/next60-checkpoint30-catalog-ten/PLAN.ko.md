# 공개 목록에서 다음 열 계약 고르기

기존 Source53 목록120개에서 첫20 초안에 들어 있지 않은 **기존 ID10개**를 골랐습니다. 새 요청 이름을 만들거나 예시를 새 요청으로 세지 않았습니다. 구체 입력52개와 원본·수정 후보·잘못된 후보의 호출 제안156개는 **미실행 기획**입니다. 새 적격 자료·라벨·역할·가중치는0입니다.

현재 Root는 반올림(ftoa)까지 적격화해 next60 자료가3개 요청·라벨8개라고 보고했습니다. 앞선 기획서의 QA 보류 표기는 당시 기록으로 보존합니다. 기존79와 비활성 모델은 그대로입니다.

| 순서 | 기존 ID · 원문 | 입력으로 확인할 차이 | 잘못된 후보 |
|---|---|---|---|
| 1 | `go53-pflag-annotation-ownership` · flag.go | 등록 배열과 getter 배열을 바꿔도 저장 annotation은 그대로 | 등록 때만 복사하고 getter는 빌린 배열 반환 |
| 2 | `go53-pflag-bounded-count` · count.go | 범위0..2의 상한 증가·음수·큰 정수 거부 후 값 보존 | 검사 전에 정수 값을 대입 |
| 3 | `go53-pflag-single-assignment` · flag.go | 같은 flag를 long/shorthand로 두 번 지정해도 첫 값 보존 | argv 철자를 별개 flag로 취급 |
| 4 | `go53-pflag-sensitive-default` · flag.go | help에는 고정 marker, 실제 기본값·파싱은 그대로 | 일반 default만 가리고 optional 값 노출 |
| 5 | `go53-pflag-deferred-function` · func.go | 뒤에서 파싱이 실패하면 콜백0, 성공하면 argv 순서대로 한 번 | 오류 때도 defer로 queue 실행 |
| 6 | `go53-pflag-text-error-value` · text.go/errors.go | flag 이름·오류 원인은 보존하고 합성 입력값은 오류문에서 제거 | 새 오류 문자열을 만들어 원인을 잃음 |
| 7 | `go53-mapstructure-hooks-errors-is` · decode_hooks.go | 실패한 두 원인을 Is/As로 찾고 nil 성공에는 즉시 멈춤 | nil 성공인데 다음 hook도 실행 |
| 8 | `go53-mapstructure-remain-conflict` · mapstructure.go | 직접/embedded 출력 필드의 remain 충돌을 데이터 대입 전에 거부 | 직접 필드만 검사 |
| 9 | `go53-mapstructure-tag-precedence` · mapstructure.go | 없는 tag와 명시적으로 빈 tag를 새 옵션에서 구별 | Get으로 둘을 같은 값으로 합침 |
| 10 | `go53-pflag-unknown-token-report` · flag.go | 실제 unknown 토큰의 원래 위치를 기록하고 값·`--` 뒤 토큰은 제외 | 대시가 붙은 모든 토큰을 unknown으로 봄 |

[JSON 제안](CATALOG10.v1.json)에 각 요청의3–6개 입력·Want 제안, 세 후보 구현 방향, 출처 SHA와 남은 조건을 넣었습니다. Want는 실행 결과로 맞춘 것이 아니며 Root의 의미·후보·구체 계약 검토 전 정답/라벨이 아닙니다. 새 API가 없다는 컴파일 오류만으로 원본 후보를 실패시키지 않습니다. 실제 원본 API를 호출하고 상태·오류·콜백 차이를 보존하는 adapter가 필요합니다.

원문 확인으로 처음 후보 두 개는 제외했습니다. typed-nil hook은 이미 원본 Compose가 보존하므로 신규 개선 근거가 없습니다. metadata 순서는 map 순회라 한 번 실행에서 우연히 정렬될 수 있고 엄격 오류 이름은 이미 정렬합니다. baseline 실패를 미리 보장하지 않습니다. Text 오류의 Is/As 원인 보존도 이미 구현돼 있으므로 신규 차이는 입력값 제거입니다.

중복 비교는 공개 메타데이터 범위로 한정했습니다. 실제79에 추가된 UUID Parse/Scan·Humanize Ordinal은 제외했고 semver/doublestar의 비학습 역할 그룹도 고르지 않았습니다. 공개 정보에는 pflag/mapstructure의 직접 출처 그룹이79에 없지만, 원래76의17개 자체 작성 요소 전체와 의미가 독립적이라는 증명은 아닙니다. 현재 native2 Or 요청은 오류문 연결을 요구합니다. 새 Join 요청은 오류 객체 원인을 보존하는 별도 coding 계약이지만 같은 hook 출처·그룹으로 묶고 기존 라벨을 재사용하지 않습니다.

열 요청은 **이미 노출된 두 그룹**을 공유합니다. pflag/Cobra는 같은 연결 그룹77, mapstructure는 그룹78입니다. 열 독립 출처나 heldout으로 주장하지 않습니다. 원문과 compiler/dependency/generation proof를 한 번씩 재사용하고, 기존 durable writer로 호출 전·반환 후 기록을 남기는 실제 batch를 준비하면 됩니다. global finalizer reset·OS 파일·서버·실시간 시계·모델이 필요하지 않습니다. 원본 package 초기화와 예상 외 오류도 실제 카운터에 남겨야 합니다.

가장 작은 실행 단위는 처음 네 직접 상태/출력 API입니다. 이미 확보한 pflag 본체를 사용하고, callback/Join/decoder 세 계약을 이어서 묶습니다. argv 좌표를 parser 안에서 보존해야 하는 마지막 요청은 그다음입니다. 실제 원본 parser를 대체하는 가짜 parser나 숨겨진 후보 상태 초기화는 만들지 않습니다. pflag BSD, mapstructure MIT, text.go의 Go1.23.4 복사 고지를 각각 유지하고 새로운 실제 compiler 선택에 연결해야 합니다. 기존 notices 핀이 새로운 실행 권한이나 포괄 라이선스 승인인 것은 아닙니다.

현재3 + 준비4 + 앞선6 + 이번10이 모두 적격화되면23개입니다. 기존 레지스트리20 + 이번10이 모두 적격화돼야30개입니다. 남은 일곱 기존 보류가 비싸거나 MPL 문제로 유지되면 다른 기존 catalog 계약 일곱 개를 같은 방식으로 고릅니다. fixtures를 늘려30이라고 하지 않습니다. 30/60 체크포인트와 도메인별 보호 평가2,400 목표를 유지합니다. 작은 입력의 코드 차이만으로 학습 성능·낮은 RAM·GPU·비용 절감을 주장하지 않습니다.

이번 작업의 Go·빌드·테스트·helper·원본·oracle·모델·보호 평가·HF·HTTP·공유 변경은0회입니다. 이전 ftoa/controller/converter/reader 작성자의 비맹검 source 기획이며 독립 정답·라벨·라이선스·컴파일 승인이 아닙니다.
