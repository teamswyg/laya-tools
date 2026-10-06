# Structure checks for bilingual training preparation

`riidolaya corpus-check` pairs Korean and English wording of one scenario. It
loads no model and performs no training. Output contains counts and structural
status only: no text, row/family identifiers, paths or text hashes.

```sh
riidolaya corpus-check --partition train --rubric-sha256 "$RUBRIC_RECEIPT_SHA256" < train.jsonl
```

Declare `train`, `validation`, `calibration` or `test` and the receipt digest of
the previously frozen rubric. The command does not hash the rubric receipt file;
a future fitting driver must verify that binding. It reads one supplied stream
and never automatically opens another partition. Supplying `test` explicitly
does read that file. A fitting driver's final-test seal is still separate work.

Every family needs exactly one `ko` and one `en` row sharing family, lineage,
partition, intent, unit description and ambiguity metadata. Source order is
retained. Repeated row IDs and repeated text SHA-256 values are rejected. Limits
are 4096 decoded text bytes, 16MiB per file and 2400 rows. Blank lines, added,
missing or duplicate JSON keys, case aliases, nulls, wrong types, NUL and invalid
raw UTF-8 are rejected. Failures return no usable partial data or source text.
Go's decoded JSON strings are used: original escape spelling is not preserved,
and strict rejection of unpaired escaped surrogates is not claimed.

The exact twenty-field input contract is the initial V4 authoring schema
`statehint-v4-original-train-seed-row-v1`. Its name contains `train`, but the
explicit partition is checked separately. There is no implicit schema expansion;
the JSON tags on `Row` declare all fields. Semantic rows require `role=prose`,
`semantic_applicable=true`, `task_level_completion=false`, `wording_index=1`,
`license=Apache-2.0`, the fixed ontology version and the declared rubric receipt
digest. A license field is an author's claim, not a rights audit. Excluded roles
such as table headers and code belong in separate controls.
Only an unclear row may declare an empty assertion-form array.

Internal `ValidateMetadata` checks declared split boundaries using a text-free
family manifest. `MatchMetadata` matches one parsed partition's families,
intents and lineages to that manifest without opening other text partitions.
Several families/intents may share a lineage, but every member must remain in
one partition. Numeric arrays and sorted slices avoid package maps, custom
locks and global caches. No comparative speed improvement is claimed.

**Structural acceptance is not label verification.** Bilingual equivalence,
visible scope, primary speech act, distinct scenarios, private-source exclusion
and publication/training rights still require separate review. Changed names or
identifiers do not establish independence. Omitted ancestry and semantic near
duplicates cannot be discovered by this checker alone.

Output fields `semantic_verified`, `training_allowed`, `model_used` and
`mutation_executed` are always `false`. A pass does not authorize training or
task-state changes. The future fitting driver must freeze complete source/data
hashes, rights and semantic review, exact partition counts, model-selection and
calibration rules before fitting.

[한국어](README.ko.md)
