# Full regression checks and historical preservation

The six optimized Go files remain byte-identical to author commit `5aab32e8a0695815e2d10e8a875e13d459f30e61`. Subsequent changes address historical source-pin regression tests and exact original text archives. Runtime verifiers and historical plans, results and thresholds are unchanged.

Three parent full race-suite attempts failed. Some initial output was truncated, so the complete failure lists for the first two attempts are not asserted. Observed failures compared historical source hashes against changed current files. The behavior/property/typed checks were repaired first, followed by the resident measurement record, scope preparation and stored-utility checks. Failed attempts remain failed; original measurements were not rerun.

Historical source/support SHA and byte counts are verified against seven exact Git blobs from `9d204c2c700505658c108297d5fa769a835a60c3`. Current compiled sources and source closure are verified separately and strictly. Current numerical replay retains the original score tolerance, candidate orders, truth, costs, fallback, groups and flags. Public CLIs continue rejecting historical execution plans against changed sources. Archives are never runtime fallbacks or features.

Independent read-only review confirmed the six optimized files, seven historical records, three additional regression checks and two new archives without blockers. It ran no duplicate training or benchmark. Author targeted race/vet checks passed. The fourth parent full `go test -race ./... -json` passed with exit0 and zero JSON failure events. Full vet, owned-source gofmt and diff checks also passed. Final GitHub CI remains a separate merge gate.

The single observed ranking allocation change, 81 to 0, is preserved. Different profiling settings prevent a causal latency/RSS improvement claim, and the increased ending Go heap remains recorded. These regression repairs add no benchmark, official fit, labels, role assignment or paid call.

[Usage and measurements](README.en.md) · [Initial targeted regression record](FINDINGS.v1.en.md) · [한국어](INTEGRATION.ko.md)
