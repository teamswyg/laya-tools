# 공개 원문을 실제 유한 행동 후보로 연결68

이 작업은 59/63의 원천 준비를 실제 입력·후보 계약으로 구체화한다. 별도 작은 영어 행동 힌트 개발 자료이며 기존72요청·216후보·unknown21·검토 정답을 변경하지 않는다. 원문 자산을 실제 후보로 쓰지 않은 준비 단계와 구분한다. 아직 실행·새 정답·학습 준비 승인은 없다.

두 원천 가족을 대상으로 **관련된 요청4개와 연결 그룹2개**를 준비한다. 요청은 프로젝트의 AI-assisted 작성이고 후보 설명은 기존 upstream 완전한 원문 문장/문법 항목이다. 원문 바이트·주석 prefix·줄바꿈·문법 기호를 그대로 입력한다. 번역·축약·paraphrase로 다양성을 만든 것으로 표현하지 않는다. 개별 저자/human-only 인증은 없다. 함수 이름이 원문에 들어가는 것을 보존하고, 그 이름에 의존한 학습 가능성도 한계로 기록한다.

Semver의 요청 두 개는 동일 유한 입력 `1.2.3`, `v1.2.3`, `1.2`만 다룬다. 첫 요청은 complete/unprefixed인 첫 입력만 accept하고 나머지 두 개를 reject하라고 요구한다. 둘째는 prefix/missing-patch를 coerce하여 세 입력을 모두 accept하라고 요구한다. 후보는 StrictNewVersion와 NewVersion 두 개이며 exact doc.go의 완전한 설명 문장을 사용한다. 두 entrypoint는 동일 original closure/type/helper/flags를 공유하므로 같은 그룹이다. 전체 오류 sentinel 우선순위·uint64/Unicode·비교·임의 SemVer 입력의 정확성은 목표가 아니다.

이번 별도 계약은 **NewVersion API를 새로 관측하는 범위**를 명시한다. 59/63의 StrictNewVersion-only 자료나57 관측을 NewVersion 실행 증명으로 확대하지 않는다. 사전 기대값: Strict accepted=[true,false,false]; NewVersion accepted=[true,true,true], 문자열=[1.2.3,1.2.3,1.2.0]. 원본 CoerceNewVersion=true, DetailedNewVersionErrors=false를 source·실행 전후값으로 확인하며 변경/병렬 호출하지 않는다. accepted는 panic이 없고 error=nil이며 Version!=nil일 때다. 실제 반환값이 다르면 Want를 고치지 않고 실패·차이를 남긴다.

Slash Match의 두 요청은 `src/main.go`, `src/lib/main.go`라는 동일 name 두 개만 다룬다. 첫 요청은 둘 다 match를 요구하고, 둘째는 main만 match·nested는 non-match를 요구한다. 후보 세 개는 `src/*.go`, `src/**/*.go`, `src/**.go`로 원본 Match를 호출하는 고정 pattern instance다. 설명은63의 g-star/g-globstar/g-mid-component 완전 원문이다. prefix/suffix와 target names는 모든 후보에 공통인 요청 계약의 전제다. pattern/후보·wrapper·helper·Match 전체 closure 관계를 같은 그룹으로 보존한다.

사전 기대 matched는 각각 [true,false], [true,true], [true,false]다. 기존57의 g01–g05와 중복된 관측은 새 고유 요청으로 세지 않는다. mid-component main의 새 관측을 기존57에 추가하거나 덮어쓰지 않는다. 반환 error=nil·supported·no panic은 관측 지원을 확인하는 부가필드이며, 후보 문장이 오류·panic·메모리 전체 계약을 설명한다고 주장하지 않는다. PathMatch·filesystem·Windows separator·malformed pattern·arbitrary names는 범위 밖이다.

실제 source worker가 만든 관측과 별도로 독립 reader가 요청 전제/후보 문구의 충실성·목표 coverage·finite Want를 검토해야 한다. source/text/captions/expectations/원문 고지·compile closure·worker 및 실행계획 SHA를 **최초 원본 API 호출 전에** 고정한다. 이 계획 작성만으로 semantic eligibility를 true로 하지 않는다. Candidate source API의 동일 helper·타입·전역 flags·파라미터 관계를 graph에 넣고 원래17그룹과의 observer/표준 import 정책을 동일하게 구분한다. source family별2요청을 서로 독립으로 세지 않는다.

최초 원본 worker는12개 API calls(semver6/glob6), concurrency1, wall5분 상한, OS peak RSS256MiB 상한을 계획한다. 실제 RSS/CPU 측정과 Go heap limit은 구분한다. synthetic callback 검사와 실제 upstream 호출은 분리한다. 모델/judge/paid API·특징 추출·역할·fit·가중치·protected-final은0이다. 실패한 작업을 성공으로 덮어쓰거나 임의 재실행하지 않으며 원인을 고치면 새 시도/장부를 남긴다.

향후 개발 fit을 위한 작성 원천의 다른 부분을 실제로 확보하는 것이 목적이다. 이 네 요청만으로 일반화·최종2400요청·human-only 작성·전체 학습 자격을 승인하지 않는다. 원래 `synthetic_single_pipeline`의 역사 기록을 보존하고, actual-used caption provenance와 제한된 혼합 개발 scope를 별도 기록한다. 전체 그룹 membership/역할/coverage, 감독 mask와 동일5% headroom은 별도 계획에 묶은 후 판단한다. 새15저장소·새2400-per-fit·모든prototype-per-role 기준이나 유리한 seed 검색은 추가하지 않는다.

MIT 전문2개와 모든 원본 artifact 고지를 함께 보존한다. 프로젝트의 새 Go 코드는 Apache-2.0이다. 모델 본체는 Git에 올리지 않고, 향후 적격 모델의 immutable HF 공개는 rights/효용/재현/CI 확인 뒤 별도 기록한다. 일반 AI-assisted Codex 협업 비용은 측정하지 않았으며 별도 모델 프로세스0을 전체비용0이라고 표현하지 않는다.
