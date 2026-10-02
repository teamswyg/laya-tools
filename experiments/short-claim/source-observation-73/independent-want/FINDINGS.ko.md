# 첫 세 원천: frozen Want/관찰 변환 독립 검토

**고정된 24개 finite Want와 관찰 변환에서 실행 전 의미 blocker를 찾지 않았다.** 앞서 observer spec·Want를 읽기 전에 고정한 독립 원문 메모를 수정하지 않고 대조했다. 이는 source-based finite 기대값 검토이며, native Got 관측·전체 동작 보장·caption/부모 라벨 qualification·학습 준비 승인이 아니다.

- **datasize 10개:** 빈 입력0, binary unit2048, suffix 공백, uint64 최댓값과 KB 곱셈 직전/직후 경계가 원문과 일치한다. `Value`·`InitialReceiver`는 uint64이므로 큰 값을 float64로 잃지 않는다. 초깃값7에서 syntax/bits→0, range→maxUint64인 오류 receiver 변화를 Want가 반영한다. main은 NumError 타입·Func·원래 Num을 확인하고 실제 `strconv.ErrRange/ErrSyntax/datasize.ErrBits`에 `errors.Is`를 적용한다. spec의 별도 bits error는 문구 생성용이며 실제 upstream sentinel identity를 대신하지 않는다. 실제 identity 판정은 observer의 위 predicate가 맡는다.
- **query 6개:** zero/repeated/nested/nil/typed-nil-pointer/int fixtures를 실제 컴파일된 struct·tag와 대조했다. empty string과 `flag,int`0 및 nested empty keys는 남고, nil Tags/Ptr,omitempty·Count,omitempty·Skip은 생략된다. repeated의 `tag=["a","b","a"]`는 그대로이며 nested bracket 이름도 일치한다. key만 정렬하고 각 value slice는 순서·중복·nil을 그대로 복사한다. nil/typed nil input은 nonnil empty map→`QueryValues=[]`, int 오류는 nil map→`QueryValues=null`+nil flag true다. callback이 없으므로 부분 map+callback 오류는 관측하지 않는다.
- **shlex 8개:** 기본 quote, 두 empty words, escaped space, comment newline, mid-word/quoted hash, 나중 quote-EOF, dangling escape-EOF, empty input의 literal 배열이 상태 머신과 일치한다. 두 malformed 사례 모두 `["ok"]`만 반환하며 unfinished `broken/tail`을 추가하지 않는다. empty input은 nonnil empty slice다. 두 error text는 달라지지 않으며 malformed_eof라는 넓은 category만으로 오류를 합쳐 숨기지 않는다.

`pure.LiteralSpecification`과 저장 Want는 main에서 byte equality로 결속된다. 24개는 **10+6+8 dispatch probes, 3 behavior goals**이며 새 문제24개나 독립 family24개가 아니다. 관찰 변환은 새 source API helper를 부르지 않는다. receiver 이후 숫자는 직접 uint64 conversion, query는 결과 map의 key sorting/copy, shlex는 반환 slice를 그대로 보존한다. 비교는 nil과 empty를 구분하는 DeepEqual이다. 오류/차이가 나도 Want를 재작성하지 않고 결과에 차이를 남기는 구조다.

범위: source 읽기상 selected probes의 no-panic 기대는 타당하지만 전체 no-panic 보장은 아니다. input bytes/fixture의 after snapshot은 이 관찰 schema에 없으므로 입력 불변성은 native 검증했다고 주장하지 않는다. lower-case `1mb`, 모든 unit, nil receiver, arbitrary callbacks, full POSIX/Unicode whitespace 등은 이번24개가 검증하지 않는다. 실패시 panic recovery 결과는 고정 redacted observation으로 남고 counter의 reserved는 호출 의도다. 초기화는 main 전 발생하므로 자식 자체 guard가 막을 수 없으며 root의 외부 선예약과 별도 init 범위가 필요하다. 현재 draft는 Frozen=false이고 실제 실행은 아직 하지 않았다.

별도 stdlib-only checker 1회/실패0으로 typed uint64를 포함한24행, handoff inline spec 동일성, 21개 closure byte/SHA, 바이너리 파일SHA/buildinfo, 이전 독립 메모의 unchanged pins를 확인했다. 바이너리 내용을 읽었지만 실행하지 않았다. 이 checker는 원문 의미 증명/compiled-instruction 증명이 아니다. 최초 metadata 요약 조회는 lowercase key를 사용해 null을 표시했으며 Capitalized schema 확인 뒤 보완했다; 작업/Want 오류가 아니다. 원본 API/init/테스트/바이너리·observer 코드 실행·모델/fit·공유 수정·Git/HF/공개 게시·보호 final 읽기 모두0이다.

독립성: 검토자는 관찰기 코드 저자가 아니며, 이전 원문 메모·넓은 source-build 목표와 과거 협업 기록에는 노출된 nonblind AI-assisted reader다. 실제 AI 협업 비용은 측정하지 않았다. 정확핀과 source/observer 위치는 `RECEIPT.json`, typed 검사 결과는 `MECHANICS.json`에 보존한다.
