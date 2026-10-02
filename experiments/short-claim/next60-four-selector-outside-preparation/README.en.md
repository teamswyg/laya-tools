# Execution control for the first real four-task observation

The planned experiment compares original and alternative code for four requests and 19 finite inputs: 57 candidate dispatches. The checks in this folder cover **fake process state and stored metadata only**. They have not executed the original worker or approved training data.

The controller enforces a single start and wait, handles time and output limits, and preserves failure evidence. A surviving partial file does not establish actual call counts or successful completion. Real execution requires separate source, input, binary and review bindings.

`preparation-v1/` is the author's unexecuted preparation. `root-tested/source/` is a separate copy with only owned Go formatting changes. Eight fake tests, race checks and vet passed locally with Go 1.27.1. CI checks the same isolated module on both operating systems; it imports no original package or model. Consult the PR checks for the actual CI result.

Saved real observations still need independent comparison with the frozen expectations before qualification. This preparation does not add four training requests, prove 57 real calls or show a performance improvement. The Go heap setting is soft and does not enforce total process or GPU memory limits.

[HANDOFF](preparation-v1/HANDOFF.v1.json) and [RESULT](root-tested/RESULT.v1.json) record the preparation and local validation scope. Original source bodies, binaries, private host paths and real-work prompts are excluded.
