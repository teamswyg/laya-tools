# Canonical concept metadata contract

The public auditor preserves the existing concept metadata schema name and exact typed structure:

`three-claims-short-core-full-provisional-concept-tranche-v1`

The authoritative field types are the `Metadata` structure and its explicitly named children in this directory's main.go. There is no catch-all extension object. Every required field must exist with its exact JSON type; unknown fields, duplicate keys, null arrays/objects outside the sole documented Correction.Prior exception, invalid UTF-8 and trailing non-JSON content reject. Optional fields are only those with `omitempty` in the typed source. This document contains no actual concept prose, source utterances, votes or labels.

For tranche N, N is an integer from1 through10. Its forty records use global registration ordinals from `40*(N-1)+1` through `40*N`, exactly once each. The tranche scope is:

- `scope.this_tranche.production_tranche_number`: N
- `scope.this_tranche.registration_ordinals`: `[40*(N-1)+1,40*N]`
- `scope.this_tranche.groups`:40; `bilingual_families`:120; `intended_rows`:240
- `scope.known_provisional_after_this_tranche`, when present: `groups=40*N`, `bilingual_families=120*N`, `intended_rows=240*N`

The full scope remains400 groups,1200 bilingual families,2400 intended rows, locales `["ko","en"]`, ten tranches of40/120/240, and `not_a_reduced_study=true`. Each group has two necessary slots, three families with frame ordinals1..3, and two connected within-group links. No study count, gate, allocation, label or wording is produced by this auditor.

The seed remains `claims-short-core-full-v1:concepts:1729`. Group IDs are `g_` plus the first24 lowercase hex characters of SHA-256 over UTF-8 `seed:group:%04d`, using the global group ordinal. Family IDs are `f_` plus the first24 lowercase hex characters of SHA-256 over `seed:family:%04d:%d`, using the global group ordinal and frame ordinal. Reusing local1..40 ordinals for later tranches is invalid. Identity uniqueness is bookkeeping only.

For every new tranche N, use the existing uniform `relation_universe` structure:

- `registered_own_ordinal_range`: the exact own forty-group range
- `currently_known_provisional_ordinal_range`: `[1,40*N]`
- `external_metadata_references`: one object for each prior tranche1..N-1
- `status`: nonempty provisional status text

Each external object has exactly `registration_ordinal_range`, `path`, and `sha256`. Its range is the exact aligned forty-group range of that prior tranche. Its path is the exact clean relative `.json` path used for that supplied input, and its SHA is the exact lowercase SHA-256 used to pin those bytes. Writers must take both values from the actual sealed input, not a conceptual or phantom registry. No external path is opened automatically.

The declared universe must be covered exactly by own groups plus the declared external files, all actually supplied to the current audit. External ranges must be forty-group aligned, nonoverlapping and wholly inside Known; repeated references and self-references reject. Every external `(range,path,SHA)` must match an actually supplied input. Missing references reject even if another supplied file contains the needed groups. This permits later tranches to refer to multiple known earlier files without adding new schema keys.

Every present Known range must be exactly `[1,ownTrancheEnd]`; the external array must contain exactly the prior tranche ranges1..N-1 and may not reference own or future tranches. Earlier declarations retain their earlier Known ranges when more inputs are supplied. Their relation endpoints cannot silently expand into later supplied records. An absent declaration remains `absent_unknown`, preserving v1 compatibility; its recorded relation endpoints still require actual supplied pinned records. Absence is not a provenance assertion.

An across-group hypothesis retains `group_ids`, matching `registration_ordinals`, `relation_hypothesis`, `rationale`, `necessary_distinction`, `review_status`, and optional `relation_dimensions`. Endpoint counts are2..400; IDs must match the global hashes and must be distinct within that hypothesis. The exact required review status remains `pending_independent_full_registry_lineage_review`. Present-universe endpoints must be inside that tranche's covered declaration. Every endpoint also resolves an actually supplied group.

Slots, family phenomenon/extra-slot tags, and relation dimensions remain lowercase ASCII identifiers. Type strings for within links and across hypotheses preserve ASCII case, including CLI/UI, with semicolon-separated identifier syntax. Whitespace/prose type strings reject. The output retains exact type spelling and pending-status counts. Candidate graph components ignore direction only to describe review connectivity; they do not certify ancestry, semantic independence, effective sample size, admission or allocation units.

