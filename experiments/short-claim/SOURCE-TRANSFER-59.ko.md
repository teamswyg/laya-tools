# 두 공개 원천의 작은 속성 전이 준비59

[English](SOURCE-TRANSFER-59.en.md) · [기계가 읽는 원천 연결](source-transfer-59.json) · [다음 준비 방향](NEXT-STORED-58.ko.md) · [기존 원본 관측57](../public-behavior/RESULTS-57.ko.md)

**두 가지 속성을 제안하고 기존 근거를 연결한 준비 자료다. 새 요청·caption·정답·허용 집합·역할 배정·학습은 아직 0이다.** 원본 소스·문서·라이선스와 저장된57 결과를 읽었으며, 원천 API·Bind·평가·Baselines·모델·유료 호출·보호 최종 접근은 이번 준비에서 0이다. `planned_reference_only`, `execution_eligible=false`, `training_ready=false`를 유지한다. 문서 생성이나 핀 기록은 실행 가능한 계약·문장 검토 완료·다양성 해소를 뜻하지 않는다.

58에서는 기존72개 개발 요청의 BM25 91회와 oracle 73회 사이에 약 19.7802%의 확인 여지가 있었다. 전체17그룹·라벨16그룹도 기존 하한15를 이미 통과했다. 이 결과를 보고 좋은 그룹을 나누거나 두 외부 원천을 새 독립 부모로 세지 않는다. 58의 순위·점수·그룹 비용은 이미 관측했으며, 미래 역할 규칙을 고정해도 미관측·블라인드 상태로 되돌아가지는 않는다.

## 무엇을 제안하나

| 제안 | 좁힌 질문 | 기존에 실제로 관측한 API | 아직 미지원인 범위 |
|---|---|---|---|
| Strict parsing | 명시한 유한 버전 입력을 임의 보정하지 않고, 실패의 nil/오류나 성공 필드를 정확히 돌려주는가 | `StrictNewVersion`,57의 `s01–s16` | permissive `NewVersion`, 보정 flags 변경, 모든 SemVer 입력의 일반 정확도 |
| Slash 경계 매칭 | 지정한 `/` 경로에서 `*`와 `/**/`가 단일 segment·zero/nested directory를 구분하는가 | `Match`,57의 `g01–g09` | `PathMatch`, Windows separator, 파일 시스템 탐색, 전체 glob 문법·오류 정책 |

초기 target은 **짧은 영어 요청과 영어 후보 설명의 유한 Go 행동 힌트**다. 이 한국어 안내는 번역 문서이며 한국어 모델 입력의 성능 검증이 아니다. 새 caption과 독립 기대값의 준비·검토, 모델 입력의 512바이트·정규화32단어 충실성, 후보 관계 감사는 후속 단계다. 이 요약 문장을 완성된 모델 입력이나 라벨로 사용하지 않는다.

## 원본 파일과 실제 동작의 연결을 구분한다

원천 핀은 [기존 manifest](../../internal/publicbehavior/upstream-manifest.json)를 사용한다. 15개 원본 artifact,99,035바이트를 보존하며 새 외부 다운로드는 없다. JSON에는 revision·module·원본/보관 경로·파일 크기·SHA를 정확히 기록했다. 원본 파일을 자르거나 다시 포맷하지 않는다.

**Semantic closure**는 행동을 결정하는 entrypoint·타입·상수·오류 sentinel·helper·관측 getter의 연결이다. **Compile closure**는 원본 파일 전체를 그대로 컴파일하기 위해 보존하는 파일 목록이다. 같은 파일에 있는 다른 함수가 컴파일된다는 사실이 그 API의 호출 권한이나 관측 근거를 추가하지 않는다. 이 기록은 정적 소스 독해이며, 새 compile/replay 검증을 실행하지 않았다.

| 원천 | 고정 revision·module | 원본 compile/고지 closure |
|---|---|---|
| [Masterminds/semver](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9) | `61fc460d28283a91c53be65c2e0f20b494ac8ad9`, `github.com/Masterminds/semver/v3`, 원본 Go1.21 | `version.go`, `collection.go`, `constraints.go`, `doc.go`, `LICENSE.txt`, `metadata/upstream.go.mod.txt`:6파일 |
| [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99) | `8b690afa33319b0a1869367f594e53977e38bc99`, `github.com/bmatcuk/doublestar/v4`, 원본 Go1.16 | `doublestar.go`, `glob.go`, `globoptions.go`, `globwalk.go`, `match.go`, `utils.go`, `validate.go`, `LICENSE`, `metadata/upstream.go.mod.txt`:9파일 |

