# Working on laya-tools

This is a public Go repository. Keep application/runtime code in Go. Python is
allowed for maintainer-only model conversion/reference experiments, never as a
runtime requirement. Codex registration must remain opt-in.

Use a branch and pull request. CI, not human reviewer approval, is the merge gate.
Trusted same-repository PRs queue for automatic squash merge; `quality` must pass.
Do not bypass branch protection or claim passing checks without actual results.
When a check fails, fix the cause and rerun the affected checks. Draft PRs opt out
of automatic merge. No new paid model evaluations without task authorization.

Before pushing, run `go test -race ./...`, `go vet ./...`, formatting, and the
redacted secret scan. Chain publication after successful verification, not after
an ignored command exit code. Native integration tests run in CI after verified
model download. Keep fake/scaffold behavior visibly distinct from native tests.

Never commit private source, user prompts from real work, credentials, local
absolute paths, model binaries, raw pprof/ORT traces, or local caches. Benchmark
only public/original fixtures; label development results and limitations. Do not
alter metrics or lower routing thresholds merely to improve reported numbers.

Model archives are immutable versioned public assets. Verify licenses, upstream
revisions, and all SHA-256 values before updating the manifest. Do not replace a
pinned asset in place. Preserve fallback behavior and explicit model overrides.

pprof measures Go memory, not total native/GPU memory. Separate those metrics.
Core ML is experimental until a real supported graph and provider traces prove
execution. Keep exact line ranges, input budgets, and truncation indicators.
