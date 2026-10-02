# Catalog10 v2: 실행 전 literal 수정 제안

이 자료는 기존 catalog의 10개 요청을 다음 native 준비에 사용할 수 있도록 구체화한 source-only 제안입니다. [JSON](CATALOG10.v2.json)에 입력 객체와 Want를 적었지만 원 함수, 후보, oracle, Go 도구, 모델은 실행하지 않았습니다. 실행 계획의 `frozen`은 false이며 독립 literal·caption 검토와 코드 작성은 남아 있습니다. 요청·후보의 truth, role, weight, pair eligibility와 qualified 값은 null입니다.

v1의 4개 파일 41,816 B는 바꾸지 않았습니다. JSON의 `preserved_v1`은 각 파일의 bytes와 SHA를 연결하고, 모든 입력은 v1의 `/ten/i/fixtures/j`를 참조합니다. 새 v2는 v1의 넓은 자연어 설명을 그대로 보존하면서 입력과 정책을 새 literal 객체로 명시합니다. 이미 존재하는 실행 fixture를 그대로 옮겼다는 주장은 아닙니다. v1의 실행 봉인된 canonical literal Want가 없는 상태도 바꾸지 않습니다. 관찰 결과를 보고 Want를 고치는 방식도 허용하지 않습니다.

10개 ID와 52개 입력의 순서는 같습니다. 요청별 입력 수는 5/6/5/5/5/5/5/5/5/6입니다. 각 요청의 후보 위치는 원본 0, 작성할 reference 1, 작성할 negative 2로 고정합니다. 따라서 30개 후보 위치와 156회 dispatch는 계획 수이며 실제 수는 0입니다. 52개 입력을 52개 부모로 세지 않습니다. 기존 pflag 7개 요청과 mapstructure 3개 요청은 두 원천 흐름의 기존 그룹 77·78을 참조합니다. 새 역할 배정이나 가족 분할은 없습니다.

| 요청 | 새 literal 범위와 남은 한계 |
| --- | --- |
| annotation ownership | getter를 `([]string, keyPresent bool, error)`로 제안합니다. 등록·교체·조회에서 복사하고 nil과 비nil empty slice를 구분합니다. 없는 key는 nil/false/nil, 없는 flag의 getter는 새 작성 sentinel을 반환하도록 정합니다. 원본 setter의 NotExistError는 유지합니다. 원본에는 이 getter ABI가 없습니다. |
| bounded count | CountVarP 등록이 0을 저장한 뒤 Start를 대입합니다. Go 1.27.1의 64-bit int와 `ParseInt(input,0,0)`를 고정합니다. 범위는 0..2이며 잘못된 입력은 commit 전에 거절합니다. 처음 다섯 입력은 Value.Set, 마지막 입력은 ordinary `-vv` Parse입니다. |
| single assignment | 한 fresh FlagSet의 한 Parse 안에서 annotation이 있는 같은 flag identity의 두 번째 setter를 거절합니다. long/shorthand 별칭은 같은 identity이며 unannotated 반복은 유지합니다. direct Set과 여러 Parse의 정책은 제외합니다. |
| sensitive help | 이름, usage, default, NoOptDefVal, hidden 상태와 정확한 help 문자열을 명시했습니다. default와 optional 표시만 redaction하며 저장 값은 바꾸지 않습니다. 문자열은 원문을 읽어 작성한 제안이고 실제 renderer 출력은 아닙니다. |
| deferred callbacks | successful Parse 뒤 argv 순서대로 callback을 실행하고 첫 callback 오류에서 멈춥니다. parse 오류·미정의 help에서는 flush하지 않습니다. literal sentinel과 trace를 명시했고 immediate mode는 유지합니다. |
| text error value | 작성할 Text fixture는 두 실패 입력에서 수신자를 바꾸지 않습니다. 이 fixture의 성질을 일반 TextUnmarshaler rollback 보장으로 확대하지 않습니다. caption은 공급 값 redaction, flag identity와 Is/As cause 보존만 주장합니다. cause 자체가 비밀 값을 포함하는 경우의 일반 sanitization은 제외합니다. |
| hook error identity | 두 실패 cause의 Join/Is, typed cause의 As, nil 출력·nil 오류의 첫 성공, 빈 hook 목록의 새 sentinel을 명시합니다. hook마다 같은 원 입력 3을 줍니다. message 연결 문자열은 이 요청의 Want가 아닙니다. |
| remain conflict | direct 및 value-embedded squash의 중복 remain을 unmatched map 쓰기 전에 거절합니다. 정상 matched field rollback, nil embedding 할당, 전체 receiver transaction이나 무할당 보장은 제외합니다. |
| tag precedence | 실제 struct-tag 문자열과 입력 map을 고정하고 IgnoreUntaggedFields=false로 둡니다. first-present는 빈 tag도 선택하며 dash는 skip합니다. legacy first-nonempty와 단일 tag 선택은 유지합니다. |
| unknown token report | 정확한 원 parser branch와 argv 좌표를 기록할 instrumentation·후보 코드가 없습니다. 6개 literal 정책은 hold입니다. 값으로 소비한 dash token, delimiter 뒤 positional, shorthand 내부 위치를 단순 별도 token scanner로 추측하면 안 됩니다. |

