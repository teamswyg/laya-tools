# 실행 전 원문·유한 기대값 독립 검토68

v3 계획은 원문 코드의 기본 플래그와 세 입력·두 이름에 대한 정적 흐름에 부합한다. 카운터 표현도 수정됐다. 그러나 **짧은 후보 캡션이 모든 유한 반환값을 설명한다고 승인하지 않는다**. 특히 NewVersion의 완전한 문장은 “coerce를 시도한다”는 설명이며, 세 입력의 성공과 정확한 정규화 문자열을 모두 열거하거나 보장하지 않는다. 실제 원본 API 관측도 아직 없다.

검토자는 `semantic_review60_prep`이다. root의68 계획이나 upstream 문구의 작성자가 아닌 별도 reader지만 blind 검토는 아니다. 이전60·61 검토,63 원문 QA,64 projection 작성,65 metadata QA와 source/metadata 노출이 있었다. 이 보고서와 아래 영어 request 초안은 검토자가 AI-assisted 방식으로 작성했으므로 그 초안에 대한 별도 작성자 독립 승인을 주장하지 않는다.

최종 계획 바인딩은 `PLAN-68.v3.ko.md`, SHA `6a3b9eb06fcd1223392dab767b7e739304927f14d239879d5cc0dbe4ac057831`이다. 최초 맡은 v2 SHA `d1de515f6b5fc5946a9fee3b74a6f1c7767483b5747326e8d65e3065dfeed4c1`와 비교하면 worker 문단 한 줄만 달라지고 요청·문구·Want 설명은 그대로다. v1의 DetailedNewVersionErrors=false 오류는 root가 실행 전 source reading으로 발견한 역사로 남긴다. 실제 선언은 CoerceNewVersion=true와 DetailedNewVersionErrors=true이며, 후자는 CoerceNewVersion=false 경로에서만 상세 오류 처리를 제어한다.

## 완전한 원문 선택과 한도

바이트 구간은0부터 시작하고 끝은 제외한다. 원문 prefix·backtick·LF·문법 기호를 수정하지 않았다. JSON에는 전체 raw text, 원본 파일 SHA, quote SHA와 줄 구간이 있다.

| 선택안 | 원본 구간 | 바이트 / normalized words | 기존63 위치 |
|---|---|---:|---|
| Strict doc.go 완전 문장 | doc.go [386,497),13–15행 |111 /17| `/quotes/0` |
| NewVersion doc.go 완전 문장 | doc.go [498,590),15–16행 |92 /16| 새68 선택;63의 선택 목록에는 없음 |
| g-star 전체 문법 항목 | match.go [197,257),15행 |60 /7| `/quotes/17` |
| g-globstar 전체 문법 항목 | match.go [258,307),16행 |49 /5| `/quotes/18` |
| g-mid-component 완전 문장 | match.go [1298,1457),40–42행 |159 /26| `/quotes/24` |

모두512B/32words 안에 든다. NewVersion 바로 뒤의 완전한 example 문장은161B/36words이므로 선택하지 않는다. 단어를 빼거나 잘라 cap을 통과시키지 않았고 전체 원문을 미선택 증거로 보존했다. MIT 고지2개는 각각1078B의 원본 전체 바이트로 별도 보존했다. 기존 source manifest의15 artifact SHA와 전체 notice도 검증했다. 이는 실제 원문 보존 근거이며 개별 저자나 human-only 작성 인증은 아니다.

## 코드 읽기로 뒷받침되는 Want와 문구의 한계

StrictNewVersion의 SplitN 세 부분 검사와 digit-only 검사는 고정 입력 `1.2.3`, `v1.2.3`, `1.2`에서 accepted `[true,false,false]`라는 사전 Want를 뒷받침한다. 이 값들은 실행 결과가 아니다. “valid 버전만 파싱” 문장은 제한을 충실히 설명하지만 모든 유효 버전의 수락이나 정확한 error/no-panic 반환 계약을 보장하는 문장으로 확대하지 않는다.

기본 CoerceNewVersion=true이면 NewVersion은 anchored loose regex와 coerceNewVersion으로 들어간다. regex는 선택적인 v prefix와 patch 생략을 허용하고 생략한 patch는0이며 String은 숫자 세 부분을 출력한다. 따라서 accepted `[true,true,true]`와 String `1.2.3,1.2.3,1.2.0`이라는 사전 Want에 정적 근거가 있다. **NewVersion은 새 관측 범위**다.57·59·63의 Strict 관측이 이를 실행했다는 근거로 사용되지 않는다. 정확한 String 값은 관측 metadata이고 request의 캡션 약속으로 간주하지 않는다. 그래도 selected 문장 자체는 “attempts”만 말하므로 prefix/missing-patch의 모든 성공 비트를 fullfinitecoverage 승인으로 바꾸지 않는다.

