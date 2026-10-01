# 공개 라이브러리의 작은 동작 주장 검증57

목표는 코드 후보가 어떤 작은 요구를 만족하는지에 대한 **검증 가능한 힌트 자료**를 늘리는 것이다. 버전 문자열 처리와 경로 패턴 매칭을 먼저 확보한다. 사용자나 에이전트가 “빌드 번호를 무시하고 버전 순서를 비교할 수 있나”, “하위 폴더까지 찾나”라고 물을 때 작은 모델이 확인할 후보를 제안하고, 실제 승인과 정답 검사는 기존 도구가 맡는 구성을 준비한다.

이 단계는 실제 원본 Go 함수와 사전 작성한 유한 입력 기대값을 대조한다. 일반적인 코드 이해, 완성형 라우터의 절감 효과, 새 학습 성공을 보여주는 실험은 아니다. 관측 결과를 보고 기대값을 고쳐 통과시키지 않는다.

## 원천과 권리

| 원천 | 고정 버전·리비전 | 사용할 동작 |
|---|---|---|
| [Masterminds/semver](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9) | v3.4.0 / `61fc460d28283a91c53be65c2e0f20b494ac8ad9` | Strict parsing, metadata, prerelease precedence |
| [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99) | v4.9.1 / `8b690afa33319b0a1869367f594e53977e38bc99` | slash/glob, escaping/character matching, pattern validation |

원본 15개 파일은 총99,035바이트다. 두 MIT LICENSE의 저작권·허용·면책 문구를 그대로 보존한다. Go 소스는 변경하지 않고 감사용 fixture에 둔다. 원래 `go.mod`는 metadata 하위에 보존하여 부모 모듈의 의존성을 바꾸지 않는다. 원본 경로와 배포 경로의 바이트·SHA 연결을 기록한다. SemVer 규범은 [2.0.0 원문](https://semver.org/spec/v2.0.0.html)을 링크하고 기대값을 별도로 작성한다.

## 여섯 속성, 두 원천 가족

1. Strict parser: 성공 시 원본 문자열·숫자 부분·prerelease·metadata, 실패 시 nil과 정확한 오류 종류를 보존한다. 특정 라이브러리의 오류 종류는 기술적 API 계약이며 SemVer 규범이 그 오류 이름을 요구한다는 뜻은 아니다.
2. Metadata: 문자열에 보존되지만 비교 우선순위에는 영향이 없는 경우를 확인한다.
3. Prerelease: 정식 버전, 숫자/문자, 점으로 나눈 접두사, ASCII 대소문자 순서를 확인한다. **uint64보다 큰 숫자 식별자도 포함**하여 규범과 구현의 차이를 놓치지 않는다.
4. Glob: 전체 이름·`/` 경계·`*`와 경로 구성요소인 `**`의 차이를 확인한다. 파일시스템과 OS별 `PathMatch`는 제외한다.
5. Escape/character: escaped wildcard와 `?`를 확인한다. ASCII 및 Unicode 경계는 입력별로 명시한다.
6. Validation: `ValidatePattern`의 bool과 `Match`의 bool/error를 각각 관측한다. 잘못된 brace의 첫 대안이 성공하는 경우를 포함한다. 모든 패턴을 먼저 거절하는 강한 정책과 원본 API의 동작 차이는 **정책 차이**로 구분한다.

원천 판독에서 큰 prerelease 숫자의 비교와 malformed brace의 조기 성공이 예상되었다. 이는 실행 전 가설이다. 기대값에는 규범·API 기술·별도 정책의 구분을 두고 원천 판독 가설을 정답으로 복사하지 않는다. alias·helper·반전 쌍·같은 요청의 입력 변형은 새로운 독립 작업이나 가족으로 세지 않는다.

## 실행 경계

Go 1.27.1, 순수 Go 감사 실행, 네트워크 없는 관측, 입력·파일 크기 제한을 사용한다. 원본 함수는 raw UTF-8 입력을 받는다. 소문자화나 구두점 제거는 API 입력에 적용하지 않는다. 정답과 source ID를 힌트 입력에 자동 채우는 경로를 만들지 않는다.

원천/관측기/계획/expectation을 고정하고 실제 binary SHA와 input freeze를 확인한 뒤 공식 수집을 한 번 수행한다. 각 속성 관측 수와 요청한 entry API 호출 수를 별도로 센다. getter/String/Original 호출은 별도 기록하며 원본 내부 helper 호출은 세지 않는다. Compare에는 두 strict parse와 한 compare가 필요하므로 한 관측을 한 함수 호출로 부풀리거나 줄이지 않는다. 오류·panic·unsupported는 별도로 보존하고 불완전 관측으로 false 라벨을 만들지 않는다. 수집 뒤 회귀 검사 재생은 새 공식 표본이 아니다.

이번 감사는 생산 `riidolaya`의 힌트 처리 경로에 라이브러리 로딩이나 lock을 추가하지 않는다. fixture와 감사기는 별도 유지보수 도구다. 메모리 soft limit이나 실행 시간 설정은 실측 자원 수치가 아니다.

## 다음 판단

기존56의 pending/unknown과 측정값을 보존한다. 신규 공개 동작 근거가 확보되어도 전체240개 원천, 최소15 관계 그룹, 필요한 비학습 효용5%, 도메인별2400개의 고유 보호 최종 요청을 달성했다고 선언하지 않는다. 지금은 새 fit·유료 모델 호출·최종 평가·가중치 배포를 하지 않는다. 다양성과 실제 확인 작업 감소를 검증한 뒤 학습 적격을 다시 판단한다.
