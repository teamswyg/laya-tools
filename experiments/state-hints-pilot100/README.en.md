# Progress, completion-report and question hints: first 100-case development test

[한국어](README.md) · [Rubric](RUBRIC.md) · [Frozen plan](PLAN.json) · [Complete results](RESULTS.json) · [Issue #155](https://github.com/teamswyg/laya-tools/issues/155)

`riidolaya state-hint-pilot` compares a local trained model, a retained parent and optional rule controls on the same cases. A separate AI author created these 100 fictional development messages and labels before predictions. We evaluated the actual published Go weights. The command never trains, selects a winner, calls Riido or installs annotations.

There are 25 progress, 25 completion-report, 25 question and five cases for each of the other five meanings. Each language has 50 cases. Labels are author interpretations, not product ground truth or independent human adjudication. One hundred distinct family identifiers do not establish statistical independence. Rows are not renamed or translated filler variants.

## Actual results

| Same 100 cases | Correct overall | Proposals | Correct proposals | Wrong proposals |
|---|---:|---:|---:|---:|
| Go v0.2 |69|16|15|1|
| Parent Go v0.1 |31|7|5|2|
| Original rules |35|22|19|3|
| Speech-act rules |44|31|28|3|

Rules use unlearned one-hot selection scores, not calibrated probabilities. The model kept its stored temperature, confidence 0.9 and margin 0.05. Higher rule coverage and higher model accuracy are separate findings. No default model or rule was replaced based on these results.

| V0.2 display | Correct classification / true cases | Correct / created proposals |
|---|---:|---:|
| Progress report |21/25|6/6|
| Completion report |10/25|3/4|
| Question |18/25|6/6|

Correct display coverage is 15/75 (20%). Overall accuracy 69%, proposal precision 15/16 and correct-display coverage 20% have different denominators. Korean scored 38/50 and English 31/50. The prior 40-case result of 85% did not persist in this new 100-case test.

A concrete failure was:

> Do completed checks have to be repeated after only the description changes?

The author label is question, but the model proposed `completion_report` at 0.98825. The parent made the same error. High scores do not prove meaning or completed work. We retained the original text, label and failure. Further training must distinguish questions mentioning completion from actual assertions, requests, hypotheses, quotations and partial completion.

Eight-intent NLL improved from 2.9244 to 0.9607, and Brier from 1.0299 to 0.4471. Parent temperature 0.7 and child temperature 0.9 mean this comparison includes calibration. A separate four-category loss groups the five excluded meanings into none without renormalizing predictions or changing proposal gates. Ten-bin calibration diagnostics describe only this small balanced synthetic set, not product base rates.

## Run locally

Go 1.27.1 can build the general command or a smaller dedicated binary. Both call the same package. Download a chosen public model revision explicitly; there is no automatic download.

```sh
go build -trimpath -o riidolaya ./cmd/riidolaya
go build -trimpath -o riido-statehint-pilot ./cmd/riido-statehint-pilot

hf download JooYoon/riidolaya-statehint-go-v0.2 statehint.rsh --revision 81335eadd9f2753d3b86c932e9b00c636714e3b4 --local-dir ./models/statehint-v0.2

./riido-statehint-pilot --cases experiments/state-hints-pilot100/cases.jsonl --model ./models/statehint-v0.2/statehint.rsh --model-sha256 ae6761dc81501b39ae17f29b48f0f0a8b35305fe11e297b9a787af0f9636df05 --compare-rules --out ./new-report.json
```

`--parent` adds a retained model. `--cases-sha256`, `--model-sha256` and `--parent-sha256` can pin the files. Wrong checksums and untrained artifacts are rejected before predictions; there is no silent rule fallback. Output files have private permissions and never overwrite previous evidence. Omit `--out` for JSON on stdout. Reports do not echo text, but case/family identifiers and the complete input hash may be sensitive; keep real-data reports private.

## Resource measurement

Two separate macOS processes repeated the same 100 cases to compare wrappers. All probabilities, metrics, proposals and metadata matched exactly. Repeating the cases adds zero new independent samples.

| One four-control process | General riidolaya | Dedicated riido-statehint-pilot |
|---|---:|---:|
| Binary file bytes |11,875,234|4,390,322|
| Peak OS RSS bytes |20,463,616|12,566,528|
| Go heap sampled before JSON encoding |1,892,600|1,650,904|

Each process called two models 100 times each and two rules 100 times each. OS measurements include model loading, JSON, synthetic catalogs and reporting, excluding live network reads, long-running operation and GPU. Do not add Go heap to RSS. V0.2's first classification/synthetic-read/aggregation segment took 384,000 ns for 100 cases; startup, loading and JSON encoding are excluded, so this is not complete-request throughput. Raw logs remain private; [curated measurements](RESOURCES.json) are public.

## Boundaries and next PDCA

Fixed three-display metrics use arrays; all eight scores and the confusion matrix remain intact. Budgets are 2,400 cases, 8 MiB input, at most 1 MiB per model and 64-reference batches. Validate the full input and budgets before callbacks. The default reader, catalog and revision 1 are explicit synthetic fixtures, not substitutes for real work revisions.

Validate the original snapshot before narrowing bindings to three meanings. Preserve original mapping conflicts. Low-confidence cases retain missing/inactive annotation diagnostics; use `annotation_checked` as their denominator. Raw reader capability is separate from pilot-disabled state planning. Status-candidate interpretation is explicitly unchecked. An injectable reader API does not prove live Riido integration.

Next work expands to at least 400 distinct contexts, annotates completion questions/partial completion/requests/quotation contrasts and connects private reader revision evidence. Public models use only publishable licensed training material. These 100 cases are exposed and cannot serve as a fresh test for later model selection. No weights changed in this round.

Source CI is pending. The added CI step downloads pinned 32 KB Go weights and actually evaluates these cases on CPU. It uses a synthetic catalog, not the Laya backbone, GPU or model training. CI verifies code and compatibility, not production model quality.