Every metadata-boundary flag must be explicitly false: utterances, targets, expected semantic counts, fixed semantic triangle, assigned length bins, assigned splits, and training/gold/reference material. Unknown utterance/message/target/split keys reject. Prose concept descriptions remain nonempty UTF-8 strings; structural checks cannot semantically prove that such prose is unrealized or establish actual-short eligibility.

The existing optional protocol, essential-scope ledger, generation/scaffold provenance and source-boundary pins retain their v1 typed fields. Source-boundary pins are format-checked declarations only. They are not opened or externally verified. Exact verification of supplied metadata pins and universe references must not be described as author ancestry verification.

Uncorrected metadata must OMIT `retained_provisional_metadata_correction` entirely. The field is optional, and omission decodes as no retained correction. Top-level `retained_provisional_metadata_correction: null` is invalid. A top-level batch0 object `{batch_number:0,status:string,input_rows_created_or_revised:0}` is also invalid: `InitialCorrection` is supported only as `prior_correction_object` inside a valid batch1 correction that reconstructs that predecessor. A writer must choose the omission form before sealing; this is an existing v1/v2 schema rule, not a post-seal rewrite requirement. The owned `TestUncorrectedMetadataOmitsCorrection` fixture confirms accepted omission and rejects null/top-level batch0.

Retained correction reconstruction remains batch1, rows0,1..32 field deltas, valid predecessor pin/byte length and one exact declared serialization. Allowed corrections retain core_event ko/en, necessary_slots ko/en, and the previously allowed family concept/caution/phenomenon/extra-slot fields. The only added paths are paired whole arrays `/across_group_relation_hypotheses/{canonicalIndex}/group_ids` and `/across_group_relation_hypotheses/{canonicalIndex}/registration_ordinals`. Both must be retained as deltas for each repaired relation; element pointers reject. Both old and new endpoints require matching global IDs/ordinals, distinct endpoints, cardinality2..400, and the explicitly declared own+prior Known namespace. Relations corrected this way require a present universe. All status/type/dimensions/rationale/distinction and provenance fields are unchanged, remain outside this whitelist, and participate unchanged in exact predecessor byte/SHA reconstruction. Array pointer indices must have canonical nonnegative decimal spelling;00,+0,-0 aliases reject. Old/current values, unchanged-value rejection, duplicate pointers, optional prior batch0 rows0 correction object, and exact predecessor reconstruction remain enforced.

The supplied audit scope is a contiguous prefix of1..10 tranches:40,80,120,...,400 groups. Each metadata input is at most204800 bytes; aggregate is at most2097152 bytes. Count is checked before any stat or content operation. All explicit inputs then receive path/pin and stat-only byte preflight before any content read/hash/decode. The opened regular file must match preflight file identity and size; actual byte length must also match. Ten maximum-size accepted files total2048000 bytes, below the aggregate cap. The larger aggregate rejection fixture deliberately uses oversized synthetic stat declarations to demonstrate pre-load rejection.

The report schema is `concept-registry-audit-text-free-v2`. Aggregate semantics preserve the registered categorical tag spelling. Source/provenance fields are `source_metadata_schema`, `scope_caps`, and `input_validation_scope`; a present input declaration is labeled `present_provisional_supplied_coverage_only`. These fields describe this tool's bounded checks only. Reports contain caller-supplied paths/pins and registered categorical tags; review them before external publication. This contract includes no actual metadata and does not constitute an audit of actual inputs.

The only explicit typed null accepted is batch1 `prior_correction_object: null`, meaning the same absence as omitting that optional field. All other typed nulls still reject. With no prior object, either the original base serialization string or exactly `UTF-8 JSON, ensure_ascii=false, separators=(comma,colon), trailing LF. Restore each old field value then omit retained_provisional_metadata_correction to reconstruct predecessor exactly.` is accepted. Both perform the identical restore-every-old-field plus omit-entire-correction transformation, followed by exact predecessor bytes/SHA verification. No other recipe text is accepted; a nonnull prior object still requires the original prior-object serialization.
