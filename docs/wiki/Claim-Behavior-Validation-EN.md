# Turning code behavior into training labels

A small model that proposes a verification order needs grounded behavior examples. We added three verified public-Go requests and actually trained under the same recipe. The new model required31 checks versus the lexical control's27 and failed the improvement gate. See [full fit results](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/data-effect-fit-79/README.en.md) for measurements and limits.

## What was actually verified

| Check | Result | Meaning |
|---|---|---|
| Time conversion and command-line slice handling | All32 fixed inputs matched expectations frozen before execution | Error causes, time values, nil versus empty slices and changed state were distinguished. These fixtures have not become training requests. |
| UUID parsing/state changes and English ordinal candidates | Nine candidates × eight inputs: all72 matched fixed expectations | Both satisfying candidates and deliberately different behaviors were executed. |
| Supervision review | Three requests, three positive and six negative candidates | Task satisfaction, caption fidelity and observation completeness were reviewed separately. |

Matching an expectation does not make a candidate correct for the request. If we predicted that a candidate would discard an error and it did, its observation passes while its behavior fails the user's request.

## Which differences the data teaches

| Request | Satisfying candidate | Observed differences in other candidates |
|---|---|---|
| Return UUID parser errors and preserve partial parser bytes | Forward the original result and error | Replace the result with zero on error, or panic instead of returning the error. |
| Preserve existing UUID state on nil, empty or erroneous input and return errors | Scan from the existing state and forward errors | Clear state before scanning, or discard errors. |
| Use English ordinal `th` for endings11,12 and13 | Preserve the original ordinal rules | Use only the last digit to produce `11st`, or always use `th` to produce `1th`. |

A negative candidate's caption must still faithfully describe its code. A caption saying that errors are discarded can be useful training data when the code actually discards errors. Inaccurate captions are a separate issue. Truth came from the fixed request, source and actual behavior, rather than text similarity.

These labels are tied to eight fixed inputs per request and the pinned source revisions. They are not empirical proof over all possible inputs. Review was AI assisted and nonblind, with code and captions both visible; it was not an independent human-blind review.

## What changed in the data-addition fit

The original76 requests retain their exact text, candidate order, truth, loss masks and roles. Only the three new requests join training. UUID Parse, Scan and their variants stay together in one connected group; Ordinal and its variants stay in another. Roles across the whole corpus are not reassigned.

The new typed manifest contains79 requests and235 candidates. One Go projection and one Fit completed, preserving the original76 requests, original training numerical prefix and full validation/calibration parts. The new model is a fresh sibling using the first FP32 fit's configuration. Seed, features, objective and quantization were unchanged. It matched the previous model at31 checks and2/10 first-candidate correctness, failing to beat the lexical control's27 checks and5/10.

This lightweight branch uses our Go features and small model. These results do not represent fine-tuning the Laya backbone or executing GPU inference. Router integration, further ternary compression and model activation each require their own evidence.

## Time and memory

The32-fixture observer took about2.56 seconds with maximum RSS18.70MiB. The72-candidate observer took about2.84 seconds with maximum RSS18.63MiB. Each ran once with no automatic retry. Whole-process measurements include startup, source checks, JSON output and disk synchronization. Per-function memory and model-inference latency were not measured.

Existing preparation budgets remain a256MiB soft Go heap limit,256MiB observed process RSS budget,300-second outside execution timeout, and64MiB total new data/model/numeric records. The heap setting is not an OS-enforced memory cap.

An additional storage audit identified67,943,460 bytes in a retained subset, exceeding the historical64MiB limit. Revalidation of2776 paths in a fixed scope yielded206,707,066 bytes; this is not a whole-machine, remote or deleted-history total. Before execution, a separate versioned policy fixed512MiB retained storage and64MiB for the new stage, preserving the historical excess. Separate paths count independently even when their hashes match.

The new projection worker took2.090s with33.45MiB OS peak RSS; the input-check/fit/comparison/persistence worker took1.244s with35.08MiB. These are actual whole-worker CPU-one measurements, distinct from Go heap or original Laya GPU measurements.

## How the result will be judged

The original validation data and four cheap controls stay fixed. The existing development-utility gate requires at least5% fewer simulated checks than the best single control, without reducing Top1 or Top3 correctness. Results on the three new training-origin requests are recorded separately.

The validation outcomes were already inspected, so the new result is a development regression comparison. Fresh protected evaluation of at least2,400 semantic requests per domain remains unmet.72 candidate executions are not72 user requests, and the79 development requests are not a final evaluation. The new model and the two earlier learned models failed utility gates and remain inactive. The next PDCA expands independent semantic requests and source groups instead of immediately retuning this small exposed validation part.

Original artifacts, plans and failures are preserved. See the [candidate preparation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/conversion-pilot-80), [original function observations](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-80) and [data preparation guide](Claim-Data-Preparation-EN) for provenance and limits. The UUID BSD-3-Clause and selected Ordinal MIT notices remain intact. Model files, host paths and raw process logs stay out of Git.
