# Source declaration references / 소스 선언 참조

This experimental Go package prepares links between retained public source
declarations and historical metadata. It is a maintainer module; an official
bounded runner and the first original-inventory run remain pending. The
[Korean plan](../../experiments/short-claim/PLAN-INVENTORY-60.ko.md),
[English plan](../../experiments/short-claim/PLAN-INVENTORY-60.en.md), and
[recipe](../../experiments/short-claim/source-inventory-recipe-60.json) describe
the current scope.

`Rebind(Input) (Report, error)` accepts four caller-pinned raw Go files, stored
root/component descriptors and the declared Go version. It uses AST names and
closed registry shapes to find declarations. It reproduces raw spans,
historical formatting/token digests and bundle recipes; it does not execute
registry callbacks or candidate functions. A matching descriptor is a syntax
and digest correspondence, not a resolved `types.Object` or runtime dispatch
proof. Transitive dependency completeness and caption meaning remain pending.

Whole enum declaration blocks may have multiple component IDs. Every original
relation is preserved. TypeSpec, FuncDecl and whole value GenDecl use their
historical node recipes; physical positions ignore `//line` remapping. Typed
standard imports remain package-wide infrastructure metadata.

The fixed policy is no cache. Physical file-hash/parse/format/normalization/
bundle attempted and returned counters are separate from completed root and
component digest comparisons. Fixed errors retain partial counters and prior
completed roots. The module uses slices without its own maps or locks; this is
an implementation fact, not a measured speed or memory improvement.

Tests use independently authored synthetic strings and expected token streams.
They never read the retained four original files or execute the original
inventory. The caller's `GoVersion` label alone does not prove the actual
compiler: the later runner must bind Go1.27.1 build metadata and the binary,
source, input and execution-plan bytes before original use. It also owns
bounded/canonical IO, output reservation, closed cardinalities and the ledger.

이 모듈은 학습 자료의 설명이 어떤 공개 구현을 가리키는지 연결하는 준비다.
설명이 실제 구현에 충실한지, 그 구현이 요청을 만족하는지는 별도로 검토한다.
현재는 실행기·봉인·원본 검증 전이며, 라벨·역할·학습·새 가중치·비용 절감
성과를 만들지 않는다. 사람 승인 절차를 추가하지 않고 기존 CI 병합 조건을 따른다.
