# Fit79 guard v2 준비

기존 v1의 저장량 검증에서 두 누락을 수정한 별도 준비 버전입니다. 원래 소스, HANDOFF, binary를 수정하지 않았습니다. 학습 결과나 모델 승인이 아닙니다.

첫째, 저장 정책 변경안은 frozen: true와 정확한 state 값 frozen_resource_policy_for_one_data_effect79_stage를 가져야 합니다. audit SHA만 넣은 pending 제안은 거부됩니다. current_registry_bytes와 SHA는 실행 계획의 유일한 retained-serialized-audit-registry 증거 pin과 같아야 합니다. 해당 audit JSON의 schema, snapshot 완료 여부, 경로별 집계, 항목 수와 합계를 검증한 뒤 모든 audit ID / bytes / SHA가 현재 소비 registry에 포함됐는지 확인합니다. 새 항목 추가는 허용됩니다.

둘째, registered_model_artifacts의 각 모델 경로 / bytes / SHA가 저장량 집계 Assets에 반드시 포함돼야 합니다. 파일 존재와 해시만 맞춰 놓고 저장 예산에서 빼는 경우를 거부합니다.

실제 pinned audit은 2,776개 경로와 206,707,066 B를 기술하며 빈 regular file이 1개 있습니다. 빈 파일도 별도 항목으로 집계하고, asset 전용 처리에서 0 B 크기와 빈 내용의 SHA를 확인합니다. 학습 입력 readPin의 양수 크기 규칙은 그대로입니다. audit의 완전성은 선택 범위와 snapshot에 한정되며 전체 기계나 삭제된 자료까지 보증하지 않습니다.

main / fit / utility / seed / role 관련 코드는 그대로이며, public source 16개는 commit 01ce47dcca16c8f82022c97354d82105b6819875 / CI 37015530233의 바이트와 일치합니다. 새 private CLI는 로컬에서 별도로 컴파일했습니다. 해당 공개 CI가 이 private CLI를 검사했다는 뜻은 아닙니다.

합성 race 검사에서 상위 test 23개와 subtest 45개가 통과했고 실패 및 skip은 0개입니다. vet, formatting, CGO 0 / trimpath build, GoList의 선택 source 확인도 통과했습니다. 모든 test는 합성 DTO / 파일 / callback을 사용합니다. 실제 corpus Fit, Score, Features, Project, RoleAssign, 모델 Encode / Decode, driver main / help 실행은 0회입니다.

Root의 frozen 변경안, 실제 79개 projection, runtime QA, 전체 소비 registry와 독립 검토가 아직 필요합니다. 기존 64 MiB 초과는 과거 실패로 유지하며 조용히 성공으로 바꾸지 않습니다. 새 stage 데이터 / 모델 / 수치 기록 예산 64 MiB, sparse payload 64 MiB, CPU 1개, 관측 RSS 256 MiB, 제한 시간 300초는 그대로입니다. 기본 64 MiB retained 한도도 그대로이며 512 MiB 분기는 Root의 실제 frozen 정책 pin이 있어야만 실행될 수 있습니다.

공개 안전 자료는 원본 Go 소스를 .go.txt로 보관합니다. binary와 모델 본체, host 절대 경로가 들어간 go.mod와 raw build log는 공개하지 않고 크기와 SHA만 남깁니다.
