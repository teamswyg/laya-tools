# Development-comment domain experiment

This maintainer command runs one fixed 2×2 study with fresh Contextual2048 linear
and 16-hidden-unit ReLU MLP models. It compares a class-balanced draw of existing
fit families with 64 original TRAINONLY development-comment families. These
source-derived AI annotations carry prior author/reviewer exposure; they are
not fresh evaluation, human truth or independent product observations.

| Arm | Backend | Additional training families |
| --- | --- | --- |
| `linear_balanced_control` | Contextual2048 linear | 8 old fit draws per intent |
| `linear_domain64` | Contextual2048 linear | 8 new domain families per intent |
| `mlp16_balanced_control` | Contextual2048 → 16 ReLU → 8 softmax | 8 old fit draws per intent |
| `mlp16_domain64` | Contextual2048 → 16 ReLU → 8 softmax | 8 new domain families per intent |

Every arm uses the same frozen 663-family base fit subset. Each profile adds 64
whole Korean/English families, giving 1,454 rows. Both architectures receive
exactly the same ordered samples within a profile. The recipe is hard CE
(alpha 0), 40 epochs, batch 32, rate .02, decay .001, seed 1729 and temperature 1:
1,840 updates per arm. Linear zero initialization and the MLP's seeded Glorot
initialization remain unchanged, so this is an architecture bundle comparison.
No parent model or recipe override is accepted.

The control ranks fit families by SHA-256 of
`dev-comment-domain-control-1729:` plus the family ID, in the fixed eight-intent
order. It takes eight whole pairs per class and cycles only if needed. Repeated
draws are labeled sampling, not new families. The new prefix and balanced prior
intentionally differ from the previous completion-heavy control.

```sh
go run ./cmd/riido-statehint-domain \
  --train TRAIN.jsonl --train-sha256 TRAIN_SHA \
  --augmentation DOMAIN.jsonl --augmentation-sha256 DOMAIN_SHA \
  --split SPLIT.json --split-sha256 SPLIT_SHA \
  --source-overlay SOURCE-GROUPS.json --source-overlay-sha256 AUDIT_SHA \
  --check
```

All paths are explicit repository-relative inputs with exact byte pins. The
original train840 and frozen group split pins are fixed. The augmentation must
contain 64 families / 128 rows / eight families per intent. Its required
`statehint-dev-comment-domain-source-groups-v1` audit binds those bytes, train840
and the split, uses `pair_id == family_id`, and identifies complete connected
original source groups/members. Every member must be frozen fit-side. A source
group cannot have multiple augmentation canonical groups; shared components
must describe the same complete group/member set. Old groups and dev assignments
are preserved. Inputs are bounded and checked before an output directory exists.
`--check` makes no model calls and writes no files.

For an authorized run, replace `--check` with
`--fit --out .cache/statehint-domain/NEW-RUN`. The run directory must be new and
has mode 0700; all artifacts have mode 0600. Optional
`--validation VALIDATION.jsonl --validation-sha256 VALIDATION_SHA` reads only the
previously exposed 120-family validation partition. There is no calibration or
final-test input. The 177-family internal dev set is also already exposed.
Both sets are **diagnostic only**: selection weights, selected arm, promotion
and deployment qualification stay zero/false. Existing family evaluation and
confidence .9 / margin .05 gates are unchanged.

Outputs preserve input bytes and pins, source audit, whole-family draws, ordered
sample receipts with class/locale counts, all eight probabilities and complete
family reports for all four arms. Each trained linear RSH v2 or MLP RSM v3
artifact is reloaded with its typed loader; every Prediction field and saved
byte must match. Failures retain an explicit `FAILED.json` rather than a complete
success report. Model bytes and raw study data belong in ignored local caches.

Compare data profiles within each architecture descriptively; do not infer
causality, fresh qualification or task completion authority. Go heap readings
are not process RSS. Measure whole-run OS peak RSS externally; arm timings are
sequential. Source additions and tests use only owned synthetic fixtures; actual
corpus results are reported separately after an authorized run.

## Actual run01

One authorized local run preserved all input pins and exact trained reload parity.
Internal dev has 177 families / 354 rows / 155 declared groups; previously exposed validation
has 120 families / 240 rows / 43 declared groups. Both are diagnostic only, with bilingual
rows and related scenarios rather than independent observations.

| Arm | Internal C precision (correct/proposed rows) | Internal eligible | Validation raw correct /240 | Validation gated correct/proposed | Validation correct C families KO/EN (each /15) |
| --- | --- | --- | --- | --- | --- |
| Linear / balanced control |25/25|yes|171/240|37/37|3/0|
| Linear / domain64 |27/27|yes|181/240|39/39|3/1|
| MLP16 / balanced control |36/38 (94.74%)|no|178/240|53/56|5/1|
| MLP16 / domain64 |35/38 (92.11%)|no|185/240|56/58|5/1|

**All four remain unqualified on exposed validation.** Its English completion
support remains 0 or 1 correct family; the linear control has zero English completion
proposals, so that locale's completion precision is undefined. The MLP's internal
false-completion family count worsened from 2 to 3 with domain data, failing the
unchanged completion-precision requirement in both profiles. Domain data raised
raw validation correct counts within both architectures but did not establish
useful completion coverage or fresh product quality. No arm was selected or
promoted; calibration and final-test calls remained zero.

The whole four-arm run took 4.01 s with OS peak RSS 27,082,752 bytes in this single local
observation, not a controlled memory comparison. Linear artifacts are 65,728 bytes;
MLP artifacts are 131,872 bytes. Inference used existing Go CPU APIs; no native/GPU runtime was added. Prior
experiments, datasets, gates and artifacts remain unchanged.
