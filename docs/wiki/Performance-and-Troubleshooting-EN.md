# Performance and troubleshooting

[한국어](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-KO) · [Home](https://github.com/teamswyg/laya-tools/wiki)

## Decide whether to enable Laya

Initial M4 Pro/24 GiB measurements, not promises for every machine:

| Path | Observation |
|---|---|
| Repository keywords only, 6-repository fixture | About 11.5 MiB peak process RSS |
| Repository preview with Laya | About 1.40 GiB RSS; model startup about 942 ms |
| Mixed 24-request preview set with Laya enabled | Median 39.669 ms, p95 54.040 ms, excluding load; includes skipped inference |
| Pure-Go shortlist, 1,000 synthetic repositories | About 0.17 ms/request; excludes index construction and inference |
| Code reranking, 8 candidates in development retrieval tests | Roughly 1.6–1.8 seconds added |

On easy synthetic repository fixtures, lexical Top-1 was 18/18 positives; Laya's raw choice was correct for 8/15 judgments and accepted suggestions were zero. Zero coverage does not establish high precision. No actual Codex savings have been demonstrated. Start lexical, then compare on labeled tasks representative of your work.

## Common problems

| Symptom | What to check |
|---|---|
| `riidolaya: command not found` | Run `./riidolaya` or `./bin/riidolaya`, or add your installation directory to PATH. |
| Example catalog/config not found | Run from a laya-tools clone; examples are not in binary-only archives. |
| Missing model/runtime | Run `riidolaya doctor`, then `riidolaya setup`; setup verifies downloads. |
| Download/checksum failure | Check network and proxy configuration; retry setup. Do not bypass checksum or TLS verification. |
| Search has no results | Use exact identifiers; check root/ignore rules. Reranking cannot recover absent candidates. |
| Searching a subdirectory fails | Install `rg`; Git-root and ordinary-directory file discovery differ. |
| Model route always uses strong | Inspect `abstained`/`reason`; default confidence is 0.9, Korean classification is unvalidated. Do not assume reducing the threshold improves outcomes. |
| Repo preview returns `candidate` with `--laya` | Check reason: unsupported script/metadata or inference unavailable can leave lexical candidates. |
| Repo preview returns `abstain` | Could be ambiguous/multi-repo, low confidence, none, or truncation. Shorten metadata and inspect the task. |
| High memory with warm services | Each process may hold a model. Reuse a process when appropriate; stop it when done. |
| Core ML initialization fails | The tested dynamic export failed; use default CPU. GPU/ANE benefits are unverified. |
| Agent cannot parse output | Use `--json`/JSONL, keep stderr separate, and inspect status rather than assuming success means recommendation. |

## Measure on your machine

```sh
riidolaya bench --iterations 30 --threads 4
riidolaya bench --cpu-profile cpu.pprof --heap-profile heap.pprof --ort-profile ort-trace
go tool pprof -top cpu.pprof
go tool pprof -top heap.pprof
/usr/bin/time -l riidolaya bench --iterations 30
```

The last command is macOS-specific. Go pprof excludes native/GPU memory; do not report Go heap as total RAM. Native calls may lack useful symbols. ORT profiling itself adds overhead. Keep profiles and private catalogs out of public issues; describe version, platform, command, and a redacted reproducible example instead.

[Full measurements](https://github.com/teamswyg/laya-tools/blob/main/docs/measurements.md) · [Repository evaluation](https://github.com/teamswyg/laya-tools/blob/main/docs/repository-routing-preview.en.md) · [Issues](https://github.com/teamswyg/laya-tools/issues)
