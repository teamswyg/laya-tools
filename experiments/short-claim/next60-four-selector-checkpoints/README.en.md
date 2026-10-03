# Observation records for the next four requests

The next batch has four distinct requests, 19 inputs and 12 candidates, with 57 proposed candidate calls. **These are not completed original observations or newly approved labels.**

The observer saves return values and error presence before reading the error string. A write failure stops later calls. A sticky writer failure also rejects a previously completed report. A surviving JSON row after interruption alone does not prove that synchronization returned.

The [synthetic check result](root-tested/RESULT.v1.json) has 16 top-level tests and five subcases: 21 passing records. Race and vet checks passed. The test module excludes the original-importing `nativebinding` and native entrypoint. One test runs 57 fake callbacks; they do not count as original observations.

CI uses `scripts/verify-next60-development.sh checkpoint` to copy the [two owned packages](root-tested/source) into an isolated module. Humans and agents can use the same reproduction command. Real original execution remains a separate phase after source, compiler, rights, binary, time and output checks.

[Outside-controller preparation](../next60-four-selector-outside-preparation/README.en.md) adds eight separate fake lifecycle and result-metadata tests. These tests are separate from the planned 57 real candidate dispatches.

[Before/after preparation](preparation-v2/README.en.md) and [original selection review](../next60-four-selector-selection/README.en.md) remain available. Later [57 actual observations and independent review](../next60-four-selector-actual-observation/README.en.md) brought the new round to seven qualified requests. This page's synthetic tests are separate from those actual observations. Training follows the checkpoints of 30 and 60.

[한국어](README.ko.md)
