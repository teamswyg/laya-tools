#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
archive="$repo_root/experiments/short-claim/next60-development-twentythree"
extra="$repo_root/experiments/short-claim/next60-extra-two-actual-observation"
previous="$repo_root/experiments/short-claim/next60-development-twentyone/data/train.jsonl"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/riido-next60-23.XXXXXX")"
work_dir="$(cd "$work_dir" && pwd -P)"
trap 'rm -rf -- "$work_dir"' EXIT

export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
export GOMAXPROCS=1 GOMEMLIMIT=256MiB CGO_ENABLED=0
mkdir -p "$work_dir/core" "$work_dir/materializer" "$work_dir/reader" "$work_dir/output"
for name in compose.go data23_test.go files.go go.mod strict.go types.go validate.go; do
  cp "$archive/source/data23/$name.txt" "$work_dir/core/$name"
done
cp "$archive/source/bridge/materializer-main.go.txt" "$work_dir/materializer/main.go"
cp "$archive/source/bridge/reader-main.go.txt" "$work_dir/reader/main.go"
cat > "$work_dir/materializer/go.mod" <<'MOD'
module example.com/riido-data23-materializer

go 1.27.1

require riido.local/data23prep v0.0.0
replace riido.local/data23prep => ../core
MOD
cat > "$work_dir/reader/go.mod" <<MOD
module example.com/riido-data23-reader

go 1.27.1

require (
  riido.local/data23prep v0.0.0
  github.com/teamswyg/laya-tools v0.0.0
)
replace riido.local/data23prep => ../core
replace github.com/teamswyg/laya-tools => $repo_root
MOD

# Synthetic controls never dispatch original candidates or models.
(cd "$work_dir/core" && CGO_ENABLED=1 go test -mod=readonly -race -p=1 -timeout=5m ./... && go vet -mod=readonly -p=1 ./...)

# Reproduce existing Root-adopted supervision, then read values with the actual
# project reader. These calls do not perform features, ranking, Fit or inference.
(cd "$work_dir/materializer" && go run -mod=readonly . \
  --previous-data "$previous" --fixtures "$extra/FIXTURES.v1.json" \
  --adoption "$extra/ROOT-QUALIFICATION.v1.json" --finite-comparison "$extra/COMPARISON.v1.json" \
  --data-output "$work_dir/output/train.jsonl" --receipt-output "$work_dir/output/materializer.json")
cmp "$work_dir/output/train.jsonl" "$archive/data/train.jsonl"
(cd "$work_dir/reader" && go run -mod=readonly . \
  --data "$work_dir/output/train.jsonl" \
  --data-sha256 39edb1bb60e88d56cb2fec271b511a5ce09a0ff8bd33ce21d0dda7defb3ce362 \
  --previous-data "$previous" --fixtures "$extra/FIXTURES.v1.json" \
  --adoption "$extra/ROOT-QUALIFICATION.v1.json" --finite-comparison "$extra/COMPARISON.v1.json" \
  --attempt-dir "$work_dir/output/reader-attempt")
cmp "$work_dir/output/reader-attempt/results.json" "$archive/INPUT-VALIDATION.v1.json"
