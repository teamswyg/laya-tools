# Pre-dispatch Want80: independent source-based review

**No concrete semantic blocker requiring a pre-dispatch Want change was found in the 24 frozen records.** This is a source-based expectation review of these inputs and channels. It does not approve actual Got equality, the observer implementation or compiled closure, broad truth, or training-label eligibility.

Reviewer: `semantic_review60_prep`. I authored source-first80, so the source notes are not independently authored; I did not author the Wants (`checkpoint_cli_45`) or observer (`task_expansion_52`). Prior source, caption, and failed-fit exposure makes this AI-assisted development review nonblind. I read the expectations against the source rather than treating the author's CHECKS as independent semantic verification.

| Scope | Source-supported expectation |
|---|---|
| Parse 0,2,4,6,7 | Successful 16-byte value and nil error. The 45-byte branch case-folds the prefix; the 38-byte branch only skips its first byte, so `[…]` is accepted by the selected Parse body. Do not substitute Validate's contract. |
| Parse 1,3 | The same late `ge` failure returns `00112233445566778899aabbccddfe00` on raw32, which assigns before checking ok, and `00112233445566778899aabbccdd0000` on standard36, which checks before assigning. Preserve both the failure slot and following byte. |
| Parse 5 | The wrong nine-byte URN prefix fails before decoding: all-zero UUID, value type `uuid.URNPrefixError`, message `invalid urn prefix: "bad:uuid:"`. |
| Scan 8–11 | Nil interface, empty string, typed nil bytes, and nonnil empty bytes retain distinct input metadata and leave the original receiver unchanged. The ‘null UUID’ comment is not an assignment. |
| Scan 12,13 | Valid text commits after successful Parse; raw16 bytes copy directly. Each fixture must start with a fresh nonnil literal receiver. |
| Scan 14,15 | Parse's local partial value is not committed after an error. Both text and recursive non16-byte paths retain the receiver and return `Scan: invalid UUID format`. Error identity and Unwrap are not observed here. |
| Ordinal 16–23 | The original int function's modulo10/modulo100 conditions and Itoa result support `0th,1st,2nd,3rd,11th,12th,13th,112th`. No error return exists, so even error_nil is null. Small nonnegative inputs do not change negative behavior or introduce an API. |

The main evidence is pinned UUID `uuid.go` (type20, sentinels51–52, URN error55–59, Parse95–145), `util.go` (xvalues20–40, xtob43–47), `sql.go` (Scan15–52), and pinned humanize `ordinals.go` (Ordinal8–25). Existing source-first `/source_clauses` byte spans were rechecked; the sentinel and URN Error body were also read in the same complete pinned source. The byte shift/or of table values255 for g and14 for e yields fe; this is static source reasoning, not an executed Got.

Concrete error types are scoped to the read Go1.27.1 errors.New/errorString.Error and fmt.Errorf sources. The two format errors and two Scan errors are `*errors.errorString`. Scan uses `%v`, not a `%w` wrapping promise. `%T` collection is separate from Error calls. Calling Error once on each of the five returned nonnil errors is a sound plan: one upstream URN method and four stdlib methods. Thus 24 direct entrypoints plus five direct Error observations give 29 tracked dispatches. Internal Scan Parse3, recursion1, formatting, and other stdlib calls are not individually instrumented return counts.

The four namespace initializers in UUID hash.go15–18 are **static call sites**. Actual init callback/return count remains null and instrumented=false. Root must reserve before child start because an in-main token check is too late to prevent import initialization. Whole-child CPU/RSS includes startup, setup, observation, and persistence; the256MiB Go soft heap is not an RSS hard cap. This review executes no original imports/init, compilation, tests, APIs, workers, models, or fits.

Metadata checks cover all24 fixture objects/IDs/original order/pointers,20 source/module/license SHA pins,16 source-clause and four startup spans, and two stdlib source SHA pins. Complete UUID BSD-3-Clause and Ordinal-only MIT licenses match the same pins. Whole humanize/WTFPL origins, additional APIs, and upstream tests remain outside scope. Got/truth/role/weight/acceptable stay null; qualification/training/production/protected-final stay false. Two source families do not establish two independent authors or training groups, nor replace2,400 unseen final requests. Observer review and later mismatch preservation remain separate.
