# Preparing a tiny hint model64–67

The aim is a small model that helps order candidate checks before LLM work. A score such as “this code may also match nested directories” could guide what to inspect first. Preserve original verification and fallback when the hint is wrong. This stage has not demonstrated faster model inference or Codex cost savings.

We are **identifying defensible training supervision and keeping related code together across development splits**. The72 requests and216 candidate positions remain the same population; preparation does not create additional examples.

| Preparation | Purpose | Current state |
|---|---|---|
| Go array projection64 | Internal maintainer module preserving truth and masks when preparing training columns | Synthetic/port verification; actual corpus projection pending |
| Complete group binding65 | Keep related requests, candidates, helpers and types together | Metadata bindings verified for17 original groups |
| Role execution tool66 | Assign whole components to train/validation/calibration | Seed and execution plan frozen; original assignment pending |
| Supervision inventory67 | Join saved truth and caption reviews to propose candidates usable in loss | First actual metadata inventory succeeded once |

The67 inventory retained support for35 positive and95 negative candidates. Keep the original labels of23 masked known candidates with future weight0. The63 unknown candidates retain null labels. All216 positions remain; removing candidates must not change the population or evaluation denominator.

Originally16 groups contain stored known truth;15 groups have any eligible candidate after the proposed masks. All9 known candidates in group52 are masked. Actual eligible groups per role remain uncomputed. If the existing9/3/3 requirements fail, retain that result rather than searching for a favorable seed or weakening masks.

The first67 metadata process took0.71 seconds and about26.2MiB peak RSS, verifying8.59MB of JSON input once. This does not measure inference, GPU memory, Go heap or resident-router speed. The public Go port passed synthetic tests and build checks; the actual observation belongs to the earlier private original binary.

## Try the inventory

Run from the repository root using Go1.27.1. No model download or Python is required.

```sh
go run ./cmd/riido-supervision --input-root . \
  --output supervision-new.json \
  --plan-sha256 c9a07943f8047b69e32d93354e812ca5b158b1dc7476264bd5b8505a86242408
```

Output must be a new file. Humans can read fixed help and failure codes; agents can use exit status and JSON. Success is `passed_metadata_proposed_supervision_only`, not training readiness. The2MiB-per-file/16MiB-total read bounds are not RSS limits.

Next, verify frozen sources, inputs, plan and actual binary after automatic CI merges, then execute one whole-component assignment. Instantiate wording from another source as actual candidates and review its bounded semantic target. Subsequently test whether checking effort or total cost decreases against existing controls on that same scope. A sufficiently large protected final population of distinct requests remains separate from development trials.

[67 usage and actual evidence](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/supervision-scope-67/README.en.md) · [Independent binding checks](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/supervision-scope-67/QA-67.v1.en.md) · [66 execution conditions](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/role-execution-preparation/README.en.md) · [Development issue](https://github.com/teamswyg/laya-tools/issues/19)