두 module 파일은 같은 원문 바이트를 metadata 이름으로 보존한다. 원본 Go directive는 원천 메타데이터이고,57의 실제 실행 환경은 Go1.27.1/Darwin arm64였다. 원본 directive를 기록한 것으로 새 언어 조건의 실행을 증명하지 않는다.

### Strict parsing의 연결

[version.go L89](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/version.go#L89)의 `StrictNewVersion`은 `Version`, `num/allowed`, `containsOnly`, `validatePrerelease`, `validateMetadata`와 여섯 원본 오류 sentinel을 사용한다. 성공 결과 전체를 연결하려면 `Original`, `String`, `Major`, `Minor`, `Patch`, `Prerelease`, `Metadata` getter도 포함한다. 정확한 선언 줄과 파일 핀은 JSON에 있다.

원본 [파싱 문서 L13](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/doc.go#L13)과 `version.go`의 L84–88 설명을 연결한다. `doc.go` L13–19는 원본 바이트 `[328,853)`, `version.go` L84–88은 `[3039,3397)`이다. 바이트는0부터 시작하고 끝은 제외하며, 줄은1부터 시작해 양끝을 포함한다. JSON의 span SHA는 이 LF 포함 원문 범위를 고정한다.

`StrictNewVersion`은 보정 flags를 읽지 않는다. 파일 전체의 regex/global/init도 원천 핀에서는 보존하되, 이 사실을 `NewVersion` 관측으로 해석하지 않는다. 기존 observer가 직접 호출한 것은 strict API다. permissive 결과를 문서나 코드 독해만으로 Got에 추가해서는 안 되며, 새 계약·관측 계획 전에는 unknown이다.

### Slash 경계의 연결

[match.go L52](https://github.com/bmatcuk/doublestar/blob/8b690afa33319b0a1869367f594e53977e38bc99/match.go#L52)의 `Match`는 `/`, validation=true, case-insensitive=false를 전달한다. `matchWithSeparator`, 재귀 `doMatchWithSeparator`, `matchRune`, `isZeroLengthPattern`, escape/alternative 탐색 함수, `utils.go:indexNextAlt`, `validate.go:doValidatePattern`, 원본 `ErrBadPattern`이 semantic closure다.

문서 L9–51의 원본 바이트는 `[76,1977)`, 핵심 문법 L15–17은 `[197,373)`이며 JSON에 span SHA가 있다. 원문 `utils.go` 전체에는 `FilepathGlob`도 있어 glob/options/walk 코드까지 compile closure에 들어간다. 그래서 원본7개 Go 파일을 유지한다. 이것이 filesystem API를 실행했거나 실행하도록 허용했다는 뜻은 아니다.

## 저장된 Want/Got만 정확하게 참조한다

[probes-57.json](../public-behavior/probes-57.json)의 `vectors[].checks`는 사전 Want이고, [results-57.json](../public-behavior/results-57.json)의 `audit.rows[].checks`는 저장된 Want/Got/일치 여부다. 새 JSON은 값을 생성하거나 큰 정수를 다시 직렬화하지 않고 **정확한 JSON pointer·field 이름·타입**으로 연결한다.

- Strict의 기존16행: `error_cause`, `error_func`, `error_kind`, `error_num`, `major/minor/patch`, `metadata`, `nil_version`, `original`, `panicked`, `prerelease`, `string`, `supported`의14필드.
- Glob의 기존9행: 오류 관련4필드와 `matched`, `panicked`, `supported`의7필드.
- 저장 row는 input을 중복 보관하지 않는다. 저장 record의 `/probes_sha256`이 고정 원본 probes 바이트 해시와 같은지 먼저 확인하고, `/vectors/N/input`에서 입력을 읽는다. 각 reference의 expected checks와 저장 row/checks/status/match pointer를 확인하고 ID·property·source family·expectation kind를 대조한다. check index를 쓰기 전에 field 이름·타입도 맞아야 한다.
- `Want/Got`은 `json.RawMessage`와 정확한 uint64/bool/string 타입을 유지한다. float64나 일반 JSON 숫자 roundtrip으로 최대 uint64를 손실하지 않는다.
- `error_message`, `Versions`, 호출 counter 등 Want가 없는 관측값이나 기본값을 새 정답으로 승격하지 않는다. 이 reference는 cold metadata이며 scorer 특징이 아니다.

해당25행은 기존57에서 모두 complete/match였다. 이것은 기존 유한 관측의 참조 상태이며 새 replay·25개 독립 요청·새 caption 정답이 아니다. 부분 필드만 같다는 이유로 전체 함수 설명을 충실하다고 판정하지 않는다.

## 원래 차이와 불확실성을 유지한다

Strict API의 nilness·sentinel·오류 검사 순서는 기술적 계약이다. SemVer 규범이 정한 오류 종류가 아니다. `s06-core-overflow`는 uint64 표현 범위를 넘는 유효한 숫자 문법의 API 거절이며 SemVer-invalid로 바꾸지 않는다. 큰 prerelease 숫자의 파싱 성공도 비교 정확도를 증명하지 않는다.

57의 비교 불일치 `p08`·`p09`는 같은 큰 숫자 비교 양상의 반전 쌍이다. `b05`는 `Match("{a,[}","a")`가 true+nil을 반환한 실제 관측과 더 강한 전체 패턴 오류 정책의 차이다. 원본 Got을 고치거나 `ValidatePattern` 선행 검사로 합성하지 않는다. 세 불일치의 정확한 저장 row pointer도 JSON에 보존했다.

새 입력/후보·더 넓은 Unicode/문법·wrapper·flags 변경·새 강한 정책은 지원이 확인되기 전까지 unknown 또는 범위 밖이다. unsupported operation·UTF-8/byte bounds 거절·panic·unclassified error를 후보 실패 라벨로 바꾸지 않는다. 새 후보의 정답·오답 대조군은 아직 작성하거나 실행하지 않았다.

## 작성 원천·그룹·라이선스의 의미

semver와 doublestar는 프로젝트 이전의 공개 원본 코드·문서다. MIT 고지의 저작권자는 semver의 Matt Butcher/Matt Farina, doublestar의 Bob Matcuk이다. 이는 특정 함수의 모든 작성자를 확인했거나 human-only 작성이라고 인증했다는 뜻은 아니다. 프로젝트의 기존 자료는 AI 지원 협업 합성 영어 작성 흐름이며, 이번 transfer caption은 아직 없다. 외부 코드나 다른 reviewer만 추가해 `synthetic_single_pipeline`을 해소했다고 하지 않는다.

Strict의 실패·오류 관측은 기존 `atomic-commit`/`error-identity`와 비교할 축이고, glob의 문맥/경계는 `prefix-balance`/`quoted-delimiters`와 비교할 축이다. 유사성만으로 동일 그룹이나 독립 그룹을 확정하지 않는다. 실제 공유 helper/타입/sentinel·복사 코드·alias·중복 caption·부모/형제·번역 관계를 전이적으로 연결한다. 같은 upstream의 다른 속성도 공유 closure를 보존한다. 새 그룹 수는 계산하거나 주장하지 않았다.

원본 [semver LICENSE](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/LICENSE.txt)와 [doublestar LICENSE](https://github.com/bmatcuk/doublestar/blob/8b690afa33319b0a1869367f594e53977e38bc99/LICENSE)는 전체 문구·저작권 고지와 함께 보존한다. 원천 라이선스의 기록은 미래 학습 자산·가중치의 배포 자격 확인을 대신하지 않는다.

기존72부모·17그룹·라벨16그룹·unknown21과 결과 원문을 바꾸지 않는다. 문장 충실성, 역할·coverage 규칙, 별도 실행 계획의 자동 검사가 필요하며 새 사람 승인 흐름을 만들지 않는다. 모델/fit/새 HF 가중치·운영 활성화는0이다. 도메인별 서로 다른 보호 최종 요청 최소2,400개는 별도의 일반화 목표이며 모든 development fit의 선행 최소라는 뜻은 아니다. protected final/CoSQA reserve는 읽거나 재배정하지 않았다.
