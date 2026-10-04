#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "$0")/.." && pwd)
audit="$repo/experiments/short-claim/next-cohort-jwt-audit"
go_bin=${RIIDOLAYA_GO_BIN:-go}
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0
export GOMAXPROCS=1 GOTRACEBACK=none
(cd "$audit" && shasum -a 256 -c SHA256SUMS)
(cd "$repo" && "$go_bin" run ./experiments/short-claim/next-cohort-jwt-audit/replay)
# Owned Unicode/error controls only: this does not invoke ParseWithClaims.
(cd "$audit/observer" && CGO_ENABLED=1 "$go_bin" test -race -count=1 ./...)
if [ "$(uname -s)" = Darwin ] && [ "$(uname -m)" = arm64 ]; then
  # Owned in-memory/temporary-file controls; no child Start, key or model.
  (cd "$audit/runner" && CGO_ENABLED=1 "$go_bin" test -race -count=1 ./...)
fi
