# Variable development input bridge — 소스 준비

이 모듈은 새 자료의 정답을 정하거나 모델을 학습하지 않습니다. Root가 별도로 고정한 계획과 실제 입력의 바이트·행 순서·문장·후보 ID·정답·가중치·출처 역할이 일치할 때만 다음 처리 함수를 한 번 호출합니다. 현재 소스와 합성 테스트만 준비했으며 Go 포맷·컴파일·테스트·실제 Reader·학습 실행은 모두 0입니다.

기존 `shortclaimdata` Reader는 `development_train`만 허용합니다. 이 모듈도 그 경계를 유지합니다. 기존 검증·보정 역할은 외부의 그룹 목록에 보존해 충돌을 검사합니다. 그룹 82·83·65535는 행 개수만큼 만든 배열의 인덱스로 쓰지 않고 정렬된 목록에서 찾습니다. Root 계획은 전체 가족의 고정 역할을 전달해야 합니다. 이 소스가 계보나 전체 등록부의 정확성을 대신 판정하지는 않습니다.

Root는 새로운 계획 스키마와 전체 자료 SHA를 외부 호출에 고정해야 합니다. 입력 안의 승인 문구만으로 통과하지 않습니다. 그렇지만 SHA 일치는 정답의 증거나 학습 승인이 아니라, 지정된 바이트와 일치한다는 뜻입니다. 이전 자료의 정확한 접두부 보존도 Root가 기존 행 SHA·순서를 계획에 유지하여 따로 확인해야 합니다.

선택된 후보는 알려진 정답과 단위 가중치만 받습니다. 미상 후보는 원래 ID·원래 순번·제외 사유·null 정답/가중치로 계획에 남고 모델 후보에는 들어가지 않습니다. 알려진 반례가 있는 부정 후보의 미상 관찰 수는 보존합니다. 여러 긍정 답을 강제로 하나로 줄이지 않으며, 모두 부정인 경우도 Root의 명시적 `no_answer` 없이 추론하지 않습니다.

`bridge.Admit`의 소비 함수는 모든 행 검증 후에만 호출됩니다. 후반 행이 잘못되어도 소비 함수는 0회입니다. 그 전에 기존 Reader는 일부 행을 읽었을 수 있습니다. 입력·문장 값은 `Batch.Input`, 출처와 정답은 별도 접근자로 분리하며, `Batch.Parent`는 고정된 값을 기존 claimfit 자료형으로 복사할 뿐 특징 추출·투영·학습을 호출하지 않습니다. 소비 함수가 패닉하면 이 모듈이 복구하지 않습니다. 내구성·실행 예산·Fit 체크포인트는 이후 Root 래퍼의 별도 책임입니다.

배치 입력 한도는 계획 1MiB·자료 1MiB·행 16KiB·256행입니다. 실제 메모리 사용량이나 속도 측정은 아직 없습니다. `module/`의 `.go.txt`는 비실행 소스입니다. 이후 Root가 `.go`로 복사하고 고정된 로컬 프로젝트를 형제 `laya-tools` 디렉터리에 연결하거나 별도 modfile을 마련해야 합니다. 이 준비가 33개 자료에서의 자동 학습이나 ~60개 시점의 학습 준비 완료를 뜻하지 않습니다.

# English

This source-only module binds a separately Root-frozen input plan to bounded development JSONL. It creates no labels, roles, groups, models, or Fit admission. Formatting, compilation, tests, actual Reader calls and training have not run.

Every exact row/candidate/metadata/supervision value and LF-terminated row SHA must match before the single consumer callback. A later-row failure may follow successful Reader calls but still leaves the consumer at zero. Unknown-only candidates retain source IDs/ordinals and null supervision in the exclusion ledger; known-negative candidates may retain unknown predicate counts. No-answer must be explicitly Root-frozen. Multiple positives are not reduced to one.

The unchanged Reader accepts new development_train rows only. Sorted sparse group and family-role tables preserve external registry roles and reject conflicts without indexing group numbers into parent-sized arrays. Root must supply and review whole-family membership; this adapter does not prove ancestry or registry completeness. Externally supplied plan/admission/data hashes bind bytes, not semantic truth or readiness. Root separately checks historical prefix preservation.

Input-only validated values and supervision/provenance accessors are separate. `Batch.Parent` copies supervision to the existing claimfit DTO without feature extraction, projection, scoring or fitting. The consumer is called once only after complete validation; fixed diagnostics replace raw supplied errors. Consumer panics, durable output/checkpoints, resource admission and future Fit intent belong to a separate Root wrapper.

Limits: plan/data 1MiB each, JSONL row 16KiB, 256 rows, eight candidates, 512 registry entries. These are input bounds, not measured peak memory or speed. The inert `.go.txt` and proposed module require Root materialization with the pinned sibling local project or a separate modfile. No automatic Fit at33 or readiness at~60 is asserted; historical Fit3/8 and logical-model3/16 limits remain unchanged.
