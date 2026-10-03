#!/usr/bin/env bash
set -euo pipefail

# Owned controls only. Original-library observation is a separate pinned run.
riido_repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
riido_prior="$riido_repo/experiments/short-claim/next60-durable-evidence"
riido_case="$riido_repo/experiments/short-claim/next60-native-wire3-execution"
riido_process_overlay="$riido_repo/experiments/short-claim/next60-native-process-cause-overlay"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off GOFLAGS=
unset RIIDO_OWNED_PROFILE_DIR
[[ $(go version) == "go version go1.27.1 "* ]]
(cd "$riido_prior" && shasum -a 256 -c SHA256SUMS >/dev/null)
(cd "$riido_case" && shasum -a 256 -c SHA256SUMS >/dev/null)
(cd "$riido_process_overlay" && shasum -a 256 -c SHA256SUMS >/dev/null)
riido_scratch=$(mktemp -d)
trap 'rm -rf "$riido_scratch"' EXIT
for riido_module in frame worker compact file baseline durable adapter; do
  mkdir -p "$riido_scratch/$riido_module"
  cp -R "$riido_prior/source/$riido_module/." "$riido_scratch/$riido_module/"
done
cp "$riido_case/source/core-overlay/union.go.txt" "$riido_scratch/worker/core/union.go.txt"
cp "$riido_case/source/core-overlay/cache_reset_test.go.txt" "$riido_scratch/worker/core/cache_reset_test.go.txt"
for riido_module in controller comparer archive compiler; do
  mkdir -p "$riido_scratch/$riido_module"
  cp -R "$riido_case/source/$riido_module/." "$riido_scratch/$riido_module/"
done
cp "$riido_process_overlay/process.go.txt" "$riido_scratch/controller/process.go.txt"
cp "$riido_process_overlay/cause_test.go.txt" "$riido_scratch/controller/cause_test.go.txt"
for riido_module in frame worker compact file baseline durable adapter controller comparer archive compiler; do
  mv "$riido_scratch/$riido_module/go.mod.template" "$riido_scratch/$riido_module/go.mod"
  while IFS= read -r -d '' riido_source; do
    mv "$riido_source" "${riido_source%.txt}"
  done < <(find "$riido_scratch/$riido_module" -name '*.go.txt' -print0)
done
cp -R "$riido_scratch/worker/core/testdata" "$riido_scratch/adapter/testdata"
for riido_module in frame worker compact file durable adapter controller comparer archive; do
  (
    cd "$riido_scratch/$riido_module"
    [[ -z $(gofmt -l .) ]]
    go test -race -count=1 ./...
    GODEBUG=panicnil=1 go test -race -count=1 ./...
    go vet ./...
  )
done
(cd "$riido_scratch/compiler" && [[ -z $(gofmt -l .) ]] && go vet ./...)
echo "Owned cache-reset, transport, saved-predicate and archive controls passed; no original/model execution or training."
