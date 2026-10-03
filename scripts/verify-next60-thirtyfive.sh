#!/usr/bin/env bash
# Prospective offline public-bundle reproduction. No native worker/model/Fit.
set -euo pipefail
umask 077
fail() { printf 'next60 offline verification: %s\n' "$*" >&2; exit 1; }
for tool in jq go cmp cp mkdir chmod wc mktemp; do
  command -v "$tool" >/dev/null 2>&1 || fail "missing tool: $tool"
done
if command -v sha256sum >/dev/null 2>&1; then hash_tool=(sha256sum)
elif command -v shasum >/dev/null 2>&1; then hash_tool=(shasum -a 256)
else fail 'missing SHA256 tool'; fi
hash_file() {
  local result
  result=$("${hash_tool[@]}" < "$1") || fail 'SHA256 read'
  printf '%s\n' "${result%% *}"
}
regular() {
  local parent
  [[ -f "$1" && ! -L "$1" ]] || fail "missing/nonregular input: $1"
  parent=$(cd -P "$(dirname "$1")" && pwd -P) || fail 'input parent'
  [[ "$parent" == "$(dirname "$1")" ]] || fail "symlink ancestor: $1"
}
check_pin() {
  regular "$1"
  [[ $(wc -c < "$1") -eq "$2" && $(hash_file "$1") == "$3" ]] || fail "input pin: $1"
}
script_dir=$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)
repo=$(cd -P "$script_dir/.." && pwd -P)
bundle="$repo/experiments/short-claim/next60-development-thirtyfive"
previous="$repo/experiments/short-claim/next60-development-thirtythree"
saved_files=(go.mod twoliteral/types.go compare/types.go compare/files.go
  compare/binding.go compare/shape.go compare/semantic.go
  compare/semantic_test.go cmd/compare/main.go)
data_files=(library.go strict.go library_test.go reservation_sync_test.go
  cmd/materialize/main.go cmd/materialize/main_test.go)
evidence_files=(FIXTURES.v1.json WANTS.v1.json CAPTIONS.v1.json ROOT-FREEZE.v2.json
  LITERAL-CORRECTION.v1.json FAMILY-ROLE-PREFREEZE.v1.json
  OUTSIDE.v1.json CHILD.v1.json SAVED-COMPARISON.v1.json)
required=(data/train.jsonl MATERIALIZATION.v1.json
  qualification/ROOT-QUALIFICATION.actual.public.v2.json
  saved-config.template.v1.json source/data35/go.mod)
for rel in "${saved_files[@]}"; do required+=("source/saved-comparer/$rel"); done
for rel in "${data_files[@]}"; do required+=("source/data35/$rel"); done
for rel in "${evidence_files[@]}"; do required+=("evidence/$rel"); done
regular "$bundle/SHA256SUMS"
verified=(); verified_count=0
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ "$line" =~ ^([0-9a-f]{64})\ \ (.+)$ ]] || fail 'malformed checksum row'
  expected=${BASH_REMATCH[1]}; rel=${BASH_REMATCH[2]}
  case "$rel" in
    ''|/*|.|..|../*|*/../*|*/..|./*|*/./*|*/.|*\\*|*//*) fail 'unsafe checksum path' ;;
  esac
  for ((i=0;i<verified_count;i++)); do
    [[ "${verified[$i]}" != "$rel" ]] || fail 'duplicate checksum path'
  done
  regular "$bundle/$rel"
  [[ $(hash_file "$bundle/$rel") == "$expected" ]] || fail "checksum: $rel"
  verified[$verified_count]="$rel"; verified_count=$((verified_count+1))
done < "$bundle/SHA256SUMS"
for needed in "${required[@]}"; do
  present=0
  for ((i=0;i<verified_count;i++)); do [[ "${verified[$i]}" != "$needed" ]] || present=1; done
  [[ "$present" -eq 1 ]] || fail "missing checksum binding: $needed"
done
check_pin "$previous/data/train.jsonl" 45390 7b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919
check_pin "$previous/MATERIALIZATION.v1.json" 4339 d3f4b4ecc8a33841d553435d81cf39854a8b3821e746ed90d710c2c3db20572a
check_pin "$bundle/qualification/ROOT-QUALIFICATION.actual.public.v2.json" 17473 024650dc9cb1e9c1c965aea40a8d31de31ebe459b645a24836eaa358d69b194f
check_pin "$bundle/evidence/SAVED-COMPARISON.v1.json" 106426 394017f724d59b6fd13f92e6ada53c20ccd9425f9a3d6b48ccb8dc8835ac87f4
[[ $(wc -c < "$bundle/data/train.jsonl") -eq 48618 ]] || fail 'frozen data35 size'
export GOPROXY=off GOSUMDB=off GOWORK=off GOTOOLCHAIN=local
export GOPRIVATE=none GONOPROXY=none GONOSUMDB=none GOVCS='*:off'
export GOENV=off GOFLAGS='-mod=mod' CGO_ENABLED=1
[[ $(go env GOVERSION) == go1.27.1 ]] || fail 'Go1.27.1 required'
created=$(mktemp -d "${TMPDIR:-/tmp}/next60-thirtyfive.XXXXXXXX")
scratch=$(cd -P "$created" && pwd -P)
cleanup() {
  [[ -n "$scratch" && "$scratch" != / && "${scratch##*/}" == next60-thirtyfive.* ]] || return 1
  rm -rf -- "$scratch"
}
trap cleanup EXIT
mkdir -p "$scratch/saved/twoliteral" "$scratch/saved/compare" "$scratch/saved/cmd/compare" \
  "$scratch/data35/cmd/materialize" "$scratch/evidence" "$scratch/bin"