나머지 46개 입력도 “실행 가능”이나 “만족” 판정이 아닙니다. ABI와 literal fixture가 명시된 제안일 뿐이며 batch용 30개 후보 위치의 코드·adapter 선택은 아직 동결되지 않았습니다. 원본 upstream 본문은 이미 보존되어 있고 읽었습니다. authored sentinel은 원본 exported sentinel로 가장하지 않습니다. 오류 kind는 이후 typed adapter가 구현할 분류 이름이고, identity는 지정한 객체 또는 Is/As 검사로 확인해야 합니다. Text에만 명시한 진단 문자열은 별도 target이며 그 외 일반 오류 문자열은 미관측 null입니다. nil slice/map은 null, 비nil empty slice는 [], empty map은 {}로 남깁니다.

훅 요청의 계보는 정확한 공개 `next60-acquisition-drafts/DRAFT-CONTRACTS.v1.json`의 `/drafts/19/related_catalog_contracts/0`입니다. native2의 `next60-mapstructure-native-or`는 기존 오류 문구를 LF로 연결하는 계약이며 이번 Join/Is 정책과 다릅니다. related link를 primary catalog ID, 새 독립 원천 또는 기존 label 재사용 근거로 세지 않습니다. 그룹 78은 공유 원문·helpers의 whole-family 경계를 유지합니다. 작성자는 이전 source·batch4·checkpoint·outside·ftoa oracle 작업과 기존 자료에 노출된 비맹검 상태입니다. v1 catalog 저자와는 다르지만 v2 literal/caption 저자이므로 자신의 검토를 독립 의미 검증으로 세지 않습니다.

JSON의 source pins는 읽은 본문 anchors입니다. 전체 compiler closure 증명이 아닙니다. pflag revision `c966cfef47379dcb01e7929504d66d94b540945b`와 mapstructure revision `52aa5c6dc1d27226460807054ca2107b2d54fb2d`의 full notices를 기존 공개 archive에서 참조합니다. pflag BSD3, mapstructure MIT와 copied Go text 소스의 BSD 고지는 별개입니다. text.go는 Go 1.23.4 flag.go에서 복사했다고 원문에 적혀 있으며 retained source와 full [Go LICENSE](https://github.com/golang/go/blob/go1.23.4/LICENSE)의 exact bytes/SHA를 JSON에 연결했습니다. 기존 archive 경로는 `experiments/short-claim/source-closure-82/attribution/Go-go1.23.4-LICENSE`입니다. 원 프로젝트 라이선스가 copied code를 재라이선스한다는 주장은 없습니다.

다음 작업은 pending 독립 literal/caption 검토 뒤 원본·reference·negative의 닫힌 코드와 explicit call schedule을 작성하는 것입니다. module/notices/compiler 및 별도 Root source/rights/resource/CI evidence를 봉인한 뒤 기존 단일 subprocess mechanics를 연결합니다. wrapper dispatch 156은 내부 API·getter·hook·Error/Is/As 호출의 전체 비용이 아닙니다. compile/controller 준비 비용, native 관찰 비용, 이후 qualification과 모델 Fit 비용을 구분해야 합니다. 관찰 전에 source 또는 fixture가 부족하면 해당 범위를 hold로 남기고, 실행 후 panic/unknown/mismatch/불완료는 보존하며 재시도나 유리한 Want 수정 없이 중단합니다.

요청과 30개 caption의 단순 ASCII 단어 수는 각각 최대 24·22입니다. 이는 metadata 계산이며 실제 lexical Normalize/Validate가 아닙니다. 다음 모델 입력은 기존 512 B/32 normalized words/candidate 최대 8 계약을 실제로 검증해야 하며 ID·그룹·Want·Got·label을 feature text에 넣지 않습니다. 새로운 threshold, seed search, 2400-per-fit 또는 all-prototypes-all-roles 조건을 추가하지 않았습니다. 이 문서 작성으로 새 부모·학습 label·역할·훈련 자격이 생기지 않습니다.

장부는 JSON 안에 포함했습니다. 파일 경로 조회 실패 4회와 협업 메시지 전달 실패 1회는 metadata 준비 기록이며 과학적 실행 실패나 retry가 아닙니다. 원본 실행·oracle·renderer·Go/build/test/gofmt·모델·보호 자료 읽기·공유/외부 변경은 모두 0입니다. 과거 v1 설명과 미해결 기록은 그대로 두고 이 새 proposal만 WRITE_STOP으로 인계합니다.
