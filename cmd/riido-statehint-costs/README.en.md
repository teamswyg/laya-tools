# Local inference costs for an already-held Go model

This maintainer command repeats four original fictional texts against a SHA-pinned
version1 `.rsh` artifact. It performs no training, accuracy evaluation, download or
operational mutation. Keep these performance fixtures outside training, selection
and golden datasets. Parent and child artifacts can use the same command.

```sh
go build -trimpath -o .cache/statehint-costs ./cmd/riido-statehint-costs
.cache/statehint-costs --model MODEL.rsh --model-sha256 SHA --iterations 500000
.cache/statehint-costs --model MODEL.rsh --model-sha256 SHA --iterations 500000 --profiles-dir NEW_PRIVATE_DIRECTORY
```

The hash, artifact format and nonzero training steps are checked first. One caller
workspace is reused. Model validation and repeated inference time are reported
separately; iterations are bounded to 4–1,000,000. Profiles use a fresh directory
and stay local. JSON reports mean time and Go allocations during the loop, without
texts, predictions, labels or host paths.

Profiling adds its own costs and allocations. A faster profiled observation does
not demonstrate optimization. Whole-process startup, IO, JSON output and shutdown
belong to separate OS measurements. Four short fixtures and one development run
do not establish p95 latency, production throughput, accuracy or whole-app RSS.
Go pprof measures Go CPU/heap, not GPU costs. Never commit raw profiles or weights.

[Development measurements of the existing v0.2 artifact](COSTS.development.json)
preserve one unprofiled 20,000-call run and separate profiled 20,000/500,000-call
runs. Their timing differences do not establish optimization or profiling speedup.

[한국어](README.ko.md)
