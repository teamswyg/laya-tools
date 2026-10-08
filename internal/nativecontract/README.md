# Joining original and amended review records

[한국어](README.ko.md)

`nativecontract.Join` checks that four review documents refer to the same material and reviewer. It distinguishes reconsidering unchanged material from independently reviewing an amended pair. It preserves the reviewer's declared verdict; it does not generate a semantic verdict.

Supply exact byte-size and SHA-256 pinned documents for the standard Review, FULL inventory acceptance, selected version and current disposition. `Expected` declares the original and selected material, producer, reviewer, read and check references. `RegisteredSourceIDs` supplies the registered IDs. Amended mode also requires the seven expected official keys. Original mode has no native key-list field and neither requires nor invents one.

```go
result, err := nativecontract.Join(nativecontract.Inputs{
    Kind:                nativecontract.Amended,
    Review:              reviewBytes,
    Acceptance:          acceptanceBytes,
    Version:             versionBytes,
    Disposition:         dispositionBytes,
    Expected:            declaredAnchors,
    RegisteredSourceIDs: sourceIDs,
})
```

Each byte input is a `PinnedBytes`: `File` supplies a relative path, size and SHA-256; `Bytes` supplies the actual document. Optional `Witnesses` additionally checks the public shapes of read start/result, check binding/report and retained held Review. The module never opens referenced paths, so the caller must separately prepare any required bytes.

Duplicate, unknown or missing keys, wrong types or schemas, changed bytes, conflicting references, reviewer mismatches and combinations that promote a hold to a pass return errors. Success copies the original document bytes and retains the old held references. Limits are 128 KiB per document, 8 MiB summed input, JSON depth 32, 256 items per collection and 400 registered IDs. At most nine input documents are supported, so the individual limits impose a smaller reachable total. Reference comparisons use sorted slices. No file, network, model or lock operations are introduced.

Only the two explicit native review formats are supported. This module does not convert every historical format or prove material meaning, rights or reviewer identity. `MeaningProven` and `TrainingEligible` remain `false` even on success. Whole-cohort QA and training admission are separate steps. Tests use authored synthetic metadata.
