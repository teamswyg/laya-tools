#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Saved token-plan evidence and synthetic controls; no trained model read.
set -euo pipefail
task_root=$(cd "$(dirname "$0")/.." && pwd)
task_packet="$task_root/experiments/short-claim/next-cohort-chi-audit/constructor-token-plan"
task_trial=$(mktemp -d)
task_cleanup() {
  task_status=$?
  trap - EXIT HUP INT TERM
  if ! rm -rf "$task_trial"; then
    printf '%s\n' 'token-plan temporary module cleanup failed' >&2
    if [ "$task_status" -eq 0 ]; then task_status=1; fi
  fi
  exit "$task_status"
}
trap task_cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
cd "$task_packet"
shasum -a 256 -c SHA256SUMS
node --input-type=module - "$task_root" "$task_packet" <<'NODE'
import fs from 'node:fs';
import crypto from 'node:crypto';
const [root, packet] = process.argv.slice(2);
const digest = b => crypto.createHash('sha256').update(b).digest('hex');
const bytes = fs.readFileSync(packet + '/SOURCE-PINS.public.v1.json');
const freeze = JSON.parse(fs.readFileSync(packet + '/FREEZE.public.v1.json'));
if (bytes.length !== freeze.source_pins.bytes || digest(bytes) !== freeze.source_pins.sha256) throw Error('source_manifest_pin');
for (const p of JSON.parse(bytes).files) {
  const b = fs.readFileSync(root + '/' + p.path);
  if (b.length !== p.bytes || digest(b) !== p.sha256) throw Error('frozen_source_pin');
}
NODE
node derivation-controls.mjs
node derive-source.mjs --repo "$task_root" \
  --baseline "$task_packet/reference/prepared-baseline.go.txt" --out "$task_trial"
cd "$task_trial"
CGO_ENABLED=0 GOFLAGS=-mod=readonly GOTOOLCHAIN=local go test -count=1 -v . \
  -tokenplan-saved "$task_packet/OBSERVATIONS.actual.public.v1.json.gz"
