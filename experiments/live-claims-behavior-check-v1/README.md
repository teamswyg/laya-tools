# Live claim behavior check v1

[한국어](README.ko.md) · [Live demo setup](https://github.com/teamswyg/laya-tools/blob/main/docs/live-claims-demo.en.md)

This bounded development diagnostic used **24 new original public-safe synthetic messages**, 12 Korean and 12 English. An AI author fixed three per-head semantic hypotheses for every text before inference. They are **not human gold, final evaluation, or an operational accuracy estimate**. The 72 head outputs share these 24 inputs. A named scope can have completion=true while another scope remains unfinished; this does not claim that the whole task is complete.

All 24 texts and hypotheses were locked at `2026-10-08T01:20:55.092Z`. Each received exactly one real HTTP `POST /api/hints`: **24 attempts, 24 HTTP 200 responses, no retry or refill**, from `2026-10-08T01:21:41.187Z` to `2026-10-08T01:21:41.211Z`. Publication made **zero additional model calls**. The current model, training steps (4240), and thresholds were unchanged; no fit, model selection, or product-state update occurred.

The before/after status and every response pin the model to SHA-256 `cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b`, confidence floor **0.9**, margin floor **0.05**, temperature **1**. Status says `research_preview` and `semantic_quality_qualified=false`.

## Recorded results

Each row covers 12 outputs. T/F/U means true/false/unknown. “Accepted positive” means final=true after the fixed numeric floors. Discrepancies compare recorded outputs with the frozen **AI hypotheses**; final discrepancies include abstention and are not an error/accuracy rate.

| Locale | Head | Expected T/F/U | Raw T/F/U | Final T/F/U | Accepted positive | Expected≠raw | Expected≠final |
|---|---|---:|---:|---:|---:|---:|---:|
| KO | response_requested | 2/9/1 | 3/9/0 | 0/3/9 | 0 | 2 | 8 |
| KO | current_activity_claimed | 2/9/1 | 3/9/0 | 0/2/10 | 0 | 2 | 9 |
| KO | completion_claimed | 3/8/1 | 3/9/0 | 0/3/9 | 0 | 2 | 8 |
| EN | response_requested | 2/9/1 | 3/9/0 | 1/3/8 | 1 | 2 | 7 |
| EN | current_activity_claimed | 2/9/1 | 4/8/0 | 1/6/5 | 1 | 3 | 5 |
| EN | completion_claimed | 3/8/1 | 5/7/0 | 0/4/8 | 0 | 2 | 7 |

Across 72 outputs: **2 true, 21 false, 49 unknown; 2 accepted positives; 13 expected≠raw and 44 expected≠final discrepancies**. All 49 final unknowns came from `low_confidence`; no raw winner was semantic unknown. A final unknown matching an expected semantic unknown therefore does not establish recognition of semantic uncertainty. The two accepted positives were EN09 response_requested (expected true) and EN04 current_activity_claimed (expected false).

EN04 is a concrete limitation:

> I am not editing the walking-tour map now. That task is paused.

Its current-activity hypothesis was **false**, but raw and final were **true**, with confidence **0.946169690250911**, margin **0.9008599783475525**, and probabilities T/F/U **0.946169690250911 / 0.04530971190335854 / 0.008520597845730313**. A negated current report passed the thresholds as positive; threshold passage is not semantic-quality qualification.

For these **24 sequential localhost POSTs only**, HTTP elapsed min/median/p95/max was **0.289 / 0.455 / 1.354 / 3.249 ms**; server-reported inference was **2 / 10 / 18 / 19 µs**. p95 uses nearest rank. HTTP time includes body consumption; `inference_us` measures the server's narrower scope. Original recording used Node built-in `fetch` and `performance.now`. These are observations of this run, with no throughput or hardware-resource claim; no Go timing parity is claimed. Unrounded numbers remain in the evidence.

## Evidence and controls

- [inputs.json](inputs.json): all 24 unchanged original texts, hypotheses, rationales, semantic definitions, and authorship dates.
- [responses.json](responses.json): complete original response bodies, parsed predictions, headers, byte counts, dates, and HTTP timings. [status-receipts.json](status-receipts.json) preserves both pin receipts; [results.json](results.json) contains the exact text-free arithmetic.
- [pre-inference-lock.json](pre-inference-lock.json), [provenance.json](provenance.json), and [exclusion.json](exclusion.json): lock chronology, original retained-artifact SHA pins, declared public metadata projection, and permanent exclusion of every original and variant from **final400/final1200/final2400, all final human-gold evaluation, and future model selectors**. These 24 inputs remain development-only.
- [SHA256SUMS](SHA256SUMS): checksums of the public projection. Only recorder actor identity and explanatory metadata were projected; texts, hypotheses, actual responses, and numeric results remain unchanged. Original source artifacts were preserved. The bundle contains no host absolute paths, private content, or model weights.

From the repository root, this Go standard-library checker reads saved evidence only and makes **no network or model calls**:

```sh
go run ./experiments/live-claims-behavior-check-v1/verify.go
go vet ./experiments/live-claims-behavior-check-v1/verify.go
```

It verifies public checksums, 24 unique ordered captures, source/public pins, lock chronology, input/exclusion digests, full-body readback, locale/head counts, fixed-threshold arithmetic, discrepancy totals, and original timing summaries. It checks captured data; it does not replay the original Node recorder.

To make a **new** manual observation, use the linked demo's pinned download/build/run instructions, inspect `/api/status` for the exact pin and floors above, then send a JSON `{"text":"..."}` request once, without retries. For example:

```sh
curl --fail --silent --show-error http://127.0.0.1:8877/api/status
curl --fail --silent --show-error \
  -H 'Content-Type: application/json' \
  --data-binary '{"text":"I am not editing the walking-tour map now. That task is paused."}' \
  http://127.0.0.1:8877/api/hints
```

Any manual replay produces **new outputs and timing observations**, not the captured run above. If repeating the cohort, use the unchanged 24 texts in `inputs.json`, exactly once each, and retain errors rather than retrying or replacing inputs.
