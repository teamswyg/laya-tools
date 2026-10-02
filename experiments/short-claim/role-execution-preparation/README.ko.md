# 전체 그룹 역할 실행 도구66 · 준비 단계

`riido-roleplan`은 같은 원천을 공유하는 요청·후보·helper를 한 덩어리로 유지해 train/validation/calibration 역할을 배정하는 유지보수 Go 도구다. 모델이 외운 유사 코드를 검증 자료에 다시 넣는 일을 막기 위해 만들었다. 일반 `riidolaya` 사용이나 Codex 연결에는 필요하지 않다. 이번 공개 단계에서는 **실제 원본 역할 배정0·학습0**이며, 초안과 자동 CI 뒤 한 번 실행할 고정 계획을 구분한다.

현재 [v2 계획](execution-plan-66.v2.draft.json)은 seed ASCII `1729`, 전체17그룹·저장 정답을 가진16그룹·72요청·216후보를 고정한다. unknown21은 정답으로 바꾸지 않는다. unknown-only 그룹64는 train provenance로만 남는다. known 그룹은 3:1:1 largest-remainder 방식으로 나누며 기존9/3/3 하한을 검사한다. 두 stored-truth coverage 집합은 같은16그룹이므로 새 조건이 아니다. 실제 loss-mask가 남기는 그룹 수는 별도 검사해야 한다.

다음 명령은 자동 CI를 통과한 소스로 Go1.27.1 / CGO0 / trimpath binary를 만든 후 사용하는 형태다. 자리표시자는 실제 경로와 SHA로 바꾼다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o riido-roleplan ./cmd/riido-roleplan
./riido-roleplan --stage prepare \
  --plan execution-plan-66.v2.draft.json --plan-sha256 PLAN_SHA256 \
  --input experiments/short-claim/whole-group-preparation/membership-proposal-65.json \
  --repo-root . --source-root . --out NEW_PREPARATION_DIRECTORY
```

prepare는 원본 역할 함수를 호출하지 않는다. execute는 exact source/input/build/binary를 결속한 **새 frozen 계획**과 하나의 고정 `--ledger`를 요구한다. [첫 frozen 계획](execution-plan-66.v2.frozen-1.json)은 Go1.27.1 / darwin-arm64 / CGO0 / trimpath / buildvcs=false로 만든 로컬 binary SHA를 고정한 새 파일이다. 아직 실행하지 않았으며 controller는 이 PR의 quality 통과·병합과 Git source/input tree 일치를 확인한 뒤 한 번만 실행한다. 다른 플랫폼의 binary를 같은 실행으로 사용하지 않는다. 실행 false인 초안은 그대로 보존한다. 실행 전에 장부·결과를 새로 예약하며, 거절 결과와 카운터를 보존한다. seed 검색·그룹 분할·mask 변경·자동 재시도는 하지 않는다. 사람의 재승인 절차를 추가한 것이 아니라 이미 허용된 작업의 자동 검증 단계를 구분한 것이다. caller가 다른 장부를 지정하면 제한을 우회할 수 있으므로 controller는 하나의 공식 실행 경로를 유지해야 한다.

선언한 Go source closure는 CLI main/loader, role.go와 additive source-identity helper 네 파일 및 go.mod/go.sum다. compiled text와 실제 source SHA를 비교하지만 compiler 신뢰 자체를 증명하지는 않는다. Git commit/blob 동결은 controller 책임이다. exact known tests만 예외로 허용하고 추가 Go 실행 소스·미등록 Go 테스트·Go 파일 symlink를 거절한다. 이 이름 검사는 `.s`/`.syso`나 전체 빌드 환경까지 검증하는 hermetic build 증명이 아니다. 이번 동결 디렉터리는 예상 Go 소스만 포함하며 실제 binary SHA를 별도로 결속했다. 파일당 입력8MiB·출력1MiB의 도구 한도와256MiB Go heap soft limit이 있다. 이 숫자는 OS RSS hard cap이나 성능 측정이 아니다.

[v2 통합 안내](HANDOFF-66.v2.ko.md)와 [장부](PREPARATION-LEDGER-66.v2.json)에 합성10개 named tests의 private/shared race 및 vet 통과를 기록했다. root는 원본 역할 호출 전에 source·plan·결속·예약·고정 오류를 읽어 검토했다. 원래 v1 allowlist가 기존 archive tests를 제외하던 문제는 실행 전 code reading으로 발견했으며 [v1 장부](PREPARATION-LEDGER-66.v1.json), 원래 source `.go.txt`와 두 layout 초안은 역사로 보존한다. 이전 setup 실패도 v1 장부에 남는다. 합성 테스트 시간은 benchmark가 아니다.

[별도 검토자](INDEPENDENT-FINDINGS-66.ko.md)는 CLI 작성자가 아니며62 모듈 작성 노출이 있는 비블라인드 reader다. 읽기 전용 verifier1회에서27파일·8,908,289bytes를 확인하고, 별도 toy 오류 검사4개의 race와 vet를 통과했다. 현재 동결 source/plan/binary의 실행 blocker는 발견하지 않았다. 탐색 읽기 실패1회와 guard/compiler/장부별 한계는 [독립 장부](INDEPENDENT-LEDGER-66.json)에 보존한다. 원본 역할 함수·CLI 실행0이다.

이 도구는 opaque digest만으로 전체 외부 관계가 없다고 증명하거나 source/text uncertainty·다른 작성 원천·학습 준비를 승인하지 않는다. 현 stored graph의 결속 범위만 검사한다. protected-final2400 목표는 별도이며 모든 개발 fit의 최소 표본 수가 아니다. source/group/role/정답/검토 정보는 learned features가 아니다. 별도 모델 프로세스0은 일반 AI-assisted 협업 비용0을 뜻하지 않으며 그 비용은 측정하지 않았다.
