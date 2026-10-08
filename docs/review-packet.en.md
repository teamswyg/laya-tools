# Read a bounded public review packet with receipts

[한국어](review-packet.ko.md)

`riido-reviewpacket` reads one regular file within a byte limit, verifies an
independently pinned SHA-256 digest, and releases the verified bytes on stdout.
It saves a durable start receipt before opening the input content and a separate
terminal result receipt before releasing any payload.

This is preparation tooling. It does not repair a failed pilot or audit, grant
S2 authority, authorize a cohort or Fit, or establish semantic review or code
correctness. All examples and tests are permanently public fixtures; they are
not formal data and do not become formal data by passing this tool.

## Command

```sh
go run ./cmd/riido-reviewpacket \
  --input ./public-fixture.go \
  --sha256 <independently-pinned-64-hex-digest> \
  --receipt ./receipts/read.json \
  --max-bytes 1048576 \
  --actor public-fixture-check
```

Replace the digest placeholder with a value independently established before
this invocation. Computing a digest from the current input and immediately
passing it back to this command does not provide an independent pin.

| Option | Contract |
| --- | --- |
| `--input` | Required path to one regular file. The final path component may not be a symlink. |
| `--sha256` | Required independent expected digest, exactly 64 hexadecimal characters. |
| `--receipt` | Required path for a new start receipt. Create its parent directory first and keep its pathname stable so the supplied paths still locate the artifacts. |
| `--max-bytes` | Input limit in bytes, from `1` through `8388608` (8 MiB); default `1048576` (1 MiB). |
| `--actor` | Optional public label. When supplied, it must be nonblank and at most 256 bytes; whitespace-only labels are rejected. It is a caller assertion, not verified identity. |

Input and receipt paths must not contain `..` parent traversal components.
Supported input-open platforms are macOS (`darwin`), `linux`, `freebsd`,
`openbsd`, `netbsd`, and `dragonfly`. Other platforms deny the content open.

There is no JSON-output flag. The receipt files are JSON for machine use;
stdout contains only the raw verified input bytes after success. Status and the
caller-supplied receipt path go to stderr. Errors use safe reasons and do not
include raw input text.

## Public fixture example

The following original fixture is exactly `package fixture` followed by a
single newline. Its digest, pinned here independently of the read invocation,
is `c9f09e0a76fbbf83847013f7d79aa09c1566778489cc499e76e3e5d72b1c8db6`.

```sh
printf 'package fixture\n' > ./public-fixture.go
mkdir -p ./receipts
go run ./cmd/riido-reviewpacket \
  --input ./public-fixture.go \
  --sha256 c9f09e0a76fbbf83847013f7d79aa09c1566778489cc499e76e3e5d72b1c8db6 \
  --receipt ./receipts/read.json \
  --actor public-fixture-check
```

On success, stdout is exactly the fixture's bytes. Preserve both
`./receipts/read.json` and `./receipts/read.json.result`. A later invocation must
use a new receipt name; existing artifacts are not overwritten or repaired.

## Read and receipt order

1. Validate options and path relationships. Invalid options, `..` parent
   traversal components, path aliases, or existing receipt artifacts can be
   rejected before a start receipt exists.
2. Exclusively create the immutable start receipt with mode `0600`, write and
   sync it, then sync its parent directory. Only after those steps succeed may
   this tool open the input content. Both receipts and directory syncs use the
   same retained parent handle even if its pathname is renamed or retargeted;
   existing parent permissions are not changed.
3. Open only a regular file. On the supported platforms the content open uses
   nonblocking and no-follow protections, rejects a final symlink, and checks
   the opened file's identity against the `lstat` and `stat` metadata. The read
   buffer is bounded to `max-bytes + 1`, allowing an oversized input to be
   rejected without an unbounded read.
4. Complete the bounded read before computing an actual SHA-256 digest. Missing
   inputs, read failures, oversized inputs, and digest mismatches are denied.
   An incomplete or oversized read is not reported as a completed input digest.
5. Exclusively create and sync the terminal sibling receipt, named by appending
   `.result` to the start receipt path, with mode `0600`. A successful result
   binds the SHA-256 digest of the start receipt's exact bytes and records the
   verification outcome. Only then may stdout receive the verified payload.

Start and result receipts omit the input/file path. The optional public actor
label may appear in the receipts. The result binds the exact stored start bytes,
not a reserialized or normalized JSON representation. Keep the pair together
when checking this invocation's evidence.

The tool performs no hidden retries. An input-capture or verification failure
after a durable start and before the terminal result produces a failed result
when storage permits it. A receipt write or sync failure suppresses
the payload and preserves existing artifacts; broken storage can prevent a
durable failure record. A start receipt without a completed result is incomplete
evidence. A successful result records verification before output, not proof that
a downstream consumer received all stdout bytes.

## What the evidence establishes

The receipts establish only **this tool's mediated intent before this input
content read, and byte verification before payload release**. They do not prove
an agent's earliest-ever read, a global start across other tools or processes,
review quality, correctness, provider/model identity, or S2 authority. The actor
label does not expand those claims.

Use this command and its fixtures to prepare and check the bounded receipt
procedure. Its success does not change preserved pilot or audit outcomes, confer
cohort/Fit eligibility, or authorize reads of formal data.
