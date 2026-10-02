# Full76 Go projection 실행기 준비

실제 코드·합성 테스트·실행 파일·동결 가능한 draft를 준비했습니다. binary를 한 번도 실행하지 않았으며 원본 Validate/Normalize/Features/Project/역할 API/모델/학습/paid API는 모두 0입니다. 저장 metadata의 SHA와 build info를 읽는 별도 stdlib 준비 helper만 한 번 성공했습니다. 실제 metadata join도 실행하지 않았습니다. 이 자료는 실행기 제작자의 준비 기록이며 독립 검토나 학습 준비 승인으로 표현하지 않습니다.

새 작성자는 이전69 driver의 독립 검토자였지만 이번76 driver 작성자입니다. roleplan62, pairlearn trace와 scope67 작성 및 이전 자료에 대한 nonblind 노출도 있습니다. 기존69 driver의 구조·희소 배열 검사를 참고했고, core claimfit.Project 구현은 바꾸지 않았습니다. 공유 repository/Git/게시 변경은 0입니다.

## 실제 고정 입력

combined70 metadata 2,694,968 bytes SHA `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21`와 실제70 역할 결과 29,864 bytes SHA `24de7b745747966107b00c6748a61dc920d97ed9cc9f98a8f22fd367cd8b6b70`를 고정했습니다. 원래72 입력의 probes56/56b·truth56b·masks67과 source68 contract·finite supervision QA도 별도 입력으로 고정해 대조합니다. Source IDs, repository·license·component·historical role 등은 옆자료이며 특징에는 원래 request/caption text만 들어갑니다. 이전66의 역할은 새76의 역할로 복사하지 않습니다.

76개 부모/226개 후보는 원래34 known +17 no_answer +21 unknown 부모에 좁은 source68 제안된 known 부모4개를 합친 자료입니다. source68의 추가10개 위치는 기존 finite 제안 label5양성/5음성으로 보존합니다. 전체 함수의 새로운 정답·문장 생성·확정적인 의미 인증을 만들지 않습니다. 실제70 결과에서 두 신규 원천 family는 모두 train으로 배정됐으며, 이 구성으로 원천 family에 대한 held-out 일반화 주장을 하지 않습니다. seed 탐색·역할 재계산은 0입니다.

## 연결과 실행 경계

Go6개 source+go.mod는 명시적인 bounded 배열/슬라이스를 사용합니다. 역할의 배열 위치를 가정하지 않고 `[76]int parent_index→row_index`를 전부 -1로 초기화해 범위·중복·누락을 거부하고 원래 JSON 순서를 보존합니다. 원래 text/acceptable/nullable label/mask는 원본에 대조하고, source68은 별도 proposed flag와 nullable proposed label에 대조합니다. 모델에 source/role/label/fidelity metadata를 특징으로 넣지 않습니다. maps/locks나 병렬 shared mutation은 사용하지 않습니다. 반환 projection의 sparse columns는 기존 Go claimfit 구현을 사용합니다.

draft의 `frozen=false` 때문에 원래 worker 실행은 거부됩니다. root는 독립 검토·소스 CI와 source/binary/input pin 검증 후 `frozen=true`로 별도 봉인해 최초 실행할 예정입니다. private plan에는 물리 경로가 있지만 [public view](execution-plan-76.v1.public-view.json)는 별도 artifact/SHA의 상대 namespace이며 실행 가능한 plan이 아닙니다. `--plan`, `--plan-sha256`, `--output-dir` 세 옵션을 모두 요구하고 추가·잘못된 인자는 고정된 실패로 끝나며 인자 값을 노출하지 않습니다.

