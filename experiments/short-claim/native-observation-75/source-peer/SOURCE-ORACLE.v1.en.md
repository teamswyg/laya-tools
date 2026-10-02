# Wrap / Marshal source-before-Want notes

These source conclusions are frozen before reading the observer author's Want, dispatch, or handoff. No library import, initialization, API, tests, or binary was executed. Reviewer `semantic_review60_prep` is independent of this observer's authorship. This is AI-assisted reading, not a human-only oracle or runtime observation.

## WrapString

`go-wordwrap/wordwrap.go:16–82` iterates runes and increments `wordBufLen`/`spaceBufLen` by one. It does not count UTF-8 bytes, graphemes, or display cells. A single-rune Unicode character or tab counts as one. Unicode whitespace enters the space buffer except NBSP U+00A0, which stays part of a word (:47–62). LF has its own branch: it is always written and resets the line count (:26–46).

The exact insertion condition is `current + wordBufLen + spaceBufLen > lim && wordBufLen < lim` (:64–69). When true, a newline is written before the pending word and pending spaces are discarded. Words are not split; the source warning about long words exceeding the limit (:10–15) matches the implementation. Width 0 never satisfies `wordBufLen < lim`, so it inserts no automatic newline. Width 1 also never satisfies that condition for a nonempty word. An oracle promising every line fits every width is incorrect.

Whitespace is neither universally retained nor universally trimmed. Pending whitespace is dropped on an inserted wrap. When there is no pending word before LF, spaces are written only if they fit the current width (:27–35). At end of input with no pending word, trailing spaces are also written only if they fit (:73–76). With a pending word, both spaces and word are written (:77–79). Leading whitespace may trigger an inserted newline. Empty input returns an empty string. The API returns only a string, with no separate error or partial-result field. It has no mutation path for the input string or external receiver; buffers are local.

Finite expectations must follow rune order, both buffers, and that exact condition. Unicode display-width wrapping, universal length bounds, and preservation of original invalid UTF-8 bytes must not be inferred. Finite fixture conclusions are not broad contract or training-label approval.

## Marshal

`godotenv/godotenv.go:183–193` constructs complete output lines, sorts them with `sort.Strings`, then joins them with LF. This is lexicographic order of complete lines, not map iteration order or an unconditional general key-order promise. No final newline is added. Nil and empty maps both return `"", nil`. The function reads the map, creates a new line slice, and neither writes the map nor reads/mutates process environment. Its explicit return always has a nil error.

`isInt(:164–178)` removes at most one leading minus, then requires a nonempty remainder consisting only of ASCII `0` through `9`. Plus, whitespace, fractions, and Unicode digits fail. Leading zeros, minus-zero, and arbitrarily long digit strings pass; no numeric conversion or overflow test occurs. Passing values are unquoted. Others use double quotes around `doubleQuoteEscape`. The comment suggesting every value is quoted (:181–182) is therefore insufficient as an implementation oracle.

Escape order is backslash, LF, CR, double quote, exclamation, dollar, and backtick (:26, :235–246). LF/CR become literal `\n`/`\r`; other listed characters acquire a backslash. Original backslashes are doubled first, preserving the distinction between literal escapes and real line breaks. This does not mean every whitespace or single quote is escaped. Keys are neither validated nor escaped, so arbitrary-key parseability or roundtrip is not guaranteed.

`parser.go` provides related escape and expansion context, but Marshal does not call `parseBytes`, `expandEscapes`, or `expandVariables`. Unmarshal/Write/Load facts cannot substitute for direct Marshal observation. Parser regex initialization was not executed in this review.

## Source and license preservation

The complete `go-wordwrap/LICENSE.md` and `godotenv/LICENCE` were read; both contain MIT permissions, copyright and warranty notices. Preserve complete notices and exact references for source-derived distribution. This does not establish human-only authorship, source diversity, independence, or training readiness. Exact source/license file hashes and byte counts are frozen in SOURCE-RECEIPT.
