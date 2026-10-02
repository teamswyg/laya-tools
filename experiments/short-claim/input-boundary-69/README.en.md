# Actual check of the existing input bounds69

[한국어](README.ko.md) · [Actual result](results-69-input-boundary.v1.json) · [Independent checks](FINDINGS-BOUNDARY-69.v1.en.md)

The existing short-claim contract allows512bytes per request/caption,32 normalized words and at most8 candidates. One audit checked all72 original requests and216 captions: all are supported. Maximum normalized word counts are32 for requests and32 for captions. Text and limits were unchanged.

One process made72 direct `Validate` calls and288 explicit length-comparison `Normalize` calls. Normalizations internal to `Validate` were not individually instrumented. This result establishes input support; it performs no training projection, fitting or new candidate ranking. Independent copied normalization math and original UTF-8 SHA checks are retained separately.

The [frozen plan](frozen-plan-69.v1.json), [pre-execution reservation](INVOCATION-69.v1.json) and [QA ledger](LEDGER-BOUNDARY-69.v1.json) are exact copies. The root separately authored the new [copy ledger](QA-COPY-LEDGER-92.v1.json). Local helpers and raw logs are excluded. The plan contains only a relative helper name and hash; this folder does not redistribute the complete execution harness. A subsequent Go projection will use the original text in a separate execution.
