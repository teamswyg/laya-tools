# Packed78 v2 source delta and saved results review

No concrete semantic/correctness blocker was identified within this review. Packed v2 scoring still took1.21–2.68 times as long as the reference array scorer within this invocation. Original v1 source, failures and metrics remain unchanged. The reviewer authored neither v1 nor v2 and is nonblind through earlier role62, input and observer work. The v2 author is task_expansion_52; the reviewer did not repeat saved author tests.

## Change scope and arithmetic contract

Kind selection moves from the per-feature coefficient call into a switch before Score's loops. All three branches retain feature order and duplicates, check valid indices, reconstruct the same FP64 coefficient and execute `sum += weight * value`. INT8 still calculates q*scale first. Absent ternary weights are+0 and still participate in multiplication, preserving0*Inf/0*NaN rather than replacing them with0. Scale is not moved after accumulation; features are not reordered or combined.

Invalid indices return ErrIndex and+0 just as v1 does, regardless of a finite/NaN partial sum. Nil/zero View still returns ErrView. New, validation, owned copying, private fields, Coefficient, uint16 prefix, OwnedPayloadBytes and format did not change. The metadata helper checked exact equality of all bytes outside Score, the test prefix containing fixture/feature functions, and benchmark source. Prior validation/ownership/prefix review therefore remains applicable to those unchanged sections. This is not proof of bit parity across all compilers.

Saved author race events show5 top tests/38 subtests/1 package pass, zero failures. Strengthened test source covers absent ternary×Inf, NaN payloads, signed-zero features, cancelling positive/negative Inf, and invalid indices following finite/Inf entries with a+0 error result across all nine fixtures. These are author synthetic checks on Go1.27.1/darwin arm64; the reviewer did not call tests, Decode, Score or trained models.

## Scoring within the v2 invocation

Each value is the median of five samples. Ratio is packed divided by legacy; positive time delta means slower. Both use the same266 public synthetic features, with setup outside score timing.

| Format | Legacy ns/op | Packed v2 ns/op | Ratio | Time increase |
|---|---:|---:|---:|---:|
| FP32 | 155.9 | 213.1 | 1.3669× | +36.69% |
| INT8 | 156.4 | 189.6 | 1.2123× | +21.23% |
| Ternary13 | 155.9 | 248.9 | 1.5965× | +59.65% |
| Ternary8192 | 156.1 | 418.8 | 2.6829× | +168.29% |

Both implementations have0 B/op and0 allocs/op in all40 score samples. The observed gap is smaller than v1's4.08–5.10 times, but v1 and v2 are separate invocations with unmeasured background load and independently uncontrolled startup state. This is not a causal attribution solely to kind hoisting or a guarantee on another CPU.

Construction legacy→packed medians are FP32 15849→6756 ns, INT8 8919→4223 ns, ternary13 9562→365.5 ns and dense12832→413.7 ns. Compared with legacy65536 B/op, packed reports41024/9536/1504/2656 B/op and2/2/3/3 allocs/op versus1. New source is unchanged; small v1/v2 construction differences are not attributed to a code optimization.

## Execution, resources and provenance limits

Saved records confirm one original root invocation/Start1/joined Wait/exit0/retry0 and16 cases×5=80 rows, preserving all original iteration counts. One successful stdlib metadata-helper execution with zero runtime failures verifies19 exact pins, test-binary Go1.27.1/darwin arm64/CGO0/trimpath with no VCS metadata, CPU1/256MiB soft heap/60-second outside guard/50-second testing timeout/run `^$`/count5/200ms/bounds. Preparation had one patch-context failure, one Perl edit syntax failure and one Go compile failure; pre-execution failed source is preserved. The helper ran once on the second Go-run attempt; no original experiment was repeated.

The combined process records real19.47/user18.55/system0.25 seconds, controller wall19.47491875 seconds and peak RSS11,288,576 B/footprint8,684,000 B. Both representations, setup and testing runtime coexist. Owned data lengths(legacy65536/packed32792,8216,1308,2330), cumulative allocated B/op, live heap and whole RSS are different measures. One RSS does not establish representation-specific RAM saving; soft heap is not an RSS hard cap.

Results cover fixed legacy→packed order and repeated warm loops on Apple M4 Pro only. No new model quality, training, LLM usage/cost saving, GPU/SIMD, serving/cold-start or production/default activation is approved. Private raw logs, binary, helper source, host paths and HF/trained-model contents are excluded from safe artifacts.

[Full numbers and raw hashes](METRICS.v2.json) · [Receipt](RECEIPT.v2.json) · [Ledger](ATTEMPT-LEDGER.v2.json).
