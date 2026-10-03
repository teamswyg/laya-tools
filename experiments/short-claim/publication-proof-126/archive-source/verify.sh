#!/usr/bin/env bash
# Only owned temporary synthetic archive tests; no retained research targets.
set -euo pipefail
here=$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)
export GOPROXY=off GOSUMDB=off GOWORK=off GOTOOLCHAIN=local GOENV=off GOFLAGS='-mod=mod' CGO_ENABLED=1
[[ $(go env GOVERSION) == go1.27.1 ]]
cd "$here"
go test -race -count=1 ./...
go vet ./...
