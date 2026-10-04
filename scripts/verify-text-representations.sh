#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
packet="$repo/experiments/short-claim/text-representation-preview"
(cd "$packet" && shasum -a 256 -c SHA256SUMS)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
go_bin=${GO:-go}
(cd "$repo" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 "$go_bin" build -mod=readonly -trimpath -o "$scratch/preview" ./cmd/riido-hintpreview)
node "$packet/compare.mjs" "$repo" "$scratch/preview" "$packet/RESULTS.actual.public.v1.json.gz"
