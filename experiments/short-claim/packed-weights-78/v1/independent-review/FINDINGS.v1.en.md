# Packed score78 source and saved synthetic record review

No concrete execution blocker was identified within this review. The source preserves the frozen 64-bit Decode acceptance contract, coefficient values, ordered Score arithmetic and owned storage. This does not approve model quality or measured speed/process-memory improvement. The reviewer did not read trained model contents or execute Features, Fit, benchmarks, new tests or original runtime APIs.

The reviewer did not author packed78. The reviewer previously implemented role62 and participated in input/observer preparation and other experiment reviews: this is a nonblind review, not proof of independent source origin or a blind quality evaluation. The packed78 author is semantic_review60_prep.

## Correctness and ownership

`New` checks RIIDOH01 magic, dimension8192, reserved bytes, kind, scale, nonzero count and exact length. Like Decode, FP32 ignores the header scale in coefficients but still requires it to be finite and nonnegative. Both +0 and -0 scale are accepted. INT8 rejects q=-128 before multiplication even with zero scale; count measures nonzero q*scale, rather than stored nonzero q. Nonzero q with scale±0 and count0 therefore remains accepted. Ternary matches the presence population to count, rejects scale±0 with any presence, and rejects unused high bits in the last sign byte. Acceptance equivalence is scoped to the frozen 64-bit Decode target, not identical error strings.

Input bytes are copied into owned storage after validation. No accessor exposes an internal slice, and read methods do not mutate storage. The caller must keep input stable during construction. Caller mutation after return and a View value copy do not mutate the immutable internal storage. The saved ternary mutation/concurrent-read race test passed in the first run; the reviewer did not repeat it.

The ternary prefix has129 uint16 entries; cumulative counts never exceed8192, so they cannot overflow. For valid indices0..8191, the word is0..127 and sign rank0..n-1. The bit0 empty lower mask and bit63 boundary are correct in source. Saved synthetic coefficient-bit checks cover empty ternary and dense8192 through last rank8191.

Score keeps feature order and duplicates, reading each FP64 coefficient before `sum += coefficient * value`. INT8 still evaluates q*scale first; an absent ternary coefficient returns the +0 left by reference Decode. There is no sign-sum-then-scale rearrangement. The saved regression distinguishes that rearrangement using fractional scale and large cancelling values. Invalid feature indices return fixed errors rather than the reference panic, an explicit difference outside valid-index arithmetic parity. NaN/Inf feature arithmetic is not normalized away. Synthetic parity on Go1.27.1/darwin arm64 does not prove bit equality across every compiler or architecture.

## Saved verification evidence

One stdlib metadata-only helper verified the handoff, eleven preparation assets and two frozen reference sources:14 exact byte/SHA pins. It also checked pass/fail counts in two saved JSONL logs. The first race run has top4 pass/1 fail and sub37 pass/1 fail; its dimension fixture wrote0 to an already-zero byte. The author's separate first vet failure with13 unkeyed-literal diagnostics is preserved. After test-only correction, the affected race run has top3/sub38 pass and zero failures. Production view.go has the same SHA before and after that correction. The reviewer did not run new tests or call Decode/Score. The author's78 synthetic Decode and74 Score calls include repetition and are not counts of new independent cases.

Two small future regression additions could be useful. They are not identified source errors or new execution gates:

- Reject q=-128 with scale+0/-0 and count0, and reject nonempty ternary with scale-0. Individual guards and zero-scale accepted cases already exist.
- Compare small ternary nonzero counts7/8/9 directly across a sign-byte boundary. The current thirteen-edge and dense8192 cases cross boundaries too; the smaller cases would localize failures.

## What is not yet measured

65,536 B is the value payload of8192 reference float64 coefficients. Packed FP32 32,792 B, INT8 8,216 B, ternary13 1,308 B and dense8192 2,330 B are calculated owned data lengths. They include the24 B wire header and ternary258 B prefix, but exclude Go struct/slice headers, allocator rounding, caller input, feature buffers, total heap and OS RSS. The handoff's abbreviated “excludes headers” should be interpreted using the README/source explanation as excluding Go headers; the wire header remains in the numbers.

The prepared benchmarks directly compare Decode/New and Score/View.Score on the same public synthetic input and266 ordered features, with Score setup outside timing. This is suitable for a kernel comparison. Constructors create different owned representations, and the packed scorer includes view/index checks and coefficient conversion work. Equal operation counts are not claimed. The control→packed order is fixed in one process, and decoded arrays coexist with views: it is not a cold-start, serving or process-RSS comparison. Future count5 results must not be counted as independent model cases. The proposed60-second outer timeout does not guarantee benchmark completion within60 seconds.

Benchmark executions remain0. There is no ns/op, alloc/op, actual heap or RSS improvement result yet. Storage-length arithmetic and measured speed/memory results must remain separate. This review also establishes no new format, compression quality, GPU/SIMD implementation, training success, serving readiness or LLM usage/cost saving.

Evidence: [MECHANICS.v1.json](MECHANICS.v1.json), [RECEIPT.v1.json](RECEIPT.v1.json), [ATTEMPT-LEDGER.v1.json](ATTEMPT-LEDGER.v1.json).
