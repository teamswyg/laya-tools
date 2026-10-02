# Whole-component role execution66: read-only review

The frozen code, plan and binary are consistent. **No concrete code or pin error blocking the original execution was found within this scope.** This is not execution or training-readiness approval. The controller remains responsible for the PR's successful CI, merge and Git source/input tree verification. No original prepare/execute, role API or fit ran in this review.

The reviewer did not author the66 CLI, loader or frozen plan, but did author the earlier public62 roleplan module and has seen development material. This is a separate CLI code review with nonblind exposure, not an independent source origin, compiler-trust proof or new semantic assessment.

## Frozen artifacts

|Artifact|Bytes / SHA256|
|---|---|
|Frozen66 v2 plan|9,559 / `f6065a1285592dfa0617415668d07e332cb10a663ec8c78f13c17c2fafa1da53`|
|Public main.go|11,948 / `b13b66b422db65982a3cc8de18f8a0107b5ef0bc8bab5fab004d778a242a97e3`|
|Public loader.go|22,518 / `be769eb783cf09744820ee556c8bbb57adae252933c998ae423e75bf695770a7`|
|Public main_test.go|13,663 / `79701206640146487b79253a23907c5e3c5c6ca08e5e14d21f7a812115df0bab`|
|Original role.go|16,123 / `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`|
|Additive source-identity helper|707 / `f1fb50eb693baa06cd1e9df7ad57275ee557a6bcc17439befbaf9617a6082224`|
|Unexecuted local binary|4,438,626 / `32aa2b34770d3b1ce2aab5e6c87429cdef7af3a6451d5bad9faba644b059baa4`|

A standalone Go verifier checked canonical frozen plan bytes and20 pins: membership65,8 original inputs,3 support files,2 binding-evidence files and6 implementation files. Repeated pins for a file serving different purposes were each checked. It also read5 exact source copies for the toy fixture and the binary. Its27 direct bounded reads totaling8,908,289 payload bytes describe **the verifier's own reads**, not CLI latency or memory. One buildInfo read and two directory metadata checks were separate.

Membership metadata preserves17 whole groups/16 known-containing groups,72 parents/216 candidates,34 known/17 no_answer/21 unknown,453 members and1,922 relationships. Metadata checks covered the original17-group list, including8 and64, the complete parent-index bijection and both known/no_answer populations covering the same16 groups. No original membership digest or role ordering was computed.

## Execution boundary and preservation

`validatePlan` fixes ASCII seed1729/hex31373239, the existing3:1:1 encoding/algorithm/domains and9/3/3 floors, unknown preservation, the identical stored-truth coverage sets, one-attempt/no-retry and fit/model/paid/final0 with readiness/features/masks=false. It verifies SHA, canonical JSON and strict shape. Execute refuses an unfrozen plan before input/role dispatch. Stored-truth coverage does not approve future loss eligibility or semantic readiness.

`loadBound` retains the exact membership65 bytes, verifies source/evidence/support/implementation pins and decodes **those originally hashed membership bytes**. The implementation loop's `b` is scoped locally and does not overwrite the original input. The Go import `github.com/teamswyg/laya-tools/internal/roleplan` connects source and module identity. Original role.go is unchanged; only the helper was added.

Reading binary SHA and buildInfo confirmed Go1.27.1, main module `github.com/teamswyg/laya-tools`, the command path, CGO0/trimpath, darwin/arm64 and absence of VCS settings. There are0 external dependency modules. The binary contains all3 exact CLI main/loader/role source byte sequences. These witnesses help establish source/metadata correspondence, but embedded text does not prove all executed instructions or independently establish compiler/toolchain trust.

After `main.go:253`, execute reserves a caller-fixed ledger with O_EXCL, writes/syncs it and creates a new output directory/file before one VerifyMembershipSet call and at most one Assign call. The same ledger/output is not reused. Membership refusal retains assign_attempts0; assignment refusal retains1. There is no fallback, new seed, component split or retry. A remaining ledger is consumed even after failure. A caller choosing another ledger is outside this per-ledger constraint; it is not a host-global enforcement mechanism.

Post-dispatch ledger/output persistence failures emit fixed errors and actual membership/assign prefix counters on stderr, never success. A ledger rewrite failing after truncate can leave incomplete JSON. Output write/Sync failure or process interruption likewise does not guarantee a completed report; the controller must not substitute success for failure. Power-loss/filesystem crash recovery was not tested.

## Bounds and current limits

Regular input files are bounded at8MiB, plans/results at1MiB and executable reads at64MiB. JSON controls cover depth128,128 object keys,512-byte keys,65536 array entries, duplicate decoded keys, unknown fields and canonical bytes. Pins use the same payload that was read and checked for bytes/SHA. GOMAXPROCS1 and a256MiB Go heap soft limit are not RSS hard caps or performance results. readBounded uses a limit+1 sentinel read and claims no separate aggregate file-payload or process-memory hard budget.

The regular-file reader follows symlinks to regular targets. It is not filesystem confinement against hostile root/parent-directory replacement; the existing controller boundary requires owned, frozen input/source roots. Pins detect altered source bytes but do not establish external filesystem or compiler trust.

The source allowlist checks exact known Go files and test exceptions. A **nonblocking wording limit** is that `checkClosedSourceRoot` examines only `.go` filenames: it is not a general closure mechanism refusing `.s`, `.syso` or every build-environment input. Both current source directories contained only expected `.go` files, and frozen binary/pins matched. Read README's broad “additional runtime files refused” within the Go source filename guard's scope. No new condition or code change is required here, and this is not promoted to a current execution blocker.

Source/text uncertainty, omitted external relationships, one synthetic caption pipeline, uncleared source diversity and training_ready=false remain. Roles66 are stored-truth development allocation only. Role-specific eligibility after67 masks and same-scope utility/headroom belong to a separate frozen step. No new floor, per-experiment2400/fresh15 requirement or all-prototypes-in-all-roles rule was added.

## Checks and ledger

The standalone metadata verifier ran once: success1/failure0. It read original text/JSON/bytes, code and binary metadata only. Original Assign, OrderDigest, AllocateCounts, MembershipDigest, VerifyMembershipSet, original CLI, model/fit/labels/roles execution all remain0.

Four necessary error controls ran against exact source copies in a private fixture: fixed codes/partial counters for short or failed writes; preservation of a consumed reservation after local persistence failure; rejection of empty/oversized/directory inputs with the regular-symlink caller limitation; and toy plan-SHA early refusal without argument disclosure or output reservation. A toy CLI early-error `run` was invoked once and never reached original input/source/binary dispatch. The original CLI author's10 tests were not rerun here.

One synthetic race check passed4 top-level tests/4 pass events, failure0, package1.603s. One vet passed; only the new audit source/test was formatted once. Test timing is not a benchmark. Different-author error controls do not make this a blind or source-origin evaluation.

Of22 reading commands, one exploratory command failed after guessing two nonexistent private directories. The actual required artifacts were present; the failed discovery is retained in the ledger. One long parallel output was truncated, and needed code ranges were reread. An empty range after EOF was not treated as missing content. Failure outputs were kept as private evidence; frozen files were unchanged.

Shared edits, Git/publication, original role/order/metadata CLI execution, features, new labels, fits, model/paid calls and protected-final reads all remain0. Ordinary AI-assisted collaboration cost is unmeasured. Safe aggregate mechanics are in `MECHANICS-66.json`; invocation/check counts are in `LEDGER-66.json`.
