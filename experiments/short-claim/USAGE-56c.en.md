# Scoped proposal preparation56c usage

[한국어](USAGE-56c.ko.md) · [Property preparation plan](PLAN-56c.en.md) · [Resident resource plan](PLAN-RESIDENT-56d.en.md) · [Existing56b results](RESULTS-56b.en.md)

`riido-scopeprep` is a Go maintainer tool that prepares small property proposals. For example, ask whether quoted commas are preserved separately from whether the whole function is correct. The command neither executes candidates nor ranks them or assigns truth. The existing [riido-shortclaim](USAGE-56.en.md) suggests verification order with `unverified_heuristic` status. Neither command decides model selection or grants execution approval.

It requires public repository sources and installed or selected Go 1.27.1. Run from the repository root with a new output directory. It uses no Python, Laya inference, GPU or authentication key. During preparation, standard-library provenance metadata comes from installed Go or an already cached exact Go 1.27.1; downloads are disabled.

```sh
mkdir -p .cache
go version
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o .cache/riido-scopeprep ./cmd/riido-scopeprep
./.cache/riido-scopeprep --stage prepare --out .cache/my-scope-proposals-56c
```

First check that `go version` reports Go 1.27.1. The tool does not interpret arbitrary repository files to produce automatic truth. It does not support `--stage audit`.

| Generated file | Meaning |
|---|---|
| `probes.json` | 4 existing delimiter parents × 3 properties = 12 proposal rows. Preserve the original 3 candidate IDs, order and source pins per row, with 36 descriptions |
| `preparation.json` and stdout | Proposal-byte hash, 19 actual compiled sources, counts, limits and stop reasons. This does not complete independent semantic review of captions or truth |

Every proposal/candidate caption has `caption_review=pending`. Text bounds require at most 512 raw/normalized bytes and 32 normalized words; text is never truncated. Library validation refuses promoting pending to reviewed or forging source, original-parent, candidate-order or table bindings. `FeatureInputs` projects request/candidate text into ranking input and excludes property IDs, original parents, sources and truth tables from features. Projection itself grants no semantic review or correctness certificate.

The three properties cover quoted commas, escapes outside quotes and zero output on specified syntax errors. Define narrow tables from 9 existing independently authored literals; preparation unit checks observe actual returns from the 4 original functions. Do not copy whole-function wrong roles into property labels. Syntax checks require the expected syntax error, zero Count and the entire empty token array, so suppressing the error cannot pass vacuously.

Literal tables are expectations for subsequent code observations. They do not approve pending captions as faithful expressions of those meanings, and no official property labels or answerable classification exist yet.

**Separate preparation unit checks from this command.** The `prepare` command creates 0 candidate evaluations, truth labels, groups, roles, fits, model calls, ranking runs, performance runs or protected-final reads. Neither 12 proposal rows nor 9 literals add independent golden requests. Preserve the 4 unknown delimiter parents and all existing56b results; do not automatically promote new proposals to known.

Writing these documents or proposals does not freeze an official property-audit execution plan, caption review or role split. No fitting before faithful captions, diverse provenance and separate utility/role plans are prepared. Keep the separate goal of at least 2,400 distinct protected-final requests per domain.

Preparation refuses stale compiled/file sources or unlisted Go implementation files in the owned packages and verifies the existing input hash. It never overwrites existing output directories/files; fixed diagnostics contain no input or private paths. The 256 MiB Go-heap limit is soft, and GOMAXPROCS 1 controls Go execution parallelism. These are not total RSS, OS-thread or GPU hard caps, nor new memory/latency measurements. Offline `go/types` provenance costs are separate from the small resident hint tool.
