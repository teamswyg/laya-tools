# Read-only assertion evidence shadow

This local Go command is an **unlearned research filter**. It records literal
reply/activity/closure candidates and supplied typed-event facts independently.
It uses no model, Fit, probability, app/DB client or task-state authority. It
cannot establish semantic quality,90% precision or Bloom-filter false-positive
bounds. Existing classifier confidence.9/margin.05 criteria stay unchanged;
this different evidence contract does not pass or replace those criteria.

Inputs are explicit local JSON byte pins, at most1MiB/100observations, each
text at most4,096UTF8bytes. The generic schema is:

```json
{
  "schema": "riido-assertion-shadow-input-v1",
  "observations": [
    {
      "id": "owned-text",
      "type": "text_observation",
      "role": "comment",
      "text": "Please explain the owned widget boundary.",
      "actor": "owned-actor",
      "unit": "owned widget",
      "timestamp": "2026-01-02T03:04:05Z",
      "version": "owned-v1",
      "source": "owned public pipeline fixture"
    },
    {
      "id": "owned-event",
      "type": "observed_event",
      "kind": "CommentObserved",
      "actor": "owned-actor",
      "unit": "owned widget",
      "timestamp": "2026-01-02T03:04:05Z",
      "version": "owned-v1",
      "source": "owned event replay"
    }
  ]
}
```

Text roles are `comment` or `body_snapshot`; they are never interchangeable.
Multiple paragraphs/chunks from a task body remain body observations, not
comments or independent examples. Unknown actor/unit/time/version is allowed
and recorded as uncertainty. A body-only input reports the missing-comment
input gap rather than inventing current work-state truth.

Typed facts are `CommentObserved`, `CommentCreated`, `ProgressObserved`,
`CommandSucceeded` and `TurnEnded`. The command records the caller's input
attestation; it does not verify authentication, membership, a production event
or database commit. Observing an existing comment is different from creating
one. No fact implies task completion. Authorized native owner-fixture review
supports read-observation ordering/skip/no-mutation boundaries only; the
public tests contain separately authored generic data. A synthetic fixture
replay is not a production event or human semantic truth.

The literal filter keeps exact source UTF8 **byte** offsets and substrings.
HTML attributes/tags and URL query punctuation are not message requests.
Quoted/code, future/conditional, negation, required-remainder or unresolved
context produces explicit uncertainty. Narrow direct reply/activity patterns
can become experimental display candidates; they remain unlearned opinions,
not reliable speech-act labels. Matching is incomplete and English/Korean
phrasing is deliberately limited. HTML entities are not rewritten, preserving
raw offsets. No confidence is fabricated. **Completion display is always withheld**,
even for an unambiguous-looking closure literal. No goal completion is claimed.

```sh
go run ./cmd/riido-statehint-assertion-shadow \
  --input OBSERVATIONS.json --input-sha256 INPUT_SHA --check

# Same pinned input; new private report directory, no app mutation.
go run ./cmd/riido-statehint-assertion-shadow \
  --input OBSERVATIONS.json --input-sha256 INPUT_SHA \
  --out .cache/statehint-assertion-shadow/NEW-RUN
```

Check mode writes no output. The new directory is0700 and `report.json`0600.
The report preserves input SHA, source/actor/unit/time/version, literal spans,
uncertainty, supplied event facts, body/comment counts and input gaps. It records
elapsed time and Go heap; heap is not OS peak RSS. Measure whole-command peak
RSS with an external process wrapper, separately from elapsed filter time.
Private inputs/evidence reports belong in ignored local storage, never Git/HF.

Owned tests cover created/observed≠done, exact Unicode/HTML spans, quotes/code,
negation/future/remainder, body-only gaps, independent candidates, input limits,
SHA/type checks and private exclusive outputs. No private source code, fixture
text, task identifiers or real prompt is included. This does not enable any
native comment pipeline or change application defaults, annotations, reactions,
labels, statuses, databases or classifier models.

Only aggregate observations from the local trials are published here. One
authorized private task body snapshot supplied 54 paragraphs as 54 body observations,
with 0 comments and 0 supplied events. Both of its 2 closure literal candidates were
withheld; display candidates were 0 and confidence was null. The body, identifiers,
URLs and evidence report remain private. This checks the input gap and read-only
behavior; it is not a semantic quality evaluation or 54 independent examples.

A separately authored original synthetic event replay supplied 3 comment texts
and 5 typed events, recording 2 reply/activity candidates. Completion display stayed
withheld, and the events had no completion authority. This synthetic trial does
not establish authentication, production behavior or native PostgreSQL fixture
truth. Both runs made 0 model calls, 0 Fit calls and 0 application state writes.

| Local run | Inner processing time | Whole command time | Externally measured peak RSS |
| --- | ---: | ---: | ---: |
| One private body input | 6.416ms | 0.60s | 10,846,208B |
| Original synthetic event replay | 0.358584ms | Below timer resolution | 6,455,296B |

These are single observations, not performance or accuracy guarantees. The
synthetic wrapper's displayed 0 does not mean a true zero duration. The goal of
useful progress, completion and question hints remains unmet. The production
comment pipeline stays OFF by default; this trial did not enable or deploy it.
