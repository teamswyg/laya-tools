# Authored telemetry fixtures

[한국어](README.ko.md)

These JSONL files were written for parser tests. They are not recorded model
responses, actual model evaluations, accepted coding tasks, or savings evidence.
They contain no real request or credential. Repository Apache-2.0 applies to
these original fixtures.

```sh
go run ./cmd/riido-taskoutcome --input examples/task-outcomes/complete-authored.jsonl
go run ./cmd/riido-taskoutcome --input examples/task-outcomes/partial-authored.jsonl
```

The first fixture has observed core totals of input 250, cached 110 and output 50.
Cached tokens are already inside input, and reasoning is already inside output;
the tool never subtracts them into billed tokens or adds them twice. Reasoning
is present for one turn only, so its observed total5 is partial and its complete
total is unknown. The second fixture preserves observed input 100/output 20, with
cached and full-trace totals unknown after a failed second turn.

`requested-model`, `requested-reasoning` and `task-label` are caller-provided safe
identifiers, not observed execution evidence. Actual model, process exit,
execution time, producer provenance and task acceptance remain `unknown`.
The tool does not launch Codex, accept an
arbitrary approval flag, publish traces, or calculate subscription use/fees.
Message, command, path and thread/item identifiers are discarded. Trace SHA-256
binds the exact input bytes, including whitespace/newlines; it does not prove
who produced them. A later independent verifier must bind any acceptance result
to the task specification, pinned base and candidate artifacts.

Missing/null usage remains unknown. Each token field reports its observed sum,
observation count and missing completed-turn count separately. A missing cached
field does not erase known input/output values. Core `usage_complete` requires
input/cached/output for every turn and no failed, open or unknown control event.
Optional reasoning completeness is separate. Parser failures emit fixed error
codes with no successful JSON report. Default limits are 64MiB trace, 1MiB line,
100000 events and 4096 turns; nested item bodies remain bounded by the line.
