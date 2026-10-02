# Native2 단회 실행 제어 코드 독립 검토

이 문서는 원본 Go 함수 23개 관측을 나중에 **한 번만** 실행하기 위한 바깥 제어 코드의 정적 검토입니다. 검토 대상 v4에서 발견된 두 구현 문제는 수정됐고, 현재 남은 구체적인 소스 차단 사유는 0개입니다. 실제 원본 실행 가능 여부는 Root가 별도로 봉인하는 권리·소스 선택·CI·자원·실행 계획에 달려 있습니다. 이 보고서는 그 승인, 기대값의 의미 승인, 원본 관측 성공을 대신하지 않습니다.

검토자는 실제 원본 빌드·기동·초기화·API 관측, production controller 빌드·기동, 가짜 자식 실행, 모델 작업, 새 HTTP 조회, 외부 게시를 모두 0회 유지했습니다. 별도의 독립 검토자도 같은 경계를 유지했습니다. 작성자가 실행한 가짜 자식 테스트의 로그를 읽어 검토했으며 다시 실행하지 않았습니다. 기존 first20 계약 검토와 Native2 관측기 검토 봉인은 수정하지 않았습니다.

## 왜 두 단계로 계획을 고정하나

아직 만들지 않은 실행 파일의 해시를 앞선 계획에 넣고, 그 실행 파일에 다시 앞선 계획 해시를 넣으면 서로의 해시가 서로를 요구합니다. 이 제어 코드는 다음 순서로 참조 방향을 정합니다.

1. Root의 **static reservation**은 이미 고정된 소스 묶음, 입력 벡터, 소스·로컬 권리 검토, CI 근거, 자원 제한과 세 경로를 참조합니다. 미래의 worker·NativePlan·Readiness·final binding 해시는 넣지 않습니다.
2. Readiness와 NativePlan은 reservation을 참조합니다. 원본 관측기의 실행 파일은 고정된 closure/vector/readiness 해시를 빌드 설정에 선언합니다.
3. 마지막 **execution binding**이 실제 reservation·plan·worker·controller 실행 파일과 controller 소스를 연결합니다. 자체 실행 파일의 해시를 소스에 다시 삽입하지 않습니다.
4. 제어 코드가 모든 연결과 제한을 확인하고 `start.once`를 독점 생성·동기화한 뒤, 다시 검사하고 한 번의 Start로 진행합니다.

이 순서에서 순환은 보이지 않습니다. JSON의 모르는 필드와 중복 키를 거절하고, 모델·보호된 평가·공개 쓰기 권한은 false로 유지하며, 관측 수는 23으로 고정합니다.

## 실제로 확인한 보호 장치

| 영역 | 확인한 코드 동작 | 해석의 범위 |
|---|---|---|
| 시작 전 검사 | closure/vector/Readiness/NativePlan 및 관련 파일의 바이트 수·SHA, 소스 목록, 선언된 Go 빌드 설정을 확인 | 원본 import/init 없이 메타데이터와 파일만 읽음 |
| 경로 | 절대 정규 경로, 모든 조상 symlink 금지, source/attempt/outside의 물리적 분리, controller 소스와 출력의 분리 | 기존 디렉터리는 inode도 대조하고 새 출력 leaf는 소문자 ASCII로 제한 |
| 단회 시작 | 새 디렉터리·O_EXCL 파일·fsync·디렉터리 sync, 이후 재검사, 직접 worker `--plan` 실행 | 쉘·help 호출·자동 재시도 없음 |
| 시작/회수 | Start 실패는 Wait0, Start 성공은 Wait1; timeout/출력 오류 때 process group에 kill을 요청하고 Wait 결과를 받음 | 시작 뒤 main 오류를 원본 init0으로 기록해서는 안 됨 |
| stdout | 받은 바이트의 수와 SHA만 기록, 원문 보관0, 12MiB 초과 감지 시 중단 | 중단 전에 더 받은 바이트도 실제 수에 포함; 강제 pipe 종료 뒤 미전달 바이트는 알 수 없음 |
| stderr | 받은 전체 바이트의 수·SHA, private prefix 최대64KiB 및 prefix SHA, 잘린 양과 I/O 실패 표시 | private 로그를 공개 자료에 복사하지 않음 |
| 원본 결과 | worker 디렉터리의 fence와 결과 파일을 검사·해시하며 밖으로 복사0 | 제어 코드는 결과 JSON의 의미나 23개 Want의 타당성을 판단하지 않음 |
| 자원 | 실행 전 CPU1·Go soft heap256MiB 환경, 300초 timer, 종료 뒤 whole-child wall 및 Darwin MaxRSS>0·≤256MiB 검사 | Go heap/GPU 측정으로 대체하지 않음; RSS는 hard limit이 아님 |
| 기록 | 바깥 메타데이터·로그1MiB 이내, 실패·partial 파일 보존, 자동 cleanup/retry 없음 | 저장 오류는 별도 exit로 보고하므로 Root는 exit와 readback을 함께 확인 |

Lock 없이 stream writer별 goroutine이 상태를 소유하고 `Cmd.Wait` 뒤 snapshot을 읽는 구조입니다. 테스트마다 가짜 자식이 같은 Go 테스트 바이너리만 다시 실행하므로 원본 패키지나 모델 초기화를 포함하지 않습니다.

