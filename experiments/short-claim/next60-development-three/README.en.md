# Next development data: three requests, eight candidate labels

[한국어](README.ko.md) · [Go reader guide](../../../pkg/shortclaimdata/README.md)

One newly observed direct-rounding request is added to the two previously qualified network/hook requests. The development subset now contains **three distinct requests and eight candidate labels: three positive, five negative**. Fifteen input fixtures and 41 candidate observations are supporting evidence, not extra requests. The previous two lines remain byte-exact.

| Request | Verified scope | Candidates | Whole training group |
|---|---|---:|---:|
| Network value setting | Five frozen IPv4 inputs | 3 | 77 |
| Retrying failing hooks | Four integer-hook configurations | 2 | 78 |
| Direct rounding of represented floats | Six frozen inputs | 3 | 76 |

The rounding candidates passed actual observation and [independent saved-result review](../next60-ftoa-actual-observation/semantic-review/REVIEW.en.md). The precise candidate satisfies all frozen cases. Witnessed counterexamples support negative labels for the original and double-rounding candidates, each with sample weight one. Unavailable original error channels remain unknown. Qualification covers the fixed two/nine-place inputs, not the original broad 0–18-place API proposal.

These are three existing IDs from the 20-request draft catalog, finitely qualified for development training. They introduce no new catalog ID or independent source family. The rounding request joins the already training-exposed humanize Ordinal whole group 76. No unseen-source generalization or protected evaluation score is provided.

## Read with Go

Pass each line of `data/train.jsonl` to `shortclaimdata.LoadDevelopmentRow`. `Example.Input()` contains model text. `Supervision()` holds ordered fixed label/weight arrays; `Metadata()` holds provenance and scope. Use **only request and candidate text** for features. API names, identifiers, source, group, labels and execution outcomes must stay out of the feature path.

The [actual validation receipt](INPUT-VALIDATION.v1.json) records three reader loads and one legacy input validation, including exact agreement of the third row with the [pinned finite input](../next60-ftoa-actual-observation/INPUTS.v2.json). The [validator source](source/validate.go.txt) checks public-data shape and correspondence without fitting or scoring. CI checks both the historical two-row reproduction and this three-row validation.

The actual 79-request corpus and previous models are unchanged. A combined collection could contain 82 requests and 23 groups; it has not been projected, scored or fitted. Checkpoints remain **30 and 60** newly qualified development requests. No new fit is started on these three alone. Protected evaluation still targets 2,400 distinct semantic requests per claimed domain; that collection is not completed.

The [immutable three-request HF release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-3-finite-v1) passed pinned download of all 71 files and viewer verification: all fields in its three rows match the original data. See the [publication record](../publication-proof-114/HF-PUBLICATION.v2.json). The previous two-request tag remains unchanged. Publication verification proceeds alongside further data acquisition; no new fit has run.

Own prose and annotations use the repository Apache 2.0 license. The selected original [MIT notice](notices/humanize-MIT.txt) is retained separately; earlier BSD/MIT notices remain in the immutable prior release. This subset excludes upstream implementation bodies, binaries, weights, private prompts/paths and traces. It provides no blanket ancestral relicensing.
