# Observing original functions against frozen expectations

Before a tiny `riidolaya` claim/hint model can suggest which candidate to verify first, its captions need behavioral evidence. This bundle records **two development requests**: paginated Union directory reads and JSON duplicate-key rejection. It is an original Go-library observation, not a model decision or a new training result.

## Actual observations

| Measure | Actual value |
| --- | ---: |
| Requests / frozen inputs / candidate captions | 2 / 13 / 6 |
| Saved and verified candidate-input rows | 39 |
| Frozen Wanted comparison | 24 pass / 12 mismatch / 3 unknown |
| Directly instrumented original methods / owned callbacks | 96 / 89 |
| Durably acknowledged frames | 528 |
| Whole-child elapsed time | 2.748737667s |
| macOS child maximum RSS, measured after Wait | 12,189,696B ≈ 11.63MiB |
| Retained executable | 4,575,474B ≈ 4.36MiB |
| Evidence file / journal | 14,303B / 404,858B |
| New Laya inference / Fit / protected rows | 0 / 0 / 0 |

This is one finite observation including process setup and Sync. It is not a repeated benchmark or Laya latency/memory measurement. GPU execution is unverified. `GOMEMLIMIT=64MiB` is advisory Go heap guidance, not a hard RSS limit. Startup and nested original-call counts remain unknown.

The39 rows are `6×3 + 7×3`. Inputs, candidates and translations do not create independent requests. Existing connected groups79/83 keep their `development_train` roles. New independent source families, the new20→60 independent cohort and protected2400 evaluation additions are all zero. Whole-work gains of at least5% and monetary savings remain unestablished.

## What the two requests mean

The first request merges overlapping lists, then reads them in multiple batches. Caching the merged list once and moving a cursor passed all six frozen inputs. Reading a positive count from the original on every call or rereading the entire list from the beginning had omissions, repetitions or different return contracts.

The second request rejects duplicate keys within each JSON object before values collapse. `a` and `\u0061` decode to the same key. Tracking decoded keys separately per object and stopping at the first duplicate passed all seven inputs. Comparing raw spelling misses the escaped alias. Converting to a Map first loses duplicate information and lacks error/location/stop channels. These are differences from the requested contract, not claims of upstream library bugs.

| Request | Candidate0 | Candidate1 | Candidate2 |
| --- | --- | --- | --- |
| Union | 6T → T | 2T·4F → F | 3T·3F → F |
| JSON | 4F·3U → F | 6T·1F → F | 7T → T |

T means all frozen predicates were confirmed. F means at least one known counterexample; associated U diagnostics remain. Unknown-only candidates do not become negative labels. [Saved comparison](NATIVE-CACHE-RESET-WANTED-COMPARISON.actual.public.v1.json) retains each predicate. Label qualification and dataset materialization are separate steps.

## Failure and correction

The first original run failed before saving its first row. Our observer reused the full cache result as the destination for a partial result, leaving inactive entries behind. Resetting the destination to `InfoReturn{}` fixed it. Six regression controls use nonempty lists. The failed executable and partial journal remain preserved. The corrected worker ran once into fresh outputs without modifying frozen Wanted or upstream source.

The comparer also separates known null from unknown, retains recovered original panics as independent counterexamples, and validates every row before advancing state. Owned fake controls remain distinct from actual original-library observation.

## Reproducing the public controls

With Go1.27.1, run:

```sh
bash scripts/verify-next60-native-wire3.sh
```

The script reconstructs earlier public wire3/core/file/journal modules in temporary directories, applies the correction, then runs race, `panicnil=1`, vet and formatting checks. Default CI uses owned fake APIs and temporary files. It does not download or execute original libraries/models or train. Check the PR separately for actual new CI status.

`source/controller` starts and reaps one original child and checks evidence, journal, ACK and genuine EOF. `source/comparer` uses saved evidence only. `source/core-overlay` contains the cache reset and regression controls. `source/archive` preserves past executables. `source/compiler` and `source/native-worker` support optional macOS original observation; default controls do not run it. This research bundle is not automatically registered in the main `riidolaya` command.

Original reproduction needs its own resource admission, frozen revisions/licenses/file SHAs and fresh output paths. Templates refer to separately supplied pinned `originals` directories. Executables, original source bodies, raw journals and host paths are not published to GitHub. Owned observer/comparer/archive source does not replace upstream licensing obligations.

## Recoverable space management

To preserve the failed worker while compiling a corrected one, two completed historical GJSON executables were archived. Raw6,339,316B became gzip2,753,963B, reclaiming3,585,353B before control metadata. The helper compares32KiB blocks directly with the still-present source and checks length, SHA, CRC, ISIZE, a single member and EOF before durable COMMIT and one unlink.

This is **stream comparison**, not an actual disk restoration. It is separate from the previous44 physically restored archives. Partial failures stay preserved and retries are not automatic. The cumulative512MiB retained research-file cap continues to apply; it is distinct from the Mac's total RAM or disk capacity.

[Run evidence](NATIVE-CACHE-RESET-RUN.actual.public.v2.json) · [First failure](NATIVE-FIRST-RUN-FAILURE.actual.public.v1.json) · [Archive readback](NATIVE-COLD-STREAM-READBACK.actual.public.v1.json) · [Issue19](https://github.com/teamswyg/laya-tools/issues/19) · [한국어](README.ko.md)

After an independent implementation matched all819 predicates, Root qualified six labels for the two requests. The [37-row development successor](../next60-development-thirtyseven/README.en.md) was generated in Go and matched all37 Reader values and bytes. The68 new unknown predicates remain linked to the prior selected scope46; new Fit remains zero. Default CI reproduces saved data and owned controls only.
