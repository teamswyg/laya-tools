#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
bundle="$(cd "$(dirname "$0")" && pwd -P)"
repo="$(cd "$bundle/../../../.." && pwd -P)"
check_sums() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check --quiet "$1"
  else
    shasum -a 256 --check "$1" >/dev/null
  fi
}
cd "$bundle"
check_sums SHA256SUMS
cd "$repo"
check_sums "$bundle/SOURCE-SHA256SUMS"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off
export CGO_ENABLED=1 GOMAXPROCS=4 GOMEMLIMIT=256MiB GODEBUG=
[[ "$(go version)" == go\ version\ go1.27.1\ * ]] || { echo 'Go1.27.1 required' >&2; exit 1; }
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-prepared-contract.XXXXXX")"
trap 'rm -rf "$work"' EXIT
cp "$bundle/source/contract_test.go.txt" "$work/contract_test.go"
cp "$bundle/source/go.mod.template" "$work/go.mod"
cd "$work"
go mod edit -replace "github.com/teamswyg/laya-tools=$repo"
go test -race -count=1 -timeout=30s -v ./...
go vet ./...
echo 'Synthetic numeric contracts passed; no learned model download or training.'
