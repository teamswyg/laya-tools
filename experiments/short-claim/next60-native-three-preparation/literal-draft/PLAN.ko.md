# 세 계약의 실행 전 literal·ABI 초안

상태는 **DRAFT, Root 미채택, 실행 준비 false**입니다. 새 자격·truth·label·role·weight·numeric group을 부여하지 않습니다. 기존 22개 가설, 5개 retired 기록 및 세 HOLD는 그대로 둡니다.

이번 선택은 기존 ordinal **2 → 7 → 1**, 세 요청과 두 whole-source family입니다. 입력은 Complement 4, JSON 5, Binary 기존 4와 별도 padding 추가 1로 **14개**입니다. 후보는 요청당 3개, 계획상 42 dispatch이며 실제 관측은 0입니다. Intended contract 구현 위치는 0/1/2로 분산했습니다. 위치·이름은 label이나 reference 우대가 아닙니다.

Complement는 원 `New(0)` 뒤 listed `Set`만 적용합니다. 입력 길이는 0/1/65/66, 명시 폭은 0/1/64/65입니다. 뒤 두 사례에서 폭과 저장 길이가 다릅니다. Want는 exact 결과 길이, words/set bits, 입력의 복사 snapshot입니다. borrowed Words는 즉시 고정 배열로 복사하며 원본 slice를 보관하지 않습니다.

JSON은 전체 문자열의 완전한 JSON 여부를 확인한 뒤 root Parse를 사용합니다. root Number이며 정수 lexeme에 dot/exponent를 허용하지 않는 좁은 제안입니다. 기존 다섯 JSON 객체를 유지합니다. exponent·malformed JSON은 현재 다섯 fixture에 없으므로 실제 coverage로 주장하지 않습니다. 원 Int는 `9007199254740993`을 Raw parseInt로 정확히 반환하는 경로가 있습니다. 변경 목적은 checked rejection이며 원 Int의 error 채널 부재는 **unknown**, 가짜 nil이나 false가 아닙니다.

Binary는 default BigEndian, 선언 최대 128bit, exact 최대 24B로 제한합니다. 네 기존 textual profile를 명시 BE64 hex와 동일 nonempty receiver로 옮겼습니다. Len1/word3의 unused-bit 사례는 **새 다섯 번째 profile**로 별도 기록하며 이전 사례를 지우지 않습니다. huge header는 원 dispatch 전에 제외하고 원 실패처럼 세지 않습니다. reject에서 receiver를 보존하는 계약과 반환 error의 실제 availability를 분리합니다.

[계획](LITERAL-PLAN.proposed.v1.json), [Want 초안](WANTS.proposed.v1.json), [ABI·availability](ABI-AND-AVAILABILITY.proposed.v1.json), [출처·license 핀](SOURCE-PINS.v1.json), [후속 범위](SUCCESSOR-AND-SCOPE.v1.json)를 함께 읽어야 합니다. 원본문과 full BSD/MIT notices는 기존 private archive에 있고 이 폴더에는 복사하지 않습니다. host 경로는 PRIVATE-BASIS 파일에만 있습니다.

own Go는 inert `.go.txt`입니다. native binder, worker/observer, durable checkpoint, compiler/linker, runtime budget 및 Root 독립 Want freeze는 모두 별도 단계입니다. 별도 oracle 준비도 같은 작성자의 source이며 독립 oracle 검증이 아닙니다. Go/gofmt/build/test/원 API/init/model/Fit/network 실행은 모두 0입니다. 첫 실행은 이 초안으로 허가되지 않습니다.
