# 원천 먼저 읽기80: UUID Parse·Scan과 작은 Ordinal

**3개 행동 목표, 2개 원천 가족, 24개 입력 초안**을 준비했다. UUID Parse와 Scan은 같은 package·helper 가족이다. Ordinal은 다른 원천의 함수 하나다. 아직 Want, adapter, binary, 실제 부모·정답·역할·가중치는 없다. 기계 자료는 [SOURCE-FIRST-80.v1.json](SOURCE-FIRST-80.v1.json), 입력만 있는 초안은 [INPUT-DRAFTS.v1.json](INPUT-DRAFTS.v1.json)이다.

| 목표 | 최대8개 입력에서 비교할 것 | 나중에 보존할 관찰 |
|---|---|---|
| UUID Parse | raw32/standard36의 정상·늦은 hex 실패, URN45의 대소문자 prefix·잘못된 prefix, braced38·다른 바깥 문자 | 반환 UUID16바이트 전체, 반환/오류/panic 상태를 별도 기록 |
| UUID.Scan | nil interface, 빈 string, typed nil bytes, nonnil empty bytes, 정상 text, raw16 bytes, 늦게 실패하는 text·bytes | 매번 새 nonnil receiver의 전후16바이트, 입력 타입·nilness, 오류/panic |
| Ordinal | 0·1·2·3·11·12·13·112 | 입력 int와 반환 문자열 전체, panic 여부 |

입력은 제안이며 정답 출력은 만들지 않았다. 제안된 직접 callback은 Parse8 + Scan8 + Ordinal8 =24개다. MustParse·Validate·Unmarshal 등의 **원문 비교**를 그 함수에 대한 직접 관찰로 바꾸지 않는다. Scan의 내부 재귀·xtob·오류의 Error() 같은 보조 호출도 직접 callback 수와 다르다. 최종 관측 변환과 호출 예산은 다음 author가 따로 고정한다. 24개의 입력 슬롯은 24개 독립 문제나 실제 호출 횟수가 아니다.

## UUID: 같은 실패처럼 보여도 반환값과 상태는 다르다

원전은 [google/uuid의 고정 revision](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d)이다. Parse의 반환형은16바이트 배열 값이다. raw32 경로는 `xtob`가 계산한 바이트를 대입한 **뒤에** 성공 flag를 검사한다. standard36 경로는 flag를 검사한 **뒤에** 대입한다. 따라서 늦은 hex 실패의 앞부분이 같아도 실패 위치의 바이트를 같거나 모두0이라고 가정하면 안 된다. 원문·table·타입·오류 선언과 exact byte span이 기계 자료에 결속돼 있다.

URN의 잘못된 prefix는 해독 전에 별도 오류로 빠진다. 길이38의 Parse는 첫 바이트를 건너뛰고 내부36 위치를 읽으며, 바깥 문자를 중괄호인지 확인하지 않는다. Validate는 이 검사를 한다. 그래서 `[...UUID...]`를 모든 API에서 같은 “invalid UUID”로 미리 분류하지 않는다. 기존79의 일반적인 invalid-input 문장을 보편적 RFC 검증 계약으로 승격하지 않고, 이후 요청은 구체적인 길이·hex·prefix 범위로 고정해야 한다. 기존 초안은 수정하지 않았다.

Scan은 nil·빈 입력에서 대입 전에 return한다. comment의 “null UUID”를 receiver 초기화로 읽으면 실제 body와 어긋난다. string은 Parse를 지역 변수로 받은 뒤 오류가 없을 때만 receiver에 대입하므로, Parse가 반환한 부분 UUID와 Scan의 receiver 상태를 합치면 안 된다.16바이트 slice는 직접 copy하고, 다른 길이의 slice는 string으로 바꿔 Scan에 위임한다. 오류 형식은 `%v`이며 `%w`로 원래 오류를 감싼다는 약속은 없다. `UUID.Scan`과 nil 초기화를 따로 다루는 `NullUUID.Scan`도 구분한다.

실제 database·driver callback, nil receiver, unsupported dynamic type, 모든 길이·separator·Unicode 조합, aliasing은 이번8입력 관찰 밖이다. 오류 타입·문자열과 panic 직렬화 역시 아직 Want로 고정하지 않았다. UUID.String 같은 원래 API를 관측용으로 숨겨 호출하지 말고 반환 배열을 직접 보존하는 방향이다.

## 컴파일 후보와 시작 초기화

