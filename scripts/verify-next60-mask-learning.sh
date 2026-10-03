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
for source in "$bundle/source/"*.go.txt; do
  cp "$source" "$work/$(basename "${source%.txt}")"
done
cp "$bundle/source/go.mod.template" "$work/go.mod"
printf '\nreplace github.com/teamswyg/laya-tools => "%s"\n' "$repo" >> "$work/go.mod"
cd "$work"
[[ -z "$(gofmt -l ./*.go)" ]] || { echo 'Unformatted source snapshot' >&2; exit 1; }
go test -race -count=1 ./...
go vet ./...
echo 'Synthetic masked sparse columns, actual Go NLL/AUC and feature bounds passed; no Golden corpus, model or Fit.'
