# Reading task usage and checking task requirements

[한국어](task-outcomes.ko.md)

A router's recommendation of a smaller model does not establish savings or task
success. Actual usage and whether the resulting files meet the requirements need
separate evidence. This guide explains `riido-taskoutcome`, which reads an
existing execution trace, and `riido-taskverify`, which checks candidate files
independently. Neither tool launches a model. Codex integration remains optional.

The [public task candidates](../benchmarks/training/public-task-candidates.json)
currently contain **0 actual `model_outcomes` records**. No new paid model run
was performed in this implementation cycle. The examples below are authored
parser inputs or ways to check public code. They are not evidence of a model's
completion rate, performance, or savings.

## Summarize an existing trace

Build the Go binaries from the repository root. Python is not required at runtime.

```sh
mkdir -p .cache/bin
go build -o .cache/bin/riido-taskoutcome ./cmd/riido-taskoutcome
go build -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
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
  for every turn and the lifecycle ends cleanly. Failed or open turns and unknown
  control events preserve observed values but prevent complete totals.
  This applies to the supplied trace; it does not prove that other attempts or
  trace files were not omitted.
- The actual model, process exit, execution time, producer provenance, and
  `task_acceptance` remain `unknown` in this summarizer.

The input SHA-256 binds the exact bytes, including whitespace and newlines. It
does not prove who produced the trace, whether a model ran, or whether the task
succeeded. An executor that automatically binds this summary and the verifier
report into an actual model outcome is outside this tool's current scope.

The summarizer does not retain messages, reasoning text, tool commands, file
bodies, local paths, or thread/item IDs in its output. Default limits are 64 MiB
per trace, 1 MiB per line, 100,000 events, and 4,096 turns. Malformed JSON,
duplicate relevant keys, numeric overflow, and exceeded limits are rejected with
fixed error codes that do not echo the source text.

## Verify the resulting files separately

`riido-taskverify` currently checks only the fixed file scope of **three tasks**.
Read the specification first. This command prints its JSON specification without
running checks or launching a model.

```sh
.cache/bin/riido-taskverify --task catalog-min-context --spec
```

| Task ID | Requirement checked | File scope and method |
| --- | --- | --- |
| `comment-budget-period` | Replace the budget-period comment with the exact specified sentence | Three pinned files: `go.mod`, catalog source, and tests. Exact change, gofmt, and pinned test execution |
| `comment-preview-authority` | Replace the comment to state that a repository suggestion never authorizes execution | One file, `pkg/reporouter/router.go`. Exact change and static gofmt check |
| `catalog-min-context` | Add optional `MinContext`, reject negative values, apply minimum-context eligibility, and preserve existing behavior | Nine pinned files covering catalog, planner, switchpolicy, and examples. Independent boundary checks and pinned catalog/planner test execution |

This file scope is the task closure. Files outside it are **unassessed**;
`accepted` does not establish correctness or safety of the entire repository or
every requirement. An unchanged base is not accepted as having performed the
requested task. Candidate-authored test assertions are not acceptance evidence.

The public base files are in `internal/taskverify/testdata/base`. Their pinned
revision is `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`; the verifier checks the
SHA-256 of every required file. Prepare a candidate directory containing the task
changes against that same base. For example, this command checks an **already
prepared** candidate:

```sh
.cache/bin/riido-taskverify \
  --task catalog-min-context \
  --base-dir internal/taskverify/testdata/base \
  --candidate-dir .cache/task-candidates/catalog-min-context
```

The command does not create that candidate or perform the task. Use the matching
task ID and candidate directory for either comment task. `--help` prints concise
usage instructions.

The two tasks requiring behavioral tests run only when macOS `sandbox-exec`
isolation is available. The verifier copies the declared public source and
pinned tests into a temporary directory, configured to avoid network access and
the host's authentication environment. If isolation is unavailable or execution
is cancelled or the overall verification deadline or output limit prevents
completion, the result is `verifier_unknown`. Actual requirement failures in
independent tests reject the candidate.
Behavioral verification of these two tasks is currently unknown on other OSes.
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

## Evidence still needed for an actual router comparison

These tools prepare two components: reading usage and checking requirements.
An actual router evaluation additionally needs real model executions against the
same task specification and base files, usage for all attempts and retries,
requested settings separated from observed execution evidence, and independent
verification bound to the candidate files. Completion rate and usage per task
must be compared together to evaluate upward or downward routing.

The authored JSONL fixtures and public code verification examples explain tool
behavior. They do not constitute actual model outcomes. The implementation scope
and validation plan are recorded in the
[precommitted plan](../experiments/task-outcomes/plan-49.json).
