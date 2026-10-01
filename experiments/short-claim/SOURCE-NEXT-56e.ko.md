# 실제 공개 원천으로 넓히는 다음 단계

56e의 유한 quoted 가족만 반복해도 데이터 다양성은 늘지 않는다. 다음은 다른 원천의 의미 계약을 실제로 구현하기 위한 조사다. GitHub API에서 고정 revision의 root tree, 라이선스와 구현 파일을 읽었다. **새 개발 요청·독립 truth·모델 결과 확보는 아직0**이며 기존120개 inventory를240개 완료로 바꾸지 않는다.

| 읽은 공개 원천 | 고정 revision | 확인한 파일 | 우선 행동 |
|---|---|---|---|
| [Masterminds/semver](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9) v3.4.0 | `61fc460d28283a91c53be65c2e0f20b494ac8ad9` | `version.go`, root tree, `LICENSE.txt` | 엄격 parsing·버전 우선순위·metadata 무시 |
| [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99) v4.9.1 | `8b690afa33319b0a1869367f594e53977e38bc99` | `match.go`, root tree, `LICENSE` | 디렉터리 glob·escape·잘못된 pattern 오류 |

두 고정 LICENSE는 MIT 문구와 저작권 표시를 포함한다. 이후 소스를 실제로 복사/배포하면 각 원본의 저작권과 라이선스 전체를 유지한다. 현재는 원문 코드를 vendor하거나 가중치에 사용하지 않았고, 이 조사만으로 전체 dependency·학습 자산의 권리 검토가 끝났다고 하지 않는다. 파일 byte SHA, 닫힌 helper/dependency, compile provenance와 NOTICE 여부는 실행 가능한 원천 manifest 단계에서 추가로 확인한다.

## 서로 다른 행동을 실제로 검사하는 계약 제안

1. **Strict semver parsing:** 유효한3개 숫자 구간만 허용하고, `v` prefix·빠진 구간·앞자리0을 거절하는 유한 요청. nil/error identity와 값 필드를 독립 literal oracle로 확인한다. permissive `NewVersion`의 전역 `CoerceNewVersion` 설정을 strict 계약과 혼합하지 않는다.
2. **Build metadata 무시:** 같은 major/minor/patch/prerelease에서 다른 metadata의 비교 결과는0이어야 한다. 문자열 자체의 equality와 의미 우선순위를 구분한다.
3. **Prerelease 순서:** 숫자 식별자는 수치로 비교하고 정상 release보다 낮으며 숫자와 문자 식별자를 구분하는 유한 입력. `Compare`, `LessThan`, `Equal` wrapper들은 같은 helper 그룹에 묶고 독립 요청처럼 부풀리지 않는다.
4. **Glob 디렉터리:** `a/*.go`와 `a/**/*.go`의 구분, zero-directory와 nested-directory 사례를 원래 `/` separator 계약으로 관측한다. macOS `PathMatch`와 `Match`를 혼합해 플랫폼 차이를 의미 학습으로 세지 않는다.
5. **Glob escape:** escape한 wildcard의 literal 동작과 한 글자 wildcard의 separator 제외를 분리한다. 길이/ASCII 범위를 먼저 제한하고 해당 입력만 검증한다.
6. **Bad-pattern error:** malformed pattern은 `ErrBadPattern`을 정확히 요구한다. `MatchUnvalidated`가 오류를 버리는 bool wrapper라는 점을 숨기지 않는다. false만 반환하면 오류 계약의 정답이 아니다.

이 행동은 upstream에 이미 구현되어 있으므로 새 코딩 과제가 미구현 상태라는 뜻이 아니다. 우선은 모델의 검증 전 설명·힌트를 위한 **코드 의미 원천**이다. 실제 코딩 완료율/모델 라우팅용 요청은 별도 baseline gap과 독립 수용 조건이 있어야 한다.

## 실행 가능하게 만드는 순서

원문 byte와 권리 고정 → source/helper/alias closure·닫힌 Go compile → 소스 관측 전 요청/설명·독립 기대값 작성 및 검토 → 원천을 바꾼 mutation 대조와 unknown 처리 → 관계 그룹 감사 → 별도 공식 truth 봉인/관측 → 후보를 모두 보존하는 비학습 효용 계획이다. 일반 원문 함수가 가진 경계를 좁은 literal 검증으로 인증하지 않는다. 세 semver 행동과 세 glob 행동이 자동으로6독립 가족이 되는 것은 아니다.

그 뒤120→240의 실제 자료 수, 저장소·언어·작성 흐름·행동 종류 분포를 확인한다. 목표 숫자를 빈 행이나 문장 변형으로 채우지 않는다. 최소15그룹·필요 효용5%·별도 도메인별 고유 보호 최종2400개 기준을 유지한다. 새 paid/model calls·fit·가중치·보호 최종 접근0. 모델 본체는 자격을 갖춘 immutable HF 배포가 있을 때만 공개하고 Git에는 넣지 않는다.
