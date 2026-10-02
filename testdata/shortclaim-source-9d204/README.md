# Historical source bytes

These text archives are exact Git blobs from commit `9d204c2c700505658c108297d5fa769a835a60c3`. They preserve historical plans' original source hashes after the validated-input optimization. They are test data, never a runtime source fallback, compiled package dependency, feature or result.

| Archive | Original repository path | SHA-256 |
|---|---|---|
| input.go.txt | pkg/shortclaim/input.go | 3b296ca0835c4397af7444a2f956aeae59638dd350cbc861bf36dd81bc210ecb |
| baseline.go.txt | pkg/shortclaim/baseline.go | ffb9f29ddc911d16b03a07cfe5c91bb99d8ae7f1d8aa1ea51edd651aee4bb00e |
| cli-main.go.txt | cmd/riido-shortclaim/main.go | 1c67c764c0faa10ab1e745593366323ffeb0302da262db85fb6aaf90075d0bd3 |
| cli-bench.go.txt | cmd/riido-shortclaim/bench.go | 97c6d1df360d48a76cd742a3e747c1210231cf3da3f1063e0a06d9827fa77dc2 |
| typed-replay-test.go.txt | cmd/riido-typedaudit/replay_test.go | 18cf0c6df82db7f4ea278c01474b630fcc2fb27a66c849b5d1b4469491f1d3e9 |

The fifth archive preserves a test file that the frozen 56b plan itself includes. The regression checks historical bytes separately from the current compiled source manifest and current semantic replay. Only the temporary replay plan/report source provenance differs. Public CLI verification continues to reject a historical plan against changed source files. Original plans, reports and hashes remain unchanged; this is neither another official observation nor a new benchmark or fit.
