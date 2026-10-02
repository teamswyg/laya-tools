# 공개 원문 문구 자산 63 — private 준비

이번에는 새 문장을 만들지 않고, 이미 보존된 두 upstream의 **완전한 문장·정의·문법 항목 25개**를 원래 bytes 그대로 추출했다. 25개 모두 512 bytes/32 normalized words 이내이며 합계 1,818 bytes, 최댓값 26단어다. 내부 줄바꿈·comment prefix·기존 오탈자를 보존했다. 새 request/caption/candidate, expected value, acceptable set, truth/unknown 변경 및 학습 준비 완료 판정은 0이다.

`quote-catalog-63.json`은 Go 배열 기반 생성 결과다. 각 quote는 repository, immutable revision, original/retained file, 전체 file bytes/SHA, zero-based `[start,end)` raw quote bytes/SHA, one-based inclusive lines, 원문 URL, MIT license 파일/SHA와 연결된다. 전체 원본 15 artifacts/99,035 bytes를 확인했고, **MIT 전문 2개도 그대로 포함**했다. 59의 기존 documentation span 4개 및 57의 strict 16행/slash 9행/287개 field 참조를 확인했다. 이 읽기는 후보 실행이나 57 재실행이 아니다.

- Semver: `Masterminds/semver@61fc460d28283a91c53be65c2e0f20b494ac8ad9`, `version.go`/`doc.go`, `StrictNewVersion`와 이미 관측한 getter 문구만.
- Doublestar: `bmatcuk/doublestar@8b690afa33319b0a1869367f594e53977e38bc99`, `match.go`, `/` 기준 `Match` 문구만.
- 원본 license와 provenance를 옆자료로 둔다. source 이름·hash·copyright·review state는 features가 아니다. MIT 보존만으로 향후 학습·weight 배포 적격성을 승인하지 않는다.

원문은 **pre-existing pinned upstream wording**이고 57의 input/finite Want는 **프로젝트 AI-assisted 작성 자료**다. 둘을 분리했다. MIT notice의 Matt Butcher/Matt Farina, Bob Matcuk은 저작권 고지이며 개별 구절의 authorship 또는 human-only 작성 여부를 확인한 결과가 아니다. 검증 여부는 false/unknown이다. 이 catalog만으로 `authoring_diversity_cleared=true`나 `synthetic_single_pipeline` 해제를 선언하지 않는다.

## 문구가 표현하는 필드와 표현하지 않는 필드

`s-doc-strict`는 엄격한 semantic-version 파싱 범위를, `s-parse-result`는 Version 또는 error라는 반환 형태를 설명한다. “valid만 파싱한다”는 문구는 모든 유효 문자열을 수용한다는 뜻이 아니다. error 정의들은 일부 입력 조건과 upstream sentinel 이름을 설명하지만 `semver_empty` 같은 **프로젝트 adapter tag**, nil 반환, 실패 시 getter 기본값 또는 error 우선순위를 명시하지 않는다. getter 문장 7개는 `major/minor/patch/original/string/prerelease/metadata`의 의미를 설명하며 각 저장 숫자·문자열을 직접 적은 것은 아니다.

| 기존 strict 행 | 연결한 원문 | 이번 문구만으로 표현되지 않는 부분 |
|---|---|---|
| s01 empty | `s-error-empty` + strict/parse 문장 | `nil_version`, adapter tag, 정확한 빈 error 부필드와 실패 getter값 |
| s02 missing patch | `s-error-invalid` + strict 문장 | 정확한 3-part 문법 및 이 입력을 거절하는 별도 완전 문법 정의 |
| s03 v prefix | `s-error-characters` + strict 문장 | v prefix를 거절하는 구체 문법; permissive `NewVersion` 결과 |
| s04 core leading zero | `s-error-leading-zero` | 실제 nil·tag·우선순위 및 어느 numeric segment인지의 전체 문법 |
| s05 empty minor | generic parse/error 문장 | 저장된 **NumError/syntax** 및 `error_func/cause/num`; invalid-character sentinel로 바꾸지 않음 |
| s06 core overflow | generic parse/error 문장 | uint64 범위와 NumError/range. 유효 decimal syntax를 SemVer-invalid로 만들지 않음 |
| s07 invalid metadata / s08 empty metadata | `s-error-metadata` | 허용 문자·empty 규칙의 완전 문법과 정확한 adapter/return 필드 |
| s09 empty prerelease / s11 empty prerelease segment | `s-error-prerelease` | segment 문법, empty 제한과 정확한 반환값 |
| s10 numeric prerelease leading zero | `s-error-leading-zero` | prerelease numeric/non-numeric 구분 전체 문법; invalid-prerelease tag로 바꾸지 않음 |
| s12 error priority | `s-error-metadata` + generic 문장 | 여러 오류가 동시에 있을 때 metadata가 우선이라는 정책 |
| s13 zero version | strict + getter 문장 | 정확한 0.0.0 필드값·nil=false·성공 시 빈 error 부필드 |
| s14 maximum core | strict + getter 문장 | uint64 최댓값과 overflow 경계. 저장 최대 정수는 RawMessage/uint64로 유지 |
| s15 fields preserved | `s-get-original` 등 getter 7개 | 원문이 exact saved values를 열거한 것은 아님; String 형식·숫자·prerelease/metadata grammar 전체 |
| s16 overflow prerelease accepted | strict + getter 문장 | 매우 큰 numeric prerelease 허용 폭·정확한 필드값·**비교 정확성** |