## 발견한 두 문제와 수정

첫째, 원본 관측기는 결과 파일을 O_EXCL로 연 뒤 첫 write를 합니다. 그 짧은 사이에는 파일 크기가 0인 것이 정상입니다. 이전 감시 코드는 실행 중에도 0바이트를 거절하여 정상 단회 관측을 오탐으로 종료할 수 있었습니다. v4는 **실행 중 0바이트 허용 / 종료 후 nonempty·fence·hash 필수**로 구분하며, 같은 Open→Write 순서의 가짜 파일 테스트가 추가됐습니다.

둘째, 문자열 경로만 비교하면 대소문자를 구분하지 않는 macOS 볼륨에서 `source`와 `SOURCE`가 같은 위치인 것을 놓칠 수 있습니다. v4는 기존 디렉터리와 조상의 inode를 대조하고, 아직 없는 출력 leaf는 소문자 ASCII만 허용하며, 같은 물리 parent의 별칭도 거절합니다. 실제 Root 경로가 겹쳤다는 발견은 아니고, 이전 코드가 약속한 물리적 분리를 보장하지 못하던 경계입니다.

작성자의 가짜 테스트 로그 v1/v2/v3/v4는 각각 top-level 테스트 12/13/14/15개 PASS입니다. 네 회 모두 실제 원본 관측0이며, 각 회의 네 가짜 자식 Start/Wait와 v2–v4의 의도적 Start 실패는 과학 실험 재시도가 아닌 제어 코드 테스트입니다. v4 15개 PASS만으로 production preflight, 실제 컴파일 입력, 원본 init, 기대값 의미가 검증됐다고 말하지 않습니다.

## 실제 실행 전에 Root가 해야 할 일

Root가 각 CI·권리·자원 근거의 **내용**을 직접 판독하고 그 판단을 Readiness에 봉인해야 합니다. 근거 파일 자체도 미래 worker·NativePlan·Readiness·final binding 해시를 역참조하지 않는 static 자료로 골라야 합니다. 제어 코드는 그 근거 파일의 ID·바이트·SHA를 읽지만 그 JSON의 CI 성공 여부나 권리의 법적 의미를 독립 해석하지 않습니다. `true` 선언은 Root가 올바른 근거를 검토했다는 책임 경계입니다.

Root가 실제 compiler selection과 빌드 명령·환경·도구 입력을 기록하고 worker/controller 바이너리를 고정해야 합니다. 해시와 `debug/buildinfo`의 선언은 정확한 연결에 도움이 되지만, 실제 컴파일 입력을 독립적으로 입증하는 증명은 아닙니다. 파일 재검사는 일반 변경을 감지하지만 악의적인 동시 변경을 커널 실행 시점까지 막는 보장은 아닙니다. 봉인 입력과 경로의 소유권을 Root가 유지해야 합니다.

단회 marker가 생성된 뒤 pin 검증·Start·Wait·기록 저장이 실패해도 그 시도는 보존합니다. marker 삭제나 새 이름으로 자동 재시도를 하지 않습니다. 바깥 프로세스가 강제 종료되면 동기화된 intent만 남고 자식 완료 상태는 불명일 수 있으므로 Root가 상태를 확인해야 합니다. 사전 CLI/입력 실패로 안전한 출력 위치를 확정하지 못한 경우에는 제어 코드가 파일을 쓰지 않고 exit2로 끝날 수 있어 Root의 상위 실행 기록도 필요합니다.

300초 제한은 성공 판정의 whole-child wall 상한이며 실패 cleanup까지 300초 안에 반환한다는 보장이 아닙니다. MaxRSS는 Darwin 직접 자식에 대한 종료 후 측정으로, 모든 descendant나 controller 자체, GPU 메모리의 합계를 뜻하지 않습니다. `GOMAXPROCS=1`도 OS CPU affinity를 뜻하지 않습니다. 실패한 결과 파일을 밖으로 복사하지 않으므로 Root의 추가 확인이 필요하면 원래 worker 출력 위치에서 상태·바이트만 조사해야 합니다.

## 대안 비교

1. **현재 v4 + Root의 고정 근거·빌드·단회 실행 계획**: 이번 23개 유한 관측에 가장 맞습니다. 수정 두 건과 별도의 근거 책임을 유지하며 실제 실행은 Root가 소유합니다.
2. **모든 proof의 의미를 controller에서 해석하는 확장**: 자동화에는 유용할 수 있지만 CI/권리/자원 스키마를 새로 고정해야 하며 이번 범위에서는 새 필요성과 근거가 없습니다. 단순 pin을 의미 검증으로 포장하지 않는 것이 우선입니다.
3. **실행을 보류하고 더 강한 격리·kernel/toolchain attestation을 추가**: 적대적인 동시 변경이나 모든 descendant의 자원 합계가 요구될 때 맞습니다. 현재 로컬 단회 관측의 보장보다 넓은 별도 실험으로 다룹니다.

소스 위치·핀·회귀 근거·후속 조건은 동봉 JSON에 있습니다. 실제 관측이나 성능 개선 수치는 이 정적 검토에서 생성하지 않았습니다.
