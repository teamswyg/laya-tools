#!/usr/bin/env bash
# Owned synthetic controls and published saved snapshots. No native worker launch.
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
temp_base=$(cd "${RUNNER_TEMP:-${TMPDIR:-/tmp}}" && pwd -P)
check_dir=$(mktemp -d "$temp_base/riido-extra-two-ci.XXXXXX")
trap 'rm -rf "$check_dir"' EXIT
archive="$repo_root/experiments/short-claim/next60-extra-two-actual-observation"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
export GOMAXPROCS=1 GOMEMLIMIT=256MiB CGO_ENABLED=0
export RIIDO_EXTRA2_PUBLIC_ARCHIVE="$archive"
for part in worker outside comparator; do
  mkdir -p "$check_dir/$part"
  cp "$archive/source/$part/go.mod.txt" "$check_dir/$part/go.mod"
  case "$part" in
    worker) dirs=(extra workerio) ;;
    outside) dirs=(. internal/protocol) ;;
    comparator) dirs=(. internal/protocol internal/semantic) ;;
  esac
  for source_dir in "${dirs[@]}"; do
    mkdir -p "$check_dir/$part/$source_dir"
    for source_file in "$archive/source/$part/$source_dir/"*.go.txt; do
      file_name=${source_file##*/}
      cp "$source_file" "$check_dir/$part/$source_dir/${file_name%.txt}"
    done
  done
  cd "$check_dir/$part"
  CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./...
  CGO_ENABLED=1 go vet -p=1 ./...
done
printf '%s\n' 'PASS: owned failure controls and all thirty saved comparison rows reproduced; no native worker, model or fit.'
