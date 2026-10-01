# Reading task usage and checking task requirements

[한국어](task-outcomes.ko.md)

A router's recommendation of a smaller model does not establish savings or task
success. Actual usage and whether the resulting files meet the requirements need
separate evidence. This guide explains `riido-taskoutcome`, which reads an
existing execution trace, and `riido-taskverify`, which checks candidate files
independently. Neither tool launches a model. Codex integration remains optional.

The [public registry](../benchmarks/training/public-task-candidates.json) has
**eight candidates in three code families**. Actually attempted scope is
**four distinct requests, two code families, one repository and 12 cumulative
owned records**. [Experiment 53](../experiments/task-outcomes/RESULTS-53.en.md)
repeated the keyword guard request twice per Luna/Sol low profile; all four
candidates passed independent closure checks with zero exit and complete whole
core usage. Four repetitions are not four requests. The first completed accepted
candidate's guard was integrated into runtime and checked with the pinned contract.

51's exact-comment rejection and support error, and 52's six accepted candidates,
five complete executions and one timeout/unknown-usage result remain preserved
historical records. Record count, candidate acceptance and completed execution
are distinct. Provider-attested model identity, money and subscription quota
remain unknown. These development cases are excluded from current training and final.

The separate `taskoutcome-event-key-bounds` parser candidate has a prepared
versioned verifier with **zero model attempts**. The [public Go source inventory](public-go-acquisition-53.en.md)
has 120 candidates from eight repositories and **zero execution-eligible tasks**;
do not add them to attempts or labels. The authored parser/verification examples
below are not savings evidence or a substitute for 2,400 final requests per domain.

## Summarize an existing trace

Build the Go binaries from the repository root. Python is not required at runtime.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskoutcome ./cmd/riido-taskoutcome
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
```

Start with the public authored fixtures to inspect the output format.

```sh
.cache/bin/riido-taskoutcome --input examples/task-outcomes/complete-authored.jsonl
.cache/bin/riido-taskoutcome --input examples/task-outcomes/partial-authored.jsonl
```

| Authored input | Observed input | Observed cached input | Observed output | Whole-trace usage |
| --- | ---: | ---: | ---: | --- |
| `complete-authored.jsonl` | 250 | 110 | 50 | Input, cached input, and output totals are known |
| `partial-authored.jsonl` | 100 | Unknown | 20 | Unknown after the second turn fails |

The first fixture contains two completed turns. Reasoning output of 5 is present
for only one turn, so its observed sum is 5 and its whole-trace reasoning total
is unknown. Missing optional reasoning does not erase known input, cached input,
or output totals.

The 110 cached input tokens are already included in the 250 input tokens.
Reasoning output is also included in output. Neither is added again, and input
minus cached input is not presented as billed tokens. The tool does not convert
tokens into fees, subscription usage, or remaining allowance.

Keep real traces in private local storage rather than a public directory. The
following paths are usage examples.

```sh
.cache/bin/riido-taskoutcome --input .cache/private-task-trace.jsonl
.cache/bin/riido-taskoutcome --input - < .cache/private-task-trace.jsonl
```

You may add short public identifiers for the task and the requested model.

```sh
.cache/bin/riido-taskoutcome \
  --input examples/task-outcomes/complete-authored.jsonl \
  --task-label authored-example \
  --requested-model example-model \
  --requested-reasoning low
