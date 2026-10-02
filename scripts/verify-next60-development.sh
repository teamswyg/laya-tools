#!/usr/bin/env bash
# Offline development-data checks and authored fake-process tests only.
# Never launches the original observation worker or a model.
set -euo pipefail

mode=${1:-data}
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
temp_base=$(cd "${RUNNER_TEMP:-${TMPDIR:-/tmp}}" && pwd -P)
check_dir=$(mktemp -d "$temp_base/riido-next60-ci.XXXXXX")
trap 'rm -rf "$check_dir"' EXIT

export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0
export GOMAXPROCS=1 GOMEMLIMIT=256MiB

case "$mode" in
  data)
    dataset="$repo_root/experiments/short-claim/next60-development-dataset"
    cp "$dataset/source/convert.go.txt" "$check_dir/convert.go"
    go run -p=1 "$check_dir/convert.go" \
      --inputs "$dataset/evidence/INPUTS.v3.json" \
      --qualified "$dataset/evidence/QUALIFIED-TRAIN-SUBSET.v1.json" \
      --out "$check_dir/train.jsonl"
    cmp "$check_dir/train.jsonl" "$dataset/data/train.jsonl"
    echo 'PASS: two finite development rows reproduced exactly; no fit or model execution.'
    three="$repo_root/experiments/short-claim/next60-development-three"
    cp "$three/source/validate.go.txt" "$check_dir/validate.go"
    go run -p=1 "$check_dir/validate.go" \
      --data "$three/data/train.jsonl" \
      --input "$repo_root/experiments/short-claim/next60-ftoa-actual-observation/INPUTS.v2.json" \
      > "$check_dir/validation.json"
    cmp "$check_dir/validation.json" "$three/INPUT-VALIDATION.v1.json"
    echo 'PASS: three finite rows/eight unit labels validated; no projection, score or fit.'
    ;;
  controller)
    # A separate Root command may select a real binary for static inspection.
    # This CI step must keep that optional fixture disabled and its parent
    # test process separate from the explicitly spawned synthetic child.
    unset RIIDO_FTOA_STATIC_WORKER RIIDO_FTOA_STATIC_BYTES RIIDO_FTOA_STATIC_SHA256
    unset RIIDO_FTOA_CLOSURE_SHA256 RIIDO_FTOA_VECTOR_SHA256 RIIDO_FTOA_READINESS_SHA256
    unset RIIDO_TEST_FAKE_CHILD RIIDO_TEST_FAKE_MODE RIIDO_TEST_FAKE_DIR
    if [[ $(uname -s) != Darwin ]]; then
      echo 'SKIP: Darwin-only authored controller tests.'
      exit 0
    fi
    source_root="$repo_root/experiments/short-claim/next60-ftoa-outside-preparation/root-tested/source"
    mkdir -p "$check_dir/module/pure"
    for name in main.go bindings_macho.go main_test.go bindings_macho_test.go go.mod; do
      cp "$source_root/$name.txt" "$check_dir/module/$name"
    done
    cp "$source_root/pure/protocol.go.txt" "$check_dir/module/pure/protocol.go"
    cd "$check_dir/module"
    # The race detector needs cgo. No original or native model package is present.
    CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./...
    CGO_ENABLED=1 go vet ./...
    echo 'PASS: authored synthetic controller tests; original worker and models not launched.'
    ;;
  *)
    echo 'Expected data or controller mode.' >&2
    exit 2
    ;;
esac
