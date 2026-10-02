# Turning code behavior into training labels

A small model that proposes an order for checking candidates needs grounded examples of their behavior. We verified public Go code, built candidates with different behaviors for the same requests, and added three labeled training requests. Model fitting and improvements in actual usage cost remain separate experiments.

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

## What changes in the next fit

The original76 requests retain their exact text, candidate order, truth, loss masks and roles. Only the three new requests join training. UUID Parse, Scan and their variants stay together in one connected group; Ordinal and its variants stay in another. Roles across the whole corpus are not reassigned.

The new typed manifest contains79 requests and235 candidates. Conversion into Go training inputs and verification that the original prefix stays unchanged have not yet executed. The next model is a fresh sibling using the first FP32 fit's configuration. The experiment changes data first, keeping the seed, features, objective and quantization unchanged.

This lightweight branch uses our Go features and small model. These results do not represent fine-tuning the Laya backbone or executing GPU inference. Router integration, further ternary compression and model activation each require their own evidence.

## Time and memory

The32-fixture observer took about2.56 seconds with maximum RSS18.70MiB. The72-candidate observer took about2.84 seconds with maximum RSS18.63MiB. Each ran once with no automatic retry. Whole-process measurements include startup, source checks, JSON output and disk synchronization. Per-function memory and model-inference latency were not measured.

Existing preparation budgets remain a256MiB soft Go heap limit,256MiB observed process RSS budget,300-second outside execution timeout, and64MiB total new data/model/numeric records. The heap setting is not an OS-enforced memory cap.

An additional storage audit identified67,943,460 bytes in a subset of retained files including separate copies. This lower bound exceeds64MiB by834,596 bytes. The historical plan did not distinguish lifetime writes from current retained storage, so new fitting is on hold while a follow-up resource plan defines accounting explicitly. Preserve the over-budget finding; identical hashes do not make separate copies free.

## How the result will be judged

The original validation data and four cheap controls stay fixed. The existing development-utility gate requires at least5% fewer simulated checks than the best single control, without reducing Top1 or Top3 correctness. Results on the three new training-origin requests are recorded separately.

The old validation outcomes have already been inspected, so the next result is a development regression comparison. A separate final set of at least2,400 fresh semantic requests has not been acquired.72 candidate executions are not72 user requests, and the79 development requests are not a final evaluation. Both existing learned models failed earlier utility gates and remain inactive; these observations did not change that state.

Original artifacts, plans and failures are preserved. See the [candidate preparation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/conversion-pilot-80), [original function observations](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-80) and [data preparation guide](Claim-Data-Preparation-EN) for provenance and limits. The UUID BSD-3-Clause and selected Ordinal MIT notices remain intact. Model files, host paths and raw process logs stay out of Git.