```

These identifiers are caller-provided requests. `--requested-model` is not
evidence of the model actually used; `observed_model` remains `unknown`. Do not
put prompts, personal information, paths, or tokens in these identifiers. A
summary of a real private task still needs a separate review before publication.

## Read the summary without confusing its evidence

- `turn.completed` records that a turn ended. It does not establish correct
  code or completion of the task.
- `usage.*.observed_total` is the sum actually read from the trace. `total`
  contains a number only when that field is known for the entire trace.
- `null` or `unknown` means unknown, not an observed **0**. Missing cached input
  alone does not discard observed input or output.
- `usage_complete` is true only when input, cached input, and output are known
  for every turn and the lifecycle ends cleanly. Failed or open turns, startup
  errors, and unknown control events preserve observed values but prevent complete
  totals.
  This applies to the supplied trace; it does not prove that other attempts or
  trace files were not omitted.
- The actual model, process exit, execution time, producer provenance, and
  `task_acceptance` remain `unknown` in this summarizer.
- `startup_error_items` counts `item.completed` error diagnostics after
  `thread.started` and before its first `turn.started`. This narrow compatibility
  handling discards raw text and IDs. A startup error leaves whole usage unknown
  even if a later turn completes. This field does not classify the cause of a
  profile availability failure.

The input SHA-256 binds the exact bytes, including whitespace and newlines. It
does not prove who produced the trace, whether a model ran, or whether the task
succeeded. This summarizer does not launch models or establish producer
provenance. That binding is provided by the separate
[opt-in executor](task-execution.en.md).

The startup compatibility fix does not alter experiment 51's second original
attempt record. Its `summary_status: parse_failed` remains the actual result of
the frozen parser at that time. Any reanalysis with a new parser needs a
separately versioned record referencing the original; it does not create missing
usage or task success.

The summarizer does not retain messages, reasoning text, tool commands, file
bodies, local paths, or thread/item IDs in its output. Default limits are 64 MiB
per trace, 1 MiB per line, 100,000 events, and 4,096 turns. Malformed JSON,
duplicate relevant keys, numeric overflow, and exceeded limits are rejected with
fixed error codes that do not echo the source text.

## Verify the resulting files separately

The four contracts below have actual model attempts. `riido-taskverify` checks their
**declared file scopes**. A separate versioned parser verifier has been prepared;
its specification/checks are versioned and it has zero model attempts.
Read the specification first. This command prints its JSON specification without
running checks or launching a model.

```sh
.cache/bin/riido-taskverify --task catalog-min-context --spec
.cache/bin/riido-taskverify --task repo-keyword-language-guard --spec
```

| Task ID | Requirement checked | File scope and method |
| --- | --- | --- |
| `comment-budget-period` | Replace the budget-period comment with the exact specified sentence | Three pinned files: `go.mod`, catalog source, and tests. Exact change, gofmt, and pinned test execution |
| `comment-preview-authority` | Replace the comment to state that a repository suggestion never authorizes execution | One file, `pkg/reporouter/router.go`. Exact change and static gofmt check |
| `catalog-min-context` | Add optional `MinContext`, reject negative values, apply minimum-context eligibility, and preserve existing behavior | Nine pinned files covering catalog, planner, switchpolicy, and examples. Independent boundary checks and pinned catalog/planner test execution |
| `repo-keyword-language-guard` | Check every selected candidate keyword for non-Latin letters before Judge; preserve the lexical result and abstention reason without calling Judge when blocked | Three files in a separate base: pinned `go.mod`, reporouter source, and tests. Independent contracts and pinned tests; all four repetitions of this request accepted in 53 |

This file scope is the task closure. Files outside it are **unassessed**;
`accepted` does not establish correctness or safety of the entire repository or
every requirement. An unchanged base is not accepted as having performed the
requested task. Candidate-authored test assertions are not acceptance evidence.

The original three tasks' public base files are in `internal/taskverify/testdata/base`. Their pinned
revision is `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`; the verifier checks the
SHA-256 of every required file. The new keyword task uses a separate base at
public revision `146b02b9c37e6d90a11386050ccfdffc036c113e`, stored in
`internal/taskverify/testdata/repo-keyword-language-guard-v1`. Its versioned
definition binds source, specification, independent contracts, and separately
pinned LICENSE/NOTICE. The original three specifications and hashes are not
changed to match the new task. Prepare a candidate directory containing the task
changes against its matching base. For example, this command checks an **already
prepared** candidate:

```sh
.cache/bin/riido-taskverify \
  --task catalog-min-context \
  --base-dir internal/taskverify/testdata/base \
  --candidate-dir .cache/task-candidates/catalog-min-context \
  --go-root /absolute/trusted/go-toolchain
