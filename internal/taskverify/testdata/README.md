# Public pinned verifier fixtures

`base/` contains only public source, tests, example JSON, LICENSE and NOTICE from
`teamswyg/laya-tools` revision `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`.
These are the project's existing Apache-2.0 public files, with their source
attribution retained. No private task, trace, model output or weights are copied.

Tests verify each required file against the SHA-256 pins in `verify.go`; they do
not depend on Git history, a network fetch, or candidate-authored assertions.
Changed candidates are authored within disposable test directories. Those
fixtures prove acceptance-check behavior, not performance of any model.
