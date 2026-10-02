# Concrete handoff for the next bounded Go loader

This is an implementation handoff, not a completed loader, Project, Features, ranking, or fit execution. The parent's separate input-bound result `53b492d0629f794785e592c90a96f3d86223dbf81891393d763a89e4a9d92d55` established that all original 72/216 inputs are supported. No new scientific condition is added.

## First scope: preserve the original 72 parents

Arrays and slices suffice for an intermediate DTO. `BoundParent` contains `OriginalIndex int`, `ParentID string`, `RawRequest string`, `Candidates []BoundCandidate`, `Truth claimfit.TruthState`, `Acceptable []int`, `WholeGroup int`, and `Role roleplan.Role`. A candidate DTO contains its original ID/Text, `NullableLabel *int`, and `LossEligible bool`. SourceID, CodeSHA, group, and role remain side metadata and never enter scorer text.

Read raw inputs in their fixed order: 48 legacy parents followed by 24 typed parents. Bind each original input to the stored truth, mask, and role at `/parents/{i}` in this audit's result. Preserve candidate length, order, ID, nullable labels, and acceptable-index order. A shortclaim input contains only `shortclaim.Schema`, the raw request, original candidate ID/Text, and a fixed provenance string.

```go
prepared, err := shortclaim.Validate(shortclaim.Input{
    Schema: shortclaim.Schema,
    Request: original.Request,
    Candidates: original.Candidates, // original ID/Text and order
    Provenance: "frozen-original-56-56b",
})
if err != nil { return zero, fixedInputError }
p := claimfit.Parent{
    Text: prepared,
    Truth: boundTruth,             // known=1, no_answer=2, unknown=3
    Acceptable: ownedAcceptable,   // original set; empty for unknown/no_answer
    Role: boundRole,               // train=0, validation=1, calibration=2
    WholeGroup: originalGroupID,
    LossEligible: boundMaskArray,  // copy existing masks; unused slots=false
}
```

Never manufacture Prepared's internal arrays. The first `claimfit.Project` validates its complete input and generates features, so new source/input/binary/plan pins and a one-attempt execution ledger must be frozen before that call. Read existing roles from 66 rather than assigning them again. Whole-group role inconsistency, duplicate parent indices, original-text SHA mismatch, or replacement of unknown labels must return fixed failures.

Train and validation rows retain every original known candidate; a false mask leaves the original label intact at weight 0. Unknown and calibration parents remain in the audit/runtime snapshot without fit rows. Do not prefill observed Project output using this audit's eligible counts: compare original denominators and output row references after execution. Preserve every no_answer and negative candidate.

Primary utility scope A retains the original acceptable sets and every candidate in all 51 known/no_answer requests. A mask does not make a check free. Report the complete-caption parent diagnostic B separately, never replacing A or 58. BM25 and oracle must use the same A denominator and unit check costs, while retaining the count of all 21 unknown parents separately. The parent will execute the first Project and same-scope utility only after a separate freeze.

## Later scope: fixed append of two families from 68

Hash and decode the same bytes from contract `818eed54f6b14b1f512025804cf7782d4a59d94ed30eeeaa46c56a270afbd0eb` and independent QA `ecc9de25fea14872257fc50af2ddb312db644d0e0b5e9ed07ffdc9f3fb119083`. The fixed parent order is prefix-reject, prefix-allow, nested-allow, nested-reject, with 2/2/3/3 candidates. Bind proposed labels and acceptable sets from the fixed QA metadata rather than executing contract Want to produce new truth. Preserve null actual weights and proposed-only status until a formal combined-data freeze.

The combined DTO owns `OriginalWholeMembership json.RawMessage`, retaining every original `canonical_members` and `relationships` item from 65, alongside two separate `ProposedWholeFamily` records. Each new record preserves its family ID, original 68 parent indices, and complete source/helper/module/license closure pins. Do not narrow the closure to a selected candidate or append only part of a source family to an existing role.

Actual combined roles require **a new complete membership mapping and a new plan using the same ASCII seed1729**. Do not automatically copy roles from 66 onto the combined corpus. If recorded evidence cannot establish a relation between new families or with the original corpus, return the exact gap rather than inventing missing edges. Opaque hashes do not prove that external edges are absent. Preserve existing floors, unknowns, and masks, and do not claim that additional samples, roles, or training are already authorized by readiness results.

Shared repository edits and new Assign/Project/Features/Baselines/API/fit/model executions are all 0 in this handoff.
