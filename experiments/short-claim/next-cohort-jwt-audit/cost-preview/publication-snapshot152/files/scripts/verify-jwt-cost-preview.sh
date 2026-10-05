#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "$0")/.." && pwd)
audit="$repo/experiments/short-claim/next-cohort-jwt-audit"
go_bin=${RIIDOLAYA_GO_BIN:-go}
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=1 GOTRACEBACK=none
(cd "$audit/cost-preview" && shasum -a 256 -c SHA256SUMS)
# Saved complete first response only; no original API, model or worker main.
(cd "$repo" && "$go_bin" run ./experiments/short-claim/next-cohort-jwt-audit/cost-replay)
# Synthetic ownership/key-context/error/deadline controls. These start a test
# binary, but never worker main or original JWT Parse/Validate/sign/key APIs.
(cd "$audit/cost-preview" && CGO_ENABLED=1 "$go_bin" test -race -count=1 ./...)
