# Development data for a small claim model

The current immutable public release has **23 distinct semantic requests and67 candidate labels**. Download [Hugging Face `next60-23-finite-v1`](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-23-finite-v1). The separately adopted development pool has **27 requests and79 labels**; its last four requests are not yet part of that23-row file.

The target is a **claim or hint model** called repeatedly with very little CPU and memory. It supplies candidate hints; tests and verified evidence determine outcomes. We are expanding verified training material. These added data checks perform zero new training or model inference. LLM acceleration and Codex token savings remain unproven.

| Unit | Public immutable release | Separately adopted pool |
|---|---:|---:|
| Distinct semantic requests | 23 | 27 |
| Candidate labels | 67:23 positive,44 negative | 79:27 positive,52 negative |
| Frozen input variants | 113 | 132 |
| All original observations | 335 | 392 |
| Selected candidate observations | 331 | 388 |
| New corpus Fits | 0 | 0 |

Input variants and candidate executions are not independent requests. All public rows are `development_train`; selected weights are1. They reuse humanize group76, pflag/Cobra group77 and mapstructure group78. The source connections for the later Afero/retryablehttp/INI requests and existing helpers must be checked **before the next materialization**. We do not assign invented numeric groups or claim unseen-source generalization.

## Download and read

Choose HF config `development` and split `train`. The actual file is `releases/next60-23-finite-v1/next60-development-twentythree/data/train.jsonl`. Root `data/train.jsonl` is the original two-request history.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  --type dataset \
  --revision 0d964a547708db596b13b0eae18d2c93dd3e3ac4 \
  --include "releases/next60-23-finite-v1/next60-development-twentythree/data/train.jsonl" \
  --local-dir ./next60-23
```

The file is **31,493bytes**, SHA-256 `39edb1bb60e88d56cb2fec271b511a5ce09a0ff8bd33ce21d0dda7defb3ce362`. Its first28,800bytes preserve the prior21 rows exactly. See [the data guide and actual reader proof](https://github.com/teamswyg/laya-tools/tree/67319e639b52d302292a0abdbbd83149c1fb0d99/experiments/short-claim/next60-development-twentythree).

[The Go reader](https://github.com/teamswyg/laya-tools/tree/67319e639b52d302292a0abdbbd83149c1fb0d99/pkg/shortclaimdata) reads one row at a time, bounded to16KiB and8 candidates. It owns fixed arrays and immutable strings without locks.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

Import `bytes` and `github.com/teamswyg/laya-tools/pkg/shortclaimdata`. Only **request and candidate text** enter model features. IDs, sources, groups, revisions and finite scope are provenance; labels and weights are supervision. The reader invokes no scoring or training. Array ownership is not a measured speed improvement.

## Actual verification

[PR119](https://github.com/teamswyg/laya-tools/pull/119) passed [all four required CI jobs](https://github.com/teamswyg/laya-tools/actions/runs/37093356839) at exact head `9d4b0b5e39628a8d7a2bb50a3eb9f9b018ccd0ef`. GitHub Actions merged `67319e639b52d302292a0abdbbd83149c1fb0d99` with the same source tree. The two new Go checks reproduce saved finite comparisons, regenerate23 rows and verify actual project reader correspondence. Existing Laya native-inference CI is a separate step.

At immutable HF commit [`0d964a547708db596b13b0eae18d2c93dd3e3ac4`](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/0d964a547708db596b13b0eae18d2c93dd3e3ac4), we downloaded and verified all **401 owned files,2,989,766bytes**, including399 current manifest payload checksums. The402 remote files contain one additional HF-managed `.gitattributes`, unchanged from21. Use root `FILE-MANIFEST.v6.json` and `SHA256SUMS.v6`; historical inventories are checked at their historical tags. Previous2/3/7/16/21 tags did not move.

The viewer returned HTTP200 with23 observed rows, identical order and every field, `truncated=false` and zero truncated cells. Its response has no commit or total-row-count field; current viewer correspondence is separate from pinned all-files proof. The first HTTP500 response is preserved and the subsequent actual200 response supplied the accepted rows. The first actual Go reader run also matched23 dispatches, returns and rows with46 durable before/after checkpoints.

The two requests newly included in23 concern cumulative string-slice entry limits and validation retaining all errors. Ten inputs×three candidates produced30 actual observations:23 satisfied,7 unsatisfied and0 unknown aggregate outcomes. Detailed conditions were125 true,15 false and1 unknown. That error-code unknown is retained. Four earlier writer-baseline observations remain excluded from training selection. Unknown is not converted to negative.

The two separately adopted requests concern `IOFS.Sub` name validation and byte-slice request-body snapshots. Their27 first observations yielded15 satisfied,12 unsatisfied and0 unknown aggregate outcomes; detailed conditions were57 true,16 false and2 unknown. Two Sub error-type unknowns are preserved alongside known contradictions. Each reference satisfied the frozen inputs and other candidates had known counterexamples, supporting separate Root finite-development adoption. This does not guarantee all paths or body types. Unmeasured original nested and startup calls remain null.

## Next steps

INI quoted values and section-deletion bounds also completed first actual observation, saved-result comparison and separate adoption over five fixed inputs each. Observation counts were9 satisfied/6 unsatisfied and10 satisfied/5 unsatisfied respectively, retaining two unknown predicates in each. State unavailable after an original getter panic remains unknown; public occurrence order is separate from private indexes. Neither concurrent atomicity nor exhaustive parser-option coverage is established. See the [detailed records](https://github.com/teamswyg/laya-tools/tree/research/next60-ini-two-qualified-and-hf23-120/experiments/short-claim/next60-ini-two-actual-observation).

The remaining three requests concern total INI input-byte budgets, bounded file reads and exclusive file writes. Frozen inputs and expectations precede first native observation, saved-result comparison and separate adoption. Preparation alone does not increase the qualified count.

Position bias remains substantial. Always choosing candidate index1 gives20/23≈87.0% per public request; always predicting negative gives44/67≈65.7% per label. The separate27-request pool gives24/27 and52/79 respectively. These denominators differ and neither control is model accuracy. Position permutations, lexical controls and connected-source separation are required before accepting improvement.

**No new corpus Fit starts before30 qualified semantic requests.** The later60-request checkpoint, protected2,400 requests per claimed domain and5% utility guard against simple controls remain. Historical79-row data,3 corpus Fits,3 logical models and failed inactive models are unchanged. We activate a hint model only after evidence supports utility. Follow [issue19](https://github.com/teamswyg/laya-tools/issues/19).

Own documentation, annotations and source use Apache-2.0; full upstream notices retain their original licenses. Public data exclude raw upstream bodies, weights, private inputs, credentials and raw journals. Unresolved historical join ancestry remains excluded; this is not blanket model-ancestry clearance.

[First two requests' training data](Native2-Training-EN) · [First native observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)
