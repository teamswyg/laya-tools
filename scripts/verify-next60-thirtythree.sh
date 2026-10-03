#!/usr/bin/env bash
# Owned synthetic controls and saved observations only; no original worker/model.
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
archive="$repo_root/experiments/short-claim/next60-development-thirtythree"
previous="$repo_root/experiments/short-claim/next60-development-thirty"
temp_base=$(cd "${RUNNER_TEMP:-${TMPDIR:-/tmp}}" && pwd -P)
check_dir=$(mktemp -d "$temp_base/riido-thirtythree-ci.XXXXXX")
trap 'rm -rf -- "$check_dir"' EXIT

export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
export GOMAXPROCS=1 GOMEMLIMIT=256MiB CGO_ENABLED=0
(cd "$archive" && shasum -a 256 -c SHA256SUMS > /dev/null)

copy_sources() {
  local source_root=$1 destination=$2 source_file relative output
  while IFS= read -r -d '' source_file; do
    relative=${source_file#"$source_root/"}
    output="$destination/${relative%.txt}"
    mkdir -p "$(dirname "$output")"
    cp "$source_file" "$output"
  done < <(find "$source_root" -type f \( -name '*.go.txt' -o -name 'go.mod.txt' \) -print0)
}

copy_sources "$archive/source/saved-comparer" "$check_dir/saved-comparer"
copy_sources "$archive/source/data33/module" "$check_dir/data33"
for source_set in saved-comparer data33; do
  # The public comparer fixture is mandatory here; absent paths fail its test.
  (cd "$check_dir/$source_set" && \
    RIIDO_NATIVE_THREE_PUBLIC_ROOT="$archive" CGO_ENABLED=1 \
      go test -mod=readonly -race -p=1 -timeout=5m ./... && \
    CGO_ENABLED=1 go vet -mod=readonly -p=1 ./...)
done

bindings=(
  --previous-data "$previous/data/train.jsonl"
  --prior-summary "$previous/MATERIALIZATION.v1.json"
  --qualification "$archive/qualification/ROOT-QUALIFICATION.v3.json"
  --qualification-sha256 c1c4a531a7e9a6197bfc0f27f091b9ff54a1e58b77b9ad04e9f20bacae1d6cb4
  --literal "$archive/evidence/LITERAL.v1.json"
  --captions "$archive/evidence/CAPTIONS.v1.json"
  --comparison "$archive/evidence/SAVED-COMPARISON.v1.json"
  --family-decision "$archive/evidence/FAMILY-ROLE-PREFREEZE.v1.json"
)
(cd "$check_dir/data33" && go run -mod=readonly -p=1 ./cmd/materialize \
  "${bindings[@]}" --out "$check_dir/materialized")
cmp "$check_dir/materialized/train.jsonl" "$archive/data/train.jsonl"
cmp "$check_dir/materialized/MATERIALIZATION.json" "$archive/MATERIALIZATION.v1.json"

copy_sources "$archive/source/data33/reader" "$check_dir/reader"
# Local paths exist only in this disposable module, never in a public manifest.
cat > "$check_dir/reader/go.mod" <<EOF
module riido.local/development33reader

go 1.27.1

require (
    github.com/teamswyg/laya-tools v0.0.0
    riido.local/development33prep v0.0.0
    golang.org/x/text v0.25.0 // indirect
)

replace github.com/teamswyg/laya-tools => "$repo_root"
replace riido.local/development33prep => "$check_dir/data33"
EOF
cp "$repo_root/go.sum" "$check_dir/reader/go.sum"
(cd "$check_dir/reader" && go vet -mod=readonly -p=1 ./... && \
  go run -mod=readonly -p=1 . "${bindings[@]}" \
  --data "$archive/data/train.jsonl" \
  --data-sha256 7b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919 \
  --data-bytes 45390 --out "$check_dir/reader-output")
cmp "$check_dir/reader-output/READER.json" "$archive/READER.v1.json"
