# Wrap/Marshal 75 saved-result checks

The standard-library Go checker ran once and passed. It did not rerun the observer, controller, original functions, or tests. All 24 saved records matched the sealed Wants at the same positions with exact typed values: 12 Wrap and 12 Marshal observations, all errors null, no panics, and no differences.

All 12 Marshal before/after input states matched. The checks distinguished a nil map from an empty non-nil map. The 64-digit integer remained exact text rather than being converted to a numeric type. The final and last partial results were byte-identical. The frozen plan differed from its draft only in the one `Frozen` token changing from false to true. Hash bindings for the 14 source, license, and observer files, the binary, and the outside execution records also matched.

The outside ledger agreed with the saved OS log. There was one child start, no retries, and completed termination and Wait. Whole-child maximum RSS was 18,677,760 bytes; peak footprint was 16,089,616 bytes. OS real/user/system times were 0.87/0.01/0.04 seconds, and controller elapsed time was 0.879832625 seconds. These observations include initialization, preflight, wrapper and function work, and checkpoint writes. They do not establish pure function latency or a performance improvement.

The checker author also authored the controller and had prior exposure to the artifacts. This is saved-result binding and arithmetic QA, not independent source-semantic judgment or independent Want authorship. The 24 finite observations cover two behavior goals; they are not 24 independent parent requests or new labels. Training readiness, caption qualification, and parent-label qualification remain false. Model, fitting, role, Features, and Project calls and shared-repository or remote mutations remain zero.

`RECEIPT.v1.json` contains the pinned inputs and findings. `ATTEMPT-LEDGER.v1.json` records the single attempt and checker pins. Raw OS logs and private paths were not copied into these notes.
