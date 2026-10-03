#!/usr/bin/env bash
# Owned comparison failure controls only. No native originals or model execution.
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
temp_base=$(cd "${RUNNER_TEMP:-${TMPDIR:-/tmp}}" && pwd -P)
check_dir=$(mktemp -d "$temp_base/riido-ini-two-ci.XXXXXX")
trap 'rm -rf -- "$check_dir"' EXIT
archive="$repo_root/experiments/short-claim/next60-ini-two-actual-observation/source/comparer"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
export GOMAXPROCS=1 GOMEMLIMIT=256MiB
cp "$archive/go.mod.txt" "$check_dir/go.mod"
for source_dir in compare internal/deleted internal/quoted internal/testdelete cmd/compare; do
  mkdir -p "$check_dir/$source_dir"
  for source_file in "$archive/$source_dir/"*.go.txt; do
    file_name=${source_file##*/}
    cp "$source_file" "$check_dir/$source_dir/${file_name%.txt}"
  done
done
cd "$check_dir"
CGO_ENABLED=1 go test -mod=readonly -race -p=1 -timeout=5m ./...
CGO_ENABLED=1 go vet -mod=readonly -p=1 ./...
printf '%s\n' 'PASS: owned synthetic INI comparison controls; no original replay, model or training.'
