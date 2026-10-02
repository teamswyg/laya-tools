# Observer 80: schema preparation before Want freeze

This is private schema design. Adapter creation, original package import/compile, native initialization/API calls, worker execution, and model/training execution remain 0. Preserve source-first80's 3 goals, 24 input fixtures, and 2 source families. Implement the adapter only after receiving the separate author's frozen Want SHA.

- UUID Parse, 8 fixtures: copy the returned fixed [16]byte directly to hex. Do not call UUID.String, MustParse, or Validate as observation helpers. Preserve partial returned bytes on errors.
- UUID Scan, 8 fixtures: copy a fresh explicit nonnil [16]byte receiver before and after the call. Distinguish nil interface, typed nil bytes, nonnil empty bytes, and dynamic input type. Direct 16-byte copying and internal recursion differ from the observer's direct call count.
- Original Ordinal, 8 fixtures: record int input and returned string. This is an explicit slice of unchanged ordinals.go, original go.mod, and the complete MIT LICENSE, not the complete humanize package.

The expected direct entrypoint budget is 24. Five returned errors are expected to produce 5 explicit Error-method observations: 1 original UUID URN method and 4 standard-library errorString methods. Internal Parse/Scan recursion and fmt formatting are not dynamically instrumented. Freeze bounded handling of unexpected errors that preserves mismatches instead of changing Wants or deleting records.

Use fmt %T for types without calling Error. Distinguish original API panic from an observed Error-method panic. Freeze a serializer that does not call arbitrary panic objects' String/Error methods. Inapplicable return channels are null; nil errors have null type/message.

Preserve 20 pins: the complete retained UUID Darwin non-JS runtime package's 15 files plus original go.mod/full BSD-3 LICENSE, and Ordinal's 1 source file plus original go.mod/full MIT LICENSE. Do not format original source. The 4 UUID namespace MustParse→Parse sites are static reservations; actual init callback/return counts remain null and uninstrumented. The outside controller must reserve before starting the child.

The proposed worker creates a fresh output directory and fsynced zero reservation/result file before callbacks, then records durable reserve/return partial snapshots. Preserve every Want, Got, and Match; do not turn differences into successes. The proposed outside controller runs one child with CPU 1, a 256 MiB Go soft heap, a 60-second process-group Kill/joined Wait guard, and 1 MiB stdout. These settings are neither measured RSS nor a hard RSS cap.

Run pure schema/stub tests only in a separate package. Never execute the main package containing original imports or `go test ./...`. Native binary compilation is allowed after Want freeze, but no binary/test invocation may execute initialization or APIs. Three logical goals and 24 finite fixtures are not 24 independent parents or new labels.
