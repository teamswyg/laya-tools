# Pre-reference family structure

`riido-familycohort` checks one explicitly supplied metadata bundle with exactly
400 Source bindings, 1,200 declared family/frame bindings (three per Source), and
2,400 supplied comments (one `ko` and one `en` per family). It generates no
content and runs no model or reference judgment. Whole Source QA remains an
external prerequisite for actual dependent authoring.

```sh
go build ./cmd/riido-familycohort
./riido-familycohort --root ./supplied-cohort --input bundle.json \
  --sha256 "$BUNDLE_SHA256" --bytes "$BUNDLE_BYTES" --out ./structure-out
```

Supply the bundle's independently fixed SHA-256 and exact encoded byte count.
The command captures this one file once with a synced `reviewpacket` start/result
pair. It checks the actual captured hash and length, strict UTF-8 and closed JSON
shape, then joins all declared IDs. It emits aggregate JSON on stdout; diagnostics
contain no supplied text, IDs or paths. The output directory must be new. Files
are synced and created exclusively with mode 0600, in directories with mode 0700.
Success retains the exact captured bytes in `BUNDLE.private.json`, their binding
and summary in `STRUCTURE.private.json`, and `SUMMARY.json`. Failed checks retain
their summary and any capture receipts; they export no validated structure.

Every field below is mandatory. Unknown, duplicate or case-alias keys, nulls,
missing fields, trailing JSON and invalid Unicode escapes are rejected.

- Bundle: `schema` = `riido-familycohort-bundle-v1`, `frame_schema` (File),
  `frame_schema_version`, `sources`, `families`, `comments`.
- Source: `source_id`, `split`, `dev_style`, `source` (File), `source_review` (File).
  Splits are `train`, `dev`, `cal`, `test`; DEV style is `short` or `general` on
  `dev` and the empty string elsewhere. There are no new split or label quotas.
- Family: `family_id`, `source_id`, `frame` (File), `frame_schema` (File).
  Its schema pin must exactly equal the bundle's supplied schema pin.
- Comment: `comment_id`, `family_id`, `source_id`, `source` (File), `frame` (File),
  `locale`, `text`. The Source ID/file and frame file must match its joined family
  and Source. Split and DEV style are inherited through that join.
- File: `path`, lowercase `sha256`, `bytes`. Paths use clean root-relative `/`
  spelling; absolute paths, parent traversal, `\`, `:` and aliases such as `./`
  are rejected. Byte counts are nonnegative; zero bytes require the empty-file
  digest. One path cannot declare conflicting hashes or byte counts.

IDs use 1–128 ASCII letters, digits, underscores or hyphens. The supplied frame
schema version is a nonempty UTF-8 string of at most 128 bytes, without surrounding
whitespace. Source/family/
comment IDs must each be unique; unknown links, missing families and repeated
locale slots reject the whole bundle. Successful output includes inherited
split/style counts, rather than independent family or comment assignments.

The encoded bundle has an **8 MiB** ceiling, while each complete decoded comment
retains a separate **4,096-byte UTF-8** allowance. Text must be valid, nonempty
and free of NUL. The command never truncates or normalizes it. All 2,400 comments
at their individual maxima can exceed the bundle ceiling: the initial transport
rejects that entire oversized bundle, without reducing the per-comment limit or
selecting a subset. Private artifacts are bounded at 16 MiB each and total
output at 64 MiB. Check actual bytes before claiming transport capacity.

Only the bundle is opened. Source, Source-review, frame and frame-schema Files
are opaque, format-checked declarations. Shared review/schema/frame files are
allowed; the checker does not resolve individual records or snapshots inside
them. Equal hashes/text do not establish independence or semantic duplication.
The supplied schema pin/version is bound, but its contents and adoption are
unverified. The eight proposed QA axes are later assessment dimensions, not
required frame fields; this tool does not invent axes, frame values or head labels.

Success means `STRUCTURAL_ONLY_QA_PENDING`. Whole Source QA, read provenance,
frame-schema semantics, meaning, provider identity/independence, rights, Source
and bilingual fidelity, naturalness, blind references, reference support and all
workflow gates remain pending. The output proves no semantic eligibility,
training authority, human Gold or model quality and must not replace those gates.

The in-memory Go API is `familycohort.Validate(Bundle) (Summary, error)`; it joins
declarations without opening files or changing caller slices. `Decode([]byte)`
adds the encoded transport bound and closed shape; `Run(root, File, out)` adds
the one actual integrity capture and private retained output. The APIs do not
register integrations or publish automatically. Tests use software fixtures only.
