# 59 결과: 문장과 리터럴 소스의 참조 준비 완료

[English](RESULTS-REFERENCES-59.en.md) · [사용법](USAGE-REFERENCES-59.ko.md) · [사전 계획](PLAN-REFERENCES-59.ko.md) · [공식 원본](caption-coverage-59.json)

[수집·실패 원장](collection-ledger-59.json) · [독립 참조 검토](reference-review-59.json)

**기존 요청과 후보 설명을 검토할 정확한 위치를 한 번의 공식 메타데이터 생성으로 연결했다. 설명이 맞다는 판정과 학습 준비는 아직 대기 중이다.** 예를 들어 검토자는 후보 설명의 원래 JSON 위치·문장 해시와 해당 계약의 Go 리터럴 표현식 위치를 함께 찾아볼 수 있다. 생성기는 후보 코드나 모델을 실행하지 않았다.

공식 시도는 **1회, 재시도 0회, exit 0**이다. 상태는 `references_generated_content_review_pending`이며 `preparation_only=true`, `content_review=pending`, `literal_payload_reified=false`, `training_ready=false`다. 공개 `caption-coverage-59.json`은 공식 원본 envelope를 바꾸지 않는 복사본이다. 파일 이름의 coverage는 참조 준비 범위를 뜻하며 의미 coverage 승인률을 뜻하지 않는다.

## 무엇을 연결했나

| 항목 | 공식 결과 |
|---|---:|
| 저장 메타데이터 Bind 시도 / 완료 | 1 / 1 |
| Go AST 파싱 시도 / 완료 | 3 / 3 |
| 기존 부모 요청 | 72 |
| 기존 후보 설명 위치 | 216 |
| 문장 참조: 요청 72 + 설명 216 | 288 |
| 계약의 raw 표현식 참조 | 18 |
| 생성 전 검증한 Git blob | 21 |

AST 파싱은 소스의 문법 구조를 읽는 작업이다. 리터럴 표현식을 실행하거나 Input/Want 값을 재구성하지 않으며, 후보별 Got이나 실패 위치를 새로 관측하지 않는다. 기존 역사적 정답·후보별 검사/실패 수·허용 후보 인덱스는 저장 기록에 연결한 metadata로 보존한다. 새로운 정답으로 만들거나 scorer 특징에 넣지 않는다.

원문 문장 SHA는 JSON을 해독한 정확한 UTF-8 바이트를 고정한다. raw 표현식 SHA는 원본 Go 파일의 `[start_byte,end_byte)` 구간을 고정한다. 이는 기존 직렬화된 리터럴 표 SHA, 형식화된 후보 선언 SHA, helper/type/sentinel을 포함하는 bundle SHA와 서로 다른 근거다. 해시 일치만으로 전체 source closure나 문장 의미를 승인하지 않는다.

## 기존 분모와 검토 대기 상태

| 보존한 자료 | 수 |
|---|---:|
| 답 있음 / 답 없음 / unknown | 34 / 17 / 21 |
| 전체 연결 그룹 / 알려진 부모가 있는 그룹 | 17 / 16 |
| 후보 2개 / 3개 / 4개인 부모 | 12 / 48 / 12 |
| 원형 | 18 |

unknown-only 그룹 ID64도 전체 그래프에 남으며, 그 부모 4개를 오답이나 답 없음으로 바꾸지 않았다. 중복 후보 위치와 원래 후보 순서도 보존했다. 72부모·216설명·288문장 참조·18표현식은 서로 다른 단위이고, 새 독립 요청이나 최종 표본 수로 합산하지 않는다.

| 아직 끝나지 않은 검토 | pending |
|---|---:|
| 요청의 계약 검토 | 72 |
| 후보 source closure | 216 |
| 후보 설명의 source fidelity | 216 |
| 후보 설명의 request-contract coverage | 216 |
| 계약 관측 필드 검토 | 18 |

잘못 동작하는 구현을 문장이 충실하게 설명할 수도 있다. 따라서 설명 충실도와 요청에 대한 정답 여부를 따로 검토해야 한다. [내용 검토 recipe](content-review-recipe-59.json)는 문장·계약·관측 필드·부정/경계와 그 근거를 읽는 다음 단계다. 이번에 검토 완료 라벨을 만들지 않았다.

[독립 참조 검토](reference-review-59.json)는 문장288개의 JSON 위치·총28,787 UTF-8바이트, 표현식18개의 원문 구간·물리적 줄, 역사적 vector 표현식148개와 그룹17개·관계 edge124개의 보존을 확인했다. 위 검토 항목 합계738개는 모두 pending으로 남는다. reviewer의 Generate·Bind·원천 API·테스트 실행은0이며 이 통과는 참조 무결성만 뜻한다.

