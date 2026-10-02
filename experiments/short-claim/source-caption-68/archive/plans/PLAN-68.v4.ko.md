# 실제 공개 원문을 쓰는 작은 행동 힌트 계약68 · v4

v3와 그 독립 검토를 수정하지 않고, Semver 학습 목표만 **고정 입력 `v1.2.3`의 leading-v 허용/거부**로 좁힌 별도 계획이다. v3에서 NewVersion의 generic attempts 문장이 세 입력의 성공과 정규화 전체를 약속하지 못했던 한계를 보존한다. 기존72요청·216후보·unknown21·17연결그룹·정답과 mask는 변경하지 않는다. 이번 계획은 새 관측과 감독 제안의 준비이며 실행 결과나 학습 준비 승인이 아니다.

서로 관련된 영어 요청4개를 두 전체 source 가족으로 묶는다. 같은 source의 두 요청을 독립 그룹으로 세지 않는다. 요청은 아래 AI-assisted 제안 문구를 root가 이번 계획의 최종 유한 입력으로 선택한 것이다. 원래 모델 입력을 번역하지 않는다.

1. `For v1.2.3, reject the leading v prefix when parsing a semantic version.`
2. `For v1.2.3, allow the optional leading v prefix when parsing a semantic version.`
3. `Using whole-name slash matching with src/ and .go fixed, match both src/main.go and src/lib/main.go, allowing zero or nested directories.`
4. `Using whole-name slash matching with src/ and .go fixed, match src/main.go but reject src/lib/main.go; wildcard matching must stay within one path segment.`

첫 두 요청의 후보는 StrictNewVersion와 NewVersion의 고정 invocation이다. Strict 설명은 원래 doc.go의 완전한 valid-v2 문장 [386,497),111B를 사용한다. NewVersion 설명은 doc.go의 package-level bullet 전체 [260,298),38B·6normalized words·SHA `5780602667a40d7af64f4f879c65194d0e046d78ff78c614f73c7dd6a012453b`다. indentation·dash·backticks·LF를 그대로 보존한다. 이 bullet은 함수 전용 주석이 아니다. **upstream library capability instantiated as pinned NewVersion(v1.2.3)**라는 provenance binding을 따로 고정한다. 같은 doc.go parsing context [328,852) 및 NewVersion→coerce→optional-v regex가 그 연결의 근거다. binding·함수명·ID·Want·정답·역할·검토문은 학습 특징에 추가하지 않는다.

후보 문장이 모든 Semver API나 임의 v 입력·error·String·panic·성능을 보장한다고 말하지 않는다. 목표의 사전 Want는 Strict가 `v1.2.3`를 reject하고 NewVersion이 allow하는 것이다. 요청1의 제안 acceptable은 Strict 위치0, 요청2는 New 위치1이다. 실제 관측 전에는 기대·제안으로만 저장한다. CoerceNewVersion=true 및 DetailedNewVersionErrors=true를 source와 실행 전후값으로 확인하며 변경하지 않는다. accepted 관측 정의는 returned && !panicked && error=nil && Version!=nil이다. 성공 객체 및 error/panic 필드는 목표 bit을 확인하는 부가 근거이며 후보 문구의 전체 계약 약속으로 보지 않는다.

원래 세 literal `1.2.3`, `v1.2.3`, `1.2`의 사전 accepted Want `[true,false,false]`와 `[true,true,true]`는 그대로 별도 aux 관측이다. NewVersion.String의 `1.2.3,1.2.3,1.2.0`도 auxiliary metadata다. missing-patch·정확한 문자열은 두 Semver 요청의 모델 목표가 아니다. v3의 부족한 caption을 새 감독으로 승인하지 않는다. NewVersion은57의 Strict-only 관측을 확대했다고 주장하지 않는 새 API 범위다.

Glob 두 요청의 name은 `src/main.go`, `src/lib/main.go`이고 후보 pattern은 고정 `src/*.go`, `src/**/*.go`, `src/**.go`다. 설명은63의 g-star·g-globstar·g-mid-component 완전 원문이다. prefix/suffix와 whole-name slash matching은 모든 후보에 공통인 요청의 명시 전제다. 사전 matched Want는 `[true,false]`, `[true,true]`, `[true,false]`; 제안 acceptable은 요청3의 위치1, 요청4의 위치0과2다. g-mid-component 원문의 txt example을 go input으로 축약/바꾸지 않는다. 원문 구조와 고정 source instance의 contextual 연결을 기록하며 모든 배치·파일시스템·Windows·잘못된 pattern·error/panic을 보장한다고 확대하지 않는다. err=nil·supported·no panic은 부가 관측이다.

기존57과 중복된 literal/관측은 원래 pointer로 표시하며 고유 요청 수를 부풀리지 않는다. mid-component main 관측은 새 위치로 구분한다. Semver의 두 후보와 요청은 Version·String observer·regex/init·검증 helper·flags를 공유한다. Glob의 세 pattern instance와 두 요청은 Match·재귀 helper·validation·sentinel을 공유한다. 기존 observer/표준 import를 전체 grouping에서 다루던 정책을 유지하며 원래 그룹을 쪼개지 않는다. 원래 membership와 연결하는 실제 union·role 배정은 이번 실행에서 하지 않는다.

첫 원본 worker의 planned entrypoint 호출은12회(Strict3/New3/Match6)이며 NewVersion.String observer3회는 별도다. Strict.String은 호출하지 않는다. package init와 내부 helper의 개별 호출 수는 계측하지 않으며 child CPU/RSS에는 초기화 비용도 포함된다. concurrency1·wall5분·OS peak RSS256MiB의 상한을 계획한다. Go heap soft limit은 OS RSS hard cap이나 RSS 측정이 아니다. 원본 worker를 compile하는 것과 실제 package init/API를 실행하는 것을 구분한다. upstream을 import하지 않는 별도 synthetic callback tests를 사용하고, 최종 입력·candidate binding·Want·15원문자산과 MIT전문2개·worker 및 전체 compile closure·binary SHA·실행계획을 **첫 원본 실행 전에** 고정한다. 실패하면 Want나 기존 결과를 고치지 않으며 automatic retry/seed search를 하지 않는다.

실제 native Got를 얻은 뒤 source/text fidelity와 **이번 좁은 target coverage**를 독립 reader가 별도로 확인해야 새 감독을 제안할 수 있다. 기존 프로젝트 직접 작성 문구와 다른 upstream wording origin을 실제 후보로 썼다는 근거를 기록하되, copyright 이름을 individual/human-only 작성 인증으로 바꾸지 않는다. source 후보2가족·관련 요청4개·재사용 literal·함수명 shortcut과 contextual attribution 한계를 유지한다. 이 작은 개발 계약은 protected-final2400요청, 일반화, 절감 또는 전체 학습 자격을 증명하지 않는다. old `synthetic_single_pipeline` history는 보존한다.

이번 native 관측에서 feature projection·role·fit·model/judge/paid trial·protected-final은0이다. 향후 통합 fit은 전체 membership/role/eligible masks/source scope와 기존5% headroom, 기존9/3/3 기준을 별도 freeze한다. 새 모든프로토타입-per-role·15저장소·2400-per-fit 숫자 기준을 추가하거나 mask/seed를 유리하게 바꾸지 않는다. 모델 본체는 Git에 올리지 않고, 적격 immutable HF publication은 실제 효용·license·재현·CI 확인 뒤 판단한다. AI-assisted 협업 비용은 측정하지 않았다. 새 Go 코드는 Apache-2.0이며 upstream 전체 MIT 고지를 보존한다.
