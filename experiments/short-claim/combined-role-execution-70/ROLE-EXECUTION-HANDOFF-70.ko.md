# 전체 76개 역할 실행기 준비

`main.go`, `loader.go`, `runner.go`와 합성 테스트를 준비했습니다. 실제 공개 역할 모듈에 연결해 빌드했지만 **원래 전체 입력에 대한 prepare/execute와 모듈 함수 호출은 0회**입니다. 현재 사용할 초안은 `execution-plan-70.draft-2.json`이며 `execution_authorized=false`입니다. 실제 배정 결과가 아니고 역할 실행의 자격·다양성·학습 준비를 승인한 자료도 아닙니다.

## 정확한 파일과 핀

| 파일 | bytes | SHA256 |
|---|---:|---|
| `execution-plan-70.draft-2.json` | 6,964 | `6988545f37f3fccc430bbc6d78e53a8946b3a25f06925dff21fec5b96c713be6` |
| `preparation-receipt-70.2.json` | 3,024 | `4ba3bc820f92a77e716ddeeaa71bb28834f92229513ffeef37bd239ea6ed8ed5` |
| `combined-role-exec` | 4,563,106 | `f4912efce64a801fe0bc7b3acda74593316980da5dab6c20fbe2507ba17381af` |
| `main.go` | 8,566 | `e7ffba1ce0df1c429d54ec1b94eb0c122646826ae2b7a4ff3e247d47ae21a026` |
| `loader.go` | 26,648 | `407891f136c1a769da89f0576c99370518f3b5f15436c3257d685c86c26e008d` |
| `runner.go` | 13,589 | `66d7d7744cf15c8f15529812e392a3465545e61f8e4f25423322e0d6ffe9dc9d` |

원래 proposal70은 2,694,968 bytes / SHA `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21`입니다. `input/combined70.json`에 같은 바이트를 복사했습니다. `input/future70draft.json`과 `input/evidence/`의 12개 원래 파일까지 전체 14개 input을 같은 방식으로 복사했습니다. 복사 장부에는 상대 경로·크기·SHA만 있으며 원래 private path는 공개용 자료에 없습니다. 원래 input은 변경하지 않았습니다. worker 단계는 이 복사본도 아직 읽어 검증하지 않았습니다.

초안1의 원래 binary/source/plan/receipt는 `frozen-draft-1/`에 남겼습니다. 초안2는 원래56b saved truth와67 mask를 별도로 결속하는 검사를 추가했습니다. 이전 loader 바이트의 SHA도 원래 초안1 핀과 맞는지 대조했습니다. 두 초안 모두 실행하지 않았으므로 원래 실패·배정·label 결과를 바꾸는 수정은 없습니다.

## 실행 전 결속과 실행 순서

CLI는 `--stage prepare|execute`, `--plan`, `--plan-sha256`, `--source-root`, `--module-root`, `--input-root`, `--run-root`만 받습니다. 인자로 actor 특징이나 임의 source callback을 받지 않습니다. schema에 없는 plan field와 알 수 없는 인자는 거절하며 진단은 고정 오류명과 숫자 counters만 출력합니다.

먼저 외부에서 정한 plan SHA와 canonical JSON, 실제 Go1.27.1/CGO0/trimpath binary SHA, worker의 컴파일된 세 원문과 디스크 source manifest, 공개 role module의 고정 원문을 결속합니다. worker/module 경로는 runtime 옵션이며 input/source 참조는 상대 경로입니다. 14개 metadata 파일은 **같은 읽은 바이트를 해시 검증한 뒤 decode**합니다. per-file 4 MiB / total 16 MiB는 maintainer 한도입니다. 원래 runtime request/caption의 512 bytes / 32 normalized words 기준을 바꾸거나 이번 도구에서 Normalize/Validate를 대신 실행하지 않습니다.

원래17 group의 전체 JSON값·raw subtree SHA는65와 대조하고 새 두 그룹은72/74·원래 append 순서 그대로 유지합니다. 각 group의 원래 parent ID/index와 새72..75 index 집합을 대조하며, 모든76 parent의 group이 유일하게 대응하는지와 저장 truth에 따른 known flag도 검사합니다. 원래72 request/candidate 순서·원문·code/bundle·allowed·null·mask를56/56b/67/66에 직접 결속합니다. 66 role 배열의 원래 group 순서는 보존하며 `original_parent_index→원래 row index` 배열 lookup만 사용합니다. 새 네 요청은 원래68v4 contract와 finite QA의 순서·caption·allowed·proposed label·mask·null weight를 유지합니다. source68의 Go/module/LICENSE 15개 assets도 disk SHA로 대조하며 해당 패키지를 import하거나 실행하지 않습니다.

