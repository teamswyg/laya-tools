# Actual supervised warm-start tuning of three-display hints: V3

[한국어](README.md) · [Original plan](PLAN.original.json) · [Reproduction path mapping](REPRODUCE-PLAN.json) · [Complete results](RESULTS.first.json)

The target is cheap progress, completion-report and question suggestions for development work. We actually continued training from Go v0.2 on 400 original synthetic messages using cross-entropy and fresh AdamW. The architecture remains 1,024 features, eight intents and FP32; the file is still 32,960 bytes. This is not Laya-derived, LoRA or distillation. Selected training made 520 updates, taking stored steps from 1,920 to 2,440 without resuming optimizer moments.

## Predeclared selection

| Candidate | Learning rate | Validation correct / 80 | Three displays + none NLL |
|---|---:|---:|---:|
| Selected v0.2 warm start |0.005|68|0.42385|
| V0.2 warm start |0.01|68|0.42690|
| V0.2 warm start |0.02|69|0.43852|
| Fresh initialization |0.02|56|0.73647|

Every arm used 40 epochs, batch32, seed1729 and decay0.001. Only validation NLL selected the candidate; higher correct count did not change the criterion. Training used eight-intent CE. Selection grouped the other five probabilities as none without changing the original scores or argmax. Separate calibration80 chose temperature1.0 from the fixed0.5–5.0 grid. Confidence0.9 and margin0.05 remained fixed. The artifact, temperature and lock were saved before the new200-case test was first parsed.

## Actual new200-case results and tradeoff

| Same new test200 | Parent v0.2 | Selected V3 |
|---|---:|---:|
| Correct overall |152/200 (76%)|158/200 (79%)|
| Progress reports |45/50|44/50|
| Completion reports |40/50|46/50|
| Questions |24/50|36/50|
| Correct / created proposals |57/57|90/94|
| Wrong completion-report proposals |0|4|
| Correct display coverage /150 target cases |38%|60%|

**Accuracy and coverage increased, but incorrect completion displays increased too.** Preserve V3 as a research candidate; no default model or policy was replaced. More proposals and fewer wrong proposals are different goals. V3 scored82/100Korean and76/100English. These are synthetic labels, not product truth or proof of completed work. The rule control scored96/200, with75/88correct proposals and13wrong completion displays; rule scores are not calibrated probabilities.

Keep the full eight-intent confusion, scores, abstentions and metadata diagnostics. Live reads/state mutations are zero. The saved file was reread, checksum-verified and compared on all200cases for exact prediction/probability parity. First stdout/stderr and all model weights remain in private cache, never Git. The selected model SHA is `6fafd22cb5af6f3c9a5ab51ae0022e67e82d7d753702a770aab5559e1804869d`.

## Data and reproduction

Two AI authors created 760 fictional rows:400train/80validation/80calibration/200test. Each primary intent has100train/50test; each of the other five has20train/10test. Languages are balanced. Author judgments are not human adjudication, product ground truth or statistically independent samples. The initial100English test drafts mirrored Korean scenarios and were replaced before predictions by different incidents/styles; draft hashes/correlation notes remain private. No outcome-driven relabeling or trimming occurred. The test is now exposed and cannot select another revision as unseen evidence.

With Go1.27.1, explicitly download the parent and use a new output directory. The reproduction plan changes only original relative cache paths to public filenames; original plan, lock and source hashes remain preserved.

```sh
hf download JooYoon/riidolaya-statehint-go-v0.2 statehint.rsh --revision 81335eadd9f2753d3b86c932e9b00c636714e3b4 --local-dir ./models/statehint-v0.2
go run ./cmd/riido-statehint-tune --plan experiments/state-hints-v3/REPRODUCE-PLAN.json --parent ./models/statehint-v0.2/statehint.rsh --driver cmd/riido-statehint-tune/main.go --out ./new-v3-run
```

Reproduction verifies exposed development results, not a new blind test. Cross-platform byte-identical floating-point results or timing are not promised. Peak OS RSS of the original single macOS Go process, including four fits/validation/calibration/test/reload, was15,974,400bytes (15.23MiB). It excludes GPU, live reads and long-running costs; do not add Go heap.

Publication-source CI and new HF model publication remain pending. Model publication would not imply deployment. A private Task reader must independently verify opaque work revisions, current catalogs and content placement. A completion report is a textual assertion, not a lifecycle fact. Original public training material and code are Apache-2.0.
