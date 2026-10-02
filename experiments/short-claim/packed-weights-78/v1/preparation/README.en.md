This private Go kernel reads needed RIIDOH01 coefficients from packed bytes instead of expanding a full FP64 array. It is separate from model-quality improvement and ternary training. Models, wire format, features and training stay unchanged. No HF model or protected data was read.

Existing Decode expands FP32, INT8 and ternary formats into 8192 FP64 coefficients: 65,536 bytes of coefficient values. `New(raw)` preserves its validation contract, then copies the input into owned storage. Caller mutations after return cannot change the view; input must remain stable during construction. `Coefficient(index)` and `Score(features)` preserve coefficient values and ordered multiplication/accumulation.

FP32 reads four bytes and converts to FP64. INT8 multiplies q by header scale first. Ternary uses the presence bitmap and a 129-entry uint16 prefix-popcount table to find sign rank. Only ternary allocates the 258-byte table. No maps, shared locks or mutable global caches are used. Calling `math/bits` alone does not prove SIMD or a particular hardware intrinsic.

| Format | Owned byte payload | Prefix table | Total payload |
|---|---:|---:|---:|
| FP32 | 32,792 B | 0 | 32,792 B |
| INT8 | 8,216 B | 0 | 8,216 B |
| Ternary, 13-nonzero synthetic example | 1,050 B | 258 B | 1,308 B |
| Ternary, 8192 nonzeros | 2,072 B | 258 B | 2,330 B |

These are **owned data lengths**, excluding struct/slice headers, allocator padding, caller-retained input, feature buffers, Go heap and OS RSS. Model file size is unchanged. Caller input and the owned copy may coexist. Packed access may trade storage for per-feature conversion/popcount work; speed, allocations and RSS still need measurement.

Match source7cea Decode's acceptance, including permissive FP32 header scale. INT8 scale ±0 can produce signed-zero coefficients from nonzero q with count0, but q=-128 is always rejected. Ternary absent bits produce +0; present bits with zero scale are rejected. Preserve finite/count/reserved/length/sign-padding checks. Retain feature order and duplicates. **Do not sum signs first and multiply by scale afterward**: a literal synthetic test distinguishes the rounding bits. Invalid feature indices return a fixed error instead of panic, outside the valid-index math contract. Nonfinite feature arithmetic is not silently changed.

Nine public synthetic formats/edge cases compare all 8192 coefficient bits and 36 reference scores; a separate literal compares one score that distinguishes arithmetic regrouping. Coverage includes empty features, repeated/unsorted indices, large/small values, signed zero/nonfinite features, 64-bit prefix boundaries, dense/empty ternary storage, caller mutation and concurrent readers. Another 29 byte-acceptance cases compare New with existing Decode. Initial checks found a test that failed to corrupt the dimension and vet warnings for unkeyed literals. Preserve that history, fix only tests and rerun the three affected race tests. Across both test invocations, synthetic reference Decode/Score calls total 78/74; repetition does not add independent cases. Production view.go is unchanged across that correction.

Benchmarks are **prepared code only: zero executions**. After root review, compare existing Decode with owned New separately from prepared-feature scoring. Call both original Score and packed Score directly; do not charge a callback to the control. Fixture generation/Decode/preparation stay outside score timing. Use CPU1 and identical public features/order. A microbenchmark cannot establish file-I/O, whole-process RSS or serving latency.

This maintainer module has a local replace path and is not a portable public runtime port. Source commit: `7cea49090727555924bf22679ffeccc4f38c9865`; decoder SHA `f27e18883f073555775ff15de4f2e72d4fcff9bb1b49dfe7b11007779ec95aff`; feature/scorer SHA `d0c906d1283617d0778039ff8bc4c6fff6cb0af3a813c48df088424d78bebb71`. FP64-bit parity was tested on synthetic inputs with Go1.27.1/darwin arm64, not proved for every compiler/architecture. Actual HF Decode/Score, Features, Fit, native observer, protected-data reads, uploads and shared edits remain0.
