# Reduce stored bytes while preserving numbers

Large JSON may reach the input cap before the next data expansion. This experiment tests **storage separately from training**. One existing76-request snapshot was encoded, read and compared once; [FP64 bits, uint16 indices, CSR and null states matched](ROUNDTRIP-HANDOFF.v1.json). All four CSR datasets, diagnostic duplication, metadata and false qualification flags remain intact.

| Format | File bytes | gzip bytes |
|---|---:|---:|
| Original pretty JSON | 8,390,462 | 594,236 |
| Minified JSON | 4,697,570 | 400,256 |
| FP64 compact | 1,978,694 | 328,206 |

The [separate size control](SAVED-BYTES-CONTROL.v1.json) shows44.01% reduction from whitespace removal and57.88% additional reduction from minified JSON to compact. The76.42% reduction from the original combines both effects. Each gzip variant was produced once with the same standard setting and compares cache/transfer size. Decompression restores the large JSON, so gzip does not automatically resolve the64 MiB decoded-input cap.

**Smaller files do not establish lower RAM.** [Whole-worker OS evidence](ROOT-OS-MEASUREMENT.v1.json) records peak RSS86,343,680B for JSON import, encode, write, read, decode and reporting together; it is not a standalone decoder measurement. Original projected payload1,928,154B and new arrays-plus-metadata1,978,406B also use different accounting scopes. Sequential warm stage timings are not serving-speed improvement ratios.

The [independent source review](independent/FINDINGS.en.md) found no blocker overturning the fixed snapshot result and no new evidence of array mismatch. It is static review without reruns. Binary decoding checks counts, physical lengths and the64 MiB limit for all four datasets before allocation. JSON import checks payload after typed allocation and may collapse duplicate keys into maps, so it should not directly become a general-input runtime. A same-codec roundtrip is not an independent source-array reader comparison.

No shared loader or runtime changed. Pure codec race/vet/build and the saved roundtrip passed, while larger data and standalone reader memory/allocation/numeric evidence remain follow-up work. Those findings should determine whether minification/gzip suffice or a separate codec justifies its maintenance cost. Existing files, models and results were not overwritten. Original features, projection, training, APIs, roles, labels and final2,400 data reads remain zero. Binaries, compact contents, raw logs and host paths are excluded from publication.

[Experiment scope](PLAN.en.md) · [Roundtrip](ROUNDTRIP-HANDOFF.v1.en.md) · [Separate control](SAVED-BYTES-CONTROL.v1.en.md) · [한국어](README.ko.md)
