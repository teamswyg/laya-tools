# V4 Go fitting driver with paired scenario evaluation

This maintainer tool trains the existing Go classifier with cross-entropy and
AdamW for development progress, completion-report and question hints. It needs
no Python or separate GPU runtime. It is separate from Laya neural-network,
LoRA and ternary training; existing models and the V3 driver stay unchanged.
These are implemented procedures, **not a newly executed training result**.
A concrete runnable plan remains pending until complete original data and review locks exist.

Humans and agents use the same command. Build with Go 1.27.1 from the clean,
verified source commit. The plan pins this executable's SHA-256, its embedded
Git revision/clean stamp and the current required Go source file hashes.
There is no automatic download or hidden model configuration.

```sh
go build -trimpath -o .cache/statehint/v4-fit ./cmd/riido-statehint-tune-v4
.cache/statehint/v4-fit --plan .cache/statehint/v4/PLAN.ready.json --check
.cache/statehint/v4-fit --plan .cache/statehint/v4/PLAN.ready.json --fit --out .cache/statehint/v4/run01
```

Declare exactly one of `--check` and `--fit`. Check mode verifies the parent
artifact and the train/validation/calibration files without model forwards,
fitting or final-test text access. Fitting is explicit. A full 2400-row ready
plan is not available yet; tiny temporary corpora cannot bypass the quota.

`Plan` JSON tags define the exact required contract, including nested objects.
Missing/unknown/duplicate/case-alias keys, nulls and wrong types are rejected.
Concrete rubric receipt/KO/EN definition, reviewed original authoring audit,
parent, source, executable and four partition hashes are mandatory. The frozen
v0.2 parent is 32960 bytes with 1920 steps. Originality, rights, semantics,
translation, ancestry and duplicate review claims bind to the exact manifest
digest. These booleans do not prove copyright or human ground truth. Build
metadata/checksums verify a trusted local build, not an externally signed attestation.

The text-free manifest checks all 1200 families/2400 row identities, paired
locales, intents, declared lineages and text hashes. Family quotas are
840/120/120/120, with each intent split 105/15/15/15. Global IDs and exact text
SHA-256 duplicates are rejected; declared lineages cannot cross partitions.
Distinct IDs do not prove semantic independence. Parsed rows are rechecked
against the manifest. Only text and expected intent become training samples;
unit/lineage metadata is never silently added to features.

Four fixed arms use independent optimizers: warm .005/.01/.02 from the parent,
and fresh .02. Each uses 40 epochs, batch 32, decay .001 and seed 1729, making
2120 updates. Fits are serial; the command limits concurrent Go CPU execution
processors to two, not the total OS thread count.
Strict wall-clock cancellation requires separate process supervision and is not claimed.

Selection ranks eligible [family reports](../statehintfamily/README.en.md) by
severity cost, eight-intent NLL, then fixed arm order. No eligible arm means a
preserved unqualified result with the test sealed. Only one arm sees a shared
temperature grid .5 through 5.0 in .1 steps. Eligible temperatures minimize
eight-intent NLL; ties keep the earlier temperature. The .9/.05 gate is fixed.

The selected artifact is saved and reloaded. Every eight-probability result,
intent, source, training-step and guard field must match on validation/calibration.
The before-test lock pins selected weights, sources, data, audits, parent,
temperature and selection metrics; file and directory sync plus exact readback
precede repeat preflight. Only then is final-test text hashed/parsed once to
compare child, frozen parent and the existing speech-act rule. Final outcomes
never select/relabel/refit. Input symlinks and duplicate inodes are rejected in a
metadata-only stage before source/data/model reads, after reading the caller's
designated plan file. A rule's forced one-hot scores are descriptive and
never eligible for the learned display gate.

Output requires a new single-name `.cache/statehint/v4/RUN` directory. Symlink
ancestry is rejected and opened directory identity checked. The private anchor
and run are mode 0700, files 0600. Existing results cannot be overwritten.
An 8MiB output budget reserves even partially failed files. Per-arm/temperature
results and failures are retained. These controls prevent local accidents under
a trusted owner; they are not authentication against a hostile same-user
process or compiler. Inputs/source are bounded workspace-confined regular files;
even hashing the final-test file is deferred until the lock.

Results remain research evidence. Synthetic intent labels do not prove work
success, authority, annotation creation or task-state transitions. There is no
network, upload, app integration or operational mutation path. Keep weights and
raw results out of Git; publish only reviewed public assets through the separate
Hugging Face release process. Results contain numbers, not input text; per-arm
files preserve complete cached eight-column scores in source order.

Current tests use original arithmetic/metadata and fake stage callbacks to
check parsing, cost, leakage and seal order. They do not execute Model.Fit/Predict
or demonstrate learned model quality.

[한국어](README.ko.md) · [Preparation plan](../../experiments/state-hints-v4/preparation/README.en.md)
