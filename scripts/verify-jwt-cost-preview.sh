#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "$0")/.." && pwd)
audit="$repo/experiments/short-claim/next-cohort-jwt-audit"
go_bin=${RIIDOLAYA_GO_BIN:-go}
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=1 GOTRACEBACK=none
# Publication 152 remains immutable. Only these five mutable repository files
# resolve to exact historical bytes; all other rows check current experiment code.
(
  cd "$audit/cost-preview"
  export LC_ALL=C
  snapshot=publication-snapshot152
  manifest_sha=85fa75f9376497fe17b516818c7059d00d3b0ac0d218e795cfe60e1f25841d0a
  snapshot_sha=361110c9c25dd435e600f6fa0911209ec5dbdbed305ba3d85de935416bd7fd35
  actual=$(shasum -a 256 SHA256SUMS); actual=${actual%% *}
  [[ "$actual" == "$manifest_sha" ]] || { echo "Changed publication-152 manifest" >&2; exit 1; }
  cmp -s SHA256SUMS "$snapshot/ORIGINAL-SHA256SUMS.publication152.txt" || { echo "Changed original manifest copy" >&2; exit 1; }
  actual=$(shasum -a 256 "$snapshot/SHA256SUMS"); actual=${actual%% *}
  [[ "$actual" == "$snapshot_sha" ]] || { echo "Changed historical snapshot manifest" >&2; exit 1; }
  (cd "$snapshot" && shasum -a 256 -c SHA256SUMS)
  rows=0; current=0; historical=0
  while IFS= read -r row; do
    hash=${row:0:64}; path=${row:66}
    [[ "$hash" =~ ^[0-9a-f]{64}$ && "${row:64:2}" == "  " && -n "$path" ]] || { echo "Invalid manifest row" >&2; exit 1; }
    case "$path" in
      ../../../../scripts/verify-jwt-cost-preview.sh|../../../../.github/workflows/ci.yml|../../../../README.md|../../../../README.en.md|../../../../.gitleaks.toml)
        resolved="$snapshot/files/${path#../../../../}"
        historical=$((historical+1)) ;;
      ../cost-replay/main.go|../cost-replay/main_test.go)
        resolved="$path"; current=$((current+1)) ;;
      *)
        case "$path" in
          /*|.*|*\\*|*[[:space:]]*|*/../*|*/./*|*/..|*/.|*//*|*[^A-Za-z0-9_./-]*)
            echo "Unapproved manifest path: $path" >&2; exit 1 ;;
        esac
        resolved="$path"; current=$((current+1)) ;;
    esac
    printf '%s  %s\n' "$hash" "$resolved" | shasum -a 256 -c -
    rows=$((rows+1))
  done < SHA256SUMS
  [[ "$rows" == 45 && "$current" == 40 && "$historical" == 5 ]] || { echo "Unexpected publication-152 row counts" >&2; exit 1; }
)
# Saved complete first response only; no original API, model or worker main.
(cd "$repo" && "$go_bin" run ./experiments/short-claim/next-cohort-jwt-audit/cost-replay)
# Synthetic ownership/key-context/error/deadline controls. These start a test
# binary, but never worker main or original JWT Parse/Validate/sign/key APIs.
(cd "$audit/cost-preview" && CGO_ENABLED=1 "$go_bin" test -race -count=1 ./...)
