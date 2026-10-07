# Concept metadata audit

This maintainer command checks explicitly pinned provisional concept metadata
before wording production. It accepts a contiguous prefix of 1–10 forty-group
tranches: 40–400 groups and 120–1,200 bilingual families. It uses only Go's
standard library and performs no model calls, inference, assignments or case
certification. The [typed contract](SCHEMA.CONTRACT.TEXT-FREE.md) preserves the
existing metadata schema; the command does not convert other schemas.

Use Go 1.27.1 from the repository root. Supply the independently recorded,
lowercase SHA-256 for every clean relative `.json` path. Repeat `-input` for
each tranche, including every earlier file referenced by a later declaration.
Files are read only; declared source pins are never opened automatically.

```sh
mkdir -p reports
GOMAXPROCS=2 GOMEMLIMIT=512MiB GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
  go run ./cmd/riido-statehint-concept-audit \
  -input metadata/tranche01.json="$TRANCHE01_SHA" \
  -input metadata/tranche02.json="$TRANCHE02_SHA" \
  > reports/concept-metadata-audit.json
```

Input count, path/pin format and regular-file size are checked before content is
read, hashed or decoded. The limits are 204,800 bytes per file and 2,097,152 bytes
in total. Opened file identity, stat size and actual bytes must agree. Strict
UTF-8, duplicate/unknown/missing keys, exact JSON types, canonical IDs, false
metadata-boundary flags and exact prior-file references are checked. A retained
batch 1 correction permits 1–32 whitelisted field deltas and must reconstruct
the predecessor's exact bytes and SHA. Whole relation endpoint arrays must be
paired; current arrays and input files remain unchanged. The sole typed null
exception is `prior_correction_object: null` inside that correction.

Success writes JSON with aggregate counts, registered categorical tags, name
length/duplicate statistics, provisional review connectivity and caller-supplied
file paths/pins. It omits per-record concept prose, IDs, targets and raw input.
Failures use fixed diagnostic codes and exit 2. Reports are local files and are
not uploaded automatically. “Text-free” does not make a report automatically
safe to publish: review its categorical tags and supplied paths/pins first.

Metadata is not actual messages, human gold or evidence of model quality.
Boundary declarations cannot prove that concept prose contains no realization.
Unique IDs and graph components cannot establish semantic independence, lineage,
effective sample size, editorial quality, admission or study validity. Source
pins inside metadata are format-checked declarations, not verified ancestry.

Owned synthetic checks require no metadata or model files:

```sh
GOMAXPROCS=2 GOMEMLIMIT=512MiB GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
  go run ./cmd/riido-statehint-concept-audit -self-test
go test -race ./cmd/riido-statehint-concept-audit
go vet ./cmd/riido-statehint-concept-audit
```

The self-test and package tests exercise byte/stat preflight, strict decoding,
40–400 group joins, exact supplied references, correction reconstruction and
output boundaries. Synthetic results establish structural behavior only; a
successful self-test is not an audit of actual metadata or a CI result.