UUID는 이미 보존된 전체 runtime 원문16개를 읽었다. Darwin non-JS 후보는 `node_js.go`를 제외한15개 +원래 go.mod +full BSD-3-Clause다. 각 raw SHA-256을 역사적 definition의 literal pin과 대조했고, 목록53에 Git blob이 있는 파일도 대조했다. 이것은 원문 파일 후보이며 아직 실제 compiler/go list로 선택된 closure나 표준 라이브러리 전체의 hermetic proof가 아니다. 원래 tests는 관찰에 필요하지 않아 실행하거나 새 기대값으로 재사용하지 않는다.

**import만 해도 main보다 먼저 원래 함수가 실행될 수 있다.** hash.go의4개 namespace 변수 initializer는 MustParse→Parse를 호출한다. 이4개는 정적으로 읽은 call site 수다. 원문 init을 계측하지 않으므로 “actual init callback4·반환4를 관측했다”고 쓰지 않는다. 미래 외부 dispatcher가 child 시작 전에 예약하고, `static startup sites reserved4 / actual init callbacks not instrumented`로 직접 probe와 분리한다. main 안의 token 확인은 시작 초기화를 막을 수 없다.

나머지 globals에는 hex table·오류 sentinel·rand.Reader 참조·pool/lock·clock/node 상태·jsonNull 등이 있다. 선택 Parse/Scan 경로에서 entropy·시간·네트워크 인터페이스·random-pool API를 부르는 코드는 없지만, 전체 package와 표준 라이브러리의 startup 비용을0이라고 주장하지 않는다. 실제 OS CPU·RSS·wall은 초기화·setup·직접 probe·관측·저장을 포함한 **whole child lifetime**으로 측정한다.

## Ordinal: 명시적 원문 slice

[humanize의 고정 revision](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e)의 원래 `Ordinal(int)` body는 `strconv.Itoa`만 호출한다. selected file에는 package 변수나 init이 없다. 그래서 원래 ordinals.go·go.mod·full MIT를 유지한 **명시적 source slice**로 함수 관찰을 제안한다. 이것을 전체 upstream library나 그 init·메모리 사용의 관찰이라고 부르지 않는다. slice 고지에는 원전 revision, 원문 SHA, 함수 하나만 남긴 package 구성, 추가 Ordinal64·다른 파일·tests를 제외했다는 사실을 적는다.

원래 함수에는 비음수 guard가 없으며 int를 받는다. 이번0..112의 작은 비음수 범위는 관찰 선택일 뿐, 음수를 지원하지 않는 새 계약이나 기존 동작을 고쳤다는 주장이 아니다. 원문 변경0이다.11·12·13 예외는 modulo100 조건이어서112를 넣어 단순 “12만 예외” 설명과 구별할 수 있는지 본다. 정답 suffix를 Want에 적거나 모델을 실행하지 않았다.

전체 humanize를 선택하면 아직 이 slice에 없는 다른 본문과 `number.go`의 gorhill/WTFPL 원전 고지를 확인해야 한다. 기계 자료는 이를 known-missing pointer로만 남긴다. 선택 Ordinal slice에 필요 없는 원문을 새로 취득하거나 복제하지 않았다. 현재 새 취득0이다.

## 고지와 아직 할 수 없는 주장

UUID **BSD-3-Clause 전체 LICENSE는1,480B**, SHA `0a8d61ed3cbfd5312326e8126c31ce9c627a283adc99131b56896d29ada04b2d`다. Humanize **MIT 전체 LICENSE는1,136B**, SHA `a973b4498c13eb74baa2a8e5c351426a6826f2fcdd909916dbe53ee2e755fd71`다. 초기 root 지시의 “UUID full MIT” 표기는 동결 전에 바로잡았다. 기존 원문·고지·기대값은 수정하지 않았다.

기존53 inventory·taskverify와55의 두 개발용 코드 수정 과제에 이미 노출된 원천이다.55 기록의 started CLI4는 과거 pilot이며 이번 실행이 아니다. 본 reader는 source-caption74 작성과76 결과에 노출됐고, 원문을 읽은 뒤 역사적 task definition/acceptance provenance도 읽었다. 새 Want author와 분리하지만 전체 과정을 blind/human-only·fresh final이라고 주장하지 않는다.

실제 parent·label·role·weight0, qualification·training·production·protected final 모두false다. 동일 UUID 두 목표, helper·alias·번역·반복 fixture는 같은 family로 유지한다. 제안2families가 독립 저자·독립 role 그룹이라는 증거는 아니다. CPU1·1worker·Go soft heap256MiB·60초·출력1MiB를 다음 실행의 상한 후보로 제안하며, soft heap을 OS RSS hard cap으로 부르지 않는다. 아직 import/init/API/test/native·모델/학습·공유 수정은0이고24입력은2,400개 미노출 도메인 final 검증을 대신하지 않는다.
