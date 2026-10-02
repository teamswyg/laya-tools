# Wrap/Marshal 75: public observation records

We compared 24 predeclared Wants with actual execution of two functions from public Go sources. All 12 `wordwrap.WrapString` probes and all 12 `godotenv.Marshal` probes matched. There were 24 original returns, no differences or panics, one child execution, and no retries.

This was a native Go 1.27.1 CPU program. It did not run a Laya model or a GPU experiment. The records connect finite inputs, source behavior, and sealed expectations. They do not add 24 independent requests or training labels, or demonstrate model generalization, caption qualification, training readiness, or usage savings. The corresponding qualification flags remain false in the original records.

| Observation | Result | Scope |
|---|---:|---|
| Sealed Wants matched | 24 / 24 | 12 Wrap + 12 Marshal probes |
| Differences / panics | 0 / 0 | Only the fixed finite inputs |
| Executions / retries | 1 / 0 | One child process |
| Maximum RSS | 18,677,760 bytes = 17.8125 MiB | Whole-child peak |
| Peak footprint | 16,089,616 bytes | A different OS metric from RSS |
| User / system CPU | 0.01 / 0.04 seconds | Whole child |
| OS real time | 0.87 seconds | Whole child |
| Controller elapsed time | 0.879832625 seconds | Separate outside record |

The OS measurements include package initialization, preflight, wrapper and function calls, and result and checkpoint writes. Pure function latency was not measured. One CPU and a 256 MiB Go soft heap setting are not a hard limit on total RSS. This single observation does not establish a performance gain or generally low resource use.

Start with the [sealed Wants](finite-want.v1.json), [actual results](observations/results.json), [outside ledger](root/ROOT-ACTUAL-LEDGER.v1.json), and [saved-result QA](saved-qa/SAVED-RECORD-QA.en.md). Null errors, nil versus empty maps, before/after input maps, and a 64-digit integer string remain distinct. The final and last partial results are byte-identical.

The [source-before-Want review](source-peer/SOURCE-ORACLE.v1.en.md) was performed by a different AI collaborator from the observer author. The [controller reviewer](controller-peer/FINDINGS.v1.en.md) authored the observer, and the [saved-result checker](saved-qa/SAVED-RECORD-QA.en.md) authored the controller. These authorship boundaries and prior exposure are not independent model evaluation or blind validation. Zero separate model calls does not mean that collaboration-agent costs were zero.

The [public-file manifest](PUBLICATION-MANIFEST.v1.json) distinguishes exact copies, derived copies, and omissions. An exact copy preserves the original bytes and SHA. The two derived execution plans replace only two private path values with public markers and are non-executable records. Their original SHAs remain explicit; derived hashes do not replace the actual execution-plan binding. Binaries, tests and helpers containing private paths, and raw logs are omitted. Preparing this bundle neither reran the observer nor published anything remotely.

Original sources and complete MIT notices are retained at PR99's fixed commit [0a25491](https://github.com/teamswyg/laya-tools/tree/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream). Preserve the full [go-wordwrap/LICENSE.md](https://github.com/teamswyg/laya-tools/blob/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream/go-wordwrap/LICENSE.md) and [godotenv/LICENCE](https://github.com/teamswyg/laya-tools/blob/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream/godotenv/LICENCE) notices. [Source references and hashes](SOURCE-REFERENCES.v1.json) bind the actual retained paths to original byte hashes. The `.go.txt` files are source-review records, not a product execution interface.
