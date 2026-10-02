# Three-request data addition: experiment 79

**Adding verified training data under the same recipe did not improve practical utility.** The new model remains an inactive failed-development artifact. Default model selection is unchanged.

The small claim model suggests which candidate to verify first. This experiment ranks short descriptions of public Go code behavior. The number of candidates checked before reaching a known correct candidate is a proxy for cost. It is not a measurement of real LLM calls, tokens, or Codex subscription savings.

## Change and comparison

We added three semantic requests about UUID parsing, extraneous-character handling, and ordinal suffixes to the training part of the original 76 requests. Each has three candidates. Previously frozen observations across 72 finite execution inputs and code/caption review established three positive and six negative candidate labels. Those 72 inputs are not 72 independent semantic requests.

The original request text, truth/unknown states, roles, sample weights and masks were preserved. The original training numerical prefix and the full validation/calibration parts were preserved too. The source-bound worker checked floating-point bits and nil/empty states; Root separately compared saved JSON request metadata, training prefixes and validation arrays. There was one Project invocation and one Fit invocation.

Training restarted from scratch with the same 8,192 FP32 coefficients, seed 1729, 50 epochs, batch 128, learning rate 0.1, L2 0.0001, and BCE loss. There was no warm start, outcome-driven threshold change, seed search or loss change. The same earliest strict minimum validation-NLL epoch rule was used.

## Actual results

The validation part contains 20 requests. Utility uses the 15 with known truth: ten answerable and five with no answer. Five unknown requests remain excluded from accuracy denominators.

| Same validation part | Total candidate checks ↓ | Correct first candidate / 10 ↑ |
| --- | ---: | ---: |
| Lexical order control | 27 | 5 |
| Previous 76-request BCE model, experiment 71 | 31 | 2 |
| New 79-request BCE model, experiment 79 | 31 | 2 |

The new model required **14.8% more checks** than the control. The frozen gate required at most 25 checks while preserving first-candidate accuracy, so it failed. It did not improve on the previous model either. Validation NLL worsened slightly from 0.66579769 to 0.66623504.

On the three new training-origin requests, the model required six checks and got one first candidate right; the lexical control required four checks and got two right. This is a training-origin diagnostic, not a generalization gate. With only three candidates, top-3 correctness is not useful evidence of success.

| Actual Mac CPU-one worker | Whole worker time | OS peak process RSS |
| --- | ---: | ---: |
| 79-request projection | 2.090 s | 33.45 MiB |
| Input verification, fit, same-validation comparison and persistence | 1.244 s | 35.08 MiB |

These times include file verification and persistence. OS RSS is distinct from Go heap. This is a small Go CPU model, not original Laya GPU execution/fine-tuning or 1.58-bit training. A 32,792-byte model file does not imply a process memory footprint of that size.

## Resource accounting

The original 64 MiB retained-storage limit was exceeded when historical copies were counted. Even the confirmed subset was 67,943,460 bytes; we do not retroactively mark it passing. Revalidating 2,776 paths in the fixed scope of existing serialized records, numerical data and tiny-model copies yielded 206,707,066 bytes. This is not a whole-Mac, remote-storage, deleted-history or lifetime-write total.

We froze a separate [versioned resource policy](actual/RESOURCE-AMENDMENT.v1.json) before execution. It allows 512 MiB retained bytes within the governed scope and 64 MiB for the new stage's data, models and numerical records. CPU one, a 256 MiB soft Go heap target, an observed 256 MiB OS RSS gate and a 300-second limit per worker were unchanged. A single result path served as checkpoint and final output; no large projection or stdout mirrors were created. The Go heap target is soft, and OS RSS is a terminal observation gate.

## Next PDCA

This is evidence that **adding these three verified requests did not improve utility**. It does not establish that data volume alone caused the failure; feature representation and objective alignment remain hypotheses. We will acquire more distinct code behaviors, request conditions and failure cases instead of repeatedly tuning against this small exposed validation part.

Follow the [acquisition plan](../../../docs/golden-set-acquisition.en.md) and [scale criteria](../../../docs/golden-set-scale.en.md). Input variants, paraphrases and training rows do not count as 2,400 semantic requests. Fresh protected evaluation of at least 2,400 requests per domain remains unmet. Exposed development validation cannot replace it. Ambiguous truth remains unknown.

Public evidence includes [actual numerical results](actual/results.json), [outside OS measurements](actual/FIT-OS-RESOURCE.v1.json), [projection preservation QA](actual/PROJECT-RUNTIME-QA.v1.json), and [independent source review](review/FIT-SOURCE-REVIEW.v2.json). Zero actual invocations in the [preparation archive](preparation-v2/README.en.md) describe that historical snapshot. Later actual execution is recorded here and under `actual/`. Model bodies stay out of GitHub; a separate inactive Hugging Face archive does not activate a default model.
