# Pre-fix file labels isolated from search inputs

Project only `instance_id` and `patch` from the two Parquet sources pinned in experiment 28. Under the [precommitted plan](label-plan-33.json), labels remain evaluation-only, outside queries, candidate construction, hint features and training. Catalog acquisition never reads patches.

| All 2,400 selected tasks | Result |
|---|---:|
| Parse failures / unsupported blocks | 0 / 0 |
| Tasks with old-file paths | 2,398 |
| Tasks without old-file paths | 2 |
| Total old paths / maximum per task | 3,912 / 30 |
| New-file blocks / duplicate paths | 64 / 0 |

The two tasks without old paths each add only one new file. Do not replace them or remove them from the denominator. The eventual evaluation policy must distinguish old-file localization applicability from all-2,400 results. The [public aggregate](label-results-33.json) contains no paths, task IDs or patch bodies.

The Go parser reads pre-hunk old-file headers, rename metadata and unambiguous binary/mode headers. `/dev/null` denotes a new file. Body text cannot become a header label. Spaces, quoted Git escapes and Unicode are supported; ambiguous/invalid paths are not guessed. This is a path-metadata parser, not a complete Git syntax or patch-applicability verifier. Changed files are not exhaustive relevance labels or ground truth for difficulty, success or decomposition.

Bound patches to 2MiB and relative paths to 4,096 bytes. Never open files or execute/apply patches. Clone short paths to avoid retaining the whole raw patch backing string. Verify projection SHA-256, allowed fields, unique IDs and row bounds before use; preserve failure rows.

Only the maintainer Parquet converter uses existing PyArrow; Go execution has no Python requirement. The converter checks source hashes, row counts and column types, writes exclusive 0600 files, and excludes questions and execution scripts. Raw patches and `labels.json` remain local.

```sh
go run ./cmd/riido-filelabels --out .cache/file-labels-new
```

Pinned query and separate patch projections are required. A new directory receives public-safe `results.json` and private `labels.json`. Both reproduce byte for byte. Tests cover header/body isolation, new files, renames, binary/Unicode paths, duplicates, invalid paths and schema/hash checks. Maintainer converter isolation, permissions and overwrite guards passed locally with PyArrow; CI does not gain a PyArrow dependency.

Actual catalog membership and ranking have not yet been evaluated. Source acquisition is complete; comparison follows the separately frozen scoring plan. Existing reserves and models are unchanged.
