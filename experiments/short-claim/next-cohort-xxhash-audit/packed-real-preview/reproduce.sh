#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
bundle="$(cd "$(dirname "$0")" && pwd -P)"
repo="$(cd "$bundle/../../../.." && pwd -P)"
: "${MODEL_FILE:?Set MODEL_FILE to the existing pinned model file}"
case "$MODEL_FILE" in /*) ;; *) MODEL_FILE="$PWD/$MODEL_FILE" ;; esac
[[ -f "$MODEL_FILE" ]] || { echo 'Model file unavailable' >&2; exit 1; }
bash "$bundle/verify.sh" >&2
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0
export GOMAXPROCS=1 GOMEMLIMIT=256MiB GODEBUG=''
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-packed-preview.XXXXXX")"
trap 'rm -rf "$work"' EXIT
cp "$bundle/source/main.v2.go.txt" "$work/main.go"
cp "$bundle/source/go.mod.template" "$work/go.mod"
mkdir "$work/checks"
cp "$bundle/source/check_test.go.txt" "$work/checks/check_test.go"
printf 'module riido-packed-replay-check\n\ngo 1.27.1\n' > "$work/checks/go.mod"
cp "$bundle/input.json" "$work/input.json"
cp "$bundle/asset.reference.json" "$work/asset.json"
ln -s "$MODEL_FILE" "$work/model.hbin"
(
  cd "$work"
  go mod edit -replace "github.com/teamswyg/laya-tools=$repo"
  go mod tidy
  go run . input.json asset.json > actual.json
  cd checks
  PACKED_BUNDLE="$bundle" ACTUAL_PACKED_RESULT="$work/actual.json" go test -count=1 -v ./... >&2
)
cat "$work/actual.json"
echo 'One existing parent repeated; timings may differ. No training or activation.' >&2
