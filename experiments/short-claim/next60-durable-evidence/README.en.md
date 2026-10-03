# Compact result persistence and verification pipeline

[한국어](README.ko.md) · [CI](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml)

A system that calls tiny hint models repeatedly also needs inexpensive surrounding code. This experiment verifies **bounded result persistence and rejection of incomplete execution evidence** before adding training data.

It does not train a model or execute Laya. All API methods are owned fakes. File creation, writes, Sync, readback and journal/ACK integration use real temporary files. The existing CI Laya inference step is separate and does not prove GPU execution.

## Changes

- Persist one compact JSON row at a time instead of retaining 39 padded structs. Keep the 64KiB file and 1,536-byte row limits.
- Preserve unknown, false, nil, empty values and error identities. Unavailable observations do not become training labels.
- Use fixed arrays for method references and call stacks. One receiver loop owns state, requiring no map or shared-state lock. Reservation order survives nested completion in a different order.
- Separate stored bytes, journal Sync, external ACK and completion. Lost ACK leaves an orphan record without promoting it to success.
- Check actual row bytes, SHA, canonical shape, task IDs and call ancestry before journaling/ACK. Check the complete header/footer/row/file hash before FINAL.

**93 top-level controls and 152 subcases** passed default/compatibility race tests, vet and formatting. The integration control handles 39 rows, 51 fake method calls and 260 frames, and joins both loops after EOF. It is not proof of the original libraries' 96 methods/91 callbacks, or new Golden observations.

## Usage

With Go 1.27.1 installed, run from the repository root. No Python or model download is required.

```sh
bash scripts/verify-next60-durable-evidence.sh
RIIDO_EVIDENCE_BENCH=1 bash scripts/verify-next60-durable-evidence.sh
```

The first command verifies pinned owned sources in a temporary directory. The second also measures readback cost. Temporary test artifacts are cleaned after execution; original failed experiment records remain preserved separately. Automation must check exit code 0 before continuing. Profile collection is disabled by default.

## Measurement scope

Apple M4 Pro, Go1.27.1, GOMAXPROCS=2, GOMEMLIMIT=64MiB; three 200ms samples. GOMEMLIMIT is an advisory Go GC target, not a hard RAM limit. The baseline is an exact copy of the earlier implementation body, evaluated on identical inputs.

| Check | Baseline median | Bounded median | Baseline B/op | Bounded B/op |
|---|---:|---:|---:|---:|
| Malformed 64KiB all-LF file, 39 references | 677,815ns | 29,907ns | 1,711,176 | 139,336 |
| Valid 39-row file | 9,008ns | 8,980ns | 12,782 | 12,783 |

The malformed case supplies the correct whole-file hash, so early hash rejection cannot explain the difference. Bounding temporary split segments reduces allocated bytes by approximately 91.9% for this input. Allocation count remains 21. No meaningful valid-39-row speed improvement is established. The valid zero-row case adds 32B/op for a spare segment slot.

These are Go readback microbenchmarks, not RSS, GPU memory, Laya inference, token savings, whole-work cost or native 1.58-bit speed. No SIMD change is included.

## Next steps and public scope

Original execution still requires license, selected dependency source, compiler artifact, process lifecycle and RAM verification. A forced 6MiB binary limit remains a proposal. Independent work/Golden groups and whole-work cost are needed before further tuning and Hugging Face model publication. New models, Fits and Golden additions remain zero.

Five cold historical census/delivery records were compressed after actual physical restoration checks, retaining the 512MiB experiment-record budget. This reclaimed storage is not model RAM improvement. Failed test causes and corrections remain under `history/`.

No original library bodies, private user work, credentials, weights, binaries, raw journals or profiles are included. `SOURCE-BINDINGS.public.v1.json`, `evidence/`, `FILE-MANIFEST.v1.json` and `SHA256SUMS` bind the owned source and evidence. Portable relative module bindings are explicitly distinguished from original local QA pins.
