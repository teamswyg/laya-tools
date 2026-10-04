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
unset ACTUAL_PACKED_RESULT
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-packed-saved.XXXXXX")"
trap 'rm -rf "$work"' EXIT
cp source/check_test.go.txt "$work/check_test.go"
printf 'module riido-packed-saved-check\n\ngo 1.27.1\n' > "$work/go.mod"
(
  cd "$work"
  PACKED_BUNDLE="$bundle" go test -count=1 -v ./...
  go vet ./...
)
echo 'Saved packed observations verified; no model or raw profile reads.'
