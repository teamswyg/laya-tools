# Actual role assignment for the original 72 requests66

[한국어](README.ko.md) · [Actual result](results-66.json) · [Independent checks](FINDINGS-RUNTIME-66.v1.en.md)

The CI-verified Go module ran once with fixed seed1729. Connected code, caption and request families stay together; all72 requests,216 candidates, truth and ordering are retained. The16 labeled groups are assigned10/3/3 to train/validation/calibration. A separate unknown-only group contributes train provenance, giving11/3/3 total groups. Unknowns do not become training negatives.

| Role | Requests | Candidates | Unknown requests |
|---|---:|---:|---:|
| train |44|132|14|
| validation |16|48|4|
| calibration |12|36|3|

First official attempts1, failures0, retries0. The OS observed0.41seconds real time and17,448,960bytes(16.64MiB) peak RSS for the whole child process. Go heap, model inference and GPU measurements are separate and absent here. `TrainingReady=false` records that role metadata alone does not complete training readiness; it does not revoke existing user authorization.

Read the [root reservation](ROOT-INVOCATION-RECEIPT-66.v1.json), [actual ledger](ROOT-ACTUAL-EXECUTION-LEDGER-66.v1.json), [worker ledger](worker-attempt-ledger-66.json), [original copy ledger](PUBLIC-COPY-LEDGER-66.v1.json) and [QA copy ledger](QA-COPY-LEDGER-92.v1.json). The earlier [preparation](../role-execution-preparation/README.en.md) retains its historical unexecuted state. Mechanical QA is nonblind and performed by a prior role implementation author; it is not independent data authorship.
