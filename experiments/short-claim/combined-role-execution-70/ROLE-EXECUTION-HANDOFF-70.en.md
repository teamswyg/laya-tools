# Preparing one role execution for all 76 requests

The worker (`main.go`, `loader.go`, `runner.go`) and synthetic tests are ready. It was compiled against the public role module, but **original-data prepare/execute stages and public-module function calls remain at 0**. Use `execution-plan-70.draft-2.json`; its `execution_authorized` field is false. This is not an assignment result, semantic/diversity approval or a declaration of fitting readiness.

## Exact files and pins

| File | Bytes | SHA256 |
|---|---:|---|
| `execution-plan-70.draft-2.json` | 6,964 | `6988545f37f3fccc430bbc6d78e53a8946b3a25f06925dff21fec5b96c713be6` |
| `preparation-receipt-70.2.json` | 3,024 | `4ba3bc820f92a77e716ddeeaa71bb28834f92229513ffeef37bd239ea6ed8ed5` |
| `combined-role-exec` | 4,563,106 | `f4912efce64a801fe0bc7b3acda74593316980da5dab6c20fbe2507ba17381af` |
| `main.go` | 8,566 | `e7ffba1ce0df1c429d54ec1b94eb0c122646826ae2b7a4ff3e247d47ae21a026` |
| `loader.go` | 26,648 | `407891f136c1a769da89f0576c99370518f3b5f15436c3257d685c86c26e008d` |
| `runner.go` | 13,589 | `66d7d7744cf15c8f15529812e392a3465545e61f8e4f25423322e0d6ffe9dc9d` |

Proposal70 is 2,694,968 bytes, SHA `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21`. `input/combined70.json` is an exact copy. `input/future70draft.json` and 12 original evidence files under `input/evidence/` bring the closed input list to 14. The receipt records relative paths, sizes and hashes, without original private paths. Original inputs were not rewritten. The worker has not yet loaded or validated these copied full inputs.

Draft1's source, binary, plan and receipt remain in `frozen-draft-1/`. Draft2 adds direct binding to original56b saved truth separately from the67 masks. The archived previous loader SHA was checked against its draft1 pin. Neither draft was executed, so no original failure, assignment or label result was changed.

## Binding and dispatch order

The CLI accepts only `--stage prepare|execute`, `--plan`, `--plan-sha256`, `--source-root`, `--module-root`, `--input-root` and `--run-root`. It accepts no actor-supplied features or source callbacks. Unknown plan fields and flags are refused. Diagnostics contain fixed error names and numerical counters.

Before dispatch, it binds the externally fixed plan SHA and canonical JSON; actual Go1.27.1 / CGO0 / trimpath binary SHA; all three compiled worker sources against their disk manifest; and the frozen public role module sources. Worker/module roots are runtime options and source/input references are relative paths. Each of the 14 metadata files is **decoded from the same bytes that passed hash verification**. The 4 MiB per-file / 16 MiB total limits bound maintainer metadata work. They do not change the existing 512-byte / 32-normalized-word runtime request/caption contract or execute Normalize/Validate here.

It checks all original 17 group JSON values and raw-subtree hashes against 65, retaining new groups 72/74 and the original append order. It checks each group's saved parent ID/index or new 72..75 index set, all 76 parent/group bijections and known flags against saved truth. The original 72 requests, candidate order/text, code/bundle pins, acceptable sets, null labels and masks bind directly to 56/56b/67/66. Original 66 role rows remain in group order; an array maps `original_parent_index` to the original JSON row. New requests retain 68v4 contract and finite QA order/captions/acceptable sets/proposed labels/masks/null weights. It also checks all 15 retained source 68 Go/module/LICENSE assets by disk SHA, without importing or executing those packages.

`prepare` writes a metadata-only result to a fresh directory and calls no public role functions. Prepare has not been executed yet. `execute` requires a separately frozen execution-authorized plan. It reserves and Syncs `role-attempt-70.json` with O_EXCL, then reserves the receipt/result under fresh `role-attempt-70/` with O_EXCL. A consumed ledger is not reset or refunded. This is one execution-root ledger, not host-global authorization.

Dispatch then checks the public module's compiled-source identity, invokes `VerifyMembershipSet` **once**, and invokes `Assign` **once** with ASCII seed1729. Before every callback, it Syncs begun-call counters into result and ledger. It calls no source/caption observers, Features, Project, Baselines, models or fitting code. Module errors and panics become fixed refusals, never successful labels. Assignment refusals preserve module counters and known fixed failure codes. Write/sync failure diagnostics also preserve identity/verification/assignment attempts and completions and the returned parent-role count.

## Conditions and limits

The frozen 62 algorithm source SHA is `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`. The manifest also pins its provenance helper, existing tests/archive tests and original module go.mod/go.sum. The worker uses the unchanged public-module 3:1:1 largest-remainder recipe, existing 9/3/3 floors and six predeclared stored-truth coverage requirements for answerable 18 / no_answer 16. Counts such as 11/4/3 are not actual assignments before execution. Original 72 roles remain historical provenance and are not copied into the new 76 assignment.

On success, parent-role mappings are returned in original combined parent-index order. Per-role reports retain parents, positions, known/no_answer/unknown/null counts and positive/negative/masked counts. Mask-qualified groups and comparison with original 9/3/3 are separate descriptive outputs. Positive-bearing counts add no gate. Hard negatives and no_answer candidates remain intact. Coverage failure atomically refuses roles: do not change the seed, membership, groups or rules, or search for a favorable seed. Source 68 labels remain narrowly proposed and actual fitting weights stay null.

CPU 1 and the 256 MiB Go soft heap limit are settings, not RSS measurements or a hard RSS cap. The worker uses a 300-second cooperative checkpoint timeout but cannot forcibly interrupt a synchronous module call in progress. **The parent OS controller must enforce the 300-second hard wall.** A forced stop leaves the last durable begun-call checkpoint. Results are capped at 64 MiB. No new speed, savings or memory improvement was measured.

The private build uses a local replace for the external public module. Its preparation go.mod and seal helper contain private build paths and must not be published unchanged. Runtime references use relative paths and root options. A later shared public port must **refreeze** module/build identity and source-manifest paths for its actual repository layout. Keep all 14 input/proposal hashes, role recipe, seed and coverage unchanged. Embedding and hashes bind source provenance; they are not independent proof of a trusted compiler.

## Checks and the parent's next step

Four race-test invocations passed 13 top-level named-test executions plus 4 nested subtests, using only synthetic pure callbacks. Three vet invocations, three binary builds and two source/input seals passed. Full original 76 loader invocation, worker prepare/execute, real module provenance/verification/Assign, source APIs, Features/Project and fitting remain at 0. Tests used no original corpus or actual assignment outcomes.

The parent can independently review draft2/source/loader, freeze source/binary/inputs and the actual execution plan, and then enable one full 76 assignment with 0 retries. Check allocation/checkpoints/refusal/coverage before separate 76 input validation, Project and fitting preparation. Source and authoring diversity remain uncleared. The ≥2,400 protected-final-per-domain objective is separate; this adds neither a positive-bearing role minimum nor a 2,400-per-fit condition. Shared-repository edits, publication and model calls remain at 0.
