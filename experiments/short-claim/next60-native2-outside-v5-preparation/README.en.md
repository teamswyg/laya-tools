# Native2 pre-execution correction v5

The previous gate searched buildinfo for three `-X` declarations. Root's actual build showed that Go1.27.1 deliberately omits these linker declarations when using `trimpath`. The gate failed before execution; original observations remain0 in this frozen preparation snapshot. The existing Root binary was read without running it.

The replacement inspects the three actual closure/vector/readiness SHA strings stored in a macOS ARM64 binary. It reads each16-byte Go string header and its referenced64-byte lowercase hexadecimal data, then compares it with the plan. Missing/duplicate symbols, invalid sections, addresses or lengths fail closed. No original function, binary help command or initialization runs. Scope is an unstripped thin little-endian Mach-O64 ARM64 executable.

Before parsing, limits are checked: file64MiB;128 commands/128KiB;64 sections;100,000 symbols;8MiB string table; symbol names1024 bytes each/8MiB copied in total;100,000 total relocations and100,000 indirect symbols. The standard buildinfo reader also invokes the Mach-O parser, so the bounded gate precedes it. Go1.27.1, module, CGO0, trimpath, darwin/arm64 and noVCS checks remain.

Final25 tests pass. Coverage includes synthetic malformed headers, missing/duplicate symbols, section/address/length errors, repeated name-copy bounds, blocking before the standard parser, and the exact existing Root binary matching all three values while rejecting an incorrect expected value. Across3 test runs, authored fake children start12 times with12 waits;3 deliberate missing-executable failures have0 waits. These are not original observations or model validation.

The execution binding must list4 controller source IDs in this order: `main.go`, `pure/protocol.go`, `go.mod`, `bindings_macho.go`. The2 test files are not runtime build inputs. The existing static reservation → readiness → native plan → execution binding → durable once intent graph is unchanged. Root must freeze new source/CI/resource proof before creating a fresh readiness/worker/binding chain. Prior v4 and the original observer seals remain intact.

Full source bytes are copied as inert `.go.txt` review material. Root owns CI, rights/resource interpretation, production compilation and the later once-only original observation. This author's standalone production build, original execution, model API, external publication and protected-set access counts are0. Root's existing builds and historical training are separate. Future actual results must be appended as a separate record.

Exact binary data inspection is not compiler-independent or kernel attestation and does not establish post-initialization runtime values or semantic correctness. Proof content interpretation remains Root-owned. Test success is not success of the23 original observations, automatic truth/role/weight assignment or approval for training/distribution. Frozen vectors, candidates and Wants are unchanged.

`ACTIVITY.v5.json` records actors and actual checks; `CORRECTIONS.v5.json` preserves correction history; `PUBLIC-FINAL-PINS.v5.json` and `SHA256SUMS` bind exact files. Public material contains no host paths, executable binaries or model weights.
