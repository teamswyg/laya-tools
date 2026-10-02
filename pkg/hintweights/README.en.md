# Optional packed-weight reader

`hintweights` is a small Go library that developers explicitly opt into. It reads 8,192 FP32, INT8, or ternary coefficients directly from owned RIIDOH01 bytes without expanding the full table to `[]float64`. Production code uses only the Go standard library. Existing routers, CLI behavior, default `Decode`/`Score`, and model policies stay unchanged.

```go
import "github.com/teamswyg/laya-tools/pkg/hintweights"

view, err := hintweights.New(rawModelBytes)
if err != nil {
    return err
}
score, err := view.Score([]hintweights.Feature{
    {Index: 0, Value: 1},
    {Index: 64, Value: 0.5},
})
```

The caller supplies `rawModelBytes` in RIIDOH01 format. This package does not read files, download assets, or select a model. Keep the input stable while `New` reads it. After return, changing the original input cannot affect the View's owned copy. Value copies of a View share immutable storage and support concurrent readers. No accessor exposes internal bytes.

| API | Meaning |
|---|---|
| `New(raw)` | Validate format, finite coefficients, count, and padding before an owned copy |
| `Coefficient(index)` | Read one coefficient as FP64 |
| `Score(features)` | Preserve input order and duplicates during FP64 multiplication and accumulation |
| `OwnedPayloadBytes()` | Count owned byte payload plus ternary prefix-table data |

A nil/zero View returns `ErrView`, an out-of-range index returns `ErrIndex`, and invalid model bytes return `ErrFormat`. Score does not skip zero-coefficient multiplication, preserving Inf/NaN behavior. Format-specific loops do not move scale after accumulation or reorder features.

Owned payload is 32,792 bytes for FP32 and 8,216 bytes for INT8. Ternary uses `24 + 1,024 + ceil(nonzero/8) + 258` bytes: 1,308 bytes for 13 nonzero coefficients and 2,330 bytes for 8,192. These counts exclude struct/slice headers, allocator padding, caller input, feature arrays, Go heap, and OS RSS. Original input and the owned copy may coexist. A smaller table does not guarantee lower CPU cost or whole-process memory.

This library is not Laya's text encoder, a trainer, or a new model. Ternary storage does not imply native 1.58-bit arithmetic or SIMD. It proves neither model accuracy, training eligibility, production qualification, nor lower Codex charges. Measure actual speed separately for the intended input and caller. Public synthetic checks cover format acceptance, coefficient/score bit equality, ownership, concurrent readers, signed zero, Inf/NaN, and sign-byte boundaries. Tests contain no trained weights or HF assets.
