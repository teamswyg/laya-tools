#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Historical source witness only. No worker, model, training or new cost sample.
set -euo pipefail

task_root=$(cd "$(dirname "$0")/.." && pwd)
task_snapshot=$(mktemp -d)
task_cleanup() {
  task_status=$?
  trap - EXIT HUP INT TERM
  if ! rm -rf "$task_snapshot"; then
    printf '%s\n' 'historical source witness cleanup failed' >&2
    if [ "$task_status" -eq 0 ]; then task_status=1; fi
  fi
  exit "$task_status"
}
trap task_cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
cd "$task_root"

# The unchanged cost-pilot verifier checks every copied byte against its original
# eleven source pins. The old Prepare is an exact public source witness, not the
# current runtime. A mismatch in any other current file still fails closed.
for task_source in \
  internal/hintlearn/learn.go \
  internal/hintlearn/model.go \
  internal/lexicalhint/audit_provenance.go \
  internal/lexicalhint/features.go \
  pkg/hintweights/doc.go \
  pkg/hintweights/view.go \
  pkg/shortclaim/audit_provenance.go \
  pkg/shortclaim/baseline.go \
  pkg/shortclaim/input.go \
  LICENSE
do
  mkdir -p "$task_snapshot/$(dirname "$task_source")"
  cp "$task_source" "$task_snapshot/$task_source"
done
mkdir -p "$task_snapshot/pkg/hintprepared"
cp experiments/short-claim/next-cohort-chi-audit/cost-pilot/reference-repo/prepared.go.txt \
  "$task_snapshot/pkg/hintprepared/prepared.go"

go run ./experiments/short-claim/next-cohort-chi-audit/cost-pilot/run \
  -repo "$task_snapshot" \
  -saved experiments/short-claim/next-cohort-chi-audit/cost-pilot/OBSERVATIONS.actual.public.v1.json.gz
