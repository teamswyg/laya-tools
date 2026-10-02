네 기존 요청의 바깥 controller 소스 준비입니다. 원래 순번 1/4/5/9, 유한 입력 19개, 후보 위치 12개, 예정 wrapper 호출 57개를 유지합니다. 입력·호출 수가 새 요청 수는 아닙니다. 실제 Go·빌드·테스트·프로세스 시작·원 init/API·모델 실행은 모두 0입니다. 모든 Go와 모듈 파일은 `.txt` 상태이며 실행 준비 완료를 뜻하지 않습니다.

[구조](ARCHITECTURE.v1.json)는 기존 파일·프로세스 처리만 재사용합니다. Ftoa의 스키마, 18관측 분모, 원형 함수, oracle, compiled closure 문자열과 권한 객체는 가져오지 않았습니다. [protocol](code/pure/protocol.go.txt)은 worker의 기존 18필드 계획, exact fixture SHA와 57순서, nullable 채널·카운터를 사용합니다. [binding 초안](BINDING.draft.private.json)과 [Root 검토 초안](ROOT-PREREVIEW.draft.json)은 frozen=false이며 원천·빌드·바이너리·검토 실제 핀이 미공급이어서 실행을 거절합니다.

Root가 활성 Go·모듈·notice·selected source 전체를 한 원천 루트에 준비한 뒤 선언한 파일 SHA와 closed Go 디렉터리를 확인합니다. Root가 저장한 selection 메타데이터는 128 packages와 upstream Go 52파일을 기록하며, humanize는 big/bytes/comma/ftoa의 4파일 slice이고 pflag/mapstructure는 해당 선택 패키지입니다. 이 저장 메타데이터 읽기는 제가 compiler를 실행하거나 전체 humanize package를 검증했다는 뜻이 아닙니다. 별도 controller 루트의 정확한 여섯 compile 파일도 확인합니다. 두 executable은 SHA와 Go 1.27.1/CGO0/trimpath/noVCS/Darwin ARM64 buildinfo, bounded Mach-O 구조를 읽습니다. [Mach-O 검사](code/macho.go.txt)는 기존 자체 작성 검사의 구조 부분만 옮겼습니다. 컴파일러 선택·저작권·CI·finite Want의 의미는 Root가 작성한 named proof의 신뢰 경계이며 hash 일치가 독립적으로 그것을 증명하지 않습니다.

[main](code/main.go.txt)은 fresh 바깥 디렉터리·부모 Sync, O_EXCL 단일 실행 intent의 쓰기/readback/fsync·디렉터리 Sync를 완료한 뒤에만 시작합니다. child 디렉터리는 worker가 만들며 시작 전에는 없어야 합니다. 최소 offline 환경만 넘기고 인증 환경을 상속하지 않습니다. CPU1/Go soft heap 256 MiB/60초/재시도0을 유지합니다. source와 출력의 문자 경로·symlink·기존 inode alias를 점검하지만 실행 중 적대적 파일 변경에 대한 kernel attestation은 아닙니다. 입력은 실행 동안 안정적이어야 합니다.

[process](code/process.go.txt)는 `/usr/bin/time -l` wrapper의 Start를 한 번 호출하고 성공하면 Wait를 한 번 호출합니다. timeout·중단·용량 초과 시 process group을 종료하고 1초 cleanup grace 안에 join을 확인합니다. join을 확인하지 못하면 실패이며 stream/호출 수를 추정하지 않습니다. running 시 absent child와 빈 result는 정상 pending입니다. worker의 reservation/journal/final 합계는 1 MiB 미만, stdout은 256 KiB, private stderr prefix는 64 KiB로 제한합니다. 관측 pipe bytes와 보존 prefix를 구분하며 초과·불완료를 숨기지 않습니다.

joined exit0, 최종 파일과 stdout exact 일치, 57행 순서·typed channel/counter/null supervision 및 완료 guard가 모두 맞을 때만 완료된 메타데이터 수를 보고합니다. 이 상태도 57만족·일반화·학습 자격을 뜻하지 않습니다. 실패 때는 최종 파일의 bounded counter prefix를 별도 reported 값으로 보존할 수 있으며 unknown 실제 수를 0으로 바꾸지 않습니다. 종료 후 남은 완전 journal 행 자체는 Sync 응답의 증거가 아니며 tail을 완료 호출 수로 승격하지 않습니다. 원 내부/getter/setup/init 실제 수는 null입니다.

RSS는 Darwin time-l의 전체 child lifetime bytes이며 startup/init/pin 검사도 포함합니다. missing·ambiguous 수치는 null로 남기고 heap 설정을 RSS 하드 제한으로 설명하지 않습니다. 256 MiB 비교는 관측 후 결과이지 live kernel 제한이 아닙니다. 바깥 대기 시간, OS real/user/sys와 내부 함수 비용도 구분합니다.

[합성 테스트 소스](code/main_test.go.txt)는 failed Start/Wait0, timeout·overflow 단일 Wait join, strict args/plan null·누락·중복·case, 빈 pending 파일, nullable OS 단위, final57/null/카운터 구조를 검사하도록 작성했으며 실행하지 않았습니다. preparer는 worker/adapters 및 이번 controller 저자이며 독립 구현·의미 승인으로 세지 않습니다. Root의 별도 소스/순수 컴파일·테스트 검토 후 실제 원천/compiler/권리/CI/resource와 두 바이너리가 봉인되어야 다음 단일 실행을 할 수 있습니다. [HANDOFF](HANDOFF.v1.json)에 파일 핀과 작성 한계가 있습니다.
