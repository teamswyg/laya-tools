# 공개 동작 검증57 결과

**실제 원본 Go 함수 관측62건을 확보했다. 사전 유한 기대값59건과 일치하고3건이 달랐다.** 두 건은 큰 prerelease 숫자 비교의 같은 실패 양상을 정방향·역방향으로 확인한 것이다. 나머지 한 건은 malformed glob의 조기 성공과 더 엄격한 검증 정책의 차이다. 모델 정확도·일반적 함수 정확도·2400개 독립 작업 결과로 해석하지 않는다.

## 무엇을 알게 되었나

| 관측 | 사전 기대 | 실제 원본 결과 | 의미 |
|---|---|---|---|
| `1.0.0-99999999999999999999` vs `1.0.0-100000000000000000000` | SemVer 숫자 비교: 앞이 작음(-1) | 앞이 큼(+1) | 둘 다 strict parse가 성공하지만 비교에서 uint64 overflow가 문자열 경로로 이동한다. 이 고정 입력의 규범 차이를 보존한다. |
| 위 쌍의 역방향 | +1 | -1 | 같은 양상을 확인하는 반전 입력이며 독립 버그·작업을 추가하지 않는다. |
| `Match("{a,[}", "a")` | 더 강한 전체 패턴 검사 정책: false + ErrBadPattern | true + nil | 첫 대안 성공 뒤 잘못된 뒤 대안을 검사하지 않는다. 원본 API의 모든 입력에 대한 의무 위반이라고 단정하지 않는다. |
| `ValidatePattern("{a,[}")` | false | false | 별도 검사 결과다. Match에 앞서 검사를 삽입하거나 결과를 합성하지 않았다. |

이는 코드 후보를 추천할 때 “숫자 버전도 비교한다”, “패턴 오류도 검사한다” 같은 넓은 주장만으로 충분하지 않은 이유를 보여준다. **입력 범위와 실패 조건을 포함한 작은 주장**을 만들고 실제 검증을 연결해야 한다. 작은 모델의 의미 이해나 비용 절감은 아직 별도 검증 대상이다.

## 관측 수와 독립성

- Strict parse16, metadata 비교5, prerelease 비교9, 경로 매칭9, escape/문자 매칭8, 오류 매칭6, pattern 검사9 =62 operation 관측.
- 요청한 entry API90회: StrictNewVersion44·Compare14·Match23·ValidatePattern9. 성공 파싱의 getter/String/Original224회는 별도다. 원본 내부 helper 호출은 계수하지 않았다.
- 기대값 분류: 규범14건 중12일치/2불일치, 기술적 API46건 모두일치, 더 강한 정책2건 중1일치/1불일치. 알려진 라이브러리 오류17건도 정확히 관측했다. unknown/panic/unsupported0.
- 사전 입력과 후보/관측기 바이트를 확인한 Git blob28개. 공식 수집1회·공식 재시도0·exit0. `HEAD` 문자열을 거절한 plan 준비 시도는 관측0이며 공식 수집 재시도가 아니다.
- 서로 다른 원천 가족2개와 유한 속성 계약6개다.62개의 입력·alias·반전·같은 계약의 변형을 독립 최종 요청으로 세지 않는다. 새 모델 입력/caption 생성·순위 평가·fit·유료 호출·가중치·보호 최종 요청은0이다.

원본 MIT 소스15개99,035바이트와 full LICENSE2개를 보존했다. 원본 module 선언은 `metadata/upstream.go.mod.txt`에 같은 바이트로 두고 부모 `go.mod`/`go.sum`을 변경하지 않았다. 원본 코드가 이미 제공하는 동작을 수집한 것이며 새로 해결한 코딩 작업6개라고 세지 않는다. [원천 manifest](../../internal/publicbehavior/upstream-manifest.json), [사전 계획](PLAN-57.ko.md), [작성·검토 기록](oracle-review-57.json)을 함께 읽는다.

## 고정과 재현

| 근거 | 바이트·SHA256 |
|---|---|
| [입력](probes-57.json) |218,192 / `c4deb9d9e33dca8ff4498531d898ea82ee05559727cbc4f00d99455910722574` |
| [실행 계획](execution-plan-57.json) |5,472 / `dfc17915b37d4ce007f48a3a8d4e6fcc693f926f6a789562f6a8fcc68c347d7b` |
| [원본 관측](results-57.json) |262,559 / `850c60266eceb613565390430adca12cea9b1a56207b803b06cb681f563978c5` |
| 실제 Darwin/arm64 Go1.27.1 CGO0 binary, Git 제외 |5,096,818 / `09e5f6079b8a436c432d313a884376a5bde2267e7afdde88c5d422f1777cc5d7` |

source freeze `9e99914f6d1e35aa9d97413f80fc6e03768e97ce`, input freeze `1462c705087f24666049f8bd31d317490c05dbeb`. JSON3개 합계486,223바이트다.19개의 embedded source/upstream artifact와7개의 support를 관측 전에 확인했다. 수집 후 `publication_test.go`·`audit_test.go`는 사전 핀에 소급하지 않는다. 회귀 재생은 새 공식 관측·독립 표본이 아니다. [사용법](USAGE-57.ko.md)에서 플랫폼 경계도 확인한다.

Go scheduling1·heap soft limit256MiB는 설정이며 RSS/CPU/GPU/지연 실측이 아니다. 이번에 Laya나 GPU를 실행하지 않았고 생산 힌트 경로에 의존성·lock을 추가하지 않았다. 기존56a/d 측정 및 pending/unknown은 보존한다. [Laya 회고](../../docs/laya-source-retrospective-57.ko.md)의 캐시·점수 설명 개선은 제안이며 이번 실행 성과에 합치지 않는다.

다음은 다른 원천의 정확한 요청/caption·후보 관계를 확보하고, 모든 후보를 보존하는 비학습 순위 효용을 따로 고정·평가하는 것이다.240개 원천 확보, 최소15그룹/필요 효용5%, 도메인별 고유 보호 최종2400개 기준을 달성했다고 선언하지 않는다. **학습 적격false**를 유지하며 새 HF 가중치를 발행하지 않았다.
