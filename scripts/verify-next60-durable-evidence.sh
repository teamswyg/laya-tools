#!/usr/bin/env bash
set -euo pipefail

# Reviewed owned controls only: no original worker, model download or training.
riido_repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
riido_case="$riido_repo/experiments/short-claim/next60-durable-evidence"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off GOFLAGS=
unset RIIDO_OWNED_PROFILE_DIR
[[ $(go version) == "go version go1.27.1 "* ]]
(cd "$riido_case" && shasum -a 256 -c SHA256SUMS >/dev/null)
riido_scratch=$(mktemp -d)
trap 'rm -rf "$riido_scratch"' EXIT
for riido_module in frame worker compact file baseline durable adapter; do
  mkdir -p "$riido_scratch/$riido_module"
  cp -R "$riido_case/source/$riido_module/." "$riido_scratch/$riido_module/"
  mv "$riido_scratch/$riido_module/go.mod.template" "$riido_scratch/$riido_module/go.mod"
  while IFS= read -r -d '' riido_source; do
    mv "$riido_source" "${riido_source%.txt}"
  done < <(find "$riido_scratch/$riido_module" -name '*.go.txt' -print0)
done
cp -R "$riido_scratch/worker/core/testdata" "$riido_scratch/adapter/testdata"
for riido_module in frame worker compact file durable adapter; do
  (
    cd "$riido_scratch/$riido_module"
    [[ -z $(gofmt -l .) ]]
    go test -race -count=1 ./...
    GODEBUG=panicnil=1 go test -race -count=1 ./...
    go vet ./...
  )
done
if [[ ${RIIDO_EVIDENCE_BENCH:-0} == 1 ]]; then
  (cd "$riido_scratch/file" && GOMAXPROCS=2 GOMEMLIMIT=64MiB go test -run '^$' -bench '^(BenchmarkMalformedReadback|BenchmarkValidReadback)$' -benchmem -benchtime=200ms -count=3)
fi
echo "Owned wire3/core/strict/file/durable/pipeline controls passed; originals/models/training not executed."
