# Historical source witness / 과거 소스 지문 대조

`prepared.go.txt`는 공개 커밋 `9ce0e73ed03b7440b5701469d8fe642efd1b3a9b`의 `pkg/hintprepared/prepared.go` 전체 바이트입니다. 5,434 B, SHA256 `c00bc5424c3014fd32aa7f3e707cb32c25a223b2ab05eb03b6e617924fc834a5`. 원래 Apache-2.0 고지를 그대로 보존합니다.

현재 Prepare가 개선되면서 과거 비용 검증기의 source pin과 현재 소스가 달라졌습니다. `scripts/verify-chi-saved-cost.sh`는 이 옛 파일과 변경되지 않은 나머지10파일을 임시 대조 폴더에 복사합니다. 변경하지 않은 기존 검증기가11파일의 전체 지문을 검사하고 과거 관측을 확인합니다. 지문·동결 파일·수치·검사 기준을 완화하지 않습니다. 임시 폴더는 종료 시 정리합니다.

This is the exact public file at the commit above, retaining its Apache-2.0 notice. The model-free saved-cost verifier requires those historical bytes. The maintainer script creates a temporary source witness with this old file and ten unchanged current files; the unchanged verifier checks all eleven original pins. This old source witness is read, not compiled or executed. No new worker/model/cost execution occurs. Frozen historical observations and thresholds remain unchanged.

현재 런타임·새 비용 결과는 [constructor-scratch](../../constructor-scratch/README.ko.md)와 [English](../../constructor-scratch/README.en.md)에 있습니다. New constructor results are separate from this historical witness.

[복구 검증 기록 / repair controls](CI-REPAIR.actual.public.v1.json): 원본은 통과했으며 같은 길이의 바이트 변조, 최신 Prepare로 교체, LICENSE 누락, 다른 소스 변조는 모두 거부됐습니다. [로컬 검사 / local QA](LOCAL-QA.actual.public.v1.json)는 전체 Go race·vet·포맷·비밀 정보 검사와 독립적인 저장 결과 검증의 통과 기록입니다. The four real saved-only negative controls rejected each altered source witness; the original passed. The first failed CI run remains recorded rather than relabeled as a pass.
