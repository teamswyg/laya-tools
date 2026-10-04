#!/usr/bin/env bash
set -euo pipefail
mode=${1:-saved}
case "$mode" in saved|replay) ;; *) exit 2 ;; esac
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
packet="$repo/experiments/short-claim/next-cohort-btree-audit"
(cd "$packet" && shasum -a 256 -c SHA256SUMS)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
node "$packet/derive-source.mjs" "$packet" "$repo" "$scratch"
go_bin=${GO:-go}
(cd "$scratch" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off "$go_bin" test -mod=readonly -race -count=1 -v .)
node "$packet/verify-saved.mjs" "$packet"
node "$packet/replay-controls.mjs"
reader="$scratch/reader"
mkdir "$reader"
cp "$packet/cohort-reader.go.txt" "$reader/main.go"
cp "$scratch/go.mod" "$reader/go.mod"
cp "$scratch/go.sum" "$reader/go.sum"
ln -s "$repo" "$reader/repo"
ln -s "$scratch/upstream" "$reader/upstream"
(cd "$reader" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off "$go_bin" run -mod=readonly . "$packet/cohort.train.v1.jsonl" "$packet/ROW-BINDINGS.public.v1.json")
if [[ "$mode" == replay ]]; then
  (cd "$scratch" && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off "$go_bin" build -mod=readonly -trimpath -o audit .)
  node - "$scratch" <<'NODE'
const fs = require('node:fs');
const {spawnSync} = require('node:child_process');
const root = process.argv[2];
const r = spawnSync(root + '/audit', ['actual', root + '/INPUTS.public.v1.json'], {
  cwd: root, timeout: 30000, maxBuffer: 131072, encoding: 'buffer'
});
// Retain the first bounded stdout/stderr and status before acceptance. CI uses
// an owned runner.temp directory and uploads it even on a semantic failure.
if (process.env.RIIDOLAYA_BTREE_FRESH) {
  const path = require('node:path');
  const destination = path.resolve(process.env.RIIDOLAYA_BTREE_FRESH);
  fs.mkdirSync(path.dirname(destination), {recursive: true});
  if (r.stdout) fs.writeFileSync(destination, r.stdout, {flag: 'wx', mode: 0o600});
  if (r.stderr && r.stderr.length) fs.writeFileSync(destination + '.stderr.txt', r.stderr, {flag: 'wx', mode: 0o600});
  fs.writeFileSync(destination + '.status.json', JSON.stringify({status:r.status,signal:r.signal,error_code:r.error?.code||null,stdout_bytes:r.stdout?.length||0,stderr_bytes:r.stderr?.length||0,original_attempts:1,model_calls:0,Fit_calls:0}) + '\n', {flag:'wx',mode:0o600});
}
if (r.status !== 0 || r.error || !r.stdout || r.stdout.length > 131072 || (r.stderr && r.stderr.length)) {
  process.stderr.write('bounded Btree semantic replay failed\n'); process.exit(1);
}
fs.writeFileSync(root + '/fresh.json', r.stdout, {flag: 'wx', mode: 0o600});
NODE
  node "$packet/verify-saved.mjs" "$packet" "$scratch/fresh.json"
fi