`prepare`는 metadata-only 결과를 새 폴더에 작성하고 공개 역할 함수를 호출하지 않습니다. 현재 prepare도 실제로 실행하지 않았습니다. `execute`는 새 frozen plan이 명시적으로 실행 상태여야 진행합니다. `role-attempt-70.json`을 O_EXCL로 예약·Sync하고 `role-attempt-70/`의 receipt/result 파일도 O_EXCL로 예약합니다. 한 실행 root의 ledger는 재설정·환급하지 않으며 host 전체의 권한 저장소라는 뜻은 아닙니다.

그 다음 공개 모듈의 compiled-source identity 확인, `VerifyMembershipSet` **1회**, `Assign(seed ASCII1729)` **1회** 순서로만 진행합니다. 각 호출 직전에 result와 ledger에 시도 counter를 Sync합니다. 원천 함수/caption 재관측·Features·Project·Baselines·모델·fit은 호출하지 않습니다. module 오류나 panic은 고정 거절로 남기며 성공 label로 바꾸지 않습니다. 배정 거절의 module 내부 counters와 알려진 고정 failure code도 보존합니다. write/sync 실패의 진단에도 identity/verification/assign 시작·완료/parent-role 수를 남깁니다.

## 조건과 한계

고정62 source SHA는 `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`입니다. helper·기존 test·archive test·원래 module go.mod/go.sum과 함께 pin합니다. 공개 모듈의 3:1:1 largest-remainder, 기존9/3/3 floor, stored-truth answerable18/no_answer16의 여섯 predeclared coverage만 사용합니다. 실제 역할을 아직 계산하지 않았으므로 `11/4/3` 같은 값을 실제로 주장하지 않습니다. 기존72 역할은 과거 provenance일 뿐 전체76에 복사하지 않습니다.

성공한 역할은 부모 index 순서로 새76 parent-role mappings을 반환합니다. role별 부모/후보/known/no_answer/unknown/null과 positive/negative/masked 수치를 함께 기록합니다. mask-qualified 그룹과 기존9/3/3의 비교는 설명용 별도 항목이고 positive-bearing 수로 새 gate를 만들지 않습니다. hard negative와 no_answer 후보는 보존합니다. coverage 실패 시 역할을 원자적으로 거절하고 seed·membership·group·기준을 고치거나 유리한 seed를 다시 찾지 않습니다. source68 label은 제한된 제안이고 실제 학습 가중치는 계속 비어 있습니다.

CPU1과 Go heap soft limit256 MiB는 설정이며 RSS 측정이나 hard RSS cap이 아닙니다. worker는300초 cooperative checkpoint timeout을 사용하지만 진행 중인 동기 module 함수를 강제 중단하지는 못합니다. **300초 hard wall은 부모 OS controller가 실행을 감독해야 합니다.** 강제 종료 시 마지막 durable checkpoint가 시작된 호출 수를 보존합니다. result payload 한도는64 MiB입니다. 새 성능·절감·메모리 개선 측정은 없습니다.

이 private build는 외부 public module을 local replace로 연결했습니다. 준비용 `go.mod`와 seal helper에는 private build 경로가 있으므로 그대로 공개하지 않습니다. 실제 실행 시 상대 참조와 source-root 옵션을 쓰며, 추후 shared public port에서는 module/build identity와 source manifest를 실제 저장소 배치에 맞춰 **새로 고정**해야 합니다. 기존14 input/proposal SHA·역할 recipe·seed·coverage는 바꾸지 않습니다. embed와 SHA는 source 결속일 뿐 trusted compiler의 독립적 증명이 아닙니다.

## 검증과 부모의 다음 한 단계

합성 fake callbacks만 사용한 race 검사4회에서 top-level named-test 실행13개와 그 안의 subtest4개를 통과했습니다. vet3회, binary build3회, source/input seal2회도 통과했습니다. 원래76 metadata 전체 loader, prepare/execute worker, 실제 모듈 provenance/verification/Assign, source API, Features/Project/fit는0회입니다. 실제 코퍼스를 합성 테스트에 넣거나 actual role 결과를 보고 테스트를 맞추지 않았습니다.

부모는 최신 draft2/source/loader의 독립 QA와 Go source/binary/input/실행 계획 동결을 마친 뒤 실행 상태의 계획을 작성할 수 있습니다. 그때 새로운 전체76 Assign은 한 번이고 재시도0입니다. 먼저 전체 배정·checkpoint/거절/coverage를 확인한 뒤 별도76 input 검증·Project와 실제 fit 준비로 진행합니다. source와 작성 흐름 다양성은 아직 cleared가 아니며 ≥2,400 protected-final/domain은 별도 목표입니다. 새로운 역할별 positive minimum이나 매-fit 2,400개 조건을 추가하지 않았습니다. 공유 저장소 수정·게시·모델 호출은0입니다.
