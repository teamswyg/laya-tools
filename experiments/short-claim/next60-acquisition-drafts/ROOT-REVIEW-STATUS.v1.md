# 후속 검토 상태 / Subsequent review status

봉인된20개 초안은 실행 결과가 아니며 적격0개를 유지합니다. 독립 원문 검토에서 다음 수정 과제를 확인했습니다. 원래 초안은 덮어쓰지 않고 검토 결과와 이후 변경을 별도 버전으로 남깁니다.

- nil decoder 초안의 오답 설명 중 typed-nil Result의 panic 주장은 원문과 맞지 않습니다. 현재 NewDecoder는 metadata 초기화 전에 이를 거부하므로 구체적인 잘못된 구현과 검사 입력을 다시 정의해야 합니다.
- exclusive-write의 경쟁 입력은 미수정·정상 구현 모두에서 첫 대상 취득 직전에 같은 독립 스케줄로 실행해야 합니다. 기존 Exists 뒤에만 경쟁을 거는 검사는 정상 O_EXCL 경로에서 경쟁 자체를 생략할 수 있습니다.
- afero Sub의 잘못된 후보 설명에서 `path.Clean("../a")`가 유효한 경로로 바뀐다는 전제는 틀립니다. 빈 문자열이 `.`으로 바뀌는 경우와 구분하고, 실제로 잘못된 구현을 다시 고정해야 합니다.
- INI 입력의 문자열은 본문이 아니라 파일명입니다. 본문 검사에는 `[]byte` 또는 reader adapter를 고정하고, 기존 옵션과 오류 동작을 함께 검증해야 합니다.
- native IPNet의5개·OrCompose의4개 예정 기대값은 원문상 일치한다는 정적 검토가 있습니다. 실행·공통 adapter·전체 closure·init·고지 검증은 아직 미완료입니다.

첫 native 관측 준비는 기존 후보 수를 유지합니다. IPNet3후보×5입력=15, OrCompose2후보×4입력=8로 총23개 예정 관측입니다. 세 후보로 통일하거나 입력을 독립 요청으로 세지 않습니다. 훈련·역할 배정·정답 라벨 확정·보호 final 접근은0이며, 다음 실행은 별도 고정 계획의 조건이 갖춰져야 합니다.

**English:** The frozen 20 drafts remain unexecuted and unqualified. Source review contradicts a proposed typed-nil panic caption: the original NewDecoder rejects that Result before metadata initialization, so a concrete faulty implementation must replace the claim. The exclusive-write race must trigger independently before the first destination acquisition on both baseline and correct routes, including the O_EXCL route. `path.Clean("../a")` remains invalid, so that candidate caption needs correction. INI strings name files; content fixtures need a frozen byte-slice or reader adapter. Native IPNet's 5 and OrCompose's 4 proposed Wants agree statically with pinned source, but common adapters, complete closure, initializers, notices and actual observation remain pending. Preparation preserves 3×5 plus 2×4=23 scheduled observations; there is no extra candidate, training, role assignment, truth label or protected-final access. The detailed independent review preserves the draft seal and must be read before implementation.

[독립 검토 / Independent review](../next60-contract-review/REVIEW.ko.md) · [English](../next60-contract-review/REVIEW.en.md)
