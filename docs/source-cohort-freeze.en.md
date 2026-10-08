# Freeze a whole source cohort

`riido-sourcecohort` prepares one complete registered source cohort. It checks
byte pins, allocation, full joins, recorded reads, evidence boundaries and
declared reviews. It does not generate sources, comments, labels or model scores.
All examples and tests are synthetic software fixtures, not admitted data.

```sh
go build -o ./riido-sourcecohort ./cmd/riido-sourcecohort
./riido-sourcecohort --default-recipe > recipe.json
./riido-sourcecohort --help
```

The v1 recipe is exactly 400 rows: eight named prelabel strata, 50 each;
35 TRAIN / 5 DEV / 5 CAL / 5 TEST per stratum; DEV has 30 short and 10 general
rows. The pinned recipe must equal this contract. A smaller cohort cannot become
ready. No subset, refill or selection is produced.

## Before source creation: allocation

Supply a caller-pinned plan, recipe and prospective registry. Every registry row
has `source_id`, `stratum`, `split`, `dev_style`, `source:null`, and an explicit
`dependencies` array. Non-DEV style is the empty string. Splits are lowercase
`train`, `dev`, `cal`, `test`. Source IDs and role IDs use ASCII letters, digits,
underscores or hyphens, at most 128 bytes.

Plan fields are `schema`, `registry`, `creation`, `reviews`, `source_schema`,
`allocation`, `split_config`, `author_id`, `checker_id`, `input_max_bytes`,
`source_max_bytes`, `freeze_max_bytes`, and `output_max_bytes`.
`schema` is `riido-sourcecohort-plan-v1`. `creation`, `reviews`, and `allocation`
are explicitly null at this phase; `source_schema` may also be null. No future
digest or nonexistent manifest is required. Each known file reference is
`{"path":"relative-file.json","sha256":"<independent digest>","bytes":<exact size>}`.
Paths are relative to `--root`; absolute paths, parent traversal and parent
symlinks escaping that root are rejected. Zero bytes require the empty-file hash.

```sh
./riido-sourcecohort --root ./new-cohort \
  --plan allocation-plan.json --sha256 <independent-plan-digest> \
  --bytes <exact-plan-size> --allocation-only \
  --out ./new-cohort/allocation-evidence
```

The command reads metadata through `reviewpacket.Capture`, verifies all 400
slots and planned dependencies, and emits the aggregate `allocation_ready`.
It writes `ALLOCATION.private.json`, `SUMMARY.json`, and fresh read receipts.
It opens no source documents. This result establishes allocation only.

## After creation and review: freeze

Use a new plan with actual creation/review/schema pins and an `allocation` pin
for the earlier `ALLOCATION.private.json`. Use a separate realized registry:
all original slot IDs, strata, splits, styles and planned dependencies must match
the allocation; every `source` now contains its actual byte pin. Keep the earlier
prospective registry and allocation unchanged.

```sh
./riido-sourcecohort --root ./new-cohort \
  --plan freeze-plan.json --sha256 <independent-plan-digest> \
  --bytes <exact-plan-size> --out ./new-cohort/freeze-evidence
```

The registry schema is `riido-sourcecohort-registry-v1`; it contains
`cohort_id`, `intended_labels:false`, and `rows`. Creation and review manifests
use `riido-sourcecohort-creation-v1` and `riido-sourcecohort-reviews-v1`, each
with `rows`. All three must join exactly once on every source ID. Missing,
extra or duplicate IDs block the entire freeze.

Creation rows bind the source file, author ID, `creator_kind:"ai_nonhuman"`,
declared `rights_allowed:true`, creation UTC, dependencies and a pinned creation
receipt. The receipt uses `riido-sourcecohort-creation-receipt-v1` and repeats
those bindings with `consulted_none`, consulted file pins and dependencies.
An explicit none declaration is required when the consultation array is empty.

Each review contains the exact source/schema file references, `reviewer_id`,
`reviewed_utc`, `complete`, `structural`, `source_scope`,
`independent_semantic_verdict`, `holds`, `evidence`, `required_dependencies`,
`observation`, `check_binding`, `read_start`, and `read_result`.
Readiness requires declared completeness/structure, scope
`complete_source_situation`, verdict `declared_pass`, and no holds. Author and
reviewer IDs must differ; this is a declared process restriction, not identity
authentication or proof of provider independence.

The original checker read must be a pinned `reviewpacket` start/result pair,
with one input-open attempt, exact source hash/size, raw start-byte linkage and
no errors. Actual UTCs must order creation, read start, read completion, terminal
decision and review. The result path is the start path plus `.result`.
These receipts establish the recorded tool operation, not a global earliest read.

Evidence spans have `kind`, `start`, `end`, using UTF-8 byte offsets, exclusive
end. They must be nonempty and land on code-point boundaries. A `source_scope`
span must cover the entire source. Other allowed kinds are `proposition`,
`opposition`, `event_anchor`, `communication_time`, `absence`, `not_applicable`.
Valid offsets cannot establish whether an excerpt supports a semantic assertion.

The pinned observation packet is retained verbatim as UTF-8 in the private join.
`check_binding` points to `riido-sourcecohort-check-binding-v1`, containing
`source`, `observation`, `source_schema`, `report`, and `executed_utc`.
The external report has `structural_valid`, `mandatory_complete`,
`source_phase_scope`, optional `failure_code`, and `scope`.
The adapter verifies those files and their binding; it does not execute or
independently reproduce the external schema checker. Its successful flags remain
declared checker claims, not official semantic QA.

All required dependencies from the registry, creation row, creation receipt and
review are checked together. Unknown, repeated-within-list, self or cross-split
IDs block the freeze. Repeated declarations across separate records are combined.
Actual dependency components are calculated with deterministic sorted joins.

## Outputs, bounds and limits

Stdout contains only the aggregate. On success the state is
`structural_provenance_freeze_ready`; the private full 400-row
`FREEZE.private.json` includes source bytes, original observation bytes,
creation/review records and checker bindings. No source text, hold reason,
private path or case ID is emitted to stdout. The full review manifest remains
pinned; declared holds are counted by category before individual source reads.
Failure retains the whole denominator and unresolved rows. It never exports a
qualified subset or hides holds. A technical failure can stop later verification;
unverified rows remain unresolved.

JSON rejects missing, unknown, duplicate/case-alias keys, wrong types, invalid
UTF-8, unpaired surrogate escapes, trailing values and unbounded nesting/arrays.
Input metadata is at most 8 MiB, source documents at most 16384 bytes, and private
freeze/output ceilings at most 64 MiB. Plan output must reserve its freeze ceiling,
4096 bytes for the next receipt pair and 65536 bytes for terminal metadata;
output minimum is 1 MiB. Initial plan reading has a separate fixed 64 MiB bootstrap
ceiling. Each operation preflights remaining space. Retained source/observation
bytes and final serialization are bounded. Outputs are new private directories
and exclusive 0600 files, with synced writes, atomic publication and directory
sync. An existing output is never overwritten. Broken storage can prevent a
terminal summary; incomplete receipts/files are retained.

Byte verification does not prove source meaning, whole-unit inventory truth,
authentic authorship, legal clearance, independent providers or human Gold.
Declared rights, AI origin and semantic verdicts remain reviewer responsibilities.
The adapter neither admits training data nor satisfies S3–S11, frame/reference,
style, support, precision, unsafe-positive, resource or usefulness gates.
