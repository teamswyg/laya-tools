#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd -P)"
bundle="$repo/experiments/short-claim/next60-mask-learning-preparation"
cd "$bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --check --quiet SHA256SUMS
else
  shasum -a 256 --check SHA256SUMS >/dev/null
fi
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-mask-controls.XXXXXX")"
trap 'rm -rf "$work"' EXIT
mkdir "$work/mask" "$work/reader" "$work/adapter"
for source in "$bundle/source/"*.go.txt; do
  cp "$source" "$work/mask/$(basename "${source%.txt}")"
done
cp "$bundle/source/go.mod.template" "$work/mask/go.mod"
printf '\nreplace github.com/teamswyg/laya-tools => "%s"\n' "$repo" >> "$work/mask/go.mod"
reader_bundle="$repo/experiments/short-claim/publication-proof-127"
cd "$reader_bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --check --quiet SHA256SUMS
else
  shasum -a 256 --check SHA256SUMS >/dev/null
fi
for name in reader.go strict.go; do
  cp "$reader_bundle/source/golden-reader/$name.txt" "$work/reader/$name"
done
printf 'module riido.local/goldenreaderprototype\n\ngo 1.27.1\n\nrequire github.com/teamswyg/laya-tools v0.0.0\n\nreplace github.com/teamswyg/laya-tools => "%s"\n' "$repo" > "$work/reader/go.mod"
for source in "$bundle/adapter/"*.go.txt; do
  cp "$source" "$work/adapter/$(basename "${source%.txt}")"
done
printf 'module github.com/teamswyg/laya-tools/internal/goldenmaskadapterprototype\n\ngo 1.27.1\n\nrequire (\n github.com/teamswyg/laya-tools v0.0.0\n riido.local/goldenreaderprototype v0.0.0\n github.com/teamswyg/laya-tools/internal/maskbridgeprototype v0.0.0\n)\n\nreplace github.com/teamswyg/laya-tools => "%s"\nreplace riido.local/goldenreaderprototype => ../reader\nreplace github.com/teamswyg/laya-tools/internal/maskbridgeprototype => ../mask\n' "$repo" > "$work/adapter/go.mod"
for module in mask adapter; do
  (
    cd "$work/$module"
    [[ -z "$(gofmt -l ./*.go)" ]] || { echo 'Unformatted source snapshot' >&2; exit 1; }
    go test -race -count=1 ./...
    go vet ./...
  )
done
echo 'Synthetic Reader-to-mask, sparse columns, actual Go NLL/AUC and feature bounds passed; no Golden corpus, model or Fit.'