새 attempt directory와 reservation/result를 exclusive로 확보한 뒤에만 metadata를 연결합니다. 첫 원본 Validate 전에 marker, 76개가 모두 준비된 뒤 단일 Project 전에 marker를 fsync합니다. 완료·error·panic은 고정된 오류와 attempts/returned/accepted prepared prefix를 result/ledger에 보존합니다. 프로세스가 밖에서 강제 종료되면 root controller와 마지막 영속 marker가 남고, 내부 함수 횟수를 추측하지 않습니다. 같은 prepare/Project와 output path의 재시도를 거부합니다. binary가 현재 자신·소스30개 pin과 Go1.27.1/CGO0/darwin-arm64/trimpath 설정을 검증하지만 이것은 hermetic compiler 증명은 아닙니다.

## 원래 역할·mask에 따른 예상 배열

| 항목 | train | validation | calibration |
|---|---:|---:|---:|
| 부모/후보 | 44 / 130 | 20 / 60 | 12 / 36 |
| unknown 부모/후보 | 13 / 39 | 5 / 15 | 3 / 9 |
| 학습 배열 행 | 91 | 45 | 0 |
| 가중치0 유지 행 | 18 | 1 | 0 |
| 양수 가중치 진단 행 | 73 | 44 | 0 |
| whole group | 12 | 4 | 3 |
| 학습행 양수 가중치 group | 10 | 4 | 0 |

예상 FeatureScans는 136행을 두 pass로 처리한272입니다. 이는 이미 관측된70 역할과 기존 mask로부터 계산한 보존 검증이며 새로운 학습 gate가 아닙니다. unknown63개 위치는 nullable audit로 유지하고 calibration36개 위치는 fit/특징 계산에서 제외합니다. 가중치0 행19개는 label과 특징을 가진 원래 행으로 유지하며 truth를 바꾸지 않습니다. 진단 Data.Excluded는0/0, unknown 분모는39/15로 구분합니다. 진단 양성/음성은22/51와13/31입니다. 진단 배열은 양수 weight 행의 정확한 복사이며 AUC는 실행하지 않습니다.

## 자원과 검증 기록

입력/source/driver 단일 파일은4 MiB, binary64 MiB, 소유 payload64 MiB, 직렬화 출력64 MiB입니다. GOMAXPROCS=1, Go heap soft 설정256 MiB를 요구합니다. 밖의 root controller가300초 process-group 한도를 담당해야 하며 이 코드 자체는 OS RSS cap이나 GPU 측정기를 제공하지 않습니다. 현재 actual memory/CPU/Project latency/FeatureScans를 측정한 것으로 주장하지 않습니다.

합성 테스트는 76개 toy 자료와 callback만 사용했습니다. 원래 단어분할·특징·Project 함수를 테스트 fixture에서 호출하지 않습니다. 첫 suite는 toy에 group20개를 잘못 넣어 scope 검사에서 top-level3개가 실패했습니다. 그때 negative subtest10개도 같은 초기 거부만 통과했으므로 유효한 개별 경계 검증으로 세지 않습니다. 이전 source/tests를 그대로 보존했고19 groups fixture와 baseline 확인을 추가한 두 번째 suite는 top-level7개/subtest10개 모두 통과했습니다(1.417s). vet1회, build1회, metadata helper1회와 helper vet1회는 성공했습니다. 준비 읽기에는 잘못된 selector1회(exit5), 존재하지 않은 추정 controller 경로1회(exit2), 의미 있는 출력이 없는 selector1회(exit0), 과대 출력 truncated batch1회가 있었고 장부에 보존합니다.

binary 4,371,426 bytes SHA `66957eb9569e345257fb12ebb7cd9201fab92312278fb8b5698d0f73b44ec161`; private draft 7,660 bytes SHA `1c66fc10cccbff5c20a9610f7554da35fb514607b78bbdb4825c6cc82fdfe0da`; [preparation pins](PREPARATION-PINS-76.v1.json) 30실행 핀/8,662,721 bytes. 소스·binary·private module/draft는 물리 경로를 포함하므로 그대로 공개하지 않습니다. 안전한 KOEN/receipt/ledger/public view만 별도 사용합니다. source fidelity·다양성·개별 인간 작성·실용 성능과 training_ready는 승인하지 않았습니다.
