# 실제 metadata 연결까지 확인한 Go projection 사전 검토

수정된 v4의 실제 metadata join과 필수 source/input/binary 핀 검증에서 미해결 concrete blocker는 없다. 원래 Prepare·Validate·Normalize·Project·Features·행동 API·역할 배정·fit는 호출하지 않았다. 이는 연결과 실행 준비 코드의 검토이며, 새로운 정답 관측·모델 결과·학습 준비 승인·수치 gate가 아니다. 먼저 불변 receipt v3를 root에 전달했고 상세 문서·장부는 후속으로 작성했다.

원래 role66의72개 row는 whole-group 순서이며, original index0..71이 각각 한 번 존재하지만28개 row는 배열 위치와 다르다. 이전 v3 join은 canonical probe index를 role array 위치로 사용해 실제 i12에서 `projection69_role_binding`으로 거부했다. 이 readonly metadata join1회와 expected-refusal race test1.539초 통과를 분리해 기록했다. test PASS가 join 성공을 뜻하지 않는다. root의 별도70 metadata 실패도 별개 기록이며 여기서 지우거나 합치지 않았다.

수정된 `roleRowsByOriginalIndex`는 `[72]int`를 -1로 초기화하고 row count, index 범위·중복·빈 필수 field·missing을 거부한다. 이후 원래 parent index로 row를 조회하며 기존 ID/truth/group/role 검사를 유지한다. 역할 JSON/DTO 배열 자체는 정렬하거나 다시 작성하지 않는다. actual corrected join1회에서 역할 DTO의 before/after JSON bytes가 동일했다. 부모72·후보216·known34·no_answer17·unknown21, role candidate132/48/36, whole groups11/3/3, known rows90/36/0, zero-weight18/1/0, unknown42/12/9, positive-weight groups9/3/0이 유지됐다. 마지막 수치는 fit 역할에서 어떤 알려진 후보든 가중치1을 가진 그룹 수이며 positive-label group 수나 새 하한이 아니다. 새68 요청4개는 추가하지 않았다.

v4 draft7594bytes SHA256 `187c19191027e6e07bf04096a04f266a4576df6bc672f99fde34f3513f3f99e1`, binary4301714bytes SHA256 `5f9beed89d8449a927a5fdf90c29f0183934f07dc07b163fd195c3ad8a014548`, author handoff v3 16061bytes SHA256 `870a88d1c2fcd78f13e0244db973d4a44e7007b1731505060cdbc5d94fd18455`가 정확히 일치했다. 실행pin30개/8158159bytes, public transitive Go14개+go.mod, handoff의79개 자산, 이전 primary4개 pin을 확인했다. 이전 v3→v4 plan의 변경은 join source pin과 binary pin뿐이며 입력8개·공유 source·다른 driver source·scope·budget·attempt directory·frozen=false는 그대로다. compiled metadata는 Go1.27.1·darwin-arm64·CGO0·trimpath true·exe·원래 module의 local replace binding과 일치한다. physical replace 경로는 private로 유지하며 hermetic compiler 증명을 주장하지 않는다.

독립 exact-copy 검사 package에서 toy10개 top-level/20개 subtest와 corrected actual metadata1개 test가 race1.630초에 통과했다. callback은 손으로 만든 sparse 값이고, 실제 join test도 decode/readPin/join/lookup만 호출했다. package vet1회와 stdlib pin helper vet1회가 통과했다. 원래 input/role/mask 파일은 수정하지 않았다. semantic annotations를 runtime feature로 사용하거나 label/no_answer/unknown을 바꾸지 않았다.

pin helper의 첫2회는 `metadata-only-preflight/...` public-safe logical ID를 driver base의 물리 위치로 잘못 가정해 거부됐다. author가 별도 private namespace의 정확한 대응을 전달했고, 기존 artifact를 변경하지 않고 세 번째 검사에서79개가 모두 일치했다. 이 검토기 거부2회와 own patch-context 실패1회는 보존했으며 원본 모델/네이티브 실패나 retry로 세지 않는다. 마지막 helper는 고유 정규파일104개/25541847bytes를 읽었다. 동일 binary build-info 읽기와 directory listing은 별도다. 추가 선택적 참조 검사를 새 실행 의존조건으로 만들지 않았다.

이전 v1/v2 receipt와 findings는 변경하지 않았다. v2의 source/pin/toy PASS는 실제 metadata 연결을 실행하지 않은 범위였으며 이번 순서 오류를 놓쳤다. 그 과거 PASS를 corpus projection 성공이나 실행 준비의 완전한 보증으로 확대하면 안 된다. 구체적인 한계는 별도 COVERAGE-LIMITATION v3 한영 문서와 old-refusal/current-join mechanics에 남겼다.

Root가 exact pins와 이 검토 scope를 확인해 별도 once-only original projection을 동결·실행할 수 있다. receipt 존재만으로 QA의 모든 의미가 검증되는 것은 아니다. CPU1은 Go scheduler,256MiB는 soft Go heap,64MiB는 보유 payload와 인코딩 후 serialized cap이며 OS RSS hard cap과 다르다. controller kill/저장장치 실패 시 외부 예약·marker·장부를 함께 보아야 한다. 검토자는69 driver/Project 작성자가 아니나62/67 작성과 기존 QA 노출로 비블라인드다. whole-membership 근거는 재사용했고 모델·학습·새 seed/roles·protected-final·Git·공유 수정·게시·readiness/diversity 승인은0이다. AI 협업 비용은 측정하지 않았다.
