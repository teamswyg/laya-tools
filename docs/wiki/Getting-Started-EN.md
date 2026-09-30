# Getting started

[한국어](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-KO) · [Home](https://github.com/teamswyg/laya-tools/wiki)

Start with keyword search. You can get a useful result without downloading Laya, connecting Codex, or supplying credentials.

## 1. Install the executable

Open [the latest release](https://github.com/teamswyg/laya-tools/releases/latest). Choose `darwin-arm64` for Apple Silicon macOS or `linux-amd64` for x86-64 Linux. Intel Macs and Linux ARM do not currently have published binaries. Download the archive and SHA256SUMS, compare the archive's SHA-256 with its matching entry, then extract it.

For example, the published v0.3.0 Apple Silicon archive can be checked and opened with:

```sh
shasum -a 256 riidolaya-v0.3.0-darwin-arm64.tar.gz
# Compare the printed digest with the matching SHA256SUMS line before continuing.
tar -xzf riidolaya-v0.3.0-darwin-arm64.tar.gz
./riidolaya version
```

On Linux use `sha256sum` and the linux-amd64 archive. Keep LICENSE, NOTICE, and licenses/ with redistributed copies. You can run `./riidolaya` directly; examples below assume you have placed the executable on PATH.

Alternatively, build from source with Go 1.27.1 and a C compiler (macOS Command Line Tools):

```sh
git clone https://github.com/teamswyg/laya-tools.git
cd laya-tools
go build -trimpath -o bin/riidolaya ./cmd/riidolaya
./bin/riidolaya version
```

For source builds, replace `riidolaya` below with `./bin/riidolaya`. Python is not required for user execution.

## 2. Get a first result without a model

From a Git repository root:

```sh
riidolaya search --root . --lexical --json 'routing confidence'
```

In laya-tools this searches the implementation. In another project, use its identifiers. Expect paths, line ranges, and excerpts, or an empty result when terms do not match. This command does not send source to a provider. For non-root/subdirectory searches, install `rg` (ripgrep).

Flags must precede the query. Use a root you intend the tool to read.

## 3. Try repository preview or model planning

The following fixture paths are in a clone of laya-tools; they are not included in a binary-only installation.

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json
```

The first returns a fictional billing repository as a keyword candidate. The second calculates a plan over fictional models/prices. Neither performs work or calls a paid model. See [Workflows](https://github.com/teamswyg/laya-tools/wiki/Workflows-EN) before interpreting their statuses.

## 4. Add local inference only if needed

```sh
riidolaya setup
riidolaya doctor
riidolaya bench --iterations 10
```

Setup downloads about 446 MiB for the current model archive, plus the runtime. The installed model file is about 572 MiB; measured process RAM with the model is about 1.4 GiB. `doctor` reports paths/existence; it is not an accuracy test. `bench` actually loads and invokes the model.

Caches: `~/Library/Caches/laya-tools` on macOS; `~/.cache/laya-tools` on Linux. `LAYA_CACHE` changes the location. Current setup also installs corrected models-v2 notices and verifies checksums. Existing `LAYA_*` variables remain supported.

Next: [Choose a workflow](https://github.com/teamswyg/laya-tools/wiki/Workflows-EN). You do not need to register MCP, start a daemon, or connect Codex just to use this tool.
