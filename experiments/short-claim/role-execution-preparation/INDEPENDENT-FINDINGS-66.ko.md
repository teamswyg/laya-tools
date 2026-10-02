# 전체 그룹 역할 실행66 · 읽기 전용 검토

현재 동결한 코드·계획·바이너리의 정합성을 검토했고, **이 범위에서 원본 실행을 막는 구체적 코드/핀 오류를 발견하지 않았다**. 이것은 실행·학습 준비 승인이 아니다. controller의 PR CI 성공·병합·Git source/input tree 검증은 별도 책임으로 남는다. 이 검토에서 원본 prepare/execute, 역할 함수와 fit을 호출하지 않았다.

검토자는66 CLI/loader/동결 계획의 작성자가 아니다. 그러나62 public roleplan 모듈 작성자이며 기존 개발 자료를 보았다. 따라서 CLI의 별도 코드 검토이지만 nonblind이며 독립 원천·독립 compiler 신뢰·새로운 의미 판단은 아니다.

## 고정 자료

|자료|bytes / SHA256|
|---|---|
|동결 계획66 v2|9,559 / `f6065a1285592dfa0617415668d07e332cb10a663ec8c78f13c17c2fafa1da53`|
|공개 main.go|11,948 / `b13b66b422db65982a3cc8de18f8a0107b5ef0bc8bab5fab004d778a242a97e3`|
|공개 loader.go|22,518 / `be769eb783cf09744820ee556c8bbb57adae252933c998ae423e75bf695770a7`|
|공개 main_test.go|13,663 / `79701206640146487b79253a23907c5e3c5c6ca08e5e14d21f7a812115df0bab`|
|원래 role.go|16,123 / `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`|
|추가 source-identity helper|707 / `f1fb50eb693baa06cd1e9df7ad57275ee557a6bcc17439befbaf9617a6082224`|
|미실행 로컬 바이너리|4,438,626 / `32aa2b34770d3b1ce2aab5e6c87429cdef7af3a6451d5bad9faba644b059baa4`|

별도 Go 검사기는 위 동결 계획의 canonical bytes와20개 pin 확인을 수행했다. membership65, 원본 입력8개, support3개, binding evidence2개, implementation6개다. 같은 파일이 다른 역할로 핀된 경우 그 핀마다 확인했다. copied toy fixture용 source5개와 바이너리도 읽었다. 직접 bounded read27회·payload8,908,289bytes는 **이 검증 도구의 읽기 집계**이며 CLI 성능/메모리 측정이 아니다. buildInfo 읽기1회와 directory metadata 검사2회는 별도다.

membership metadata는 전체17그룹/known-containing16그룹,72부모/216후보, known34/no_answer17/unknown21,453members/1,922relationships를 유지했다. 그룹8과64를 포함한 원래17그룹 목록·부모 index의 완전한 대응·known/no_answer가 같은16그룹을 덮는 사실을 metadata로 확인했다. 원래 membership digest나 역할 order를 계산하지 않았다.

## 실행 경계와 보존

`validatePlan`은 seed ASCII1729/hex31373239, 기존3:1:1 encoding/algorithm/domain과9/3/3 하한, unknown 보존, 같은 두 stored-truth coverage 집합, one-attempt/no-retry, fit/model/paid/final0 및 readiness/features/masks=false를 고정한다. SHA·canonical JSON·엄격 shape를 검증하고 execute는 unfrozen 계획을 input/role dispatch 전에 거절한다. 저장 truth coverage는 미래 loss mask의 적격성이나 의미 준비를 승인하지 않는다.

`loadBound`는 membership65 exact bytes를 먼저 저장하고 source/evidence/support/implementation pins를 확인한 뒤 **처음 해시한 membership bytes**를 decode한다. implementation loop의 `b`는 내부 scope에 있어 input bytes를 덮어쓰지 않는다. source/body 입력·module 경로 결속은 Go import `github.com/teamswyg/laya-tools/internal/roleplan`으로 연결된다. 기존 role.go는 변경하지 않았고 helper만 추가됐다.

바이너리의 실제 SHA와 buildInfo를 읽어 Go1.27.1, main module `github.com/teamswyg/laya-tools`, command path, CGO0/trimpath, darwin/arm64와 VCS settings 부재를 확인했다. external dependency module 수는0이다. 바이너리 안에 exact CLI main/loader/role source bytes3개가 있음을 확인했다. 이 source witness는 코드와 metadata의 대응을 돕지만 embed text가 실제 명령 실행을 전부 증명하거나 compiler/toolchain 신뢰를 독립적으로 증명하지는 않는다.

`main.go:253` 이후 execute 경로는 하나의 caller-fixed ledger를 O_EXCL로 예약·write·Sync하고 새 출력 디렉터리/파일을 확보한 다음, VerifyMembershipSet1회와 최대 Assign1회로 이어진다. 같은 장부/출력은 재사용하지 않는다. membership refusal은 assign_attempts0, assignment refusal은1을 보존하며 fallback·새 seed·그룹 분할·retry를 호출하지 않는다. 실패 뒤 장부 파일이 남으면 consumed 상태다. 새 장부를 고르는 별도 호출은 controller 밖에서 막는 host-global 장치가 아니다.

