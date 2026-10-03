#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd -P)"
bundle="$repo/experiments/short-claim/next60-union-duplicate-preparation"
cd "$bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --check --quiet SHA256SUMS
else
  shasum -a 256 --check SHA256SUMS >/dev/null
fi
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-union-controls.XXXXXX")"
trap 'rm -rf "$work"' EXIT
mkdir "$work/frame" "$work/artifact-seal" "$work/durable"
for name in frame.go writer.go writer_test.go encoding_bound_test.go required_corrections_test.go; do
  cp "$bundle/source/frame/$name.txt" "$work/frame/$name"
done
for name in seal.go links.go seal_test.go; do
  cp "$bundle/source/artifact-seal/$name.txt" "$work/artifact-seal/$name"
done
for name in protocol.go protocol_test.go; do
  cp "$bundle/source/durable/$name.txt" "$work/durable/$name"
done
for module in frame artifact-seal durable; do
  cp "$bundle/source/$module/go.mod.template" "$work/$module/go.mod"
  (
    cd "$work/$module"
    [[ -z "$(gofmt -l ./*.go)" ]] || { echo 'Unformatted source snapshot' >&2; exit 1; }
    go test -race -count=1 ./...
    go vet ./...
    if [[ "$module" != artifact-seal ]]; then
      GODEBUG=panicnil=1 go test -race -count=1 ./...
    fi
  )
done
echo 'Owned wire, fake durable ACK and ordinary file-drift controls passed; no original worker, model or Fit.'
