// Package hintweights optionally reads owned packed RIIDOH01/8192 coefficient
// tables without expanding them to a full FP64 slice. It supports FP32, INT8,
// and ternary bitmap/sign storage using standard Go only. Score preserves the
// supplied feature order and FP64 multiply/add order; packing is not native
// ternary arithmetic or SIMD. This library neither encodes text nor trains or
// activates a model. Existing routers and default model paths do not select it.
package hintweights
