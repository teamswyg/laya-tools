#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd -P)"
bundle="$repo/experiments/short-claim/publication-proof-127"
cd "$bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --check --quiet SHA256SUMS
else
  shasum -a 256 --check SHA256SUMS >/dev/null
fi
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-source-controls.XXXXXX")"
trap 'rm -rf "$work"' EXIT
for kind in leaf-archive golden-reader; do
  module="$work/$kind"
  mkdir "$module"
  for source in "$bundle/source/$kind/"*.go.txt; do
    cp "$source" "$module/$(basename "${source%.txt}")"
  done
  if [[ "$kind" == leaf-archive ]]; then
    cp "$bundle/source/$kind/go.mod.template" "$module/go.mod"
  else
    printf 'module riido.local/goldenreaderprototype\n\ngo 1.27.1\n\nrequire github.com/teamswyg/laya-tools v0.0.0\n\nreplace github.com/teamswyg/laya-tools => "%s"\n' "$repo" > "$module/go.mod"
  fi
  (
    cd "$module"
    [[ -z "$(gofmt -l ./*.go)" ]] || { echo 'Unformatted source snapshot' >&2; exit 1; }
    go test -race -count=1 ./...
    go vet ./...
  )
done
echo 'Owned temporary archive and synthetic Golden Reader controls passed; no research-target cleanup, original observer, model or Fit.'
