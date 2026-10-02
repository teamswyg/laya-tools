첫 두 native 요청을 **원래 Go 함수로 관측하기 위한 준비물**입니다. IPNet 3후보×5입력=15개, OrCompose 2후보×4입력=8개로 총 23개를 예약할 수 있게 설계했습니다. 현재 worker 빌드·시작·원래 함수·원래 패키지 초기화·후보 교차평가·학습·모델 API·보호 세트 접근·외부 게시가 모두 0입니다. 기존 20개 초안과 기대값을 바꾸지 않았고 후보나 라벨도 추가하지 않았습니다.

이 단계는 작은 힌트 모델을 학습할 정답을 확정하기 전, 실제 원문이 유한 입력에서 어떤 값을 반환하는지 기록할 준비입니다. 함수나 파서를 새로 구현해서 원래 함수라고 부르지 않습니다. 나중에 Root가 원천·권리·독립 검토·CI·자원을 동결했을 때만 한 번 시작할 수 있습니다. 이 문서나 준비 계획만으로 실행·업로드·모델 활성화가 허용되지 않습니다.

`source/`는 실행하지 않는 .go.txt 소스 사본입니다. private module에는 기존 Source81의 전체 upstream 소스와 go.mod를 바이트 그대로 복사했습니다. Go1.27.1에서 원문 Go 파일 48개가 선택되고 두 과거 버전 파일은 제외됩니다. 전체 소스 목록에는 이 두 파일과 순수 테스트 파일도 핀으로 남지만 실제 runtime 선택 파일로 세지 않습니다. Root가 별도로 현재 worker의 GoList 메타데이터를 한 번 확인했습니다. 123개 패키지=표준 118+원문 3+자체 2이며 선택된 표준 파일 860개가 기존 핀에 모두 포함됩니다. 준비 담당자의 GoList는 0이며 worker build info·binary도 아직 없습니다. `ROOT-SELECTION-RECEIPT.v1.json`은 이 후속 사실을 이전 준비 기록과 구분합니다.

IPNet 후보는 공개 FlagSet.IPNet/IP/IPMask 생성자로 각각의 원래 Value를 만들고 **그 Value.Set을 직접** 호출합니다. 각 입력은 새로운 초기 상태를 사용합니다. IPNet 초기 IP=[192,0,2,0], mask=[255,255,255,0]이며, 성공 후 실제 IP·mask 바이트를 독립 숫자 배열로 복사합니다. IP와 IPMask도 고유 타입·상태를 기록하고 임의로 network 결과를 합성하지 않습니다. Flag.Changed는 진단 값일 뿐 성공 판정으로 쓰지 않습니다. String이나 GetIPNet 등 상태를 다시 파싱하는 메서드는 추가 호출하지 않습니다.

OrCompose 후보는 원래 OrComposeDecodeHookFunc와 ComposeDecodeHookFunc를 생성하고, 원래 DecodeHookExec로 실행하도록 설계했습니다. 관측용 콜백은 입력 4와 유효한 int target을 받으며 E1/E2 오류, 9/99 또는 nil을 반환하는 제한된 fixture입니다. 그 콜백은 composition 알고리즘을 재구현한 것이 아닙니다. 입력이 이전 결과로 바뀌는지, 호출 순서·횟수, nil+nil 성공, E1\nE2\n 실패 메시지, 빈 hook에서의 nonnil 빈 메시지 오류를 구분해서 저장합니다. 현재 return domain은 int/null로 제한했으며 그 밖의 동적 타입까지 보존한다고 주장하지 않습니다.

| 준비 항목 | 상태 |
|---|---|
| native goal / 후보 / 입력 / 예약 관측 | 2 / 3·2 / 5·4 / 23 |
| 실제 관측·초기화·worker build/start | 모두 0 |
| 순수 protocol 테스트 | 첫 5개 PASS, gate 보강 후 6개 PASS |
| 원래 봉인 Input/Want/설명/revision/symbol 필드 대조 | 동일; 변경 메타데이터 5종 거절 |
| 메타데이터 형식만 채운 큰 더미 결과 크기 | 62,310B, 실제 Got/결과는 저장하지 않음 |
| 결과 cap / outside cap | 12MiB / 1MiB |
| CPU / RSS cap / soft Go heap / whole child 시간 | 1 / 256MiB / 256MiB / 300초 |

