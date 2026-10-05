#!/usr/bin/env bash
# Source only until reviewed and run by ordinary public CI. Never executes worker main.
set -euo pipefail
cd "$(dirname "$0")"
mode=${1:-normal}
case "$mode" in normal|race|producer|publication) ;; *) printf '%s\n' 'unsupported CI mode' >&2; exit 2;; esac
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off GOVCS='*:off'
export GOFLAGS='' GOEXPERIMENT='' GODEBUG=goindex=0 GOTRACEBACK=none
export CGO_ENABLED=0
if [ "$mode" = race ]; then export CGO_ENABLED=1; fi
proof="ci-artifacts/proof/$mode"
private="ci-private/$mode"
if [ "$mode" = publication ]; then
  proof="ci-artifacts/proof"
  test -d "$proof"
else
  test ! -e "$proof"
fi
test ! -e "$private"
mkdir -p "$proof" "$private"
sha() { shasum -a 256 "$1" | awk '{print $1}'; }
capture() {
  local label=$1; shift
  test ! -e "$proof/$label.stdout"
  printf '%s\n' "$@" > "$proof/$label.argv"
  set +e
  "$@" > "$proof/$label.stdout" 2> "$proof/$label.stderr"
  local status=$?
  set -e
  printf '%s\n' "$status" > "$proof/$label.status"
  # Full first stdout/stderr/status exist before this gate. No automatic retry.
  if [ "$status" -ne 0 ]; then printf '%s\n' "CI first stage failed: $label (status $status)" >&2; return "$status"; fi
}
capture_private() {
  local label=$1; shift
  test ! -e "$private/$label.stdout"
  set +e
  "$@" > "$private/$label.stdout" 2> "$private/$label.stderr"
  local status=$?
  set -e
  printf '%s\n' "$status" > "$private/$label.status"
  # Private CI-only first locators precede parsing; no public host-path claim.
  if [ "$status" -ne 0 ]; then printf '%s\n' "CI private first stage failed: $label (status $status)" >&2; return "$status"; fi
}
if [ "$mode" = publication ]; then
  # Scan full retained raw logs; unsafe first data remains private and blocks upload.
  capture_private publication-gate node ci-proof.mjs publication all "$proof" "$private"
  exit 0
fi
capture source-check shasum -a 256 -c SHA256SUMS
cp SHA256SUMS "$proof/SOURCE45.sha256"
capture version go version
capture environment go env -json GOVERSION GOOS GOARCH CGO_ENABLED GOTOOLCHAIN GOWORK GOENV GODEBUG GOFLAGS GOEXPERIMENT
# Host paths only locate the installed SDK/cache. They are not uploaded or published.
capture_private sdk-locations go env -json GOROOT GOTOOLDIR GOCACHE
host_os=$(awk -F '"' '$2 == "GOOS" {print $4}' "$proof/environment.stdout")
host_arch=$(awk -F '"' '$2 == "GOARCH" {print $4}' "$proof/environment.stdout")
if [ "$mode" = producer ]; then
  test "$host_os" = darwin
  test "$host_arch" = arm64
  test "${RUNNER_OS:-}" = macOS
fi
capture_private environment-proof node ci-proof.mjs environment "$mode" "$proof" "$private"
# This template emits portable package/source rows, never Dir or host paths.
template='{{if .Error}}E|{{.ImportPath}}{{"\n"}}{{end}}P|{{.ImportPath}}|{{.Standard}}|{{with .Module}}{{.Path}}{{end}}{{"\n"}}{{if .Standard}}{{range .GoFiles}}S|{{$.ImportPath}}|GoFiles|{{.}}{{"\n"}}{{end}}{{range .CgoFiles}}S|{{$.ImportPath}}|CgoFiles|{{.}}{{"\n"}}{{end}}{{range .CFiles}}S|{{$.ImportPath}}|CFiles|{{.}}{{"\n"}}{{end}}{{range .CXXFiles}}S|{{$.ImportPath}}|CXXFiles|{{.}}{{"\n"}}{{end}}{{range .MFiles}}S|{{$.ImportPath}}|MFiles|{{.}}{{"\n"}}{{end}}{{range .HFiles}}S|{{$.ImportPath}}|HFiles|{{.}}{{"\n"}}{{end}}{{range .FFiles}}S|{{$.ImportPath}}|FFiles|{{.}}{{"\n"}}{{end}}{{range .SFiles}}S|{{$.ImportPath}}|SFiles|{{.}}{{"\n"}}{{end}}{{range .SwigFiles}}S|{{$.ImportPath}}|SwigFiles|{{.}}{{"\n"}}{{end}}{{range .SwigCXXFiles}}S|{{$.ImportPath}}|SwigCXXFiles|{{.}}{{"\n"}}{{end}}{{range .SysoFiles}}S|{{$.ImportPath}}|SysoFiles|{{.}}{{"\n"}}{{end}}{{range .EmbedFiles}}S|{{$.ImportPath}}|EmbedFiles|{{.}}{{"\n"}}{{end}}{{end}}'
if [ "$mode" = race ]; then
  capture selected-source-rows go list -mod=readonly -deps -test -race -f "$template" .
elif [ "$mode" = producer ]; then
  capture selected-source-rows go list -mod=readonly -deps -f "$template" .
else
  capture selected-source-rows go list -mod=readonly -deps -test -f "$template" .
fi
# Retain full generated source, using a private CI-only cache locator.
# Its unportable absolute path is deliberately not a public raw-Go-list claim.
if [ "$mode" != producer ]; then
capture_private testmain-locator go list -mod=readonly -test -f '{{if eq .ImportPath "riido.example/jwt151-cost-preview.test"}}{{range .GoFiles}}{{printf "%s\n" .}}{{end}}{{end}}' .
fi
capture_private sources-proof node ci-proof.mjs sources "$mode" "$proof" "$private"
if [ "$mode" = normal ]; then
  capture owned-tests go test -p=1 -mod=readonly -trimpath -buildvcs=false -vet=off -count=1 -timeout=60s -v .
  capture owned-vet go vet -p=1 -mod=readonly -trimpath -buildvcs=false .
elif [ "$mode" = race ]; then
  capture owned-race go test -p=1 -mod=readonly -trimpath -buildvcs=false -race -vet=off -count=1 -timeout=60s -v .
else
  # The producer is CGO0 Darwin/arm64. It does not inherit CGO1 race closure.
  test -s ci-artifacts/proof/normal/PROOF.json
  mkdir -p ci-artifacts/bin
  test ! -e ci-artifacts/bin/retrieval-preview
  capture build go build -p=1 -mod=readonly -trimpath -buildvcs=false '-ldflags=-s -w' -o ci-artifacts/bin/retrieval-preview .
  capture buildinfo go version -m -json ci-artifacts/bin/retrieval-preview
  # Prospective local transfer gate; a larger binary is a retained CI failure.
  binary_bytes=$(wc -c < ci-artifacts/bin/retrieval-preview | tr -d ' ')
  test "$binary_bytes" -le 4194304
fi
capture source-after shasum -a 256 -c SHA256SUMS
capture_private finish-proof node ci-proof.mjs finish "$mode" "$proof" "$private"
