원래 함수의 유한 동작을 한 번 관측하기 전에, 프로세스 시작 **밖에서** 조건을 확인하는 실행 제어기의 준비물입니다. 첫 두 요청은 IPNet 3후보×5입력과 OrCompose 2후보×4입력, 총 23개 관측입니다. 23행을 23개 새 의미 요청이나 학습 라벨로 세지 않습니다. 원문 worker의 초기화·함수 실행, 모델 API, 보호 세트, 외부 게시와 production 실행은 모두 0입니다.

worker 내부의 승인 확인은 main보다 먼저 실행되는 패키지 초기화를 막을 수 없습니다. 그래서 이 제어기가 소스·입력·CI·권리·자원·디렉터리·바이너리를 확인한 뒤 한 번만 시작하도록 설계했습니다. Go 표준 라이브러리와 기존의 자체 pure protocol만 사용하며, 원문 패키지나 모델을 import하지 않습니다. `source/`는 실행하지 않는 소스 사본이고, 별도 private module에 실제 소스가 있습니다.

핀의 순환을 피하려고 순서를 나눴습니다.

1. Root가 정적 reservation을 먼저 고정합니다. closure/vector, 원천·local rights 검토, 정적 CI, 자원, 고정 디렉터리만 들어갑니다. worker·NativePlan·Readiness·실행 binding 해시는 들어가지 않습니다.
2. Readiness가 그 reservation의 SHA를 참조하고 NativePlan이 Readiness를 참조합니다. worker는 이 세 핀을 선언하도록 나중에 Root가 빌드합니다.
3. 이후 execution binding이 실제 worker/제어기 바이너리, NativePlan, 소스의 SHA를 연결합니다. 이 binding의 해시를 Readiness나 reservation으로 되돌리지 않습니다.
4. 모든 조건 확인 후 `start.once`를 새 파일로 쓰고 파일·디렉터리를 fsync한 다음, 고정한 worker에 `--plan`만 전달합니다.

Source/CI/Resource proof 파일 자체도 나중 worker·Readiness·NativePlan 해시에 의존하면 안 됩니다. 이 간접 순환까지 Root가 문서를 고를 때 검토해야 합니다. 제어기는 proof의 **바이트 핀**을 확인합니다. CI의 적합성이나 라이선스 결론의 뜻을 스스로 다시 판독하지 않으며, 그 해석은 정확한 근거를 고정한 Root의 책임입니다.

| 조건 | 고정 범위 |
|---|---|
| worker 시작 | 같은 reservation의 고정 output 디렉터리에서 최대 1회; 자동 재시작 0 |
| Start/Wait | Start가 성공했을 때만 Wait 정확히 1회; Start 실패는 Wait 0 |
| Go 실행 동시성 / soft heap | GOMAXPROCS=1 / GOMEMLIMIT=256MiB |
| whole-child 시간 / 관측 RSS | 300초 / Darwin MaxRSS 256MiB |
| result / outside 파일 cap | 12MiB / 1MiB |
| stdout 원문 / private stderr | 원문 보관 0 / 최대 64KiB prefix |
| 모델·보호 세트·게시 | 허용 0 |

모든 절대 경로의 조상에서 symlink를 거절하고, SourceRoot·AttemptDir·outside를 서로 분리합니다. 문자열 경로 비교에 더해 기존 조상의 inode를 비교해서 대소문자 별칭도 확인합니다. 새 output의 마지막 이름은 소문자 ASCII·숫자·`-`·`_`만 허용해 아직 없는 이름의 대소문자/Unicode 정규화 별칭을 피합니다. 제어기 소스 디렉터리도 output과 물리적으로 분리되어야 합니다. 처음부터 존재하는 output 디렉터리는 다시 쓰지 않습니다.

시작 전에 모든 원문 closure 파일, 실제 frozen draft의 Input/Want/설명/revision/symbol, 예약 23개 shape, Readiness를 확인합니다. 바이너리에서는 Go1.27.1·darwin/arm64·CGO0·trimpath·VCS 정보 없음과 worker의 세 `-X` 선언을 확인하도록 설계했습니다. 현재 worker buildinfo 검사는 **정확히 세 `-X` 항목**을 요구합니다. Root가 다른 linker flag를 원하면 별도 소스 검토와 버전이 필요합니다. 선언과 파일 SHA는 compiler-independent attestation이나 커널의 실제 실행 inode 증명은 아닙니다. actual build tool/cache 전체 확인은 Root의 후속 빌드 계측에 남깁니다.

프로세스 그룹을 만들어 timeout·출력 cap·중단 신호에서 그룹 SIGKILL을 요청하고, 한 번의 Wait로 파이프 읽기를 마칩니다. stdout은 개수·해시만 계산하고 원문을 저장하지 않습니다. stderr prefix만 private로 보관합니다. `results.json`이나 실패 조각을 다른 곳에 복사하지 않고 종료 뒤 크기·해시와 reservation에 묶인 inner dispatch fence를 읽습니다. 결과 JSON의 뜻·정답·모델 품질은 이 단계에서 판정하지 않습니다.

실행 중 `results.json`이 0바이트일 수 있습니다. 원래 worker의 새 파일 생성과 첫 Write 사이 정상 구간이라서 허용하며, 종료 후에는 nonempty·cap·해시를 강제합니다. 이 경계를 처음 코드가 잘못 거절했던 사실과 보강 이력을 `AUTHORING-LEDGER.v1.json`에 보존했습니다. 기존 controller 원본이나 native2 observer 봉인은 바꾸지 않았습니다.

테스트에서는 원문 worker 대신 이 테스트 바이너리의 더미 자식만 시작했습니다. 최종 15개 상위 테스트 PASS에는 정상 종료, timeout 그룹 kill, stdout 초과, stderr prefix, 실패한 Start/Wait0, 빈 파일의 정상 생성 구간과 물리 경로 검사가 포함됩니다. 테스트용 짧은 시간·작은 cap은 production plan에 노출되지 않습니다. standalone production controller build와 원문 worker build/start는 0입니다. Go test를 위한 자체 소스의 컴파일과 더미 프로세스 실행은 별도로 기록합니다.

RSS는 직접 자식의 전체 시작·핀 확인·관측·저장을 포함한 Darwin wait4 MaxRSS이며 종료 후 gate입니다. OS의 강제 RSS 제한, 전체 자식 그룹의 완전한 peak, GPU 메모리 또는 Go heap으로 해석하지 않습니다. GOMAXPROCS=1도 OS CPU affinity는 아닙니다. timeout 뒤의 파이프 정리로 실패한 cleanup이 300초를 조금 넘을 수 있지만 성공 결과는 전체 자식 wall≤300초여야 합니다. 제어기가 SIGKILL되면 fsync된 once만 남을 수 있어 자식 종료와 최종 개수는 unknown으로 남깁니다. 자동 재시작으로 이 불확실성을 덮지 않습니다.

공개 템플릿은 false/빈 핀으로 실행 불가능하게 준비했습니다. 이 준비물이나 자원 계획은 실행·학습·활성화·배포를 허용하지 않습니다. Root가 정확한 소스/CI/권리/자원을 고정하고 실제 두 바이너리 크기를 측정한 뒤 별도로 production 빌드와 단 한 번의 원문 관측을 소유합니다. 실제 결과는 이 0 상태의 역사 기록을 고치지 않고 후속 영수증으로 추가해야 합니다.
