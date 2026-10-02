# Source-first80: UUID Parse/Scan and small Ordinal

Prepare **three behavior goals, two source families and24 input drafts**. UUID Parse and Scan share one package/helper family; Ordinal comes from another upstream source. No Wants, adapter, binary, actual parents, labels, roles or weights exist. See [SOURCE-FIRST-80.v1.json](SOURCE-FIRST-80.v1.json) and the input-only [INPUT-DRAFTS.v1.json](INPUT-DRAFTS.v1.json).

| Goal | Eight proposed inputs | Later observations to preserve |
|---|---|---|
| UUID Parse | raw32/standard36 valid and late hex cases; URN45 prefix casefold/wrong prefix; braced38/other outer characters | Complete returned16-byte UUID, return/error/panic separately |
| UUID.Scan | nil interface, empty string, typed nil and nonnil empty bytes, valid text, raw16, late bad text and bytes | Fresh nonnil receiver before/after, input type/nilness, errors/panic |
| Ordinal | 0,1,2,3,11,12,13,112 | Input int, complete output text, panic status |

These are proposed inputs, not expected outputs. Proposed direct callbacks are Parse8 + Scan8 + Ordinal8 =24. **Source comparisons** with MustParse, Validate or Unmarshal are not observations of those APIs. Recursive Scan, xtob and observation-time Error() methods differ from direct callbacks. A later author must freeze transformations and the final call budget.24 fixture slots are neither24 independent parents nor an actual invocation count.

## Distinct returned values and receiver state

At [the pinned UUID revision](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d), Parse returns a fixed16-byte value. raw32 assigns xtob's calculated byte **before** checking success; standard36 checks success **before** assignment. Equal decoded prefixes do not imply equal failed-slot bytes or zero everywhere after the prefix. The machine proposal binds exact byte spans, type/table/helper declarations and complete source hashes.

Wrong URN prefixes fail before decoding. At length38, Parse skips the first byte and decodes the inner36 positions without checking the outer characters as braces. Validate performs that check. Do not preclassify square-bracket input as the same invalid input for every API. The existing79 generic invalid-input wording is not a universal RFC-validator contract; a later request must specify its finite length/hex/prefix policy. Historical drafts remain unchanged.

Scan returns before assignment on nil or empty input. Its comment mentioning a null UUID does not reset its receiver. It parses strings into a local value and commits only without an error, so Parse's partial UUID and Scan's receiver state must remain separate. Exactly16 input bytes copy directly; other byte lengths delegate to Scan(string(src)). `%v` formats an error without a `%w` wrapping promise. UUID.Scan and NullUUID.Scan have different state/reset contracts.

Database/driver callbacks, nil receivers, unsupported dynamic types, exhaustive length/separator/Unicode cases and aliasing remain outside these eight fixtures. Error type/message and panic serialization are not yet Wants. Preserve returned array bytes directly instead of silently calling UUID.String as an observer helper.

## Candidate source closure and initialization

All sixteen retained UUID runtime source files were read. The Darwin non-JS candidate contains15 files, excluding node_js.go, plus the original go.mod and complete BSD-3-Clause license. Raw hashes match historical definition literals; known inventory53 Git blobs also match. This is a source-file candidate, not an executed compiler/go-list selection or hermetic proof of the complete standard-library closure. Original tests are unnecessary for this observation and were not executed or reused as new expectations.

**Import can execute original functions before main.** Four namespace variable initializers in hash.go call MustParse→Parse. Four is a source-read call-site count. Without instrumenting original initialization, do not claim observed actual init callbacks or returns4. The future external dispatcher reserves before child start and records `static startup sites reserved4 / actual init callbacks not instrumented` separately from direct probes. An in-main token check cannot prevent initialization.

Other globals include the hex table, error sentinels, rand.Reader reference, pool/locks, clock/node values and jsonNull. Selected Parse/Scan source paths call no entropy, time, network-interface or random-pool API, but that does not establish zero full-package/standard-library startup work. Measure OS CPU/RSS/wall over **whole child lifetime**, including initialization, setup, probes, observation and persistence.

## Explicit original Ordinal slice

[Original Ordinal(int)](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/ordinals.go) calls only strconv.Itoa. The selected file has no package variables or init functions. Propose an **explicit source slice** retaining unchanged ordinals.go, original go.mod and the full MIT license. This does not observe the complete upstream package or its initialization/footprint. Its provenance notice must identify the revision, raw hashes, single-function package composition and exclusion of added Ordinal64, other files and tests.

The original signature accepts int and has no nonnegative guard. Choosing small nonnegative values0..112 does not redefine negative inputs as unsupported or repair their existing suffix behavior. Source edits0. The teen exception is modulo100;112 helps distinguish that condition from a description checking only the literal12. No suffix Wants or model runs were produced.

A full-humanize profile would require other missing bodies and number.go's gorhill/WTFPL origin notices. Those remain known-missing pointers. Files unnecessary for the selected slice were not acquired or copied; new acquisition0.

## Rights, exposure and limits

UUID's complete **BSD-3-Clause LICENSE is1,480B**, SHA `0a8d61ed3cbfd5312326e8126c31ce9c627a283adc99131b56896d29ada04b2d`. Humanize's complete **MIT LICENSE is1,136B**, SHA `a973b4498c13eb74baa2a8e5c351426a6826f2fcdd909916dbe53ee2e755fd71`. Root's preliminary “UUID full MIT” wording was corrected before freeze. Existing sources, notices and expectation records remain untouched.

Both sources have inventory53/taskverify and prior55 coding-task exposure. Four started CLI attempts in55 are a historical pilot, not new80 execution. This reader authored source-caption74, saw76 results and read historical task-definition/acceptance provenance after original source reading. The later Want author is separate; the entire pipeline is not blind, human-only or untouched final.

Actual parents/labels/roles/weights0; qualification, training, production and protected-final flags false. Keep both UUID goals, helpers, aliases, locales and repeated fixtures in one family. Two proposed families are not proved independent author/role groups. Proposed future bounds are one worker, CPU1, a256MiB Go soft heap,60 seconds and1MiB output. A soft heap is not an OS RSS hard cap. Original import/init/API/tests/native, models/fits and shared edits are0.24 input drafts do not replace the untouched2,400-per-domain final target.
