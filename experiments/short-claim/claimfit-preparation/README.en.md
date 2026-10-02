# Preparing Go training columns from short-claim supervision

[claimfit.Project](../../../internal/claimfit/projection.go) moves caller-verified, frozen text, truth, roles and masks into existing Go training columns. It is not a loader, semantic eligibility policy, role allocator or trainer. This public port verifies code integration using synthetic inputs; it does not project the live corpus, fit a model, run inference or measure performance. [한국어](README.ko.md)

Each `Parent` supplies the full `shortclaim.Prepared`, original `Truth`, original-order `Acceptable []int`, `roleplan.Role`, `WholeGroup int` and candidate `LossEligible [8]bool`. An internal driver can use the following call. This is an `internal` package; an external-project CLI/API is not included in this port.

```go
projection, err := claimfit.Project([]claimfit.Parent{
    {
        Text: prepared,                 // full original, separately bound by caller
        Truth: claimfit.Known,
        Acceptable: []int{2, 0},         // original set and order
        Role: roleplan.DevelopmentTrain,
        WholeGroup: frozenGroup,
        LossEligible: [8]bool{false, true, true},
    },
})
```

`projection.Parents` retains every original input/candidate, order/count and complete acceptable set. Known gives label 1 to every acceptable candidate and 0 to the rest; no_answer retains label 0 for every candidate. Masked known rows remain physical rows with their original labels/features and weight 0. Unknown's `NullableLabel.Known=false` means a **null label**: reading `Positive=false` alone as negative is incorrect. Unknown creates no fit rows. Calibration preserves full original truth/candidates but enters neither fitting nor training AUC diagnostics.

Development/validation use existing `pairlearn.Dataset` sparse columns: owned `Offsets`, `uint16 Indices`, `float64 Values`, `Labels`, `Groups`, explicit `SampleWeights` and original parent/candidate `RowRef` arrays. Features reuse `hintlearn.Features(original Request, original Candidate.Text)`. IDs, provenance, truth, roles, groups, masks and review metadata are not feature arguments. No class balancing, parent weighting, listwise loss or deletion of masked known rows is introduced. Physical retention of zero-weight rows is not claimed equivalent to row deletion during training.

`DevelopmentAUC`/`ValidationAUC` copy only positive-weight rows and report original fit, masked, unknown-candidate and positive/negative denominators. Existing AUC on this view is an **unweighted diagnostic** on the selected rows. It does not replace full-candidate ranking, fallback, Top, costs or existing 5% utility validation. Synthetic checks verified view contents/denominators without AUC scoring or Fit. Successful empty/all-zero-weight projection also does not imply successful fitting or readiness.

The same `WholeGroup` across roles is rejected, including unknown, all-masked and calibration parents. This does not prove caller group numbers match original components, members outside the batch are complete, or existing 9/3/3 coverage holds. `RoleEnumContract` describes only the current train=0/validation=1/calibration=2 bridge. Source/text/truth/review SHA, whole membership, supervision/evaluation eligibility, roleplan source, actual seed/coverage and Fit driver/config must be frozen separately by the caller. No roles, seed or new scientific gate are created here.

Complete prepared **output payload** is capped at 64MiB, counting runtime parents/retained strings, slice headers, both fit and both diagnostic sparse arrays and RowRefs. The first feature pass determines exact size/finiteness and checks overflow/limits before allocating complete result columns; the second fills them with the same feature path. This is not a 64MiB peak RSS guarantee including caller inputs, feature/sort/row-plan scratch or allocator slack. CPU/GPU/pprof measurements were not performed. Returned arrays are mutable caller-owned data; concurrent input mutation and integrity after edits remain caller responsibilities.

Eight publicly safe original/independent-QA archives are protected by an [exact SHA/size archive guard](../../../internal/claimfit/archives_test.go). Originals were not rewritten. Historical test modules/READMEs containing personal paths and raw test output are withheld, with their hashes retained in the ledger. The [independent findings](INDEPENDENT-FINDINGS-64.en.md) and [ledger](INDEPENDENT-AUDIT-LEDGER-64.json) concern the prototype, not a new independent review of this port. The porter authored the prototype; another reader must independently inspect the public changes.

The first staging Go1.27.1 race execution passed 14 top-level tests/29 pass events including subtests, with zero failures; one vet execution also passed. These comprise seven prototype tests, five prior independent synthetic tests and two archive/enum tests. [PORT-LEDGER-64.v1.json](PORT-LEDGER-64.v1.json) records actual attempts, original/ported source pins, publication scope and limits. These are synthetic API checks, not approval of live eligibility, fitting/readiness or performance. AI assisted implementation/review and ordinary collaboration cost was not measured. Separate judge model/API, training, weights and protected-final work remain zero.
