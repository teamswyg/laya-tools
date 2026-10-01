# 120 public Go task candidates: sources for executable contracts

[한국어](public-go-acquisition-53.ko.md) · [Canonical JSON](../benchmarks/training/public-go-acquisition-53.json) · [2,400-request acquisition plan](golden-set-acquisition.en.md)

We authored **120 behavioral request candidates from eight public Go repositories**. We read each pinned revision, actual source files, function/type declarations, root LICENSE, file-origin signals and NOTICE presence. **Zero candidates are currently execution-eligible.** The count is authored request drafts, not model attempts, difficulty answers, training labels or independent statistical samples.

This expands source breadth so small development trials do not carry the performance claim. All material has already been read during development and is excluded from protected final. It does not count as acquiring the separate 2,400 requests per claimed domain or the training, selection and calibration datasets. These are sources for checking actual Go coding completion, not labels for repository selection or task decomposition.

## What was acquired

| Public repository | Candidates | Inspected pinned revision | Actual root license |
|---|---:|---|---|
| [spf13/pflag](https://github.com/spf13/pflag/commit/c966cfef47379dcb01e7929504d66d94b540945b) | 15 | `c966cfef4737` | [BSD-3-Clause](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/LICENSE) |
| [spf13/afero](https://github.com/spf13/afero/commit/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e) | 15 | `eb6a92826ea5` | [Apache-2.0](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/LICENSE.txt) |
| [spf13/cobra](https://github.com/spf13/cobra/commit/adbc8813901bba65827259daa8e22ff94ec1f30e) | 15 | `adbc8813901b` | [Apache-2.0](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/LICENSE.txt) |
| [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp/commit/fd004584a46724fae09e2f21d7c382e15c893f42) | 15 | `fd004584a467` | [MPL-2.0](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/LICENSE) |
| [go-viper/mapstructure](https://github.com/go-viper/mapstructure/commit/52aa5c6dc1d27226460807054ca2107b2d54fb2d) | 15 | `52aa5c6dc1d2` | [MIT](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/LICENSE) |
| [go-ini/ini](https://github.com/go-ini/ini/commit/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08) | 15 | `e2db55b0e088` | [Apache-2.0](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/LICENSE) |
| [dustin/go-humanize](https://github.com/dustin/go-humanize/commit/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e) | 15 | `a1b4e66b9a6d` | [MIT](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/LICENSE) + number.go WTFPL v2 |
| [google/uuid](https://github.com/google/uuid/commit/2d3c2a9cc518326daf99a383f07c4d3c44317e4d) | 15 | `2d3c2a9cc518` | [BSD-3-Clause](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/LICENSE) |

The recorded inspection evidence covers 97 original files, including LICENSE and go.mod. Edit anchors span flag parsing, command trees, virtual file systems, HTTP retries, struct decoding, INI configuration, number/time formatting and UUIDs. Fifteen candidates per repository is a convenience allocation for initial collection, not an estimate of real user task frequencies.

Every candidate has exactly the status `source_candidate_only`. Each records `contract.status=pending`, `execution_eligible=false`, `training_label_eligible=false`, `final_eligible=false` and `model_executed=false`. Pending license review is a separate field. Pending contract work does not confer execution eligibility.

The original prompts are **newly authored English behavior requests**. Both language tables summarize the same requests; use the ID to locate the exact prompt and proposed acceptance criteria in JSON. We did not copy upstream issues, comments or test bodies, or increase the count by renaming the same bug. This inventory contains no source-code bodies, model weights, private material, real user requests, credentials or raw execution traces. It is not a list of demonstrated defects in the current implementations.

## Gates between a candidate and an executable contract

No candidate becomes a fixed model-execution task until these gates pass.

1. Freeze unambiguous input, output, failure and compatibility policies. Revise or exclude already-satisfied, oversized or solution-template duplicates.
2. Materialize the smallest needed closure of editable files, support code, platform files, dependencies, toolchain and independent tests. Current file lists are **inspected edit anchors**, not a compiled closure.
3. Preserve the complete closure SHA-256, revision, LICENSE/NOTICE and copied/derived-origin notices. Git blob SHA-1 fields below trace upstream provenance and are not complete closure hashes.
4. Verify that an independent completion checker rejects the unchanged baseline, accepts a separately authored correct reference and rejects known incorrect implementations. Existing tests or model-authored tests alone do not establish completion.
5. After applicable checks and public-data review, seal profiles, attempt limits, time, concurrency and stop conditions in a separate execution plan. Keep reference/mutant verification separate from actual model outcomes.

The JSON acceptance criteria are **proposals**. No contract has been implemented or sealed, and baseline rejection, reference acceptance and mutant rejection have not been executed. No fast/standard/strong labels were assigned. Actual capability labels require requested-profile attempts with independent completion checks, process/failure status and complete usage accounting.

## Outstanding license work

Pinned original LICENSE files are linked in the table. No NOTICE file was found in the eight complete non-truncated repository trees. That does not establish absence of NOTICE requirements in support files, dependencies or copied origins. Our authored prompts and explanations are Apache-2.0; upstream licenses are not changed. This record is not blanket clearance of repositories, derivative datasets or model-training rights.

- **go-retryablehttp:** The root and Go SPDX headers are MPL-2.0. Specify source/executable distribution, modified-file scope, source availability, notices and dependencies for the actual closure. The [MPL text](https://www.mozilla.org/en-US/MPL/2.0/) and [official FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/) are recorded. Public availability does not make the imported material Apache-2.0-only.
- **go-humanize number.go:** Record a separate origin alongside root MIT. The [pinned gorhill gist revision](https://gist.github.com/gorhill/5285193/cc36e402754549e658cf1d7255fd84f3c284f063) actually contains the [WTFPL v2 text](https://gist.githubusercontent.com/gorhill/5285193/raw/f45156c00aace6db928e0d89660606e6bb500231/%7ELICENSE). Confirmed origin evidence and the still-pending combined provenance/distribution manifest are distinct facts.
- **Standard-library origin signals:** pflag text.go states copying from Go 1.23.4 flag.go; afero match.go states adaptation from filepath. The Go Authors copyright in afero ioutil.go and encoding/json copying statement in INI struct.go are also recorded. Original revisions and notice scope for the included portions need further review; these signals alone do not establish a violation.
- **Multiple copyright holders and dependencies:** Preserve/review the contributors in afero util.go/mem/file.go and Cobra's pflag/documentation dependencies. Apply [Apache-2.0](https://www.apache.org/licenses/LICENSE-2.0) modification, redistribution and notice conditions to the actual distribution scope.

Specific file signals are attached to candidates editing those files. Enlarging a closure requires reviewing support-code, dependency and test licenses for every candidate. Rights and contract review both remain pending.

## Connected groups and the existing seven candidates

Initially keep candidates from one repository together. Pinned Cobra [go.mod](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/go.mod) depends on pflag v1.0.9, so these repositories are conservatively connected. The inventory's pflag HEAD is not assumed to equal that dependency version. There are **seven provisional repository/dependency groups**, not seven proven independent final groups.

Further source, copied-origin, fork, shared support-code, solution-template and parent/child/sibling review can merge groups. Atomicity, resource bounds, ownership and clock injection can be semantically related across repositories. JSON mechanism tags aid that review; they are not unique difficulty answers. Repository-generalization evaluation separates whole repositories and keeps connected groups from crossing train, selection, calibration and final.

The historical seven at acquisition time comprised three `catalog-budget` and four `repo-preview` candidates. The [current registry](../benchmarks/training/public-task-candidates.json) adds one separate `taskoutcome-parser` candidate, making **eight candidates in three families**; the new parser has a prepared versioned verifier with zero model attempts. Actual keyword-task repetitions are recorded separately in [results 53](../experiments/task-outcomes/RESULTS-53.en.md). Do not overwrite the inventory JSON's historical seven-candidate relationship or the new 120-source count. New upstream anchors differ from existing laya-tools files, but budget, snapshot, filter and state concepts remain potentially related, with deduplication/group review pending. Preserve existing frozen requests and attempt records; do not add these 120 to historical executions or model-label counts.

## First ten contracts to develop

This is a proposed **development order for compact, concrete independent checks**, not measured model difficulty, training answers or authorization to launch inference. Feasibility and source-support review may reorder or exclude candidates.

| Order | ID | Behavior | Reason to inspect first |
|---:|---|---|---|
| 1 | `go53-humanize-ordinal64` | Humanize Ordinal64 | A single-file integer-boundary API allows a compact independent arithmetic specification. |
| 2 | `go53-pflag-map-snapshot` | pflag Map Snapshot | External mutation can verify ownership while a new API preserves the existing getter. |
| 3 | `go53-mapstructure-decoder-nil-config` | Mapstructure Decoder Nil Config | Input preconditions and errors can be checked through compact public calls. |
| 4 | `go53-cobra-active-help-lines` | Cobra Active Help Lines | Fixed strings can check completion-protocol line boundaries. |
| 5 | `go53-afero-iofs-sub-validation` | Afero Iofs Sub Validation | Independent inputs can check fs path grammar and error shape. |
| 6 | `go53-ini-section-delete-index` | INI Section Delete Index | Negative, out-of-range and valid indexes plus preserved state are directly observable. |
| 7 | `go53-uuid-canonical-parse` | UUID Canonical Parse | Fixed byte vectors can check the accepted grammar and preserve existing Parse compatibility. |
| 8 | `go53-uuid-sql-null-reset` | UUID SQL Null Reset | A prefilled receiver makes successful reset and failed-scan atomicity observable. |
| 9 | `go53-humanize-strict-comma-parse` | Humanize Strict Comma Parse | Independent counterexamples can cover integer grammar and numeric range. |
| 10 | `go53-pflag-canonical-network` | pflag Canonical Network | Fixed addresses can test an explicit network/host-bit policy. |

Larger state-transition, overlay, concurrency and documentation-dependency candidates require wider closures. Imported MPL material and copied-origin scopes must be resolved before the affected candidates become execution-eligible. While developing these ten, continue acquiring 120→240+ development signal sources and separately protected 2,400-request final sources; do not report candidates as measured labels.

## Complete list of 120 candidates

Every row is a development source candidate with no execution eligibility. Distinct family names do not prove independence. Pinned function/type line anchors, Git blob SHA-1, original English prompts and proposed acceptance criteria are in [JSON](../benchmarks/training/public-go-acquisition-53.json).

### spf13/pflag

Root BSD-3-Clause and Go Authors headers were inspected. text.go states copying from Go 1.23.4 flag.go, so original notices and closure scope need separate confirmation.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-pflag-normalization-collision` | pflag Normalization Collision | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.flag_identity` |
| `go53-pflag-unknown-token-report` | pflag Unknown Token Report | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.argument_provenance` |
| `go53-pflag-single-assignment` | pflag Single Assignment | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.assignment_policy` |
| `go53-pflag-parse-stop-callback` | pflag Parse Stop Callback | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.parser_control` |
| `go53-pflag-help-display-width` | pflag Help Display Width | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.display_columns` |
| `go53-pflag-sensitive-default` | pflag Sensitive Default | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.help_disclosure` |
| `go53-pflag-annotation-ownership` | pflag Annotation Ownership | [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.metadata_ownership` |
| `go53-pflag-bounded-count` | pflag Bounded Count | [count.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/count.go) | `cli.count_accumulation` |
| `go53-pflag-bounded-string-slice` | pflag Bounded String Slice | [string_slice.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/string_slice.go) | `cli.collection_budget` |
| `go53-pflag-map-snapshot` | pflag Map Snapshot | [string_to_string.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/string_to_string.go) | `cli.map_snapshot` |
| `go53-pflag-deferred-function` | pflag Deferred Function | [func.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/func.go), [flag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/flag.go) | `cli.side_effect_staging` |
| `go53-pflag-go-bridge-conflict` | pflag Go Bridge Conflict | [golangflag.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/golangflag.go) | `cli.stdlib_bridge` |
| `go53-pflag-text-error-value` | pflag Text Error Value | [text.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/text.go) | `cli.custom_value_error` |
| `go53-pflag-local-time` | pflag Local Time | [time.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/time.go) | `cli.time_location` |
| `go53-pflag-canonical-network` | pflag Canonical Network | [ipnet.go](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/ipnet.go) | `cli.network_canonicality` |

### spf13/afero

Root Apache-2.0 and several contributor headers were inspected. The filepath-derived signal in match.go and Go Authors provenance in ioutil.go require further origin/notice review. Preserve the multiple copyrights in util.go and mem/file.go.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-afero-iofs-sub-validation` | Afero Iofs Sub Validation | [iofs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/iofs.go) | `fs.subdirectory_contract` |
| `go53-afero-exclusive-safe-write` | Afero Exclusive Safe Write | [util.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/util.go) | `fs.exclusive_creation` |
| `go53-afero-atomic-replacement` | Afero Atomic Replacement | [ioutil.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/ioutil.go) | `fs.atomic_replacement` |
| `go53-afero-bounded-read` | Afero Bounded Read | [ioutil.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/ioutil.go) | `fs.read_budget` |
| `go53-afero-union-pagination` | Afero Union Pagination | [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.overlay_enumeration` |
| `go53-afero-copy-metadata` | Afero Copy Metadata | [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.copy_metadata` |
| `go53-afero-cache-invalidation` | Afero Cache Invalidation | [cacheOnReadFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/cacheOnReadFs.go) | `fs.cache_invalidation` |
| `go53-afero-cache-clock` | Afero Cache Clock | [cacheOnReadFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/cacheOnReadFs.go) | `fs.time_dependency` |
| `go53-afero-copy-on-write-tombstone` | Afero Copy On Write Tombstone | [copyOnWriteFs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/copyOnWriteFs.go), [unionFile.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/unionFile.go) | `fs.overlay_deletion` |
| `go53-afero-readonly-truncate` | Afero Readonly Truncate | [readonlyfs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/readonlyfs.go) | `fs.open_capability` |
| `go53-afero-regexp-rename` | Afero Regexp Rename | [regexpfs.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/regexpfs.go) | `fs.filtered_rename` |
| `go53-afero-mem-seek-contract` | Afero Mem Seek Contract | [mem/file.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/mem/file.go) | `fs.file_position` |
| `go53-afero-mem-subtree-rename` | Afero Mem Subtree Rename | [memmap.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/memmap.go) | `fs.tree_relocation` |
| `go53-afero-glob-limit` | Afero Glob Limit | [match.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/match.go) | `fs.pattern_expansion` |
| `go53-afero-symlink-lexical-base` | Afero Symlink Lexical Base | [basepath.go](https://github.com/spf13/afero/blob/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/basepath.go) | `fs.symlink_policy` |

### spf13/cobra

Inspected sources have Apache-2.0 headers. cobra.go mentions inspiration from other tools; this is recorded without treating it as proof of copying. pflag and documentation dependencies require separate notice review.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-cobra-checked-tree-insertion` | Cobra Checked Tree Insertion | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.command_graph` |
| `go53-cobra-command-name-conflict` | Cobra Command Name Conflict | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.command_namespace` |
| `go53-cobra-ranked-suggestions` | Cobra Ranked Suggestions | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go), [cobra.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/cobra.go) | `cli.diagnostic_ranking` |
| `go53-cobra-completion-state` | Cobra Completion State | [flag_groups.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/flag_groups.go), [completions.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/completions.go) | `cli.completion_isolation` |
| `go53-cobra-context-preflight` | Cobra Context Preflight | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.execution_cancellation` |
| `go53-cobra-error-writer` | Cobra Error Writer | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.output_failure` |
| `go53-cobra-validator-all-errors` | Cobra Validator All Errors | [args.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/args.go) | `cli.validation_composition` |
| `go53-cobra-required-annotation` | Cobra Required Annotation | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.annotation_schema` |
| `go53-cobra-command-finalizers` | Cobra Command Finalizers | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go), [cobra.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/cobra.go) | `cli.lifecycle_cleanup` |
| `go53-cobra-group-registration` | Cobra Group Registration | [command.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/command.go) | `cli.group_registry` |
| `go53-cobra-cancel-completion` | Cobra Cancel Completion | [completions.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/completions.go) | `cli.completion_cancellation` |
| `go53-cobra-active-help-lines` | Cobra Active Help Lines | [active_help.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/active_help.go) | `cli.completion_protocol` |
| `go53-cobra-markdown-filename` | Cobra Markdown Filename | [doc/md_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/md_docs.go) | `cli.documentation_paths` |
| `go53-cobra-man-fixed-date` | Cobra Man Fixed Date | [doc/man_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/man_docs.go) | `cli.documentation_clock` |
| `go53-cobra-yaml-command-metadata` | Cobra Yaml Command Metadata | [doc/yaml_docs.go](https://github.com/spf13/cobra/blob/adbc8813901bba65827259daa8e22ff94ec1f30e/doc/yaml_docs.go) | `cli.documentation_schema` |

### hashicorp/go-retryablehttp

Root MPL-2.0 and Go SPDX headers were inspected. Imported/distributed closure eligibility remains pending until source/executable distribution, modified files, notices, source availability and dependencies are specified.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-retryhttp-idempotent-policy` | Retry HTTP Idempotent Policy | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.retry_safety` |
| `go53-retryhttp-server-delay-cap` | Retry HTTP Server Delay Cap | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.server_backpressure` |
| `go53-retryhttp-duration-overflow` | Retry HTTP Duration Overflow | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.duration_representation` |
| `go53-retryhttp-header-selection` | Retry HTTP Header Selection | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.header_resolution` |
| `go53-retryhttp-deadline-backoff` | Retry HTTP Deadline Backoff | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.deadline_budget` |
| `go53-retryhttp-exhaustion-error` | Retry HTTP Exhaustion Error | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.failure_telemetry` |
| `go53-retryhttp-request-body-snapshot` | Retry HTTP Request Body Snapshot | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.body_ownership` |
| `go53-retryhttp-bounded-reader-body` | Retry HTTP Bounded Reader Body | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.body_buffer_budget` |
| `go53-retryhttp-prepare-error-close` | Retry HTTP Prepare Error Close | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.retry_resource_lifetime` |
| `go53-retryhttp-response-hook-attempts` | Retry HTTP Response Hook Attempts | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.hook_observability` |
| `go53-retryhttp-redacted-query` | Retry HTTP Redacted Query | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.log_projection` |
| `go53-retryhttp-clock-backoff` | Retry HTTP Clock Backoff | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.clock_injection` |
| `go53-retryhttp-deterministic-jitter` | Retry HTTP Deterministic Jitter | [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.randomness_injection` |
| `go53-retryhttp-roundtripper-default` | Retry HTTP Roundtripper Default | [roundtripper.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/roundtripper.go) | `http.adapter_initialization` |
| `go53-retryhttp-cert-wrapped-errors` | Retry HTTP Cert Wrapped Errors | [cert_error_go120.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/cert_error_go120.go), [client.go](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/client.go) | `http.error_type_resolution` |

### go-viper/mapstructure

Root MIT was inspected. No different license signal was found in inspected core Go files; this is not blanket clearance of the repository, derivatives or training rights.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-mapstructure-decoder-nil-config` | Mapstructure Decoder Nil Config | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.configuration_precondition` |
| `go53-mapstructure-depth-budget` | Mapstructure Depth Budget | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.recursion_budget` |
| `go53-mapstructure-cycle-input` | Mapstructure Cycle Input | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.graph_cycle` |
| `go53-mapstructure-collection-budget` | Mapstructure Collection Budget | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.container_budget` |
| `go53-mapstructure-normalized-key-collision` | Mapstructure Normalized Key Collision | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.key_identity` |
| `go53-mapstructure-unused-order` | Mapstructure Unused Order | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.metadata_order` |
| `go53-mapstructure-partial-result-policy` | Mapstructure Partial Result Policy | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.transactional_output` |
| `go53-mapstructure-hooks-errors-is` | Mapstructure Hooks Errors Is | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.alternative_hooks` |
| `go53-mapstructure-typed-nil-hook` | Mapstructure Typed Nil Hook | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.hook_value_lifetime` |
| `go53-mapstructure-delimited-escape` | Mapstructure Delimited Escape | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.delimited_text` |
| `go53-mapstructure-lossless-numeric` | Mapstructure Lossless Numeric | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.numeric_exactness` |
| `go53-mapstructure-hook-location` | Mapstructure Hook Location | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.temporal_context` |
| `go53-mapstructure-remain-conflict` | Mapstructure Remain Conflict | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.struct_schema` |
| `go53-mapstructure-tag-precedence` | Mapstructure Tag Precedence | [mapstructure.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/mapstructure.go) | `decode.tag_resolution` |
| `go53-mapstructure-unmarshal-ownership` | Mapstructure Unmarshal Ownership | [decode_hooks.go](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | `decode.custom_unmarshal` |

### go-ini/ini

Root and inspected sources have Apache-2.0 notices. struct.go states copying with modifications from encoding/json/encode.go; original revision and notices remain separately pending.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-ini-parse-byte-budget` | INI Parse Byte Budget | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.input_budget` |
| `go53-ini-quoted-comment` | INI Quoted Comment | [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.lexical_comments` |
| `go53-ini-section-duplicate-policy` | INI Section Duplicate Policy | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go), [parser.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | `config.section_identity` |
| `go53-ini-interpolation-cycle` | INI Interpolation Cycle | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.interpolation_graph` |
| `go53-ini-shadow-snapshot` | INI Shadow Snapshot | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.shadow_history` |
| `go53-ini-reader-reload-policy` | INI Reader Reload Policy | [data_source.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/data_source.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.source_lifetime` |
| `go53-ini-append-transaction` | INI Append Transaction | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.source_transaction` |
| `go53-ini-section-delete-index` | INI Section Delete Index | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.indexed_removal` |
| `go53-ini-key-rename` | INI Key Rename | [section.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/section.go), [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.key_relocation` |
| `go53-ini-effective-parent-keys` | INI Effective Parent Keys | [section.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/section.go) | `config.inheritance_resolution` |
| `go53-ini-save-atomic` | INI Save Atomic | [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.persistence_atomicity` |
| `go53-ini-write-style-instance` | INI Write Style Instance | [ini.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/ini.go), [file.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/file.go) | `config.format_state` |
| `go53-ini-struct-required-tag` | INI Struct Required Tag | [struct.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/struct.go) | `config.struct_presence` |
| `go53-ini-struct-nil-target` | INI Struct Nil Target | [struct.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/struct.go) | `config.reflection_precondition` |
| `go53-ini-strict-list-index-errors` | INI Strict List Index Errors | [key.go](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/key.go) | `config.typed_list_diagnostics` |

### dustin/go-humanize

API metadata says NOASSERTION but the actual root is MIT. The pinned gorhill origin of number.go contains WTFPL v2. Retain origin headers, separate license evidence and distribution scope rather than labeling it MIT-only.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-humanize-exact-fractional-bytes` | Humanize Exact Fractional Bytes | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go) | `format.exact_fractional_bytes` |
| `go53-humanize-byte-rates` | Humanize Byte Rates | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go), [bigbytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bigbytes.go) | `format.rational_byte_rates` |
| `go53-humanize-explicit-byte-unit` | Humanize Explicit Byte Unit | [bytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bytes.go) | `format.explicit_byte_unit` |
| `go53-humanize-bigbytes-precision` | Humanize Bigbytes Precision | [bigbytes.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/bigbytes.go), [big.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/big.go) | `format.arbitrary_precision_bytes` |
| `go53-humanize-strict-comma-parse` | Humanize Strict Comma Parse | [comma.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/comma.go) | `format.grouped_integer_grammar` |
| `go53-humanize-append-comma` | Humanize Append Comma | [comma.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/comma.go), [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go), [big.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/big.go) | `format.append_number_buffer` |
| `go53-humanize-int64-template` | Humanize int64 Template | [number.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/number.go) | `format.exact_integer_template` |
| `go53-humanize-precise-ftoa` | Humanize Precise Ftoa | [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go) | `format.float_precision_rounding` |
| `go53-humanize-scientific-si` | Humanize Scientific SI | [si.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/si.go) | `format.scientific_si_grammar` |
| `go53-humanize-rounded-si-promotion` | Humanize Rounded SI Promotion | [si.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/si.go), [ftoa.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ftoa.go) | `format.rounded_si_promotion` |
| `go53-humanize-relative-table-validation` | Humanize Relative Table Validation | [times.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/times.go) | `time.checked_magnitude_table` |
| `go53-humanize-duration-parts` | Humanize Duration Parts | [times.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/times.go) | `time.bounded_duration_decomposition` |
| `go53-humanize-ordinal64` | Humanize Ordinal64 | [ordinals.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ordinals.go) | `format.signed_ordinal` |
| `go53-humanize-case-plural` | Humanize Case Plural | [english/words.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/english/words.go) | `text.case_preserving_plural` |
| `go53-humanize-quoted-word-series` | Humanize Quoted Word Series | [english/words.go](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/english/words.go) | `text.escaped_series` |

### google/uuid

Root BSD-3-Clause and Google copyright/LICENSE-linking headers in inspected Go files were verified. Platform files, tests and notice scope of the actual closure are not yet frozen.

| ID | Proposed behavior | Inspected edit anchors | Review family |
|---|---|---|---|
| `go53-uuid-canonical-parse` | UUID Canonical Parse | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.canonical_text` |
| `go53-uuid-atomic-batch-parse` | UUID Atomic Batch Parse | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.transactional_batch` |
| `go53-uuid-append-text` | UUID Append Text | [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.append_serialization` |
| `go53-uuid-sorted-unique` | UUID Sorted Unique | [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.sorted_deduplication` |
| `go53-uuid-bounded-random-batch` | UUID Bounded Random Batch | [version4.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version4.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.bounded_entropy_batch` |
| `go53-uuid-write-text` | UUID Write Text | [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.short_write_io` |
| `go53-uuid-sql-null-reset` | UUID SQL Null Reset | [sql.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/sql.go) | `identifier.sql_scan_state` |
| `go53-uuid-owned-binary-value` | UUID Owned Binary Value | [sql.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/sql.go), [marshal.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/marshal.go) | `identifier.sql_binary_ownership` |
| `go53-uuid-nullable-binary` | UUID Nullable Binary | [null.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/null.go) | `identifier.nullable_encoding` |
| `go53-uuid-checked-hash` | UUID Checked Hash | [hash.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/hash.go) | `identifier.checked_hash_io` |
| `go53-uuid-v1-v6-conversion` | UUID V1 V6 Conversion | [version1.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version1.go), [version6.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version6.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [util.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/util.go) | `identifier.reversible_time_layout` |
| `go53-uuid-checked-timestamp` | UUID Checked Timestamp | [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go) | `identifier.checked_timestamp` |
| `go53-uuid-explicit-v7-time` | UUID Explicit V7 Time | [version7.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go), [version4.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version4.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go) | `identifier.explicit_time_generation` |
| `go53-uuid-instance-v7-generator` | UUID Instance V7 Generator | [version7.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [uuid.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/uuid.go) | `identifier.isolated_monotonic_generator` |
| `go53-uuid-checked-dce-domain` | UUID Checked DCE Domain | [dce.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/dce.go), [version1.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version1.go), [time.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/time.go), [node.go](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/node.go) | `identifier.dce_domain_policy` |

The [RFC 9562 text](https://www.rfc-editor.org/rfc/rfc9562.html) can provide independent UUID bit-layout evidence. Instance monotonicity and post-failure state behavior are project-selected contracts, not claimed mandatory properties of all UUID implementations.