for rel in "${saved_files[@]}"; do cp "$bundle/source/saved-comparer/$rel" "$scratch/saved/$rel"; done
for rel in "${data_files[@]}"; do cp "$bundle/source/data35/$rel" "$scratch/data35/$rel"; done
for rel in "${evidence_files[@]}"; do cp "$bundle/evidence/$rel" "$scratch/evidence/$rel"; done
chmod 600 "$scratch/evidence/OUTSIDE.v1.json" "$scratch/evidence/CHILD.v1.json"
quoted_repo=$(jq -Rnr --arg p "$repo" '$p | @json')
replace_seen=0; module_seen=0
while IFS= read -r line || [[ -n "$line" ]]; do
  case "$line" in
    'module riido.local/development35prep') module_seen=$((module_seen+1)); printf '%s\n' "$line" ;;
    'module '*) fail 'wrong materializer module' ;;
    'replace github.com/teamswyg/laya-tools => ../../../../..'|'replace github.com/teamswyg/laya-tools => "../../../../.."')
      replace_seen=$((replace_seen+1)); printf 'replace github.com/teamswyg/laya-tools => %s\n' "$quoted_repo" ;;
    'replace '*) fail 'unexpected materializer replace' ;;
    *) printf '%s\n' "$line" ;;
  esac
done < "$bundle/source/data35/go.mod" > "$scratch/data35/go.mod"
[[ "$replace_seen" -eq 1 && "$module_seen" -eq 1 ]] || fail 'missing/duplicate module binding'
regular "$repo/go.sum"; cp "$repo/go.sum" "$scratch/data35/go.sum"
# Rebuild Go struct key order; retain all template fields except pin.path.
jq -ce --arg e "$scratch/evidence" '
  def p($v;$name):
    if ($v|keys)==(["id","path","bytes","sha256"]|sort) and $v.path==("evidence/"+$name)
    then {id:$v.id,path:($e+"/"+$name),bytes:$v.bytes,sha256:$v.sha256}
    else error("bad saved pin/template path") end;
  if keys!=(["schema","outside","child","fixtures","wants","captions","Root_freeze","correction","family","Root_collector_exit_code","plan_sha256","worker_sha256","controller_sha256"]|sort)
  then error("bad saved config fields") else . end | . as $t | {
    schema:$t.schema,outside:p($t.outside;"OUTSIDE.v1.json"),child:p($t.child;"CHILD.v1.json"),
    fixtures:p($t.fixtures;"FIXTURES.v1.json"),wants:p($t.wants;"WANTS.v1.json"),
    captions:p($t.captions;"CAPTIONS.v1.json"),Root_freeze:p($t.Root_freeze;"ROOT-FREEZE.v2.json"),
    correction:p($t.correction;"LITERAL-CORRECTION.v1.json"),family:p($t.family;"FAMILY-ROLE-PREFREEZE.v1.json"),
    Root_collector_exit_code:$t.Root_collector_exit_code,plan_sha256:$t.plan_sha256,
    worker_sha256:$t.worker_sha256,controller_sha256:$t.controller_sha256}
' "$bundle/saved-config.template.v1.json" |
  jq -Rr 'gsub("&";"\\u0026")|gsub("<";"\\u003c")|gsub(">";"\\u003e")|gsub("\u2028";"\\u2028")|gsub("\u2029";"\\u2029")' \
  > "$scratch/saved-config.json"
saved_config_sha=$(hash_file "$scratch/saved-config.json")
jq -n --arg previous "$previous/data/train.jsonl" --arg metadata "$previous/MATERIALIZATION.v1.json" \
  --arg qualification "$bundle/qualification/ROOT-QUALIFICATION.actual.public.v2.json" \
  --arg comparison "$scratch/evidence/SAVED-COMPARISON.v1.json" '
  {schema:"riido-development35-config-v1",previous_data:$previous,prior_metadata:$metadata,
   qualification:$qualification,qualification_sha256:"024650dc9cb1e9c1c965aea40a8d31de31ebe459b645a24836eaa358d69b194f",comparison:$comparison}
' > "$scratch/data35-config.json"
data_config_sha=$(hash_file "$scratch/data35-config.json")
(cd "$scratch/saved"; go test -race -count=1 ./...; go vet ./...; go build -o "$scratch/bin/saved-compare" ./cmd/compare)
(cd "$scratch/data35"; go test -race -count=1 ./...; go vet ./...; go build -o "$scratch/bin/materialize" ./cmd/materialize)
"$scratch/bin/saved-compare" --config "$scratch/saved-config.json" --config-sha256 "$saved_config_sha" --out "$scratch/saved-comparison.json"
cmp "$scratch/saved-comparison.json" "$bundle/evidence/SAVED-COMPARISON.v1.json"
"$scratch/bin/materialize" --config "$scratch/data35-config.json" --config-sha256 "$data_config_sha" --out "$scratch/materialized"
[[ $(wc -c < "$scratch/materialized/train.jsonl") -eq 48618 ]] || fail 'reproduced data35 size'
cmp "$scratch/materialized/train.jsonl" "$bundle/data/train.jsonl"
jq -e --arg sha "$data_config_sha" '.config_sha256==$sha' "$scratch/materialized/MATERIALIZATION.v1.json" >/dev/null
jq -e '.config_sha256 | test("^[0-9a-f]{64}$")' "$bundle/MATERIALIZATION.v1.json" >/dev/null
jq -S 'del(.config_sha256)' "$scratch/materialized/MATERIALIZATION.v1.json" > "$scratch/receipt.actual.json"
jq -S 'del(.config_sha256)' "$bundle/MATERIALIZATION.v1.json" > "$scratch/receipt.frozen.json"
cmp "$scratch/receipt.actual.json" "$scratch/receipt.frozen.json"
printf '%s\n' 'next60 development35 and saved GJSON offline reproduction verified'
