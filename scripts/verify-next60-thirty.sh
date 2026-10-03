#!/usr/bin/env bash
# Offline saved-data and owned synthetic checks. No original worker or model.
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
archive="$repo_root/experiments/short-claim/next60-development-thirty"
temp_base=$(cd "${RUNNER_TEMP:-${TMPDIR:-/tmp}}" && pwd -P)
check_dir=$(mktemp -d "$temp_base/riido-thirty-ci.XXXXXX")
trap 'rm -rf -- "$check_dir"' EXIT

export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
export GOMAXPROCS=1 GOMEMLIMIT=256MiB CGO_ENABLED=0

# The immutable archive remains separate from the formatted project command.
(cd "$archive" && shasum -a 256 -c SHA256SUMS > /dev/null)
for source_set in data27-30 durable-worker outside-controller saved-comparer normalization; do
  source_root="$archive/source/$source_set"
  while IFS= read -r -d '' source_file; do
    relative=${source_file#"$source_root/"}
    output="$check_dir/$source_set/${relative%.txt}"
    mkdir -p "$(dirname "$output")"
    cp "$source_file" "$output"
  done < <(find "$source_root" -type f \( -name '*.go.txt' -o -name 'go.mod.txt' \) -print0)
  if [[ "$source_set" == durable-worker || "$source_set" == saved-comparer ]]; then
    mkdir -p "$check_dir/$source_set/basis"
    cp "$archive/evidence/FIXTURE-BINDINGS.v1.json" "$check_dir/$source_set/basis/FIXTURE-BINDINGS.v1.json"
  fi
  # Tests use owned fake callbacks/files/processes and saved literals only.
  (cd "$check_dir/$source_set/module" && \
    CGO_ENABLED=1 go test -mod=readonly -race -p=1 -timeout=5m ./... && \
    CGO_ENABLED=1 go vet -mod=readonly -p=1 ./...)
done

mkdir -p "$check_dir/output"
(cd "$check_dir/data27-30/module" && go run -mod=readonly -p=1 ./cmd/materialize \
  --mode 30 \
  --previous-data "$repo_root/experiments/short-claim/next60-development-twentythree/data/train.jsonl" \
  --subbody-adoption "$archive/qualification/SUB-BODY.v2.json" \
  --ini-adoption "$archive/qualification/INI-TWO.v1.json" \
  --group-decision "$archive/qualification/GROUP-DECISION.mode30.v2.json" \
  --group-decision-sha256 5c82aa14b01f34bc4af08e06c0d565eb5459cee27e55db82a2304e3c6d28f00d \
  --additional-adoption "$archive/qualification/REMAINING-THREE.v1.json" \
  --data-output "$check_dir/output/train.jsonl" \
  --receipt-output "$check_dir/output/materialization.json")
cmp "$check_dir/output/train.jsonl" "$archive/data/train.jsonl"
cmp "$check_dir/output/materialization.json" "$archive/MATERIALIZATION.v1.json"

cd "$repo_root"
go run -mod=readonly -p=1 ./cmd/riido-developmentverify \
  --data "$archive/data/train.jsonl" --metadata "$archive/MATERIALIZATION.v1.json" \
  --finite "$archive/evidence/REMAINING-THREE-PREDICATES.v1.json"
