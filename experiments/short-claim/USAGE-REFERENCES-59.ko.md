# 문장 참조 준비 도구 사용법

[English](USAGE-REFERENCES-59.en.md) · [공식 결과와 해석](RESULTS-REFERENCES-59.ko.md) · [사전 계획](PLAN-REFERENCES-59.ko.md)

[수집·실패 원장](collection-ledger-59.json) · [독립 참조 검토](reference-review-59.json)

`riido-captionref`는 유지보수자나 에이전트가 **설명을 검토할 원문과 근거를 찾는 Go 도구**다. 원래 JSON 위치·문장 해시·Go 리터럴 표현식 위치를 연결한다. Python이나 모델 가중치가 필요하지 않으며, 후보 코드·Laya·순위·학습을 실행하지 않는다. 빌드 성공이나 참조 생성 성공은 문장 의미·라이선스·모델 배포 자격의 승인이 아니다.

## 결과를 읽고 검토를 시작하기

일반 사용자는 [caption-coverage-59.json](caption-coverage-59.json)과 [결과 설명](RESULTS-REFERENCES-59.ko.md)을 읽으면 된다. 공식 생성을 다시 실행할 필요가 없다.

1. `generation_counters`에서 Bind1/1·AST3/3·부모72·설명216·계약18을 확인한다.
2. `references.parents`의 `request` 또는 후보의 `caption`에서 원본 파일·JSON pointer·해독한 UTF-8 바이트 수와 SHA를 찾는다. 문구와 순서를 바꾸지 않고 해당 원문을 읽는다.
3. 같은 원형의 `references.contracts`에서 역사적 입력 범위와 `literal_source_reference`를 읽는다. byte offset은0부터 시작하고 끝은 제외한다. 줄 번호는 원본 physical line이며 양끝을 포함한다.
4. [내용 검토 recipe](content-review-recipe-59.json)에 따라 구현 설명·요청 범위·관측 필드·부정/경계를 따로 검토하고, 문장과 원천 근거 위치·남은 불확실성을 기록한다. 현재 모든 검토는 pending이다.

허용 후보와 후보별 검사/실패 수는 저장된 역사적 정답이다. 참조 도구가 새로 실행한 결과가 아니다. unknown21과 unknown-only 그룹64는 그대로 유지하며, ID·그룹·정답·literal·검토 상태를 scorer 특징에 넣지 않는다. 잘못된 구현을 충실히 묘사한 caption도 정답 후보라는 뜻은 아니다.

## 코드 확인과 CI 재현의 범위

저장소 루트에서 Go1.27.1로 도구와 저장 참조의 회귀 검사를 실행할 수 있다.

```sh
GOTOOLCHAIN=go1.27.1 go test -race ./internal/captionref ./cmd/riido-captionref
```

단위 검사는 작은 소유 fixture로 AST·바이트 구간·핀·입출력 제한·출력 보존·실패 카운터를 확인한다. 공식 결과 뒤 추가한 [frozen 회귀 검사](../../cmd/riido-captionref/frozen_test.go)는 기존 입력에서 `references` 본문과 `generation_counters`를 정확하게 재생·비교하는 별도 검사다. 이때 저장 메타데이터 Bind도 수행하지만 원천 API·순위·모델은 호출하지 않는다. 점수나 부동소수점 허용 오차는 없다. 원래 Darwin envelope·소스·입력·계획·바이너리 핀은 별도로 확인하며, Linux/macOS CI 실행 파일이 원래 Darwin 파일과 같아야 하는 것은 아니다.

공식 결과 이후 로컬 frozen 회귀 검사와 전체 race/vet는 통과했다. [독립 참조 검토](reference-review-59.json)의 읽기 전용 검토자는 Generate·Bind·테스트를 실행하지 않았다. 반복 검사는 공식 시도·독립 요청·새 truth를 추가하지 않고 [수집 원장](collection-ledger-59.json)의 수집 횟수와 구분한다. 후속 회귀 파일을 사전 지원 파일9개에 소급하지 않는다. 최종 GitHub CI는 이 문서 작성 시점에 대기 중이며 실제 해당 실행에서 따로 확인해야 한다.

