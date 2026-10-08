# Join FULL acceptance to a derived communication graph

`internal/acceptedfullbridge.Join` closes the software gap between two different
artifacts: an independent review accepts a **complete native inventory**, while a
communication plan refers to a **derived graph**. The FULL inventory remains the
acceptance anchor. A graph digest never replaces its digest. This pure Go
endpoint uses supplied pinned bytes, performs no file/network access, and writes
no producer, reviewer or acceptance records.

The endpoint executes `inventoryprojection.Project` with exact FULL/config/native/
mapping/adapter/Source bytes. Supplied graph bytes must exactly equal the derived
graph; supplied projection-binding bytes must close and exactly match all seven
binding fields. Source bytes are required here so spans can be checked against
actual UTF-8 boundaries. Each input has an independently expected relative path,
SHA-256 and exact byte size.

The newly defined wire contracts are explicit **generic declarations**:

| Contract | Required links and checks |
| --- | --- |
| `Version` | Original/amended inventory selection and independent first/fresh review lineage; Source ID; Source, observation, FULL, graph, projection binding, producer, current standard Review, independent FULL acceptance and retained-history pins. |
| `Producer` | Same inventory selection, Source, observation and FULL; inventory-predecessor history only; producer ID and actor match the FULL coder declaration. No producer acceptance flag exists. |
| `Acceptance` | Exact selected FULL, observation, producer and current Review; the current Review's reviewer ID; `declared_pass`; scope `complete_full_inventory_official_and_support`; complete native official-field keyset; history. Acceptance time cannot precede current Review time. |
| `History` | Same Source; separate prior FULL, observation, producer, Review and version pins; `held` or `superseded` plus `inventory_predecessor` or `review_predecessor`. Referenced historical bodies remain unopened and unqualified. |
| `Frame` | Separate accepted FULL and projection-binding pins plus a `semanticframe.Frame` plan using `riido-projected-family-plan-v1`. Its inventory pin is the derived graph; its native-contract pin and all selected-version links must match. |

The current standard `sourcecohort.Review` must declare a complete structural
pass with no holds and the exact Source/observation/dependencies. Supplied public
reviewpacket start/result bytes must join its actual declared reviewer actor,
Source digest and size, one recorded read attempt and ordered timestamps. Its
separate checker binding/report must join the exact observation and Source
schema with an execution time between read completion and Review. These checks
verify declared lineage; they neither authenticate an actor nor prove that a
recorded read was an agent's first-ever read, or rerun the historical checker.

Declared producer and reviewer/read actors must also differ when role IDs
differ. This generic contract treats a bare ASCII `label` and `/root/label` as
the same actor label, case sensitively. Other namespaces or nested paths are
rejected rather than guessed; supporting them needs an explicit contract.
Different accepted labels still do not authenticate distinct people, agents or
providers. Alias checks only reject known self-review declarations.

Original inventory selection permits no inventory predecessor; amended selection
requires one. Separately, `first_review` permits no review predecessor and
`fresh_review` requires one. A review predecessor keeps the exact same FULL,
observation and producer with a different Review/version pin. This permits a
fresh independent review of an unchanged original inventory while retaining an
old procedural hold. Retained history is never recursively accepted or promoted
to pass. A current
held Review cannot be used as current acceptance. Duplicate/conflicting pins are
checked across supplied inputs, FULL provenance, historical references, review
methods/receipts and frame methods. Exact repeated pins are valid; one path cannot
claim different contents.

A support-only amendment may preserve graph bytes while changing FULL bytes.
Then the FULL/config/projection/version/producer/acceptance/frame links must be
updated explicitly. Reusing an old FULL acceptance is rejected even if graph
bytes are identical. This is why `semanticframe.NormalizeAccepted` V3 is not used
to manufacture graph acceptance. `semanticframe.JoinGraphPlan` shares only graph
and plan structure checks and expressly establishes no FULL acceptance by itself.

```go
result, err := acceptedfullbridge.Join(inputs)
// inputs supply pinned bytes; the function opens and writes no artifacts.
// result.Projection retains exact FULL bytes and the independently pinned graph.
```

Inputs are rejected whole above 128 KiB each, Source above 16 KiB, receipts above
8 KiB, or aggregate bytes above 8 MiB. Acceptance/history lists have at most 256
entries; registered Source IDs at most 400. A token preflight bounds JSON
collections to 256 entries and depth to 32 before typed allocation. Keys/IDs are
bounded, string decoding is bounded, and the shared typed plan checks account for
compact JSON escaping before sorting or splitting pointers. Errors contain fixed
codes without caller words, IDs or paths. Runtime code uses bounded arrays/sorted
slices and introduces no global lock, Python or model dependency.

**Actual historical integration remains incomplete.** Existing native record
schemas cannot be guessed from these generic declarations. A safe text-free
export is still needed for exact historical selected-version, producer,
independent FULL acceptance, retained/current disposition wire shapes: schema
literals, field names/types/nullability, exact FULL/observation/reviewer links,
original/amended lineage and native schema/official keyset. A reviewed adapter
must preserve and pin original records and distinguish complete official/support
acceptance from producer flags or structural reports. This endpoint does not
perform that export or adaptation, project admitted data, author frames, produce
head references or train a model. Whole Source QA is still a prerequisite to
actual projection/authoring.

Success reports `FULL_ACCEPTANCE_AND_GRAPH_PLAN_DECLARATIONS_JOINED_QA_PENDING`.
`MeaningProven` and `TrainingEligible` stay false. Meaning, fidelity, whole-cohort
QA, rights, authenticated/provider independence and real usefulness are still
unproved. Tests use original public software fixtures, including identical-graph
support amendments, held history, full/graph substitutions, actor mismatches,
partial acceptance scope, conflicting pins and resource limits.
