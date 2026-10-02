# Reader74 independent saved-record review

The saved execution records pass the consistency check. This is not a second numeric-array read or a rerun of a reader, codec, observer, or model. Reviewer `semantic_review60_prep` is independent of reader/controller authorship, with prior codec and preflight exposure; the review is not blind across the whole pipeline. Reading and review were AI-assisted.

The frozen plan differs from the original draft only in the single `Frozen: false` to `true` token. The earlier independent preflight receipt remains unchanged. Controller final/partial records and each child's final/partial records are byte-identical. Recorded JSON and compact starts are one each, with zero retries, exit 0, no timeout, and returned `Wait`. JSON preflight/materialization each show 1 attempt/1 return; compact read/Decode each show 1/1; validation shows 1/1 per child. Completion state and both success fields were checked, rather than treating exit 0 alone as success.

All 24 column summaries—four datasets by six columns—match in count, nil flag, and hash. Row counts are 91/45/73/44, with 67,104/37,975/37,916/37,657 nonzeros. All columns in this snapshot are nonnil. Source JSON, metadata, and complete-bit digests agree. The recorded lengths yield 1,814,648B of numeric arrays plus 163,758B metadata = the reported 1,978,406B owned payload, within 64MiB. This accounting differs from the original projection's payload measure.

| Saved measurement | JSON | compact |
| --- | ---: | ---: |
| Whole-child OS real/user/sys, seconds | 0.49 / 0.08 / 0.01 | 0.05 / 0.01 / 0.00 |
| Controller child wall, seconds | 0.50100075 | 0.058869916 |
| Maximum RSS, B | 19,120,128 | 19,300,352 |
| Peak footprint, B | 16,269,840 | 15,745,528 |
| Final Go HeapAlloc, B | 4,660,664 | 4,765,512 |
| Cumulative Go TotalAlloc, B | 44,840,120 | 21,608,800 |

OS numbers match the saved raw time logs. Compact's cumulative allocation was lower in this execution, but maximum RSS was 180,224B higher and final heap was 104,848B higher. RAM savings are therefore not established. Displayed `0.00` time is rounded, not proof of zero CPU use.

Controller hash reads warmed files before the fixed JSON-then-compact order, with one child per format. OS metrics cover the whole child lifetime, including source/file pin checks, reading, validation, hashing, and reporting. Shorter time was observed, but this does not establish a causal 10x format speedup, cold I/O, large-data scaling, or serving performance. HeapAlloc is a final heap observation; TotalAlloc is cumulative allocation, not peak heap or RSS. Go soft 256MiB and 64MiB owned-payload bounds are not hard OS RSS limits.

One stdlib saved-record helper attempt succeeded, checking 184 conditions and 16 saved-file pins. Array digests were not independently recomputed. Digest interpretation relies on the earlier static source review and must not be described as a second independent numeric-array validation. The minimal environment is declared in the root invocation record; this reviewer did not independently observe historical process environments or counts.

New source API calls, input-array reads, reader/parser/codec executions, retests, Features, Project, Fit, roles, labels, models, and HF operations are all zero. Training readiness, performance approval, and protected final-2400 completion remain false. Zero model/paid calls means no separate inference/API execution; it does not mean AI-assisted collaboration was free.
