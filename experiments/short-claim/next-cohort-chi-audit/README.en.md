# Chi routing order and a small claim model

[한국어](README.ko.md) · [Optional Go API](../../../../pkg/hintprepared/README.en.md) · [CI](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml)

This experiment tests how similar code descriptions can hide a consequential difference in operation order. A small model proposes the order of candidate checks; the actual contract checks make the final judgment. Scores are not probabilities.

**Result: Go API calculation parity passed, but the existing model's recommendation quality was inadequate.** It placed the candidate passing all checks fifth; BM25 placed it first. The previously failed model stays inactive. No fitting, new model publication or default-policy change occurred.

## The problem

Consider receiving the name `a%2Fb` through `GET /u/{id}`. Decoding the whole path first produces `/u/a/b`, adding a routing segment. Matching first and decoding only the captured name preserves a single name, `a/b`.

The contract matches RawPath first when available. Only when decoding is opted in and original RawPath is nonempty does it PathUnescape the matched id once. It updates both chi.URLParam and Request.PathValue, preserving original URL Path and RawPath. With no RawPath or opt-out, captured values remain unchanged.

| Candidate | Mechanism | Contract passes out of 9 |
|---|---|---:|
| post_path | One post-match PathUnescape | 9 |
| post_query | One post-match QueryUnescape | 8 |
| pre_path | Decode the whole path first on a cloned request | 8 |
| twice_path | Two post-match PathUnescape passes | 8 |
| unchanged | Keep captured names encoded | 6 |

QueryUnescape changes literal `+` into space. Twice decoding removes `%2F` that must remain after one pass over `a%252Fb`. Decoding the whole path first turns an encoded slash into a routing separator. These are different normal results conflicting with the requirement, rather than errors or panics.

The nine inputs cover encoded slash, nested encoding, plus, space, opt-out, plain/no-RawPath, literal percent/no-RawPath, and unmatched paths. This finite experiment explicitly sets URL fields. Network request parsing, malformed escapes, regex, mounts, concurrency and hidden whole-request state are outside its scope.

## What ran

Local CPU execution used Go1.27.1 and pinned Chi sources. Each trial created a fresh request/router and captured both parameter APIs inside the live route or NotFound handler. Literal Wants were authored before execution and independently reviewed from source.

All 45 trials returned normally. All 533 explicitly wrapped API calls returned, with no errors or panics. The ledger excludes hidden calls and field writes. Complete observation JSON is 80,431 B, stored as 2,575 B gzip; CRC, full EOF decode, full JSON and recomputed counts were checked. See [raw observations](route/OBSERVATIONS.actual.public.v1.json.gz) and [literal-Want comparison](route/COMPARISON.actual.public.v1.json).

A separate run read the existing inactive FP32 claim model once and compared the reference calculation with the [Prepared API](../../../../pkg/hintprepared/README.en.md). All five score bits and order indices matched exactly, and actual product input validation passed. These are existing hashed-feature claim weights. No Laya encoder, MPS/GPU execution or training occurred here.

| Ordering method | First candidate | Rank of the 9/9 candidate |
|---|---|---:|
| Existing FP32 model | pre_path | 5 |
| Prepared API with the same weights | pre_path | 5 |
| BM25 | post_path | 1 |
| Ordered lexical control | post_path | 1 |
| Fixed input order | post_path | 1 |

Fixed order is a control. narrow_rule rejects this natural-language grammar with `unsupported_rule_request` and uses BM25 fallback; it is not a separate rule-model success. Ranks 5 and 1 are not measured LLM calls or task-completion times. [Complete scores/order](preview/OBSERVATIONS.actual.public.v1.json) and [comparison with explicit fallback](preview/COMPARISON.actual.public.v2.json) preserve all candidates.

## Reproduce

