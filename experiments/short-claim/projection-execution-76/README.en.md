# First preparation of all 76 requests

The first Go preparation of the complete 76-request development set succeeded. Fitting and ranking utility are separate observations. This preparation does not run the Laya encoder or a GPU.

| Role | Requests | Candidates | Prepared fit rows | Zero-weight rows | Unknown candidates |
|---|---:|---:|---:|---:|---:|
| train | 44 | 130 | 91 | 18 | 39 |
| validation | 20 | 60 | 45 | 1 | 15 |
| calibration | 12 | 36 | 0 | 0 | 9 |

The original 72 texts, labels, candidate order and masks remain unchanged. Four new requests use separately identified proposed supervision from finite public upstream observations. Both new source groups entered train; validation therefore does not demonstrate transfer to those new sources. All 63 candidates from 21 unknown requests retain null labels and remain outside the fitting loss. The 19 known zero-weight rows remain physically present.

Direct Validate returned 76 times, Project once, with 272 feature scans. Owned returned payload was 1,928,154 B (about 1.84 MiB), whole-child peak RSS 50,036,736 B (about 47.72 MiB), and outside wall time 0.475047292 s. CPU 1, soft Go heap 256 MiB and outside 300 s timeout were configured. This is one preparation observation, not Go-heap/GPU memory, resident inference latency or LLM savings.

The original result is 8,390,462 B, SHA256 `f68bff5f48747a66f038c17262568c1bd91251a8c81b3f46f7ae5090c8fe4b99`. Full feature arrays, private plans, binaries and logs are excluded from Git. The [actual ledger](actual/ROOT-ACTUAL-LEDGER-76.v1.json) and [saved-array review](runtime-qa/FINDINGS-RUNTIME-76.v1.en.md) distinguish execution from review. The reviewer differs from the preparation-worker author but previously authored claimfit and the 69/71 drivers; this is a disclosed nonblind consistency review. Feature semantics were not independently extracted again.

`preparation/execution-plan-76.v1.public-view.json` is a preparation-time public derivative. It does not replace the actual frozen private plan, whose SHA is in the actual ledger. Historical preparation states remain unchanged.

This execution made zero fits, new role assignments, new original labels, paid model calls or protected-final reads. A subsequent primary FP32 fit uses the original fixed settings and all validation candidates in the declared scope, with separate records. These 76 development requests form 19 related whole groups; they do not replace the protected final target of at least 2,400 distinct requests per domain. AI-assisted preparation and review cost remains unmeasured.

[한국어](README.ko.md) · [Whole roles 70](../combined-role-execution-70/README.en.md) · [Role review](../role-audit-70/README.en.md)
