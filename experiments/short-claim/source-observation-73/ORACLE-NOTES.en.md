# Independent oracle notes for three sources — before native execution

This source/text review is independent of observer authoring. I read the three pinned upstream files and the source-build handoff; I have not read the observer's `pure/spec.go`, Want, or execution code. The handoff's broad contract drafts were visible, so that part is nonblind. These are **independent predictions/scope proposals**, not frozen official Want, observed Got, new training labels, or readiness approval. Original APIs, package/import initialization, upstream tests, binaries, models, and fitting remain unexecuted in this review.

## datasize: magnitude, error identity, and receiver state are separate

`UnmarshalText` lines116–217 reads an ASCII integer prefix followed by a unit. Units use powers of two (lines13–19), not decimal1000. It trims the unit suffix, but not leading numeric whitespace; fractions/signs are unsupported. The exact forms `Kb/Mb/Gb/Tb/Pb/Eb` branch to the bits error before lowercasing, so `1Mb` and `1mb` differ. Empty input skips the loop and succeeds as zero.

| Raw example | Predicted receiver decimal | Predicted error |
|---|---:|---|
| `2 KB` | 2048 | none |
| `1Mb` | 0 | ErrBits |
| `1mb` | 1048576 | none |
| empty string | 0 | none |
| ` 1KB` or `1.5MB` | 0 | ErrSyntax |
| `18446744073709551615` | 18446744073709551615 | none |
| `18446744073709551616` | 18446744073709551615 | ErrRange |
| `18014398509481983KB` | 18446744073709550592 | none |
| `18014398509481984KB` | 18446744073709551615 | ErrRange |

Use independently authored integer arithmetic for n×unit and the0..2^64−1 boundary, then check suffix/error precedence against source branches. Preserve uint64 or decimal strings; float64 loses integer precision. Error observations must separate `*strconv.NumError`, `Func="UnmarshalText"`, `Num=the complete original input`, and the actual `Err` sentinel identity, rather than comparing only messages. Syntax/bits set the receiver to0; range sets maxUint64 (lines206–216). A universal preserve-on-error/atomic-state contract **contradicts this source**. Initial receiver7 exposes that distinction. Input-byte snapshots differ from receiver snapshots; direct `uint64(receiver)` conversion needs no additional upstream helper calls. Nil receivers, unsafe aliases, and additional Parse/MustParse/String APIs are excluded.

## querystring: key presence differs from per-key value order

The primary oracle is `url.Values`: **key absence/presence plus exact value-slice order, duplicates, and empty strings**. Sorting keys for reporting is harmless; sorting/deduplicating slices or checking only Get's first value loses behavior. Encoded URL ordering is not directly specified by this API, so the documentation's encoded example alone is not the map oracle.

For fixed exported primitive/value-struct fixtures: `url:"-"` and empty+omitempty are omitted; untagged empty strings retain `[""]`; **empty slices/arrays are omitted even without omitempty** (lines216–220). Default slices preserve repeated-value order through Add (lines248–253), as do multiple declared fields sharing a key. Booleans default to `false/true`, with the int option `0/1`. Nested `user.name` becomes `user[name]`, not a flattened key. Nil scalar pointers without omitempty yield an empty string; nonnil pointers to zero survive omitempty because the pointer itself is nonempty (lines185–186,208–214,283–290,321–335).

Top-level non-struct inputs return **nil map+error** (lines140–141). Nil interfaces/typed nil pointers return a **nonnil empty map+nil error** (lines126–137). There is no receiver. Input snapshots for fixed primitive fixtures are sufficient; do not expand that to arbitrary custom-object immutability.

`Values` returns its accumulated map together with a reflectValue error (lines144–145), but those error points depend on custom Encoder and nested/embedded propagation (lines201–204,264–277). A callback-free fixture must not pretend to observe **partial map+error**. Custom Encoder/IsZero/Stringer, cycles, arbitrary reflection types, and time options are outside the bounded draft. The broad omitempty description does not establish universal key presence for empty slices.

## shlex: return completed prefix, not unfinished word

The source uses fixed ASCII separators(space/tab/CR/LF), quote modes, escape, and comment states. Full POSIX execution/substitution/escape semantics are not an oracle. Prefer **manual state-boundary reasoning and fixed literal token arrays** to a reference that simply duplicates the implementation.

`one "two three" four` gives `one`, `two three`, `four`; `'' ""` yields two empty tokens; `a""b` yields one `ab`. A start-state `#` begins a comment, but the in-word `#` in `a#b` remains literal. Backslash is literal inside single quotes; in double quotes/outside it makes the next rune literal. A generic shell library or strings.Fields is therefore unsuitable as a universal oracle.

Later malformed quotes/dangling escapes return **already completed tokens+error**. `ok "unfinished` gives `["ok"]`+quote-EOF error. An input ending in a backslash immediately after `ok`, with no separating whitespace, has no completed token and gives `[]`+escape-EOF error. A token must reach whitespace/normal EOF to count as completed. Tokenizer constructs partial tokens on errors, but Lexer.Next discards them (lines154–156), and Split returns its previous slice (lines403–415). Appending the incomplete word or deleting the previous prefix both contradict the source.

Quote-EOF and escape-EOF use two fixed source messages (lines284/302,320/345). They are not exported sentinels, so a newly created same-message error does not prove identity. Split converts terminating io.EOF into nil error. Empty/whitespace/comment-only input returns a **nonnil empty slice+nil error**. This API has no caller receiver and accepts a string. Injected Reader errors, direct Tokenizer calls, nil Readers, and broader Unicode whitespace are excluded.

## Minimal comparison when author Want is frozen

Keep normal values, error class/identity, nil-versus-empty, completed prefix, and receiver changes in separate fields. Finite probes do not immediately establish whole-source behavior or candidate satisfaction labels; multiple probes do not multiply independent source families. Next compare the frozen author spec/Want against this unchanged memo, reporting concrete source-backed differences before original execution.

Source pins: datasize `51293273…583d9`(aa82cc1e…); query `644ee90c…83ff7`(965d79f2…); shlex `f34d676e…220b`(e7afc7fb…). `SOURCE-REVIEW-RECEIPT.json` preserves full hashes, line references, and scope. No upstream/handoff authorship or human-only diversity claim is made. AI-assisted reading cost was not measured.
