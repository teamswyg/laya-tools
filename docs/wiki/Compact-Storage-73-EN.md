# Reduce storage while preserving training numbers

A separate storage/readback experiment on the existing76-request snapshot preserved FP64 bits, indices, CSR and null states. File sizes are below.

| Format | Original bytes | gzip bytes |
|---|---:|---:|
| Pretty JSON | 8,390,462 | 594,236 |
| Minified JSON | 4,697,570 | 400,256 |
| FP64 compact | 1,978,694 | 328,206 |

Whitespace removal accounts for44.01% reduction; compact adds57.88% versus minified JSON. The76.42% reduction from the original combines both. Gzip compares cache/transfer bytes, without changing the decoded-JSON input limit.

Files and RAM differ. The complete JSON-import/compact-roundtrip/report worker peaked at about82.3 MiB RSS. This does not establish standalone decoder memory or inference speed gains. Shared loaders, runtime, fits, labels and limits remain unchanged. Next, verify numbers and memory separately before choosing simple compression or a dedicated codec.

Follow-up74 read each format once in a fresh child. All24 column bit/count/null digests and metadata matched. Peak RSS was JSON18.23MiB versus compact18.41MiB: **no compact RAM saving was demonstrated**. Cumulative allocations were44.84MB versus21.61MB, while final Go heap was slightly higher for compact. Fixed JSON-then-compact order, prefetched files and one sample do not establish a general speedup. Do not compare these readers directly with the historical82.3MiB whole roundtrip. Compact remains a file-cache preview; the default loader is unchanged. [Actual standalone measurements and review](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/standalone-reader-74).

[Bilingual details, frozen plan, original result and controls](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/compact-storage-73) · [한국어](Compact-Storage-73-KO) · [New-source observations](Source-Observation-73-EN)
