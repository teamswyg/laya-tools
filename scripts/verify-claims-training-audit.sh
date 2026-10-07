#!/usr/bin/env bash
# Reproduce one immutable own-public-training aggregate; no model or Fit.
set -euo pipefail

audit_directory='.cache/public-claims-training-audit'
audit_training="$audit_directory/training.jsonl"
test ! -L .cache
test ! -L "$audit_directory"
mkdir -p "$audit_directory"
test ! -L "$audit_training"

audit_download=''
audit_report=$(mktemp "$audit_directory/report.XXXXXX")
trap 'rm -f "$audit_report"; if [[ -n "$audit_download" ]]; then rm -f "$audit_download"; fi' EXIT

if [[ ! -e "$audit_training" ]]; then
  audit_download=$(mktemp "$audit_directory/download.XXXXXX")
  curl --fail --silent --show-error --location --max-time 60 --max-filesize 8388608 \
    'https://huggingface.co/JooYoon/riidolaya-development-claim-hints-mlp16-v0.1/resolve/31705b576851edba30f10c0e26fd22793ffc6f56/training.original.ai-reviewed.jsonl' \
    --output "$audit_download"
  mv "$audit_download" "$audit_training"
  audit_download=''
fi

go run ./cmd/riido-statehint-claims-training-audit \
  --training "$audit_training" \
  --training-sha256 c5b5a0adbc721d73213d4565c6a7c430094f31b8e7502fb5249e7ed40a3da837 \
  > "$audit_report"
if ! cmp experiments/state-hints-training-audit/RESULTS.development.json "$audit_report"; then
  cat "$audit_report"
  exit 1
fi
echo 'Pinned public training aggregate reproduced; no model inference or Fit.'
