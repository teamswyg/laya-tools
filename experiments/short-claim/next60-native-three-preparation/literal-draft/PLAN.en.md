# Pre-observation literal and ABI draft for three contracts

Status: **DRAFT, not adopted by Root, execution_ready=false**. No truth, label, role, weight, numeric group or qualification is assigned. The prior 22 hypotheses, five retirements and three HOLDs remain unchanged.

The selection uses existing ordinals **2 → 7 → 1**: three requests in two whole-source families. Four Complement inputs, five JSON inputs and four original Binary profiles plus one separately declared padding profile give **14 finite inputs**, nine candidates and 42 proposed dispatches. Actual observations are zero. Intended contract implementations occupy positions 0/1/2; position and name do not confer a label or reference privilege.

Complement inputs use original `New(0)` followed by listed `Set` operations. Stored lengths are 0/1/65/66, versus explicit widths 0/1/64/65. The latter two differ. Literal Wants include exact result length, words/set bits and an unchanged copied receiver snapshot. Borrowed Words are immediately copied into fixed arrays, never retained.

JSON requires a complete document, root Parse and a Number integer lexeme without a dot or exponent. All five old input objects remain unchanged. Exponent and malformed-document cases are not among these five, so that coverage is not claimed. Existing Int already has a Raw parseInt path returning `9007199254740993` exactly. The changed contract is checked rejection. Its original error channel is unavailable: **unknown**, not a fabricated nil or false.

Binary uses default BigEndian, at most 128 declared bits and at most 24 exact encoded bytes. The four textual profiles receive explicit BE64 hex and the same nonempty receiver. Len1/word3 is a **new fifth padding profile**, separately declared before observation. Huge headers are excluded before original dispatch and cannot count as observed baseline failures. Transactional rejection and error-channel availability remain separate.

Read the [plan](LITERAL-PLAN.proposed.v1.json), [draft Wants](WANTS.proposed.v1.json), [ABI/availability](ABI-AND-AVAILABILITY.proposed.v1.json), [source/license pins](SOURCE-PINS.v1.json) and [successor scope](SUCCESSOR-AND-SCOPE.v1.json) together. Original bodies and full BSD/MIT notices remain in existing private archives; none are copied here. Host locations occur only in PRIVATE-BASIS.

Owned Go files are inert `.go.txt`. Native binder, worker, durable checkpoints, compiler/linker, runtime budget and Root's independent Want freeze remain separate pending steps. The separate oracle preparation has the same author and is not independent oracle verification. Go/gofmt/build/test, original API/init, model/Fit and network executions are all zero. This draft never authorizes a first run.
