# First three sources: saved-result comparison

One stdlib Go checker invocation completed: one success and zero failures. No original API, observer, or controller was rerun. The checker author also authored the controller, so this is **saved-file verification**, not independent runtime execution or broad semantic approval.

All 24 Probe IDs, ordering, inputs, initial receivers, and Wants exactly matched the immutable pre-execution Want. Complete Got observations and Matches flags agreed: 10 datasize, 6 query, and 8 shlex probes; 24 returns, zero panics, and zero differences. Three `uint64` maximum values were compared without float conversion. The eight observed errors exactly matched their expected Kind, Text, Function, Input, and Cause; errors were preserved rather than discarded.

The checker distinguished one nil query map, two allocated-empty maps, three nonempty maps, repeated value ordering, and one allocated-empty shlex slice. Final and partial results were byte-identical. The frozen plan changed only one Frozen token from the draft. Result, plan, Want, controller ledger, and private-log SHA bindings matched.

Saved OS fields also matched the log: child real 0.85 seconds, user 0.01 seconds, sys 0.04 seconds, maximum RSS 18,743,296 bytes, footprint 16,122,360 bytes, and controller elapsed 0.854884 seconds. These are retained whole-child observations, not new measurements. They include initialization, preflight, wrappers, and checkpoint I/O; they do not establish pure API or Laya inference performance or a hard RSS cap.

The 24 probes are finite inputs for three behavior goals, not 24 new distinct parent requests. New parents, labels, roles, models, fits, shared-file changes, and publication remain zero. Caption, parent-label, broad-truth, and training qualification remain false. `RECEIPT.v1.json`, `INVOCATION.v1.json`, and `LEDGER.v1.json` preserve exact input bindings and the single checker attempt.
