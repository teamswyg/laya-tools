# Whole-component role execution66 · preparation

`riido-roleplan` is a maintainer Go tool that keeps requests, candidates and helpers sharing a source together when assigning train, validation and calibration roles. It prevents closely related code from leaking across development splits. It is optional for ordinary `riidolaya` and Codex use. This publication contains **zero original role assignments and zero fits**, keeping drafts separate from the frozen plan for one execution after automatic CI.

The current [v2 plan](execution-plan-66.v2.draft.json) fixes ASCII seed `1729`,17 whole groups,16 groups containing stored known truth,72 requests and216 candidate positions. All21 unknown parents remain unknown; unknown-only group64 is train provenance. Known groups use largest-remainder 3:1:1 allocation with the existing9/3/3 floors. Both stored-truth coverage sets contain the same16 groups and add no stronger condition. Actual groups surviving loss masks require a separate check.

Build from automatically CI-verified source using Go1.27.1 / CGO0 / trimpath. Replace placeholders with actual paths and SHA values:

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o riido-roleplan ./cmd/riido-roleplan
./riido-roleplan --stage prepare \
  --plan execution-plan-66.v2.draft.json --plan-sha256 PLAN_SHA256 \
  --input experiments/short-claim/whole-group-preparation/membership-proposal-65.json \
  --repo-root . --source-root . --out NEW_PREPARATION_DIRECTORY
```

Prepare calls no original role function. Execute requires a **new frozen plan** binding exact sources, inputs, build and binary, plus one fixed `--ledger`. The [first frozen plan](execution-plan-66.v2.frozen-1.json) is a new file pinning the local binary built with Go1.27.1 / darwin-arm64 / CGO0 / trimpath / buildvcs=false. It remains unexecuted. The controller will verify this PR's quality pass, merge and exact Git source/input tree before its one execution. A binary for another platform is not the same execution. Keep the disabled draft unchanged. Reserve new ledger/output files before execution and retain refusals and counters. No seed search, group splitting, mask changes or automatic retries occur. This is an automatic verification distinction within existing authorization, not an additional human approval process. A caller can bypass a per-ledger limit by choosing another file; the controller must preserve one official invocation path.

The declared Go source closure is CLI main/loader, role.go and the additive source-identity helper, plus go.mod/go.sum. Compiled text is compared with source SHA values; this does not independently prove compiler trust. Git commit/blob freezing remains the controller's responsibility. Only exact known test filenames are exceptions; additional Go runtime sources, unlisted Go tests and Go-file symlinks are refused. This filename check is not a hermetic-build proof covering `.s`/`.syso` files or every build-environment input. The inspected frozen directories contain only expected Go sources, with the actual binary digest bound separately. The tool bounds each input file to8MiB and output to1MiB and uses a256MiB Go heap soft limit. These are neither OS RSS measurements nor an RSS hard cap.

The [v2 handoff](HANDOFF-66.v2.en.md) and [ledger](PREPARATION-LEDGER-66.v2.json) record private/shared race checks of10 named synthetic tests and vet passes. Before original role calls, root reviewed source, plan bindings, reservations and fixed diagnostics. Code reading found that the original v1 allowlist omitted existing archive tests; the [v1 ledger](PREPARATION-LEDGER-66.v1.json), source `.go.txt` files and both layout drafts remain historical. The earlier setup failure remains in that ledger. Test elapsed times are not benchmarks.

A [separate reviewer](INDEPENDENT-FINDINGS-66.en.md), not the CLI author but previously exposed as the62 module author, performed a nonblind review. One read-only verifier checked27 files/8,908,289bytes; four separate toy error tests passed race and vet. No current frozen-source/plan/binary execution blocker was found. The exploratory read failure and guard/compiler/per-ledger limits remain in the [independent ledger](INDEPENDENT-LEDGER-66.json). Original role APIs and CLI execution remain0.

The tool verifies the declared stored graph; opaque hashes do not prove all external edges absent. It grants no source/text certainty, different authoring origin or training readiness. The separate2400 protected-final target is not a minimum for every development fit. Source/group/role/truth/review metadata never become learned features. Zero separate model processes does not mean zero ordinary AI-assisted collaboration cost; that cost is unmeasured.
