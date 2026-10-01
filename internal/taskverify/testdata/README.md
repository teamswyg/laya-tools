# Public pinned verifier fixtures

`base/` contains only public source, tests, example JSON, LICENSE and NOTICE from
`teamswyg/laya-tools` revision `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`.
These are the project's existing Apache-2.0 public files, with their source
attribution retained. No private task, trace, model output or weights are copied.

Tests verify each required file against the SHA-256 pins in `verify.go`; they do
not depend on Git history, a network fetch, or candidate-authored assertions.
Changed candidates are authored within disposable test directories. Those
fixtures prove acceptance-check behavior, not performance of any model.

`repo-keyword-language-guard-v1/` is a separate immutable original-source fixture
from public revision `146b02b9c37e6d90a11386050ccfdffc036c113e`. Its source closure
is `go.mod`, `pkg/reporouter/router.go`, and `pkg/reporouter/router_test.go`.
Apache-2.0 `LICENSE` and project `NOTICE` are separately pinned staging files;
candidate attribution files are outside source-acceptance scope. Versioned pins,
the prompt, independent contract digest, and exact package/test terminal-pass
requirements are bound by `TaskDefinition` and its `TaskSpec.DefinitionSHA256`.
The candidate never supplies the behavioral acceptance assertions. Existing
50/51 task definitions, source closures, source pins and spec hashes are unchanged.

`taskoutcome-event-key-bounds-v1/` adds another code family, the original
public JSONL summarizer source/tests and module at revision
`ee72334166e2962b0821d5198a50fdd13f92ab29`. Its Apache-2.0 LICENSE/NOTICE are pinned
separately. The independently supplied contract validates duplicate decoded
envelope keys, inclusive count/byte bounds, per-event reset, numeric lifecycle
and usage meaning, and opaque nested payload preservation. Named terminal passes
are required in `pkg/taskoutcome`; repo-preview test evidence cannot satisfy this
contract. Authored baseline/correct/incorrect candidates are verifier controls,
not model executions or additional independent golden requests.
