# Standalone reader74 실행 전 독립 읽기 검토

고정 draft/소스/바이너리 범위에서 실행 전 blocker를 찾지 않았다. 검토자는 reader/controller 작성자가 아니지만 기존 codec/projection 검토에 노출된 nonblind AI 보조 검토자다. 입력 snapshot이나 compact 본문은 읽지 않았고 worker/controller/helper/테스트/Features/Fit/모델도 실행하지 않았다. 실제 데이터가 검사에 통과했다거나 두 결과가 이미 같다는 승인이 아니다.

새 JSON 경로는 기존 ImportJSON을 부르지 않는다. 첫 pass는 numeric token의 타입·범위·유한값·각 열 count·null과 CSR offset/행 길이를 확인한다. 네 dataset 전체의 metadata+typed-array 합계가64MiB 이내인지를 검증한 뒤 두 번째 pass에서 배열을 만들며, 같은 source SHA/shape/metadata여야 반환한다. 모든 객체에서 escape decoding 뒤 exact duplicate key를 거절한다. 작은 typed control은 lowercase-key 중복도 먼저 거절한다. Compact는 기존 Decode의 정확한 source pins를 유지한다. 양쪽 Summarize는 동일 구조/finite 검사와 strict metadata 재직렬화 byte equality를 적용한다.

Digest는 네 dataset의 여섯 열 모두를 포함한다. Dataset/column 위치·count·nil flag와 원래 순서의 int64/uint16/FP64 bits를 기록하므로 null/빈 배열, -0, 값·순서 차이를 숨기지 않는다. 전체 Summary 비교에는 원문 source SHA, exact metadata SHA/길이, owned payload와 각 column SHA도 포함한다. SHA 기반 검사이며 기존 codec 왕복을 반복하는 것이 아니다. 새 digest를 실제 데이터에서 계산하지는 않았다.

Controller는 source/binary/input pins를 먼저 검증하고 Frozen만 false→true로 바꾼다. 새 exclusive attempt, frozen plan, fsync된 controller prefix와 각 child intent/directory를 Start 전에 만든다. JSON 다음 compact 각각 한 child이며 자동 재시도 경로가 없다. 각60초 후 process group SIGKILL과 Wait로 연결된다. 실패/불완전 출력은 유효 prefix 또는 raw file refs로 남고 종료·timeout·overflow를 성공으로 바꾸지 않는다. Worker 역시 각 단계 호출 의도/반환 prefix를 별도로 fsync한다. 중단 시 prefix는 마지막 내구성 있는 기록이며 실제 함수 진입 횟수의 완전한 증명은 아니다.

운영 경계 하나를 명확히 해야 한다. Controller는 child 실패나 digest mismatch를 `completed_two_children_failed_or_mismatched_no_approval`로 기록한 뒤에도 종료코드0을 반환한다. 종료코드는 **기록 완료**를 뜻할 수 있으며 비교 성공은 `BothCompleted && ArrayAndMetadataDigestsEqual`과 final state로 판정해야 한다. 이 계획에서 기록 상태를 직접 확인한다면 blocker가 아니다. Shell exit0만으로 동등성 성공을 판단하면 안 된다.

Go1.27.1/Darwin arm64/CGO0/trimpath와 두 binary path/dependency identity,20개 source/plan/test/binary byte/SHA pins를 읽기 검증했다. Source hash와 buildinfo는 출처 결속이며 hermetic compiler 증명은 아니다. 기존 합성8 tests의 설계만 읽었고 저자의 test1/vet1/build2를 이 검토자의 실행으로 세지 않았다.

두 입력을 controller가 먼저 hash하므로 warm-file, JSON→compact 고정 순서의 한 표본이다. OS real/user/sys/RSS/footprint는 worker의 source/binary 확인·읽기·검증/digest·출력까지 포함한다. Retained64MiB는 tokenizer/scratch/slice headers/allocator/encoded input/전체 RSS를 포함하지 않으며 Go256MiB도 soft target이다. Kill 뒤 Wait와 사전/사후 처리 때문에 controller 전체120초 hard return을 보장하지 않는다. GPU/LLM 비용·모델 효용·128/512 확장·causal 형식 이익·학습/최종 성공은 이번 범위 밖이다.