## 봉인과 실행 이력

| 원본 provenance | 값 |
|---|---|
| Source commit | `75a9776230f7eaef3294247cc10bb7f13342f102` |
| Input commit | `69e9ddc51e218da029572e2bb463500dc36cd143` |
| 실행 계획 | 4,681 B; SHA256 `a321d625e986ec7e621b827a0616b7492291f624bd395bc32b424a0ad1178819` |
| 공식 실행 파일 | 5,202,770 B; SHA256 `03efc5808316128e0c336887907ee61d4aa9bab59cc764b8401e53d4786a21ec` |
| 원본 결과 | 360,591 B; SHA256 `7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29` |

원본 실행 환경은 Go1.27.1 / Darwin arm64, CGO0, trimpath 빌드다. 생성 전에 컴파일된 Go 소스 5개·지원 파일 9개·입력 6개·실행 계획 1개의 Git 바이트를 확인했다. 공식 계획 준비는 1회이며 Bind·Generate는 0이다. 출력은 새 디렉터리와 배타적인 `results.json`으로 예약했다. 이 장치는 경로 재사용을 막으며 전체 컴퓨터에서 한 번만 실행되도록 만드는 전역 lock은 아니다.

[prototype ledger](prototype-ledger-59.json)는 별도 준비 실행 2회·Bind 2회를 보존한다. 첫 실행은 수동 원형 이름 세 개가 역사적 이름과 달라 panic으로 종료했고 출력은 0이었다. 당시 실행 파일 SHA는 기록되지 않았다. 원래 이름을 사용한 두 번째 실행은 private 참조를 만들었으며 의미 검토는 pending이었다. 이 둘을 공식 시도나 새 정답으로 더하지 않는다.

실행 전 전체 race/vet와 공개 자료 검사도 통과했다. CLI의 최초 `GOTOOLCHAIN=local` race/vet 확인 두 건은 기본 Go1.27.0의 버전 guard에서 테스트 실행 전에 중단됐고, 설치된 Go1.27.1로 다시 검사해 통과했다. 결과 이후에는 [frozen 회귀 검사](../../cmd/riido-captionref/frozen_test.go)와 로컬 전체 race/vet가 통과했다. 저장 메타데이터 회귀 반복은 공식 시도·새 독립 자료에 더하지 않는다. [수집 원장](collection-ledger-59.json)은 prototype+공식 수집 Bind3회와 후속 회귀를 구분한다. 후속 `frozen_test.go`는 원래 지원 파일9개에 소급하지 않으며 이 문서 작성 시점에는 최종 GitHub CI 완료를 선언하지 않는다.

## 다음에 사용할 방법과 남은 경계

[원천 전이 준비](SOURCE-TRANSFER-59.ko.md)는 semver strict parsing과 glob slash 경계의 두 scoped proposal이다. 기존 원천 artifact 15개와 저장된57의 25개 관측·287개 Want/Got 필드 참조를 연결했으며, 새 다운로드·API 관측·부모·caption·truth·허용 집합은 0이다. 이 작성 원천과 프로젝트의 AI 지원 합성 caption 작성 과정을 구분한다. 외부 코드나 다른 reviewer만으로 `synthetic_single_pipeline`을 해소하지 않는다.

[역할 recipe](role-recipe-59.json)는 `unassigned`다. 미래에 연결 요소 전체를 배정하는 규칙을 기록했지만 membership/seed 봉인·coverage·실제 배정은 아직 없다. 기존 train9 / validation3 / calibration3 하한을 유지하며 transfer0을 허용한다. 이미 관측한58의 점수로 그룹이나 seed를 고르지 않고, 역할 이름으로 기존 자료를 미관측/블라인드 검증으로 바꾸지 않는다.

`no_roles_plan`·`synthetic_single_pipeline`을 유지한다. 새 라벨·독립 요청·역할·순위·원천 API·모델/유료 호출·fit·가중치·보호 최종 접근·운영 활성화는 모두 0이다. 이번 도구 자체는 Laya를 실행하지 않으며 기존 Laya CI의 기록을 변경하지 않는다. Go scheduling1·heap soft limit256MiB는 설정이며 RSS/CPU/GPU·지연 실측이나 LLM 토큰·요금 절감 근거가 아니다.

도메인별 서로 다른 보호 최종 요청 최소2,400개는 별도 일반화 목표다. 모든 개발 fit의 선행 최소 수가 아니며 이번 참조 생성으로 채워지지도 않는다. 다음 내용 검토와 별도 학습 계획은 기존 자동 검사·CI 정책 안에서 진행하며 새 사람 승인 절차를 만들지 않는다.
