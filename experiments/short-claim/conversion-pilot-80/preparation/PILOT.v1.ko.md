80의 완료된 원문 관측을 **요청 초안3개와 후보9개**로 연결했다. 실제 parent·학습 truth·역할·마스크·가중치는 아직0이며 관련 필드는null이다. 기존80의24fixture, Want와 Got, 원문20개 핀, 기존76자료를 바꾸지 않았다. 모델·기준선 채점도0이다.

| 요청 초안 | 전체 후보3개 | 제안 acceptable index |
|---|---|---:|
| 선택된 UUID 문자들의 오류·부분 반환 bytes 보존 | 오류면zero / 원문결과그대로 / 오류면panic | [1] |
| 선택된 DB 입력에서 receiver 보존·성공commit·오류반환 | 원문결과그대로 / 먼저clear / 오류discard | [0] |
| 작은 비음수 rank의 decimal digits·teen suffix 예외 | last-digit만 / constant-th / 원래Ordinal | [2] |

정답 위치를0/1/2에 분산한 순서 제안은 채점 전에 고정하며 점수를 보고 바꾸지 않는다. 원천 계열은UUID와Ordinal 둘이고 Parse/Scan은서로다른목표지만같은sourcefamily다. 이 세 초안은 독립그룹3개나2400평가, 학습 효용을 뜻하지 않는다. 아직 실제그룹소속도null이다.

각 후보는 [구체적인 Go 함수](candidates.go.txt)와 연결했다. 이 파일은 inert 텍스트이고 module·adapter·binary를 만들거나 실행하지 않았다. Parse에서 오류 후값을zero로 바꾸거나panic하는 변환, Scan에서 먼저zero로 시작하거나error를 버리는 변환, Ordinal에서teen예외를 생략하거나항상th를 붙이는 변환이다. 잘못된 후보도 그 코드의 동작을 충실히 설명하는 caption을 제안했다. 반대 문구만 만들어진 원문 함수의 관측값처럼 가장하지 않는다. Go 컴파일 가능성과 새 후보 실제실행 결과는 아직검증되지 않았다.

[제안 JSON](CONVERSION-PROPOSAL.v1.json)은 네축을 나눠 둔다. code-SAT에는 좁은finite요청을 만족할지와 반례fixture를 제안하지만actual은null이다. caption fidelity는새후처리까지설명하는작성자제안이고peer판단은null이다. eligible-mask는원래67정책에따른적격가능성만기록하고실제loss-mask/label/weight는null이다. channel completeness는기존80의필요채널존재와새후보미검증을분리한다. 예를들어Scan은receiver가같아도오류를숨기면요청을만족하지않는다. unknown/unsupported를새negative truth로변환하지않는다.

요청과caption은ASCII로작성했고원래NormalizeText의ASCII분할recipe를별도전사해바이트/단어를확인했다. 요청25/24/25words,caption16~22words이며512B/32words내다. 원래Normalize/Validate/Features를호출한결과는아니므로publicinput실제검증필드는null이다. runtime에는request/caption만보내고source/코드/fixture/Want/Got/정답제안/role/mask는옆자료로둔다.

다음은Root의원문peer검토와[별도관찰preplan](CANDIDATE-VERIFICATION-PREPLAN.v1.json)이다. 이preplan은미봉인·미실행이다. 같은24입력의모든후보를확인하면candidate invocation72개이고native는Parse24+Scan24+Ordinal8=56회,나머지Ordinal16개는새owned함수다. 예측상반환error의명시적Error호출10회,보수적상한48회를native primary와구분한다. error-valuedpanic의message를다시Error로읽지않고null로보존하는계획이며,원래MustParse panic을관측했다고쓰지않는다. UUIDstartup4지점과Scan내부재귀/Parse/fmt호출수는정적근거이고실측은null이다.

이후실제candidate근거·caption적격·typedscopedacceptable이연결돼야fixed_order/lexical_ordered/BM25/narrow_rule을동일한전체후보에서한번비교할계획이다. unsupportednarrow_rule의정상fallback을보존한다. strongestcontrol/tie/order와oracle분모·기존5%headroom/utility,Top1/Top3조건은그대로며지금checks/headroom/score는null이다. 이pilot은fit을승인하지않는다. 기존9/3/3역할하한·보호된2400final범위는바꾸지않고,cal/final을train으로옮기거나옛val을unseen이라고부르지않는다. baseline이이미oracle이면그slice를학습할이유가없다.

작성자는80Want작성자와기존62/71/76작업노출이있는비맹검제안자다. 독립source/정답승인이나실제모델품질실험이아니다. 원20핀41,784B와저장된24Want/Got의연결만읽기대조했다. 원본import/compile/init/API·원80observer·후보실행·Score/Features/Fit·paid/HF·공유편집모두0이다. fullBSD/MIT[고지](NOTICE.md)를보존하고새derivative를독립source로세지않는다. 자세한바이트핀·실제준비시도는[receipt](RECEIPT.v1.json)와[ledger](LEDGER.v1.json)에둔다. [English](PILOT.v1.en.md)
