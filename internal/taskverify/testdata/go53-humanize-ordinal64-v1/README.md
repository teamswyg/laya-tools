# Minimal public ordinal fixture

These five files are copied without source edits from
[dustin/go-humanize at a1b4e66](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e):

| File | Git blob SHA-1 | Copied scope |
| --- | --- | --- |
| `go.mod` | `611d24fa1a8560ca99ba83263ec3d6a12d85c3e9` | Complete original module declaration; no external requirements. |
| `ordinals.go` | `43d88a861950eac85b0f742a59621f92345d7109` | Complete original `Ordinal(int)` source; imports only standard-library `strconv`. |
| `ordinals_test.go` | `c478d5c9bdafae580e0bc7f6283766ae6c48d4a6` | Complete original ordinal test. |
| `common_test.go` | `fc7db151640401c8b56bce644e294076bdff9fe0` | Complete original `testList` helper required by that test; imports only `testing`. |
| `LICENSE` | `8d9a94a90680d9fc114a1b3a2b4123c233c324af` | Complete original MIT copyright and permission text. |

This is a closed ordinal source subset, not the whole upstream package or a
claim that all upstream functions are assessed. No `number.go` or code from its
separate gorhill/WTFPL origin is copied, compiled or needed by this closure.
The copied Go files contain no separate license/origin notice. The complete
non-truncated pinned upstream tree contains `LICENSE` and no `NOTICE`.

Keep this fixture's MIT license with any redistribution of the copied files.
The repository's Apache-2.0 license does not replace that upstream license.
This README is an authored scope record, not an upstream file. This review does
not provide blanket clearance for other files, dependencies or model training.

The original API has no `Ordinal64`. The request is authored project behavior:
new signed int64 ordinals use the magnitude for suffix selection, while the
existing int API keeps its original behavior, including negative `th` suffixes.
Independent acceptance assertions and authored correct/wrong temporary changes
are verifier controls for one request, not model outcomes or additional tasks.

Execution uses a trusted offline Go1.27.1 toolchain and a minimal compiler module
with the original module identity. The original Go1.21 `go.mod` bytes remain
pinned staging evidence; the temporary compiler module is a harness adaptation.
