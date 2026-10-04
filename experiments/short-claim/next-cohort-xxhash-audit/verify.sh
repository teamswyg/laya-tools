#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
bundle="$(cd "$(dirname "$0")" && pwd -P)"
cd "$bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --check --quiet SHA256SUMS
else
  shasum -a 256 --check SHA256SUMS >/dev/null
fi
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-xxhash-audit.XXXXXX")"
trap 'rm -rf "$work"' EXIT
mkdir "$work/xxhash"
cp source/main.v2.go.txt "$work/main.go"
cp source/verify_test.go.txt "$work/verify_test.go"
cp source/go.mod.template "$work/go.mod"
cp frozen.json "$work/frozen.json"
gzip -dc observations.actual.json.gz > "$work/observations.actual.json"
cp source/xxhash/go.mod.template "$work/xxhash/go.mod"
cp source/xxhash/LICENSE.txt "$work/xxhash/LICENSE.txt"
for name in xxhash.go xxhash_other.go xxhash_unsafe.go; do
  cp "source/xxhash/$name.txt" "$work/xxhash/$name"
done
(
  cd "$work"
  [[ -z "$(gofmt -l main.go verify_test.go)" ]] || { echo 'Unformatted owned observer' >&2; exit 1; }
  go test -tags=purego -count=1 ./...
  go vet -tags=purego ./...
  go run -tags=purego . frozen.json > actual.json.gz
)
cmp observations.actual.json.gz "$work/actual.json.gz"
echo '75 finite candidate trials reproduced exactly; one parent, audit only, no model or Fit.'
