# Source observation80: public source archive and input proposal

This archive preserves source for a later bounded comparison of UUID parsing, receiver changes, and small nonnegative ordinals. The [frozen explanation](preparation/SOURCE-FIRST-80.en.md) and [input drafts](preparation/INPUT-DRAFTS.v1.json) are copied unchanged. This preparation neither authors nor executes Wants, adapters, source APIs, or training. It does not report the later state of a separate observation task.

| Source | Preserved scope | Complete license |
|---|---|---|
| [google/uuid](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d) | The fixed revision's 15 Darwin non-JS runtime sources and original go.mod | [BSD-3-Clause](upstream/google-uuid/LICENSE) |
| [dustin/go-humanize](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e) | Original ordinals.go and go.mod only; an explicit source slice, not the complete upstream package | [MIT](upstream/humanize-ordinal-slice/LICENSE) |

Go sources use inert `.go.txt` names, and module files use `.mod.txt` names. Their bodies, copyright headers, comments, build tags, and module contents are unchanged. This is neither a runnable library port nor a compiler-closure proof. JavaScript UUID source, upstream tests, added Ordinal64, observer helpers, and binaries are excluded. Other humanize files and the separate WTFPL origin associated with `number.go` are outside this Ordinal slice and were not newly acquired.

All 15 UUID files retain their Google copyright and BSD notices. Both complete upstream LICENSE files accompany the sources. [NOTICE](NOTICE.md) and the [manifest](ARCHIVE-MANIFEST.v1.json) identify origins, file notices, and SHA-256 pins. The hosting repository's Apache-2.0 license does not relicense these upstream sources. Read the relevant complete license for redistribution conditions.

The proposal has **3 behavior goals, 2 source families, and 24 input slots**: UUID Parse, UUID Scan, and Ordinal. The two UUID goals and repeated fixtures do not become independent source families. Existing null truth, role, and weight fields and false qualification, training, production, and protected-final states are unchanged.

UUID `hash.go` contains four statically read namespace-initializer MustParse→Parse call sites. Original init was not instrumented; this is not an observed count of four init callbacks or returns. A future dispatcher must reserve that startup boundary before starting the child, and separately measure CPU/RSS over the entire child lifetime, including initialization. This archive preparation performs zero original imports/init, APIs, tests, compilation, native observations, model calls, or fits.

[SHA256SUMS](SHA256SUMS) checks copied source and preparation bytes. The [copy ledger](COPY-LEDGER.v1.json) describes this metadata preparation only. The prior helper compilation failure remains in the [historical ledger](preparation/LEDGER.v1.json). A separate author must freeze Wants before later observation and candidate-satisfaction review. This archive is not a label, training eligibility, generalization approval, or a replacement for 2,400 unseen final requests.
