# v-prefix만 다루는 새 계약에 대한 비판 검토68

새 계약의 목표를 고정 입력 `v1.2.3`의 **leading v를 허용하는가/거부하는가**로 좁히면, 기존 upstream 문구로 의미를 연결할 수 있다. 다만 소개 bullet은 semver **라이브러리의 capability**이며 NewVersion 함수 주석이 아니다. 출처를 그 수준 그대로 표시하고 NewVersion에 대한 source binding을 명시해야 한다.

원문 전체 bullet은 다음 raw bytes다. 실제 후보에서는 두 칸 들여쓰기·dash·backtick·끝 LF까지 그대로 유지한다.

```text
  - Optionally work with a `v` prefix
```

doc.go의9행, byte `[260,298)`,38B/6normalized words, SHA `5780602667a40d7af64f4f879c65194d0e046d78ff78c614f73c7dd6a012453b`다. 전체 원본 doc.go SHA는 `ad987af91a2a2c8dd5d01a38364263c53d294918670bed23edfad865d14afc22`다. 축약이나 새로운 paraphrase 없이512B/32words cap에 든다. JSON에는 전체 MIT 고지도 원본1078B 그대로 있다.

문단 연결은 추측으로 함수 이름을 붙이는 방식이 아니다. 같은 doc.go의 parsing 문단은 StrictNewVersion가 valid-v2만 파싱하고 NewVersion이 leading v가 있는 버전을 coerce한다고 설명한다. pinned NewVersion→coerceNewVersion→anchored loose regex도 optional `v?` 경로를 실제로 갖는다. 따라서 **package-level optional-prefix capability를 pinned NewVersion(v1.2.3)에 instantiate한다**는 명시적인 candidate 계약에는 source/doc 근거가 있다. Strict의 기존 완전 문장과 digit-only 검사, leading v가 SemVer 자체에 없다는 원문 comment는 같은 입력에 대한 반대 목표를 뒷받침한다.

이 연결 아래에서 v-prefix 허용/거부라는 좁은 semantic property는 contextual coverage가 있다. bullet 단독으로 “NewVersion라는 함수의 모든 입력 성공을 보장한다”거나 “모든 semver API가 prefix를 허용한다”고 해석하면 unsupported다. 실제 candidate→NewVersion 연결이 없거나 보고서가 임의로 주장하기만 하면 coverage gap을 그대로 남겨야 한다. 함수 binding은 provenance이며 runtime 문구에 새로운 설명을 덧붙이지 않는다.

제안 영어 요청은 다음 두 개다. 검토자 작성 AI-assisted 제안이므로 root의 최종 계약 freeze나 별도 작성자 독립 승인을 대신하지 않는다.

1. `For v1.2.3, reject the leading v prefix when parsing a semantic version.` —72B/14words.
2. `For v1.2.3, allow the optional leading v prefix when parsing a semantic version.` —80B/15words.

두 요청의 model target은 이 literal의 prefix 행동 하나뿐이다. 기존 세 literal 관측 Want와 String auxiliary metadata는 유지할 수 있지만 missing-patch 성공·정확한 문자열·error·panic·메모리 성능은 이 caption target에 포함하지 않는다. 예상 Strict reject/New allow는 정적 source reading이며 실제 accepted Got나 새 정답이 아니다. `1.2.3` control 입력은 prefix를 제외한 부분이 파싱됨을 관측할 수 있는 aux 자리로 남는다.

반대 요청 두 개와 Strict/NewVersion는 동일 Version/검증 helper/regex/init/flags 및 String observer를 공유하는 한 whole family다. glob 계약은 이 변경으로 넓어지지 않는다. 다른 upstream wording origin을 확보할 수 있다는 가설과 human-only·independent population·학습 준비 승인도 구분한다.

v3 report SHA `e1b2eb0e247fd22d98734d14dce535b4d85bf261f1d3eb75344a05d2bf65b814`를 입력으로 확인했으며 v3 generic-coercion fullcoverage 한계와 모든 v3 artifact는 수정하지 않았다. 여기서 v4 최종 계획을 읽거나 승인한 것은 아니다. root가 최종 정확한 입력·quote binding·target/aux schema·source/worker를 freeze한 뒤의 실제 실행은 별도 기록해야 한다.

이 단계의 private metadata serialization은1회 성공/실패0이며 원본 API/init/AST/parser/formatter/actual feature/role/fit/inference/judge/paid 실행0이다. 새 truth/acceptable set·shared edit·Git·외부 게시0이다. AI-assisted Codex 검토의 일반 비용은 측정하지 않았고 별도 API/process0을 전체 AI 비용0으로 표현하지 않는다.
