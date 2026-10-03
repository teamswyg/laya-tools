# Reproduce and read 23 development requests in Go

We actually materialized **23 development requests with 67 candidate labels (+23/−44)** in Go and verified that the project's reader preserves text, labels, weights, and source metadata for every row. [data/train.jsonl](data/train.jsonl) is **31,493 bytes**. The previous21-row, 28,800-byte prefix remains byte-exact.

The purpose is to test whether a small model can hint at candidate verification order. This does not establish final approval, certified truth, or prose-generation ability. The tools copy supervision already adopted by Root from finite execution evidence; neither tool assigns new labels.

## Use

From the repository root:

```sh
bash scripts/verify-next60-extra-two.sh
bash scripts/verify-next60-twentythree.sh
```

The first check recomputes the additional two requests' saved finite comparisons. The second restores authored Go modules into a temporary directory, runs synthetic controls, appends two rows to the existing21, compares data bytes, reads all23 through the real project reader, and compares the stored reader result. Humans and agents use the same commands and failure exit codes. These checks require no Python, original-candidate execution, Laya download, inference, or Fit. Go1.27.1 and system C tools for race checks are required.

| Observed check | Value |
| --- | --- |
| Distinct requests | 23 |
| Candidate labels | 67 (+23/−44), unit weights |
| Frozen input variants | 113 |
| Retained original / selected observations | 335 / 331 |
| Actual reader calls, returns, value matches | 23 each |
| Acknowledged before/after reader checkpoints | 46 |
| Retained unknown predicate in the additional two requests | 1 (`B_code`) |
| New Fit or model calls | 0 |

[ROOT-ACTUAL-CORRESPONDENCE.v1.json](ROOT-ACTUAL-CORRESPONDENCE.v1.json) records the first actual materialization and reader run. [MATERIALIZATION.v1.json](MATERIALIZATION.v1.json) and [INPUT-VALIDATION.v1.json](INPUT-VALIDATION.v1.json) are the tool returns. The reader Summary's `output_*` flags are materializer-specific and remain false. ReaderRecorder separately reserves fresh files and confirms46 checkpoints followed by final Sync/Close/directory Sync; those Summary flags are not a reader persistence failure.

`source/` contains authored Go sources restored by the check. [SOURCE-ARCHIVE.v1.json](SOURCE-ARCHIVE.v1.json) maps archive locations to original source fingerprints. `preparation/` and `peer/` retain **pre-execution** preparation and separate source review. Their zero actual-generation counts describe that historical stage and are distinct from the later first-execution evidence. The public bundle excludes raw reader journals, host execution paths, binaries, and models.

## Limits

The reference occupies position2 in20/23 requests;44/67 candidate labels are negative. These differently denominated position/class controls are not model accuracy. Both new requests reuse connected semantic group77; more requests do not mean more source domains. Features must use request/candidate text without labels or source IDs.

A new Fit begins only after30 qualified requests. The60-request development checkpoint and2,400 protected evaluation requests per claimed domain remain separate plans. This is not protected evaluation or evidence of user cost/time savings. This first-execution record predates CI119 and HF23 publication; consult actual publication evidence for their status.
