# Source80 finite expectations for 72 candidate positions

[WANTS.v1.json](WANTS.v1.json) seals **pre-execution expectations** for the original 24 inputs across nine candidate functions, including the candidates' intended failures. These are not labels, role assignments, or actual candidate observations.

Order is goal 0..2 → candidate 0..2 → original fixture 0..7: `ordinal=g*24+c*8+f`, `original_fixture_ordinal=g*8+f`. Original fixture IDs, full inputs, and Want/Got/input-observation references remain connected. Metadata checks bind 20 original source pins, the original 24 Want/observation rows, and nine candidate byte/line spans. Future model inputs remain request+caption; code, IDs, Wants, saved Got, and source metadata are not hidden features.

| Goal and candidate | Frozen expectations |
| --- | --- |
| ParseZeroOnError | The three errors return zero UUID while retaining the error. |
| ParsePreserve | Forward original Parse values/errors, preserving the distinction between the raw32 late-error `…ddfe00` and standard36 `…dd0000`. |
| ParsePanicOnError | The three errors become error-valued panics. Returned UUID/error channels are null; only panic presence/type remain. panic_text is null and no extra Error/String/Format is invoked. This does not reproduce MustParse's string-valued panic. |
| ScanPreserve | Preserve ff..00 for nil/empty/error; return decoded bytes on success. |
| ScanClearBefore | Fixtures 0,1,2,3,6,7 return zero; successful fixtures 4,5 return decoded bytes. before is the **ff..00 wrapper input**, not its zero local receiver immediately before inner Scan. after is the wrapper's returned UUID. |
| ScanSuppressError | Retain the returned state while replacing the two errors with nil; error_type/message become null too. |
| OrdinalLastDigit | `0th,1st,2nd,3rd,11st,12nd,13rd,112nd`. |
| OrdinalConstantTH | `0th,1th,2th,3th,11th,12th,13th,112th`. |
| OrdinalPreserve | Retain all eight original outputs. Negative inputs and whole-package humanize equivalence are outside scope. |

The existing pure.Channels has **11 JSON fields**. It stays unchanged; `Want_observed.nonstring_panic_text_unobserved` separately records the unobserved error-valued panic text, true only for the three Parse panic rows. Expected dispatch invokes Error once only on returned nonnil errors; suppressed errors and panic payloads add no call. Input metadata distinguishes nil interface, typed nil bytes, and nonnil empty bytes.

Expected primary 72 = normal returns 69 + panics 3; explicit returned Error 10 = UUID URN 2 + stdlib errorString 8. Thus the **candidate-level tracked expectation is 82, maximum 120 (72+48)**. Source-inferred inner original Parse/Scan/Ordinal calls 56 (24/24/8) and owned Ordinal 16 are separate static estimates, not measured counters. `56+48=104` uses a different denominator. Four UUID startup source sites, Scan recursion, and formatting internals have uninstrumented/null actual return counts.

Author `semantic_review60_prep` prepared original Source80 but did not author these candidate functions or the observer. This is AI-assisted, nonblind authorship with prior base Got exposure. Assembly uses the frozen base Wants and the manually authored [LITERAL-OVERRIDES.v1.json](LITERAL-OVERRIDES.v1.json); saved Got only verifies original references. This is neither a native behavior simulator nor independent training-truth qualification.

One stdlib-only metadata assembly attempt passed with zero failures; a separate JSON null/count check passed. Original APIs/import/init/compile/tests/native/observer reruns, Features/Score/Fit/HF, and shared changes remain zero. Got/truth/role/weight/label/mask/acceptable remain null in all 72 rows, with actual parent creation false. Three planned requests and two upstream families do not establish 72 independent parents, a 2400-request final set, or training readiness. This AI-assisted collaboration's cost was not measured.

Licensing distinguishes the full retained UUID Darwin 15-file source + original go.mod/BSD-3-Clause from the humanize Ordinal-only source slice + original go.mod/MIT. The existing 20 pins and full notices remain referenced. No full humanize/WTFPL scope, additional acquisition, or blanket clearance is claimed. Root owns independent peer review and any separately frozen execution.
