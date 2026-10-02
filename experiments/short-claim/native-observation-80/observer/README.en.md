# Observation80 pre-execution preparation archive

This archive preserves preparation for a **Go native CPU observer** covering three behavior goals in 24 finite fixtures: 8 UUID Parse inputs, 8 UUID Scan inputs and 8 original Ordinal inputs. It is not Laya/GPU/model inference or a promotion of training data. Original preparation documents are preserved byte-exact. This archive does not read or report Root's later actual execution results.

New Go and module source are archived as `.go.txt` and `.mod.txt`, so they do not activate a public package. Copyright, comments and source bodies are unchanged. Historical source versions remain inert text. Relative paths in the [original handoff](HANDOFF.v1.json) describe the original private workspace filenames. Public filenames are mapped by `source_path`→`archive_path` in the [copy ledger](COPY-LEDGER.v1.json). The original handoff was not rewritten to pretend it had used this public layout.

| Preparation scope | Record |
|---|---|
| Logical behavior goals / source families / finite fixtures | 3 / 2 / 24 |
| Expected callbacks | 24 direct APIs + 5 explicit Error calls = 29 |
| Conservative callback bound | At most 24 Error calls, 48 tracked dispatches |
| Original namespace initialization | 4 static sites; actual callback/return counts null |
| Preparation checks | 3 pure race invocations, 3 vet invocations, 2 compile-only builds PASS |
| Original init/API/test/worker execution during preparation | 0 |
| New labels/roles/weights/Features/Fit/models/protected-final reads | 0 |

29 and 48 are an expectation and an upper bound, not measured counts. Unexpected errors retain their actual dynamic types rather than being falsely assigned a known class. Undispatched Got/Matches remain null. UUID arrays, Scan receiver before/after, nil interfaces, typed nil bytes and nonnil empty bytes remain distinct. Panic formatting does not dispatch extra Error/String methods. Failures, differences and the limits of the last successful checkpoint remain recorded. The [preparation guide](HANDOFF.v1.en.md) and [ledger](PREPARATION-LEDGER.v1.json) explain these boundaries.

The [UUID source and complete BSD-3-Clause license](../upstream/google-uuid/LICENSE) and [Ordinal slice and complete MIT license](../upstream/humanize-ordinal-slice/LICENSE) remain in the existing [source archive](../README.en.md). The selection contains 15 fixed-revision UUID runtime files with the original go.mod, and one original humanize ordinals.go with its original go.mod. The 20 original assets and [source-first records](../preparation/SOURCE-FIRST-80.v1.json) were verified by hash and byte equality without copying them again. The observer's project-authored code uses the host repository's Apache-2.0 terms, which do not relicense upstream BSD/MIT code. This does not clear the full humanize package or other borrowed origins in `number.go`. The existing [NOTICE](../NOTICE.md) remains part of the archive.

Want is an [exact copy](WANTS.v1.json) sealed by a separate author. That Want author's [observer mechanics review](peer-static-review/RECEIPT.v1.json) is nonblind source reading, not a new independent oracle or human-blind test. Exact pins for the separate source-only Want peer and existing source archive are in [external references](EXTERNAL-REFERENCES.v1.json). Repeated fixtures do not become independent parents, and training/production qualification remains false.

Draft/frozen plans containing host paths, binaries, private helpers and raw test/build metadata are excluded from publication. The ledger retains their hashes and exclusion reasons. CPU 1, a 256 MiB Go soft heap setting, a 60-second outside timeout and no retries are settings, not a measured RSS hard limit. Original package initialization occurs before main. This archive does not replace Root's execution decision or prove generalization, performance or cost savings.

[ARCHIVE-MANIFEST](ARCHIVE-MANIFEST.v1.json) and [SHA256SUMS](SHA256SUMS) bind the archived bytes. Publication is not complete. This work only prepares a private stage and verifies metadata; actual results, workers, original tests and models were not rerun.
