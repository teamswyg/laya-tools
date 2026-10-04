#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
bundle="$(cd "$(dirname "$0")" && pwd -P)"
repo="$(cd "$bundle/../../../.." && pwd -P)"
: "${MODEL_FILE:?Set MODEL_FILE to the existing pinned model file}"
case "$MODEL_FILE" in /*) ;; *) MODEL_FILE="$PWD/$MODEL_FILE" ;; esac
[[ -f "$MODEL_FILE" ]] || { echo 'Model file unavailable' >&2; exit 1; }
bash "$bundle/verify.sh"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-hint-preview.XXXXXX")"
trap 'rm -rf "$work"' EXIT
cp "$bundle/source/main.go.txt" "$work/main.go"
cp "$bundle/source/go.mod.template" "$work/go.mod"
cp "$bundle/input.json" "$work/input.json"
cp "$bundle/asset.reference.json" "$work/asset.json"
ln -s "$MODEL_FILE" "$work/model.hbin"
(
  cd "$work"
  go mod edit -replace "github.com/teamswyg/laya-tools=$repo"
  go mod tidy
  go run . input.json asset.json > actual.json
)
cmp "$bundle/observations.actual.public.v1.json" "$work/actual.json"
echo 'One post-hoc reference repeated exactly; no training or activation.'
