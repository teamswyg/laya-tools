#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "$0")/.." && pwd -P)
proof="$repo_root/experiments/short-claim/publication-proof-125"
task_tmp=$(mktemp -d)
task_tmp=$(cd "$task_tmp" && pwd -P)
trap 'rm -rf "$task_tmp"' EXIT
task_go=${RIIDOLAYA_GO:-go}
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
export GOMAXPROCS=${GOMAXPROCS:-2}
for module in input-bridge bias; do
  stage="$task_tmp/$module"
  mkdir -p "$stage"
  if [ "$module" = input-bridge ]; then
    name=variablefitbridge
    source_paths=(bridge/types.go.txt bridge/plan.go.txt bridge/admit.go.txt bridge/bridge_test.go.txt)
  else
    name=developmentbiasaudit
    source_paths=(audit/audit.go.txt audit/audit_test.go.txt audit/public_frozen_test.go.txt cmd/biasaudit/main.go.txt)
  fi
  for relative in "${source_paths[@]}"; do
    destination="$stage/${relative%.txt}"
    mkdir -p "$(dirname "$destination")"
    cp "$proof/source/$module/$relative" "$destination"
  done
  cat > "$stage/go.mod" <<MOD
module github.com/teamswyg/laya-tools/experiments/$name

go 1.27.1

require github.com/teamswyg/laya-tools v0.0.0

replace github.com/teamswyg/laya-tools => $repo_root
MOD
  cp "$repo_root/go.sum" "$stage/go.sum"
  (
    cd "$stage"
    if [ "$module" = bias ]; then
      export RIIDOLAYA_BIAS_FROZEN="$proof/bias-audit/results.json"
      export RIIDOLAYA_BIAS_DATA="$repo_root/experiments/short-claim/next60-development-thirtythree/data/train.jsonl"
    fi
    "$task_go" test -race -count=1 -p=1 -timeout=2m ./...
    "$task_go" vet ./...
  )
done
printf '%s\n' 'Synthetic input controls and full nonlearned33 bias reproduction passed; no model fit/original observer.'
