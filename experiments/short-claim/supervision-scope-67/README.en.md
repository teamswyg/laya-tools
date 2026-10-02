# Supervision scope inventory67 with original truth preserved

`riido-supervision` is a Go command that joins saved requests, candidates, truth and content assessments to **propose which candidates can contribute to loss for the original target**. It neither trains nor runs a model. It preserves original truth and the full candidate set instead of turning unsupported descriptions into new negatives.

The target is a small hint model linking the **complete explicit conditions** in finite English requests to candidate descriptions. A request with multiple acceptable answers retains the complete acceptable set. `no_answer` retains its original negative candidates; `unknown` receives no label. A description covering only part of a request does not allow borrowing existing function truth for a different target.

## First actual inventory

The parent execution ledger records one successful metadata inventory, zero failures and zero retries. The counts below come from the unchanged [original result](results-67.json) and [completion ledger](ROOT-ACTUAL-EXECUTION-LEDGER-67.v1.json).

|Item|Original population or proposed result|
|---|---:|
|Parent requests / candidate positions / whole groups|72 / 216 / 17|
|Known or no_answer parents / unknown parents|51 / 21|
|Original entries / supplemental audit entries|738 / 432|
|Verified files / bytes / review records|21 / 8,586,507 / 1,170|
|Eligible positive / negative candidates|35 / 95|
|Known candidates retaining labels with mask=false|23|
|Unknown candidates with nullable labels|63|
|Groups with any eligible candidate|15|
|Groups with any eligible positive candidate|12|

`35 + 95 + 23 + 63 = 216`: every candidate position remains. All9 known candidates in group52 retain mask=false and are not deleted. The original population has16 groups containing stored known truth. That count differs from15 groups with any eligible candidate. The12 positive-bearing groups are descriptive; they do not introduce a new training floor.

This is **proposed supervision metadata**. Actual feature projection, role assignment, seed selection, fitting, model/paid calls, original source behavior APIs and protected final reads all remain0. Both `training_ready` and `authoring_diversity_cleared` remain false. Role-specific eligible coverage and the existing utility/headroom calculation for the same scope have not run. The result must not trigger mask relaxation or favorable seed search.

## Usage

Run from the repository root with Go1.27.1. Each of the21 original JSON inputs must match its fixed bytes and SHA. This directory's archive alone does not substitute for those inputs.

```sh
go run ./cmd/riido-supervision \
  --input-root . \
  --output supervision-67-new.json \
  --plan-sha256 c9a07943f8047b69e32d93354e812ca5b158b1dc7476264bd5b8505a86242408
```

To use a binary, build it as follows. Python, network requests and model downloads are not runtime requirements.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o riido-supervision ./cmd/riido-supervision
./riido-supervision --help
```

`--output` must be a new file. Existing files are never overwritten, and failure causes no automatic retry. Inputs are read only as regular files within the read-only root. The loader enforces2MiB per file and16MiB total. These are payload read bounds, not process RSS or Go heap limits.

People can read fixed help and failure codes; agents can read the exit status and JSON envelope. Success is exit0 with `passed_metadata_proposed_supervision_only`. Validation failure is exit1 with a `failed` envelope containing an empty report and a valid-prefix verification ledger. Invalid flags or missing required arguments yield exit2. An output reservation or persistence failure can leave no complete envelope. Diagnostics do not echo supplied paths or argument values.

## Proposed mask rule

The parent request contract, prototype contract observation, candidate source closure and caption fidelity must all be `consistent_with_scoped_evidence`. Original label1 additionally needs consistent candidate coverage; label0 needs coverage equal to `contradicts_scoped_evidence`. Omission, uncertainty, pending states or coverage inconsistent with saved truth produce mask=false.

Original acceptable sets, candidate order, evaluation denominators and whole groups stay intact. Known/no_answer0/1 labels are not rewritten; unknown labels stay null. Supplemental observations and negative boundaries are audit records, not overrides of original coverage or truth. A source-closure-only record does not become caption fidelity approval. Public source names, digests and pointers are provenance alongside the data, not runtime features or a scalar approval of a candidate.

Step67 creates no fit rows. It reports the later projection contract: retain masked known candidates with weight0 and retain unknown only for audit. Dropping part of a group must not be used to meet role or fit floors. The original56 plan's9/3/3 role floors and utility5% for the same scope are unchanged. This step adds no requirement for2400 examples in every experiment or any new diversity count.

## Resource observation and limits

The original private binary's single cold metadata process measured real0.71s, user CPU0.13s, system CPU0.01s and peak RSS27,508,736bytes. A separate peak memory footprint metric was24,887,824bytes; it is not the same metric as RSS. Raw diagnostic logs remain private, with only their hashes in the completion ledger.

These measurements do not cover model inference, model memory, GPU memory, Go heap, warm router latency or a general performance distribution. The public port has only been built and tested with synthetic fixtures: original-input runs0. The actual one-run inventory used the original private binary.

Normal joins and counts use arrays, slices and sorting without new locks. Reproducing individual review-record canonical JSON SHA values uses the existing hash recipe's map representation. This step contains no benchmark establishing SIMD or inference improvements.

The AI-assisted assessment and loader author had seen original assessments and truth. Synthetic checks are author-side verification, not an independent origin or blind evaluation. The loader verifies saved bindings and mask rules; it does not establish the assessments' semantic correctness, source diversity or training readiness. Ordinary AI collaboration cost is unmeasured.

## Artifacts and reproduction scope

- [Original frozen plan](PLAN-67.ko.md): unchanged content and bytes.
- [Original result](results-67.json): exact original envelope copy.
- [Original reservation](ROOT-INVOCATION-RECEIPT-1.json) and [completion ledger](ROOT-ACTUAL-EXECUTION-LEDGER-67.v1.json): preserve their distinct reservation and completion states.
- [Preparation ledger](PREPARATION-LEDGER-67.json): preserves the historical pre-execution count0. Read the completion ledger for current execution counts.
- `prototype-*.txt`: exact text archives of the original helper, synthetic tests and module file. No binary is published.
- [Public invocation description](INVOCATION-PUBLIC-67.json): a new artifact replacing private paths with placeholders. Its SHA is not claimed to equal the private invocation plan.
- [Port ledger](PORT-LEDGER-67.json) and [artifact provenance](PROVENANCE-67.json): distinguish bytes and hashes of original, public-port and transformed artifacts.

This is a maintainer metadata tool. Codex registration remains a separate opt-in; the port changes no current router or model. CI checks and publication follow the repository's existing process.