strict 16행의 기존 체크 14필드는 `error_cause/error_func/error_kind/error_num`, `major/metadata/minor/nil_version/original/panicked/patch/prerelease/string/supported`다. quote topic 연결은 어느 필드의 뜻을 일부 표현하는지의 옆자료다. **어느 quote도 14필드를 한 번에 승인하는 scalar가 아니다.** 원본 Want/Got/matches와 pointer는 유지했으며 새 oracle를 만들지 않았다. `error_message`, counter, Versions 등 Want가 없던 관측을 새 truth로 승격하지 않았다.

`g-star`는 non-path-separator sequence, `g-globstar`는 `/**/`의 zero-or-more directories, `g-forward-slash`는 `/` 분할, `g-whole-name`은 전체 name matching을 표현한다. `g-mid-component`는 원래 완전한 한 문장으로 `path/to/**.txt`와 `path/to/*.txt`의 같은 결과를 명시하며 26단어다. 원문을 줄여 만든 caption이 아니다.

| 기존 slash 행 | 연결한 원문 | 표현 범위의 한계 |
|---|---|---|
| g01 one segment / g02 does not cross slash | `g-star`, `/` 분할 | 개별 literal Want 자체를 문구가 열거하지 않음 |
| g03 zero directory / g04 nested directory | `g-globstar`, `g-component` | `/**/` 문법 항목의 좁은 범위; 전체 glob grammar 보증 아님 |
| g05 mid-component double star | `g-mid-component`, `g-component`, `g-star` | 원래 문장에 나온 pattern 형식의 의미; 모든 pattern 정책으로 확대하지 않음 |
| g06 trailing globstar zero | entrypoint/전체 name/component 문구 | **`src/**`가 `src`에 match하는 경계를 짧은 `/**/` 문구가 명시하지 않음**; 기존 관측만 보존 |
| g07 whole name anchor | `g-whole-name`, `g-star` | substring이 아닌 전체 matching이라는 범위 |
| g08 star empty name | `g-star` | any-sequence 문구; 빈 name의 literal Want를 직접 적은 문구는 아님 |
| g09 globstar crosses slash | entrypoint/전체 name/component 문구 | **slash 없이 전체 pattern인 `**`를 짧은 `/**/` 항목이 명시하지 않음**; 기존 관측만 보존 |

slash 9행의 기존 체크는 네 error 필드와 `matched/panicked/supported`다. quote는 `matched` 의미의 일부를 표현하지만 저장된 no-panic, wrapper 지원 여부, 빈 error 부필드를 보증하지 않는다. `g-error`는 upstream의 ErrBadPattern 설명만 보존한 참고 문장이다. 기존 malformed-pattern discrepancy를 지우거나 새 strict validation gate를 만든 것이 아니다. `NewVersion`, `PathMatch`, filesystem 및 다른 property의 신규 동작·expected는 만들지 않았다.

## 실제 실행 기록과 다음 사용 경계

원문 생성 helper 2회는 모두 exit 0, independent byte/reference verifier 1회도 exit 0였다. 첫 generator 결과 뒤 self-review에서 row ID 별칭 및 sentinel/topic 과대 연결을 발견했다. 첫 catalog는 `quote-catalog-63.attempt1.json`으로 보존하고 보류했으며, 연결을 수정한 최종 catalog를 새 exclusive file로 생성했다. **첫 self-review 실패 1회**와 helper execution failure 0을 구분했다. synthetic race 명령 3회는 모두 통과했으며 실제 test 실행은 2회, 마지막 명령의 test 결과는 cache 재사용이다. 마지막 실제 race는 1.606초다. vet 명령 3회도 모두 통과했다. verifier는 원본 bytes와 25 quote/16+9행/287필드를 확인했다; 내용 전반의 승인이나 새 독립 origin 실험이 아니다.

자세한 attempt/실패 ledger는 `PREPARATION-LEDGER-63.json`, 기계 검증 receipt는 `quote-verification-receipt-63.json`에 있다. 원본 API/AST/formatter/Rebind/SourcePins/Bind/Generate/behavior/model/fit/paid/role/seed/newlabel/candidate/acceptable/Git/shared edit/publication은 0이다. 원문 문구의 임의 축약·번역·재작성은 0이며 이 한영 메모는 분석 설명이지 model input이 아니다.

다음 작업은 이 자산을 이용할 때의 좁은 입력·완전한 후보 보존·감독 적격성·group 관계를 기존 계획에 따라 고정하는 것이다. 정확한 upstream wording을 확보한 사실과 학습 준비·diversity 승인·성능 입증은 별개다. 이번 준비에는 새 gate·human approval·readiness 판정을 추가하지 않았다.
