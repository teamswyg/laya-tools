# Native2 v5 independent review

The exact final v5 source has no remaining concrete blocker in this static review. This finding permits Root to continue preparation; it supplies no execution readiness, CI pass or semantic result.

Root's real Go1.27.1 build exposed a gap in v4: trimpath intentionally removes linker declarations from buildinfo. Fifteen fake tests and the earlier source review did not prove actual binary compatibility. The old v4 seal stays intact. v5 reads the actual three SHA strings from an unstripped thin little-endian ARM64 Mach-O file, compares their16-byte Go headers and64-byte backing data with the plan, and retains the version/module/architecture/CGO/trimpath/noVCS checks.

The peer review found that symbol-count/string-table limits alone could multiply long symbol-name copies. The correction bounds each name to1024 bytes and aggregate names to8MiB before the standard parser. Load commands, symbols, indirect symbols, relocations, sections and ranges are also bounded before allocation. The author then found that buildinfo itself parsed Mach-O before this gate. I withdrew my premature clearance; final v5 bounds both self and worker before buildinfo. The correction ledger preserves that attribution and review mistake.

Static metadata of Root's existing held binary matches all three values. The final author tests report25 passes; across three stages22/24/25 pass, authored fake starts12 times with12 waits and3 deliberate missing-start failures. The peer reran no tests or binaries. Peer metadata tools ran nm once, otool once and xxd five times. Original, controller, model, HTTP, compilation and publication counts are0 for this review.

The reservation → readiness → plan → exact binding → durable once marker graph, directory separation, start/wait accounting and resource rules remain unchanged. Production inventory now has four ordered inputs, including bindings_macho.go. Root must freeze a new prospective source/CI/resource/readiness/binary/binding chain and perform actual once execution separately. Old readiness and binary hashes must not be reused as proof of a new chain.

This is static data comparison under immutable-input ownership, not compiler/kernel or post-initialization attestation. A soft256MiB Go memory limit is not an absolute RSS bound; whole-child Darwin RSS needs actual observation. Rights and proof content remain Root-owned. No Wants, labels, weights or observation results changed. REVIEW.v5.json binds exact sources and scope; CORRECTIONS.v5.json and ACTIVITY.v5.json retain limitations and failed metadata lookups.
