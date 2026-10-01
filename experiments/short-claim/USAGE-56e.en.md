# Using the finite-input property audit

`riido-propertyaudit` is an offline Go maintainer tool aligning **request meaning with truth evidence** for tiny claim-model development. It loads no Laya or model weights. Humans and agents can follow request→candidate source→listed input→expected/observed values. It does not authorize general function correctness or safety.

The easiest check from the repository with Go1.27.1 is:

```sh
go test ./internal/finiteproperty ./cmd/riido-propertyaudit
```

This reads public evidence and replays source behavior. Linux/macOS CI checks observation equality, adding no official experiment/model calls. The preserved Darwin executable was not thereby executed on Linux.

## Historical collection procedure

This describes the completed56e collection. A new experiment needs a new plan/version preserving historical records. Keep executables out of Git; record build recipes/hashes. Require real source/input Git commits and preserve the same executable between preparation/audit.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o PRIVATE_BINARY ./cmd/riido-propertyaudit
PRIVATE_BINARY --stage prepare --source-commit SOURCE_COMMIT --out NEW_PREPARATION
# Freeze dataset.json and plan.json bytes in a separate input commit, then:
PRIVATE_BINARY --stage audit --input-commit INPUT_COMMIT --in NEW_PREPARATION --out NEW_AUDIT
```

The official run also used the environment in the [frozen build recipe](build-recipe-56e.json). Preparation emits dataset/plan with zero candidate execution. Audit requires equality with actual Git blobs at `experiments/short-claim/probes-56e.json` and `execution-plan-56e.json`. Existing output directories, tampered data, changed executables and bad input commits refuse. The CLI does not prevent repetition into every new directory, so maintain an honest official-attempt/failure ledger.

## Interpreting results

[Raw evidence](results-56e.json) includes full expected/got observations in `source_results`. `rows` preserves original candidate order and evidence indices. `matches_finite_literals` means agreement on listed inputs only, not other inputs, general function truth or model probability. `finite_no_answer` does not instruct removal or reduced checking: preserve candidates and fallback.

This is truth material for subsequent utility/training planning. Training eligibility false, fits/models/rankings0 and new independent parents0. Read [English results](RESULTS-56e.en.md), [Korean](RESULTS-56e.ko.md) and [next public-source contracts](SOURCE-NEXT-56e.en.md) for scope and remaining work.
