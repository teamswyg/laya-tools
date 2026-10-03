#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "$0")/.." && pwd -P)
proof="$repo_root/experiments/short-claim/next60-development37-bias-audit"
task_tmp=$(mktemp -d)
task_tmp=$(cd "$task_tmp" && pwd -P)
trap 'rm -rf "$task_tmp"' EXIT
task_go=${RIIDOLAYA_GO:-go}
task_gofmt=${RIIDOLAYA_GOFMT:-gofmt}
if [[ "$task_go" == */* ]] && [ -z "${RIIDOLAYA_GOFMT:-}" ]; then
  task_gofmt="$(dirname "$task_go")/gofmt"
fi
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=${GOMAXPROCS:-2}
for name in main.go main_test.go public_frozen_test.go; do
  cp "$proof/source/$name.txt" "$task_tmp/$name"
done
cat > "$task_tmp/go.mod" <<MOD
module riido.local/development37biasaudit

go 1.27.1

require github.com/teamswyg/laya-tools v0.0.0

replace github.com/teamswyg/laya-tools => $repo_root
MOD
cp "$repo_root/go.sum" "$task_tmp/go.sum"
task_out=${RIIDOLAYA_BIAS37_ACTUAL:-$task_tmp/actual.json}
mkdir -p "$(dirname "$task_out")"
task_parent=$(cd "$(dirname "$task_out")" && pwd -P)
task_out="$task_parent/$(basename "$task_out")"
(
  cd "$task_tmp"
  task_format=$("$task_gofmt" -l main.go main_test.go public_frozen_test.go)
  test -z "$task_format"
  env -u RIIDOLAYA_BIAS37_ACTUAL -u RIIDOLAYA_BIAS37_FROZEN "$task_go" test -race -count=1 -timeout=60s ./...
  env -u RIIDOLAYA_BIAS37_ACTUAL -u RIIDOLAYA_BIAS37_FROZEN GODEBUG=panicnil=1 "$task_go" test -race -count=1 -timeout=60s ./...
  "$task_go" vet ./...
  "$task_go" run . --data "$repo_root/experiments/short-claim/next60-development-thirtyseven/data/train.jsonl" --out "$task_out"
  RIIDOLAYA_BIAS37_ACTUAL="$task_out" RIIDOLAYA_BIAS37_FROZEN="$proof/results.json" "$task_go" test -count=1 -run '^TestSaved' -v ./...
)
printf '%s\n' 'Saved37 platform matrix and complete derived report verified; actual scores preserved; no original/model/Fit or whole-work claim.'