## 이미 끝난 공식 수집 절차

아래는 **과거 공식1회를 만든 절차의 기록**이다. 기존 출력 경로를 재사용하거나 새 경로만 골라 공식59를 다시 실행하라는 안내가 아니다. 원래 소스 commit은 `75a9776230f7eaef3294247cc10bb7f13342f102`, 입력 commit은 `69e9ddc51e218da029572e2bb463500dc36cd143`이다.

```sh
GOTOOLCHAIN=go1.27.1 CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o .cache/captionref59-driver ./cmd/riido-captionref
.cache/captionref59-driver prepare --repo . --source-commit 75a9776230f7eaef3294247cc10bb7f13342f102 --out .cache/captionref59-plan.json
# 생성한 plan 바이트를 experiments/short-claim/execution-plan-59.json에 그대로 고정하고 input commit을 만든 뒤:
.cache/captionref59-driver references --repo . --input-commit 69e9ddc51e218da029572e2bb463500dc36cd143 --out .cache/captionref59-official/results.json
```

계획 준비부터 공식 생성까지 같은 Go1.27.1/CGO0/trimpath/Darwin arm64 실행 파일을 사용했다. 기존 계획은 실제 실행 파일 SHA까지 고정하므로 다른 빌드나 플랫폼의 파일로 원래 공식 CLI 실행을 재현할 수 없다. 실행 파일은 로컬 private 경로에 두며 Git에 올리지 않는다. 새 실험은 별도 version·사전 계획·봉인·시도 원장으로 준비해야 한다.

`prepare`는 Bind/Generate0이며 소스5·지원9·입력6의 Git blob20개를 검사한다. `references`는 소스5·지원9·입력6·계획1의21개를 먼저 검사하고, 새 디렉터리와 배타적인 `results.json`을 예약한 뒤 Bind와 AST 참조 생성을 한다. 입력/결과는 각각1MiB 이하, 바이너리는64MiB 이하, Git 검사는 파일별5초다. 출력 경로 재사용 방지는 컴퓨터 전체의 공식 실행 횟수 제한을 대신하지 않는다.

실패하면 `state=incomplete`, 고정 `failure_code/failure_cause`와 단계별 카운터를 남긴다. 출력 인코딩·쓰기가 실패해 파일에 온전히 남지 못하면 같은 카운터가 오류 진단에 남는다. 실패를 삭제하거나 자동 재시도하여 실행0처럼 만들지 않는다. [prototype ledger](prototype-ledger-59.json)의 준비2회·Bind2회와 첫 실패도 공식1회와 별도로 보존한다.

## 다음 준비 자료

[원천 전이 설명](SOURCE-TRANSFER-59.ko.md)과 [기계가 읽는 원천 연결](source-transfer-59.json)은 semver/glob의 기존 원문과 저장된57 관측을 연결한 두 scoped proposal이다. 새 요청·caption·정답·원천 API 호출은0이다. [역할 recipe](role-recipe-59.json)는 실제 배정 전이며 train9/validation3/calibration3 하한과 transfer0 허용을 유지한다. 내용 검토·관계 그래프·coverage·membership/seed 봉인이 다음 작업이다.

`training_ready=false`, `no_roles_plan`, `synthetic_single_pipeline`을 함께 읽는다. 이번 자료로 모델 효용·토큰/요금 절감·실측 메모리/CPU/GPU나 한국어 입력 성능을 주장하지 않는다. 서로 다른 보호 최종 요청2,400개는 별도 목표이며 모든 개발 fit의 선행 최소가 아니다. 기존 Laya CI는 별도 기록이고 이번 도구 자체의 모델 호출은0이다.
