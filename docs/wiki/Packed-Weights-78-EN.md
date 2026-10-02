# Choosing a compact coefficient table in Go

`pkg/hintweights` does not generate text or train a model. It is an optional Go API that reads an existing RIIDOH01 table of 8,192 coefficients directly from packed bytes instead of expanding every coefficient into an FP64 array. It does not run the whole Laya text encoder within that footprint.

An FP64 coefficient array contains 65,536 data bytes. Owned payload in the new reader is 32,792 bytes for FP32, 8,216 for INT8, and 1,308/2,330 for ternary with 13/8,192 nonzero coefficients. These figures exclude structs, allocator rounding, caller input, feature arrays, Go heap and process RSS. Input and the owned copy can coexist during construction.

There is a speed tradeoff. With 266 public synthetic features, scoring in the first private prototype was about 4–5 times slower than the existing array. After separating loops by storage format, median score time was 213.1ns for FP32, 189.6ns for INT8 and 248.9/418.8ns for sparse/dense ternary, about 1.21–2.68 times slower than the array. All measured scoring samples had 0 bytes/op and 0 allocs/op. These are warm synthetic repetitions on one Mac; versions ran in separate processes. The public package port was not separately benchmarked.

Developers can try it when coefficient storage matters and retain the existing path when scoring speed matters. Default routers do not select it automatically. `New` validates input before making an owned copy; keep input stable during construction. Returned Views share read-only storage and support concurrent reads without a separate lock. Feature order, duplicates and FP64 multiplication/accumulation are preserved. Zero-coefficient products are retained to preserve Inf/NaN behavior.

[Usage and API](https://github.com/teamswyg/laya-tools/blob/main/pkg/hintweights/README.en.md) · [Original measurements and failed attempts](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/packed-weights-78). Bitmap/sign storage is not native 1.58-bit arithmetic, bulk SIMD or a full SoA implementation. Real application RAM, model accuracy and LLM usage savings require separate validation.

Development sources are also expanding. [Observation77](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-77) observed 24 finite logfmt/percent-escape inputs once, matching the expectations. These are not 24 independent tasks or new training labels. [Proposal79](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/source-growth-79) considers six families including UUID and humanize while separating source acquisition, licensing, related groups, observations and training eligibility. Protected 2,400-case final evaluation and utility of a new learned model remain incomplete.

[한국어](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-KO)
