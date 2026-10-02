# Source82: source preparation before execution

Missing source files were acquired so Source81's four behavior goals can later be checked against original Go code. No original function or model was executed at this stage. Acquisition, compilation, behavior verification and training readiness are separate milestones.

There were **46 new HTTP requests, all successful, with zero retries**: 44 source bodies and two metadata responses. Thirteen historical bodies were linked without changing their bytes. The retained source, module and license inventory contains **57 files and 332,806 bytes**. Source inspection proposes 48 runtime files for Go 1.27.1; this is not yet the compiler's actual selection or a linkage result.

Read the historical failure wording together with the [separate correction](CHRONOLOGY-CLARIFICATION.v1.en.md). The actual order was **46 successful HTTP requests → two record-writer compilation failures → successful third compilation and record generation**. Each failed writer made zero additional HTTP or original API calls and produced no output records. The original 72-file stage remains unchanged; the correction and its [receipt](CHRONOLOGY-CLARIFICATION-RECEIPT.v1.json) are separate additions.

## Evidence

- [Acquisition plan](ACQUISITION-PLAN-82.v1.json) and [actual result](ACQUISITION-RESULT-82.v1.json): frozen limits and complete request results
- [Source review](SOURCE-REVIEW-82.v1.json): build conditions, imports, copied code and observation scope
- [Original notices](NOTICE-82.md): mapstructure MIT, pflag BSD and BSD notices on copied Go code
- [Stage handoff](PUBLIC-STAGE-HANDOFF-82.v1.json) and [checksums](SHA256SUMS): unchanged historical preparation inventory

Fields such as `external_publication: 0` describe the historical preparation stage. This page and the correction are later additions, excluded from that stage's original checksum denominator. Its source-first README files also remain byte-identical.

The project's Apache license does not replace upstream MIT/BSD notices. Original headers and complete notices are retained, including Go Authors 2022 code in mapstructure and Go Authors code in pflag. An unresolved original Go revision for copied code remains an explicit gap. There are zero new actual user requests, labels, roles, weights or fits. These files do not satisfy the protected final minimum of 2,400 requests.
