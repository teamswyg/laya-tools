# 첫 개발용 fit의 입력 고정 준비 71

이 산출물은 기존 71 v4 학습 실행기의 입력을 연결하는 **메타데이터 준비**다. 학습·특징 생성·기준 도구 실행·역할 재배정·모델 입출력 API는 호출하지 않았다. 계획은 아직 `frozen=false`, `caller_frozen_bounded_development_prerequisites=false`다. 최초 실제 fit은 부모 실행자가 별도의 실행 감시기를 한 번 호출하는 작업이다.

`bundle-v1/execution-plan-71.json`은 15,812바이트, SHA256 `87ed7c3aa0ffd7e2afe8024610b0ac4bc7d1702ead97f9aa342a3e0dbd43cc44`다. `bundle-v1/BOUND-EVIDENCE-MANIFEST-71.json`은 9,630바이트, SHA256 `11721878d37398162493113c890850e1f287911e99b009b61f45a326c29c846c`다. 입력·드라이버 코드·바이너리를 계획 폴더에 상대 경로로 새로 복사했고 원래 바이트와 SHA를 보존했다. 저장소 원래 코드 15개와 `original-plan56`은 명시적인 저장소 루트 기준 상대 경로다. 실제 실행 계획의 실행 폴더 값과 복사한 드라이버의 모듈 파일에는 비공개 경로가 있으므로 원본을 그대로 공개하지 않는다.

연결한 주요 증거는 combined76 `4e9d36db…`, actual-role70 `24de7b74…`, 역할 독립 검토 `34751236…`, 원래 71 v4 독립 검토 `f4e7b715…`, 76 사전 검토 `502f148e…`, 실제 76 projection `f68bff5f…`, 실제 projection 독립 검토 `962b977d…`, 원래 truth56b·masks67·두 probe 파일·contract68·finiteqa68이다. 정확한 바이트·전체 SHA·경로는 manifest에 있다. 새 truth나 label은 만들지 않았다. 네 추가 공개 원천 요청의 감독 정보는 여전히 좁은 유한 관측에서 도출한 제안 범위이며, 문구 다양성 전체 승인이나 새로운 원천에 대한 일반화 증거로 바꾸지 않았다.

학습 설정은 FP32, 8192 차원, seed 1729, 최대 50 epoch, batch 128, learning rate 0.1, L2 0.0001 그대로다. 이번 실행은 하나의 fit만 허용한다. 원래 캠페인의 최대 8 fit·16 artifact 예산은 바꾸지 않았다. CPU 1, Go heap soft 256 MiB, 바깥 시간 제한 300초, 새 결과 64 MiB 제한을 유지했다.

고정된 76부모·226후보는 known 38, no_answer 17, unknown 21이다. 기존 known/no_answer 51부모와 추가 제안 known 4부모를 합쳐 55부모이며 후보 전체와 원래 acceptable 집합을 유지한다. 학습에는 91행(그중 loss 0인 18행), 검증에는 45행(그중 loss 0인 1행)이 남는다. unknown 후보 39/15는 각각 학습/검증 감사 분모로 보존하고 학습 loss에 넣지 않는다. calibration의 known 후보 27개와 unknown 9개는 감사용으로 보존하되 fit 입력은 0행이다. 손실 마스킹은 정답 변경이나 후보 삭제가 아니다.

주 비교 A는 같은 검증 역할의 known/no_answer 부모에 있는 모든 후보를 사용하며 loss 0 후보도 포함한다. `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule` 순서를 유지한다. 점수가 같으면 원래 후보 순서, 가장 강한 기준 도구의 총 검사 수가 같으면 이 고정 도구 순서를 유지한다. 원래 5% headroom 검사를 다시 같은 범위에서 계산하고, 부족하면 fit 전에 멈춘다. legacy(<48), typed(48–71), upstream_finite(72–75), 원래 known/no_answer의 mask 영향 부모라는 strata는 저장 메타데이터에서 기계적으로 고정했다. 이 정보는 진단 전용이며 특징이나 새로운 학습 조건이 아니다. 추가 두 원천 가족은 실제 역할 결과에서 모두 train이므로 이 실행에서 원천 가족을 달리한 검증 성능은 측정하지 못한다.

범위를 제한한 caller evidence는 이 binder 저자가 부모의 개발용 준비 지시에 맞춰 작성한 **제안 기록**이다. 독립 검토자가 작성한 학습 승인 기록으로 가장하지 않는다. 실행자는 이 기록과 기존 독립 QA를 읽고 개발용 최초 실행 범위만 확인한다. 전체 corpus의 `TrainingReady`, 전체 다양성, 일반화, 공개 모델 적격성, 제품 배포 승인은 모두 false다. 실행 감시기는 계획의 두 false만 true로 바꿔 같은 bundle 안에 독점 생성하고, fsync한 예약 기록 이후에만 자식을 시작한다.

실행 감시기는 프로세스 그룹에 대해 300초에 SIGKILL을 보내고 이후 Wait를 수행한다. 정확한 전체 감시기 종료 시간이 300초 이하라고 보장하지 않는다. Go heap 제한은 soft이며 RSS hard cap이 아니다. 실제 전체 자식의 OS RSS는 바이트, 시간은 초로 기록하고 256 MiB 예산 이내/초과/미관측을 표시한다. 비정상 종료에서는 저장된 결과 접두사만 기록하며 실행 중이던 trainer 내부 호출 수를 추정하지 않는다. 결과의 Schema/State/Counts, frozen plan SHA, Go 버전을 확인하고 private 모델은 정확한 32,792바이트·SHA만 기록한다. 효용 실패와 fit 완료는 별도로 기록한다.

32,792바이트는 FP32 모델 직렬화 크기(24바이트 헤더+8192×4)다. 실제 Go decoder는 `[]float64` 계수 backing array 65,536바이트를 만든다. trainer의 FP64 shadow와 gradient도 각각 65,536바이트이고, 양자화 값·best copy·입력·JSON·allocator 등의 메모리가 추가된다. FP32 저장 형식이 이 모든 메모리를 FP32로 줄인다는 주장은 하지 않는다. 이번 준비는 저비용 힌트의 효용을 시험할 수 있게 연결한 것이며 효용이나 모델 메모리 성능을 이미 입증한 결과가 아니다.

저자는 76 worker·이 71 binder/감시기의 제작자이고 71 fitter 제작자나 독립 검토자가 아니다. 이전 62/67 작업과 69 검토에 노출돼 있어 blind review가 아니다. 순수 stdlib synthetic 검사는 binder 6개 test/35개 subtest, 감시기 최종 6개 test가 통과했고 race/vet/build를 마쳤다. 메타데이터 spec 준비 1회와 binder 실행 1회가 성공했다. 감시기 실제 실행·fit·모델/paid API·원본 source/role/Validate/Features/Project API·Git/외부 게시 실행은 모두 0회다. 실패한 조회 두 명령(없는 디렉터리 3개/파일 1개)과 patch 문맥 불일치 두 건은 준비 장부에 보존했다.
