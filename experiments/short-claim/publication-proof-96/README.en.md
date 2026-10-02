# Public research archives and readback verification

The first FP32 model failed utility. To preserve that evidence, the [Hugging Face research archive](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33) contains original coefficients, numerical results, bilingual documentation, license notices and checksums. Model bytes are excluded from Git. All9 project files,135,328B including the32,792B model, matched after a fresh download. Tag `failed-fp32-71-v1` resolves to the same commit. The public research collection preserves its original15 items and adds this one failed archive.

Archiving is separate from qualification. The27-to31 check regression, Top1 5-to2 and all original false qualification/production/publication flags remain unchanged. There is no automatic pipeline, widget, default activation or paid serving. The [publication and readback proof](HF-PUBLICATION-71.v1.json) binds immutable commit and file hashes.

PR94's [Korean Wiki](https://github.com/teamswyg/laya-tools/wiki/First-Claim-Fit-71-KO) and [English Wiki](https://github.com/teamswyg/laya-tools/wiki/First-Claim-Fit-71-EN) were published and checked against CI-qualified documents using remote Git bytes. The [Wiki proof](WIKI-PUBLICATION-94.v1.json) preserves the successful helper verification. Root's first readback used an incorrectly transcribed Wiki commit and safely failed; a second check used the actual remote commit and passed. There was one Wiki push. The earlier failure is not erased into the successful helper record.

The first Issue19 comment comparison failed because the CLI added a trailing newline; exact raw-body readback subsequently matched. The comment was posted once. Before HF publication, one CI JSON assertion used fields from a different response shape and failed; the corrected assertion passed before any remote mutation. These metadata handoff failures are not training retries or changes in model utility.

[한국어](README.ko.md)