post-dispatch ledger/output 저장 실패는 고정 오류와 실제 membership/assign prefix counters를 stderr에 내며 성공으로 처리하지 않는다. ledger 업데이트가 truncate 뒤 실패하면 JSON이 불완전할 수 있다. 출력 write/Sync 실패나 프로세스 중단도 완료 report를 보장하지 않으므로 controller는 실패 상태를 성공으로 대체하지 않아야 한다. 이 검토는 전원 장애·filesystem crash 복구를 실험하지 않았다.

## 상한과 현재 한계

입력 regular file은8MiB, plan/result는1MiB, 실행 binary 읽기는64MiB 상한이다. JSON은 depth128·object keys128·key512bytes·array65536와 duplicate decoded key/unknown-field/canonical 검사를 적용한다. pins는 bytes/SHA가 맞는 같은 read payload를 사용한다. GOMAXPROCS1과256MiB Go heap soft limit은 RSS hard cap이나 성능 결과가 아니다. readBounded는 limit+1 sentinel read를 사용하며 총 file payload 합계/프로세스 전체 메모리의 별도 hard budget을 주장하지 않는다.

일반 file reader는 path symlink의 regular target을 따른다. root/parent directory의 악의적인 교체를 막는 filesystem confinement 도구가 아니며, 호출자가 고정·관리하는 source/input root가 필요하다는 기존 controller 한계가 있다. read pin은 source 변조를 검출하지만 외부 filesystem나 compiler의 신뢰를 증명하지 않는다.

source allowlist는 정확한 알려진 Go 파일과 test 예외를 검사한다. **비차단 설명 한계**로, `checkClosedSourceRoot`는 `.go` 파일만 검사하므로 `.s`/`.syso`나 모든 build-environment 입력을 거절하는 일반 source-closure 도구는 아니다. 현재 두 source 디렉터리는 예상한 `.go` 파일만 존재했고 동결 binary/pins가 맞았다. README의 “추가 runtime 파일 거절”을 Go source filename guard 범위로 읽어야 한다. 새 조건/코드 변경을 요구하거나 현재 실행 blocker로 승격하지 않는다.

source closure/text 불확실성, 빠진 외부 관계, 한 synthetic caption pipeline, 원천 다양성 uncleared와 training_ready=false는 유지한다.66 역할은 stored truth 기준의 development 배정일 뿐이며,67 mask 이후 역할별 적격 범위와 same-scope 효용/headroom은 별도 고정 단계다. 새 하한·매 실험2400/fresh15·모든 prototype을 모든 role에 요구하는 조건을 추가하지 않았다.

## 검사와 장부

단독 metadata verifier1회 성공/실패0. 원본 자료의 텍스트·JSON/bytes·source와 바이너리 metadata만 읽었다. 원본 Assign/OrderDigest/AllocateCounts/MembershipDigest/VerifyMembershipSet, 원본 CLI, model/fit/labels/roles 실행은0이다.

필요한 오류 경로4개만 exact source의 private copy에서 합성 시험했다. 짧거나 실패한 write의 고정 코드/partial counters, 저장 실패 후 consumed ledger 보존, empty/oversize/directory 거절과 regular symlink target 한계, toy plan-SHA 실패의 인수 비노출/예약 부작용0이다. toy CLI의 early-error `run` 호출1회가 있으며 원본 input/source/binary dispatch로 이어지지 않았다. 기존 CLI 작성자의10개 테스트는 여기서 다시 실행하지 않았다.

합성 race1회:4 top-level/4 pass events/실패0, package1.603초. vet1회 성공, 새 audit source/test만 gofmt1회. 테스트 시간은 성능 benchmark가 아니다. 원본 CLI 작성과 다른 검토자가 작성한 오류 테스트지만 blind/source-origin 평가는 아니다.

읽기 명령22회 중 탐색1회가 실패했다. 필요 경로를 받은 뒤 두 private 디렉터리를 추측해 조회했고 둘이 없었다. 이는 원래 필수 artifact의 결손이 아니며 실패 장부에 남겼다. 긴 병렬 출력 한 번이 잘려 필요한 코드 범위를 재읽었고, 파일 끝 이후의 빈 출력은 결손으로 해석하지 않았다. 모든 실패 출력은 private logs에 별도 기억했고 기존 frozen 파일을 수정하지 않았다.

공유 edit/Git/외부 게시, 원본 role/order/metadata CLI 실행·features·새 라벨·fit·model/paid·protected-final 읽기는 모두0이다. 일반 AI-assisted 협업 비용은 측정하지 않았다. 재현 가능한 안전 aggregate는 `MECHANICS-66.json`, 실행/검사 counts는 `LEDGER-66.json`에 기록한다.