더미 크기는 모든 행에 16바이트 IP 상태·마스크·최대 세 콜백·256바이트 오류 메시지·64바이트 타입 문자열까지 채운 **형식 추정**입니다. 실제 출력 크기나 모든 입력에 대한 증명된 상한이 아닙니다. 12MiB는 cap이며 12MiB를 이미 소비했다는 뜻도 아닙니다. Root가 실제 결과·복사·로그·다운로드의 바이트를 경로별로 다시 기록해야 합니다. 개발 도구의 executable 크기도 모델이나 과학적 결과 크기와 따로 계산합니다.

Go1.27.1의 기존 saved selection에 포함된 표준 패키지 118개와 현재 cached source/toolchain 파일 863개의 해시를 읽었습니다. 원문 source 6,886,387바이트와 compiler/tool binaries는 읽기만 했고 새로 복사하지 않았습니다. Root의 후속 GoList는 현재 선택을 연결하지만 컴파일·링크·초기화 또는 의미 검증을 증명하지 않습니다. 추가 cached 도구 핀도 실제 사용한 tool/cache 전체 목록이나 compiler-independent attestation은 아닙니다. 실제 빌드의 자식 도구 실행 횟수는 계측 전 unknown입니다. 순수 test의 testing import는 native runtime selection 밖으로 구분했습니다. 메타 준비 첫 시도는 이 구분이 없어 중단됐고, 동일 SHA 부분 복사본을 보존하여 정정 후 다시 진행했습니다. 원래 함수나 모델의 실패·재시도는 아니며 모두 기록했습니다.

프로그램 안의 readiness 검사는 원래 패키지의 **pre-main 초기화**를 막지 못합니다. pflag.CommandLine의 NewFlagSet 같은 초기화가 main보다 먼저 일어납니다. 따라서 Root의 outside controller가 process 시작 전에 소스·vector·Want 검토·라이선스·CI·binary/build info·전체 자원·once 예약을 확인해야 합니다. SourceRoot와 AttemptDir의 모든 조상 경로에 symlink가 없고 서로 분리됐는지도 outside에서 검사합니다. whole wall/RSS는 이 시작·핀 읽기·초기화·결과 저장을 전부 포함해야 합니다. 코드의 GOMAXPROCS와 Go soft heap은 OS RSS 강제 제한을 대신하지 않습니다.

worker는 원문 API를 부르기 전에 closure, readiness, vector 핀과 기존 draft의 **실제 필드**를 비교하고, 한 번만 만들 수 있는 dispatch 파일로 이중 시작을 막도록 설계했습니다. row 수만 맞으면 출처 검증이 끝났다고 보지 않습니다. 이 파일도 pre-main 초기화를 막을 수 없으므로 outside once 예약이 먼저 필요합니다. 공개 readiness/plan template의 승인 값은 모두 false이고, 실제 실행 계획은 만들어지지 않았습니다.

라이선스는 `RIGHTS.v1.json`과 보존한 원문 고지를 확인합니다. private 복사본에는 MIT/BSD 루트 고지, Go2009/2012·Go1.27.1 고지와 Source82 notice chain을 유지했습니다. local finite observation 검토, 미래 학습권 검토, 원문/worker 재배포 검토를 분리했으며 아직 각 clearance를 확정하지 않았습니다. Go2022 복사 계보의 미확정 부분은 과거 버전의 제외 파일과 함께 남깁니다. 루트 LICENSE가 존재한다는 이유로 모든 권리를 확보했다고 표시하지 않습니다.

독립 검토는 현재 유한 범위의 타입별 state·callback·error 채널이 원문에 충실해 보인다고 판단했습니다. 이는 정적 source review이며 실제 실행 승인이나 학습 정답 판정이 아닙니다. 입력·기대값이나 범위를 수정해야 할 경우 기존 봉인을 보존하고 별도 버전으로 기록해야 합니다. 다른 18개 coding 초안의 수정·관측은 이번 범위에 포함하지 않았습니다.
