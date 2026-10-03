#!/usr/bin/env bash
set -euo pipefail
riido_repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
riido_case="$riido_repo/experiments/short-claim/next60-development-thirtyseven"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off GOFLAGS=
[[ $(go version) == "go version go1.27.1 "* ]]
(cd "$riido_case" && shasum -a 256 -c SHA256SUMS >/dev/null)
riido_created=$(mktemp -d)
riido_scratch=$(cd "$riido_created" && pwd -P)
trap 'rm -rf "$riido_scratch"' EXIT
cp "$riido_case/source/data37/main.go.txt" "$riido_scratch/main.go"
sed "s|@REPOSITORY@|$riido_repo|g" "$riido_case/source/data37/go.mod.template" > "$riido_scratch/go.mod"
(
  cd "$riido_scratch"
  [[ -z $(gofmt -l .) ]]
  go vet ./...
  go run . "$riido_repo/experiments/short-claim/next60-development-thirtyfive/data/train.jsonl" \
    "$riido_case/ROOT-QUALIFIED-ROWS.v1.json" \
    3d4b71af90e22c34c35efab9832e7b7cff2105bb6ac982f1b679b34c37d49925 \
    "$riido_scratch/train.jsonl" "$riido_scratch/receipt.json"
)
cmp "$riido_scratch/train.jsonl" "$riido_case/data/train.jsonl"
cmp "$riido_scratch/receipt.json" "$riido_case/MATERIALIZATION.actual.public.v1.json"
echo "37 saved development rows and full project Reader values matched; no original/model/Fit execution."
