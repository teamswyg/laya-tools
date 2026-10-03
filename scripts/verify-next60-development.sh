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
    cd "$repo_root"
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
    seven="$repo_root/experiments/short-claim/next60-development-seven"
    cp "$seven/source/validate.go.txt" "$check_dir/validate-seven.go"
    go run -p=1 "$check_dir/validate-seven.go" \
      --data "$seven/data/train.jsonl" \
      --data-sha256 1ee18e02958953452bf4f4c0e072652aede266c095f1866a6e71209058b1c5bc \
      --fixtures "$repo_root/experiments/short-claim/next60-four-selector-checkpoints/preparation-v2/FIXTURES.v1.json" \
      --previous-data "$three/data/train.jsonl" \
      > "$check_dir/validation-seven.json"
    cmp "$check_dir/validation-seven.json" "$seven/INPUT-VALIDATION.v1.json"
    echo 'PASS: seven finite rows/twenty unit labels; preserved three-row prefix; no fit.'
    sixteen="$repo_root/experiments/short-claim/next60-development-sixteen"
    cp "$sixteen/source/materializer/main.go.txt" "$check_dir/materialize-sixteen.go"
    go run -p=1 "$check_dir/materialize-sixteen.go" \
      --previous-data "$seven/data/train.jsonl" \
      --catalog "$repo_root/experiments/short-claim/next60-catalog10-literal-correction/CATALOG10.v2.json" \
      --qualification "$sixteen/ROOT-QUALIFICATION.v1.json" \
      --out "$check_dir/train-sixteen.jsonl" > "$check_dir/materialize-sixteen.json"
    cmp "$check_dir/train-sixteen.jsonl" "$sixteen/data/train.jsonl"
    cmp "$check_dir/materialize-sixteen.json" "$sixteen/MATERIALIZATION.v1.json"
    cp "$sixteen/source/validator/validate.go.txt" "$check_dir/validate-sixteen.go"
    cp "$sixteen/source/validator/reader-main.go.txt" "$check_dir/reader-sixteen.go"
    go run -p=1 "$check_dir/validate-sixteen.go" "$check_dir/reader-sixteen.go" \
      --data "$sixteen/data/train.jsonl" \
      --data-sha256 3222e493f4b5bfaf4de008a115a3969209b493d85e115df209394b2e1f0077a4 \
      --previous-data "$seven/data/train.jsonl" \
      --catalog "$repo_root/experiments/short-claim/next60-catalog10-literal-correction/CATALOG10.v2.json" \
      --adoption "$sixteen/ROOT-QUALIFICATION.v1.json" \
      --adoption-sha256 b2d5a732d7f5d9238883060a881fc0c8b21b08c5b26004a0c27ef374be0aa9d4 \
      --finite-comparison "$repo_root/experiments/short-claim/next60-nine-actual-observation/independent-semantic-review/FINITE-COMPARISON.v1.json" \
      > "$check_dir/validation-sixteen.json"
    cmp "$check_dir/validation-sixteen.json" "$sixteen/INPUT-VALIDATION.v1.json"
    echo 'PASS: sixteen finite rows/47 unit labels; exact seven-row prefix and evidence; no fit.'
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
  checkpoint)
    source_root="$repo_root/experiments/short-claim/next60-four-selector-checkpoints/root-tested/source"
    mkdir -p "$check_dir/module/batchprep" "$check_dir/module/checkpoint"
    cp "$source_root/go.mod.txt" "$check_dir/module/go.mod"
    for part in batchprep checkpoint; do
      for source_file in "$source_root/$part/"*.go.txt; do
        file_name=${source_file##*/}
        cp "$source_file" "$check_dir/module/$part/${file_name%.txt}"
      done
    done
    cd "$check_dir/module"
    # Exactly these two owned packages; no original imports or native main.
    CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./batchprep ./checkpoint
    CGO_ENABLED=1 go vet -p=1 ./batchprep ./checkpoint
    echo 'PASS: synthetic checkpoint failure controls; no original57 observation or model execution.'
    ;;
  outside)
    source_root="$repo_root/experiments/short-claim/next60-four-selector-outside-preparation/root-tested/source"
    mkdir -p "$check_dir/module/pure"
    for name in main.go files.go process.go macho.go main_test.go go.mod; do
      cp "$source_root/$name.txt" "$check_dir/module/$name"
    done
    cp "$source_root/pure/protocol.go.txt" "$check_dir/module/pure/protocol.go"
    cd "$check_dir/module"
    CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./...
    CGO_ENABLED=1 go vet -p=1 ./...
    echo 'PASS: fake lifecycle and metadata tests; no native worker or real child process.'
    ;;
  catalog9)
    archive="$repo_root/experiments/short-claim/next60-nine-actual-observation"
    export RIIDO_CATALOG9_TEST_CATALOG="$repo_root/experiments/short-claim/next60-catalog10-literal-correction/CATALOG10.v2.json"
    mkdir -p "$check_dir/outside/internal/protocol" "$check_dir/comparator"
    for name in main.go main_test.go go.mod; do
      cp "$archive/source/outside/$name.txt" "$check_dir/outside/$name"
    done
    for name in protocol.go protocol_test.go; do
      cp "$archive/source/outside/internal/protocol/$name.txt" "$check_dir/outside/internal/protocol/$name"
    done
    for name in compare.go compare_test.go go.mod; do
      cp "$archive/source/comparator/$name.txt" "$check_dir/comparator/$name"
    done
    # Both modules contain owned/std-library code only. These tests never
    # launch the original worker or a real child process.
    for part in outside comparator; do
      cd "$check_dir/$part"
      CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./...
      CGO_ENABLED=1 go vet -p=1 ./...
    done
    cd "$check_dir/comparator"
    go run -p=1 . \
      --catalog "$repo_root/experiments/short-claim/next60-catalog10-literal-correction/CATALOG10.v2.json" \
      --results "$archive/evidence/SAVED-OBSERVATIONS.v1.json" \
      --results-sha256 5b0634522473882a7823c33565d3db68a6f11efd28e2b3474a59f8d6d57bfb99 \
      --output "$check_dir/recomparison.json"
    cmp "$check_dir/recomparison.json" "$archive/independent-semantic-review/FINITE-COMPARISON.v1.json"
    echo 'PASS: synthetic controls and exact saved-result comparison; no native worker or model.'
    ;;
  sixteen)
    archive="$repo_root/experiments/short-claim/next60-development-sixteen/source"
    for part in materializer validator; do
      mkdir -p "$check_dir/$part"
      for source_file in "$archive/$part/"*.go.txt "$archive/$part/go.mod.txt"; do
        file_name=${source_file##*/}
        # The reader bridge imports this repository and has a main function;
        # the separate stdlib validator module tests its core without it.
        [[ "$file_name" == reader-main.go.txt ]] && continue
        cp "$source_file" "$check_dir/$part/${file_name%.txt}"
      done
      cd "$check_dir/$part"
      CGO_ENABLED=1 go test -race -p=1 -timeout=5m ./...
      CGO_ENABLED=1 go vet -p=1 ./...
    done
    echo 'PASS: synthetic sixteen-row generation and validation controls; no worker or model.'
    ;;
  *)
    echo 'Expected data, controller, checkpoint, outside, catalog9 or sixteen mode.' >&2
    exit 2
    ;;
esac