Match는 slash·validation=true·caseInsensitive=false를 사용한다. plain star의 backtrack은 slash를 건너지 못하고, segment-start double star 뒤 slash는0개 또는 nested directory를 허용한다. ** 뒤 문자가 `.`이면 ordinary star로 처리한다. 고정 패턴 `src/*.go`, `src/**/*.go`, `src/**.go`와 이름 `src/main.go`, `src/lib/main.go`에 대한 matched Want는 각각 `[true,false]`, `[true,true]`, `[true,false]`로 코드 흐름과 부합한다.

g-star·g-globstar는 문법 항목이므로 `src/`·`.go`·whole-name slash matching을 공유 계약으로 명시해야 완전한 pattern instance에 연결된다. g-mid-component의 실제 예는 `path/to/**.txt`이며 `src/**.go`를 직접 열거하지 않는다. 같은 구조라는 제한된 source 연결은 가능하지만 전체 문구가 모든 배치·오류·panic·Windows·filesystem 동작을 보장한다고 확대하지 않는다. 반환 err=nil·supported·no panic은 별도로 검증할 관측 필드다.

## 제안 영어 입력과 연결 범위

원래 모델 입력을 번역하지 않는다. 아래는 새 영어 **제안 초안**이다. 최종 입력/후보 binding/Want/worker freeze를 대신하거나 새 정답을 만들지 않는다.

1. `For inputs 1.2.3, v1.2.3, and 1.2, accept only the complete unprefixed version; reject both other inputs.` —105B/21words.
2. `For inputs 1.2.3, v1.2.3, and 1.2, accept all three, coercing the prefix or missing patch into a semantic version.` —114B/24words.
3. `Using whole-name slash matching with src/ and .go fixed, match both src/main.go and src/lib/main.go, allowing zero or nested directories.` —137B/25words.
4. `Using whole-name slash matching with src/ and .go fixed, match src/main.go but reject src/lib/main.go; wildcard matching must stay within one path segment.` —155B/28words.

Semver의 두 entrypoint·Version·String observer·검증 helper·regex/init/flags는 하나의 연결 가족이다. glob의 세 pattern instance·Match·재귀 helper·validation·sentinel도 하나의 연결 가족이다. 요청을 둘씩 나누거나 wrapper를 추가해 독립 가족 수를 늘리지 않는다. 원래 membership에 연결하는 실제 graph union/역할 배정은 하지 않았다. utils.go의 전체 compile closure는 filesystem API 실행 승인이 아니다.

g01–g05와 Strict의 prefix/missing-patch 관측은 원래 JSON pointer로 중복을 표시했다. 이는 새 Got나 정답으로 복사되지 않는다. NewVersion 세 입력과 mid-component main 위치는 새 관측 자리다. 원래72/216·unknown21·17groups/known16·roots와 prior review는 변경하지 않았다.

## 실행 전 남는 구체 범위

v3는 Strict3/New3/Match6의 entrypoint12회와 NewVersion String observer3회를 구분한다. Strict.String은 호출하지 않는다. initializer/helper의 개별 호출 수를 계측했다는 주장을 하지 않고 child CPU/RSS에 초기화 비용을 포함한다. 합성 test package는 upstream을 import하지 않았다.

원래 계획에 따라 최종 영어 입력과 caption→entrypoint/pattern 연결, Want와 관측 schema, 전체 원문/license/compile closure/worker/plan SHA를 첫 실행 전에 고정해야 한다. 이 보고서는 준비성·fit·일반화·LLM 비용 절감 승인이 아니다. 실제 원문을 후보로 사용하면 프로젝트 직접 작성 문구와 다른 wording origin을 기록할 수 있지만, 두 관련 가족·함수 이름 shortcut·재사용 literal만으로 human-only·독립 population·학습 자격이 증명되지 않는다.

AI-assisted source/text review1회와 metadata serializer1회 성공을 수행했다. 합성 test 첫 시도는 자체 코드의 쉼표 누락으로 compile 실패했고 수정 후3test가 race 통과했다. 첫 실패 파일과 출력도 남겼다. 정확한 ledger와 receipt에 실제 시도/실패를 기록한다. 원본 API/init/AST/parser/formatter·actual Features·role·fit·model/judge/paid API·protected-final·shared edit·Git·외부 게시0이다. 이는 별도 프로세스/API의 범위이며 일반 Codex 협업에 사용된 AI나 그 비용이0이라는 뜻은 아니다. 협업 비용은 측정하지 않았다.
