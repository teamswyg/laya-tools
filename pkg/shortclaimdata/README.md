# A small Go reader for public development rows

`pkg/shortclaimdata` separates one public development request into model text, supervision and source bookkeeping. It simplifies future training code and reduces the chance of accidentally turning labels or source IDs into text features.

- `Example.Input()` returns the existing `shortclaim.ValidatedInput`: checked and normalized request/candidate text. Candidate IDs remain bookkeeping.
- `Example.Supervision()` returns label and weight arrays in candidate order. Every candidate requires an explicit Boolean label and the integer literal `1` as its weight.
- `Example.Metadata()` returns the request ID, connected group, source family/revision and finite observation scope. Keep these fields out of encoder text.

Input provenance is fixed to `finite-development-v1`. Original source bookkeeping remains in metadata; the adapted input is not byte-identical to the original input file.

## Use

Import `io` and the project's `pkg/shortclaim` and `pkg/shortclaimdata` packages. This function reads a row without running inference or training.

```go
func readRow(r io.Reader) (shortclaim.ValidatedInput, shortclaimdata.Supervision, error) {
    example, err := shortclaimdata.LoadDevelopmentRow(r)
    if err != nil {
        return shortclaim.ValidatedInput{}, shortclaimdata.Supervision{}, err
    }
    return example.Input(), example.Supervision(), nil
}
```

Supply exactly one JSON object, at most 16 KiB, with 1–8 candidates. For JSONL, pass one bounded row at a time. Only `development_train` is accepted. Missing fields, duplicates, case aliases and invalid Unicode are rejected. Fixed errors retain no supplied text or paths.

The reader owns immutable strings and fixed arrays. Labels and weights occupy separate arrays, and accessors return value copies. Editing a returned array cannot change the original example. The reader retains no shared mutable maps/slices and uses no locks.

## What has been checked

Local race tests on Go 1.27.1 passed 46 test/subtest records with zero failures/skips; vet also passed. A separate source author reviewed the implementation and found no blocker. The [validation record](../../experiments/short-claim/publication-proof-111/READER-LOCAL-QA.v1.json) separates these local checks from the change's CI.

Metadata-invariance tests check that changes to bookkeeping and labels leave normalized text unchanged. They do not measure model quality, speed, memory savings or GPU performance.

The current local [development data](../../experiments/short-claim/next60-development-twentythree/README.en.md) contains23 requests and67 labels (+23/−44). The [actual correspondence check](../../experiments/short-claim/next60-development-twentythree/INPUT-VALIDATION.v1.json) records23 reader calls, returns and value matches. The last verified [HF release contains21 requests](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-21-finite-v1); consult separate publication evidence for23-request release status. Earlier subsets remain preserved. The113 finite fixtures and335 original observations do not increase request counts or weights;331 observations are selected for supervision, without converting unknowns to false. This is not a2,400-request protected evaluation set. Structural acceptance grants no rights, evaluation role, semantic truth or training authority.
