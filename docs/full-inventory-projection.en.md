# Retain full inventory while deriving a graph

`internal/inventoryprojection` is a pure Go preparation interface. It copies a
typed graph from a supplied native inventory while retaining the exact original
inventory bytes, pin and complete typed view. No filesystem/network access,
historical adaptation, accepted-version record, reviewer verdict or training
permission is generated. Tests use original public software fixtures only.

```go
result, err := inventoryprojection.Project(inventoryprojection.Inputs{
    Config: configBytes, Full: fullBytes, Native: nativeContractBytes,
    Mapping: mappingContractBytes, Adapter: declaredAdapterSourceBytes,
    Source: nil,
})
```

Each input is `PinnedBytes{File, Bytes}` with an independently expected relative
path, SHA-256 and exact size. `Config` is closed and contains
`selected_full_inventory`, `adapter_source`, `adapter_version`,
`mapping_contract`, `native_contract`, and `graph_output_path`. Its pins must
match all supplied inputs. The supported version is `riido-inventoryprojection-v1`.
Adapter bytes establish a declared pinned identity, not proof that those source
bytes produced the executing binary.

The separately pinned `NativeContract` contains `native_inventory_schema` and
the complete `official_field_keys` set. There are no guessed historical schema
literals or field names. The closed `MappingContract` contains `graph_schema`,
`copied_fields`, and `full_only_fields`; values must exactly match the public
`SupportedMapping()` definition, including list order. A free schema string or
hidden field-key configuration cannot replace these contracts.

`Result.Binding` contains only full/graph/config/native/mapping/adapter pins and
adapter version. The graph is derived under its own digest and output path.
Source/observation/producer/review/acceptance/predecessor/selection lineage joins
are not produced. Those belong to later explicit historical adapters. Current
semanticframe V3 cannot consume historical full-inventory acceptance by replacing
its inventory with this graph.

Every graph item, description, open polarity/kind/time value, nullable parent,
evidence span and array position is copied. Full-only negative IDs, combined
negation explanation, official-field support, unresolved questions, provenance
and completeness declarations remain in `Result.Full` and exact
`RetainedFullBytes`. Official fields retain original object order and each raw
support value. `Full`'s Go serialization is not the native object representation;
use retained bytes when referencing the original. Returned objects belong to the
caller; preserve pinned bytes rather than mutating them. A support-only change
may leave graph bytes unchanged while changing full/config pins, which is valid.

Mechanical checks cover explicit endpoints in graph and full-only support,
negative-ID membership, duplicate IDs, referent parent cycles, self dependencies
and conflicting same-path pins across supplied/unopened references. Exact pin
repetitions are allowed. Negative-ID membership cannot establish negative meaning.
Dependency IDs are retained; registration, split and whole-cohort closure remain
later checks. Declared completeness, inspection, status, value and unresolved
fields are preserved without judging their semantics. Omitted/implicit governing
relations still need full-inventory fidelity review.

With `Source:nil`, numeric span ranges are checked against the declared Source
size; `UTF8SpansVerified` stays false. Supplying the exact pinned Source bytes
also checks UTF-8 validity and code-point boundaries for every supplied span,
including full-only support. Empty evidence arrays are preserved and prove no
completeness. Span checks cannot establish that an excerpt supports a claim.

All inputs are at most 128 KiB each, Source at most 16 KiB, and the aggregate
ceiling is 8 MiB. Arrays/objects have at most 256 entries, nesting at most 32,
keys/IDs at most 128 bytes, and decoded strings at most 65536 bytes.
IDs and official support keys use ASCII letters, digits, underscores or hyphens.
Before typed list allocation, a bounded preflight checks collection/depth limits. Before
encoding a graph, its compact JSON size is checked, including HTML/control
escaping; the output ceiling is 128 KiB. Oversized inputs are rejected whole.
Missing/duplicate/case-alias/unknown native fields, null arrays, malformed UTF-8,
unpaired surrogate values and trailing JSON are rejected. Errors contain fixed
codes, never caller text or paths. Arrays/sorted lookup slices replace hot maps;
there are no global locks or model-runtime dependencies.

`MeaningProven` and `TrainingEligible` always remain false. No full Source QA,
rights, authenticated independence, semantic acceptance, human Gold, frame
eligibility or model usefulness is established. Actual whole Source QA remains
required before real projection/authoring. No new approval gate is added.
