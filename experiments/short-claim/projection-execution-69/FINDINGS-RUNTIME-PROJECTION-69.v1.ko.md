# 저장된 원본 72개 projection69 실행 결과 검토

원본 결과의 핀, 배열 연결, 제외 범위와 자원 기록을 읽기 전용으로 대조했고 이 범위에서는 문제가 없습니다. 원본 실행은 root가 한 번 수행했습니다. 이 검토는 원본 실행을 재현하거나 새 정답을 만들거나 학습·성능·다양성 준비를 승인하는 실험이 아닙니다.

정확한 원본 결과는 7,699,810 bytes, SHA256 `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`입니다. root 동결 계획 SHA `1811c6c098468acd48e774075263103af4bc87a6f42e398cc7920abe2edb8d84`와 예약·시작 전 단계·완료 장부를 연결했습니다. draft→frozen은 `frozen:false→true`만 바뀌었습니다. 실행 입력·소스·binary 30개 핀 8,158,159 bytes와 전체 참조 42개 15,882,779 bytes를 실제 파일 SHA/bytes와 대조했습니다. private 경로와 원본 시간 로그는 공개용 기록에 넣지 않았습니다.

## 보존된 데이터

| 항목 | development | validation | calibration |
|---|---:|---:|---:|
| 부모 | 44 | 16 | 12 |
| 후보 위치 | 132 | 48 | 36 |
| unknown 후보 | 42 | 12 | 9 |
| 학습용 배열 행 | 90 | 36 | 0 |
| 가중치 0인 유지 행 | 18 | 1 | 0 |
| whole group | 11 | 3 | 3 |
| 가중치가 양수인 행이 있는 group | 9 | 3 | 0 |

총 72개 부모/216개 후보에서 known 34개, no_answer 17개, unknown 21개가 그대로입니다. 부모별 후보 수는 2개/3개/4개인 부모가 각각 12개/48개/12개이며 원래 순서를 유지합니다. 원래 role JSON의 28개 비위치 행은 original index로 연결했고 원래 JSON 순서를 바꾸지 않았습니다. 원문 request/caption/ID, acceptable의 전체 순서·집합, group·role, loss mask와 nullable label을 핀한 입력에 대조했습니다. unknown 후보 63개는 null 의미를 유지하고 calibration 후보 36개는 RuntimeParent에 남지만 fit 배열과 특징 계산 대상에 들어가지 않습니다. 알려진 행을 가중치 0으로 보존하는 것은 새 음성 정답을 만드는 작업이 아닙니다.

## 저장 배열과 진단 범위

희소 열의 offsets, index/value 길이, 행별 index 유일성·8192 차원 경계·finite 값, 원래 parent/candidate 순서와 label/group/weight를 검증했습니다. development/validation 특징 항목은 64,411/33,278개입니다. 가중치 0인 19개 행도 특징과 label을 가진 원래 행으로 남습니다. 배열 group의 역할 누출을 확인했습니다.

진단 배열은 양수 가중치 행 72개/35개만 정확히 복사하며, 특징 indices와 float64 values까지 bit 단위로 원본 배열의 해당 행과 같습니다. 진단 특징 항목은 35,223/32,960개입니다. 진단 `Data.Excluded=0/0`과 `UnknownAuditCandidates=42/12`는 서로 다른 의미를 그대로 지킵니다. 진단 양성/음성 label 수는 20/52, 10/25입니다. AUC 계산은 0입니다.

저장된 normalized request 72개와 caption 216개를 이미 핀한 lexicalhint 문자열 분할 recipe의 stdlib 복사로 대조했습니다. 이는 저장 문자열 확인 288개이며 원래 NormalizeText/Validate/Features/Project 함수 재호출은 0입니다. 특징 값의 수학적 생성 의미는 재계산하지 않았습니다. 그 의미를 독립적으로 입증한 것으로 확대하지 않습니다.

## 자원 수치의 의미

| 기록 | 값 | 범위 |
|---|---:|---|
| 소유 배열·헤더·문자열 PayloadBytes | 1,774,364 bytes ≈ 1.692 MiB | 저장 배열 길이와 동일 Go ABI struct 폭으로 재산출; allocator/scratch 제외 |
| 원본 JSON 직렬화 크기 | 7,699,810 bytes ≈ 7.343 MiB | 결과 파일 크기 |
| 전체 child 최대 RSS | 56,573,952 bytes ≈ 53.953 MiB | macOS time의 관측 bytes, hard cap 아님 |
| 전체 child 최대 footprint | 49,857,016 bytes ≈ 47.547 MiB | macOS 관측 bytes |
| Go heap soft 설정 | 268,435,456 bytes = 256 MiB | 설정이며 실제 heap 관측값 아님 |
| 전체 child 표시 real/user/sys | 0.46/0.07/0.01 seconds | 입력 읽기·검증·Project·직렬화를 포함한 전체 프로세스 |
| controller wall | 0.465949875 seconds | controller 관측 전체 대기 시간 |

owned payload와 JSON 출력은 기존 64 MiB 각각의 한도 안입니다. 원래 direct Validate 72번, Project 1번, 반환된 FeatureScans 252번 기록은 행 126개의 두 pass와 맞습니다. Project 내부 ValidatePrepared/normalization 개별 횟수는 기록되지 않았으며 그 값을 새로 추정하지 않았습니다. 직접 Project만의 시간, 실제 Go heap, GPU 사용량, AI가 준비한 비용은 측정되지 않았습니다. 이 수치는 학습·모델 추론 벤치마크가 아닙니다.

## 실패·저자 노출·한계

독립 helper 첫 시도는 검토자 코드가 모든 부모를 후보 3개로 잘못 가정해 `runtime69_parent_binding`에서 거부됐습니다. 원본 데이터 문제가 아닙니다. 최초 source/tests/module·호출 계획·실패 기록을 그대로 보존했고, 부모별 원래 후보 수로 수정한 두 번째 metadata helper는 통과했습니다. 합성 race 두 번은 모두 통과했고 최신 suite는 top-level 6개/subtest 6개, vet 두 번 통과입니다. 별도 실패 read/tool/test 0입니다. 실제 helper 첫 실패와 pass의 compile/tool 시간은 준비 비용이며 원본 실행 시간과 합치지 않습니다.

검토자는 driver69/claimfit.Project 작성자가 아니지만 pairlearn trace, roleplan62, scope67 작성과 이전 자료를 본 nonblind 노출이 있습니다. 별도 원천 저자 독립성·blind 의미 평가를 주장하지 않습니다. 저장 상태를 대조하는 이 검토로 concurrent mutation의 불변성, hermetic compiler, 문구 충실성, source 다양성, 모델 유용성을 입증하지 않습니다. 원래 truth/mask/role은 바꾸지 않았고 새 수치 gate, 역할, label, seed, 모델, fit, paid API, native 실행, protected final 읽기, 공유 수정·Git·게시 모두 0입니다. `training_ready=false`, `source_diversity_cleared=false`를 유지합니다.

관련 파일: [실행 결과 검토 receipt](RECEIPT-RUNTIME-PROJECTION-69.v1.json), [배열·핀·자원 mechanics](MECHANICS-RUNTIME-PROJECTION-69.v1.json), [검토 장부](LEDGER-RUNTIME-PROJECTION-69.v1.json), [첫 helper 거부 기록](HELPER-ATTEMPT-1-REFUSAL.json).
