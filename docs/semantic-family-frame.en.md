# Source-grounded family frames

`internal/semanticframe` is a small Go interface for checking a communication
plan against an immutable, byte-pinned inventory. It helps prevent a later
comment from accidentally using a different inventory revision, dropping a
declared dependency or resolving an ambiguity that was meant to remain open.
It does not generate text, assign claim states or qualify training data.

This interface supports **new generic declaration formats only**. Historical
private coder, amendment and reviewer records are not supported by implicit
conversion. Explicit reviewed adapters, actual whole Source QA and independent
frame/comment meaning reviews remain necessary before real authoring. No actual
cohort, model or private records are included in these tests or documentation.

## What a frame says

A V3 frame identifies its Source, family and sibling slot (1–3), then pins the
selected Source, observation, typed inventory, producer declaration, standard
Source review and separate accepted-version evidence. It also pins the
definitions and inventory schema source as opaque method references.

The communication plan selects **whole inventory items**, for example
`/propositions/0` or `/referent_support/0`. It gives each selected item one role:

- `utterance_content`: intended content of the comment.
- `governing_constraint`: retained context governing the wording.
- `unavailable_source_context`: Source context unavailable to the comment reader.

These roles are not question/progress/completion labels. A Source fact does not
automatically become an assertion in the comment. A plan also carries bounded
Source evidence and a private description. Preserved ambiguities identify
selected inventory items and say `preserve_without_resolution`. Three siblings
are not required to communicate three different claim states.

## Supplying evidence

Callers capture immutable bytes using their existing read-receipt workflow and
supply `PinnedBytes{File, Bytes}`. The expected relative path, SHA-256 and byte
count must come from independently retained pins. The package performs no file
opens, storage writes, network requests or receipt generation.

```go
frame, err := semanticframe.DecodeFrame(frameBytes)
if err != nil {
    return err
}
accepted, err := semanticframe.NormalizeAccepted("original", suppliedInputs)
if err != nil {
    return err
}
summary, err := semanticframe.JoinFrame(frame, accepted, registeredSourceIDs)
```

`suppliedInputs` contains the exact version-binding, Source, observation,
inventory, producer, standard-review and independent-inventory-acceptance bytes.
Amended selections additionally contain retained predecessor version bindings.

The supported generic wire contracts are:

- Frame: `riido-semantic-family-frame-draft-v3`.
- Version: `riido-accepted-inventory-version-binding-draft-v1`.
- Inventory: `riido-inventory-graph-declaration-v1`.
- Producer: `riido-inventory-original-producer-declaration-v1` or
  `riido-inventory-amended-producer-declaration-v1`.
- Independent inventory acceptance:
  `riido-inventory-acceptance-declaration-v1`.
- Standard Source review: the public `sourcecohort.Review` shape.

Exported Go types define the closed required fields. Original and amended
producer schemas are distinct. `NormalizeAccepted` checks every supplied digest
and size, then joins the selected observation/inventory/producer/review and each
independent inventory acceptance declaration to the exact same Source/version.
The producer ID must differ from declared reviewer IDs. An accepted filename,
producer flag or standard Source review cannot replace separate inventory
acceptance evidence. Original mode has no predecessor; amended mode requires
at least one distinct previous inventory version binding for the same Source.
Predecessor bytes are retained inputs and their metadata is checked; their
referenced old artifacts are not recursively opened or requalified. Prior held
versions must stay unchanged in the caller's historical records.
Here `amended` means an inventory-version amendment. A fresh review of an
unchanged original inventory still selects `original` and pins its new review
and independent acceptance evidence; it is not an inventory amendment.

Byte verification establishes **the supplied declarations and their joins**.
It does not authenticate a producer/reviewer or establish that a declared pass
is correct or independently performed. Observation bytes are verified as UTF-8
but their schema/checker is not executed. Review read receipts and method files
remain unopened references. Their pin metadata, review UTC, read-result path,
whole UTF-8 evidence boundary and declared dependencies must be consistent;
this is not complete Source-review validation. The package does not reproduce
`sourcecohort`'s whole-cohort provenance checks.
All captured and declared references, including unopened review/predecessor
metadata and frame methods, share one retained path/pin consistency registry.
Exact repetitions are allowed; one path declaring different hashes or byte
counts is rejected even across different roles.

## Structural safeguards and limits

Inventory item order is retained because pointer indices bind to its digest.
The graph keeps open polarity, referent kind and time relation strings; it
introduces no new semantic enum. IDs are checked for duplicates, referent
parents for cycles, and proposition/referent/opposition/time/constraint edges
for missing endpoints. Each selected item's declared endpoints must also be
selected in the plan, possibly under a different disclosure role. Complete
Source dependencies must name externally supplied registered Source IDs.
Forward declared edges are checked; implicit governing meanings, omitted edges
and same-split dependency requirements still need independent review/workflow
checks. Bounds cannot establish semantic fidelity.

Source boundaries cover `[0, Source.bytes)`. Evidence spans use UTF-8 byte
offsets with exclusive end, must be nonempty and land on code-point boundaries.
Whole-item pointers reject nested fields, aliases, leading-zero indices,
duplicates and conflicting roles. Errors contain only fixed codes, never caller
text, paths or identifiers. Caller slices are not mutated.

Sources are at most 16 KiB; each frame, generic inventory or evidence artifact
is at most 128 KiB; combined normalization inputs are at most 8 MiB. Each graph
category, plan, ambiguity list and evidence list is limited to 256 items.
Registered Source IDs have a separate ceiling of 400. Oversized inputs are
rejected whole. Closed JSON rejects missing, duplicate, case-alias and unknown
fields, null arrays, invalid UTF-8/surrogates and trailing values.
Direct typed `JoinFrame` callers receive the same 128 KiB compact encoded-size
ceiling. Scalar and collection limits plus an allocation-free JSON size count
run before sorting or pointer parsing; oversized strings are rejected before
splitting or decoding hashes. This measures bounded Go work, not process RSS,
native memory or GPU use.

Success says `STRUCTURAL_DECLARATIONS_JOINED_QA_PENDING`, with
`meaning_proven:false` and `training_eligible:false`. It cannot establish
utterance force, truth, bilingual fidelity, naturalness, rights, authentication,
human Gold, independent providers or whole Source QA. Eight later QA axes are
assessments, not eight frame fields or expected head states. Independent
frame/comment reviews and blind text-only references remain separate later
steps. `familycohort` remains unchanged and structural-only. There is no new
approval gate, runtime model or training operation in this package.
