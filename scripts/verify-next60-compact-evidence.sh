#!/usr/bin/env bash
set -euo pipefail

# Materialize only reviewed owned sources. No upstream worker, inference, model
# download, Fit or corpus qualification. Profiles stay disabled in this path.
riido_repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
riido_case="$riido_repo/experiments/short-claim/next60-compact-evidence-preparation"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off GOFLAGS=
unset RIIDO_OWNED_PROFILE_DIR
if [[ $(go version) != "go version go1.27.1 "* ]]; then
  echo "Go 1.27.1 is required" >&2
  exit 1
fi
(cd "$riido_case" && shasum -a 256 -c SHA256SUMS >/dev/null)
riido_scratch=$(mktemp -d)
trap 'rm -rf "$riido_scratch"' EXIT
mkdir -p "$riido_scratch/frame" "$riido_scratch/worker/core" "$riido_scratch/compact"
cp "$riido_case/source/frame/go.mod.template" "$riido_scratch/frame/go.mod"
for riido_name in encoding_bound_test frame original_call_test required_corrections_test writer writer_test; do
  cp "$riido_case/source/frame/$riido_name.go.txt" "$riido_scratch/frame/$riido_name.go"
done
cp "$riido_case/source/worker/go.mod.template" "$riido_scratch/worker/go.mod"
for riido_name in context duplicate failure_test fixtures original_binding_test result_budget_test run types union; do
  cp "$riido_case/source/worker/core/$riido_name.go.txt" "$riido_scratch/worker/core/$riido_name.go"
done
cp "$riido_case/source/compact/go.mod.template" "$riido_scratch/compact/go.mod"
for riido_name in compact compact_test profile_test; do
  cp "$riido_case/source/compact/$riido_name.go.txt" "$riido_scratch/compact/$riido_name.go"
done
for riido_module in frame worker compact; do
  (
    cd "$riido_scratch/$riido_module"
    [[ -z $(gofmt -l .) ]]
    go test -race -count=1 ./...
    go vet ./...
  )
done
for riido_module in frame worker; do
  (cd "$riido_scratch/$riido_module" && GODEBUG=panicnil=1 go test -race -count=1 ./...)
done
echo "Owned frame/marker/compact controls passed; no original worker, model or Fit executed."
