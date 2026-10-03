# Recording small observations accurately

[![CI](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml) · [한국어](README.ko.md) · [Progress](https://github.com/teamswyg/laya-tools/issues/19)

Owned Go preparation for riidolaya's low-resource experiments. Repeated tiny-model use also requires inexpensive copying and serialization. Native-shaped return/panic, owned copy failures, thunk outcomes and acknowledgements remain distinct so incorrect observations do not become training labels.

With Go1.27.1, run `bash scripts/verify-next60-compact-evidence.sh` from the repository root. Humans and agents use the same checksum and owned-control path:34 top-level tests plus five worker subcases. It needs no upstream packages, Laya/GPU/model download or training. Inert text sources are not wired into the production riidolaya command.

| Owned measurement | Padded representation | Compact prototype |
|---|---:|---:|
| Identical one-row snapshot/encoding, median of3 | 12.958µs | 1.133µs |
| Allocated bytes/op | 5,888 | 586 |
| Allocations/op | 3 | 7 |

Compact includes snapshot creation. Allocation count increases while bytes and time decrease. This measures neither file Sync/hashing nor native/model inference, whole-work cost, RSS or GPU memory. Raw pprof stays private. Cumulative allocation space is not peak process memory.

The owned padded39-row result was179,946 bytes. The prototype encodes active reads/visits/views and preserves availability, knownness, null and value independently. Available-unknown is1; known-false is3. Error bits retain their own meanings. This compresses observation state, not model weights; it is not1.58-bit quantization or a new Fit.

Each row is bounded to1,536 bytes including LF. Conservative numeric/array supersets measure882-byte Union and889-byte JSON rows;39 rows plus4,096 reserved header/footer bytes total38,631. Multiplying the global maxima gives the more conservative38,641, within the65,536-byte result reserve. Actual file code must enforce those bounds without dropping unknown data.

The next binding freezes the snapshot before candidate AFTER, stores it, and ties offset/length/SHA to the journal. File Sync precedes FINAL's whole-file SHA. A hash proves byte correspondence, not native provenance or semantic correctness. Strict external decoding, mask semantics, wire3, file/process lifecycle and original execution remain pending.

Initial failures and source snapshots remain in history. Owned worker source imports no original package and uses frame wire2. Compact ExecutionGate remains ErrPending. No weights, upstream source bodies, binaries, raw profiles or private inputs are included. New original/model executions, Fits and Golden rows:0. Existing35 development cases and the proposed20→60/protected2,400/domain evaluations remain separate.