From the repository root, using Go1.27.1:

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/replay
```

Use `-go` and `-root` for a different Go executable or packet location. The verifier checks fixed source/notice/input pins, compiles in an owned temporary directory, compares the replay against public observations and literal Wants, then removes the temporary executable. No model download or authentication is required. Output reports one parent, 45 trials, 533 explicit calls and the 9/8/8/8/6 matrix.

An actual local invocation matched the full observations and left no new temporary directory. See the [completed execution](replay/EXECUTION.actual.public.v2.json). The same CPU replay step has been added to CI; check the PR checks for its execution and outcome. CI success covers code checks and this finite contract's reproducibility. Model generalization, corpus admission and default activation have separate criteria. Existing Laya CI and this Chi replay execute different targets.

The first replay attempt failed because the verifier's temporary-directory setting caused Go to ignore its module. Repair review also found that a promoted `ReadFrom` could bypass the output cap. The module root and compiler temporary directory are now separate, and the buffer is a named field. Actual controls checked exact-fit and oversize output through both copy paths. The [first source](replay/FIRST-REPLAY-FAILURE.source.go.txt) and [failure receipt](replay/EXECUTION.actual.public.v1.json) remain available. Chi inputs, literal Wants and the model comparison were unchanged.

## Cost and memory interpretation

Execution receipts preserve Darwin's whole-process timing and raw maximum resident-size report. They include startup, file reads, routing or model calculations, and output/encoding. No units were converted, and the two different process values are not a model-memory reduction comparison. CPU time rounded to 0.00 does not mean zero CPU work. GPU, MPS and pprof were not measured in these runs.

GOMAXPROCS and soft Go heap targets were bounded. GOMEMLIMIT does not cap RSS or system memory. Temporary compilation trees were removed; no executable copy was retained. The unchanged 512 MiB retained logical-file budget was maintained by compressing closed public copies with full-byte recoverability. That file budget is separate from RAM.

## Sample and next tuning criteria

**This is one parent.** Nine inputs, five mechanisms, 45 trials and six ordering records are not 45 independent Golden tasks or newly qualified independent parents. It is an exposed development audit and cannot move to protected-final evaluation.

The PDCA finding is that encoding order and guards must be distinguished better than surface word overlap. Before training, match captions to code and review related parents/children/helpers as a whole group. Add qualified parents from materially different sources, then tune on development data. Criteria include actual verification cost before finding the right candidate versus BM25. Independent 2400 evaluation and whole-task/LLM savings remain unproven.

Caption normalization also exposed a boundary: punctuation/camel splitting gave 40/40/41/40/34 words in the original draft, exceeding the 32-word product limit. Separate v2/v3 inputs fixed the bounds and clarified the guard, decoded Path assignment and raw-routing fallback before any model run. Original inputs were preserved; nothing was silently truncated or relabeled. The final request has 31 normalized words and candidates 27/27/32/27/24, accepted by actual product validation.

## Sources and licenses

[Pinned go-chi/chi](https://github.com/go-chi/chi/tree/167e1e3bd039d060696b99c8da4e876ae04f42c1) remains MIT with its complete notice. tree.go's Armon Dadgar radix attribution and the [complete Armon MIT notice](source/ARMON-NOTICE.txt) are retained. The Armon revision is a notice comparator, not an asserted historical predecessor or copied extent. Owned observation code and authored inputs/captions follow repository Apache-2.0; see [NOTICE](NOTICE).

The [immutable HF model revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46) holds the model body. No model body, private locator, profiles or raw execution journals are included here. [PACKET](PACKET.public.v1.json), execution freezes and SHA256SUMS connect public files and pins. A freeze's pre-execution/pending text records its historical state; EXECUTION/COMPARISON records the completed runs.

Scanning decompressed evidence flagged the public Go filename `dir_plan9.go` as a false positive. Its exact value and pinned inventory context were verified before a rerun with a temporary configuration limited to these inputs. Repository scanning rules were unchanged and the rerun found no leaks. The [first scan](replay/SCAN-DECOMPRESSED-FIRST-FAILURE.actual.public.v1.json) and [disposition/rerun](replay/SCAN-DECOMPRESSED-DISPOSITION.actual.public.v1.json) are preserved.