```

The command does not create that candidate or perform the task. Use the matching
task ID and candidate directory for either comment task. `--help` prints concise
usage instructions.
For the new keyword task, use `--task repo-keyword-language-guard` with its separate
`--base-dir` above and a candidate prepared against that base. `--spec` shows the
current task's base revision and pins without launching a model.
Set `--go-root` to the absolute path of a trusted Go 1.27.1 installation containing
`bin/go`. A packaged `-trimpath` binary without a default Go installation needs
this option for behavioral checks. Unavailable toolchain resolution or isolated
Go execution remains `verifier_unknown`, not an incorrect candidate label. This
is separate from the comment task's static checks without candidate execution.

The three tasks requiring behavioral tests run only when macOS `sandbox-exec`
isolation is available. The verifier copies the declared public source and
pinned tests into a temporary directory, configured to avoid network access and
the host's authentication environment. If isolation is unavailable or execution
is cancelled or the overall verification deadline or output limit prevents
completion, the result is `verifier_unknown`. Actual requirement failures in
independent tests reject the candidate.
Behavioral verification of these three tasks is currently unknown on other OSes.
`comment-preview-authority` uses static checks without executing candidate code,
so it is separate from that restriction. This is not a general security verifier
that approves arbitrary repositories or arbitrary code execution.

The verification JSON binds the task specification, base, and candidate-file
hashes to check results. `independent_tests` counts passed Go test/subtest events,
not distinct task-request samples. Read `status` and the individual check codes rather than
treating every `accepted: false` as an incorrect implementation.
`verifier_unknown` may indicate an unavailable verification environment.
Candidate code outside the supported structure is also reported as unknown with
`unsupported_candidate_shape`. For example, an implementation outside the
allowed base import set may be beyond this narrow verifier's scope; that alone
does not establish that the implementation fails its requirements.

For the new keyword task, changed imports, `init` functions, and compiler
directives also exceed the supported shape. Acceptance requires the named
independent contract tests to actually pass in their exact package, rather than
candidate-authored assertions. Source acceptance and staged LICENSE/NOTICE hashes
are separate; this does not assess candidate files outside the declared closure.

| Command | Exit code | Meaning |
| --- | ---: | --- |
| `riido-taskoutcome` | 0 | Summary or help output succeeded; this is not task success |
| `riido-taskoutcome` | 1 | Input, metadata validation, parsing, limit, or output error |
| `riido-taskoutcome` | 2 | Invalid argument syntax, such as an unknown option |
| `riido-taskverify` verification | 0 | Requirements accepted within the declared task closure |
| `riido-taskverify` verification | 1 | Candidate rejected |
| `riido-taskverify` verification | 2 | Invalid invocation, base files, output, or related call error |
| `riido-taskverify` verification | 3 | Isolation or behavioral verification unknown |
| `riido-taskverify --spec` or `--help` | 0 | Specification or help output succeeded; no candidate result |

Agents should read both JSON fields and exit codes. Successful parsing does not
complete a task, and `--spec` exit code 0 does not approve a candidate.

## Separate actual attempts from historical authored fixtures

[Experiment 50](../experiments/task-outcomes/RESULTS-50.en.md) stopped before a main
coding-model process because of executor packaging.
[Experiment 51](../experiments/task-outcomes/RESULTS-51.en.md) connected two owned
CLI attempts on one comment task. Candidate rejection in the first and profile
support failure plus startup-parser incompatibility in the second left the
remaining four entries unexecuted. First-attempt usage is complete; second-attempt
usage is unknown. No comparable two-profile capability results or savings were
established.

[Experiment 52](../experiments/task-outcomes/RESULTS-52.en.md) records all six
attempts under a separate plan requesting available `gpt-6-sol` and `gpt-6-luna`
profiles, both with `low`. Both candidates passed independent closure checks on
all three requests, but only five attempts exited zero with complete core usage.
The Sol attempt on `catalog-min-context` reached the 120-second deadline; its
accepted candidate is not relabeled as an implementation failure, and whole usage
stays `unknown`. Only the two comment requests have complete usage for both
profiles. This does not assume that the larger profile is always better or
establish a general savings rate.

The historical 51/52 subtotal is eight records over three requests. Through 53,
the current count is **12 records across four requests**, in two actually attempted
families and one repository. All four keyword-contract attempts in 53 were
accepted with zero exit and complete whole core usage. Accepted candidates are
52's six plus 53's four; the ten complete-core-usage records include 51's rejected
first attempt. One support failure and one timeout remain visible. The actual
provider-attested model stays unknown; the parser candidate has no model attempt.

53's sums are 339,805 input including 280,576 cached input, and 4,890 output
including 76 reasoning. Luna had less total input, but uncached input was
Luna 31,425 versus Sol 27,804. Do not double-count subsets or infer actual savings.

Actual routing evaluation still needs independently accepted outcomes from
executable profiles on the same specification and base, usage across all
attempts/retries, and requested/observed evidence. Compare completion rate with
whole-task usage to evaluate upward and downward routing. These development
records do not replace the [golden-set design](golden-set-scale.en.md)'s target
of at least 2,400 final requests per evaluated domain.
See the [acquisition plan](golden-set-acquisition.en.md) for source collection,
grouped splits, and comparison-execution budgets.

The authored JSONL fixtures and public code verification examples explain tool
behavior. They do not constitute actual model outcomes. The original tool scope
and validation plan are in historical
[plan 49](../experiments/task-outcomes/plan-49.json).

[Actual tool execution on authored fixtures](../experiments/task-outcomes/fixtures-49.json)
records two summaries, acceptance of the exact comment change, and rejection of
an unchanged base. Tool exit codes are separate from model process exits; there
are **zero actual model outcomes in that historical authored record**, separate
from experiments 51/52/53's later owned attempts.

Source acquisition and executable contracts are separated in the [120-candidate public Go survey](public-go-acquisition-53.en.md).
